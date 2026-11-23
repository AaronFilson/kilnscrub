package control

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

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

// Daemon executes the closed-loop Controller against a model.HardwareHAL at cfg.LoopInterval
// until its context is canceled.
type Daemon struct {
	controller *Controller
	hal        model.HardwareHAL
	interval   time.Duration

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

	d := &Daemon{
		controller: ctrl,
		hal:        hal,
		interval:   cfg.LoopInterval,
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
	lastTick := time.Now()
	d.runTick(ctx, d.interval)

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

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
			}
			d.runTick(ctx, dt)
		}
	}
}

func (d *Daemon) runTick(ctx context.Context, dt time.Duration) {
	status, err := d.controller.Tick(ctx, d.hal, dt)
	d.ticks.Add(1)

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

// Status returns the most recent model.ControlStatus snapshot.
func (d *Daemon) Status() model.ControlStatus {
	return d.controller.Status()
}

// Controller returns the underlying Controller instance.
func (d *Daemon) Controller() *Controller {
	return d.controller
}

// ResetInterlock delegates an interlock reset request to the underlying Controller.
func (d *Daemon) ResetInterlock() error {
	return d.controller.ResetInterlock()
}

// Running reports whether Run is currently executing.
func (d *Daemon) Running() bool {
	return d.running.Load()
}

// Ticks returns the total number of control ticks completed by this Daemon.
func (d *Daemon) Ticks() uint64 {
	return d.ticks.Load()
}
