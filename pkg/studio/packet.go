package studio

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
)

// StudioEvaluation is the unified studio-kit output combining engineering physics,
// COTS component snapping & BOM, urban neighbor readiness, Phase 7 architecture options,
// closed-loop recirculation chemistry, structured in-browser shop plan, and Markdown guides.
type StudioEvaluation struct {
	Project               StudioProject             `json:"project"`
	Design                model.ScrubberDesign      `json:"design"`
	COTS                  COTSSummary               `json:"cots"`
	Urban                 UrbanAssessment           `json:"urban"`
	ArchitectureOptions   ArchitectureOptionsResult `json:"architecture_options"`
	Recirculation         RecirculationPlan         `json:"recirculation"`
	ShopPlan              ShopPlan                  `json:"shop_plan"`
	BuildGuideMD          string                    `json:"build_guide_md"`
	OperationsChecklistMD string                    `json:"operations_checklist_md"`
}

// Evaluator orchestrates scrubber design, COTS mapping, urban site readiness,
// Phase 7 selectable options, recirculation, shop plan generation, build/firing guides,
// and downloadable Build Packet (.zip) bundling.
type Evaluator struct {
	designer model.Designer
	exporter model.CADExporter
}

// NewEvaluator constructs an Evaluator with the provided Designer and CADExporter
// (falling back to default scrubber.NewDesigner and cad.NewExporter when nil).
func NewEvaluator(designer model.Designer, exporter model.CADExporter) *Evaluator {
	if designer == nil {
		designer = scrubber.NewDesigner()
	}
	if exporter == nil {
		exporter = cad.NewExporter()
	}
	return &Evaluator{
		designer: designer,
		exporter: exporter,
	}
}

// Evaluate validates project, solves the 4-stage scrubber physics, maps COTS hardware,
// computes urban readiness checks, Phase 7 selectable options, closed-loop recirculation,
// structured shop plans, and generates tailored build & firing Markdown guides.
func (e *Evaluator) Evaluate(project StudioProject) (StudioEvaluation, error) {
	if err := project.Validate(); err != nil {
		return StudioEvaluation{}, fmt.Errorf("studio evaluate: %w", err)
	}
	project = project.Normalize()

	effectiveInput := project.EffectiveDesignInput()
	design, err := e.designer.Design(effectiveInput)
	if err != nil {
		return StudioEvaluation{}, fmt.Errorf("studio design: %w", err)
	}

	cots := BuildCOTSSummary(project, design)
	archOpts := EvaluateArchitectureOptions(project, design, cots)
	recirc := BuildRecirculationPlan(project, design, cots)
	urban := AssessUrban(project, design, cots)
	shopPlan := BuildShopPlan(project, design, cots, urban, archOpts)

	eval := StudioEvaluation{
		Project:             project,
		Design:              design,
		COTS:                cots,
		Urban:               urban,
		ArchitectureOptions: archOpts,
		Recirculation:       recirc,
		ShopPlan:            shopPlan,
	}
	eval.BuildGuideMD = GenerateBuildGuide(eval)
	eval.OperationsChecklistMD = GenerateOperationsChecklist(eval)
	return eval, nil
}

// PacketFilename returns a clean, filesystem-safe zip filename for a StudioProject.
func PacketFilename(project StudioProject) string {
	slug := strings.TrimSpace(project.PresetID)
	if slug == "" {
		slug = strings.TrimSpace(project.StudioName)
	}
	if slug == "" {
		slug = "studio"
	}
	var b strings.Builder
	for _, r := range strings.ToLower(slug) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	clean := strings.Trim(b.String(), "-_")
	if clean == "" {
		clean = "studio"
	}
	return fmt.Sprintf("kilnscrub-%s-packet.zip", clean)
}

// --- Daemon config JSON schema matching configs/kilnscrub.example.json without importing internal/ ---

type daemonRegisterSpec struct {
	Address uint16  `json:"address"`
	Scale   float64 `json:"scale"`
	Signed  bool    `json:"signed"`
}

type daemonRegisterMap struct {
	TempUnit             string             `json:"temp_unit"`
	DraftPa              daemonRegisterSpec `json:"draft_pa"`
	QuenchInletTemp      daemonRegisterSpec `json:"quench_inlet_temp"`
	QuenchOutletTemp     daemonRegisterSpec `json:"quench_outlet_temp"`
	WaterFlowGPM         daemonRegisterSpec `json:"water_flow_gpm"`
	SumpPH               daemonRegisterSpec `json:"sump_ph"`
	BlowerSpeedFeedback  daemonRegisterSpec `json:"blower_speed_feedback"`
	BlowerVFDAddr        uint16             `json:"blower_vfd_addr"`
	QuenchSprayValveAddr uint16             `json:"quench_spray_valve_addr"`
	OutputScale          float64            `json:"output_scale"`
	BypassDamperCoilAddr uint16             `json:"bypass_damper_coil_addr"`
	AlarmLatchCoilAddr   uint16             `json:"alarm_latch_coil_addr"`
}

type daemonModbusConfig struct {
	Device                     string            `json:"device"`
	BaudRate                   int               `json:"baud_rate"`
	SlaveID                    byte              `json:"slave_id"`
	TimeoutNs                  time.Duration     `json:"timeout_ns"`
	UseHoldingRegisters        bool              `json:"use_holding_registers"`
	BypassCoilDeenergizeToTrip bool              `json:"bypass_coil_deenergize_to_trip,omitempty"`
	Registers                  daemonRegisterMap `json:"registers"`
}

type daemonSimConfig struct {
	NominalInletTempK     float64       `json:"nominal_inlet_temp_k"`
	BaselineQuenchTempK   float64       `json:"baseline_quench_temp_k"`
	MaxWaterFlowGPM       float64       `json:"max_water_flow_gpm"`
	NominalWaterFlowGPM   float64       `json:"nominal_water_flow_gpm"`
	BlowerMaxSuctionPa    float64       `json:"blower_max_suction_pa"`
	BlowerTimeConstNs     time.Duration `json:"blower_time_const_ns"`
	DraftTimeConstNs      time.Duration `json:"draft_time_const_ns"`
	ValveTimeConstNs      time.Duration `json:"valve_time_const_ns"`
	QuenchTimeConstNs     time.Duration `json:"quench_time_const_ns"`
	HeatSensitivityK      float64       `json:"heat_sensitivity_k"`
	CoolingGainKPerGPM    float64       `json:"cooling_gain_k_per_gpm"`
	DryGasOutletTempK     float64       `json:"dry_gas_outlet_temp_k"`
	InitialSumpPH         float64       `json:"initial_sump_ph"`
	DraftNoiseAmplitudePa float64       `json:"draft_noise_amplitude_pa"`
	DraftNoiseFreqHz      float64       `json:"draft_noise_freq_hz"`
	AutoStepIntervalNs    time.Duration `json:"auto_step_interval_ns"`
}

type daemonFileConfig struct {
	Description string              `json:"description"`
	HALMode     string              `json:"hal_mode"`
	Control     model.ControlConfig `json:"control"`
	Modbus      daemonModbusConfig  `json:"modbus"`
	Sim         daemonSimConfig     `json:"sim"`
}

func buildDaemonJSON(eval StudioEvaluation) ([]byte, error) {
	ctrl := model.DefaultControlConfig()
	ctrl.SetTargetDraftInWC(eval.Design.Input.TargetKilnDraftInWC)

	// Keep quench setpoint safely above adiabatic saturation and below the 210 F CPVC trip.
	spF := math.Max(188.0, eval.Design.Quench.OutletTemp.Fahrenheit()+4.0)
	if spF > 202.0 {
		spF = 202.0
	}
	ctrl.QuenchSetpoint = model.FahrenheitToKelvin(spF)
	ctrl.Sync()
	if err := ctrl.Validate(); err != nil {
		return nil, fmt.Errorf("validate generated daemon control config: %w", err)
	}

	nomWater := math.Max(0.5, round2(eval.Design.Quench.TotalSprayWaterGPM))
	maxWater := math.Max(2.5, round2(nomWater*2.0))

	cfg := daemonFileConfig{
		Description: fmt.Sprintf("%s (%.1f cu.ft., %s) Control Daemon Config — switch hal_mode to \"modbus\" for physical RS-485 PLC/VFD hardware",
			eval.Project.StudioName, eval.Design.Input.KilnVolumeCuFt, eval.Project.FiringCone),
		HALMode: "sim",
		Control: ctrl,
		Modbus: daemonModbusConfig{
			Device:              "/dev/ttyUSB0",
			BaudRate:            9600,
			SlaveID:             1,
			TimeoutNs:           500 * time.Millisecond,
			UseHoldingRegisters: false,
			Registers: daemonRegisterMap{
				TempUnit:             "F",
				DraftPa:              daemonRegisterSpec{Address: 0, Scale: 0.01, Signed: true},
				QuenchInletTemp:      daemonRegisterSpec{Address: 1, Scale: 0.1, Signed: true},
				QuenchOutletTemp:     daemonRegisterSpec{Address: 2, Scale: 0.1, Signed: true},
				WaterFlowGPM:         daemonRegisterSpec{Address: 3, Scale: 0.01, Signed: false},
				SumpPH:               daemonRegisterSpec{Address: 4, Scale: 0.01, Signed: false},
				BlowerSpeedFeedback:  daemonRegisterSpec{Address: 5, Scale: 0.0001, Signed: false},
				BlowerVFDAddr:        100,
				QuenchSprayValveAddr: 101,
				OutputScale:          10000,
				BypassDamperCoilAddr: 0,
				AlarmLatchCoilAddr:   1,
			},
		},
		Sim: daemonSimConfig{
			NominalInletTempK:     float64(eval.Design.InletGas.Temperature),
			BaselineQuenchTempK:   float64(ctrl.QuenchSetpoint),
			MaxWaterFlowGPM:       maxWater,
			NominalWaterFlowGPM:   nomWater,
			BlowerMaxSuctionPa:    30.0,
			BlowerTimeConstNs:     400 * time.Millisecond,
			DraftTimeConstNs:      300 * time.Millisecond,
			ValveTimeConstNs:      300 * time.Millisecond,
			QuenchTimeConstNs:     2000 * time.Millisecond,
			HeatSensitivityK:      55.0,
			CoolingGainKPerGPM:    22.0,
			DryGasOutletTempK:     float64(model.FahrenheitToKelvin(380.0)),
			InitialSumpPH:         eval.Design.Input.SumpPHTarget,
			DraftNoiseAmplitudePa: 0.6,
			DraftNoiseFreqHz:      4.0,
			AutoStepIntervalNs:    0,
		},
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal daemon.json: %w", err)
	}
	return append(out, '\n'), nil
}

func buildPacketReadme(eval StudioEvaluation) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "KILNSCRUB COMPLETE BUILD PACKET — %s\n", eval.Project.StudioName)
	b.WriteString(strings.Repeat("=", 72) + "\n\n")
	fmt.Fprintf(&b, "Preset ID:        %s\n", eval.Project.PresetID)
	fmt.Fprintf(&b, "Kiln Volume:      %.1f cu.ft. (%s / %.0f F peak, %.0f hr firing, %.1f hr salting)\n",
		eval.Design.Input.KilnVolumeCuFt, eval.Project.FiringCone,
		eval.Design.Input.PeakExitTempF, eval.Project.FiringDurationHours, eval.Project.EffectiveSaltingHours())
	fmt.Fprintf(&b, "Vapor Glaze Mode: %s | Build Path: %s | Corrosion Tier: %s\n",
		eval.Project.VaporGlazeMode, eval.Project.BuildPath, CorrosionTierLabel(eval.Project.CorrosionTier))
	fmt.Fprintf(&b, "Stage Options:    %s | %s | %s | U-Turn=%v | SharedSump=%v | Controls=%s\n",
		eval.Project.StageAlignment, eval.Project.PackedBedLayout, eval.Project.VenturiLayout,
		eval.Project.CompactFootprint, eval.Project.SharedBaffledSump, eval.Project.ControlTier)
	fmt.Fprintf(&b, "Urban Readiness:  %s\n", strings.ToUpper(eval.Urban.OverallStatus))
	fmt.Fprintf(&b, "Estimated COTS:   $%.2f USD (Delta %+.2f USD vs baseline | %.1f%% of flood)\n",
		eval.COTS.EstimatedTotalUSD, eval.ArchitectureOptions.CostDeltaUSD, eval.COTS.ActualPercentOfFlood)
	fmt.Fprintf(&b, "Footprint:        %.1f ft x %.1f ft (%.1f sq.ft., %.1f%% length savings)\n",
		eval.ArchitectureOptions.FootprintLengthFt, eval.ArchitectureOptions.FootprintWidthFt,
		eval.ArchitectureOptions.FootprintAreaSqFt, eval.ArchitectureOptions.FootprintSavingsPct)
	fmt.Fprintf(&b, "Electrical:       %s (%.1f A)\n", eval.Urban.Electrical.CircuitLabel, eval.Urban.Electrical.RequiredAmps)
	fmt.Fprintf(&b, "Water & Recirc:   %.2f GPM peak spray | Closed-loop saves %.1f%% water (dump every %d firing(s) / SG >= %.3f)\n",
		eval.Urban.Water.RequiredGPM, eval.Recirculation.WaterSavedPct,
		eval.Recirculation.FiringsBeforeFullDump, eval.Recirculation.DumpSpecificGravity)
	fmt.Fprintf(&b, "Neighbor Noise:   %.1f dBA at %.0f ft property line (limit %.0f dBA)\n",
		eval.Urban.Noise.PropertyLineDBA, eval.Urban.Noise.PropertyLineFt, eval.Urban.Noise.UrbanLimitDBA)
	fmt.Fprintf(&b, "Winter Plume:     ~%.0f ft bare -> ~%.0f ft with %.0f CFM squirrel-cage HV diluter (reheat for 0 ft: %.1f kW)\n",
		eval.Urban.Plume.BarePlumeLengthFt, eval.Urban.Plume.DilutedPlumeLengthFt,
		eval.Urban.Plume.DilutionCFMNeeded, eval.Urban.Plume.DilutionReheatKW)
	fmt.Fprintf(&b, "Bulk Baking Soda: %s\n\n", eval.Urban.Drain.BulkPackagingLabel)

	b.WriteString("FILES INCLUDED IN THIS PACKET\n")
	b.WriteString("-----------------------------\n")
	b.WriteString("1.  BUILD_GUIDE.md          Step-by-step fabrication, assembly & cold-water commissioning guide\n")
	b.WriteString("2.  OPERATIONS_CHECKLIST.md Firing-day safety, draft, recirculation & drain neutralization checklist\n")
	b.WriteString("3.  bom_cots.md             Off-the-shelf (COTS) hardware shopping list (Markdown)\n")
	b.WriteString("4.  bom_cots.csv            Off-the-shelf (COTS) hardware shopping list (Spreadsheet CSV)\n")
	b.WriteString("5.  bom_engineered.csv      Sheet-metal fabrication BOM (Spreadsheet CSV)\n")
	b.WriteString("6.  scrubber.stl            3D assembly model (millimeters) for visual review & layout\n")
	b.WriteString("7.  scrubber.dxf            2D AutoCAD R12 flat-pattern cut sheets (millimeters) for CNC/fab\n")
	b.WriteString("8.  shop_plan.json          Structured in-browser Complete Plan Viewer parts, cut sheets & clearances\n")
	b.WriteString("9.  project.json            Saved studio project settings (re-import into kilnscrub UI/CLI)\n")
	b.WriteString("10. daemon.json             Pre-tuned kilnscrub control daemon config (sim or Modbus RS-485)\n")
	b.WriteString("11. README_PACKET.txt       This overview file\n\n")
	b.WriteString("SAFETY REMINDER (NFPA 86 / IPC 608 / NEC 430): Always install and test the\n")
	b.WriteString("independent hardwired 200 F snap-disc + 24VDC de-energize-to-trip safety relay\n")
	b.WriteString("driving the spring-return fail-open hot-gas bypass damper (ahead of the 210 F /\n")
	b.WriteString("0.25 GPM software backup), IPC 608 Air Gap/RPZ backflow preventer, and NEC 430.102\n")
	b.WriteString("lockable sight disconnect before firing.\n")
	return []byte(b.String())
}

// BuildPacketZip evaluates the project and packages all 11 fabrication, shopping,
// guide, CAD, structured shop plan, and control configuration files into an in-memory .zip archive.
func (e *Evaluator) BuildPacketZip(project StudioProject) ([]byte, StudioEvaluation, error) {
	eval, err := e.Evaluate(project)
	if err != nil {
		return nil, StudioEvaluation{}, err
	}

	stlBytes, err := e.exporter.ExportSTL(eval.Design)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet export STL: %w", err)
	}
	dxfBytes, err := e.exporter.ExportDXF(eval.Design)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet export DXF: %w", err)
	}
	bomCOTSCSV, err := e.exporter.FormatBOMCSV(eval.COTS.BOM)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet format COTS BOM CSV: %w", err)
	}
	bomCOTSMD, err := e.exporter.FormatBOMMarkdown(eval.COTS.BOM)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet format COTS BOM MD: %w", err)
	}
	engBOM, err := e.exporter.GenerateBOM(eval.Design)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet generate engineered BOM: %w", err)
	}
	bomEngCSV, err := e.exporter.FormatBOMCSV(engBOM)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet format engineered BOM CSV: %w", err)
	}
	shopPlanJSON, err := json.MarshalIndent(eval.ShopPlan, "", "  ")
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet marshal shop_plan.json: %w", err)
	}
	shopPlanJSON = append(shopPlanJSON, '\n')

	projectJSON, err := json.MarshalIndent(eval.Project, "", "  ")
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet marshal project.json: %w", err)
	}
	projectJSON = append(projectJSON, '\n')

	daemonJSON, err := buildDaemonJSON(eval)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet build daemon.json: %w", err)
	}

	type packetEntry struct {
		name string
		data []byte
	}
	entries := []packetEntry{
		{name: "scrubber.stl", data: stlBytes},
		{name: "scrubber.dxf", data: dxfBytes},
		{name: "shop_plan.json", data: shopPlanJSON},
		{name: "bom_cots.csv", data: bomCOTSCSV},
		{name: "bom_cots.md", data: bomCOTSMD},
		{name: "bom_engineered.csv", data: bomEngCSV},
		{name: "BUILD_GUIDE.md", data: []byte(eval.BuildGuideMD)},
		{name: "OPERATIONS_CHECKLIST.md", data: []byte(eval.OperationsChecklistMD)},
		{name: "project.json", data: projectJSON},
		{name: "daemon.json", data: daemonJSON},
		{name: "README_PACKET.txt", data: buildPacketReadme(eval)},
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	now := time.Now().UTC()

	for _, entry := range entries {
		hdr := &zip.FileHeader{
			Name:     entry.name,
			Method:   zip.Deflate,
			Modified: now,
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			_ = zw.Close()
			return nil, StudioEvaluation{}, fmt.Errorf("packet zip create %s: %w", entry.name, err)
		}
		if _, err := w.Write(entry.data); err != nil {
			_ = zw.Close()
			return nil, StudioEvaluation{}, fmt.Errorf("packet zip write %s: %w", entry.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet zip close: %w", err)
	}

	return buf.Bytes(), eval, nil
}
