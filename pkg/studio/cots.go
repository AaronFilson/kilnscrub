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

// Schedule 80 CPVC pipe nominal sizes (ASTM F441, vacuum-rated for -20 inWC ID suction)
// up to 16", transitioning to vacuum-rated filament-wound FRP for 18"-30" columns
// (open-head 55-gal HDPE drums are restricted strictly to atmospheric vented sump tanks).
var packedCPVCCatalog = []pipeSpec{
	{6, 5.761, "CPVC80-06", 185, 3.6},
	{8, 7.625, "CPVC80-08", 275, 5.3},
	{10, 9.564, "CPVC80-10", 395, 7.5},
	{12, 11.376, "CPVC80-12", 525, 10.1},
	{14, 12.500, "CPVC80-14", 665, 11.9},
	{16, 14.314, "CPVC80-16", 835, 14.6},
	{18, 18.000, "FRP-18", 1120, 11.5},
	{24, 24.000, "FRP-24", 1590, 15.8},
	{30, 30.000, "FRP-30", 2150, 20.5},
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
		if spec.skuSuffix == "FRP-18" || spec.skuSuffix == "FRP-24" || spec.skuSuffix == "FRP-30" {
			cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
			weight := round2(spec.kgPerFt * heightFt)
			part = COTSPart{
				SKU:         "COL-" + spec.skuSuffix,
				Name:        fmt.Sprintf(`%.0f" Vacuum-Rated Filament-Wound FRP Absorber Column (%.1f ft)`, spec.nominalIn, heightFt),
				NominalIn:   spec.nominalIn,
				Material:    "Filament-Wound Vinyl-Ester FRP (ASTM E84 Class 1, Vacuum-Rated to -25 inWC)",
				Source:      "Ershigs / Beetle Plastics / US Plastic Industrial FRP",
				UnitCostUSD: cost,
				WeightKg:    weight,
				Notes:       fmt.Sprintf("Vacuum-rated FRP column (true bore %.1f\", snapped from %.1f\") avoids thin-wall HDPE drum collapse under -20 inWC suction at 184 F; protected by 200 F hardwired snap-disc + 210 F software interlock.", spec.actualID, reqIn),
			}
		} else {
			cost := round2(spec.baseCost * math.Max(0.85, heightFt/8.0))
			weight := round2(spec.kgPerFt * heightFt)
			part = COTSPart{
				SKU:         "HARVEL-" + spec.skuSuffix,
				Name:        fmt.Sprintf(`%.0f" Sch 80 CPVC Pipe Column (ID %.2f", %.1f ft)`, spec.nominalIn, spec.actualID, heightFt),
				NominalIn:   spec.nominalIn,
				Material:    "Schedule 80 CPVC (ASTM F441, ASTM E84 Class 1, 200 F max, Vacuum-Rated)",
				Source:      "SupplyHouse / US Plastic / McMaster-Carr",
				UnitCostUSD: cost,
				WeightKg:    weight,
				Notes:       fmt.Sprintf("Vacuum-rated Sch 80 CPVC pipe (true bore %.2f\", snapped from %.1f\"); requires independent hardwired 200 F manual-reset snap-disc + 210 F software backup bypass interlock.", spec.actualID, reqIn),
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
			Notes:       fmt.Sprintf("Atmospheric vented sump ONLY (never under blower vacuum); covers %.1f gal required holdup with 100%% ash-settling & NaHCO3 mixing reserve.", sumpGal),
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
			Notes:       fmt.Sprintf("Atmospheric vented sump ONLY (never under blower vacuum); covers %.1f gal required holdup with easy top access for baking soda dosing and ash basket.", sumpGal),
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
			Notes:       fmt.Sprintf("Atmospheric vented co-op sump for %.1f gal holdup with bulkheads for recirculation & cooled blowdown.", sumpGal),
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

	var voltageStr, vfdSKU string
	switch {
	case hp <= 0.5:
		voltageStr = "120V 1-Ph Input VFD in NEMA 4X Enclosure"
		vfdSKU = "GS21-10P5"
	case hp <= 1.0:
		if project.AvailableVoltage >= 240 {
			voltageStr = "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure"
			vfdSKU = "GS21-21P0"
		} else {
			voltageStr = "120V 1-Ph Input (30A) VFD in NEMA 4X Enclosure"
			vfdSKU = "GS21-11P0"
		}
	case hp <= 2.0:
		voltageStr = "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure"
		vfdSKU = "GS21-22P0"
	case hp <= 3.0:
		voltageStr = "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure"
		vfdSKU = "GS21-23P0"
	default:
		voltageStr = "240V 3-Ph / Derated 1-Ph VFD in NEMA 4X Enclosure"
		vfdSKU = "GS21-25P0"
	}

	const nema4XEnclosureCost = 165.0 // NEC 110.28 NEMA 4X fiberglass/polycarbonate housing for IP20 GS21 (or upgrade to GS21X washdown)
	var blowerCost, vfdCost, weightKg float64
	switch {
	case hp <= 0.5:
		blowerCost, vfdCost, weightKg = 640, 175+nema4XEnclosureCost, 25.0
	case hp <= 1.0:
		blowerCost, vfdCost, weightKg = 860, 215+nema4XEnclosureCost, 32.0
	case hp <= 1.5:
		blowerCost, vfdCost, weightKg = 1120, 265+nema4XEnclosureCost, 39.0
	case hp <= 2.0:
		blowerCost, vfdCost, weightKg = 1340, 310+nema4XEnclosureCost, 47.0
	case hp <= 3.0:
		blowerCost, vfdCost, weightKg = 1720, 395+nema4XEnclosureCost, 61.0
	default:
		blowerCost = round2(1720.0 + 360.0*(hp-3.0))
		vfdCost = round2(395.0 + nema4XEnclosureCost + 110.0*(hp-3.0))
		weightKg = round2(61.0 + 12.0*(hp-3.0))
	}

	mat := "Powder-Coated Steel Housing / 316L SS Radial Wheel + IP20 GS21 VFD inside NEMA 4X Enclosure"
	sku := fmt.Sprintf("PB-RAD-%.2gHP-%s-N4X", hp, vfdSKU)
	source := "Grainger (Dayton/Cincinnati Fan) + AutomationDirect"
	if heavy {
		blowerCost = round2(blowerCost * 1.45)
		mat = "Vinyl-Ester FRP Housing & Chemical-Duty Radial Wheel + IP20 GS21 VFD inside NEMA 4X Enclosure"
		sku = fmt.Sprintf("FRP-RAD-%.2gHP-%s-N4X", hp, vfdSKU)
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
		Notes:       fmt.Sprintf("High-head radial pressure blower sized for %.0f ACFM @ %.1f inWC static pressure; %s VFD (IP20 housed in NEC 110.28 NEMA 4X enclosure, or upgrade to standalone GS21X washdown VFD) accepts 0-10V / Modbus PID draft control.", design.Blower.FlowACFM, design.Blower.TotalStaticPressureInWC, vfdSKU),
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

func selectPlumeDiluter(project StudioProject, design model.ScrubberDesign) COTSPart {
	cfm, coneIn := RecommendedDilutionCFM(design)
	var cost, weight, amps float64
	switch {
	case cfm <= 500.0:
		cost, weight, amps = 275.0, 11.5, 2.2
	case cfm <= 1000.0:
		cost, weight, amps = 385.0, 16.5, 4.1
	case cfm <= 1500.0:
		cost, weight, amps = 495.0, 22.0, 6.5
	case cfm <= 2000.0:
		cost, weight, amps = 595.0, 27.0, 8.8
	default:
		cost, weight, amps = 765.0, 34.0, 12.5
	}
	teeMat := "Dry Clean-Air Forward-Curved Blower + CPVC / Vinyl-Ester FRP / 316L SS Stack Mixing Tee & Taper Cone (No Galvanized Zinc in Wet Exhaust)"
	if project.VaporGlazeMode == VaporGlazeSalt {
		teeMat = "Dry Clean-Air Intake Blower + Acid-Resistant Sch 80 CPVC / Vinyl-Ester FRP / 316L SS Stack Mixing Tee & Taper Cone (Zinc-Free for Salt HCl)"
	}
	return COTSPart{
		SKU:         fmt.Sprintf("SQ-DILUTE-%.0fCFM-%.0fIN", cfm, coneIn),
		Name:        fmt.Sprintf(`%.0f CFM 120V Squirrel-Cage Plume Dilution Blower + CPVC/FRP/316L Mixing Tee & %.0f" HV Exit Cone`, cfm, coneIn),
		NominalIn:   coneIn,
		Material:    teeMat,
		Source:      "Grainger (Dayton 1TDT/4C) + US Plastic / SupplyHouse CPVC/FRP",
		UnitCostUSD: cost,
		WeightKg:    weight,
		Notes:       fmt.Sprintf("120V (~%.1f A, 0.7-1.2 inWC) mounted DOWNSTREAM of Stage 4 ID blower (never in place of the %.1f inWC radial blower). Wetted stack mixing tee & %.0f\" high-velocity exit cone (~3,400 FPM) use CPVC/FRP/316L (never galvanized steel, which corrodes in HCl condensate); add optional cold-intake kiln-jacket warm-air shroud for near-zero visible plume.", amps, design.Blower.TotalStaticPressureInWC, coneIn),
	}
}

func selectSensorKit(design model.ScrubberDesign) COTSPart {
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()))
	damperCost := round2(265.0 + 22.0*ductIn)
	hardwiredSafetyCost := 228.0 // 200 F manual-reset snap-disc ($35) + 24VDC latching safety relay ($125) + NEMA 4X E-stop/reset/strobe ($68)
	return COTSPart{
		SKU:         "KIT-MODBUS-SAFE-V2",
		Name:        "Modbus RTU Sensor, Hardwired 200 F NFPA 86 Safety Relay & Fail-Safe Bypass Kit",
		NominalIn:   ductIn,
		Material:    "316L SS Thermowells / NEMA 4X / 24VDC Latching Safety Relay / Spring-Return Damper",
		Source:      "AutomationDirect / Dwyer / Belimo / Honeywell / Atlas Scientific",
		UnitCostUSD: round2(415.0 + damperCost + hardwiredSafetyCost),
		WeightKg:    13.5,
		Notes:       fmt.Sprintf(`Includes independent NFPA 86 hardwired 200 F manual-reset bimetallic snap-disc + 24VDC de-energize-to-trip latching safety relay + NEMA 4X E-Stop, Type-K flue TC, Pt100 quench RTD (210 F software backup), Dwyer +/-2 inWC draft transmitter, 0.25 GPM paddle flow switch, pH probe, isolated RS-485 USB adapter, and %.0f" spring-return fail-open bypass damper.`, ductIn),
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
		d := selectPlumeDiluter(project, design)
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
			Category:    "Liquid & Chemical Handling",
			Name:        "IPC Section 608 Potable Water Backflow Prevention (Air-Gap Break Tank or ASSE 1013 RPZ)",
			Material:    "Lead-Free Bronze RPZ / HDPE Break Tank with >=1\" Air Gap",
			Quantity:    1,
			Unit:        "ea",
			Spec:        "[IPC608-RPZ-AG] Mandatory IPC 608 / 608.15.1 high-hazard potable water isolation (engineered >=1\" vertical air gap above overflow rim OR certified ASSE 1013 RPZ backflow preventer) (SupplyHouse / Watts / Zurn)",
			UnitCostUSD: 185.0,
			WeightKg:    3.5,
		},
		model.BOMItem{
			Category:    "Safety & Instrumentation",
			Name:        "NEC 430.102 Lockable NEMA 4X Local Disconnect Switch & 3% VFD Input Line Reactor",
			Material:    "NEMA 4X Non-Fused Disconnect / Copper-Wound Line Reactor",
			Quantity:    1,
			Unit:        "assy",
			Spec:        "[NEC430-DISC-LR] Padlockable line-side disconnect within sight (<=50 ft) of VFD & blower + 3% input line reactor for NEC 210.8(F) HF-GFCI stability (Square D / AutomationDirect)",
			UnitCostUSD: 215.0,
			WeightKg:    4.8,
		},
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
			Material:    "316 SS / Ceramic Fiber / PTFE / EPDM / CPVC",
			Quantity:    1,
			Unit:        "lot",
			Spec:        ">=2300 F alumina-silica ceramic fiber rope/gaskets for dry hot flue upstream of Stage 1 (NO silicone/PTFE on 2400 F flue!), expanded PTFE/EPDM gaskets strictly downstream (<=200 F), band clamps/bolts, CPVC Sch 80 primer & cement, drain ball valve, thermometer (IPC 702.5 <=140 F check), pH test strips",
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
			"NFPA 86 & ASTM F441: Stage 3 Sch 80 CPVC/FRP components MUST be protected by an independent hardwired 200 F manual-reset snap-disc + 24VDC de-energize-to-trip latching safety relay driving the spring-return fail-open hot-gas bypass damper (ahead of the 210 F / 0.25 GPM software backup).",
			"NEC 430.102 / 430.122 / 210.8(F) / 110.28: Size VFD branch circuit at >=125% VFD rated 1-phase input current, install lockable NEMA 4X sight disconnect (<=50 ft), house IP20 VFD in NEMA 4X enclosure, and use line-side HF-rated GFCI + input line reactor.",
			"IPC 608 / 702.5 / 803.1 & NFPA 91: Isolate potable water via Air Gap (>=1\") or ASSE 1013 RPZ; cool drain/blowdown <=140 F and neutralize to pH 6.5-8.5 (>=5.0 per 40 CFR 403.5) before sanitary discharge; include NFPA 91 duct cleanouts/low-point drains and ASTM E84 Class 1 plastic ducting.",
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
