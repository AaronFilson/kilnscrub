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

func TestAllPresetsEvaluateCleanly(t *testing.T) {
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
	}

	// Verify electric_raku_6 is 120V friendly.
	raku, _ := PresetByID("electric_raku_6")
	rakuEval, err := eval.Evaluate(raku.Project)
	if err != nil {
		t.Fatalf("raku evaluate: %v", err)
	}
	if rakuEval.Urban.Electrical.RequiredVoltage != 120 || !rakuEval.Urban.Electrical.FitsAvailableVoltage {
		t.Errorf("electric_raku_6 should fit 120V, got %+v", rakuEval.Urban.Electrical)
	}

	// Verify coop_wood_48 uses heavy_c276_frp and passes corrosion cleanly.
	coop, _ := PresetByID("coop_wood_48")
	coopEval, err := eval.Evaluate(coop.Project)
	if err != nil {
		t.Fatalf("coop evaluate: %v", err)
	}
	if coopEval.Urban.Corrosion.Status != StatusOK || !strings.Contains(coopEval.Urban.Corrosion.MaterialSummary, "C-276") {
		t.Errorf("coop_wood_48 corrosion expected OK with C-276, got %+v", coopEval.Urban.Corrosion)
	}
}

func TestParseLoadSaveProjectPartialJSON(t *testing.T) {
	// Partial JSON referencing a preset should inherit that preset's defaults.
	partial := []byte(`{"preset_id":"gas_soda_14","studio_name":"Backyard Test","property_line_ft":45}`)
	proj, err := ParseProject(partial)
	if err != nil {
		t.Fatalf("ParseProject partial failed: %v", err)
	}
	if proj.StudioName != "Backyard Test" {
		t.Errorf("StudioName = %q, want Backyard Test", proj.StudioName)
	}
	if proj.PropertyLineFt != 45 {
		t.Errorf("PropertyLineFt = %v, want 45", proj.PropertyLineFt)
	}
	if proj.VaporGlazeMode != VaporGlazeSoda {
		t.Errorf("VaporGlazeMode = %q, want %q", proj.VaporGlazeMode, VaporGlazeSoda)
	}
	if proj.Input.KilnVolumeCuFt != 14.0 {
		t.Errorf("KilnVolumeCuFt = %v, want 14.0", proj.Input.KilnVolumeCuFt)
	}

	// Partial JSON without preset_id inherits DefaultProject().
	proj2, err := ParseProject([]byte(`{"firing_duration_hours": 12}`))
	if err != nil {
		t.Fatalf("ParseProject default-seeded failed: %v", err)
	}
	if proj2.FiringDurationHours != 12 || proj2.Input.KilnVolumeCuFt != 24.0 {
		t.Errorf("unexpected default-seeded project: %+v", proj2)
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
	if _, err := ParseProject([]byte(`{"available_voltage": 480}`)); err == nil {
		t.Error("expected error for invalid voltage 480")
	}
	if _, err := ParseProject([]byte(`{"vapor_glaze_mode": "chlorine"}`)); err == nil {
		t.Error("expected error for invalid vapor_glaze_mode")
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
		// Verify BOM totals and line weights are consistent.
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

func TestUrbanNoiseVoltageAndBakingSodaMath(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	// 1. 24 cu.ft. wood/salt kiln on 120V must flag action_needed for electrical (requires 240V >= 1.5 HP).
	proj := DefaultProject()
	proj.AvailableVoltage = 120
	proj.IncludeSilencer = false
	proj.PropertyLineFt = 20.0
	proj.FiringDurationHours = 16.0

	res, err := eval.Evaluate(proj)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if res.Urban.Electrical.RequiredVoltage != 240 {
		t.Errorf("expected RequiredVoltage=240 for 24 cu.ft. kiln, got %d", res.Urban.Electrical.RequiredVoltage)
	}
	if res.Urban.Electrical.FitsAvailableVoltage {
		t.Error("expected FitsAvailableVoltage=false when 120V supplied for 240V blower")
	}
	if res.Urban.Electrical.Status != StatusActionNeeded {
		t.Errorf("expected electrical status action_needed, got %q", res.Urban.Electrical.Status)
	}

	// Verify inverse-square noise attenuation math: L2 = L1 - 20*log10(d2/3).
	wantBareLine := round1(res.Urban.Noise.BareDBA3Ft - 20.0*math.Log10(20.0/3.0))
	if math.Abs(res.Urban.Noise.PropertyLineDBA-wantBareLine) > 0.11 {
		t.Errorf("unsilenced PropertyLineDBA = %.1f, want %.1f (bare3ft=%.1f)",
			res.Urban.Noise.PropertyLineDBA, wantBareLine, res.Urban.Noise.BareDBA3Ft)
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
	if resSil.Urban.Electrical.Status != StatusOK || !resSil.Urban.Electrical.FitsAvailableVoltage {
		t.Errorf("expected 240V electrical OK, got %+v", resSil.Urban.Electrical)
	}

	// Verify baking soda math: NaHCO3ConsumptionKgPerHr * duration / KgPerLb, ceil(/4) boxes.
	wantLbs := round1(resSil.Design.PackedBed.NaHCO3ConsumptionKgPerHr * proj.FiringDurationHours / model.KgPerLb)
	wantBoxes := int(math.Ceil(wantLbs / 4.0))
	if resSil.Urban.Drain.BakingSodaLbsPerFiring != wantLbs {
		t.Errorf("BakingSodaLbsPerFiring = %.1f, want %.1f", resSil.Urban.Drain.BakingSodaLbsPerFiring, wantLbs)
	}
	if resSil.Urban.Drain.BakingSoda4LbBoxes != wantBoxes {
		t.Errorf("BakingSoda4LbBoxes = %d, want %d", resSil.Urban.Drain.BakingSoda4LbBoxes, wantBoxes)
	}
	if wantBoxes < 12 {
		t.Errorf("expected >= 12 boxes for 16-hr 2 lbs/hr salt firing, got %d", wantBoxes)
	}
}

func TestSodaVsSaltEliminatesHCl(t *testing.T) {
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
		t.Errorf("expected 316L salt corrosion status caution, got %q", saltEval.Urban.Corrosion.Status)
	}

	// Switch to soda mode while retaining raw Input.SaltRateLbsPerHr=2.0 -> effective salt must be 0.
	sodaProj := saltProj
	sodaProj.VaporGlazeMode = VaporGlazeSoda
	sodaEval, err := eval.Evaluate(sodaProj)
	if err != nil {
		t.Fatalf("soda evaluate: %v", err)
	}
	if sodaEval.Project.VaporGlazeMode != VaporGlazeSoda {
		t.Errorf("expected preserved mode soda, got %q", sodaEval.Project.VaporGlazeMode)
	}
	if sodaEval.Design.Input.SaltRateLbsPerHr != 0 {
		t.Errorf("expected effective SaltRateLbsPerHr=0 in soda mode, got %v", sodaEval.Design.Input.SaltRateLbsPerHr)
	}
	if sodaEval.Design.PackedBed.HClInletPPM != 0 {
		t.Errorf("expected 0 HCl ppm in soda mode, got %v", sodaEval.Design.PackedBed.HClInletPPM)
	}
	if sodaEval.Urban.Drain.BakingSodaLbsPerFiring != 0 || sodaEval.Urban.Drain.BakingSoda4LbBoxes != 0 {
		t.Errorf("expected 0 baking soda demand in soda mode, got %+v", sodaEval.Urban.Drain)
	}
	if sodaEval.Urban.Corrosion.Status != StatusOK {
		t.Errorf("expected soda corrosion status ok, got %q", sodaEval.Urban.Corrosion.Status)
	}
	if !strings.Contains(sodaEval.Urban.Corrosion.SodaComparison, "100%") {
		t.Errorf("expected SodaComparison to highlight 100%% HCl elimination: %s", sodaEval.Urban.Corrosion.SodaComparison)
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

func TestBuildPacketZipContainsAllTenValidFiles(t *testing.T) {
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
	for _, key := range []string{"project", "design", "cots", "urban", "build_guide_md", "operations_checklist_md"} {
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

	// Reference 24 cu.ft. kiln in 40 F winter ambient.
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
	// Tangent-clear reheat for 24 cu.ft. @ 40 F / 60% RH matches plume-analyst (~48.6 kW).
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

	// Enabling IncludePlumeDiluter adds the COTS squirrel-cage assembly, raises BOM total,
	// updates BuildGuideMD, and clears the 30 ft property line cleanly.
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
