package studio

import (
	"math"
	"strings"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
)

func TestPhase16SolidFuelsSO2AndSulfuricDewPoint(t *testing.T) {
	designer := scrubber.NewDesigner()

	fuels := []struct {
		fuel       string
		saltRate   float64
		wantMinSO2 float64
		wantMinHCl float64
	}{
		{model.FuelCoalSalt, 2.0, 700.0, 200.0},
		{model.FuelCoal, 0.0, 700.0, 0.0},
		{model.FuelPeatSalt, 1.5, 250.0, 150.0},
		{model.FuelPeat, 0.0, 250.0, 0.0},
		{model.FuelBiomassAgSalt, 1.5, 70.0, 150.0},
		{model.FuelBiomassAg, 0.0, 70.0, 0.0},
	}

	for _, tc := range fuels {
		in := model.DefaultSaltKilnInput()
		in.FuelType = tc.fuel
		in.SaltRateLbsPerHr = tc.saltRate

		d, err := designer.Design(in)
		if err != nil {
			t.Fatalf("Design(%s) failed: %v", tc.fuel, err)
		}
		if math.Abs(d.InletGas.SumMoleFractions()-1.0) > 1e-9 {
			t.Errorf("%s inlet mole fraction sum = %.12f, want 1.0", tc.fuel, d.InletGas.SumMoleFractions())
		}
		if math.Abs(d.QuenchedGas.SumMoleFractions()-1.0) > 1e-9 {
			t.Errorf("%s quenched mole fraction sum = %.12f, want 1.0", tc.fuel, d.QuenchedGas.SumMoleFractions())
		}
		if math.Abs(d.OutletGas.SumMoleFractions()-1.0) > 1e-9 {
			t.Errorf("%s outlet mole fraction sum = %.12f, want 1.0", tc.fuel, d.OutletGas.SumMoleFractions())
		}
		if d.PackedBed.SO2InletPPM < tc.wantMinSO2 {
			t.Errorf("%s SO2InletPPM = %.1f, want >= %.1f", tc.fuel, d.PackedBed.SO2InletPPM, tc.wantMinSO2)
		}
		if d.PackedBed.SO2RemovalEfficiency < 0.90 {
			t.Errorf("%s SO2RemovalEfficiency = %.3f, want >= 0.90", tc.fuel, d.PackedBed.SO2RemovalEfficiency)
		}
		if d.PackedBed.SulfurAcidDewPointF < 240.0 || d.PackedBed.SulfurAcidDewPointF > 330.0 {
			t.Errorf("%s SulfurAcidDewPointF = %.1f F, want in [240, 330] F", tc.fuel, d.PackedBed.SulfurAcidDewPointF)
		}
		if tc.wantMinHCl > 0 && d.PackedBed.HClInletPPM < tc.wantMinHCl {
			t.Errorf("%s HClInletPPM = %.1f, want >= %.1f", tc.fuel, d.PackedBed.HClInletPPM, tc.wantMinHCl)
		}
		if d.PackedBed.NaHCO3ConsumptionKgPerHr <= 0 {
			t.Errorf("%s NaHCO3ConsumptionKgPerHr must be > 0 to neutralize SO2", tc.fuel)
		}
	}

	// Coal+Salt must demand substantially more NaHCO3 than Wood+Salt at the same salt rate due to SO2.
	woodIn := model.DefaultSaltKilnInput()
	coalIn := model.DefaultSaltKilnInput()
	coalIn.FuelType = model.FuelCoalSalt
	dWood, err := designer.Design(woodIn)
	if err != nil {
		t.Fatalf("wood_salt design: %v", err)
	}
	dCoal, err := designer.Design(coalIn)
	if err != nil {
		t.Fatalf("coal_salt design: %v", err)
	}
	if dCoal.PackedBed.NaHCO3ConsumptionKgPerHr <= dWood.PackedBed.NaHCO3ConsumptionKgPerHr*1.3 {
		t.Errorf("expected coal_salt NaHCO3 (%.2f kg/hr) to exceed wood_salt (%.2f kg/hr) by >30%% due to SO2",
			dCoal.PackedBed.NaHCO3ConsumptionKgPerHr, dWood.PackedBed.NaHCO3ConsumptionKgPerHr)
	}
}

func TestPhase16GlobalUrbanRegionsAndAltitudePhysics(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())

	regions := []struct {
		region      string
		fuel        string
		currency    string
		wantMaxKPa  float64
		wantMinKPa  float64
		alkali      string
		draftMode   string
		afterburner bool
	}{
		{LocalRegionKathmandu, model.FuelCoal, "NPR", 87.0, 85.0, LocalAlkaliSlakedLime, LocalDraftSteamEjectorSubsonic, true},
		{LocalRegionLagosAccra, model.FuelBiomassAgSalt, "NGN", 102.0, 100.0, LocalAlkaliTronaKanwa, LocalDraftSteamEjectorLP, true},
		{LocalRegionTonalaPuebla, model.FuelWood, "MXN", 84.0, 82.0, LocalAlkaliTronaKanwa, LocalDraftNaturalChimney, true},
		{LocalRegionCuscoAndes, model.FuelPeat, "PEN", 68.0, 65.0, LocalAlkaliSlakedLime, LocalDraftSteamEjectorLP, true},
		{LocalRegionKhurjaDhaka, model.FuelCoalSalt, "INR", 102.0, 100.0, LocalAlkaliSlakedLime, LocalDraftSteamEjectorHP, true},
		{LocalRegionCairoFez, model.FuelBiomassAg, "EGP", 102.0, 100.0, LocalAlkaliTronaKanwa, LocalDraftNaturalChimney, true},
		{LocalRegionHighlandPeat, model.FuelPeatSalt, "GBP", 102.0, 100.0, LocalAlkaliWoodAshLye, LocalDraftNaturalChimney, true},
	}

	for _, tc := range regions {
		proj := GlobalUrbanSolidFuelProject(tc.region, tc.fuel, tc.draftMode)
		proj.LocalAlkali = tc.alkali
		proj.LocalAfterburner = tc.afterburner

		res, err := eval.Evaluate(proj)
		if err != nil {
			t.Fatalf("region=%s fuel=%s Evaluate failed: %v", tc.region, tc.fuel, err)
		}
		fr := res.Frugal
		if fr == nil || !fr.Enabled {
			t.Fatalf("region=%s expected enabled FrugalAssessment", tc.region)
		}
		if fr.Region != tc.region {
			t.Errorf("region=%s got %q", tc.region, fr.Region)
		}
		if !strings.Contains(fr.LocalCurrencyEstimate, tc.currency) {
			t.Errorf("region=%s currency estimate %q missing %q", tc.region, fr.LocalCurrencyEstimate, tc.currency)
		}
		if fr.AmbientPressureKPa < tc.wantMinKPa || fr.AmbientPressureKPa > tc.wantMaxKPa {
			t.Errorf("region=%s AmbientPressureKPa=%.1f want in [%.1f, %.1f]", tc.region, fr.AmbientPressureKPa, tc.wantMinKPa, tc.wantMaxKPa)
		}
		if math.Abs(fr.EstimatedLocalCostUSD-res.COTS.EstimatedTotalUSD) > 0.02 {
			t.Errorf("region=%s cost mismatch: Frugal=%.2f vs COTS=%.2f", tc.region, fr.EstimatedLocalCostUSD, res.COTS.EstimatedTotalUSD)
		}
		if fr.AfterburnerEnabled != tc.afterburner {
			t.Errorf("region=%s AfterburnerEnabled=%v want %v", tc.region, fr.AfterburnerEnabled, tc.afterburner)
		}
		if fr.RegionalMachineryNote == "" {
			t.Errorf("region=%s missing RegionalMachineryNote", tc.region)
		}
		if model.IsSolidDirtyFuel(tc.fuel) {
			if fr.SO2InletPPM <= 0 || fr.SO2RemovalPct < 85.0 {
				t.Errorf("region=%s fuel=%s SO2InletPPM=%.1f SO2RemovalPct=%.1f", tc.region, tc.fuel, fr.SO2InletPPM, fr.SO2RemovalPct)
			}
			if fr.AlkaliKgPerFiring <= 0 {
				t.Errorf("region=%s fuel=%s expected >0 alkali kg/firing for SO2 neutralization", tc.region, tc.fuel)
			}
		}
	}

	// Verify extreme altitude physics at Cusco (3,400 m / 66.8 kPa) vs sea level (Khurja 101.3 kPa):
	// Cusco thin air expands local flue gas volume by +52%, requiring a wider stoneware column than sea level!
	cuscoProj := GlobalUrbanSolidFuelProject(LocalRegionCuscoAndes, model.FuelPeat, LocalDraftNaturalChimney)
	seaProj := GlobalUrbanSolidFuelProject(LocalRegionKhurjaDhaka, model.FuelPeat, LocalDraftNaturalChimney)
	cuscoRes, err := eval.Evaluate(cuscoProj)
	if err != nil {
		t.Fatalf("cusco eval: %v", err)
	}
	seaRes, err := eval.Evaluate(seaProj)
	if err != nil {
		t.Fatalf("sea eval: %v", err)
	}
	if cuscoRes.Frugal.BedDiameterIn <= seaRes.Frugal.BedDiameterIn {
		t.Errorf("expected Cusco 3400m column diameter (%.0f in) > sea-level (%.0f in) due to +52%% lower air density",
			cuscoRes.Frugal.BedDiameterIn, seaRes.Frugal.BedDiameterIn)
	}
}

func TestPhase16TronaKanwaAndAfterburnerParity(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())

	// Verify Trona / Kanwa / Tequesquite (Na2CO3·NaHCO3·2H2O, 3 eq/mol) vs Soda Ash & Baking Soda on coal_salt.
	base := GlobalUrbanSolidFuelProject(LocalRegionLagosAccra, model.FuelCoalSalt, LocalDraftSteamEjectorLP)
	base.LocalAlkali = LocalAlkaliTronaKanwa
	resTrona, err := eval.Evaluate(base)
	if err != nil {
		t.Fatalf("trona eval: %v", err)
	}
	if resTrona.Frugal.AlkaliID != LocalAlkaliTronaKanwa || !strings.Contains(resTrona.Frugal.AlkaliFormula, "Na2CO3") {
		t.Errorf("unexpected trona alkali metadata: %+v", resTrona.Frugal)
	}
	if resTrona.Frugal.AlkaliKgPerFiring <= 5.0 {
		t.Errorf("expected substantial trona mass for coal_salt (HCl+SO2), got %.2f kg", resTrona.Frugal.AlkaliKgPerFiring)
	}

	// Verify toggling LocalAfterburner off reduces cost and triggers tar warning on coal_salt.
	noAB := base
	noAB.LocalAfterburner = false
	resNoAB, err := eval.Evaluate(noAB)
	if err != nil {
		t.Fatalf("no afterburner eval: %v", err)
	}
	if resTrona.Frugal.EstimatedLocalCostUSD <= resNoAB.Frugal.EstimatedLocalCostUSD {
		t.Errorf("expected afterburner cost (%.2f) > without afterburner (%.2f)",
			resTrona.Frugal.EstimatedLocalCostUSD, resNoAB.Frugal.EstimatedLocalCostUSD)
	}
	if math.Abs(resNoAB.Frugal.EstimatedLocalCostUSD-resNoAB.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("no-afterburner cost mismatch: Frugal=%.2f vs COTS=%.2f",
			resNoAB.Frugal.EstimatedLocalCostUSD, resNoAB.COTS.EstimatedTotalUSD)
	}
	if !strings.Contains(strings.ToLower(resNoAB.Frugal.TarMitigationNote), "tar") {
		t.Errorf("expected tar warning when coal_salt runs without afterburner, got %q", resNoAB.Frugal.TarMitigationNote)
	}
}
