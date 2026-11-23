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

func TestUSCodeComplianceC1ThroughC14(t *testing.T) {
	eval := NewEvaluator(nil, nil)

	// C1 & C2: NEC 430.122(A) 125% VFD 1-phase input sizing + Plume Diluter amps + NEC 430.102/210.8(F)/400.12.
	refProj := DefaultProject() // 24 cu.ft. wood/salt -> 2.0 HP blower
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

	// Plume diluter adds electrical draw.
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

	// 1.0 HP on 120V must flag 30A / caution per NEC 430.122(A), whereas on 240V it uses 240V 20A OK.
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

	// C3: Real AutomationDirect GS21 SKUs + NEMA 4X enclosure, never truncated GS21-10P/GS21-20P.
	if strings.Contains(refEval.COTS.BlowerVFD.SKU, "GS21-20P") || !strings.Contains(refEval.COTS.BlowerVFD.SKU, "GS21-22P0") {
		t.Errorf("C3: 2.0 HP VFD SKU expected GS21-22P0, got %q", refEval.COTS.BlowerVFD.SKU)
	}
	if !strings.Contains(refEval.COTS.BlowerVFD.Material, "NEMA 4X") {
		t.Errorf("C3: VFD must specify NEMA 4X enclosure, got %q", refEval.COTS.BlowerVFD.Material)
	}

	// C4: IPC Section 608 backflow prevention in water summary, COTS BOM, and guides.
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
	if !hasRPZItem {
		t.Error("C4: COTS BOM missing IPC 608 backflow preventer item")
	}
	if !hasDisconnectItem {
		t.Error("C2: COTS BOM missing NEC 430.102 disconnect & line reactor item")
	}
	if !strings.Contains(refEval.BuildGuideMD, "IPC Section 608") || !strings.Contains(refEval.OperationsChecklistMD, "IPC 608") {
		t.Error("C4: guides missing IPC 608 backflow checklist items")
	}

	// C5: IPC 702.5 <=140 F cooldown + 40 CFR 403.5(b)(2) pH >= 5.0.
	if !strings.Contains(refEval.Urban.Drain.Summary, "140 F") || !strings.Contains(refEval.Urban.Drain.Summary, "40 CFR 403.5") {
		t.Errorf("C5: DrainAssessment missing IPC 702.5 <=140 F or 40 CFR 403.5: %s", refEval.Urban.Drain.Summary)
	}
	if !strings.Contains(refEval.OperationsChecklistMD, "<= 140 F") || !strings.Contains(refEval.OperationsChecklistMD, "40 CFR 403.5") {
		t.Error("C5: OperationsChecklistMD missing IPC 702.5 <=140 F or 40 CFR 403.5")
	}

	// C6: No thin-walled 55-gal HDPE drum as vacuum packed-bed column; 200 F hardwired snap-disc documented.
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

	// C8: Salt mode never calls exhaust "harmless water vapor" and flags <30 ft setback.
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

	// C9, C11, C12, C13: No 650 F RTV silicone recommendation, >=2300 F ceramic gaskets, 36" NFPA 211, NFPA 91 cleanout, zinc-free diluter tee, OSHA 1910.95 noise.
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

	// Cover StudioProject.Validate invalid arms.
	badProjects := []func(*StudioProject){
		func(p *StudioProject) { p.FiringDurationHours = 0 },
		func(p *StudioProject) { p.BuildPath = "riveted" },
		func(p *StudioProject) { p.CorrosionTier = "carbon_steel" },
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

	// Cover all COTS sizing tiers: unsnapped hobby & heavy, large sump (>45 gal), 3 HP & 5+ HP VFDs,
	// all plume diluter CFM tiers (500, 1000, 1500, 2000, 3000), FRP-18/24/30 columns, large nozzles,
	// and all urban assessment branches (water caution/action, noise caution/action silenced & unsilenced,
	// plume moderate/low/action, electrical 0.5/1.5/3/5+ HP).
	for _, cfm := range []float64{60, 140, 240, 340, 550} {
		d := model.ScrubberDesign{}
		d.Blower.FlowACFM = cfm
		d.Blower.TotalStaticPressureInWC = 20.0
		gotCFM, cone := RecommendedDilutionCFM(d)
		if gotCFM <= 0 || cone <= 0 {
			t.Errorf("RecommendedDilutionCFM(%.0f) = (%.0f, %.0f)", cfm, gotCFM, cone)
		}
		part := selectPlumeDiluter(DefaultProject(), d)
		if part.UnitCostUSD <= 0 {
			t.Errorf("selectPlumeDiluter(%.0f) invalid cost", cfm)
		}
		if a := diluterAmps120V(d); a <= 0 {
			t.Errorf("diluterAmps120V(%.0f) <= 0", cfm)
		}
	}

	// Direct COTS & electrical unit checks for 0/0.5/0.75/1.5/3.0/6.0 HP and large sump/nozzles.
	for _, hp := range []float64{0, 0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 7.5} {
		for _, v := range []int{120, 240} {
			p := DefaultProject()
			p.AvailableVoltage = v
			p.IncludePlumeDiluter = true
			d := model.ScrubberDesign{}
			d.Blower.RecommendedMotorHP = hp
			d.Blower.FlowACFM = 260
			d.Blower.TotalStaticPressureInWC = 21
			ea := assessElectrical(p, d)
			if ea.RequiredAmps <= 0 || ea.CircuitLabel == "" {
				t.Errorf("assessElectrical(hp=%v, v=%d) invalid: %+v", hp, v, ea)
			}
			vfd := selectBlowerVFD(p, d)
			if vfd.UnitCostUSD <= 0 {
				t.Errorf("selectBlowerVFD(hp=%v, v=%d) invalid", hp, v)
			}
		}
	}

	// Large 100-gal sump, large nozzle tiers (>0.75 & >1.8 GPM/nozzle), FRP-18/24/30 column snapping, and SnapToCOTS=false.
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

	sump := selectSumpTank(bigDesign)
	if sump.SKU != "USPLASTIC-100T" {
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

	colFRP, _, _ := selectPackedColumn(DefaultProject(), bigDesign)
	if !strings.Contains(colFRP.SKU, "FRP-24") {
		t.Errorf("expected FRP-24 vacuum column for 20\" ID, got %q", colFRP.SKU)
	}

	// Unsnapped hobby & heavy evaluations + urban water/noise/plume moderate/mild/tight cases.
	unsnapHobby := DefaultProject()
	unsnapHobby.SnapToCOTS = false
	unsnapHobby.GardenHoseGPM = 0.8 // triggers water action_needed
	unsnapHobby.PropertyLineFt = 4.0
	unsnapHobby.IncludeSilencer = true
	unsnapHobby.AmbientWinterF = 55.0 // moderate plume
	if ev, err := eval.Evaluate(unsnapHobby); err != nil {
		t.Fatalf("unsnapHobby evaluate: %v", err)
	} else {
		if ev.Urban.Water.Status != StatusActionNeeded {
			t.Errorf("expected water action_needed, got %q", ev.Urban.Water.Status)
		}
		if ev.Urban.Noise.Status != StatusActionNeeded {
			t.Errorf("expected noise action_needed at 4 ft even silenced, got %q (%.1f dBA)", ev.Urban.Noise.Status, ev.Urban.Noise.PropertyLineDBA)
		}
	}

	unsnapHeavy := DefaultProject()
	unsnapHeavy.SnapToCOTS = false
	unsnapHeavy.CorrosionTier = CorrosionTierHeavyC276FRP
	unsnapHeavy.VaporGlazeMode = VaporGlazeNone
	unsnapHeavy.GardenHoseGPM = 1.4 // >70% caution
	unsnapHeavy.AmbientWinterF = 75.0
	unsnapHeavy.PropertyLineFt = 8.0 // <10 ft IMC 501.3.1 Item 2 action_needed
	if ev, err := eval.Evaluate(unsnapHeavy); err != nil {
		t.Fatalf("unsnapHeavy evaluate: %v", err)
	} else {
		if ev.Urban.Water.Status != StatusCaution {
			t.Errorf("expected water caution, got %q", ev.Urban.Water.Status)
		}
		if ev.Urban.Plume.Status != StatusActionNeeded {
			t.Errorf("expected non-salt plume action_needed at 8 ft (<10 ft IMC), got %q", ev.Urban.Plume.Status)
		}
	}

	// Moderate & mild plume OK cases + noise caution (55-60 dBA).
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
	mildProj.PropertyLineFt = 25.0 // unsilenced ~58-60 dBA
	if ev, err := eval.Evaluate(mildProj); err != nil {
		t.Fatalf("mildProj evaluate: %v", err)
	} else if ev.Urban.Plume.VisiblePlumeRisk != "low" {
		t.Errorf("expected low plume risk at 72 F, got %q", ev.Urban.Plume.VisiblePlumeRisk)
	}
}
