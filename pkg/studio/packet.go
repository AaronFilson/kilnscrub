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
	Frugal                *FrugalAssessment         `json:"frugal,omitempty"`
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

	frugal := EvaluateFrugal(project, design)
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
		Frugal:              frugal,
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

	// Keep quench setpoint safely above adiabatic saturation and below the aligned 200 F CPVC trip.
	spF := math.Max(186.0, eval.Design.Quench.OutletTemp.Fahrenheit()+4.0)
	if spF > 195.0 {
		spF = 195.0
	}
	ctrl.QuenchSetpoint = model.FahrenheitToKelvin(spF)
	ctrl.Sync()
	if err := ctrl.Validate(); err != nil {
		return nil, fmt.Errorf("validate generated daemon control config: %w", err)
	}

	nomWater := math.Max(0.5, round2(eval.Design.Quench.TotalSprayWaterGPM))
	maxWater := math.Max(2.5, round2(nomWater*2.0))
	maxSuctionPa := round1(math.Max(30.0, math.Abs(float64(ctrl.TargetDraftPa))*2.5))

	cfg := daemonFileConfig{
		Description: fmt.Sprintf("%s (%.1f cu.ft., %s) Control Daemon Config — switch hal_mode to \"modbus\" for physical RS-485 PLC/VFD hardware",
			eval.Project.StudioName, eval.Design.Input.KilnVolumeCuFt, eval.Project.FiringCone),
		HALMode: "sim",
		Control: ctrl,
		Modbus: daemonModbusConfig{
			Device:                     "/dev/ttyUSB0",
			BaudRate:                   19200,
			SlaveID:                    1,
			TimeoutNs:                  500 * time.Millisecond,
			UseHoldingRegisters:        false,
			BypassCoilDeenergizeToTrip: true,
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
			BlowerMaxSuctionPa:    maxSuctionPa,
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
	fmt.Fprintf(&b, "Stage Options:    %s | %s (%.0f\" pipe) | %s | U-Turn=%v | SharedSump=%v | Controls=%s\n",
		eval.Project.StageAlignment, eval.Project.PackedBedLayout, eval.Project.ParallelPipeSizeIn, eval.Project.VenturiLayout,
		eval.Project.CompactFootprint, eval.Project.SharedBaffledSump, eval.Project.ControlTier)
	fmt.Fprintf(&b, "Budget Options:   Blower=%s (%+.0f USD) | Packing=%s (%+.0f USD) | Shop Compressor=%.0f SCFM (Eductor feasible=%v, Air-Atomize feasible=%v)\n",
		eval.Project.BlowerDriveOption, eval.ArchitectureOptions.BlowerDrive.CostDeltaUSD,
		eval.Project.PackingMediaOption, eval.ArchitectureOptions.PackingMedia.CostDeltaUSD,
		eval.ArchitectureOptions.Compressor.AvailableSCFM,
		eval.ArchitectureOptions.Compressor.EductorFeasible, eval.ArchitectureOptions.Compressor.AtomizeFeasible)
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
	fmt.Fprintf(&b, "Bulk Baking Soda: %s\n", eval.Urban.Drain.BulkPackagingLabel)
	if eval.Frugal != nil && eval.Frugal.Enabled {
		fmt.Fprintf(&b, "Local Ceramic:    %s | Draft=%s | Zero-Electricity=%v (%.0f W) | Chimney=%.1f m (%.3f inWC avail vs %.3f inWC req)\n",
			eval.Frugal.RegionLabel, eval.Frugal.DraftMode, eval.Frugal.ZeroElectricity, eval.Frugal.PowerWatts,
			eval.Frugal.ChimneyHeightM, eval.Frugal.AvailableDraftInWC, eval.Frugal.RequiredDraftInWC)
		if eval.Frugal.SteamEnabled {
			barFmt := fmt.Sprintf("%.2f barg", eval.Frugal.SteamPressureBar)
			if eval.Frugal.SteamPressureBar < 1.0 {
				barFmt = fmt.Sprintf("%.3f barg", eval.Frugal.SteamPressureBar)
			}
			fmt.Fprintf(&b, "Steam-Jet Ejector: %.0f psig (%s) | Nozzle=%.1f m/s (Mach %.2f, %.2f mm) | Steam=%.1f kg/hr (%.1f kW vs %.1f kW kiln flue waste heat) | Ejector=%.3f inWC (%.1f:1)\n",
				eval.Frugal.SteamPressurePSIG, barFmt, eval.Frugal.SteamNozzleVelMS, eval.Frugal.SteamNozzleMach,
				eval.Frugal.SteamNozzleThroatMM, eval.Frugal.SteamFlowKgHr, eval.Frugal.SteamThermalKW, eval.Frugal.KilnWasteHeatKW,
				eval.Frugal.EjectorDraftInWC, eval.Frugal.EjectorEntrainmentRatio)
		}
		fmt.Fprintf(&b, "Local Cost/Chem:  %s | Alkali=%s (%s): %.2f kg/firing | Capture: %.1f%% HCl / %.0f%% submicron NaCl\n",
			eval.Frugal.LocalCurrencyEstimate, eval.Frugal.AlkaliName, eval.Frugal.AlkaliFormula,
			eval.Frugal.AlkaliKgPerFiring, eval.Frugal.HClRemovalPct, eval.Frugal.SubmicronNaClCapturePct)
	}
	b.WriteString("\n")

	isLocal := eval.Project.IsLocalCeramic()
	includeDaemon := !isLocal && eval.Project.ControlTier != ControlTierElectromechanicalOnly

	b.WriteString("FILES INCLUDED IN THIS PACKET\n")
	b.WriteString("-----------------------------\n")
	idx := 1
	addEntry := func(name, desc string) {
		fmt.Fprintf(&b, "%-3d %-23s %s\n", idx, name, desc)
		idx++
	}
	addEntry("BUILD_GUIDE.md", "Step-by-step fabrication, assembly & commissioning guide")
	addEntry("OPERATIONS_CHECKLIST.md", "Firing-day safety, draft, recirculation & disposal checklist")
	addEntry("bom_cots.md", "Tailored hardware & materials shopping list (Markdown)")
	addEntry("bom_cots.csv", "Tailored hardware & materials shopping list (Spreadsheet CSV)")
	if !isLocal {
		addEntry("bom_engineered.csv", "Baseline sheet-metal fabrication BOM (Spreadsheet CSV)")
		addEntry("scrubber.stl", "Baseline 4-stage 3D assembly model (millimeters)")
		addEntry("scrubber.dxf", "Baseline 4-stage 2D AutoCAD R12 flat-pattern cut sheets (millimeters)")
	}
	addEntry("shop_plan.svg", "Standalone architectural blueprint schematic (SVG, tailored to active build path)")
	addEntry("shop_plan.json", "Structured Complete Systems Plan (parts, cuts, electrical, plumbing, bracing, acoustics)")
	addEntry("project.json", "Saved studio project settings (re-import into kilnscrub UI/CLI)")
	if includeDaemon {
		addEntry("daemon.json", "Pre-tuned kilnscrub control daemon config (sim or Modbus RS-485)")
	}
	addEntry("README_PACKET.txt", "This overview file")
	b.WriteString("\n")

	if isLocal {
		b.WriteString("NOTE ON TAILORED PACKET CONTENTS (Local Ceramic / Masonry Build):\n")
		b.WriteString("Sheet-metal STL/DXF files (scrubber.stl, scrubber.dxf, bom_engineered.csv) and\n")
		b.WriteString("electric Modbus daemon.json are intentionally omitted because this build uses\n")
		b.WriteString("local firebrick masonry, stoneware Raschig rings, and gravity/steam-jet hardware.\n")
		b.WriteString("Use shop_plan.svg, shop_plan.json, bom_cots.csv, and BUILD_GUIDE.md for layout.\n\n")
		b.WriteString("SAFETY REMINDER (Masonry / Gravity / Certified Steam): Cross-brace the 3.5 m\n")
		b.WriteString("overhead gravity tank stand (200-500 kg wet load), keep >= 36\" clearance from\n")
		b.WriteString("hot flues to combustibles (NFPA 211), NEVER fabricate DIY welded-drum steam\n")
		b.WriteString("boilers, and cool sump liquor <= 140 F & neutralize to pH 6.5-8.5 before disposal.\n")
	} else {
		p := eval.Project
		if p.StageAlignment != StageAlignmentInlineIDBlower ||
			p.PackedBedLayout != PackedBedLayoutSingleColumn ||
			p.VenturiLayout != VenturiLayoutSingleThroat ||
			p.CompactFootprint ||
			p.IsDeluxe() {
			b.WriteString("NOTE ON BASELINE CAD FILES: scrubber.stl and scrubber.dxf represent the baseline\n")
			b.WriteString("inline 4-stage vessel geometry. Refer to shop_plan.svg, shop_plan.json, and\n")
			b.WriteString("BUILD_GUIDE.md for your selected Phase 7-9 options (eductor, Uniseal parallel\n")
			b.WriteString("bundle, mini-venturi manifold, U-turn skid, or Deluxe ANSI flanges/skid).\n\n")
		}
		if !includeDaemon {
			b.WriteString("NOTE ON CONTROLS (Electromechanical Only): daemon.json is omitted because this\n")
			b.WriteString("build uses a 100% hardwired 24VDC safety relay panel with zero software daemon.\n\n")
		}
		b.WriteString("SAFETY REMINDER (NFPA 86 / IPC 608 / NEC 430): Always install and test the\n")
		b.WriteString("independent hardwired 200 F snap-disc + 24VDC de-energize-to-trip safety relay\n")
		b.WriteString("driving the spring-return fail-open hot-gas bypass damper (aligned with the 200 F /\n")
		b.WriteString("0.25 GPM software backup), IPC 608 Air Gap/RPZ backflow preventer, and NEC 430.102\n")
		b.WriteString("lockable sight disconnect before firing.\n")
	}
	return []byte(b.String())
}

// BuildPacketZip evaluates the project and packages tailored fabrication, shopping,
// guide, schematic SVG, CAD (when applicable), structured shop plan, and control
// configuration files into an in-memory .zip archive.
func (e *Evaluator) BuildPacketZip(project StudioProject) ([]byte, StudioEvaluation, error) {
	eval, err := e.Evaluate(project)
	if err != nil {
		return nil, StudioEvaluation{}, err
	}

	isLocal := eval.Project.IsLocalCeramic()
	includeDaemon := !isLocal && eval.Project.ControlTier != ControlTierElectromechanicalOnly

	bomCOTSCSV, err := e.exporter.FormatBOMCSV(eval.COTS.BOM)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet format COTS BOM CSV: %w", err)
	}
	bomCOTSMD, err := e.exporter.FormatBOMMarkdown(eval.COTS.BOM)
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet format COTS BOM MD: %w", err)
	}
	shopPlanJSON, err := json.MarshalIndent(eval.ShopPlan, "", "  ")
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet marshal shop_plan.json: %w", err)
	}
	shopPlanJSON = append(shopPlanJSON, '\n')
	shopPlanSVG := RenderShopPlanSVG(eval.ShopPlan)

	projectJSON, err := json.MarshalIndent(eval.Project, "", "  ")
	if err != nil {
		return nil, StudioEvaluation{}, fmt.Errorf("packet marshal project.json: %w", err)
	}
	projectJSON = append(projectJSON, '\n')

	type packetEntry struct {
		name string
		data []byte
	}
	var entries []packetEntry

	if !isLocal {
		stlBytes, err := e.exporter.ExportSTL(eval.Design)
		if err != nil {
			return nil, StudioEvaluation{}, fmt.Errorf("packet export STL: %w", err)
		}
		dxfBytes, err := e.exporter.ExportDXF(eval.Design)
		if err != nil {
			return nil, StudioEvaluation{}, fmt.Errorf("packet export DXF: %w", err)
		}
		engBOM, err := e.exporter.GenerateBOM(eval.Design)
		if err != nil {
			return nil, StudioEvaluation{}, fmt.Errorf("packet generate engineered BOM: %w", err)
		}
		bomEngCSV, err := e.exporter.FormatBOMCSV(engBOM)
		if err != nil {
			return nil, StudioEvaluation{}, fmt.Errorf("packet format engineered BOM CSV: %w", err)
		}
		entries = append(entries,
			packetEntry{name: "scrubber.stl", data: stlBytes},
			packetEntry{name: "scrubber.dxf", data: dxfBytes},
			packetEntry{name: "bom_engineered.csv", data: bomEngCSV},
		)
	}

	entries = append(entries,
		packetEntry{name: "shop_plan.svg", data: shopPlanSVG},
		packetEntry{name: "shop_plan.json", data: shopPlanJSON},
		packetEntry{name: "bom_cots.csv", data: bomCOTSCSV},
		packetEntry{name: "bom_cots.md", data: bomCOTSMD},
		packetEntry{name: "BUILD_GUIDE.md", data: []byte(eval.BuildGuideMD)},
		packetEntry{name: "OPERATIONS_CHECKLIST.md", data: []byte(eval.OperationsChecklistMD)},
		packetEntry{name: "project.json", data: projectJSON},
	)

	if includeDaemon {
		daemonJSON, err := buildDaemonJSON(eval)
		if err != nil {
			return nil, StudioEvaluation{}, fmt.Errorf("packet build daemon.json: %w", err)
		}
		entries = append(entries, packetEntry{name: "daemon.json", data: daemonJSON})
	}
	entries = append(entries, packetEntry{name: "README_PACKET.txt", data: buildPacketReadme(eval)})

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
