package control

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// InterlockState is an immutable snapshot of the fail-safe interlock state machine.
type InterlockState struct {
	BypassDamperOpen bool
	AlarmLatched     bool
	TripReason       string
	FaultActive      bool
	ActiveFault      string
	PHWarning        bool
	PHWarningDetail  string
}

// Interlock enforces hard fail-safe bypass-damper latching based on model.SafetyLimits.
// Hard bypass trips latch BypassDamperOpen and AlarmLatched until all hard faults clear
// and Reset() is explicitly invoked. Sump pH excursions produce a warning only and never
// open the bypass damper or latch a hard bypass trip.
type Interlock struct {
	mu                sync.RWMutex
	limits            model.SafetyLimits
	latched           bool
	tripReason        string
	faultActive       bool
	activeFaultReason string
	phWarning         bool
	phWarningDetail   string
	hasEvaluated      bool
}

// NewInterlock validates limits and returns a new fail-safe interlock state machine.
func NewInterlock(limits model.SafetyLimits) (*Interlock, error) {
	if err := limits.Validate(); err != nil {
		return nil, fmt.Errorf("invalid safety limits: %w", err)
	}
	return &Interlock{
		limits: limits,
	}, nil
}

// Evaluate inspects the latest sensor snapshot (and optional HAL read error) at timestamp now,
// updates hard-fault latching and pH warning status, and returns the resulting InterlockState.
func (i *Interlock) Evaluate(readings model.SensorReadings, readErr error, now time.Time) InterlockState {
	if now.IsZero() {
		now = time.Now()
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.hasEvaluated = true
	fault := i.detectHardFaultLocked(readings, readErr, now)
	i.activeFaultReason = fault
	i.faultActive = fault != model.TripNone

	if i.faultActive && !i.latched {
		i.latched = true
		i.tripReason = fault
	}

	// Sump pH out-of-range is warning-only; it never opens the bypass damper or latches a hard trip.
	if readErr == nil && isFinite(readings.SumpPH) {
		switch {
		case readings.SumpPH < i.limits.MinSumpPH:
			i.phWarning = true
			i.phWarningDetail = fmt.Sprintf("sump pH %.2f below minimum %.2f", readings.SumpPH, i.limits.MinSumpPH)
		case readings.SumpPH > i.limits.MaxSumpPH:
			i.phWarning = true
			i.phWarningDetail = fmt.Sprintf("sump pH %.2f above maximum %.2f", readings.SumpPH, i.limits.MaxSumpPH)
		default:
			i.phWarning = false
			i.phWarningDetail = ""
		}
	} else {
		i.phWarning = false
		i.phWarningDetail = ""
	}

	return i.stateLocked()
}

// Trip immediately latches the fail-safe bypass damper and alarm for the given fault reason
// (for example, when writing actuator commands to hardware fails).
func (i *Interlock) Trip(reason string) InterlockState {
	i.mu.Lock()
	defer i.mu.Unlock()

	if reason == "" || reason == model.TripNone {
		reason = model.TripSensorFault
	}
	i.hasEvaluated = true
	i.faultActive = true
	i.activeFaultReason = reason
	if !i.latched {
		i.latched = true
		i.tripReason = reason
	}
	return i.stateLocked()
}

// Reset clears a latched interlock trip only if all hard faults have cleared in the most
// recent evaluation. Returns an error and keeps the interlock latched if a fault is still active.
func (i *Interlock) Reset() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.latched {
		return nil
	}
	if !i.hasEvaluated || i.faultActive {
		reason := i.activeFaultReason
		if reason == model.TripNone {
			reason = i.tripReason
		}
		return fmt.Errorf("cannot reset interlock while fault %q is still active", reason)
	}

	i.latched = false
	i.tripReason = model.TripNone
	return nil
}

// State returns a thread-safe snapshot of the current interlock state.
func (i *Interlock) State() InterlockState {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.stateLocked()
}

// Limits returns a copy of the configured SafetyLimits.
func (i *Interlock) Limits() model.SafetyLimits {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.limits
}

func (i *Interlock) stateLocked() InterlockState {
	return InterlockState{
		BypassDamperOpen: i.latched,
		AlarmLatched:     i.latched,
		TripReason:       i.tripReason,
		FaultActive:      i.faultActive,
		ActiveFault:      i.activeFaultReason,
		PHWarning:        i.phWarning,
		PHWarningDetail:  i.phWarningDetail,
	}
}

func (i *Interlock) detectHardFaultLocked(readings model.SensorReadings, readErr error, now time.Time) string {
	// 1. HAL read error, non-finite sensor readings, or missing/stale/future timestamp.
	if readErr != nil {
		return model.TripSensorFault
	}
	if !readingsAreFinite(readings) {
		return model.TripSensorFault
	}
	if readings.Timestamp.IsZero() {
		return model.TripSensorFault
	}
	age := now.Sub(readings.Timestamp)
	if age > i.limits.SensorFaultTimeout || age < -i.limits.SensorFaultTimeout {
		return model.TripSensorFault
	}

	// 1b. Open-thermocouple / wire-break / under-range sensor plausibility guard.
	// Real 16-bit Modbus RTD/TC modules report finite sentinel values (e.g. 0x8000 / -3276.8 F or 0 F)
	// on open sensor; without a lower plausibility bound, the quench PID would drive spray to minimum.
	const (
		minPlausibleOutletTempK = 273.15 // 32 F / 0 C (wet quench tower cannot operate frozen)
		minPlausibleInletTempK  = 250.0  // -9.7 F
	)
	if float64(readings.QuenchOutletTemp) < minPlausibleOutletTempK ||
		float64(readings.QuenchInletTemp) < minPlausibleInletTempK ||
		float64(readings.WaterFlow) < 0 ||
		readings.WaterFlowGPM < 0 ||
		readings.BlowerSpeedFraction < -0.05 {
		return model.TripSensorFault
	}

	// 2. CPVC thermal protection: quench outlet temperature >= MaxQuenchOutletTemp (default 200 F).
	if readings.QuenchOutletTemp >= i.limits.MaxQuenchOutletTemp {
		return model.TripHighQuenchTemp
	}

	// 3. Quench water loss: water flow < MinWaterFlow (default 0.25 GPM).
	if readings.WaterFlow < i.limits.MinWaterFlow {
		return model.TripWaterLoss
	}

	// 4. Excessive draft magnitude: |DraftPa| > MaxDraftMagnitudePa.
	if math.Abs(float64(readings.DraftPa)) > float64(i.limits.MaxDraftMagnitudePa) {
		return model.TripOverDraft
	}

	return model.TripNone
}

func readingsAreFinite(r model.SensorReadings) bool {
	return isFinite(float64(r.DraftPa)) &&
		isFinite(r.DraftInWC) &&
		isFinite(float64(r.QuenchInletTemp)) &&
		isFinite(float64(r.QuenchOutletTemp)) &&
		isFinite(float64(r.WaterFlow)) &&
		isFinite(r.WaterFlowGPM) &&
		isFinite(r.SumpPH) &&
		isFinite(r.BlowerSpeedFraction)
}
