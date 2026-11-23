package control

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.einride.tech/pid"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Controller operating states reported in model.ControlStatus.State.
const (
	StateNormal  = "NORMAL"
	StateWarning = "WARNING"
	StateTripped = "TRIPPED"
)

// Controller implements dual closed-loop PID control (induced-draft blower VFD and quench spray
// valve) with low-pass draft filtering and fail-safe interlock enforcement.
//
// Direct-acting sign convention:
// go.einride.tech/pid computes e = ReferenceSignal - ActualSignal. Both kiln draft control
// (higher/less-negative pressure -> increase blower VFD) and quench cooling (higher outlet
// temperature -> increase spray valve) are direct-acting (PV - SP). Therefore both loops pass
// ReferenceSignal = actual (PV) and ActualSignal = setpoint (SP) so positive gains in
// model.DefaultControlConfig() produce negative feedback rather than runaway.
type Controller struct {
	mu           sync.RWMutex
	cfg          model.ControlConfig
	draftFilter  *DraftFilter
	interlock    *Interlock
	draftPID     pid.AntiWindupController
	quenchPID    pid.AntiWindupController
	draftBias    float64
	quenchBias   float64
	status       model.ControlStatus
	phWarning    bool
	phWarningMsg string
}

// NewController validates cfg and builds a ready-to-run closed-loop scrubber controller.
func NewController(cfg model.ControlConfig) (*Controller, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid control config: %w", err)
	}

	draftFilter, err := NewDraftFilter(cfg.DraftFilterCutoffHz, cfg.LoopInterval)
	if err != nil {
		return nil, err
	}

	interlock, err := NewInterlock(cfg.Safety)
	if err != nil {
		return nil, err
	}

	// Nominal operating biases placed inside (MinOutput, MaxOutput) so both positive and
	// negative errors immediately produce symmetric proportional action without startup dip.
	draftBias := cfg.DraftPID.MinOutput + 0.30*(cfg.DraftPID.MaxOutput-cfg.DraftPID.MinOutput)
	quenchBias := cfg.QuenchPID.MinOutput + 0.30*(cfg.QuenchPID.MaxOutput-cfg.QuenchPID.MinOutput)

	c := &Controller{
		cfg:         cfg,
		draftFilter: draftFilter,
		interlock:   interlock,
		draftPID:    newAntiWindupPID(cfg.DraftPID),
		quenchPID:   newAntiWindupPID(cfg.QuenchPID),
		draftBias:   draftBias,
		quenchBias:  quenchBias,
		status: model.ControlStatus{
			State: StateNormal,
		},
	}
	return c, nil
}

func newAntiWindupPID(g model.PIDGains) pid.AntiWindupController {
	return pid.AntiWindupController{
		Config: pid.AntiWindupControllerConfig{
			ProportionalGain:    g.Kp,
			IntegralGain:        g.Ki,
			DerivativeGain:      g.Kd,
			AntiWindUpGain:      g.AntiWindUpGain,
			LowPassTimeConstant: g.DerivativeFilterTimeConst,
			MinOutput:           g.MinOutput,
			MaxOutput:           g.MaxOutput,
		},
	}
}

// Step executes a single deterministic control calculation from a sensor snapshot and optional
// read error, updating internal filter, interlock, PID, and status state.
func (c *Controller) Step(readings model.SensorReadings, readErr error, dt time.Duration, now time.Time) model.ActuatorCommands {
	if now.IsZero() {
		now = time.Now()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if dt <= 0 {
		dt = c.cfg.LoopInterval
	}

	// 1. Evaluate fail-safe interlock on raw sensor readings first.
	ilState := c.interlock.Evaluate(readings, readErr, now)
	c.phWarning = ilState.PHWarning
	c.phWarningMsg = ilState.PHWarningDetail

	// 2. Update filtered draft only when raw draft reading is valid.
	var filteredDraft model.Pascal
	if readErr == nil && isFinite(float64(readings.DraftPa)) {
		filteredDraft = c.draftFilter.UpdateWithDT(readings.DraftPa, dt)
	} else {
		filteredDraft = c.draftFilter.Value()
	}

	// 3. Enforce fail-safe outputs whenever the interlock is latched.
	var cmd model.ActuatorCommands
	if ilState.BypassDamperOpen {
		// Prevent integrator windup while tripped.
		c.draftPID.Reset()
		c.quenchPID.Reset()
		cmd = model.ActuatorCommands{
			BlowerVFD:        0.0,                       // Shut down ID blower to isolate downstream CPVC/PP
			QuenchSprayValve: c.cfg.QuenchPID.MaxOutput, // Full spray cooling fail-safe
			BypassDamperOpen: true,
			AlarmLatched:     true,
			TripReason:       ilState.TripReason,
		}
		c.status = model.ControlStatus{
			State:             StateTripped,
			FilteredDraftPa:   filteredDraft,
			FilteredDraftInWC: filteredDraft.InWC(),
			Readings:          readings,
			Commands:          cmd,
			InterlockActive:   true,
			UpdatedAt:         now,
		}
		return cmd
	}

	// 4. Direct-acting PID updates: e = PV - SP => ReferenceSignal: actual, ActualSignal: setpoint.
	c.draftPID.Update(pid.AntiWindupControllerInput{
		ReferenceSignal:   float64(filteredDraft),
		ActualSignal:      float64(c.cfg.TargetDraftPa),
		FeedForwardSignal: c.draftBias,
		SamplingInterval:  dt,
	})
	blowerOut := clampOutput(c.draftPID.State.ControlSignal, c.cfg.DraftPID.MinOutput, c.cfg.DraftPID.MaxOutput, c.draftBias)

	c.quenchPID.Update(pid.AntiWindupControllerInput{
		ReferenceSignal:   float64(readings.QuenchOutletTemp),
		ActualSignal:      float64(c.cfg.QuenchSetpoint),
		FeedForwardSignal: c.quenchBias,
		SamplingInterval:  dt,
	})
	valveOut := clampOutput(c.quenchPID.State.ControlSignal, c.cfg.QuenchPID.MinOutput, c.cfg.QuenchPID.MaxOutput, c.quenchBias)

	cmd = model.ActuatorCommands{
		BlowerVFD:        blowerOut,
		QuenchSprayValve: valveOut,
		BypassDamperOpen: false,
		AlarmLatched:     false,
		TripReason:       model.TripNone,
	}

	state := StateNormal
	if ilState.PHWarning {
		state = StateWarning
	}

	c.status = model.ControlStatus{
		State:             state,
		FilteredDraftPa:   filteredDraft,
		FilteredDraftInWC: filteredDraft.InWC(),
		Readings:          readings,
		Commands:          cmd,
		InterlockActive:   false,
		UpdatedAt:         now,
	}
	return cmd
}

// Tick reads sensors from hal, runs Step, and applies the resulting ActuatorCommands to hal.
// Even when ReadSensors fails, Tick latches TripSensorFault and still drives fail-safe commands
// to hal before returning the status and any HAL error.
func (c *Controller) Tick(ctx context.Context, hal model.HardwareHAL, dt time.Duration) (model.ControlStatus, error) {
	if hal == nil {
		return model.ControlStatus{}, fmt.Errorf("hardware HAL is nil")
	}

	readings, readErr := hal.ReadSensors(ctx)
	if readErr != nil && ctx != nil && ctx.Err() != nil &&
		(errors.Is(readErr, context.Canceled) || errors.Is(readErr, context.DeadlineExceeded)) {
		return c.Status(), readErr
	}
	cmd := c.Step(readings, readErr, dt, time.Now())
	applyErr := hal.ApplyCommands(ctx, cmd)

	status := c.Status()
	if readErr != nil {
		if applyErr != nil {
			return status, fmt.Errorf("read sensors: %w; apply fail-safe commands: %v", readErr, applyErr)
		}
		return status, fmt.Errorf("read sensors: %w", readErr)
	}
	if applyErr != nil {
		return status, fmt.Errorf("apply commands: %w", applyErr)
	}
	return status, nil
}

// ResetInterlock attempts to unlatch a tripped interlock. It returns an error if any hard fault
// is still active. Upon successful reset, PID integrators and the draft filter are reset so
// closed-loop control resumes cleanly without windup or stale filter transients.
func (c *Controller) ResetInterlock() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.interlock.Reset(); err != nil {
		return err
	}
	c.draftPID.Reset()
	c.quenchPID.Reset()
	c.draftFilter.Reset()
	c.status.InterlockActive = false
	c.status.Commands.BypassDamperOpen = false
	c.status.Commands.AlarmLatched = false
	c.status.Commands.TripReason = model.TripNone
	if c.phWarning {
		c.status.State = StateWarning
	} else {
		c.status.State = StateNormal
	}
	return nil
}

// Status returns a thread-safe snapshot of the latest controller telemetry.
func (c *Controller) Status() model.ControlStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// InterlockState returns a thread-safe snapshot of the underlying interlock state machine.
func (c *Controller) InterlockState() InterlockState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.interlock.State()
}

// PHWarning reports whether a sump pH warning (non-tripping) is currently active along with its detail.
func (c *Controller) PHWarning() (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.phWarning, c.phWarningMsg
}

// Config returns a copy of the controller's validated ControlConfig.
func (c *Controller) Config() model.ControlConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cfg
}

// SetTargetDraftInWC updates the closed-loop kiln draft setpoint (in inWC, negative = suction)
// after validating it against ControlConfig rules.
func (c *Controller) SetTargetDraftInWC(inwc float64) error {
	_, err := c.SetSetpoints(&inwc, nil)
	return err
}

// SetQuenchSetpointF updates the closed-loop quench outlet temperature setpoint (in degrees F)
// after validating it against ControlConfig rules and MaxQuenchOutletTemp.
func (c *Controller) SetQuenchSetpointF(tempF float64) error {
	_, err := c.SetSetpoints(nil, &tempF)
	return err
}

// SetSetpoints atomically validates and applies optional runtime updates to TargetDraftInWC
// and/or QuenchSetpoint (specified in degrees Fahrenheit), returning the updated ControlConfig.
func (c *Controller) SetSetpoints(targetDraftInWC *float64, quenchSetpointF *float64) (model.ControlConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	next := c.cfg
	if targetDraftInWC != nil {
		next.SetTargetDraftInWC(*targetDraftInWC)
	}
	if quenchSetpointF != nil {
		if !isFinite(*quenchSetpointF) {
			return c.cfg, fmt.Errorf("quench_setpoint_f must be finite, got %v", *quenchSetpointF)
		}
		next.QuenchSetpoint = model.FahrenheitToKelvin(*quenchSetpointF)
	}
	if err := next.Validate(); err != nil {
		return c.cfg, fmt.Errorf("invalid setpoint update: %w", err)
	}
	c.cfg = next
	return c.cfg, nil
}

func clampOutput(val, minOut, maxOut, fallback float64) float64 {
	if !isFinite(val) {
		val = fallback
	}
	if val < minOut {
		return minOut
	}
	if val > maxOut {
		return maxOut
	}
	return val
}
