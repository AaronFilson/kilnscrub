package studio

import (
	"math"
	"strings"
	"testing"
)

func TestPhase13ValidationMutualExclusivity(t *testing.T) {
	// 1. Steam PPCP / Syphon require >= 20 psig steam; Sump-Pool requires 60 psig HP strictly!
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

	// Sump-pool no-pump MUST fail at 20 psig LP (insufficient draft for high-capture trough + sieve trays).
	p20Pool := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorLP)
	p20Pool.LocalPumpMode = LocalPumpSumpPoolNoPump
	if err := p20Pool.Validate(); err == nil {
		t.Errorf("expected 20 psig LP + sump_pool_no_pump to fail Validate()")
	}

	// 2. Boiler source safety matrix (Phase 14):
	// - stage1_feedwater_preheater (0 psig vented): allowed on 6, 20, 60 psig; rejected on natural_chimney.
	// - kiln_exhaust_coil (<2 L monotube): allowed ONLY on 20 psig LP (banned at 60 psig & 6 psig!).
	// - stage1_water_jacket (ASME/IBR <=25 L): allowed on 20 & 60 psig; rejected on 6 psig & natural_chimney.
	for _, src := range []string{LocalBoilerStage1FeedwaterPreheater, LocalBoilerKilnExhaustCoil, LocalBoilerStage1WaterJacket} {
		p0 := OffGridCeramicProject(LocalRegionNairobi)
		p0.LocalBoilerSource = src
		if err := p0.Validate(); err == nil {
			t.Errorf("expected natural_chimney + %s to fail Validate()", src)
		}
	}
	p6Pre := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorSubsonic)
	p6Pre.LocalBoilerSource = LocalBoilerStage1FeedwaterPreheater
	if err := p6Pre.Validate(); err != nil {
		t.Errorf("expected 6 psig + stage1_feedwater_preheater to pass Validate(), got %v", err)
	}
	p6Jkt := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorSubsonic)
	p6Jkt.LocalBoilerSource = LocalBoilerStage1WaterJacket
	if err := p6Jkt.Validate(); err == nil {
		t.Errorf("expected 6 psig + stage1_water_jacket to fail Validate()")
	}
	p60Coil := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)
	p60Coil.LocalBoilerSource = LocalBoilerKilnExhaustCoil
	if err := p60Coil.Validate(); err == nil {
		t.Errorf("expected 60 psig HP + kiln_exhaust_coil to fail Validate() (No-Monotube-at-60-psig rule)")
	}
	p20Coil := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorLP)
	p20Coil.LocalBoilerSource = LocalBoilerKilnExhaustCoil
	if err := p20Coil.Validate(); err != nil {
		t.Errorf("expected 20 psig LP + kiln_exhaust_coil to pass Validate(), got %v", err)
	}
	p60Jkt := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorHP)
	p60Jkt.LocalBoilerSource = LocalBoilerStage1WaterJacket
	if err := p60Jkt.Validate(); err != nil {
		t.Errorf("expected 60 psig HP + stage1_water_jacket to pass Validate(), got %v", err)
	}

	// 3. Deluxe energy/water-saving upgrades require build_path=custom_fab_deluxe,
	// and variable-throat VFD strictly requires inline_id_blower + single_throat + VFD blower.
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

	dlxEd := cots
	dlxEd.StageAlignment = StageAlignmentEductorDryFan
	if err := dlxEd.Validate(); err == nil {
		t.Errorf("expected eductor_dry_fan + deluxe_variable_throat_vfd to fail Validate()")
	}
	dlxPar := cots
	dlxPar.VenturiLayout = VenturiLayoutParallelMiniVenturis
	if err := dlxPar.Validate(); err == nil {
		t.Errorf("expected parallel_mini_venturis + deluxe_variable_throat_vfd to fail Validate()")
	}
	dlxRegen := cots
	dlxRegen.BlowerDriveOption = BlowerDriveRegenSideChannel
	if err := dlxRegen.Validate(); err == nil {
		t.Errorf("expected regen_side_channel + deluxe_variable_throat_vfd to fail Validate()")
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

	// A0. 0 psig Vented Stage 1 Feedwater Preheater on 6 psig Subsonic Steamer
	preProj := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorSubsonic)
	preProj.LocalBoilerSource = LocalBoilerStage1FeedwaterPreheater
	preEval, err := ev.Evaluate(preProj)
	if err != nil {
		t.Fatalf("Evaluate 6 psig + stage1_feedwater_preheater failed: %v", err)
	}
	if preEval.Frugal.BoilerWaterSavedPct < 6.0 || preEval.Frugal.BoilerWaterSavedPct > 12.0 {
		t.Errorf("expected 0 psig feedwater preheater to save 6-12%% evap water, got %.1f%%", preEval.Frugal.BoilerWaterSavedPct)
	}
	if !strings.Contains(preEval.Frugal.BoilerCorrosionNote, "0 psig") {
		t.Errorf("expected 0 psig safety note on feedwater preheater, got %q", preEval.Frugal.BoilerCorrosionNote)
	}
	if math.Abs(preEval.Frugal.EstimatedLocalCostUSD-preEval.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("Preheater cost parity mismatch: frugal=%.2f vs cots=%.2f", preEval.Frugal.EstimatedLocalCostUSD, preEval.COTS.EstimatedTotalUSD)
	}

	// A. ASME/IBR <=25 L Stage 1 Water Jacket + Mechanical Steam PPCP Recirc Pump
	ppcpProj := baseProj
	ppcpProj.ZeroElectricity = false // Even if ZeroElectricity=false, steam PPCP draws 0 W!
	ppcpProj.LocalBoilerSource = LocalBoilerStage1WaterJacket
	ppcpProj.LocalPumpMode = LocalPumpSteamPPCP
	ppcpEval, err := ev.Evaluate(ppcpProj)
	if err != nil {
		t.Fatalf("Evaluate Stage 1 jacket + PPCP failed: %v", err)
	}
	if ppcpEval.Frugal.BoilerWaterSavedPct < 18.0 || ppcpEval.Frugal.BoilerWaterSavedPct > 38.0 {
		t.Errorf("expected energy-conserved Stage 1 jacket water savings in [18, 38]%%, got %.1f%%", ppcpEval.Frugal.BoilerWaterSavedPct)
	}
	if !strings.Contains(ppcpEval.Frugal.BoilerCorrosionNote, "UG-28") {
		t.Errorf("expected ASME UG-28 external-pressure stiffener note on Stage 1 jacket, got %q", ppcpEval.Frugal.BoilerCorrosionNote)
	}
	if ppcpEval.Frugal.TreadleMinutesPerHr != 0 || ppcpEval.Frugal.Jerrycans20LPerHr != 0 {
		t.Errorf("expected PPCP to eliminate manual pumping, got treadle=%.1f jerrycans=%d", ppcpEval.Frugal.TreadleMinutesPerHr, ppcpEval.Frugal.Jerrycans20LPerHr)
	}
	if ppcpEval.Frugal.PowerWatts != 0 || ppcpEval.Urban.Electrical.RequiredVoltage != 0 || len(ppcpEval.ShopPlan.Electrical) != 0 {
		t.Errorf("expected 0 W / 0V / 0 electrical circuits for steam PPCP, got W=%.0f V=%d circuits=%d", ppcpEval.Frugal.PowerWatts, ppcpEval.Urban.Electrical.RequiredVoltage, len(ppcpEval.ShopPlan.Electrical))
	}
	if ppcpEval.Frugal.SteamThermalKW <= baseEval.Frugal.SteamThermalKW {
		t.Errorf("expected PPCP steam draw to increase SteamThermalKW (got %.1f vs base %.1f)", ppcpEval.Frugal.SteamThermalKW, baseEval.Frugal.SteamThermalKW)
	}
	if math.Abs(ppcpEval.Frugal.EstimatedLocalCostUSD-ppcpEval.COTS.EstimatedTotalUSD) > 0.02 {
		t.Errorf("PPCP cost parity mismatch: frugal=%.2f vs cots=%.2f", ppcpEval.Frugal.EstimatedLocalCostUSD, ppcpEval.COTS.EstimatedTotalUSD)
	}

	// B. 20 psig LP Kiln Exhaust Coil + Steam-Jet Cold Makeup Syphon
	syphProj := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorLP)
	syphProj.LocalBoilerSource = LocalBoilerKilnExhaustCoil
	syphProj.LocalPumpMode = LocalPumpSteamJetSyphon
	syphEval, err := ev.Evaluate(syphProj)
	if err != nil {
		t.Fatalf("Evaluate 20 psig coil + syphon failed: %v", err)
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

	// C. High-Capture Zero-Pump, Zero-Water-Tower Sump-Pool Scrubber (60 psig HP)
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
	if !poolEval.Frugal.DraftFeasible || poolEval.Frugal.DraftMarginInWC < 1.5 {
		t.Errorf("expected 60 psig HP sump-pool to have >=1.5 inWC draft margin (avail=%.3f vs req=%.3f)", poolEval.Frugal.AvailableDraftInWC, poolEval.Frugal.RequiredDraftInWC)
	}
	if poolEval.Frugal.HClRemovalPct < 96.0 || poolEval.Frugal.SubmicronNaClCapturePct < 76.0 {
		t.Errorf("expected high-capture sump-pool >=96%% HCl & >=76%% NaCl, got HCl=%.1f%% NaCl=%.1f%%", poolEval.Frugal.HClRemovalPct, poolEval.Frugal.SubmicronNaClCapturePct)
	}
	if strings.Contains(poolEval.Frugal.Summary, " 0 L") || strings.Contains(poolEval.Frugal.Summary, "0.0 ft") {
		t.Errorf("sump_pool_no_pump Summary must not contain nonsensical '0 L' / '0.0 ft' text: %q", poolEval.Frugal.Summary)
	}
	if poolEval.Recirculation.SumpVolumeGal < 90.0 {
		t.Errorf("expected dual sump pool working volume >=90 gal, got %.1f", poolEval.Recirculation.SumpVolumeGal)
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
	// Exact cost parity between COTS BOM delta and DeluxeSpec.EcoCostDeltaUSD (duty-scaled!)
	actualDelta := round2(ecoEval.COTS.EstimatedTotalUSD - baseEval.COTS.EstimatedTotalUSD)
	if math.Abs(actualDelta-sp.EcoCostDeltaUSD) > 0.02 {
		t.Errorf("COTS eco delta = %.2f != DeluxeSpec.EcoCostDeltaUSD = %.2f", actualDelta, sp.EcoCostDeltaUSD)
	}
	if sp.WaterSavedPct < 65.0 || sp.WaterSavedGalPerFiring <= 0 {
		t.Errorf("expected combined Stage 1 jacket + ceramic condensing reclaimer water savings >=65%%, got %.1f%% (%.1f gal)", sp.WaterSavedPct, sp.WaterSavedGalPerFiring)
	}
	if sp.EnergySavedPct < 50.0 || sp.EnergySavedKWhPerFiring <= 0 {
		t.Errorf("expected variable-throat VFD + Stage 1 pre-cooler energy savings >=50%%, got %.1f%% (%.1f kWh)", sp.EnergySavedPct, sp.EnergySavedKWhPerFiring)
	}
	if sp.RecoveredHeatKW < 55.0 {
		t.Errorf("expected psychrometric Stage 1 + ceramic condenser recovered heat >=55 kW, got %.1f kW", sp.RecoveredHeatKW)
	}

	// Verify ceramic-coated / SiC lining (NO bare 316L SS in condensing reclaimer!)
	foundCeramicCond := false
	for _, it := range ecoEval.COTS.BOM.Items {
		if strings.Contains(it.Spec, "DLX-COND-RECLAIM") {
			foundCeramicCond = true
			if !strings.Contains(it.Material, "Ceramic") || !strings.Contains(it.Material, "SiC") {
				t.Errorf("expected ceramic-enamel / SiC material on condensing reclaimer, got %q", it.Material)
			}
			if strings.Contains(it.Material, "316L") {
				t.Errorf("condensing reclaimer material must NOT specify bare 316L SS: %q", it.Material)
			}
		}
	}
	if !foundCeramicCond {
		t.Errorf("expected DLX-COND-RECLAIM item in Deluxe Eco BOM")
	}

	// Verify propagation into Urban Plume (150 F exit & shorter winter plume), Urban Water & Recirculation!
	if ecoEval.Urban.Plume.ExhaustTempF != 150.0 {
		t.Errorf("expected condensing reclaimer to lower Urban.Plume.ExhaustTempF to 150.0 F, got %.1f F", ecoEval.Urban.Plume.ExhaustTempF)
	}
	if ecoEval.Urban.Plume.BarePlumeLengthFt >= baseEval.Urban.Plume.BarePlumeLengthFt {
		t.Errorf("expected condensing reclaimer to shrink bare winter plume (eco=%.1f vs base=%.1f)", ecoEval.Urban.Plume.BarePlumeLengthFt, baseEval.Urban.Plume.BarePlumeLengthFt)
	}
	if ecoEval.Urban.Water.TotalFiringGallons >= baseEval.Urban.Water.TotalFiringGallons {
		t.Errorf("expected Deluxe Eco water recovery to reduce Urban.Water.TotalFiringGallons (eco=%.1f vs base=%.1f)", ecoEval.Urban.Water.TotalFiringGallons, baseEval.Urban.Water.TotalFiringGallons)
	}
	if ecoEval.Recirculation.RecircMakeupWaterGal >= baseEval.Recirculation.RecircMakeupWaterGal {
		t.Errorf("expected Deluxe Eco water recovery to reduce Recirculation.RecircMakeupWaterGal (eco=%.1f vs base=%.1f)", ecoEval.Recirculation.RecircMakeupWaterGal, baseEval.Recirculation.RecircMakeupWaterGal)
	}
	if len(ecoEval.ShopPlan.Parts) <= len(baseEval.ShopPlan.Parts) {
		t.Errorf("expected Deluxe Eco upgrades to appear in ShopPlan.Parts")
	}
}
