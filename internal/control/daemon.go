package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// ErrFaultInjectionUnsupported is returned by Daemon.InjectFault when the active HAL does not
// support interactive fault injection (for example, physical Modbus RTU hardware).
var ErrFaultInjectionUnsupported = errors.New("fault injection is not supported by active hardware HAL")

type faultInjector interface {
	InjectFault(action string, value float64) error
}

// DaemonOption customizes a Daemon constructed with NewDaemon.
type DaemonOption func(*Daemon)

// WithTelemetryCallback registers a callback invoked after every control loop tick with the
// latest model.ControlStatus snapshot.
func WithTelemetryCallback(fn func(model.ControlStatus)) DaemonOption {
	return func(d *Daemon) {
		d.onTelemetry = fn
	}
}

// WithErrorCallback registers a callback invoked when a HAL ReadSensors or ApplyCommands error
// occurs during a tick. The safety loop continues running and driving fail-safe commands.
func WithErrorCallback(fn func(error)) DaemonOption {
	return func(d *Daemon) {
		d.onError = fn
	}
}

// WithHALMode overrides the HAL mode label reported by Daemon.HALMode() (e.g. "sim", "modbus").
func WithHALMode(mode string) DaemonOption {
	return func(d *Daemon) {
		if trimmed := strings.ToLower(strings.TrimSpace(mode)); trimmed != "" {
			d.halMode = trimmed
		}
	}
}

// Daemon executes the closed-loop Controller against a model.HardwareHAL at cfg.LoopInterval
// until its context is canceled, fanning out live telemetry via an internal TelemetryHub.
type Daemon struct {
	controller *Controller
	hal        model.HardwareHAL
	halMode    string
	interval   time.Duration
	hub        *TelemetryHub

	mu          sync.RWMutex
	onTelemetry func(model.ControlStatus)
	onError     func(error)
	running     atomic.Bool
	ticks       atomic.Uint64
}

// NewDaemon validates cfg and constructs a Daemon bound to hal.
func NewDaemon(cfg model.ControlConfig, hal model.HardwareHAL, opts ...DaemonOption) (*Daemon, error) {
	if hal == nil {
		return nil, fmt.Errorf("hardware HAL must not be nil")
	}
	ctrl, err := NewController(cfg)
	if err != nil {
		return nil, err
	}

	mode := "custom"
	switch hal.(type) {
	case *SimHAL:
		mode = HALModeSim
	case *ModbusHAL:
		mode = HALModeModbus
	}

	d := &Daemon{
		controller: ctrl,
		hal:        hal,
		halMode:    mode,
		interval:   cfg.LoopInterval,
		hub:        NewTelemetryHub(defaultHubBufferSize),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}
	return d, nil
}

// SetTelemetryCallback replaces the telemetry callback in a thread-safe manner.
func (d *Daemon) SetTelemetryCallback(fn func(model.ControlStatus)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onTelemetry = fn
}

// SetErrorCallback replaces the HAL error callback in a thread-safe manner.
func (d *Daemon) SetErrorCallback(fn func(error)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onError = fn
}

// Run executes the closed-loop control ticker until ctx is canceled, returning nil on clean
// context cancellation. Concurrent calls to Run on the same Daemon return an error.
func (d *Daemon) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !d.running.CompareAndSwap(false, true) {
		return fmt.Errorf("control daemon is already running")
	}
	defer d.running.Store(false)

	if err := ctx.Err(); err != nil {
		return nil
	}

	// Execute an initial control tick immediately at startup.
	d.runTick(ctx, d.interval)
	lastTick := time.Now()

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	// On clean shutdown of a physical Modbus controller, drive fail-safe actuator outputs
	// (blower stopped, quench spray open, bypass damper open) so the PLC never stays latched
	// at its last running command without an active software supervisor.
	defer func() {
		if d.halMode == HALModeModbus {
			shutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			_ = d.hal.ApplyCommands(shutCtx, d.controller.FailSafeCommand(model.TripSensorFault))
		}
	}()

	maxDT := d.interval * 5
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			if ctx.Err() != nil {
				return nil
			}
			dt := now.Sub(lastTick)
			lastTick = now
			if dt <= 0 {
				dt = d.interval
			} else if dt > maxDT {
				dt = maxDT
			}
			d.runTick(ctx, dt)
		}
	}
}

func (d *Daemon) runTick(ctx context.Context, dt time.Duration) {
	status, err := d.controller.Tick(ctx, d.hal, dt)
	d.ticks.Add(1)

	// Non-blocking fan-out to all live telemetry subscribers (e.g. HTTP SSE clients).
	d.hub.Publish(status)

	d.mu.RLock()
	onTelemetry := d.onTelemetry
	onError := d.onError
	d.mu.RUnlock()

	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		if onError != nil {
			onError(err)
		}
	}
	if onTelemetry != nil {
		onTelemetry(status)
	}
}

// Subscribe registers a live telemetry stream subscriber and returns its channel and
// an idempotent unsubscribe function. Slow subscribers drop old samples instead of
// blocking the control loop.
func (d *Daemon) Subscribe() (<-chan model.ControlStatus, func()) {
	return d.hub.Subscribe()
}

// Hub returns the underlying TelemetryHub for live telemetry subscribers.
func (d *Daemon) Hub() *TelemetryHub {
	return d.hub
}

// Status returns the most recent model.ControlStatus snapshot.
func (d *Daemon) Status() model.ControlStatus {
	return d.controller.Status()
}

// Config returns a copy of the active validated model.ControlConfig.
func (d *Daemon) Config() model.ControlConfig {
	return d.controller.Config()
}

// HALMode returns the active HAL mode label ("sim", "modbus", etc.).
func (d *Daemon) HALMode() string {
	return d.halMode
}

// SupportsFaultInjection reports whether the underlying HAL supports interactive fault injection.
func (d *Daemon) SupportsFaultInjection() bool {
	_, ok := d.hal.(faultInjector)
	return ok
}

// Controller returns the underlying Controller instance.
func (d *Daemon) Controller() *Controller {
	return d.controller
}

// ResetInterlock delegates an interlock reset request to the underlying Controller and
// immediately publishes the updated status to live subscribers on success.
func (d *Daemon) ResetInterlock() error {
	if err := d.controller.ResetInterlock(); err != nil {
		return err
	}
	d.hub.Publish(d.controller.Status())
	return nil
}

// SetTargetDraftInWC updates the live kiln draft setpoint in inches of water column (negative = suction).
func (d *Daemon) SetTargetDraftInWC(inwc float64) error {
	return d.controller.SetTargetDraftInWC(inwc)
}

// SetQuenchSetpointF updates the live quench outlet temperature setpoint in degrees Fahrenheit.
func (d *Daemon) SetQuenchSetpointF(tempF float64) error {
	return d.controller.SetQuenchSetpointF(tempF)
}

// SetSetpoints atomically validates and applies optional targetDraftInWC and/or quenchSetpointF
// updates, returning the updated ControlConfig.
func (d *Daemon) SetSetpoints(targetDraftInWC *float64, quenchSetpointF *float64) (model.ControlConfig, error) {
	return d.controller.SetSetpoints(targetDraftInWC, quenchSetpointF)
}

// InjectFault triggers or clears a simulated fault on the underlying HAL when supported
// (e.g. "heat_spike", "water_loss", "clear_water_loss", "sensor_fault", "clear_sensor_fault",
// "sump_ph", "clear_all"). Returns ErrFaultInjectionUnsupported on physical hardware HALs.
func (d *Daemon) InjectFault(action string, value float64) error {
	fi, ok := d.hal.(faultInjector)
	if !ok {
		return ErrFaultInjectionUnsupported
	}
	return fi.InjectFault(action, value)
}

// Running reports whether Run is currently executing.
func (d *Daemon) Running() bool {
	return d.running.Load()
}

// Ticks returns the total number of control ticks completed by this Daemon.
func (d *Daemon) Ticks() uint64 {
	return d.ticks.Load()
}
