package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/thermo"
)

// Traffic-light readiness status levels.
const (
	StatusOK           = "ok"
	StatusCaution      = "caution"
	StatusActionNeeded = "action_needed"
)

// ElectricalAssessment evaluates blower/pump circuit voltage and amperage against studio power.
type ElectricalAssessment struct {
	Status               string  `json:"status"`
	RequiredVoltage      int     `json:"required_voltage"`
	RequiredAmps         float64 `json:"required_amps"`
	CircuitLabel         string  `json:"circuit_label"`
	Summary              string  `json:"summary"`
	FitsAvailableVoltage bool    `json:"fits_available_voltage"`
}

// WaterAssessment evaluates continuous quench spray/makeup demand against a studio garden hose.
type WaterAssessment struct {
	Status             string  `json:"status"`
	RequiredGPM        float64 `json:"required_gpm"`
	TotalFiringGallons float64 `json:"total_firing_gallons"`
	Summary            string  `json:"summary"`
}

// NoiseAssessment predicts inverse-square blower sound level at the property line.
type NoiseAssessment struct {
	Status              string  `json:"status"`
	BareDBA3Ft          float64 `json:"bare_dba_3ft"`
	SilencedDBA3Ft      float64 `json:"silenced_dba_3ft"`
	PropertyLineDBA     float64 `json:"property_line_dba"`
	PropertyLineFt      float64 `json:"property_line_ft"`
	UrbanLimitDBA       float64 `json:"urban_limit_dba"`
	SilencerRecommended bool    `json:"silencer_recommended"`
	Summary             string  `json:"summary"`
}

// PlumeAssessment evaluates cold-weather visible steam condensation, stack height,
// psychrometric dilution/reheat demand, and squirrel-cage plume-shortening performance.
type PlumeAssessment struct {
	Status               string  `json:"status"`
	ExhaustTempF         float64 `json:"exhaust_temp_f"`
	DewPointF            float64 `json:"dew_point_f"`
	VisiblePlumeRisk     string  `json:"visible_plume_risk"`
	MinStackHeightFt     float64 `json:"min_stack_height_ft"`
	DilutionCFMNeeded    float64 `json:"dilution_cfm_needed"`
	DilutionReheatKW     float64 `json:"dilution_reheat_kw"`
	ColdDilutionWorks    bool    `json:"cold_dilution_works"`
	BarePlumeLengthFt    float64 `json:"bare_plume_length_ft"`
	DilutedPlumeLengthFt float64 `json:"diluted_plume_length_ft"`
	SquirrelCageVerdict  string  `json:"squirrel_cage_verdict"`
	Summary              string  `json:"summary"`
}

// DrainAssessment sizes bulk 15-lb bags / 50-lb sacks of baking soda (NaHCO3) for the
// active peak salting window and provides IPC 702.5 / 40 CFR 403.5 drain neutralization guidance.
type DrainAssessment struct {
	Status                 string  `json:"status"`
	SaltingHours           float64 `json:"salting_hours"`
	BakingSodaLbsPerFiring float64 `json:"baking_soda_lbs_per_firing"`
	BakingSoda15LbBags     int     `json:"baking_soda_15lb_bags"`
	BakingSoda50LbSacks    int     `json:"baking_soda_50lb_sacks"`
	BulkPackagingLabel     string  `json:"bulk_packaging_label"`
	TargetDrainPH          string  `json:"target_drain_ph"`
	Summary                string  `json:"summary"`
}

// CorrosionAssessment provides practical alloy selection, buffered flush, and weld inspection guidance.
type CorrosionAssessment struct {
	Status             string `json:"status"`
	Tier               string `json:"tier"`
	TierLabel          string `json:"tier_label"`
	MaterialSummary    string `json:"material_summary"`
	SodaComparison     string `json:"soda_comparison"`
	InspectionInterval string `json:"inspection_interval"`
}

// UrbanAssessment aggregates all 6 backyard / urban home-studio readiness checks.
type UrbanAssessment struct {
	OverallStatus string               `json:"overall_status"`
	Electrical    ElectricalAssessment `json:"electrical"`
	Water         WaterAssessment      `json:"water"`
	Noise         NoiseAssessment      `json:"noise"`
	Plume         PlumeAssessment      `json:"plume"`
	Drain         DrainAssessment      `json:"drain"`
	Corrosion     CorrosionAssessment  `json:"corrosion"`
}

func statusRank(s string) int {
	switch s {
	case StatusActionNeeded:
		return 2
	case StatusCaution:
		return 1
	default:
		return 0
	}
}

func worstStatus(statuses ...string) string {
	best := StatusOK
	for _, s := range statuses {
		if statusRank(s) > statusRank(best) {
			best = s
		}
	}
	return best
}

// AssessUrban performs electrical, water, neighbor-noise, steam-plume, neutralized-drain,
// and corrosion maintenance checks grounded in validated scrubber physics.
func AssessUrban(project StudioProject, design model.ScrubberDesign, cots COTSSummary) UrbanAssessment {
	p := project.Normalize()
	elec := assessElectrical(p, design)
	water := assessWater(p, design)
	noise := assessNoise(p, design)
	plume := assessPlume(p, design)
	drain := assessDrain(p, design)
	corr := assessCorrosion(p, design, drain)

	overall := worstStatus(
		elec.Status,
		water.Status,
		noise.Status,
		plume.Status,
		drain.Status,
		corr.Status,
	)

	return UrbanAssessment{
		OverallStatus: overall,
		Electrical:    elec,
		Water:         water,
		Noise:         noise,
		Plume:         plume,
		Drain:         drain,
		Corrosion:     corr,
	}
}

func diluterAmps120V(design model.ScrubberDesign) float64 {
	cfm, _ := RecommendedDilutionCFM(design)
	switch {
	case cfm <= 500.0:
		return 2.2
	case cfm <= 1000.0:
		return 4.1
	case cfm <= 1500.0:
		return 6.5
	case cfm <= 2000.0:
		return 8.8
	default:
		return 12.5
	}
}

func assessElectrical(project StudioProject, design model.ScrubberDesign) ElectricalAssessment {
	hp := design.Blower.RecommendedMotorHP
	blowerRole := "blower"
	if project.StageAlignment == StageAlignmentEductorDryFan {
		ed := analyzeEductor(project, design)
		if ed.DryFanMotorHP > 0 {
			hp = ed.DryFanMotorHP
		}
		blowerRole = "dry motive fan"
	}
	if hp <= 0 {
		hp = 0.5
	}

	var reqV int
	var vfdRatedInputA float64
	var auxAmps float64
	var diluterNote string
	diluterA120 := 0.0
	if project.IncludePlumeDiluter {
		diluterA120 = diluterAmps120V(design)
		diluterNote = fmt.Sprintf(" Includes +%.1f A (120V) downstream plume diluter load.", diluterA120)
	}

	// NEC 430.122(A): VFD line-side conductors must be sized >= 125% of VFD rated 1-phase
	// input current (not motor FLC), plus auxiliary recirc pump & control loads.
	switch {
	case hp <= 0.5:
		reqV = 120
		if hp <= 0.25 {
			vfdRatedInputA = 7.2 // GS21-10P5 (1/4 HP load class @ 120V 1-ph)
		} else {
			vfdRatedInputA = 11.1 // GS21-10P5 rated 1-ph VT input @ 120V
		}
		auxAmps = 3.0 // ~250W mag-drive recirc pump + ~50W instrumentation @ 120V
	case hp <= 1.0:
		if project.AvailableVoltage >= 240 {
			reqV = 240
			vfdRatedInputA = 9.7 // GS21-21P0 rated 1-ph VT input @ 240V
			auxAmps = 1.5
		} else {
			reqV = 120
			if hp <= 0.75 {
				vfdRatedInputA = 15.5 // GS21-11P0 derated @ 120V 1-ph
			} else {
				vfdRatedInputA = 20.6 // GS21-11P0 rated 1-ph VT input @ 120V
			}
			auxAmps = 3.0
		}
	case hp <= 1.5:
		reqV = 240
		vfdRatedInputA = 15.5 // GS21-22P0 @ 1.5 HP 240V 1-ph
		auxAmps = 1.5
	case hp <= 2.0:
		reqV = 240
		vfdRatedInputA = 18.5 // GS21-22P0 rated 1-ph VT input @ 240V
		auxAmps = 1.5
	case hp <= 3.0:
		reqV = 240
		vfdRatedInputA = 27.5 // GS21-23P0 rated 1-ph VT input @ 240V
		auxAmps = 1.5
	default:
		reqV = 240
		vfdRatedInputA = 27.5 + 7.25*(hp-3.0) // >=5 HP requires 50-60A 1-ph derated or 3-phase service
		auxAmps = 1.5
	}

	diluterCircuitA := diluterA120
	if reqV == 240 {
		diluterCircuitA = diluterA120 * 0.5 // equivalent 240V power draw (feed diluter via dedicated 120V leg)
	}
	reqAmps := round1(1.25*vfdRatedInputA + auxAmps + diluterCircuitA)

	var circuitLabel string
	var fits bool
	var status string
	var summary string

	const necNotes = " Install in fixed conduit (NEC 400.12) with a lockable NEMA 4X sight disconnect <=50 ft (NEC 430.102) and a line-side HF-rated GFCI breaker + 3% input line reactor (NEC 210.8(F))."

	if reqV == 120 {
		fits = true // Both 120V and 240V studios have 120V service.
		if reqAmps <= 20.0 {
			circuitLabel = "120V 20A dedicated VFD circuit (12 AWG, NEC 430.122)"
			status = StatusOK
			summary = fmt.Sprintf("%.2g HP %s VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity — fits a %s.%s%s",
				hp, blowerRole, vfdRatedInputA, reqAmps, circuitLabel, diluterNote, necNotes)
		} else {
			circuitLabel = "120V 30A dedicated VFD circuit (10 AWG) — or 240V 20A"
			status = StatusCaution
			summary = fmt.Sprintf("%.2g HP %s VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity (exceeds standard 120V 20A garage outlet per NEC 430.122(A)) — install a 120V 30A circuit or upgrade to 240V 20A.%s%s",
				hp, blowerRole, vfdRatedInputA, reqAmps, diluterNote, necNotes)
		}
	} else {
		switch {
		case reqAmps <= 20.0:
			circuitLabel = "240V 20A single-phase VFD circuit (12 AWG, NEC 430.122)"
		case reqAmps <= 30.0:
			circuitLabel = "240V 30A single-phase VFD circuit (10 AWG, NEC 430.122)"
		case reqAmps <= 40.0:
			circuitLabel = "240V 40A single-phase VFD circuit (8 AWG, NEC 430.122)"
		case reqAmps <= 50.0:
			circuitLabel = "240V 50A single-phase VFD circuit (6 AWG, NEC 430.122)"
		default:
			circuitLabel = "240V 60A single-phase (or 3-phase) VFD circuit (NEC 430.122)"
		}
		fits = project.AvailableVoltage >= 240
		if fits {
			status = StatusOK
			summary = fmt.Sprintf("%.2g HP %s VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity on 240V — ready for your %s.%s%s",
				hp, blowerRole, vfdRatedInputA, reqAmps, circuitLabel, diluterNote, necNotes)
		} else {
			status = StatusActionNeeded
			summary = fmt.Sprintf("%.2g HP %s requires a %s (%.1f A min ampacity per NEC 430.122(A)), but studio power is set to 120V. Add a 240V breaker/subpanel or downsize kiln/venturi throat velocity.%s%s",
				hp, blowerRole, circuitLabel, reqAmps, diluterNote, necNotes)
		}
	}

	return ElectricalAssessment{
		Status:               status,
		RequiredVoltage:      reqV,
		RequiredAmps:         reqAmps,
		CircuitLabel:         circuitLabel,
		Summary:              summary,
		FitsAvailableVoltage: fits,
	}
}

func assessWater(project StudioProject, design model.ScrubberDesign) WaterAssessment {
	reqGPM := round2(design.Quench.TotalSprayWaterGPM)
	evapGPM := design.Quench.EvaporatedWaterGPM
	blowdownGPM := 0.03
	if project.VaporGlazeMode == VaporGlazeSalt {
		blowdownGPM = 0.08
	}
	sumpFillGal := math.Max(25.0, float64(design.PackedBed.SumpVolume)/model.CubicMetersPerUSGallon)
	totalGal := round1(sumpFillGal + (evapGPM+blowdownGPM)*60.0*project.FiringDurationHours)

	const ipc608Note = " MANDATORY (IPC Section 608): isolate potable water via an engineered Air Gap (>=1\" above break-tank overflow rim, IPC 608.15.1) or certified ASSE 1013 RPZ backflow preventer — never connect a garden hose directly to a pressurized chemical manifold or submerged float valve."
	status := StatusOK
	var summary string
	switch {
	case reqGPM > project.GardenHoseGPM:
		status = StatusActionNeeded
		summary = fmt.Sprintf("Peak quench spray (%.2f GPM) exceeds your %.1f GPM supply. Add a dedicated 3/4\" supply or break/buffer tank (~%.0f gal total per %.0f-hr firing).%s", reqGPM, project.GardenHoseGPM, totalGal, project.FiringDurationHours, ipc608Note)
	case reqGPM > 0.70*project.GardenHoseGPM:
		status = StatusCaution
		summary = fmt.Sprintf("Peak quench spray (%.2f GPM) uses >70%% of your %.1f GPM supply — avoid sharing the spigot during firing (~%.0f gal total per %.0f-hr firing).%s", reqGPM, project.GardenHoseGPM, totalGal, project.FiringDurationHours, ipc608Note)
	default:
		summary = fmt.Sprintf("Water flow capacity OK: %.2f GPM peak spray (%.2f GPM evaporated) vs %.1f GPM available; ~%.0f gal total per %.0f-hr firing (closed-loop recirculation cuts makeup water 70-85%% vs once-through).%s", reqGPM, evapGPM, project.GardenHoseGPM, totalGal, project.FiringDurationHours, ipc608Note)
	}

	return WaterAssessment{
		Status:             status,
		RequiredGPM:        reqGPM,
		TotalFiringGallons: totalGal,
		Summary:            summary,
	}
}

func assessNoise(project StudioProject, design model.ScrubberDesign) NoiseAssessment {
	const (
		refDistFt       = 3.0
		urbanLimitDBA   = 60.0
		silencerDropDBA = 13.0
	)
	shaftHP := math.Max(0.15, design.Blower.ShaftPowerHP)
	spInWC := math.Max(1.0, design.Blower.TotalStaticPressureInWC)
	if project.StageAlignment == StageAlignmentEductorDryFan {
		ed := analyzeEductor(project, design)
		if ed.DryFanShaftHP > 0 {
			shaftHP = math.Max(0.15, ed.DryFanShaftHP)
			spInWC = math.Max(1.0, ed.DryFanStaticInWC)
		}
	}

	bare3Ft := round1(74.0 + 6.5*math.Log10(shaftHP) + 0.18*spInWC)
	silenced3Ft := round1(bare3Ft - silencerDropDBA)

	d2 := math.Max(refDistFt, project.PropertyLineFt)
	atten := 20.0 * math.Log10(d2/refDistFt)
	bareAtLine := bare3Ft - atten

	sourceDBA := bare3Ft
	if project.IncludeSilencer {
		sourceDBA = silenced3Ft
	}
	propDBA := round1(math.Max(25.0, sourceDBA-atten))
	silencerRec := bareAtLine > 55.0

	var status string
	var summary string
	switch {
	case propDBA > urbanLimitDBA:
		status = StatusActionNeeded
		if !project.IncludeSilencer {
			summary = fmt.Sprintf("Unsilenced blower reaches %.1f dBA at %.0f ft property line (above %.0f dBA residential daytime limit / 50-55 dBA nighttime ordinances). Enable the acoustic discharge silencer/box to drop ~13 dBA.", propDBA, d2, urbanLimitDBA)
		} else {
			summary = fmt.Sprintf("Even silenced, blower reaches %.1f dBA at %.0f ft (above %.0f dBA limit). Relocate blower indoors/behind a masonry wall or increase setback.", propDBA, d2, urbanLimitDBA)
		}
	case propDBA > 55.0:
		status = StatusCaution
		summary = fmt.Sprintf("Predicted %.1f dBA at %.0f ft property line — below the %.0f dBA daytime limit, but may exceed 50-55 dBA overnight ordinances during long firings. Acoustic box / silencer keeps late-night firings neighbor-friendly.", propDBA, d2, urbanLimitDBA)
	default:
		status = StatusOK
		modeLabel := "with acoustic silencer"
		if !project.IncludeSilencer {
			modeLabel = "bare (no silencer)"
		}
		summary = fmt.Sprintf("Quiet backyard operation: %.1f dBA at %.0f ft property line (%s) — comfortably under the %.0f dBA daytime limit (verify 50-55 dBA overnight local ordinance).", propDBA, d2, modeLabel, urbanLimitDBA)
	}

	if !project.IncludeSilencer && bare3Ft >= 78.0 {
		summary += fmt.Sprintf(" OSHA 29 CFR 1910.95 HEARING WARNING: unsilenced near-field blower (%.1f dBA at 3 ft, rising >=85 dBA near reflective walls) can exceed the 85 dBA 8-hr TWA action level — wear hearing protection or install the acoustic enclosure.", bare3Ft)
	}

	return NoiseAssessment{
		Status:              status,
		BareDBA3Ft:          bare3Ft,
		SilencedDBA3Ft:      silenced3Ft,
		PropertyLineDBA:     propDBA,
		PropertyLineFt:      d2,
		UrbanLimitDBA:       urbanLimitDBA,
		SilencerRecommended: silencerRec,
		Summary:             summary,
	}
}

// waterPsatSafePa evaluates IAPWS-95 saturation vapor pressure via pkg/thermo.WaterPsat
// above the triple point (273.16 K) and Magnus-over-ice below freezing so sub-32 F
// winter sites are handled smoothly.
func waterPsatSafePa(tK model.Kelvin) float64 {
	if float64(tK) >= 273.16 {
		if ps, err := thermo.WaterPsat(tK); err == nil {
			return float64(ps)
		}
	}
	t := float64(tK)
	return 611.657 * math.Exp(22.452*(t-273.16)/math.Max(t-0.61, 1.0))
}

func humidityRatioSat(tK model.Kelvin, pAmbPa float64) float64 {
	ps := waterPsatSafePa(tK)
	if ps >= 0.99*pAmbPa {
		ps = 0.99 * pAmbPa
	}
	y := ps / pAmbPa
	return 0.6245 * y / math.Max(1.0-y, 1e-6)
}

func assessPlume(project StudioProject, design model.ScrubberDesign) PlumeAssessment {
	exhaustF := round1(design.OutletGas.Temperature.Fahrenheit())
	dewF := round1(design.Quench.DewPoint.Fahrenheit())

	minStackFt := 10.0
	if design.Input.KilnVolumeCuFt >= 40.0 {
		minStackFt = 15.0
	} else if design.Input.KilnVolumeCuFt >= 20.0 || project.PropertyLineFt < 35.0 {
		minStackFt = 12.0
	}

	// Psychrometric state using IAPWS-95 WaterPsat (conservative 60% winter RH).
	const rhAmb = 0.60
	pAmb := float64(design.Input.AmbientPressurePa)
	if pAmb <= 0 {
		pAmb = float64(model.StandardPressurePa)
	}
	tAmbK := model.FahrenheitToKelvin(project.AmbientWinterF)
	tExhK := design.OutletGas.Temperature
	if tExhK <= tAmbK+1.0 {
		tExhK = tAmbK + 1.0
	}

	yWAmb := math.Min(0.95, rhAmb*waterPsatSafePa(tAmbK)/pAmb)
	wAmb := 0.6245 * yWAmb / math.Max(1.0-yWAmb, 1e-6)

	mH2O := design.OutletGas.MolarFlowMolPerSec * design.OutletGas.H2O * 0.01801528
	mTotal := float64(design.OutletGas.MassFlow)
	if mTotal <= 0 {
		mTotal = math.Max(mH2O*1.8, 1e-4)
	}
	mDry := math.Max(1e-6, mTotal-mH2O)
	wExh := mH2O / mDry

	// 1. Check whether unheated cold dilution stays below saturation along the mixing line.
	coldDilutionWorks := true
	for _, frac := range []float64{0.15, 0.35, 0.50, 0.70, 0.85} {
		tMix := model.Kelvin(float64(tAmbK) + frac*(float64(tExhK)-float64(tAmbK)))
		wMix := wAmb + frac*(wExh-wAmb)
		if wMix > humidityRatioSat(tMix, pAmb)+1e-6 {
			coldDilutionWorks = false
			break
		}
	}

	// 2. Tangent-clear reheat power (kW) invariant across dilution air temperature (0 ft fog).
	sMin := math.Inf(1)
	tMaxScan := math.Min(372.0, float64(tExhK))
	for t := float64(tAmbK) + 0.5; t <= tMaxScan; t += 0.5 {
		wSatT := humidityRatioSat(model.Kelvin(t), pAmb)
		slope := (wSatT - wAmb) / (t - float64(tAmbK))
		if slope > 0 && slope < sMin {
			sMin = slope
		}
	}
	qReheatKW := 0.0
	if !math.IsInf(sMin, 1) && sMin > 0 && mH2O > mDry*wAmb {
		qReq := (mH2O-mDry*wAmb)*(1.01/sMin) - (mDry*1.01+mH2O*1.89)*(float64(tExhK)-float64(tAmbK))
		if qReq > 0 {
			qReheatKW = round1(qReq)
		}
	}

	// 3. Ambient dilution airflow required for RH <= 100% (used for turbulent plume-length scaling).
	rhoAmb := (pAmb * 0.028964) / (model.UniversalGasConst * float64(tAmbK))
	wSatAmb := humidityRatioSat(tAmbK, pAmb)
	deltaWSat := math.Max(wSatAmb-wAmb, 1e-4)
	mAmbClearKgS := 0.85 * math.Max(0.0, mH2O-mDry*wAmb) / deltaWSat
	qAmbClearM3S := mAmbClearKgS / math.Max(rhoAmb, 0.5)
	qAmbClearCFM := model.CubicMetersPerSec(qAmbClearM3S).CFM()

	// Practical 120V high-velocity squirrel-cage dilution blower sizing (~5x-6x scrubber ACFM).
	dilutionCFM, coneIn := RecommendedDilutionCFM(design)
	shroudCFM250 := 0.0
	if qReheatKW > 0 {
		tShroudK := float64(model.FahrenheitToKelvin(250.0))
		if tShroudK > float64(tAmbK)+5.0 {
			mShroud := qReheatKW / (1.01 * (tShroudK - float64(tAmbK)))
			shroudCFM250 = round1(model.CubicMetersPerSec(mShroud / math.Max(rhoAmb, 0.5)).CFM())
		}
	}

	// 4. Visible horizontal plume reach (5 mph Neutral Class D wind, calibrated to Briggs/Fischer).
	bareLenFt := 0.0
	dilutedLenFt := 0.0
	if !coldDilutionWorks && qAmbClearM3S > 0 {
		bareLenFt = round1(10.3 * math.Sqrt(qAmbClearM3S))
		if qAmbClearCFM > dilutionCFM {
			qRemRatio := (qAmbClearCFM - dilutionCFM) / qAmbClearCFM
			exhShare := design.Blower.FlowACFM / math.Max(design.Blower.FlowACFM+dilutionCFM, 1.0)
			dilutedLenFt = round1(math.Max(0.5, bareLenFt*0.42*math.Pow(qRemRatio, 1.35)*math.Pow(exhShare, 0.15)))
		}
	}

	squirrelVerdict := fmt.Sprintf(
		"A 120V squirrel-cage blower CANNOT replace the main Stage 4 radial ID blower (%.1f inWC static pressure exceeds the ~1.5-2.5 inWC squirrel-cage stall limit), "+
			"but it DOES hold up downstream at ~0.7-1.2 inWC in a stack mixing tee + %.0f\" high-velocity exit cone (~3,400 FPM). "+
			"Unheated cold air alone stays supersaturated at the stack exit (cold_dilution_works=false), yet a %.0f CFM 120V squirrel-cage HV jet cuts visible winter plume reach from ~%.0f ft bare down to ~%.0f ft (vs your %.0f ft property line). "+
			"For 0 ft tangent-clear winter invisibility (%.1f kW equivalent), pull ~250 F waste heat from a kiln-jacket/flue shroud (~%.0f CFM, blower mounted on the cool intake side) — avoid electric duct heaters.",
		design.Blower.TotalStaticPressureInWC, coneIn, dilutionCFM, bareLenFt, dilutedLenFt,
		project.PropertyLineFt, qReheatKW, shroudCFM250,
	)

	var risk, status, summary string
	activeLenFt := bareLenFt
	if project.IncludePlumeDiluter {
		activeLenFt = dilutedLenFt
	}

	hclOutPPM := design.PackedBed.HClOutletPPM
	isSaltHCl := project.VaporGlazeMode == VaporGlazeSalt || hclOutPPM > 0.0
	var exhaustSafetyNote string
	var setbackActionNeeded bool
	if isSaltHCl {
		exhaustSafetyNote = fmt.Sprintf(
			" SALT-FIRING HCl NOTICE: post-scrubber stack exit contains ~%.1f ppm HCl (OSHA 29 CFR 1910.1000 ceiling PEL = 5 ppm — NEVER discharge indoors or into a breathing zone). Discharge strictly outdoors via a >=%.0f ft elevated vertical stack meeting IMC 510 / 501.3.1 hazardous exhaust setbacks (>=30 ft property line, >=10 ft operable windows/air intakes, >=6 ft walls/roofs).",
			hclOutPPM, minStackFt,
		)
		if project.PropertyLineFt < 30.0 {
			setbackActionNeeded = true
			exhaustSafetyNote += fmt.Sprintf(" ACTION NEEDED: %.0f ft property line is below the 30 ft IMC 501.3.1 hazardous HCl exhaust setback.", project.PropertyLineFt)
		}
	} else {
		exhaustSafetyNote = fmt.Sprintf(
			" Exhaust is %.0f F water vapor with 0 ppm HCl (Soda/Non-salt mode); run a >=%.0f ft outdoor stack with condensate drain meeting IMC 501.3.1 setbacks (>=10 ft property line, >=10 ft operable windows/air intakes, >=3 ft walls/roofs).",
			exhaustF, minStackFt,
		)
		if project.PropertyLineFt < 10.0 {
			setbackActionNeeded = true
			exhaustSafetyNote += fmt.Sprintf(" ACTION NEEDED: %.0f ft property line is below the 10 ft IMC 501.3.1 exhaust setback.", project.PropertyLineFt)
		}
	}

	switch {
	case project.AmbientWinterF <= 45.0:
		risk = "high"
		if project.IncludePlumeDiluter && dilutedLenFt <= 0.75*project.PropertyLineFt {
			status = StatusOK
			summary = fmt.Sprintf("Squirrel-cage HV plume diluter (%.0f CFM) shrinks winter steam reach from ~%.0f ft bare down to ~%.0f ft — dissipating cleanly before your %.0f ft property line.%s",
				dilutionCFM, bareLenFt, dilutedLenFt, project.PropertyLineFt, exhaustSafetyNote)
		} else {
			status = StatusCaution
			summary = fmt.Sprintf("In %.0f F winter air, saturated %.0f F scrubber exhaust forms a ~%.0f ft visible white steam cloud. Adding a %.0f CFM 120V squirrel-cage diluter + %.0f\" HV cone shrinks visible reach to ~%.0f ft (vs %.0f ft property line).%s",
				project.AmbientWinterF, exhaustF, bareLenFt, dilutionCFM, coneIn, dilutedLenFt, project.PropertyLineFt, exhaustSafetyNote)
		}
	case project.AmbientWinterF <= 65.0:
		risk = "moderate"
		if project.IncludePlumeDiluter || activeLenFt <= 0.50*project.PropertyLineFt {
			status = StatusOK
		} else {
			status = StatusCaution
		}
		summary = fmt.Sprintf("Cool %.0f F air creates a ~%.0f ft bare white steam wisp (~%.0f ft with %.0f CFM squirrel-cage HV diluter).%s",
			project.AmbientWinterF, bareLenFt, dilutedLenFt, dilutionCFM, exhaustSafetyNote)
	default:
		risk = "low"
		status = StatusOK
		summary = fmt.Sprintf("Mild %.0f F ambient air disperses %.0f F steam quickly (~%.0f ft bare, ~%.0f ft diluted).%s",
			project.AmbientWinterF, exhaustF, bareLenFt, dilutedLenFt, exhaustSafetyNote)
	}
	if setbackActionNeeded {
		status = StatusActionNeeded
	}

	return PlumeAssessment{
		Status:               status,
		ExhaustTempF:         exhaustF,
		DewPointF:            dewF,
		VisiblePlumeRisk:     risk,
		MinStackHeightFt:     minStackFt,
		DilutionCFMNeeded:    dilutionCFM,
		DilutionReheatKW:     qReheatKW,
		ColdDilutionWorks:    coldDilutionWorks,
		BarePlumeLengthFt:    bareLenFt,
		DilutedPlumeLengthFt: dilutedLenFt,
		SquirrelCageVerdict:  squirrelVerdict,
		Summary:              summary,
	}
}

func assessDrain(project StudioProject, design model.ScrubberDesign) DrainAssessment {
	const targetPH = "6.5 - 8.5"
	saltingHrs := project.EffectiveSaltingHours()
	nahco3Kg := design.PackedBed.NaHCO3ConsumptionKgPerHr * saltingHrs
	lbs := round1(nahco3Kg / model.KgPerLb)
	bags15 := 0
	sacks50 := 0
	if lbs > 0 {
		bags15 = int(math.Ceil(lbs / 15.0))
		sacks50 = int(math.Ceil(lbs / 50.0))
	}

	if lbs <= 0 {
		return DrainAssessment{
			Status:                 StatusOK,
			SaltingHours:           0,
			BakingSodaLbsPerFiring: 0,
			BakingSoda15LbBags:     0,
			BakingSoda50LbSacks:    0,
			BulkPackagingLabel:     "0 lbs required (zero HCl mode)",
			TargetDrainPH:          targetPH,
			Summary:                fmt.Sprintf("Zero HCl neutralization required (%s mode: 0 lbs baking soda). Cool sump/blowdown water to <=140 F (IPC 702.5 PVC/ABS drain limit), strain ash/solids, and confirm pH %s (federal 40 CFR 403.5(b)(2) floor >=5.0) before sanitary drain release.", project.VaporGlazeMode, targetPH),
		}
	}

	firingsPerSack := maxInt(1, int(math.Floor(50.0/lbs)))
	bagWord := "bag"
	if bags15 > 1 {
		bagWord = "bags"
	}
	bulkLabel := fmt.Sprintf("%d × 15 lb bulk %s (~%.1f lbs used over %.1f hr salting; 1 × 50 lb sack covers ~%d firings)",
		bags15, bagWord, lbs, saltingHrs, firingsPerSack)

	return DrainAssessment{
		Status:                 StatusCaution,
		SaltingHours:           saltingHrs,
		BakingSodaLbsPerFiring: lbs,
		BakingSoda15LbBags:     bags15,
		BakingSoda50LbSacks:    sacks50,
		BulkPackagingLabel:     bulkLabel,
		TargetDrainPH:          targetPH,
		Summary:                fmt.Sprintf("Stock %.1f lbs bulk baking soda (%s) for the %.1f-hr active salting window (15%% pH buffer). Cool/temper sump and live blowdown to <=140 F before PVC/ABS drains (IPC 702.5), settle wood ash, verify neutralized pH %s (IPC 803.1; federal 40 CFR 403.5(b)(2) strictly prohibits pH < 5.0), and check local POTW chloride/heavy-metal sludge limits.", lbs, bulkLabel, saltingHrs, targetPH),
	}
}

func assessCorrosion(project StudioProject, design model.ScrubberDesign, drain DrainAssessment) CorrosionAssessment {
	hclPPM := design.PackedBed.HClInletPPM
	tierLabel := CorrosionTierLabel(project.CorrosionTier)
	saltingHrs := project.EffectiveSaltingHours()

	switch project.VaporGlazeMode {
	case VaporGlazeSoda:
		return CorrosionAssessment{
			Status:             StatusOK,
			Tier:               project.CorrosionTier,
			TierLabel:          tierLabel,
			MaterialSummary:    fmt.Sprintf("%s: Soda ash (Na2CO3) vapor glazing generates zero HCl acid gas — standard 316L SS quench/venturi + Sch 80 CPVC packed column avoids chloride pitting and delivers long service life.", tierLabel),
			SodaComparison:     "Zero HCl acid gas load (0 ppm HCl, 0 lbs baking soda required) allows standard 316L SS and Sch 80 CPVC hardware; periodically rinse dried soda deposits from spray nozzles and the 50-mesh pump strainer.",
			InspectionInterval: "Every 25-30 firings (annual weld, gasket & spray-nozzle inspection)",
		}
	case VaporGlazeNone:
		return CorrosionAssessment{
			Status:             StatusOK,
			Tier:               project.CorrosionTier,
			TierLabel:          tierLabel,
			MaterialSummary:    fmt.Sprintf("%s: No salt/acid vapor load — standard 316L SS hot quench/venturi + Sch 80 CPVC wet column provides long service life against combustion smoke, ash, and non-acidic fumes.", tierLabel),
			SodaComparison:     "Zero HCl acid gas load (0 ppm HCl, 0 lbs baking soda required) allows standard 316L SS and Sch 80 CPVC hardware; clean the 50-mesh ash strainer and check spray cone uniformity every 25-30 firings.",
			InspectionInterval: "Every 25-30 firings (clean ash strainer and verify spray pattern)",
		}
	default: // VaporGlazeSalt
		if project.CorrosionTier == CorrosionTierHeavyC276FRP {
			return CorrosionAssessment{
				Status:             StatusOK,
				Tier:               project.CorrosionTier,
				TierLabel:          tierLabel,
				MaterialSummary:    fmt.Sprintf("%s: Alloy C-276 (UNS N10276) / refractory-lined quench & venturi throat + Derakane vinyl-ester FRP column & blower wheel — immune to hot wet HCl pitting & Cl-SCC at %.0f ppm HCl (~%.0f F quench).", tierLabel, hclPPM, design.Quench.OutletTemp.Fahrenheit()),
				SodaComparison:     fmt.Sprintf("Salt-firing maintenance guidance: maintain buffered sump pH 7.5-8.5 throughout the %.1f-hr salting window (%.1f lbs bulk NaHCO3 per firing), run a 5-minute clean-water flush during cooldown so NaCl crystals never foul nozzles, and inspect PTFE flange gaskets & refractory liner every 20-25 firings.", saltingHrs, drain.BakingSodaLbsPerFiring),
				InspectionInterval: "Every 20-25 salt firings (inspect PTFE gaskets, refractory liner & nozzle orifices)",
			}
		}
		return CorrosionAssessment{
			Status:             StatusCaution,
			Tier:               project.CorrosionTier,
			TierLabel:          tierLabel,
			MaterialSummary:    fmt.Sprintf("%s — HOT-WET-HCl CARE: Unbuffered hot wet HCl (%.0f ppm @ %.0f F) accelerates chloride pitting & Cl-SCC in 316L SS above 140 F. ALWAYS recirculate alkaline NaHCO3 liquor (pH 8.0-8.5) to BOTH quench & venturi, use bolted replaceable Stage 1-2 spools, or select Heavy Duty C-276/FRP for frequent community firings.", tierLabel, hclPPM, design.Quench.OutletTemp.Fahrenheit()),
			SodaComparison:     fmt.Sprintf("Salt-firing maintenance guidance: pre-charge the sump to pH 8.0-8.5 before lighting, dose %.1f lbs bulk baking soda across the %.1f-hr salting window, run a 5-minute buffered pH 8 flush during cooldown to clear chloride films, and inspect 316L weld seams & flange crevices every 5-10 salt firings.", drain.BakingSodaLbsPerFiring, saltingHrs),
			InspectionInterval: "Every 5-10 salt firings (inspect 316L weld seams, flange crevices & wall thickness)",
		}
	}
}
