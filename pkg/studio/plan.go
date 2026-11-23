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

// DeluxeFabSpec summarizes the Custom Fabrication Deluxe hardware specification and cost delta.
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
}

// ShopPlan provides complete, self-contained in-browser assembly plans, numbered
// parts legend, flat-pattern cut sheets, clearances, and full systems schedules
// (electrical, plumbing, pump station, bracing, acoustics, and deluxe spec).
type ShopPlan struct {
	Title            string              `json:"title"`
	BuildPath        string              `json:"build_path"`
	IsDeluxe         bool                `json:"is_deluxe"`
	StageAlignment   string              `json:"stage_alignment"`
	PackedBedLayout  string              `json:"packed_bed_layout"`
	VenturiLayout    string              `json:"venturi_layout"`
	CompactFootprint bool                `json:"compact_footprint"`
	OverallLengthIn  float64             `json:"overall_length_in"`
	OverallWidthIn   float64             `json:"overall_width_in"`
	OverallHeightIn  float64             `json:"overall_height_in"`
	Parts            []PlanPart          `json:"parts"`
	CutList          []CutSheetItem      `json:"cut_list"`
	Clearances       PlanClearances      `json:"clearances"`
	Electrical       []ElectricalCircuit `json:"electrical"`
	Plumbing         []PlumbingLine      `json:"plumbing"`
	PumpStation      PumpStationSpec     `json:"pump_station"`
	Bracing          []BracingMember     `json:"bracing"`
	Acoustics        []AcousticZone      `json:"acoustics"`
	AcousticSummary  AcousticSummary     `json:"acoustic_summary"`
	DeluxeSpec       *DeluxeFabSpec      `json:"deluxe_spec,omitempty"`
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

	p6Valves := `2" Cleanout Tee, Swing Check / Gravity Slope (>= 1/4" per ft)`
	p6Notes := "Gravity return from Stage 1 quench bottom, Stage 3 absorber base, and stack 1/2\" condensate loop into baffled ash-settling chamber."
	if arch.Eductor.Enabled {
		p6Valves = `2" Deep U-Trap Liquid Seal (>= 8" Seal Leg) + Cleanout Plug`
		p6Notes = fmt.Sprintf("EDUCTOR POSITIVE-PRESSURE SEAL: >= 8\" U-trap water leg exceeds +%.2f inWC Stage 3 positive pressure so wet exhaust/acid mist never blows back into the vented sump.", arch.Eductor.PositivePressureInWC)
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
			FlowGPM:        recircGPM,
			PressurePSI:    2.0,
			ValvesFittings: "Bulkhead Fitting, True-Union Isolation Ball Valve, 50-Mesh Y-Strainer",
			Notes:          "Mount pump below sump minimum liquid level for positive flooded suction (mag-drive pumps must NEVER run dry).",
		},
		{
			ID:             "P3",
			Label:          "Pump Discharge -> Flow Switch -> Stage 1 Quench Manifold",
			Fluid:          "Pressurized Quench Spray Liquor",
			From:           "Recirculation Pump Discharge Manifold",
			To:             fmt.Sprintf("Stage 1 Quench Spray Nozzles (%d×)", maxInt(design.Quench.NozzleCount, 2)),
			SizeIn:         0.75,
			Material:       quenchHoseMat,
			FlowGPM:        quenchGPM,
			PressurePSI:    65.0,
			ValvesFittings: "Check Valve, 0.25 GPM Safety Flow Switch, Pressure Gauge, Throttle Valve",
			Notes:          "60-100 psig fine-atomizing header (<=200 um droplets); 0.25 GPM flow switch wired directly into E3 hardwired NFPA 86 safety loop.",
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
			PressurePSI:    40.0,
			ValvesFittings: "Needle/Ball Throttle Valve, 316L/C-276 Injection Nozzles, Union",
			Notes:          fmt.Sprintf("Delivers %.2f GPM to Stage 2 throat at L/G=%.4f for submicron fume capture.", venturiGPM, design.Venturi.LGRatio),
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
			FlowGPM:        round2(recircGPM + quenchGPM),
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
		FlowGPM:           recircGPM,
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

	// Compute exact cost delta comparing custom_fab_deluxe vs equivalent cots_no_weld configuration.
	var deluxeTotal, cotsTotal float64
	if isDeluxe {
		deluxeTotal = cots.EstimatedTotalUSD
		cotsProj := p
		cotsProj.BuildPath = BuildPathCOTSNoWeld
		cotsTotal = BuildCOTSSummary(cotsProj, design).EstimatedTotalUSD
	} else {
		cotsProj := p
		cotsProj.BuildPath = BuildPathCOTSNoWeld
		if p.BuildPath == BuildPathCOTSNoWeld {
			cotsTotal = cots.EstimatedTotalUSD
		} else {
			cotsTotal = BuildCOTSSummary(cotsProj, design).EstimatedTotalUSD
		}
		dlxProj := p
		dlxProj.BuildPath = BuildPathCustomFabDeluxe
		deluxeTotal = BuildCOTSSummary(dlxProj, design).EstimatedTotalUSD
	}
	costDelta := round2(deluxeTotal - cotsTotal)

	vesselMat := `TIG-Welded Heavy-Gauge 316L SS (11 ga / Sch 10S) Quench & Venturi + Filament-Wound Derakane Vinyl-Ester FRP Packed Column w/ Structural Gusset Rings`
	if p.CorrosionTier == CorrosionTierHeavyC276FRP {
		vesselMat = `TIG-Welded Heavy-Gauge Alloy C-276 (11 ga / Sch 10S) Quench & Venturi + Filament-Wound Derakane Vinyl-Ester FRP Packed Column w/ Structural Gusset Rings`
	}

	summary := fmt.Sprintf(
		"Turnkey shop-fabricated Deluxe package (+$%.2f vs No-Weld COTS): TIG-welded heavy-gauge hot stages with ANSI/ASME B16.5 Class 150 lap-joint flanges, Derakane FRP column with bolted manway & removable demister cassette, welded 2\"×2\"×3/16\" steel forklift skid, UL 508A NEMA 4X panel, and 2\" Rockwool acoustic baffle box + tuned stack silencer (-22 dBA -> %.1f dBA @ %.0f ft).",
		costDelta, urban.Noise.PropertyLineDBA, urban.Noise.PropertyLineFt,
	)

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
		Highlights: []string{
			"Back-purged TIG heavy-gauge 316L SS (or Alloy C-276) quench & venturi with ANSI B16.5 Class 150 lap-joint flanges",
			"2× borosilicate sight-glass ports + 2× bolted NFPA 91 quick-release cleanout doors + 1\" ceramic-fiber/SS hot-flue jacket",
			"Filament-wound Derakane vinyl-ester FRP column with structural gusset rings, bolted 12\" packing manway & removable demister cassette",
			"Sloped-bottom baffled settling sump (2\" drain ball valve) + heavy-duty sealless PVDF magnetic-drive pump station",
			"Welded 2\"×2\"×3/16\" structural steel forklift skid + diagonal K-bracing + concrete wedge anchors + flexible PTFE/EPDM vibration boots",
			"2\" Rockwool acoustic blower baffle enclosure w/ cooling-air labyrinth + tuned absorptive stack silencer (-22 dBA combined attenuation)",
			"UL 508A-style NEMA 4X pre-wired panel with through-door NEC 430.102 disconnect, 3% line reactor, HF-GFCI & FM/UL 200 F high-limit safety relay",
		},
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
		plateSideIn := round1(math.Max(16.0, math.Ceil(math.Sqrt(float64(arch.ParallelBed.PipeCount)))*(arch.ParallelBed.PipeNominalIn+1.25)+4.0))
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
		flangeOD := round1(quenchDiaIn + 6.0)
		boxSide := round1(math.Max(26.0, cots.BlowerVFD.NominalIn+14.0))
		kBraceLen := round1(math.Hypot(colDiaIn+8.0, colHIn*0.45))
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
				Notes:        fmt.Sprintf(`Cut 4× vertical posts @ %.1f" H + 8× diagonal K-braces @ %.1f" L (45° mitered ends) around Stage 3 column.`, colHIn, kBraceLen),
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
