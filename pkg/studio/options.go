package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
)

// EductorAnalysis quantifies the Dry-Fan Vertical-Downflow Eductor architecture:
// Quench exits out & down to a flanged Tee where a clean dry ambient-air radial fan
// drives motive air straight down into a vertical static venturi -> 90-deg cyclonic
// sweep elbow -> bottom plenum of Stage 3 Packed Bed -> top stack (no wet ID blower).
type EductorAnalysis struct {
	Enabled               bool     `json:"enabled"`
	Feasible              bool     `json:"feasible"`
	QuenchedACFM          float64  `json:"quenched_acfm"`
	MotiveAirCFM          float64  `json:"motive_air_cfm"`
	EntrainmentRatio      float64  `json:"entrainment_ratio"`
	TotalMixedACFM        float64  `json:"total_mixed_acfm"`
	MixedTempF            float64  `json:"mixed_temp_f"`
	ThermalMarginGainF    float64  `json:"thermal_margin_gain_f"`
	DryFanStaticInWC      float64  `json:"dry_fan_static_inwc"`
	DryFanShaftHP         float64  `json:"dry_fan_shaft_hp"`
	DryFanMotorHP         float64  `json:"dry_fan_motor_hp"`
	DryFanCostUSD         float64  `json:"dry_fan_cost_usd"`
	WetBlowerSavedUSD     float64  `json:"wet_blower_saved_usd"`
	NetBlowerDeltaUSD     float64  `json:"net_blower_delta_usd"`
	EnlargedBedDiameterIn float64  `json:"enlarged_bed_diameter_in"`
	EnlargedBedFloodPct   float64  `json:"enlarged_bed_flood_pct"`
	PositivePressureInWC  float64  `json:"positive_pressure_inwc"`
	EductorEfficiencyPct  float64  `json:"eductor_efficiency_pct"`
	CyclonicSweepElbowIn  float64  `json:"cyclonic_sweep_elbow_in"`
	Summary               string   `json:"summary"`
	Warnings              []string `json:"warnings"`
}

// ParallelBedAnalysis quantifies replacing a single large industrial packed column
// with a shell-less Uniseal flat-plate bundle of N parallel 2", 3", or 4" stock
// plumbing pipes (NO external large pipe shell or industrial flanges) and packing
// sized to keep wall-channeling ratio D/d_p >= 7.5.
type ParallelBedAnalysis struct {
	Enabled             bool     `json:"enabled"`
	ShelllessPlenum     bool     `json:"shellless_plenum"`
	PipeMaterial        string   `json:"pipe_material"`
	PipeNominalIn       float64  `json:"pipe_nominal_in"`
	PipeActualIDIn      float64  `json:"pipe_actual_id_in"`
	PipeCount           int      `json:"pipe_count"`
	PackingDiameterIn   float64  `json:"packing_diameter_in"`
	WallRatioDToDp      float64  `json:"wall_ratio_d_to_dp"`
	BedDPInWC           float64  `json:"bed_dp_inwc"`
	SingleColumnDPInWC  float64  `json:"single_column_dp_inwc"`
	ActualFloodPct      float64  `json:"actual_flood_pct"`
	SingleColumnCostUSD float64  `json:"single_column_cost_usd"`
	ManifoldCostUSD     float64  `json:"manifold_cost_usd"`
	CostDeltaUSD        float64  `json:"cost_delta_usd"`
	Summary             string   `json:"summary"`
	Notes               []string `json:"notes"`
}

// ParallelVenturiAnalysis quantifies splitting Stage 2 into N (2-4) parallel smaller
// venturi throats built from off-the-shelf 316L sanitary concentric reducers.
type ParallelVenturiAnalysis struct {
	Enabled          bool     `json:"enabled"`
	ThroatCount      int      `json:"throat_count"`
	ThroatDiameterIn float64  `json:"throat_diameter_in"`
	InletFittingIn   float64  `json:"inlet_fitting_in"`
	ThroatDPInWC     float64  `json:"throat_dp_inwc"`
	SingleDPInWC     float64  `json:"single_dp_inwc"`
	PluggingRisk     string   `json:"plugging_risk"`
	SingleCostUSD    float64  `json:"single_cost_usd"`
	MultiCostUSD     float64  `json:"multi_cost_usd"`
	CostDeltaUSD     float64  `json:"cost_delta_usd"`
	TurndownBenefit  string   `json:"turndown_benefit"`
	Summary          string   `json:"summary"`
	Notes            []string `json:"notes"`
}

// BlowerDriveAnalysis evaluates the selected Phase 8 blower / motive drive architecture
// (industrial radial+VFD, side-channel ring blower, dual 3-stage tangential bypass blowers,
// shop-compressor twin-fluid air-atomize + medium-pressure radial blower, or compressor eductor).
type BlowerDriveAnalysis struct {
	Option             string   `json:"option"`
	Label              string   `json:"label"`
	FanCostUSD         float64  `json:"fan_cost_usd"`
	BaselineFanCostUSD float64  `json:"baseline_fan_cost_usd"`
	CostDeltaUSD       float64  `json:"cost_delta_usd"`
	RequiredStaticInWC float64  `json:"required_static_inwc"`
	ThroatVelocityMS   float64  `json:"throat_velocity_m_s"`
	DropletUm          float64  `json:"droplet_um"`
	MotorHP            float64  `json:"motor_hp"`
	Voltage            int      `json:"voltage"`
	Feasible           bool     `json:"feasible"`
	Summary            string   `json:"summary"`
	Caveats            []string `json:"caveats"`
}

// CompressorAnalysis evaluates a studio's existing shop air compressor in BOTH roles:
// (1) driving main scrubber airflow via an air-amplifier eductor (usually infeasible at
// studio scale due to backpressure entrainment collapse), and (2) powering a $75 twin-fluid
// air-atomizing venturi nozzle (25 um droplets, ~4-6 SCFM @ 40 PSI, slashing venturi dP
// and blower cost by $1,000+). Always evaluated on every StudioEvaluation.
type CompressorAnalysis struct {
	AvailableSCFM       float64 `json:"available_scfm"`
	EductorSCFMNeeded   float64 `json:"eductor_scfm_needed"`
	EductorHPNeeded     float64 `json:"eductor_hp_needed"`
	EductorFeasible     bool    `json:"eductor_feasible"`
	EductorVerdict      string  `json:"eductor_verdict"`
	AtomizeSCFMNeeded   float64 `json:"atomize_scfm_needed"`
	AtomizeFeasible     bool    `json:"atomize_feasible"`
	AtomizeDropletUm    float64 `json:"atomize_droplet_um"`
	AtomizeDPSavedInWC  float64 `json:"atomize_dp_saved_inwc"`
	AtomizeCostSavedUSD float64 `json:"atomize_cost_saved_usd"`
	AtomizeVerdict      string  `json:"atomize_verdict"`
}

// PackingMediaAnalysis evaluates Stage 3 random packing media choices:
// commercial PP Pall rings, bulk aquaculture bio-balls, or DIY potter-extruded stoneware
// ceramic Raschig rings (honestly computing higher irrigated Ergun bed dP and weight).
type PackingMediaAnalysis struct {
	Option               string  `json:"option"`
	Label                string  `json:"label"`
	DiameterIn           float64 `json:"diameter_in"`
	Voidage              float64 `json:"voidage"`
	BedDPInWC            float64 `json:"bed_dp_inwc"`
	MediaCostUSD         float64 `json:"media_cost_usd"`
	BaselineMediaCostUSD float64 `json:"baseline_media_cost_usd"`
	CostDeltaUSD         float64 `json:"cost_delta_usd"`
	Summary              string  `json:"summary"`
}

// ArchitectureOptionsResult summarizes all Phase 7 & Phase 8 selectable cost, time,
// footprint, blower drive, packing media, compressor, and stage-design options.
type ArchitectureOptionsResult struct {
	StageAlignment      string                  `json:"stage_alignment"`
	PackedBedLayout     string                  `json:"packed_bed_layout"`
	VenturiLayout       string                  `json:"venturi_layout"`
	CompactFootprint    bool                    `json:"compact_footprint"`
	SharedBaffledSump   bool                    `json:"shared_baffled_sump"`
	ControlTier         string                  `json:"control_tier"`
	BaselineCostUSD     float64                 `json:"baseline_cost_usd"`
	SelectedCostUSD     float64                 `json:"selected_cost_usd"`
	CostDeltaUSD        float64                 `json:"cost_delta_usd"`
	ControlDeltaUSD     float64                 `json:"control_delta_usd"`
	SumpDeltaUSD        float64                 `json:"sump_delta_usd"`
	FootprintLengthFt   float64                 `json:"footprint_length_ft"`
	FootprintWidthFt    float64                 `json:"footprint_width_ft"`
	FootprintAreaSqFt   float64                 `json:"footprint_area_sqft"`
	BaselineLengthFt    float64                 `json:"baseline_length_ft"`
	FootprintSavingsPct float64                 `json:"footprint_savings_pct"`
	Eductor             EductorAnalysis         `json:"eductor"`
	ParallelBed         ParallelBedAnalysis     `json:"parallel_bed"`
	ParallelVenturi     ParallelVenturiAnalysis `json:"parallel_venturi"`
	BlowerDrive         BlowerDriveAnalysis     `json:"blower_drive"`
	Compressor          CompressorAnalysis      `json:"compressor"`
	PackingMedia        PackingMediaAnalysis    `json:"packing_media"`
	SavingsSummary      []string                `json:"savings_summary"`
	Warnings            []string                `json:"warnings"`
}

var stdNemaMotorHPs = []float64{
	0.25, 0.33, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 7.5, 10.0, 15.0, 20.0,
}

func pickNemaMotorHP(shaftHP float64) float64 {
	req := shaftHP * 1.20
	for _, hp := range stdNemaMotorHPs {
		if hp >= req {
			return hp
		}
	}
	return math.Ceil(req/5.0) * 5.0
}

// dryFanUnitCost returns budgetary COTS pricing for a clean dry-air steel/aluminum
// radial pressure blower + VFD in NEMA 4X enclosure (significantly cheaper than wet FRP/316L).
func dryFanUnitCost(hp float64) (fanCost, vfdCost, weightKg float64) {
	const nema4X = 165.0
	switch {
	case hp <= 0.5:
		return 345, 175 + nema4X, 18.0
	case hp <= 1.0:
		return 460, 215 + nema4X, 24.0
	case hp <= 1.5:
		return 585, 265 + nema4X, 30.0
	case hp <= 2.0:
		return 695, 310 + nema4X, 36.0
	case hp <= 3.0:
		return 885, 395 + nema4X, 46.0
	case hp <= 5.0:
		return 1180, 520 + nema4X, 62.0
	default:
		return round2(1180 + 210*(hp-5.0)), round2(520 + nema4X + 110*(hp-5.0)), round2(62.0 + 10.0*(hp-5.0))
	}
}

// mediumPressureAirAtomizeBlowerCost returns budgetary COTS pricing for the medium-pressure
// cast-aluminum radial blower (~10-14 inWC) + simple VFD/NEMA 4X + $75 twin-fluid air-atomizing nozzle.
func mediumPressureAirAtomizeBlowerCost(project StudioProject, hp float64) (totalCost, weightKg float64) {
	const (
		twinFluidNozzleCost = 75.0 // COTS 316L twin-fluid air-atomizing venturi nozzle + 40 PSI regulator
	)
	var blowerCost, vfdDamperCost float64
	switch {
	case hp <= 0.5:
		blowerCost, vfdDamperCost, weightKg = 295.0, 255.0, 15.0
	case hp <= 1.5:
		blowerCost, vfdDamperCost, weightKg = 360.0, 315.0, 20.0
	case hp <= 2.0:
		blowerCost, vfdDamperCost, weightKg = 440.0, 360.0, 26.0
	default:
		blowerCost = round2(440.0 + 140.0*(hp-2.0))
		vfdDamperCost = round2(360.0 + 95.0*(hp-2.0))
		weightKg = round2(26.0 + 7.0*(hp-2.0))
	}
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		blowerCost = round2(blowerCost * 1.30)
	}
	return round2(blowerCost + vfdDamperCost + twinFluidNozzleCost), weightKg
}

func baselineProjectFor(p StudioProject) StudioProject {
	baseProj := p.Normalize()
	baseProj.StageAlignment = StageAlignmentInlineIDBlower
	baseProj.PackedBedLayout = PackedBedLayoutSingleColumn
	baseProj.VenturiLayout = VenturiLayoutSingleThroat
	baseProj.CompactFootprint = false
	baseProj.SharedBaffledSump = true
	baseProj.ControlTier = ControlTierRPiModbus
	baseProj.BlowerDriveOption = BlowerDriveIndustrialRadialVFD
	baseProj.ParallelPipeSizeIn = 4.0
	baseProj.PackingMediaOption = PackingMediaPPPallRings
	return baseProj
}

func resolveBaselineDesign(p StudioProject, activeDesign model.ScrubberDesign) model.ScrubberDesign {
	if p.BlowerDriveOption != BlowerDriveCompressorAirAtomize {
		return activeDesign
	}
	baseProj := baselineProjectFor(p)
	if baseDesign, err := scrubber.NewDesigner().Design(baseProj.EffectiveDesignInput()); err == nil {
		return baseDesign
	}
	return activeDesign
}

// solveMixedAirState solves the psychrometric equilibrium temperature T_mix [K],
// mixed volumetric flow Q_mix [m^3/s], mixed density rho_mix [kg/m^3], and mixed
// gas mass flow m_mix [kg/s] when quenched flue gas mixes with dry ambient motive air.
func solveMixedAirState(quenched model.GasComposition, qMotiveM3S float64, pAmbPa float64) (tMixK model.Kelvin, qMixM3S, rhoMix, mMix float64) {
	tAmbK := float64(model.FahrenheitToKelvin(70.0)) // conservative studio ambient for thermal margin
	tQuenchK := float64(quenched.Temperature)
	if tQuenchK <= tAmbK+1.0 {
		tQuenchK = tAmbK + 1.0
	}
	if pAmbPa <= 0 {
		pAmbPa = float64(model.StandardPressurePa)
	}

	const (
		mwAir = 0.028964
		mwH2O = 0.01801528
		cpAir = 1006.0    // J/(kg*K)
		cpVap = 1900.0    // J/(kg*K)
		hfg   = 2360000.0 // latent heat of condensation ~130-150 F [J/kg]
	)
	rhoAmb := (pAmbPa * mwAir) / (model.UniversalGasConst * tAmbK)
	mMotive := qMotiveM3S * rhoAmb
	nMotive := mMotive / mwAir

	nQuench := quenched.MolarFlowMolPerSec
	nH2OIn := nQuench * quenched.H2O
	nDryQuench := math.Max(1e-6, nQuench-nH2OIn)
	mH2OIn := nH2OIn * mwH2O
	mQuench := float64(quenched.MassFlow)
	mDryQuench := math.Max(1e-6, mQuench-mH2OIn)

	// Energy residual f(T) = heat released by quench cooling + condensation - heat absorbed by dry motive air.
	// At T = tAmbK, f(T) > 0; at T = tQuenchK, f(T) < 0.
	nDryTotal := nDryQuench + nMotive
	evalState := func(tK float64) (res float64, nH2OVap float64) {
		pSat := waterPsatSafePa(model.Kelvin(tK))
		if pSat > 0.95*pAmbPa {
			pSat = 0.95 * pAmbPa
		}
		nH2OSat := nDryTotal * pSat / math.Max(pAmbPa-pSat, 100.0)
		nH2OVap = math.Min(nH2OIn, nH2OSat)
		mCond := math.Max(0.0, (nH2OIn-nH2OVap)*mwH2O)
		qRel := (mDryQuench*1040.0+mH2OIn*cpVap)*(tQuenchK-tK) + mCond*hfg
		qAbs := mMotive * cpAir * (tK - tAmbK)
		return qRel - qAbs, nH2OVap
	}

	lo, hi := tAmbK, tQuenchK
	var nH2OFinal float64
	for i := 0; i < 40; i++ {
		mid := 0.5 * (lo + hi)
		r, nVap := evalState(mid)
		nH2OFinal = nVap
		if r > 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	tFinal := 0.5 * (lo + hi)
	nTotal := nDryTotal + nH2OFinal
	mMix = mDryQuench + mMotive + nH2OFinal*mwH2O
	qMixM3S = nTotal * model.UniversalGasConst * tFinal / pAmbPa
	rhoMix = mMix / math.Max(qMixM3S, 1e-6)
	return model.Kelvin(tFinal), qMixM3S, rhoMix, mMix
}

func analyzeEductor(project StudioProject, design model.ScrubberDesign) EductorAnalysis {
	enabled := project.StageAlignment == StageAlignmentEductorDryFan
	quenchedQ := float64(design.QuenchedGas.VolumetricFlow)
	quenchedACFM := design.QuenchedGas.VolumetricFlow.CFM()
	if quenchedQ <= 0 || quenchedACFM <= 0 {
		quenchedACFM = math.Max(50.0, design.Blower.FlowACFM)
		quenchedQ = float64(model.CFMToCubicMetersPerSec(quenchedACFM))
	}

	pAmb := float64(design.Input.AmbientPressurePa)
	if pAmb <= 0 {
		pAmb = float64(model.StandardPressurePa)
	}

	// Upstream suction load the eductor must draw (hot flue duct + quench + target kiln draft).
	suctionInWC := design.Blower.Breakdown.DuctInWC + design.Blower.Breakdown.QuenchInWC + design.Blower.Breakdown.KilnDraftInWC
	if suctionInWC < 0.5 {
		suctionInWC = 0.8
	}

	// Initial entrainment ratio R = Q_motive / Q_quenched (~1.25-1.55 depending on positive bed load).
	basePosInWC := design.PackedBed.TotalPressureDropInWC + 0.60 // +0.35 sweep elbow + 0.25 stack exit
	entrainRatio := math.Max(1.25, math.Min(1.60, 1.25+0.08*basePosInWC))
	qMotiveM3S := entrainRatio * quenchedQ
	motiveCFM := model.CubicMetersPerSec(qMotiveM3S).CFM()

	// Exact psychrometric mixed state entering the 90-deg sweep elbow & Stage 3 packed bed.
	tMixK, qMixM3S, rhoMix, mMix := solveMixedAirState(design.QuenchedGas, qMotiveM3S, pAmb)
	mixedTempF := round1(tMixK.Fahrenheit())
	quenchTempF := design.Quench.OutletTemp.Fahrenheit()
	if quenchTempF <= 0 {
		quenchTempF = 184.0
	}
	thermalMarginGainF := round1(math.Max(0.0, quenchTempF-mixedTempF))
	totalMixedACFM := round1(model.CubicMetersPerSec(qMixM3S).CFM())

	// Size enlarged Stage 3 packed column for mixed flow using Eckert flooding (<= 75% flood).
	fp := design.Input.PackingFactor
	if fp <= 0 {
		fp = 170.0
	}
	rhoL := float64(model.WaterDensity)
	qRecircMix := math.Max(0.0028*qMixM3S, float64(design.PackedBed.RecirculationFlow))
	vFloodMix := 2.4
	if vf, err := scrubber.EckertFloodingVelocity(qRecircMix*rhoL, mMix, rhoMix, rhoL, fp, 0.001); err == nil && vf > 0 {
		vFloodMix = float64(vf)
	}
	// Pick the smallest COTS column in the catalog whose actual bore stays safely <= 72% of Eckert flood (strictly <= 75%).
	catalog := packedCPVCCatalog
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		catalog = packedFRPCatalog
	}
	enlargedDiamIn := catalog[len(catalog)-1].nominalIn
	enlargedIDIn := catalog[len(catalog)-1].actualID
	for _, spec := range catalog {
		dM := float64(model.InchesToMeters(spec.actualID))
		area := math.Pi * dM * dM / 4.0
		pct := (qMixM3S / area) / vFloodMix * 100.0
		if pct <= 72.0 {
			enlargedDiamIn = spec.nominalIn
			enlargedIDIn = spec.actualID
			break
		}
	}
	enlargedArea := math.Pi * math.Pow(float64(model.InchesToMeters(enlargedIDIn)), 2) / 4.0
	vSupMix := qMixM3S / math.Max(enlargedArea, 1e-6)
	enlargedFloodPct := round1((vSupMix / vFloodMix) * 100.0)

	// Recompute irrigated Ergun bed dP + demister dP + 90-deg sweep elbow + stack exit for positive pressure.
	packDiamM := float64(design.Input.PackingDiameter)
	if packDiamM <= 0 {
		packDiamM = 0.0254
	}
	voidage := design.Input.PackingVoidage
	if voidage <= 0 || voidage >= 1 {
		voidage = 0.90
	}
	packHeightM := float64(design.PackedBed.PackingHeight)
	if packHeightM <= 0 {
		packHeightM = 0.65
	}
	muG := float64(design.QuenchedGas.Viscosity)
	if muG <= 0 {
		muG = 1.85e-5
	}
	specArea := 207.0 * (0.0254 / packDiamM)
	sauterDp := 6.0 * (1.0 - voidage) / specArea
	bedDPPa := 200.0
	if dryDP, err := scrubber.ErgunPressureDrop(sauterDp, voidage, vSupMix, rhoMix, muG, packHeightM); err == nil {
		bedDPPa = float64(dryDP) * 1.70
	}
	demisterDPPa := math.Max(35.0, 55.0*0.5*rhoMix*vSupMix*vSupMix)
	sweepElbowPa := float64(model.InWCToPascal(0.40))
	stackExitPa := float64(model.InWCToPascal(0.25))
	positivePressureInWC := round2(model.Pascal(bedDPPa + demisterDPPa + sweepElbowPa + stackExitPa).InWC())

	// Jet-eductor momentum/energy balance (ASHRAE / Karassik jet-pump efficiency eta_ed = 21%).
	// eta_ed = (Q_s * dP_lift) / (Q_m * (P_m - P_out)) -> P_m = P_out + (Q_s/Q_m)*(dP_lift/eta_ed) + throat_spray_loss.
	const etaEductor = 0.21
	dpLiftInWC := suctionInWC + positivePressureInWC
	throatSprayMixLossInWC := 3.5 // wet particulate atomization mixing loss in vertical downflow throat
	dryFanStaticInWC := round1(positivePressureInWC + (1.0/entrainRatio)*(dpLiftInWC/etaEductor) + throatSprayMixLossInWC)

	const dryFanEff = 0.65
	dryFanShaftHP := round2((motiveCFM * dryFanStaticInWC) / (6356.0 * dryFanEff))
	dryFanMotorHP := pickNemaMotorHP(dryFanShaftHP)

	fanCost, vfdCost, _ := dryFanUnitCost(dryFanMotorHP)
	dryFanCostUSD := round2(fanCost + vfdCost)
	wetBlowerPart := selectWetBlowerVFD(project, resolveBaselineDesign(project, design))
	wetBlowerSavedUSD := wetBlowerPart.UnitCostUSD
	netBlowerDeltaUSD := round2(dryFanCostUSD - wetBlowerSavedUSD)

	sweepElbowIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()/2.0)*2.0)
	feasible := dryFanMotorHP <= 10.0 && enlargedFloodPct <= 75.0

	warnings := []string{
		fmt.Sprintf("OUTDOOR-ONLY POSITIVE PRESSURE WARNING: Downstream of the vertical downflow eductor (90-deg sweep elbow, Stage 3 packed bed, and stack) operates at +%.2f inWC POSITIVE pressure (unlike negative-pressure ID operation). Any loose flange or gasket leaks wet exhaust/HCl OUTWARD — site strictly outdoors with bolted full-face PTFE gaskets and soap-bubble pressure test before firing!", positivePressureInWC),
	}
	if project.VaporGlazeMode == VaporGlazeSalt {
		warnings = append(warnings, "Salt Mode Pressure Seal Check: Verify all CPVC/FRP flanges and sump water-seal trap depth (>= 6\" liquid seal head > positive pressure) so positive pressure never blows acid mist back through the sump drain.")
	}
	if !feasible {
		warnings = append(warnings, fmt.Sprintf("High motive airflow (%.0f CFM @ %.1f inWC -> %.1f HP) makes an inline ID blower more energy-efficient at this kiln scale.", motiveCFM, dryFanStaticInWC, dryFanMotorHP))
	}

	summary := fmt.Sprintf(
		"Clean dry radial fan (%.0f CFM @ %.1f inWC, %.2g HP, $%.0f vs $%.0f wet blower) pushes ambient air straight down into a vertical static venturi -> %.0f\" 90-deg cyclonic sweep elbow -> bottom of enlarged %.0f\" packed bed (%.1f%% flood). Mixing cools Stage 3 inlet from %.0f F to %.1f F (+%.1f F CPVC thermal margin), eliminates wet-blower corrosion completely, and shifts Stage 3 to +%.2f inWC positive pressure (outdoor-only).",
		motiveCFM, dryFanStaticInWC, dryFanMotorHP, dryFanCostUSD, wetBlowerSavedUSD,
		sweepElbowIn, enlargedDiamIn, enlargedFloodPct, quenchTempF, mixedTempF, thermalMarginGainF, positivePressureInWC,
	)

	return EductorAnalysis{
		Enabled:               enabled,
		Feasible:              feasible,
		QuenchedACFM:          round1(quenchedACFM),
		MotiveAirCFM:          round1(motiveCFM),
		EntrainmentRatio:      round2(entrainRatio),
		TotalMixedACFM:        totalMixedACFM,
		MixedTempF:            mixedTempF,
		ThermalMarginGainF:    thermalMarginGainF,
		DryFanStaticInWC:      dryFanStaticInWC,
		DryFanShaftHP:         dryFanShaftHP,
		DryFanMotorHP:         dryFanMotorHP,
		DryFanCostUSD:         dryFanCostUSD,
		WetBlowerSavedUSD:     wetBlowerSavedUSD,
		NetBlowerDeltaUSD:     netBlowerDeltaUSD,
		EnlargedBedDiameterIn: enlargedDiamIn,
		EnlargedBedFloodPct:   enlargedFloodPct,
		PositivePressureInWC:  positivePressureInWC,
		EductorEfficiencyPct:  round1(etaEductor * 100.0),
		CyclonicSweepElbowIn:  sweepElbowIn,
		Summary:               summary,
		Warnings:              warnings,
	}
}

type parallelPipeGeom struct {
	nominalIn     float64
	actualIDIn    float64
	packingDiamIn float64
	fpSmall       float64
	epsSmall      float64
	costPerFt     float64
	packUnitCost  float64
	pipeMaterial  string
	isPVC         bool
}

func resolveParallelPipeGeom(project StudioProject, eductor EductorAnalysis) parallelPipeGeom {
	sizeIn := project.ParallelPipeSizeIn
	if sizeIn != 2.0 && sizeIn != 3.0 && sizeIn != 4.0 {
		sizeIn = 4.0
	}
	usePVC := eductor.Enabled && eductor.MixedTempF <= 140.0 && project.CorrosionTier != CorrosionTierHeavyC276FRP

	switch sizeIn {
	case 2.0:
		if usePVC {
			return parallelPipeGeom{
				nominalIn:     2.0,
				actualIDIn:    2.067,
				packingDiamIn: 0.25, // D/d_p = 8.27 >= 7.5
				fpSmall:       520.0,
				epsSmall:      0.82,
				costPerFt:     1.35,
				packUnitCost:  112.0,
				pipeMaterial:  "Schedule 40 PVC (ASTM D1785, <=140 F Eductor-Cooled)",
				isPVC:         true,
			}
		}
		return parallelPipeGeom{
			nominalIn:     2.0,
			actualIDIn:    1.939,
			packingDiamIn: 0.25, // D/d_p = 7.76 >= 7.5
			fpSmall:       520.0,
			epsSmall:      0.82,
			costPerFt:     4.20,
			packUnitCost:  112.0,
			pipeMaterial:  "Schedule 80 CPVC (ASTM F441, 200 F max)",
			isPVC:         false,
		}
	case 3.0:
		if usePVC {
			return parallelPipeGeom{
				nominalIn:     3.0,
				actualIDIn:    3.068,
				packingDiamIn: 0.375, // D/d_p = 8.18 >= 7.5
				fpSmall:       390.0,
				epsSmall:      0.85,
				costPerFt:     2.25,
				packUnitCost:  104.0,
				pipeMaterial:  "Schedule 40 PVC (ASTM D1785, <=140 F Eductor-Cooled)",
				isPVC:         true,
			}
		}
		return parallelPipeGeom{
			nominalIn:     3.0,
			actualIDIn:    2.900,
			packingDiamIn: 0.375, // D/d_p = 7.73 >= 7.5
			fpSmall:       390.0,
			epsSmall:      0.85,
			costPerFt:     6.20,
			packUnitCost:  104.0,
			pipeMaterial:  "Schedule 80 CPVC (ASTM F441, 200 F max)",
			isPVC:         false,
		}
	default: // 4.0" sweet spot
		if usePVC {
			return parallelPipeGeom{
				nominalIn:     4.0,
				actualIDIn:    4.026,
				packingDiamIn: 0.50, // D/d_p = 8.05 >= 7.5
				fpSmall:       310.0,
				epsSmall:      0.87,
				costPerFt:     3.20,
				packUnitCost:  98.0,
				pipeMaterial:  "Schedule 40 PVC (ASTM D1785, <=140 F Eductor-Cooled)",
				isPVC:         true,
			}
		}
		return parallelPipeGeom{
			nominalIn:     4.0,
			actualIDIn:    3.826,
			packingDiamIn: 0.50, // D/d_p = 7.65 >= 7.5
			fpSmall:       310.0,
			epsSmall:      0.87,
			costPerFt:     8.50,
			packUnitCost:  98.0,
			pipeMaterial:  "Schedule 80 CPVC (ASTM F441, 200 F max)",
			isPVC:         false,
		}
	}
}

func analyzeParallelBed(project StudioProject, design model.ScrubberDesign, eductor EductorAnalysis) ParallelBedAnalysis {
	p := project.Normalize()
	enabled := p.PackedBedLayout == PackedBedLayoutParallelManifold

	// Determine active volumetric flow & gas properties (accounts for eductor mixed flow if both are selected).
	qGas := float64(design.QuenchedGas.VolumetricFlow)
	gMass := float64(design.QuenchedGas.MassFlow)
	rhoG := float64(design.QuenchedGas.Density)
	muG := float64(design.QuenchedGas.Viscosity)
	if qGas <= 0 {
		qGas = 0.10
	}
	if gMass <= 0 {
		gMass = 0.10
	}
	if rhoG <= 0 {
		rhoG = 0.90
	}
	if muG <= 0 {
		muG = 1.85e-5
	}
	if eductor.Enabled {
		pAmb := float64(design.Input.AmbientPressurePa)
		if pAmb <= 0 {
			pAmb = float64(model.StandardPressurePa)
		}
		qMotive := float64(model.CFMToCubicMetersPerSec(eductor.MotiveAirCFM))
		_, qMix, rhoMix, mMix := solveMixedAirState(design.QuenchedGas, qMotive, pAmb)
		qGas, rhoG, gMass = qMix, rhoMix, mMix
	}

	geom := resolveParallelPipeGeom(p, eductor)
	wallRatio := round2(geom.actualIDIn / geom.packingDiamIn)
	pipeIDM := float64(model.InchesToMeters(geom.actualIDIn))
	pipeAreaM2 := math.Pi * pipeIDM * pipeIDM / 4.0
	rhoL := float64(model.WaterDensity)
	qRecirc := math.Max(0.0028*qGas, float64(design.PackedBed.RecirculationFlow))

	vFlood := 1.8
	if vf, err := scrubber.EckertFloodingVelocity(qRecirc*rhoL, gMass, rhoG, rhoL, geom.fpSmall, 0.001); err == nil && vf > 0 {
		vFlood = float64(vf)
	}

	// Choose pipe count N so superficial gas velocity stays safely <= 70% of Eckert flooding (strictly <= 75%).
	reqTotalArea := qGas / (0.68 * vFlood)
	pipeCount := int(math.Ceil(reqTotalArea / pipeAreaM2))
	if pipeCount < 2 {
		pipeCount = 2
	}
	for (qGas/(float64(pipeCount)*pipeAreaM2))/vFlood*100.0 > 75.0 {
		pipeCount++
	}

	totalArea := float64(pipeCount) * pipeAreaM2
	vSup := qGas / totalArea
	actualFloodPct := round1((vSup / vFlood) * 100.0)

	// Recompute irrigated Sauter/Ergun bed pressure drop for the smaller packing.
	packHeightM := float64(design.PackedBed.PackingHeight)
	if packHeightM <= 0 {
		packHeightM = 0.65
	}
	packDiamM := float64(model.InchesToMeters(geom.packingDiamIn))
	specArea := 207.0 * (0.0254 / packDiamM)
	sauterDp := 6.0 * (1.0 - geom.epsSmall) / specArea
	bedDPInWC := design.PackedBed.BedPressureDrop.InWC()
	if dryDP, err := scrubber.ErgunPressureDrop(sauterDp, geom.epsSmall, vSup, rhoG, muG, packHeightM); err == nil {
		bedDPInWC = round2(model.Pascal(float64(dryDP) * 1.70).InWC())
	}
	singleDPInWC := round2(design.PackedBed.BedPressureDrop.InWC())

	// Shell-less Uniseal / flat-plate manifold bundle economics (NO external 14" pipe, NO $250 industrial flanges!).
	// Riser pipe length only spans packing bed depth + 1.0 ft freeboard/spray clearance (bottom plate sits over sump plenum).
	singleColPart, singleIDIn, _ := selectSinglePackedColumn(p, design, eductor)
	singlePackPart, singlePackCuFt := selectPackingMedia(singleIDIn, design)
	singleCost := round2(singleColPart.UnitCostUSD + singlePackCuFt*singlePackPart.UnitCostUSD)

	riserHeightFt := math.Max(3.5, model.Meters(packHeightM).Feet()+1.0)
	pipesCost := float64(pipeCount) * riserHeightFt * geom.costPerFt
	if p.CorrosionTier == CorrosionTierHeavyC276FRP {
		pipesCost *= 1.25 // chemical-grade CPVC / FRP-wrapped risers
	}
	const (
		collectorPlatesCost = 95.0 // flat 1/2" HDPE/CPVC top & bottom collector plates (no external drum/shell!)
		unisealUnitCost     = 5.50 // rubber pipe-to-plate Uniseal seals (2 per pipe: top + bottom)
		sprayManifoldCost   = 65.0 // balanced PVC/CPVC calibrated spray distribution header
	)
	unisealsCost := 2.0 * float64(pipeCount) * unisealUnitCost
	manifoldHardwareCost := round2(pipesCost + collectorPlatesCost + unisealsCost + sprayManifoldCost)

	packVolCuFt := math.Max(1.0, math.Ceil(model.CubicMeters(totalArea*packHeightM).CuFt()))
	// Evaluate packing cost for parallel bed using baseline PP small rings so parallel_bed.cost_delta_usd
	// cleanly isolates shell-less manifold savings vs single column (packing_media.cost_delta_usd isolates media swap).
	smallPackCost := packVolCuFt * geom.packUnitCost
	manifoldCost := round2(manifoldHardwareCost + smallPackCost)
	costDelta := round2(manifoldCost - singleCost)

	notes := []string{
		fmt.Sprintf("SHELL-LESS UNISEAL BUNDLE (NO EXTERNAL LARGE PIPE): %d bare %.0f\" %s risers (%.1f ft each) held between top & bottom 1/2\" HDPE/CPVC flat collector plates ($95) with %d Uniseal rubber pipe seals ($5.50 ea = $%.0f) and a $65 spray header — eliminates specialty 14\"–16\" pipe, $250 industrial flanges, and LTL pallet freight.",
			pipeCount, geom.nominalIn, geom.pipeMaterial, riserHeightFt, 2*pipeCount, unisealsCost),
		fmt.Sprintf("Wall-channeling check: %.3f\" pipe ID / %.3f\" packing = D/d_p %.2f (>= 7.5 required to prevent gas bypassing down the pipe wall).", geom.actualIDIn, geom.packingDiamIn, wallRatio),
		fmt.Sprintf("MANDATORY LIQUID DISTRIBUTION HEADER: Install the balanced %d-port calibrated orifice/nozzle spray header ($65) above the top collector plate so every %.0f\" riser receives equal %.2f GPM wetting flow (prevents dry channeling in one pipe while another floods).", pipeCount, geom.nominalIn, (qRecirc/model.CubicMetersPerSecPerGPM)/float64(pipeCount)),
		fmt.Sprintf("Pressure drop trade-off: %.3f\" packing raises irrigated Ergun bed dP from %.2f inWC (single column, 1\" rings) to %.2f inWC.", geom.packingDiamIn, singleDPInWC, bedDPInWC),
	}
	if geom.isPVC {
		notes = append(notes, fmt.Sprintf("CONSERVATIVE THERMAL WARNING (Sch 40 PVC @ $%.2f/ft): Standard PVC softens above 140 F (ASTM D1785). Permitted here ONLY because dry-fan eductor air dilution cools Stage 3 inlet to %.1f F — keep the hardwired safety bypass active and switch to Sch 80 CPVC if ambient studio temps exceed 95 F.", geom.costPerFt, eductor.MixedTempF))
	}
	if geom.nominalIn == 2.0 && pipeCount >= 30 {
		notes = append(notes, fmt.Sprintf("HIGH PIPE-COUNT / LABOR WARNING (2\" pipes): %d parallel 2\" pipes require %d hole-saw cuts and %d Uniseals ($%.0f in seals alone) plus tighter 1/4\" packing (%.2f inWC bed dP). Switch to the 4\" pipe sweet spot (~11–12 pipes) for lower labor and real net cost savings!", pipeCount, 2*pipeCount, 2*pipeCount, unisealsCost, bedDPInWC))
	}

	summary := fmt.Sprintf(
		"Shell-less Uniseal bundle of %d × %.0f\" %s risers (ID %.3f\", NO external large pipe) with %.3g\" packing (D/d_p = %.2f, %.1f%% flood, %.2f inWC bed dP). Total $%.0f vs $%.0f single %.0f\" column (%+.0f USD).",
		pipeCount, geom.nominalIn, geom.pipeMaterial, geom.actualIDIn, geom.packingDiamIn, wallRatio, actualFloodPct, bedDPInWC,
		manifoldCost, singleCost, singleColPart.NominalIn, costDelta,
	)

	return ParallelBedAnalysis{
		Enabled:             enabled,
		ShelllessPlenum:     true,
		PipeMaterial:        geom.pipeMaterial,
		PipeNominalIn:       geom.nominalIn,
		PipeActualIDIn:      geom.actualIDIn,
		PipeCount:           pipeCount,
		PackingDiameterIn:   geom.packingDiamIn,
		WallRatioDToDp:      wallRatio,
		BedDPInWC:           bedDPInWC,
		SingleColumnDPInWC:  singleDPInWC,
		ActualFloodPct:      actualFloodPct,
		SingleColumnCostUSD: singleCost,
		ManifoldCostUSD:     manifoldCost,
		CostDeltaUSD:        costDelta,
		Summary:             summary,
		Notes:               notes,
	}
}

func analyzeParallelVenturi(project StudioProject, design model.ScrubberDesign) ParallelVenturiAnalysis {
	enabled := project.VenturiLayout == VenturiLayoutParallelMiniVenturis
	singleThroatIn := design.Venturi.ThroatDiameter.Inches()
	if singleThroatIn <= 0 {
		singleThroatIn = 2.0
	}

	// Select N in [2, 4] parallel throats matching standard COTS sanitary concentric reducers (2"->1"->2" or 3"->1.5"->3").
	throatCount := 2
	if singleThroatIn/math.Sqrt(4.0) >= 1.05 {
		throatCount = 4
	} else if singleThroatIn/math.Sqrt(3.0) >= 1.00 {
		throatCount = 3
	}
	throatDiamIn := round2(singleThroatIn / math.Sqrt(float64(throatCount)))
	inletFittingIn := 2.0
	if throatDiamIn > 1.20 {
		inletFittingIn = 3.0
	}

	// Recompute venturi dP for smaller individual throat length using Yung-Calvert droplet acceleration + splitter plenum loss.
	vThroat := float64(design.Venturi.ThroatVelocity)
	if vThroat <= 0 {
		vThroat = 75.0
	}
	lg := design.Venturi.LGRatio
	if lg <= 0 {
		lg = 0.001
	}
	dDropM := float64(design.Venturi.DropletDiameter)
	if dDropM <= 0 {
		dDropM = float64(scrubber.NukiyamaTanasawaDropletDiameter(vThroat, lg))
	}
	rhoG := float64(design.QuenchedGas.Density)
	if rhoG <= 0 {
		rhoG = 0.90
	}
	muG := float64(design.QuenchedGas.Viscosity)
	if muG <= 0 {
		muG = 1.85e-5
	}
	rhoL := float64(model.WaterDensity)
	throatLenM := math.Max(0.05, 1.5*float64(model.InchesToMeters(throatDiamIn)))
	beta := scrubber.VenturiDropletAccelFactor(throatLenM, dDropM, vThroat, rhoG, rhoL, muG)
	dpAccel := beta * rhoL * vThroat * vThroat * lg
	dpDiffuser := 0.25 * 0.5 * rhoG * vThroat * vThroat
	dpSplitter := float64(model.InWCToPascal(0.35))
	throatDPInWC := round2(model.Pascal(dpAccel + dpDiffuser + dpSplitter).InWC())
	singleDPInWC := round2(design.Venturi.PressureDropInWC)

	// Plugging risk assessment: flag "High" if throat < 1.0" (25.4 mm) in wood/ash service!
	isWood := design.Input.FuelType == model.FuelWood || design.Input.FuelType == model.FuelWoodSalt
	pluggingRisk := "Low"
	if throatDiamIn < 1.0 && isWood {
		pluggingRisk = "High"
	} else if throatDiamIn < 1.0 || (throatDiamIn < 1.25 && isWood) {
		pluggingRisk = "Moderate"
	}

	// Economics comparison: N off-the-shelf sanitary concentric reducers (zero custom cone rolling) vs single venturi.
	singleCost := estimateSingleVenturiCost(project, design)
	perThroatCost := 64.0 // 2x 316L sanitary concentric reducers + tri-clamp throat tee + EPDM/PTFE gaskets
	headerCost := 52.0
	if inletFittingIn >= 3.0 {
		perThroatCost = 78.0
		headerCost = 68.0
	}
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		perThroatCost = round2(perThroatCost * 2.4)
		headerCost = round2(headerCost * 2.2)
	}
	multiCost := round2(float64(throatCount)*perThroatCost + headerCost)
	costDelta := round2(multiCost - singleCost)

	turndownBenefit := fmt.Sprintf("Cap 1–%d throat(s) with solid tri-clamp blanks during bisque/low-fire to maintain full %.0f m/s particulate capture velocity at %.0f%% turndown.",
		throatCount-1, vThroat, (1.0/float64(throatCount))*100.0)

	notes := []string{
		fmt.Sprintf("Zero custom cone rolling: built from %d × off-the-shelf %.0f\" -> %.2f\" -> %.0f\" sanitary 316L concentric reducers & tri-clamp tees.", throatCount, inletFittingIn, throatDiamIn, inletFittingIn),
		turndownBenefit,
	}
	if pluggingRisk == "High" {
		notes = append(notes, fmt.Sprintf("HIGH WOOD-ASH PLUGGING RISK: %.2f\" (%.1f mm) throat is under 1.0\" (25 mm) in wood-ash service — install a coarse ember/ash screen upstream and inspect tri-clamp throats every firing, or stay with a single larger venturi throat.", throatDiamIn, throatDiamIn*25.4))
	} else if pluggingRisk == "Moderate" {
		notes = append(notes, fmt.Sprintf("Moderate fouling watch: %.2f\" (%.1f mm) throat — quick-release tri-clamp ends allow 2-minute flush-out between firings.", throatDiamIn, throatDiamIn*25.4))
	}

	summary := fmt.Sprintf(
		"%d parallel COTS sanitary mini-venturis (%.0f\" -> %.2f\" throat -> %.0f\", %.1f inWC dP vs %.1f inWC single). Plugging risk: %s. Hardware $%.0f vs $%.0f single (%+.0f USD) with zero sheet-metal cone rolling.",
		throatCount, inletFittingIn, throatDiamIn, inletFittingIn, throatDPInWC, singleDPInWC, pluggingRisk, multiCost, singleCost, costDelta,
	)

	return ParallelVenturiAnalysis{
		Enabled:          enabled,
		ThroatCount:      throatCount,
		ThroatDiameterIn: throatDiamIn,
		InletFittingIn:   inletFittingIn,
		ThroatDPInWC:     throatDPInWC,
		SingleDPInWC:     singleDPInWC,
		PluggingRisk:     pluggingRisk,
		SingleCostUSD:    singleCost,
		MultiCostUSD:     multiCost,
		CostDeltaUSD:     costDelta,
		TurndownBenefit:  turndownBenefit,
		Summary:          summary,
		Notes:            notes,
	}
}

// activeTwinFluidD50Microns computes the Calvert NaCl fume cut diameter d_50 [um]
// under active 40 PSI twin-fluid air-atomizing nozzle injection (counter-current/cross-flow
// relative jet velocity v_rel = max(75 m/s, v_throat + 27 m/s) on 25 um SMD droplets).
func activeTwinFluidD50Microns(atomizeDesign model.ScrubberDesign) float64 {
	muG := float64(atomizeDesign.QuenchedGas.Viscosity)
	if muG <= 0 {
		muG = 1.85e-5
	}
	dDropM := float64(atomizeDesign.Venturi.DropletDiameter)
	if dDropM <= 0 {
		dDropM = float64(model.MicronsToMeters(25.0))
	}
	lg := atomizeDesign.Venturi.LGRatio
	if lg <= 0 {
		lg = 0.0012
	}
	vRel := math.Max(75.0, float64(atomizeDesign.Venturi.ThroatVelocity)+27.0)
	if d50M, err := scrubber.CalvertCutDiameter(muG, dDropM, vRel, float64(model.NaClParticleDensity), lg, float64(model.WaterDensity), scrubber.DefaultCalvertFactor, scrubber.DefaultMeanFreePathM); err == nil && d50M > 0 {
		return round2(model.Meters(d50M).Microns())
	}
	return round2(atomizeDesign.Venturi.CutDiameterD50Microns)
}

func eductorEntrainmentRatio(dpInWC float64) float64 {
	switch {
	case dpInWC <= 4.0:
		return 3.5
	case dpInWC >= 15.0:
		return 1.5
	default:
		// Entrainment collapses smoothly from ~3.5:1 @ 4 inWC down to ~1.5:1 @ 15+ inWC.
		return 3.5 - (dpInWC-4.0)*(2.0/11.0)
	}
}

// analyzeCompressor evaluates the studio's existing shop air compressor in BOTH roles:
// (1) main airflow compressed-air eductor, and (2) twin-fluid 25 um air-atomizing venturi nozzle.
// Always computed on every Evaluate() call.
func analyzeCompressor(project StudioProject, activeDesign, baseDesign model.ScrubberDesign) (CompressorAnalysis, model.ScrubberDesign) {
	p := project.Normalize()
	availSCFM := round1(p.ShopCompressorSCFM)
	if availSCFM <= 0 {
		availSCFM = 15.0
	}

	// Role 1: Main scrubber airflow via air amplifier / compressed-air eductor.
	qACFM := baseDesign.Blower.FlowACFM
	if qACFM <= 0 {
		qACFM = 100.0
	}
	baseDPInWC := baseDesign.Blower.TotalStaticPressureInWC
	if baseDPInWC <= 0 {
		baseDPInWC = 15.0
	}
	entRatio := eductorEntrainmentRatio(baseDPInWC)
	edSCFM := round1(qACFM / entRatio)
	edHP := round1(edSCFM / 3.8)
	edFeasible := edSCFM <= availSCFM

	var edVerdict string
	if edFeasible {
		edVerdict = fmt.Sprintf(
			"FEASIBLE on this small low-dP kiln: moving %.0f ACFM against %.1f inWC (%.1f:1 entrainment) requires %.1f SCFM (~%.1f HP) vs your %.1f SCFM shop compressor — verify 100%% continuous-duty rating and add an inline moisture/oil separator.",
			qACFM, baseDPInWC, entRatio, edSCFM, edHP, availSCFM,
		)
	} else {
		edVerdict = fmt.Sprintf(
			"NOT FEASIBLE for main scrubber airflow: moving %.0f ACFM against %.1f inWC backpressure collapses air-amplifier entrainment to %.1f:1, requiring %.1f SCFM (~%.1f HP industrial rotary-screw compressor) vs your %.1f SCFM studio compressor (%.1fx shortfall). Use your shop compressor for a 25 um twin-fluid air-atomizing nozzle instead!",
			qACFM, baseDPInWC, entRatio, edSCFM, edHP, availSCFM, edSCFM/math.Max(availSCFM, 1.0),
		)
	}

	// Role 2: Twin-fluid air-atomizing venturi nozzle (4-6 SCFM @ 40 PSI -> 25 um SMD droplets, 48 m/s throat).
	var atomizeDesign model.ScrubberDesign
	if p.BlowerDriveOption == BlowerDriveCompressorAirAtomize {
		atomizeDesign = activeDesign
	} else {
		atomProj := p
		atomProj.BlowerDriveOption = BlowerDriveCompressorAirAtomize
		if d, err := scrubber.NewDesigner().Design(atomProj.EffectiveDesignInput()); err == nil {
			atomizeDesign = d
		} else {
			atomizeDesign = baseDesign
		}
	}

	ventGPM := atomizeDesign.Venturi.LiquidFlowGPM
	atomizeSCFM := round1(math.Max(4.0, math.Min(6.0, 4.0+1.5*ventGPM)))
	atomizeFeasible := availSCFM >= atomizeSCFM
	atomizeDropletUm := round1(atomizeDesign.Venturi.DropletDiameter.Microns())
	if atomizeDropletUm <= 0 {
		atomizeDropletUm = 25.0
	}
	dpSaved := round2(math.Max(0.0, baseDesign.Blower.TotalStaticPressureInWC-atomizeDesign.Blower.TotalStaticPressureInWC))

	baseFan := selectWetBlowerVFD(baselineProjectFor(p), baseDesign)
	atomFanCost, _ := mediumPressureAirAtomizeBlowerCost(p, atomizeDesign.Blower.RecommendedMotorHP)
	costSaved := round2(math.Max(0.0, baseFan.UnitCostUSD-atomFanCost))

	activeD50 := activeTwinFluidD50Microns(atomizeDesign)
	var atomVerdict string
	if atomizeFeasible {
		atomVerdict = fmt.Sprintf(
			"FEASIBLE & HIGH SAVINGS: A $75 twin-fluid air-atomizing nozzle needs only %.1f SCFM @ 40 PSI (%.0f%% of your %.1f SCFM compressor), makes %.0f um SMD droplets at %.0f m/s throat velocity (active twin-fluid jet shear v_rel~75 m/s achieves Calvert d50 = %.2f um parity vs %.2f um baseline; passive u_L0=0 would slip to %.2f um, so aim nozzle cross-flow/counter-current!), cuts venturi dP %.1f -> %.1f inWC (total static %.1f -> %.1f inWC, saving %.1f inWC), and unlocks a $360 medium-pressure cast-aluminum blower saving $%.0f!",
			atomizeSCFM, (atomizeSCFM/availSCFM)*100.0, availSCFM, atomizeDropletUm,
			float64(atomizeDesign.Venturi.ThroatVelocity), activeD50, baseDesign.Venturi.CutDiameterD50Microns, atomizeDesign.Venturi.CutDiameterD50Microns,
			baseDesign.Venturi.PressureDropInWC, atomizeDesign.Venturi.PressureDropInWC,
			baseDesign.Blower.TotalStaticPressureInWC, atomizeDesign.Blower.TotalStaticPressureInWC,
			dpSaved, costSaved,
		)
	} else {
		atomVerdict = fmt.Sprintf(
			"Requires %.1f SCFM @ 40 PSI for 25 um air-atomizing (slightly above your %.1f SCFM setting) — upgrade to a standard 5 HP 60-gal studio compressor (>= 10 SCFM) to unlock %.1f inWC dP reduction and $%.0f blower savings.",
			atomizeSCFM, availSCFM, dpSaved, costSaved,
		)
	}

	return CompressorAnalysis{
		AvailableSCFM:       availSCFM,
		EductorSCFMNeeded:   edSCFM,
		EductorHPNeeded:     edHP,
		EductorFeasible:     edFeasible,
		EductorVerdict:      edVerdict,
		AtomizeSCFMNeeded:   atomizeSCFM,
		AtomizeFeasible:     atomizeFeasible,
		AtomizeDropletUm:    atomizeDropletUm,
		AtomizeDPSavedInWC:  dpSaved,
		AtomizeCostSavedUSD: costSaved,
		AtomizeVerdict:      atomVerdict,
	}, atomizeDesign
}

func voltageForHP(project StudioProject, hp float64) int {
	if hp <= 0.5 || (hp <= 1.0 && project.AvailableVoltage < 230) {
		return 120
	}
	return 240
}

func analyzeBlowerDrive(
	project StudioProject,
	activeDesign model.ScrubberDesign,
	baseDesign model.ScrubberDesign,
	eductor EductorAnalysis,
	comp CompressorAnalysis,
	atomizeDesign model.ScrubberDesign,
) BlowerDriveAnalysis {
	p := project.Normalize()
	baseFan := selectWetBlowerVFD(baselineProjectFor(p), baseDesign)
	baseFanCost := baseFan.UnitCostUSD

	switch p.BlowerDriveOption {
	case BlowerDriveRegenSideChannel:
		qCFM := activeDesign.Blower.FlowACFM
		reqStatic := round1(activeDesign.Blower.TotalStaticPressureInWC)
		if eductor.Enabled {
			qCFM = eductor.MotiveAirCFM
			reqStatic = eductor.DryFanStaticInWC
		}
		var hp, cost float64
		switch {
		case qCFM <= 120.0:
			hp, cost = 1.5, 380.0
		case qCFM <= 280.0:
			hp, cost = 3.0, 460.0
		default:
			hp, cost = 5.0, 520.0
		}
		feasible := qCFM <= 340.0 && reqStatic <= 65.0
		voltage := 240
		if hp <= 1.5 && p.AvailableVoltage == 120 {
			voltage = 120
		}
		delta := round2(cost - baseFanCost)
		var caveats []string
		if !eductor.Enabled {
			caveats = append(caveats, "DRY-AIR CORROSION CAVEAT: Cast-aluminum side-channel regenerative blowers have tight (~0.005\") impeller clearances intended for dry clean air. Using one wet on acidic HCl exhaust without the Dry-Fan Eductor risks rapid aluminum pitting and salt-crystal seizure — pair with Stage Alignment = eductor_dry_fan or flush/purge dry after every firing.")
		} else {
			caveats = append(caveats, "Dry-air side placement (paired with Vertical Downflow Eductor): 100% dry ambient air avoids wet HCl contact on the cast-aluminum impeller while leveraging 60+ inWC static headroom.")
		}
		caveats = append(caveats, "Use a bleed/relief blast-gate damper (never dead-head) and coarse intake filter; ring blowers run at fixed speed unless paired with a VFD.")
		if !feasible {
			caveats = append(caveats, fmt.Sprintf("Required airflow (%.0f CFM) exceeds single 3 HP side-channel ring blower capacity (~280–340 CFM); use dual parallel ring blowers or an industrial radial blower.", qCFM))
		}
		summary := fmt.Sprintf(
			"COTS %.2g HP cast-aluminum side-channel regenerative ring blower ($%.0f vs $%.0f baseline -> %+.0f USD) delivers 60+ inWC static headroom for %.0f CFM @ %.1f inWC.",
			hp, cost, baseFanCost, delta, qCFM, reqStatic,
		)
		return BlowerDriveAnalysis{
			Option:             BlowerDriveRegenSideChannel,
			Label:              "Side-Channel Regenerative Ring Blower (2–3 HP)",
			FanCostUSD:         cost,
			BaselineFanCostUSD: baseFanCost,
			CostDeltaUSD:       delta,
			RequiredStaticInWC: reqStatic,
			ThroatVelocityMS:   round1(float64(activeDesign.Venturi.ThroatVelocity)),
			DropletUm:          round1(activeDesign.Venturi.DropletDiameter.Microns()),
			MotorHP:            hp,
			Voltage:            voltage,
			Feasible:           feasible,
			Summary:            summary,
			Caveats:            caveats,
		}

	case BlowerDriveDualBypassBlowers:
		qCFM := activeDesign.Blower.FlowACFM
		reqStatic := round1(activeDesign.Blower.TotalStaticPressureInWC)
		var cost, hp float64
		switch {
		case qCFM <= 130.0:
			cost, hp = 380.0, 1.5
		case qCFM <= 280.0:
			cost, hp = 420.0, 2.0
		default:
			cost, hp = 460.0, 2.5
		}
		feasible := qCFM <= 285.0 && reqStatic <= 35.0
		voltage := 120
		if p.AvailableVoltage >= 230 {
			voltage = 240
		}
		delta := round2(cost - baseFanCost)
		caveats := []string{
			"Tangential bypass discharge keeps cooling air over the motor windings completely isolated from wet acidic exhaust (never substitute thru-flow shop-vac motors!).",
			"CARBON BRUSH WEAR ITEM: Universal-motor carbon brushes last ~650–800 hrs (~45 sixteen-hour firings); budget $15/pair for a quick brush swap every ~45 firings.",
			"High-RPM (~18,000 RPM) tangential fans emit higher-pitched whine (~82–86 dBA bare) — pair with the insulated acoustic blower box/silencer for residential yards.",
		}
		if !feasible {
			caveats = append(caveats, fmt.Sprintf("Kiln exhaust flow (%.0f ACFM) exceeds 2x Ametek Lamb 3-stage parallel capacity (~240–260 CFM @ 25+ inWC); use 4x staged units or an industrial radial blower.", qCFM))
		}
		summary := fmt.Sprintf(
			"2× COTS Ametek Lamb 3-stage tangential bypass blowers in parallel (~240–260 CFM @ 25+ inWC, $%.0f total vs $%.0f baseline -> %+.0f USD). Tangential discharge isolates motor windings from wet exhaust; swap $15 brushes every ~45 firings.",
			cost, baseFanCost, delta,
		)
		return BlowerDriveAnalysis{
			Option:             BlowerDriveDualBypassBlowers,
			Label:              "Dual 3-Stage Tangential Bypass Blowers (2× Ametek Lamb)",
			FanCostUSD:         cost,
			BaselineFanCostUSD: baseFanCost,
			CostDeltaUSD:       delta,
			RequiredStaticInWC: reqStatic,
			ThroatVelocityMS:   round1(float64(activeDesign.Venturi.ThroatVelocity)),
			DropletUm:          round1(activeDesign.Venturi.DropletDiameter.Microns()),
			MotorHP:            hp,
			Voltage:            voltage,
			Feasible:           feasible,
			Summary:            summary,
			Caveats:            caveats,
		}

	case BlowerDriveCompressorAirAtomize:
		hp := atomizeDesign.Blower.RecommendedMotorHP
		if hp <= 0 {
			hp = 1.0
		}
		cost, _ := mediumPressureAirAtomizeBlowerCost(p, hp)
		delta := round2(cost - baseFanCost)
		reqStatic := round1(atomizeDesign.Blower.TotalStaticPressureInWC)
		vThroat := round1(float64(atomizeDesign.Venturi.ThroatVelocity))
		dropletUm := round1(atomizeDesign.Venturi.DropletDiameter.Microns())
		voltage := voltageForHP(p, hp)
		activeD50 := activeTwinFluidD50Microns(atomizeDesign)
		caveats := []string{
			fmt.Sprintf("Requires %.1f SCFM @ 40 PSI continuous from your existing studio shop compressor (%.0f%% of %.1f SCFM capacity) through an inline air-pressure regulator & moisture trap.", comp.AtomizeSCFMNeeded, (comp.AtomizeSCFMNeeded/math.Max(comp.AvailableSCFM, 1.0))*100.0, comp.AvailableSCFM),
			fmt.Sprintf("Calvert physics parity verified: 25 um twin-fluid jet shear (v_rel ~75 m/s @ 40 PSI in a %.0f m/s throat) achieves d50 = %.2f um (beating %.2f um @ 70 m/s baseline; passive u_L0=0 would slip to %.2f um) while cutting venturi dP %.1f -> %.1f inWC and total static %.1f -> %.1f inWC.", vThroat, activeD50, baseDesign.Venturi.CutDiameterD50Microns, atomizeDesign.Venturi.CutDiameterD50Microns, baseDesign.Venturi.PressureDropInWC, atomizeDesign.Venturi.PressureDropInWC, baseDesign.Blower.TotalStaticPressureInWC, atomizeDesign.Blower.TotalStaticPressureInWC),
			"If compressed air fails mid-firing, venturi particulate efficiency temporarily drops until air returns (Stage 1 quench cooling and Stage 3 HCl absorption continue normally).",
		}
		summary := fmt.Sprintf(
			"$75 twin-fluid air-atomizing nozzle (%.1f SCFM @ 40 PSI -> 25 um SMD) slashes total static %.1f -> %.1f inWC at 48 m/s (active twin-fluid d50=%.2f um parity), unlocking a $360 medium-pressure cast-aluminum radial blower + simple VFD ($%.0f total vs $%.0f baseline -> %+.0f USD).",
			comp.AtomizeSCFMNeeded, baseDesign.Blower.TotalStaticPressureInWC, reqStatic, activeD50, cost, baseFanCost, delta,
		)
		return BlowerDriveAnalysis{
			Option:             BlowerDriveCompressorAirAtomize,
			Label:              "Shop Compressor Air-Atomize (25 µm) + Medium-Pressure Radial Blower",
			FanCostUSD:         cost,
			BaselineFanCostUSD: baseFanCost,
			CostDeltaUSD:       delta,
			RequiredStaticInWC: reqStatic,
			ThroatVelocityMS:   vThroat,
			DropletUm:          dropletUm,
			MotorHP:            hp,
			Voltage:            voltage,
			Feasible:           comp.AtomizeFeasible,
			Summary:            summary,
			Caveats:            caveats,
		}

	case BlowerDriveCompressorEductor:
		const eductorNozzleCost = 185.0 // COTS 316L air amplifier / venturi eductor hardware only (assumes existing compressor)
		delta := round2(eductorNozzleCost - baseFanCost)
		voltage := 240
		if comp.EductorHPNeeded <= 2.0 {
			voltage = 120
		}
		caveats := []string{
			comp.EductorVerdict,
			"Compressed-air amplification has only ~15–20% overall wire-to-air thermodynamic efficiency (vs ~62% for a direct radial blower), consuming 3–4× more electricity per firing.",
			"Only viable for tiny <= 6 cu.ft. low-dP test kilns; for 14–48 cu.ft. studio kilns, use your shop compressor for a 25 um twin-fluid air-atomizing nozzle instead (4–6 SCFM).",
		}
		summary := fmt.Sprintf(
			"Shop compressor main-airflow eductor needs %.1f SCFM (~%.1f HP) vs %.1f SCFM available (feasible=%v). Hardware nozzle is $%.0f, but backpressure entrainment collapse makes main airflow impractical above tiny <= 6 cu.ft. low-dP test kilns.",
			comp.EductorSCFMNeeded, comp.EductorHPNeeded, comp.AvailableSCFM, comp.EductorFeasible, eductorNozzleCost,
		)
		return BlowerDriveAnalysis{
			Option:             BlowerDriveCompressorEductor,
			Label:              "Shop Compressor Main Eductor / Air Amplifier (Feasibility Check)",
			FanCostUSD:         eductorNozzleCost,
			BaselineFanCostUSD: baseFanCost,
			CostDeltaUSD:       delta,
			RequiredStaticInWC: round1(baseDesign.Blower.TotalStaticPressureInWC),
			ThroatVelocityMS:   round1(float64(baseDesign.Venturi.ThroatVelocity)),
			DropletUm:          round1(baseDesign.Venturi.DropletDiameter.Microns()),
			MotorHP:            comp.EductorHPNeeded,
			Voltage:            voltage,
			Feasible:           comp.EductorFeasible,
			Summary:            summary,
			Caveats:            caveats,
		}

	default: // BlowerDriveIndustrialRadialVFD
		if eductor.Enabled {
			dryPart := selectDryMotiveBlowerVFD(p, eductor)
			delta := round2(dryPart.UnitCostUSD - baseFanCost)
			return BlowerDriveAnalysis{
				Option:             BlowerDriveIndustrialRadialVFD,
				Label:              "Clean Dry-Air Industrial Radial Blower + VFD (Eductor Motive)",
				FanCostUSD:         dryPart.UnitCostUSD,
				BaselineFanCostUSD: baseFanCost,
				CostDeltaUSD:       delta,
				RequiredStaticInWC: eductor.DryFanStaticInWC,
				ThroatVelocityMS:   round1(float64(activeDesign.Venturi.ThroatVelocity)),
				DropletUm:          round1(activeDesign.Venturi.DropletDiameter.Microns()),
				MotorHP:            eductor.DryFanMotorHP,
				Voltage:            voltageForHP(p, eductor.DryFanMotorHP),
				Feasible:           eductor.Feasible,
				Summary:            fmt.Sprintf("Clean dry-air %.2g HP radial blower + GS21 VFD in NEMA 4X ($%.0f vs $%.0f wet baseline -> %+.0f USD) driving vertical downflow eductor at %.1f inWC.", eductor.DryFanMotorHP, dryPart.UnitCostUSD, baseFanCost, delta, eductor.DryFanStaticInWC),
				Caveats:            []string{"Mounted strictly on dry ambient intake side (zero acid contact); Stage 3 downstream operates under positive pressure (outdoor-only)."},
			}
		}
		hp := activeDesign.Blower.RecommendedMotorHP
		if hp <= 0 {
			hp = 1.0
		}
		return BlowerDriveAnalysis{
			Option:             BlowerDriveIndustrialRadialVFD,
			Label:              "Industrial Radial Pressure Blower + VFD (Baseline)",
			FanCostUSD:         baseFanCost,
			BaselineFanCostUSD: baseFanCost,
			CostDeltaUSD:       0.0,
			RequiredStaticInWC: round1(activeDesign.Blower.TotalStaticPressureInWC),
			ThroatVelocityMS:   round1(float64(activeDesign.Venturi.ThroatVelocity)),
			DropletUm:          round1(activeDesign.Venturi.DropletDiameter.Microns()),
			MotorHP:            hp,
			Voltage:            voltageForHP(p, hp),
			Feasible:           true,
			Summary:            fmt.Sprintf("Baseline %.2g HP corrosion-duty radial pressure blower + GS21 VFD in NEMA 4X ($%.0f) delivering %.0f ACFM @ %.1f inWC with smooth PID kiln draft control.", hp, baseFanCost, activeDesign.Blower.FlowACFM, activeDesign.Blower.TotalStaticPressureInWC),
			Caveats:            []string{"Highest upfront fan cost; select Compressor Air-Atomize, Dual 3-Stage Bypass, or Side-Channel Ring Blower to save $1,000–$1,400."},
		}
	}
}

type packingMediaSpec struct {
	option       string
	label        string
	sku          string
	diameterIn   float64
	voidage      float64
	specAreaRef  float64 // specific surface area a_p at 1" reference [m^2/m^3]
	unitCostUSD  float64
	weightKgCuFt float64
	material     string
	source       string
}

func resolvePackingMediaSpec(option string, parallelBed ParallelBedAnalysis) (sel packingMediaSpec, base packingMediaSpec) {
	if parallelBed.Enabled {
		baseUnitCost := 98.0
		baseVoidage := 0.87
		baseSKU := "PACK-PP-PALL-050IN"
		if parallelBed.PackingDiameterIn < 0.30 {
			baseUnitCost = 112.0
			baseVoidage = 0.82
			baseSKU = "PACK-PP-SAD-025IN"
		} else if parallelBed.PackingDiameterIn < 0.45 {
			baseUnitCost = 104.0
			baseVoidage = 0.85
			baseSKU = "PACK-PP-SAD-0375IN"
		}
		base = packingMediaSpec{
			option:       PackingMediaPPPallRings,
			label:        fmt.Sprintf(`Commercial %.3g" PP Pall Rings / Saddles`, parallelBed.PackingDiameterIn),
			sku:          baseSKU,
			diameterIn:   parallelBed.PackingDiameterIn,
			voidage:      baseVoidage,
			specAreaRef:  207.0,
			unitCostUSD:  baseUnitCost,
			weightKgCuFt: 2.8,
			material:     fmt.Sprintf("Polypropylene (PP), %.0f%% Voidage, D/d_p=%.2f", baseVoidage*100.0, parallelBed.WallRatioDToDp),
			source:       "McMaster-Carr / Jaeger / US Plastic",
		}
		switch option {
		case PackingMediaAquacultureBioballs:
			sel = packingMediaSpec{
				option:       PackingMediaAquacultureBioballs,
				label:        fmt.Sprintf(`Aquaculture PP Bio-Media (%.3g" Mini Bio-Balls)`, parallelBed.PackingDiameterIn),
				sku:          "PACK-BIOBALL-MINI",
				diameterIn:   parallelBed.PackingDiameterIn,
				voidage:      0.85,
				specAreaRef:  185.0,
				unitCostUSD:  38.0,
				weightKgCuFt: 2.9,
				material:     fmt.Sprintf("Bulk Aquaculture Polypropylene Bio-Media, 85%% Voidage, D/d_p=%.2f", parallelBed.WallRatioDToDp),
				source:       "Pentair AES / Bulk Aquaculture Supply",
			}
		case PackingMediaDIYCeramicRings:
			sel = packingMediaSpec{
				option:       PackingMediaDIYCeramicRings,
				label:        fmt.Sprintf(`DIY Kiln-Fired Stoneware Raschig Rings (%.3g")`, parallelBed.PackingDiameterIn),
				sku:          "PACK-DIY-CERAMIC-RING",
				diameterIn:   parallelBed.PackingDiameterIn,
				voidage:      0.70,
				specAreaRef:  240.0,
				unitCostUSD:  18.0,
				weightKgCuFt: 21.0, // heavy unglazed stoneware rings (~46 lb/cu.ft.)
				material:     fmt.Sprintf("Unglazed Stoneware Ceramic Raschig Rings (70%% Voidage, 2000 F & Acid Proof, D/d_p=%.2f)", parallelBed.WallRatioDToDp),
				source:       "Studio Clay Extruder + Kiln Bisque/Stoneware Firing",
			}
		default:
			sel = base
		}
		return sel, base
	}

	base = packingMediaSpec{
		option:       PackingMediaPPPallRings,
		label:        `Commercial 1" PP Pall Rings`,
		sku:          "PACK-PP-PALL-1IN",
		diameterIn:   1.0,
		voidage:      0.90,
		specAreaRef:  207.0,
		unitCostUSD:  88.0, // small-lot US distributor 1-10 cu.ft. pricing
		weightKgCuFt: 2.5,
		material:     "Polypropylene (PP), 90% Voidage, Fp=52/ft",
		source:       "McMaster-Carr / Jaeger / US Plastic",
	}
	switch option {
	case PackingMediaAquacultureBioballs:
		sel = packingMediaSpec{
			option:       PackingMediaAquacultureBioballs,
			label:        `Aquaculture PP Bio-Media (1"–1.25" Bio-Balls)`,
			sku:          "PACK-BIOBALL-1IN",
			diameterIn:   1.125,
			voidage:      0.85,
			specAreaRef:  185.0,
			unitCostUSD:  38.0,
			weightKgCuFt: 2.9,
			material:     "Bulk Aquaculture Polypropylene Bio-Balls (85% Voidage, <=200 F)",
			source:       "Pentair AES / Bulk Pond & Aquaculture Supply",
		}
	case PackingMediaDIYCeramicRings:
		sel = packingMediaSpec{
			option:       PackingMediaDIYCeramicRings,
			label:        `DIY Kiln-Fired Stoneware Ceramic Raschig Rings (3/4")`,
			sku:          "PACK-DIY-CERAMIC-075",
			diameterIn:   0.75,
			voidage:      0.70,
			specAreaRef:  240.0,
			unitCostUSD:  18.0,
			weightKgCuFt: 21.0, // heavy unglazed stoneware (~46 lb/cu.ft.)
			material:     "Potter-Extruded Unglazed Stoneware Ceramic Rings (70% Voidage, Acid-Proof & 2000 F-Proof)",
			source:       "Studio Clay Extruder + Potter's Kiln ($18/cu.ft. clay + firing)",
		}
	default:
		sel = base
	}
	return sel, base
}

func analyzePackingMedia(
	project StudioProject,
	design model.ScrubberDesign,
	eductor EductorAnalysis,
	parallelBed ParallelBedAnalysis,
) (PackingMediaAnalysis, packingMediaSpec, float64) {
	p := project.Normalize()
	sel, base := resolvePackingMediaSpec(p.PackingMediaOption, parallelBed)

	// Determine active gas flow, cross-sectional area, and packing volume.
	qGas := float64(design.QuenchedGas.VolumetricFlow)
	rhoG := float64(design.QuenchedGas.Density)
	muG := float64(design.QuenchedGas.Viscosity)
	if qGas <= 0 {
		qGas = 0.10
	}
	if rhoG <= 0 {
		rhoG = 0.90
	}
	if muG <= 0 {
		muG = 1.85e-5
	}
	if eductor.Enabled {
		pAmb := float64(design.Input.AmbientPressurePa)
		if pAmb <= 0 {
			pAmb = float64(model.StandardPressurePa)
		}
		qMotive := float64(model.CFMToCubicMetersPerSec(eductor.MotiveAirCFM))
		_, qMix, rhoMix, _ := solveMixedAirState(design.QuenchedGas, qMotive, pAmb)
		qGas, rhoG = qMix, rhoMix
	}

	packHeightM := float64(design.PackedBed.PackingHeight)
	if packHeightM <= 0 {
		packHeightM = 0.65
	}
	var totalAreaM2 float64
	if parallelBed.Enabled {
		pipeIDM := float64(model.InchesToMeters(parallelBed.PipeActualIDIn))
		totalAreaM2 = float64(parallelBed.PipeCount) * math.Pi * pipeIDM * pipeIDM / 4.0
	} else {
		_, actualIDIn, _ := selectSinglePackedColumn(p, design, eductor)
		colDiaM := float64(model.InchesToMeters(actualIDIn))
		totalAreaM2 = math.Pi * colDiaM * colDiaM / 4.0
	}
	vSup := qGas / math.Max(totalAreaM2, 1e-6)
	packCuFt := math.Max(1.0, math.Ceil(model.CubicMeters(totalAreaM2*packHeightM).CuFt()))

	// Exact irrigated Sauter/Ergun bed pressure drop for selected packing media.
	dM := float64(model.InchesToMeters(sel.diameterIn))
	specArea := sel.specAreaRef * (0.0254 / dM)
	sauterDp := 6.0 * (1.0 - sel.voidage) / specArea
	bedDPInWC := round2(design.PackedBed.BedPressureDrop.InWC())
	if dryDP, err := scrubber.ErgunPressureDrop(sauterDp, sel.voidage, vSup, rhoG, muG, packHeightM); err == nil {
		bedDPInWC = round2(model.Pascal(float64(dryDP) * 1.70).InWC())
	}

	mediaCost := round2(packCuFt * sel.unitCostUSD)
	baseCost := round2(packCuFt * base.unitCostUSD)
	costDelta := round2(mediaCost - baseCost)
	totalWeightLbs := round1(packCuFt * sel.weightKgCuFt / model.KgPerLb)

	var summary string
	switch sel.option {
	case PackingMediaAquacultureBioballs:
		summary = fmt.Sprintf(
			"%.0f cu.ft. bulk aquaculture PP bio-media ($38/cu.ft., %.0f%% voidage, %.2f inWC irrigated Ergun bed dP, ~%.0f lbs): $%.0f vs $%.0f commercial PP Pall rings (%+.0f USD savings).",
			packCuFt, sel.voidage*100.0, bedDPInWC, totalWeightLbs, mediaCost, baseCost, costDelta,
		)
	case PackingMediaDIYCeramicRings:
		summary = fmt.Sprintf(
			"%.0f cu.ft. potter-extruded unglazed stoneware Raschig rings (%.3g\", $18/cu.ft. clay+firing, 70%% voidage): 100%% acid-proof & 2000 F-proof ($%.0f vs $%.0f -> %+.0f USD). Trade-off: lower 70%% voidage raises irrigated Ergun bed dP to %.2f inWC and adds ~%.0f lbs bed weight (use reinforced FRP/316L support grate!).",
			packCuFt, sel.diameterIn, mediaCost, baseCost, costDelta, bedDPInWC, totalWeightLbs,
		)
	default:
		summary = fmt.Sprintf(
			"%.0f cu.ft. %s ($%.0f/cu.ft., %.0f%% voidage, %.2f inWC irrigated Ergun bed dP, ~%.0f lbs): lowest bed pressure drop and highest HCl transfer efficiency.",
			packCuFt, sel.label, sel.unitCostUSD, sel.voidage*100.0, bedDPInWC, totalWeightLbs,
		)
	}

	return PackingMediaAnalysis{
		Option:               sel.option,
		Label:                sel.label,
		DiameterIn:           sel.diameterIn,
		Voidage:              sel.voidage,
		BedDPInWC:            bedDPInWC,
		MediaCostUSD:         mediaCost,
		BaselineMediaCostUSD: baseCost,
		CostDeltaUSD:         costDelta,
		Summary:              summary,
	}, sel, packCuFt
}

// EvaluateArchitectureOptions computes the complete Phase 7 & Phase 8 trade-off summary,
// exact baseline-vs-selected COTS cost delta, footprint savings, blower drive, packing media,
// shop compressor feasibility, and warnings.
func EvaluateArchitectureOptions(project StudioProject, design model.ScrubberDesign, selectedCOTS COTSSummary) ArchitectureOptionsResult {
	p := project.Normalize()
	baseProj := baselineProjectFor(p)
	baseDesign := resolveBaselineDesign(p, design)

	baseCOTS := BuildCOTSSummary(baseProj, baseDesign)
	eductor := analyzeEductor(p, design)
	parallelBed := analyzeParallelBed(p, design, eductor)
	parallelVenturi := analyzeParallelVenturi(p, design)
	compressor, atomizeDesign := analyzeCompressor(p, design, baseDesign)
	blowerDrive := analyzeBlowerDrive(p, design, baseDesign, eductor, compressor, atomizeDesign)
	packingMedia, _, _ := analyzePackingMedia(p, design, eductor, parallelBed)

	// Individual control-tier and shared-sump deltas for transparent breakdown.
	controlDelta := 0.0
	if p.ControlTier == ControlTierElectromechanicalOnly {
		rpiKit := selectSensorKitForTier(ControlTierRPiModbus, design)
		emKit := selectSensorKitForTier(ControlTierElectromechanicalOnly, design)
		controlDelta = round2(emKit.UnitCostUSD - rpiKit.UnitCostUSD)
	}
	sumpDelta := 0.0
	if !p.SharedBaffledSump {
		sumpDelta = round2(dualSumpExtraCost(design))
	}

	// Footprint geometry (Linear vs U-Turn Folded Skid).
	baseLenFt := round1(math.Max(5.5, design.Envelope.Length.Feet()))
	baseWidthFt := round1(math.Max(2.2, design.Envelope.Width.Feet()))
	if eductor.Enabled {
		// Eliminating downstream horizontal blower scroll saves ~1.2 ft length.
		baseLenFt = round1(math.Max(4.8, baseLenFt-1.2))
	}
	fpLenFt := baseLenFt
	fpWidthFt := baseWidthFt
	fpSavingsPct := 0.0
	if p.CompactFootprint {
		fpLenFt = round1(baseLenFt * 0.53) // ~47% linear length reduction for narrow urban side-yards
		fpWidthFt = round1(math.Max(baseWidthFt*1.45, baseWidthFt+1.3))
		fpSavingsPct = round1((1.0 - fpLenFt/baseLenFt) * 100.0)
	}
	fpAreaSqFt := round1(fpLenFt * fpWidthFt)

	costDelta := round2(selectedCOTS.EstimatedTotalUSD - baseCOTS.EstimatedTotalUSD)

	var savings []string
	var warnings []string

	if blowerDrive.Option != BlowerDriveIndustrialRadialVFD {
		savings = append(savings, fmt.Sprintf("%s: %+.0f USD vs industrial radial+VFD ($%.0f vs $%.0f; %.1f inWC static, feasible=%v).",
			blowerDrive.Label, blowerDrive.CostDeltaUSD, blowerDrive.FanCostUSD, blowerDrive.BaselineFanCostUSD, blowerDrive.RequiredStaticInWC, blowerDrive.Feasible))
		if !blowerDrive.Feasible {
			warnings = append(warnings, fmt.Sprintf("Blower Drive Feasibility Warning (%s): %s", blowerDrive.Label, blowerDrive.Summary))
		}
		if blowerDrive.Option == BlowerDriveRegenSideChannel && !eductor.Enabled {
			warnings = append(warnings, blowerDrive.Caveats[0])
		}
	}
	if eductor.Enabled {
		savings = append(savings, fmt.Sprintf("Dry-Fan Downflow Eductor: %+.0f USD blower delta (eliminates wet acid-duty blower; +%.1f F CPVC thermal margin; Stage 3 enlarged to %.0f\" @ %.1f%% flood).",
			eductor.NetBlowerDeltaUSD, eductor.ThermalMarginGainF, eductor.EnlargedBedDiameterIn, eductor.EnlargedBedFloodPct))
		warnings = append(warnings, eductor.Warnings...)
	}
	if parallelBed.Enabled {
		savings = append(savings, fmt.Sprintf("Shell-Less Uniseal %.0f\" Parallel Bundle (%d risers, %s): %+.0f USD vs single column; D/d_p = %.2f, %.2f inWC bed dP.",
			parallelBed.PipeNominalIn, parallelBed.PipeCount, parallelBed.PipeMaterial, parallelBed.CostDeltaUSD, parallelBed.WallRatioDToDp, parallelBed.BedDPInWC))
		warnings = append(warnings, fmt.Sprintf("Parallel Bed Distribution: Use calibrated %d-port top liquid distributor so no riser runs dry while another floods.", parallelBed.PipeCount))
		if parallelBed.PipeNominalIn == 2.0 && parallelBed.PipeCount >= 30 {
			warnings = append(warnings, fmt.Sprintf("Parallel 2\" Pipe Count Warning: %d pipes require %d Uniseals — consider 4\" pipes for lower labor and maximum savings.", parallelBed.PipeCount, 2*parallelBed.PipeCount))
		}
	}
	if packingMedia.Option != PackingMediaPPPallRings {
		savings = append(savings, fmt.Sprintf("%s: %+.0f USD vs commercial PP Pall rings ($%.0f vs $%.0f; voidage %.0f%%, bed dP %.2f inWC).",
			packingMedia.Label, packingMedia.CostDeltaUSD, packingMedia.MediaCostUSD, packingMedia.BaselineMediaCostUSD, packingMedia.Voidage*100.0, packingMedia.BedDPInWC))
		if packingMedia.Option == PackingMediaDIYCeramicRings {
			warnings = append(warnings, fmt.Sprintf("DIY Ceramic Rings Trade-Off: 70%% voidage raises irrigated bed dP to %.2f inWC and adds heavy stoneware weight (~46 lb/cu.ft.) — use a reinforced FRP/316L support grate.", packingMedia.BedDPInWC))
		}
	}
	if parallelVenturi.Enabled && !eductor.Enabled {
		savings = append(savings, fmt.Sprintf("Parallel COTS Mini-Venturis (%d × %.2f\" throats): %+.0f USD, zero cone rolling, staged low-fire turndown.",
			parallelVenturi.ThroatCount, parallelVenturi.ThroatDiameterIn, parallelVenturi.CostDeltaUSD))
		if parallelVenturi.PluggingRisk == "High" {
			warnings = append(warnings, fmt.Sprintf("Parallel Mini-Venturi Ash Plugging Risk: HIGH (%.2f\" throat < 1.0\" in wood-ash service).", parallelVenturi.ThroatDiameterIn))
		}
	}
	if p.CompactFootprint {
		savings = append(savings, fmt.Sprintf("Compact U-Turn Folded Skid: cuts linear length %.1f%% (%.1f ft -> %.1f ft × %.1f ft W = %.1f sq.ft.) via 180-deg return bend (+$65).",
			fpSavingsPct, baseLenFt, fpLenFt, fpWidthFt, fpAreaSqFt))
	}
	if p.SharedBaffledSump {
		savings = append(savings, "Shared Baffled Sump + Single Mag-Drive Pump (active): avoids ~$365+ second tank & pump cost.")
	} else {
		savings = append(savings, fmt.Sprintf("Dual Dedicated Sumps (separate quench-ash + absorber sumps): +$%.0f for extra drum, pump & piping.", sumpDelta))
	}
	if p.ControlTier == ControlTierElectromechanicalOnly {
		savings = append(savings, fmt.Sprintf("Electromechanical Safety Only: saves $%.0f vs Raspberry Pi Modbus (retains mandatory NFPA 86 hardwired 200 F snap-disc, 0.25 GPM flow switch, latching safety relay & spring-return bypass damper + analog Magnehelic & manual VFD dial).",
			math.Abs(controlDelta)))
	}
	if p.IsDeluxe() {
		savings = append(savings, "Custom Fabrication Deluxe Mode active: TIG-welded heavy-gauge 316L/C-276 quench & venturi w/ ANSI B16.5 Class 150 flanges, filament-wound Derakane FRP column w/ bolted manway & demister cassette, welded 2\"×2\"×3/16\" steel forklift skid + K-bracing, UL 508A NEMA 4X panel, and 2\" Rockwool acoustic baffle box + tuned silencer (-22 dBA).")
	}
	if p.IsLocalCeramic() {
		fr := EvaluateFrugal(p, design)
		savings = append(savings, fmt.Sprintf("Local Ceramic / Masonry Mode (%s, draft=%s, zero_electricity=%v): %s — %.0f\" stoneware bed (%.3f inWC Ergun dP), %.1f m chimney (%.3f inWC avail vs %.3f inWC req), %.0f W electric, %.1f%% HCl / %.0f%% submicron NaCl capture.",
			fr.RegionLabel, fr.DraftMode, fr.ZeroElectricity, fr.LocalCurrencyEstimate, fr.BedDiameterIn, fr.BedDPInWC, fr.ChimneyHeightM, fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.PowerWatts, fr.HClRemovalPct, fr.SubmicronNaClCapturePct))
	}
	if len(savings) == 0 {
		savings = append(savings, "Standard baseline inline configuration active (toggle budget blower drive, shell-less 4\" parallel bundle, DIY ceramic rings, U-turn skid, or electromechanical controls to compare savings).")
	}

	return ArchitectureOptionsResult{
		StageAlignment:      p.StageAlignment,
		PackedBedLayout:     p.PackedBedLayout,
		VenturiLayout:       p.VenturiLayout,
		CompactFootprint:    p.CompactFootprint,
		SharedBaffledSump:   p.SharedBaffledSump,
		ControlTier:         p.ControlTier,
		BaselineCostUSD:     baseCOTS.EstimatedTotalUSD,
		SelectedCostUSD:     selectedCOTS.EstimatedTotalUSD,
		CostDeltaUSD:        costDelta,
		ControlDeltaUSD:     controlDelta,
		SumpDeltaUSD:        sumpDelta,
		FootprintLengthFt:   fpLenFt,
		FootprintWidthFt:    fpWidthFt,
		FootprintAreaSqFt:   fpAreaSqFt,
		BaselineLengthFt:    baseLenFt,
		FootprintSavingsPct: fpSavingsPct,
		Eductor:             eductor,
		ParallelBed:         parallelBed,
		ParallelVenturi:     parallelVenturi,
		BlowerDrive:         blowerDrive,
		Compressor:          compressor,
		PackingMedia:        packingMedia,
		SavingsSummary:      savings,
		Warnings:            warnings,
	}
}
