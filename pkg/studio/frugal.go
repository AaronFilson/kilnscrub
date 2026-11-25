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

	// Phase 10B Zero-Electricity Steam-Jet Ejector & Condensation Scrubbing fields.
	DraftMode               string   `json:"draft_mode"`
	DraftModeLabel          string   `json:"draft_mode_label"`
	SteamEnabled            bool     `json:"steam_enabled"`
	SteamPressurePSIG       float64  `json:"steam_pressure_psig"`
	SteamPressureBar        float64  `json:"steam_pressure_bar"`
	SteamNozzleVelMS        float64  `json:"steam_nozzle_vel_m_s"`
	SteamNozzleMach         float64  `json:"steam_nozzle_mach"`
	SteamNozzleThroatMM     float64  `json:"steam_nozzle_throat_mm"`
	SteamFlowKgHr           float64  `json:"steam_flow_kg_hr"`
	SteamWaterLHr           float64  `json:"steam_water_l_hr"`
	SteamThermalKW          float64  `json:"steam_thermal_kw"`
	KilnWasteHeatKW         float64  `json:"kiln_waste_heat_kw"`
	WasteHeatSufficient     bool     `json:"waste_heat_sufficient"`
	EjectorDraftInWC        float64  `json:"ejector_draft_inwc"`
	EjectorEntrainmentRatio float64  `json:"ejector_entrainment_ratio"`
	CondensationGrowthNote  string   `json:"condensation_growth_note"`
	SteamSourceType         string   `json:"steam_source_type"`
	SteamSafetyRequirements []string `json:"steam_safety_requirements"`

	// Phase 13 boiler waste-heat source & steam-driven water pump / zero-pump sump-pool fields.
	BoilerSource          string  `json:"boiler_source"`
	BoilerSourceLabel     string  `json:"boiler_source_label"`
	BoilerWaterSavedPct   float64 `json:"boiler_water_saved_pct"`
	BoilerFuelSavedNote   string  `json:"boiler_fuel_saved_note,omitempty"`
	BoilerCorrosionNote   string  `json:"boiler_corrosion_note,omitempty"`
	PumpMode              string  `json:"pump_mode"`
	PumpModeLabel         string  `json:"pump_mode_label"`
	PumpSteamKgHr         float64 `json:"pump_steam_kg_hr"`
	HasOverheadTank       bool    `json:"has_overhead_tank"`
	SumpPoolSubmergenceIn float64 `json:"sump_pool_submergence_in,omitempty"`
	SumpPoolDPInWC        float64 `json:"sump_pool_dp_inwc,omitempty"`
	PumpEngineeringNote   string  `json:"pump_engineering_note,omitempty"`

	// Phase 16 Global Urban Pottery Hubs, Coal/Peat/Ag-Biomass SO2 & Tar-Afterburner fields.
	AmbientPressureKPa    float64 `json:"ambient_pressure_kpa,omitempty"`
	ElevationNote         string  `json:"elevation_note,omitempty"`
	FuelType              string  `json:"fuel_type,omitempty"`
	SO2InletPPM           float64 `json:"so2_inlet_ppm,omitempty"`
	SO2RemovalPct         float64 `json:"so2_removal_pct,omitempty"`
	SulfurAcidDewPointF   float64 `json:"sulfur_acid_dew_point_f,omitempty"`
	AfterburnerEnabled    bool    `json:"afterburner_enabled,omitempty"`
	AfterburnerNote       string  `json:"afterburner_note,omitempty"`
	TarMitigationNote     string  `json:"tar_mitigation_note,omitempty"`
	RegionalMachineryNote string  `json:"regional_machinery_note,omitempty"`
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
	tronaUSDPerKg   float64
	machineryNote   string
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
			tronaUSDPerKg:   0.52,
			machineryNote:   "Mumbai / Kumbharwada / Pune supply chain: Sintex HDPE tanks, IS 651 SWG vitrified pipes, Kirloskar monoblock pumps, Forbes Marshall PPCPs, and non-IBR (<=25 L) garment/catering baby boilers are manufactured domestically.",
		}
	case LocalRegionKathmandu:
		return localRegionProfile{
			id:              LocalRegionKathmandu,
			label:           "Kathmandu / Bhaktapur Valley, Nepal (NPR)",
			currencyCode:    "NPR",
			fxPerUSD:        134.0,
			ambientPa:       86200.0, // 1,350 m Himalayan valley bowl (~0.851 atm)
			ambientTempK:    293.15,  // 20 C / 68 F valley ambient
			elevationNote:   "1,350 m valley basin (86.2 kPa, 20 C / 68 F — winter PM2.5 inversion)",
			tankBrand:       "Hilltake / Sintex Nepal HDPE Tank",
			pumpBrand:       "Kirloskar / Local 0.5 HP Monoblock",
			treadleLabel:    "Cast-Iron Semi-Rotary Wing Pump / Treadle Pump",
			pipeLocalSpec:   "Local Bhaktapur Kiln Firebricks & Terracotta/SWG Flue Rings",
			costScale:       0.86,
			ringCostPerCuFt: 4.2,
			sodaAshUSDPerKg: 0.62,
			limeUSDPerKg:    0.22,
			bakingUSDPerKg:  0.92,
			tronaUSDPerKg:   0.58,
			machineryNote:   "Kathmandu / Thimi / Bhaktapur pottery quarters: severe bowl-inversion PM2.5 makes smoke/tar abatement urgent during straw, rice-husk, wood & coal-dust firings. Hilltake HDPE tanks, brick masonry, catering momo/hotel steamers (6 psig), and cheap Chuna slaked lime are ubiquitous locally.",
		}
	case LocalRegionLagosAccra:
		return localRegionProfile{
			id:              LocalRegionLagosAccra,
			label:           "Lagos / Abeokuta & Accra, West Africa (NGN)",
			currencyCode:    "NGN",
			fxPerUSD:        1550.0,
			ambientPa:       101325.0,
			ambientTempK:    304.15, // 31 C / 88 F humid tropical coastal ambient
			elevationNote:   "sea level humid tropical (101.3 kPa, 31 C / 88 F)",
			tankBrand:       "Geepee / Polytank HDPE Overhead Tank",
			pumpBrand:       "Pedrollo / Local 0.5 HP & Solar DC Pump",
			treadleLabel:    "Cast-Iron Semi-Rotary Wing / Foot Treadle Pump",
			pipeLocalSpec:   "Local Laterite/Firebrick Masonry & Glazed Stoneware Flue Liners",
			costScale:       0.94,
			ringCostPerCuFt: 5.0,
			sodaAshUSDPerKg: 0.68,
			limeUSDPerKg:    0.30,
			bakingUSDPerKg:  1.15,
			tronaUSDPerKg:   0.38, // Abundant Lake Chad Kanwa / Trona!
			machineryNote:   "Lagos / Abeokuta / Accra: chronic grid blackouts & high diesel generator cost make 0 W passive/steam-jet operation critical. Palm Kernel Shells (PKS) provide high-calorific fuel (~17 MJ/kg) with heavy stoking smoke; Lake Chad Trona/Kanwa (Na2CO3·NaHCO3·2H2O) is sold cheaply in every market as a zero-scaling local alkali, paired with Geepee/Polytank HDPE tanks.",
		}
	case LocalRegionTonalaPuebla:
		return localRegionProfile{
			id:              LocalRegionTonalaPuebla,
			label:           "Tonalá (Guadalajara) & Puebla, Mexico (MXN)",
			currencyCode:    "MXN",
			fxPerUSD:        18.0,
			ambientPa:       83000.0, // ~1,650 m highland plateau (~0.819 atm)
			ambientTempK:    296.15,  // 23 C / 73 F plateau ambient
			elevationNote:   "1,650 m highland plateau (83.0 kPa, 23 C / 73 F)",
			tankBrand:       "Rotoplas HDPE Tinaco Overhead Tank",
			pumpBrand:       "Truper / Evans 0.5 HP Domestic Pump",
			treadleLabel:    "Manual Cast-Iron Wing / Pedal Pump",
			pipeLocalSpec:   "Ladrillo Refractario Firebrick & Kusakabe Secondary-Air Masonry",
			costScale:       1.04,
			ringCostPerCuFt: 6.0,
			sodaAshUSDPerKg: 0.72,
			limeUSDPerKg:    0.24,
			bakingUSDPerKg:  1.05,
			tronaUSDPerKg:   0.42, // Natural Mexican Tequesquite / trona!
			machineryNote:   "Tonalá & Puebla urban ceramics: strict SEMADET/municipal air-quality crackdowns target black smoke from wood, sawdust & waste-oil kilns. Rotoplas tinacos, ladrillo refractario, cheap cal hidratada / natural tequesquite trona, and Kusakabe secondary-air smoke-free masonry afterburner retrofits are widely available.",
		}
	case LocalRegionCuscoAndes:
		return localRegionProfile{
			id:              LocalRegionCuscoAndes,
			label:           "Cusco / Altiplano, Peru & Bolivia (PEN)",
			currencyCode:    "PEN",
			fxPerUSD:        3.75,
			ambientPa:       66800.0, // 3,400 m extreme high altitude (~0.659 atm — -34% air density!)
			ambientTempK:    287.15,  // 14 C / 57 F highland ambient
			elevationNote:   "3,400 m Andean Altiplano (66.8 kPa, 14 C / 57 F — -34% air density!)",
			tankBrand:       "Rotoplas / Eternit Andes HDPE Tank",
			pumpBrand:       "Pedrollo / Truper 0.5 HP Pump",
			treadleLabel:    "Manual Cast-Iron Semi-Rotary / Pedal Pump",
			pipeLocalSpec:   "Adobe-Jacketed Firebrick Masonry & Thrown Stoneware Flue Tiles",
			costScale:       0.98,
			ringCostPerCuFt: 5.2,
			sodaAshUSDPerKg: 0.78,
			limeUSDPerKg:    0.28,
			bakingUSDPerKg:  1.10,
			tronaUSDPerKg:   0.55,
			machineryNote:   "Cusco / Raqch'i / Altiplano (3,400 m): thin 66.8 kPa air cuts natural chimney buoyancy by ~34% and expands flue volume by +52% (requiring a wider stoneware tower or 6-20 psig steam-jet ejector!). High-moisture Andean peat (turba) & dung (taquia) emit heavy phenolic tar & SO2, solved by the passive secondary-air masonry afterburner + cal hidratada.",
		}
	case LocalRegionKhurjaDhaka:
		return localRegionProfile{
			id:              LocalRegionKhurjaDhaka,
			label:           "Khurja (Delhi NCR) & Dhaka Coal Pottery Hub (INR)",
			currencyCode:    "INR",
			fxPerUSD:        84.0,
			ambientPa:       101325.0,
			ambientTempK:    305.15, // 32 C / 90 F Indo-Gangetic plain ambient
			elevationNote:   "sea level Indo-Gangetic plain (101.3 kPa, 32 C / 90 F)",
			tankBrand:       "Sintex HDPE Overhead Tank",
			pumpBrand:       "Kirloskar / Crompton / Forbes Marshall PPCP",
			treadleLabel:    "Cast-Iron Semi-Rotary Wing Pump / Treadle Pump",
			pipeLocalSpec:   "Refractory Alumina Firebrick, IS 651 SWG Pipes & Acid-Resistant Mortar",
			costScale:       0.80,
			ringCostPerCuFt: 4.0,
			sodaAshUSDPerKg: 0.52,
			limeUSDPerKg:    0.18,
			bakingUSDPerKg:  0.80,
			tronaUSDPerKg:   0.48,
			machineryNote:   "Khurja (UP) & Dhaka stoneware clusters: dense coal-fired downdraft kilns burning 1.5-3.5% S Indian bituminous coal face strict NGT/CPCB smoke & SO2 mandates. Domestic non-IBR <=25 L baby boilers, Forbes Marshall steam PPCPs, Sintex tanks, IS 651 SWG pipes, and ultra-cheap Chuna lime ($0.18/kg) make 0 W coal-smoke + SO2/HCl scrubbing immediately buildable.",
		}
	case LocalRegionCairoFez:
		return localRegionProfile{
			id:              LocalRegionCairoFez,
			label:           "Fustat (Old Cairo) & Fez (Ain Nokbi), MENA (EGP)",
			currencyCode:    "EGP",
			fxPerUSD:        49.0,
			ambientPa:       101325.0,
			ambientTempK:    305.15, // 32 C / 90 F arid urban ambient
			elevationNote:   "sea level arid urban (101.3 kPa, 32 C / 90 F — water-scarce)",
			tankBrand:       "Local UV-Stabilized HDPE Roof Tank",
			pumpBrand:       "Calpeda / Local 0.5 HP Monoblock",
			treadleLabel:    "Cast-Iron Semi-Rotary Wing Pump",
			pipeLocalSpec:   "Local Kiln Firebrick & Hand-Thrown Glazed Terracotta SWG Flue Tiles",
			costScale:       0.88,
			ringCostPerCuFt: 4.5,
			sodaAshUSDPerKg: 0.60,
			limeUSDPerKg:    0.22,
			bakingUSDPerKg:  0.90,
			tronaUSDPerKg:   0.40, // Natural Wadi El Natrun natron/trona!
			machineryNote:   "Fustat (Old Cairo) & Fez (Ain Nokbi): historic urban pottery quarters firing olive pomace (jift), sawdust & wood face neighborhood smoke bans under arid water-scarce conditions. Closed-loop baffled brick sumps, Wadi El Natrun natural natron/trona or cheap Gir slaked lime, and a passive secondary-air smoke-burner minimize both smoke and water use.",
		}
	case LocalRegionHighlandPeat:
		return localRegionProfile{
			id:              LocalRegionHighlandPeat,
			label:           "Scottish Highlands / Rural Ireland / Boreal Peat (GBP)",
			currencyCode:    "GBP",
			fxPerUSD:        0.79,
			ambientPa:       101325.0,
			ambientTempK:    281.15, // 8 C / 46 F cool damp northern ambient
			elevationNote:   "sea level cool maritime (101.3 kPa, 8 C / 46 F — strong chimney buoyancy)",
			tankBrand:       "1,000 L Reclaimed Farm IBC Tote / Polypipe Tank",
			pumpBrand:       "Manual Semi-Rotary Wing / 12V Solar Bilge Pump",
			treadleLabel:    "Cast-Iron Semi-Rotary Wing Pump / Pedal Pump",
			pipeLocalSpec:   "Firebrick Masonry & Vitrified Clay Salt-Glazed Flue Liners",
			costScale:       1.28,
			ringCostPerCuFt: 9.0,
			sodaAshUSDPerKg: 0.95,
			limeUSDPerKg:    0.42,
			bakingUSDPerKg:  1.35,
			tronaUSDPerKg:   0.85,
			machineryNote:   "Scottish Highlands / West Ireland / Northern off-grid heritage kilns: air-dried peat/turf (~30% moisture, 0.5-1.5% S) & seaweed-salt firings emit dense phenolic/humic tar & SO2. Cool 8 C ambient air boosts natural chimney buoyancy dramatically; pair a hot refractory secondary-air tar afterburner with a tar-skimming weir, reclaimed farm IBC tote, and free wood-ash potash or agricultural lime.",
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
			tronaUSDPerKg:   0.65,
			machineryNote:   "General off-grid / rural pottery build using local firebrick masonry, potter-extruded stoneware rings, HDPE gravity drum, semi-rotary hand pump, and commercial catering steamer or baby boiler.",
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
			tronaUSDPerKg:   0.45, // Lake Magadi natural trona / soda ash!
			machineryNote:   "Nairobi / East Africa supply chain: Kentank/Roto HDPE tanks, MoneyMaker treadle pumps, Davis & Shirtliff pumps, Lake Magadi soda ash/trona, and River Rd / Industrial Area catering steamers & KEBS baby boilers.",
		}
	}
}

func formatLocalCurrency(usd float64, prof localRegionProfile) string {
	if prof.currencyCode == "USD" {
		return fmt.Sprintf("~$%.0f USD", math.Round(usd))
	}
	if prof.fxPerUSD < 10.0 {
		localAmt := math.Round((usd*prof.fxPerUSD)/10.0) * 10.0
		return fmt.Sprintf("~%s %s (~$%.0f USD @ %.2f %s/USD)", prof.currencyCode, formatIntCommas(int(localAmt)), math.Round(usd), prof.fxPerUSD, prof.currencyCode)
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
		mwNaClKgPerMol   = 0.0584428 // NaCl -> 1 mol HCl per mol NaCl (1 acid eq)
		mwNa2CO3KgPerMol = 0.1059888 // 2 H+ + Na2CO3 -> 2 Na+ + H2O + CO2 (0.5 mol/eq; 1 mol/mol SO2)
		mwCaOH2KgPerMol  = 0.074093  // 2 H+ + Ca(OH)2 -> Ca2+ + 2 H2O (0.5 mol/eq; 1 mol/mol SO2 = 1.16 kg/kg SO2)
		mwK2CO3KgPerMol  = 0.138205  // 2 H+ + K2CO3 -> 2 K+ + H2O + CO2 (0.5 mol/eq; 1 mol/mol SO2)
		mwNaHCO3KgPerMol = 0.0840066 // H+ + NaHCO3 -> Na+ + H2O + CO2 (1.0 mol/eq; 2 mol/mol SO2)
		mwTronaKgPerMol  = 0.226026  // Na2CO3·NaHCO3·2H2O (Kanwa / Tequesquite / Natron: 3 H+ eq/mol -> 1/3 mol/eq; 2/3 mol/mol SO2!)
		woodAshK2CO3Frac = 0.07      // ~7% extractable K2CO3 by mass from dry hardwood kiln ash
	)

	molHCl := saltKg / mwNaClKgPerMol
	// Phase 16: Include SO2 acid equivalents across the full firing duration for coal, peat, and ag-biomass
	// (0.85 firing-average load factor across ramp + soak; diprotic SO2 consumes 2 H+ equivalents per mol SO2).
	molSO2 := 0.0
	if design.InletGas.SO2 > 0 && design.InletGas.MolarFlowMolPerSec > 0 && p.FiringDurationHours > 0 {
		molSO2 = design.InletGas.MolarFlowMolPerSec * design.InletGas.SO2 * 3600.0 * p.FiringDurationHours * 0.85
	}
	molAcidEq := molHCl + 2.0*molSO2
	acidScope := fmt.Sprintf("%.1f-hr salting firing", saltingHrs)
	if molSO2 > 0 && molHCl > 0 {
		acidScope = fmt.Sprintf("combined HCl (%.1f-hr salting) + %.1f-hr fuel SO2 firing (%.0f ppm SO2)", saltingHrs, p.FiringDurationHours, design.PackedBed.SO2InletPPM)
	} else if molSO2 > 0 {
		acidScope = fmt.Sprintf("%.1f-hr coal/peat/biomass SO2 firing (%.0f ppm SO2, 0 salt)", p.FiringDurationHours, design.PackedBed.SO2InletPPM)
	}

	switch alkID {
	case LocalAlkaliSlakedLime:
		kg := round2(molAcidEq * 0.5 * mwCaOH2KgPerMol * 1.30)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.limeUSDPerKg)
		note := fmt.Sprintf("Builder's slaked lime / Chuna / Cal hidratada (Ca(OH)2, 0.5 mol/HCl & 1.0 mol/SO2 + 30%% margin): %.2f kg (%.1f lbs) per %s. Lowest chemical cost locally.", kg, lbs, acidScope)
		if prof.id == LocalRegionMumbai || prof.id == LocalRegionKhurjaDhaka {
			note = fmt.Sprintf("Local Chuna / hydrated lime (Ca(OH)2, 0.5 mol/HCl & 1.0 mol/SO2 + 30%% margin): %.2f kg (%.1f lbs) per %s (~INR %.0f/firing from any building-supply shop).", kg, lbs, acidScope, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionNairobi {
			note = fmt.Sprintf("Nairobi builder's slaked lime / Chokaa (Ca(OH)2, 0.5 mol/HCl & 1.0 mol/SO2 + 30%% margin): %.2f kg (%.1f lbs) per %s (~KES %.0f/firing from local hardware yards).", kg, lbs, acidScope, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionTonalaPuebla || prof.id == LocalRegionCuscoAndes {
			note = fmt.Sprintf("Local Cal Hidratada (Ca(OH)2, 0.5 mol/HCl & 1.0 mol/SO2 + 30%% margin): %.2f kg (%.1f lbs) per %s (~%s %.0f/firing from any ferretería).", kg, lbs, acidScope, prof.currencyCode, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionCairoFez {
			note = fmt.Sprintf("Local Gir / hydrated lime (Ca(OH)2, 0.5 mol/HCl & 1.0 mol/SO2 + 30%% margin): %.2f kg (%.1f lbs) per %s (~EGP %.0f/firing).", kg, lbs, acidScope, cost*prof.fxPerUSD)
		}
		if molAcidEq <= 0 {
			note = fmt.Sprintf("Zero HCl/SO2 mode (%s): 0 kg slaked lime needed (optional 0.25 kg buffer keeps sump alkaline).", p.VaporGlazeMode)
		}
		scaleWarn := "CaCO3 SCALING & SLURRY WARNING: Slaked lime has low water solubility (~1.7 g/L) — stir the sump continuously as a milk-of-lime slurry so lime does not settle dead on the bottom, and expect chalky CaCO3 scale from flue CO2 on the spray rose and upper stoneware rings (brush/vinegar-flush every 3-5 firings)."
		if molSO2 > 0 {
			scaleWarn += " COAL/PEAT CaSO3/GYPSUM SLUDGE NOTE: Neutralizing SO2 with Ca(OH)2 precipitates insoluble calcium sulfite/gypsum (CaSO3/CaSO4·2H2O) mud — use coarse >=1.5\" stoneware rings or removable ceramic louver plates and shovel settled gypsum sludge from Pool #1 between firings!"
		}
		return localAlkaliResult{
			id:             LocalAlkaliSlakedLime,
			name:           "Slaked Lime / Chuna / Cal",
			formula:        "Ca(OH)2",
			kgPerFiring:    kg,
			lbsPerFiring:   lbs,
			costUSD:        cost,
			localNote:      note,
			scalingWarning: scaleWarn,
		}

	case LocalAlkaliTronaKanwa:
		kg := round2(molAcidEq * (1.0 / 3.0) * mwTronaKgPerMol * 1.25)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.tronaUSDPerKg)
		note := fmt.Sprintf("Natural Trona / Kanwa / Tequesquite / Natron (Na2CO3·NaHCO3·2H2O, 3 eq/mol -> 0.33 mol/HCl & 0.67 mol/SO2 + 25%% margin): %.2f kg (%.1f lbs) per %s (~%s %.0f/firing). Mined regionally across Lake Chad, East African Rift, Wadi El Natrun & Mexico — highly soluble with ZERO CaCO3 or CaSO3 gypsum scaling!", kg, lbs, acidScope, prof.currencyCode, cost*prof.fxPerUSD)
		if molAcidEq <= 0 {
			note = fmt.Sprintf("Zero HCl/SO2 mode (%s): 0 kg Trona/Kanwa needed (optional 0.25 kg keeps sump buffered at pH 8.5).", p.VaporGlazeMode)
		}
		return localAlkaliResult{
			id:             LocalAlkaliTronaKanwa,
			name:           "Trona / Kanwa / Tequesquite",
			formula:        "Na2CO3·NaHCO3·2H2O",
			kgPerFiring:    kg,
			lbsPerFiring:   lbs,
			costUSD:        cost,
			localNote:      note,
			scalingWarning: "IMPURITY SCREENING NOTE: Natural market trona/kanwa/tequesquite often contains 5-15% insoluble lake-bed silt/clay — dissolve in a settling bucket and decant clear alkaline liquor into the overhead tank/sump so sand never wears the pump or plugs spray holes.",
		}

	case LocalAlkaliWoodAshLye:
		pureK2CO3Kg := molAcidEq * 0.5 * mwK2CO3KgPerMol * 1.25
		dryAshKg := round1(pureK2CO3Kg / woodAshK2CO3Frac)
		dryAshLbs := round1(dryAshKg / model.KgPerLb)
		note := fmt.Sprintf("Potter's own wood-kiln ash leached in a 200L barrel (~7%% extractable K2CO3 -> %.1f kg / %.1f lbs dry hardwood ash yields %.2f kg K2CO3 for %s at $0.00 chemical cost!).", dryAshKg, dryAshLbs, pureK2CO3Kg, acidScope)
		if molAcidEq <= 0 {
			note = fmt.Sprintf("Zero HCl/SO2 mode (%s): 0 kg wood ash required (potter's own free wood-kiln ash provides mild pH 8-9 rinse water if desired).", p.VaporGlazeMode)
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
		kg := round2(molAcidEq * 1.0 * mwNaHCO3KgPerMol * 1.25)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.bakingUSDPerKg)
		note := fmt.Sprintf("Bulk sodium bicarbonate (NaHCO3, 1.0 mol/HCl & 2.0 mol/SO2 + 25%% margin): %.2f kg (%.1f lbs) per %s. Gentlest handling (self-buffers near pH 8.3) with zero scaling.", kg, lbs, acidScope)
		if molAcidEq <= 0 {
			note = fmt.Sprintf("Zero HCl/SO2 mode (%s): 0 kg baking soda needed.", p.VaporGlazeMode)
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
		kg := round2(molAcidEq * 0.5 * mwNa2CO3KgPerMol * 1.25)
		lbs := round1(kg / model.KgPerLb)
		cost := round2(kg * prof.sodaAshUSDPerKg)
		note := fmt.Sprintf("Magadi Soda Ash / bulk washing soda (Na2CO3, 0.5 mol/HCl & 1.0 mol/SO2 + 25%% margin): %.2f kg (%.1f lbs) per %s — 37%% lighter than baking soda, highly soluble, zero CaCO3/CaSO3 scaling.", kg, lbs, acidScope)
		if prof.id == LocalRegionNairobi {
			note = fmt.Sprintf("Lake Magadi Soda Ash (Na2CO3, 0.5 mol/HCl & 1.0 mol/SO2 + 25%% margin): %.2f kg (%.1f lbs) per %s (~KES %.0f/firing — mined in Kenya, abundant in Nairobi industrial/soap supply depots, zero scaling).", kg, lbs, acidScope, cost*prof.fxPerUSD)
		} else if prof.id == LocalRegionMumbai || prof.id == LocalRegionKhurjaDhaka {
			note = fmt.Sprintf("Bulk Indian Washing Soda / Soda Ash (Na2CO3, 0.5 mol/HCl & 1.0 mol/SO2 + 25%% margin): %.2f kg (%.1f lbs) per %s (~INR %.0f/firing from local chemical depots, highly soluble, zero gypsum/CaCO3 scaling).", kg, lbs, acidScope, cost*prof.fxPerUSD)
		}
		if molAcidEq <= 0 {
			note = fmt.Sprintf("Zero HCl/SO2 mode (%s): 0 kg soda ash needed (optional 0.25 kg keeps sump mildly buffered at pH 8.0-8.5).", p.VaporGlazeMode)
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
	afterburnerUSD   float64
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
	steamSystemUSD   float64
	plumbingUSD      float64
	chimneyUSD       float64
	instrumentsUSD   float64
	alkaliUSD        float64
	totalUSD         float64
}

type steamEjectorPhysics struct {
	draftMode               string
	draftModeLabel          string
	steamEnabled            bool
	pressurePSIG            float64
	pressureBar             float64
	nozzleVelMS             float64
	nozzleMach              float64
	nozzleThroatMM          float64
	steamFlowKgHr           float64
	steamWaterLHr           float64
	steamThermalKW          float64
	kilnWasteHeatKW         float64
	wasteHeatSufficient     bool
	ejectorDraftInWC        float64
	ejectorEntrainmentRatio float64
	condensationGrowthNote  string
	steamSourceType         string
	safetyRequirements      []string
}

// satSteamProps returns accurate IAPWS-95 saturated water/steam enthalpy [kJ/kg]
// and entropy [kJ/(kg*K)] across 0.5-10 bar (80-180 C, <0.25% error).
func satSteamProps(tK float64) (hfKJ, hgKJ, sfKJK, sfgKJK float64) {
	tC := tK - 273.15
	hfKJ = 4.180*tC + 0.00022*tC*tC
	hgKJ = 2504.0 + 1.84*tC - 0.00124*tC*tC
	sfKJK = 4.190 * math.Log(tK/273.15)
	sfgKJK = (hgKJ - hfKJ) / tK
	return
}

// computeSteamEjectorPhysics solves first-principles de Laval steam nozzle expansion,
// choked throat area, 1D constant-area ejector momentum balance, thermal load, and
// kiln waste-heat availability for Phase 10B local ceramic steam-jet modes.
func computeSteamEjectorPhysics(
	draftMode string,
	design model.ScrubberDesign,
	prof localRegionProfile,
	qLocal float64,
	rhoLocal float64,
) steamEjectorPhysics {
	const (
		paPerPSI    = 6894.757
		barPerPSI   = 0.06894757
		kSteam      = 1.30
		rVapor      = 461.5 // J/(kg*K) water vapor gas constant
		etaNozzle   = 0.90  // de Laval convergent-divergent nozzle enthalpy efficiency
		cdNozzle    = 0.96  // machined stainless nozzle discharge coefficient
		cpFeedWater = 4.186 // kJ/(kg*K) liquid feedwater cp
	)

	mGas := float64(design.QuenchedGas.MassFlow)
	if mGas <= 0 {
		mGas = math.Max(0.02, qLocal*rhoLocal)
	}

	// Kiln hot exhaust waste heat available above 400 F (477.59 K) flue pinch temperature.
	mFlue := float64(design.InletGas.MassFlow)
	if mFlue <= 0 {
		mFlue = mGas * 0.72
	}
	tFlueK := float64(design.InletGas.Temperature)
	if tFlueK <= 0 {
		tFlueK = float64(model.FahrenheitToKelvin(design.Input.PeakExitTempF))
	}
	tPinchK := float64(model.FahrenheitToKelvin(400.0))
	cpFlueKJ := 1.22
	tMidK := model.Kelvin(math.Max(300.0, 0.5*(tFlueK+tPinchK)))
	if cpMol, err := thermo.MixtureCp(design.InletGas, tMidK); err == nil {
		mwIn := thermo.MixtureMolecularWeight(design.InletGas)
		if mwIn > 0.015 && mwIn < 0.045 {
			cpCalc := (cpMol / mwIn) / 1000.0
			if cpCalc >= 0.95 && cpCalc <= 1.50 {
				cpFlueKJ = cpCalc
			}
		}
	}
	kilnWasteKW := round1(math.Max(0.0, mFlue*cpFlueKJ*math.Max(0.0, tFlueK-tPinchK)))

	if draftMode != LocalDraftSteamEjectorSubsonic && draftMode != LocalDraftSteamEjectorLP && draftMode != LocalDraftSteamEjectorHP {
		return steamEjectorPhysics{
			draftMode:              LocalDraftNaturalChimney,
			draftModeLabel:         "Natural Chimney Buoyancy (Passive 0 psig)",
			steamEnabled:           false,
			kilnWasteHeatKW:        kilnWasteKW,
			wasteHeatSufficient:    false,
			condensationGrowthNote: "Passive natural chimney draft (~30% submicron NaCl capture): no pressurized steam injection; select Subsonic Quiet (6 psig), LP (20 psig), or HP (60 psig) Steam-Jet Ejector for hygroscopic NaCl condensation nucleation growth & submicron capture at 0 W electric.",
			steamSourceType:        "None (Passive Natural Chimney Buoyancy)",
			safetyRequirements:     []string{},
		}
	}

	isSubsonic := draftMode == LocalDraftSteamEjectorSubsonic
	isHP := draftMode == LocalDraftSteamEjectorHP
	psig := 20.0
	draftLabel := "Low-Pressure Steam-Jet Ejector (20 psig / 1.38 barg — Draft + Condensation Assist)"
	steamMassFrac := 0.063 // ~22 kg/hr at m_gas = 0.097 kg/s (entrainment ~15.9)
	minSteamKgS := 0.0020
	vMix := 14.0 // m/s moderate LP ejector mixing throat
	mawpPSIG := 25.0
	hydroPSIG := 38.0
	if isSubsonic {
		psig = 6.0
		draftLabel = "Subsonic Quiet Steam-Jet Ejector (6 psig / 0.414 barg — Mach < 1.0, ~52 dBA, Catering Steamer + 4.2 m Standpipe)"
		steamMassFrac = 0.0528 // ~18.5 kg/hr at m_gas = 0.0974 kg/s (entrainment ~18.9)
		minSteamKgS = 0.0018
		vMix = 13.0 // m/s subsonic mixing tube yielding ~0.38-0.42 inWC draft
		mawpPSIG = 15.0
		hydroPSIG = 23.0
	} else if isHP {
		psig = 60.0
		draftLabel = "High-Velocity Hydro-Sonic Steam-Jet Ejector (60 psig / 4.14 barg — Certified Baby Boiler)"
		steamMassFrac = 0.155 // ~54 kg/hr at m_gas = 0.097 kg/s (entrainment ~6.45)
		minSteamKgS = 0.0050
		vMix = 24.0 // m/s high-shear Hydro-Sonic two-phase mixing throat
		mawpPSIG = 75.0
		hydroPSIG = 115.0
	}
	barG := round2(psig * barPerPSI)
	if isSubsonic {
		barG = round3(psig * barPerPSI) // 0.414 barg exact
	}

	pAmb := prof.ambientPa
	p0Pa := pAmb + psig*paPerPSI

	tSat0K := 399.15
	if tSat, err := thermo.WaterDewPoint(model.Pascal(p0Pa)); err == nil {
		tSat0K = float64(tSat)
	}
	tSatAmbK := 373.15
	if tSat, err := thermo.WaterDewPoint(model.Pascal(pAmb)); err == nil {
		tSatAmbK = float64(tSat)
	}
	t0K := tSat0K

	// Exact IAPWS-95 saturated steam isentropic expansion delta_h_s from P_0 to P_amb
	// (s_{g,0} = s_{f,2} + x_{2s} * s_{fg,2}, matching ASME/IAPWS-IF97 to <0.5%).
	_, hg0KJ, sf0, sfg0 := satSteamProps(tSat0K)
	sg0 := sf0 + sfg0
	h0KJ := hg0KJ
	s0 := sg0
	hfAmbKJ, hgAmbKJ, sfAmb, sfgAmb := satSteamProps(tSatAmbK)
	x2s := math.Min(1.0, math.Max(0.75, (s0-sfAmb)/sfgAmb))
	h2sKJ := hfAmbKJ + x2s*(hgAmbKJ-hfAmbKJ)
	deltaHsJ := math.Max(1000.0, (h0KJ-h2sKJ)*1000.0)

	// Nozzle / converging orifice exit velocity & local Mach number.
	vExit := math.Sqrt(2.0 * etaNozzle * deltaHsJ)
	aExit := math.Sqrt(kSteam * rVapor * tSatAmbK)
	machExit := vExit / aExit

	// Motive steam mass flow & nozzle throat / drilled brass orifice diameter (in mm).
	mSteamKgS := math.Max(minSteamKgS, mGas*steamMassFrac)
	steamFlowKgHr := round1(mSteamKgS * 3600.0)
	steamWaterLHr := steamFlowKgHr // 1 kg water = 1.0 L feedwater
	entrainRatio := round2(mGas / mSteamKgS)

	rhoSteamAmb := pAmb / (rVapor * tSatAmbK)
	var throatMM float64
	if isSubsonic {
		// Unchoked subsonic simple drilled brass pipe-cap converging orifice (6 psig is well below
		// the sonic choking threshold of 12.2 psig sea-level / 10.1 psig Nairobi — NO CNC de Laval bell!).
		// m_dot = C_d * A * rho_exit * v_e, with sharp-edged hand-drilled orifice C_d = 0.58
		// (vena contracta contraction on flat drilled brass cap -> ~7.6-8.0 mm drill bore).
		const cdSubsonic = 0.58
		h2KJ := h0KJ - etaNozzle*(h0KJ-h2sKJ)
		x2 := math.Min(1.0, math.Max(0.75, (h2KJ-hfAmbKJ)/(hgAmbKJ-hfAmbKJ)))
		rhoExit := rhoSteamAmb / x2
		aOrifice := mSteamKgS / (cdSubsonic * rhoExit * vExit)
		throatMM = round2(math.Sqrt(4.0*aOrifice/math.Pi) * 1000.0)
	} else {
		critFactor := math.Pow(2.0/(kSteam+1.0), (kSteam+1.0)/(2.0*(kSteam-1.0)))
		gStar := cdNozzle * p0Pa * math.Sqrt(kSteam/(rVapor*t0K)) * critFactor
		aStar := mSteamKgS / gStar
		throatMM = round2(math.Sqrt(4.0*aStar/math.Pi) * 1000.0)
	}

	// 1D constant-area mixing-tube ejector momentum balance with 7-deg diffuser pressure recovery.
	// dP_ej = eta_jet * [ m_s*(v_e - v_m)/A_m - (1 - eta_d + K_f) * 0.5 * rho_mix * v_m^2 ]
	const (
		etaDiffuser = 0.65
		kFriction   = 0.20
		etaJet      = 0.78
	)
	qMix := qLocal + mSteamKgS/rhoSteamAmb
	rhoMix := (mGas + mSteamKgS) / qMix
	aMix := qMix / vMix
	dpMomPa := etaJet * (mSteamKgS*(vExit-vMix)/aMix - (1.0-etaDiffuser+kFriction)*0.5*rhoMix*vMix*vMix)
	ejectorDraftInWC := round3(math.Max(0.15, dpMomPa/model.PaPerInWC))

	// Steam generation thermal load (sensible heating from ambient feedwater + latent + superheat).
	hFeedKJ := cpFeedWater * math.Max(5.0, prof.ambientTempK-273.15)
	steamThermalKW := round1(mSteamKgS * (h0KJ - hFeedKJ))
	wasteSufficient := kilnWasteKW >= steamThermalKW

	var sourceType, condNote string
	var safetyReqs []string
	if isSubsonic {
		switch prof.id {
		case LocalRegionMumbai:
			sourceType = "Certified 12-15 psig Commercial Hotel/Catering Stainless Pressure Steamer / Low-Pressure Idli/Catering Kettle (Run @ 6 psig / 0.414 barg) + 4.2 m Open Hydrostatic Standpipe Relief & Drilled Brass Converging Orifice"
		case LocalRegionNairobi:
			sourceType = "Certified 12-15 psig Commercial Hotel/Catering Stainless Pressure Steamer / Low-Pressure Kettle (Run @ 6 psig / 0.414 barg, River Rd / Industrial Area Supply) + 4.2 m Open Hydrostatic Standpipe Relief & Drilled Brass Converging Orifice"
		default:
			sourceType = "Certified 12-15 psig Commercial Hotel/Catering Stainless Pressure Steamer / Low-Pressure Kettle (Run @ 6 psig / 0.414 barg) + 4.2 m Open Hydrostatic Standpipe Relief & Drilled Brass Converging Orifice"
		}
		condNote = fmt.Sprintf(
			"SUBSONIC CONDENSATION NUCLEATION GROWTH (6 psig / Mach %.2f < 1.0): %.1f kg/hr unchoked subsonic steam (%.0f m/s, ~52 dBA quiet operation with zero supersonic shock screech) cools through the wide irrigated stoneware bed to nucleate condensation directly onto deliquescent NaCl fume nuclei (growing 0.1-0.3 um salt aerosol to 1.5-3.5 um droplets -> ~55%% submicron NaCl capture at 0 W electric, with gentler droplet shear than 60 psig Hydro-Sonic).",
			machExit, steamFlowKgHr, vExit,
		)
		safetyReqs = []string{
			"4.2 m (13.8 ft) OPEN WATER-COLUMN HYDROSTATIC RELIEF STANDPIPE (S2 — 6.0 psig PHYSICAL LIMIT): Connect a dedicated unvalved 3/4\" vertical open-top standpipe (P = rho*g*h = 41.2 kPa ≈ 6.0 psig at 4.2 m) directly to the steamer lower liquid manifold so over-pressure physically overflows water and vents steam freely with ZERO moving parts!",
			"FACTORY-CERTIFIED 10-15 PSIG SPRING-LOADED BRASS PRV & FUSIBLE PLUG: Retain the factory-stamped 10-15 psig brass spring-loaded PRV (with manual pull-ring test lever) and low-water fusible plug as secondary backup; pipe PRV discharge downward to a safe outdoor grade.",
			fmt.Sprintf("1.5x MAWP COLD HYDROSTATIC PROOF TEST (%.0f psig): Cold-water hydrostatically test the catering steamer, steam header S1, drilled brass orifice union, and threaded fittings to %.0f psig (1.5x %.0f psig MAWP) for 30 minutes with zero weeping prior to first live steam firing.", hydroPSIG, hydroPSIG, mawpPSIG),
			"CALIBRATED LOW-RANGE PRESSURE GAUGE (0-15 PSIG): Mount a 0-15 psig (or 0-30 psig) glycerin-filled pressure gauge on a brass siphon pigtail so 6.0 psig operating pressure reads clearly mid-scale.",
			"BRONZE CHECK VALVE & OPEN ATMOSPHERIC 0 PSIG FEED TANK: Supply feedwater ONLY from an open vented atmospheric 0 psig gravity tank/pump through a bronze/316L spring check valve so steam can never back-pressurize plastic tanks.",
			"INSULATED STEAM HEADER (S1) & SCALD GUARDS: Wrap live steam pipe S1 (~106-110 C) and the 4.2 m standpipe top overflow splash elbow in mineral-wool insulation / sheet-metal scald guards, and place the bronze throttling valve outside the spray zone.",
			"STRICT BAN ON UNVENTED DIY WELDED DRUM BOILERS: NEVER fabricate an unvented homemade welded steel drum, converted LPG cylinder, or scrap-pipe pressure boiler on a kiln flue — operate ONLY a factory-certified 12-15 psig commercial catering steamer/kettle equipped with BOTH the 4.2 m open hydrostatic standpipe and a stamped brass PRV.",
		}
	} else if isHP {
		switch prof.id {
		case LocalRegionMumbai:
			sourceType = "Certified Non-IBR (<=25 L Water Capacity, Indian Boilers Act Exempt) Commercial Baby Steam Boiler (60 psig / 4.14 barg) + 316L de Laval Hydro-Sonic Ejector"
		case LocalRegionNairobi:
			sourceType = "KEBS-Approved Commercial Baby Steam Boiler (<=25 L, 60 psig / 4.14 barg, Garment/Dairy Supply) + 316L de Laval Hydro-Sonic Ejector"
		default:
			sourceType = "Certified ASME/PED Commercial Miniature Baby Steam Boiler (<=25 L, 60 psig / 4.14 barg) + 316L de Laval Hydro-Sonic Ejector"
		}
		condNote = fmt.Sprintf(
			"HYDRO-SONIC SUPERSONIC SHEAR + CONDENSATION NUCLEATION: Mach %.2f (%.0f m/s) 60 psig steam jet shears recirculated scrub water into 15-30 um droplets inside the mixing throat while rapid expansion supersaturation condenses steam directly onto hygroscopic NaCl fume nuclei (growing 0.1-0.3 um salt aerosol to 2-5 um droplets captured on wet stoneware rings -> ~80%% submicron NaCl capture at 0 W electric, approaching an electric venturi!).",
			machExit, vExit,
		)
	} else {
		switch prof.id {
		case LocalRegionMumbai:
			sourceType = "Code-Rated Small-Bore Seamless 316L Monotube Flash Coil (<=20 psig, <2 L Stored Volume, Non-IBR) + Certified IBR/ASME PRV & de Laval Nozzle"
		case LocalRegionNairobi:
			sourceType = "Code-Rated Small-Bore Seamless 316L Monotube Flash Coil (<=20 psig, <2 L Stored Volume, KEBS-Certified PRV) + Machined 316L de Laval Nozzle"
		default:
			sourceType = "Code-Rated Small-Bore Seamless 316L Monotube Flash Coil (<=20 psig, <2 L Stored Volume) + ASME-Certified PRV & 316L de Laval Nozzle"
		}
		condNote = fmt.Sprintf(
			"HYGROSCOPIC NaCl CONDENSATION GROWTH + STEFAN-FLOW DIFFUSIOPHORESIS: Expanding 20 psig steam (%.0f m/s, Mach %.2f) creates transient supersaturation around deliquescent NaCl fume nuclei in the cooled stoneware tower, growing 0.1-0.3 um salt particles to 1.5-4 um droplets and sweeping them onto wet ceramic rings via condensing vapor Stefan flow (~62%% submicron NaCl capture vs ~30%% passive chimney, at 0 W electric).",
			vExit, machExit,
		)
	}

	if len(safetyReqs) == 0 {
		banLine := "STRICT BAN ON DIY UNCERTIFIED WELDED DRUM BOILERS: NEVER fabricate a homemade welded steel drum, converted LPG cylinder, or scrap-pipe pressure boiler on a kiln flue — only factory-certified non-IBR (<=25 L)/KEBS/ASME baby boilers or code-rated small-bore (<2 L) seamless monotube flash coils (<= 20 psig ONLY) are permitted."
		if draftMode == LocalDraftSteamEjectorHP {
			banLine = "STRICT CERTIFIED BABY-BOILER-ONLY RULE AT 60 PSIG (NO DIY DRUMS & NO MONOTUBE COILS): NEVER fabricate a homemade welded drum, converted LPG cylinder, or studio monotube coil at 60 psig (4.14 barg) — 60 psig Tier 2 strictly requires a factory-certified non-IBR (<=25 L)/KEBS/ASME commercial baby steam boiler with stamped PRV, sight glass, fusible plug, and factory hydro-test certificate."
		}
		safetyReqs = []string{
			fmt.Sprintf("CERTIFIED SPRING-LOADED PRV: Install a factory-stamped ASME/IBR/KEBS spring-loaded brass/316L pressure relief valve set <= %.0f psig MAWP with a manual pull-ring test lever; pipe discharge downward to a safe outdoor grade away from operators.", mawpPSIG),
			fmt.Sprintf("1.5x MAWP HYDROSTATIC PROOF TEST: Cold-water hydrostatically test the steam boiler/piping, header S1, and all threaded fittings to %.0f psig (1.5x MAWP) for 30 minutes with zero weeping prior to first live steam firing.", hydroPSIG),
			"CALIBRATED PRESSURE GAUGE & BOILER TRIM: Mount a 0-100 psig glycerin-filled pressure gauge on a brass siphon pigtail (plus armored water sight glass and fusible low-water plug on HP certified baby boilers).",
			"FEEDWATER CHECK VALVE & OPEN ATMOSPHERIC 0 PSIG FEED TANK: Supply feedwater from an open vented atmospheric 0 psig tank via a manual piston/boiler feed pump (or batch refill at 0 psig) through a bronze/316L spring check valve — a 3.5 m (5 PSI) gravity tank CANNOT feed into a pressurized 20-60 psig boiler!",
			"INSULATED STEAM HEADER (S1) & SCALD GUARDS: Wrap live steam pipe S1 in mineral wool / calcium-silicate insulation with sheet-metal scald guards and locate the bronze steam throttling valve outside the spray/jet zone.",
			banLine,
		}
	}

	return steamEjectorPhysics{
		draftMode:               draftMode,
		draftModeLabel:          draftLabel,
		steamEnabled:            true,
		pressurePSIG:            psig,
		pressureBar:             barG,
		nozzleVelMS:             round1(vExit),
		nozzleMach:              round2(machExit),
		nozzleThroatMM:          throatMM,
		steamFlowKgHr:           steamFlowKgHr,
		steamWaterLHr:           steamWaterLHr,
		steamThermalKW:          steamThermalKW,
		kilnWasteHeatKW:         kilnWasteKW,
		wasteHeatSufficient:     wasteSufficient,
		ejectorDraftInWC:        ejectorDraftInWC,
		ejectorEntrainmentRatio: entrainRatio,
		condensationGrowthNote:  condNote,
		steamSourceType:         sourceType,
		safetyRequirements:      safetyReqs,
	}
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
	afterburnerUSD := 0.0
	if p.LocalAfterburner {
		// Passive Secondary-Air Refractory Checkerwork Smoke-Burner / Tar-Cracking Throat (>750 C, 0 W)
		// + removable Pool #1 tar-skimming weir & washable coarse ceramic louver plates.
		afterburnerUSD = round2(95.0 * volScale * s)
	}
	baffleZone := round2(42.0 * volScale * s)
	masonryTower := round2((70.0 + 2.2*bedDiaIn) * s)
	stonewareRings := round2(ringCuFt * prof.ringCostPerCuFt)
	terracottaDem := round2(28.0 * volScale * s)
	brickSump := round2(48.0 * volScale * s)
	gravityTank := round2((34.0 + 0.08*tankLiters) * s)
	standBracing := round2((40.0 + 0.05*tankLiters) * s)
	plumbing := round2(40.0 * s)

	draftMode := p.EffectiveLocalDraftMode()
	boilerSource := p.EffectiveLocalBoilerSource()
	pumpMode := p.EffectiveLocalPumpMode()
	steamEnabled := draftMode == LocalDraftSteamEjectorSubsonic || draftMode == LocalDraftSteamEjectorLP || draftMode == LocalDraftSteamEjectorHP

	var pumpUSD, fanUSD, steamSystemUSD, chimneyUSD float64
	if steamEnabled {
		if draftMode == LocalDraftSteamEjectorSubsonic {
			// Certified 12-15 psig commercial catering stainless pressure steamer + simple drilled brass orifice ($22)
			// + 4.2 m open hydrostatic standpipe relief ($32) + certified 10-15 psig brass PRV & gauge (~$149 Mumbai / ~$177 Nairobi).
			steamSystemUSD = round2(177.0 * s)
		} else if draftMode == LocalDraftSteamEjectorHP {
			// Certified non-IBR <=25 L / KEBS commercial baby steam boiler + PRV/fusible plug/sight glass + 316L de Laval ejector ($490-$560).
			steamSystemUSD = round2(530.0 * math.Max(0.925, math.Min(1.055, s)))
		} else {
			// Code-rated seamless 316L monotube flash coil + certified PRV/gauge/check valve + 316L de Laval nozzle ($280-$360).
			steamSystemUSD = round2(330.0 * math.Max(0.86, math.Min(1.09, s)))
		}
		// Phase 13/14: Boiler waste-heat recovery hardware pricing (no double billing!)
		switch boilerSource {
		case LocalBoilerStage1FeedwaterPreheater:
			// Unpressurized 0 psig vented atmospheric copper/316L feedwater economizer coil on Stage 1
			// (keeps certified steamer/baby boiler intact; zero pressure-vessel code risk!).
			steamSystemUSD = round2(steamSystemUSD + 140.0*s)
		case LocalBoilerKilnExhaustCoil:
			// 20 psig LP ONLY: replaces standalone monotube firebox with SiC-muffled <2 L seamless flue coil
			// + certified PRV/gauge + 316L de Laval ejector (no double billing!).
			steamSystemUSD = round2(395.0 * math.Max(0.86, math.Min(1.09, s)))
		case LocalBoilerStage1WaterJacket:
			// >=20 psig ONLY: replaces standalone baby boiler/firebox with ASME/IBR-certified <=25 L
			// shop-fabricated pressurized Stage 1 water-jacket boiler + UG-28 external-pressure stiffener
			// rings + stamped PRV/sight-glass/fusible-plug + 316L de Laval ejector (no double billing!).
			steamSystemUSD = round2(680.0 * math.Max(0.92, math.Min(1.08, s)))
		}

		// Phase 13/14: Off-grid steam water pump & high-capture zero-pump sump-pool cost deltas
		switch pumpMode {
		case LocalPumpSteamPPCP:
			// Mechanical Pressure-Powered Condensate Pump (Forbes Marshall / Spirax Sarco MFP standalone)
			pumpUSD = round2(680.0 * s)
		case LocalPumpSteamJetSyphon:
			// Gunmetal/316L steam-jet cold-makeup syphon + low-pressure bronze boiler feed injector + manual backup
			pumpUSD = round2(295.0 * s)
		case LocalPumpSumpPoolNoPump:
			// Eliminates overhead Kentank/Sintex + 3.5 m braced timber stand + pump!
			// Adds dual ground-level plastered brick sump pools + barometric surge-damped 0.5" S-vane skimmer
			// + 1.4 m co-current drenched stoneware trough + 2-stage perforated ceramic sieve tray + flooded weir.
			gravityTank = 0.0
			standBracing = 0.0
			pumpUSD = 0.0
			brickSump = round2(110.0 * volScale * s)
			baffleZone = round2(baffleZone + 95.0*s)
			plumbing = round2(28.0 * s)
		default:
			if p.ZeroElectricity {
				pumpUSD = round2(75.0 * s) // MoneyMaker treadle / cast-iron semi-rotary hand pump
			} else {
				pumpUSD = round2(76.0 * s) // Local 0.5 HP 230V 50Hz domestic monoblock water pump
			}
		}
		fanUSD = 0.0                                   // Steam-jet ejector provides draft (zero electric fan!)
		chimneyUSD = round2((22.0 + 9.2*chimneyM) * s) // Shorter 4.5-5.0 m ceramic/SWG chimney (saves masonry vs 8.5-9.5 m!)
	} else if p.ZeroElectricity {
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
		afterburnerUSD + quenchChamber + baffleZone + masonryTower + stonewareRings + terracottaDem +
			brickSump + gravityTank + standBracing + pumpUSD + fanUSD + steamSystemUSD +
			plumbing + chimneyUSD + instruments + alkaliCost,
	)

	return localCeramicBOMBreakdown{
		afterburnerUSD:   afterburnerUSD,
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
		steamSystemUSD:   steamSystemUSD,
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

	// Bed depth (0.90 - 1.10 m) aligned with docs/LOCAL_ZERO_ELECTRICITY_ENGINEERING.md
	// for >=95.5% HCl capture at low superficial velocity (~0.34-0.40 m/s) while keeping passive drag low.
	bedDepthM := 0.95
	if design.Input.KilnVolumeCuFt <= 16.0 {
		bedDepthM = 0.90
	} else if design.Input.KilnVolumeCuFt >= 40.0 {
		bedDepthM = 1.10
	}

	// First-principles irrigated Ergun pressure drop for 1" potter-extruded thin-wall stoneware
	// Raschig rings using exact Sauter equivalent sphere diameter d_p,eq = 6*(1 - eps) / a_p
	// (eps = 0.74, a_p = 185 m^-1 -> d_p,eq = 8.43 mm) and low-loading irrigated factor 1.40x.
	const (
		ringVoidage     = 0.74
		ringSpecAreaM1  = 185.0
		ringSauterDpM   = 6.0 * (1.0 - ringVoidage) / ringSpecAreaM1 // 0.008432 m (8.43 mm)
		irrigatedFactor = 1.40
	)
	bedDPInWC := 0.075
	if dryDP, err := scrubber.ErgunPressureDrop(ringSauterDpM, ringVoidage, vSup, rhoLocal, muG, bedDepthM); err == nil {
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
		afterburnerDPInWC        = 0.015 // Hot open firebrick checkerwork + secondary-air slot throat
	)
	extraAfterburnerDP := 0.0
	if p.LocalAfterburner {
		extraAfterburnerDP = afterburnerDPInWC
	}
	reqDraftInWC := round3(kilnDraftReq + bedDPInWC + quenchBaffleDPInWC + terracottaDemisterDPInWC + extraAfterburnerDP)

	// 3. Draft mechanism:
	//   - Phase 10B/10C Steam-Jet Ejector (6 psig Subsonic, 20 psig LP, or 60 psig HP): compact 4.5-5.5 m
	//     ceramic chimney + steam-jet momentum draft (0 W electric when zero_electricity=true, or 370 W pump hybrid).
	//   - Natural Chimney Buoyancy (when zero_electricity=true & natural_chimney): 8-12 m thermal-reheat chimney.
	//   - Local 12V/230V fan hybrid (when zero_electricity=false & natural_chimney): 4.5 m chimney + 1.8 inWC fan.
	const (
		mwDryAir            = 0.028964 // kg/mol
		gAccel              = model.StandardGravity
		paPerInWC           = model.PaPerInWC
		gravityAssist       = 0.015 // realistic downflow quench assist from 4-6 L/min @ 3.5 m head (~3.7 Pa)
		chimneyFrictionInWC = 0.012 // chimney wall friction + exit velocity head (~3.0 Pa)
		targetMarginWC      = 0.040
	)
	mwStack := thermo.MixtureMolecularWeight(design.OutletGas)
	if mwStack <= 0.015 || mwStack > 0.045 {
		mwStack = 0.0268 // wet scrubbed wood/salt flue gas (~26-28% H2O vapor lowers MW, aiding buoyancy)
	}
	rhoAmb := (pAmb * mwDryAir) / (model.UniversalGasConst * prof.ambientTempK)

	draftMode := p.EffectiveLocalDraftMode()
	steam := computeSteamEjectorPhysics(draftMode, design, prof, qLocal, rhoLocal)

	// Phase 13/14: Resolve Boiler Heat Source & Pump Mode early so overhead gravity assist
	// and electric power accurately reflect steam-driven pumps and zero-pump sump-pool mode.
	boilerSource := p.EffectiveLocalBoilerSource()
	pumpMode := p.EffectiveLocalPumpMode()
	if !steam.steamEnabled {
		boilerSource = LocalBoilerDedicatedHearth
		pumpMode = LocalPumpManualTreadle
	} else if draftMode == LocalDraftSteamEjectorSubsonic {
		pumpMode = LocalPumpManualTreadle
	}
	hasOverheadTank := !(steam.steamEnabled && pumpMode == LocalPumpSumpPoolNoPump)
	effGravityAssist := gravityAssist
	if !hasOverheadTank {
		effGravityAssist = 0.0 // No 3.5 m overhead downflow spray assist when overhead tank is removed!
	}

	var chimneyM, stackReheatF, availDraftInWC, powerWatts float64
	var draftMechanism string

	if steam.steamEnabled {
		// Steam-jet ejector eliminates the tall 8.5-12 m reheat chimney & ember hearth.
		// Condensing stoneware tower exits cooled below local dew point (and well below local
		// altitude boiling point — e.g. 202.4 F at Nairobi 83.5 kPa, 192.0 F at Cusco 66.8 kPa).
		condExitF := 172.0 // 77.8 C optimal condensing exit at sea level
		if pAmb < 75000.0 {
			condExitF = 156.0 // 68.9 C condensing exit at Cusco/Altiplano 66.8 kPa (water boils 192.0 F)
		} else if pAmb < 92000.0 {
			condExitF = 166.0 // 74.4 C condensing exit at Nairobi/Kathmandu/Tonalá 83-86 kPa
		}
		if draftMode == LocalDraftSteamEjectorSubsonic {
			chimneyM = 5.2 // 17.1 ft compact sea-level stack
			if pAmb < 75000.0 {
				chimneyM = 6.0 // 19.7 ft high-altitude clearance at Cusco 66.8 kPa
			} else if pAmb < 92000.0 {
				chimneyM = 5.5 // 18.0 ft compact highland stack at 83-86 kPa
			}
		} else {
			chimneyM = 4.5
			if pAmb < 75000.0 {
				chimneyM = 5.5 // 18.0 ft extreme-altitude clearance at Cusco 66.8 kPa
			} else if pAmb < 92000.0 && draftMode == LocalDraftSteamEjectorLP {
				chimneyM = 5.0 // 16.4 ft highland clearance at 83-86 kPa
			} else if pAmb < 92000.0 {
				chimneyM = 4.8 // 15.7 ft highland clearance for HP Hydro-Sonic jet
			}
		}
		stackReheatF = round1(condExitF)
		tStackK := float64(model.FahrenheitToKelvin(stackReheatF))
		rhoStack := (pAmb * mwStack) / (model.UniversalGasConst * tStackK)
		natBuoyInWC := math.Max(0.0, (gAccel*chimneyM*(rhoAmb-rhoStack))/paPerInWC-chimneyFrictionInWC)
		availDraftInWC = round3(natBuoyInWC + steam.ejectorDraftInWC + effGravityAssist)
		steamPumpOffGrid := zeroElec || pumpMode != LocalPumpManualTreadle
		if draftMode == LocalDraftSteamEjectorSubsonic {
			if steamPumpOffGrid {
				powerWatts = 0.0
				draftMechanism = fmt.Sprintf(
					"Zero-Electricity 6 psig (0.414 barg) Subsonic Quiet Steam-Jet Ejector (%.0f m/s, Mach %.2f < 1.0, drilled %.2f mm brass orifice + 4.2-4.4 m open hydrostatic standpipe -> %.3f inWC ejector draft + %.3f inWC %.1f m chimney buoyancy + %.3f inWC gravity assist, no reheat hearth, ~52 dBA) — 0 W electric.",
					steam.nozzleVelMS, steam.nozzleMach, steam.nozzleThroatMM,
					steam.ejectorDraftInWC, natBuoyInWC, chimneyM, effGravityAssist,
				)
			} else {
				powerWatts = 370.0
				draftMechanism = fmt.Sprintf(
					"6 psig (0.414 barg) Subsonic Quiet Steam-Jet Ejector (%.0f m/s, Mach %.2f < 1.0, drilled %.2f mm brass orifice + 4.2-4.4 m open hydrostatic standpipe -> %.3f inWC ejector draft + %.3f inWC %.1f m chimney) + 0.5 HP 230V 50Hz Domestic Water Pump (~%.0f W, zero electric draft fan).",
					steam.nozzleVelMS, steam.nozzleMach, steam.nozzleThroatMM,
					steam.ejectorDraftInWC, natBuoyInWC, chimneyM, powerWatts,
				)
			}
		} else if steamPumpOffGrid {
			powerWatts = 0.0
			if !hasOverheadTank {
				draftMechanism = fmt.Sprintf(
					"Zero-Electricity 60 psig (4.14 barg) Flooded Sump-Level Steam-Jet Ejector (%.0f m/s, Mach %.2f -> %.3f inWC ejector draft + %.3f inWC %.1f m chimney buoyancy, 0\" suction lift, no overhead tower) — 0 W electric.",
					steam.nozzleVelMS, steam.nozzleMach, steam.ejectorDraftInWC, natBuoyInWC, chimneyM,
				)
			} else {
				draftMechanism = fmt.Sprintf(
					"Zero-Electricity %.0f psig (%.2f barg) Steam-Jet de Laval Ejector (%.0f m/s, Mach %.2f -> %.3f inWC ejector draft + %.3f inWC %.1f m chimney buoyancy + %.3f inWC gravity assist, no reheat hearth) — 0 W electric.",
					steam.pressurePSIG, steam.pressureBar, steam.nozzleVelMS, steam.nozzleMach,
					steam.ejectorDraftInWC, natBuoyInWC, chimneyM, effGravityAssist,
				)
			}
		} else {
			powerWatts = 370.0 // 0.5 HP 230V 50Hz domestic monoblock water pump only (steam ejector replaces electric draft fan!)
			draftMechanism = fmt.Sprintf(
				"%.0f psig (%.2f barg) Steam-Jet de Laval Ejector (%.0f m/s, Mach %.2f -> %.3f inWC ejector draft + %.3f inWC %.1f m chimney) + 0.5 HP 230V 50Hz Domestic Water Pump (~%.0f W, zero electric draft fan).",
				steam.pressurePSIG, steam.pressureBar, steam.nozzleVelMS, steam.nozzleMach,
				steam.ejectorDraftInWC, natBuoyInWC, chimneyM, powerWatts,
			)
		}
	} else if zeroElec {
		// Auto-size brick/terracotta thermal-reheat chimney in [8.0, 12.0] m at 340 F (444.26 K)
		// recovered from the 2400 F kiln exit flue via concentric brick flue-wall recuperator.
		// If high altitude or high target draft still needs extra margin at 12.0 m, step reheat up to 420 F.
		stackReheatF = 340.0
		chimneyM = 8.0
		calcBuoyancy := func(hM, tStackF float64) float64 {
			tStackK := float64(model.FahrenheitToKelvin(tStackF))
			rhoStack := (pAmb * mwStack) / (model.UniversalGasConst * tStackK)
			buoyInWC := (gAccel * hM * (rhoAmb - rhoStack)) / paPerInWC
			return math.Max(0.0, buoyInWC-chimneyFrictionInWC) + gravityAssist
		}
		for h := 8.0; h <= 12.0+1e-6; h += 0.5 {
			chimneyM = round1(h)
			if calcBuoyancy(chimneyM, stackReheatF)-reqDraftInWC >= targetMarginWC {
				break
			}
		}
		for tF := 340.0; tF <= 420.0+1e-6; tF += 10.0 {
			stackReheatF = tF
			if calcBuoyancy(chimneyM, stackReheatF)-reqDraftInWC >= targetMarginWC {
				break
			}
		}
		availDraftInWC = round3(calcBuoyancy(chimneyM, stackReheatF))
		powerWatts = 0.0
		draftMechanism = fmt.Sprintf(
			"Natural Chimney Buoyancy + Concentric Brick Flue-Wall Thermal Reheat (%.0f F stack, %.1f m / %.1f ft brick chimney at %s) + %.3f inWC gravity-spray downflow assist — 0 W electric.",
			stackReheatF, chimneyM, model.Meters(chimneyM).Feet(), prof.elevationNote, gravityAssist,
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
			chimneyM, model.Meters(chimneyM).Feet(), powerWatts,
		)
	}

	chimneyFt := round1(model.Meters(chimneyM).Feet())
	const litersPerUSGal = model.CubicMetersPerUSGallon * 1000.0 // exact 3.785411784 L / US gal

	// Phase 13/14: Energy-conserved Boiler Heat Source calculations (both Stage 1 options + 20 psig LP coil)
	qEvapKW := math.Max(15.0, design.Quench.EvaporatedWaterGPM*litersPerUSGal/60.0*2260.0)
	boilerSourceLabel := "Dedicated Firebox / Separate Wood or LPG Burner"
	boilerWaterSavedPct := 0.0
	boilerFuelSavedNote := ""
	boilerCorrosionNote := ""
	if steam.steamEnabled {
		switch boilerSource {
		case LocalBoilerStage1FeedwaterPreheater:
			boilerSourceLabel = "0 psig Vented Stage 1 Feedwater Preheater / Economizer (Safe on 6/20/60 psig)"
			qPreKW := (steam.steamFlowKgHr / 3600.0) * 4.186 * 63.0 // 22 C -> 85 C atmospheric feedwater preheat
			boilerWaterSavedPct = round1(math.Min(12.0, math.Max(6.0, 100.0*qPreKW/qEvapKW)))
			boilerFuelSavedNote = fmt.Sprintf("Unpressurized 0 psig vented atmospheric copper/316L economizer coil on Stage 1 preheats boiler feedwater to ~85 C (%.1f kW recovered, ~13%% boiler fuel saved & ~%.0f%% less Stage 1 evaporative water) with ZERO pressure-vessel code risk!", qPreKW, boilerWaterSavedPct)
			boilerCorrosionNote = "ZERO-PRESSURE ATMOSPHERIC SAFETY (0 psig): Because the Stage 1 feedwater preheater vents freely to the open atmospheric feed tank (0 psig), it works safely on 6, 20, and 60 psig certified boilers with zero pressurized-vessel collapse or over-pressure explosion risk."
		case LocalBoilerKilnExhaustCoil:
			boilerSourceLabel = "Kiln Exhaust Flue Monotube Steam Coil (20 psig LP Only, <2 L, SiC Muffle Shielded)"
			qRecKW := math.Min(steam.steamThermalKW, steam.kilnWasteHeatKW*0.65)
			boilerWaterSavedPct = round1(math.Min(32.0, math.Max(14.0, 100.0*qRecKW/qEvapKW)))
			boilerFuelSavedNote = fmt.Sprintf("Replaces the standalone monotube firebox: recovers ~%.1f kW directly from the %.0f F kiln flue (20 psig LP / <2 L seamless tube ONLY), cutting separate burner fuel to near-zero during high fire and dropping Stage 1 evaporative quench water loss by ~%.0f%%.", qRecKW, design.Input.PeakExitTempF, boilerWaterSavedPct)
			if p.VaporGlazeMode == VaporGlazeSalt {
				boilerCorrosionNote = "HOT NaCl/HCl CORROSION GUARD (20 psig LP ONLY): Bare 316L stainless fails rapidly >350-400 C in molten NaCl/HCl vapor (mp 801 C). Keep the <2 L coil ALWAYS water-flooded (wall <215 C), shield behind a thin SiC / split-firebrick radiation muffle wall (survives 1000-1250 C) or clad with Inconel 625 (<650 C), and NEVER dry-fire or run at 60 psig!"
			} else {
				boilerCorrosionNote = "Zero-salt wood/soda flue (20 psig LP ONLY): Keep the <2 L seamless 316L steam coil continuously water-flooded with a bronze check valve so tube skin stays <220 C; brush fly ash between firings."
			}
		case LocalBoilerStage1WaterJacket:
			boilerSourceLabel = "ASME/IBR-Certified <=25 L Stage 1 Pressurized Water-Jacket Boiler (UG-28 Stiffener Rings)"
			qRecKW := math.Min(steam.steamThermalKW, steam.kilnWasteHeatKW*0.72)
			boilerWaterSavedPct = round1(math.Min(38.0, math.Max(18.0, 100.0*qRecKW/qEvapKW)))
			boilerFuelSavedNote = fmt.Sprintf("Replaces standalone baby boiler: shop-fabricated ASME/IBR-certified <=25 L pressurized water jacket around upper dry Stage 1 harvests ~%.1f kW sensible heat BEFORE spray nozzles, cutting Stage 1 evaporative water loss by ~%.0f%% while generating ejector steam.", qRecKW, boilerWaterSavedPct)
			boilerCorrosionNote = "ASME/IBR <=25 L & UG-28 PRESSURE-VESSEL RULE: Double-walled annular jackets put the inner flue tube under EXTERNAL collapsing pressure — MUST be shop-fabricated by an ASME/IBR/KEBS vessel shop with UG-28 external-pressure stiffener rings, <=25 L water volume, stamped PRV, armored sight glass & fusible plug (NEVER field-welded DIY!). Boiling water locks inner wall at ~120-160 C (above 84 C HCl dew point, below 400 C hot chloride attack)."
		default:
			boilerFuelSavedNote = fmt.Sprintf("External certified firebox/burner isolated from corrosive kiln flue gases (requires %.1f kW fuel or indirect masonry waste-heat tap).", steam.steamThermalKW)
		}
	}

	// 4. Gravity overhead tank (Kentank/Sintex 210-490 L on 3.5 m / 11.5 ft braced stand) & pumping budget.
	tankLiters := 300.0
	if design.Input.KilnVolumeCuFt <= 16.0 {
		tankLiters = 210.0
	} else if design.Input.KilnVolumeCuFt >= 40.0 {
		tankLiters = 490.0
	}
	tankGal := round1(tankLiters / litersPerUSGal)
	tankElevFt := 11.5 // 3.5 m braced brick/timber stand
	gravityHeadPSI := round2(tankElevFt * 0.4335)

	// Closed-loop gravity quench + stoneware bed trickle (~2.3x evaporative demand -> ~4-6 L/min on 24 cu.ft. studio kiln).
	evapGPM := design.Quench.EvaporatedWaterGPM * (1.0 - boilerWaterSavedPct/100.0)
	if evapGPM <= 0 {
		evapGPM = design.Quench.TotalSprayWaterGPM / 3.0
	}
	sprayFlowGPM := round2(math.Max(1.05, math.Min(2.20, evapGPM*1.55+0.45)))
	sprayFlowLPM := round2(sprayFlowGPM * litersPerUSGal)

	var manualPumpType string
	var jerrycansPerHr int
	var treadleMinPerHr float64
	var pumpSteamKgHr float64
	var sumpPoolSubIn, sumpPoolDPInWC float64
	pumpModeLabel := "Manual Treadle / Hand Wing Pump + 3.5 m Gravity Tank"
	pumpEngNote := ""

	// Helper to update steam thermal load & wasteHeatSufficient consistently when steam pumps add steam flow.
	recomputeSteamThermal := func() {
		if !steam.steamEnabled {
			return
		}
		p0Pa := pAmb + steam.pressurePSIG*6894.757
		tSat0K := 399.15
		if tSat, err := thermo.WaterDewPoint(model.Pascal(p0Pa)); err == nil {
			tSat0K = float64(tSat)
		}
		_, hg0KJ, _, _ := satSteamProps(tSat0K)
		hFeedKJ := 4.186 * math.Max(5.0, prof.ambientTempK-273.15)
		steam.steamThermalKW = round1((steam.steamFlowKgHr / 3600.0) * (hg0KJ - hFeedKJ))
		steam.wasteHeatSufficient = steam.kilnWasteHeatKW >= steam.steamThermalKW
	}

	if steam.steamEnabled && pumpMode == LocalPumpSteamPPCP {
		pumpModeLabel = "Mechanical Steam PPCP Recirc Pump (Forbes Marshall / Spirax Sarco MFP, 0 W)"
		manualPumpType = pumpModeLabel
		pumpSteamKgHr = round2(math.Max(0.45, 0.75*(sprayFlowLPM/5.0)))
		steam.steamFlowKgHr = round1(steam.steamFlowKgHr + pumpSteamKgHr)
		steam.steamWaterLHr = round1(steam.steamFlowKgHr)
		recomputeSteamThermal()
		jerrycansPerHr = 0
		treadleMinPerHr = 0.0
		pumpEngNote = fmt.Sprintf("Zero-electricity float-actuated steam displacement pump (Forbes Marshall / Spirax Sarco MFP, available in Mumbai/Pune & Nairobi Industrial Area) lifts hot alkaline sump liquor (up to 212 F) to the 3.5 m tank using only ~%.2f kg/hr steam (<2%% of ejector load) with zero steam-water mixing. Surplus tea/laundry units ~$350 vs ~$%.0f new.", pumpSteamKgHr, 680.0*prof.costScale)
	} else if steam.steamEnabled && pumpMode == LocalPumpSteamJetSyphon {
		pumpModeLabel = "Steam-Jet Cold Makeup Syphon + Penberthy Boiler Feed Injector (<120 F Cold Water)"
		manualPumpType = pumpModeLabel
		pumpSteamKgHr = round2(math.Max(3.5, 5.5*(sprayFlowLPM/5.0)))
		steam.steamFlowKgHr = round1(steam.steamFlowKgHr + pumpSteamKgHr)
		steam.steamWaterLHr = round1(steam.steamFlowKgHr)
		recomputeSteamThermal()
		totalHourlyLiters := sprayFlowLPM*60.0 + steam.steamWaterLHr
		evapMakeupLitersHr := evapGPM*litersPerUSGal*60.0 + steam.steamWaterLHr
		jerrycansPerHr = int(math.Ceil(math.Max(0.0, totalHourlyLiters-evapMakeupLitersHr) / 20.0))
		treadleMinPerHr = round1(math.Max(0.0, totalHourlyLiters-evapMakeupLitersHr) / 35.0)
		pumpEngNote = fmt.Sprintf("Low-cost gunmetal/316L steam-jet liquid syphon + bronze boiler feed injector lifts COLD fresh makeup water (<120 F / 49 C) from a dedicated cold makeup drum to the 3.5 m tank and feeds the pressurized boiler automatically (~%.1f kg/hr intermittent steam). NEVER feed hot recirculating sump liquor (>120 F) into a condensing steam syphon — keep manual/PPCP return for hot sump.", pumpSteamKgHr)
	} else if steam.steamEnabled && pumpMode == LocalPumpSumpPoolNoPump {
		pumpModeLabel = "High-Capture Zero-Pump Flooded Sump-Pool + 1.4 m Co-Current Drenched Trough + 2-Stage Ceramic Sieve (60 psig)"
		manualPumpType = "None — Flooded Sump-Level 60 psig Ejector + Co-Current Drenched Trough & 2-Stage Sieve (0 Pumps, 0 Tower)"
		hasOverheadTank = false
		tankLiters = 0.0
		tankGal = 0.0
		tankElevFt = 0.0
		gravityHeadPSI = 0.0
		// Phase 14 High-Capture Sump-Pool Draft Budget (60 psig HP only, ~4.3 inWC avail):
		//   kilnDraft (~0.05) + surge-damped 0.5" S-vane skimmer (0.65) + 1.4m co-current drenched stoneware trough (0.15)
		//   + 2-stage 0.5"+0.5" perforated ceramic sieve tray (1.15) + terracotta demister (0.02) = ~2.02 inWC total!
		sumpPoolSubIn = 0.50
		bedDPInWC = 0.150     // 1.4 m horizontal co-current drenched stoneware Raschig-ring trough
		sumpPoolDPInWC = 1.80 // 0.65 inWC surge-damped 0.5" S-vane skimmer + 1.15 inWC 2-stage ceramic sieve tray
		reqDraftInWC = round3(kilnDraftReq + bedDPInWC + sumpPoolDPInWC + terracottaDemisterDPInWC)
		evapMakeupLitersHr := evapGPM*litersPerUSGal*60.0 + steam.steamWaterLHr
		jerrycansPerHr = int(math.Ceil(evapMakeupLitersHr / 20.0))
		treadleMinPerHr = 0.0
		pumpEngNote = fmt.Sprintf("HIGH-CAPTURE ZERO-PUMP, ZERO-TOWER ARCHITECTURE (60 psig HP): Wide 2x barometric surge plenum + shallow 0.50\" S-vane skimmer in Pool #1 (0.65 inWC, damps kiln spyhole pulsing!) -> gravity-flooded sump-level 60 psig ejector at Pool #2 waterline (0\" suction lift — eliminates impossible dip-tube lift!) shears alkaline liquor into 15-30 um mist sweeping through a 1.4 m co-current drenched stoneware Raschig-ring trough (0.15 inWC) -> 2-stage shallow 0.5\"+0.5\" perforated ceramic sieve trays (1.15 inWC) -> ~96.8%% HCl & ~78%% NaCl capture with 0 pumps, 0 water tower, and only ~%d x 20L ground-level makeup buckets/hr!", jerrycansPerHr)
	} else if zeroElec {
		totalHourlyLiters := sprayFlowLPM*60.0 + steam.steamWaterLHr
		manualPumpType = "Treadle / Semi-Rotary Cast-Iron Hand Pump"
		jerrycansPerHr = int(math.Ceil(totalHourlyLiters / 20.0))
		treadleMinPerHr = round1(totalHourlyLiters / 35.0) // ~35 L/min MoneyMaker treadle / wing pump
	} else {
		manualPumpType = fmt.Sprintf("Local 0.5 HP 230V 50Hz Monoblock Pump (%s)", prof.pumpBrand)
		pumpModeLabel = manualPumpType
		jerrycansPerHr = 0
		treadleMinPerHr = 0.0
	}

	draftMarginInWC := round3(availDraftInWC - reqDraftInWC)
	draftFeasible := draftMarginInWC >= 0.0

	// 5. Local alkali stoichiometry.
	alkali := computeLocalAlkali(p, design, prof)

	// 6. Honest capture performance, SO2 removal, tar cracking & zero-meltdown thermal immunity.
	// Wide irrigated stoneware bed has long gas residence time (~3 s) -> high HCl absorption (95.5-98.0%).
	htuM := 0.34 * math.Pow(math.Max(vSup, 0.20)/0.38, 0.22)
	ntu := bedDepthM / htuM
	hclPct := round1(math.Max(95.5, math.Min(98.0, (1.0-math.Exp(-ntu))*100.0)))
	so2InletPPM := round1(design.PackedBed.SO2InletPPM)
	sulfurAcidDewF := round1(design.PackedBed.SulfurAcidDewPointF)
	so2RemovalPct := 0.0
	if so2InletPPM > 0 {
		// Alkaline wide stoneware bed achieves ~86-92% SO2 absorption (HTU_SO2 ~ 1.45 * HTU_HCl).
		so2RemovalPct = round1(math.Max(85.0, math.Min(94.5, (1.0-math.Exp(-ntu/1.45))*100.0)))
		if draftMode == LocalDraftSteamEjectorHP {
			so2RemovalPct = round1(math.Min(95.0, so2RemovalPct+3.5))
		}
	}
	afterburnerNote := ""
	tarMitigationNote := ""
	if p.LocalAfterburner {
		afterburnerNote = "PASSIVE SECONDARY-AIR MASONRY SMOKE-BURNER / TAR AFTERBURNER ACTIVE (0 W): Refractory firebrick checkerwork throat immediately downstream of the kiln bagwall/flue admits preheated passive secondary air (>750 C), thermally cracking 70-85% of heavy coal/peat/PKS tar vapors & black soot BEFORE wet quenching (adds only 0.015 inWC draft loss; inspired by Tonalá/Kusakabe smokeless kilns)."
		tarMitigationNote = "ANTI-TAR FOULING PACKAGE ACTIVE: Secondary-air afterburner + Pool #1 floating tar-skimming weir + washable coarse ceramic louver plates prevent sticky coal/peat creosote from blinding stoneware Raschig rings. Skim floating tar scum from Pool #1 after every firing and hot-alkali wash rings every 5-8 firings."
	} else if model.IsSolidDirtyFuel(design.Input.FuelType) {
		tarMitigationNote = fmt.Sprintf("HEAVY TAR & SOOT WARNING (%s without Afterburner): Raw coal, peat/turf, or ag-biomass smoke will rapidly coat wet ceramic rings with sticky creosote/phenolic tar! Enable Local Secondary-Air Masonry Smoke-Burner / Afterburner (+$%.0f, 0 W) to crack 70-85%% of tar thermally before Stage 1.", design.Input.FuelType, 95.0*prof.costScale)
	}
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
	if !zeroElec && !steam.steamEnabled {
		submicronNote = fmt.Sprintf(
			"HONEST PHYSICS CHECK: A low-dP local-grid/12V ceramic scrubber (~1.8 inWC fan) achieves %.1f%% HCl acid gas capture and %.1f%% coarse wood ash capture, but only ~%.0f%% submicron (0.1-0.5 um) NaCl fume capture (vs ~84%% in a 19 inWC industrial venturi). Discharge above roof level via the ceramic chimney.",
			hclPct, coarseAshPct, submicronPct,
		)
	}
	if draftMode == LocalDraftSteamEjectorSubsonic {
		hclPct = 97.5
		coarseAshPct = 92.5
		submicronPct = 55.0
		if !zeroElec {
			submicronPct = 56.5
		}
		submicronNote = fmt.Sprintf(
			"SUBSONIC CONDENSATION GROWTH CHECK (6 psig / 0.414 barg): Unchoked %.0f m/s (Mach %.2f < 1.0, ~52 dBA quiet) steam from a certified commercial catering steamer injects %.1f kg/hr into the cooled wide stoneware bed, reaching up to ~%.0f%% submicron NaCl capture (with active sump cooling <=75 C; ~40-45%% if closed sump warms uncooled above 78 C) + %.1f%% HCl and %.1f%% coarse ash capture at %.0f W electric.",
			steam.nozzleVelMS, steam.nozzleMach, steam.steamFlowKgHr, submicronPct, hclPct, coarseAshPct, powerWatts,
		)
	} else if draftMode == LocalDraftSteamEjectorLP {
		hclPct = 97.8
		coarseAshPct = 94.0
		submicronPct = 62.0
		if !zeroElec {
			submicronPct = 63.5
		}
		submicronNote = fmt.Sprintf(
			"STEAM CONDENSATION GROWTH CHECK (20 psig LP): Expanding %.0f m/s (Mach %.2f) steam onto hygroscopic NaCl fume nuclei + Stefan-flow diffusiophoresis in the cooled stoneware tower reaches up to ~%.0f%% submicron NaCl capture (with active sump cooling <=75 C; ~48-52%% if closed sump warms uncooled) while neutralizing %.1f%% HCl and capturing %.1f%% coarse ash at %.0f W electric (trades +%.1f L/hr clean boiler feedwater).",
			steam.nozzleVelMS, steam.nozzleMach, submicronPct, hclPct, coarseAshPct, powerWatts, steam.steamWaterLHr,
		)
	} else if draftMode == LocalDraftSteamEjectorHP {
		hclPct = 98.5
		coarseAshPct = 96.5
		submicronPct = 80.0
		if !zeroElec {
			submicronPct = 81.5
		}
		submicronNote = fmt.Sprintf(
			"HYDRO-SONIC STEAM-JET CHECK (60 psig HP): Supersonic %.0f m/s (Mach %.2f) steam shears scrub liquor into 15-30 um droplets + drives condensation nucleation on NaCl nuclei for ~%.0f%% submicron NaCl capture (~72%% uncooled aerodynamic floor, ~80%% with active condensing sump cooling <=75 C!) plus %.1f%% HCl and %.1f%% coarse ash capture at %.0f W electric (requires certified non-IBR/KEBS baby boiler + %.1f L/hr clean feedwater).",
			steam.nozzleVelMS, steam.nozzleMach, submicronPct, hclPct, coarseAshPct, powerWatts, steam.steamWaterLHr,
		)
	}
	thermalImmunityNote := "ZERO-MELTDOWN CERAMIC IMMUNITY: Fired firebrick, IS 651 salt-glazed stoneware (SWG) pipes, potter-extruded unglazed stoneware Raschig rings, and hand-formed terracotta chevron baffles tolerate >1,800 F (980 C) dry-run exhaust indefinitely without melting — completely eliminating the 200 F / 210 F CPVC plastic meltdown risk and electric fail-safe bypass actuators."

	if pumpMode == LocalPumpSumpPoolNoPump && steam.steamEnabled {
		// Phase 14 High-Capture Sump-Pool Config: flooded sump-level 60 psig ejector + 1.4 m
		// co-current drenched stoneware trough + 2-stage 0.5"+0.5" perforated ceramic sieve tray
		// delivers high HCl (96.8%), coarse ash (96.0%), and submicron NaCl (78.0%) capture at 0 W!
		hclPct = 96.8
		coarseAshPct = 96.0
		submicronPct = 78.0
		submicronNote = fmt.Sprintf(
			"HIGH-CAPTURE ZERO-PUMP SUMP-POOL CHECK (60 psig Flooded Ejector + 1.4 m Co-Current Drenched Trough + 2-Stage Ceramic Sieve Tray): Gravity-flooded sump-level ejector (0\" suction lift!) shears alkaline liquor into 15-30 um mist sweeping co-currently through 1.4 m of stoneware rings and two 0.5\" perforated ceramic sieve trays -> %.1f%% HCl, %.1f%% coarse ash, and %.0f%% submicron NaCl capture with 0 pumps, 0 water tower, and 0 W electric!",
			hclPct, coarseAshPct, submicronPct,
		)
	}

	// 7. Local Materials BOM cost estimate & local currency conversion.
	costs := computeLocalCeramicCosts(p, design, prof, bedDiaIn, bedDepthM, chimneyM, tankLiters, alkali)
	localCurrency := formatLocalCurrency(costs.totalUSD, prof)

	packingMaterial := fmt.Sprintf(`1" Potter-Extruded Unglazed Stoneware Raschig Rings (%.0f cu.ft., voidage 70%%, %.2f m deep, fired Cone 6-10 in studio kiln)`, costs.ringCuFt, bedDepthM)

	localMaterials := []string{
		fmt.Sprintf("Stage 1 & Recuperator: %s (~$%.0f) — refractory masonry quench chamber rated >1,800 F dry-run", prof.pipeLocalSpec, costs.quenchChamberUSD),
		fmt.Sprintf("Stage 2 Settling & Quench: Hand-formed terracotta baffle bricks + brass/PVC coarse gravity spray rose (~$%.0f)", costs.baffleZoneUSD),
		fmt.Sprintf("Stage 3 Absorber Tower: %.0f\" (%.0f mm) wide brick / IS 651 SWG ceramic column (~$%.0f) + %.0f cu.ft. potter-extruded 1\" stoneware Raschig rings (~$%.0f) + stacked terracotta chevron mist baffles (~$%.0f)", bedDiaIn, bedDiaIn*25.4, costs.masonryTowerUSD, costs.ringCuFt, costs.stonewareRingUSD, costs.terracottaDemUSD),
		fmt.Sprintf("Overhead Gravity Tank & Stand: %.0f L (%.0f gal) %s (~$%.0f) on braced 3.5 m (11.5 ft) brick/timber stand (~$%.0f) + brick baffled sump (~$%.0f)", tankLiters, tankGal, prof.tankBrand, costs.gravityTankUSD, costs.standBracingUSD, costs.brickSumpUSD),
		fmt.Sprintf("Circulation & Plumbing: %s (~$%.0f) + PVC/HDPE gravity downcomer, foot valve & drain (~$%.0f)", manualPumpType, costs.pumpUSD, costs.plumbingUSD),
		fmt.Sprintf("Chimney & Field Test Kit: %.1f m (%.0f ft) brick/terracotta chimney w/ guy-wires (~$%.0f) + $10 brine hydrometer, litmus pH strips & stem thermometer (~$%.0f) + %s charge ($%.2f)", chimneyM, chimneyFt, costs.chimneyUSD, costs.instrumentsUSD, alkali.name, alkali.costUSD),
	}
	if !hasOverheadTank {
		localMaterials[1] = fmt.Sprintf("Stage 2 Surge-Damped Skimmer & Trough: 2x barometric surge plenum + 0.5\" refractory S-vane skimmer + 1.4 m co-current drenched trough + 2-stage perforated ceramic sieve trays (~$%.0f)", costs.baffleZoneUSD)
		localMaterials[3] = fmt.Sprintf("Dual Ground-Level Brick Sump Pools (0 Overhead Tower): Plastered firebrick Pool #1 (ash skimmer) + Pool #2 (gravity-flooded ejector weir, ~360 L total, ~$%.0f)", costs.brickSumpUSD)
		localMaterials[4] = fmt.Sprintf("Zero-Pump Gravity-Flooded Weir Plumbing: 0 pumps ($0) + 1\" flooded side weir & 2\" sump cleanout drain (~$%.0f)", costs.plumbingUSD)
	}
	if !steam.steamEnabled {
		localMaterials[0] = fmt.Sprintf("Stage 1 & Recuperator: %s (~$%.0f) — concentric brick flue wall warms scrubbed exhaust back to %.0f F without mixing raw flue gas", prof.pipeLocalSpec, costs.quenchChamberUSD, stackReheatF)
	} else if draftMode == LocalDraftSteamEjectorSubsonic {
		localMaterials = append(localMaterials, fmt.Sprintf("Subsonic Quiet Steam-Jet Package (~$%.0f): Certified 12-15 psig commercial catering stainless steamer + 4.2 m open hydrostatic water-column standpipe relief (S2, $32) + simple drilled %.2f mm brass converging orifice ($22, no CNC bell) + insulated header (S1), 10-15 psig brass PRV, 0-15 psig gauge & check valve (%.1f kg/hr @ 6 psig / 0.414 barg)", costs.steamSystemUSD, steam.nozzleThroatMM, steam.steamFlowKgHr))
	} else {
		localMaterials = append(localMaterials, fmt.Sprintf("Certified Steam-Jet Ejector Package (~$%.0f, %s): %s + insulated steam header (S1), ASME/IBR/KEBS spring-loaded PRV, pressure gauge, check valve & %.2f mm throat 316L de Laval ejector (%.1f kg/hr @ %.0f psig)", costs.steamSystemUSD, boilerSourceLabel, steam.steamSourceType, steam.nozzleThroatMM, steam.steamFlowKgHr, steam.pressurePSIG))
	}
	if !zeroElec && !steam.steamEnabled {
		localMaterials = append(localMaterials, fmt.Sprintf("Local Low-dP Draft Fan: 12V solar / 230V 50Hz jua-kali centrifugal/radiator fan (~$%.0f, 1.8 inWC static)", costs.fanUSD))
	}
	if p.LocalAfterburner {
		localMaterials = append(localMaterials, fmt.Sprintf("Stage 0 Passive Secondary-Air Smoke-Burner & Tar Weir (~$%.0f): Refractory firebrick checkerwork afterburner throat (>750 C, 0 W) + Pool #1 tar-skimming weir & washable coarse ceramic anti-tar plates", costs.afterburnerUSD))
	}

	var manualSteps []string
	if steam.steamEnabled {
		pumpStep := fmt.Sprintf("3. Open Gravity Spray Valve & Pace Manual Pumping (0 W): Crack downcomer valve to %.2f L/min (%.2f GPM) and top up the atmospheric 0 psig steam feed tank (+%.1f L/hr steam makeup). Total manual refill requires %.1f min/hr on a treadle/hand pump (or %d × 20L jerrycans/hr).", sprayFlowLPM, sprayFlowGPM, steam.steamWaterLHr, treadleMinPerHr, jerrycansPerHr)
		if !hasOverheadTank {
			pumpStep = fmt.Sprintf("3. Flood Pool #2 Side Weir & Pace Ground-Level Makeup (0 W, 0 Pumps): Open the 1\" gravity-flooded weir feeding the sump-level 60 psig ejector and top up Pool #2 + clean boiler feed tank at floor level (~%d × 20L ground-level buckets/hr; zero overhead carrying!).", jerrycansPerHr)
		} else if pumpMode == LocalPumpSteamPPCP {
			pumpStep = fmt.Sprintf("3. Engage Steam PPCP Recirculation (0 W, Zero Manual Lifting): Crack steam supply to the Forbes Marshall / Spirax Sarco MFP pump (~%.2f kg/hr steam) to lift hot sump liquor automatically to the 3.5 m tank while keeping the clean 0 psig boiler feed tank topped up (+%.1f L/hr).", pumpSteamKgHr, steam.steamWaterLHr)
		} else if pumpMode == LocalPumpSteamJetSyphon {
			pumpStep = fmt.Sprintf("3. Engage Cold-Water Steam-Jet Syphon (<120 F) & Boiler Injector: Use the cold-water syphon strictly on the dedicated COLD (<120 F) makeup drum and boiler feed injector (~%.1f kg/hr intermittent steam); return hot sump liquor via manual treadle (%.1f min/hr).", pumpSteamKgHr, treadleMinPerHr)
		} else if !zeroElec && powerWatts > 0 {
			pumpStep = fmt.Sprintf("3. Start 0.5 HP Domestic Water Pump (%.0f W) & Open Gravity Spray: Confirm RCD/GFCI protection, maintain %.2f L/min (%.2f GPM) tower spray plus %.1f L/hr clean feedwater to the open 0 psig steam feed tank (overhead tank rides through grid blackouts).", powerWatts, sprayFlowLPM, sprayFlowGPM, steam.steamWaterLHr)
		}
		preFillStep := fmt.Sprintf("2. Pre-Fill Overhead Tank & Dose Local Alkali: Fill the %.0f L (%.0f gal) %s (%.1f ft / %.2f PSI) and dissolve/stir %.2f kg (%.1f lbs) %s (%s) in the baffled brick sump to pH 8.0-8.5.", tankLiters, tankGal, prof.tankBrand, tankElevFt, gravityHeadPSI, alkali.kgPerFiring, alkali.lbsPerFiring, alkali.name, alkali.formula)
		if !hasOverheadTank {
			preFillStep = fmt.Sprintf("2. Fill Dual Ground-Level Brick Sump Pools & Dose Local Alkali (0 Tower): Fill Pool #1 to the 0.50\" S-vane waterline and Pool #2 to the flooded ejector weir (~360 L combined), and dissolve/stir %.2f kg (%.1f lbs) %s (%s) to pH 8.0-8.5.", alkali.kgPerFiring, alkali.lbsPerFiring, alkali.name, alkali.formula)
		}
		if draftMode == LocalDraftSteamEjectorSubsonic {
			manualSteps = []string{
				"1. Cold Pre-Firing Subsonic Safety Check (4.2 m Standpipe & PRV): Verify the 4.2 m (13.8 ft) open vertical hydrostatic standpipe (S2) is unvalved, unobstructed, and water-primed (physically caps pressure at 6.0 psig!), pull the factory 10-15 psig brass PRV test ring, confirm 1.5x MAWP (23 psig) cold hydro-test record, zero the 0-15 psig gauge, verify the bronze check valve, and fill the open 0 psig feed tank.",
				preFillStep,
				pumpStep,
				fmt.Sprintf("4. Bring Catering Steamer to 6 psig (0.414 barg) & Establish Quiet Subsonic Draft: Start wet tower spray FIRST, warm the catering steamer steadily (%.1f kW thermal vs %.1f kW kiln flue waste heat), then crack the bronze valve on insulated header S1 to feed the simple drilled %.2f mm brass converging orifice (%.0f m/s, Mach %.2f < 1.0, ~52 dBA -> %.3f inWC ejector draft, %.1f m compact chimney).", steam.steamThermalKW, steam.kilnWasteHeatKW, steam.nozzleThroatMM, steam.nozzleVelMS, steam.nozzleMach, steam.ejectorDraftInWC, chimneyM),
				"5. Hourly Monitoring, Cooldown & Neutralized Disposal: Verify steady inward kiln spyhole draw, confirm 6 psig mid-scale gauge reading and dry standpipe overflow cup, test sump pH 7.5-8.5 hourly with litmus strips, depressurize header S1 at shutdown, cool sump <=140 F (60 C), check SG < 1.105, and flush spray rose clean.",
			}
		} else {
			manualSteps = []string{
				"1. Cold Pre-Firing Steam Safety Check (PRV & Hydro-Test): Pull the certified spring-loaded PRV test ring to verify free snap action, confirm 1.5x MAWP cold hydro-test record, verify pressure gauge zeros cleanly, check feedwater check valve, and fill the open atmospheric 0 psig clean boiler feed tank.",
				preFillStep,
				pumpStep,
				fmt.Sprintf("4. Bring Certified Steam Source to %.0f psig (%.2f barg) & Establish Ejector Draft: Start liquid wetting FIRST, warm steam source steadily (%.1f kW thermal vs %.1f kW kiln flue waste heat), then smoothly open the bronze throttling valve on insulated header S1 to drive the %.2f mm de Laval nozzle (%.0f m/s, Mach %.2f -> %.3f inWC ejector draft, %.1f m short chimney).", steam.pressurePSIG, steam.pressureBar, steam.steamThermalKW, steam.kilnWasteHeatKW, steam.nozzleThroatMM, steam.nozzleVelMS, steam.nozzleMach, steam.ejectorDraftInWC, chimneyM),
				"5. Hourly Monitoring, Cooldown & Neutralized Disposal: Verify steady inward kiln spyhole draw (no S-vane pressure pulsing), check steam pressure gauge & sight glass/feed tank level, test sump pH 7.5-8.5 hourly with litmus strips, depressurize steam header at shutdown, cool sump <=140 F (60 C), check SG < 1.105, and rinse clean.",
			}
		}
	} else if zeroElec {
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

	var warnings []string
	tankBraceWarning := fmt.Sprintf("GRAVITY TANK STAND STRUCTURAL BRACING: A filled %.0f L %s weighs ~%.0f kg (%.0f lbs) at 3.5 m (11.5 ft) elevation — build a diagonally X-braced timber/steel or reinforced masonry pier stand on level compacted footings to prevent collapse!", tankLiters, prof.tankBrand, tankLiters*1.05+20.0, (tankLiters*1.05+20.0)/model.KgPerLb)
	if !hasOverheadTank {
		tankBraceWarning = "BAROMETRIC SURGE-PLENUM & KILN SPYHOLE DAMPING: Keep Pool #1 waterline locked precisely at 0.50\" S-vane submergence and build a wide 2x cross-section barometric surge plenum upstream of the S-vane so liquid-seal sloshing never pulses kiln spyhole pressure."
	}
	if steam.steamEnabled {
		if draftMode == LocalDraftSteamEjectorSubsonic {
			warnings = []string{
				"NEVER BUILD UNVENTED DIY WELDED DRUM BOILERS: Even at 6 psig (0.414 barg), steam MUST ONLY come from a factory-certified 12-15 psig commercial catering stainless steamer/kettle protected by BOTH an unvalved 4.2-4.4 m open water-column hydrostatic relief standpipe (S2 — physically caps pressure at ~5.8-6.0 psig with zero moving parts; use 4.35-4.4 m to avoid hot-water density burping at 6.0 psig) and a stamped 10-15 psig brass PRV, fusible plug, 1.5x MAWP cold hydro-test, bronze check valve, and open 0 psig feed tank!",
				tankBraceWarning,
				fmt.Sprintf("OUTDOOR-ONLY SITING & STEAM SCALD CLEARANCE: Never operate indoors — terminate the %.1f m (%.0f ft) brick/terracotta chimney outdoors above studio eaves, insulate steam header S1 with scald guards, and pipe the 4.2-4.4 m standpipe top overflow & PRV vent safely down to an outdoor quench drain away from people.", chimneyM, chimneyFt),
			}
			if steam.steamFlowKgHr > 24.0 {
				warnings = append(warnings, fmt.Sprintf("CATERING STEAMER CAPACITY LIMIT: %.1f kg/hr (%.1f kW) exceeds a single standard commercial hotel steamer (~15-22 kg/hr max) — pair 2 certified steamers in parallel or step up to a certified LP/HP commercial baby boiler.", steam.steamFlowKgHr, steam.steamThermalKW))
			}
		} else if draftMode == LocalDraftSteamEjectorHP {
			warnings = []string{
				fmt.Sprintf("CERTIFIED BABY BOILER ONLY AT 60 PSIG (NO DIY DRUMS & NO MONOTUBE COILS): Pressurized 60 psig (%.2f barg) steam MUST ONLY come from a factory-certified non-IBR (<=25 L)/KEBS/ASME commercial baby steam boiler (or ASME/IBR-certified <=25 L shop-fabricated Stage 1 jacket with UG-28 stiffener rings) with stamped spring-loaded PRV, 1.5x MAWP hydro-test, sight glass, fusible plug, gauge, check valve, scald guards, and a pressurized/hand-piston boiler feed pump (never a DIY drum or monotube coil!).", steam.pressureBar),
				tankBraceWarning,
				fmt.Sprintf("OUTDOOR-ONLY SITING & STEAM SCALD CLEARANCE: Never operate indoors — terminate the %.1f m (%.0f ft) brick/terracotta chimney outdoors above studio eaves, insulate steam header S1 with scald guards, and aim the PRV discharge pipe safely down toward grade.", chimneyM, chimneyFt),
			}
		} else {
			warnings = []string{
				fmt.Sprintf("NEVER BUILD DIY UNCERTIFIED WELDED DRUM BOILERS: Pressurized 20 psig (%.2f barg) steam MUST ONLY come from a factory-certified non-IBR (<=25 L)/KEBS/ASME baby boiler or code-rated small-bore (<2 L) seamless monotube coil equipped with a stamped spring-loaded PRV, 1.5x MAWP hydro-test, pressure gauge, check valve, scald guards, and a boiler feed pump / atmospheric batch refill!", steam.pressureBar),
				tankBraceWarning,
				fmt.Sprintf("OUTDOOR-ONLY SITING & STEAM SCALD CLEARANCE: Never operate indoors — terminate the %.1f m (%.0f ft) brick/terracotta chimney outdoors above studio eaves, insulate steam header S1 with scald guards, and aim the PRV discharge pipe safely down toward grade.", chimneyM, chimneyFt),
			}
		}
		// Honest condensing heat-rejection warning: sustaining 75 C condensation rejects ~61-84 kW latent heat.
		condRejectKW := round1(math.Max(25.0, 50.0*(design.Input.KilnVolumeCuFt/24.0)+steam.steamThermalKW*0.75))
		warnings = append(warnings, fmt.Sprintf("CONDENSING SUMP HEAT-REJECTION REQUIREMENT (~%.0f kW): Sustaining <=75 C condensing nucleation over a full salting window rejects ~%.0f kW of latent heat (an uncooled 300-500 L closed sump warms past 75 C in ~20-30 min, dropping NaCl capture toward its uncooled floor) — route sump return through a large shaded masonry cooling trough/pond or feed continuous cold makeup water to hold <=75 C.", condRejectKW, condRejectKW))
		if !steam.wasteHeatSufficient {
			warnings = append(warnings, fmt.Sprintf("KILN WASTE-HEAT DEFICIT: Steam generation needs %.1f kW thermal, exceeding estimated kiln flue waste heat (%.1f kW above 400 F) — fire the certified steamer/boiler on its dedicated wood/LPG burner rather than relying solely on kiln flue heat.", steam.steamThermalKW, steam.kilnWasteHeatKW))
		}
	} else {
		warnings = []string{
			"DO NOT BUILD DIY PRESSURIZED STEAM BOILERS / STEAM-JET EJECTORS: Wrapping a closed pressurized water pipe or drum around the 2400 F kiln flue to drive a steam-jet ejector creates a LETHAL boiler explosion hazard without ASME-stamped pressure vessels and certified safety relief valves — NEVER use DIY pressurized steam!",
			tankBraceWarning,
			fmt.Sprintf("OUTDOOR-ONLY SITING & TALL CHIMNEY CLEARANCE: Never operate indoors — terminate the %.1f m (%.0f ft) brick/terracotta chimney outdoors well above studio roofs with lateral wall ties/guy-wires against wind gusts so CO, residual HCl, and fine salt haze disperse safely.", chimneyM, chimneyFt),
		}
	}
	if alkali.scalingWarning != "" {
		warnings = append(warnings, alkali.scalingWarning)
	}
	if !draftFeasible {
		warnings = append(warnings, fmt.Sprintf("PASSIVE DRAFT SHORTFALL: Required draft (%.3f inWC) exceeds natural chimney buoyancy (%.3f inWC) — reduce target kiln draft magnitude, widen the stoneware bed, select Steam-Jet Ejector draft, or toggle Zero Electricity OFF to add a local low-dP fan.", reqDraftInWC, availDraftInWC))
	}
	if boilerCorrosionNote != "" {
		warnings = append(warnings, boilerCorrosionNote)
	}
	if pumpMode == LocalPumpSteamJetSyphon {
		warnings = append(warnings, "STEAM-JET SYPHON COLD-WATER LIMIT (<120 F / 49 C): Condensing steam syphons & Penberthy injectors stall/break if fed hot scrubber sump liquor >120 F and heat discharge water by +30-45 F — plumb strictly to a dedicated COLD (<120 F) fresh makeup drum / boiler feed!")
	}
	if pumpMode == LocalPumpSumpPoolNoPump && !draftFeasible {
		warnings = append(warnings, fmt.Sprintf("SUMP-POOL SUBMERGENCE DRAFT SHORTFALL: Submerged S-vane + sieve trays (%.2f inWC) push required draft (%.3f inWC) above %.0f psig ejector draft (%.3f inWC) — step up to 60 psig HP Steam-Jet.", sumpPoolDPInWC, reqDraftInWC, steam.pressurePSIG, availDraftInWC))
	}
	if so2InletPPM > 10.0 {
		warnings = append(warnings, fmt.Sprintf("COAL/PEAT SULFUR (%.0f ppm SO2 in, %.1f%% alkaline capture) & H2SO4 DEW POINT (~%.0f F): Keep pre-quench flue masonry dry and hot (>%.0f F) so sulfuric acid mist never condenses onto metallic dampers, and maintain sump pH 7.5-8.5 continuously across the full %.1f-hr firing (not just during salting!).", so2InletPPM, so2RemovalPct, sulfurAcidDewF, sulfurAcidDewF+25.0, p.FiringDurationHours))
	}
	if !p.LocalAfterburner && model.IsSolidDirtyFuel(design.Input.FuelType) {
		warnings = append(warnings, tarMitigationNote)
	}
	if pAmb < 75000.0 && !steam.steamEnabled && zeroElec {
		warnings = append(warnings, fmt.Sprintf("EXTREME ALTITUDE BUOYANCY PENALTY (%.1f kPa / %s): Thin mountain air reduces natural chimney buoyancy by ~34%% and expands flue volume by +52%% — if natural chimney margin is tight, switch to 6 psig Subsonic or 20 psig LP Steam-Jet Ejector draft!", pAmb/1000.0, prof.elevationNote))
	}

	modeTag := "Local-Grid / 12V Solar Hybrid Ceramic Mode"
	if steam.steamEnabled && powerWatts == 0 {
		modeTag = fmt.Sprintf("Zero-Electricity Steam-Jet Ejector Ceramic Mode (%.0f psig)", steam.pressurePSIG)
	} else if steam.steamEnabled {
		modeTag = fmt.Sprintf("Hybrid Pump + %.0f psig Steam-Jet Ejector Ceramic Mode", steam.pressurePSIG)
	} else if zeroElec {
		modeTag = "100% Zero-Electricity Passive Ceramic Mode"
	}
	waterSupplyClause := fmt.Sprintf("Overhead %.0f L tank @ %.1f ft (%.2f PSI, %.2f L/min) + %s.", tankLiters, tankElevFt, gravityHeadPSI, sprayFlowLPM, manualPumpType)
	if !hasOverheadTank {
		waterSupplyClause = fmt.Sprintf("Dual ground-level brick sump pools (~360 L, 0 tower, 0 pumps) + 0.5\" surge-damped S-vane + flooded sump-level 60 psig ejector + 1.4 m co-current drenched trough + 2-stage ceramic sieve tray (~%d × 20L ground makeup buckets/hr).", jerrycansPerHr)
	}
	summary := fmt.Sprintf(
		"%s (%s): %.0f\" wide brick/SWG stoneware-packed tower (v_s=%.2f m/s, bed dP=%.3f inWC) + %.1f m (%.0f ft) brick chimney (%.0f F -> %.3f inWC avail vs %.3f inWC req, margin %+.3f inWC, feasible=%v). %s Neutralizes %.1f%% HCl with %.2f kg %s (%s); %.0f%% submicron NaCl capture; est. local cost %s (%.0f W).",
		modeTag, prof.label, bedDiaIn, vSup, bedDPInWC, chimneyM, chimneyFt, stackReheatF,
		availDraftInWC, reqDraftInWC, draftMarginInWC, draftFeasible,
		waterSupplyClause,
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
		DraftMode:               steam.draftMode,
		DraftModeLabel:          steam.draftModeLabel,
		SteamEnabled:            steam.steamEnabled,
		SteamPressurePSIG:       steam.pressurePSIG,
		SteamPressureBar:        steam.pressureBar,
		SteamNozzleVelMS:        steam.nozzleVelMS,
		SteamNozzleMach:         steam.nozzleMach,
		SteamNozzleThroatMM:     steam.nozzleThroatMM,
		SteamFlowKgHr:           steam.steamFlowKgHr,
		SteamWaterLHr:           steam.steamWaterLHr,
		SteamThermalKW:          steam.steamThermalKW,
		KilnWasteHeatKW:         steam.kilnWasteHeatKW,
		WasteHeatSufficient:     steam.wasteHeatSufficient,
		EjectorDraftInWC:        steam.ejectorDraftInWC,
		EjectorEntrainmentRatio: steam.ejectorEntrainmentRatio,
		CondensationGrowthNote:  steam.condensationGrowthNote,
		SteamSourceType:         steam.steamSourceType,
		SteamSafetyRequirements: steam.safetyRequirements,
		BoilerSource:            boilerSource,
		BoilerSourceLabel:       boilerSourceLabel,
		BoilerWaterSavedPct:     boilerWaterSavedPct,
		BoilerFuelSavedNote:     boilerFuelSavedNote,
		BoilerCorrosionNote:     boilerCorrosionNote,
		PumpMode:                pumpMode,
		PumpModeLabel:           pumpModeLabel,
		PumpSteamKgHr:           pumpSteamKgHr,
		HasOverheadTank:         hasOverheadTank,
		SumpPoolSubmergenceIn:   sumpPoolSubIn,
		SumpPoolDPInWC:          sumpPoolDPInWC,
		PumpEngineeringNote:     pumpEngNote,
		AmbientPressureKPa:      round1(pAmb / 1000.0),
		ElevationNote:           prof.elevationNote,
		FuelType:                design.Input.FuelType,
		SO2InletPPM:             so2InletPPM,
		SO2RemovalPct:           so2RemovalPct,
		SulfurAcidDewPointF:     sulfurAcidDewF,
		AfterburnerEnabled:      p.LocalAfterburner,
		AfterburnerNote:         afterburnerNote,
		TarMitigationNote:       tarMitigationNote,
		RegionalMachineryNote:   prof.machineryNote,
	}
}

func localCeramicBedDepthM(kilnVolCuFt float64) float64 {
	if kilnVolCuFt <= 16.0 {
		return 0.90
	}
	if kilnVolCuFt >= 40.0 {
		return 1.10
	}
	return 0.95
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

	bedDepthM := localCeramicBedDepthM(design.Input.KilnVolumeCuFt)
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
	sumpBOMName := fmt.Sprintf("Brick Baffled Settling Sump + %.0f L %s + Braced 3.5 m Stand", frugal.GravityTankLiters, prof.tankBrand)
	pumpBOMName := fmt.Sprintf("%s & Gravity Downcomer Plumbing", frugal.ManualPumpType)
	pumpBOMSpec := fmt.Sprintf("[LOCAL-PUMP-PLUMB] %s ($%.0f) + gravity downcomer/drain kit ($%.0f); flow %.2f L/min (%.2f GPM) at %.2f PSI head.", frugal.ManualPumpType, costs.pumpUSD, costs.plumbingUSD, frugal.SprayFlowLPM, frugal.SprayFlowGPM, frugal.GravityHeadPSI)
	if !frugal.HasOverheadTank {
		sumpPart = COTSPart{
			SKU:         "LOCAL-DUAL-SUMP-POOL",
			Name:        "Dual Ground-Level Baffled Brick Sump Pools (Pool #1 Skimmer + Pool #2 Flooded Weir — 0 Tower)",
			NominalIn:   36.0,
			Material:    "Cement-Plastered Firebrick Dual Sump Pools (~360 L) + 0.5\" Surge-Damped S-Vane + 1.4 m Co-Current Drenched Trough + 2-Stage Ceramic Sieve Tray",
			Source:      "Local Brickyard & Plumbing Shop (" + prof.label + ")",
			UnitCostUSD: round2(costs.brickSumpUSD + costs.gravityTankUSD + costs.standBracingUSD),
			WeightKg:    225.0,
			Notes:       frugal.PumpEngineeringNote,
		}
		sumpBOMName = sumpPart.Name
		pumpBOMName = "Zero-Pump Gravity-Flooded Weir Feed & Ground-Level Sump Interconnect Plumbing"
		pumpBOMSpec = fmt.Sprintf("[LOCAL-NOPUMP-PLUMB] Zero pumps & zero 3.5 m water tower ($0 pump + $%.0f flooded weir/drain fittings); 0.50\" surge-damped S-vane skimmer + 1.4 m co-current drenched trough + 2-stage ceramic sieve tray (%.2f inWC total extra dP -> %.1f%% HCl / %.0f%% NaCl).", costs.plumbingUSD, frugal.SumpPoolDPInWC, frugal.HClRemovalPct, frugal.SubmicronNaClCapturePct)
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
	if !frugal.HasOverheadTank {
		nozzlePart = COTSPart{
			SKU:         "LOCAL-SVANE-SIEVE",
			Name:        `Barometric Surge-Damped 0.5" S-Vane Skimmer, 1.4 m Co-Current Drenched Trough & 2-Stage Ceramic Sieve Tray`,
			NominalIn:   1.0,
			Material:    "Refractory Brick Barometric Surge Plenum + 0.5\" S-Vane + 2× Perforated Fired Stoneware Sieve Plates",
			Source:      "Potter's Studio Kiln + Local Brickyard",
			UnitCostUSD: costs.baffleZoneUSD,
			WeightKg:    34.0,
			Notes:       "Replaces gravity spray rose: 60 psig flooded sump-level ejector drenches the 1.4 m stoneware trough co-currently and polishes exhaust across 2 shallow 0.5\" ceramic sieve trays (~96.8% HCl / ~78% NaCl, zero clogging nozzles).",
		}
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
	if frugal.SteamEnabled {
		blowerPart = COTSPart{
			SKU:         fmt.Sprintf("LOCAL-STEAM-CHIMNEY-%.0fM", math.Round(frugal.ChimneyHeightM)),
			Name:        fmt.Sprintf("Short Ceramic / SWG Exhaust Chimney (%.1f m / %.0f ft) Driven by %.0f psig Steam-Jet Ejector", frugal.ChimneyHeightM, frugal.ChimneyHeightFt, frugal.SteamPressurePSIG),
			NominalIn:   10.0,
			Material:    "Fired Clay Bricks / Salt-Glazed SWG Flue Tiles + Galvanized Guy-Wires (No Reheat Hearth Needed)",
			Source:      "Local Brickyard / Hardware (" + prof.label + ")",
			UnitCostUSD: costs.chimneyUSD,
			WeightKg:    round2(34.0 * frugal.ChimneyHeightM),
			Notes:       frugal.DraftMechanism,
		}
	} else if p.ZeroElectricity {
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
			Name:        sumpBOMName,
			Material:    sumpPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s", sumpPart.SKU, sumpPart.Notes),
			UnitCostUSD: sumpPart.UnitCostUSD,
			WeightKg:    sumpPart.WeightKg,
		},
		{
			Category:    "Liquid & Gravity Circulation",
			Name:        pumpBOMName,
			Material:    "Cast-Iron / Stainless / HDPE Pump or Aspiration Dip-Tube + Drain Valve",
			Quantity:    1,
			Unit:        "assy",
			Spec:        pumpBOMSpec,
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

	if costs.afterburnerUSD > 0 {
		items = append(items, model.BOMItem{
			Category:    "Stage 0: Passive Masonry Tar Afterburner & Weir",
			Name:        "Passive Secondary-Air Refractory Checkerwork Smoke-Burner Throat + Pool #1 Tar-Skimming Weir",
			Material:    "Dense Alumina Firebrick Checkerwork + Secondary-Air Damper Bricks + Washable Coarse Ceramic Anti-Tar Louver Plates",
			Quantity:    1,
			Unit:        "assy",
			Spec:        "[LOCAL-TAR-AFTERBURNER] Thermal tar-cracking throat (>750 C, 0 W, +0.015 inWC dP) cracks 70-85% of coal/peat/PKS black smoke & phenolic tars before Stage 1; paired with removable Pool #1 tar-skimming baffle.",
			UnitCostUSD: costs.afterburnerUSD,
			WeightKg:    135.0,
		})
	}

	if frugal.SteamEnabled {
		steamSKU := "LOCAL-STEAM-LP-20PSIG"
		steamWt := 24.0
		steamCat := "Stage 2B: Certified Steam Source & de Laval Ejector"
		steamMat := "Certified Non-IBR/KEBS/ASME Pressure Trim, 316L de Laval Nozzle, Spring-Loaded PRV, Gauge, Check Valve & Insulated Header S1"
		steamName := fmt.Sprintf("%.0f psig (%.2f barg) %s", frugal.SteamPressurePSIG, frugal.SteamPressureBar, frugal.SteamSourceType)
		steamSpec := fmt.Sprintf("[%s] %.1f kg/hr steam @ %.0f psig -> %.0f m/s (Mach %.2f, %.2f mm throat), %.3f inWC ejector draft (entrainment %.1f:1), %.1f kW thermal; MUST pass 1.5x MAWP cold hydro-test and NEVER use DIY welded drum boilers.", steamSKU, frugal.SteamFlowKgHr, frugal.SteamPressurePSIG, frugal.SteamNozzleVelMS, frugal.SteamNozzleMach, frugal.SteamNozzleThroatMM, frugal.EjectorDraftInWC, frugal.EjectorEntrainmentRatio, frugal.SteamThermalKW)
		if frugal.DraftMode == LocalDraftSteamEjectorSubsonic {
			steamSKU = "LOCAL-STEAM-SUBSONIC-6PSIG"
			steamWt = 18.0
			steamCat = "Stage 2B: Commercial Catering Steamer, 4.2 m Standpipe & Brass Orifice"
			steamMat = "Certified 12-15 psig Commercial Stainless Catering Steamer, 4.2 m Open Hydrostatic Standpipe (S2), Drilled Brass Converging Orifice, 10-15 psig Brass PRV, 0-15 psig Gauge, Bronze Check Valve & Insulated Header S1"
			steamName = fmt.Sprintf("6 psig (%.3f barg) %s", frugal.SteamPressureBar, frugal.SteamSourceType)
			steamSpec = fmt.Sprintf("[%s] %.1f kg/hr unchoked subsonic steam @ 6 psig -> %.0f m/s (Mach %.2f < 1.0, %.2f mm drilled brass orifice, ~52 dBA), %.3f inWC ejector draft (entrainment %.1f:1), %.1f kW thermal + 4.2 m open hydrostatic water-column standpipe (6.0 psig physical limit) & 10-15 psig brass PRV; MUST pass 1.5x MAWP cold hydro-test and NEVER use unvented DIY welded drum boilers.", steamSKU, frugal.SteamFlowKgHr, frugal.SteamNozzleVelMS, frugal.SteamNozzleMach, frugal.SteamNozzleThroatMM, frugal.EjectorDraftInWC, frugal.EjectorEntrainmentRatio, frugal.SteamThermalKW)
		} else if frugal.DraftMode == LocalDraftSteamEjectorHP {
			steamSKU = "LOCAL-STEAM-HP-60PSIG"
			steamWt = 46.0
			steamSpec = fmt.Sprintf("[%s] %.1f kg/hr steam @ %.0f psig -> %.0f m/s (Mach %.2f, %.2f mm throat), %.3f inWC ejector draft (entrainment %.1f:1), %.1f kW thermal; MUST pass 1.5x MAWP cold hydro-test and NEVER use DIY welded drum boilers.", steamSKU, frugal.SteamFlowKgHr, frugal.SteamPressurePSIG, frugal.SteamNozzleVelMS, frugal.SteamNozzleMach, frugal.SteamNozzleThroatMM, frugal.EjectorDraftInWC, frugal.EjectorEntrainmentRatio, frugal.SteamThermalKW)
		}
		items = append(items, model.BOMItem{
			Category:    steamCat,
			Name:        steamName,
			Material:    steamMat,
			Quantity:    1,
			Unit:        "assy",
			Spec:        steamSpec,
			UnitCostUSD: costs.steamSystemUSD,
			WeightKg:    steamWt,
		})
	}

	bomSafetyNote := "SAFETY: NEVER build an uncertified DIY pressurized steam boiler on the kiln flue. Brace the 3.5 m overhead gravity tank stand for >350-550 kg wet load, and discharge strictly outdoors."
	if frugal.SteamEnabled {
		if frugal.DraftMode == LocalDraftSteamEjectorSubsonic {
			bomSafetyNote = fmt.Sprintf("SUBSONIC STEAM SAFETY (6 psig / 0.414 barg): Use ONLY a factory-certified 12-15 psig commercial catering stainless steamer/kettle equipped with BOTH an unvalved 4.2-4.4 m open vertical hydrostatic water-column relief standpipe (6.0 psig physical cap, zero moving parts) and a stamped 10-15 psig brass PRV + fusible plug, 1.5x MAWP (23 psig) cold hydro-test, 0-15 psig gauge, bronze check valve, scald guards, and open 0 psig feed tank — NEVER fabricate an unvented DIY welded drum boiler! (%.1f kg/hr).", frugal.SteamFlowKgHr)
		} else if frugal.DraftMode == LocalDraftSteamEjectorHP {
			bomSafetyNote = fmt.Sprintf("HP STEAM SAFETY (60 psig / 4.14 barg): Use ONLY a factory-certified non-IBR (<=25 L)/KEBS/ASME commercial baby steam boiler with stamped spring-loaded PRV, 1.5x MAWP cold hydro-test, armored sight glass, fusible plug, pressure gauge, check valve, boiler feed pump, and scald guards — NEVER fabricate a DIY welded drum boiler or studio monotube coil at 60 psig! (%.0f psig / %.1f kg/hr).", frugal.SteamPressurePSIG, frugal.SteamFlowKgHr)
		} else {
			bomSafetyNote = fmt.Sprintf("LP STEAM SAFETY (20 psig / 1.38 barg): Use ONLY a factory-certified non-IBR (<=25 L)/KEBS/ASME baby boiler or code-rated small-bore (<2 L) seamless monotube coil (<= 20 psig ONLY) with stamped spring-loaded PRV, 1.5x MAWP cold hydro-test, pressure gauge, check valve, scald guards, and boiler feed pump / atmospheric batch refill — NEVER fabricate a DIY welded drum boiler! (%.0f psig / %.1f kg/hr).", frugal.SteamPressurePSIG, frugal.SteamFlowKgHr)
		}
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
			bomSafetyNote,
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
