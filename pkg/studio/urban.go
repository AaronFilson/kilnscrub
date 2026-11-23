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

// DrainAssessment converts NaHCO3 neutralization into grocery-store boxes and drain pH guidance.
type DrainAssessment struct {
	Status                 string  `json:"status"`
	BakingSodaLbsPerFiring float64 `json:"baking_soda_lbs_per_firing"`
	BakingSoda4LbBoxes     int     `json:"baking_soda_4lb_boxes"`
	TargetDrainPH          string  `json:"target_drain_ph"`
	Summary                string  `json:"summary"`
}

// CorrosionAssessment surfaces 316L SS hot-wet-HCl pitting guidance and Soda-vs-Salt comparison.
type CorrosionAssessment struct {
	Status             string `json:"status"`
	Tier               string `json:"tier"`
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
// and corrosion/soda-vs-salt checks grounded in validated scrubber physics.
func AssessUrban(project StudioProject, design model.ScrubberDesign, cots COTSSummary) UrbanAssessment {
	elec := assessElectrical(project, design)
	water := assessWater(project, design)
	noise := assessNoise(project, design)
	plume := assessPlume(project, design)
	drain := assessDrain(project, design)
	corr := assessCorrosion(project, design, drain)

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
			summary = fmt.Sprintf("%.2g HP blower VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity — fits a %s.%s%s",
				hp, vfdRatedInputA, reqAmps, circuitLabel, diluterNote, necNotes)
		} else {
			circuitLabel = "120V 30A dedicated VFD circuit (10 AWG) — or 240V 20A"
			status = StatusCaution
			summary = fmt.Sprintf("%.2g HP blower VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity (exceeds standard 120V 20A garage outlet per NEC 430.122(A)) — install a 120V 30A circuit or upgrade to 240V 20A.%s%s",
				hp, vfdRatedInputA, reqAmps, diluterNote, necNotes)
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
			summary = fmt.Sprintf("%.2g HP blower VFD (%.1f A rated in x 125%%) + pump requires %.1f A min ampacity on 240V — ready for your %s.%s%s",
				hp, vfdRatedInputA, reqAmps, circuitLabel, diluterNote, necNotes)
		} else {
			status = StatusActionNeeded
			summary = fmt.Sprintf("%.2g HP blower requires a %s (%.1f A min ampacity per NEC 430.122(A)), but studio power is set to 120V. Add a 240V breaker/subpanel or downsize kiln/venturi throat velocity.%s%s",
				hp, circuitLabel, reqAmps, diluterNote, necNotes)
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
		// Continuous/batch blowdown to limit dissolved NaCl TDS buildup per ENGINEERING_VALIDATION.md B6.
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
		summary = fmt.Sprintf("Water flow capacity OK: %.2f GPM peak spray (%.2f GPM evaporated) vs %.1f GPM available; ~%.0f gal total per %.0f-hr firing (incl. sump fill & blowdown).%s", reqGPM, evapGPM, project.GardenHoseGPM, totalGal, project.FiringDurationHours, ipc608Note)
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
	// Account for slight thermal warming of the diluted plume at clearing (~0.85 mass factor vs isothermal).
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

	// Verdict on squirrel-cage blower suitability.
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
	nahco3Kg := design.PackedBed.NaHCO3ConsumptionKgPerHr * project.FiringDurationHours
	lbs := round1(nahco3Kg / model.KgPerLb)
	boxes := 0
	if lbs > 0 {
		boxes = int(math.Ceil(lbs / 4.0))
	}

	if lbs <= 0 {
		return DrainAssessment{
			Status:                 StatusOK,
			BakingSodaLbsPerFiring: 0,
			BakingSoda4LbBoxes:     0,
			TargetDrainPH:          targetPH,
			Summary:                fmt.Sprintf("Zero HCl neutralization required (%s mode: 0 lbs baking soda). Cool sump/blowdown water to <=140 F (IPC 702.5 PVC/ABS drain limit), strain ash/solids, and confirm pH %s (federal 40 CFR 403.5(b)(2) floor >=5.0) before sanitary drain release.", project.VaporGlazeMode, targetPH),
		}
	}

	return DrainAssessment{
		Status:                 StatusCaution,
		BakingSodaLbsPerFiring: lbs,
		BakingSoda4LbBoxes:     boxes,
		TargetDrainPH:          targetPH,
		Summary:                fmt.Sprintf("Stock %.1f lbs baking soda (%d x 4-lb Arm & Hammer boxes) for a %.0f-hr salt firing (15%% pH buffer). Cool/temper sump and live blowdown to <=140 F before PVC/ABS drains (IPC 702.5), settle wood ash, verify neutralized pH %s (IPC 803.1; federal 40 CFR 403.5(b)(2) strictly prohibits pH < 5.0), and check local POTW chloride/heavy-metal sludge limits.", lbs, boxes, project.FiringDurationHours, targetPH),
	}
}

func assessCorrosion(project StudioProject, design model.ScrubberDesign, drain DrainAssessment) CorrosionAssessment {
	hclPPM := design.PackedBed.HClInletPPM
	// Reference comparison: stoichiometric NaHCO3 for 2.0 lbs/hr salt over this firing duration
	// (1.485 kg/hr = 3.274 lbs/hr NaHCO3 with 15% buffer, matching validated reference case).
	refSaltRate := project.Input.SaltRateLbsPerHr
	if refSaltRate <= 0 {
		refSaltRate = 2.0
	}
	equivLbsNaHCO3 := round1((refSaltRate / 2.0) * 3.274 * project.FiringDurationHours)
	equivBoxes := int(math.Ceil(equivLbsNaHCO3 / 4.0))

	switch project.VaporGlazeMode {
	case VaporGlazeSoda:
		return CorrosionAssessment{
			Status:             StatusOK,
			Tier:               project.CorrosionTier,
			MaterialSummary:    "Soda ash (Na2CO3) vapor glazing generates ZERO HCl acid gas — standard 316L SS quench/venturi + CPVC packed column avoids hot wet chloride pitting and lasts for years.",
			SodaComparison:     fmt.Sprintf("Choosing Soda (Na2CO3) instead of %.1f lbs/hr Salt (NaCl) eliminates 100%% of HCl acid gas (0 ppm HCl) and saves ~%.1f lbs (%d x 4-lb boxes) of baking soda per %.0f-hr firing while eliminating 316L chloride stress-corrosion cracking.", refSaltRate, equivLbsNaHCO3, equivBoxes, project.FiringDurationHours),
			InspectionInterval: "Every 25-30 firings (annual weld, gasket & spray-nozzle inspection)",
		}
	case VaporGlazeNone:
		return CorrosionAssessment{
			Status:             StatusOK,
			Tier:               project.CorrosionTier,
			MaterialSummary:    "No salt/acid vapor load — 316L SS hot quench/venturi + CPVC/HDPE wet column provides long service life against combustion smoke, ash, and non-acidic fumes.",
			SodaComparison:     "Zero HCl acid gas (0 ppm HCl, 0 lbs NaHCO3). If you add vapor glazing later, choose Soda (Na2CO3) over Salt (NaCl) to avoid hot wet HCl corrosion.",
			InspectionInterval: "Every 25-30 firings (clean ash strainer and verify spray pattern)",
		}
	default: // VaporGlazeSalt
		if project.CorrosionTier == CorrosionTierHeavyC276FRP {
			return CorrosionAssessment{
				Status:             StatusOK,
				Tier:               project.CorrosionTier,
				MaterialSummary:    fmt.Sprintf("Industrial acid-proof tier: Alloy C-276 (UNS N10276) / refractory-lined quench & venturi throat + Derakane vinyl-ester FRP column & blower wheel — immune to hot wet HCl pitting & Cl-SCC at %.0f ppm HCl (~%.0f F quench).", hclPPM, design.Quench.OutletTemp.Fahrenheit()),
				SodaComparison:     fmt.Sprintf("Salt (NaCl) produces %.0f ppm post-quench HCl requiring %.1f lbs (%d x 4-lb boxes) baking soda per firing. Switching to Soda (Na2CO3) drops HCl to 0 ppm and allows lower-cost 316L/CPVC hardware.", hclPPM, drain.BakingSodaLbsPerFiring, drain.BakingSoda4LbBoxes),
				InspectionInterval: "Every 20-25 salt firings (inspect PTFE gaskets, refractory liner & nozzle orifices)",
			}
		}
		return CorrosionAssessment{
			Status:             StatusCaution,
			Tier:               project.CorrosionTier,
			MaterialSummary:    fmt.Sprintf("316L SS HOT-WET-HCl WARNING: Unbuffered hot wet HCl (%.0f ppm @ %.0f F) causes rapid chloride pitting & Cl-SCC in 316L SS above 140 F. ALWAYS recirculate alkaline NaHCO3 liquor (pH 8.0-8.5) to BOTH quench & venturi, use bolted replaceable Stage 1-2 spools, or upgrade to Heavy C-276/FRP.", hclPPM, design.Quench.OutletTemp.Fahrenheit()),
			SodaComparison:     fmt.Sprintf("Switching from Salt (NaCl) to Soda (Na2CO3) eliminates 100%% of HCl acid gas (%.0f ppm -> 0 ppm), saves %.1f lbs (%d x 4-lb boxes) of baking soda per firing, and removes 316L chloride pitting risk entirely.", hclPPM, drain.BakingSodaLbsPerFiring, drain.BakingSoda4LbBoxes),
			InspectionInterval: "Every 5-10 salt firings (inspect 316L weld seams, flange crevices & wall thickness)",
		}
	}
}
