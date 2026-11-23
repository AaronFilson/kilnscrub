package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// COTSPart describes a real commercial off-the-shelf (COTS) hardware component
// suitable for studio-scale kiln scrubber assembly.
type COTSPart struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	NominalIn   float64 `json:"nominal_in"`
	Material    string  `json:"material"`
	Source      string  `json:"source"`
	UnitCostUSD float64 `json:"unit_cost_usd"`
	WeightKg    float64 `json:"weight_kg"`
	Notes       string  `json:"notes"`
}

// COTSSummary captures the snapped off-the-shelf component selections,
// recalculated column flooding percentage, and complete COTS Bill of Materials.
type COTSSummary struct {
	Snapped              bool      `json:"snapped"`
	QuenchVessel         COTSPart  `json:"quench_vessel"`
	PackedColumn         COTSPart  `json:"packed_column"`
	SumpTank             COTSPart  `json:"sump_tank"`
	SprayNozzle          COTSPart  `json:"spray_nozzle"`
	PackingMedia         COTSPart  `json:"packing_media"`
	BlowerVFD            COTSPart  `json:"blower_vfd"`
	Silencer             *COTSPart `json:"silencer,omitempty"`
	PlumeDiluter         *COTSPart `json:"plume_diluter,omitempty"`
	SensorKit            COTSPart  `json:"sensor_kit"`
	ActualPercentOfFlood float64   `json:"actual_percent_of_flood"`
	EstimatedTotalUSD    float64   `json:"estimated_total_usd"`
	BOM                  model.BOM `json:"bom"`
}

type pipeSpec struct {
	nominalIn float64
	actualID  float64
	skuSuffix string
	baseCost  float64
	kgPerFt   float64
}

var quench316LCatalog = []pipeSpec{
	{6, 6.00, "06", 245, 4.2},
	{8, 8.00, "08", 315, 5.6},
	{10, 10.00, "10", 395, 7.1},
	{12, 12.00, "12", 485, 8.6},
	{14, 14.00, "14", 590, 10.4},
	{16, 16.00, "16", 710, 12.1},
	{18, 18.00, "18", 845, 14.2},
	{24, 24.00, "24", 1220, 19.5},
}

var quenchC276Catalog = []pipeSpec{
	{6, 6.00, "06", 980, 9.5},
	{8, 8.00, "08", 1240, 12.5},
	{10, 10.00, "10", 1560, 16.0},
	{12, 12.00, "12", 1890, 19.5},
	{14, 14.00, "14", 2280, 23.0},
	{16, 16.00, "16", 2720, 27.0},
	{18, 18.00, "18", 3190, 31.5},
	{24, 24.00, "24", 4450, 44.0},
}

// Schedule 40 CPVC/PVC pipe nominal sizes with true bore IDs, plus 55-gal HDPE drum (22.5" ID)
// and 30" FRP/HDPE cylinder for co-op scale columns.
var packedCPVCCatalog = []pipeSpec{
	{6, 6.065, "CPVC-06", 145, 2.9},
	{8, 7.981, "CPVC-08", 215, 4.3},
	{10, 10.020, "CPVC-10", 320, 6.1},
	{12, 11.938, "CPVC-12", 435, 8.1},
	{14, 13.124, "CPVC-14", 560, 9.6},
	{16, 15.000, "CPVC-16", 715, 11.8},
	{22.5, 22.500, "HDPE-55D", 195, 3.8},
	{30, 30.000, "FRP-30", 1180, 14.0},
}

var packedFRPCatalog = []pipeSpec{
	{6, 6.00, "FRP-06", 420, 3.4},
	{8, 8.00, "FRP-08", 540, 4.6},
	{10, 10.00, "FRP-10", 680, 5.9},
	{12, 12.00, "FRP-12", 835, 7.4},
	{14, 14.00, "FRP-14", 990, 8.8},
	{16, 16.00, "FRP-16", 1175, 10.4},
	{18, 18.00, "FRP-18", 1360, 12.0},
	{24, 24.00, "FRP-24", 1890, 16.5},
	{30, 30.00, "FRP-30", 2540, 21.5},
}

func round2(v float64) float64 {
	return math.Round(v*100.0) / 100.0
}

func round1(v float64) float64 {
	return math.Round(v*10.0) / 10.0
}

func snapPipe(catalog []pipeSpec, reqDiamIn float64) (pipeSpec, bool) {
	for _, spec := range catalog {
		if spec.actualID >= reqDiamIn-1e-6 {
			return spec, true
		}
	}
	return catalog[len(catalog)-1], false
}

func selectQuenchVessel(project StudioProject, design model.ScrubberDesign) COTSPart {
	reqIn := design.Quench.TowerDiameter.Inches()
	heightFt := design.Quench.TowerHeight.Feet()
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP

	if !project.SnapToCOTS {
		mat := "316L Stainless Steel (14 ga)"
		sku := fmt.Sprintf("FAB-Q316-%.0f", math.Round(reqIn))
		source := "Local sheet-metal fabricator (scrubber.dxf)"
		cost := round2(280.0 + 36.0*reqIn + 42.0*heightFt)
		weight := round2(math.Max(14.0, 0.85*reqIn*heightFt))
		notes := fmt.Sprintf("Custom-rolled %.1f\" ID x %.1f ft quench shell from DXF pattern.", reqIn, heightFt)
		if heavy {
			mat = "Alloy C-276 / Castable Refractory-Lined Shell"
			sku = fmt.Sprintf("FAB-QC276-%.0f", math.Round(reqIn))
			source = "Corrosion Materials / Refractory Process Fab"
			cost = round2(1100.0 + 140.0*reqIn + 115.0*heightFt)
			weight = round2(math.Max(32.0, 1.9*reqIn*heightFt))
			notes = fmt.Sprintf("Acid-proof %.1f\" ID x %.1f ft Alloy C-276 / refractory-lined quench tower.", reqIn, heightFt)
		}
		return COTSPart{
			SKU:         sku,
			Name:        fmt.Sprintf("Custom Quench Tower Shell (%.1f\" ID x %.1f ft)", reqIn, heightFt),
			NominalIn:   round1(reqIn),
			Material:    mat,
			Source:      source,
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       notes,
		}
	}

	if heavy {
		spec, _ := snapPipe(quenchC276Catalog, reqIn)
		cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
		weight := round2(spec.kgPerFt * heightFt)
		return COTSPart{
			SKU:         "C276-QSP-" + spec.skuSuffix,
			Name:        fmt.Sprintf(`%.0f" Alloy C-276 / Refractory-Lined Quench Spool (%.1f ft)`, spec.nominalIn, heightFt),
			NominalIn:   spec.nominalIn,
			Material:    "Alloy C-276 (UNS N10276) / Castable Refractory Liner",
			Source:      "Corrosion Materials / Industrial Process Pipe",
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       fmt.Sprintf("Immune to hot wet HCl pitting & Cl-SCC up to 2400 F inlet; bolted ANSI flange ends (snapped from %.1f\").", reqIn),
		}
	}

	spec, _ := snapPipe(quench316LCatalog, reqIn)
	cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
	weight := round2(spec.kgPerFt * heightFt)
	notes := fmt.Sprintf("Standard %.0f\" flanged 316L SS sanitary/flue pipe spool (snapped from %.1f\").", spec.nominalIn, reqIn)
	if project.VaporGlazeMode == VaporGlazeSalt {
		notes += " Use alkaline NaHCO3 quench spray (pH 8.0-8.5) and treat as a replaceable bolted spool."
	}
	return COTSPart{
		SKU:         "SS316L-FLUE-" + spec.skuSuffix,
		Name:        fmt.Sprintf(`%.0f" 316L Stainless Sanitary/Flue Quench Pipe (%.1f ft)`, spec.nominalIn, heightFt),
		NominalIn:   spec.nominalIn,
		Material:    "316L Stainless Steel (16/14 ga Flanged)",
		Source:      "McMaster-Carr / Grainger / VentingPipe",
		UnitCostUSD: cost,
		WeightKg:    weight,
		Notes:       notes,
	}
}

func selectPackedColumn(project StudioProject, design model.ScrubberDesign) (COTSPart, float64, float64) {
	reqIn := design.PackedBed.ColumnDiameter.Inches()
	heightFt := design.PackedBed.ColumnHeight.Feet()
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP

	if !project.SnapToCOTS {
		mat := "CPVC / Polypropylene Sheet"
		sku := fmt.Sprintf("FAB-COL-%.0f", math.Round(reqIn))
		source := "Local plastics fabricator (scrubber.dxf)"
		cost := round2(310.0 + 38.0*reqIn + 35.0*heightFt)
		weight := round2(math.Max(16.0, 0.72*reqIn*heightFt))
		if heavy {
			mat = "Derakane Vinyl-Ester FRP"
			sku = fmt.Sprintf("FAB-FRP-%.0f", math.Round(reqIn))
			source = "Industrial FRP Fabricator (scrubber.dxf)"
			cost = round2(620.0 + 75.0*reqIn + 55.0*heightFt)
			weight = round2(math.Max(22.0, 0.95*reqIn*heightFt))
		}
		part := COTSPart{
			SKU:         sku,
			Name:        fmt.Sprintf("Custom Packed-Bed Column (%.1f\" ID x %.1f ft)", reqIn, heightFt),
			NominalIn:   round1(reqIn),
			Material:    mat,
			Source:      source,
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       fmt.Sprintf("Exact calculated column diameter %.1f\" (un-snapped).", reqIn),
		}
		return part, reqIn, design.PackedBed.PercentOfFlood
	}

	var spec pipeSpec
	var part COTSPart
	if heavy {
		spec, _ = snapPipe(packedFRPCatalog, reqIn)
		cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
		weight := round2(spec.kgPerFt * heightFt)
		part = COTSPart{
			SKU:         "DERAKANE-" + spec.skuSuffix,
			Name:        fmt.Sprintf(`%.0f" Vinyl-Ester FRP Packed Column (%.1f ft)`, spec.nominalIn, heightFt),
			NominalIn:   spec.nominalIn,
			Material:    "Derakane Vinyl-Ester FRP (210 F Wet Acid Rated)",
			Source:      "Ershigs / Beetle Plastics / Industrial FRP",
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       fmt.Sprintf("Filament-wound chemical-grade FRP column (ID %.1f\", snapped from %.1f\").", spec.actualID, reqIn),
		}
	} else {
		spec, _ = snapPipe(packedCPVCCatalog, reqIn)
		if spec.skuSuffix == "HDPE-55D" {
			// Two stacked/bolted open-head 55-gal HDPE drums form a tall low-cost studio column.
			drums := math.Max(2.0, math.Ceil(heightFt/2.8))
			part = COTSPart{
				SKU:         "ULINE-S-10757-COL",
				Name:        fmt.Sprintf(`55-Gal Open-Head HDPE Drum Column (22.5" ID, %.0fx stacked)`, drums),
				NominalIn:   spec.nominalIn,
				Material:    "Chemical-Grade HDPE (22.5\" ID)",
				Source:      "Uline / Grainger / McMaster-Carr",
				UnitCostUSD: round2(95.0 * drums),
				WeightKg:    round2(10.5 * drums),
				Notes:       fmt.Sprintf("Bolted flange-joined 55-gal HDPE drum sections (22.5\" ID, snapped from %.1f\"); keep quench <= 185 F.", reqIn),
			}
		} else if spec.skuSuffix == "FRP-30" {
			cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
			weight := round2(spec.kgPerFt * heightFt)
			part = COTSPart{
				SKU:         "COL-FRP-30",
				Name:        fmt.Sprintf(`30" HDPE / FRP Absorber Column (%.1f ft)`, heightFt),
				NominalIn:   spec.nominalIn,
				Material:    "HDPE / Vinyl-Ester FRP",
				Source:      "US Plastic / Industrial FRP",
				UnitCostUSD: cost,
				WeightKg:    weight,
				Notes:       fmt.Sprintf("Large-bore column snapped from %.1f\" for community/co-op flow.", reqIn),
			}
		} else {
			cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
			weight := round2(spec.kgPerFt * heightFt)
			part = COTSPart{
				SKU:         "HARVEL-" + spec.skuSuffix,
				Name:        fmt.Sprintf(`%.0f" Sch 40/80 CPVC Pipe Column (ID %.2f", %.1f ft)`, spec.nominalIn, spec.actualID, heightFt),
				NominalIn:   spec.nominalIn,
				Material:    "Schedule 40/80 CPVC (ASTM F441, 200 F max)",
				Source:      "SupplyHouse / US Plastic / McMaster-Carr",
				UnitCostUSD: cost,
				WeightKg:    weight,
				Notes:       fmt.Sprintf("Standard COTS CPVC pipe (true bore %.2f\", snapped from %.1f\"); requires 210 F bypass interlock.", spec.actualID, reqIn),
			}
		}
	}

	// Recalculate superficial velocity and actual % of flood for the snapped bore ID.
	actualDiamM := float64(model.InchesToMeters(spec.actualID))
	actualArea := math.Pi * actualDiamM * actualDiamM / 4.0
	actualPctFlood := design.PackedBed.PercentOfFlood
	if actualArea > 0 && design.PackedBed.FloodingVelocity > 0 && design.QuenchedGas.VolumetricFlow > 0 {
		vSupActual := float64(design.QuenchedGas.VolumetricFlow) / actualArea
		actualPctFlood = round2((vSupActual / float64(design.PackedBed.FloodingVelocity)) * 100.0)
	}
	return part, spec.actualID, actualPctFlood
}

func selectSumpTank(design model.ScrubberDesign) COTSPart {
	sumpGal := float64(design.PackedBed.SumpVolume) / model.CubicMetersPerUSGallon
	switch {
	case sumpGal <= 22.0:
		return COTSPart{
			SKU:         "ULINE-S-16913",
			Name:        "30-Gallon Open-Head HDPE Chemical Sump Drum",
			NominalIn:   19.0,
			Material:    "High-Density Polyethylene (HDPE)",
			Source:      "Uline / Global Industrial",
			UnitCostUSD: 72.0,
			WeightKg:    7.5,
			Notes:       fmt.Sprintf("Covers %.1f gal required holdup with 100%% ash-settling & NaHCO3 mixing reserve.", sumpGal),
		}
	case sumpGal <= 45.0:
		return COTSPart{
			SKU:         "ULINE-S-10757",
			Name:        "55-Gallon Open-Head HDPE Chemical Sump Drum",
			NominalIn:   22.5,
			Material:    "High-Density Polyethylene (HDPE)",
			Source:      "Uline / Grainger / McMaster-Carr",
			UnitCostUSD: 96.0,
			WeightKg:    10.5,
			Notes:       fmt.Sprintf("Covers %.1f gal required holdup; easy top access for baking soda dosing and ash basket.", sumpGal),
		}
	default:
		return COTSPart{
			SKU:         "USPLASTIC-100T",
			Name:        "100-Gallon Heavy-Duty Poly Chemical Sump Tank",
			NominalIn:   28.0,
			Material:    "Cross-Linked HDPE (XLPE)",
			Source:      "US Plastic Corp / Ace Roto-Mold",
			UnitCostUSD: 315.0,
			WeightKg:    22.0,
			Notes:       fmt.Sprintf("Co-op scale sump for %.1f gal holdup with bulkheads for recirculation & blowdown.", sumpGal),
		}
	}
}

func selectSprayNozzle(project StudioProject, design model.ScrubberDesign) COTSPart {
	gpm := design.Quench.TotalSprayWaterGPM
	count := design.Quench.NozzleCount
	if count <= 0 {
		count = 2
	}
	perNozzle := gpm / float64(count)
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP

	nomIn := 0.25
	sizeLabel := `1/4" NPT`
	if perNozzle > 0.75 {
		nomIn = 0.375
		sizeLabel = `3/8" NPT`
	}
	if perNozzle > 1.8 {
		nomIn = 0.5
		sizeLabel = `1/2" NPT`
	}

	if heavy {
		return COTSPart{
			SKU:         fmt.Sprintf("SSCO-FJ-C276-%d", int(nomIn*1000)),
			Name:        fmt.Sprintf("Hastelloy C-276 Fine-Atomizing Full-Cone Nozzle (%s)", sizeLabel),
			NominalIn:   nomIn,
			Material:    "Alloy C-276 (Hastelloy)",
			Source:      "Spraying Systems Co. / BETE",
			UnitCostUSD: 145.0,
			WeightKg:    0.45,
			Notes:       fmt.Sprintf("Sized for %.2f GPM/nozzle (%.2f GPM total quench); run at 60-100 psig or air-assist for <=200 um droplets.", perNozzle, gpm),
		}
	}
	return COTSPart{
		SKU:         fmt.Sprintf("MC-316FJ-%d", int(nomIn*1000)),
		Name:        fmt.Sprintf("316L SS Full-Cone Fine-Atomizing Spray Nozzle (%s)", sizeLabel),
		NominalIn:   nomIn,
		Material:    "316L Stainless Steel",
		Source:      "McMaster-Carr / Spraying Systems Co.",
		UnitCostUSD: 58.0,
		WeightKg:    0.35,
		Notes:       fmt.Sprintf("Sized for %.2f GPM/nozzle (%.2f GPM total quench); use >=60 psig pressure or twin-fluid misting to ensure complete evaporation.", perNozzle, gpm),
	}
}

func selectPackingMedia(actualColIDIn float64, design model.ScrubberDesign) (COTSPart, float64) {
	colDiaM := float64(model.InchesToMeters(actualColIDIn))
	colArea := math.Pi * colDiaM * colDiaM / 4.0
	packVolM3 := colArea * float64(design.PackedBed.PackingHeight)
	if packVolM3 <= 0 {
		packVolM3 = float64(design.PackedBed.PackingVolume)
	}
	packCuFt := math.Max(1.0, math.Ceil(model.CubicMeters(packVolM3).CuFt()))

	part := COTSPart{
		SKU:         "PACK-PP-PALL-1IN",
		Name:        `1" Polypropylene Pall Rings (1 cu.ft. Bulk Bag)`,
		NominalIn:   1.0,
		Material:    "Polypropylene (PP), 90% Voidage, Fp=52/ft",
		Source:      "McMaster-Carr / Jaeger / Amazon Industrial",
		UnitCostUSD: 48.0,
		WeightKg:    2.5,
		Notes:       fmt.Sprintf("Order %.0f cu.ft. to fill %.1f\" ID column to %.1f\" bed depth (%.2f m). Bio-balls acceptable for zero-acid smoke-only mode.", packCuFt, actualColIDIn, design.PackedBed.PackingHeight.Inches(), float64(design.PackedBed.PackingHeight)),
	}
	return part, packCuFt
}

func selectBlowerVFD(project StudioProject, design model.ScrubberDesign) COTSPart {
	hp := design.Blower.RecommendedMotorHP
	if hp <= 0 {
		hp = 1.0
	}
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP

	voltageStr := "120V 1-Ph Input VFD"
	vfdSKU := "GS21-10P"
	if hp > 1.0 {
		voltageStr = "240V 1-Ph Input -> 3-Ph Motor VFD"
		vfdSKU = "GS21-20P"
	}

	var blowerCost, vfdCost, weightKg float64
	switch {
	case hp <= 0.5:
		blowerCost, vfdCost, weightKg = 640, 175, 22.0
	case hp <= 1.0:
		blowerCost, vfdCost, weightKg = 860, 215, 29.0
	case hp <= 1.5:
		blowerCost, vfdCost, weightKg = 1120, 265, 36.0
	case hp <= 2.0:
		blowerCost, vfdCost, weightKg = 1340, 310, 44.0
	case hp <= 3.0:
		blowerCost, vfdCost, weightKg = 1720, 395, 58.0
	default:
		blowerCost = round2(1720.0 + 360.0*(hp-3.0))
		vfdCost = round2(395.0 + 110.0*(hp-3.0))
		weightKg = round2(58.0 + 12.0*(hp-3.0))
	}

	mat := "Powder-Coated Steel Housing / 316L SS Radial Wheel + NEMA 4X VFD"
	sku := fmt.Sprintf("PB-RAD-%.2gHP-%s", hp, vfdSKU)
	source := "Grainger (Dayton/Cincinnati Fan) + AutomationDirect"
	if heavy {
		blowerCost = round2(blowerCost * 1.45)
		mat = "Vinyl-Ester FRP Housing & Chemical-Duty Radial Wheel + NEMA 4X VFD"
		sku = fmt.Sprintf("FRP-RAD-%.2gHP-%s", hp, vfdSKU)
		source = "New York Blower / Hartzell FRP + AutomationDirect"
	}

	return COTSPart{
		SKU:         sku,
		Name:        fmt.Sprintf("%.2g HP Radial-Blade Pressure Blower + %s", hp, voltageStr),
		NominalIn:   round1(design.Blower.WheelDiameter.Inches()),
		Material:    mat,
		Source:      source,
		UnitCostUSD: round2(blowerCost + vfdCost),
		WeightKg:    weightKg,
		Notes:       fmt.Sprintf("High-head radial pressure blower sized for %.0f ACFM @ %.1f inWC static pressure; VFD accepts 0-10V / Modbus PID draft control.", design.Blower.FlowACFM, design.Blower.TotalStaticPressureInWC),
	}
}

func selectSilencer(design model.ScrubberDesign) COTSPart {
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()/2.0)*2.0)
	cost := round2(220.0 + 22.0*ductIn)
	weight := round2(8.0 + 1.1*ductIn)
	return COTSPart{
		SKU:         fmt.Sprintf("SIL-ACOUSTIC-%02.0f", ductIn),
		Name:        fmt.Sprintf(`%.0f" Inline Acoustic Discharge Silencer & Insulated Blower Box`, ductIn),
		NominalIn:   ductIn,
		Material:    "Galvanized/FRP Outer Shell, Moisture-Resistant Acoustic Baffle",
		Source:      "Grainger / Fantech / DIY Rockwool Enclosure",
		UnitCostUSD: cost,
		WeightKg:    weight,
		Notes:       "Reduces blower discharge & casing noise by ~13 dBA for residential property-line compliance (<=60 dBA).",
	}
}

// RecommendedDilutionCFM picks a standard 120V forward-curved squirrel-cage blower capacity
// (~5x-6x scrubber ACFM) for high-velocity downstream stack plume dilution.
func RecommendedDilutionCFM(design model.ScrubberDesign) (cfm float64, coneIn float64) {
	target := design.Blower.FlowACFM * 6.0
	switch {
	case target <= 500.0:
		return 500.0, 6.0
	case target <= 1050.0:
		return 1000.0, 8.0
	case target <= 1650.0:
		return 1500.0, 10.0
	case target <= 2250.0:
		return 2000.0, 10.0
	default:
		return 3000.0, 12.0
	}
}

func selectPlumeDiluter(design model.ScrubberDesign) COTSPart {
	cfm, coneIn := RecommendedDilutionCFM(design)
	var cost, weight, amps float64
	switch {
	case cfm <= 500.0:
		cost, weight, amps = 245.0, 11.5, 2.2
	case cfm <= 1000.0:
		cost, weight, amps = 345.0, 16.5, 4.1
	case cfm <= 1500.0:
		cost, weight, amps = 445.0, 22.0, 6.5
	case cfm <= 2000.0:
		cost, weight, amps = 540.0, 27.0, 8.8
	default:
		cost, weight, amps = 695.0, 34.0, 12.5
	}
	return COTSPart{
		SKU:         fmt.Sprintf("SQ-DILUTE-%.0fCFM-%.0fIN", cfm, coneIn),
		Name:        fmt.Sprintf(`%.0f CFM 120V Squirrel-Cage Plume Dilution Blower + Mixing Tee & %.0f" HV Exit Cone`, cfm, coneIn),
		NominalIn:   coneIn,
		Material:    "Galvanized Steel Forward-Curved Blower + Galvanized/316L Stack Mixing Tee & Taper Cone",
		Source:      "Grainger (Dayton 1TDT/4C) / SupplyHouse HVAC",
		UnitCostUSD: cost,
		WeightKg:    weight,
		Notes:       fmt.Sprintf("120V (~%.1f A, 0.7-1.2 inWC) mounted DOWNSTREAM of Stage 4 ID blower (never in place of the 21 inWC radial blower). %.0f\" high-velocity exit cone (~3,400 FPM) cuts winter plume reach ~3x-4x; add optional cold-intake kiln-jacket warm-air shroud for near-zero visible plume.", amps, coneIn),
	}
}

func selectSensorKit(design model.ScrubberDesign) COTSPart {
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()))
	damperCost := round2(265.0 + 22.0*ductIn)
	return COTSPart{
		SKU:         "KIT-MODBUS-SAFE-V1",
		Name:        "Modbus RTU Sensor & Fail-Safe Bypass Interlock Kit",
		NominalIn:   ductIn,
		Material:    "316L SS Thermowells / NEMA 4X / Spring-Return 24VDC Damper",
		Source:      "AutomationDirect / Dwyer / Belimo / Atlas Scientific",
		UnitCostUSD: round2(415.0 + damperCost),
		WeightKg:    11.5,
		Notes:       fmt.Sprintf(`Includes Type-K flue TC, Pt100 quench RTD (210 F cutoff), Dwyer +/-2 inWC draft transmitter, Hall/paddlewheel flow switch (0.25 GPM cutoff), pH probe, RS-485 USB adapter, and %.0f" spring-return fail-open bypass damper.`, ductIn),
	}
}

// BuildCOTSSummary snaps (when requested) the engineered scrubber geometry to standard
// commercial off-the-shelf hardware, recalculates actual % of flooding velocity, and
// builds a complete COTS Bill of Materials.
func BuildCOTSSummary(project StudioProject, design model.ScrubberDesign) COTSSummary {
	quenchPart := selectQuenchVessel(project, design)
	colPart, actualColIDIn, actualPctFlood := selectPackedColumn(project, design)
	sumpPart := selectSumpTank(design)
	nozzlePart := selectSprayNozzle(project, design)
	packingPart, packCuFt := selectPackingMedia(actualColIDIn, design)
	blowerPart := selectBlowerVFD(project, design)
	sensorPart := selectSensorKit(design)

	var silencerPtr *COTSPart
	if project.IncludeSilencer {
		s := selectSilencer(design)
		silencerPtr = &s
	}

	var diluterPtr *COTSPart
	if project.IncludePlumeDiluter {
		d := selectPlumeDiluter(design)
		diluterPtr = &d
	}

	nozzleQty := float64(design.Quench.NozzleCount + 2)
	if nozzleQty < 3 {
		nozzleQty = 4
	}

	// Venturi throat & transition assembly (no-weld bolted reducers vs welded DXF cones).
	ventMat := "316L Stainless Steel"
	ventCost := round2(240.0 + 18.0*design.Venturi.InletDiameter.Inches())
	ventWeight := round2(7.5 + 0.9*design.Venturi.InletDiameter.Inches())
	ventName := "Bolted Concentric Reducer Venturi & Throat Injection Spool"
	ventSpec := fmt.Sprintf(`No-weld bolted reducers: %.1f" -> %.2f" throat -> %.1f" with 2x NPT throat injection ports`,
		design.Venturi.InletDiameter.Inches(), design.Venturi.ThroatDiameter.Inches(), design.Venturi.OutletDiameter.Inches())
	if project.BuildPath == BuildPathWeldedDXF {
		ventName = "Rolled & Welded Venturi Cones & Throat Spool (from scrubber.dxf)"
		ventSpec = fmt.Sprintf(`DXF flat-pattern cones: %.0fmm -> %.1fmm throat -> %.0fmm + bolted flanges`,
			design.Venturi.InletDiameter.Millimeters(), design.Venturi.ThroatDiameter.Millimeters(), design.Venturi.OutletDiameter.Millimeters())
		ventCost = round2(ventCost * 1.15)
	}
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		ventMat = "Alloy C-276 Throat / Silicon Carbide Liner + FRP Diffuser"
		ventCost = round2(ventCost * 2.6)
		ventWeight = round2(ventWeight * 1.35)
	}

	// Chemical-resistant mag-drive recirculation pump.
	recircGPM := design.PackedBed.RecirculationFlowGPM
	if recircGPM <= 0 {
		recircGPM = math.Max(4.0, design.Quench.TotalSprayWaterGPM*3.0)
	}
	pumpCost := round2(265.0 + 14.0*recircGPM)
	pumpWeight := round2(7.5 + 0.30*recircGPM)

	// 3-pass polypropylene chevron mist eliminator pad sized to actual column ID.
	colAreaSqFt := math.Pi * (actualColIDIn / 12.0) * (actualColIDIn / 12.0) / 4.0
	demisterCost := round2(145.0 + 95.0*colAreaSqFt)
	demisterWeight := round2(math.Max(2.0, 2.8*colAreaSqFt))

	// Gaskets, clamps, CPVC cement, hardware kit.
	hardwareCost := 195.0
	hardwareName := "No-Weld Couplings, EPDM/PTFE Gaskets, CPVC Cement & 316 SS Clamps"
	if project.BuildPath == BuildPathWeldedDXF {
		hardwareCost = 245.0
		hardwareName = "PTFE Full-Face Flange Gaskets & 316 SS Hex Bolt/Nut Kit"
	}

	items := []model.BOMItem{
		{
			Category:    "Stage 1: Quench Vessel",
			Name:        quenchPart.Name,
			Material:    quenchPart.Material,
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("[%s] %s (%s)", quenchPart.SKU, quenchPart.Notes, quenchPart.Source),
			UnitCostUSD: quenchPart.UnitCostUSD,
			WeightKg:    quenchPart.WeightKg,
		},
		{
			Category:    "Stage 2: Wet Venturi",
			Name:        ventName,
			Material:    ventMat,
			Quantity:    1,
			Unit:        "assy",
			Spec:        ventSpec,
			UnitCostUSD: ventCost,
			WeightKg:    ventWeight,
		},
		{
			Category:    "Stage 3: Packed-Bed Column",
			Name:        colPart.Name,
			Material:    colPart.Material,
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("[%s] %s (%s)", colPart.SKU, colPart.Notes, colPart.Source),
			UnitCostUSD: colPart.UnitCostUSD,
			WeightKg:    colPart.WeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Column",
			Name:        sumpPart.Name,
			Material:    sumpPart.Material,
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("[%s] %s (%s)", sumpPart.SKU, sumpPart.Notes, sumpPart.Source),
			UnitCostUSD: sumpPart.UnitCostUSD,
			WeightKg:    sumpPart.WeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Column",
			Name:        packingPart.Name,
			Material:    packingPart.Material,
			Quantity:    packCuFt,
			Unit:        "cu.ft.",
			Spec:        fmt.Sprintf("[%s] %s (%s)", packingPart.SKU, packingPart.Notes, packingPart.Source),
			UnitCostUSD: packingPart.UnitCostUSD,
			WeightKg:    round2(packCuFt * packingPart.WeightKg),
		},
		{
			Category:    "Stage 3: Packed-Bed Column",
			Name:        "3-Pass Chevron Mist Eliminator Pad & Packing Support Grate",
			Material:    "Polypropylene (PP) / FRP Grating",
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf(`Cut to %.1f" ID column bore; captures >99%% entrained mist droplets`, actualColIDIn),
			UnitCostUSD: demisterCost,
			WeightKg:    demisterWeight,
		},
		{
			Category:    "Liquid & Chemical Handling",
			Name:        nozzlePart.Name,
			Material:    nozzlePart.Material,
			Quantity:    nozzleQty,
			Unit:        "ea",
			Spec:        fmt.Sprintf("[%s] %d quench + 2 venturi throat nozzles; %s (%s)", nozzlePart.SKU, int(nozzleQty)-2, nozzlePart.Notes, nozzlePart.Source),
			UnitCostUSD: nozzlePart.UnitCostUSD,
			WeightKg:    round2(nozzleQty * nozzlePart.WeightKg),
		},
		{
			Category:    "Liquid & Chemical Handling",
			Name:        "Seal-Less Mag-Drive Recirculation Pump & Ash Strainer",
			Material:    "Polypropylene / EPDM / Ceramic Shaft",
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("Rated %.1f GPM @ 35 ft TDH (March / Iwaki / Little Giant TE-MD) + 50-mesh inline Y-strainer", recircGPM),
			UnitCostUSD: pumpCost,
			WeightKg:    pumpWeight,
		},
		{
			Category:    "Stage 4: Blower & VFD",
			Name:        blowerPart.Name,
			Material:    blowerPart.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s (%s)", blowerPart.SKU, blowerPart.Notes, blowerPart.Source),
			UnitCostUSD: blowerPart.UnitCostUSD,
			WeightKg:    blowerPart.WeightKg,
		},
	}

	if silencerPtr != nil {
		items = append(items, model.BOMItem{
			Category:    "Stage 4: Blower & VFD",
			Name:        silencerPtr.Name,
			Material:    silencerPtr.Material,
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("[%s] %s (%s)", silencerPtr.SKU, silencerPtr.Notes, silencerPtr.Source),
			UnitCostUSD: silencerPtr.UnitCostUSD,
			WeightKg:    silencerPtr.WeightKg,
		})
	}

	if diluterPtr != nil {
		items = append(items, model.BOMItem{
			Category:    "Stage 4: Blower & VFD",
			Name:        diluterPtr.Name,
			Material:    diluterPtr.Material,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("[%s] %s (%s)", diluterPtr.SKU, diluterPtr.Notes, diluterPtr.Source),
			UnitCostUSD: diluterPtr.UnitCostUSD,
			WeightKg:    diluterPtr.WeightKg,
		})
	}

	items = append(items,
		model.BOMItem{
			Category:    "Safety & Instrumentation",
			Name:        sensorPart.Name,
			Material:    sensorPart.Material,
			Quantity:    1,
			Unit:        "kit",
			Spec:        fmt.Sprintf("[%s] %s (%s)", sensorPart.SKU, sensorPart.Notes, sensorPart.Source),
			UnitCostUSD: sensorPart.UnitCostUSD,
			WeightKg:    sensorPart.WeightKg,
		},
		model.BOMItem{
			Category:    "Hardware & Plumbing",
			Name:        hardwareName,
			Material:    "316 SS / PTFE / EPDM / CPVC",
			Quantity:    1,
			Unit:        "lot",
			Spec:        "High-temp silicone/PTFE flue gaskets, band clamps/bolts, CPVC Sch 80 primer & cement, drain ball valve, pH test strips",
			UnitCostUSD: hardwareCost,
			WeightKg:    8.5,
		},
	)

	title := fmt.Sprintf("Studio COTS Hardware Bill of Materials — %s (%.1f cu.ft.)",
		project.StudioName, design.Input.KilnVolumeCuFt)
	bom := model.BOM{
		Title: title,
		Items: items,
		Notes: []string{
			"All line-item WeightKg values represent TOTAL line-item weight (already multiplied by Quantity).",
			"COTS pricing reflects realistic US studio/industrial distributor budgetary estimates (excluding freight and tax).",
			fmt.Sprintf("Packed column snapped=%v -> %.1f\" actual ID operating at %.1f%% of Eckert flooding velocity (design limit <= 75%%).", project.SnapToCOTS, actualColIDIn, actualPctFlood),
			"Stage 3 CPVC/HDPE components MUST be protected by the hardwired 210 F / 0.25 GPM fail-open hot-gas bypass damper.",
		},
	}
	bom.RecalculateTotals()

	return COTSSummary{
		Snapped:              project.SnapToCOTS,
		QuenchVessel:         quenchPart,
		PackedColumn:         colPart,
		SumpTank:             sumpPart,
		SprayNozzle:          nozzlePart,
		PackingMedia:         packingPart,
		BlowerVFD:            blowerPart,
		Silencer:             silencerPtr,
		PlumeDiluter:         diluterPtr,
		SensorKit:            sensorPart,
		ActualPercentOfFlood: actualPctFlood,
		EstimatedTotalUSD:    round2(bom.TotalCostUSD),
		BOM:                  bom,
	}
}
