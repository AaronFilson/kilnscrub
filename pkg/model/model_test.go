package model

import (
	"encoding/json"
	"math"
	"testing"
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

	limits := DefaultSafetyLimits()
	if !almostEqual(limits.MaxQuenchOutletTemp.Fahrenheit(), 210.0, 1e-9) {
		t.Errorf("expected default MaxQuenchOutletTemp 210 F (372.04 K), got %.2f F (%.2f K)",
			limits.MaxQuenchOutletTemp.Fahrenheit(), limits.MaxQuenchOutletTemp)
	}
	if TripHighQuenchTemp == "" || TripWaterLoss == "" || TripSensorFault == "" {
		t.Error("trip reason constants must be non-empty")
	}

	cfg := DefaultControlConfig()
	if cfg.TargetDraftInWC != -0.05 {
		t.Errorf("expected control target draft -0.05 inWC, got %v", cfg.TargetDraftInWC)
	}
	if !almostEqual(cfg.TargetDraftPa.InWC(), -0.05, 1e-9) {
		t.Errorf("control target draft Pa/inWC mismatch: %v", cfg.TargetDraftPa.InWC())
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

	bom := BOM{
		Title: "Test BOM",
		Items: []BOMItem{
			{Category: "Stage 1", Name: "Quench Shell", Material: "316L SS", Quantity: 2, Unit: "sheet", UnitCostUSD: 150, WeightKg: 25},
			{Category: "Stage 3", Name: "Pall Rings", Material: "PP", Quantity: 5, Unit: "cu.ft.", UnitCostUSD: 40, WeightKg: 8},
		},
	}
	bom.RecalculateTotals()
	if !almostEqual(bom.TotalCostUSD, 500.0, 1e-9) || !almostEqual(bom.TotalWeightKg, 33.0, 1e-9) {
		t.Fatalf("BOM totals mismatch: cost=%v weight=%v", bom.TotalCostUSD, bom.TotalWeightKg)
	}
}
