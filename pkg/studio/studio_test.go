package studio

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
)

func TestConeToFahrenheit(t *testing.T) {
	cases := []struct {
		cone string
		want float64
		ok   bool
	}{
		{"Cone 06", 1830.0, true},
		{"06", 1830.0, true},
		{"cone 6", 2232.0, true},
		{"Cone 8", 2305.0, true},
		{"Cone 10", 2381.0, true},
		{"10", 2381.0, true},
		{"Cone 11", 2400.0, true},
		{"Cone 99", 0, false},
	}
	for _, tc := range cases {
		got, ok := ConeToFahrenheit(tc.cone)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ConeToFahrenheit(%q) = (%v, %v), want (%v, %v)", tc.cone, got, ok, tc.want, tc.ok)
		}
	}
	if len(SupportedCones()) != 5 {
		t.Fatalf("expected 5 supported cones, got %d", len(SupportedCones()))
	}
}

func TestAllPresetsEvaluateCleanlyAndToneFree(t *testing.T) {
	presets := Presets()
	if len(presets) != 5 {
		t.Fatalf("expected 5 presets, got %d", len(presets))
	}

	wantIDs := []string{
		"electric_raku_6",
		"gas_soda_14",
		"wood_salt_24",
		"coop_wood_48",
		"foundry_glass_16",
	}

	bannedPhrases := []string{
		"Hobby " + "316L",
		"4-" + "lb",
		"gro" + "cery",
		"Switching from Salt " + "(NaCl) to Soda",
	}

	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())
	for i, wantID := range wantIDs {
		if presets[i].ID != wantID {
			t.Errorf("preset[%d].ID = %q, want %q", i, presets[i].ID, wantID)
		}
		found, ok := PresetByID(wantID)
		if !ok || found.ID != wantID {
			t.Fatalf("PresetByID(%q) failed", wantID)
		}
		res, err := eval.Evaluate(found.Project)
		if err != nil {
			t.Fatalf("preset %q Evaluate failed: %v", wantID, err)
		}
		if res.COTS.EstimatedTotalUSD <= 0 {
			t.Errorf("preset %q COTS total USD <= 0: %v", wantID, res.COTS.EstimatedTotalUSD)
		}
		if res.COTS.ActualPercentOfFlood <= 0 || res.COTS.ActualPercentOfFlood > 75.0 {
			t.Errorf("preset %q ActualPercentOfFlood = %.2f out of (0, 75]", wantID, res.COTS.ActualPercentOfFlood)
		}
		if res.BuildGuideMD == "" || res.OperationsChecklistMD == "" {
			t.Errorf("preset %q missing generated markdown guides", wantID)
		}

		// Verify zero banned phrases across full serialized evaluation & markdown guides.
		rawJSON, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("preset %q marshal: %v", wantID, err)
		}
		haystack := string(rawJSON) + "\n" + res.BuildGuideMD + "\n" + res.OperationsChecklistMD + "\n" + string(buildPacketReadme(res))
		for _, banned := range bannedPhrases {
			if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
				t.Errorf("preset %q output contains banned phrase %q", wantID, banned)
			}
		}
	}

	// Verify electric_raku_6 is 120V friendly and uses canonical standard_316l.
	raku, _ := PresetByID("electric_raku_6")
	rakuEval, err := eval.Evaluate(raku.Project)
	if err != nil {
		t.Fatalf("raku evaluate: %v", err)
	}
	if rakuEval.Project.CorrosionTier != CorrosionTierStandard316L {
		t.Errorf("electric_raku_6 corrosion_tier = %q, want %q", rakuEval.Project.CorrosionTier, CorrosionTierStandard316L)
	}
	if rakuEval.Urban.Corrosion.TierLabel != "Standard 316L SS (Light / Occasional)" {
		t.Errorf("unexpected tier_label: %q", rakuEval.Urban.Corrosion.TierLabel)
	}
	if rakuEval.Urban.Electrical.RequiredVoltage != 120 || !rakuEval.Urban.Electrical.FitsAvailableVoltage {
		t.Errorf("electric_raku_6 should fit 120V, got %+v", rakuEval.Urban.Electrical)
	}

	// Verify coop_wood_48 uses heavy_c276_frp, salting_hours=4.0, and passes corrosion cleanly.
	coop, _ := PresetByID("coop_wood_48")
	if coop.Project.SaltingHours != 4.0 {
		t.Errorf("coop_wood_48 salting_hours = %v, want 4.0", coop.Project.SaltingHours)
	}
	coopEval, err := eval.Evaluate(coop.Project)
	if err != nil {
		t.Fatalf("coop evaluate: %v", err)
	}
	if coopEval.Urban.Corrosion.Status != StatusOK || !strings.Contains(coopEval.Urban.Corrosion.MaterialSummary, "C-276") {
		t.Errorf("coop_wood_48 corrosion expected OK with C-276, got %+v", coopEval.Urban.Corrosion)
	}
}

func TestParseLoadSaveProjectPartialAndBackwardCompat(t *testing.T) {
	// Partial JSON referencing a preset should inherit that preset's defaults.
	partial := []byte(`{"preset_id":"gas_soda_14","studio_name":"Backyard Test","property_line_ft":45}`)
	proj, err := ParseProject(partial)
	if err != nil {
		t.Fatalf("ParseProject partial failed: %v", err)
	}
	if proj.StudioName != "Backyard Test" || proj.PropertyLineFt != 45 || proj.VaporGlazeMode != VaporGlazeSoda || proj.Input.KilnVolumeCuFt != 14.0 {
		t.Errorf("unexpected partial project: %+v", proj)
	}

	// Backward-compatible legacy corrosion_tier alias must parse cleanly and normalize to standard_316l.
	legacyJSON := []byte(`{"preset_id":"wood_salt_24","corrosion_tier":"hobby_316l","salting_hours":25.0,"firing_duration_hours":12.0}`)
	legacyProj, err := ParseProject(legacyJSON)
	if err != nil {
		t.Fatalf("ParseProject legacy corrosion_tier failed: %v", err)
	}
	if legacyProj.CorrosionTier != CorrosionTierStandard316L {
		t.Errorf("legacy corrosion_tier not normalized: got %q, want %q", legacyProj.CorrosionTier, CorrosionTierStandard316L)
	}
	if legacyProj.SaltingHours != 12.0 {
		t.Errorf("salting_hours should clamp to firing_duration_hours=12.0, got %v", legacyProj.SaltingHours)
	}

	// Validate directly accepts CorrosionTierHobby316L as well.
	directLegacy := DefaultProject()
	directLegacy.CorrosionTier = CorrosionTierHobby316L
	if err := directLegacy.Validate(); err != nil {
		t.Errorf("Validate rejected legacy CorrosionTierHobby316L: %v", err)
	}

	// Partial JSON without preset_id inherits DefaultProject().
	proj2, err := ParseProject([]byte(`{"firing_duration_hours": 12}`))
	if err != nil {
		t.Fatalf("ParseProject default-seeded failed: %v", err)
	}
	if proj2.FiringDurationHours != 12 || proj2.Input.KilnVolumeCuFt != 24.0 || proj2.SaltingHours != 3.0 {
		t.Errorf("unexpected default-seeded project: %+v", proj2)
	}
	if proj2.StageAlignment != StageAlignmentInlineIDBlower || !proj2.SharedBaffledSump || proj2.ControlTier != ControlTierRPiModbus {
		t.Errorf("unexpected Phase 7 defaults: %+v", proj2)
	}

	// Round-trip SaveProject / LoadProject.
	tmpFile := filepath.Join(t.TempDir(), "project.json")
	if err := SaveProject(tmpFile, proj); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}
	loaded, err := LoadProject(tmpFile)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if loaded.StudioName != proj.StudioName || loaded.Input.KilnVolumeCuFt != proj.Input.KilnVolumeCuFt {
		t.Errorf("loaded project mismatch: got %+v, want %+v", loaded, proj)
	}

	// Invalid values should fail validation.
	for _, bad := range []string{
		`{"available_voltage": 480}`,
		`{"vapor_glaze_mode": "chlorine"}`,
		`{"stage_alignment": "turbo_jet"}`,
		`{"packed_bed_layout": "spiral"}`,
		`{"venturi_layout": "quad_cone"}`,
		`{"control_tier": "bluetooth"}`,
		`{"salting_hours": -2}`,
	} {
		if _, err := ParseProject([]byte(bad)); err == nil {
			t.Errorf("expected error for invalid JSON %s", bad)
		}
	}
}

func TestCOTSSnappingRecalculatesFloodLEDesign(t *testing.T) {
	eval := NewEvaluator(nil, nil)
	for _, preset := range Presets() {
		proj := preset.Project
		proj.SnapToCOTS = true
		res, err := eval.Evaluate(proj)
		if err != nil {
			t.Fatalf("Evaluate(%s): %v", preset.ID, err)
		}
		if !res.COTS.Snapped {
			t.Errorf("%s: expected COTS.Snapped=true", preset.ID)
		}
		if res.COTS.ActualPercentOfFlood > res.Design.PackedBed.PercentOfFlood+1e-6 {
			t.Errorf("%s: snapped ActualPercentOfFlood (%.3f%%) > design PercentOfFlood (%.3f%%)",
				preset.ID, res.COTS.ActualPercentOfFlood, res.Design.PackedBed.PercentOfFlood)
		}
		var sumCost, sumWeight float64
		for _, item := range res.COTS.BOM.Items {
			if item.WeightKg <= 0 {
				t.Errorf("%s: BOM item %q has non-positive WeightKg %v", preset.ID, item.Name, item.WeightKg)
			}
			sumCost += item.TotalCostUSD
			sumWeight += item.WeightKg
		}
		if math.Abs(sumCost-res.COTS.BOM.TotalCostUSD) > 0.01 {
			t.Errorf("%s: BOM cost sum %.2f != TotalCostUSD %.2f", preset.ID, sumCost, res.COTS.BOM.TotalCostUSD)
		}
		if math.Abs(sumWeight-res.COTS.BOM.TotalWeightKg) > 0.01 {
			t.Errorf("%s: BOM weight sum %.2f != TotalWeightKg %.2f", preset.ID, sumWeight, res.COTS.BOM.TotalWeightKg)
		}
	}
}

func TestUrbanNoiseVoltageAndPeakSaltingBulkBakingSoda(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	// 1. 24 cu.ft. wood/salt kiln on 120V must flag action_needed for electrical (requires 240V >= 1.5 HP).
	proj := DefaultProject()
	proj.AvailableVoltage = 120
	proj.IncludeSilencer = false
	proj.PropertyLineFt = 20.0
	proj.FiringDurationHours = 16.0
	proj.SaltingHours = 3.0

	res, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if res.Urban.Electrical.RequiredVoltage != 240 || res.Urban.Electrical.FitsAvailableVoltage || res.Urban.Electrical.Status != StatusActionNeeded {
		t.Errorf("expected 240V action_needed on 120V service, got %+v", res.Urban.Electrical)
	}

	// Verify inverse-square noise attenuation math: L2 = L1 - 20*log10(d2/3).
	wantBareLine := round1(res.Urban.Noise.BareDBA3Ft - 20.0*math.Log10(20.0/3.0))
	if math.Abs(res.Urban.Noise.PropertyLineDBA-wantBareLine) > 0.11 {
		t.Errorf("unsilenced PropertyLineDBA = %.1f, want %.1f", res.Urban.Noise.PropertyLineDBA, wantBareLine)
	}

	// Enabling silencer reduces property-line dBA by ~13 dBA.
	proj.IncludeSilencer = true
	proj.AvailableVoltage = 240
	resSil, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("Evaluate silenced: %v", err)
	}
	drop := res.Urban.Noise.PropertyLineDBA - resSil.Urban.Noise.PropertyLineDBA
	if drop < 12.5 || drop > 13.5 {
		t.Errorf("expected ~13 dBA silencer drop, got %.1f dBA", drop)
	}

	// Verify peak-window bulk baking soda math: uses SaltingHours (3.0 hrs -> ~9.8 lbs), NOT 16 hrs (~52.4 lbs)!
	wantLbs := round1(resSil.Design.PackedBed.NaHCO3ConsumptionKgPerHr * proj.SaltingHours / model.KgPerLb)
	if resSil.Urban.Drain.SaltingHours != 3.0 {
		t.Errorf("Drain.SaltingHours = %.1f, want 3.0", resSil.Urban.Drain.SaltingHours)
	}
	if resSil.Urban.Drain.BakingSodaLbsPerFiring != wantLbs {
		t.Errorf("BakingSodaLbsPerFiring = %.1f, want %.1f", resSil.Urban.Drain.BakingSodaLbsPerFiring, wantLbs)
	}
	if wantLbs < 9.0 || wantLbs > 11.0 {
		t.Errorf("expected ~9.8 lbs NaHCO3 for 3.0 hr peak salting, got %.1f lbs", wantLbs)
	}
	if resSil.Urban.Drain.BakingSoda15LbBags != 1 || resSil.Urban.Drain.BakingSoda50LbSacks != 1 {
		t.Errorf("expected 1x 15lb bag & 1x 50lb sack, got %+v", resSil.Urban.Drain)
	}
	if !strings.Contains(resSil.Urban.Drain.BulkPackagingLabel, "15 lb") || !strings.Contains(resSil.Urban.Drain.BulkPackagingLabel, "50 lb") {
		t.Errorf("unexpected BulkPackagingLabel: %q", resSil.Urban.Drain.BulkPackagingLabel)
	}
}

func TestVaporGlazeModesAndNonPreachyMaintenance(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	saltProj := DefaultProject()
	saltProj.VaporGlazeMode = VaporGlazeSalt
	saltProj.Input.SaltRateLbsPerHr = 2.0
	saltEval, err := eval.Evaluate(saltProj)
	if err != nil {
		t.Fatalf("salt evaluate: %v", err)
	}
	if saltEval.Design.PackedBed.HClInletPPM <= 500 {
		t.Fatalf("expected substantial salt HCl ppm, got %.1f", saltEval.Design.PackedBed.HClInletPPM)
	}
	if saltEval.Urban.Corrosion.Status != StatusCaution {
		t.Errorf("expected standard_316l salt corrosion status caution, got %q", saltEval.Urban.Corrosion.Status)
	}
	if !strings.Contains(saltEval.Urban.Corrosion.SodaComparison, "pH 8") || !strings.Contains(saltEval.Urban.Corrosion.SodaComparison, "weld") {
		t.Errorf("expected practical buffered flush & weld inspection guidance, got %q", saltEval.Urban.Corrosion.SodaComparison)
	}

	// Switch to soda mode -> zero HCl, zero baking soda, practical zero-HCl note.
	sodaProj := saltProj
	sodaProj.VaporGlazeMode = VaporGlazeSoda
	sodaEval, err := eval.Evaluate(sodaProj)
	if err != nil {
		t.Fatalf("soda evaluate: %v", err)
	}
	if sodaEval.Design.Input.SaltRateLbsPerHr != 0 || sodaEval.Design.PackedBed.HClInletPPM != 0 {
		t.Errorf("expected 0 HCl in soda mode, got rate=%v ppm=%v", sodaEval.Design.Input.SaltRateLbsPerHr, sodaEval.Design.PackedBed.HClInletPPM)
	}
	if sodaEval.Urban.Drain.BakingSodaLbsPerFiring != 0 || sodaEval.Urban.Drain.BakingSoda15LbBags != 0 || sodaEval.Urban.Drain.BakingSoda50LbSacks != 0 {
		t.Errorf("expected 0 bulk soda in soda mode, got %+v", sodaEval.Urban.Drain)
	}
	if sodaEval.Urban.Corrosion.Status != StatusOK || !strings.Contains(sodaEval.Urban.Corrosion.SodaComparison, "Zero HCl") {
		t.Errorf("expected zero-HCl note in soda mode, got %+v", sodaEval.Urban.Corrosion)
	}
}

func TestPhase7EductorParallelBedParallelVenturiCompactAndControls(t *testing.T) {
	eval := NewEvaluator(nil, nil)
	baseProj := DefaultProject()
	baseEval, err := eval.Evaluate(baseProj)
	if err != nil {
		t.Fatalf("baseline evaluate: %v", err)
	}

	// 1. Dry-Fan Vertical-Downflow Eductor.
	edProj := baseProj
	edProj.StageAlignment = StageAlignmentEductorDryFan
	edEval, err := eval.Evaluate(edProj)
	if err != nil {
		t.Fatalf("eductor evaluate: %v", err)
	}
	ed := edEval.ArchitectureOptions.Eductor
	if !ed.Enabled || !ed.Feasible {
		t.Errorf("expected enabled & feasible eductor on 24 cu.ft. kiln, got %+v", ed)
	}
	if ed.EntrainmentRatio < 1.20 || ed.EntrainmentRatio > 1.65 {
		t.Errorf("EntrainmentRatio = %.2f out of [1.20, 1.65]", ed.EntrainmentRatio)
	}
	if ed.MixedTempF >= edEval.Design.Quench.OutletTemp.Fahrenheit()-20.0 {
		t.Errorf("expected cooler mixed_temp_f (%.1f F) vs quench (%.1f F)", ed.MixedTempF, edEval.Design.Quench.OutletTemp.Fahrenheit())
	}
	if ed.EnlargedBedFloodPct <= 0 || ed.EnlargedBedFloodPct > 75.0 {
		t.Errorf("EnlargedBedFloodPct = %.1f%% out of (0, 75]", ed.EnlargedBedFloodPct)
	}
	if ed.PositivePressureInWC <= 0.5 || ed.DryFanCostUSD >= ed.WetBlowerSavedUSD {
		t.Errorf("unexpected eductor pressure/cost: +%.2f inWC, dry=$%.0f vs wet=$%.0f", ed.PositivePressureInWC, ed.DryFanCostUSD, ed.WetBlowerSavedUSD)
	}
	if len(ed.Warnings) == 0 || !strings.Contains(ed.Warnings[0], "OUTDOOR-ONLY POSITIVE PRESSURE") {
		t.Errorf("missing positive-pressure outdoor warning: %+v", ed.Warnings)
	}
	if !strings.Contains(edEval.BuildGuideMD, "Dry-Fan Vertical-Downflow Static Venturi Eductor") || !strings.Contains(edEval.BuildGuideMD, "Soap-Bubble") {
		t.Error("BuildGuideMD missing eductor & positive-pressure soap-bubble commissioning steps")
	}

	// 2. Parallel 4" Pipe Packed-Bed Manifold.
	pbProj := baseProj
	pbProj.PackedBedLayout = PackedBedLayoutParallelManifold
	pbEval, err := eval.Evaluate(pbProj)
	if err != nil {
		t.Fatalf("parallel bed evaluate: %v", err)
	}
	pb := pbEval.ArchitectureOptions.ParallelBed
	if !pb.Enabled || pb.PipeNominalIn != 4.0 || pb.PipeCount < 6 {
		t.Errorf("unexpected parallel bed sizing for 24 cu.ft.: %+v", pb)
	}
	if pb.WallRatioDToDp < 7.5 || pb.ActualFloodPct > 75.0 {
		t.Errorf("parallel bed constraint violation: D/d_p=%.2f, flood=%.1f%%", pb.WallRatioDToDp, pb.ActualFloodPct)
	}
	if pb.BedDPInWC <= pb.SingleColumnDPInWC {
		t.Errorf("expected smaller 1/2\" packing to raise bed dP vs 1\" rings: %.2f vs %.2f", pb.BedDPInWC, pb.SingleColumnDPInWC)
	}
	if !strings.Contains(strings.Join(pb.Notes, " "), "LIQUID DISTRIBUTION HEADER") {
		t.Errorf("parallel bed missing mandatory liquid distribution header note: %+v", pb.Notes)
	}

	// Tiny 6 cu.ft. kiln with explicit 3" parallel pipe uses 3" pipe + 3/8" packing (D/d_p >= 7.5).
	raku, _ := PresetByID("electric_raku_6")
	rakuPB := raku.Project
	rakuPB.PackedBedLayout = PackedBedLayoutParallelManifold
	rakuPB.ParallelPipeSizeIn = 3.0
	rakuPBEval, err := eval.Evaluate(rakuPB)
	if err != nil {
		t.Fatalf("raku parallel bed evaluate: %v", err)
	}
	if rakuPBEval.ArchitectureOptions.ParallelBed.PipeNominalIn != 3.0 || rakuPBEval.ArchitectureOptions.ParallelBed.WallRatioDToDp < 7.5 {
		t.Errorf("expected 3\" pipe & D/d_p >= 7.5 for 6 cu.ft. kiln, got %+v", rakuPBEval.ArchitectureOptions.ParallelBed)
	}

	// Combined Eductor + Parallel Bed couples mixed airflow into the parallel manifold!
	comboProj := baseProj
	comboProj.StageAlignment = StageAlignmentEductorDryFan
	comboProj.PackedBedLayout = PackedBedLayoutParallelManifold
	comboEval, err := eval.Evaluate(comboProj)
	if err != nil {
		t.Fatalf("eductor+parallel bed evaluate: %v", err)
	}
	if comboEval.ArchitectureOptions.ParallelBed.PipeCount <= pb.PipeCount {
		t.Errorf("eductor mixed flow should require more parallel pipes (%d) than inline (%d)",
			comboEval.ArchitectureOptions.ParallelBed.PipeCount, pb.PipeCount)
	}

	// 3. Parallel Mini-Venturis + Wood-Ash Plugging Risk check.
	pvProj := baseProj
	pvProj.VenturiLayout = VenturiLayoutParallelMiniVenturis
	pvEval, err := eval.Evaluate(pvProj)
	if err != nil {
		t.Fatalf("parallel venturi evaluate: %v", err)
	}
	pv := pvEval.ArchitectureOptions.ParallelVenturi
	if !pv.Enabled || pv.ThroatCount < 2 || pv.ThroatCount > 4 || pv.ThroatDPInWC <= 0 {
		t.Errorf("unexpected parallel venturi result: %+v", pv)
	}
	// Small wood kiln triggers High plugging risk (< 1.0" throat in wood-ash service).
	smallWood := baseProj
	smallWood.Input.KilnVolumeCuFt = 6.0
	smallWood.VenturiLayout = VenturiLayoutParallelMiniVenturis
	smallWoodEval, err := eval.Evaluate(smallWood)
	if err != nil {
		t.Fatalf("small wood parallel venturi evaluate: %v", err)
	}
	if smallWoodEval.ArchitectureOptions.ParallelVenturi.PluggingRisk != "High" {
		t.Errorf("expected High plugging risk for sub-1\" wood-ash throat (%.2f\"), got %q",
			smallWoodEval.ArchitectureOptions.ParallelVenturi.ThroatDiameterIn,
			smallWoodEval.ArchitectureOptions.ParallelVenturi.PluggingRisk)
	}

	// 4. Compact U-Turn Footprint + Dual Sump + Electromechanical-Only Controls.
	saverProj := baseProj
	saverProj.CompactFootprint = true
	saverProj.SharedBaffledSump = true
	saverProj.ControlTier = ControlTierElectromechanicalOnly
	saverEval, err := eval.Evaluate(saverProj)
	if err != nil {
		t.Fatalf("saver evaluate: %v", err)
	}
	if saverEval.ArchitectureOptions.FootprintSavingsPct < 44.0 || saverEval.ArchitectureOptions.FootprintSavingsPct > 51.0 {
		t.Errorf("FootprintSavingsPct = %.1f%%, want ~45-50%%", saverEval.ArchitectureOptions.FootprintSavingsPct)
	}
	ctrlSavings := -saverEval.ArchitectureOptions.ControlDeltaUSD
	if ctrlSavings < 900.0 || ctrlSavings > 1200.0 {
		t.Errorf("electromechanical savings = $%.0f, want $900-$1,200", ctrlSavings)
	}
	if !strings.Contains(saverEval.COTS.SensorKit.Notes, "200 F") || !strings.Contains(saverEval.BuildGuideMD, "electromechanical_only") {
		t.Error("electromechanical mode must preserve mandatory NFPA 86 200 F hardwired snap-disc")
	}

	// Dual dedicated sumps (shared_baffled_sump=false) increases cost vs baseline.
	dualSumpProj := baseProj
	dualSumpProj.SharedBaffledSump = false
	dualSumpEval, err := eval.Evaluate(dualSumpProj)
	if err != nil {
		t.Fatalf("dual sump evaluate: %v", err)
	}
	if dualSumpEval.ArchitectureOptions.SumpDeltaUSD <= 250.0 || dualSumpEval.COTS.EstimatedTotalUSD <= baseEval.COTS.EstimatedTotalUSD {
		t.Errorf("expected dual sump to raise cost: delta=$%.0f, total=$%.0f vs base=$%.0f",
			dualSumpEval.ArchitectureOptions.SumpDeltaUSD, dualSumpEval.COTS.EstimatedTotalUSD, baseEval.COTS.EstimatedTotalUSD)
	}

	t.Logf("24 cu.ft. Economics Summary: baseline=$%.2f | eductor=$%.2f (delta=%+.2f, dryFan=%.2gHP, mixed=%.1fF, bed=%.0f\"@%.1f%%) | parallelBed=$%.2f (%dx%.0f\" pipes, bedDP=%.2finWC) | parallelVenturi=$%.2f (%dx%.2f\" throats, risk=%s) | uTurn+EM=$%.2f (len %.1fft->%.1fft, -%.1f%%)",
		baseEval.COTS.EstimatedTotalUSD,
		edEval.COTS.EstimatedTotalUSD, edEval.ArchitectureOptions.CostDeltaUSD, ed.DryFanMotorHP, ed.MixedTempF, ed.EnlargedBedDiameterIn, ed.EnlargedBedFloodPct,
		pbEval.COTS.EstimatedTotalUSD, pb.PipeCount, pb.PipeNominalIn, pb.BedDPInWC,
		pvEval.COTS.EstimatedTotalUSD, pv.ThroatCount, pv.ThroatDiameterIn, pv.PluggingRisk,
		saverEval.COTS.EstimatedTotalUSD, baseEval.ArchitectureOptions.FootprintLengthFt, saverEval.ArchitectureOptions.FootprintLengthFt, saverEval.ArchitectureOptions.FootprintSavingsPct)
}

func TestPhase7RecirculationAndShopPlan(t *testing.T) {
	eval := NewEvaluator(nil, nil)
	proj := DefaultProject()
	ev, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}

	// RecirculationPlan checks.
	rc := ev.Recirculation
	if rc.SumpVolumeGal < 24.0 || rc.OnceThroughWaterGal <= rc.RecircMakeupWaterGal {
		t.Errorf("invalid water balance: %+v", rc)
	}
	if rc.WaterSavedPct < 65.0 || rc.WaterSavedPct > 92.0 {
		t.Errorf("WaterSavedPct = %.1f%% out of expected range", rc.WaterSavedPct)
	}
	if rc.SaltingHours != 3.0 || rc.BakingSoda15LbBags != 1 || rc.BakingSoda50LbSacks != 1 {
		t.Errorf("unexpected recirculation bulk soda: %+v", rc)
	}
	if rc.DumpTDSLimitGL != 150.0 || rc.DumpSpecificGravity != 1.105 || rc.FiringsBeforeFullDump < 1 {
		t.Errorf("unexpected dump thresholds: %+v", rc)
	}
	if rc.EndOfFiringTDSGL <= 0 || rc.HClNeutralizedLbs <= 0 || rc.InitialChargeLbs <= 0 {
		t.Errorf("expected positive salt neutralization chemistry: %+v", rc)
	}
	triggersText := strings.Join(rc.DumpTriggers, " ")
	if !strings.Contains(triggersText, "1.105") || !strings.Contains(triggersText, "140 F") || !strings.Contains(triggersText, "IPC 702.5") {
		t.Errorf("dump triggers missing hydrometer SG or IPC 702.5 rule: %s", triggersText)
	}

	// ShopPlan checks.
	sp := ev.ShopPlan
	if len(sp.Parts) < 9 || len(sp.CutList) < 6 {
		t.Errorf("expected rich ShopPlan parts (%d) and cut list (%d)", len(sp.Parts), len(sp.CutList))
	}
	if sp.Clearances.FlueToCombustiblesIn != 36.0 || sp.Clearances.PropertyLineMinFt != 30.0 || sp.Clearances.StackAboveRoofFt != 6.0 || sp.Clearances.ServiceAisleIn != 36.0 {
		t.Errorf("unexpected salt ShopPlan clearances: %+v", sp.Clearances)
	}
	foundFrustum := false
	for _, item := range sp.CutList {
		if item.Shape == "frustum" {
			foundFrustum = true
			if item.OuterRadiusIn <= item.InnerRadiusIn || item.SweepDeg <= 0 || item.SweepDeg > 360 {
				t.Errorf("invalid frustum flat pattern %q: R_in=%v R_out=%v sweep=%v", item.Name, item.InnerRadiusIn, item.OuterRadiusIn, item.SweepDeg)
			}
		}
	}
	if !foundFrustum {
		t.Error("expected at least one unfolded frustum cone in ShopPlan.CutList")
	}

	// Non-salt ShopPlan uses 10 ft property line setback.
	soda, _ := PresetByID("gas_soda_14")
	sodaEv, err := eval.Evaluate(soda.Project)
	if err != nil {
		t.Fatalf("soda evaluate: %v", err)
	}
	if sodaEv.ShopPlan.Clearances.PropertyLineMinFt != 10.0 {
		t.Errorf("non-salt PropertyLineMinFt = %v, want 10.0", sodaEv.ShopPlan.Clearances.PropertyLineMinFt)
	}
}

func TestBuildGuideAdaptsToBuildPath(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	cotsProj := DefaultProject()
	cotsProj.BuildPath = BuildPathCOTSNoWeld
	cotsEval, err := eval.Evaluate(cotsProj)
	if err != nil {
		t.Fatalf("cots evaluate: %v", err)
	}
	if !strings.Contains(cotsEval.BuildGuideMD, "No-Weld COTS Hardware") {
		t.Error("expected BuildGuideMD to reflect cots_no_weld build path")
	}
	if !strings.Contains(cotsEval.OperationsChecklistMD, "200 F") || !strings.Contains(cotsEval.OperationsChecklistMD, "0.25 GPM") {
		t.Error("expected OperationsChecklistMD to include 200 F and 0.25 GPM trip tests")
	}
	if !strings.Contains(cotsEval.OperationsChecklistMD, "1.105") || !strings.Contains(cotsEval.OperationsChecklistMD, "15 lb") {
		t.Error("expected OperationsChecklistMD to include brine hydrometer SG and 15 lb bulk bags")
	}

	weldProj := DefaultProject()
	weldProj.BuildPath = BuildPathWeldedDXF
	weldProj.CorrosionTier = CorrosionTierHeavyC276FRP
	weldEval, err := eval.Evaluate(weldProj)
	if err != nil {
		t.Fatalf("welded evaluate: %v", err)
	}
	if !strings.Contains(weldEval.BuildGuideMD, "Rolled Sheet Metal & DXF Fabrication") {
		t.Error("expected BuildGuideMD to reflect welded_dxf build path")
	}
	if !strings.Contains(weldEval.BuildGuideMD, "ERNiCrMo-4") {
		t.Error("expected BuildGuideMD to reference Alloy C-276 welding filler for heavy_c276_frp")
	}
}

func TestBuildPacketZipContainsAllElevenValidFiles(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())
	proj := DefaultProject()

	zipBytes, ev, err := eval.BuildPacketZip(proj)
	if err != nil {
		t.Fatalf("BuildPacketZip failed: %v", err)
	}
	if len(zipBytes) == 0 {
		t.Fatal("zipBytes is empty")
	}
	if ev.Project.PresetID != "wood_salt_24" {
		t.Errorf("unexpected eval preset_id %q", ev.Project.PresetID)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader failed: %v", err)
	}

	wantFiles := []string{
		"scrubber.stl",
		"scrubber.dxf",
		"shop_plan.svg",
		"shop_plan.json",
		"bom_cots.csv",
		"bom_cots.md",
		"bom_engineered.csv",
		"BUILD_GUIDE.md",
		"OPERATIONS_CHECKLIST.md",
		"project.json",
		"daemon.json",
		"README_PACKET.txt",
	}

	got := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", f.Name, err)
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(rc); err != nil {
			_ = rc.Close()
			t.Fatalf("read zip entry %s: %v", f.Name, err)
		}
		_ = rc.Close()
		got[f.Name] = buf.Bytes()
	}

	if len(got) != len(wantFiles) {
		t.Fatalf("zip contains %d files, want %d", len(got), len(wantFiles))
	}
	for _, name := range wantFiles {
		data, ok := got[name]
		if !ok {
			t.Errorf("zip missing required file %q", name)
			continue
		}
		if len(data) == 0 {
			t.Errorf("zip entry %q is empty", name)
		}
	}

	// Verify shop_plan.json unmarshals cleanly into ShopPlan.
	var parsedPlan ShopPlan
	if err := json.Unmarshal(got["shop_plan.json"], &parsedPlan); err != nil {
		t.Fatalf("packet shop_plan.json invalid JSON: %v", err)
	}
	if len(parsedPlan.Parts) == 0 || len(parsedPlan.CutList) == 0 {
		t.Errorf("packet shop_plan.json missing parts or cut_list: %+v", parsedPlan)
	}

	// Verify project.json round-trips cleanly via ParseProject.
	if _, err := ParseProject(got["project.json"]); err != nil {
		t.Errorf("packet project.json failed ParseProject: %v", err)
	}

	// Verify daemon.json has valid control config matching model.ControlConfig.
	var daemonParsed struct {
		HALMode string              `json:"hal_mode"`
		Control model.ControlConfig `json:"control"`
	}
	if err := json.Unmarshal(got["daemon.json"], &daemonParsed); err != nil {
		t.Fatalf("packet daemon.json invalid JSON: %v", err)
	}
	if daemonParsed.HALMode != "sim" {
		t.Errorf("daemon.json hal_mode = %q, want sim", daemonParsed.HALMode)
	}
	if err := daemonParsed.Control.Validate(); err != nil {
		t.Errorf("daemon.json control config invalid: %v", err)
	}

	// Verify StudioEvaluation JSON tags match the exact cross-package contract.
	evalJSON, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal StudioEvaluation: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(evalJSON, &raw); err != nil {
		t.Fatalf("unmarshal StudioEvaluation raw: %v", err)
	}
	for _, key := range []string{"project", "design", "cots", "urban", "architecture_options", "recirculation", "shop_plan", "build_guide_md", "operations_checklist_md"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("StudioEvaluation JSON missing top-level key %q", key)
		}
	}

	if fn := PacketFilename(proj); fn != "kilnscrub-wood_salt_24-packet.zip" {
		t.Errorf("PacketFilename = %q, want kilnscrub-wood_salt_24-packet.zip", fn)
	}
}

func TestPlumeDilutionSquirrelCagePsychrometrics(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	proj := DefaultProject()
	proj.AmbientWinterF = 40.0
	proj.PropertyLineFt = 30.0
	proj.IncludePlumeDiluter = false

	bareEval, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("bare evaluate: %v", err)
	}
	pl := bareEval.Urban.Plume
	if pl.ColdDilutionWorks {
		t.Error("expected ColdDilutionWorks=false for 184 F saturated exhaust in 40 F winter air")
	}
	if pl.DilutionReheatKW < 42.0 || pl.DilutionReheatKW > 56.0 {
		t.Errorf("DilutionReheatKW = %.1f kW, want ~48.6 kW (within 42..56)", pl.DilutionReheatKW)
	}
	if pl.DilutionCFMNeeded != 1500.0 {
		t.Errorf("DilutionCFMNeeded = %.0f CFM, want 1500 CFM for 24 cu.ft. kiln", pl.DilutionCFMNeeded)
	}
	if pl.BarePlumeLengthFt < 30.0 || pl.BarePlumeLengthFt > 50.0 {
		t.Errorf("BarePlumeLengthFt = %.1f ft, want ~38.5 ft (within 30..50)", pl.BarePlumeLengthFt)
	}
	if pl.DilutedPlumeLengthFt <= 0 || pl.DilutedPlumeLengthFt >= pl.BarePlumeLengthFt*0.5 {
		t.Errorf("expected HV squirrel-cage diluter to cut plume reach >2x: bare=%.1f ft, diluted=%.1f ft",
			pl.BarePlumeLengthFt, pl.DilutedPlumeLengthFt)
	}
	if !strings.Contains(pl.SquirrelCageVerdict, "inWC") || !strings.Contains(pl.SquirrelCageVerdict, "shroud") {
		t.Errorf("SquirrelCageVerdict missing expected engineering guidance: %s", pl.SquirrelCageVerdict)
	}
	if bareEval.COTS.PlumeDiluter != nil {
		t.Error("expected COTS.PlumeDiluter=nil when IncludePlumeDiluter=false")
	}

	proj.IncludePlumeDiluter = true
	dilEval, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("diluted evaluate: %v", err)
	}
	if dilEval.COTS.PlumeDiluter == nil {
		t.Fatal("expected COTS.PlumeDiluter != nil when IncludePlumeDiluter=true")
	}
	if dilEval.COTS.EstimatedTotalUSD <= bareEval.COTS.EstimatedTotalUSD {
		t.Errorf("expected BOM total to increase with PlumeDiluter: %.2f vs %.2f",
			dilEval.COTS.EstimatedTotalUSD, bareEval.COTS.EstimatedTotalUSD)
	}
	if !strings.Contains(dilEval.BuildGuideMD, "Squirrel-Cage Plume Dilution Assembly") ||
		!strings.Contains(dilEval.BuildGuideMD, "shroud") {
		t.Error("expected BuildGuideMD to include squirrel-cage mixing tee & kiln-jacket shroud steps")
	}
	if dilEval.Urban.Plume.Status != StatusOK {
		t.Errorf("expected plume status ok when diluted length (%.1f ft) clears 30 ft property line, got %q",
			dilEval.Urban.Plume.DilutedPlumeLengthFt, dilEval.Urban.Plume.Status)
	}

	// Sub-freezing ambient (10 F) must evaluate cleanly without WaterPsat triple-point error.
	proj.AmbientWinterF = 10.0
	coldEval, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("sub-freezing evaluate failed: %v", err)
	}
	if coldEval.Urban.Plume.BarePlumeLengthFt <= bareEval.Urban.Plume.BarePlumeLengthFt {
		t.Errorf("expected colder 10 F air to lengthen bare plume vs 40 F: %.1f vs %.1f",
			coldEval.Urban.Plume.BarePlumeLengthFt, bareEval.Urban.Plume.BarePlumeLengthFt)
	}
}

func TestUSCodeComplianceC1ThroughC14(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	refProj := DefaultProject()
	refEval, err := eval.Evaluate(refProj)
	if err != nil {
		t.Fatalf("ref evaluate: %v", err)
	}
	if !strings.Contains(refEval.Urban.Electrical.CircuitLabel, "240V 30A") {
		t.Errorf("C1: 2.0 HP 240V blower must size to 240V 30A per NEC 430.122(A), got %q (%.1f A)",
			refEval.Urban.Electrical.CircuitLabel, refEval.Urban.Electrical.RequiredAmps)
	}
	if refEval.Urban.Electrical.RequiredAmps < 23.0 || refEval.Urban.Electrical.RequiredAmps > 30.0 {
		t.Errorf("C1: 2.0 HP 240V NEC 430.122 ampacity = %.1f A, want ~24.6 A", refEval.Urban.Electrical.RequiredAmps)
	}
	for _, want := range []string{"NEC 430.102", "NEC 210.8(F)", "NEC 400.12"} {
		if !strings.Contains(refEval.Urban.Electrical.Summary, want) {
			t.Errorf("C2: electrical summary missing %q: %s", want, refEval.Urban.Electrical.Summary)
		}
	}

	dilProj := refProj
	dilProj.IncludePlumeDiluter = true
	dilEval, err := eval.Evaluate(dilProj)
	if err != nil {
		t.Fatalf("diluter evaluate: %v", err)
	}
	if dilEval.Urban.Electrical.RequiredAmps <= refEval.Urban.Electrical.RequiredAmps {
		t.Errorf("C1: IncludePlumeDiluter must increase RequiredAmps: %.1f vs %.1f",
			dilEval.Urban.Electrical.RequiredAmps, refEval.Urban.Electrical.RequiredAmps)
	}
	if !strings.Contains(dilEval.Urban.Electrical.Summary, "plume diluter") {
		t.Errorf("C1: electrical summary should note plume diluter load: %s", dilEval.Urban.Electrical.Summary)
	}

	soda, _ := PresetByID("gas_soda_14")
	soda120 := soda.Project
	soda120.AvailableVoltage = 120
	soda120Eval, err := eval.Evaluate(soda120)
	if err != nil {
		t.Fatalf("soda 120V evaluate: %v", err)
	}
	if !strings.Contains(soda120Eval.Urban.Electrical.CircuitLabel, "120V 30A") || soda120Eval.Urban.Electrical.Status != StatusCaution {
		t.Errorf("C1: 1.0 HP on 120V expected 120V 30A caution, got %+v", soda120Eval.Urban.Electrical)
	}
	if !strings.Contains(soda120Eval.COTS.BlowerVFD.SKU, "GS21-11P0") {
		t.Errorf("C3: 1.0 HP 120V VFD SKU expected GS21-11P0, got %q", soda120Eval.COTS.BlowerVFD.SKU)
	}

	if strings.Contains(refEval.COTS.BlowerVFD.SKU, "GS21-20P") || !strings.Contains(refEval.COTS.BlowerVFD.SKU, "GS21-22P0") {
		t.Errorf("C3: 2.0 HP VFD SKU expected GS21-22P0, got %q", refEval.COTS.BlowerVFD.SKU)
	}
	if !strings.Contains(refEval.COTS.BlowerVFD.Material, "NEMA 4X") {
		t.Errorf("C3: VFD must specify NEMA 4X enclosure, got %q", refEval.COTS.BlowerVFD.Material)
	}

	if !strings.Contains(refEval.Urban.Water.Summary, "IPC Section 608") || !strings.Contains(refEval.Urban.Water.Summary, "ASSE 1013") {
		t.Errorf("C4: WaterAssessment missing IPC 608 / ASSE 1013: %s", refEval.Urban.Water.Summary)
	}
	hasRPZItem := false
	hasDisconnectItem := false
	for _, it := range refEval.COTS.BOM.Items {
		if strings.Contains(it.Spec, "IPC608-RPZ-AG") {
			hasRPZItem = true
		}
		if strings.Contains(it.Spec, "NEC430-DISC-LR") {
			hasDisconnectItem = true
		}
	}
	if !hasRPZItem || !hasDisconnectItem {
		t.Errorf("C2/C4: COTS BOM missing RPZ (%v) or disconnect (%v)", hasRPZItem, hasDisconnectItem)
	}
	if !strings.Contains(refEval.BuildGuideMD, "IPC Section 608") || !strings.Contains(refEval.OperationsChecklistMD, "IPC 608") {
		t.Error("C4: guides missing IPC 608 backflow checklist items")
	}

	if !strings.Contains(refEval.Urban.Drain.Summary, "140 F") || !strings.Contains(refEval.Urban.Drain.Summary, "40 CFR 403.5") {
		t.Errorf("C5: DrainAssessment missing IPC 702.5 <=140 F or 40 CFR 403.5: %s", refEval.Urban.Drain.Summary)
	}
	if !strings.Contains(refEval.OperationsChecklistMD, "<= 140 F") || !strings.Contains(refEval.OperationsChecklistMD, "40 CFR 403.5") {
		t.Error("C5: OperationsChecklistMD missing IPC 702.5 <=140 F or 40 CFR 403.5")
	}

	for _, p := range Presets() {
		ev, err := eval.Evaluate(p.Project)
		if err != nil {
			t.Fatalf("preset %s: %v", p.ID, err)
		}
		if strings.Contains(ev.COTS.PackedColumn.SKU, "HDPE") {
			t.Errorf("C6: preset %s selected non-vacuum-rated HDPE drum for packed column: %s", p.ID, ev.COTS.PackedColumn.SKU)
		}
	}
	if !strings.Contains(refEval.COTS.SensorKit.Notes, "200 F") || !strings.Contains(refEval.BuildGuideMD, "200 F") {
		t.Error("C6: SensorKit and BuildGuideMD must specify independent hardwired 200 F CPVC snap-disc")
	}

	if strings.Contains(strings.ToLower(refEval.Urban.Plume.Summary), "harmless water vapor") {
		t.Errorf("C8: salt-mode plume summary must NEVER call HCl exhaust harmless water vapor: %s", refEval.Urban.Plume.Summary)
	}
	if !strings.Contains(refEval.Urban.Plume.Summary, "OSHA 29 CFR 1910.1000") || !strings.Contains(refEval.Urban.Plume.Summary, "IMC 510") {
		t.Errorf("C8: salt-mode plume summary missing OSHA/IMC 510 notice: %s", refEval.Urban.Plume.Summary)
	}
	tightSalt := refProj
	tightSalt.PropertyLineFt = 20.0
	tightSaltEval, err := eval.Evaluate(tightSalt)
	if err != nil {
		t.Fatalf("tightSalt evaluate: %v", err)
	}
	if tightSaltEval.Urban.Plume.Status != StatusActionNeeded {
		t.Errorf("C8: salt mode with 20 ft property line (< 30 ft IMC 501.3.1) expected action_needed, got %q", tightSaltEval.Urban.Plume.Status)
	}

	if strings.Contains(refEval.BuildGuideMD, "rated 650 F") {
		t.Error("C9: BuildGuideMD must not recommend 650 F RTV silicone on a 2400 F flue")
	}
	for _, want := range []string{"2300 F", `36"`, "NFPA 211", "NFPA 91"} {
		if !strings.Contains(refEval.BuildGuideMD, want) {
			t.Errorf("C9/C11: BuildGuideMD missing %q", want)
		}
	}
	if strings.Contains(dilEval.COTS.PlumeDiluter.Material, "Galvanized/316L") || !strings.Contains(dilEval.COTS.PlumeDiluter.Material, "CPVC") {
		t.Errorf("C12: plume diluter mixing tee must be CPVC/FRP/316L (not galvanized): %s", dilEval.COTS.PlumeDiluter.Material)
	}
	unsilProj := refProj
	unsilProj.IncludeSilencer = false
	unsilEval, err := eval.Evaluate(unsilProj)
	if err != nil {
		t.Fatalf("unsilenced evaluate: %v", err)
	}
	if !strings.Contains(unsilEval.Urban.Noise.Summary, "OSHA 29 CFR 1910.95") {
		t.Errorf("C13: unsilenced blower missing OSHA 29 CFR 1910.95 hearing warning: %s", unsilEval.Urban.Noise.Summary)
	}
}

func TestStudioBranchCoverageAndEdgeCases(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	// Cover all EffectiveDesignInput switch arms.
	effCases := []struct {
		mode string
		fuel string
		want string
	}{
		{VaporGlazeNone, "", model.FuelGas},
		{VaporGlazeNone, model.FuelGasSalt, model.FuelGas},
		{VaporGlazeSoda, model.FuelWoodSalt, model.FuelWood},
		{VaporGlazeSalt, "", model.FuelWoodSalt},
		{VaporGlazeSalt, model.FuelWood, model.FuelWoodSalt},
		{VaporGlazeSalt, model.FuelGas, model.FuelGasSalt},
	}
	for _, tc := range effCases {
		p := DefaultProject()
		p.VaporGlazeMode = tc.mode
		p.Input.FuelType = tc.fuel
		got := p.EffectiveDesignInput()
		if got.FuelType != tc.want {
			t.Errorf("EffectiveDesignInput(mode=%q, fuel=%q) = %q, want %q", tc.mode, tc.fuel, got.FuelType, tc.want)
		}
	}

	// Cover Normalize & EffectiveSaltingHours zero/empty arms.
	emptyOpts := DefaultProject()
	emptyOpts.CorrosionTier = ""
	emptyOpts.StageAlignment = ""
	emptyOpts.PackedBedLayout = ""
	emptyOpts.VenturiLayout = ""
	emptyOpts.ControlTier = ""
	emptyOpts.SaltingHours = 0
	norm := emptyOpts.Normalize()
	if norm.CorrosionTier != CorrosionTierStandard316L || norm.StageAlignment != StageAlignmentInlineIDBlower ||
		norm.PackedBedLayout != PackedBedLayoutSingleColumn || norm.VenturiLayout != VenturiLayoutSingleThroat ||
		norm.ControlTier != ControlTierRPiModbus || norm.SaltingHours != 3.0 {
		t.Errorf("Normalize defaults failed: %+v", norm)
	}
	if hrs := emptyOpts.EffectiveSaltingHours(); hrs != 3.0 {
		t.Errorf("EffectiveSaltingHours(0 in salt) = %v, want 3.0", hrs)
	}
	emptyOpts.SaltingHours = 50.0
	if hrs := emptyOpts.EffectiveSaltingHours(); hrs != emptyOpts.FiringDurationHours {
		t.Errorf("EffectiveSaltingHours clamp = %v, want %v", hrs, emptyOpts.FiringDurationHours)
	}

	// Cover StudioProject.Validate invalid arms.
	badProjects := []func(*StudioProject){
		func(p *StudioProject) { p.FiringDurationHours = 0 },
		func(p *StudioProject) { p.SaltingHours = -1 },
		func(p *StudioProject) { p.BuildPath = "riveted" },
		func(p *StudioProject) { p.CorrosionTier = "carbon_steel" },
		func(p *StudioProject) { p.StageAlignment = "bad_align" },
		func(p *StudioProject) { p.PackedBedLayout = "bad_bed" },
		func(p *StudioProject) { p.VenturiLayout = "bad_vent" },
		func(p *StudioProject) { p.ControlTier = "bad_ctrl" },
		func(p *StudioProject) { p.PropertyLineFt = 1.0 },
		func(p *StudioProject) { p.AmbientWinterF = -100.0 },
		func(p *StudioProject) { p.GardenHoseGPM = 0 },
		func(p *StudioProject) { p.Input.KilnVolumeCuFt = -5 },
	}
	for i, mutate := range badProjects {
		p := DefaultProject()
		mutate(&p)
		if err := p.Validate(); err == nil {
			t.Errorf("badProjects[%d]: expected validation error", i)
		}
		if _, err := eval.Evaluate(p); err == nil {
			t.Errorf("badProjects[%d]: Evaluate expected error", i)
		}
	}

	// ParseProject edge cases: empty input, blank studio_name, bad JSON.
	if p, err := ParseProject([]byte("   ")); err != nil || p.PresetID != "wood_salt_24" {
		t.Errorf("ParseProject empty = (%+v, %v)", p, err)
	}
	if p, err := ParseProject([]byte(`{"studio_name":""}`)); err != nil || p.StudioName != "Studio Kiln Scrubber" {
		t.Errorf("ParseProject blank name = (%q, %v)", p.StudioName, err)
	}
	if _, err := ParseProject([]byte(`{not-json`)); err == nil {
		t.Error("ParseProject bad JSON expected error")
	}
	if _, err := LoadProject(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("LoadProject missing file expected error")
	}
	badSave := DefaultProject()
	badSave.AvailableVoltage = 999
	if err := SaveProject(filepath.Join(t.TempDir(), "bad.json"), badSave); err == nil {
		t.Error("SaveProject invalid project expected error")
	}

	// PacketFilename edge cases: space, empty preset_id, punctuation-only.
	pSpace := DefaultProject()
	pSpace.PresetID = ""
	pSpace.StudioName = "My Studio Kiln!"
	if fn := PacketFilename(pSpace); fn != "kilnscrub-my-studio-kiln-packet.zip" {
		t.Errorf("PacketFilename(space) = %q", fn)
	}
	pEmpty := DefaultProject()
	pEmpty.PresetID = ""
	pEmpty.StudioName = "!!!"
	if fn := PacketFilename(pEmpty); fn != "kilnscrub-studio-packet.zip" {
		t.Errorf("PacketFilename(empty) = %q", fn)
	}
	pBlank := DefaultProject()
	pBlank.PresetID = ""
	pBlank.StudioName = ""
	if fn := PacketFilename(pBlank); fn != "kilnscrub-studio-packet.zip" {
		t.Errorf("PacketFilename(blank) = %q", fn)
	}

	// Cover COTS sizing tiers, dryFanUnitCost tiers, pickNemaMotorHP, and plume diluter CFM tiers.
	for _, cfm := range []float64{60, 140, 240, 340, 550} {
		d := model.ScrubberDesign{}
		d.Blower.FlowACFM = cfm
		d.Blower.TotalStaticPressureInWC = 20.0
		gotCFM, cone := RecommendedDilutionCFM(d)
		if gotCFM <= 0 || cone <= 0 {
			t.Errorf("RecommendedDilutionCFM(%.0f) = (%.0f, %.0f)", cfm, gotCFM, cone)
		}
		if part := selectPlumeDiluter(DefaultProject(), d); part.UnitCostUSD <= 0 {
			t.Errorf("selectPlumeDiluter(%.0f) invalid cost", cfm)
		}
		if a := diluterAmps120V(d); a <= 0 {
			t.Errorf("diluterAmps120V(%.0f) <= 0", cfm)
		}
	}

	for _, hp := range []float64{0, 0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 7.5, 25.0} {
		_ = pickNemaMotorHP(hp)
		fc, vc, wt := dryFanUnitCost(hp)
		if fc <= 0 || vc <= 0 || wt <= 0 {
			t.Errorf("dryFanUnitCost(%v) invalid: %v %v %v", hp, fc, vc, wt)
		}
		for _, v := range []int{120, 240} {
			p := DefaultProject()
			p.AvailableVoltage = v
			p.IncludePlumeDiluter = true
			d := model.ScrubberDesign{}
			d.Blower.RecommendedMotorHP = hp
			d.Blower.FlowACFM = 260
			d.Blower.TotalStaticPressureInWC = 21
			if ea := assessElectrical(p, d); ea.RequiredAmps <= 0 || ea.CircuitLabel == "" {
				t.Errorf("assessElectrical(hp=%v, v=%d) invalid: %+v", hp, v, ea)
			}
			if vfd := selectBlowerVFD(p, d); vfd.UnitCostUSD <= 0 {
				t.Errorf("selectBlowerVFD(hp=%v, v=%d) invalid", hp, v)
			}
		}
	}

	// Large 100-gal sump, large nozzle tiers, FRP-24 column snapping, heavy eductor, heavy parallel venturi, and SnapToCOTS=false.
	bigDesign := model.ScrubberDesign{}
	bigDesign.Quench.TowerDiameter = model.InchesToMeters(16)
	bigDesign.Quench.TowerHeight = model.FeetToMeters(10)
	bigDesign.Quench.TotalSprayWaterGPM = 4.5
	bigDesign.Quench.NozzleCount = 2
	bigDesign.PackedBed.ColumnDiameter = model.InchesToMeters(20)
	bigDesign.PackedBed.ColumnHeight = model.FeetToMeters(12)
	bigDesign.PackedBed.PackingHeight = model.FeetToMeters(5)
	bigDesign.PackedBed.SumpVolume = model.CubicMeters(60.0 * model.CubicMetersPerUSGallon)
	bigDesign.PackedBed.FloodingVelocity = 3.0
	bigDesign.QuenchedGas.VolumetricFlow = 0.25
	bigDesign.Blower.RecommendedMotorHP = 3.0
	bigDesign.Blower.FlowACFM = 420
	bigDesign.Blower.TotalStaticPressureInWC = 22

	if sump := selectSumpTank(bigDesign); sump.SKU != "USPLASTIC-100T" {
		t.Errorf("expected 100-gal sump for 60 gal holdup, got %q", sump.SKU)
	}
	if nz := selectSprayNozzle(DefaultProject(), bigDesign); nz.NominalIn != 0.5 {
		t.Errorf("expected 1/2\" nozzle for 2.25 GPM/nozzle, got %v", nz.NominalIn)
	}
	bigDesign.Quench.TotalSprayWaterGPM = 2.0
	if nz := selectSprayNozzle(DefaultProject(), bigDesign); nz.NominalIn != 0.375 {
		t.Errorf("expected 3/8\" nozzle for 1.0 GPM/nozzle, got %v", nz.NominalIn)
	}
	bigDesign.Quench.NozzleCount = 0
	_ = selectSprayNozzle(DefaultProject(), bigDesign)

	if colFRP, _, _ := selectPackedColumn(DefaultProject(), bigDesign); !strings.Contains(colFRP.SKU, "FRP-24") {
		t.Errorf("expected FRP-24 vacuum column for 20\" ID, got %q", colFRP.SKU)
	}

	// Heavy C-276 + eductor + parallel bed + parallel venturi + unsnapped coverage.
	heavyCombo := DefaultProject()
	heavyCombo.CorrosionTier = CorrosionTierHeavyC276FRP
	heavyCombo.StageAlignment = StageAlignmentEductorDryFan
	heavyCombo.SnapToCOTS = false
	if _, err := eval.Evaluate(heavyCombo); err != nil {
		t.Fatalf("heavyCombo unsnapped eductor evaluate: %v", err)
	}
	heavyCombo.SnapToCOTS = true
	heavyCombo.PackedBedLayout = PackedBedLayoutParallelManifold
	if _, err := eval.Evaluate(heavyCombo); err != nil {
		t.Fatalf("heavyCombo snapped eductor+parallel evaluate: %v", err)
	}
	heavyPV := DefaultProject()
	heavyPV.CorrosionTier = CorrosionTierHeavyC276FRP
	heavyPV.VenturiLayout = VenturiLayoutParallelMiniVenturis
	heavyPV.Input.KilnVolumeCuFt = 48.0
	if ev, err := eval.Evaluate(heavyPV); err != nil {
		t.Fatalf("heavyPV 48 cuft evaluate: %v", err)
	} else if ev.ArchitectureOptions.ParallelVenturi.ThroatCount < 3 {
		t.Errorf("expected 3-4 throats for 48 cuft parallel venturi, got %d", ev.ArchitectureOptions.ParallelVenturi.ThroatCount)
	}

	// Unsnapped standard & heavy evaluations + urban water/noise/plume edge cases.
	unsnapStd := DefaultProject()
	unsnapStd.SnapToCOTS = false
	unsnapStd.GardenHoseGPM = 0.8
	unsnapStd.PropertyLineFt = 4.0
	unsnapStd.IncludeSilencer = true
	unsnapStd.AmbientWinterF = 55.0
	if ev, err := eval.Evaluate(unsnapStd); err != nil {
		t.Fatalf("unsnapStd evaluate: %v", err)
	} else {
		if ev.Urban.Water.Status != StatusActionNeeded || ev.Urban.Noise.Status != StatusActionNeeded {
			t.Errorf("expected water & noise action_needed, got water=%q noise=%q", ev.Urban.Water.Status, ev.Urban.Noise.Status)
		}
	}

	unsnapHeavy := DefaultProject()
	unsnapHeavy.SnapToCOTS = false
	unsnapHeavy.CorrosionTier = CorrosionTierHeavyC276FRP
	unsnapHeavy.VaporGlazeMode = VaporGlazeNone
	unsnapHeavy.GardenHoseGPM = 1.4
	unsnapHeavy.AmbientWinterF = 75.0
	unsnapHeavy.PropertyLineFt = 8.0
	if ev, err := eval.Evaluate(unsnapHeavy); err != nil {
		t.Fatalf("unsnapHeavy evaluate: %v", err)
	} else {
		if ev.Urban.Water.Status != StatusCaution || ev.Urban.Plume.Status != StatusActionNeeded {
			t.Errorf("expected water caution & plume action_needed, got water=%q plume=%q", ev.Urban.Water.Status, ev.Urban.Plume.Status)
		}
	}

	modProj := DefaultProject()
	modProj.VaporGlazeMode = VaporGlazeSoda
	modProj.AmbientWinterF = 55.0
	modProj.PropertyLineFt = 120.0
	modProj.IncludeSilencer = false
	if ev, err := eval.Evaluate(modProj); err != nil {
		t.Fatalf("modProj evaluate: %v", err)
	} else if ev.Urban.Plume.Status != StatusOK {
		t.Errorf("expected moderate plume OK at 120 ft, got %q", ev.Urban.Plume.Status)
	}

	mildProj := modProj
	mildProj.AmbientWinterF = 72.0
	mildProj.PropertyLineFt = 25.0
	if ev, err := eval.Evaluate(mildProj); err != nil {
		t.Fatalf("mildProj evaluate: %v", err)
	} else if ev.Urban.Plume.VisiblePlumeRisk != "low" {
		t.Errorf("expected low plume risk at 72 F, got %q", ev.Urban.Plume.VisiblePlumeRisk)
	}

	// Cover degenerate cylinder frustum fallback in buildFrustumCutItem & frustumBoundingBoxIn sweep > 270 deg.
	cylItem := buildFrustumCutItem(3, "Equal Ends", "316L SS", 6.0, 6.0, 10.0, 1, "test")
	if cylItem.Shape != "cylinder" || cylItem.FlatWidthIn <= 0 {
		t.Errorf("expected cylinder fallback for equal ends, got %+v", cylItem)
	}
	w, h := frustumBoundingBoxIn(2.0, 5.0, 1.8*math.Pi)
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive frustum bbox, got (%v, %v)", w, h)
	}
}

func TestPhase8BudgetBlowerDrivesShelllessPipesPackingAndCompressor(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())
	baseProj := DefaultProject()
	baseEval, err := eval.Evaluate(baseProj)
	if err != nil {
		t.Fatalf("baseline evaluate: %v", err)
	}

	// Verify JSON contract keys on architecture_options.
	archBytes, err := json.Marshal(baseEval.ArchitectureOptions)
	if err != nil {
		t.Fatalf("marshal architecture_options: %v", err)
	}
	var archRaw map[string]any
	if err := json.Unmarshal(archBytes, &archRaw); err != nil {
		t.Fatalf("unmarshal architecture_options: %v", err)
	}
	for _, key := range []string{"blower_drive", "compressor", "packing_media", "parallel_bed"} {
		if _, ok := archRaw[key]; !ok {
			t.Errorf("architecture_options JSON missing key %q", key)
		}
	}

	// 1. Shop Compressor always evaluated on EVERY Evaluate() call.
	comp := baseEval.ArchitectureOptions.Compressor
	if comp.AvailableSCFM != 15.0 {
		t.Errorf("default AvailableSCFM = %.1f, want 15.0", comp.AvailableSCFM)
	}
	if comp.EductorFeasible {
		t.Errorf("24 cu.ft. kiln should mark compressor eductor_feasible=false, got %+v", comp)
	}
	if comp.EductorSCFMNeeded < 65.0 || comp.EductorHPNeeded < 17.0 {
		t.Errorf("24 cu.ft. kiln eductor should need >=65 SCFM & >=17 HP, got %.1f SCFM / %.1f HP", comp.EductorSCFMNeeded, comp.EductorHPNeeded)
	}
	if !comp.AtomizeFeasible || comp.AtomizeSCFMNeeded < 4.0 || comp.AtomizeSCFMNeeded > 6.0 {
		t.Errorf("24 cu.ft. kiln air-atomizing should be feasible with 4-6 SCFM, got %+v", comp)
	}
	if comp.AtomizeDropletUm != 25.0 || comp.AtomizeDPSavedInWC < 6.0 || comp.AtomizeCostSavedUSD < 950.0 {
		t.Errorf("air-atomize savings below expected: %+v", comp)
	}

	// Tiny <= 6 cu.ft. low-dP test kiln marks compressor eductor_feasible=true!
	tinyKiln, _ := PresetByID("electric_raku_6")
	tinyProj := tinyKiln.Project
	tinyProj.Input.VenturiThroatVelocity = 32.0 // low-dP test kiln (~4-5 inWC)
	tinyProj.BlowerDriveOption = BlowerDriveCompressorEductor
	tinyEval, err := eval.Evaluate(tinyProj)
	if err != nil {
		t.Fatalf("tiny compressor eductor evaluate: %v", err)
	}
	if !tinyEval.ArchitectureOptions.Compressor.EductorFeasible || !tinyEval.ArchitectureOptions.BlowerDrive.Feasible {
		t.Errorf("tiny 6 cu.ft. low-dP kiln should be feasible on 15 SCFM shop compressor, got comp=%+v drive=%+v",
			tinyEval.ArchitectureOptions.Compressor, tinyEval.ArchitectureOptions.BlowerDrive)
	}

	// Undersized 3 SCFM compressor marks atomize_feasible=false.
	lowCompProj := baseProj
	lowCompProj.ShopCompressorSCFM = 3.0
	lowCompEval, err := eval.Evaluate(lowCompProj)
	if err != nil {
		t.Fatalf("lowComp evaluate: %v", err)
	}
	if lowCompEval.ArchitectureOptions.Compressor.AtomizeFeasible {
		t.Errorf("3 SCFM compressor should not satisfy 4-6 SCFM air-atomizing, got %+v", lowCompEval.ArchitectureOptions.Compressor)
	}

	// 2. All 5 BlowerDriveOption modes on 24 cu.ft. reference kiln.
	// 2a. compressor_air_atomize: exact Calvert active twin-fluid d_50 <= 0.22 um parity, slashed dP, saves ~$1,000+!
	airAtomProj := baseProj
	airAtomProj.BlowerDriveOption = BlowerDriveCompressorAirAtomize
	airAtomEval, err := eval.Evaluate(airAtomProj)
	if err != nil {
		t.Fatalf("compressor_air_atomize evaluate: %v", err)
	}
	activeD50 := activeTwinFluidD50Microns(airAtomEval.Design)
	if activeD50 > 0.22 {
		t.Errorf("compressor_air_atomize active twin-fluid d_50 = %.3f um, want <= 0.22 um parity", activeD50)
	}
	if airAtomEval.Design.Venturi.DropletDiameter.Microns() != 25.0 || airAtomEval.Design.Venturi.ThroatVelocity != 48.0 {
		t.Errorf("compressor_air_atomize expected 25 um @ 48 m/s, got %.1f um @ %.1f m/s",
			airAtomEval.Design.Venturi.DropletDiameter.Microns(), airAtomEval.Design.Venturi.ThroatVelocity)
	}
	if airAtomEval.Design.Venturi.PressureDropInWC >= baseEval.Design.Venturi.PressureDropInWC-6.0 {
		t.Errorf("compressor_air_atomize venturi dP = %.2f inWC did not drop enough vs baseline %.2f inWC",
			airAtomEval.Design.Venturi.PressureDropInWC, baseEval.Design.Venturi.PressureDropInWC)
	}
	if airAtomEval.Design.Blower.TotalStaticPressureInWC > 14.0 {
		t.Errorf("compressor_air_atomize total static = %.2f inWC, want ~12-13.5 inWC", airAtomEval.Design.Blower.TotalStaticPressureInWC)
	}
	if airAtomEval.ArchitectureOptions.BlowerDrive.CostDeltaUSD > -950.0 || !airAtomEval.ArchitectureOptions.BlowerDrive.Feasible {
		t.Errorf("compressor_air_atomize should save ~$1,000+ and be feasible, got %+v", airAtomEval.ArchitectureOptions.BlowerDrive)
	}
	if !strings.Contains(airAtomEval.BuildGuideMD, "Twin-Fluid Air-Atomizing") {
		t.Error("BuildGuideMD missing twin-fluid air-atomizing instructions")
	}

	// 2b. regen_side_channel: wet inline warns about dry-air corrosion; eductor mode clears wet caveat!
	regenWetProj := baseProj
	regenWetProj.BlowerDriveOption = BlowerDriveRegenSideChannel
	regenWetEval, err := eval.Evaluate(regenWetProj)
	if err != nil {
		t.Fatalf("regen_side_channel wet evaluate: %v", err)
	}
	rwDrive := regenWetEval.ArchitectureOptions.BlowerDrive
	if !rwDrive.Feasible || rwDrive.FanCostUSD < 380.0 || rwDrive.FanCostUSD > 520.0 {
		t.Errorf("regen_side_channel cost/feasibility unexpected: %+v", rwDrive)
	}
	if !strings.Contains(strings.Join(rwDrive.Caveats, " "), "DRY-AIR CORROSION CAVEAT") {
		t.Errorf("regen_side_channel wet mode missing dry-air corrosion caveat: %+v", rwDrive.Caveats)
	}

	regenDryProj := regenWetProj
	regenDryProj.StageAlignment = StageAlignmentEductorDryFan
	regenDryEval, err := eval.Evaluate(regenDryProj)
	if err != nil {
		t.Fatalf("regen_side_channel eductor evaluate: %v", err)
	}
	if strings.Contains(strings.Join(regenDryEval.ArchitectureOptions.BlowerDrive.Caveats, " "), "DRY-AIR CORROSION CAVEAT") {
		t.Error("regen_side_channel with eductor_dry_fan should not emit wet corrosion caveat")
	}

	// 2c. dual_bypass_blowers: $380-$460, isolated windings, $15 brush swap every ~45 firings.
	dualProj := baseProj
	dualProj.BlowerDriveOption = BlowerDriveDualBypassBlowers
	dualEval, err := eval.Evaluate(dualProj)
	if err != nil {
		t.Fatalf("dual_bypass_blowers evaluate: %v", err)
	}
	dbDrive := dualEval.ArchitectureOptions.BlowerDrive
	if !dbDrive.Feasible || dbDrive.FanCostUSD < 380.0 || dbDrive.FanCostUSD > 460.0 || dbDrive.CostDeltaUSD > -1200.0 {
		t.Errorf("dual_bypass_blowers unexpected: %+v", dbDrive)
	}
	if !strings.Contains(strings.Join(dbDrive.Caveats, " "), "$15") || !strings.Contains(strings.Join(dbDrive.Caveats, " "), "45 firings") {
		t.Errorf("dual_bypass_blowers missing $15 brush swap / ~45 firings caveat: %+v", dbDrive.Caveats)
	}

	// 2d. compressor_eductor_only on 24 cu.ft. kiln -> feasible=false!
	compEdProj := baseProj
	compEdProj.BlowerDriveOption = BlowerDriveCompressorEductor
	compEdEval, err := eval.Evaluate(compEdProj)
	if err != nil {
		t.Fatalf("compressor_eductor_only evaluate: %v", err)
	}
	if compEdEval.ArchitectureOptions.BlowerDrive.Feasible {
		t.Errorf("compressor_eductor_only on 24 cu.ft. kiln must be infeasible: %+v", compEdEval.ArchitectureOptions.BlowerDrive)
	}

	// Large 48 cu.ft. co-op kiln flags infeasibility for single ring blower & dual bypass blowers.
	coop, _ := PresetByID("coop_wood_48")
	for _, opt := range []string{BlowerDriveRegenSideChannel, BlowerDriveDualBypassBlowers} {
		p := coop.Project
		p.BlowerDriveOption = opt
		ev, err := eval.Evaluate(p)
		if err != nil {
			t.Fatalf("coop %s evaluate: %v", opt, err)
		}
		if ev.ArchitectureOptions.BlowerDrive.Feasible {
			t.Errorf("48 cu.ft. co-op kiln should exceed single %s capacity", opt)
		}
	}

	// 3. Shell-Less Uniseal Parallel 2", 3", 4" Pipes.
	// 3a. 4" Sch 80 CPVC shell-less bundle genuinely SAVES money (negative cost_delta_usd, $400-$1,100)!
	par4Proj := baseProj
	par4Proj.PackedBedLayout = PackedBedLayoutParallelManifold
	par4Proj.ParallelPipeSizeIn = 4.0
	par4Eval, err := eval.Evaluate(par4Proj)
	if err != nil {
		t.Fatalf("parallel 4in evaluate: %v", err)
	}
	pb4 := par4Eval.ArchitectureOptions.ParallelBed
	if !pb4.Enabled || !pb4.ShelllessPlenum || !strings.Contains(pb4.PipeMaterial, "CPVC") {
		t.Errorf("4\" parallel bed missing shellless_plenum or CPVC pipe_material: %+v", pb4)
	}
	if pb4.CostDeltaUSD > -400.0 || pb4.CostDeltaUSD < -1100.0 {
		t.Errorf("4\" shell-less CPVC parallel bed CostDeltaUSD = %.2f, want in [-1100, -400]", pb4.CostDeltaUSD)
	}
	if par4Eval.COTS.EstimatedTotalUSD >= baseEval.COTS.EstimatedTotalUSD-400.0 {
		t.Errorf("4\" shell-less parallel BOM total ($%.2f) should save >=$400 vs baseline ($%.2f)",
			par4Eval.COTS.EstimatedTotalUSD, baseEval.COTS.EstimatedTotalUSD)
	}

	// 3b. 4" Sch 40 PVC when eductor_dry_fan cools mixed gas <= 140 F (even larger savings + thermal warning)!
	par4PVCProj := par4Proj
	par4PVCProj.StageAlignment = StageAlignmentEductorDryFan
	par4PVCEval, err := eval.Evaluate(par4PVCProj)
	if err != nil {
		t.Fatalf("eductor+parallel 4in PVC evaluate: %v", err)
	}
	pb4PVC := par4PVCEval.ArchitectureOptions.ParallelBed
	if !strings.Contains(pb4PVC.PipeMaterial, "PVC") || !strings.Contains(strings.Join(pb4PVC.Notes, " "), "140 F") {
		t.Errorf("eductor <=140 F parallel 4\" should select Sch 40 PVC with thermal warning: %+v", pb4PVC)
	}
	if pb4PVC.CostDeltaUSD >= pb4.CostDeltaUSD {
		t.Errorf("eductor 4\" PVC savings (%.2f) should beat inline 4\" CPVC savings (%.2f)", pb4PVC.CostDeltaUSD, pb4.CostDeltaUSD)
	}

	// 3c. 2" parallel pipes warns honestly about 40+ pipes & labor vs 4" sweet spot.
	par2Proj := baseProj
	par2Proj.PackedBedLayout = PackedBedLayoutParallelManifold
	par2Proj.ParallelPipeSizeIn = 2.0
	par2Eval, err := eval.Evaluate(par2Proj)
	if err != nil {
		t.Fatalf("parallel 2in evaluate: %v", err)
	}
	pb2 := par2Eval.ArchitectureOptions.ParallelBed
	if pb2.PipeNominalIn != 2.0 || pb2.PipeCount < 40 || pb2.WallRatioDToDp < 7.5 {
		t.Errorf("2\" parallel bed expected >=40 pipes & D/d_p>=7.5, got %+v", pb2)
	}
	if !strings.Contains(strings.Join(pb2.Notes, " "), "HIGH PIPE-COUNT") {
		t.Errorf("2\" parallel bed missing high pipe-count/labor warning: %+v", pb2.Notes)
	}

	// 3d. 3" parallel pipes (inline CPVC & eductor PVC).
	for _, align := range []string{StageAlignmentInlineIDBlower, StageAlignmentEductorDryFan} {
		par3Proj := baseProj
		par3Proj.StageAlignment = align
		par3Proj.PackedBedLayout = PackedBedLayoutParallelManifold
		par3Proj.ParallelPipeSizeIn = 3.0
		if ev, err := eval.Evaluate(par3Proj); err != nil || ev.ArchitectureOptions.ParallelBed.PipeNominalIn != 3.0 {
			t.Fatalf("parallel 3in (%s) failed: err=%v, pb=%+v", align, err, ev.ArchitectureOptions.ParallelBed)
		}
	}
	// Also test 2" PVC under eductor for full branch coverage.
	par2PVC := par2Proj
	par2PVC.StageAlignment = StageAlignmentEductorDryFan
	if ev, err := eval.Evaluate(par2PVC); err != nil || !strings.Contains(ev.ArchitectureOptions.ParallelBed.PipeMaterial, "PVC") {
		t.Fatalf("parallel 2in eductor PVC failed: err=%v, mat=%q", err, ev.ArchitectureOptions.ParallelBed.PipeMaterial)
	}

	// 4. All 3 PackingMediaOption modes.
	bioProj := baseProj
	bioProj.PackingMediaOption = PackingMediaAquacultureBioballs
	bioEval, err := eval.Evaluate(bioProj)
	if err != nil {
		t.Fatalf("aquaculture_bioballs evaluate: %v", err)
	}
	bioPack := bioEval.ArchitectureOptions.PackingMedia
	if bioPack.Voidage != 0.85 || bioPack.CostDeltaUSD > -150.0 || bioPack.CostDeltaUSD < -250.0 {
		t.Errorf("aquaculture_bioballs expected voidage=0.85 and -$150..-$250 savings, got %+v", bioPack)
	}

	diyProj := baseProj
	diyProj.PackingMediaOption = PackingMediaDIYCeramicRings
	diyEval, err := eval.Evaluate(diyProj)
	if err != nil {
		t.Fatalf("diy_ceramic_rings evaluate: %v", err)
	}
	diyPack := diyEval.ArchitectureOptions.PackingMedia
	if diyPack.Voidage != 0.70 || diyEval.Design.PackedBed.PercentOfFlood > 75.0 || diyEval.Design.PackedBed.ColumnDiameter <= baseEval.Design.PackedBed.ColumnDiameter || diyPack.CostDeltaUSD >= bioPack.CostDeltaUSD {
		t.Errorf("diy_ceramic_rings expected voidage=0.70, column upsized to stay <=75%% flood (%.1f%%), and lower media cost than bioballs: %+v", diyEval.Design.PackedBed.PercentOfFlood, diyPack)
	}
	if diyEval.COTS.PackingMedia.WeightKg <= baseEval.COTS.PackingMedia.WeightKg*5.0 {
		t.Errorf("diy_ceramic_rings must reflect heavy stoneware weight: %.1f kg/cuft vs %.1f kg/cuft",
			diyEval.COTS.PackingMedia.WeightKg, baseEval.COTS.PackingMedia.WeightKg)
	}
	if !strings.Contains(diyEval.BuildGuideMD, "Stoneware Rings Support Note") {
		t.Error("BuildGuideMD missing DIY stoneware support grate note")
	}

	// 5. Full Budget Studio Combo: shell-less 4" parallel + DIY ceramic rings + air-atomize + U-turn + electromechanical!
	budgetProj := baseProj
	budgetProj.PackedBedLayout = PackedBedLayoutParallelManifold
	budgetProj.ParallelPipeSizeIn = 4.0
	budgetProj.PackingMediaOption = PackingMediaDIYCeramicRings
	budgetProj.BlowerDriveOption = BlowerDriveCompressorAirAtomize
	budgetProj.CompactFootprint = true
	budgetProj.ControlTier = ControlTierElectromechanicalOnly
	budgetEval, err := eval.Evaluate(budgetProj)
	if err != nil {
		t.Fatalf("budget studio combo evaluate: %v", err)
	}
	netSavings := baseEval.COTS.EstimatedTotalUSD - budgetEval.COTS.EstimatedTotalUSD
	if netSavings < 2400.0 {
		t.Errorf("Budget Studio Combo net savings = $%.2f, want >= $2,400 (base=$%.2f, budget=$%.2f)",
			netSavings, baseEval.COTS.EstimatedTotalUSD, budgetEval.COTS.EstimatedTotalUSD)
	}
	if math.Abs(budgetEval.ArchitectureOptions.CostDeltaUSD-(-netSavings)) > 0.02 {
		t.Errorf("Budget CostDeltaUSD (%.2f) mismatch with exact BOM delta (%.2f)",
			budgetEval.ArchitectureOptions.CostDeltaUSD, -netSavings)
	}
	// Verify ALL mandatory NEC 430 / IPC 608 / IPC 702.5 / NFPA 86 safety items remain intact!
	hasRPZ, hasDisc, hasNFPA86, hasIPC702 := false, false, false, false
	for _, it := range budgetEval.COTS.BOM.Items {
		if strings.Contains(it.Spec, "IPC608-RPZ-AG") {
			hasRPZ = true
		}
		if strings.Contains(it.Spec, "NEC430-DISC-LR") {
			hasDisc = true
		}
		if strings.Contains(it.Spec, "200 F") && strings.Contains(it.Spec, "NFPA 86") {
			hasNFPA86 = true
		}
		if strings.Contains(it.Spec, "IPC 702.5") {
			hasIPC702 = true
		}
	}
	if !hasRPZ || !hasDisc || !hasNFPA86 || !hasIPC702 {
		t.Errorf("Budget combo dropped mandatory code safety items: RPZ=%v NEC430=%v NFPA86=%v IPC702=%v",
			hasRPZ, hasDisc, hasNFPA86, hasIPC702)
	}

	// Verify zero banned strings across Budget Studio Combo output & guides!
	bannedPhrases := []string{"Hobby " + "316L", "4-" + "lb", "gro" + "cery", "Switching from Salt " + "(NaCl) to Soda"}
	budgetJSON, _ := json.Marshal(budgetEval)
	haystack := string(budgetJSON) + "\n" + budgetEval.BuildGuideMD + "\n" + budgetEval.OperationsChecklistMD + "\n" + string(buildPacketReadme(budgetEval))
	for _, banned := range bannedPhrases {
		if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
			t.Errorf("Budget Studio Combo output contains banned phrase %q", banned)
		}
	}

	// 6. Validate Phase 8 invalid inputs & defaults.
	for _, bad := range []string{
		`{"blower_drive_option": "jet_turbine"}`,
		`{"parallel_pipe_size_in": 5}`,
		`{"packing_media_option": "gravel"}`,
		`{"shop_compressor_scfm": -10}`,
		`{"shop_compressor_scfm": 500}`,
	} {
		if _, err := ParseProject([]byte(bad)); err == nil {
			t.Errorf("expected validation error for %s", bad)
		}
	}
	// Parallel + bioballs & heavy air-atomize + small 120V air-atomize branch coverage.
	parBio := par4Proj
	parBio.PackingMediaOption = PackingMediaAquacultureBioballs
	if _, err := eval.Evaluate(parBio); err != nil {
		t.Fatalf("parallel + bioballs evaluate: %v", err)
	}
	heavyAtom := airAtomProj
	heavyAtom.CorrosionTier = CorrosionTierHeavyC276FRP
	heavyAtom.Input.KilnVolumeCuFt = 48.0
	if _, err := eval.Evaluate(heavyAtom); err != nil {
		t.Fatalf("heavy 48 cuft air-atomize evaluate: %v", err)
	}
	smallRegen := tinyKiln.Project
	smallRegen.BlowerDriveOption = BlowerDriveRegenSideChannel
	if _, err := eval.Evaluate(smallRegen); err != nil {
		t.Fatalf("small 6 cuft regen evaluate: %v", err)
	}
	smallDual := tinyKiln.Project
	smallDual.BlowerDriveOption = BlowerDriveDualBypassBlowers
	if _, err := eval.Evaluate(smallDual); err != nil {
		t.Fatalf("small 6 cuft dual bypass evaluate: %v", err)
	}
	smallAtom := tinyKiln.Project
	smallAtom.BlowerDriveOption = BlowerDriveCompressorAirAtomize
	if _, err := eval.Evaluate(smallAtom); err != nil {
		t.Fatalf("small 6 cuft air-atomize evaluate: %v", err)
	}

	t.Logf("Phase 8 24 cu.ft. Cost-Savings Table:\n"+
		"  Baseline (industrial_radial_vfd, 14\" CPVC, PP Pall, RPi):   $%.2f (venturi dP=%.2f, total=%.2f inWC, d50=%.3f um)\n"+
		"  Compressor Air-Atomize (25 um, 48 m/s, $75 nozzle+MP fan):  $%.2f (delta=%+.2f | venturi dP=%.2f, total=%.2f inWC, active d50=%.2f um, passive d50=%.3f um)\n"+
		"  Dual 3-Stage Tangential Bypass (2x Ametek Lamb):            $%.2f (delta=%+.2f)\n"+
		"  Side-Channel Ring Blower (wet inline / eductor dry):        $%.2f (delta=%+.2f) / $%.2f\n"+
		"  Compressor Main Eductor (15 SCFM studio compressor):        feasible=%v (needs %.1f SCFM / %.1f HP)\n"+
		"  Shell-Less 4\" CPVC Uniseal Bundle (%d pipes):               $%.2f (delta=%+.2f | bed dP=%.2f inWC)\n"+
		"  Shell-Less 4\" PVC Uniseal + Eductor (%d pipes, %.1f F):     $%.2f (parallel delta=%+.2f)\n"+
		"  Shell-Less 2\" CPVC Uniseal (%d pipes, high labor):          $%.2f (parallel delta=%+.2f)\n"+
		"  Aquaculture Bio-Balls ($38/cu.ft., eps=0.85):               $%.2f (media delta=%+.2f | bed dP=%.2f inWC)\n"+
		"  DIY Stoneware Ceramic Rings ($18/cu.ft., eps=0.70):         $%.2f (media delta=%+.2f | bed dP=%.2f inWC, wt=%.0f kg/cuft)\n"+
		"  BUDGET STUDIO COMBO (4\" Uniseal+DIY Rings+AirAtom+UTurn+EM): $%.2f (TOTAL SAVED: -$%.2f!)",
		baseEval.COTS.EstimatedTotalUSD, baseEval.Design.Venturi.PressureDropInWC, baseEval.Design.Blower.TotalStaticPressureInWC, baseEval.Design.Venturi.CutDiameterD50Microns,
		airAtomEval.COTS.EstimatedTotalUSD, airAtomEval.ArchitectureOptions.CostDeltaUSD, airAtomEval.Design.Venturi.PressureDropInWC, airAtomEval.Design.Blower.TotalStaticPressureInWC, activeD50, airAtomEval.Design.Venturi.CutDiameterD50Microns,
		dualEval.COTS.EstimatedTotalUSD, dualEval.ArchitectureOptions.CostDeltaUSD,
		regenWetEval.COTS.EstimatedTotalUSD, regenWetEval.ArchitectureOptions.CostDeltaUSD, regenDryEval.COTS.EstimatedTotalUSD,
		comp.EductorFeasible, comp.EductorSCFMNeeded, comp.EductorHPNeeded,
		pb4.PipeCount, par4Eval.COTS.EstimatedTotalUSD, pb4.CostDeltaUSD, pb4.BedDPInWC,
		pb4PVC.PipeCount, par4PVCEval.ArchitectureOptions.Eductor.MixedTempF, par4PVCEval.COTS.EstimatedTotalUSD, pb4PVC.CostDeltaUSD,
		pb2.PipeCount, par2Eval.COTS.EstimatedTotalUSD, pb2.CostDeltaUSD,
		bioEval.COTS.EstimatedTotalUSD, bioPack.CostDeltaUSD, bioPack.BedDPInWC,
		diyEval.COTS.EstimatedTotalUSD, diyPack.CostDeltaUSD, diyPack.BedDPInWC, diyEval.COTS.PackingMedia.WeightKg,
		budgetEval.COTS.EstimatedTotalUSD, netSavings,
	)
}

func TestPhase9CustomFabDeluxeAndCompleteSystemsShopPlan(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())

	// 1. ParseProject & Validate accept "custom_fab_deluxe".
	dlxParsed, err := ParseProject([]byte(`{"preset_id":"wood_salt_24","build_path":"custom_fab_deluxe"}`))
	if err != nil {
		t.Fatalf("ParseProject custom_fab_deluxe failed: %v", err)
	}
	if dlxParsed.BuildPath != BuildPathCustomFabDeluxe || !dlxParsed.IsDeluxe() {
		t.Fatalf("expected BuildPathCustomFabDeluxe, got %q", dlxParsed.BuildPath)
	}

	cotsProj := DefaultProject()
	cotsProj.BuildPath = BuildPathCOTSNoWeld
	cotsProj.IncludeSilencer = true
	cotsEval, err := eval.Evaluate(cotsProj)
	if err != nil {
		t.Fatalf("cots evaluate: %v", err)
	}

	dlxEval, err := eval.Evaluate(dlxParsed)
	if err != nil {
		t.Fatalf("deluxe evaluate: %v", err)
	}

	// 2. Verify Deluxe BOM upgrades, weight/cost invariants, and mandatory code compliance items.
	if dlxEval.COTS.EstimatedTotalUSD <= cotsEval.COTS.EstimatedTotalUSD+3000.0 {
		t.Errorf("expected Deluxe BOM ($%.2f) to exceed COTS ($%.2f) by >$3,000", dlxEval.COTS.EstimatedTotalUSD, cotsEval.COTS.EstimatedTotalUSD)
	}
	var sumCost, sumWeight float64
	var hasSightCleanout, hasSteelSkid, hasUL508A, hasANSI150, hasRPZ, hasDisc, hasNFPA86, hasIPC702 bool
	for _, it := range dlxEval.COTS.BOM.Items {
		if it.WeightKg <= 0 {
			t.Errorf("Deluxe BOM item %q has non-positive WeightKg %v", it.Name, it.WeightKg)
		}
		sumCost += it.TotalCostUSD
		sumWeight += it.WeightKg
		blob := it.Name + " " + it.Material + " " + it.Spec
		if strings.Contains(blob, "DLX-SIGHT-CLNOUT-INSUL") && strings.Contains(blob, "borosilicate") {
			hasSightCleanout = true
		}
		if strings.Contains(blob, "DLX-SKID-KBRACE-VBOOT") && strings.Contains(blob, "K-bracing") {
			hasSteelSkid = true
		}
		if strings.Contains(blob, "DLX-UL508A-N4X-PANEL") && strings.Contains(blob, "1/16-DIN") {
			hasUL508A = true
		}
		if strings.Contains(blob, "B16.5 Class 150") && strings.Contains(blob, "PTFE envelope") {
			hasANSI150 = true
		}
		if strings.Contains(blob, "IPC608-RPZ-AG") {
			hasRPZ = true
		}
		if strings.Contains(blob, "NEC 430.102") {
			hasDisc = true
		}
		if strings.Contains(blob, "NEC430-DISC-LR") {
			t.Errorf("Deluxe BOM must not double-count separate NEC430-DISC-LR when DLX-UL508A-N4X-PANEL includes through-door disconnect")
		}
		if strings.Contains(blob, "200 F") && strings.Contains(blob, "NFPA 86") {
			hasNFPA86 = true
		}
		if strings.Contains(blob, "IPC 702.5") {
			hasIPC702 = true
		}
	}
	if math.Abs(sumCost-dlxEval.COTS.BOM.TotalCostUSD) > 0.01 || math.Abs(sumWeight-dlxEval.COTS.BOM.TotalWeightKg) > 0.01 {
		t.Errorf("Deluxe BOM totals mismatch: cost=%.2f/%.2f wt=%.2f/%.2f", sumCost, dlxEval.COTS.BOM.TotalCostUSD, sumWeight, dlxEval.COTS.BOM.TotalWeightKg)
	}
	if !hasSightCleanout || !hasSteelSkid || !hasUL508A || !hasANSI150 || !hasRPZ || !hasDisc || !hasNFPA86 || !hasIPC702 {
		t.Errorf("Deluxe BOM missing required items: sight=%v skid=%v ul508a=%v ansi150=%v rpz=%v disc=%v nfpa86=%v ipc702=%v",
			hasSightCleanout, hasSteelSkid, hasUL508A, hasANSI150, hasRPZ, hasDisc, hasNFPA86, hasIPC702)
	}
	if dlxEval.COTS.Silencer == nil || !strings.Contains(dlxEval.COTS.Silencer.SKU, "DLX-ACOUSTIC-ENC") {
		t.Errorf("expected Deluxe acoustic enclosure silencer, got %+v", dlxEval.COTS.Silencer)
	}
	if dlxEval.COTS.SumpTank.SKU != "DLX-SUMP-SLOPED-35" {
		t.Errorf("expected DLX-SUMP-SLOPED-35 for 24 cu.ft. Deluxe, got %q", dlxEval.COTS.SumpTank.SKU)
	}

	// 3. Verify Deluxe acoustic attenuation (~9 dBA extra over standard silenced, ~22 dBA over bare).
	extraDrop3Ft := cotsEval.Urban.Noise.SilencedDBA3Ft - dlxEval.Urban.Noise.SilencedDBA3Ft
	if extraDrop3Ft < 8.0 || extraDrop3Ft > 10.0 {
		t.Errorf("expected Deluxe extra 3ft attenuation in [8, 10] dBA, got %.1f dBA (COTS=%.1f, Deluxe=%.1f)",
			extraDrop3Ft, cotsEval.Urban.Noise.SilencedDBA3Ft, dlxEval.Urban.Noise.SilencedDBA3Ft)
	}
	extraDropLine := cotsEval.Urban.Noise.PropertyLineDBA - dlxEval.Urban.Noise.PropertyLineDBA
	if extraDropLine < 8.0 || extraDropLine > 10.0 {
		t.Errorf("expected Deluxe extra property-line attenuation in [8, 10] dBA, got %.1f dBA (COTS=%.1f, Deluxe=%.1f)",
			extraDropLine, cotsEval.Urban.Noise.PropertyLineDBA, dlxEval.Urban.Noise.PropertyLineDBA)
	}
	if dlxEval.ShopPlan.AcousticSummary.TotalReductionDBA != 22.0 {
		t.Errorf("expected Deluxe TotalReductionDBA=22.0, got %.1f", dlxEval.ShopPlan.AcousticSummary.TotalReductionDBA)
	}
	if cotsEval.ShopPlan.AcousticSummary.TotalReductionDBA != 13.0 {
		t.Errorf("expected COTS silenced TotalReductionDBA=13.0, got %.1f", cotsEval.ShopPlan.AcousticSummary.TotalReductionDBA)
	}

	// 4. Verify Deluxe Build Guide section (TIG weld prep, ANSI B16.5 bolt torque, FRP curing, skid anchoring, acoustic baffle, UL 508A & NEC 110.26).
	for _, want := range []string{
		"Custom Fabrication Deluxe (`custom_fab_deluxe`)",
		"## 7. Custom Fabrication Deluxe Build Guide",
		"ER316L",
		"B16.5 Class 150",
		"3-stage star pattern",
		"Derakane Vinyl-Ester FRP",
		"concrete wedge anchors",
		"Cooling Labyrinth",
		"UL 508A",
		"NEC 110.26",
	} {
		if !strings.Contains(dlxEval.BuildGuideMD, want) {
			t.Errorf("Deluxe BuildGuideMD missing %q", want)
		}
	}

	// 5. Verify exact Phase 9 ShopPlan JSON contract (matching p9-web-deluxe-plan).
	spBytes, err := json.Marshal(dlxEval.ShopPlan)
	if err != nil {
		t.Fatalf("marshal ShopPlan: %v", err)
	}
	var spRaw map[string]any
	if err := json.Unmarshal(spBytes, &spRaw); err != nil {
		t.Fatalf("unmarshal ShopPlan raw: %v", err)
	}
	for _, key := range []string{
		"build_path", "is_deluxe", "clearances", "electrical", "plumbing",
		"pump_station", "bracing", "acoustics", "acoustic_summary", "deluxe_spec",
	} {
		if _, ok := spRaw[key]; !ok {
			t.Errorf("ShopPlan JSON missing top-level key %q", key)
		}
	}
	clrRaw := spRaw["clearances"].(map[string]any)
	for _, key := range []string{
		"flue_to_combustibles_in", "property_line_min_ft", "stack_above_roof_ft", "service_aisle_in",
		"nec_panel_depth_in", "nec_panel_width_in", "pump_service_in", "cleanout_swing_in",
		"hose_bend_radius_in", "wiring_access_note",
	} {
		if _, ok := clrRaw[key]; !ok {
			t.Errorf("ShopPlan.clearances JSON missing key %q", key)
		}
	}
	sp := dlxEval.ShopPlan
	if sp.BuildPath != BuildPathCustomFabDeluxe || !sp.IsDeluxe {
		t.Errorf("unexpected ShopPlan build_path=%q is_deluxe=%v", sp.BuildPath, sp.IsDeluxe)
	}
	if sp.Clearances.NECPanelDepthIn != 36.0 || sp.Clearances.NECPanelWidthIn != 30.0 ||
		sp.Clearances.PumpServiceIn != 24.0 || sp.Clearances.CleanoutSwingIn != 24.0 || sp.Clearances.HoseBendRadiusIn != 6.0 {
		t.Errorf("unexpected expanded clearances: %+v", sp.Clearances)
	}

	// Verify electrical E1..E5 (E3 is_safety_loop=true), plumbing P1..P7, pump_station, bracing B1..B5, acoustics S1..S3, deluxe_spec.
	if len(sp.Electrical) != 5 {
		t.Errorf("expected 5 electrical circuits (E1..E5) without diluter, got %d", len(sp.Electrical))
	}
	foundSafetyLoop := false
	for _, c := range sp.Electrical {
		if c.ID == "E3" {
			foundSafetyLoop = c.IsSafetyLoop && c.Voltage == 24
		}
		if c.ID == "E1" && (c.Voltage != dlxEval.Urban.Electrical.RequiredVoltage || c.Amps != dlxEval.Urban.Electrical.RequiredAmps) {
			t.Errorf("E1 voltage/amps (%dV %.1fA) mismatch urban.electrical (%dV %.1fA)", c.Voltage, c.Amps, dlxEval.Urban.Electrical.RequiredVoltage, dlxEval.Urban.Electrical.RequiredAmps)
		}
	}
	if !foundSafetyLoop {
		t.Errorf("missing 24VDC E3 hardwired safety loop in %+v", sp.Electrical)
	}
	if len(sp.Plumbing) != 7 {
		t.Errorf("expected 7 plumbing lines (P1..P7), got %d", len(sp.Plumbing))
	}
	if sp.PumpStation.FlowGPM <= 0 || sp.PumpStation.HeadFt < 40.0 || !strings.Contains(sp.PumpStation.BackflowPreventer, "ASSE 1013") {
		t.Errorf("unexpected Deluxe pump_station: %+v", sp.PumpStation)
	}
	if len(sp.Bracing) != 5 {
		t.Errorf("expected 5 bracing members (B1..B5), got %d", len(sp.Bracing))
	}
	if len(sp.Acoustics) != 3 {
		t.Errorf("expected 3 acoustic zones (S1..S3), got %d", len(sp.Acoustics))
	}
	var sumZoneDBA float64
	for _, z := range sp.Acoustics {
		sumZoneDBA += z.AttenuationDBA
	}
	if math.Abs(sumZoneDBA-sp.AcousticSummary.TotalReductionDBA) > 0.01 {
		t.Errorf("Deluxe acoustic zone sum %.1f != TotalReductionDBA %.1f", sumZoneDBA, sp.AcousticSummary.TotalReductionDBA)
	}
	if sp.DeluxeSpec == nil || !sp.DeluxeSpec.Enabled || len(sp.DeluxeSpec.Highlights) < 6 {
		t.Fatalf("unexpected DeluxeSpec: %+v", sp.DeluxeSpec)
	}
	exactDelta := round2(dlxEval.COTS.EstimatedTotalUSD - cotsEval.COTS.EstimatedTotalUSD)
	if math.Abs(sp.DeluxeSpec.CostDeltaUSD-exactDelta) > 0.02 {
		t.Errorf("DeluxeSpec.CostDeltaUSD=%.2f != exact delta %.2f", sp.DeluxeSpec.CostDeltaUSD, exactDelta)
	}
	if cotsEval.ShopPlan.DeluxeSpec == nil || cotsEval.ShopPlan.DeluxeSpec.Enabled || math.Abs(cotsEval.ShopPlan.DeluxeSpec.CostDeltaUSD-exactDelta) > 0.02 {
		t.Errorf("COTS ShopPlan.DeluxeSpec should preview disabled state with exact delta %.2f, got %+v", exactDelta, cotsEval.ShopPlan.DeluxeSpec)
	}

	// Verify expanded Deluxe cut_list includes ANSI B16.5 flanges, steel skid, K-bracing & acoustic baffle panels.
	if len(sp.CutList) <= len(cotsEval.ShopPlan.CutList)+3 {
		t.Errorf("expected expanded Deluxe cut_list (%d) vs COTS (%d)", len(sp.CutList), len(cotsEval.ShopPlan.CutList))
	}
	cutText := ""
	for _, ci := range sp.CutList {
		cutText += ci.Name + " " + ci.Notes + "\n"
	}
	for _, want := range []string{"ANSI/ASME B16.5 Class 150", "Forklift Skid", "K-Bracing", "Acoustic Blower Baffle Enclosure"} {
		if !strings.Contains(cutText, want) {
			t.Errorf("Deluxe CutList missing %q", want)
		}
	}

	// 6. Verify strict rejection of contradictory combos (Deluxe + ParallelManifold; Eductor + AirAtomize),
	// and verify valid combined features: COTS + Parallel Bed + Compressor Air-Atomize + DIY Ceramic Rings + Plume Diluter + Dual Sump!
	badDeluxeCombo := dlxParsed
	badDeluxeCombo.PackedBedLayout = PackedBedLayoutParallelManifold
	if _, err := eval.Evaluate(badDeluxeCombo); err == nil {
		t.Errorf("expected custom_fab_deluxe + parallel_manifold to be rejected by Validate()")
	}
	badEductorCombo := DefaultProject()
	badEductorCombo.StageAlignment = StageAlignmentEductorDryFan
	badEductorCombo.BlowerDriveOption = BlowerDriveCompressorAirAtomize
	if _, err := eval.Evaluate(badEductorCombo); err == nil {
		t.Errorf("expected eductor_dry_fan + compressor_air_atomize to be rejected by Validate()")
	}

	comboProj := DefaultProject()
	comboProj.PackedBedLayout = PackedBedLayoutParallelManifold
	comboProj.BlowerDriveOption = BlowerDriveCompressorAirAtomize
	comboProj.PackingMediaOption = PackingMediaDIYCeramicRings
	comboProj.IncludePlumeDiluter = true
	comboProj.SharedBaffledSump = false
	comboEval, err := eval.Evaluate(comboProj)
	if err != nil {
		t.Fatalf("combo evaluate: %v", err)
	}
	csp := comboEval.ShopPlan
	// E6 diluter electrical circuit present.
	hasE6 := false
	for _, c := range csp.Electrical {
		if c.ID == "E6" {
			hasE6 = true
		}
	}
	if !hasE6 {
		t.Error("expected E6 plume diluter circuit when IncludePlumeDiluter=true")
	}
	// A1 compressor air line, P5 parallel manifold, P6 barometric seal leg present.
	hasA1, hasP5Parallel, hasP6UTrap := false, false, false
	for _, pl := range csp.Plumbing {
		if pl.ID == "A1" && pl.PressurePSI == 40.0 {
			hasA1 = true
		}
		if pl.ID == "P5" && strings.Contains(pl.To, "Parallel-Pipe Spray Manifold") {
			hasP5Parallel = true
		}
		if pl.ID == "P6" && strings.Contains(pl.Notes, "barometric water-seal") {
			hasP6UTrap = true
		}
	}
	if !hasA1 || !hasP5Parallel || !hasP6UTrap {
		t.Errorf("combo plumbing interaction checks failed: A1=%v P5par=%v P6utrap=%v", hasA1, hasP5Parallel, hasP6UTrap)
	}
	// S4 diluter acoustic zone present.
	hasS4 := false
	for _, z := range csp.Acoustics {
		if z.ID == "S4" {
			hasS4 = true
		}
	}
	if !hasS4 {
		t.Error("expected S4 diluter acoustic zone when IncludePlumeDiluter=true")
	}
	// B4 adapts to eductor frame support (and P6 switches to >= 8" positive-pressure U-trap) in valid EductorDryFan config.
	edProj := DefaultProject()
	edProj.StageAlignment = StageAlignmentEductorDryFan
	edProj.PackedBedLayout = PackedBedLayoutParallelManifold
	edEval, err := eval.Evaluate(edProj)
	if err != nil {
		t.Fatalf("eductor evaluate: %v", err)
	}
	hasB4Eductor := false
	for _, b := range edEval.ShopPlan.Bracing {
		if b.ID == "B4" && strings.Contains(b.Component, "Eductor") {
			hasB4Eductor = true
		}
	}
	if !hasB4Eductor {
		t.Error("expected B4 bracing member to adapt to Downflow Eductor Frame Support")
	}
	// DIY ceramic rings increases B3 wet column load substantially vs PP Pall rings on identical single column!
	ceramicSingle := dlxParsed
	ceramicSingle.PackingMediaOption = PackingMediaDIYCeramicRings
	ceramicSingleEval, err := eval.Evaluate(ceramicSingle)
	if err != nil {
		t.Fatalf("ceramicSingle evaluate: %v", err)
	}
	var b3PP, b3Cer float64
	for _, b := range dlxEval.ShopPlan.Bracing {
		if b.ID == "B3" {
			b3PP = b.LoadLbs
		}
	}
	for _, b := range ceramicSingleEval.ShopPlan.Bracing {
		if b.ID == "B3" {
			b3Cer = b.LoadLbs
		}
	}
	if b3Cer <= b3PP+80.0 {
		t.Errorf("expected DIY stoneware ceramic rings B3 wet load (%.1f lbs) to exceed PP Pall rings (%.1f lbs) by >80 lbs", b3Cer, b3PP)
	}

	// 7. Verify Heavy C-276 Deluxe, small 35-gal Deluxe, large 110-gal Deluxe, unsnapped Deluxe, and welded_dxf preview branch coverage.
	heavyDlx, _ := PresetByID("coop_wood_48")
	heavyDlxProj := heavyDlx.Project
	heavyDlxProj.BuildPath = BuildPathCustomFabDeluxe
	heavyDlxProj.IncludePlumeDiluter = true
	heavyDlxEval, err := eval.Evaluate(heavyDlxProj)
	if err != nil {
		t.Fatalf("heavy 48 cuft deluxe evaluate: %v", err)
	}
	if !strings.Contains(heavyDlxEval.COTS.QuenchVessel.SKU, "DLX-QC276") || !strings.Contains(heavyDlxEval.BuildGuideMD, "ERNiCrMo-4") {
		t.Errorf("expected Alloy C-276 Deluxe quench & ERNiCrMo-4 guide, got SKU=%q", heavyDlxEval.COTS.QuenchVessel.SKU)
	}

	raku, _ := PresetByID("electric_raku_6")
	smallDlx := raku.Project
	smallDlx.BuildPath = BuildPathCustomFabDeluxe
	smallDlx.ControlTier = ControlTierElectromechanicalOnly
	smallDlx.IncludeSilencer = false
	smallDlxEval, err := eval.Evaluate(smallDlx)
	if err != nil {
		t.Fatalf("small 6 cuft deluxe evaluate: %v", err)
	}
	if smallDlxEval.COTS.SumpTank.SKU != "DLX-SUMP-SLOPED-35" || smallDlxEval.ShopPlan.AcousticSummary.TotalReductionDBA != 22.0 {
		t.Errorf("unexpected small Deluxe sump/acoustics: sump=%q red=%.1f", smallDlxEval.COTS.SumpTank.SKU, smallDlxEval.ShopPlan.AcousticSummary.TotalReductionDBA)
	}

	unsnapDlx := dlxParsed
	unsnapDlx.SnapToCOTS = false
	unsnapDlx.StageAlignment = StageAlignmentEductorDryFan
	if _, err := eval.Evaluate(unsnapDlx); err != nil {
		t.Fatalf("unsnapped Deluxe eductor evaluate: %v", err)
	}

	// Cover medium 60-gal & large 110-gal Deluxe sump + ansi150FlangeNote + breakerAndGaugeForAmps branches.
	medD := dlxEval.Design
	medD.PackedBed.SumpVolume = model.CubicMeters(30.0 * model.CubicMetersPerUSGallon)
	if s60 := selectSumpTankForProject(dlxParsed, medD); s60.SKU != "DLX-SUMP-SLOPED-60" {
		t.Errorf("expected DLX-SUMP-SLOPED-60 for 30 gal holdup, got %q", s60.SKU)
	}
	bigD := dlxEval.Design
	bigD.PackedBed.SumpVolume = model.CubicMeters(60.0 * model.CubicMetersPerUSGallon)
	if s110 := selectSumpTankForProject(dlxParsed, bigD); s110.SKU != "DLX-SUMP-SLOPED-110" {
		t.Errorf("expected DLX-SUMP-SLOPED-110 for 60 gal holdup, got %q", s110.SKU)
	}
	for _, nom := range []float64{6, 8, 10, 12, 14, 16, 18, 24} {
		if note := ansi150FlangeNote(nom); !strings.Contains(note, "BCD") {
			t.Errorf("ansi150FlangeNote(%v) missing BCD: %q", nom, note)
		}
	}
	for _, a := range []float64{15, 25, 35, 45, 55} {
		if brk, g := breakerAndGaugeForAmps(a); brk <= 0 || g == "" {
			t.Errorf("breakerAndGaugeForAmps(%v) invalid: %d %q", a, brk, g)
		}
	}
	// Cover workingSumpVolumeGal for all 3 Deluxe SKUs.
	for _, sku := range []string{"DLX-SUMP-SLOPED-35", "DLX-SUMP-SLOPED-60", "DLX-SUMP-SLOPED-110"} {
		if gal := workingSumpVolumeGal(COTSSummary{SumpTank: COTSPart{SKU: sku}}, dlxEval.Design); gal <= 0 {
			t.Errorf("workingSumpVolumeGal(%q) = %v", sku, gal)
		}
	}

	// 8. Verify BuildPacketZip contains enriched shop_plan.json with all Phase 9 systems layers!
	zipBytes, _, err := eval.BuildPacketZip(dlxParsed)
	if err != nil {
		t.Fatalf("Deluxe BuildPacketZip failed: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	foundShopPlan := false
	for _, f := range zr.File {
		if f.Name == "shop_plan.json" {
			foundShopPlan = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open shop_plan.json: %v", err)
			}
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(rc)
			_ = rc.Close()
			var zPlan ShopPlan
			if err := json.Unmarshal(buf.Bytes(), &zPlan); err != nil {
				t.Fatalf("unmarshal packet shop_plan.json: %v", err)
			}
			if !zPlan.IsDeluxe || len(zPlan.Electrical) < 5 || len(zPlan.Plumbing) < 7 || len(zPlan.Bracing) < 5 || len(zPlan.Acoustics) < 3 || zPlan.DeluxeSpec == nil {
				t.Errorf("packet shop_plan.json missing Phase 9 systems data: %+v", zPlan)
			}
		}
	}
	if !foundShopPlan {
		t.Fatal("Deluxe packet missing shop_plan.json")
	}

	// 9. Verify zero banned strings across Deluxe evaluation, guides, and packet README.
	bannedPhrases := []string{"Hobby " + "316L", "4-" + "lb", "gro" + "cery", "Switching from Salt " + "(NaCl) to Soda"}
	dlxJSON, _ := json.Marshal(dlxEval)
	haystack := string(dlxJSON) + "\n" + dlxEval.BuildGuideMD + "\n" + dlxEval.OperationsChecklistMD + "\n" + string(buildPacketReadme(dlxEval))
	for _, banned := range bannedPhrases {
		if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
			t.Errorf("Deluxe output contains banned phrase %q", banned)
		}
	}

	t.Logf("Phase 9 24 cu.ft. Deluxe vs COTS Comparison:\n"+
		"  No-Weld COTS (silenced):    $%.2f | bare=%.1f dBA @ 3ft -> silenced=%.1f dBA @ 3ft -> %.1f dBA @ %.0fft (-%.1f dBA)\n"+
		"  Custom Fab Deluxe:          $%.2f (delta=+$%.2f) | bare=%.1f dBA @ 3ft -> treated=%.1f dBA @ 3ft -> %.1f dBA @ %.0fft (-%.1f dBA)\n"+
		"  B3 Column Wet Load:         PP Pall=%.1f lbs | DIY Ceramic Rings=%.1f lbs",
		cotsEval.COTS.EstimatedTotalUSD, cotsEval.Urban.Noise.BareDBA3Ft, cotsEval.Urban.Noise.SilencedDBA3Ft, cotsEval.Urban.Noise.PropertyLineDBA, cotsEval.Urban.Noise.PropertyLineFt, cotsEval.ShopPlan.AcousticSummary.TotalReductionDBA,
		dlxEval.COTS.EstimatedTotalUSD, sp.DeluxeSpec.CostDeltaUSD, dlxEval.Urban.Noise.BareDBA3Ft, dlxEval.Urban.Noise.SilencedDBA3Ft, dlxEval.Urban.Noise.PropertyLineDBA, dlxEval.Urban.Noise.PropertyLineFt, sp.AcousticSummary.TotalReductionDBA,
		b3PP, b3Cer,
	)
}

func TestPhase10LocalCeramicNairobiMumbaiAndZeroElectricity(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())

	// 1. Standard powered build always populates Frugal preview with enabled=false.
	stdEval, err := eval.Evaluate(DefaultProject())
	if err != nil {
		t.Fatalf("default evaluate: %v", err)
	}
	if stdEval.Frugal == nil || stdEval.Frugal.Enabled || stdEval.Frugal.ZeroElectricity {
		t.Fatalf("expected non-nil Frugal preview with enabled=false on standard build, got %+v", stdEval.Frugal)
	}

	// 2. Verify all 3 regions (Nairobi, Mumbai, General Local) in 100% Zero-Electricity mode.
	bannedPhrases := []string{"Hobby " + "316L", "4-" + "lb", "gro" + "cery", "Switching from Salt " + "(NaCl) to Soda"}

	regionChecks := []struct {
		region   string
		currency string
	}{
		{LocalRegionNairobi, "KES"},
		{LocalRegionMumbai, "INR"},
		{LocalRegionGeneral, "USD"},
	}

	var nairobiEval, mumbaiEval StudioEvaluation
	for _, rc := range regionChecks {
		proj := OffGridCeramicProject(rc.region)
		ev, err := eval.Evaluate(proj)
		if err != nil {
			t.Fatalf("OffGridCeramicProject(%s) evaluate failed: %v", rc.region, err)
		}
		if rc.region == LocalRegionNairobi {
			nairobiEval = ev
		} else if rc.region == LocalRegionMumbai {
			mumbaiEval = ev
		}

		fr := ev.Frugal
		if fr == nil || !fr.Enabled || !fr.ZeroElectricity {
			t.Fatalf("%s: expected enabled zero-electricity FrugalAssessment, got %+v", rc.region, fr)
		}
		if fr.Region != rc.region || !strings.Contains(fr.LocalCurrencyEstimate, rc.currency) {
			t.Errorf("%s: unexpected region/currency: region=%q est=%q", rc.region, fr.Region, fr.LocalCurrencyEstimate)
		}
		if fr.PowerWatts != 0 {
			t.Errorf("%s: zero_electricity power_watts = %v, want 0", rc.region, fr.PowerWatts)
		}
		if !fr.DraftFeasible || fr.AvailableDraftInWC <= fr.RequiredDraftInWC || fr.DraftMarginInWC < 0.039 {
			t.Errorf("%s: passive draft must be feasible with >=0.04 inWC margin: avail=%.3f req=%.3f margin=%.3f",
				rc.region, fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftMarginInWC)
		}
		if fr.ChimneyHeightM < 8.0 || fr.ChimneyHeightM > 12.0 {
			t.Errorf("%s: ChimneyHeightM = %.1f out of [8.0, 12.0]", rc.region, fr.ChimneyHeightM)
		}
		if math.Abs(fr.ChimneyHeightFt-round1(fr.ChimneyHeightM*3.28084)) > 0.11 {
			t.Errorf("%s: ChimneyHeightFt mismatch: %.1f vs %.1f", rc.region, fr.ChimneyHeightFt, fr.ChimneyHeightM*3.28084)
		}
		if fr.BedSuperficialVelMS < 0.30 || fr.BedSuperficialVelMS > 0.46 {
			t.Errorf("%s: BedSuperficialVelMS = %.2f out of [0.30, 0.46]", rc.region, fr.BedSuperficialVelMS)
		}
		if fr.BedDPInWC < 0.03 || fr.BedDPInWC > 0.10 {
			t.Errorf("%s: BedDPInWC = %.3f out of [0.03, 0.10]", rc.region, fr.BedDPInWC)
		}
		if fr.GravityTankGal < 55.0 || fr.GravityTankGal > 130.0 || fr.GravityTankLiters < 200.0 || fr.GravityTankLiters > 500.0 {
			t.Errorf("%s: gravity tank out of bounds: %.1f gal / %.0f L", rc.region, fr.GravityTankGal, fr.GravityTankLiters)
		}
		if math.Abs(fr.GravityHeadPSI-round2(fr.TankElevationFt*0.4335)) > 0.02 {
			t.Errorf("%s: GravityHeadPSI mismatch: %.2f vs %.2f", rc.region, fr.GravityHeadPSI, fr.TankElevationFt*0.4335)
		}
		if fr.Jerrycans20LPerHr < 12 || fr.Jerrycans20LPerHr > 20 || fr.TreadleMinutesPerHr < 6.0 || fr.TreadleMinutesPerHr > 12.0 {
			t.Errorf("%s: unexpected manual pumping labor: %d jerrycans/hr, %.1f min/hr", rc.region, fr.Jerrycans20LPerHr, fr.TreadleMinutesPerHr)
		}
		if fr.HClRemovalPct < 95.0 || fr.HClRemovalPct > 98.5 {
			t.Errorf("%s: HClRemovalPct = %.1f%% out of [95.0, 98.5]", rc.region, fr.HClRemovalPct)
		}
		if fr.CoarseAshCapturePct < 88.0 || fr.CoarseAshCapturePct > 94.0 {
			t.Errorf("%s: CoarseAshCapturePct = %.1f%% out of [88.0, 94.0]", rc.region, fr.CoarseAshCapturePct)
		}
		if fr.SubmicronNaClCapturePct < 25.0 || fr.SubmicronNaClCapturePct >= 45.0 {
			t.Errorf("%s: SubmicronNaClCapturePct = %.1f%% must be honest (<45%%)", rc.region, fr.SubmicronNaClCapturePct)
		}
		if fr.EstimatedLocalCostUSD < 350.0 || fr.EstimatedLocalCostUSD > 1100.0 {
			t.Errorf("%s: EstimatedLocalCostUSD = $%.2f out of [$350, $1100]", rc.region, fr.EstimatedLocalCostUSD)
		}
		if math.Abs(fr.EstimatedLocalCostUSD-ev.COTS.EstimatedTotalUSD) > 0.02 {
			t.Errorf("%s: Frugal cost ($%.2f) != COTS BOM total ($%.2f)", rc.region, fr.EstimatedLocalCostUSD, ev.COTS.EstimatedTotalUSD)
		}
		// Verify BOM weight/cost invariants & positive weights.
		var sumCost, sumWeight float64
		for _, it := range ev.COTS.BOM.Items {
			if it.WeightKg <= 0 {
				t.Errorf("%s: BOM item %q has non-positive WeightKg %v", rc.region, it.Name, it.WeightKg)
			}
			sumCost += it.TotalCostUSD
			sumWeight += it.WeightKg
		}
		if math.Abs(sumCost-ev.COTS.BOM.TotalCostUSD) > 0.02 || math.Abs(sumWeight-ev.COTS.BOM.TotalWeightKg) > 0.02 {
			t.Errorf("%s: BOM totals mismatch: cost %.2f/%.2f wt %.2f/%.2f", rc.region, sumCost, ev.COTS.BOM.TotalCostUSD, sumWeight, ev.COTS.BOM.TotalWeightKg)
		}

		// Verify Urban electrical (0V / 0W) and whisper-quiet noise (~45 dBA bare).
		if ev.Urban.Electrical.Status != StatusOK || ev.Urban.Electrical.RequiredVoltage != 0 || ev.Urban.Electrical.RequiredAmps != 0 ||
			!strings.Contains(ev.Urban.Electrical.CircuitLabel, "0V / 0W") {
			t.Errorf("%s: unexpected zero-elec Urban.Electrical: %+v", rc.region, ev.Urban.Electrical)
		}
		if ev.Urban.Noise.Status != StatusOK || ev.Urban.Noise.BareDBA3Ft != 45.0 {
			t.Errorf("%s: unexpected zero-elec Urban.Noise: %+v", rc.region, ev.Urban.Noise)
		}

		// Verify ShopPlan: 0 electrical circuits ([] in JSON, not null), 0 NEC panel clearance, P1..P5, B1..B5.
		if len(ev.ShopPlan.Electrical) != 0 || ev.ShopPlan.Electrical == nil {
			t.Errorf("%s: expected empty non-nil ShopPlan.Electrical, got %#v", rc.region, ev.ShopPlan.Electrical)
		}
		if ev.ShopPlan.Clearances.NECPanelDepthIn != 0 || ev.ShopPlan.Clearances.NECPanelWidthIn != 0 {
			t.Errorf("%s: expected 0 NEC panel clearance when off-grid, got %+v", rc.region, ev.ShopPlan.Clearances)
		}
		if len(ev.ShopPlan.Plumbing) != 5 || len(ev.ShopPlan.Bracing) != 5 || len(ev.ShopPlan.Parts) != 9 || len(ev.ShopPlan.CutList) < 6 {
			t.Errorf("%s: unexpected ShopPlan counts: parts=%d cuts=%d plumb=%d brace=%d",
				rc.region, len(ev.ShopPlan.Parts), len(ev.ShopPlan.CutList), len(ev.ShopPlan.Plumbing), len(ev.ShopPlan.Bracing))
		}
		if ev.ShopPlan.PumpStation.HP != 0 || ev.ShopPlan.PumpStation.Voltage != 0 {
			t.Errorf("%s: expected 0 HP / 0 V manual pump station, got %+v", rc.region, ev.ShopPlan.PumpStation)
		}

		// Verify DIY steam boiler safety rejection & guide sections.
		for _, want := range []string{
			"STEAM BOILER",
			"## 7. Local Ceramic & Zero-Electricity Build + Manual Firing Guide",
			"Raschig Rings",
			"Gravity Tank Stand",
		} {
			if !strings.Contains(strings.ToUpper(ev.BuildGuideMD), strings.ToUpper(want)) {
				t.Errorf("%s: BuildGuideMD missing %q", rc.region, want)
			}
		}
		if !strings.Contains(ev.OperationsChecklistMD, "Zero-Steam-Boiler") {
			t.Errorf("%s: OperationsChecklistMD missing Zero-Steam-Boiler safety rule", rc.region)
		}

		// Verify zero banned strings across full JSON, guides, and README.
		rawJSON, err := json.Marshal(ev)
		if err != nil {
			t.Fatalf("%s: marshal eval: %v", rc.region, err)
		}
		haystack := string(rawJSON) + "\n" + ev.BuildGuideMD + "\n" + ev.OperationsChecklistMD + "\n" + string(buildPacketReadme(ev))
		for _, banned := range bannedPhrases {
			if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
				t.Errorf("%s: output contains banned phrase %q", rc.region, banned)
			}
		}
	}

	// Nairobi high altitude (83.5 kPa) requires a taller chimney than sea-level Mumbai (101.3 kPa).
	if nairobiEval.Frugal.ChimneyHeightM < mumbaiEval.Frugal.ChimneyHeightM {
		t.Errorf("expected Nairobi highland chimney (%.1f m) >= Mumbai sea-level chimney (%.1f m)",
			nairobiEval.Frugal.ChimneyHeightM, mumbaiEval.Frugal.ChimneyHeightM)
	}

	// 3. Verify all 4 local alkalis & exact molar stoichiometry on 24 cu.ft. wood/salt kiln (6.0 lbs = 2.72155 kg NaCl -> 46.568 mol HCl).
	molHCl := (2.0 * 3.0 * model.KgPerLb) / 0.0584428
	alkaliCases := []struct {
		id         string
		wantKg     float64
		wantZero   bool
		wantWarn   bool
		formulaSub string
	}{
		{LocalAlkaliSodaAsh, round2(molHCl * 0.5 * 0.1059888 * 1.25), false, false, "Na2CO3"},
		{LocalAlkaliSlakedLime, round2(molHCl * 0.5 * 0.074093 * 1.30), false, true, "Ca(OH)2"},
		{LocalAlkaliWoodAshLye, round1((molHCl * 0.5 * 0.138205 * 1.25) / 0.07), true, true, "K2CO3"},
		{LocalAlkaliBakingSoda, round2(molHCl * 1.0 * 0.0840066 * 1.25), false, false, "NaHCO3"},
	}
	for _, ac := range alkaliCases {
		proj := OffGridCeramicProject(LocalRegionMumbai)
		proj.LocalAlkali = ac.id
		ev, err := eval.Evaluate(proj)
		if err != nil {
			t.Fatalf("alkali %s evaluate: %v", ac.id, err)
		}
		fr := ev.Frugal
		if fr.AlkaliID != ac.id || !strings.Contains(fr.AlkaliFormula, ac.formulaSub) {
			t.Errorf("alkali %s mismatch: id=%q formula=%q", ac.id, fr.AlkaliID, fr.AlkaliFormula)
		}
		if math.Abs(fr.AlkaliKgPerFiring-ac.wantKg) > 0.11 {
			t.Errorf("alkali %s kg=%.2f, want %.2f", ac.id, fr.AlkaliKgPerFiring, ac.wantKg)
		}
		if ac.wantZero && fr.AlkaliCostUSD != 0 {
			t.Errorf("alkali %s expected $0 cost, got $%.2f", ac.id, fr.AlkaliCostUSD)
		}
		if !ac.wantZero && fr.AlkaliCostUSD <= 0 {
			t.Errorf("alkali %s expected positive cost, got $%.2f", ac.id, fr.AlkaliCostUSD)
		}
		if ac.wantWarn && fr.ScalingWarning == "" {
			t.Errorf("alkali %s expected non-empty scaling_warning", ac.id)
		}
		if !ac.wantWarn && fr.ScalingWarning != "" {
			t.Errorf("alkali %s expected empty scaling_warning, got %q", ac.id, fr.ScalingWarning)
		}
	}

	// Non-salt (soda) mode yields 0 kg local alkali across all 4 alkalis.
	for _, alk := range []string{LocalAlkaliSodaAsh, LocalAlkaliSlakedLime, LocalAlkaliWoodAshLye, LocalAlkaliBakingSoda} {
		proj := OffGridCeramicProject(LocalRegionNairobi)
		proj.VaporGlazeMode = VaporGlazeSoda
		proj.LocalAlkali = alk
		ev, err := eval.Evaluate(proj)
		if err != nil {
			t.Fatalf("soda mode alkali %s: %v", alk, err)
		}
		if ev.Frugal.AlkaliKgPerFiring != 0 || ev.Frugal.AlkaliCostUSD != 0 {
			t.Errorf("soda mode alkali %s should be 0 kg / $0, got %.2f kg / $%.2f", alk, ev.Frugal.AlkaliKgPerFiring, ev.Frugal.AlkaliCostUSD)
		}
	}

	// 4. Hybrid Local Ceramic mode (zero_electricity: false -> 230V 50Hz domestic pump + local fan + 4.5 m chimney).
	hybridProj := OffGridCeramicProject(LocalRegionMumbai)
	hybridProj.ZeroElectricity = false
	hybridProj.AvailableVoltage = 230
	hybridProj.IncludeSilencer = true
	hybridEval, err := eval.Evaluate(hybridProj)
	if err != nil {
		t.Fatalf("hybrid ceramic evaluate: %v", err)
	}
	hfr := hybridEval.Frugal
	if !hfr.Enabled || hfr.ZeroElectricity {
		t.Fatalf("expected enabled hybrid FrugalAssessment, got %+v", hfr)
	}
	if hfr.PowerWatts < 450.0 || hfr.PowerWatts > 650.0 {
		t.Errorf("hybrid PowerWatts = %.0f, want [450, 650]", hfr.PowerWatts)
	}
	if hfr.ChimneyHeightM != 4.5 || !strings.Contains(hfr.ManualPumpType, "230V 50Hz") || hfr.Jerrycans20LPerHr != 0 {
		t.Errorf("unexpected hybrid specs: chimney=%.1f m pump=%q jerrycans=%d", hfr.ChimneyHeightM, hfr.ManualPumpType, hfr.Jerrycans20LPerHr)
	}
	if len(hybridEval.ShopPlan.Electrical) != 2 || hybridEval.ShopPlan.PumpStation.HP != 0.5 || hybridEval.ShopPlan.PumpStation.Voltage != 230 {
		t.Errorf("unexpected hybrid ShopPlan electrical/pump: circuits=%d pump=%+v", len(hybridEval.ShopPlan.Electrical), hybridEval.ShopPlan.PumpStation)
	}

	// 120V hybrid & ZeroElectricity-only toggle on default COTS project (auto-activates ceramic passive mode!).
	hybrid120 := hybridProj
	hybrid120.AvailableVoltage = 120
	hybrid120.LocalRegion = LocalRegionGeneral
	if ev120, err := eval.Evaluate(hybrid120); err != nil || ev120.Urban.Electrical.RequiredVoltage != 120 {
		t.Fatalf("120V hybrid evaluate failed: err=%v elec=%+v", err, ev120.Urban.Electrical)
	}

	// Contradictory ZeroElectricity=true on non-local BuildPath must be rejected by Validate().
	badZeroProj := DefaultProject()
	badZeroProj.ZeroElectricity = true
	badZeroProj.AvailableVoltage = 0
	if _, err := eval.Evaluate(badZeroProj); err == nil {
		t.Fatalf("expected ZeroElectricity=true with build_path=cots_no_weld to fail Validate()")
	}
	okZeroProj := badZeroProj
	okZeroProj.BuildPath = BuildPathLocalCeramic
	autoZeroEval, err := eval.Evaluate(okZeroProj)
	if err != nil {
		t.Fatalf("ZeroElectricity=true local ceramic evaluate failed: %v", err)
	}
	if !autoZeroEval.Frugal.Enabled || !autoZeroEval.Frugal.ZeroElectricity || len(autoZeroEval.ShopPlan.Electrical) != 0 {
		t.Errorf("ZeroElectricity=true local ceramic should have 0 circuits: %+v", autoZeroEval.Frugal)
	}

	// 5. Small (6 cu.ft.) and Large (48 cu.ft.) ceramic kilns + passive draft shortfall edge case.
	smallCer := OffGridCeramicProject(LocalRegionNairobi)
	smallCer.Input.KilnVolumeCuFt = 6.0
	if evSmall, err := eval.Evaluate(smallCer); err != nil || evSmall.Frugal.GravityTankLiters != 210.0 {
		t.Fatalf("small ceramic evaluate failed: err=%v tank=%v", err, evSmall.Frugal.GravityTankLiters)
	}
	bigCer := OffGridCeramicProject(LocalRegionMumbai)
	bigCer.Input.KilnVolumeCuFt = 48.0
	if evBig, err := eval.Evaluate(bigCer); err != nil || evBig.Frugal.GravityTankLiters != 490.0 {
		t.Fatalf("big ceramic evaluate failed: err=%v tank=%v", err, evBig.Frugal.GravityTankLiters)
	}
	shortfallProj := OffGridCeramicProject(LocalRegionNairobi)
	shortfallProj.Input.TargetKilnDraftInWC = -0.35
	shortfallEval, err := eval.Evaluate(shortfallProj)
	if err != nil {
		t.Fatalf("shortfall evaluate: %v", err)
	}
	if shortfallEval.Frugal.DraftFeasible {
		t.Errorf("expected DraftFeasible=false for extreme -0.35 inWC target draft, got avail=%.3f req=%.3f",
			shortfallEval.Frugal.AvailableDraftInWC, shortfallEval.Frugal.RequiredDraftInWC)
	}

	// 6. Verify exact JSON contract keys matching p10-web-frugal & BuildPacketZip round-trip.
	evalBytes, err := json.Marshal(nairobiEval)
	if err != nil {
		t.Fatalf("marshal nairobiEval: %v", err)
	}
	var topRaw map[string]any
	if err := json.Unmarshal(evalBytes, &topRaw); err != nil {
		t.Fatalf("unmarshal topRaw: %v", err)
	}
	frugalRaw, ok := topRaw["frugal"].(map[string]any)
	if !ok {
		t.Fatalf("StudioEvaluation JSON missing 'frugal' object")
	}
	for _, key := range []string{
		"enabled", "zero_electricity", "region", "region_label",
		"chimney_height_ft", "chimney_height_m", "stack_reheat_temp_f",
		"available_draft_inwc", "required_draft_inwc", "draft_margin_inwc", "draft_feasible", "draft_mechanism",
		"bed_diameter_in", "bed_superficial_vel_m_s", "bed_dp_inwc", "packing_material",
		"gravity_tank_gal", "gravity_tank_liters", "tank_elevation_ft", "gravity_head_psi",
		"spray_flow_gpm", "spray_flow_lpm", "manual_pump_type", "jerrycans_20l_per_hr", "treadle_minutes_per_hr",
		"alkali_id", "alkali_name", "alkali_formula", "alkali_kg_per_firing", "alkali_lbs_per_firing",
		"alkali_cost_usd", "alkali_local_note", "scaling_warning",
		"hcl_removal_pct", "coarse_ash_capture_pct", "submicron_nacl_capture_pct", "submicron_honest_note", "thermal_immunity_note",
		"estimated_local_cost_usd", "local_currency_estimate", "power_watts",
		"manual_steps", "local_materials", "warnings", "summary",
	} {
		if _, exists := frugalRaw[key]; !exists {
			t.Errorf("FrugalAssessment JSON missing required contract key %q", key)
		}
	}
	spRaw := topRaw["shop_plan"].(map[string]any)
	if elecArr, ok := spRaw["electrical"].([]any); !ok || len(elecArr) != 0 {
		t.Errorf("expected off-grid shop_plan.electrical to marshal as empty JSON array [], got %#v", spRaw["electrical"])
	}

	if zipBytes, _, err := eval.BuildPacketZip(OffGridCeramicProject(LocalRegionNairobi)); err != nil || len(zipBytes) == 0 {
		t.Fatalf("Off-Grid BuildPacketZip failed: err=%v len=%d", err, len(zipBytes))
	}

	// 7. Validation checks for invalid Phase 10 fields and ParseProject round-trip.
	if parsed, err := ParseProject([]byte(`{"preset_id":"wood_salt_24","build_path":"local_ceramic_masonry","zero_electricity":true,"local_region":"mumbai","local_alkali":"slaked_lime"}`)); err != nil {
		t.Fatalf("ParseProject local_ceramic_masonry failed: %v", err)
	} else if parsed.LocalRegion != LocalRegionMumbai || parsed.LocalAlkali != LocalAlkaliSlakedLime || !parsed.ZeroElectricity {
		t.Errorf("unexpected parsed Phase 10 project: %+v", parsed)
	}
	for _, bad := range []string{
		`{"local_region": "mars"}`,
		`{"local_alkali": "bleach"}`,
		`{"available_voltage": 0, "zero_electricity": false}`,
	} {
		if _, err := ParseProject([]byte(bad)); err == nil {
			t.Errorf("expected validation error for %s", bad)
		}
	}
	// Cover formatIntCommas helper edge cases.
	if formatIntCommas(-12500) != "-12,500" || formatIntCommas(500) != "500" || formatIntCommas(105300) != "105,300" {
		t.Errorf("formatIntCommas unexpected output")
	}

	t.Logf("Phase 10 24 cu.ft. Zero-Electricity Ceramic Summary:\n"+
		"  Nairobi (83.5 kPa, Magadi Na2CO3):  %s | %0.f\" bed (v_s=%.2f m/s, dP=%.3f inWC) | %.1f m chimney (%.0f F) -> %.3f inWC avail vs %.3f req (+%.3f) | %.0f L tank (%.2f PSI, %.2f L/min -> %d jerrycans/hr, %.1f min/hr treadle) | %.2f kg Na2CO3 | HCl=%.1f%% Ash=%.1f%% Submicron=%.0f%%\n"+
		"  Mumbai  (101.3 kPa, Washing Soda):  %s | %0.f\" bed (v_s=%.2f m/s, dP=%.3f inWC) | %.1f m chimney (%.0f F) -> %.3f inWC avail vs %.3f req (+%.3f) | %.0f L tank (%.2f PSI, %.2f L/min -> %d jerrycans/hr, %.1f min/hr treadle) | %.2f kg Na2CO3 | HCl=%.1f%% Ash=%.1f%% Submicron=%.0f%%",
		nairobiEval.Frugal.LocalCurrencyEstimate, nairobiEval.Frugal.BedDiameterIn, nairobiEval.Frugal.BedSuperficialVelMS, nairobiEval.Frugal.BedDPInWC,
		nairobiEval.Frugal.ChimneyHeightM, nairobiEval.Frugal.StackReheatTempF, nairobiEval.Frugal.AvailableDraftInWC, nairobiEval.Frugal.RequiredDraftInWC, nairobiEval.Frugal.DraftMarginInWC,
		nairobiEval.Frugal.GravityTankLiters, nairobiEval.Frugal.GravityHeadPSI, nairobiEval.Frugal.SprayFlowLPM, nairobiEval.Frugal.Jerrycans20LPerHr, nairobiEval.Frugal.TreadleMinutesPerHr,
		nairobiEval.Frugal.AlkaliKgPerFiring, nairobiEval.Frugal.HClRemovalPct, nairobiEval.Frugal.CoarseAshCapturePct, nairobiEval.Frugal.SubmicronNaClCapturePct,
		mumbaiEval.Frugal.LocalCurrencyEstimate, mumbaiEval.Frugal.BedDiameterIn, mumbaiEval.Frugal.BedSuperficialVelMS, mumbaiEval.Frugal.BedDPInWC,
		mumbaiEval.Frugal.ChimneyHeightM, mumbaiEval.Frugal.StackReheatTempF, mumbaiEval.Frugal.AvailableDraftInWC, mumbaiEval.Frugal.RequiredDraftInWC, mumbaiEval.Frugal.DraftMarginInWC,
		mumbaiEval.Frugal.GravityTankLiters, mumbaiEval.Frugal.GravityHeadPSI, mumbaiEval.Frugal.SprayFlowLPM, mumbaiEval.Frugal.Jerrycans20LPerHr, mumbaiEval.Frugal.TreadleMinutesPerHr,
		mumbaiEval.Frugal.AlkaliKgPerFiring, mumbaiEval.Frugal.HClRemovalPct, mumbaiEval.Frugal.CoarseAshCapturePct, mumbaiEval.Frugal.SubmicronNaClCapturePct,
	)
}

func TestPhase10BZeroElectricitySteamJetEjectorAndCondensation(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())
	bannedPhrases := []string{"Hobby " + "316L", "4-" + "lb", "gro" + "cery", "Switching from Salt " + "(NaCl) to Soda"}

	// 1. Presets count stays 5, and default/empty local_draft_mode normalizes cleanly to natural_chimney.
	if len(Presets()) != 5 {
		t.Fatalf("expected 5 curated presets, got %d", len(Presets()))
	}
	passiveNB := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftNaturalChimney)
	if passiveNB.LocalDraftMode != LocalDraftNaturalChimney || passiveNB.IsSteamEjector() {
		t.Fatalf("unexpected natural_chimney helper: %+v", passiveNB)
	}
	defaultSteam := OffGridSteamCeramicProject(LocalRegionGeneral, "")
	if defaultSteam.LocalDraftMode != LocalDraftSteamEjectorLP || !defaultSteam.IsSteamEjector() {
		t.Fatalf("empty draftMode on OffGridSteamCeramicProject should default to steam_ejector_lp, got %q", defaultSteam.LocalDraftMode)
	}

	// 2. Evaluate natural_chimney, steam_ejector_lp, and steam_ejector_hp across Nairobi, Mumbai, and General Local.
	for _, region := range []string{LocalRegionNairobi, LocalRegionMumbai, LocalRegionGeneral} {
		evPassive, err := eval.Evaluate(OffGridCeramicProject(region))
		if err != nil {
			t.Fatalf("%s passive evaluate: %v", region, err)
		}
		evLP, err := eval.Evaluate(OffGridSteamCeramicProject(region, LocalDraftSteamEjectorLP))
		if err != nil {
			t.Fatalf("%s LP steam evaluate: %v", region, err)
		}
		evHP, err := eval.Evaluate(OffGridSteamCeramicProject(region, LocalDraftSteamEjectorHP))
		if err != nil {
			t.Fatalf("%s HP steam evaluate: %v", region, err)
		}

		fp, flp, fhp := evPassive.Frugal, evLP.Frugal, evHP.Frugal
		if fp.SteamEnabled || fp.DraftMode != LocalDraftNaturalChimney || fp.SteamNozzleVelMS != 0 || fp.EjectorDraftInWC != 0 {
			t.Errorf("%s passive should have steam_enabled=false & 0 steam metrics: %+v", region, fp)
		}
		if !flp.SteamEnabled || flp.DraftMode != LocalDraftSteamEjectorLP || flp.SteamPressurePSIG != 20.0 || math.Abs(flp.SteamPressureBar-1.38) > 0.02 {
			t.Errorf("%s LP steam metadata wrong: %+v", region, flp)
		}
		if !fhp.SteamEnabled || fhp.DraftMode != LocalDraftSteamEjectorHP || fhp.SteamPressurePSIG != 60.0 || math.Abs(fhp.SteamPressureBar-4.14) > 0.02 {
			t.Errorf("%s HP steam metadata wrong: %+v", region, fhp)
		}

		// Exact physics monotonicity: v_nozzle_hp > v_nozzle_lp > 0, Mach_hp > Mach_lp > 1,
		// ejector_draft_hp > ejector_draft_lp > 0, submicron_hp > submicron_lp > submicron_passive,
		// chimney_steam < chimney_passive.
		if !(fhp.SteamNozzleVelMS > flp.SteamNozzleVelMS && flp.SteamNozzleVelMS > 500.0 && flp.SteamNozzleVelMS < 560.0 && fhp.SteamNozzleVelMS > 695.0 && fhp.SteamNozzleVelMS < 785.0) {
			t.Errorf("%s nozzle velocity monotonicity/bounds failed: LP=%.1f HP=%.1f", region, flp.SteamNozzleVelMS, fhp.SteamNozzleVelMS)
		}
		if !(fhp.SteamNozzleMach > flp.SteamNozzleMach && flp.SteamNozzleMach >= 1.05 && flp.SteamNozzleMach <= 1.20 && fhp.SteamNozzleMach >= 1.48 && fhp.SteamNozzleMach <= 1.68) {
			t.Errorf("%s nozzle Mach bounds failed: LP=%.2f HP=%.2f", region, flp.SteamNozzleMach, fhp.SteamNozzleMach)
		}
		if flp.SteamNozzleThroatMM < 3.5 || flp.SteamNozzleThroatMM > 6.5 || fhp.SteamNozzleThroatMM < 3.5 || fhp.SteamNozzleThroatMM > 6.5 {
			t.Errorf("%s throat diameter out of physical bounds: LP=%.2f mm HP=%.2f mm", region, flp.SteamNozzleThroatMM, fhp.SteamNozzleThroatMM)
		}
		if flp.SteamFlowKgHr < 18.0 || flp.SteamFlowKgHr > 26.0 || fhp.SteamFlowKgHr < 48.0 || fhp.SteamFlowKgHr > 62.0 {
			t.Errorf("%s steam flow out of 24 cu.ft. bounds: LP=%.1f HP=%.1f kg/hr", region, flp.SteamFlowKgHr, fhp.SteamFlowKgHr)
		}
		if flp.SteamWaterLHr != flp.SteamFlowKgHr || fhp.SteamWaterLHr != fhp.SteamFlowKgHr {
			t.Errorf("%s steam_water_l_hr should match steam_flow_kg_hr", region)
		}
		if flp.EjectorDraftInWC < 0.70 || flp.EjectorDraftInWC > 1.25 || fhp.EjectorDraftInWC < 3.80 || fhp.EjectorDraftInWC > 6.00 {
			t.Errorf("%s ejector draft out of bounds: LP=%.3f HP=%.3f inWC", region, flp.EjectorDraftInWC, fhp.EjectorDraftInWC)
		}
		if !(flp.EjectorEntrainmentRatio > fhp.EjectorEntrainmentRatio && fhp.EjectorEntrainmentRatio > 4.0) {
			t.Errorf("%s entrainment ratio unexpected: LP=%.2f HP=%.2f", region, flp.EjectorEntrainmentRatio, fhp.EjectorEntrainmentRatio)
		}
		if !(flp.SteamThermalKW > 12.0 && flp.SteamThermalKW < 22.0 && fhp.SteamThermalKW > 32.0 && fhp.SteamThermalKW < 48.0) {
			t.Errorf("%s steam thermal kW unexpected: LP=%.1f HP=%.1f", region, flp.SteamThermalKW, fhp.SteamThermalKW)
		}
		if flp.KilnWasteHeatKW <= 40.0 || !flp.WasteHeatSufficient || !fhp.WasteHeatSufficient {
			t.Errorf("%s expected ample Cone 11 kiln waste heat: avail=%.1f kW LP_suf=%v HP_suf=%v", region, flp.KilnWasteHeatKW, flp.WasteHeatSufficient, fhp.WasteHeatSufficient)
		}
		if !(fhp.SubmicronNaClCapturePct > flp.SubmicronNaClCapturePct && flp.SubmicronNaClCapturePct > fp.SubmicronNaClCapturePct) {
			t.Errorf("%s submicron NaCl capture monotonicity failed: passive=%.1f LP=%.1f HP=%.1f",
				region, fp.SubmicronNaClCapturePct, flp.SubmicronNaClCapturePct, fhp.SubmicronNaClCapturePct)
		}
		if flp.SubmicronNaClCapturePct < 58.0 || flp.SubmicronNaClCapturePct > 66.0 || fhp.SubmicronNaClCapturePct < 76.0 || fhp.SubmicronNaClCapturePct > 84.0 {
			t.Errorf("%s submicron capture out of targets: LP=%.1f%% HP=%.1f%%", region, flp.SubmicronNaClCapturePct, fhp.SubmicronNaClCapturePct)
		}
		if !(flp.ChimneyHeightM < fp.ChimneyHeightM && fhp.ChimneyHeightM < fp.ChimneyHeightM && flp.ChimneyHeightM >= 4.5 && flp.ChimneyHeightM <= 5.5 && fhp.ChimneyHeightM >= 4.5 && fhp.ChimneyHeightM <= 5.5) {
			t.Errorf("%s steam chimney height should drop to [4.5, 5.5] m vs passive %.1f m: LP=%.1f HP=%.1f",
				region, fp.ChimneyHeightM, flp.ChimneyHeightM, fhp.ChimneyHeightM)
		}
		if !flp.DraftFeasible || !fhp.DraftFeasible || flp.DraftMarginInWC <= 0.5 || fhp.DraftMarginInWC <= 3.0 {
			t.Errorf("%s steam ejector draft margin should be generous: LP=%+.3f HP=%+.3f", region, flp.DraftMarginInWC, fhp.DraftMarginInWC)
		}

		// Zero-electricity invariants: 0 W electric, 0 circuits in ShopPlan, exact BOM cost match, steam parts/plumbing/bracing present.
		for _, ev := range []StudioEvaluation{evLP, evHP} {
			fr := ev.Frugal
			if fr.PowerWatts != 0 || ev.Urban.Electrical.RequiredVoltage != 0 || ev.Urban.Electrical.RequiredAmps != 0 {
				t.Errorf("%s %s off-grid must draw 0 W / 0 V / 0 A: power=%.1f elec=%+v", region, fr.DraftMode, fr.PowerWatts, ev.Urban.Electrical)
			}
			if len(ev.ShopPlan.Electrical) != 0 || ev.ShopPlan.Electrical == nil {
				t.Errorf("%s %s off-grid must have empty non-nil ShopPlan.Electrical", region, fr.DraftMode)
			}
			if !ev.ShopPlan.SteamEjectorEnabled || ev.ShopPlan.LocalDraftMode != fr.DraftMode {
				t.Errorf("%s %s ShopPlan steam flags mismatch: mode=%q enabled=%v", region, fr.DraftMode, ev.ShopPlan.LocalDraftMode, ev.ShopPlan.SteamEjectorEnabled)
			}
			if math.Abs(fr.EstimatedLocalCostUSD-ev.COTS.EstimatedTotalUSD) > 0.02 {
				t.Errorf("%s %s Frugal cost ($%.2f) != COTS BOM total ($%.2f)", region, fr.DraftMode, fr.EstimatedLocalCostUSD, ev.COTS.EstimatedTotalUSD)
			}
			if len(ev.ShopPlan.Parts) != 11 || len(ev.ShopPlan.Plumbing) != 7 || len(ev.ShopPlan.Bracing) != 6 {
				t.Errorf("%s %s expected steam ShopPlan parts=11 plumbing=7 bracing=6, got %d/%d/%d",
					region, fr.DraftMode, len(ev.ShopPlan.Parts), len(ev.ShopPlan.Plumbing), len(ev.ShopPlan.Bracing))
			}
			if len(fr.SteamSafetyRequirements) < 5 {
				t.Errorf("%s %s expected detailed steam_safety_requirements, got %d", region, fr.DraftMode, len(fr.SteamSafetyRequirements))
			}
			// Verify S1 steam header in Plumbing and B6 in Bracing.
			hasS1 := false
			for _, pl := range ev.ShopPlan.Plumbing {
				if pl.ID == "S1" && pl.PressurePSI == fr.SteamPressurePSIG {
					hasS1 = true
				}
			}
			if !hasS1 {
				t.Errorf("%s %s missing S1 insulated steam header in ShopPlan.Plumbing", region, fr.DraftMode)
			}
			// Verify Urban water reflects steam makeup water and acoustic level matches LP (62 dBA) / HP (68 dBA).
			if !strings.Contains(ev.Urban.Water.Summary, "L/hr") {
				t.Errorf("%s %s Urban.Water.Summary missing steam L/hr makeup note: %q", region, fr.DraftMode, ev.Urban.Water.Summary)
			}
			wantDBA := 62.0
			if fr.DraftMode == LocalDraftSteamEjectorHP {
				wantDBA = 68.0
			}
			if ev.Urban.Noise.BareDBA3Ft != wantDBA {
				t.Errorf("%s %s expected bare %.1f dBA steam hiss, got %.1f dBA", region, fr.DraftMode, wantDBA, ev.Urban.Noise.BareDBA3Ft)
			}
			// Verify BuildGuideMD and OperationsChecklistMD document PRV pull-ring, condensation growth, and reject DIY drum boilers.
			for _, want := range []string{"7.6 Zero-Electricity Steam-Jet Ejector", "PRV", "1.5", "DIY", "Condensation"} {
				if !strings.Contains(ev.BuildGuideMD, want) {
					t.Errorf("%s %s BuildGuideMD missing %q", region, fr.DraftMode, want)
				}
			}
			if !strings.Contains(ev.OperationsChecklistMD, "PRV") || !strings.Contains(ev.OperationsChecklistMD, "Zero DIY Welded Drum Boilers") {
				t.Errorf("%s %s OperationsChecklistMD missing PRV / certified boiler checklist", region, fr.DraftMode)
			}
			// Zero banned strings.
			raw, _ := json.Marshal(ev)
			haystack := string(raw) + "\n" + ev.BuildGuideMD + "\n" + ev.OperationsChecklistMD + "\n" + string(buildPacketReadme(ev))
			for _, banned := range bannedPhrases {
				if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
					t.Errorf("%s %s output contains banned phrase %q", region, fr.DraftMode, banned)
				}
			}
		}

		// Verify cost delta bounds: LP adds +$280-$360 less chimney savings (~$235-$330 net), HP adds +$490-$560 less chimney savings (~$440-$530 net).
		lpDelta := flp.EstimatedLocalCostUSD - fp.EstimatedLocalCostUSD
		hpDelta := fhp.EstimatedLocalCostUSD - fp.EstimatedLocalCostUSD
		if lpDelta < 220.0 || lpDelta > 370.0 || hpDelta < 430.0 || hpDelta > 570.0 || hpDelta <= lpDelta {
			t.Errorf("%s unexpected local steam cost deltas: LP=+$%.2f HP=+$%.2f", region, lpDelta, hpDelta)
		}
	}

	// 3. Hybrid mode (zero_electricity=false + steam_ejector_hp): 230V domestic water pump (370 W, 1 circuit E1, no fan E2).
	hybridHP := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorHP)
	hybridHP.ZeroElectricity = false
	hybridHP.AvailableVoltage = 230
	hybridHP.IncludeSilencer = true
	hybridHP.PropertyLineFt = 5.0 // trigger Caution/ActionNeeded acoustic branch coverage
	hybEval, err := eval.Evaluate(hybridHP)
	if err != nil {
		t.Fatalf("hybrid steam HP evaluate: %v", err)
	}
	if hybEval.Frugal.PowerWatts != 370.0 || len(hybEval.ShopPlan.Electrical) != 1 || hybEval.ShopPlan.Electrical[0].ID != "E1" {
		t.Errorf("hybrid steam HP expected 370 W & 1 circuit (E1 pump only), got %.0f W & %d circuits", hybEval.Frugal.PowerWatts, len(hybEval.ShopPlan.Electrical))
	}
	hybrid120 := hybridHP
	hybrid120.AvailableVoltage = 120
	hybrid120.PropertyLineFt = 10.0
	if ev120, err := eval.Evaluate(hybrid120); err != nil || ev120.Urban.Electrical.RequiredVoltage != 120 {
		t.Fatalf("hybrid 120V steam evaluate failed: err=%v elec=%+v", err, ev120.Urban.Electrical)
	}

	// 4. Dynamic waste_heat_sufficient=false edge case: low-temp flue + HP steam demand.
	lowHeat := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)
	lowHeat.Input.PeakExitTempF = 650.0
	lowHeatEval, err := eval.Evaluate(lowHeat)
	if err != nil {
		t.Fatalf("low-heat HP evaluate: %v", err)
	}
	if lowHeatEval.Frugal.WasteHeatSufficient {
		t.Errorf("expected waste_heat_sufficient=false at 650 F flue (waste=%.1f kW vs steam=%.1f kW)",
			lowHeatEval.Frugal.KilnWasteHeatKW, lowHeatEval.Frugal.SteamThermalKW)
	}

	// 5. Verify exact 18 new JSON contract keys in FrugalAssessment + 2 in ShopPlan + BuildPacketZip.
	hpNairobi, err := eval.Evaluate(OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP))
	if err != nil {
		t.Fatalf("hpNairobi evaluate: %v", err)
	}
	rawJSON, err := json.Marshal(hpNairobi)
	if err != nil {
		t.Fatalf("marshal hpNairobi: %v", err)
	}
	var top map[string]any
	if err := json.Unmarshal(rawJSON, &top); err != nil {
		t.Fatalf("unmarshal top: %v", err)
	}
	frRaw := top["frugal"].(map[string]any)
	for _, key := range []string{
		"draft_mode", "draft_mode_label", "steam_enabled",
		"steam_pressure_psig", "steam_pressure_bar",
		"steam_nozzle_vel_m_s", "steam_nozzle_mach", "steam_nozzle_throat_mm",
		"steam_flow_kg_hr", "steam_water_l_hr", "steam_thermal_kw",
		"kiln_waste_heat_kw", "waste_heat_sufficient",
		"ejector_draft_inwc", "ejector_entrainment_ratio",
		"condensation_growth_note", "steam_source_type", "steam_safety_requirements",
	} {
		if _, ok := frRaw[key]; !ok {
			t.Errorf("FrugalAssessment JSON missing Phase 10B contract key %q", key)
		}
	}
	spRaw := top["shop_plan"].(map[string]any)
	if spRaw["local_draft_mode"] != LocalDraftSteamEjectorHP || spRaw["steam_ejector_enabled"] != true {
		t.Errorf("ShopPlan JSON missing Phase 10B keys: local_draft_mode=%v steam_ejector_enabled=%v", spRaw["local_draft_mode"], spRaw["steam_ejector_enabled"])
	}
	if zipBytes, _, err := eval.BuildPacketZip(OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)); err != nil || len(zipBytes) == 0 {
		t.Fatalf("Steam-Jet BuildPacketZip failed: err=%v len=%d", err, len(zipBytes))
	}

	// 6. Validation & ParseProject round-trip: valid steam modes, empty build_path with steam mode, and invalid draft mode error.
	if parsed, err := ParseProject([]byte(`{"preset_id":"wood_salt_24","build_path":"local_ceramic_masonry","zero_electricity":true,"local_region":"nairobi","local_draft_mode":"steam_ejector_hp"}`)); err != nil {
		t.Fatalf("ParseProject steam_ejector_hp failed: %v", err)
	} else if parsed.LocalDraftMode != LocalDraftSteamEjectorHP {
		t.Errorf("unexpected parsed local_draft_mode: %q", parsed.LocalDraftMode)
	}
	emptyBP := DefaultProject()
	emptyBP.BuildPath = ""
	emptyBP.LocalDraftMode = LocalDraftSteamEjectorLP
	if err := emptyBP.Validate(); err != nil {
		t.Errorf("expected empty build_path with steam_ejector_lp to validate cleanly, got: %v", err)
	}
	if emptyBP.Normalize().BuildPath != BuildPathLocalCeramic {
		t.Errorf("expected empty build_path with steam_ejector_lp to normalize to local_ceramic_masonry")
	}
	emptyBPNoSteam := DefaultProject()
	emptyBPNoSteam.BuildPath = ""
	emptyBPNoSteam.LocalDraftMode = ""
	if err := emptyBPNoSteam.Validate(); err == nil {
		t.Errorf("expected empty build_path without steam mode to fail validation")
	}
	if _, err := ParseProject([]byte(`{"local_draft_mode":"rocket_nozzle"}`)); err == nil {
		t.Errorf("expected validation error for unknown local_draft_mode")
	}

	// Log exact summary table for engineering verification.
	evNBLP, _ := eval.Evaluate(OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorLP))
	evMBLP, _ := eval.Evaluate(OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorLP))
	evMBHP, _ := eval.Evaluate(OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorHP))
	t.Logf("Phase 10B Zero-Electricity Steam-Jet Ejector Summary (24 cu.ft. Cone 11 Wood-Salt):\n"+
		"  Nairobi LP (20 psig, 83.5 kPa): v_e=%.1f m/s (M=%.2f), d*=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Cost=%s\n"+
		"  Nairobi HP (60 psig, 83.5 kPa): v_e=%.1f m/s (M=%.2f), d*=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Cost=%s\n"+
		"  Mumbai  LP (20 psig, 101.3kPa): v_e=%.1f m/s (M=%.2f), d*=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Cost=%s\n"+
		"  Mumbai  HP (60 psig, 101.3kPa): v_e=%.1f m/s (M=%.2f), d*=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Cost=%s",
		evNBLP.Frugal.SteamNozzleVelMS, evNBLP.Frugal.SteamNozzleMach, evNBLP.Frugal.SteamNozzleThroatMM, evNBLP.Frugal.SteamFlowKgHr, evNBLP.Frugal.SteamThermalKW, evNBLP.Frugal.KilnWasteHeatKW, evNBLP.Frugal.EjectorDraftInWC, evNBLP.Frugal.EjectorEntrainmentRatio, evNBLP.Frugal.ChimneyHeightM, evNBLP.Frugal.SubmicronNaClCapturePct, evNBLP.Frugal.LocalCurrencyEstimate,
		hpNairobi.Frugal.SteamNozzleVelMS, hpNairobi.Frugal.SteamNozzleMach, hpNairobi.Frugal.SteamNozzleThroatMM, hpNairobi.Frugal.SteamFlowKgHr, hpNairobi.Frugal.SteamThermalKW, hpNairobi.Frugal.KilnWasteHeatKW, hpNairobi.Frugal.EjectorDraftInWC, hpNairobi.Frugal.EjectorEntrainmentRatio, hpNairobi.Frugal.ChimneyHeightM, hpNairobi.Frugal.SubmicronNaClCapturePct, hpNairobi.Frugal.LocalCurrencyEstimate,
		evMBLP.Frugal.SteamNozzleVelMS, evMBLP.Frugal.SteamNozzleMach, evMBLP.Frugal.SteamNozzleThroatMM, evMBLP.Frugal.SteamFlowKgHr, evMBLP.Frugal.SteamThermalKW, evMBLP.Frugal.KilnWasteHeatKW, evMBLP.Frugal.EjectorDraftInWC, evMBLP.Frugal.EjectorEntrainmentRatio, evMBLP.Frugal.ChimneyHeightM, evMBLP.Frugal.SubmicronNaClCapturePct, evMBLP.Frugal.LocalCurrencyEstimate,
		evMBHP.Frugal.SteamNozzleVelMS, evMBHP.Frugal.SteamNozzleMach, evMBHP.Frugal.SteamNozzleThroatMM, evMBHP.Frugal.SteamFlowKgHr, evMBHP.Frugal.SteamThermalKW, evMBHP.Frugal.KilnWasteHeatKW, evMBHP.Frugal.EjectorDraftInWC, evMBHP.Frugal.EjectorEntrainmentRatio, evMBHP.Frugal.ChimneyHeightM, evMBHP.Frugal.SubmicronNaClCapturePct, evMBHP.Frugal.LocalCurrencyEstimate,
	)
}

func TestPhase10CSubsonicQuietSteamJetEjector(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())
	bannedPhrases := []string{"Hobby " + "316L", "4-" + "lb", "gro" + "cery", "Switching from Salt " + "(NaCl) to Soda"}

	// 1. Helper & predicate checks for steam_ejector_subsonic.
	subNBProj := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorSubsonic)
	if subNBProj.LocalDraftMode != LocalDraftSteamEjectorSubsonic || !subNBProj.IsSteamEjector() || subNBProj.EffectiveLocalDraftMode() != LocalDraftSteamEjectorSubsonic {
		t.Fatalf("unexpected subsonic project helpers: %+v", subNBProj)
	}

	// 2. Evaluate all 4 modes across Nairobi, Mumbai, and General Local and verify strict physics,
	// capture, acoustic, and cost monotonicity!
	for _, region := range []string{LocalRegionNairobi, LocalRegionMumbai, LocalRegionGeneral} {
		evPassive, err := eval.Evaluate(OffGridCeramicProject(region))
		if err != nil {
			t.Fatalf("%s passive evaluate: %v", region, err)
		}
		evSub, err := eval.Evaluate(OffGridSteamCeramicProject(region, LocalDraftSteamEjectorSubsonic))
		if err != nil {
			t.Fatalf("%s subsonic evaluate: %v", region, err)
		}
		evLP, err := eval.Evaluate(OffGridSteamCeramicProject(region, LocalDraftSteamEjectorLP))
		if err != nil {
			t.Fatalf("%s LP evaluate: %v", region, err)
		}
		evHP, err := eval.Evaluate(OffGridSteamCeramicProject(region, LocalDraftSteamEjectorHP))
		if err != nil {
			t.Fatalf("%s HP evaluate: %v", region, err)
		}

		fp, fs, flp, fhp := evPassive.Frugal, evSub.Frugal, evLP.Frugal, evHP.Frugal

		// Core subsonic metadata & pressure.
		if !fs.SteamEnabled || fs.DraftMode != LocalDraftSteamEjectorSubsonic || fs.SteamPressurePSIG != 6.0 || fs.SteamPressureBar != 0.414 {
			t.Errorf("%s subsonic metadata wrong: enabled=%v mode=%q psig=%.1f bar=%.3f",
				region, fs.SteamEnabled, fs.DraftMode, fs.SteamPressurePSIG, fs.SteamPressureBar)
		}

		// Strict Mach monotonicity: 0 < Mach_subsonic < 1.0 < Mach_lp < Mach_hp.
		if !(0.0 < fs.SteamNozzleMach && fs.SteamNozzleMach < 1.0 && 1.0 < flp.SteamNozzleMach && flp.SteamNozzleMach < fhp.SteamNozzleMach) {
			t.Errorf("%s Mach monotonicity failed: sub=%.2f LP=%.2f HP=%.2f",
				region, fs.SteamNozzleMach, flp.SteamNozzleMach, fhp.SteamNozzleMach)
		}
		if region == LocalRegionNairobi {
			if fs.SteamNozzleVelMS < 345.0 || fs.SteamNozzleVelMS > 360.0 || fs.SteamNozzleMach < 0.72 || fs.SteamNozzleMach > 0.78 {
				t.Errorf("Nairobi subsonic velocity/Mach out of IAPWS-95 bounds: v=%.1f m/s M=%.2f", fs.SteamNozzleVelMS, fs.SteamNozzleMach)
			}
			if fs.ChimneyHeightM != 5.5 {
				t.Errorf("Nairobi subsonic chimney expected 5.5 m, got %.1f m", fs.ChimneyHeightM)
			}
		} else {
			if fs.SteamNozzleVelMS < 322.0 || fs.SteamNozzleVelMS > 335.0 || fs.SteamNozzleMach < 0.67 || fs.SteamNozzleMach > 0.72 {
				t.Errorf("%s subsonic velocity/Mach out of IAPWS-95 bounds: v=%.1f m/s M=%.2f", region, fs.SteamNozzleVelMS, fs.SteamNozzleMach)
			}
			if fs.ChimneyHeightM != 5.2 {
				t.Errorf("%s subsonic chimney expected 5.2 m, got %.1f m", region, fs.ChimneyHeightM)
			}
		}

		// Drilled brass converging orifice sizing (7.5-9.0 mm) & 1D momentum balance.
		if fs.SteamNozzleThroatMM < 7.5 || fs.SteamNozzleThroatMM > 9.0 {
			t.Errorf("%s subsonic brass orifice diameter out of [7.5, 9.0] mm: %.2f mm", region, fs.SteamNozzleThroatMM)
		}
		if fs.SteamFlowKgHr < 17.0 || fs.SteamFlowKgHr > 20.5 || fs.SteamWaterLHr != fs.SteamFlowKgHr {
			t.Errorf("%s subsonic steam flow out of bounds: %.1f kg/hr (water=%.1f L/hr)", region, fs.SteamFlowKgHr, fs.SteamWaterLHr)
		}
		if fs.SteamThermalKW < 12.0 || fs.SteamThermalKW > 15.0 || !fs.WasteHeatSufficient {
			t.Errorf("%s subsonic thermal load unexpected: %.1f kW vs %.1f kW kiln flue (suf=%v)",
				region, fs.SteamThermalKW, fs.KilnWasteHeatKW, fs.WasteHeatSufficient)
		}
		if fs.EjectorEntrainmentRatio < 17.5 || fs.EjectorEntrainmentRatio > 20.5 {
			t.Errorf("%s subsonic entrainment ratio unexpected: %.2f:1", region, fs.EjectorEntrainmentRatio)
		}
		if fs.EjectorDraftInWC < 0.35 || fs.EjectorDraftInWC > 0.45 {
			t.Errorf("%s subsonic ejector draft out of [0.35, 0.45] inWC: %.3f inWC", region, fs.EjectorDraftInWC)
		}
		if !fs.DraftFeasible || fs.DraftMarginInWC < 0.22 {
			t.Errorf("%s subsonic draft margin should be generous (>=+0.22 inWC): avail=%.3f req=%.3f margin=%+.3f",
				region, fs.AvailableDraftInWC, fs.RequiredDraftInWC, fs.DraftMarginInWC)
		}

		// Strict submicron NaCl capture monotonicity: 30% < 55% < 62% < 80%.
		if fp.SubmicronNaClCapturePct != 30.0 || fs.SubmicronNaClCapturePct != 55.0 || flp.SubmicronNaClCapturePct != 62.0 || fhp.SubmicronNaClCapturePct != 80.0 {
			t.Errorf("%s submicron NaCl capture monotonicity mismatch: passive=%.1f sub=%.1f LP=%.1f HP=%.1f",
				region, fp.SubmicronNaClCapturePct, fs.SubmicronNaClCapturePct, flp.SubmicronNaClCapturePct, fhp.SubmicronNaClCapturePct)
		}
		if fs.HClRemovalPct != 97.5 || fs.CoarseAshCapturePct != 92.5 {
			t.Errorf("%s subsonic HCl/ash targets mismatch: HCl=%.1f%% ash=%.1f%%", region, fs.HClRemovalPct, fs.CoarseAshCapturePct)
		}

		// Strict acoustic monotonicity: 45 dBA < 52 dBA < 62 dBA < 68 dBA (and ~29.5 dBA @ 40 ft).
		if evPassive.Urban.Noise.BareDBA3Ft != 45.0 || evSub.Urban.Noise.BareDBA3Ft != 52.0 || evLP.Urban.Noise.BareDBA3Ft != 62.0 || evHP.Urban.Noise.BareDBA3Ft != 68.0 {
			t.Errorf("%s acoustic monotonicity mismatch: passive=%.1f sub=%.1f LP=%.1f HP=%.1f",
				region, evPassive.Urban.Noise.BareDBA3Ft, evSub.Urban.Noise.BareDBA3Ft, evLP.Urban.Noise.BareDBA3Ft, evHP.Urban.Noise.BareDBA3Ft)
		}
		if math.Abs(evSub.Urban.Noise.PropertyLineDBA-29.5) > 0.2 {
			t.Errorf("%s subsonic 40 ft property line noise expected ~29.5 dBA, got %.1f dBA", region, evSub.Urban.Noise.PropertyLineDBA)
		}

		// Strict cost monotonicity: cost_passive < cost_subsonic < cost_lp < cost_hp & BOM match.
		if !(fp.EstimatedLocalCostUSD < fs.EstimatedLocalCostUSD && fs.EstimatedLocalCostUSD < flp.EstimatedLocalCostUSD && flp.EstimatedLocalCostUSD < fhp.EstimatedLocalCostUSD) {
			t.Errorf("%s cost monotonicity failed: passive=$%.2f sub=$%.2f LP=$%.2f HP=$%.2f",
				region, fp.EstimatedLocalCostUSD, fs.EstimatedLocalCostUSD, flp.EstimatedLocalCostUSD, fhp.EstimatedLocalCostUSD)
		}
		if math.Abs(fs.EstimatedLocalCostUSD-evSub.COTS.EstimatedTotalUSD) > 0.02 {
			t.Errorf("%s subsonic Frugal cost ($%.2f) != COTS BOM total ($%.2f)", region, fs.EstimatedLocalCostUSD, evSub.COTS.EstimatedTotalUSD)
		}
		if region == LocalRegionNairobi && (fs.EstimatedLocalCostUSD < 920.0 || fs.EstimatedLocalCostUSD > 990.0) {
			t.Errorf("Nairobi subsonic total cost expected ~$955, got $%.2f", fs.EstimatedLocalCostUSD)
		}
		if region == LocalRegionMumbai && (fs.EstimatedLocalCostUSD < 755.0 || fs.EstimatedLocalCostUSD > 815.0) {
			t.Errorf("Mumbai subsonic total cost expected ~$785, got $%.2f", fs.EstimatedLocalCostUSD)
		}

		// Zero-electricity & ShopPlan invariants: 0 W, 0 electrical circuits, S1 + S2 + P6 plumbing, brass orifice part.
		if fs.PowerWatts != 0 || evSub.Urban.Electrical.RequiredVoltage != 0 || evSub.Urban.Electrical.RequiredAmps != 0 {
			t.Errorf("%s subsonic off-grid must draw 0 W / 0 V / 0 A: power=%.1f elec=%+v", region, fs.PowerWatts, evSub.Urban.Electrical)
		}
		if len(evSub.ShopPlan.Electrical) != 0 || evSub.ShopPlan.Electrical == nil {
			t.Errorf("%s subsonic off-grid must have empty non-nil ShopPlan.Electrical", region)
		}
		if !evSub.ShopPlan.SteamEjectorEnabled || evSub.ShopPlan.LocalDraftMode != LocalDraftSteamEjectorSubsonic || evSub.ShopPlan.VenturiLayout != "subsonic_brass_orifice_condensation" {
			t.Errorf("%s subsonic ShopPlan flags mismatch: enabled=%v mode=%q vent=%q",
				region, evSub.ShopPlan.SteamEjectorEnabled, evSub.ShopPlan.LocalDraftMode, evSub.ShopPlan.VenturiLayout)
		}
		if len(evSub.ShopPlan.Parts) != 11 || len(evSub.ShopPlan.Plumbing) != 8 || len(evSub.ShopPlan.Bracing) != 6 {
			t.Errorf("%s subsonic ShopPlan expected parts=11 plumbing=8 (P1..P5,S1,S2,P6) bracing=6, got %d/%d/%d",
				region, len(evSub.ShopPlan.Parts), len(evSub.ShopPlan.Plumbing), len(evSub.ShopPlan.Bracing))
		}
		var hasS1, hasS2, hasP6 bool
		for _, pl := range evSub.ShopPlan.Plumbing {
			switch pl.ID {
			case "S1":
				if pl.PressurePSI == 6.0 {
					hasS1 = true
				}
			case "S2":
				if strings.Contains(pl.Label, "4.2 m") && strings.Contains(strings.ToLower(pl.Label), "hydrostatic") {
					hasS2 = true
				}
			case "P6":
				hasP6 = true
			}
		}
		if !hasS1 || !hasS2 || !hasP6 {
			t.Errorf("%s subsonic ShopPlan.Plumbing missing S1/S2/P6: S1=%v S2=%v P6=%v", region, hasS1, hasS2, hasP6)
		}

		// Safety requirements, guide, checklist & zero banned strings.
		if len(fs.SteamSafetyRequirements) < 7 {
			t.Errorf("%s subsonic expected >=7 safety requirements, got %d", region, len(fs.SteamSafetyRequirements))
		}
		safetyJoined := strings.Join(fs.SteamSafetyRequirements, "\n")
		for _, must := range []string{"4.2 m", "6.0 psig", "PRV", "fusible", "1.5x MAWP", "check valve", "DIY"} {
			if !strings.Contains(strings.ToLower(safetyJoined), strings.ToLower(must)) {
				t.Errorf("%s subsonic steam_safety_requirements missing %q", region, must)
			}
		}
		for _, want := range []string{"0.414 barg", "4.2 m", "Drilled Brass Converging Orifice", "Mach < 1.0", "52 dBA", "DIY"} {
			if !strings.Contains(evSub.BuildGuideMD, want) {
				t.Errorf("%s subsonic BuildGuideMD missing %q", region, want)
			}
		}
		if !strings.Contains(evSub.OperationsChecklistMD, "4.2 m") || !strings.Contains(evSub.OperationsChecklistMD, "Zero Unvented DIY Welded Drum Boilers") {
			t.Errorf("%s subsonic OperationsChecklistMD missing standpipe/DIY safety items", region)
		}
		raw, _ := json.Marshal(evSub)
		haystack := string(raw) + "\n" + evSub.BuildGuideMD + "\n" + evSub.OperationsChecklistMD + "\n" + string(buildPacketReadme(evSub))
		for _, banned := range bannedPhrases {
			if strings.Contains(strings.ToLower(haystack), strings.ToLower(banned)) {
				t.Errorf("%s subsonic output contains banned phrase %q", region, banned)
			}
		}
	}

	// 3. Hybrid pump mode (zero_electricity=false + steam_ejector_subsonic): 370 W pump only, 1 circuit E1, 56.5% NaCl.
	hybSub := OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorSubsonic)
	hybSub.ZeroElectricity = false
	hybSub.AvailableVoltage = 230
	hybSub.IncludeSilencer = true
	hybEv, err := eval.Evaluate(hybSub)
	if err != nil {
		t.Fatalf("hybrid subsonic evaluate: %v", err)
	}
	if hybEv.Frugal.PowerWatts != 370.0 || len(hybEv.ShopPlan.Electrical) != 1 || hybEv.Frugal.SubmicronNaClCapturePct != 56.5 || hybEv.Urban.Noise.SilencedDBA3Ft != 48.0 {
		t.Errorf("hybrid subsonic unexpected metrics: power=%.0f elec=%d nacl=%.1f sil=%.1f",
			hybEv.Frugal.PowerWatts, len(hybEv.ShopPlan.Electrical), hybEv.Frugal.SubmicronNaClCapturePct, hybEv.Urban.Noise.SilencedDBA3Ft)
	}

	// 4. ParseProject round-trip, empty build_path normalization, exact 63 FrugalAssessment JSON keys, & BuildPacketZip.
	parsed, err := ParseProject([]byte(`{"preset_id":"wood_salt_24","build_path":"local_ceramic_masonry","zero_electricity":true,"local_region":"nairobi","local_draft_mode":"steam_ejector_subsonic"}`))
	if err != nil || parsed.LocalDraftMode != LocalDraftSteamEjectorSubsonic {
		t.Fatalf("ParseProject steam_ejector_subsonic failed: err=%v mode=%q", err, parsed.LocalDraftMode)
	}
	emptyBP := DefaultProject()
	emptyBP.BuildPath = ""
	emptyBP.LocalDraftMode = LocalDraftSteamEjectorSubsonic
	if err := emptyBP.Validate(); err != nil || emptyBP.Normalize().BuildPath != BuildPathLocalCeramic {
		t.Errorf("empty build_path with steam_ejector_subsonic failed: err=%v bp=%q", err, emptyBP.Normalize().BuildPath)
	}
	evNairobiSub, err := eval.Evaluate(parsed)
	if err != nil {
		t.Fatalf("evNairobiSub evaluate: %v", err)
	}
	rawJSON, _ := json.Marshal(evNairobiSub)
	var top map[string]any
	if err := json.Unmarshal(rawJSON, &top); err != nil {
		t.Fatalf("unmarshal evNairobiSub: %v", err)
	}
	frRaw := top["frugal"].(map[string]any)
	if len(frRaw) < 63 {
		t.Errorf("expected at least 63 FrugalAssessment JSON fields intact, got %d", len(frRaw))
	}
	if zipBytes, _, err := eval.BuildPacketZip(parsed); err != nil || len(zipBytes) == 0 {
		t.Fatalf("Subsonic BuildPacketZip failed: err=%v len=%d", err, len(zipBytes))
	}

	evMumbaiSub, _ := eval.Evaluate(OffGridSteamCeramicProject(LocalRegionMumbai, LocalDraftSteamEjectorSubsonic))
	t.Logf("Phase 10C Subsonic Quiet Steam-Jet Summary (6 psig / 0.414 barg, 24 cu.ft. Cone 11 Wood-Salt):\n"+
		"  Nairobi (83.5 kPa):  v_e=%.1f m/s (M=%.2f < 1), d_orif=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Noise=%.1f->%.1f dBA, Cost=%s\n"+
		"  Mumbai  (101.3 kPa): v_e=%.1f m/s (M=%.2f < 1), d_orif=%.2f mm, m_s=%.1f kg/hr (%.1f kW vs %.1f kW flue), dP_ej=%.3f inWC (ER=%.1f), H=%.1f m, NaCl=%.0f%%, Noise=%.1f->%.1f dBA, Cost=%s",
		evNairobiSub.Frugal.SteamNozzleVelMS, evNairobiSub.Frugal.SteamNozzleMach, evNairobiSub.Frugal.SteamNozzleThroatMM, evNairobiSub.Frugal.SteamFlowKgHr, evNairobiSub.Frugal.SteamThermalKW, evNairobiSub.Frugal.KilnWasteHeatKW, evNairobiSub.Frugal.EjectorDraftInWC, evNairobiSub.Frugal.EjectorEntrainmentRatio, evNairobiSub.Frugal.ChimneyHeightM, evNairobiSub.Frugal.SubmicronNaClCapturePct, evNairobiSub.Urban.Noise.BareDBA3Ft, evNairobiSub.Urban.Noise.PropertyLineDBA, evNairobiSub.Frugal.LocalCurrencyEstimate,
		evMumbaiSub.Frugal.SteamNozzleVelMS, evMumbaiSub.Frugal.SteamNozzleMach, evMumbaiSub.Frugal.SteamNozzleThroatMM, evMumbaiSub.Frugal.SteamFlowKgHr, evMumbaiSub.Frugal.SteamThermalKW, evMumbaiSub.Frugal.KilnWasteHeatKW, evMumbaiSub.Frugal.EjectorDraftInWC, evMumbaiSub.Frugal.EjectorEntrainmentRatio, evMumbaiSub.Frugal.ChimneyHeightM, evMumbaiSub.Frugal.SubmicronNaClCapturePct, evMumbaiSub.Urban.Noise.BareDBA3Ft, evMumbaiSub.Urban.Noise.PropertyLineDBA, evMumbaiSub.Frugal.LocalCurrencyEstimate,
	)
}

func TestTailoredBuildPacketAndPhysicsHonestyRegressions(t *testing.T) {
	eval := NewEvaluator(scrubber.NewDesigner(), cad.NewExporter())

	// 1. Local ceramic off-grid packet omits sheet-metal STL/DXF/BOM and daemon.json, and bundles shop_plan.svg.
	localProj := OffGridSteamCeramicProject(LocalRegionNairobi, LocalDraftSteamEjectorHP)
	zipBytes, evLocal, err := eval.BuildPacketZip(localProj)
	if err != nil {
		t.Fatalf("local BuildPacketZip: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	names := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, forbidden := range []string{"scrubber.stl", "scrubber.dxf", "bom_engineered.csv", "daemon.json"} {
		if names[forbidden] {
			t.Errorf("local_ceramic_masonry packet must omit %q", forbidden)
		}
	}
	if !names["shop_plan.svg"] || !names["shop_plan.json"] || !names["BUILD_GUIDE.md"] {
		t.Errorf("local_ceramic_masonry packet missing core visual/guide files: %+v", names)
	}

	// 2. Local ceramic guide & ops checklist must not leak electric VFD / NEC 430 / 24VDC damper instructions.
	for _, leaked := range []string{"NEC 430.102", "jog VFD at 15 Hz", "24 VDC spring-return"} {
		if strings.Contains(evLocal.BuildGuideMD, leaked) || strings.Contains(evLocal.OperationsChecklistMD, leaked) {
			t.Errorf("local_ceramic_masonry guide/checklist leaked electric text %q", leaked)
		}
	}

	// 3. 60 psig HP steam-jet strictly bans studio monotube coils & enforces realistic condensing exit temp (< 180 F).
	if strings.Contains(strings.ToLower(evLocal.Frugal.SteamSourceType), "monotube") {
		t.Errorf("60 psig HP steam source must not endorse monotube coil: %q", evLocal.Frugal.SteamSourceType)
	}
	if evLocal.Frugal.StackReheatTempF >= 180.0 {
		t.Errorf("60 psig Nairobi condensing exit temp %.1f F must stay well below 202.4 F boiling point", evLocal.Frugal.StackReheatTempF)
	}

	// 4. Electromechanical-only electric build omits daemon.json but retains STL/DXF/SVG.
	emProj := DefaultProject()
	emProj.ControlTier = ControlTierElectromechanicalOnly
	emZip, _, err := eval.BuildPacketZip(emProj)
	if err != nil {
		t.Fatalf("electromechanical BuildPacketZip: %v", err)
	}
	emZR, _ := zip.NewReader(bytes.NewReader(emZip), int64(len(emZip)))
	emNames := make(map[string]bool, len(emZR.File))
	for _, f := range emZR.File {
		emNames[f.Name] = true
	}
	if emNames["daemon.json"] {
		t.Errorf("electromechanical_only packet must omit daemon.json")
	}
	if !emNames["scrubber.stl"] || !emNames["shop_plan.svg"] {
		t.Errorf("electromechanical_only packet must retain scrubber.stl and shop_plan.svg")
	}
}
