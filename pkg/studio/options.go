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
// with N parallel 3" or 4" Schedule 80 CPVC stock plumbing pipes and smaller packing
// sized to keep wall-channeling ratio D/d_p >= 7.5.
type ParallelBedAnalysis struct {
	Enabled             bool     `json:"enabled"`
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

// ArchitectureOptionsResult summarizes all Phase 7 selectable cost, time, footprint,
// and stage-design options and their exact impact on BOM total, footprint, and physics.
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
	wetBlowerPart := selectWetBlowerVFD(project, design)
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

func analyzeParallelBed(project StudioProject, design model.ScrubberDesign, eductor EductorAnalysis) ParallelBedAnalysis {
	enabled := project.PackedBedLayout == PackedBedLayoutParallelManifold

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

	// Pick 3" Sch 80 CPVC (2.900" ID) for tiny kilns (<= 8 cu.ft.) or 4" Sch 80 CPVC (3.826" ID) standard.
	pipeNominalIn := 4.0
	pipeActualIDIn := 3.826
	packingDiamIn := 0.50 // 1/2" PP Pall rings/saddles -> D/d_p = 7.65 >= 7.5
	fpSmall := 310.0      // 1/2" PP random packing factor [1/m]
	epsSmall := 0.87
	pipeCostPerFt := 10.50
	packingCostPerCuFt := 62.0

	if design.Input.KilnVolumeCuFt <= 8.0 && !eductor.Enabled {
		pipeNominalIn = 3.0
		pipeActualIDIn = 2.900
		packingDiamIn = 0.375 // 3/8" PP saddles/rings -> D/d_p = 7.73 >= 7.5
		fpSmall = 390.0
		epsSmall = 0.85
		pipeCostPerFt = 7.50
		packingCostPerCuFt = 68.0
	}

	wallRatio := round2(pipeActualIDIn / packingDiamIn)
	pipeIDM := float64(model.InchesToMeters(pipeActualIDIn))
	pipeAreaM2 := math.Pi * pipeIDM * pipeIDM / 4.0
	rhoL := float64(model.WaterDensity)
	qRecirc := math.Max(0.0028*qGas, float64(design.PackedBed.RecirculationFlow))

	vFlood := 1.8
	if vf, err := scrubber.EckertFloodingVelocity(qRecirc*rhoL, gMass, rhoG, rhoL, fpSmall, 0.001); err == nil && vf > 0 {
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
	packDiamM := float64(model.InchesToMeters(packingDiamIn))
	specArea := 207.0 * (0.0254 / packDiamM)
	sauterDp := 6.0 * (1.0 - epsSmall) / specArea
	bedDPInWC := design.PackedBed.BedPressureDrop.InWC()
	if dryDP, err := scrubber.ErgunPressureDrop(sauterDp, epsSmall, vSup, rhoG, muG, packHeightM); err == nil {
		bedDPInWC = round2(model.Pascal(float64(dryDP) * 1.70).InWC())
	}
	singleDPInWC := round2(design.PackedBed.BedPressureDrop.InWC())

	// Economics comparison: N stock Sch 80 CPVC pipes + top/bottom manifold plenums + liquid distribution header + smaller packing
	// vs single large column + 1" Pall rings.
	singleColPart, singleIDIn, _ := selectSinglePackedColumn(project, design, eductor)
	singlePackPart, singlePackCuFt := selectPackingMedia(singleIDIn, design)
	singleCost := round2(singleColPart.UnitCostUSD + singlePackCuFt*singlePackPart.UnitCostUSD)

	colHeightFt := math.Max(4.0, design.PackedBed.ColumnHeight.Feet())
	pipesCost := float64(pipeCount) * colHeightFt * pipeCostPerFt
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		pipesCost *= 1.35 // heavy chemical-grade Sch 80 CPVC / FRP-wrapped risers
	}
	manifoldPlenumsCost := 135.0 + 24.0*float64(pipeCount) // top/bottom collector headers + bulkheads/unions
	liquidHeaderCost := 45.0 + 11.0*float64(pipeCount)     // equal-flow orifice/spray header for each riser
	packVolCuFt := math.Max(1.0, math.Ceil(model.CubicMeters(totalArea*packHeightM).CuFt()))
	smallPackCost := packVolCuFt * packingCostPerCuFt
	manifoldCost := round2(pipesCost + manifoldPlenumsCost + liquidHeaderCost + smallPackCost)
	costDelta := round2(manifoldCost - singleCost)

	notes := []string{
		fmt.Sprintf("Wall-channeling check: %.3f\" pipe ID / %.3f\" packing = D/d_p %.2f (>= 7.5 required to prevent gas bypassing down the pipe wall).", pipeActualIDIn, packingDiamIn, wallRatio),
		fmt.Sprintf("MANDATORY LIQUID DISTRIBUTION HEADER: Install a balanced %.0f-port calibrated orifice/nozzle spray header at the top of the manifold so every %.0f\" riser receives equal %.2f GPM wetting flow (prevents dry channeling in one pipe while another floods).", float64(pipeCount), pipeNominalIn, (qRecirc/model.CubicMetersPerSecPerGPM)/float64(pipeCount)),
		fmt.Sprintf("Pressure drop trade-off: %.3f\" packing raises irrigated Ergun bed dP from %.2f inWC (single column, 1\" rings) to %.2f inWC.", packingDiamIn, singleDPInWC, bedDPInWC),
	}

	summary := fmt.Sprintf(
		"%d × %.0f\" Sch 80 CPVC stock plumbing pipes (ID %.3f\") filled with %.3g\" PP packing (D/d_p = %.2f, %.1f%% flood, %.2f inWC bed dP). Manifold total $%.0f vs $%.0f single %.0f\" column (%+.0f USD) — eliminates specialty large-diameter pipe lead times and freight.",
		pipeCount, pipeNominalIn, pipeActualIDIn, packingDiamIn, wallRatio, actualFloodPct, bedDPInWC,
		manifoldCost, singleCost, singleColPart.NominalIn, costDelta,
	)

	return ParallelBedAnalysis{
		Enabled:             enabled,
		PipeNominalIn:       pipeNominalIn,
		PipeActualIDIn:      pipeActualIDIn,
		PipeCount:           pipeCount,
		PackingDiameterIn:   packingDiamIn,
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

// EvaluateArchitectureOptions computes the complete Phase 7 trade-off summary,
// exact baseline-vs-selected COTS cost delta, footprint savings, and warnings.
func EvaluateArchitectureOptions(project StudioProject, design model.ScrubberDesign, selectedCOTS COTSSummary) ArchitectureOptionsResult {
	p := project.Normalize()

	// Build baseline (all default Phase 7 options) COTS summary for exact apples-to-apples cost delta.
	baseProj := p
	baseProj.StageAlignment = StageAlignmentInlineIDBlower
	baseProj.PackedBedLayout = PackedBedLayoutSingleColumn
	baseProj.VenturiLayout = VenturiLayoutSingleThroat
	baseProj.CompactFootprint = false
	baseProj.SharedBaffledSump = true
	baseProj.ControlTier = ControlTierRPiModbus

	baseCOTS := BuildCOTSSummary(baseProj, design)
	eductor := analyzeEductor(p, design)
	parallelBed := analyzeParallelBed(p, design, eductor)
	parallelVenturi := analyzeParallelVenturi(p, design)

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

	if eductor.Enabled {
		savings = append(savings, fmt.Sprintf("Dry-Fan Downflow Eductor: %+.0f USD blower delta (eliminates wet acid-duty blower; +%.1f F CPVC thermal margin; Stage 3 enlarged to %.0f\" @ %.1f%% flood).",
			eductor.NetBlowerDeltaUSD, eductor.ThermalMarginGainF, eductor.EnlargedBedDiameterIn, eductor.EnlargedBedFloodPct))
		warnings = append(warnings, eductor.Warnings...)
	}
	if parallelBed.Enabled {
		savings = append(savings, fmt.Sprintf("Parallel %.0f\" Pipe Packed-Bed Manifold (%d risers, %.3g\" packing): %+.0f USD vs single column; D/d_p = %.2f, %.2f inWC bed dP.",
			parallelBed.PipeNominalIn, parallelBed.PipeCount, parallelBed.PackingDiameterIn, parallelBed.CostDeltaUSD, parallelBed.WallRatioDToDp, parallelBed.BedDPInWC))
		warnings = append(warnings, fmt.Sprintf("Parallel Bed Distribution: Use calibrated %d-port top liquid distributor so no riser runs dry while another floods.", parallelBed.PipeCount))
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
	if len(savings) == 0 {
		savings = append(savings, "Standard baseline inline configuration active (toggle eductor, parallel manifold, U-turn skid, or electromechanical controls to compare savings).")
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
		SavingsSummary:      savings,
		Warnings:            warnings,
	}
}
