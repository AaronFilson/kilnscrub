package studio

import (
	"math"
	"strings"
	"testing"
)

func TestPhase13ValidationMutualExclusivity(t *testing.T) {
	// 1. Steam PPCP / Syphon / Sump-Pool require >= 20 psig steam (reject 0 psig & 6 psig).
	for _, pumpMode := range []string{LocalPumpSteamPPCP, LocalPumpSteamJetSyphon, LocalPumpSumpPoolNoPump} {
		p0 := OffGridCeramicProject(LocalRegionNairobi)
		p0.LocalPumpMode = pumpMode
		if err := p0.Validate(); err == nil {
			t.Errorf("expected natural_chimney + %s to fail Validate()", pumpMode)
		}

		p6 := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorSubsonic)
		p6.LocalPumpMode = pumpMode
		if err := p6.Validate(); err == nil {
			t.Errorf("expected 6 psig subsonic + %s to fail Validate()", pumpMode)
		}

		p60 := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)
		p60.LocalPumpMode = pumpMode
		if err := p60.Validate(); err != nil {
			t.Errorf("expected 60 psig HP + %s to pass Validate(), got %v", pumpMode, err)
		}
	}

	// 2. Non-default boiler heat source requires steam ejector mode.
	for _, src := range []string{LocalBoilerKilnExhaustCoil, LocalBoilerStage1WaterJacket} {
		p0 := OffGridCeramicProject(LocalRegionNairobi)
		p0.LocalBoilerSource = src
		if err := p0.Validate(); err == nil {
			t.Errorf("expected natural_chimney + %s to fail Validate()", src)
		}
		p20 := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorLP)
		p20.LocalBoilerSource = src
		if err := p20.Validate(); err != nil {
			t.Errorf("expected 20 psig LP + %s to pass Validate(), got %v", src, err)
		}
	}

	// 3. Deluxe energy/water-saving upgrades require build_path=custom_fab_deluxe.
	cots := DefaultProject()
	cots.DeluxeCondensingReclaimer = true
	if err := cots.Validate(); err == nil {
		t.Errorf("expected cots_no_weld + deluxe_condensing_reclaimer to fail Validate()")
	}
	cots.BuildPath = BuildPathCustomFabDeluxe
	cots.DeluxeVariableThroatVFD = true
	cots.DeluxeStage1HeatRecovery = true
	if err := cots.Validate(); err != nil {
		t.Errorf("expected custom_fab_deluxe + all 3 eco upgrades to pass Validate(), got %v", err)
	}
}

func TestPhase13BoilerSourcesAndSteamPumps(t *testing.T) {
	ev := NewEvaluator(nil, nil)

	// Baseline 60 psig HP steam ejector with manual treadle pump + dedicated hearth
	baseProj := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)
	baseEval, err := ev.Evaluate(baseProj)
	if err != nil {
		t.Fatalf("Evaluate base 60 psig failed: %v", err)
	}
	if !baseEval.Frugal.HasOverheadTank || baseEval.Frugal.GravityTankLiters <= 0 {
		t.Fatalf("expected baseline to have overhead tank, got %+v", baseEval.Frugal)
	}

	// A. Stage 1 Water Jacket + Mechanical Steam PPCP Recirc Pump
	ppcpProj := baseProj
	ppcpProj.LocalBoilerSource = LocalBoilerStage1WaterJacket
	ppcpProj.LocalPumpMode = LocalPumpSteamPPCP
	ppcpEval, err := ev.Evaluate(ppcpProj)
	if err != nil {
		t.Fatalf("Evaluate Stage 1 jacket + PPCP failed: %v", err)
	}
	if ppcpEval.Frugal.BoilerWaterSavedPct < 35.0 {
		t.Errorf("expected Stage 1 water jacket to save >=35%% evap water, got %.1f%%", ppcpEval.Frugal.BoilerWaterSavedPct)
	}
	if ppcpEval.Frugal.TreadleMinutesPerHr != 0 || ppcpEval.Frugal.Jerrycans20LPerHr != 0 {
		t.Errorf("expected PPCP to eliminate manual pumping, got treadle=%.1f jerrycans=%d", ppcpEval.Frugal.TreadleMinutesPerHr, ppcpEval.Frugal.Jerrycans20LPerHr)
	}
	if ppcpEval.Frugal.PumpSteamKgHr <= 0 || ppcpEval.Frugal.PumpSteamKgHr > 2.0 {
		t.Errorf("unexpected PPCP steam draw: %.2f kg/hr", ppcpEval.Frugal.PumpSteamKgHr)
	}
	if math.Abs(ppcpEval.Frugal.EstimatedLocalCostUSD-ppcpEval.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("PPCP cost parity mismatch: frugal=%.2f vs cots=%.2f", ppcpEval.Frugal.EstimatedLocalCostUSD, ppcpEval.COTS.EstimatedTotalUSD)
	}

	// B. Kiln Exhaust Coil + Steam-Jet Cold Makeup Syphon
	syphProj := baseProj
	syphProj.LocalBoilerSource = LocalBoilerKilnExhaustCoil
	syphProj.LocalPumpMode = LocalPumpSteamJetSyphon
	syphEval, err := ev.Evaluate(syphProj)
	if err != nil {
		t.Fatalf("Evaluate coil + syphon failed: %v", err)
	}
	if !strings.Contains(syphEval.Frugal.BoilerCorrosionNote, "316L") || !strings.Contains(syphEval.Frugal.BoilerCorrosionNote, "SiC") {
		t.Errorf("expected salt-flue hot corrosion warning mentioning 316L and SiC, got %q", syphEval.Frugal.BoilerCorrosionNote)
	}
	foundColdWarn := false
	for _, w := range syphEval.Frugal.Warnings {
		if strings.Contains(w, "<120 F") {
			foundColdWarn = true
			break
		}
	}
	if !foundColdWarn {
		t.Errorf("expected <120 F cold-water warning for steam_jet_syphon")
	}
	if math.Abs(syphEval.Frugal.EstimatedLocalCostUSD-syphEval.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("Syphon cost parity mismatch: frugal=%.2f vs cots=%.2f", syphEval.Frugal.EstimatedLocalCostUSD, syphEval.COTS.EstimatedTotalUSD)
	}

	// C. Zero-Pump, Zero-Water-Tower Sump-Pool Self-Induced Scrubber (60 psig)
	poolProj := baseProj
	poolProj.LocalPumpMode = LocalPumpSumpPoolNoPump
	poolEval, err := ev.Evaluate(poolProj)
	if err != nil {
		t.Fatalf("Evaluate sump_pool_no_pump failed: %v", err)
	}
	if poolEval.Frugal.HasOverheadTank {
		t.Errorf("expected HasOverheadTank=false in sump_pool_no_pump mode")
	}
	if poolEval.Frugal.GravityTankLiters != 0 || poolEval.Frugal.TankElevationFt != 0 {
		t.Errorf("expected 0 overhead tank & elevation in sump_pool_no_pump, got %.0f L @ %.1f ft", poolEval.Frugal.GravityTankLiters, poolEval.Frugal.TankElevationFt)
	}
	if !poolEval.Frugal.DraftFeasible {
		t.Errorf("expected 60 psig HP steam-jet to easily drive 1.5\" submerged sump-pool curtain (avail=%.3f vs req=%.3f)", poolEval.Frugal.AvailableDraftInWC, poolEval.Frugal.RequiredDraftInWC)
	}
	if poolEval.Frugal.EstimatedLocalCostUSD >= baseEval.Frugal.EstimatedLocalCostUSD {
		t.Errorf("expected sump_pool_no_pump (%.2f) to cost less than overhead tank + pump baseline (%.2f)", poolEval.Frugal.EstimatedLocalCostUSD, baseEval.Frugal.EstimatedLocalCostUSD)
	}
	if math.Abs(poolEval.Frugal.EstimatedLocalCostUSD-poolEval.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("Sump-pool cost parity mismatch: frugal=%.2f vs cots=%.2f", poolEval.Frugal.EstimatedLocalCostUSD, poolEval.COTS.EstimatedTotalUSD)
	}
}

func TestPhase13DeluxeEcoUpgrades(t *testing.T) {
	ev := NewEvaluator(nil, nil)

	baseDeluxe := DefaultProject()
	baseDeluxe.BuildPath = BuildPathCustomFabDeluxe
	baseEval, err := ev.Evaluate(baseDeluxe)
	if err != nil {
		t.Fatalf("Evaluate base deluxe failed: %v", err)
	}

	ecoDeluxe := baseDeluxe
	ecoDeluxe.DeluxeCondensingReclaimer = true
	ecoDeluxe.DeluxeVariableThroatVFD = true
	ecoDeluxe.DeluxeStage1HeatRecovery = true
	ecoEval, err := ev.Evaluate(ecoDeluxe)
	if err != nil {
		t.Fatalf("Evaluate eco deluxe failed: %v", err)
	}

	sp := ecoEval.ShopPlan.DeluxeSpec
	if sp == nil || !sp.Enabled {
		t.Fatalf("expected enabled DeluxeSpec")
	}
	if sp.EcoCostDeltaUSD != 4950.0 {
		t.Errorf("EcoCostDeltaUSD = %.2f, want 4950.00", sp.EcoCostDeltaUSD)
	}
	actualDelta := round2(ecoEval.COTS.EstimatedTotalUSD - baseEval.COTS.EstimatedTotalUSD)
	if math.Abs(actualDelta-4950.0) > 0.02 {
		t.Errorf("COTS delta = %.2f, want 4950.00", actualDelta)
	}
	if sp.WaterSavedPct < 75.0 || sp.WaterSavedGalPerFiring <= 0 {
		t.Errorf("expected combined Stage 1 jacket + condensing reclaimer water savings >=75%%, got %.1f%% (%.1f gal)", sp.WaterSavedPct, sp.WaterSavedGalPerFiring)
	}
	if sp.EnergySavedPct < 50.0 || sp.EnergySavedKWhPerFiring <= 0 {
		t.Errorf("expected variable-throat VFD + Stage 1 pre-cooler energy savings >=50%%, got %.1f%% (%.1f kWh)", sp.EnergySavedPct, sp.EnergySavedKWhPerFiring)
	}
	if sp.RecoveredHeatKW < 40.0 {
		t.Errorf("expected combined recovered studio waste heat >=40 kW, got %.1f kW", sp.RecoveredHeatKW)
	}
}
