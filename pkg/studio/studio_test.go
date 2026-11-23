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

	// Tiny 6 cu.ft. kiln with parallel bed uses 3" pipe + 3/8" packing (D/d_p >= 7.5).
	raku, _ := PresetByID("electric_raku_6")
	rakuPB := raku.Project
	rakuPB.PackedBedLayout = PackedBedLayoutParallelManifold
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
	if !strings.Contains(cotsEval.OperationsChecklistMD, "210 F") || !strings.Contains(cotsEval.OperationsChecklistMD, "0.25 GPM") {
		t.Error("expected OperationsChecklistMD to include 210 F and 0.25 GPM trip tests")
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
