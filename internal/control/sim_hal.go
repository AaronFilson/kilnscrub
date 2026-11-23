package control

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// SimHALConfig parameterizes the simulated studio kiln + 4-stage scrubber plant dynamics.
type SimHALConfig struct {
	// NominalInletTemp is the baseline kiln exhaust gas temperature entering the quench tower.
	NominalInletTemp model.Kelvin
	// BaselineQuenchTemp is the adiabatic saturation equilibrium outlet temp at nominal load & flow.
	BaselineQuenchTemp model.Kelvin
	// MaxWaterFlowGPM is the full-open (valve=1.0) quench spray water capacity in GPM.
	MaxWaterFlowGPM float64
	// NominalWaterFlowGPM is the steady-state spray flow at NominalInletTemp that maintains BaselineQuenchTemp.
	NominalWaterFlowGPM float64
	// BlowerMaxSuctionPa is the positive suction magnitude produced at BlowerVFD = 1.0.
	BlowerMaxSuctionPa float64
	// BlowerTimeConst is the first-order VFD + blower inertia time constant.
	BlowerTimeConst time.Duration
	// DraftTimeConst is the first-order kiln/duct pressure lag time constant.
	DraftTimeConst time.Duration
	// ValveTimeConst is the first-order quench spray valve actuator time constant.
	ValveTimeConst time.Duration
	// QuenchTimeConst is the first-order thermal time constant of the quench outlet gas.
	QuenchTimeConst time.Duration
	// HeatSensitivityK is the open-loop quench outlet temperature rise per unit normalized inlet heat load increase.
	HeatSensitivityK float64
	// CoolingGainKPerGPM is the quench outlet cooling sensitivity in K per GPM of spray water.
	CoolingGainKPerGPM float64
	// DryGasOutletTemp is the equilibrium quench outlet temperature when spray water flow is lost.
	DryGasOutletTemp model.Kelvin
	// InitialSumpPH is the starting scrubber sump pH (typically ~8.0 with NaHCO3).
	InitialSumpPH float64
	// DraftNoiseAmplitudePa adds deterministic high-frequency sensor flutter to DraftPa when > 0.
	DraftNoiseAmplitudePa float64
	// DraftNoiseFreqHz is the sinusoidal noise frequency in Hz used when DraftNoiseAmplitudePa > 0.
	DraftNoiseFreqHz float64
	// AutoStepInterval, when > 0, causes ApplyCommands to advance the plant physics automatically
	// by this fixed step (useful for real-time Daemon operation). When 0, ApplyCommands uses
	// elapsed wall clock since the last update (clamped) unless Step(dt) is driven explicitly.
	AutoStepInterval time.Duration
}

// DefaultSimHALConfig returns physically calibrated dynamics matching model.DefaultControlConfig().
func DefaultSimHALConfig() SimHALConfig {
	return SimHALConfig{
		NominalInletTemp:    model.FahrenheitToKelvin(1800.0), // ~1255.37 K
		BaselineQuenchTemp:  model.FahrenheitToKelvin(165.0),  // ~347.04 K
		MaxWaterFlowGPM:     2.5,
		NominalWaterFlowGPM: 1.0,
		BlowerMaxSuctionPa:  30.0,
		BlowerTimeConst:     400 * time.Millisecond,
		DraftTimeConst:      300 * time.Millisecond,
		ValveTimeConst:      300 * time.Millisecond,
		QuenchTimeConst:     2 * time.Second,
		HeatSensitivityK:    75.0,
		CoolingGainKPerGPM:  22.0,
		DryGasOutletTemp:    model.FahrenheitToKelvin(380.0), // ~466.48 K (> 210 F)
		InitialSumpPH:       8.0,
	}
}

// SimHAL is a thread-safe simulated HardwareHAL modeling kiln firing transients,
// induced-draft blower aerodynamics, quench evaporative cooling, and fault injection hooks.
type SimHAL struct {
	mu  sync.RWMutex
	cfg SimHALConfig

	// Dynamic plant states.
	blowerSpeed      float64
	sprayValvePos    float64
	draftPa          model.Pascal
	quenchInletTemp  model.Kelvin
	quenchOutletTemp model.Kelvin
	waterFlowGPM     float64
	sumpPH           float64
	lastCmd          model.ActuatorCommands

	// Disturbances & timed spikes.
	heatSpikeDeltaK    model.Kelvin
	draftDisturbancePa model.Pascal
	spikeRemaining     time.Duration
	spikePermanent     bool
	simTime            time.Duration
	lastWallUpdate     time.Time
	manualStepping     bool
	clockFn            func() time.Time

	// Fault injection hooks.
	waterLossActive bool
	sensorErr       error
	forceNaN        bool
	staleAge        time.Duration
}

var _ model.HardwareHAL = (*SimHAL)(nil)

// NewSimHAL creates a thread-safe simulated scrubber plant initialized at nominal steady state.
func NewSimHAL(cfg SimHALConfig) *SimHAL {
	if cfg.NominalInletTemp <= 300 {
		cfg = DefaultSimHALConfig()
	}
	if cfg.MaxWaterFlowGPM <= 0 {
		cfg.MaxWaterFlowGPM = 2.5
	}
	if cfg.NominalWaterFlowGPM <= 0 {
		cfg.NominalWaterFlowGPM = 1.0
	}
	if cfg.BlowerMaxSuctionPa <= 0 {
		cfg.BlowerMaxSuctionPa = 30.0
	}
	if cfg.BlowerTimeConst <= 0 {
		cfg.BlowerTimeConst = 400 * time.Millisecond
	}
	if cfg.DraftTimeConst <= 0 {
		cfg.DraftTimeConst = 300 * time.Millisecond
	}
	if cfg.ValveTimeConst <= 0 {
		cfg.ValveTimeConst = 300 * time.Millisecond
	}
	if cfg.QuenchTimeConst <= 0 {
		cfg.QuenchTimeConst = 2 * time.Second
	}
	if cfg.DryGasOutletTemp <= cfg.BaselineQuenchTemp {
		cfg.DryGasOutletTemp = model.FahrenheitToKelvin(380.0)
	}
	if cfg.InitialSumpPH == 0 {
		cfg.InitialSumpPH = 8.0
	}

	initBlower := math.Abs(float64(model.InWCToPascal(-0.05))) / cfg.BlowerMaxSuctionPa
	if initBlower > 1.0 {
		initBlower = 1.0
	}
	initValve := cfg.NominalWaterFlowGPM / cfg.MaxWaterFlowGPM
	if initValve > 1.0 {
		initValve = 1.0
	}
	initDraft := model.Pascal(-cfg.BlowerMaxSuctionPa * initBlower)

	return &SimHAL{
		cfg:              cfg,
		blowerSpeed:      initBlower,
		sprayValvePos:    initValve,
		draftPa:          initDraft,
		quenchInletTemp:  cfg.NominalInletTemp,
		quenchOutletTemp: cfg.BaselineQuenchTemp,
		waterFlowGPM:     cfg.NominalWaterFlowGPM,
		sumpPH:           cfg.InitialSumpPH,
		lastCmd: model.ActuatorCommands{
			BlowerVFD:        initBlower,
			QuenchSprayValve: initValve,
		},
		clockFn: time.Now,
	}
}

// SetManualStepping toggles explicit Step(dt) mode (disabling automatic wall-clock stepping in ApplyCommands).
func (s *SimHAL) SetManualStepping(manual bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manualStepping = manual
}

// SetClock overrides the wall clock used for SensorReadings.Timestamp (useful in deterministic tests).
func (s *SimHAL) SetClock(fn func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fn == nil {
		s.clockFn = time.Now
		return
	}
	s.clockFn = fn
}

// SetDraftNoise configures deterministic sinusoidal sensor noise on DraftPa.
func (s *SimHAL) SetDraftNoise(amplitudePa, freqHz float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.DraftNoiseAmplitudePa = amplitudePa
	s.cfg.DraftNoiseFreqHz = freqHz
}

// InjectHeatSpike simulates a wood-stoking thermal and positive-pressure puff disturbance.
// If duration <= 0, the disturbance persists until ClearHeatSpike() is called.
func (s *SimHAL) InjectHeatSpike(deltaInletTempK model.Kelvin, draftPulsePa model.Pascal, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heatSpikeDeltaK = deltaInletTempK
	s.draftDisturbancePa = draftPulsePa
	if duration <= 0 {
		s.spikePermanent = true
		s.spikeRemaining = 0
	} else {
		s.spikePermanent = false
		s.spikeRemaining = duration
	}
}

// ClearHeatSpike removes any active heat/draft stoking disturbance.
func (s *SimHAL) ClearHeatSpike() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heatSpikeDeltaK = 0
	s.draftDisturbancePa = 0
	s.spikePermanent = false
	s.spikeRemaining = 0
}

// InjectWaterLoss simulates quench supply water loss (true) or restoration (false).
func (s *SimHAL) InjectWaterLoss(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waterLossActive = active
	if active {
		s.waterFlowGPM = 0
	} else {
		s.waterFlowGPM = s.sprayValvePos * s.cfg.MaxWaterFlowGPM
	}
}

// InjectSensorError causes ReadSensors to return err (pass nil to clear).
func (s *SimHAL) InjectSensorError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sensorErr = err
}

// InjectNaNReading causes ReadSensors to emit NaN for QuenchOutletTemp when active.
func (s *SimHAL) InjectNaNReading(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forceNaN = active
}

// InjectStaleTimestamp offsets SensorReadings.Timestamp backwards by age (pass 0 to clear).
func (s *SimHAL) InjectStaleTimestamp(age time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staleAge = age
}

// SetSumpPH sets the simulated scrubber sump pH reading.
func (s *SimHAL) SetSumpPH(ph float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sumpPH = ph
}

// SetQuenchOutletTemp directly sets the simulated quench outlet temperature (for fault-injection tests).
func (s *SimHAL) SetQuenchOutletTemp(t model.Kelvin) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quenchOutletTemp = t
}

// SetDraftPa directly sets the simulated kiln draft pressure in Pascals (for fault-injection tests).
func (s *SimHAL) SetDraftPa(p model.Pascal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.draftPa = p
}

// LastCommands returns the most recent ActuatorCommands applied to SimHAL.
func (s *SimHAL) LastCommands() model.ActuatorCommands {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastCmd
}

// ReadSensors implements model.HardwareHAL.
func (s *SimHAL) ReadSensors(ctx context.Context) (model.SensorReadings, error) {
	if err := ctx.Err(); err != nil {
		return model.SensorReadings{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sensorErr != nil {
		return model.SensorReadings{}, s.sensorErr
	}

	draft := s.draftPa
	if s.cfg.DraftNoiseAmplitudePa > 0 && s.cfg.DraftNoiseFreqHz > 0 {
		noise := s.cfg.DraftNoiseAmplitudePa * math.Sin(2.0*math.Pi*s.cfg.DraftNoiseFreqHz*s.simTime.Seconds())
		draft += model.Pascal(noise)
	}

	outTemp := s.quenchOutletTemp
	if s.forceNaN {
		outTemp = model.Kelvin(math.NaN())
	}

	ts := s.clockFn()
	if s.staleAge > 0 {
		ts = ts.Add(-s.staleAge)
	}

	waterFlowM3s := model.GPMToCubicMetersPerSec(s.waterFlowGPM)

	return model.SensorReadings{
		DraftPa:             draft,
		DraftInWC:           draft.InWC(),
		QuenchInletTemp:     s.quenchInletTemp,
		QuenchOutletTemp:    outTemp,
		WaterFlow:           waterFlowM3s,
		WaterFlowGPM:        s.waterFlowGPM,
		SumpPH:              s.sumpPH,
		BlowerSpeedFraction: s.blowerSpeed,
		Timestamp:           ts,
	}, nil
}

// ApplyCommands implements model.HardwareHAL.
func (s *SimHAL) ApplyCommands(ctx context.Context, cmd model.ActuatorCommands) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastCmd = cmd

	if s.manualStepping {
		return nil
	}

	var dt time.Duration
	if s.cfg.AutoStepInterval > 0 {
		dt = s.cfg.AutoStepInterval
	} else {
		now := s.clockFn()
		if !s.lastWallUpdate.IsZero() {
			dt = now.Sub(s.lastWallUpdate)
			if dt < 0 {
				dt = 0
			}
			if dt > time.Second {
				dt = time.Second
			}
		} else {
			dt = 100 * time.Millisecond
		}
		s.lastWallUpdate = now
	}

	if dt > 0 {
		s.stepLocked(dt)
	}
	return nil
}

// Step advances the simulated kiln and scrubber dynamics by dt.
func (s *SimHAL) Step(dt time.Duration) {
	if dt <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stepLocked(dt)
}

func (s *SimHAL) stepLocked(dt time.Duration) {
	dtSec := dt.Seconds()
	s.simTime += dt

	// Expire timed wood-stoking disturbances.
	activeHeatSpike := s.heatSpikeDeltaK
	activeDraftPulse := s.draftDisturbancePa
	if !s.spikePermanent && s.spikeRemaining > 0 {
		if dt >= s.spikeRemaining {
			s.spikeRemaining = 0
			s.heatSpikeDeltaK = 0
			s.draftDisturbancePa = 0
		} else {
			s.spikeRemaining -= dt
		}
	}

	// Clamp commanded actuator targets to [0, 1].
	targetBlower := clampFloat(s.lastCmd.BlowerVFD, 0, 1)
	targetValve := clampFloat(s.lastCmd.QuenchSprayValve, 0, 1)

	// 1. Actuator first-order lags.
	alphaBlower := expAlpha(dtSec, s.cfg.BlowerTimeConst.Seconds())
	s.blowerSpeed += alphaBlower * (targetBlower - s.blowerSpeed)

	alphaValve := expAlpha(dtSec, s.cfg.ValveTimeConst.Seconds())
	s.sprayValvePos += alphaValve * (targetValve - s.sprayValvePos)

	// 2. Water flow responds to spray valve unless water supply is lost.
	if s.waterLossActive {
		s.waterFlowGPM = 0
	} else {
		s.waterFlowGPM = s.sprayValvePos * s.cfg.MaxWaterFlowGPM
	}

	// 3. Kiln draft dynamics: higher blower speed produces more negative draft (suction).
	// When the fail-safe bypass damper is open, blower suction is isolated from the kiln.
	effectiveSuction := s.cfg.BlowerMaxSuctionPa * s.blowerSpeed
	if s.lastCmd.BypassDamperOpen {
		effectiveSuction = 0
	}
	targetDraftPa := float64(activeDraftPulse) - effectiveSuction
	alphaDraft := expAlpha(dtSec, s.cfg.DraftTimeConst.Seconds())
	s.draftPa = model.Pascal(float64(s.draftPa) + alphaDraft*(targetDraftPa-float64(s.draftPa)))

	// 4. Quench thermal dynamics.
	s.quenchInletTemp = s.cfg.NominalInletTemp + activeHeatSpike
	var targetOutTemp float64
	if s.waterLossActive || s.waterFlowGPM < 0.05 {
		targetOutTemp = float64(s.cfg.DryGasOutletTemp)
	} else {
		denom := float64(s.cfg.NominalInletTemp) - 300.0
		if denom <= 0 {
			denom = 950.0
		}
		heatLoad := (float64(s.quenchInletTemp) - 300.0) / denom
		targetOutTemp = float64(s.cfg.BaselineQuenchTemp) +
			s.cfg.HeatSensitivityK*(heatLoad-1.0) -
			s.cfg.CoolingGainKPerGPM*(s.waterFlowGPM-s.cfg.NominalWaterFlowGPM)
		// Clamp physical equilibrium between spray water minimum (~315 K) and dry gas temperature.
		targetOutTemp = clampFloat(targetOutTemp, 315.0, float64(s.cfg.DryGasOutletTemp))
	}
	alphaQuench := expAlpha(dtSec, s.cfg.QuenchTimeConst.Seconds())
	s.quenchOutletTemp = model.Kelvin(float64(s.quenchOutletTemp) + alphaQuench*(targetOutTemp-float64(s.quenchOutletTemp)))
}

func expAlpha(dtSec, tauSec float64) float64 {
	if tauSec <= 0 {
		return 1.0
	}
	return 1.0 - math.Exp(-dtSec/tauSec)
}

func clampFloat(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}
