package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// PlanPart represents a numbered assembly part in the interactive Shop Plan Viewer.
type PlanPart struct {
	Num          int     `json:"num"`
	ID           string  `json:"id"`
	Label        string  `json:"label"`
	Stage        string  `json:"stage"`
	Material     string  `json:"material"`
	DimensionsIn string  `json:"dimensions_in"`
	DimensionsMM string  `json:"dimensions_mm"`
	SKU          string  `json:"sku"`
	Qty          float64 `json:"qty"`
	CostUSD      float64 `json:"cost_usd"`
	Notes        string  `json:"notes"`
}

// CutSheetItem describes a single pipe cut length or exact unfolded sheet-metal
// flat pattern (cylinder rectangle or frustum annular sector from pkg/cad math).
type CutSheetItem struct {
	PartNum       int     `json:"part_num"`
	Name          string  `json:"name"`
	Shape         string  `json:"shape"`
	Material      string  `json:"material"`
	FlatWidthIn   float64 `json:"flat_width_in"`
	FlatHeightIn  float64 `json:"flat_height_in"`
	OuterRadiusIn float64 `json:"outer_radius_in"`
	InnerRadiusIn float64 `json:"inner_radius_in"`
	SweepDeg      float64 `json:"sweep_deg"`
	Qty           int     `json:"qty"`
	Notes         string  `json:"notes"`
}

// PlanClearances specifies mandatory code setbacks, NEC electrical working space,
// pump/cleanout maintenance envelopes, hose bend radius, and shop service aisles.
type PlanClearances struct {
	FlueToCombustiblesIn float64 `json:"flue_to_combustibles_in"`
	PropertyLineMinFt    float64 `json:"property_line_min_ft"`
	StackAboveRoofFt     float64 `json:"stack_above_roof_ft"`
	ServiceAisleIn       float64 `json:"service_aisle_in"`
	NECPanelDepthIn      float64 `json:"nec_panel_depth_in"`
	NECPanelWidthIn      float64 `json:"nec_panel_width_in"`
	PumpServiceIn        float64 `json:"pump_service_in"`
	CleanoutSwingIn      float64 `json:"cleanout_swing_in"`
	HoseBendRadiusIn     float64 `json:"hose_bend_radius_in"`
	WiringAccessNote     string  `json:"wiring_access_note"`
}

// ElectricalCircuit specifies one field circuit or hardwired safety loop on the
// Complete Systems Installation Plan.
type ElectricalCircuit struct {
	ID           string  `json:"id"`
	Label        string  `json:"label"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Voltage      int     `json:"voltage"`
	Amps         float64 `json:"amps"`
	BreakerAmps  int     `json:"breaker_amps"`
	WireGauge    string  `json:"wire_gauge"`
	Conduit      string  `json:"conduit"`
	NECRef       string  `json:"nec_ref"`
	IsSafetyLoop bool    `json:"is_safety_loop"`
	Notes        string  `json:"notes"`
}

// PlumbingLine specifies one liquid, drain, or air-atomizing line on the
// Complete Systems Installation Plan.
type PlumbingLine struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	Fluid          string  `json:"fluid"`
	From           string  `json:"from"`
	To             string  `json:"to"`
	SizeIn         float64 `json:"size_in"`
	Material       string  `json:"material"`
	FlowGPM        float64 `json:"flow_gpm"`
	PressurePSI    float64 `json:"pressure_psi"`
	ValvesFittings string  `json:"valves_fittings"`
	Notes          string  `json:"notes"`
}

// PumpStationSpec describes the flooded-suction chemical recirculation pump station,
// strainer, backflow protection, and cool-down drain provisions.
type PumpStationSpec struct {
	Location          string  `json:"location"`
	PumpType          string  `json:"pump_type"`
	FlowGPM           float64 `json:"flow_gpm"`
	HeadFt            float64 `json:"head_ft"`
	HP                float64 `json:"hp"`
	Voltage           int     `json:"voltage"`
	SuctionSizeIn     float64 `json:"suction_size_in"`
	DischargeSizeIn   float64 `json:"discharge_size_in"`
	Strainer          string  `json:"strainer"`
	BackflowPreventer string  `json:"backflow_preventer"`
	DrainSpec         string  `json:"drain_spec"`
}

// BracingMember specifies a structural skid, column brace, pipe saddle, or vibration mount.
type BracingMember struct {
	ID         string  `json:"id"`
	Component  string  `json:"component"`
	MemberSpec string  `json:"member_spec"`
	Mounting   string  `json:"mounting"`
	LoadLbs    float64 `json:"load_lbs"`
	AnchorSpec string  `json:"anchor_spec"`
	Notes      string  `json:"notes"`
}

// AcousticZone describes one noise-attenuation treatment zone in the acoustic budget.
type AcousticZone struct {
	ID             string  `json:"id"`
	Location       string  `json:"location"`
	Treatment      string  `json:"treatment"`
	Material       string  `json:"material"`
	ThicknessIn    float64 `json:"thickness_in"`
	AttenuationDBA float64 `json:"attenuation_dba"`
	Notes          string  `json:"notes"`
}

// AcousticSummary summarizes bare vs treated near-field and property-line dBA performance.
type AcousticSummary struct {
	BareDBA3Ft        float64 `json:"bare_dba_3ft"`
	TreatedDBA3Ft     float64 `json:"treated_dba_3ft"`
	PropertyLineDBA   float64 `json:"property_line_dba"`
	PropertyLineFt    float64 `json:"property_line_ft"`
	TotalReductionDBA float64 `json:"total_reduction_dba"`
}

// DeluxeFabSpec summarizes the Custom Fabrication Deluxe hardware specification,
// cost delta, and Phase 13 optional energy/water-saving upgrades.
type DeluxeFabSpec struct {
	Enabled         bool     `json:"enabled"`
	VesselMaterial  string   `json:"vessel_material"`
	FlangeStandard  string   `json:"flange_standard"`
	SightGlasses    string   `json:"sight_glasses"`
	CleanoutHatches string   `json:"cleanout_hatches"`
	SkidFrame       string   `json:"skid_frame"`
	PanelStandard   string   `json:"panel_standard"`
	PumpUpgrade     string   `json:"pump_upgrade"`
	CostDeltaUSD    float64  `json:"cost_delta_usd"`
	Summary         string   `json:"summary"`
	Highlights      []string `json:"highlights"`

	// Phase 13 optional Deluxe Energy & Water Saving Upgrades
	CondensingReclaimer     bool     `json:"condensing_reclaimer,omitempty"`
	VariableThroatVFD       bool     `json:"variable_throat_vfd,omitempty"`
	Stage1HeatRecovery      bool     `json:"stage1_heat_recovery,omitempty"`
	EcoCostDeltaUSD         float64  `json:"eco_cost_delta_usd,omitempty"`
	WaterSavedPct           float64  `json:"water_saved_pct,omitempty"`
	WaterSavedGalPerFiring  float64  `json:"water_saved_gal_per_firing,omitempty"`
	EnergySavedPct          float64  `json:"energy_saved_pct,omitempty"`
	EnergySavedKWhPerFiring float64  `json:"energy_saved_kwh_per_firing,omitempty"`
	RecoveredHeatKW         float64  `json:"recovered_heat_kw,omitempty"`
	EcoSummary              string   `json:"eco_summary,omitempty"`
	EcoHighlights           []string `json:"eco_highlights,omitempty"`
}

// ShopPlan provides complete, self-contained in-browser assembly plans, numbered
// parts legend, flat-pattern cut sheets, clearances, and full systems schedules
// (electrical, plumbing, pump station, bracing, acoustics, and deluxe spec).
type ShopPlan struct {
	Title               string              `json:"title"`
	BuildPath           string              `json:"build_path"`
	IsDeluxe            bool                `json:"is_deluxe"`
	LocalDraftMode      string              `json:"local_draft_mode,omitempty"`
	SteamEjectorEnabled bool                `json:"steam_ejector_enabled,omitempty"`
	StageAlignment      string              `json:"stage_alignment"`
	PackedBedLayout     string              `json:"packed_bed_layout"`
	VenturiLayout       string              `json:"venturi_layout"`
	CompactFootprint    bool                `json:"compact_footprint"`
	OverallLengthIn     float64             `json:"overall_length_in"`
	OverallWidthIn      float64             `json:"overall_width_in"`
	OverallHeightIn     float64             `json:"overall_height_in"`
	Parts               []PlanPart          `json:"parts"`
	CutList             []CutSheetItem      `json:"cut_list"`
	Clearances          PlanClearances      `json:"clearances"`
	Electrical          []ElectricalCircuit `json:"electrical"`
	Plumbing            []PlumbingLine      `json:"plumbing"`
	PumpStation         PumpStationSpec     `json:"pump_station"`
	Bracing             []BracingMember     `json:"bracing"`
	Acoustics           []AcousticZone      `json:"acoustics"`
	AcousticSummary     AcousticSummary     `json:"acoustic_summary"`
	DeluxeSpec          *DeluxeFabSpec      `json:"deluxe_spec,omitempty"`
}

func frustumBoundingBoxIn(rIn, rOut, thetaRad float64) (w, h float64) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	angles := []float64{0, thetaRad}
	for _, card := range []float64{0.5 * math.Pi, math.Pi, 1.5 * math.Pi} {
		if card > 0 && card < thetaRad {
			angles = append(angles, card)
		}
	}
	for _, ang := range angles {
		ca, sa := math.Cos(ang), math.Sin(ang)
		for _, r := range []float64{rIn, rOut} {
			px, py := r*ca, r*sa
			if px < minX {
				minX = px
			}
			if py < minY {
				minY = py
			}
			if px > maxX {
				maxX = px
			}
			if py > maxY {
				maxY = py
			}
		}
	}
	return round2(maxX - minX), round2(maxY - minY)
}

func buildFrustumCutItem(partNum int, name, material string, d1In, d2In, axialIn float64, qty int, notes string) CutSheetItem {
	dev, err := cad.ComputeFrustumDevelopment(d1In, d2In, axialIn)
	if err != nil || dev.IsCylinder {
		circ := round2(math.Pi * math.Max(d1In, d2In))
		return CutSheetItem{
			PartNum:      partNum,
			Name:         name,
			Shape:        "cylinder",
			Material:     material,
			FlatWidthIn:  circ,
			FlatHeightIn: round2(axialIn),
			Qty:          qty,
			Notes:        notes,
		}
	}
	w, h := frustumBoundingBoxIn(dev.InnerRadius, dev.OuterRadius, dev.ThetaRad)
	return CutSheetItem{
		PartNum:       partNum,
		Name:          name,
		Shape:         "frustum",
		Material:      material,
		FlatWidthIn:   w,
		FlatHeightIn:  h,
		OuterRadiusIn: round2(dev.OuterRadius),
		InnerRadiusIn: round2(dev.InnerRadius),
		SweepDeg:      round2(dev.ThetaDeg),
		Qty:           qty,
		Notes:         notes,
	}
}

// ansi150FlangeNote returns standard ASME B16.5 Class 150 bolt-circle drilling specs.
func ansi150FlangeNote(nomIn float64) string {
	switch {
	case nomIn <= 6.5:
		return `ANSI/ASME B16.5 Class 150 6" lap-joint flange: 11.00" OD, 8× 3/4" studs on 9.50" BCD`
	case nomIn <= 8.5:
		return `ANSI/ASME B16.5 Class 150 8" lap-joint flange: 13.50" OD, 8× 3/4" studs on 11.75" BCD`
	case nomIn <= 10.5:
		return `ANSI/ASME B16.5 Class 150 10" lap-joint flange: 16.00" OD, 12× 7/8" studs on 14.25" BCD`
	case nomIn <= 12.5:
		return `ANSI/ASME B16.5 Class 150 12" lap-joint flange: 19.00" OD, 12× 7/8" studs on 17.00" BCD`
	case nomIn <= 14.5:
		return `ANSI/ASME B16.5 Class 150 14" lap-joint flange: 21.00" OD, 12× 1" studs on 18.75" BCD`
	case nomIn <= 16.5:
		return `ANSI/ASME B16.5 Class 150 16" lap-joint flange: 23.50" OD, 16× 1" studs on 21.25" BCD`
	case nomIn <= 18.5:
		return `ANSI/ASME B16.5 Class 150 18" lap-joint flange: 25.00" OD, 16× 1-1/8" studs on 22.75" BCD`
	default:
		return `ANSI/ASME B16.5 Class 150 24" lap-joint flange: 32.00" OD, 20× 1-1/4" studs on 29.50" BCD`
	}
}

// ansi150FlangeOD returns the ASME B16.5 Class 150 flange outside diameter in inches.
func ansi150FlangeOD(nomIn float64) float64 {
	switch {
	case nomIn <= 6.5:
		return 11.0
	case nomIn <= 8.5:
		return 13.5
	case nomIn <= 10.5:
		return 16.0
	case nomIn <= 12.5:
		return 19.0
	case nomIn <= 14.5:
		return 21.0
	case nomIn <= 16.5:
		return 23.5
	case nomIn <= 18.5:
		return 25.0
	case nomIn <= 20.5:
		return 27.5
	case nomIn <= 24.5:
		return 32.0
	default:
		return round1(nomIn + 7.5)
	}
}

func breakerAndGaugeForAmps(amps float64) (breaker int, gauge string) {
	switch {
	case amps <= 20.0:
		return 20, "12 AWG THHN/THWN-2 Copper"
	case amps <= 30.0:
		return 30, "10 AWG THHN/THWN-2 Copper"
	case amps <= 40.0:
		return 40, "8 AWG THHN/THWN-2 Copper"
	case amps <= 50.0:
		return 50, "6 AWG THHN/THWN-2 Copper"
	default:
		return 60, "4 AWG THHN/THWN-2 Copper"
	}
}

func buildElectricalSchedule(
	p StudioProject,
	design model.ScrubberDesign,
	urban UrbanAssessment,
	arch ArchitectureOptionsResult,
) []ElectricalCircuit {
	isDeluxe := p.IsDeluxe()
	mainBreaker, mainGauge := breakerAndGaugeForAmps(urban.Electrical.RequiredAmps)

	panelDest := "NEMA 4X Control/VFD Panel"
	if isDeluxe {
		panelDest = "UL 508A NEMA 4X Pre-Wired Panel (Through-Door Disconnect)"
	}

	// E2 blower destination adapts to eductor & blower drive selection.
	blowerTo := fmt.Sprintf("Stage 4 Radial Blower Motor (%.2g HP)", arch.BlowerDrive.MotorHP)
	blowerConduit := `3/4" LFMC Wet-Location (Shielded VFD Cable)`
	blowerWire := "14/4 AWG Shielded VFD Cable (3-Ph + Ground)"
	if arch.BlowerDrive.MotorHP >= 3.0 {
		blowerWire = "12/4 AWG Shielded VFD Cable (3-Ph + Ground)"
	}
	blowerAmps := round1(math.Max(2.2, arch.BlowerDrive.MotorHP*3.6))
	blowerVolts := 230
	if urban.Electrical.RequiredVoltage == 120 && arch.BlowerDrive.MotorHP <= 0.5 {
		blowerVolts = 230 // GS21-10P5 outputs 230V 3-phase to 3-phase motor from 120V 1-ph input
	}
	blowerNotes := "Run shielded VFD motor cable in dedicated metallic/LFMC conduit >=12\" away from sensor & 24VDC control wiring to prevent EMI."
	switch {
	case arch.Eductor.Enabled && p.BlowerDriveOption == BlowerDriveIndustrialRadialVFD:
		blowerTo = fmt.Sprintf("Dry-Air Eductor Motive Blower Motor (%.2g HP)", arch.Eductor.DryFanMotorHP)
		blowerAmps = round1(math.Max(2.8, arch.Eductor.DryFanMotorHP*3.6))
		blowerNotes = "Dry ambient motive fan mounted upstream of vertical downflow eductor; shielded VFD cable separated >=12\" from sensor leads."
	case p.BlowerDriveOption == BlowerDriveDualBypassBlowers:
		blowerTo = "2× Parallel Ametek Lamb 3-Stage Tangential Bypass Blowers"
		blowerVolts = arch.BlowerDrive.Voltage
		blowerAmps = round1(math.Max(8.0, arch.BlowerDrive.MotorHP*6.5))
		blowerWire = "12 AWG THHN/THWN-2 Copper (Staged Relay Feed)"
		blowerNotes = "Dedicated staging relay feed to dual parallel tangential bypass blowers; keep tangential motor cooling vents clear."
	case p.BlowerDriveOption == BlowerDriveRegenSideChannel:
		blowerTo = fmt.Sprintf("Side-Channel Regenerative Ring Blower (%.2g HP)", arch.BlowerDrive.MotorHP)
		blowerNotes = "High-static ring blower feed; maintain bleed/relief blast gate and motor overload protection."
	case p.BlowerDriveOption == BlowerDriveCompressorAirAtomize:
		blowerTo = fmt.Sprintf("Medium-Pressure Air-Atomized Radial Blower (%.2g HP)", arch.BlowerDrive.MotorHP)
		blowerNotes = "Lower-static radial blower enabled by 25 um twin-fluid air-atomizing nozzle; shielded VFD cable in dedicated LFMC."
	case p.BlowerDriveOption == BlowerDriveCompressorEductor:
		blowerTo = "Existing Shop Air Compressor Feed (No Stage 4 Blower Motor)"
		blowerVolts = arch.BlowerDrive.Voltage
		blowerNotes = "Compressed-air amplifier has no electric fan motor; verify existing studio shop compressor circuit capacity."
	}

	safetyFrom := "200 F Snap-Disc + 0.25 GPM Flow Switch + NEMA 4X E-Stop"
	safetyTo := "24VDC Latching Safety Relay -> Belimo Fail-Open Bypass Damper & VFD STO"
	safetyNotes := "De-energize-to-trip hardwired NFPA 86 safety loop; overrides software/PLC so power loss, >200 F quench outlet, or <0.25 GPM water loss immediately springs bypass damper OPEN and stops VFD."
	if isDeluxe {
		safetyFrom = "FM/UL 1/16-DIN 200 F High-Limit + 0.25 GPM Flow Transmitter + E-Stop"
		safetyTo = "UL 508A Latching Safety Relay -> Belimo Fail-Open Damper, VFD STO & Stack Alarm Horn"
		safetyNotes = "Independent FM/UL 1/16-DIN 200 F high-limit controller + hardwired 24VDC de-energize-to-trip safety relay drives Belimo spring-return fail-open bypass damper, VFD Safe-Torque-Off, and pilot stack light/alarm horn."
	}

	pumpAmps := 3.0
	pumpTo := "Stage 3 Magnetic-Drive Recirculation Pump"
	pumpNotes := "Dedicated 120V GFCI circuit for flooded-suction sealless mag-drive pump; mount weatherproof GFCI receptacle or hardwire in LFMC."
	if !p.SharedBaffledSump {
		pumpAmps = 5.5
		pumpTo = "Dual Sump Magnetic-Drive Recirculation Pumps (2×)"
		pumpNotes = "Feeds both dedicated ash-settling and absorber mag-drive pumps on 120V GFCI protection."
	}

	sensorLabel := "Sensor & Modbus RS-485 Shielded Instrumentation Loop"
	sensorFrom := "NEMA 4X Control Panel (24VDC / RS-485)"
	sensorTo := "Kiln Flue dP Tap, Quench Type-K TC, Pt100 RTD & Sump pH Probe"
	sensorNotes := "Physically separated >=12\" from VFD power conduit (NEC 725.136); ground RS-485 shield at control panel end only."
	if p.ControlTier == ControlTierElectromechanicalOnly {
		sensorLabel = "Hardwired Status & Analog Instrumentation Wiring"
		sensorTo = "Quench Stem/High-Limit Well, Magnehelic Gauge Port & Manual VFD Potentiometer"
		sensorNotes = "Electromechanical-only tier: zero software required; shielded 3-conductor VFD speed potentiometer lead routed separately from motor power."
	}
	if isDeluxe {
		sensorTo = "Dual Type-K Thermowells, Inline Flow Transmitter, Flow-Through pH Probe, dP Transmitter & Pilot Stack Light"
		sensorNotes = "UL 508A segregated low-voltage wireway; shielded twisted-pair RS-485 & Type-K extension cable in dedicated instrumentation conduit >=12\" from VFD motor leads."
	}

	circuits := []ElectricalCircuit{
		{
			ID:           "E1",
			Label:        "Main Studio Feeder -> Sight Disconnect -> Control/VFD Panel",
			From:         "Main Studio Breaker Panel",
			To:           panelDest,
			Voltage:      urban.Electrical.RequiredVoltage,
			Amps:         urban.Electrical.RequiredAmps,
			BreakerAmps:  mainBreaker,
			WireGauge:    mainGauge,
			Conduit:      `3/4" LFMC / EMT Wet-Location Conduit`,
			NECRef:       "NEC 430.122(A) / 430.102 / 210.8(F)",
			IsSafetyLoop: false,
			Notes:        fmt.Sprintf("Sized at >=125%% VFD rated input (%s) via lockable sight disconnect <=50 ft, HF-GFCI breaker & 3%% input line reactor.", urban.Electrical.CircuitLabel),
		},
		{
			ID:           "E2",
			Label:        "VFD / Drive Output -> Blower Motor",
			From:         panelDest,
			To:           blowerTo,
			Voltage:      blowerVolts,
			Amps:         blowerAmps,
			BreakerAmps:  mainBreaker,
			WireGauge:    blowerWire,
			Conduit:      blowerConduit,
			NECRef:       "NEC 430.22 / 400.12",
			IsSafetyLoop: false,
			Notes:        blowerNotes,
		},
		{
			ID:           "E3",
			Label:        "24VDC Hardwired NFPA 86 Safety Interlock Loop",
			From:         safetyFrom,
			To:           safetyTo,
			Voltage:      24,
			Amps:         1.5,
			BreakerAmps:  5,
			WireGauge:    "18 AWG Red Tray Cable (Class 1/2 Control)",
			Conduit:      `1/2" LFMC Dedicated Safety Conduit`,
			NECRef:       "NFPA 86 / NEC 725",
			IsSafetyLoop: true,
			Notes:        safetyNotes,
		},
		{
			ID:           "E4",
			Label:        "Recirculation Pump 120V GFCI Circuit",
			From:         panelDest,
			To:           pumpTo,
			Voltage:      120,
			Amps:         pumpAmps,
			BreakerAmps:  15,
			WireGauge:    "12 AWG THHN/THWN-2 Copper",
			Conduit:      `1/2" LFMC Wet-Location`,
			NECRef:       "NEC 210.8(F) / 400.12",
			IsSafetyLoop: false,
			Notes:        pumpNotes,
		},
		{
			ID:           "E5",
			Label:        sensorLabel,
			From:         sensorFrom,
			To:           sensorTo,
			Voltage:      24,
			Amps:         0.5,
			BreakerAmps:  5,
			WireGauge:    "22 AWG Shielded Twisted Pair + Type-K Ext",
			Conduit:      `1/2" Dedicated Instrumentation Conduit`,
			NECRef:       "NEC 725.136 / EIA-485",
			IsSafetyLoop: false,
			Notes:        sensorNotes,
		},
	}

	if p.IncludePlumeDiluter {
		dilAmps := round1(diluterAmps120V(design))
		dilBreaker := 15
		if dilAmps > 12.0 {
			dilBreaker = 20
		}
		cfm, coneIn := RecommendedDilutionCFM(design)
		circuits = append(circuits, ElectricalCircuit{
			ID:           "E6",
			Label:        fmt.Sprintf("Plume Dilution Squirrel-Cage Fan (%.0f CFM, 120V)", cfm),
			From:         panelDest,
			To:           fmt.Sprintf(`Downstream %.0f CFM Plume Diluter Fan + %.0f" HV Cone`, cfm, coneIn),
			Voltage:      120,
			Amps:         dilAmps,
			BreakerAmps:  dilBreaker,
			WireGauge:    "12 AWG THHN/THWN-2 Copper",
			Conduit:      `1/2" LFMC Wet-Location`,
			NECRef:       "NEC 210.8(F) / 430.22",
			IsSafetyLoop: false,
			Notes:        "Dedicated 120V GFCI branch for downstream squirrel-cage plume dilution blower (mounted strictly downstream of Stage 4 blower).",
		})
	}

	return circuits
}

func buildPlumbingSchedule(
	p StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	arch ArchitectureOptionsResult,
) ([]PlumbingLine, PumpStationSpec) {
	isDeluxe := p.IsDeluxe()

	recircGPM := design.PackedBed.RecirculationFlowGPM
	if recircGPM <= 0 {
		recircGPM = math.Max(4.0, design.Quench.TotalSprayWaterGPM*3.0)
	}
	recircGPM = round2(recircGPM)
	quenchGPM := round2(math.Max(0.4, design.Quench.TotalSprayWaterGPM))
	venturiGPM := round2(math.Max(0.2, design.Venturi.LiquidFlowGPM))
	makeupGPM := round2(math.Max(0.5, design.Quench.EvaporatedWaterGPM*1.5))

	suctionSizeIn := 1.0
	if recircGPM >= 6.0 || isDeluxe {
		suctionSizeIn = 1.5
	}
	dischargeSizeIn := 0.75
	if recircGPM >= 12.0 {
		dischargeSizeIn = 1.0
	}

	quenchHoseMat := `Sch 80 CPVC / 200 F Braided EPDM Hose`
	if isDeluxe {
		quenchHoseMat = `3/4" PTFE-Lined SS-Braided Chemical Hose + 316L Quick-Disconnects`
	}

	p5To := "Stage 3 Packed-Bed Top Liquid Distributor"
	p5Valves := "True-Union Ball Valve, Throttle Valve, Flow Meter"
	p5Notes := fmt.Sprintf("Counter-current absorber wetting header (%.1f GPM across column cross-section); maintain >=6\" hose bend radius.", recircGPM)
	if arch.ParallelBed.Enabled {
		p5To = fmt.Sprintf("Stage 3 Pressurized %d-Port Parallel-Pipe Spray Manifold", arch.ParallelBed.PipeCount)
		p5Valves = fmt.Sprintf("True-Union Ball Valve + Calibrated %d-Port Equal-Orifice Header", arch.ParallelBed.PipeCount)
		p5Notes = fmt.Sprintf("MANDATORY balanced distribution: splits %.1f GPM evenly (~%.2f GPM/pipe) across all %d × %.0f\" risers to prevent dry channeling.",
			recircGPM, recircGPM/float64(maxInt(arch.ParallelBed.PipeCount, 1)), arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn)
	}

	totalPumpGPM := round2(recircGPM + quenchGPM + venturiGPM)
	evapGPM := round2(design.Quench.EvaporatedWaterGPM)
	gravityReturnGPM := round2(math.Max(recircGPM, totalPumpGPM-evapGPM))

	p6Valves := `2" Cleanout Tee, >=26"-30" Vertical Barometric Water-Seal Leg + Gravity Slope (>= 1/4" per ft)`
	p6Notes := fmt.Sprintf("NEGATIVE-PRESSURE BAROMETRIC SEAL: Stage 3 operates under -%.1f inWC ID blower suction — install a >=%.0f\" vertical barometric water-seal leg (submerged into sump liquid below) so blower vacuum cannot pull brine up into the column or suck air backward through the drain.",
		design.Blower.TotalStaticPressureInWC, math.Max(26.0, math.Ceil(design.Blower.TotalStaticPressureInWC+6.0)))
	if arch.Eductor.Enabled {
		sealIn := math.Max(8.0, math.Ceil(arch.Eductor.PositivePressureInWC+4.0))
		p6Valves = fmt.Sprintf(`2" Deep U-Trap Liquid Seal (>= %.0f" Seal Leg) + Cleanout Plug`, sealIn)
		p6Notes = fmt.Sprintf("EDUCTOR POSITIVE-PRESSURE SEAL: >= %.0f\" U-trap water leg exceeds +%.2f inWC Stage 3 positive pressure so wet exhaust/acid mist never blows back into the vented sump.", sealIn, arch.Eductor.PositivePressureInWC)
	}

	p2Strainer := "50-Mesh Y-Strainer"
	if isDeluxe {
		p2Strainer = "20-40 Mesh True-Union Y-Strainer (prevents hot 184 F suction cavitation)"
	}
	pumpHeadPSI := 15.2
	if isDeluxe {
		pumpHeadPSI = 19.5
	}

	lines := []PlumbingLine{
		{
			ID:             "P1",
			Label:          "Potable Makeup Supply (IPC 608 Backflow Protected)",
			Fluid:          "Potable Cold Water",
			From:           "Studio Water Spigot / Supply",
			To:             "IPC 608 ASSE 1013 RPZ / >=1\" Air Gap -> Sump Float Valve",
			SizeIn:         0.75,
			Material:       `3/4" Copper/PEX + Lead-Free Bronze ASSE 1013 RPZ (or >=1" Air Gap)`,
			FlowGPM:        makeupGPM,
			PressurePSI:    55.0,
			ValvesFittings: "Shutoff Ball Valve, ASSE 1013 RPZ / Air-Gap Break Tank, Float Valve",
			Notes:          "MANDATORY IPC Section 608 high-hazard potable isolation; never connect a hose directly to pressurized chemical lines.",
		},
		{
			ID:             "P2",
			Label:          "Sump Flooded Suction -> Recirculation Pump",
			Fluid:          "Buffered NaHCO3 Scrubber Liquor (pH 7.5-8.5)",
			From:           "Baffled Sump Clean Compartment (Flooded Suction)",
			To:             "Recirculation Pump Suction Port",
			SizeIn:         suctionSizeIn,
			Material:       `Sch 80 CPVC / Wire-Reinforced EPDM Chemical Suction Hose`,
			FlowGPM:        totalPumpGPM,
			PressurePSI:    2.0,
			ValvesFittings: "Bulkhead Fitting, True-Union Isolation Ball Valve, " + p2Strainer,
			Notes:          fmt.Sprintf("Total shared suction = %.2f GPM (Stage 1 %.2f + Stage 2 %.2f + Stage 3 %.2f GPM). Mount pump below sump minimum liquid level for positive flooded suction (mag-drive pumps must NEVER run dry).", totalPumpGPM, quenchGPM, venturiGPM, recircGPM),
		},
		{
			ID:             "P3",
			Label:          "Pump Discharge (or Dedicated Booster) -> Flow Switch -> Stage 1 Quench Manifold",
			Fluid:          "Pressurized Quench Spray Liquor",
			From:           "Recirculation Pump Discharge Manifold (or 60-100 PSI Fine-Mist Booster)",
			To:             fmt.Sprintf("Stage 1 Quench Spray Nozzles (%d×)", maxInt(design.Quench.NozzleCount, 2)),
			SizeIn:         0.75,
			Material:       quenchHoseMat,
			FlowGPM:        quenchGPM,
			PressurePSI:    pumpHeadPSI,
			ValvesFittings: "Check Valve, 0.25 GPM Safety Flow Switch, Pressure Gauge, Throttle Valve",
			Notes:          fmt.Sprintf("Standard recirc header runs %.1f PSI (add dedicated 60-100 psig booster or twin-fluid air-atomize nozzle for <=200 um droplets in short towers); 0.25 GPM flow switch wired directly into E3 hardwired NFPA 86 safety loop.", pumpHeadPSI),
		},
		{
			ID:             "P4",
			Label:          "Pump Discharge Branch -> Stage 2 Venturi Spray Injection",
			Fluid:          "Venturi Throat Atomizing Liquor",
			From:           "Recirculation Pump Discharge Branch",
			To:             "Stage 2 Venturi Throat Opposed Spray Ports",
			SizeIn:         0.5,
			Material:       `1/2" Sch 80 CPVC / 316L SS Manifold + Braided Hose`,
			FlowGPM:        venturiGPM,
			PressurePSI:    pumpHeadPSI,
			ValvesFittings: "Needle/Ball Throttle Valve, 316L/C-276 Injection Nozzles, Union",
			Notes:          fmt.Sprintf("Delivers %.2f GPM to Stage 2 throat at L/G=%.4f (%.1f PSI manifold feed; gas-stream shear atomizes droplets at throat).", venturiGPM, design.Venturi.LGRatio, pumpHeadPSI),
		},
	}

	if p.BlowerDriveOption == BlowerDriveCompressorAirAtomize {
		lines = append(lines, PlumbingLine{
			ID:             "A1",
			Label:          "Shop Compressor Air Line -> Twin-Fluid Venturi Nozzle",
			Fluid:          fmt.Sprintf("Clean Regulated Shop Air (%.1f SCFM @ 40 PSI)", arch.Compressor.AtomizeSCFMNeeded),
			From:           fmt.Sprintf("Studio Shop Air Compressor (%.0f SCFM)", arch.Compressor.AvailableSCFM),
			To:             "Stage 2 Twin-Fluid 25 um Air-Atomizing Venturi Nozzle",
			SizeIn:         0.375,
			Material:       `3/8" Industrial Rubber/PU Air Hose + 316L Swivel Fitting`,
			FlowGPM:        0.0,
			PressurePSI:    40.0,
			ValvesFittings: "Coalescing Filter/Moisture Trap, 40 PSI Regulator, Gauge, Check Valve",
			Notes:          fmt.Sprintf("Produces 25 um SMD droplets at 48 m/s throat velocity (active twin-fluid d50=%.2f um parity), cutting total static dP by %.1f inWC.", activeTwinFluidD50Microns(design), arch.Compressor.AtomizeDPSavedInWC),
		})
	}

	lines = append(lines,
		PlumbingLine{
			ID:             "P5",
			Label:          "Pump Discharge Riser -> Stage 3 Top Liquid Distributor",
			Fluid:          "Counter-Current Absorber Liquor",
			From:           "Recirculation Pump Discharge Manifold",
			To:             p5To,
			SizeIn:         dischargeSizeIn,
			Material:       `Sch 80 CPVC Vertical Riser + Pipe Clamps`,
			FlowGPM:        recircGPM,
			PressurePSI:    15.0,
			ValvesFittings: p5Valves,
			Notes:          p5Notes,
		},
		PlumbingLine{
			ID:             "P6",
			Label:          "Gravity Return Drain (Stage 1 + Stage 3 + Stack Condensate)",
			Fluid:          "Hot Scrubbed Liquor & Ash Slurry (~180 F)",
			From:           "Stage 1 Quench Base, Stage 3 Sump Cone & Stack Drain Loop",
			To:             "Baffled Sump Settling Chamber",
			SizeIn:         2.0,
			Material:       `2" Sch 80 CPVC Gravity Drain Pipe (200 F Rated)`,
			FlowGPM:        gravityReturnGPM,
			PressurePSI:    1.0,
			ValvesFittings: p6Valves,
			Notes:          p6Notes,
		},
		PlumbingLine{
			ID:             "P7",
			Label:          `Sump Bottom Cool-Down Drain (2" Ball Valve -> Neutralized Floor Drain)`,
			Fluid:          "Cooled Neutralized Brine Effluent (<=140 F, pH 6.5-8.5)",
			From:           "Sump Bottom Drain Bulkhead",
			To:             "Ash Filter Bag -> Sanitary Floor Drain",
			SizeIn:         2.0,
			Material:       `2" Sch 80 CPVC + Full-Port True-Union Ball Valve`,
			FlowGPM:        12.0,
			PressurePSI:    2.0,
			ValvesFittings: `2" Full-Port Ball Valve, Thermometer Well (IPC 702.5 <=140 F), pH Test Port`,
			Notes:          "MANDATORY IPC 702.5 & 40 CFR 403.5: cool liquor <=140 F, strain wood ash/sludge, and verify pH 6.5-8.5 (>=5.0 federal floor) before opening drain valve.",
		},
	)

	pumpType := "Seal-Less Mag-Drive Recirculation Pump (PP/EPDM, March / Little Giant TE-MD)"
	headFt := 35.0
	pumpHP := 0.25
	if recircGPM > 10.0 {
		pumpHP = 0.5
	}
	strainer := `50-Mesh Inline CPVC/PP Y-Strainer (Clean Before Every Firing)`
	drainSpec := `2" Sump Drain Ball Valve -> Cool <=140 F (IPC 702.5) + Verify pH 6.5-8.5 (40 CFR 403.5)`
	if isDeluxe {
		pumpType = "Heavy-Duty Magnetic-Drive Sealless Chemical Pump (PVDF/Ceramic, Iwaki MD / March)"
		headFt = 45.0
		if pumpHP < 0.5 {
			pumpHP = 0.5
		}
		strainer = `316L/CPVC 20-40 Mesh True-Union Y-Strainer w/ Flush Ball Valve & Pressure Gauge`
		drainSpec = `Sloped-Bottom Sump 2" Full-Port True-Union Drain Ball Valve + Removable Ash Basket -> IPC 702.5 <=140 F & pH 6.5-8.5 Check`
	}
	location := "Adjacent to Atmospheric Baffled Sump (Flooded Suction, 24\" Front Service Access)"
	if !p.SharedBaffledSump {
		location = "Dual Flooded-Suction Pump Pads (Ash-Settling Sump + Absorber Sump, 24\" Service Access)"
	}

	pumpStation := PumpStationSpec{
		Location:          location,
		PumpType:          pumpType,
		FlowGPM:           totalPumpGPM,
		HeadFt:            headFt,
		HP:                pumpHP,
		Voltage:           120,
		SuctionSizeIn:     suctionSizeIn,
		DischargeSizeIn:   dischargeSizeIn,
		Strainer:          strainer,
		BackflowPreventer: `IPC Section 608 Certified ASSE 1013 RPZ or Engineered >=1" Vertical Air-Gap Break Tank`,
		DrainSpec:         drainSpec,
	}

	return lines, pumpStation
}

func buildBracingSchedule(
	p StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	arch ArchitectureOptionsResult,
) []BracingMember {
	isDeluxe := p.IsDeluxe()

	quenchWetLbs := round1(cots.QuenchVessel.WeightKg/model.KgPerLb + 38.0)

	// Extract actual packing media & demister total line weights from COTS BOM.
	packingKg := 0.0
	demisterKg := 4.0
	for _, it := range cots.BOM.Items {
		if it.Unit == "cu.ft." {
			packingKg += it.WeightKg
		}
		if it.Category == "Stage 3: Packed-Bed Column" && it.Unit == "assy" {
			demisterKg = it.WeightKg
		}
	}
	bedHoldupWaterLbs := round1(math.Max(18.0, design.PackedBed.PackingVolume.CuFt()*6.5))
	colWetLbs := round1((cots.PackedColumn.WeightKg+packingKg+demisterKg)/model.KgPerLb + bedHoldupWaterLbs)
	packingLbs := round1(packingKg / model.KgPerLb)

	sumpGal := workingSumpVolumeGal(cots, design)
	if !p.SharedBaffledSump {
		sumpGal *= 1.6
	}
	sumpWetLbs := round1(cots.SumpTank.WeightKg/model.KgPerLb + sumpGal*8.6)
	blowerLbs := round1(cots.BlowerVFD.WeightKg / model.KgPerLb)
	if cots.Silencer != nil {
		blowerLbs = round1(blowerLbs + cots.Silencer.WeightKg/model.KgPerLb)
	}
	skidSelfLbs := 125.0
	if isDeluxe {
		skidSelfLbs = 235.0
	}
	totalSkidLbs := round1(quenchWetLbs + colWetLbs + sumpWetLbs + blowerLbs + skidSelfLbs)

	b1Spec := `1-5/8" Galvanized Unistrut / Pressure-Treated Heavy Base Skid Frame`
	b1Anchor := `4× 1/2"×4" 316 SS Concrete Wedge Anchors (min 3" embedment)`
	b1Notes := fmt.Sprintf("Supports total wet operating load (~%.0f lbs including %.0f-gal filled sump); level within 1/8\"/ft for gravity drain slope.", totalSkidLbs, sumpGal)
	if isDeluxe {
		b1Spec = `Welded 2"×2"×3/16" Galvanized/Powder-Coated Structural Steel Forklift Skid`
		b1Anchor = `6× 1/2"×4" 316 SS Concrete Wedge Anchors (40 ft-lb torque, min 3" embedment)`
		b1Notes = fmt.Sprintf("Welded forklift-pocket steel skid rated for %.0f lbs wet operating mass (incl. %.0f-gal sloped sump) with perimeter gussets & leveling pads.", totalSkidLbs, sumpGal)
	}

	b3Spec := `1-5/8" Galvanized Unistrut Vertical Column Cage + Diagonal Wind/Seismic Struts`
	if isDeluxe {
		b3Spec = `Welded 2"×2"×3/16" Steel Column Cage + 1-1/2" Diagonal K-Bracing & FRP Gusset Ring Clamps`
	}
	b3Notes := fmt.Sprintf("Carries %.0f lbs wet column load (includes %.0f lbs %s media + %.0f lbs irrigated liquid holdup) and resists 90 mph wind / 0.3g lateral overturning.",
		colWetLbs, packingLbs, arch.PackingMedia.Label, bedHoldupWaterLbs)
	if p.PackingMediaOption == PackingMediaDIYCeramicRings {
		b3Notes += " HEAVY CERAMIC MEDIA REINFORCEMENT: Stoneware rings (~46 lb/cu.ft.) require heavy FRP bar grating or 316L cross-bars under the bed!"
	}

	b4Comp := "Overhead Stage 3 -> Stage 4 Duct Hanger & Pipe Saddles"
	b4Spec := `3/8" 316 SS Threaded Rod Clevis Hangers + Cushioned Pipe Saddles (<= 4 ft spacing)`
	b4Load := round1(math.Max(35.0, design.Envelope.DuctDiameter.Inches()*6.5))
	b4Notes := "Prevents dead-weight duct sag on Stage 3 top flange and Stage 4 blower inlet; slope wet horizontal runs >=1/4\" per ft toward low-point drain."
	if arch.Eductor.Enabled {
		b4Comp = "Downflow Eductor Tee & 90° Cyclonic Sweep Elbow Support Frame"
		b4Spec = `Welded 2"×2" Steel Cantilever Cradle + Cushioned Pipe Saddles & Thrust Tie-Rod`
		b4Load = round1(math.Max(65.0, arch.Eductor.CyclonicSweepElbowIn*9.5))
		b4Notes = fmt.Sprintf("Supports overhead dry motive fan & vertical downflow eductor thrust into %.0f\" 90° sweep elbow (+%.2f inWC positive pressure).", arch.Eductor.CyclonicSweepElbowIn, arch.Eductor.PositivePressureInWC)
	}

	return []BracingMember{
		{
			ID:         "B1",
			Component:  "Base Structural Skid Frame",
			MemberSpec: b1Spec,
			Mounting:   "Level 4\"+ Concrete Pad / Masonry Slab",
			LoadLbs:    totalSkidLbs,
			AnchorSpec: b1Anchor,
			Notes:      b1Notes,
		},
		{
			ID:         "B2",
			Component:  "Stage 1 Quench Vertical Standoffs & Thermal Expansion Saddle",
			MemberSpec: `1-1/2"×1-1/2"×1/8" 316L/Steel Vertical Standoffs + Slotted Thermal Expansion Saddle`,
			Mounting:   "Bolted to Base Skid with Slotted Top Collar",
			LoadLbs:    quenchWetLbs,
			AnchorSpec: `4× 3/8" 316 SS Hex Bolts w/ Locking Nuts + Ceramic Thermal Isolation Pads`,
			Notes:      "Slotted upper guide allows ~3/16\" vertical thermal growth from 2400 F flue inlet without stressing Stage 2 venturi flange.",
		},
		{
			ID:         "B3",
			Component:  "Stage 3 Packed-Bed Tall Column Frame & Diagonal K-Bracing",
			MemberSpec: b3Spec,
			Mounting:   "4-Post Vertical Cage Bolted/Welded to Skid + Mid-Height Collar",
			LoadLbs:    colWetLbs,
			AnchorSpec: `4× 1/2" 316 SS Baseplate Bolts + Rubber-Lined Column Clamps`,
			Notes:      b3Notes,
		},
		{
			ID:         "B4",
			Component:  b4Comp,
			MemberSpec: b4Spec,
			Mounting:   "Tied to Stage 3 Column Frame / Skid Uprights",
			LoadLbs:    b4Load,
			AnchorSpec: `3/8" 316 SS Double-Nutted Hanger Rods & Beam Clamps`,
			Notes:      b4Notes,
		},
		{
			ID:         "B5",
			Component:  "Blower Pedestal & Flexible PTFE/EPDM Vibration Isolation Boots",
			MemberSpec: `Heavy Steel Blower Pedestal + 1" Ribbed Neoprene Isolation Pads + Flexible PTFE/EPDM Inlet & Outlet Boots`,
			Mounting:   "Elastomeric Isolation Mounts on Skid End-Bay",
			LoadLbs:    blowerLbs,
			AnchorSpec: `4× 3/8" Isolated Stud Mounts + 316 SS Constant-Torque Band Clamps`,
			Notes:      "Decouples rotating blower vibration from rigid CPVC/FRP column and exhaust stack, eliminating structure-borne hum and flange fatigue.",
		},
	}
}

func buildAcousticsSchedule(
	p StudioProject,
	design model.ScrubberDesign,
	urban UrbanAssessment,
	arch ArchitectureOptionsResult,
) ([]AcousticZone, AcousticSummary) {
	isDeluxe := p.IsDeluxe()
	silencingActive := p.IncludeSilencer || isDeluxe

	blowerLoc := "Stage 4 Induced-Draft Blower Housing"
	if arch.Eductor.Enabled {
		blowerLoc = "Upstream Dry-Air Motive Eductor Blower Housing"
	}

	s1Atten, s2Atten, s3Atten := 0.0, 0.0, 0.0
	s1Note := "Not enabled — enable Acoustic Silencer (-13 dBA) or Custom Fab Deluxe (-22 dBA) to cut blower casing noise."
	s2Note := "Not enabled — add inline absorptive discharge silencer above blower vibration boot to cut stack roar."
	s3Note := "Recommended upgrade — 1\" ribbed neoprene mounts and flexible PTFE/EPDM duct boots break structure-borne hum."

	if isDeluxe {
		s1Atten, s2Atten, s3Atten = 10.0, 9.0, 3.0
		s1Note = "ACTIVE (Deluxe): 2\" Rockwool Safe'n'Sound mineral wool behind perforated metal inner liner with baffled TEFC motor cooling-air intake & exhaust labyrinth (prevents motor overheating while blocking direct line-of-sight noise)."
		s2Note = "ACTIVE (Deluxe): Tuned absorptive perforated-core vertical stack silencer mounted above discharge vibration boot and below diluter tee."
		s3Note = "ACTIVE (Deluxe): 1\" ribbed neoprene isolation pads under blower pedestal & pump baseplate plus flexible PTFE/EPDM inlet/outlet boots break structure-borne resonance."
	} else if p.IncludeSilencer {
		s1Atten, s2Atten, s3Atten = 6.0, 5.0, 2.0
		s1Note = "ACTIVE: Insulated 2\" Rockwool Safe'n'Sound blower acoustic baffle box with unobstructed baffled motor-cooling vents."
		s2Note = "ACTIVE: Inline absorptive discharge silencer installed on vertical exhaust stack above flexible vibration boot."
		s3Note = "ACTIVE: Elastomeric vibration isolation pads under blower & pump plus flexible EPDM duct collars."
	}

	zones := []AcousticZone{
		{
			ID:             "S1",
			Location:       blowerLoc,
			Treatment:      "Acoustic Blower Baffle Box w/ Baffled Cooling-Air Labyrinth",
			Material:       `2" Rockwool Safe'n'Sound Mineral Wool + Perforated Metal Liner`,
			ThicknessIn:    2.0,
			AttenuationDBA: s1Atten,
			Notes:          s1Note,
		},
		{
			ID:             "S2",
			Location:       "Vertical Discharge Stack (Above Vibration Boot, Below Diluter Tee)",
			Treatment:      "Inline Tuned Absorptive Discharge Stack Silencer",
			Material:       "Moisture-Resistant Acoustic Pack + Perforated Inner Core (FRP/316L/Galv)",
			ThicknessIn:    2.0,
			AttenuationDBA: s2Atten,
			Notes:          s2Note,
		},
		{
			ID:             "S3",
			Location:       "Blower Pedestal, Pump Baseplate & Blower Duct Transitions",
			Treatment:      "Elastomeric Vibration Isolation Pads & Flexible Duct Boots",
			Material:       `1" Ribbed Neoprene Pads + Flexible PTFE/EPDM Vibration Boots`,
			ThicknessIn:    1.0,
			AttenuationDBA: s3Atten,
			Notes:          s3Note,
		},
	}

	if p.IncludePlumeDiluter {
		cfm, _ := RecommendedDilutionCFM(design)
		zones = append(zones, AcousticZone{
			ID:             "S4",
			Location:       fmt.Sprintf("Plume Diluter Squirrel-Cage Fan Inlet (%.0f CFM)", cfm),
			Treatment:      "Diluter Fan Inlet Acoustic Duct Liner",
			Material:       `1" Closed-Cell Foil-Faced Acoustic Duct Liner`,
			ThicknessIn:    1.0,
			AttenuationDBA: 4.0,
			Notes:          "Attenuates auxiliary 120V squirrel-cage intake roar by ~4 dBA locally so enabling plume dilution adds negligible net property-line dBA.",
		})
	}

	treated3Ft := urban.Noise.BareDBA3Ft
	if silencingActive {
		treated3Ft = urban.Noise.SilencedDBA3Ft
	}
	totalRed := round1(math.Max(0.0, urban.Noise.BareDBA3Ft-treated3Ft))

	summary := AcousticSummary{
		BareDBA3Ft:        urban.Noise.BareDBA3Ft,
		TreatedDBA3Ft:     treated3Ft,
		PropertyLineDBA:   urban.Noise.PropertyLineDBA,
		PropertyLineFt:    urban.Noise.PropertyLineFt,
		TotalReductionDBA: totalRed,
	}
	return zones, summary
}

func buildDeluxeSpec(
	p StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	urban UrbanAssessment,
) *DeluxeFabSpec {
	isDeluxe := p.IsDeluxe()

	// Compute exact cost delta & acoustic result comparing custom_fab_deluxe vs equivalent cots_no_weld configuration.
	var deluxeTotal, cotsTotal float64
	deluxePropDBA := urban.Noise.PropertyLineDBA
	deluxePropFt := urban.Noise.PropertyLineFt
	if isDeluxe {
		deluxeTotal = cots.EstimatedTotalUSD
		cotsProj := p
		cotsProj.BuildPath = BuildPathCOTSNoWeld
		cotsProj.ZeroElectricity = false
		cotsProj.DeluxeCondensingReclaimer = false
		cotsProj.DeluxeVariableThroatVFD = false
		cotsProj.DeluxeStage1HeatRecovery = false
		cotsTotal = BuildCOTSSummary(cotsProj, design).EstimatedTotalUSD
	} else {
		cotsProj := p
		cotsProj.BuildPath = BuildPathCOTSNoWeld
		cotsProj.ZeroElectricity = false
		cotsProj.DeluxeCondensingReclaimer = false
		cotsProj.DeluxeVariableThroatVFD = false
		cotsProj.DeluxeStage1HeatRecovery = false
		if p.BuildPath == BuildPathCOTSNoWeld && !p.ZeroElectricity {
			cotsTotal = cots.EstimatedTotalUSD
		} else {
			cotsTotal = BuildCOTSSummary(cotsProj, design).EstimatedTotalUSD
		}
		dlxProj := p
		dlxProj.BuildPath = BuildPathCustomFabDeluxe
		dlxProj.ZeroElectricity = false
		if dlxProj.PackedBedLayout == PackedBedLayoutParallelManifold {
			dlxProj.PackedBedLayout = PackedBedLayoutSingleColumn
		}
		dlxCOTS := BuildCOTSSummary(dlxProj, design)
		deluxeTotal = dlxCOTS.EstimatedTotalUSD
		dlxNoise := assessNoise(dlxProj, design)
		deluxePropDBA = dlxNoise.PropertyLineDBA
		deluxePropFt = dlxNoise.PropertyLineFt
	}
	costDelta := round2(deluxeTotal - cotsTotal)

	// Phase 13: Quantify optional Deluxe Energy & Water Saving Upgrades
	ecoCost := 0.0
	waterSavedFrac := 0.0
	energySavedFrac := 0.0
	recoveredKW := 0.0
	var ecoHighlights []string

	evapGalPerFiring := design.Quench.EvaporatedWaterGPM * 60.0 * p.FiringDurationHours
	if evapGalPerFiring <= 0 {
		evapGalPerFiring = 240.0
	}
	motorKW := (design.Blower.ShaftPowerHP * 0.7457) / 0.88 // ~88% motor+VFD efficiency
	if motorKW <= 0 {
		motorKW = 1.25
	}
	baselineKWh := motorKW * p.FiringDurationHours
	nonSaltFrac := 0.75
	if p.FiringDurationHours > 0 {
		nonSaltFrac = math.Max(0.25, math.Min(0.95, (p.FiringDurationHours-p.EffectiveSaltingHours())/p.FiringDurationHours))
	}
	sensibleKW := math.Max(18.0, 58.0*(design.Input.KilnVolumeCuFt/24.0))

	if isDeluxe && p.DeluxeStage1HeatRecovery {
		ecoCost += 1650.0
		waterSavedFrac += 0.40 * (1.0 - waterSavedFrac)
		energySavedFrac += 0.07 * (1.0 - energySavedFrac)
		s1KW := round1(sensibleKW * 0.40)
		recoveredKW += s1KW
		ecoHighlights = append(ecoHighlights, fmt.Sprintf("Stage 1 Water-Jacket Pre-Cooler (+$1,650): harvests ~%.1f kW sensible heat before spray nozzles, cutting evaporative water loss ~40%% and blower load ~7%%", s1KW))
	}
	if isDeluxe && p.DeluxeCondensingReclaimer {
		ecoCost += 1850.0
		waterSavedFrac += 0.70 * (1.0 - waterSavedFrac)
		condKW := round1(sensibleKW * 0.55)
		recoveredKW += condKW
		ecoHighlights = append(ecoHighlights, fmt.Sprintf("Closed-Loop Condensing Plume Abater & Water Reclaimer (+$1,850): cools saturated exhaust 184 F -> ~120 F, reclaiming ~70%% of evaporated water back to sump, slashing winter plume >75%% & capturing ~%.1f kW low-grade hydronic heat", condKW))
	}
	if isDeluxe && p.DeluxeVariableThroatVFD {
		ecoCost += 1450.0
		// Affinity law P ~ N^3: widening throat during non-salting hours drops VFD speed ~48% (~82% power drop during off-peak hours).
		vfdSave := nonSaltFrac * 0.78
		energySavedFrac += vfdSave * (1.0 - energySavedFrac)
		ecoHighlights = append(ecoHighlights, fmt.Sprintf("Auto Variable-Throat Venturi + VFD Turndown (+$1,450): widens throat during %.1f non-salting hrs (dP ~19 -> ~3.8 inWC), cutting firing blower electricity ~%.0f%% and off-peak noise -10 dBA", p.FiringDurationHours-p.EffectiveSaltingHours(), vfdSave*100.0))
	}

	waterSavedPct := round1(waterSavedFrac * 100.0)
	waterSavedGal := round1(evapGalPerFiring * waterSavedFrac)
	energySavedPct := round1(energySavedFrac * 100.0)
	energySavedKWh := round1(baselineKWh * energySavedFrac)
	recoveredKW = round1(recoveredKW)
	ecoSummary := ""
	if ecoCost > 0 {
		ecoSummary = fmt.Sprintf(
			"Deluxe Eco Upgrades (+$%.0f): saves ~%.0f gal/firing (%.0f%% evap water reclaimed), ~%.1f kWh/firing (%.0f%% blower electricity saved), and recovers ~%.1f kW usable studio waste heat.",
			ecoCost, waterSavedGal, waterSavedPct, energySavedKWh, energySavedPct, recoveredKW,
		)
	}

	vesselMat := `TIG-Welded Heavy-Gauge 316L SS (11 ga / Sch 10S) Quench & Venturi + Filament-Wound Derakane Vinyl-Ester FRP Packed Column w/ Structural Gusset Rings`
	if p.CorrosionTier == CorrosionTierHeavyC276FRP {
		vesselMat = `TIG-Welded Heavy-Gauge Alloy C-276 (11 ga / Sch 10S) Quench & Venturi + Filament-Wound Derakane Vinyl-Ester FRP Packed Column w/ Structural Gusset Rings`
	}

	summary := fmt.Sprintf(
		"Turnkey shop-fabricated Deluxe package (+$%.2f vs No-Weld COTS): TIG-welded heavy-gauge hot stages with ANSI/ASME B16.5 Class 150 lap-joint flanges, Derakane FRP column with bolted manway & removable demister cassette, welded 2\"×2\"×3/16\" steel forklift skid, UL 508A NEMA 4X panel, and 2\" Rockwool acoustic baffle box + tuned stack silencer (-22 dBA -> %.1f dBA @ %.0f ft).",
		costDelta, deluxePropDBA, deluxePropFt,
	)
	if ecoSummary != "" {
		summary += " " + ecoSummary
	}

	return &DeluxeFabSpec{
		Enabled:         isDeluxe,
		VesselMaterial:  vesselMat,
		FlangeStandard:  `Machined ANSI/ASME B16.5 Class 150 Lap-Joint Flanges + PTFE Envelope Gaskets + 316 SS Studs`,
		SightGlasses:    `2× Flanged Borosilicate Sight-Glass Ports (Stage 1 Quench Spray + Stage 3 Packing Bed)`,
		CleanoutHatches: `2× Bolted NFPA 91 Quick-Release Cleanout Doors + 1" Ceramic-Fiber/SS Hot-Flue Insulation Jacket`,
		SkidFrame:       `Welded 2"×2"×3/16" Galvanized/Powder-Coated Steel Forklift Skid + Diagonal K-Bracing + Pipe Saddles + Concrete Anchors + PTFE/EPDM Vibration Boots`,
		PanelStandard:   `UL 508A-Style NEMA 4X Pre-Wired Panel w/ Through-Door NEC 430.102 Disconnect, 3% Line Reactor, HF-GFCI, FM/UL 1/16-DIN 200 F High-Limit, Dual Type-K Thermowells, Flow Transmitter, Flow-Through pH & Pilot Stack Light/Horn`,
		PumpUpgrade:     `Sloped-Bottom Baffled Chemical Settling Sump w/ 2" Drain Ball Valve + Heavy-Duty Sealless PVDF Mag-Drive Recirculation Pump Station`,
		CostDeltaUSD:    costDelta,
		Summary:         summary,
		Highlights: append([]string{
			"Back-purged TIG heavy-gauge 316L SS (or Alloy C-276) quench & venturi with ANSI B16.5 Class 150 lap-joint flanges",
			"2× borosilicate sight-glass ports + 2× bolted NFPA 91 quick-release cleanout doors + 1\" ceramic-fiber/SS hot-flue jacket",
			"Filament-wound Derakane vinyl-ester FRP column with structural gusset rings, bolted 12\" packing manway & removable demister cassette",
			"Sloped-bottom baffled settling sump (2\" drain ball valve) + heavy-duty sealless PVDF magnetic-drive pump station",
			"Welded 2\"×2\"×3/16\" structural steel forklift skid + diagonal K-bracing + concrete wedge anchors + flexible PTFE/EPDM vibration boots",
			"2\" Rockwool acoustic blower baffle enclosure w/ cooling-air labyrinth + tuned absorptive stack silencer (-22 dBA combined attenuation)",
			"UL 508A-style NEMA 4X pre-wired panel with through-door NEC 430.102 disconnect, 3% line reactor, HF-GFCI & FM/UL 200 F high-limit safety relay",
		}, ecoHighlights...),
		CondensingReclaimer:     isDeluxe && p.DeluxeCondensingReclaimer,
		VariableThroatVFD:       isDeluxe && p.DeluxeVariableThroatVFD,
		Stage1HeatRecovery:      isDeluxe && p.DeluxeStage1HeatRecovery,
		EcoCostDeltaUSD:         round2(ecoCost),
		WaterSavedPct:           waterSavedPct,
		WaterSavedGalPerFiring:  waterSavedGal,
		EnergySavedPct:          energySavedPct,
		EnergySavedKWhPerFiring: energySavedKWh,
		RecoveredHeatKW:         recoveredKW,
		EcoSummary:              ecoSummary,
		EcoHighlights:           ecoHighlights,
	}
}

// BuildShopPlan generates the structured ShopPlan (parts callouts, exact cut list
// with pkg/cad frustum flat-pattern developments, expanded code clearances, and full
// electrical, plumbing, pump station, bracing, acoustic, and Deluxe spec data).
func BuildShopPlan(
	project StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	urban UrbanAssessment,
	arch ArchitectureOptionsResult,
) ShopPlan {
	p := project.Normalize()
	if p.IsLocalCeramic() {
		return buildLocalCeramicShopPlan(p, design, cots, urban, arch)
	}
	isDeluxe := p.IsDeluxe()

	ductIn := round1(design.Envelope.DuctDiameter.Inches())
	ductLenIn := round1(design.Input.DuctLength.Inches())
	quenchDiaIn := round1(cots.QuenchVessel.NominalIn)
	quenchHIn := round1(design.Quench.TowerHeight.Inches())
	colDiaIn := round1(cots.PackedColumn.NominalIn)
	colHIn := round1(design.PackedBed.ColumnHeight.Inches())
	stackHIn := round1(urban.Plume.MinStackHeightFt * 12.0)

	ventInIn := round1(design.Venturi.InletDiameter.Inches())
	ventThIn := round2(design.Venturi.ThroatDiameter.Inches())
	ventOutIn := round1(design.Venturi.OutletDiameter.Inches())
	ventConvIn := round1(design.Venturi.ConvergingLength.Inches())
	ventThLenIn := round1(design.Venturi.ThroatLength.Inches())
	ventDivIn := round1(design.Venturi.DivergingLength.Inches())

	overallLenIn := round1(arch.FootprintLengthFt * 12.0)
	overallWidthIn := round1(arch.FootprintWidthFt * 12.0)
	overallHeightIn := round1(math.Max(design.Envelope.Height.Inches(), math.Max(quenchHIn, colHIn)+18.0))

	// Find venturi, pump, and demister costs from COTS BOM for exact traceability.
	ventCostUSD := estimateSingleVenturiCost(p, design)
	if arch.ParallelVenturi.Enabled && !arch.Eductor.Enabled {
		ventCostUSD = arch.ParallelVenturi.MultiCostUSD
	}
	if arch.Eductor.Enabled {
		ventCostUSD = round2(285.0 + 24.0*arch.Eductor.CyclonicSweepElbowIn)
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			ventCostUSD = round2(ventCostUSD * 2.2)
		}
	}

	// Build numbered Parts Legend [1]..[9].
	stage2Label := "Stage 2: Wet Venturi Particulate Scrubber"
	stage2SKU := "VENT-COTS-RED"
	if isDeluxe {
		stage2Label = "Stage 2: TIG-Welded Heavy-Gauge Venturi w/ ANSI B16.5 Class 150 Flanges"
		stage2SKU = "DLX-VENT-ANSI150"
	}
	stage2DimsIn := fmt.Sprintf(`%.1f" -> %.2f" throat -> %.1f" (L %.1f")`, ventInIn, ventThIn, ventOutIn, ventConvIn+ventThLenIn+ventDivIn)
	stage2DimsMM := fmt.Sprintf("%.0f -> %.1f -> %.0f mm (L %.0f mm)", ventInIn*25.4, ventThIn*25.4, ventOutIn*25.4, (ventConvIn+ventThLenIn+ventDivIn)*25.4)
	stage2Qty := 1.0
	stage2Notes := fmt.Sprintf("%.0f m/s throat, d50=%.2f um NaCl fume; 2x opposed throat spray ports.", design.Venturi.ThroatVelocity, design.Venturi.CutDiameterD50Microns)
	if arch.Eductor.Enabled {
		stage2Label = "Stage 2: Vertical Downflow Static Venturi Eductor & 90° Cyclonic Sweep Elbow"
		stage2SKU = fmt.Sprintf("EDUCT-VSTAT-%.0f", arch.Eductor.CyclonicSweepElbowIn)
		stage2DimsIn = fmt.Sprintf(`%.0f" Tee -> %.2f" downflow throat -> %.0f" 90° sweep elbow`, arch.Eductor.CyclonicSweepElbowIn, ventThIn, arch.Eductor.CyclonicSweepElbowIn)
		stage2DimsMM = fmt.Sprintf("%.0f mm Tee -> %.1f mm throat -> %.0f mm sweep", arch.Eductor.CyclonicSweepElbowIn*25.4, ventThIn*25.4, arch.Eductor.CyclonicSweepElbowIn*25.4)
		stage2Notes = fmt.Sprintf("Dry ambient motive air (%.0f CFM) enters top vertically, entrains quenched gas (%.0f ACFM), and turns 90° into Stage 3 bottom plenum (+%.2f inWC positive pressure).",
			arch.Eductor.MotiveAirCFM, arch.Eductor.QuenchedACFM, arch.Eductor.PositivePressureInWC)
	} else if arch.ParallelVenturi.Enabled {
		stage2Label = fmt.Sprintf("Stage 2: %d× Parallel COTS Sanitary Mini-Venturis", arch.ParallelVenturi.ThroatCount)
		stage2SKU = fmt.Sprintf("SAN-MINI-VENT-%dX", arch.ParallelVenturi.ThroatCount)
		stage2DimsIn = fmt.Sprintf(`%d × (%.0f" -> %.2f" -> %.0f" tri-clamp)`, arch.ParallelVenturi.ThroatCount, arch.ParallelVenturi.InletFittingIn, arch.ParallelVenturi.ThroatDiameterIn, arch.ParallelVenturi.InletFittingIn)
		stage2DimsMM = fmt.Sprintf("%d × (%.0f -> %.1f -> %.0f mm)", arch.ParallelVenturi.ThroatCount, arch.ParallelVenturi.InletFittingIn*25.4, arch.ParallelVenturi.ThroatDiameterIn*25.4, arch.ParallelVenturi.InletFittingIn*25.4)
		stage2Qty = float64(arch.ParallelVenturi.ThroatCount)
		stage2Notes = fmt.Sprintf("Zero cone rolling; plugging risk: %s. %s", arch.ParallelVenturi.PluggingRisk, arch.ParallelVenturi.TurndownBenefit)
	}

	stage3Label := fmt.Sprintf("Stage 3: %s", cots.PackedColumn.Name)
	stage3DimsIn := fmt.Sprintf(`%.1f" nom × %.1f" H (%.1f" packing)`, colDiaIn, colHIn, design.PackedBed.PackingHeight.Inches())
	stage3DimsMM := fmt.Sprintf("%.0f mm × %.0f mm H (%.0f mm packing)", colDiaIn*25.4, colHIn*25.4, design.PackedBed.PackingHeight.Millimeters())
	stage3Qty := 1.0
	stage3Notes := fmt.Sprintf("%.1f%% of Eckert flooding velocity (limit <= 75%%); media: %s (voidage %.0f%%, %.2f inWC bed dP) & 3-pass chevron demister.",
		cots.ActualPercentOfFlood, arch.PackingMedia.Label, arch.PackingMedia.Voidage*100.0, arch.PackingMedia.BedDPInWC)
	if arch.ParallelBed.Enabled {
		riserHIn := round1(math.Max(42.0, design.PackedBed.PackingHeight.Inches()+12.0))
		stage3Label = fmt.Sprintf(`Stage 3: Shell-Less Uniseal Bundle (%d × %.0f" %s)`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn, arch.ParallelBed.PipeMaterial)
		stage3DimsIn = fmt.Sprintf(`%d × %.0f" (ID %.3f") × %.1f" H + 2× flat plates + %d Uniseals`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn, arch.ParallelBed.PipeActualIDIn, riserHIn, 2*arch.ParallelBed.PipeCount)
		stage3DimsMM = fmt.Sprintf("%d × %.1f mm ID × %.0f mm H (NO external large shell)", arch.ParallelBed.PipeCount, arch.ParallelBed.PipeActualIDIn*25.4, riserHIn*25.4)
		stage3Qty = float64(arch.ParallelBed.PipeCount)
		stage3Notes = fmt.Sprintf("Shell-less flat-plate Uniseal bundle (%.1f%% flood, D/d_p=%.2f, %.2f inWC bed dP) with %s and balanced %d-port top liquid spray manifold.",
			arch.ParallelBed.ActualFloodPct, arch.ParallelBed.WallRatioDToDp, arch.PackingMedia.BedDPInWC, arch.PackingMedia.Label, arch.ParallelBed.PipeCount)
	}

	stage4Label := fmt.Sprintf("Stage 4: %s", cots.BlowerVFD.Name)
	stage4DimsIn := fmt.Sprintf(`%.1f" wheel, %.0f ACFM @ %.1f inWC`, cots.BlowerVFD.NominalIn, design.Blower.FlowACFM, arch.BlowerDrive.RequiredStaticInWC)
	stage4DimsMM := fmt.Sprintf("%.0f mm wheel, %.2g HP (%dV)", cots.BlowerVFD.NominalIn*25.4, arch.BlowerDrive.MotorHP, arch.BlowerDrive.Voltage)
	stage4Notes := cots.BlowerVFD.Notes
	if arch.Eductor.Enabled && p.BlowerDriveOption == BlowerDriveIndustrialRadialVFD {
		stage4Label = fmt.Sprintf("Stage 4 (Upstream Dry Motive): %s", cots.BlowerVFD.Name)
		stage4DimsIn = fmt.Sprintf(`%.0f CFM dry motive @ %.1f inWC (%.2g HP)`, arch.Eductor.MotiveAirCFM, arch.Eductor.DryFanStaticInWC, arch.Eductor.DryFanMotorHP)
		stage4DimsMM = fmt.Sprintf("%.0f m3/hr dry clean air, %.2g HP", arch.Eductor.MotiveAirCFM*1.699, arch.Eductor.DryFanMotorHP)
	}

	sumpLabel := cots.SumpTank.Name + " + Mag-Drive Pump"
	sumpQty := 1.0
	if !p.SharedBaffledSump {
		sumpLabel = "Dual Dedicated Sumps (Ash Settling + Absorber) + 2× Mag-Drive Pumps"
		sumpQty = 2.0
	}

	parts := []PlanPart{
		{
			Num:          1,
			ID:           "inlet_flue",
			Label:        "Hot Flue Inlet Duct & NFPA 86 Fail-Safe Bypass Tee",
			Stage:        "Inlet",
			Material:     cots.QuenchVessel.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" dia × %.1f" L`, ductIn, ductLenIn),
			DimensionsMM: fmt.Sprintf("%.0f mm dia × %.0f mm L", ductIn*25.4, ductLenIn*25.4),
			SKU:          fmt.Sprintf("FLUE-TEE-%02.0f", ductIn),
			Qty:          1,
			CostUSD:      round2(145.0 + 18.0*ductIn),
			Notes:        "Seal with >=2300 F ceramic fiber gaskets (NO silicone/PTFE); keep >=36\" clearance to combustibles (NFPA 211).",
		},
		{
			Num:          2,
			ID:           "quench_tower",
			Label:        "Stage 1: " + cots.QuenchVessel.Name,
			Stage:        "Stage 1",
			Material:     cots.QuenchVessel.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" nom × %.1f" H`, quenchDiaIn, quenchHIn),
			DimensionsMM: fmt.Sprintf("%.0f mm × %.0f mm H", quenchDiaIn*25.4, quenchHIn*25.4),
			SKU:          cots.QuenchVessel.SKU,
			Qty:          1,
			CostUSD:      cots.QuenchVessel.UnitCostUSD,
			Notes:        fmt.Sprintf("Co-current downward spray cools %.0f F -> %.1f F via %d × %s (%.2f GPM).", design.Input.PeakExitTempF, design.Quench.OutletTemp.Fahrenheit(), maxInt(design.Quench.NozzleCount, 2), cots.SprayNozzle.Name, design.Quench.TotalSprayWaterGPM),
		},
		{
			Num:          3,
			ID:           "venturi_stage",
			Label:        stage2Label,
			Stage:        "Stage 2",
			Material:     design.Venturi.Material,
			DimensionsIn: stage2DimsIn,
			DimensionsMM: stage2DimsMM,
			SKU:          stage2SKU,
			Qty:          stage2Qty,
			CostUSD:      ventCostUSD,
			Notes:        stage2Notes,
		},
		{
			Num:          4,
			ID:           "packed_bed",
			Label:        stage3Label,
			Stage:        "Stage 3",
			Material:     cots.PackedColumn.Material,
			DimensionsIn: stage3DimsIn,
			DimensionsMM: stage3DimsMM,
			SKU:          cots.PackedColumn.SKU,
			Qty:          stage3Qty,
			CostUSD:      cots.PackedColumn.UnitCostUSD,
			Notes:        stage3Notes,
		},
		{
			Num:          5,
			ID:           "sump_pump",
			Label:        sumpLabel,
			Stage:        "Stage 3 Sump",
			Material:     cots.SumpTank.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" dia vented atmospheric sump`, cots.SumpTank.NominalIn),
			DimensionsMM: fmt.Sprintf("%.0f mm dia atmospheric sump", cots.SumpTank.NominalIn*25.4),
			SKU:          cots.SumpTank.SKU,
			Qty:          sumpQty,
			CostUSD:      round2(cots.SumpTank.UnitCostUSD + arch.SumpDeltaUSD),
			Notes:        "Atmospheric vented sump ONLY (never under blower vacuum); baffled ash settling zone + 50-mesh pump Y-strainer.",
		},
		{
			Num:          6,
			ID:           "blower_vfd",
			Label:        stage4Label,
			Stage:        "Stage 4",
			Material:     cots.BlowerVFD.Material,
			DimensionsIn: stage4DimsIn,
			DimensionsMM: stage4DimsMM,
			SKU:          cots.BlowerVFD.SKU,
			Qty:          1,
			CostUSD:      cots.BlowerVFD.UnitCostUSD,
			Notes:        stage4Notes,
		},
		{
			Num:          7,
			ID:           "safety_controls",
			Label:        cots.SensorKit.Name,
			Stage:        "Safety & Controls",
			Material:     cots.SensorKit.Material,
			DimensionsIn: fmt.Sprintf(`%.0f" fail-open bypass damper + NEMA 4X panel`, cots.SensorKit.NominalIn),
			DimensionsMM: fmt.Sprintf("%.0f mm bypass damper + 24VDC safety loop", cots.SensorKit.NominalIn*25.4),
			SKU:          cots.SensorKit.SKU,
			Qty:          1,
			CostUSD:      cots.SensorKit.UnitCostUSD,
			Notes:        fmt.Sprintf("Control tier: %s. Includes mandatory NFPA 86 hardwired 200 F snap-disc/high-limit, 0.25 GPM flow switch & latching safety relay.", p.ControlTier),
		},
		{
			Num:          8,
			ID:           "exhaust_stack",
			Label:        "Outdoor Vertical Exhaust Stack (+ Silencer / Diluter)",
			Stage:        "Exhaust",
			Material:     "Sch 80 CPVC / Vinyl-Ester FRP (ASTM E84 Class 1)",
			DimensionsIn: fmt.Sprintf(`%.1f" dia × %.0f" H (%.0f ft min)`, ductIn, stackHIn, urban.Plume.MinStackHeightFt),
			DimensionsMM: fmt.Sprintf("%.0f mm dia × %.0f mm H", ductIn*25.4, stackHIn*25.4),
			SKU:          fmt.Sprintf("STACK-CPVC-%02.0f", ductIn),
			Qty:          1,
			CostUSD: func() float64 {
				c := round2(110.0 + 12.0*urban.Plume.MinStackHeightFt)
				if cots.Silencer != nil {
					c += cots.Silencer.UnitCostUSD
				}
				if cots.PlumeDiluter != nil {
					c += cots.PlumeDiluter.UnitCostUSD
				}
				return round2(c)
			}(),
			Notes: fmt.Sprintf("Terminate outdoors >=%.0f ft above grade with 1/2\" bottom condensate drain loop returning to sump.", urban.Plume.MinStackHeightFt),
		},
		{
			Num:          9,
			ID:           "code_utilities",
			Label:        "IPC 608 Backflow Preventer & NEC 430.102 Lockable Sight Disconnect",
			Stage:        "Utilities",
			Material:     "Lead-Free Bronze RPZ / Air-Gap + NEMA 4X Disconnect + 3% Line Reactor",
			DimensionsIn: `3/4" water isolation + NEMA 4X local disconnect <=50 ft`,
			DimensionsMM: "20 mm potable isolation + NEMA 4X sight disconnect",
			SKU:          "CODE-IPC608-NEC430",
			Qty:          1,
			CostUSD:      400.0,
			Notes:        fmt.Sprintf("Mandatory code utility kit: %s + IPC 608 Air Gap (>=1\") or ASSE 1013 RPZ.", urban.Electrical.CircuitLabel),
		},
	}

	// Build exact Cut Sheet (pipe cut lengths + unfolded cylinder/frustum flat patterns via pkg/cad).
	flueCutNotes := fmt.Sprintf(`Cut %.1f" dia pipe to %.1f" length (flat wrap %.2f" × %.1f"); add NFPA 91 cleanout hatch.`, ductIn, ductLenIn, math.Pi*ductIn, ductLenIn)
	quenchCutNotes := fmt.Sprintf(`%.1f" dia × %.1f" tall shell (flat sheet %.2f" × %.1f"). Drill %d top spray ports.`, quenchDiaIn, quenchHIn, math.Pi*quenchDiaIn, quenchHIn, maxInt(design.Quench.NozzleCount, 2))
	if isDeluxe {
		flueCutNotes += fmt.Sprintf(" Wrap with 1\" ceramic-fiber + SS jacket; %s.", ansi150FlangeNote(ductIn))
		quenchCutNotes += fmt.Sprintf(" Back-purged TIG seam; add borosilicate sight-glass & bolted NFPA 91 cleanout door; %s.", ansi150FlangeNote(quenchDiaIn))
	}

	cutList := []CutSheetItem{
		{
			PartNum:      1,
			Name:         "Inlet Hot Flue Duct Pipe",
			Shape:        "pipe",
			Material:     cots.QuenchVessel.Material,
			FlatWidthIn:  round2(math.Pi * ductIn),
			FlatHeightIn: ductLenIn,
			Qty:          1,
			Notes:        flueCutNotes,
		},
		{
			PartNum:      2,
			Name:         "Stage 1 Quench Tower Shell",
			Shape:        "cylinder",
			Material:     cots.QuenchVessel.Material,
			FlatWidthIn:  round2(math.Pi * quenchDiaIn),
			FlatHeightIn: quenchHIn,
			Qty:          1,
			Notes:        quenchCutNotes,
		},
	}

	if arch.Eductor.Enabled {
		edIn := arch.Eductor.CyclonicSweepElbowIn
		cutList = append(cutList,
			buildFrustumCutItem(3, "Downflow Eductor Converging Cone", design.Venturi.Material, edIn, ventThIn, ventConvIn, 1,
				fmt.Sprintf(`Vertical downflow cone %.1f" -> %.2f" throat (unfolded annular sector).`, edIn, ventThIn)),
			buildFrustumCutItem(3, "Downflow Eductor Diffuser Cone", design.Venturi.Material, ventThIn, edIn, ventDivIn, 1,
				fmt.Sprintf(`Vertical downflow diffuser %.2f" -> %.1f" into 90° cyclonic sweep elbow.`, ventThIn, edIn)),
			CutSheetItem{
				PartNum:      3,
				Name:         `90° Long-Radius Cyclonic Sweep Elbow`,
				Shape:        "pipe",
				Material:     "Sch 80 CPVC / 316L SS",
				FlatWidthIn:  round2(math.Pi * edIn),
				FlatHeightIn: round2(1.5 * edIn * 0.5 * math.Pi),
				Qty:          1,
				Notes:        fmt.Sprintf(`%.0f" long-radius 90° sweep elbow with bottom 1.5" drain stub to sump (cyclonic droplet separation).`, edIn),
			},
		)
	} else if arch.ParallelVenturi.Enabled {
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      3,
				Name:         fmt.Sprintf(`Parallel Sanitary Mini-Venturi Spools (%.0f"->%.2f"->%.0f")`, arch.ParallelVenturi.InletFittingIn, arch.ParallelVenturi.ThroatDiameterIn, arch.ParallelVenturi.InletFittingIn),
				Shape:        "pipe",
				Material:     design.Venturi.Material,
				FlatWidthIn:  round2(math.Pi * arch.ParallelVenturi.ThroatDiameterIn),
				FlatHeightIn: round2(math.Max(3.0, 2.5*arch.ParallelVenturi.ThroatDiameterIn)),
				Qty:          arch.ParallelVenturi.ThroatCount,
				Notes:        fmt.Sprintf("Assemble %d identical COTS sanitary concentric reducer spools with tri-clamp tees (zero cone rolling).", arch.ParallelVenturi.ThroatCount),
			},
		)
	} else {
		ventFlangeSuffix := ""
		if isDeluxe {
			ventFlangeSuffix = " " + ansi150FlangeNote(ventInIn) + "."
		}
		cutList = append(cutList,
			buildFrustumCutItem(3, "Stage 2 Venturi Converging Cone", design.Venturi.Material, ventInIn, ventThIn, ventConvIn, 1,
				fmt.Sprintf(`Unfolded 25° included converging cone (%.1f" -> %.2f", axial %.1f").%s`, ventInIn, ventThIn, ventConvIn, ventFlangeSuffix)),
			CutSheetItem{
				PartNum:      3,
				Name:         "Stage 2 Venturi Throat Spool",
				Shape:        "cylinder",
				Material:     design.Venturi.Material,
				FlatWidthIn:  round2(math.Pi * ventThIn),
				FlatHeightIn: ventThLenIn,
				Qty:          1,
				Notes:        fmt.Sprintf(`%.2f" ID × %.1f" L throat spool with 2× opposed NPT injection bosses.`, ventThIn, ventThLenIn),
			},
			buildFrustumCutItem(3, "Stage 2 Venturi Diverging Diffuser Cone", design.Venturi.Material, ventThIn, ventOutIn, ventDivIn, 1,
				fmt.Sprintf(`Unfolded 7° included pressure-recovery diffuser (%.2f" -> %.1f", axial %.1f").%s`, ventThIn, ventOutIn, ventDivIn, ventFlangeSuffix)),
		)
	}

	if arch.ParallelBed.Enabled {
		riserHIn := round1(math.Max(42.0, design.PackedBed.PackingHeight.Inches()+12.0))
		// Uniseals require >= +2.0" pitch over nominal pipe size (e.g., 6.0" pitch for 4" Uniseals w/ 5.75" rubber flange OD).
		plateSideIn := round1(math.Max(16.0, math.Ceil(math.Sqrt(float64(arch.ParallelBed.PipeCount)))*(arch.ParallelBed.PipeNominalIn+2.0)+4.0))
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      4,
				Name:         fmt.Sprintf(`Shell-Less Parallel %.0f" Risers (%s)`, arch.ParallelBed.PipeNominalIn, arch.ParallelBed.PipeMaterial),
				Shape:        "pipe",
				Material:     arch.ParallelBed.PipeMaterial,
				FlatWidthIn:  round2(math.Pi * arch.ParallelBed.PipeActualIDIn),
				FlatHeightIn: riserHIn,
				Qty:          arch.ParallelBed.PipeCount,
				Notes:        fmt.Sprintf(`Cut %d equal lengths of %.0f" stock pipe to %.1f" each (NO external large shell); fill %.1f" depth with %s.`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn, riserHIn, design.PackedBed.PackingHeight.Inches(), arch.PackingMedia.Label),
			},
			CutSheetItem{
				PartNum:      4,
				Name:         `Top & Bottom Flat Collector Plates (1/2" HDPE/CPVC + Uniseal Ports)`,
				Shape:        "plate",
				Material:     `1/2" HDPE / CPVC Flat Sheet + Rubber Uniseals`,
				FlatWidthIn:  plateSideIn,
				FlatHeightIn: plateSideIn,
				Qty:          2,
				Notes:        fmt.Sprintf(`Cut 2× %.1f" × %.1f" flat plates and hole-saw %d ports per plate for %d Uniseal rubber pipe seals ($5.50 ea).`, plateSideIn, plateSideIn, arch.ParallelBed.PipeCount, 2*arch.ParallelBed.PipeCount),
			},
		)
	} else {
		topRedH := round1(math.Max(6.0, (colDiaIn-ductIn)*0.9))
		colCutNotes := fmt.Sprintf(`%.1f" nom × %.1f" tall absorber shell; support grate at 12" above inlet, packing depth %.1f".`, colDiaIn, colHIn, design.PackedBed.PackingHeight.Inches())
		if isDeluxe {
			colCutNotes += fmt.Sprintf(" Add bolted 12\" packing manway, borosilicate sight glass, structural gusset rings & removable demister cassette; %s.", ansi150FlangeNote(colDiaIn))
		}
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      4,
				Name:         "Stage 3 Packed-Bed Column Shell",
				Shape:        "cylinder",
				Material:     cots.PackedColumn.Material,
				FlatWidthIn:  round2(math.Pi * colDiaIn),
				FlatHeightIn: colHIn,
				Qty:          1,
				Notes:        colCutNotes,
			},
			buildFrustumCutItem(4, "Stage 3 Top Column-to-Stack Reducer Cone", cots.PackedColumn.Material, colDiaIn, ductIn, topRedH, 1,
				fmt.Sprintf(`Transitions %.1f" absorber column to %.1f" outlet duct/stack.`, colDiaIn, ductIn)),
		)
	}

	cutList = append(cutList,
		CutSheetItem{
			PartNum:      8,
			Name:         "Outdoor Exhaust Stack Pipe",
			Shape:        "pipe",
			Material:     "Sch 80 CPVC / FRP (ASTM E84 Class 1)",
			FlatWidthIn:  round2(math.Pi * ductIn),
			FlatHeightIn: stackHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`%.1f" dia × %.0f" (%.0f ft) vertical outdoor stack with 1/2" condensate drain loop.`, ductIn, stackHIn, urban.Plume.MinStackHeightFt),
		},
	)

	// Expand CutList in Deluxe mode with ANSI flange rings, structural steel skid frame cuts,
	// column K-bracing struts, and acoustic blower baffle box panels.
	if isDeluxe {
		flangeOD := ansi150FlangeOD(quenchDiaIn)
		boxSide := round1(math.Max(26.0, cots.BlowerVFD.NominalIn+14.0))
		kBraceLen := round1(math.Hypot(colDiaIn+8.0, colHIn*0.45))
		kBraceAngleDeg := round1(math.Atan2(colHIn*0.45, colDiaIn+8.0) * 180.0 / math.Pi)
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      2,
				Name:         "ANSI/ASME B16.5 Class 150 Lap-Joint Flange Rings & PTFE Envelope Gaskets",
				Shape:        "plate",
				Material:     "316L SS Machined Backing Rings + PTFE Envelope Gaskets",
				FlatWidthIn:  flangeOD,
				FlatHeightIn: flangeOD,
				Qty:          6,
				Notes:        fmt.Sprintf("%s; %s; star-pattern torque in 3 stages with anti-seize on 316 SS studs.", ansi150FlangeNote(quenchDiaIn), ansi150FlangeNote(colDiaIn)),
			},
			CutSheetItem{
				PartNum:      5,
				Name:         `Welded Steel Forklift Skid Rails & Cross-Members (2"×2"×3/16" Tube)`,
				Shape:        "tube",
				Material:     `ASTM A500 2"×2"×3/16" Galvanized/Powder-Coated Steel Tubing`,
				FlatWidthIn:  overallWidthIn,
				FlatHeightIn: overallLenIn,
				Qty:          2,
				Notes:        fmt.Sprintf(`Cut 2× main rails @ %.1f" L + 5× cross-members @ %.1f" W + forklift channel pockets; weld 6× 1/2" anchor pads.`, overallLenIn, overallWidthIn),
			},
			CutSheetItem{
				PartNum:      4,
				Name:         `Stage 3 Column Cage Uprights & Diagonal K-Bracing Struts`,
				Shape:        "tube",
				Material:     `2"×2"×3/16" Steel Uprights + 1-1/2"×1-1/2"×1/8" Diagonal K-Braces`,
				FlatWidthIn:  kBraceLen,
				FlatHeightIn: colHIn,
				Qty:          4,
				Notes:        fmt.Sprintf(`Cut 4× vertical posts @ %.1f" H + 8× diagonal K-braces @ %.1f" L (%.1f° mitered ends) around Stage 3 column.`, colHIn, kBraceLen, kBraceAngleDeg),
			},
			CutSheetItem{
				PartNum:      6,
				Name:         `Deluxe Acoustic Blower Baffle Enclosure Panels (2" Rockwool + Perforated Liner)`,
				Shape:        "panel",
				Material:     `16 ga Powder-Coated Steel Outer + 2" Rockwool Mineral Wool + Perforated Metal Inner Liner`,
				FlatWidthIn:  boxSide,
				FlatHeightIn: boxSide,
				Qty:          6,
				Notes:        fmt.Sprintf(`Cut 6× %.1f" × %.1f" enclosure panels + 2× baffled cooling-air labyrinth plenums (intake & exhaust) for TEFC motor cooling without line-of-sight noise leak.`, boxSide, boxSide),
			},
		)
	}

	propMinFt := 10.0
	if p.VaporGlazeMode == VaporGlazeSalt || design.PackedBed.HClOutletPPM > 0 {
		propMinFt = 30.0 // IMC 501.3.1 hazardous HCl exhaust setback
	}

	wiringNote := `NEC 110.26(A): Maintain unobstructed 36" depth × 30" width × 78" height working space in front of NEMA 4X Control/VFD Panel and NEC 430.102 lockable sight disconnect (<=50 ft); route 24VDC NFPA 86 safety loop & RS-485 instrumentation conduit >=12" from VFD motor leads.`
	if isDeluxe {
		wiringNote = `NEC 110.26(A) / UL 508A: Maintain unobstructed 36" depth × 30" width × 78" height working space in front of the UL 508A NEMA 4X pre-wired panel (through-door lockable NEC 430.102 disconnect); segregate shielded VFD power leads >=12" from 24VDC FM/UL 200 F safety loop & instrumentation conduit.`
	}

	electrical := buildElectricalSchedule(p, design, urban, arch)
	plumbing, pumpStation := buildPlumbingSchedule(p, design, cots, arch)
	bracing := buildBracingSchedule(p, design, cots, arch)
	acoustics, acousticSummary := buildAcousticsSchedule(p, design, urban, arch)
	deluxeSpec := buildDeluxeSpec(p, design, cots, urban)

	return ShopPlan{
		Title:            fmt.Sprintf("Complete Shop Plan — %s (%.1f cu.ft.)", p.StudioName, design.Input.KilnVolumeCuFt),
		BuildPath:        p.BuildPath,
		IsDeluxe:         isDeluxe,
		StageAlignment:   p.StageAlignment,
		PackedBedLayout:  p.PackedBedLayout,
		VenturiLayout:    p.VenturiLayout,
		CompactFootprint: p.CompactFootprint,
		OverallLengthIn:  overallLenIn,
		OverallWidthIn:   overallWidthIn,
		OverallHeightIn:  overallHeightIn,
		Parts:            parts,
		CutList:          cutList,
		Clearances: PlanClearances{
			FlueToCombustiblesIn: 36.0,
			PropertyLineMinFt:    propMinFt,
			StackAboveRoofFt:     6.0,
			ServiceAisleIn:       36.0,
			NECPanelDepthIn:      36.0,
			NECPanelWidthIn:      30.0,
			PumpServiceIn:        24.0,
			CleanoutSwingIn:      24.0,
			HoseBendRadiusIn:     6.0,
			WiringAccessNote:     wiringNote,
		},
		Electrical:      electrical,
		Plumbing:        plumbing,
		PumpStation:     pumpStation,
		Bracing:         bracing,
		Acoustics:       acoustics,
		AcousticSummary: acousticSummary,
		DeluxeSpec:      deluxeSpec,
	}
}

// buildLocalCeramicShopPlan produces the complete masonry, stoneware-packed tower,
// elevated overhead gravity tank stand, manual/local pump plumbing (P1..P5), tall
// thermal-reheat brick chimney, bracing (B1..B5), and 0-circuit (when off-grid)
// ShopPlan for Phase 10 Local Ceramic / Zero-Electricity builds.
func buildLocalCeramicShopPlan(
	p StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	urban UrbanAssessment,
	arch ArchitectureOptionsResult,
) ShopPlan {
	fr := EvaluateFrugal(p, design)
	prof := resolveLocalRegionProfile(p.EffectiveLocalRegion())

	bedDiaIn := fr.BedDiameterIn
	bedDepthIn := round1(localCeramicBedDepthM(design.Input.KilnVolumeCuFt) / model.MetersPerInch)
	towerHIn := round1(bedDepthIn + 34.0) // lower plenum + stoneware bed + spray clearance + terracotta chevron baffles
	chimneyHIn := round1(fr.ChimneyHeightFt * 12.0)
	quenchDiaIn := round1(cots.QuenchVessel.NominalIn)
	quenchHIn := round1(math.Max(48.0, design.Quench.TowerHeight.Inches()*0.75))
	ductIn := round1(math.Max(10.0, design.Envelope.DuctDiameter.Inches()+2.0))

	overallLenIn := round1(math.Max(96.0, quenchDiaIn+bedDiaIn+56.0))
	overallWidthIn := round1(math.Max(48.0, bedDiaIn+26.0))
	overallHeightIn := round1(math.Max(fr.TankElevationFt*12.0+42.0, towerHIn+18.0))

	tankWetKg := round1(fr.GravityTankLiters*1.05 + 22.0)
	tankWetLbs := round1(tankWetKg / model.KgPerLb)
	ringCuFt := math.Max(4.0, math.Ceil(model.CubicMeters(math.Pi*math.Pow(float64(model.InchesToMeters(bedDiaIn)), 2)/4.0*(bedDepthIn*model.MetersPerInch)).CuFt()))
	towerWetLbs := round1((cots.PackedColumn.WeightKg + ringCuFt*21.0 + 40.0) / model.KgPerLb)
	chimneyLbs := round1(cots.BlowerVFD.WeightKg / model.KgPerLb)
	quenchLbs := round1(cots.QuenchVessel.WeightKg / model.KgPerLb)
	totalPadLbs := round1(quenchLbs + towerWetLbs + tankWetLbs + chimneyLbs + 350.0)

	flueLabel := "Refractory Brick Kiln Flue & Concentric Flue-Wall Thermal Recuperator"
	flueDimIn := fmt.Sprintf(`%.0f" masonry flue × 48" L (concentric brick reheat wall)`, ductIn)
	flueNotes := fmt.Sprintf("Transfers waste heat from %.0f F kiln flue across a split-firebrick wall to warm scrubbed exhaust back to %.0f F without mixing raw gas.", design.Input.PeakExitTempF, fr.StackReheatTempF)
	chimneyPartLabel := fmt.Sprintf("Stage 4: Tall Brick / Terracotta Thermal-Reheat Chimney (%.1f m / %.0f ft)", fr.ChimneyHeightM, fr.ChimneyHeightFt)
	if fr.SteamEnabled {
		flueLabel = "Refractory Brick Kiln Inlet Flue (Direct to Quench, No Reheat Hearth)"
		flueDimIn = fmt.Sprintf(`%.0f" masonry flue × 48" L (refractory brick inlet)`, ductIn)
		flueNotes = fmt.Sprintf("Conveys %.0f F kiln flue gas into the brick quench chamber (steam-jet ejector replaces the tall thermal-reheat recuperator).", design.Input.PeakExitTempF)
		chimneyPartLabel = fmt.Sprintf("Stage 4: Compact Brick / SWG Exhaust Chimney (%.1f m / %.0f ft, Steam-Jet Driven)", fr.ChimneyHeightM, fr.ChimneyHeightFt)
	} else if !p.ZeroElectricity {
		flueLabel = "Refractory Brick Kiln Inlet Flue (Hybrid Fan-Assisted Mode)"
		flueDimIn = fmt.Sprintf(`%.0f" masonry flue × 48" L (refractory brick inlet)`, ductIn)
		flueNotes = fmt.Sprintf("Conveys %.0f F kiln flue gas into the brick quench chamber (local low-dP draft fan replaces tall thermal-reheat chimney).", design.Input.PeakExitTempF)
		chimneyPartLabel = fmt.Sprintf("Stage 4: Compact Brick / SWG Exhaust Chimney (%.1f m / %.0f ft, Fan-Assisted)", fr.ChimneyHeightM, fr.ChimneyHeightFt)
	}

	parts := []PlanPart{
		{
			Num:          1,
			ID:           "brick_flue_recuperator",
			Label:        flueLabel,
			Stage:        "Inlet",
			Material:     prof.pipeLocalSpec,
			DimensionsIn: flueDimIn,
			DimensionsMM: fmt.Sprintf("%.0f mm masonry flue × 1220 mm L", ductIn*25.4),
			SKU:          "LOCAL-BRICK-FLUE",
			Qty:          1,
			CostUSD:      round2(cots.QuenchVessel.UnitCostUSD * 0.40),
			Notes:        flueNotes,
		},
		{
			Num:          2,
			ID:           "masonry_quench_chamber",
			Label:        "Stage 1: Masonry Settling & Gravity-Spray Quench Chamber",
			Stage:        "Stage 1",
			Material:     cots.QuenchVessel.Material,
			DimensionsIn: fmt.Sprintf(`%.0f" sq/dia masonry × %.0f" H`, quenchDiaIn, quenchHIn),
			DimensionsMM: fmt.Sprintf("%.0f mm × %.0f mm H", quenchDiaIn*25.4, quenchHIn*25.4),
			SKU:          cots.QuenchVessel.SKU,
			Qty:          1,
			CostUSD:      round2(cots.QuenchVessel.UnitCostUSD * 0.60),
			Notes:        "All-ceramic firebrick chamber tolerates >1,800 F dry-run heat with zero plastic meltdown risk; includes brick ash cleanout arch.",
		},
		{
			Num:          3,
			ID:           "terracotta_baffles_rose",
			Label:        "Stage 2: Terracotta Settling Baffles & Brass/PVC Gravity Spray Rose",
			Stage:        "Stage 2",
			Material:     cots.SprayNozzle.Material,
			DimensionsIn: fmt.Sprintf(`3× baffle tiers + 1" gravity spray rose (%.2f GPM @ %.2f PSI)`, fr.SprayFlowGPM, fr.GravityHeadPSI),
			DimensionsMM: fmt.Sprintf("3× terracotta baffles + 25 mm spray rose (%.2f L/min)", fr.SprayFlowLPM),
			SKU:          cots.SprayNozzle.SKU,
			Qty:          1,
			CostUSD:      cots.SprayNozzle.UnitCostUSD,
			Notes:        "Low-resistance coarse droplet quench & fly-ash knockdown (0.030 inWC dP) driven purely by overhead gravity head.",
		},
		{
			Num:          4,
			ID:           "wide_stoneware_tower",
			Label:        fmt.Sprintf(`Stage 3: %.0f" Wide Masonry / SWG Ceramic Packed Tower & 1" Stoneware Rings`, bedDiaIn),
			Stage:        "Stage 3",
			Material:     cots.PackedColumn.Material,
			DimensionsIn: fmt.Sprintf(`%.0f" bore × %.0f" H (%.0f" stoneware ring bed + terracotta chevron baffles)`, bedDiaIn, towerHIn, bedDepthIn),
			DimensionsMM: fmt.Sprintf("%.0f mm bore × %.0f mm H (%.0f mm bed)", bedDiaIn*25.4, towerHIn*25.4, bedDepthIn*25.4),
			SKU:          cots.PackedColumn.SKU,
			Qty:          1,
			CostUSD:      round2(cots.PackedColumn.UnitCostUSD + ringCuFt*cots.PackingMedia.UnitCostUSD),
			Notes:        fmt.Sprintf("Wide low-velocity bore (v_s=%.2f m/s) keeps irrigated Ergun bed dP down to %.3f inWC while achieving %.1f%% HCl neutralization.", fr.BedSuperficialVelMS, fr.BedDPInWC, fr.HClRemovalPct),
		},
		{
			Num:          5,
			ID:           "brick_baffled_sump",
			Label:        "Plastered Brick / Glazed-Tile Baffled Settling Sump",
			Stage:        "Stage 3 Sump",
			Material:     "Cement-Plastered Brick / Glazed Stoneware Tile Baffled Basin",
			DimensionsIn: `36" L × 24" W × 24" D (~55 gal / 208 L working volume)`,
			DimensionsMM: "915 mm × 610 mm × 610 mm (~208 L)",
			SKU:          "LOCAL-BRICK-SUMP",
			Qty:          1,
			CostUSD:      round2(cots.SumpTank.UnitCostUSD * 0.35),
			Notes:        fmt.Sprintf("Weir baffle traps settled wood ash & lime/alkali sediment before suction strainer; dose %.2f kg %s (%s) per firing.", fr.AlkaliKgPerFiring, fr.AlkaliName, fr.AlkaliFormula),
		},
		{
			Num:          6,
			ID:           "overhead_gravity_tank",
			Label:        fmt.Sprintf("Elevated %.0f L (%.0f gal) %s on Braced 3.5 m Stand", fr.GravityTankLiters, fr.GravityTankGal, prof.tankBrand),
			Stage:        "Gravity Feed",
			Material:     fmt.Sprintf("UV-Stabilized HDPE Tank + Diagonally Braced Timber/Brick Stand (%.0f kg Wet Load)", tankWetKg),
			DimensionsIn: fmt.Sprintf(`%.0f L tank @ %.1f ft elevation (%.2f PSI static head)`, fr.GravityTankLiters, fr.TankElevationFt, fr.GravityHeadPSI),
			DimensionsMM: fmt.Sprintf("%.0f L tank @ 3500 mm elevation (%.2f PSI)", fr.GravityTankLiters, fr.GravityHeadPSI),
			SKU:          "LOCAL-GRAV-TANK",
			Qty:          1,
			CostUSD:      round2(cots.SumpTank.UnitCostUSD * 0.65),
			Notes:        fmt.Sprintf("MANDATORY STRUCTURAL BRACING: Filled tank weighs ~%.0f kg (%.0f lbs) — cross-brace all 4 stand legs on compacted concrete/brick footings!", tankWetKg, tankWetLbs),
		},
		{
			Num:          7,
			ID:           "circulation_pump",
			Label:        fr.ManualPumpType,
			Stage:        "Pumping",
			Material:     "Cast-Iron / HDPE Manual Pump (or Local 0.5 HP 230V 50Hz Domestic Monoblock Pump)",
			DimensionsIn: fmt.Sprintf(`%.2f GPM (%.2f L/min) closed-loop lift to %.1f ft`, fr.SprayFlowGPM, fr.SprayFlowLPM, fr.TankElevationFt),
			DimensionsMM: fmt.Sprintf("%.2f L/min lift to 3.5 m overhead tank", fr.SprayFlowLPM),
			SKU:          "LOCAL-PUMP-UNIT",
			Qty:          1,
			CostUSD:      round2(75.0 * prof.costScale),
			Notes: func() string {
				if p.ZeroElectricity {
					return fmt.Sprintf("Zero-electricity manual closed-loop refill: %.1f min/hr on treadle/hand pump (or %d × 20L jerrycans/hr).", fr.TreadleMinutesPerHr, fr.Jerrycans20LPerHr)
				}
				if fr.SteamEnabled {
					return fmt.Sprintf("Local 230V 50Hz domestic monoblock pump (%.0f W, zero electric draft fan — steam ejector drives draft) automates closed-loop tank refill; overhead gravity tank provides blackout ride-through.", fr.PowerWatts)
				}
				return fmt.Sprintf("Local 230V 50Hz domestic monoblock pump (%.0f W w/ low-dP fan) automates closed-loop tank refill; overhead gravity tank provides blackout ride-through.", fr.PowerWatts)
			}(),
		},
		{
			Num:          8,
			ID:           "reheat_brick_chimney",
			Label:        chimneyPartLabel,
			Stage:        "Exhaust",
			Material:     cots.BlowerVFD.Material,
			DimensionsIn: fmt.Sprintf(`12" flue × %.0f" H (%.1f ft / %.1f m)`, chimneyHIn, fr.ChimneyHeightFt, fr.ChimneyHeightM),
			DimensionsMM: fmt.Sprintf("300 mm flue × %.0f mm H (%.1f m)", fr.ChimneyHeightM*1000.0, fr.ChimneyHeightM),
			SKU:          cots.BlowerVFD.SKU,
			Qty:          1,
			CostUSD:      cots.BlowerVFD.UnitCostUSD,
			Notes:        fmt.Sprintf("Available draft %.3f inWC vs %.3f inWC required (margin %+.3f inWC, feasible=%v); guy-wire / wall-tie every 3 m.", fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftMarginInWC, fr.DraftFeasible),
		},
		{
			Num:          9,
			ID:           "field_test_alkali_kit",
			Label:        "Field Hydrometer, Litmus pH Strips, Stem Thermometer & Local Alkali Charge",
			Stage:        "Monitoring & Chemistry",
			Material:     cots.SensorKit.Material,
			DimensionsIn: `Zero-electronics field kit (SG 1.000-1.200 + pH 1-14 strips)`,
			DimensionsMM: fmt.Sprintf("%.2f kg %s (%s) per firing", fr.AlkaliKgPerFiring, fr.AlkaliName, fr.AlkaliFormula),
			SKU:          cots.SensorKit.SKU,
			Qty:          1,
			CostUSD:      cots.SensorKit.UnitCostUSD,
			Notes:        fr.AlkaliLocalNote,
		},
	}

	if fr.SteamEnabled {
		steamCost := 0.0
		for _, it := range cots.BOM.Items {
			if it.Category == "Stage 2B: Certified Steam Source & de Laval Ejector" ||
				it.Category == "Stage 2B: Commercial Catering Steamer, 4.2 m Standpipe & Brass Orifice" {
				steamCost = it.UnitCostUSD
			}
		}
		isSubsonic := fr.DraftMode == LocalDraftSteamEjectorSubsonic
		boilerSKU := "LOCAL-MONOTUBE-20PSIG"
		boilerLabel := "Stage 2B: Code-Rated Small-Bore 316L Monotube Flash Coil + Certified PRV & Gauge (20 psig)"
		boilerDimsIn := `3/8" seamless 316L monotube (<2 L volume) + 25 psig ASME/IBR/KEBS PRV + 0-100 psig gauge`
		boilerDimsMM := "10 mm seamless 316L coil (<2 L) + certified PRV & siphon gauge"
		ejectorID := "de_laval_steam_ejector"
		ejectorLabel := fmt.Sprintf("Stage 2B: Machined 316L de Laval Steam-Jet Ejector Throat (%.2f mm) & Mixing Diffuser", fr.SteamNozzleThroatMM)
		ejectorMat := "Machined 316L Stainless de Laval Nozzle + Insulated Steam Header (S1) + Stoneware/316L Mixing Diffuser"
		ejectorDimsIn := fmt.Sprintf(`%.3f" (%.2f mm) choked throat, %.0f m/s (Mach %.2f) -> %.3f inWC draft`, fr.SteamNozzleThroatMM/25.4, fr.SteamNozzleThroatMM, fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.EjectorDraftInWC)
		ejectorDimsMM := fmt.Sprintf("%.2f mm choked throat, %.0f m/s (Mach %.2f), entrainment %.1f:1", fr.SteamNozzleThroatMM, fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.EjectorEntrainmentRatio)
		ejectorSKU := "LOCAL-DELAVAL-EJECTOR-316L"
		if isSubsonic {
			boilerSKU = "LOCAL-CATERING-STEAMER-6PSIG"
			boilerLabel = "Stage 2B: Certified Commercial Catering Stainless Steamer (6 psig / 0.414 barg) + 4.2 m Open Hydrostatic Standpipe (S2)"
			boilerDimsIn = `12-15 psig commercial catering steamer (run @ 6 psig) + 13.8 ft (4.2 m) open water-column standpipe + 10-15 psig brass PRV & 0-15 psig gauge`
			boilerDimsMM = "Commercial catering steamer (0.414 barg) + 4,200 mm open hydrostatic standpipe (6 psig physical cap) + 10-15 psig brass PRV"
			ejectorID = "subsonic_brass_orifice_ejector"
			ejectorLabel = fmt.Sprintf("Stage 2B: Simple Drilled Brass Converging Orifice (%.2f mm, Unchoked Mach %.2f) & Masonry Mixing Diffuser", fr.SteamNozzleThroatMM, fr.SteamNozzleMach)
			ejectorMat = "Simple Drilled 1/2\" Brass Pipe-Cap Converging Orifice (No CNC Bell) + Insulated Low-Pressure Header (S1) + Stoneware Mixing Diffuser"
			ejectorDimsIn = fmt.Sprintf(`%.3f" (%.2f mm) drilled brass orifice, %.0f m/s (Mach %.2f < 1.0, ~52 dBA) -> %.3f inWC draft`, fr.SteamNozzleThroatMM/25.4, fr.SteamNozzleThroatMM, fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.EjectorDraftInWC)
			ejectorDimsMM = fmt.Sprintf("%.2f mm drilled brass orifice (unchoked Mach %.2f), entrainment %.1f:1", fr.SteamNozzleThroatMM, fr.SteamNozzleMach, fr.EjectorEntrainmentRatio)
			ejectorSKU = "LOCAL-BRASS-ORIFICE-SUBSONIC"
		} else if fr.DraftMode == LocalDraftSteamEjectorHP {
			boilerSKU = "LOCAL-BABY-BOILER-60PSIG"
			boilerLabel = "Stage 2B: Certified Non-IBR (<=25 L) / KEBS Commercial Baby Steam Boiler + PRV & Sight Glass (60 psig)"
			boilerDimsIn = `<=25 L certified baby boiler (60 psig) + 75 psig PRV, fusible plug, sight glass & check valve`
			boilerDimsMM = "<=25 L non-IBR / KEBS baby boiler (4.14 barg) + certified safety trim"
		}
		parts = append(parts,
			PlanPart{
				Num:          10,
				ID:           "certified_steam_source",
				Label:        boilerLabel,
				Stage:        "Stage 2B Steam",
				Material:     fr.SteamSourceType,
				DimensionsIn: boilerDimsIn,
				DimensionsMM: boilerDimsMM,
				SKU:          boilerSKU,
				Qty:          1,
				CostUSD:      round2(steamCost * 0.72),
				Notes:        fmt.Sprintf("%.1f kg/hr steam (%.1f kW thermal vs %.1f kW kiln flue waste heat); MUST pass 1.5x MAWP cold hydro-test. NEVER build unvented DIY welded drum boilers!", fr.SteamFlowKgHr, fr.SteamThermalKW, fr.KilnWasteHeatKW),
			},
			PlanPart{
				Num:          11,
				ID:           ejectorID,
				Label:        ejectorLabel,
				Stage:        "Stage 2B Steam",
				Material:     ejectorMat,
				DimensionsIn: ejectorDimsIn,
				DimensionsMM: ejectorDimsMM,
				SKU:          ejectorSKU,
				Qty:          1,
				CostUSD:      round2(steamCost * 0.28),
				Notes:        fr.CondensationGrowthNote,
			},
		)
	}

	cutList := []CutSheetItem{
		{
			PartNum:      1,
			Name:         "Refractory Brick Quench & Concentric Flue-Wall Recuperator Courses",
			Shape:        "masonry",
			Material:     "Local Firebrick & Refractory Clay / Lime-Sand Mortar",
			FlatWidthIn:  quenchDiaIn,
			FlatHeightIn: quenchHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`Lay %.0f" internal masonry chamber × %.0f" H with a 65 mm split-firebrick heat-transfer wall separating %.0f F raw flue from clean chimney reheat duct (zero gas cross-leakage).`, quenchDiaIn, quenchHIn, design.Input.PeakExitTempF),
		},
		{
			PartNum:      4,
			Name:         fmt.Sprintf(`Wide Masonry / SWG Absorber Tower (%.0f" Bore × %.0f" H)`, bedDiaIn, towerHIn),
			Shape:        "cylinder",
			Material:     prof.pipeLocalSpec,
			FlatWidthIn:  round2(math.Pi * bedDiaIn),
			FlatHeightIn: towerHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`Construct %.0f" (%.0f mm) wide brick or IS 651 SWG clay-pipe absorber with a heavy firebrick/stoneware packing support arch 14" above the lower gas inlet.`, bedDiaIn, bedDiaIn*25.4),
		},
		{
			PartNum:      4,
			Name:         `Potter-Extruded 1" Unglazed Stoneware Raschig Rings`,
			Shape:        "ceramic",
			Material:     "Grogged Stoneware Clay Fired to Cone 6-10 (70% Voidage)",
			FlatWidthIn:  1.0,
			FlatHeightIn: 1.0,
			Qty:          int(math.Round(ringCuFt * 950.0)), // ~950 rings per cu.ft.
			Notes:        fmt.Sprintf(`Extrude & cut ~%.0f hollow 1" OD × 1" L stoneware rings (%.0f cu.ft. total bed volume, %.0f" bed depth); bisque/stoneware fire tucked inside normal studio kiln loads.`, ringCuFt*950.0, ringCuFt, bedDepthIn),
		},
		{
			PartNum:      4,
			Name:         "Hand-Formed Terracotta 3-Pass Chevron Mist Baffle Tiles",
			Shape:        "plate",
			Material:     "Unglazed Fired Terracotta Louver Slabs (3/8\" thick)",
			FlatWidthIn:  round1(bedDiaIn - 1.0),
			FlatHeightIn: 6.0,
			Qty:          12,
			Notes:        `Form 12 louvered terracotta baffle strips set at 45° alternating angles above the gravity spray rose to knock down mist droplets at <=0.020 inWC dP.`,
		},
		{
			PartNum:      6,
			Name:         "Braced 3.5 m (11.5 ft) Overhead Gravity Tank Stand Timber/Steel Legs & X-Braces",
			Shape:        "frame",
			Material:     `4"×4" Treated Cypress/Eucalyptus Timber or 2"×2" Mild Steel Angle + Diagonal X-Bracing`,
			FlatWidthIn:  42.0,
			FlatHeightIn: 138.0,
			Qty:          4,
			Notes:        fmt.Sprintf(`Cut 4× 138" (3.5 m) corner posts + 8× diagonal X-braces + bolted deck joists rated safely for %.0f kg (%.0f lbs) filled tank weight.`, tankWetKg, tankWetLbs),
		},
		{
			PartNum:      8,
			Name:         fmt.Sprintf("Tall Brick / SWG Thermal-Reheat Chimney (%.1f m / %.0f ft)", fr.ChimneyHeightM, fr.ChimneyHeightFt),
			Shape:        "masonry",
			Material:     "Fired Clay Bricks / 300 mm IS 651 SWG Flue Tiles + Mortar + Galvanized Guy-Wires",
			FlatWidthIn:  38.0,
			FlatHeightIn: chimneyHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`Build %.1f m (%.0f ft) vertical chimney on reinforced footing with cleanout door at base and lateral wall ties / 3-way guy-wires every 3 m.`, fr.ChimneyHeightM, fr.ChimneyHeightFt),
		},
	}
	if fr.SteamEnabled {
		cutList[len(cutList)-1].Name = fmt.Sprintf("Short Brick / SWG Exhaust Chimney (%.1f m / %.0f ft, Steam-Jet Driven)", fr.ChimneyHeightM, fr.ChimneyHeightFt)
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			cutList = append(cutList, CutSheetItem{
				PartNum:      11,
				Name:         fmt.Sprintf("Insulated Low-Pressure Steam Header (S1), 4.2-4.4 m Open Hydrostatic Standpipe (S2) & Drilled Brass Converging Orifice (%.2f mm)", fr.SteamNozzleThroatMM),
				Shape:        "pipe",
				Material:     `1/2" Copper/Brass Steam Header (S1) + 3/4" × 4.2-4.4 m Open Vertical Standpipe (S2) + Hand-Drilled 1/2" Brass Pipe-Cap Orifice (No CNC Bell)`,
				FlatWidthIn:  round2(math.Pi * 0.75),
				FlatHeightIn: 165.4, // 4.2 m vertical open hydrostatic standpipe
				Qty:          1,
				Notes:        fmt.Sprintf(`Drill %.2f mm center bore in standard 1/2" brass pipe cap (no CNC de Laval bell needed); plumb unvalved 4.2-4.4 m (13.8-14.4 ft) vertical open water-column standpipe S2 directly to steamer liquid space (physically caps pressure at ~6.0 psig), insulate header S1, and hydro-test to 1.5x MAWP (23 psig).`, fr.SteamNozzleThroatMM),
			})
		} else {
			cutList = append(cutList, CutSheetItem{
				PartNum:      11,
				Name:         fmt.Sprintf("Insulated Steam Header (S1) & Machined 316L de Laval Ejector Mixing Throat (%.2f mm Nozzle)", fr.SteamNozzleThroatMM),
				Shape:        "pipe",
				Material:     `1/2" Sch 40 Seamless 316L/Steel Steam Pipe + Mineral-Wool Scald Guard + 316L Mixing Diffuser`,
				FlatWidthIn:  round2(math.Pi * 0.5),
				FlatHeightIn: 36.0,
				Qty:          1,
				Notes:        fmt.Sprintf(`Machine %.2f mm choked 316L de Laval nozzle centered coaxially in the mixing throat; insulate S1 steam header with scald guards and hydro-test to 1.5x MAWP before firing.`, fr.SteamNozzleThroatMM),
			})
		}
	}

	// Electrical schedule: strictly EMPTY (0 circuits) when zero_electricity == true!
	electrical := []ElectricalCircuit{}
	if !p.ZeroElectricity {
		reqV := urban.Electrical.RequiredVoltage
		if reqV <= 0 {
			reqV = 230
		}
		electrical = []ElectricalCircuit{
			{
				ID:           "E1",
				Label:        fmt.Sprintf("Local %dV 50Hz RCD/GFCI Circuit -> 0.5 HP Domestic Monoblock Pump", reqV),
				From:         "Studio Domestic Breaker / RCD Panel",
				To:           fr.ManualPumpType,
				Voltage:      reqV,
				Amps:         round1(370.0 / float64(reqV) * 1.25),
				BreakerAmps:  10,
				WireGauge:    "1.5 mm² / 14 AWG Weatherproof Flexible Cable in Conduit",
				Conduit:      "20 mm PVC Outdoor Conduit w/ Weatherproof Isolator Switch",
				NECRef:       "IEC 60364 / Local 30mA RCD Protection",
				IsSafetyLoop: false,
				Notes:        "Powers the 0.5 HP (370 W) domestic monoblock recirculation pump; overhead gravity tank maintains flow during grid blackouts.",
			},
		}
		if !fr.SteamEnabled {
			electrical = append(electrical, ElectricalCircuit{
				ID:           "E2",
				Label:        "Local Low-dP Draft Fan (12V Solar / 230V Jua-Kali Fan)",
				From:         "12V Solar Charge Controller / Local RCD Circuit",
				To:           "Low-Static Ceramic Chimney Assist Fan (~150 W, 1.8 inWC)",
				Voltage:      12,
				Amps:         12.5,
				BreakerAmps:  20,
				WireGauge:    "4 mm² / 12 AWG DC Twin Cable (or 1.5 mm² AC)",
				Conduit:      "20 mm PVC Outdoor Conduit",
				NECRef:       "Local Low-Voltage DC / RCD Protected",
				IsSafetyLoop: false,
				Notes:        "Provides 1.8 inWC mechanical draft with shorter 4.5 m ceramic chimney; all-ceramic body stays safe even if fan stops.",
			})
		}
	}

	// Plumbing schedule P1..P5 for gravity tank + manual treadle/hand pump (or local 0.5 HP pump).
	pumpLineLabel := "Manual Treadle / Hand Pump Lift Riser -> Overhead Gravity Tank"
	pumpPressure := 8.0
	if !p.ZeroElectricity {
		pumpLineLabel = "Local 0.5 HP Monoblock Pump Lift Riser -> Overhead Gravity Tank"
		pumpPressure = 18.0
	}
	plumbing := []PlumbingLine{
		{
			ID:             "P1",
			Label:          "Makeup Water Supply (Bucket / Hose w/ Air-Gap Fill) -> Overhead Gravity Tank",
			Fluid:          "Fresh Makeup Water",
			From:           "Studio Tap / Borehole / Jerrycan Fill",
			To:             fmt.Sprintf("Elevated %.0f L %s (Top Air-Gap Fill)", fr.GravityTankLiters, prof.tankBrand),
			SizeIn:         0.75,
			Material:       `3/4" HDPE / PVC Pipe (or 20L Jerrycan Top Fill w/ Air Gap)`,
			FlowGPM:        round2(math.Max(0.35, design.Quench.EvaporatedWaterGPM*1.4)),
			PressurePSI:    15.0,
			ValvesFittings: "Top Air-Gap Fill Pipe (>=1\" above overflow rim) + Overflow Drain",
			Notes:          "Replaces evaporative quench loss (~1-2 twenty-liter jerrycans/hr) while keeping potable supply air-gapped from alkaline scrubber liquor.",
		},
		{
			ID:             "P2",
			Label:          "Brick Baffled Sump Suction -> Foot Valve -> Circulation Pump",
			Fluid:          fmt.Sprintf("Alkaline %s (%s) Scrubber Liquor (pH 7.5-8.5)", fr.AlkaliName, fr.AlkaliFormula),
			From:           "Brick Baffled Sump Clean Chamber",
			To:             fr.ManualPumpType,
			SizeIn:         1.25,
			Material:       `1-1/4" Reinforced PVC / HDPE Suction Hose + Brass Foot Valve & Mesh Basket`,
			FlowGPM:        fr.SprayFlowGPM,
			PressurePSI:    1.0,
			ValvesFittings: "Coarse Cloth/Mesh Ash Basket, Brass Foot Valve, Isolation Valve",
			Notes:          "Keep foot valve elevated 4\" off sump floor behind the weir baffle so settled wood ash and lime sediment never enter the pump.",
		},
		{
			ID:             "P3",
			Label:          pumpLineLabel,
			Fluid:          "Recirculated Alkaline Scrubber Liquor",
			From:           fr.ManualPumpType,
			To:             fmt.Sprintf("Elevated %.0f L %s @ %.1f ft (3.5 m)", fr.GravityTankLiters, prof.tankBrand, fr.TankElevationFt),
			SizeIn:         1.0,
			Material:       `1" Class-D PVC / HDPE Riser Pipe Strapped to Tank Stand`,
			FlowGPM:        round2(math.Max(fr.SprayFlowGPM, 9.2)), // ~35 L/min intermittent treadle burst rate
			PressurePSI:    pumpPressure,
			ValvesFittings: "Swing Check Valve, Union, Stand Pipe Clamps",
			Notes: func() string {
				if p.ZeroElectricity {
					return fmt.Sprintf("Closed-loop manual lift: pump %.1f min/hr on treadle/hand pump (or carry %d × 20L jerrycans/hr) to keep overhead gravity tank topped up.", fr.TreadleMinutesPerHr, fr.Jerrycans20LPerHr)
				}
				return "Automated 0.5 HP monoblock pump refill with overhead tank float/overflow bypass back to sump."
			}(),
		},
		{
			ID:             "P4",
			Label:          "Gravity Tank Downcomer -> Brass/PVC Coarse Spray Rose (Quench + Stoneware Bed)",
			Fluid:          "Gravity-Fed Alkaline Spray Liquor",
			From:           fmt.Sprintf("Elevated %.0f L %s (%.2f PSI Gravity Head)", fr.GravityTankLiters, prof.tankBrand, fr.GravityHeadPSI),
			To:             "Stage 1-2 Quench Rose & Stage 3 Wide Stoneware Bed Top Distributor",
			SizeIn:         1.0,
			Material:       `1" PVC / HDPE Gravity Downcomer + Brass/PVC Coarse Spray Rose`,
			FlowGPM:        fr.SprayFlowGPM,
			PressurePSI:    fr.GravityHeadPSI,
			ValvesFittings: "Full-Port PVC Throttle Ball Valve, Cleanout Tee, Removable Spray Rose",
			Notes:          fmt.Sprintf("Delivers %.2f L/min (%.2f GPM) at %.2f PSI gravity head — large 3-4 mm spray-rose holes resist plugging from ash or local alkali slurry.", fr.SprayFlowLPM, fr.SprayFlowGPM, fr.GravityHeadPSI),
		},
		{
			ID:             "P5",
			Label:          "Gravity Brick Channel Return & Cool-Down Sump Cleanout Drain",
			Fluid:          "Cooled Neutralized Brine Effluent (<=140 F, pH 6.5-8.5)",
			From:           "Stage 1 Quench Base & Stage 3 Stoneware Tower Base -> Brick Sump -> Drain",
			To:             "Settling Bucket / Ash Filter -> Approved Neutralized Disposal",
			SizeIn:         2.0,
			Material:       `2" Glazed Clay / PVC Gravity Return Channel + Sump Drain Plug/Valve`,
			FlowGPM:        round2(fr.SprayFlowGPM + 1.0),
			PressurePSI:    1.0,
			ValvesFittings: "Sloped Masonry Return Weir, 2\" Drain Valve, Stem Thermometer & Litmus pH Check",
			Notes:          "Cool sump <=140 F (60 C), confirm pH 6.5-8.5 with litmus strips & SG < 1.105 for multi-firing reuse, and strain wood ash before draining.",
		},
	}

	if fr.SteamEnabled {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			plumbing = append(plumbing,
				PlumbingLine{
					ID:             "S1",
					Label:          fmt.Sprintf("Insulated Low-Pressure Steam Line (S1) -> 10-15 psig PRV & Gauge -> Drilled Brass Converging Orifice (6 psig / 0.414 barg, %.1f kg/hr)", fr.SteamFlowKgHr),
					Fluid:          fmt.Sprintf("Unchoked Subsonic Saturated Steam (6 psig / 0.414 barg, %.0f m/s, Mach %.2f < 1.0 @ %.2f mm drilled brass orifice)", fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.SteamNozzleThroatMM),
					From:           "Certified 12-15 psig Commercial Catering Stainless Steamer / Low-Pressure Kettle",
					To:             "Stage 2B Simple Drilled Brass Converging Orifice & Stoneware Mixing Diffuser",
					SizeIn:         0.5,
					Material:       `1/2" Copper / Seamless Brass Steam Line + Mineral-Wool Insulation & Metal Scald Guard`,
					FlowGPM:        round2(fr.SteamWaterLHr / 60.0 / 3.785411784),
					PressurePSI:    fr.SteamPressurePSIG,
					ValvesFittings: "Factory 10-15 psig Brass Spring-Loaded PRV (w/ Pull-Ring), 0-15 psig Glycerin Gauge w/ Siphon, Bronze Throttling Valve",
					Notes:          fmt.Sprintf("1.5x MAWP (23 psig) cold hydro-tested header S1; simple drilled %.2f mm brass pipe-cap orifice needs NO CNC de Laval bell. Delivers %.3f inWC quiet subsonic draft (~52 dBA, %.1f:1 entrainment) at %.0f W electric.", fr.SteamNozzleThroatMM, fr.EjectorDraftInWC, fr.EjectorEntrainmentRatio, fr.PowerWatts),
				},
				PlumbingLine{
					ID:             "S2",
					Label:          "4.2 m (13.8 ft cold / 4.4 m hot water) Open Vertical Hydrostatic Water-Column Relief Standpipe (S2 — 6.0 psig Physical Cap)",
					Fluid:          "Open Atmospheric Water Column / Automatic Over-Pressure Steam Vent (Zero Moving Parts)",
					From:           "Certified Catering Steamer Lower Liquid Manifold (Unvalved Direct Connection)",
					To:             "Open-Top Atmospheric Overflow Elbow @ +4.2-4.4 m Elevation -> Outdoor Splash Drain",
					SizeIn:         0.75,
					Material:       `3/4" Rigid Copper / Galvanized Steel Vertical Standpipe w/ Unobstructed Open Overflow Bend`,
					FlowGPM:        0.0,
					PressurePSI:    fr.SteamPressurePSIG,
					ValvesFittings: "ZERO Shutoff Valves Permitted — Open Top Atmospheric Overflow Cup & Downward Splash Deflector",
					Notes:          "Static hydrostatic water-column safety relief (0 GPM normal flow): P_max = rho*g*h (~5.96 psig cold @ 4.2 m; 4.35-4.4 m prevents hot-water density burping at 6.0 psig). Physically caps steamer pressure with zero moving parts, backing up the factory 10-15 psig brass PRV.",
				},
				PlumbingLine{
					ID:             "P6",
					Label:          fmt.Sprintf("Open 0 psig Clean Feed Tank -> Hand-Piston / Boiler Feed Pump -> Check Valve -> Catering Steamer (+%.1f L/hr)", fr.SteamWaterLHr),
					Fluid:          "Clean Soft/Filtered Boiler Feedwater (Separate from Alkaline Sump!)",
					From:           "Open Atmospheric 0 psig Clean Feedwater Tank + Hand-Piston Feed Pump (or 0 psig Batch Refill)",
					To:             "Certified Commercial Catering Steamer Feed Inlet",
					SizeIn:         0.5,
					Material:       `1/2" Copper / 316L Seamless Feed Tubing + Bronze Spring Check Valve`,
					FlowGPM:        round2(fr.SteamWaterLHr / 60.0 / 3.785411784),
					PressurePSI:    round1(fr.SteamPressurePSIG + 5.0),
					ValvesFittings: "Fine Mesh Strainer, Manual Piston Feed Pump (or Depressurized Batch Fill), Bronze Spring Check Valve, Water-Level Sight Glass",
					Notes:          "Requires feed pressure > 6 psig (a 3.5 m / 5 PSI gravity tank cannot push water into a 6 psig steamer); bronze check valve prevents steam backflow into the 0 psig clean feed tank.",
				},
			)
		} else {
			feedSource := "Certified Non-IBR Baby Boiler (60 psig — NO Monotube Coils)"
			if fr.DraftMode == LocalDraftSteamEjectorLP {
				feedSource = "Certified Monotube Coil (<2 L) or Non-IBR Baby Boiler (20 psig)"
			}
			plumbing = append(plumbing,
				PlumbingLine{
					ID:             "S1",
					Label:          fmt.Sprintf("Insulated Live Steam Header (S1) -> PRV & Gauge -> de Laval Ejector (%.0f psig / %.1f kg/hr)", fr.SteamPressurePSIG, fr.SteamFlowKgHr),
					Fluid:          fmt.Sprintf("Saturated / Dry Motive Steam (%.0f psig / %.2f barg, %.0f m/s @ %.2f mm nozzle)", fr.SteamPressurePSIG, fr.SteamPressureBar, fr.SteamNozzleVelMS, fr.SteamNozzleThroatMM),
					From:           feedSource,
					To:             "Stage 2B 316L de Laval Steam-Jet Ejector Mixing Throat",
					SizeIn:         0.5,
					Material:       `1/2" Sch 40 Seamless 316L/Steel Pipe + Mineral-Wool Insulation & Metal Scald Guard`,
					FlowGPM:        round2(fr.SteamWaterLHr / 60.0 / 3.785411784),
					PressurePSI:    fr.SteamPressurePSIG,
					ValvesFittings: "ASME/IBR/KEBS Certified Spring-Loaded PRV (w/ Pull-Ring), 0-100 psig Glycerin Gauge w/ Siphon, Bronze Throttling Valve",
					Notes:          fmt.Sprintf("1.5x MAWP cold hydro-tested header S1; pipe PRV discharge safely down toward grade. Delivers %.3f inWC steam-jet ejector draft (%.1f:1 entrainment) at %.0f W electric.", fr.EjectorDraftInWC, fr.EjectorEntrainmentRatio, fr.PowerWatts),
				},
				PlumbingLine{
					ID:             "P6",
					Label:          fmt.Sprintf("Open 0 psig Clean Feed Tank -> Boiler Feed Pump -> Check Valve -> Steam Source (+%.1f L/hr)", fr.SteamWaterLHr),
					Fluid:          "Clean Soft/Filtered Boiler Feedwater (Separate from Alkaline Sump!)",
					From:           "Open Atmospheric 0 psig Clean Feed Tank + Manual Piston / Boiler Feed Pump",
					To:             feedSource + " Feed Inlet",
					SizeIn:         0.5,
					Material:       `1/2" Copper / 316L Seamless Feed Tubing + Bronze/316L Spring Check Valve`,
					FlowGPM:        round2(fr.SteamWaterLHr / 60.0 / 3.785411784),
					PressurePSI:    round1(fr.SteamPressurePSIG + 10.0),
					ValvesFittings: "Fine Mesh Strainer, High-Head Hand-Piston or Boiler Feed Pump, Bronze/316L Spring Check Valve, Isolation Valve, Sight Glass",
					Notes:          fmt.Sprintf("Boiler feed pump MUST exceed %.0f psig operating pressure (a 3.5 m / 5 PSI gravity tank cannot inject into a pressurized boiler!); check valve isolates the 0 psig atmospheric clean tank.", fr.SteamPressurePSIG),
				},
			)
		}
	}

	pumpHP := 0.0
	pumpVolt := 0
	if !p.ZeroElectricity {
		pumpHP = 0.5
		pumpVolt = urban.Electrical.RequiredVoltage
		if pumpVolt <= 0 {
			pumpVolt = 230
		}
	}
	pumpStation := PumpStationSpec{
		Location:          fmt.Sprintf("Beside Plastered Brick Baffled Sump below %.1f ft Overhead Tank Stand", fr.TankElevationFt),
		PumpType:          fr.ManualPumpType,
		FlowGPM:           fr.SprayFlowGPM,
		HeadFt:            round1(fr.TankElevationFt + 3.5),
		HP:                pumpHP,
		Voltage:           pumpVolt,
		SuctionSizeIn:     1.25,
		DischargeSizeIn:   1.0,
		Strainer:          "Woven Cloth / Coarse Mesh Foot-Valve Basket in Baffled Sump Clean Chamber",
		BackflowPreventer: `Engineered >=1" Open Air-Gap Top Fill into Overhead Gravity Tank (Zero Direct Pressurized Connection)`,
		DrainSpec:         "Cool Sump <=140 F (60 C) + Litmus pH 6.5-8.5 Check + Ash Settling Before Disposal",
	}

	bracing := []BracingMember{
		{
			ID:         "B1",
			Component:  "Reinforced Masonry Footing & Slab Pad",
			MemberSpec: `150 mm (6") Concrete Slab / Compacted Stone Masonry Foundation`,
			Mounting:   "Level Outdoor Grade w/ Perimeter Storm Drainage",
			LoadLbs:    totalPadLbs,
			AnchorSpec: "Continuous Mortar Bed & Footing Pads Under Brick Chambers & Tank Stand",
			Notes:      fmt.Sprintf("Supports ~%.0f lbs combined wet masonry, stoneware packing, brick sump & chimney mass without differential settling.", totalPadLbs),
		},
		{
			ID:         "B2",
			Component:  "Concentric Firebrick Flue-Wall Recuperator & Quench Buckstays",
			MemberSpec: `2"×2" Mild Steel Corner Angles + 1/2" Threaded Tie-Rods (or Buttressed Brick Piers)`,
			Mounting:   "Perimeter Clamping Around Hot Firebrick Quench & Flue Wall",
			LoadLbs:    quenchLbs,
			AnchorSpec: "Spring-Tensioned Tie-Rods Allowing Brick Thermal Expansion at 2400 F",
			Notes:      "Prevents mortar-joint hairline cracks from thermal expansion between the 2400 F flue inlet and concentric brick reheat wall.",
		},
		{
			ID:         "B3",
			Component:  "Wide Stoneware-Packed Masonry Tower & Heavy Ceramic Support Arch",
			MemberSpec: `Double-Wythe Brick / Wire-Banded SWG Wall + Firebrick Arch Packing Grate`,
			Mounting:   "Bonded Directly to Masonry Slab Footing",
			LoadLbs:    towerWetLbs,
			AnchorSpec: "External Galvanized Steel Banding Straps Every 600 mm Vertical",
			Notes:      fmt.Sprintf("Carries %.0f lbs wet tower load (including ~%.0f lbs heavy potter-extruded stoneware rings @ ~46 lb/cu.ft. + water holdup).", towerWetLbs, ringCuFt*21.0/model.KgPerLb),
		},
		{
			ID:         "B4",
			Component:  fmt.Sprintf("Elevated %.0f L Overhead Gravity Tank Stand Bracing (3.5 m / 11.5 ft)", fr.GravityTankLiters),
			MemberSpec: `4× 4"×4" Treated Hardwood/Steel Corner Posts + Full Diagonal X-Bracing + Bolted Deck Joists`,
			Mounting:   "Anchored Concrete Pier Footings (450 mm Deep)",
			LoadLbs:    tankWetLbs,
			AnchorSpec: `4× Concrete Pier Footings + M12 Through-Bolts on All Diagonal X-Braces`,
			Notes:      fmt.Sprintf("CRITICAL LIFE-SAFETY BRACING: Filled %.0f L %s weighs ~%.0f kg (%.0f lbs) at 3.5 m elevation — never mount on unbraced loose blocks or rotting poles!", fr.GravityTankLiters, prof.tankBrand, tankWetKg, tankWetLbs),
		},
		{
			ID:         "B5",
			Component:  fmt.Sprintf("Brick / Terracotta Chimney Guy-Wires & Wall Ties (%.1f m / %.0f ft)", fr.ChimneyHeightM, fr.ChimneyHeightFt),
			MemberSpec: `3-Way 6 mm Galvanized Steel Guy-Wires + Collar Bands Every 3 m (or Masonry Wall Ties)`,
			Mounting:   "Turnbuckle Anchors into Concrete Deadmen / Adjacent Masonry Wall",
			LoadLbs:    chimneyLbs,
			AnchorSpec: "Galvanized Masonry Collar Bands + Turnbuckles + Anchor Stakes",
			Notes:      fmt.Sprintf("Stabilizes the %.1f m (%.0f ft) brick/SWG chimney against storm wind overturning and seismic lateral sway.", fr.ChimneyHeightM, fr.ChimneyHeightFt),
		},
	}

	if fr.SteamEnabled {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			bracing = append(bracing, BracingMember{
				ID:         "B6",
				Component:  "Certified 6 psig Catering Steamer Cradle, Header S1 Pipe Clamps & 4.2 m Open Standpipe (S2) Straps",
				MemberSpec: `2"×2" Mild Steel Steamer Stand + Vertical Wall/Post Straps Every 1.2 m Along 4.2 m Standpipe S2 + Scald Shield`,
				Mounting:   "Bolted to Masonry Pad Outside Wet Spray Zone w/ Vertical Standpipe S2 & PRV Overflow Guide",
				LoadLbs:    48.0,
				AnchorSpec: `4× M10 Masonry Expansion Anchors + 4× 3/4" Pipe Straps Supporting 4.2 m Open Hydrostatic Standpipe S2`,
				Notes:      "Anchors certified catering steamer, holds the 4.2 m vertical open hydrostatic relief standpipe S2 plumb and kink-free, supports insulated header S1, and directs standpipe/PRV overflow safely to grade.",
			})
		} else {
			steamLoadLbs := 65.0
			if fr.DraftMode == LocalDraftSteamEjectorHP {
				steamLoadLbs = 125.0
			}
			bracing = append(bracing, BracingMember{
				ID:         "B6",
				Component:  fmt.Sprintf("Certified %.0f psig Steam Source Cradle, Header S1 Pipe Clamps & de Laval Ejector Thrust Mount", fr.SteamPressurePSIG),
				MemberSpec: `2"×2" Mild Steel Boiler Cradle + Anchored Steam Header Pipe Clamps + Sheet-Metal Scald Shield`,
				Mounting:   "Bolted to Masonry Pad Outside Wet Spray Zone w/ PRV Discharge Guide",
				LoadLbs:    steamLoadLbs,
				AnchorSpec: `4× M10 Masonry Expansion Anchors + Thermal-Expansion Pipe Guides on S1`,
				Notes:      "Anchors certified steam coil/baby boiler, absorbs de Laval ejector jet reaction thrust, supports insulated header S1, and anchors PRV vent pipe downward away from operator walkways.",
			})
		}
	}

	var acoustics []AcousticZone
	if fr.SteamEnabled {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			acoustics = []AcousticZone{
				{
					ID:             "S1",
					Location:       "6 psig Subsonic Quiet Steam-Jet Brass Orifice (Mach < 1.0, Zero Shock Screech) & Masonry Diffuser",
					Treatment:      "Thick Brick / SWG Masonry Shroud + Inherently Quiet Subsonic Lighthill v^8 Jet",
					Material:       "4.5\" Fired Brick / SWG Acoustic Mass Wall + Optional Mineral-Wool Diffuser Wrap",
					ThicknessIn:    4.5,
					AttenuationDBA: round1(math.Max(0.0, urban.Noise.BareDBA3Ft-urban.Noise.SilencedDBA3Ft)),
					Notes:          fmt.Sprintf("Whisper-quiet unchoked subsonic steam jet (%.1f dBA bare @ 3 ft via Lighthill v^8 drop & zero supersonic shock screech) -> %.1f dBA at %.0f ft property line.", urban.Noise.BareDBA3Ft, urban.Noise.PropertyLineDBA, urban.Noise.PropertyLineFt),
				},
			}
		} else {
			acoustics = []AcousticZone{
				{
					ID:             "S1",
					Location:       fmt.Sprintf("%.0f psig de Laval Steam-Jet Ejector Throat & Masonry Mixing Diffuser", fr.SteamPressurePSIG),
					Treatment:      "Thick Brick / SWG Masonry Shroud + Optional Mineral-Wool Diffuser Acoustic Wrap",
					Material:       "4.5\" Fired Brick / SWG Acoustic Mass Wall + High-Temp Mineral Wool",
					ThicknessIn:    4.5,
					AttenuationDBA: round1(math.Max(0.0, urban.Noise.BareDBA3Ft-urban.Noise.SilencedDBA3Ft)),
					Notes:          fmt.Sprintf("Modest steam-jet hiss (~%.1f dBA bare @ 3 ft — quieter than a 79 dBA 2 HP blower) -> %.1f dBA at %.0f ft property line.", urban.Noise.BareDBA3Ft, urban.Noise.PropertyLineDBA, urban.Noise.PropertyLineFt),
				},
			}
		}
	} else if p.ZeroElectricity {
		acoustics = []AcousticZone{
			{
				ID:             "S1",
				Location:       "Natural Chimney Buoyancy & Gravity Water Trickle (Zero Electric Blower)",
				Treatment:      "Inherently Whisper-Quiet Passive Ceramic Operation (No Silencer Needed)",
				Material:       "Thick Firebrick / Masonry Acoustic Mass Walls",
				ThicknessIn:    4.5,
				AttenuationDBA: 0.0,
				Notes:          fmt.Sprintf("Zero rotating blower machinery: ~%.1f dBA at 3 ft (gentle water trickle) -> %.1f dBA at %.0f ft property line.", urban.Noise.BareDBA3Ft, urban.Noise.PropertyLineDBA, urban.Noise.PropertyLineFt),
			},
		}
	} else {
		acoustics = []AcousticZone{
			{
				ID:             "S1",
				Location:       "Local 0.5 HP Domestic Monoblock Pump & Low-dP Assist Fan",
				Treatment:      "Rubber Anti-Vibration Pad + Simple Masonry Wind Screen",
				Material:       "Recycled Tire Rubber Pad + Brick Baffle Wall",
				ThicknessIn:    1.0,
				AttenuationDBA: round1(math.Max(0.0, urban.Noise.BareDBA3Ft-urban.Noise.SilencedDBA3Ft)),
				Notes:          fmt.Sprintf("Quiet local hybrid operation: %.1f dBA bare -> %.1f dBA at %.0f ft property line.", urban.Noise.BareDBA3Ft, urban.Noise.PropertyLineDBA, urban.Noise.PropertyLineFt),
			},
		}
	}
	treated3Ft := urban.Noise.BareDBA3Ft
	if (fr.SteamEnabled || !p.ZeroElectricity) && p.IncludeSilencer {
		treated3Ft = urban.Noise.SilencedDBA3Ft
	}
	acousticSummary := AcousticSummary{
		BareDBA3Ft:        urban.Noise.BareDBA3Ft,
		TreatedDBA3Ft:     treated3Ft,
		PropertyLineDBA:   urban.Noise.PropertyLineDBA,
		PropertyLineFt:    urban.Noise.PropertyLineFt,
		TotalReductionDBA: round1(math.Max(0.0, urban.Noise.BareDBA3Ft-treated3Ft)),
	}

	propMinFt := 10.0
	if p.VaporGlazeMode == VaporGlazeSalt || design.PackedBed.HClOutletPPM > 0 {
		propMinFt = 30.0
	}

	necDepth := 0.0
	necWidth := 0.0
	wiringNote := `0V / OFF-GRID (Zero Electricity): No electrical panel, VFD, or NEC 110.26 working clearance required. Maintain >=36" hot-flue clearance to combustibles (NFPA 211), >=36" service aisle around brick cleanouts & treadle pump, and inspect braced overhead gravity tank stand before every firing.`
	if fr.SteamEnabled && p.ZeroElectricity {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			wiringNote = `0V / OFF-GRID SUBSONIC STEAM-JET EJECTOR (6 psig, Zero Electricity): No electrical panel, VFD, or NEC 110.26 working clearance required. Maintain >=36" hot-flue & steam-header S1 clearance, keep the 4.2 m open hydrostatic standpipe (S2) completely unvalved & unobstructed, keep clear access to the 10-15 psig PRV pull-ring & bronze throttling valve, and inspect braced overhead tank stand before every firing.`
		} else {
			wiringNote = `0V / OFF-GRID STEAM-JET EJECTOR (Zero Electricity): No electrical panel, VFD, or NEC 110.26 working clearance required. Maintain >=36" hot-flue & steam-header S1 clearance, clear access to the certified PRV pull-ring & bronze throttling valve, and inspect braced overhead tank stand before every firing.`
		}
	} else if !p.ZeroElectricity && fr.SteamEnabled {
		necDepth = 36.0
		necWidth = 30.0
		wiringNote = `Hybrid Steam-Ejector Power (~370 W water pump only; zero electric draft fan): Mount weatherproof RCD/GFCI isolator switch in clear view of the 0.5 HP 230V 50Hz domestic pump with >=36" unobstructed dry access, well clear of live steam header S1 and PRV discharge.`
	} else if !p.ZeroElectricity {
		necDepth = 36.0
		necWidth = 30.0
		wiringNote = `Local Hybrid Power (~520 W): Mount weatherproof RCD/GFCI isolator switch in clear view of the 0.5 HP 230V 50Hz domestic pump and low-dP fan with >=36" unobstructed dry access.`
	}

	stageAlign := "passive_masonry_buoyancy"
	ventLayout := "terracotta_baffle_quench"
	if fr.SteamEnabled {
		stageAlign = "steam_jet_ejector_draft"
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			ventLayout = "subsonic_brass_orifice_condensation"
		} else {
			ventLayout = "de_laval_steam_ejector_condensation"
		}
	}

	return ShopPlan{
		Title:               fmt.Sprintf("Local Ceramic Shop Plan (%s) — %s (%.1f cu.ft.)", prof.label, p.StudioName, design.Input.KilnVolumeCuFt),
		BuildPath:           BuildPathLocalCeramic,
		IsDeluxe:            false,
		LocalDraftMode:      fr.DraftMode,
		SteamEjectorEnabled: fr.SteamEnabled,
		StageAlignment:      stageAlign,
		PackedBedLayout:     "wide_masonry_stoneware_bed",
		VenturiLayout:       ventLayout,
		CompactFootprint:    p.CompactFootprint,
		OverallLengthIn:     overallLenIn,
		OverallWidthIn:      overallWidthIn,
		OverallHeightIn:     overallHeightIn,
		Parts:               parts,
		CutList:             cutList,
		Clearances: PlanClearances{
			FlueToCombustiblesIn: 36.0,
			PropertyLineMinFt:    propMinFt,
			StackAboveRoofFt:     math.Max(6.0, round1(fr.ChimneyHeightFt-10.0)),
			ServiceAisleIn:       36.0,
			NECPanelDepthIn:      necDepth,
			NECPanelWidthIn:      necWidth,
			PumpServiceIn:        24.0,
			CleanoutSwingIn:      24.0,
			HoseBendRadiusIn:     6.0,
			WiringAccessNote:     wiringNote,
		},
		Electrical:      electrical,
		Plumbing:        plumbing,
		PumpStation:     pumpStation,
		Bracing:         bracing,
		Acoustics:       acoustics,
		AcousticSummary: acousticSummary,
		DeluxeSpec:      buildDeluxeSpec(p, design, cots, urban),
	}
}
