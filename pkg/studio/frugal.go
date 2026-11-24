package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
	"github.com/AaronFilson/kilnscrub/pkg/thermo"
)

// FrugalAssessment quantifies the Phase 10 Local Ceramic / Masonry (Nairobi / Mumbai)
// and 100% Zero-Electricity passive scrubber architecture from first principles:
//   - Natural chimney buoyancy draft: dP = g * H * (rho_amb - rho_stack) / 249.08891
//   - Wide stoneware-packed masonry column irrigated Ergun pressure drop
//   - Elevated overhead gravity tank head & manual treadle / hand-pump labor budget
//   - Exact local alkali stoichiometry (Magadi soda ash, chuna slaked lime, wood-ash lye, baking soda)
//   - Honest HCl vs coarse ash vs submicron NaCl capture performance and >1800 F ceramic immunity
type FrugalAssessment struct {
	Enabled                 bool     `json:"enabled"`
	ZeroElectricity         bool     `json:"zero_electricity"`
	Region                  string   `json:"region"`
	RegionLabel             string   `json:"region_label"`
	ChimneyHeightFt         float64  `json:"chimney_height_ft"`
	ChimneyHeightM          float64  `json:"chimney_height_m"`
	StackReheatTempF        float64  `json:"stack_reheat_temp_f"`
	AvailableDraftInWC      float64  `json:"available_draft_inwc"`
	RequiredDraftInWC       float64  `json:"required_draft_inwc"`
	DraftMarginInWC         float64  `json:"draft_margin_inwc"`
	DraftFeasible           bool     `json:"draft_feasible"`
	DraftMechanism          string   `json:"draft_mechanism"`
	BedDiameterIn           float64  `json:"bed_diameter_in"`
	BedSuperficialVelMS     float64  `json:"bed_superficial_vel_m_s"`
	BedDPInWC               float64  `json:"bed_dp_inwc"`
	PackingMaterial         string   `json:"packing_material"`
	GravityTankGal          float64  `json:"gravity_tank_gal"`
	GravityTankLiters       float64  `json:"gravity_tank_liters"`
	TankElevationFt         float64  `json:"tank_elevation_ft"`
	GravityHeadPSI          float64  `json:"gravity_head_psi"`
	SprayFlowGPM            float64  `json:"spray_flow_gpm"`
	SprayFlowLPM            float64  `json:"spray_flow_lpm"`
	ManualPumpType          string   `json:"manual_pump_type"`
	Jerrycans20LPerHr       int      `json:"jerrycans_20l_per_hr"`
	TreadleMinutesPerHr     float64  `json:"treadle_minutes_per_hr"`
	AlkaliID                string   `json:"alkali_id"`
	AlkaliName              string   `json:"alkali_name"`
	AlkaliFormula           string   `json:"alkali_formula"`
	AlkaliKgPerFiring       float64  `json:"alkali_kg_per_firing"`
	AlkaliLbsPerFiring      float64  `json:"alkali_lbs_per_firing"`
	AlkaliCostUSD           float64  `json:"alkali_cost_usd"`
	AlkaliLocalNote         string   `json:"alkali_local_note"`
	ScalingWarning          string   `json:"scaling_warning"`
	HClRemovalPct           float64  `json:"hcl_removal_pct"`
	CoarseAshCapturePct     float64  `json:"coarse_ash_capture_pct"`
	SubmicronNaClCapturePct float64  `json:"submicron_nacl_capture_pct"`
	SubmicronHonestNote     string   `json:"submicron_honest_note"`
	ThermalImmunityNote     string   `json:"thermal_immunity_note"`
	EstimatedLocalCostUSD   float64  `json:"estimated_local_cost_usd"`
	LocalCurrencyEstimate   string   `json:"local_currency_estimate"`
	PowerWatts              float64  `json:"power_watts"`
	ManualSteps             []string `json:"manual_steps"`
	LocalMaterials          []string `json:"local_materials"`
	Warnings                []string `json:"warnings"`
	Summary                 string   `json:"summary"`
}

func round3(v float64) float64 {
	return math.Round(v*1000.0) / 1000.0
}

type localRegionProfile struct {
	id              string
	label           string
	currencyCode    string
	fxPerUSD        float64
	ambientPa       float64 // barometric pressure [Pa] accounting for elevation (e.g., Nairobi 1,795 m)
	ambientTempK    float64 // representative daytime ambient temperature [K]
	elevationNote   string
	tankBrand       string
	pumpBrand       string
	treadleLabel    string
	pipeLocalSpec   string
	costScale       float64
	ringCostPerCuFt float64
	sodaAshUSDPerKg float64
	limeUSDPerKg    float64
	bakingUSDPerKg  float64
}

func resolveLocalRegionProfile(region string) localRegionProfile {
	switch region {
	case LocalRegionMumbai:
		return localRegionProfile{
			id:              LocalRegionMumbai,
			label:           "Mumbai / Western India (INR)",
			currencyCode:    "INR",
			fxPerUSD:        84.0,
			ambientPa:       101325.0,
			ambientTempK:    303.15, // 30 C / 86 F coastal ambient
			elevationNote:   "sea level (101.3 kPa, 30 C / 86 F)",
			tankBrand:       "Sintex HDPE Overhead Tank",
			pumpBrand:       "Kirloskar / Crompton",
			treadleLabel:    "Cast-Iron Semi-Rotary Hand Pump / Treadle Pump",
			pipeLocalSpec:   "IS 651 Salt-Glazed Stoneware (SWG) Clay Pipes & Firebrick Masonry",
			costScale:       0.84,
			ringCostPerCuFt: 4.5,
			sodaAshUSDPerKg: 0.58,
			limeUSDPerKg:    0.20,
			bakingUSDPerKg:  0.85,
		}
	case LocalRegionGeneral:
		return localRegionProfile{
			id:              LocalRegionGeneral,
			label:           "General Local Pottery / Rural Studio (USD)",
			currencyCode:    "USD",
			fxPerUSD:        1.0,
			ambientPa:       101325.0,
			ambientTempK:    293.15, // 20 C / 68 F mild ambient
			elevationNote:   "sea level (101.3 kPa, 20 C / 68 F)",
			tankBrand:       "HDPE Gravity Drum / Overhead Water Tank",
			pumpBrand:       "Kirloskar / Davis & Shirtliff / Local 0.5 HP",
			treadleLabel:    "Treadle / Semi-Rotary Cast-Iron Hand Pump",
			pipeLocalSpec:   "Local Firebrick Masonry & Thrown/Glazed Stoneware Flue Tiles",
			costScale:       1.12,
			ringCostPerCuFt: 8.0,
			sodaAshUSDPerKg: 0.80,
			limeUSDPerKg:    0.35,
			bakingUSDPerKg:  1.15,
		}
	default: // LocalRegionNairobi
		return localRegionProfile{
			id:              LocalRegionNairobi,
			label:           "Nairobi / East Africa (KES)",
			currencyCode:    "KES",
			fxPerUSD:        130.0,
			ambientPa:       83500.0, // 1,660 m highland elevation (~0.824 atm)
			ambientTempK:    295.15,  // 22 C / 72 F highland ambient
			elevationNote:   "1,660 m highland elevation (83.5 kPa, 22 C / 72 F)",
			tankBrand:       "Kentank / Roto HDPE Overhead Tank",
			pumpBrand:       "Davis & Shirtliff",
			treadleLabel:    "MoneyMaker Treadle / Semi-Rotary Cast-Iron Hand Pump",
			pipeLocalSpec:   "Local Firebrick Masonry & Thrown/Salt-Glazed Stoneware Flue Tiles",
			costScale:       1.0,
			ringCostPerCuFt: 5.5,
			sodaAshUSDPerKg: 0.55,
			limeUSDPerKg:    0.25,
			bakingUSDPerKg:  1.10,
		}
	}
}

func formatLocalCurrency(usd float64, prof localRegionProfile) string {
	if prof.currencyCode == "USD" {
		return fmt.Sprintf("~$%.0f USD", math.Round(usd))
	}
	localAmt := math.Round((usd*prof.fxPerUSD)/100.0) * 100.0
	return fmt.Sprintf("~%s %s (~$%.0f USD @ %.0f %s/USD)", prof.currencyCode, formatIntCommas(int(localAmt)), math.Round(usd), prof.fxPerUSD, prof.currencyCode)
}

func formatIntCommas(n int) string {
	if n < 0 {
		return "-" + formatIntCommas(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	rem := len(s) % 3
	if rem == 0 {
		rem = 3
	}
	out := s[:rem]
	for i := rem; i < len(s); i += 3 {
		out += "," + s[i:i+3]
	}
	return out
}

type localAlkaliResult struct {
	id             string
	name           string
	formula        string
	kgPerFiring    float64
	lbsPerFiring   float64
	costUSD        float64
	localNote      string
	scalingWarning string
}

func computeLocalAlkali(p StudioProject, design model.ScrubberDesign, prof localRegionProfile) localAlkaliResult {
	alkID := p.EffectiveLocalAlkali()
	saltingHrs := p.EffectiveSaltingHours()
	saltKg := 0.0
	if p.VaporGlazeMode == VaporGlazeSalt && saltingHrs > 0 && design.Input.SaltRateLbsPerHr > 0 {
		saltKg = design.Input.SaltRateLbsPerHr * saltingHrs * model.KgPerLb
	}

	const (
		mwNaClKgPerMol   = 0.0584428 // NaCl -> 1 mol HCl per mol NaCl
		mwNa2CO3KgPerMol = 0.1059888 // 2 HCl + Na2CO3 -> 2 NaCl + H2O + CO2 (0.5 mol/mol HCl)
		mwCaOH2KgPerMol  = 0.074093  // 2 HCl + Ca(OH)2 -> CaCl2 + 2 H2O (0.5 mol/mol HCl)
		mwK2CO3KgPerMol  = 0.138205  // 2 HCl + K2CO3 -> 2 KCl + H2O + CO2 (0.5 mol/mol HCl)
		mwNaHCO3KgPerMol = 0.0840066 // HCl + NaHCO3 -> NaCl + H2O + CO2 (1.0 mol/mol HCl)
		woodAshK2CO3Frac = 0.07      // ~7% extractable K2CO3 by mass from dry hardwood kiln ash
	)

	molHCl := saltKg / mwNaClKgPerMol

	switch alkID {
	case LocalAlkaliSlakedLime:
		kg := round2(molHCl * 0.5 * mwCaOH2KgPerMol * 1.30)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.limeUSDPerKg)
		note := fmt.Sprintf("Builder's slaked lime / Chuna (Ca(OH)2, 0.5 mol/mol HCl + 30%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing. Lowest chemical cost in Mumbai & Nairobi hardware shops.", kg, lbs, saltingHrs)
		if prof.id == LocalRegionMumbai {
			note = fmt.Sprintf("Local Mumbai Chuna / hydrated lime (Ca(OH)2, 0.5 mol/mol HCl + 30%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing (~INR %.0f/firing from any building-supply shop).", kg, lbs, saltingHrs, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionNairobi {
			note = fmt.Sprintf("Nairobi builder's slaked lime / Chokaa (Ca(OH)2, 0.5 mol/mol HCl + 30%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing (~KES %.0f/firing from local hardware yards).", kg, lbs, saltingHrs, cost*prof.fxPerUSD)
		}
		if molHCl <= 0 {
			note = fmt.Sprintf("Zero HCl mode (%s): 0 kg slaked lime needed (optional 0.25 kg buffer keeps sump alkaline).", p.VaporGlazeMode)
		}
		return localAlkaliResult{
			id:             LocalAlkaliSlakedLime,
			name:           "Slaked Lime / Chuna",
			formula:        "Ca(OH)2",
			kgPerFiring:    kg,
			lbsPerFiring:   lbs,
			costUSD:        cost,
			localNote:      note,
			scalingWarning: "CaCO3 SCALING & SLURRY WARNING: Slaked lime has low water solubility (~1.7 g/L) — stir the sump continuously as a milk-of-lime slurry so lime does not settle dead on the bottom, and expect chalky CaCO3 scale from flue CO2 on the spray rose and upper stoneware rings (brush/vinegar-flush every 3-5 firings).",
		}

	case LocalAlkaliWoodAshLye:
		pureK2CO3Kg := molHCl * 0.5 * mwK2CO3KgPerMol * 1.25
		dryAshKg := round1(pureK2CO3Kg / woodAshK2CO3Frac)
		dryAshLbs := round1(dryAshKg / model.KgPerLb)
		note := fmt.Sprintf("Potter's own wood-kiln ash leached in a 200L barrel (~7%% extractable K2CO3 -> %.1f kg / %.1f lbs dry hardwood ash yields %.2f kg K2CO3 for %.1f-hr salting at $0.00 chemical cost!).", dryAshKg, dryAshLbs, pureK2CO3Kg, saltingHrs)
		if molHCl <= 0 {
			note = fmt.Sprintf("Zero HCl mode (%s): 0 kg wood ash required (potter's own free wood-kiln ash provides mild pH 8-9 rinse water if desired).", p.VaporGlazeMode)
		}
		return localAlkaliResult{
			id:             LocalAlkaliWoodAshLye,
			name:           "Wood-Ash Potash Lye (Dry Ash to Leach)",
			formula:        "K2CO3 (from ~7% wood ash)",
			kgPerFiring:    dryAshKg,
			lbsPerFiring:   dryAshLbs,
			costUSD:        0.0,
			localNote:      note,
			scalingWarning: "ASH FILTRATION & CAUSTIC CARE: Steep dry hardwood ash in a 200L barrel for 24 hrs and decant clear amber potash lye through woven cloth before pouring into the gravity tank so charcoal grit and insoluble CaCO3 ash never plug the spray rose. Fresh wood-ash lye is caustic (pH 11-13) — wear rubber gloves and eye protection.",
		}

	case LocalAlkaliBakingSoda:
		kg := round2(molHCl * 1.0 * mwNaHCO3KgPerMol * 1.25)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.bakingUSDPerKg)
		note := fmt.Sprintf("Bulk sodium bicarbonate (NaHCO3, 1.0 mol/mol HCl + 25%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing. Gentlest handling (self-buffers near pH 8.3) with zero scaling.", kg, lbs, saltingHrs)
		if molHCl <= 0 {
			note = fmt.Sprintf("Zero HCl mode (%s): 0 kg baking soda needed.", p.VaporGlazeMode)
		}
		return localAlkaliResult{
			id:             LocalAlkaliBakingSoda,
			name:           "Bulk Baking Soda",
			formula:        "NaHCO3",
			kgPerFiring:    kg,
			lbsPerFiring:   lbs,
			costUSD:        cost,
			localNote:      note,
			scalingWarning: "",
		}

	default: // LocalAlkaliSodaAsh
		kg := round2(molHCl * 0.5 * mwNa2CO3KgPerMol * 1.25)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.sodaAshUSDPerKg)
		note := fmt.Sprintf("Magadi Soda Ash / bulk washing soda (Na2CO3, 0.5 mol/mol HCl + 25%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing — 37%% lighter than baking soda, highly soluble, zero CaCO3 scaling.", kg, lbs, saltingHrs)
		if prof.id == LocalRegionNairobi {
			note = fmt.Sprintf("Lake Magadi Soda Ash (Na2CO3, 0.5 mol/mol HCl + 25%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing (~KES %.0f/firing — mined in Kenya, abundant in Nairobi industrial/soap supply depots, zero CaCO3 scaling).", kg, lbs, saltingHrs, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionMumbai {
			note = fmt.Sprintf("Bulk Mumbai Washing Soda / Soda Ash (Na2CO3, 0.5 mol/mol HCl + 25%% margin): %.2f kg (%.1f lbs) per %.1f-hr salting firing (~INR %.0f/firing from Crawford Market / chemical depots, highly soluble, zero scaling).", kg, lbs, saltingHrs, cost*prof.fxPerUSD)
		}
		if molHCl <= 0 {
			note = fmt.Sprintf("Zero HCl mode (%s): 0 kg soda ash needed (optional 0.25 kg keeps sump mildly buffered at pH 8.0-8.5).", p.VaporGlazeMode)
		}
		return localAlkaliResult{
			id:             LocalAlkaliSodaAsh,
			name:           "Magadi Soda Ash / Washing Soda",
			formula:        "Na2CO3",
			kgPerFiring:    kg,
			lbsPerFiring:   lbs,
			costUSD:        cost,
			localNote:      note,
			scalingWarning: "",
		}
	}
}

type localCeramicBOMBreakdown struct {
	quenchChamberUSD float64
	baffleZoneUSD    float64
	masonryTowerUSD  float64
	stonewareRingUSD float64
	ringCuFt         float64
	terracottaDemUSD float64
	brickSumpUSD     float64
	gravityTankUSD   float64
	standBracingUSD  float64
	pumpUSD          float64
	fanUSD           float64
	plumbingUSD      float64
	chimneyUSD       float64
	instrumentsUSD   float64
	alkaliUSD        float64
	totalUSD         float64
}

func computeLocalCeramicCosts(
	p StudioProject,
	design model.ScrubberDesign,
	prof localRegionProfile,
	bedDiaIn float64,
	bedDepthM float64,
	chimneyM float64,
	tankLiters float64,
	alkali localAlkaliResult,
) localCeramicBOMBreakdown {
	volScale := math.Max(0.75, math.Min(1.50, math.Pow(design.Input.KilnVolumeCuFt/24.0, 0.35)))
	s := prof.costScale

	bedDiaM := float64(model.InchesToMeters(bedDiaIn))
	bedVolM3 := math.Pi * bedDiaM * bedDiaM / 4.0 * bedDepthM
	ringCuFt := math.Max(4.0, math.Ceil(model.CubicMeters(bedVolM3).CuFt()))

	quenchChamber := round2(110.0 * volScale * s)
	baffleZone := round2(42.0 * volScale * s)
	masonryTower := round2((70.0 + 2.2*bedDiaIn) * s)
	stonewareRings := round2(ringCuFt * prof.ringCostPerCuFt)
	terracottaDem := round2(28.0 * volScale * s)
	brickSump := round2(48.0 * volScale * s)
	gravityTank := round2((34.0 + 0.08*tankLiters) * s)
	standBracing := round2((40.0 + 0.05*tankLiters) * s)
	plumbing := round2(40.0 * s)

	var pumpUSD, fanUSD, chimneyUSD float64
	if p.ZeroElectricity {
		pumpUSD = round2(75.0 * s) // MoneyMaker treadle / cast-iron semi-rotary hand pump
		fanUSD = 0.0
		chimneyUSD = round2((25.0 + 9.2*chimneyM) * s) // 8-12 m brick/SWG flue chimney + guy wires
	} else {
		pumpUSD = round2(76.0 * s)                     // Local 0.5 HP 230V 50Hz domestic monoblock pump
		fanUSD = round2(65.0 * s)                      // Local 12V solar / 230V jua-kali low-dP draft fan
		chimneyUSD = round2((20.0 + 9.2*chimneyM) * s) // Shorter 4.5 m ceramic/SWG chimney
	}

	instruments := round2(22.0 * s) // $10 brine hydrometer + litmus pH strips + stem thermometer
	alkaliCost := round2(alkali.costUSD)

	total := round2(
		quenchChamber + baffleZone + masonryTower + stonewareRings + terracottaDem +
			brickSump + gravityTank + standBracing + pumpUSD + fanUSD + plumbing +
			chimneyUSD + instruments + alkaliCost,
	)

	return localCeramicBOMBreakdown{
		quenchChamberUSD: quenchChamber,
		baffleZoneUSD:    baffleZone,
		masonryTowerUSD:  masonryTower,
		stonewareRingUSD: stonewareRings,
		ringCuFt:         ringCuFt,
		terracottaDemUSD: terracottaDem,
		brickSumpUSD:     brickSump,
		gravityTankUSD:   gravityTank,
		standBracingUSD:  standBracing,
		pumpUSD:          pumpUSD,
		fanUSD:           fanUSD,
		plumbingUSD:      plumbing,
		chimneyUSD:       chimneyUSD,
		instrumentsUSD:   instruments,
		alkaliUSD:        alkaliCost,
		totalUSD:         total,
	}
}

// EvaluateFrugal computes the first-principles Local Ceramic / Masonry (Nairobi / Mumbai)
// and Zero-Electricity passive scrubber physics assessment for any StudioProject.
// Always returned non-nil on StudioEvaluation with Enabled = project.IsLocalCeramic()
// so both off-grid builds and standard powered builds can preview local/off-grid metrics.
func EvaluateFrugal(project StudioProject, design model.ScrubberDesign) *FrugalAssessment {
	p := project.Normalize()
	enabled := p.IsLocalCeramic()
	zeroElec := p.ZeroElectricity
	prof := resolveLocalRegionProfile(p.EffectiveLocalRegion())

	// 1. Wide brick/terracotta stoneware-packed column sized for low superficial velocity (v_s ~ 0.35-0.45 m/s).
	qGas := float64(design.QuenchedGas.VolumetricFlow)
	rhoG := float64(design.QuenchedGas.Density)
	muG := float64(design.QuenchedGas.Viscosity)
	if qGas <= 0 {
		qGas = 0.11
	}
	if rhoG <= 0 {
		rhoG = 0.87
	}
	if muG <= 0 {
		muG = 1.85e-5
	}

	// Scale gas volume to local ambient pressure (e.g., Nairobi 81.5 kPa vs sea-level 101.3 kPa)
	// while conserving molar flow so local column velocity and buoyancy density are consistent.
	pDesign := float64(design.QuenchedGas.Pressure)
	if pDesign <= 0 {
		pDesign = float64(model.StandardPressurePa)
	}
	pAmb := prof.ambientPa
	qLocal := qGas * (pDesign / pAmb)
	rhoLocal := rhoG * (pAmb / pDesign)

	const targetVs = 0.40 // m/s sweet spot in [0.35, 0.45] m/s for ultra-low Ergun dP
	reqArea := qLocal / targetVs
	rawDiaIn := model.Meters(math.Sqrt(4.0 * reqArea / math.Pi)).Inches()
	// Snap to practical 2" (50 mm) masonry/SWG column increments, min 14" so small kilns stay practical.
	bedDiaIn := math.Max(14.0, math.Ceil(rawDiaIn/2.0)*2.0)
	bedDiaM := float64(model.InchesToMeters(bedDiaIn))
	bedAreaM2 := math.Pi * bedDiaM * bedDiaM / 4.0
	vSup := qLocal / bedAreaM2

	// Bed depth (1.1 - 1.4 m) scaled smoothly with kiln volume for >=95.5% HCl capture at low velocity.
	bedDepthM := 1.20
	if design.Input.KilnVolumeCuFt <= 16.0 {
		bedDepthM = 1.10
	} else if design.Input.KilnVolumeCuFt >= 40.0 {
		bedDepthM = 1.35
	}

	// Real irrigated Ergun pressure drop for 1" (0.0254 m) potter-extruded unglazed stoneware
	// Raschig rings (voidage eps = 0.70, Eckert irrigated liquid-holdup factor = 1.70).
	// Using nominal hollow-ring diameter dp = 0.0254 m matches the empirical Eckert GPDC chart
	// for 1" ceramic Raschig rings at low gas velocity (~0.05-0.08 inWC over 1.2 m).
	const (
		ringNomM        = 0.0254
		ringVoidage     = 0.70
		irrigatedFactor = 1.70
	)
	bedDPInWC := 0.065
	if dryDP, err := scrubber.ErgunPressureDrop(ringNomM, ringVoidage, vSup, rhoLocal, muG, bedDepthM); err == nil {
		bedDPInWC = round3(model.Pascal(float64(dryDP) * irrigatedFactor).InWC())
	}

	// 2. Total passive draft budget: |target_kiln_draft| + bed_dp + quench/baffle (0.03) + terracotta demister (0.02).
	kilnDraftReq := math.Abs(design.Input.TargetKilnDraftInWC)
	if kilnDraftReq <= 0 {
		kilnDraftReq = 0.05
	}
	const (
		quenchBaffleDPInWC       = 0.030
		terracottaDemisterDPInWC = 0.020
	)
	reqDraftInWC := round3(kilnDraftReq + bedDPInWC + quenchBaffleDPInWC + terracottaDemisterDPInWC)

	// 3. Natural chimney buoyancy draft (when zero_electricity=true) OR local 12V/230V fan hybrid (when false).
	// Ideal gas density: rho = P * M / (R * T), buoyancy dP = g * H * (rho_amb - rho_stack) / 249.08891.
	const (
		mwDryAir       = 0.028964 // kg/mol
		gAccel         = model.StandardGravity
		paPerInWC      = model.PaPerInWC
		gravityAssist  = 0.025 // modest downward quench-spray entrainment assist from 3.5 m gravity head
		targetMarginWC = 0.040
	)
	mwStack := thermo.MixtureMolecularWeight(design.OutletGas)
	if mwStack <= 0.015 || mwStack > 0.045 {
		mwStack = 0.0268 // wet scrubbed wood/salt flue gas (~26-28% H2O vapor lowers MW, aiding buoyancy)
	}
	rhoAmb := (pAmb * mwDryAir) / (model.UniversalGasConst * prof.ambientTempK)

	var chimneyM, stackReheatF, availDraftInWC, powerWatts float64
	var draftMechanism string

	if zeroElec {
		// Auto-size brick/terracotta thermal-reheat chimney in [8.0, 12.0] m at 340 F (444.26 K)
		// recovered from the 2400 F kiln exit flue via concentric brick flue-wall recuperator.
		// If high altitude or high target draft still needs extra margin at 12.0 m, step reheat up to 390 F.
		stackReheatF = 340.0
		chimneyM = 8.0
		calcBuoyancy := func(hM, tStackF float64) float64 {
			tStackK := float64(model.FahrenheitToKelvin(tStackF))
			rhoStack := (pAmb * mwStack) / (model.UniversalGasConst * tStackK)
			buoyInWC := (gAccel * hM * (rhoAmb - rhoStack)) / paPerInWC
			return buoyInWC + gravityAssist
		}
		for h := 8.0; h <= 12.0+1e-6; h += 0.5 {
			chimneyM = round1(h)
			if calcBuoyancy(chimneyM, stackReheatF)-reqDraftInWC >= targetMarginWC {
				break
			}
		}
		for tF := 340.0; tF <= 400.0 && calcBuoyancy(chimneyM, tF)-reqDraftInWC < targetMarginWC; tF += 10.0 {
			stackReheatF = tF
		}
		availDraftInWC = round3(calcBuoyancy(chimneyM, stackReheatF))
		powerWatts = 0.0
		draftMechanism = fmt.Sprintf(
			"Natural Chimney Buoyancy + Concentric Brick Flue-Wall Thermal Reheat (%.0f F stack, %.1f m / %.1f ft brick chimney at %s) + %.3f inWC gravity-spray downflow assist — 0 W electric.",
			stackReheatF, chimneyM, chimneyM*3.28084, prof.elevationNote, gravityAssist,
		)
	} else {
		// Local-grid / 12V solar hybrid: shorter 4.5 m ceramic/SWG chimney + local low-dP fan + 0.5 HP 230V 50Hz pump.
		chimneyM = 4.5
		stackReheatF = round1(design.OutletGas.Temperature.Fahrenheit())
		if stackReheatF <= 0 {
			stackReheatF = 184.0
		}
		tStackK := float64(model.FahrenheitToKelvin(stackReheatF))
		rhoStack := (pAmb * mwStack) / (model.UniversalGasConst * tStackK)
		natBuoyInWC := (gAccel * chimneyM * (rhoAmb - rhoStack)) / paPerInWC
		const localFanStaticInWC = 1.80 // local 12V solar / 230V jua-kali low-dP centrifugal/radiator fan
		availDraftInWC = round3(localFanStaticInWC + natBuoyInWC)
		powerWatts = 520.0 // 370 W (0.5 HP 230V 50Hz domestic monoblock pump) + 150 W local low-dP fan
		draftMechanism = fmt.Sprintf(
			"Local Low-dP 12V Solar / 230V Jua-Kali Draft Fan (1.80 inWC) + %.1f m (%.1f ft) Ceramic Chimney + 0.5 HP 230V 50Hz Domestic Monoblock Pump (~%.0f W total).",
			chimneyM, chimneyM*3.28084, powerWatts,
		)
	}

	chimneyFt := round1(chimneyM * 3.28084)
	draftMarginInWC := round3(availDraftInWC - reqDraftInWC)
	draftFeasible := draftMarginInWC >= 0.0

	// 4. Gravity overhead tank (Kentank/Sintex 210-490 L on 3.5 m / 11.5 ft braced stand) & manual pumping budget.
	tankLiters := 300.0
	if design.Input.KilnVolumeCuFt <= 16.0 {
		tankLiters = 210.0
	} else if design.Input.KilnVolumeCuFt >= 40.0 {
		tankLiters = 490.0
	}
	tankGal := round1(tankLiters / 3.785411784)
	const tankElevFt = 11.5 // 3.5 m braced brick/timber stand
	gravityHeadPSI := round2(tankElevFt * 0.4335)

	// Closed-loop gravity quench + stoneware bed trickle (~2.3x evaporative demand -> ~4-6 L/min on 24 cu.ft. studio kiln).
	evapGPM := design.Quench.EvaporatedWaterGPM
	if evapGPM <= 0 {
		evapGPM = design.Quench.TotalSprayWaterGPM / 3.0
	}
	sprayFlowGPM := round2(math.Max(1.05, math.Min(2.20, evapGPM*1.55+0.45)))
	sprayFlowLPM := round2(sprayFlowGPM * 3.78541)

	var manualPumpType string
	var jerrycansPerHr int
	var treadleMinPerHr float64
	if zeroElec {
		manualPumpType = "Treadle / Semi-Rotary Cast-Iron Hand Pump"
		jerrycansPerHr = int(math.Ceil(sprayFlowLPM * 60.0 / 20.0))
		treadleMinPerHr = round1(sprayFlowLPM * 60.0 / 35.0) // ~35 L/min MoneyMaker treadle / wing pump
	} else {
		manualPumpType = fmt.Sprintf("Local 0.5 HP 230V 50Hz Monoblock Pump (%s)", prof.pumpBrand)
		jerrycansPerHr = 0
		treadleMinPerHr = 0.0
	}

	// 5. Local alkali stoichiometry.
	alkali := computeLocalAlkali(p, design, prof)

	// 6. Honest capture performance & zero-meltdown thermal immunity.
	// Wide irrigated stoneware bed has long gas residence time (~3 s) -> high HCl absorption (95.5-98.0%).
	htuM := 0.34 * math.Pow(math.Max(vSup, 0.20)/0.38, 0.22)
	ntu := bedDepthM / htuM
	hclPct := round1(math.Max(95.5, math.Min(98.0, (1.0-math.Exp(-ntu))*100.0)))
	coarseAshPct := 90.5
	if bedDepthM >= 1.30 {
		coarseAshPct = 92.5
	} else if bedDepthM <= 1.12 {
		coarseAshPct = 88.5
	}
	submicronPct := 30.0
	if !zeroElec {
		submicronPct = 33.5 // modest improvement with higher local-pump spray pressure
	}
	submicronNote := fmt.Sprintf(
		"HONEST PHYSICS CHECK: Passive zero-electricity draft (~%.2f inWC) cannot power a high-energy 19 inWC / 70 m/s venturi throat (needed for ~84%% submicron NaCl capture). Expect ~%.0f%% capture of 0.1-0.5 um submicron NaCl fume (via condensation growth, diffusiophoresis & wet stoneware impaction) while achieving %.1f%% HCl acid neutralization and %.1f%% coarse ash capture — rely on the tall %.1f m (%.0f ft) thermal-reheat chimney to disperse residual fine salt haze well above roofs.",
		availDraftInWC, submicronPct, hclPct, coarseAshPct, chimneyM, chimneyFt,
	)
	if !zeroElec {
		submicronNote = fmt.Sprintf(
			"HONEST PHYSICS CHECK: A low-dP local-grid/12V ceramic scrubber (~1.8 inWC fan) achieves %.1f%% HCl acid gas capture and %.1f%% coarse wood ash capture, but only ~%.0f%% submicron (0.1-0.5 um) NaCl fume capture (vs ~84%% in a 19 inWC industrial venturi). Discharge above roof level via the ceramic chimney.",
			hclPct, coarseAshPct, submicronPct,
		)
	}
	thermalImmunityNote := "ZERO-MELTDOWN CERAMIC IMMUNITY: Fired firebrick, IS 651 salt-glazed stoneware (SWG) pipes, potter-extruded unglazed stoneware Raschig rings, and hand-formed terracotta chevron baffles tolerate >1,800 F (980 C) dry-run exhaust indefinitely without melting — completely eliminating the 200 F / 210 F CPVC plastic meltdown risk and electric fail-safe bypass actuators."

	// 7. Local Materials BOM cost estimate & local currency conversion.
	costs := computeLocalCeramicCosts(p, design, prof, bedDiaIn, bedDepthM, chimneyM, tankLiters, alkali)
	localCurrency := formatLocalCurrency(costs.totalUSD, prof)

	packingMaterial := fmt.Sprintf(`1" Potter-Extruded Unglazed Stoneware Raschig Rings (%.0f cu.ft., voidage 70%%, %.2f m deep, fired Cone 6-10 in studio kiln)`, costs.ringCuFt, bedDepthM)

	localMaterials := []string{
		fmt.Sprintf("Stage 1 & Recuperator: %s (~$%.0f) — concentric brick flue wall warms scrubbed exhaust back to %.0f F without mixing raw flue gas", prof.pipeLocalSpec, costs.quenchChamberUSD, stackReheatF),
		fmt.Sprintf("Stage 2 Settling & Quench: Hand-formed terracotta baffle bricks + brass/PVC coarse gravity spray rose (~$%.0f)", costs.baffleZoneUSD),
		fmt.Sprintf("Stage 3 Absorber Tower: %.0f\" (%.0f mm) wide brick / IS 651 SWG ceramic column (~$%.0f) + %.0f cu.ft. potter-extruded 1\" stoneware Raschig rings (~$%.0f) + stacked terracotta chevron mist baffles (~$%.0f)", bedDiaIn, bedDiaIn*25.4, costs.masonryTowerUSD, costs.ringCuFt, costs.stonewareRingUSD, costs.terracottaDemUSD),
		fmt.Sprintf("Overhead Gravity Tank & Stand: %.0f L (%.0f gal) %s (~$%.0f) on braced 3.5 m (11.5 ft) brick/timber stand (~$%.0f) + brick baffled sump (~$%.0f)", tankLiters, tankGal, prof.tankBrand, costs.gravityTankUSD, costs.standBracingUSD, costs.brickSumpUSD),
		fmt.Sprintf("Circulation & Plumbing: %s (~$%.0f) + PVC/HDPE gravity downcomer, foot valve & drain (~$%.0f)", manualPumpType, costs.pumpUSD, costs.plumbingUSD),
		fmt.Sprintf("Chimney & Field Test Kit: %.1f m (%.0f ft) brick/terracotta chimney w/ guy-wires (~$%.0f) + $10 brine hydrometer, litmus pH strips & stem thermometer (~$%.0f) + %s charge ($%.2f)", chimneyM, chimneyFt, costs.chimneyUSD, costs.instrumentsUSD, alkali.name, alkali.costUSD),
	}
	if !zeroElec {
		localMaterials = append(localMaterials, fmt.Sprintf("Local Low-dP Draft Fan: 12V solar / 230V 50Hz jua-kali centrifugal/radiator fan (~$%.0f, 1.8 inWC static)", costs.fanUSD))
	}

	var manualSteps []string
	if zeroElec {
		manualSteps = []string{
			fmt.Sprintf("1. Pre-Warm the %.1f m (%.0f ft) Chimney Before Salting: Light the kiln slowly (or burn a 5-min dry-wood kindling primer in the chimney-base cleanout) until stack reheat reaches >280 F so natural buoyancy draft (%.3f inWC avail vs %.3f inWC req) establishes cleanly.", chimneyM, chimneyFt, availDraftInWC, reqDraftInWC),
			fmt.Sprintf("2. Pre-Fill Overhead Tank & Dose Local Alkali: Pump %.0f L (%.0f gal) of water into the elevated %s (%.1f ft / %.2f PSI) and dissolve/stir %.2f kg (%.1f lbs) %s (%s) in the baffled brick sump to pH 8.0-8.5.", tankLiters, tankGal, prof.tankBrand, tankElevFt, gravityHeadPSI, alkali.kgPerFiring, alkali.lbsPerFiring, alkali.name, alkali.formula),
			fmt.Sprintf("3. Open Gravity Spray Valve & Pace Manual Pumping: Crack the downcomer valve to %.2f L/min (%.2f GPM) over the quench baffle & stoneware rings. Closed-loop refill requires %.1f min/hr on a treadle/hand pump (or %d × 20L jerrycans/hr) during the active salting window.", sprayFlowLPM, sprayFlowGPM, treadleMinPerHr, jerrycansPerHr),
			"4. Monitor Draft, pH & Sump Level Hourly: Check flame draw at the kiln spyhole (gentle inward pull), dip a litmus pH strip every 45-60 min (keep pH 7.5-8.5), and top up evaporated water (~1-2 jerrycans/hr).",
			"5. Cool & Neutralize Before Disposal: Let sump water cool to <=140 F (60 C), settle ash, verify pH 6.5-8.5 with litmus strips, check brine density with the $10 hydrometer (dump when SG >= 1.105), and flush the spray rose clean.",
		}
	} else {
		manualSteps = []string{
			fmt.Sprintf("1. Verify 230V 50Hz / 12V Solar Power & Pre-Charge Sump: Confirm RCD/GFCI protection on the %s (~%.0f W total w/ fan), fill the brick baffled sump & %.0f L overhead buffer tank, and dose %.2f kg (%.1f lbs) %s (%s) to pH 8.0-8.5.", manualPumpType, powerWatts, tankLiters, alkali.kgPerFiring, alkali.lbsPerFiring, alkali.name, alkali.formula),
			fmt.Sprintf("2. Start Wet Spray & Low-dP Fan Before High-Fire: Start the 0.5 HP monoblock pump (%.2f L/min / %.2f GPM) and local draft fan (1.8 inWC available vs %.3f inWC required) — adjust blast-gate damper for gentle %.2f inWC kiln suction.", sprayFlowLPM, sprayFlowGPM, reqDraftInWC, design.Input.TargetKilnDraftInWC),
			"3. Power-Outage Resilience: If local grid power drops mid-firing, the elevated overhead gravity tank keeps quench/bed spray flowing for 45-60 min while the all-ceramic firebrick/stoneware body tolerates >1,800 F dry heat without melting — switch seamlessly to manual hand-pump refill if outage continues!",
			"4. Hourly pH & Ash Check: Dip a litmus pH strip hourly (keep pH 7.5-8.5), skim floating ash from the baffled brick sump, and verify clean spray rose flow.",
			"5. Cool-Down & Neutralized Disposal: Cool sump to <=140 F (60 C), verify pH 6.5-8.5 & hydrometer SG < 1.105 for multi-firing reuse, and rinse the pump & spray rose with clean water.",
		}
	}

	warnings := []string{
		"DO NOT BUILD DIY PRESSURIZED STEAM BOILERS / STEAM-JET EJECTORS: Wrapping a closed pressurized water pipe or drum around the 2400 F kiln flue to drive a steam-jet ejector creates a LETHAL boiler explosion hazard without ASME-stamped pressure vessels and certified safety relief valves — NEVER use DIY pressurized steam!",
		fmt.Sprintf("GRAVITY TANK STAND STRUCTURAL BRACING: A filled %.0f L %s weighs ~%.0f kg (%.0f lbs) at 3.5 m (11.5 ft) elevation — build a diagonally X-braced timber/steel or reinforced masonry pier stand on level compacted footings to prevent collapse!", tankLiters, prof.tankBrand, tankLiters*1.05+20.0, (tankLiters*1.05+20.0)/model.KgPerLb),
		fmt.Sprintf("OUTDOOR-ONLY SITING & TALL CHIMNEY CLEARANCE: Never operate indoors — terminate the %.1f m (%.0f ft) brick/terracotta chimney outdoors well above studio roofs with lateral wall ties/guy-wires against wind gusts so CO, residual HCl, and fine salt haze disperse safely.", chimneyM, chimneyFt),
	}
	if alkali.scalingWarning != "" {
		warnings = append(warnings, alkali.scalingWarning)
	}
	if !draftFeasible {
		warnings = append(warnings, fmt.Sprintf("PASSIVE DRAFT SHORTFALL: Required draft (%.3f inWC) exceeds natural chimney buoyancy (%.3f inWC) — reduce target kiln draft magnitude, widen the stoneware bed, or toggle Zero Electricity OFF to add a local low-dP fan.", reqDraftInWC, availDraftInWC))
	}

	modeTag := "Local-Grid / 12V Solar Hybrid Ceramic Mode"
	if zeroElec {
		modeTag = "100% Zero-Electricity Passive Ceramic Mode"
	}
	summary := fmt.Sprintf(
		"%s (%s): %.0f\" wide brick/SWG stoneware-packed tower (v_s=%.2f m/s, Ergun bed dP=%.3f inWC) + %.1f m (%.0f ft) brick chimney (%.0f F reheat -> %.3f inWC avail vs %.3f inWC req, margin %+.3f inWC, feasible=%v). Overhead %.0f L tank @ %.1f ft (%.2f PSI, %.2f L/min) + %s. Neutralizes %.1f%% HCl with %.2f kg %s (%s); honest %.0f%% submicron NaCl capture; est. local cost %s (%.0f W).",
		modeTag, prof.label, bedDiaIn, vSup, bedDPInWC, chimneyM, chimneyFt, stackReheatF,
		availDraftInWC, reqDraftInWC, draftMarginInWC, draftFeasible,
		tankLiters, tankElevFt, gravityHeadPSI, sprayFlowLPM, manualPumpType,
		hclPct, alkali.kgPerFiring, alkali.name, alkali.formula, submicronPct, localCurrency, powerWatts,
	)

	return &FrugalAssessment{
		Enabled:                 enabled,
		ZeroElectricity:         zeroElec,
		Region:                  prof.id,
		RegionLabel:             prof.label,
		ChimneyHeightFt:         chimneyFt,
		ChimneyHeightM:          chimneyM,
		StackReheatTempF:        stackReheatF,
		AvailableDraftInWC:      availDraftInWC,
		RequiredDraftInWC:       reqDraftInWC,
		DraftMarginInWC:         draftMarginInWC,
		DraftFeasible:           draftFeasible,
		DraftMechanism:          draftMechanism,
		BedDiameterIn:           bedDiaIn,
		BedSuperficialVelMS:     round2(vSup),
		BedDPInWC:               bedDPInWC,
		PackingMaterial:         packingMaterial,
		GravityTankGal:          tankGal,
		GravityTankLiters:       tankLiters,
		TankElevationFt:         tankElevFt,
		GravityHeadPSI:          gravityHeadPSI,
		SprayFlowGPM:            sprayFlowGPM,
		SprayFlowLPM:            sprayFlowLPM,
		ManualPumpType:          manualPumpType,
		Jerrycans20LPerHr:       jerrycansPerHr,
		TreadleMinutesPerHr:     treadleMinPerHr,
		AlkaliID:                alkali.id,
		AlkaliName:              alkali.name,
		AlkaliFormula:           alkali.formula,
		AlkaliKgPerFiring:       alkali.kgPerFiring,
		AlkaliLbsPerFiring:      alkali.lbsPerFiring,
		AlkaliCostUSD:           alkali.costUSD,
		AlkaliLocalNote:         alkali.localNote,
		ScalingWarning:          alkali.scalingWarning,
		HClRemovalPct:           hclPct,
		CoarseAshCapturePct:     coarseAshPct,
		SubmicronNaClCapturePct: submicronPct,
		SubmicronHonestNote:     submicronNote,
		ThermalImmunityNote:     thermalImmunityNote,
		EstimatedLocalCostUSD:   costs.totalUSD,
		LocalCurrencyEstimate:   localCurrency,
		PowerWatts:              powerWatts,
		ManualSteps:             manualSteps,
		LocalMaterials:          localMaterials,
		Warnings:                warnings,
		Summary:                 summary,
	}
}

// buildLocalCeramicCOTSSummary builds a genuine Nairobi/Mumbai/Local Pottery BOM & COTSSummary
// when build_path == "local_ceramic_masonry" or zero_electricity == true, matching
// FrugalAssessment.EstimatedLocalCostUSD to the cent with realistic positive weights.
func buildLocalCeramicCOTSSummary(project StudioProject, design model.ScrubberDesign, frugal *FrugalAssessment) COTSSummary {
	p := project.Normalize()
	if frugal == nil {
		frugal = EvaluateFrugal(p, design)
	}
	prof := resolveLocalRegionProfile(p.EffectiveLocalRegion())
	alkali := computeLocalAlkali(p, design, prof)

	bedDepthM := 1.20
	if design.Input.KilnVolumeCuFt <= 16.0 {
		bedDepthM = 1.10
	} else if design.Input.KilnVolumeCuFt >= 40.0 {
		bedDepthM = 1.35
	}
	costs := computeLocalCeramicCosts(p, design, prof, frugal.BedDiameterIn, bedDepthM, frugal.ChimneyHeightM, frugal.GravityTankLiters, alkali)

	// Recalculate actual % of flooding velocity for the wide masonry column (very conservative ~12-25%).
	bedDiaM := float64(model.InchesToMeters(frugal.BedDiameterIn))
	bedArea := math.Pi * bedDiaM * bedDiaM / 4.0
	actualPctFlood := 18.0
	if bedArea > 0 && design.PackedBed.FloodingVelocity > 0 && design.QuenchedGas.VolumetricFlow > 0 {
		vSup := float64(design.QuenchedGas.VolumetricFlow) / bedArea
		actualPctFlood = round2(math.Min(70.0, math.Max(5.0, (vSup/float64(design.PackedBed.FloodingVelocity))*100.0)))
	}

	quenchPart := COTSPart{
		SKU:         "LOCAL-BRICK-QUENCH",
		Name:        "Refractory Firebrick Quench & Concentric Flue-Wall Reheat Chamber",
		NominalIn:   round1(math.Max(14.0, design.Quench.TowerDiameter.Inches()+4.0)),
		Material:    prof.pipeLocalSpec + " (>1,800 F Dry-Run Immune)",
		Source:      "Local Brickyard / Builders Depot (" + prof.label + ")",
		UnitCostUSD: costs.quenchChamberUSD,
		WeightKg:    round2(240.0 * math.Max(0.75, design.Input.KilnVolumeCuFt/24.0)),
		Notes:       fmt.Sprintf("Fired firebrick & refractory mortar quench/recuperator chamber recovers waste heat from %.0f F flue to warm clean exhaust to %.0f F with zero plastic meltdown risk.", design.Input.PeakExitTempF, frugal.StackReheatTempF),
	}

	colPart := COTSPart{
		SKU:         fmt.Sprintf("LOCAL-SWG-TOWER-%02.0f", frugal.BedDiameterIn),
		Name:        fmt.Sprintf(`%.0f" Wide Brick / Salt-Glazed Stoneware (SWG) Absorber Tower`, frugal.BedDiameterIn),
		NominalIn:   frugal.BedDiameterIn,
		Material:    prof.pipeLocalSpec + " (Acid-Proof & >1,800 F Rated)",
		Source:      "Local Brickyard / IS 651 SWG Pipe Supplier (" + prof.label + ")",
		UnitCostUSD: costs.masonryTowerUSD,
		WeightKg:    round2(310.0 * (frugal.BedDiameterIn / 24.0)),
		Notes:       fmt.Sprintf("Wide %.0f\" bore holds superficial velocity to %.2f m/s (irrigated Ergun bed dP = %.3f inWC, %.1f%% flood) for natural chimney buoyancy draft.", frugal.BedDiameterIn, frugal.BedSuperficialVelMS, frugal.BedDPInWC, actualPctFlood),
	}

	sumpPart := COTSPart{
		SKU:         "LOCAL-BRICK-SUMP",
		Name:        "Plastered Brick / Glazed-Tile Baffled Settling Sump + Braced Overhead Gravity Tank",
		NominalIn:   30.0,
		Material:    fmt.Sprintf("Cement-Plastered Brick Baffled Sump + %.0f L %s on Braced 3.5 m Stand", frugal.GravityTankLiters, prof.tankBrand),
		Source:      "Local Hardware / Tank Depot (" + prof.label + ")",
		UnitCostUSD: round2(costs.brickSumpUSD + costs.gravityTankUSD + costs.standBracingUSD),
		WeightKg:    round2(145.0 + 0.08*frugal.GravityTankLiters),
		Notes:       fmt.Sprintf("Atmospheric baffled settling sump + elevated %.0f L (%.0f gal) %s @ %.1f ft (%.2f PSI gravity head) on diagonally braced stand rated for >%.0f kg wet load.", frugal.GravityTankLiters, frugal.GravityTankGal, prof.tankBrand, frugal.TankElevationFt, frugal.GravityHeadPSI, frugal.GravityTankLiters*1.1+25.0),
	}

	nozzlePart := COTSPart{
		SKU:         "LOCAL-GRAV-ROSE",
		Name:        "Terracotta Settling Baffles & Brass/PVC Coarse Gravity Spray Rose",
		NominalIn:   1.0,
		Material:    "Hand-Formed Terracotta Baffles + Brass / PVC Low-Head Spray Rose",
		Source:      "Potter's Studio Kiln + Local Plumbing Shop",
		UnitCostUSD: costs.baffleZoneUSD,
		WeightKg:    18.0,
		Notes:       fmt.Sprintf("Operates cleanly at %.2f PSI gravity head (%.2f L/min / %.2f GPM) with large non-clogging orifices for coarse quench & stoneware bed wetting.", frugal.GravityHeadPSI, frugal.SprayFlowLPM, frugal.SprayFlowGPM),
	}

	packingPart := COTSPart{
		SKU:         "LOCAL-STONEWARE-RINGS-1IN",
		Name:        `1" Potter-Extruded Unglazed Stoneware Raschig Rings (1 cu.ft.)`,
		NominalIn:   1.0,
		Material:    "Unglazed Grogged Stoneware Clay Fired to Cone 6-10 (70% Voidage, Acid & >1,800 F Proof)",
		Source:      "Potter's Own Clay Extruder & Regular Kiln Firings",
		UnitCostUSD: prof.ringCostPerCuFt,
		WeightKg:    21.0,
		Notes:       frugal.PackingMaterial,
	}

	var blowerPart COTSPart
	if p.ZeroElectricity {
		blowerPart = COTSPart{
			SKU:         fmt.Sprintf("LOCAL-CHIMNEY-%.0fM", math.Round(frugal.ChimneyHeightM)),
			Name:        fmt.Sprintf("Natural-Draft Brick / Terracotta Thermal-Reheat Chimney (%.1f m / %.0f ft) + Guy-Wires (0 W)", frugal.ChimneyHeightM, frugal.ChimneyHeightFt),
			NominalIn:   12.0,
			Material:    "Fired Clay Bricks / Salt-Glazed SWG Flue Tiles + Galvanized Guy-Wires & Wall Ties",
			Source:      "Local Brickyard / Hardware (" + prof.label + ")",
			UnitCostUSD: costs.chimneyUSD,
			WeightKg:    round2(38.0 * frugal.ChimneyHeightM),
			Notes:       frugal.DraftMechanism,
		}
	} else {
		blowerPart = COTSPart{
			SKU:         "LOCAL-HYBRID-FAN-CHIMNEY",
			Name:        fmt.Sprintf("Local 12V Solar / 230V Low-dP Draft Fan + %.1f m (%.0f ft) Brick/SWG Chimney", frugal.ChimneyHeightM, frugal.ChimneyHeightFt),
			NominalIn:   10.0,
			Material:    "Local Jua-Kali / Automotive-Style Low-dP Fan (1.8 inWC) + Ceramic Flue Chimney",
			Source:      "Local Electrical / Auto-Spares & Brickyard (" + prof.label + ")",
			UnitCostUSD: round2(costs.fanUSD + costs.chimneyUSD),
			WeightKg:    round2(8.5 + 32.0*frugal.ChimneyHeightM),
			Notes:       frugal.DraftMechanism,
		}
	}

	sensorPart := COTSPart{
		SKU:         "LOCAL-FIELD-TEST-KIT",
		Name:        "Field Hydrometer, Litmus pH Strips, Stem Thermometer & Local Alkali Charge",
		NominalIn:   0.0,
		Material:    fmt.Sprintf("Brine Hydrometer (SG 1.000-1.200) + Litmus pH Strips + Stem Thermometer + %s (%s)", alkali.name, alkali.formula),
		Source:      "Local Chemical / Agro-Vet / Hardware Shop (" + prof.label + ")",
		UnitCostUSD: round2(costs.instrumentsUSD + costs.alkaliUSD),
		WeightKg:    round2(math.Max(1.5, 1.2+alkali.kgPerFiring)),
		Notes:       fmt.Sprintf("Zero electronics required: verify sump pH 7.5-8.5 with litmus strips, monitor brine dump threshold (SG >= 1.105) with $10 hydrometer, and stage %.2f kg %s.", alkali.kgPerFiring, alkali.name),
	}

	items := []model.BOMItem{
		{
			Category:    "Stage 1: Brick Quench & Reheat",
			Name:        quenchPart.Name,
			Material:    quenchPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", quenchPart.SKU, quenchPart.Notes),
			UnitCostUSD: quenchPart.UnitCostUSD,
			WeightKg:    quenchPart.WeightKg,
		},
		{
			Category:    "Stage 2: Terracotta Baffles & Gravity Spray",
			Name:        nozzlePart.Name,
			Material:    nozzlePart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", nozzlePart.SKU, nozzlePart.Notes),
			UnitCostUSD: nozzlePart.UnitCostUSD,
			WeightKg:    nozzlePart.WeightKg,
		},
		{
			Category:    "Stage 3: Wide Masonry / SWG Absorber",
			Name:        colPart.Name,
			Material:    colPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", colPart.SKU, colPart.Notes),
			UnitCostUSD: colPart.UnitCostUSD,
			WeightKg:    colPart.WeightKg,
		},
		{
			Category:    "Stage 3: Wide Masonry / SWG Absorber",
			Name:        packingPart.Name,
			Material:    packingPart.Material,
			Quantity:    costs.ringCuFt,
			Unit:        "cu.ft.",
			Spec:        fmt.Sprintf("[%s] %s", packingPart.SKU, packingPart.Notes),
			UnitCostUSD: packingPart.UnitCostUSD,
			WeightKg:    round2(costs.ringCuFt * packingPart.WeightKg),
		},
		{
			Category:    "Stage 3: Wide Masonry / SWG Absorber",
			Name:        "Hand-Formed Terracotta Chevron Mist Baffles & Stoneware Support Arch",
			Material:    "Fired Unglazed Terracotta Louver Tiles & Firebrick Packing Grate",
			Quantity:    1,
			Unit:        "assy",
			Spec:        "[LOCAL-TERRA-DEMIST] Stacked 3-pass terracotta chevron mist baffles (0.020 inWC dP) + heavy firebrick/stoneware arch grate supporting ceramic ring weight",
			UnitCostUSD: costs.terracottaDemUSD,
			WeightKg:    26.0,
		},
		{
			Category:    "Liquid & Gravity Circulation",
			Name:        fmt.Sprintf("Brick Baffled Settling Sump + %.0f L %s + Braced 3.5 m Stand", frugal.GravityTankLiters, prof.tankBrand),
			Material:    sumpPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", sumpPart.SKU, sumpPart.Notes),
			UnitCostUSD: sumpPart.UnitCostUSD,
			WeightKg:    sumpPart.WeightKg,
		},
		{
			Category:    "Liquid & Gravity Circulation",
			Name:        fmt.Sprintf("%s & Gravity Downcomer Plumbing", frugal.ManualPumpType),
			Material:    "Cast-Iron / HDPE Pump + PVC/HDPE Gravity Downcomer, Foot Valve & Drain Valve",
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[LOCAL-PUMP-PLUMB] %s ($%.0f) + gravity downcomer/drain kit ($%.0f); flow %.2f L/min (%.2f GPM) at %.2f PSI head.", frugal.ManualPumpType, costs.pumpUSD, costs.plumbingUSD, frugal.SprayFlowLPM, frugal.SprayFlowGPM, frugal.GravityHeadPSI),
			UnitCostUSD: round2(costs.pumpUSD + costs.plumbingUSD),
			WeightKg:    18.5,
		},
		{
			Category:    "Stage 4: Chimney & Passive Draft",
			Name:        blowerPart.Name,
			Material:    blowerPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", blowerPart.SKU, blowerPart.Notes),
			UnitCostUSD: blowerPart.UnitCostUSD,
			WeightKg:    blowerPart.WeightKg,
		},
		{
			Category:    "Field Monitoring & Local Alkali",
			Name:        sensorPart.Name,
			Material:    sensorPart.Material,
			Quantity:    1,
			Unit:        "kit",
			Spec:        fmt.Sprintf("[%s] %s", sensorPart.SKU, sensorPart.Notes),
			UnitCostUSD: sensorPart.UnitCostUSD,
			WeightKg:    sensorPart.WeightKg,
		},
	}

	bom := model.BOM{
		Title: fmt.Sprintf("Local Ceramic & Masonry Bill of Materials — %s (%s, %.1f cu.ft.)",
			p.StudioName, prof.label, design.Input.KilnVolumeCuFt),
		Items: items,
		Notes: []string{
			"All line-item WeightKg values represent TOTAL line-item weight (already multiplied by Quantity).",
			fmt.Sprintf("Local pricing estimated for %s: %s.", prof.label, frugal.LocalCurrencyEstimate),
			frugal.ThermalImmunityNote,
			frugal.SubmicronHonestNote,
			"SAFETY: NEVER build an uncertified DIY pressurized steam boiler on the kiln flue. Brace the 3.5 m overhead gravity tank stand for >350-550 kg wet load, and discharge strictly outdoors.",
		},
	}
	bom.RecalculateTotals()

	return COTSSummary{
		Snapped:              true,
		QuenchVessel:         quenchPart,
		PackedColumn:         colPart,
		SumpTank:             sumpPart,
		SprayNozzle:          nozzlePart,
		PackingMedia:         packingPart,
		BlowerVFD:            blowerPart,
		Silencer:             nil,
		PlumeDiluter:         nil,
		SensorKit:            sensorPart,
		ActualPercentOfFlood: actualPctFlood,
		EstimatedTotalUSD:    round2(bom.TotalCostUSD),
		BOM:                  bom,
	}
}
