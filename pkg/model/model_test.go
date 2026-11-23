package model

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestUnitConversions(t *testing.T) {
	t.Run("Temperature", func(t *testing.T) {
		cases := []struct {
			f float64
			k Kelvin
		}{
			{32.0, 273.15},
			{212.0, 373.15},
			{210.0, 372.0388888888889},
			{2400.0, 1588.7055555555555},
			{165.0, 347.0388888888889},
			{188.0, 359.81666666666666},
		}
		for _, tc := range cases {
			gotK := FahrenheitToKelvin(tc.f)
			if !almostEqual(float64(gotK), float64(tc.k), 1e-6) {
				t.Errorf("FahrenheitToKelvin(%v) = %v, want %v", tc.f, gotK, tc.k)
			}
			gotF := KelvinToFahrenheit(gotK)
			if !almostEqual(gotF, tc.f, 1e-9) || !almostEqual(gotK.Fahrenheit(), tc.f, 1e-9) {
				t.Errorf("round-trip F->K->F for %v got %v", tc.f, gotF)
			}
		}
		if !almostEqual(float64(CelsiusToKelvin(20.0)), 293.15, 1e-9) {
			t.Fatalf("CelsiusToKelvin(20) failed")
		}
		if !almostEqual(Kelvin(293.15).Celsius(), 20.0, 1e-9) {
			t.Fatalf("Kelvin.Celsius failed")
		}
	})

	t.Run("PressureInWC", func(t *testing.T) {
		p := InWCToPascal(-0.05)
		if !almostEqual(float64(p), -12.4544455, 1e-6) {
			t.Errorf("InWCToPascal(-0.05) = %v", p)
		}
		if !almostEqual(p.InWC(), -0.05, 1e-12) {
			t.Errorf("round-trip inWC failed: %v", p.InWC())
		}
	})

	t.Run("FlowCFMAndGPM", func(t *testing.T) {
		qCFM := CFMToCubicMetersPerSec(1000.0)
		if !almostEqual(float64(qCFM), 0.4719474432, 1e-9) {
			t.Errorf("CFMToCubicMetersPerSec(1000) = %v", qCFM)
		}
		if !almostEqual(qCFM.CFM(), 1000.0, 1e-9) {
			t.Errorf("CFM round-trip failed: %v", qCFM.CFM())
		}

		qGPM := GPMToCubicMetersPerSec(10.0)
		if !almostEqual(float64(qGPM), 6.30901964e-4, 1e-12) {
			t.Errorf("GPMToCubicMetersPerSec(10) = %v", qGPM)
		}
		if !almostEqual(qGPM.GPM(), 10.0, 1e-9) {
			t.Errorf("GPM round-trip failed: %v", qGPM.GPM())
		}
	})

	t.Run("MassLengthVolumePower", func(t *testing.T) {
		m := LbsPerHrToKgPerSec(2.0)
		if !almostEqual(m.LbsPerHr(), 2.0, 1e-12) {
			t.Errorf("lbs/hr round-trip failed: %v", m.LbsPerHr())
		}
		if !almostEqual(float64(InchesToMeters(1.0)), 0.0254, 1e-12) {
			t.Errorf("InchesToMeters(1) failed")
		}
		if !almostEqual(InchesToMeters(12.0).Feet(), 1.0, 1e-12) {
			t.Errorf("12 in != 1 ft")
		}
		if !almostEqual(FeetToMeters(1.0).Inches(), 12.0, 1e-12) {
			t.Errorf("1 ft != 12 in")
		}
		if !almostEqual(MicronsToMeters(50.0).Microns(), 50.0, 1e-12) {
			t.Errorf("Microns round-trip failed")
		}
		if !almostEqual(MillimetersToMeters(25.4).Inches(), 1.0, 1e-12) {
			t.Errorf("25.4 mm != 1 in")
		}
		if !almostEqual(FeetPerSecToMetersPerSec(100.0).FeetPerSec(), 100.0, 1e-12) {
			t.Errorf("ft/s round-trip failed")
		}
		if !almostEqual(CuFtToCubicMeters(24.0).CuFt(), 24.0, 1e-12) {
			t.Errorf("cu.ft. round-trip failed")
		}
		if !almostEqual(SqFtToSquareMeters(10.0).SqFt(), 10.0, 1e-12) {
			t.Errorf("sq.ft. round-trip failed")
		}
		if !almostEqual(HPToWatts(5.0).HP(), 5.0, 1e-12) {
			t.Errorf("HP round-trip failed")
		}
	})
}

func TestDefaultsAndValidation(t *testing.T) {
	in := DefaultSaltKilnInput()
	if err := in.Validate(); err != nil {
		t.Fatalf("DefaultSaltKilnInput() failed validation: %v", err)
	}
	if in.KilnVolumeCuFt != 24.0 || in.PeakExitTempF != 2400.0 || in.SaltRateLbsPerHr != 2.0 {
		t.Fatalf("unexpected reference salt kiln defaults: %+v", in)
	}
	if in.TargetKilnDraftInWC != -0.05 {
		t.Fatalf("expected target draft -0.05 inWC, got %v", in.TargetKilnDraftInWC)
	}

	bad := in
	bad.KilnVolumeCuFt = 0
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for KilnVolumeCuFt=0")
	}

	bad = in
	bad.TargetHClRemoval = 1.0
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for TargetHClRemoval=1.0")
	}

	bad = in
	bad.TargetKilnDraftInWC = 0.05
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for positive TargetKilnDraftInWC=+0.05")
	}

	bad = in
	bad.TargetKilnDraftInWC = -2.5
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for TargetKilnDraftInWC=-2.5")
	}

	okZeroDraft := in
	okZeroDraft.TargetKilnDraftInWC = 0
	if err := okZeroDraft.Validate(); err != nil {
		t.Errorf("expected TargetKilnDraftInWC=0 to be allowed, got %v", err)
	}

	bad = in
	bad.PackingFactor = 0
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for PackingFactor=0")
	}

	bad = in
	bad.PackingFactor = -10
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for PackingFactor=-10")
	}

	for _, ft := range []string{"", FuelWoodSalt, FuelWood, FuelGasSalt, FuelGas} {
		okFuel := in
		okFuel.FuelType = ft
		if err := okFuel.Validate(); err != nil {
			t.Errorf("expected FuelType=%q to pass validation, got %v", ft, err)
		}
	}
	bad = in
	bad.FuelType = "coal"
	if err := bad.Validate(); err == nil {
		t.Error("expected validation error for unknown FuelType")
	}

	limits := DefaultSafetyLimits()
	if err := limits.Validate(); err != nil {
		t.Fatalf("DefaultSafetyLimits().Validate() failed: %v", err)
	}
	if !almostEqual(limits.MaxQuenchOutletTemp.Fahrenheit(), 210.0, 1e-9) {
		t.Errorf("expected default MaxQuenchOutletTemp 210 F (372.04 K), got %.2f F (%.2f K)",
			limits.MaxQuenchOutletTemp.Fahrenheit(), limits.MaxQuenchOutletTemp)
	}
	if TripHighQuenchTemp == "" || TripWaterLoss == "" || TripSensorFault == "" {
		t.Error("trip reason constants must be non-empty")
	}

	cfg := DefaultControlConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultControlConfig().Validate() failed: %v", err)
	}
	if cfg.TargetDraftInWC != -0.05 {
		t.Errorf("expected control target draft -0.05 inWC, got %v", cfg.TargetDraftInWC)
	}
	if !almostEqual(cfg.QuenchSetpoint.Fahrenheit(), 188.0, 1e-9) {
		t.Errorf("expected default QuenchSetpoint 188.0 F (above 183.9 F adiabatic floor), got %.2f F", cfg.QuenchSetpoint.Fahrenheit())
	}
	if !almostEqual(cfg.TargetDraftPa.InWC(), -0.05, 1e-9) {
		t.Errorf("control target draft Pa/inWC mismatch: %v", cfg.TargetDraftPa.InWC())
	}
	expectedTau := 1.0 / (2.0 * math.Pi * cfg.DraftFilterCutoffHz)
	if !almostEqual(cfg.DraftFilterTimeConst.Seconds(), expectedTau, 1e-6) {
		t.Errorf("default DraftFilterTimeConst (%v) != 1/(2*pi*%.2f) (%.6fs)",
			cfg.DraftFilterTimeConst, cfg.DraftFilterCutoffHz, expectedTau)
	}
	if cfg.DraftPID.DerivativeFilterTimeConst <= 0 || cfg.QuenchPID.DerivativeFilterTimeConst <= 0 {
		t.Errorf("default PID DerivativeFilterTimeConst must be > 0")
	}
	if cfg.DraftPID.AntiWindUpGain <= 0 || cfg.QuenchPID.AntiWindUpGain <= 0 {
		t.Errorf("default PID AntiWindUpGain must be > 0")
	}
}

func TestDesignInputRejectsNaNAndInf(t *testing.T) {
	nonFinites := []struct {
		label string
		val   float64
	}{
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}
	mutators := []struct {
		name string
		set  func(d *DesignInput, v float64)
	}{
		{"KilnVolumeCuFt", func(d *DesignInput, v float64) { d.KilnVolumeCuFt = v }},
		{"PeakExitTempF", func(d *DesignInput, v float64) { d.PeakExitTempF = v }},
		{"SaltRateLbsPerHr", func(d *DesignInput, v float64) { d.SaltRateLbsPerHr = v }},
		{"ExcessAirFraction", func(d *DesignInput, v float64) { d.ExcessAirFraction = v }},
		{"VenturiThroatVelocity", func(d *DesignInput, v float64) { d.VenturiThroatVelocity = MetersPerSec(v) }},
		{"VenturiLGRatio", func(d *DesignInput, v float64) { d.VenturiLGRatio = v }},
		{"VenturiDropletDiameter", func(d *DesignInput, v float64) { d.VenturiDropletDiameter = Meters(v) }},
		{"TargetHClRemoval", func(d *DesignInput, v float64) { d.TargetHClRemoval = v }},
		{"PackingDiameter", func(d *DesignInput, v float64) { d.PackingDiameter = Meters(v) }},
		{"PackingVoidage", func(d *DesignInput, v float64) { d.PackingVoidage = v }},
		{"PackingFactor", func(d *DesignInput, v float64) { d.PackingFactor = v }},
		{"TargetKilnDraftInWC", func(d *DesignInput, v float64) { d.TargetKilnDraftInWC = v }},
		{"AmbientTempK", func(d *DesignInput, v float64) { d.AmbientTempK = Kelvin(v) }},
		{"AmbientPressurePa", func(d *DesignInput, v float64) { d.AmbientPressurePa = Pascal(v) }},
		{"SprayWaterTempK", func(d *DesignInput, v float64) { d.SprayWaterTempK = Kelvin(v) }},
		{"SumpPHTarget", func(d *DesignInput, v float64) { d.SumpPHTarget = v }},
		{"DuctLength", func(d *DesignInput, v float64) { d.DuctLength = Meters(v) }},
		{"DuctDiameter", func(d *DesignInput, v float64) { d.DuctDiameter = Meters(v) }},
	}

	for _, m := range mutators {
		for _, nf := range nonFinites {
			d := DefaultSaltKilnInput()
			m.set(&d, nf.val)
			if err := d.Validate(); err == nil {
				t.Errorf("Validate() accepted %s for %s", nf.label, m.name)
			}
		}
	}
}

func TestControlConfigSyncAndValidation(t *testing.T) {
	cfg := DefaultControlConfig()
	cfg.TargetDraftInWC = -0.10 // Mutate only InWC without syncing Pa.
	if err := cfg.Validate(); err == nil {
		t.Error("expected Validate() error on diverged TargetDraftInWC vs TargetDraftPa")
	}
	cfg.Sync()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected Validate() to pass after Sync(), got %v", err)
	}
	if !almostEqual(float64(cfg.TargetDraftPa), float64(InWCToPascal(-0.10)), 1e-9) {
		t.Errorf("Sync() did not update TargetDraftPa: got %v", cfg.TargetDraftPa)
	}

	cfg.SetTargetDraftInWC(-0.08)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("SetTargetDraftInWC failed validation: %v", err)
	}

	// Diverged LPF cutoff vs time constant must fail validation until Sync().
	cfg.DraftFilterCutoffHz = 1.0
	if err := cfg.Validate(); err == nil {
		t.Error("expected Validate() error when DraftFilterCutoffHz and DraftFilterTimeConst disagree")
	}
	cfg.Sync()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected Sync() to reconcile LPF cutoff and tau, got %v", err)
	}

	// Zero DerivativeFilterTimeConst must fail validation (protects einride/pid div-by-zero).
	badPID := DefaultControlConfig()
	badPID.DraftPID.DerivativeFilterTimeConst = 0
	if err := badPID.Validate(); err == nil {
		t.Error("expected Validate() error for zero DraftPID.DerivativeFilterTimeConst")
	}

	// QuenchSetpoint >= MaxQuenchOutletTemp must fail validation.
	badQuench := DefaultControlConfig()
	badQuench.QuenchSetpoint = badQuench.Safety.MaxQuenchOutletTemp
	if err := badQuench.Validate(); err == nil {
		t.Error("expected Validate() error when QuenchSetpoint >= MaxQuenchOutletTemp")
	}

	// Cutoff/tau helper round-trip.
	tau := CutoffHzToTimeConst(0.5)
	fc := TimeConstToCutoffHz(tau)
	if !almostEqual(fc, 0.5, 1e-6) {
		t.Errorf("CutoffHzToTimeConst/TimeConstToCutoffHz round-trip got %v", fc)
	}
	if CutoffHzToTimeConst(0) != 0 || TimeConstToCutoffHz(0) != 0 {
		t.Error("expected zero for non-positive LPF parameters")
	}
	if tau < 318*time.Millisecond || tau > 319*time.Millisecond {
		t.Errorf("unexpected tau for 0.5 Hz: %v", tau)
	}
}

func TestDesignJSONRoundTripAndHelpers(t *testing.T) {
	design := ScrubberDesign{
		Input: DefaultSaltKilnInput(),
		InletGas: GasComposition{
			N2: 0.72, CO2: 0.12, H2O: 0.11, O2: 0.048, HCl: 0.002,
		},
		Blower: BlowerResult{
			Flow:     CFMToCubicMetersPerSec(350),
			FlowACFM: 350,
			Breakdown: PressureDropBreakdown{
				Duct:      50,
				Quench:    250,
				Venturi:   3500,
				PackedBed: 600,
				Demister:  150,
				KilnDraft: 12.45,
			},
		},
	}

	if !almostEqual(design.InletGas.SumMoleFractions(), 1.0, 1e-12) {
		t.Errorf("expected mole fractions sum 1.0, got %v", design.InletGas.SumMoleFractions())
	}
	if !almostEqual(float64(design.Blower.Breakdown.TotalPa()), 4562.45, 1e-9) {
		t.Errorf("unexpected pressure drop total: %v", design.Blower.Breakdown.TotalPa())
	}

	data, err := json.Marshal(design)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var decoded ScrubberDesign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if decoded.Input.KilnVolumeCuFt != design.Input.KilnVolumeCuFt ||
		decoded.Blower.Breakdown.Venturi != design.Blower.Breakdown.Venturi ||
		decoded.Blower.FlowACFM != 350 {
		t.Fatalf("JSON round-trip mismatch: %+v", decoded)
	}

	var nilBOM *BOM
	nilBOM.RecalculateTotals() // Must not panic.

	bom := BOM{
		Title: "Test BOM",
		Items: []BOMItem{
			{Category: "Stage 1", Name: "Quench Shell", Material: "316L SS", Quantity: 2, Unit: "sheet", UnitCostUSD: 150, WeightKg: 25},
			{Category: "Stage 3", Name: "Pall Rings", Material: "PP", Quantity: 5, Unit: "cu.ft.", UnitCostUSD: 40, WeightKg: 8},
			{Category: "Hardware", Name: "Lump-sum Gaskets", Material: "PTFE", Quantity: 1, Unit: "lot", UnitCostUSD: 0, TotalCostUSD: 75, WeightKg: 1.5},
		},
	}
	bom.RecalculateTotals()
	if !almostEqual(bom.TotalCostUSD, 575.0, 1e-9) || !almostEqual(bom.TotalWeightKg, 34.5, 1e-9) {
		t.Fatalf("BOM totals mismatch: cost=%v weight=%v", bom.TotalCostUSD, bom.TotalWeightKg)
	}

	// Editing Quantity/UnitCostUSD must recompute TotalCostUSD (no stale totals), while preserving lump-sum item.
	bom.Items[0].Quantity = 3
	bom.Items[1].UnitCostUSD = 50
	bom.RecalculateTotals()
	if !almostEqual(bom.Items[0].TotalCostUSD, 450.0, 1e-9) || !almostEqual(bom.Items[1].TotalCostUSD, 250.0, 1e-9) {
		t.Fatalf("BOM item TotalCostUSD not recomputed: %+v", bom.Items)
	}
	if !almostEqual(bom.Items[2].TotalCostUSD, 75.0, 1e-9) {
		t.Fatalf("BOM lump-sum item TotalCostUSD overwritten: %v", bom.Items[2].TotalCostUSD)
	}
	if !almostEqual(bom.TotalCostUSD, 775.0, 1e-9) {
		t.Fatalf("BOM recomputed TotalCostUSD mismatch: got %v want 775", bom.TotalCostUSD)
	}
}
