package scrubber

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

type scrubberOracleFixtures struct {
	ErgunPackedBed []struct {
		DpM     float64 `json:"dp_m"`
		Voidage float64 `json:"voidage"`
		VsMS    float64 `json:"vs_m_s"`
		RhoKgM3 float64 `json:"rho_kg_m3"`
		MuPaS   float64 `json:"mu_Pa_s"`
		LM      float64 `json:"L_m"`
		DPPa    float64 `json:"dP_Pa"`
	} `json:"ergun_packed_bed"`
	CalvertVenturi []struct {
		MuGPaS        float64 `json:"mu_g_Pa_s"`
		DDropletM     float64 `json:"d_droplet_m"`
		VThroatMS     float64 `json:"v_throat_m_s"`
		RhoPKgM3      float64 `json:"rho_p_kg_m3"`
		LG            float64 `json:"L_G"`
		RhoLKgM3      float64 `json:"rho_L_kg_m3"`
		FCalvert      float64 `json:"f_calvert"`
		MeanFreePathM float64 `json:"mean_free_path_m"`
		D50M          float64 `json:"d50_m"`
	} `json:"calvert_venturi"`
}

func loadScrubberOracle(t *testing.T) scrubberOracleFixtures {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "oracle_fixtures.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read oracle fixtures %s: %v", path, err)
	}
	var fx scrubberOracleFixtures
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatalf("unmarshal oracle fixtures: %v", err)
	}
	return fx
}

func relErr(got, want float64) float64 {
	if want == 0 {
		return math.Abs(got - want)
	}
	return math.Abs(got-want) / math.Abs(want)
}

func TestErgunAgainstOracle(t *testing.T) {
	fx := loadScrubberOracle(t)
	const maxRelTol = 1e-6

	for _, c := range fx.ErgunPackedBed {
		t.Run(fmt.Sprintf("vs=%.1fm_s", c.VsMS), func(t *testing.T) {
			dp, err := ErgunPressureDrop(c.DpM, c.Voidage, c.VsMS, c.RhoKgM3, c.MuPaS, c.LM)
			if err != nil {
				t.Fatalf("ErgunPressureDrop: %v", err)
			}
			if r := relErr(float64(dp), c.DPPa); r > maxRelTol {
				t.Errorf("Ergun dP = %.12g Pa, want %.12g Pa (relErr=%.3e > %.1e)", dp, c.DPPa, r, maxRelTol)
			}
		})
	}
}

func TestCalvertD50AgainstOracle(t *testing.T) {
	fx := loadScrubberOracle(t)
	const maxRelTol = 1e-6

	for _, c := range fx.CalvertVenturi {
		t.Run(fmt.Sprintf("v=%.0f_LG=%.4f", c.VThroatMS, c.LG), func(t *testing.T) {
			d50, err := CalvertCutDiameter(
				c.MuGPaS, c.DDropletM, c.VThroatMS, c.RhoPKgM3,
				c.LG, c.RhoLKgM3, c.FCalvert, c.MeanFreePathM,
			)
			if err != nil {
				t.Fatalf("CalvertCutDiameter: %v", err)
			}
			if r := relErr(d50, c.D50M); r > maxRelTol {
				t.Errorf("Calvert d50 = %.12g m, want %.12g m (relErr=%.3e > %.1e)", d50, c.D50M, r, maxRelTol)
			}
			// Verify Pt(d50) == 0.5.
			pt := CalvertPenetration(d50, c.MuGPaS, c.DDropletM, c.VThroatMS, c.RhoPKgM3, c.LG, c.RhoLKgM3, c.FCalvert, c.MeanFreePathM)
			if math.Abs(pt-0.5) > 1e-9 {
				t.Errorf("Pt(d50) = %.12f, want 0.5", pt)
			}
		})
	}
}

func TestCalvertFSmallXStability(t *testing.T) {
	// Ensure small-x Taylor branch is smooth, negative, finite, and matches direct log1p at boundary u = 0.05.
	const f = 0.5
	kpTiny := 1e-10
	fTiny := CalvertF(kpTiny, f)
	if math.IsNaN(fTiny) || math.IsInf(fTiny, 0) || fTiny >= 0 {
		t.Fatalf("CalvertF(1e-10, 0.5) = %v, expected finite negative value", fTiny)
	}

	// Continuity across u = 0.05 (x = 0.035, kp = 0.07).
	kpBelow := 0.069999999
	kpAbove := 0.070000001
	fBelow := CalvertF(kpBelow, f)
	fAbove := CalvertF(kpAbove, f)
	if r := relErr(fBelow, fAbove); r > 1e-6 {
		t.Errorf("CalvertF discontinuity at Taylor boundary: below=%.12g, above=%.12g (rel=%.3e)", fBelow, fAbove, r)
	}
}

func TestVenturiDropletAccelFactor(t *testing.T) {
	// 24 cu.ft. reference case: L_t = 0.08 m, d_d = 50 um, v_t = 70 m/s, rho_g = 0.806, rho_L = 998.2, mu_g = 1.54e-5.
	beta := VenturiDropletAccelFactor(0.08, 50e-6, 70.0, 0.806, 998.2, 1.54e-5)
	if beta < 0.70 || beta > 0.75 {
		t.Errorf("VenturiDropletAccelFactor(ref) = %.4f, want in [0.70, 0.75]", beta)
	}
	// Longer throat must accelerate droplets further (monotonically increasing, bounded <= 1.0).
	betaLong := VenturiDropletAccelFactor(0.15, 50e-6, 70.0, 0.806, 998.2, 1.54e-5)
	if betaLong <= beta || betaLong > 1.0 {
		t.Errorf("expected beta(0.15m)=%.4f in (%.4f, 1.0]", betaLong, beta)
	}
	// Larger droplets accelerate less over the same throat length.
	betaCoarse := VenturiDropletAccelFactor(0.08, 109e-6, 70.0, 0.806, 998.2, 1.54e-5)
	if betaCoarse >= beta || betaCoarse <= 0 {
		t.Errorf("expected beta(109um)=%.4f in (0, %.4f)", betaCoarse, beta)
	}
}

func TestDefault24CuFtWoodSaltKilnDesign(t *testing.T) {
	d := NewDesigner()
	input := model.DefaultSaltKilnInput()
	design, err := d.Design(input)
	if err != nil {
		t.Fatalf("Design(DefaultSaltKilnInput): %v", err)
	}

	// 1. Gas compositions normalized.
	for name, g := range map[string]model.GasComposition{
		"inlet":    design.InletGas,
		"quenched": design.QuenchedGas,
		"outlet":   design.OutletGas,
	} {
		if math.Abs(g.SumMoleFractions()-1.0) > 1e-9 {
			t.Errorf("%s mole fraction sum = %.12f, want 1.0", name, g.SumMoleFractions())
		}
		if g.VolumetricFlow <= 0 || g.Density <= 0 || g.Viscosity <= 0 {
			t.Errorf("%s invalid thermophysical state: Q=%v, rho=%v, mu=%v", name, g.VolumetricFlow, g.Density, g.Viscosity)
		}
	}

	// 2. Stage 1 Quench checks.
	if design.Quench.OutletTemp >= model.DefaultSafetyLimits().MaxQuenchOutletTemp {
		t.Errorf("quench outlet %.2f K >= CPVC limit %.2f K", design.Quench.OutletTemp, model.DefaultSafetyLimits().MaxQuenchOutletTemp)
	}
	if design.Quench.TotalSprayWaterGPM <= design.Quench.EvaporatedWaterGPM {
		t.Errorf("total spray (%.3f GPM) must exceed evaporated (%.3f GPM)", design.Quench.TotalSprayWaterGPM, design.Quench.EvaporatedWaterGPM)
	}

	// 3. Stage 2 Venturi checks: NaCl d50 > WoodAsh d50 (since rho_wood_ash = 2400 > rho_nacl = 2160).
	if design.Venturi.CutDiameterD50 <= design.Venturi.WoodAshCutDiameterD50 {
		t.Errorf("expected NaCl d50 (%.4f um) > wood ash d50 (%.4f um)", design.Venturi.CutDiameterD50Microns, design.Venturi.WoodAshCutDiameterD50Microns)
	}
	if len(design.Venturi.EfficiencyCurve) < 10 {
		t.Errorf("expected >= 10 points on efficiency curve, got %d", len(design.Venturi.EfficiencyCurve))
	}
	var prevAsh, prevNaCl float64
	for _, pt := range design.Venturi.EfficiencyCurve {
		if pt.WoodAshEfficiency < 0 || pt.WoodAshEfficiency > 1 || pt.NaClFumeEfficiency < 0 || pt.NaClFumeEfficiency > 1 {
			t.Errorf("efficiency point out of [0, 1] at %.2f um: ash=%v, nacl=%v", pt.ParticleDiameterMicrons, pt.WoodAshEfficiency, pt.NaClFumeEfficiency)
		}
		if pt.WoodAshEfficiency < prevAsh || pt.NaClFumeEfficiency < prevNaCl {
			t.Errorf("efficiency curve not monotonic at %.2f um", pt.ParticleDiameterMicrons)
		}
		prevAsh = pt.WoodAshEfficiency
		prevNaCl = pt.NaClFumeEfficiency
	}
	if design.Venturi.OverallWoodAshEfficiency <= 0 || design.Venturi.OverallWoodAshEfficiency > 1 {
		t.Errorf("OverallWoodAshEfficiency out of (0, 1]: %v", design.Venturi.OverallWoodAshEfficiency)
	}
	if design.Venturi.OverallNaClFumeEfficiency <= 0 || design.Venturi.OverallNaClFumeEfficiency > 1 {
		t.Errorf("OverallNaClFumeEfficiency out of (0, 1]: %v", design.Venturi.OverallNaClFumeEfficiency)
	}
	// A2: Yung-Calvert (1977) throat droplet acceleration gives ~18.5-19.5 inWC for 50 um droplet / 80 mm throat.
	if design.Venturi.PressureDropInWC < 17.5 || design.Venturi.PressureDropInWC > 20.5 {
		t.Errorf("Venturi PressureDropInWC = %.2f, want realistic Yung-Calvert band [17.5, 20.5] inWC", design.Venturi.PressureDropInWC)
	}

	// 4. Stage 3 Packed-Bed checks: flooding <= 75% (0..100 scale), HCl removal >= 0.98 (0..1 scale).
	if design.PackedBed.PercentOfFlood <= 0 || design.PackedBed.PercentOfFlood > 75.0 {
		t.Errorf("PercentOfFlood = %.2f%%, must be in (0, 75.0]", design.PackedBed.PercentOfFlood)
	}
	// A1: Irrigated Sauter-diameter Ergun bed dP must land in the Eckert GPDC irrigated band (~1.2-2.0 inWC).
	bedDPInWC := design.PackedBed.BedPressureDrop.InWC()
	if bedDPInWC < 1.20 || bedDPInWC > 2.00 {
		t.Errorf("PackedBed BedPressureDrop = %.3f inWC, want GPDC irrigated band [1.20, 2.00] inWC", bedDPInWC)
	}
	if design.PackedBed.HClRemovalEfficiency < 0.98 || design.PackedBed.HClRemovalEfficiency > 1.0 {
		t.Errorf("HClRemovalEfficiency = %.4f, must be in [0.98, 1.0]", design.PackedBed.HClRemovalEfficiency)
	}
	if relErr(float64(design.PackedBed.PackingHeight), design.PackedBed.NTU*float64(design.PackedBed.HTU)) > 1e-9 {
		t.Errorf("PackingHeight (%v) != NTU (%v) * HTU (%v)", design.PackedBed.PackingHeight, design.PackedBed.NTU, design.PackedBed.HTU)
	}
	if design.PackedBed.NaHCO3ConsumptionKgPerHr <= 0 {
		t.Errorf("NaHCO3ConsumptionKgPerHr must be > 0 for salt kiln, got %v", design.PackedBed.NaHCO3ConsumptionKgPerHr)
	}

	// 5. Stage 4 Blower contract checks: NO double-counting demister, positive kiln draft magnitude,
	// signed negative echo in TargetKilnDraftInWC, Efficiency on (0, 1].
	b := design.Blower
	if b.Breakdown.PackedBed != design.PackedBed.BedPressureDrop {
		t.Errorf("Breakdown.PackedBed (%v) != PackedBed.BedPressureDrop (%v) — possible demister double-count", b.Breakdown.PackedBed, design.PackedBed.BedPressureDrop)
	}
	if b.Breakdown.Demister != design.PackedBed.DemisterPressureDrop {
		t.Errorf("Breakdown.Demister (%v) != PackedBed.DemisterPressureDrop (%v)", b.Breakdown.Demister, design.PackedBed.DemisterPressureDrop)
	}
	if b.Breakdown.KilnDraft <= 0 {
		t.Errorf("Breakdown.KilnDraft must be positive draw magnitude, got %v", b.Breakdown.KilnDraft)
	}
	if b.TargetKilnDraftInWC != input.TargetKilnDraftInWC {
		t.Errorf("Blower.TargetKilnDraftInWC = %v, want signed echo %v", b.TargetKilnDraftInWC, input.TargetKilnDraftInWC)
	}
	if math.Abs(float64(b.Breakdown.TotalPa()-b.TotalStaticPressure)) > 1e-9 {
		t.Errorf("Breakdown.TotalPa() (%v) != TotalStaticPressure (%v)", b.Breakdown.TotalPa(), b.TotalStaticPressure)
	}
	if b.Efficiency <= 0 || b.Efficiency > 1.0 {
		t.Errorf("Blower.Efficiency = %v, must be on (0, 1]", b.Efficiency)
	}
	if b.RecommendedMotorHP < b.ShaftPowerHP {
		t.Errorf("RecommendedMotorHP (%v) < ShaftPowerHP (%v)", b.RecommendedMotorHP, b.ShaftPowerHP)
	}
	// Combined A1+A2 realistic 24 cu.ft. total static pressure (~20.5-22.5 inWC) and motor (2.0-3.0 HP).
	if b.TotalStaticPressureInWC < 20.0 || b.TotalStaticPressureInWC > 23.0 {
		t.Errorf("Blower TotalStaticPressureInWC = %.2f, want realistic band [20.0, 23.0] inWC", b.TotalStaticPressureInWC)
	}
	if b.RecommendedMotorHP < 2.0 || b.RecommendedMotorHP > 3.0 {
		t.Errorf("Blower RecommendedMotorHP = %.2f HP, want [2.0, 3.0] HP", b.RecommendedMotorHP)
	}
	if len(b.FanCurve) < 10 || len(b.SystemCurve) != len(b.FanCurve) {
		t.Errorf("unexpected fan/system curve lengths: %d / %d", len(b.FanCurve), len(b.SystemCurve))
	}

	// A4: Verify exact 2400 F reference temperature scaling (1588.7056 - 293.15 K).
	wantRefDeltaT := float64(model.FahrenheitToKelvin(2400.0)) - 293.15
	if math.Abs(refPeakDeltaTK-wantRefDeltaT) > 1e-12 {
		t.Errorf("refPeakDeltaTK = %.12f, want %.12f", refPeakDeltaTK, wantRefDeltaT)
	}

	// 6. Envelope dimensions positive.
	if design.Envelope.Length <= 0 || design.Envelope.Width <= 0 || design.Envelope.Height <= 0 || design.Envelope.FootprintArea <= 0 {
		t.Errorf("invalid envelope dimensions: %+v", design.Envelope)
	}
}

func TestAllFuelTypesAndAutoDroplet(t *testing.T) {
	d := NewDesigner()
	fuels := []struct {
		fuel     string
		saltRate float64
	}{
		{model.FuelWoodSalt, 2.0},
		{model.FuelWood, 0.0},
		{model.FuelGasSalt, 1.5},
		{model.FuelGas, 0.0},
		{"", 2.0}, // Empty defaults to wood_salt
	}

	for _, tc := range fuels {
		t.Run("fuel_"+tc.fuel, func(t *testing.T) {
			in := model.DefaultSaltKilnInput()
			in.FuelType = tc.fuel
			in.SaltRateLbsPerHr = tc.saltRate
			in.VenturiDropletDiameter = 0 // Exercise Nukiyama-Tanasawa auto droplet sizing.

			des, err := d.Design(in)
			if err != nil {
				t.Fatalf("Design(%q): %v", tc.fuel, err)
			}
			if des.Venturi.DropletDiameter <= 0 {
				t.Errorf("auto droplet diameter must be > 0, got %v", des.Venturi.DropletDiameter)
			}
			if des.PackedBed.PercentOfFlood > 75.0 {
				t.Errorf("PercentOfFlood = %.2f > 75.0", des.PackedBed.PercentOfFlood)
			}
			if tc.saltRate == 0 && des.PackedBed.NaHCO3ConsumptionKgPerHr != 0 {
				t.Errorf("expected 0 NaHCO3 consumption for salt_rate=0, got %v", des.PackedBed.NaHCO3ConsumptionKgPerHr)
			}
		})
	}
}

func TestDesignInputValidationErrors(t *testing.T) {
	d := NewDesigner()

	cases := []struct {
		name   string
		mutate func(*model.DesignInput)
	}{
		{"zero_volume", func(in *model.DesignInput) { in.KilnVolumeCuFt = 0 }},
		{"nan_temp", func(in *model.DesignInput) { in.PeakExitTempF = math.NaN() }},
		{"positive_draft", func(in *model.DesignInput) { in.TargetKilnDraftInWC = 0.05 }},
		{"bad_fuel", func(in *model.DesignInput) { in.FuelType = "plutonium" }},
		{"zero_packing_factor", func(in *model.DesignInput) { in.PackingFactor = 0 }},
		{"invalid_hcl_target", func(in *model.DesignInput) { in.TargetHClRemoval = 1.0 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := model.DefaultSaltKilnInput()
			tc.mutate(&in)
			if _, err := d.Design(in); err == nil {
				t.Errorf("expected validation error for %s, got nil", tc.name)
			}
		})
	}
}
