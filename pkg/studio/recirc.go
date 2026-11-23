package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// RecirculationPlan models closed-loop aqueous sump chemistry across a firing,
// water conservation vs once-through scrubbing, bulk 15-lb/50-lb baking soda sizing
// for the active peak salting window, dissolved-solids TDS accumulation, and exact
// field dump triggers (brine hydrometer SG, pH, ash depth, and IPC 702.5 <=140 F rule).
type RecirculationPlan struct {
	SumpVolumeGal              float64  `json:"sump_volume_gal"`
	OnceThroughWaterGal        float64  `json:"once_through_water_gal"`
	RecircMakeupWaterGal       float64  `json:"recirc_makeup_water_gal"`
	WaterSavedPct              float64  `json:"water_saved_pct"`
	SaltingHours               float64  `json:"salting_hours"`
	HClNeutralizedLbs          float64  `json:"hcl_neutralized_lbs"`
	BakingSodaLbsPerFiring     float64  `json:"baking_soda_lbs_per_firing"`
	BakingSoda15LbBags         int      `json:"baking_soda_15lb_bags"`
	BakingSoda50LbSacks        int      `json:"baking_soda_50lb_sacks"`
	InitialChargeLbs           float64  `json:"initial_charge_lbs"`
	EndOfFiringTDSGL           float64  `json:"end_of_firing_tds_g_l"`
	DumpTDSLimitGL             float64  `json:"dump_tds_limit_g_l"`
	DumpSpecificGravity        float64  `json:"dump_specific_gravity"`
	FiringsBeforeFullDump      int      `json:"firings_before_full_dump"`
	TopUpSodaLbsBetweenFirings float64  `json:"top_up_soda_lbs_between_firings"`
	DumpTriggers               []string `json:"dump_triggers"`
	ScheduleSteps              []string `json:"schedule_steps"`
	Summary                    string   `json:"summary"`
}

func workingSumpVolumeGal(cots COTSSummary, design model.ScrubberDesign) float64 {
	// Derive working liquid fill (~80% of selected drum/tank capacity) from selected COTS sump.
	switch cots.SumpTank.SKU {
	case "ULINE-S-16913":
		return 24.0 // 30-gal drum working fill
	case "ULINE-S-10757":
		return 45.0 // 55-gal drum working fill
	case "USPLASTIC-100T":
		return 85.0 // 100-gal tank working fill
	case "DLX-SUMP-SLOPED-35":
		return 28.0 // 35-gal Deluxe sloped-bottom baffled sump working fill
	case "DLX-SUMP-SLOPED-60":
		return 48.0 // 60-gal Deluxe sloped-bottom baffled sump working fill
	case "DLX-SUMP-SLOPED-110":
		return 90.0 // 110-gal Deluxe sloped-bottom baffled sump working fill
	default:
		gal := float64(design.PackedBed.SumpVolume) / model.CubicMetersPerUSGallon
		return round1(math.Max(25.0, gal))
	}
}

// BuildRecirculationPlan computes closed-loop sump water savings, peak-window baking soda
// dosing, dissolved NaCl+NaHCO3 TDS accumulation, and multi-firing dump schedule.
func BuildRecirculationPlan(project StudioProject, design model.ScrubberDesign, cots COTSSummary) RecirculationPlan {
	p := project.Normalize()
	const (
		dumpTDSLimitGL      = 150.0 // g/L safe operating ceiling (well below 360 g/L NaCl saturation)
		dumpSpecificGravity = 1.105 // ~14% brine SG measurable with a $12 brine/salinity hydrometer
		hclMWKgPerMol       = 0.03646094
		naclMWGPerMol       = 58.4428
		nahco3MWGPerMol     = 84.0066
	)

	sumpGal := workingSumpVolumeGal(cots, design)
	if !p.SharedBaffledSump {
		// Dual-sump arrangement holds extra working volume (~1.6x single sump).
		sumpGal = round1(sumpGal * 1.6)
	}
	sumpLiters := sumpGal * 3.785411784

	saltingHrs := p.EffectiveSaltingHours()
	firingHrs := p.FiringDurationHours

	// HCl neutralized & NaHCO3 reagent consumed ONLY during the active peak salting window.
	nHClRemovedMolS := design.QuenchedGas.MolarFlowMolPerSec * design.QuenchedGas.HCl * design.PackedBed.HClRemovalEfficiency
	hclNeutralizedKg := nHClRemovedMolS * hclMWKgPerMol * 3600.0 * saltingHrs
	hclNeutralizedLbs := round2(hclNeutralizedKg / model.KgPerLb)

	nahco3Kg := design.PackedBed.NaHCO3ConsumptionKgPerHr * saltingHrs
	bakingSodaLbs := round1(nahco3Kg / model.KgPerLb)
	bags15 := 0
	sacks50 := 0
	if bakingSodaLbs > 0 {
		bags15 = int(math.Ceil(bakingSodaLbs / 15.0))
		sacks50 = int(math.Ceil(bakingSodaLbs / 50.0))
	}

	initialChargeLbs := 0.0
	if bakingSodaLbs > 0 {
		// Pre-charge sump to pH 8.0-8.5 (~35% of firing demand, min 2.5 lbs, capped by 0.12 lb/gal solubility comfort).
		initialChargeLbs = round1(math.Max(2.5, math.Min(bakingSodaLbs*0.35, sumpGal*0.12)))
	}

	// Dissolved-solids TDS increase per firing [g/L]:
	//   HCl + NaHCO3 -> NaCl(aq) + H2O + CO2(g)
	// Plus 15% unreacted NaHCO3 buffer and soluble wood-ash potassium/carbonate salts.
	naclGrams := nHClRemovedMolS * 3600.0 * saltingHrs * naclMWGPerMol
	stoichNaHCO3Grams := nHClRemovedMolS * 3600.0 * saltingHrs * nahco3MWGPerMol
	excessNaHCO3Grams := math.Max(0.0, nahco3Kg*1000.0-stoichNaHCO3Grams)

	solubleAshGrams := 35.0 // minor combustion / soda dust baseline
	isWood := design.Input.FuelType == model.FuelWood || design.Input.FuelType == model.FuelWoodSalt
	if isWood {
		solubleAshGrams = 18.0 * firingHrs // ~18 g/hr soluble K2CO3/wood-ash minerals
	} else if p.VaporGlazeMode == VaporGlazeSoda {
		solubleAshGrams = 32.0 * firingHrs // airborne Na2CO3 vapor glaze carryover
	}

	endTDSGL := round1((naclGrams + excessNaHCO3Grams + solubleAshGrams) / math.Max(sumpLiters, 1.0))

	// Firings before full dump: governed by TDS ceiling (150 g/L / SG 1.105) AND ash-sludge hygiene limit.
	maxHygieneFirings := 10
	if isWood {
		maxHygieneFirings = 6
	} else if p.VaporGlazeMode == VaporGlazeSoda {
		maxHygieneFirings = 8
	}
	firingsBeforeDump := maxHygieneFirings
	if endTDSGL > 0.1 {
		tdsFirings := int(math.Floor(dumpTDSLimitGL / endTDSGL))
		if tdsFirings < firingsBeforeDump {
			firingsBeforeDump = tdsFirings
		}
	}
	if firingsBeforeDump < 1 {
		firingsBeforeDump = 1
	}

	// Water conservation comparison: once-through spray (quench + venturi without recirculation)
	// vs closed-loop recirculation (evaporative makeup + amortized sump fill/dump per firing).
	evapGPM := design.Quench.EvaporatedWaterGPM
	quenchSprayGPM := design.Quench.TotalSprayWaterGPM
	venturiGPM := design.Venturi.LiquidFlowGPM
	onceThroughGPM := quenchSprayGPM + venturiGPM
	if onceThroughGPM <= evapGPM*1.5 {
		onceThroughGPM = math.Max(quenchSprayGPM+venturiGPM, evapGPM*3.5)
	}
	onceThroughGal := round1(onceThroughGPM * 60.0 * firingHrs)
	evapMakeupGal := evapGPM * 60.0 * firingHrs
	amortizedDumpGal := sumpGal / float64(firingsBeforeDump)
	recircMakeupGal := round1(evapMakeupGal + amortizedDumpGal)

	waterSavedPct := 0.0
	if onceThroughGal > 0 {
		waterSavedPct = round1(math.Max(0.0, math.Min(98.0, (1.0-recircMakeupGal/onceThroughGal)*100.0)))
	}

	dumpTriggers := []string{
		fmt.Sprintf("Brine Hydrometer >= %.3f SG (~%.0f g/L dissolved NaCl/NaHCO3 — verifiable in 10 seconds with a $12 brine/salinity hydrometer)", dumpSpecificGravity, dumpTDSLimitGL),
		fmt.Sprintf("Calculated dissolved TDS >= %.0f g/L (well below 360 g/L NaCl saturation so salt never cakes inside the mag-drive pump or spray nozzles)", dumpTDSLimitGL),
		"Sump pH stays < 7.0 after stirring in scheduled baking soda top-up (alkalinity buffer exhausted)",
		"Settled wood-ash / silica sludge depth exceeds 4\" in the baffled settling compartment",
		"MANDATORY DRAIN RULE (IPC 702.5 & 40 CFR 403.5): ALWAYS cool sump liquor to <= 140 F (60 C) and verify pH 6.5–8.5 (strictly >= 5.0) before draining to sanitary sewer",
	}

	var steps []string
	if bakingSodaLbs > 0 {
		steps = []string{
			fmt.Sprintf("1. Pre-Firing Charge: Fill baffled sump to %.0f gal working level and dissolve %.1f lbs initial baking soda buffer (target pH 8.0–8.5).", sumpGal, initialChargeLbs),
			fmt.Sprintf("2. Peak Salting Window (%.1f hrs): Dose remaining ~%.1f lbs baking soda gradually over the %.1f-hr salting window (%.1f lbs total -> %d × 15 lb bulk bag; 1 × 50 lb sack covers ~%d firings).",
				saltingHrs, round1(math.Max(0.0, bakingSodaLbs-initialChargeLbs)), saltingHrs, bakingSodaLbs, bags15, maxInt(1, int(math.Floor(50.0/bakingSodaLbs)))),
			fmt.Sprintf("3. Evaporative Makeup: Potable water (via IPC 608 Air Gap / ASSE 1013 RPZ) automatically replaces ~%.0f gal evaporated steam over the %.0f-hr firing — saving %.1f%% water vs once-through (%.0f gal -> %.0f gal).",
				evapMakeupGal, firingHrs, waterSavedPct, onceThroughGal, recircMakeupGal),
			fmt.Sprintf("4. Between-Firing Reuse: Leave cooled neutral brine in sump for up to %d firing(s) (~%.1f g/L TDS added per firing); skim/strain ash basket and top up %.1f lbs baking soda before the next salt firing.",
				firingsBeforeDump, endTDSGL, bakingSodaLbs),
			fmt.Sprintf("5. Full Dump & Rinse (every %d firing(s) or SG >= %.3f): Cool <= 140 F (IPC 702.5), confirm pH 6.5–8.5, strain ash sludge, drain to sanitary sewer, and run 5-min fresh-water flush.",
				firingsBeforeDump, dumpSpecificGravity),
		}
	} else {
		steps = []string{
			fmt.Sprintf("1. Pre-Firing Fill (%s mode, 0 ppm HCl): Fill baffled sump to %.0f gal with clean tap water (optional 1 cup baking soda keeps pH ~8.0).", p.VaporGlazeMode, sumpGal),
			fmt.Sprintf("2. Closed-Loop Makeup: Replaces only ~%.0f gal evaporated quench water over the %.0f-hr firing — saving %.1f%% water vs once-through (%.0f gal -> %.0f gal).",
				evapMakeupGal, firingHrs, waterSavedPct, onceThroughGal, recircMakeupGal),
			fmt.Sprintf("3. Multi-Firing Reuse: Reuse settled water for up to %d firings (~%.1f g/L dissolved ash/soda per firing), cleaning the 50-mesh pump strainer between firings.",
				firingsBeforeDump, endTDSGL),
			"4. Full Dump & Rinse: Cool <= 140 F (IPC 702.5), strain settled ash solids, confirm pH 6.5–8.5 (40 CFR 403.5), drain cleanly, and refill.",
		}
	}

	var summary string
	if bakingSodaLbs > 0 {
		summary = fmt.Sprintf(
			"Closed-loop %.0f-gal baffled sump saves %.1f%% water vs once-through (%.0f gal makeup vs %.0f gal once-through). Neutralizing %.2f lbs HCl over a %.1f-hr salting window uses %.1f lbs bulk baking soda (%d × 15 lb bag; 1 × 50 lb sack covers ~%d firings) and adds ~%.1f g/L TDS — reuse sump for up to %d firing(s) before full dump at SG >= %.3f (150 g/L).",
			sumpGal, waterSavedPct, recircMakeupGal, onceThroughGal, hclNeutralizedLbs, saltingHrs,
			bakingSodaLbs, bags15, maxInt(1, int(math.Floor(50.0/bakingSodaLbs))), endTDSGL, firingsBeforeDump, dumpSpecificGravity,
		)
	} else {
		summary = fmt.Sprintf(
			"Closed-loop %.0f-gal baffled sump saves %.1f%% water vs once-through (%.0f gal makeup vs %.0f gal once-through) with 0 lbs baking soda required (%s mode). Reuse settled sump water up to %d firings (~%.1f g/L dissolved solids/firing) before cooling <= 140 F and dumping.",
			sumpGal, waterSavedPct, recircMakeupGal, onceThroughGal, p.VaporGlazeMode, firingsBeforeDump, endTDSGL,
		)
	}

	return RecirculationPlan{
		SumpVolumeGal:              sumpGal,
		OnceThroughWaterGal:        onceThroughGal,
		RecircMakeupWaterGal:       recircMakeupGal,
		WaterSavedPct:              waterSavedPct,
		SaltingHours:               saltingHrs,
		HClNeutralizedLbs:          hclNeutralizedLbs,
		BakingSodaLbsPerFiring:     bakingSodaLbs,
		BakingSoda15LbBags:         bags15,
		BakingSoda50LbSacks:        sacks50,
		InitialChargeLbs:           initialChargeLbs,
		EndOfFiringTDSGL:           endTDSGL,
		DumpTDSLimitGL:             dumpTDSLimitGL,
		DumpSpecificGravity:        dumpSpecificGravity,
		FiringsBeforeFullDump:      firingsBeforeDump,
		TopUpSodaLbsBetweenFirings: bakingSodaLbs,
		DumpTriggers:               dumpTriggers,
		ScheduleSteps:              steps,
		Summary:                    summary,
	}
}
