package control

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func nominalReadings(now time.Time) model.SensorReadings {
	draft := model.InWCToPascal(-0.05)
	water := model.GPMToCubicMetersPerSec(1.0)
	return model.SensorReadings{
		DraftPa:             draft,
		DraftInWC:           draft.InWC(),
		QuenchInletTemp:     model.FahrenheitToKelvin(1800.0),
		QuenchOutletTemp:    model.FahrenheitToKelvin(188.0),
		WaterFlow:           water,
		WaterFlowGPM:        water.GPM(),
		SumpPH:              8.0,
		BlowerSpeedFraction: 0.42,
		Timestamp:           now,
	}
}

// --- 1. LPF DraftFilter Tests ---

func TestDraftFilterGuards(t *testing.T) {
	if _, err := NewDraftFilter(0, 100*time.Millisecond); err == nil {
		t.Fatalf("expected error for cutoffHz == 0 (divide-by-zero guard)")
	}
	if _, err := NewDraftFilter(-1.0, 100*time.Millisecond); err == nil {
		t.Fatalf("expected error for negative cutoffHz")
	}
	if _, err := NewDraftFilter(math.NaN(), 100*time.Millisecond); err == nil {
		t.Fatalf("expected error for NaN cutoffHz")
	}
	if _, err := NewDraftFilter(math.Inf(1), 100*time.Millisecond); err == nil {
		t.Fatalf("expected error for Inf cutoffHz")
	}
	if _, err := NewDraftFilter(0.5, 0); err == nil {
		t.Fatalf("expected error for dt == 0")
	}
}

func TestDraftFilterFirstSampleSeedingAndReset(t *testing.T) {
	f, err := NewDraftFilter(0.5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewDraftFilter: %v", err)
	}
	if f.Seeded() {
		t.Fatalf("expected filter not to be seeded before first sample")
	}

	target := model.InWCToPascal(-0.05) // -12.4544 Pa
	first := f.Update(target)
	if math.Abs(float64(first-target)) > 1e-12 {
		t.Fatalf("first sample must seed directly without jump from 0: got %.6f Pa, want %.6f Pa", first, target)
	}
	if !f.Seeded() {
		t.Fatalf("expected filter to be seeded after first sample")
	}

	// Steady constant input must stay locked at target with zero drift.
	for i := 0; i < 20; i++ {
		out := f.Update(target)
		if math.Abs(float64(out-target)) > 1e-10 {
			t.Fatalf("tick %d: steady output drifted: got %.6f, want %.6f", i, out, target)
		}
	}

	// Non-finite sample must not poison filter state.
	ignored := f.Update(model.Pascal(math.NaN()))
	if math.IsNaN(float64(ignored)) || math.Abs(float64(ignored-target)) > 1e-10 {
		t.Fatalf("NaN sample poisoned filter: got %v", ignored)
	}

	// Reset must allow clean re-seeding at a new operating point.
	f.Reset()
	if f.Seeded() {
		t.Fatalf("expected Seeded() == false after Reset()")
	}
	newPoint := model.InWCToPascal(-0.10) // -24.9089 Pa
	reseeded := f.Update(newPoint)
	if math.Abs(float64(reseeded-newPoint)) > 1e-12 {
		t.Fatalf("first sample after Reset must seed cleanly: got %.6f, want %.6f", reseeded, newPoint)
	}
}

func TestDraftFilterNoiseAttenuation(t *testing.T) {
	const (
		cutoffHz = 0.5
		dt       = 10 * time.Millisecond // 100 Hz sampling
		noiseHz  = 10.0                  // 20x above cutoff -> strong attenuation
		ampPa    = 5.0
	)
	f, err := NewDraftFilter(cutoffHz, dt)
	if err != nil {
		t.Fatalf("NewDraftFilter: %v", err)
	}
	meanPa := float64(model.InWCToPascal(-0.05))

	// Warm up for 2 seconds, then measure RMS ripple over 2 seconds.
	var rawSqSum, filtSqSum float64
	const warmupTicks = 200
	const measureTicks = 200
	for k := 0; k < warmupTicks+measureTicks; k++ {
		tSec := float64(k) * dt.Seconds()
		raw := meanPa + ampPa*math.Sin(2.0*math.Pi*noiseHz*tSec)
		filt := float64(f.Update(model.Pascal(raw)))
		if k >= warmupTicks {
			rawErr := raw - meanPa
			filtErr := filt - meanPa
			rawSqSum += rawErr * rawErr
			filtSqSum += filtErr * filtErr
		}
	}

	rawRMS := math.Sqrt(rawSqSum / measureTicks)
	filtRMS := math.Sqrt(filtSqSum / measureTicks)
	ratio := filtRMS / rawRMS
	// Theoretical first-order LPF attenuation at f/fc = 20 is ~1/sqrt(1+400) ~= 0.05.
	if ratio > 0.15 {
		t.Fatalf("insufficient high-frequency attenuation: rawRMS=%.4f Pa, filtRMS=%.4f Pa, ratio=%.4f (want <= 0.15)",
			rawRMS, filtRMS, ratio)
	}
}

// --- 2. Interlock State Machine Tests ---

func TestInterlockTripsAndResetClearance(t *testing.T) {
	limits := model.DefaultSafetyLimits()
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mutate     func(r *model.SensorReadings)
		readErr    error
		wantReason string
	}{
		{
			name: "high quench temp exact 210F boundary",
			mutate: func(r *model.SensorReadings) {
				r.QuenchOutletTemp = limits.MaxQuenchOutletTemp // exactly 210 F / 372.04 K
			},
			wantReason: model.TripHighQuenchTemp,
		},
		{
			name: "high quench temp above 210F",
			mutate: func(r *model.SensorReadings) {
				r.QuenchOutletTemp = model.FahrenheitToKelvin(215.0)
			},
			wantReason: model.TripHighQuenchTemp,
		},
		{
			name: "water loss below 0.25 GPM",
			mutate: func(r *model.SensorReadings) {
				r.WaterFlow = model.GPMToCubicMetersPerSec(0.24)
				r.WaterFlowGPM = 0.24
			},
			wantReason: model.TripWaterLoss,
		},
		{
			name: "over-draft negative magnitude",
			mutate: func(r *model.SensorReadings) {
				r.DraftPa = -(limits.MaxDraftMagnitudePa + 1.0)
				r.DraftInWC = r.DraftPa.InWC()
			},
			wantReason: model.TripOverDraft,
		},
		{
			name: "over-draft positive pressure magnitude",
			mutate: func(r *model.SensorReadings) {
				r.DraftPa = limits.MaxDraftMagnitudePa + 1.0
				r.DraftInWC = r.DraftPa.InWC()
			},
			wantReason: model.TripOverDraft,
		},
		{
			name:       "HAL read error",
			mutate:     func(r *model.SensorReadings) {},
			readErr:    errors.New("i2c bus timeout"),
			wantReason: model.TripSensorFault,
		},
		{
			name: "NaN sensor reading",
			mutate: func(r *model.SensorReadings) {
				r.QuenchOutletTemp = model.Kelvin(math.NaN())
			},
			wantReason: model.TripSensorFault,
		},
		{
			name: "Inf sensor reading",
			mutate: func(r *model.SensorReadings) {
				r.DraftPa = model.Pascal(math.Inf(-1))
			},
			wantReason: model.TripSensorFault,
		},
		{
			name: "stale sensor timestamp",
			mutate: func(r *model.SensorReadings) {
				r.Timestamp = now.Add(-(limits.SensorFaultTimeout + 100*time.Millisecond))
			},
			wantReason: model.TripSensorFault,
		},
		{
			name: "zero sensor timestamp",
			mutate: func(r *model.SensorReadings) {
				r.Timestamp = time.Time{}
			},
			wantReason: model.TripSensorFault,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			il, err := NewInterlock(limits)
			if err != nil {
				t.Fatalf("NewInterlock: %v", err)
			}

			r := nominalReadings(now)
			tc.mutate(&r)
			st := il.Evaluate(r, tc.readErr, now)

			if !st.BypassDamperOpen || !st.AlarmLatched {
				t.Fatalf("expected bypass open & alarm latched, got %+v", st)
			}
			if st.TripReason != tc.wantReason {
				t.Fatalf("got TripReason=%q, want %q", st.TripReason, tc.wantReason)
			}

			// Reset MUST fail while the fault is still active.
			if err := il.Reset(); err == nil {
				t.Fatalf("expected Reset() to fail while fault %q is active", tc.wantReason)
			}
			if !il.State().BypassDamperOpen {
				t.Fatalf("interlock must remain latched after rejected Reset()")
			}

			// Healthy reading clears active fault, but interlock MUST remain latched until explicit Reset().
			good := nominalReadings(now.Add(100 * time.Millisecond))
			stAfterClear := il.Evaluate(good, nil, now.Add(100*time.Millisecond))
			if !stAfterClear.BypassDamperOpen || !stAfterClear.AlarmLatched || stAfterClear.TripReason != tc.wantReason {
				t.Fatalf("interlock must stay latched after fault clearance until Reset(): %+v", stAfterClear)
			}
			if stAfterClear.FaultActive {
				t.Fatalf("expected FaultActive == false after healthy reading")
			}

			// Now explicit Reset() MUST succeed.
			if err := il.Reset(); err != nil {
				t.Fatalf("Reset() after fault clearance failed: %v", err)
			}
			stAfterReset := il.State()
			if stAfterReset.BypassDamperOpen || stAfterReset.AlarmLatched || stAfterReset.TripReason != model.TripNone {
				t.Fatalf("expected clean unlatched state after Reset(), got %+v", stAfterReset)
			}
		})
	}
}

func TestInterlockSumpPHWarningDoesNotTripBypass(t *testing.T) {
	cfg := model.DefaultControlConfig()
	ctrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	now := time.Now()

	for _, badPH := range []float64{5.2, 10.4} {
		r := nominalReadings(now)
		r.SumpPH = badPH
		cmd := ctrl.Step(r, nil, cfg.LoopInterval, now)
		if cmd.BypassDamperOpen || cmd.AlarmLatched || cmd.TripReason != model.TripNone {
			t.Fatalf("pH=%.1f MUST NOT trip bypass damper or latch alarm, got %+v", badPH, cmd)
		}
		warn, detail := ctrl.PHWarning()
		if !warn || detail == "" {
			t.Fatalf("expected pH warning active for pH=%.1f", badPH)
		}
		status := ctrl.Status()
		if status.State != StateWarning || status.InterlockActive {
			t.Fatalf("expected State=%q & InterlockActive=false for pH=%.1f, got %+v", StateWarning, badPH, status)
		}
	}

	// Returning pH to [MinSumpPH, MaxSumpPH] clears the warning automatically.
	good := nominalReadings(now.Add(cfg.LoopInterval))
	good.SumpPH = 8.0
	ctrl.Step(good, nil, cfg.LoopInterval, now.Add(cfg.LoopInterval))
	if warn, _ := ctrl.PHWarning(); warn {
		t.Fatalf("expected pH warning to clear when pH returns to 8.0")
	}
	if ctrl.Status().State != StateNormal {
		t.Fatalf("expected State=%q after pH returns to normal, got %q", StateNormal, ctrl.Status().State)
	}
}

// --- 3. Direct-Acting PID Sign & Anti-Windup / Output Clamping Tests ---

func TestDirectActingPIDSignNoRunaway(t *testing.T) {
	cfg := model.DefaultControlConfig()
	now := time.Now()

	// Baseline controller at exact setpoints.
	baseCtrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	baseCmd := baseCtrl.Step(nominalReadings(now), nil, cfg.LoopInterval, now)

	// Draft direct-acting check:
	// Higher (less negative) draft pressure (e.g. -2 Pa vs -12.45 Pa) -> MUST INCREASE BlowerVFD.
	highDraftCtrl, _ := NewController(cfg)
	rHighDraft := nominalReadings(now)
	rHighDraft.DraftPa = -2.0
	rHighDraft.DraftInWC = rHighDraft.DraftPa.InWC()
	cmdHighDraft := highDraftCtrl.Step(rHighDraft, nil, cfg.LoopInterval, now)
	if cmdHighDraft.BlowerVFD <= baseCmd.BlowerVFD {
		t.Fatalf("draft PID sign inverted! less-negative draft (-2 Pa vs %.2f Pa) produced BlowerVFD=%.4f <= baseline %.4f",
			cfg.TargetDraftPa, cmdHighDraft.BlowerVFD, baseCmd.BlowerVFD)
	}

	// Lower (more negative) draft pressure (e.g. -25 Pa vs -12.45 Pa) -> MUST DECREASE BlowerVFD.
	lowDraftCtrl, _ := NewController(cfg)
	rLowDraft := nominalReadings(now)
	rLowDraft.DraftPa = -25.0
	rLowDraft.DraftInWC = rLowDraft.DraftPa.InWC()
	cmdLowDraft := lowDraftCtrl.Step(rLowDraft, nil, cfg.LoopInterval, now)
	if cmdLowDraft.BlowerVFD >= baseCmd.BlowerVFD {
		t.Fatalf("draft PID sign inverted! over-suction (-25 Pa vs %.2f Pa) produced BlowerVFD=%.4f >= baseline %.4f",
			cfg.TargetDraftPa, cmdLowDraft.BlowerVFD, baseCmd.BlowerVFD)
	}

	// Quench cooling direct-acting check:
	// Higher quench outlet temp (e.g. 196 F vs 188 F, below 200 F trip) -> MUST INCREASE QuenchSprayValve.
	hotQuenchCtrl, _ := NewController(cfg)
	rHot := nominalReadings(now)
	rHot.QuenchOutletTemp = model.FahrenheitToKelvin(196.0)
	cmdHot := hotQuenchCtrl.Step(rHot, nil, cfg.LoopInterval, now)
	if cmdHot.QuenchSprayValve <= baseCmd.QuenchSprayValve {
		t.Fatalf("quench PID sign inverted! hotter outlet (196 F vs 188 F) produced QuenchSprayValve=%.4f <= baseline %.4f",
			cmdHot.QuenchSprayValve, baseCmd.QuenchSprayValve)
	}

	// Cooler quench outlet temp (e.g. 170 F vs 188 F) -> MUST DECREASE QuenchSprayValve.
	coolQuenchCtrl, _ := NewController(cfg)
	rCool := nominalReadings(now)
	rCool.QuenchOutletTemp = model.FahrenheitToKelvin(170.0)
	cmdCool := coolQuenchCtrl.Step(rCool, nil, cfg.LoopInterval, now)
	if cmdCool.QuenchSprayValve >= baseCmd.QuenchSprayValve {
		t.Fatalf("quench PID sign inverted! cooler outlet (170 F vs 188 F) produced QuenchSprayValve=%.4f >= baseline %.4f",
			cmdCool.QuenchSprayValve, baseCmd.QuenchSprayValve)
	}

	// Verify outputs are strictly finite (no NaN from DerivativeFilterTimeConst) and clamped to [MinOutput, MaxOutput].
	for _, cmd := range []model.ActuatorCommands{baseCmd, cmdHighDraft, cmdLowDraft, cmdHot, cmdCool} {
		if math.IsNaN(cmd.BlowerVFD) || cmd.BlowerVFD < cfg.DraftPID.MinOutput || cmd.BlowerVFD > cfg.DraftPID.MaxOutput {
			t.Fatalf("BlowerVFD out of bounds [%v, %v]: %v", cfg.DraftPID.MinOutput, cfg.DraftPID.MaxOutput, cmd.BlowerVFD)
		}
		if math.IsNaN(cmd.QuenchSprayValve) || cmd.QuenchSprayValve < cfg.QuenchPID.MinOutput || cmd.QuenchSprayValve > cfg.QuenchPID.MaxOutput {
			t.Fatalf("QuenchSprayValve out of bounds [%v, %v]: %v", cfg.QuenchPID.MinOutput, cfg.QuenchPID.MaxOutput, cmd.QuenchSprayValve)
		}
	}
}

// --- 4. Closed-Loop Wood-Stoking Heat Spike Rejection Test ---

func TestClosedLoopWoodStokingHeatSpikeRejection(t *testing.T) {
	cfg := model.DefaultControlConfig()
	ctrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}

	simCfg := DefaultSimHALConfig()
	simCfg.DraftNoiseAmplitudePa = 0.8 // Add realistic draft flutter
	simCfg.DraftNoiseFreqHz = 8.0
	hal := NewSimHAL(simCfg)
	hal.SetManualStepping(true)

	simNow := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	hal.SetClock(func() time.Time { return simNow })

	ctx := context.Background()
	dt := cfg.LoopInterval // 100 ms

	runSteps := func(n int) (maxTempK model.Kelvin, maxDraftErrPa float64, lastReadings model.SensorReadings, lastCmd model.ActuatorCommands) {
		for i := 0; i < n; i++ {
			simNow = simNow.Add(dt)
			readings, err := hal.ReadSensors(ctx)
			if err != nil {
				t.Fatalf("ReadSensors: %v", err)
			}
			cmd := ctrl.Step(readings, nil, dt, simNow)
			if err := hal.ApplyCommands(ctx, cmd); err != nil {
				t.Fatalf("ApplyCommands: %v", err)
			}
			hal.Step(dt)

			if readings.QuenchOutletTemp > maxTempK {
				maxTempK = readings.QuenchOutletTemp
			}
			dErr := math.Abs(float64(ctrl.Status().FilteredDraftPa - cfg.TargetDraftPa))
			if dErr > maxDraftErrPa {
				maxDraftErrPa = dErr
			}
			lastReadings = readings
			lastCmd = cmd
		}
		return
	}

	// Phase 1: 5 seconds nominal warm-up to steady state.
	_, _, baseReadings, baseCmd := runSteps(50)
	if ctrl.Status().InterlockActive {
		t.Fatalf("unexpected interlock trip during nominal warm-up: %+v", ctrl.Status())
	}
	// Verify PID modulates cleanly around 188 F without saturating the spray valve at 100%.
	if baseCmd.QuenchSprayValve >= 0.85 || baseCmd.QuenchSprayValve <= cfg.QuenchPID.MinOutput {
		t.Fatalf("expected nominal spray valve to modulate cleanly off saturation bounds, got %.3f (temp=%.2f F)",
			baseCmd.QuenchSprayValve, baseReadings.QuenchOutletTemp.Fahrenheit())
	}

	// Phase 2: Inject severe wood-stoking heat spike (+333 K inlet surge -> 2400 F)
	// plus positive kiln pressure puff (+8 Pa) sustained for 20 seconds.
	hal.InjectHeatSpike(model.Kelvin(333.33), model.Pascal(8.0), 0)
	peakTempK, _, spikeReadings, spikeCmd := runSteps(200)

	if ctrl.Status().InterlockActive || spikeCmd.BypassDamperOpen {
		t.Fatalf("closed-loop PID failed to prevent bypass trip during wood-stoking spike: peak=%.2f F, status=%+v",
			peakTempK.Fahrenheit(), ctrl.Status())
	}
	// Peak quench outlet temp must stay safely below 210 F (372.04 K) CPVC limit.
	if peakTempK >= cfg.Safety.MaxQuenchOutletTemp {
		t.Fatalf("quench outlet temp exceeded 210 F limit during stoking spike: peak=%.2f F", peakTempK.Fahrenheit())
	}
	// Verify both actuators responded in the direct-acting direction (increased).
	if spikeCmd.QuenchSprayValve <= baseCmd.QuenchSprayValve+0.10 {
		t.Fatalf("expected spray valve to increase substantially during heat spike: base=%.3f, spike=%.3f",
			baseCmd.QuenchSprayValve, spikeCmd.QuenchSprayValve)
	}
	if spikeCmd.BlowerVFD <= baseCmd.BlowerVFD+0.10 {
		t.Fatalf("expected blower VFD to increase substantially during stoking puff: base=%.3f, spike=%.3f",
			baseCmd.BlowerVFD, spikeCmd.BlowerVFD)
	}
	// Verify closed-loop regulation settled close to setpoints under sustained stoking disturbance.
	tempErrK := math.Abs(float64(spikeReadings.QuenchOutletTemp - cfg.QuenchSetpoint))
	if tempErrK > 2.0 {
		t.Fatalf("quench temp did not settle near 188 F setpoint: got %.2f F (err=%.2f K)",
			spikeReadings.QuenchOutletTemp.Fahrenheit(), tempErrK)
	}
	if spikeCmd.QuenchSprayValve >= 0.99 {
		t.Fatalf("expected spray valve to remain unsaturated under 2400 F stoking load at 188 F setpoint, got %.3f", spikeCmd.QuenchSprayValve)
	}
	draftErrPa := math.Abs(float64(ctrl.Status().FilteredDraftPa - cfg.TargetDraftPa))
	if draftErrPa > 1.5 {
		t.Fatalf("draft did not settle near -0.05 inWC setpoint: got %.2f Pa (target %.2f Pa, err=%.2f Pa)",
			ctrl.Status().FilteredDraftPa, cfg.TargetDraftPa, draftErrPa)
	}

	// Phase 3: Clear stoking spike and verify recovery back to nominal without runaway.
	hal.ClearHeatSpike()
	_, _, postReadings, _ := runSteps(150)
	if math.Abs(float64(postReadings.QuenchOutletTemp-cfg.QuenchSetpoint)) > 2.0 {
		t.Fatalf("quench temp failed to recover after spike cleared: got %.2f F", postReadings.QuenchOutletTemp.Fahrenheit())
	}
	if math.Abs(float64(ctrl.Status().FilteredDraftPa-cfg.TargetDraftPa)) > 1.5 {
		t.Fatalf("draft failed to recover after spike cleared: got %.2f Pa", ctrl.Status().FilteredDraftPa)
	}
}

func TestClosedLoopWaterLossFailSafeAndRecovery(t *testing.T) {
	cfg := model.DefaultControlConfig()
	ctrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	hal := NewSimHAL(DefaultSimHALConfig())
	hal.SetManualStepping(true)
	simNow := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	hal.SetClock(func() time.Time { return simNow })
	ctx := context.Background()
	dt := cfg.LoopInterval

	// Run one nominal tick.
	r, _ := hal.ReadSensors(ctx)
	cmd := ctrl.Step(r, nil, dt, simNow)
	_ = hal.ApplyCommands(ctx, cmd)

	// Inject water loss -> must immediately latch TripWaterLoss and fail-safe actuators.
	hal.InjectWaterLoss(true)
	simNow = simNow.Add(dt)
	r, _ = hal.ReadSensors(ctx)
	cmd = ctrl.Step(r, nil, dt, simNow)
	_ = hal.ApplyCommands(ctx, cmd)

	if !cmd.BypassDamperOpen || !cmd.AlarmLatched || cmd.TripReason != model.TripWaterLoss {
		t.Fatalf("expected TripWaterLoss fail-safe latch, got %+v", cmd)
	}
	if cmd.BlowerVFD != 0.0 || cmd.QuenchSprayValve != cfg.QuenchPID.MaxOutput {
		t.Fatalf("unexpected fail-safe actuator outputs: BlowerVFD=%.3f (want 0), Valve=%.3f (want %.3f)",
			cmd.BlowerVFD, cmd.QuenchSprayValve, cfg.QuenchPID.MaxOutput)
	}

	// Reset must fail while water is still lost.
	if err := ctrl.ResetInterlock(); err == nil {
		t.Fatalf("expected ResetInterlock to fail while water loss persists")
	}

	// Restore water and evaluate a healthy tick, then reset.
	hal.InjectWaterLoss(false)
	hal.SetQuenchOutletTemp(cfg.QuenchSetpoint)
	simNow = simNow.Add(dt)
	r, _ = hal.ReadSensors(ctx)
	cmd = ctrl.Step(r, nil, dt, simNow)
	if !cmd.BypassDamperOpen {
		t.Fatalf("controller must remain tripped after water restoration until explicit ResetInterlock()")
	}

	if err := ctrl.ResetInterlock(); err != nil {
		t.Fatalf("ResetInterlock failed after water restoration: %v", err)
	}
	simNow = simNow.Add(dt)
	r, _ = hal.ReadSensors(ctx)
	cmd = ctrl.Step(r, nil, dt, simNow)
	if cmd.BypassDamperOpen || cmd.AlarmLatched || ctrl.Status().InterlockActive {
		t.Fatalf("expected normal operation after ResetInterlock, got cmd=%+v status=%+v", cmd, ctrl.Status())
	}

	// Verify a severe unmitigated thermal spike still trips >= 210 F (TripHighQuenchTemp).
	hal.InjectHeatSpike(model.Kelvin(900.0), model.Pascal(0), 0)
	trippedHighTemp := false
	for i := 0; i < 100; i++ {
		simNow = simNow.Add(dt)
		r, _ = hal.ReadSensors(ctx)
		cmd = ctrl.Step(r, nil, dt, simNow)
		_ = hal.ApplyCommands(ctx, cmd)
		hal.Step(dt)
		if cmd.BypassDamperOpen && cmd.TripReason == model.TripHighQuenchTemp {
			trippedHighTemp = true
			break
		}
	}
	if !trippedHighTemp {
		t.Fatalf("expected severe thermal spike to trip %s (>= 210 F), last status=%+v", model.TripHighQuenchTemp, ctrl.Status())
	}
}

// --- 5. Daemon Concurrency & Clean Shutdown Test (-race) ---

func TestDaemonLoopAndRaceSafety(t *testing.T) {
	cfg := model.DefaultControlConfig()
	cfg.LoopInterval = 10 * time.Millisecond

	simCfg := DefaultSimHALConfig()
	simCfg.AutoStepInterval = 100 * time.Millisecond
	hal := NewSimHAL(simCfg)

	var (
		cbMu       sync.Mutex
		cbCount    int
		lastStatus model.ControlStatus
		errCount   int
	)

	d, err := NewDaemon(cfg, hal,
		WithTelemetryCallback(func(st model.ControlStatus) {
			cbMu.Lock()
			cbCount++
			lastStatus = st
			cbMu.Unlock()
		}),
		WithErrorCallback(func(err error) {
			cbMu.Lock()
			errCount++
			cbMu.Unlock()
		}),
	)
	if err != nil {
		t.Fatalf("NewDaemon: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- d.Run(ctx)
	}()

	// Concurrent readers & disturbance injectors to exercise -race thoroughly.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 40; i++ {
			_ = d.Status()
			_ = d.Ticks()
			_ = d.Running()
			time.Sleep(2 * time.Millisecond)
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		hal.InjectHeatSpike(model.Kelvin(200), model.Pascal(5), 30*time.Millisecond)
		time.Sleep(25 * time.Millisecond)
		hal.SetSumpPH(6.0) // warning only
		time.Sleep(15 * time.Millisecond)
		hal.SetSumpPH(8.0)
	}()
	wg.Wait()

	// Deterministically verify sensor error latching and fault clearance while Daemon runs.
	hal.InjectSensorError(errors.New("transient ADC glitch"))
	waitFor(t, time.Second, func() bool {
		st := d.Status()
		return st.InterlockActive && st.Commands.TripReason == model.TripSensorFault
	}, "daemon did not latch TripSensorFault")

	hal.InjectSensorError(nil)
	waitFor(t, time.Second, func() bool {
		return !d.Controller().InterlockState().FaultActive
	}, "daemon did not observe sensor fault clearance")

	cancel()
	select {
	case runErr := <-errCh:
		if runErr != nil {
			t.Fatalf("Daemon.Run returned unexpected error on clean cancel: %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Daemon.Run did not shut down cleanly after context cancel")
	}

	if d.Running() {
		t.Fatalf("expected Daemon.Running() == false after shutdown")
	}
	if d.Ticks() < 5 {
		t.Fatalf("expected >= 5 control ticks, got %d", d.Ticks())
	}

	cbMu.Lock()
	gotCB := cbCount
	gotErrs := errCount
	gotLast := lastStatus
	cbMu.Unlock()

	if gotCB < 5 {
		t.Fatalf("expected telemetry callback to fire >= 5 times, got %d", gotCB)
	}
	if gotErrs == 0 {
		t.Fatalf("expected error callback to record injected HAL sensor error")
	}
	if !gotLast.InterlockActive || gotLast.Commands.TripReason != model.TripSensorFault {
		t.Fatalf("expected Daemon to latch TripSensorFault on injected HAL error, got %+v", gotLast)
	}

	// Verify fault cleared on subsequent healthy tick and daemon interlock can be reset cleanly.
	if err := d.ResetInterlock(); err != nil {
		t.Fatalf("expected ResetInterlock to succeed after transient sensor error cleared: %v", err)
	}
	if d.Status().InterlockActive {
		t.Fatalf("expected InterlockActive == false after ResetInterlock")
	}
}

func TestSimHALFaultHooksAndConfigValidation(t *testing.T) {
	// Invalid config with zero DerivativeFilterTimeConst must be rejected.
	badCfg := model.DefaultControlConfig()
	badCfg.DraftPID.DerivativeFilterTimeConst = 0
	if _, err := NewController(badCfg); err == nil {
		t.Fatalf("expected NewController to reject zero DerivativeFilterTimeConst")
	}

	// Exercise SimHAL NaN and stale timestamp injection via Controller.Tick.
	cfg := model.DefaultControlConfig()
	ctrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	hal := NewSimHAL(DefaultSimHALConfig())
	ctx := context.Background()

	hal.InjectNaNReading(true)
	st, err := ctrl.Tick(ctx, hal, cfg.LoopInterval)
	if err != nil {
		t.Fatalf("unexpected HAL error on NaN reading: %v", err)
	}
	if !st.InterlockActive || st.Commands.TripReason != model.TripSensorFault {
		t.Fatalf("expected TripSensorFault on NaN reading, got %+v", st)
	}

	hal.InjectNaNReading(false)
	_, _ = ctrl.Tick(ctx, hal, cfg.LoopInterval)
	if err := ctrl.ResetInterlock(); err != nil {
		t.Fatalf("ResetInterlock after NaN cleared: %v", err)
	}

	hal.InjectStaleTimestamp(cfg.Safety.SensorFaultTimeout + time.Second)
	st, err = ctrl.Tick(ctx, hal, cfg.LoopInterval)
	if err != nil {
		t.Fatalf("unexpected HAL error on stale reading: %v", err)
	}
	if !st.InterlockActive || st.Commands.TripReason != model.TripSensorFault {
		t.Fatalf("expected TripSensorFault on stale reading, got %+v", st)
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal(msg)
}

type failWriteHAL struct {
	readings       model.SensorReadings
	failNextWrites int
	applied        []model.ActuatorCommands
}

func (h *failWriteHAL) ReadSensors(_ context.Context) (model.SensorReadings, error) {
	return h.readings, nil
}

func (h *failWriteHAL) ApplyCommands(_ context.Context, cmd model.ActuatorCommands) error {
	h.applied = append(h.applied, cmd)
	if h.failNextWrites > 0 {
		h.failNextWrites--
		return errors.New("simulated modbus write timeout")
	}
	return nil
}

func TestSprint1FailSafeHardening(t *testing.T) {
	cfg := model.DefaultControlConfig()
	if got := cfg.Safety.MaxQuenchOutletTemp.Fahrenheit(); math.Abs(got-200.0) > 0.05 {
		t.Fatalf("default MaxQuenchOutletTemp = %.2f F, want aligned 200.0 F", got)
	}

	// 1. ApplyCommands failure must latch TripSensorFault and issue best-effort fail-safe write.
	ctrl, err := NewController(cfg)
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	hal := &failWriteHAL{
		readings: model.SensorReadings{
			Timestamp:        time.Now(),
			DraftPa:          model.InWCToPascal(-0.05),
			DraftInWC:        -0.05,
			QuenchInletTemp:  model.FahrenheitToKelvin(2200),
			QuenchOutletTemp: model.FahrenheitToKelvin(186),
			WaterFlow:        model.GPMToCubicMetersPerSec(1.5),
			WaterFlowGPM:     1.5,
			SumpPH:           8.0,
		},
		failNextWrites: 1,
	}
	st, err := ctrl.Tick(context.Background(), hal, 100*time.Millisecond)
	if err == nil {
		t.Fatalf("expected ApplyCommands error to be returned")
	}
	if !st.InterlockActive || !st.Commands.BypassDamperOpen || st.Commands.TripReason != model.TripSensorFault {
		t.Fatalf("ApplyCommands failure must latch fail-safe bypass: %+v", st)
	}
	if len(hal.applied) < 2 || !hal.applied[len(hal.applied)-1].BypassDamperOpen {
		t.Fatalf("expected best-effort fail-safe ApplyCommands retry, got %+v", hal.applied)
	}

	// 2. Open-thermocouple / sub-freezing reading must trip TripSensorFault.
	ctrl2, _ := NewController(cfg)
	hal2 := &failWriteHAL{
		readings: model.SensorReadings{
			Timestamp:        time.Now(),
			DraftPa:          model.InWCToPascal(-0.05),
			DraftInWC:        -0.05,
			QuenchInletTemp:  model.FahrenheitToKelvin(2200),
			QuenchOutletTemp: 0, // open TC / 0 K
			WaterFlow:        model.GPMToCubicMetersPerSec(1.5),
			WaterFlowGPM:     1.5,
			SumpPH:           8.0,
		},
	}
	st2, err := ctrl2.Tick(context.Background(), hal2, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error on open-TC reading: %v", err)
	}
	if !st2.InterlockActive || st2.Commands.TripReason != model.TripSensorFault {
		t.Fatalf("expected TripSensorFault on open-TC reading, got %+v", st2)
	}

	// 3. Unknown JSON keys in DaemonConfig must be rejected strictly.
	badCfgJSON := `{"hal_mode":"sim","modbus":{"bypass_coil_deenergize_to_tirp":true}}`
	if _, err := ParseDaemonConfig([]byte(badCfgJSON)); err == nil {
		t.Fatalf("expected ParseDaemonConfig to reject typo/unknown JSON key")
	}
}
