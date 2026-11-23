package model

import (
	"math"
	"testing"
	"time"
)

func TestModelBranchCoverageAndValidation(t *testing.T) {
	// 1. DesignInput.Validate error branches.
	badInputs := []func(*DesignInput){
		func(d *DesignInput) { d.KilnVolumeCuFt = 0 },
		func(d *DesignInput) { d.AmbientTempK = 100 },
		func(d *DesignInput) { d.PeakExitTempF = math.NaN() },
		func(d *DesignInput) { d.PeakExitTempF = 50 }, // <= ambient
		func(d *DesignInput) { d.PeakExitTempF = 5000 },
		func(d *DesignInput) { d.SaltRateLbsPerHr = -1 },
		func(d *DesignInput) { d.ExcessAirFraction = 6.0 },
		func(d *DesignInput) { d.FuelType = "diesel" },
		func(d *DesignInput) { d.VenturiThroatVelocity = 0 },
		func(d *DesignInput) { d.VenturiLGRatio = 0 },
		func(d *DesignInput) { d.VenturiDropletDiameter = -1 },
		func(d *DesignInput) { d.TargetHClRemoval = 1.0 },
		func(d *DesignInput) { d.PackingDiameter = 0 },
		func(d *DesignInput) { d.PackingVoidage = 1.0 },
		func(d *DesignInput) { d.PackingFactor = 0 },
		func(d *DesignInput) { d.TargetKilnDraftInWC = 0.1 },
		func(d *DesignInput) { d.AmbientPressurePa = 1000 },
		func(d *DesignInput) { d.SprayWaterTempK = 250 },
		func(d *DesignInput) { d.SumpPHTarget = 12.0 },
		func(d *DesignInput) { d.DuctLength = -1 },
	}
	for i, mutate := range badInputs {
		in := DefaultSaltKilnInput()
		mutate(&in)
		if err := in.Validate(); err == nil {
			t.Errorf("badInputs[%d]: expected validation error", i)
		}
	}

	// 2. SafetyLimits.Validate error branches.
	badSafety := []func(*SafetyLimits){
		func(s *SafetyLimits) { s.MaxQuenchOutletTemp = 200 },
		func(s *SafetyLimits) { s.MinWaterFlow = -1 },
		func(s *SafetyLimits) { s.MinSumpPH = 10; s.MaxSumpPH = 8 },
		func(s *SafetyLimits) { s.SensorFaultTimeout = 0 },
		func(s *SafetyLimits) { s.MaxDraftMagnitudePa = 0 },
	}
	for i, mutate := range badSafety {
		s := DefaultSafetyLimits()
		mutate(&s)
		if err := s.Validate(); err == nil {
			t.Errorf("badSafety[%d]: expected validation error", i)
		}
	}

	// 3. PIDGains.Validate error branches.
	badPID := []func(*PIDGains){
		func(g *PIDGains) { g.Kp = -1 },
		func(g *PIDGains) { g.Ki = -1 },
		func(g *PIDGains) { g.Kd = -1 },
		func(g *PIDGains) { g.AntiWindUpGain = -1 },
		func(g *PIDGains) { g.DerivativeFilterTimeConst = 0 },
		func(g *PIDGains) { g.MinOutput = 0.8; g.MaxOutput = 0.5 },
	}
	for i, mutate := range badPID {
		g := DefaultControlConfig().DraftPID
		mutate(&g)
		if err := g.Validate("pid"); err == nil {
			t.Errorf("badPID[%d]: expected validation error", i)
		}
	}

	// 4. Cutoff/TimeConst guards, nil ControlConfig receivers, Sync reverse paths, ControlConfig.Validate errors.
	if CutoffHzToTimeConst(0) != 0 || CutoffHzToTimeConst(math.NaN()) != 0 {
		t.Error("CutoffHzToTimeConst invalid should return 0")
	}
	if TimeConstToCutoffHz(0) != 0 {
		t.Error("TimeConstToCutoffHz(0) should return 0")
	}

	var nilCfg *ControlConfig
	nilCfg.SetTargetDraftInWC(-0.05)
	nilCfg.Sync()

	// Reverse sync: TargetDraftPa -> TargetDraftInWC and DraftFilterTimeConst -> DraftFilterCutoffHz.
	rev := DefaultControlConfig()
	rev.TargetDraftInWC = 0
	rev.TargetDraftPa = InWCToPascal(-0.10)
	rev.DraftFilterCutoffHz = 0
	rev.DraftFilterTimeConst = 200 * time.Millisecond
	rev.Sync()
	if math.Abs(rev.TargetDraftInWC-(-0.10)) > 1e-6 {
		t.Errorf("reverse TargetDraftInWC = %v, want -0.10", rev.TargetDraftInWC)
	}
	if rev.DraftFilterCutoffHz <= 0 {
		t.Errorf("reverse DraftFilterCutoffHz = %v, want > 0", rev.DraftFilterCutoffHz)
	}
	if err := rev.Validate(); err != nil {
		t.Errorf("reverse synced ControlConfig should validate: %v", err)
	}

	badCtrl := []func(*ControlConfig){
		func(c *ControlConfig) { c.TargetDraftInWC = 0.5 },
		func(c *ControlConfig) { c.TargetDraftPa = Pascal(math.NaN()) },
		func(c *ControlConfig) { c.TargetDraftPa = -999 }, // mismatch
		func(c *ControlConfig) { c.DraftFilterCutoffHz = 0 },
		func(c *ControlConfig) { c.DraftFilterTimeConst = 0 },
		func(c *ControlConfig) { c.DraftFilterTimeConst = 10 * time.Second }, // mismatch
		func(c *ControlConfig) { c.Safety.SensorFaultTimeout = 0 },
		func(c *ControlConfig) { c.QuenchSetpoint = 500 },
		func(c *ControlConfig) { c.LoopInterval = 0 },
		func(c *ControlConfig) { c.DraftPID.Kp = -1 },
		func(c *ControlConfig) { c.QuenchPID.Kp = -1 },
	}
	for i, mutate := range badCtrl {
		c := DefaultControlConfig()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("badCtrl[%d]: expected validation error", i)
		}
	}

	// 5. BOM nil & lump-sum RecalculateTotals + units helpers.
	var nilBOM *BOM
	nilBOM.RecalculateTotals()
	b := BOM{Items: []BOMItem{{Quantity: 1, UnitCostUSD: 0, TotalCostUSD: 123.45, WeightKg: 5.0}}}
	b.RecalculateTotals()
	if b.TotalCostUSD != 123.45 || b.TotalWeightKg != 5.0 {
		t.Errorf("lump-sum BOM totals mismatch: %+v", b)
	}

	// Exercise remaining unit conversion methods.
	if Meters(1.0).Inches() <= 0 || FeetToMeters(1.0) <= 0 || CubicMeters(1.0).CuFt() <= 0 {
		t.Error("unit conversion check failed")
	}
	if SqMetersToSqFeet := SquareMeters(1.0); SqMetersToSqFeet <= 0 {
		t.Error("SquareMeters check failed")
	}
}
