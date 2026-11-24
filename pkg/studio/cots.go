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
// up to 16", transitioning to vacuum-rated filament-wound FRP for 18"-30" columns.
// Sizes >= 12" include realistic industrial Van Stone ANSI flanges and LTL pallet freight.
// (Open-head 55-gal HDPE drums are restricted strictly to atmospheric vented sump tanks.)
var packedCPVCCatalog = []pipeSpec{
	{6, 5.761, "CPVC80-06", 215, 3.6},
	{8, 7.625, "CPVC80-08", 345, 5.3},
	{10, 9.564, "CPVC80-10", 545, 7.5},
	{12, 11.376, "CPVC80-12", 890, 10.1},
	{14, 12.500, "CPVC80-14", 1245, 11.9},
	{16, 14.314, "CPVC80-16", 1595, 14.6},
	{18, 18.000, "FRP-18", 1890, 11.5},
	{24, 24.000, "FRP-24", 2490, 15.8},
	{30, 30.000, "FRP-30", 3250, 20.5},
}

var packedFRPCatalog = []pipeSpec{
	{6, 6.00, "FRP-06", 460, 3.4},
	{8, 8.00, "FRP-08", 620, 4.6},
	{10, 10.00, "FRP-10", 840, 5.9},
	{12, 12.00, "FRP-12", 1180, 7.4},
	{14, 14.00, "FRP-14", 1540, 8.8},
	{16, 16.00, "FRP-16", 1860, 10.4},
	{18, 18.00, "FRP-18", 2150, 12.0},
	{24, 24.00, "FRP-24", 2890, 16.5},
	{30, 30.00, "FRP-30", 3740, 21.5},
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

	if project.IsDeluxe() {
		nomIn := round1(reqIn)
		if project.SnapToCOTS {
			if heavy {
				spec, _ := snapPipe(quenchC276Catalog, reqIn)
				nomIn = spec.nominalIn
			} else {
				spec, _ := snapPipe(quench316LCatalog, reqIn)
				nomIn = spec.nominalIn
			}
		}
		if heavy {
			cost := round2(2450.0 + 165.0*nomIn + 145.0*heightFt)
			weight := round2(math.Max(45.0, 2.20*nomIn*heightFt))
			return COTSPart{
				SKU:         fmt.Sprintf("DLX-QC276-%02.0f", nomIn),
				Name:        fmt.Sprintf(`%.0f" TIG-Welded Heavy-Gauge Alloy C-276 Quench Tower w/ ANSI B16.5 Class 150 Flanges (%.1f ft)`, nomIn, heightFt),
				NominalIn:   nomIn,
				Material:    "Heavy-Gauge Alloy C-276 (11 ga / Sch 10S, TIG Back-Purged) + ANSI B16.5 Class 150 Lap-Joint Flanges + PTFE Envelope Gaskets",
				Source:      "Custom TIG Process Fabricator (scrubber.dxf + ANSI B16.5)",
				UnitCostUSD: cost,
				WeightKg:    weight,
				Notes:       fmt.Sprintf("Custom Fab Deluxe quench tower (%.0f\" nom, calc %.1f\"): TIG back-purged Alloy C-276 shell, machined ANSI/ASME B16.5 Class 150 lap-joint flanges, PTFE envelope gaskets, 2× borosilicate sight-glass ports, 2× bolted NFPA 91 quick-release cleanout doors, and 1\" ceramic-fiber + stainless hot-flue insulation jacket.", nomIn, reqIn),
			}
		}
		cost := round2(1150.0 + 58.0*nomIn + 65.0*heightFt)
		weight := round2(math.Max(28.0, 1.45*nomIn*heightFt))
		notes := fmt.Sprintf("Custom Fab Deluxe quench tower (%.0f\" nom, calc %.1f\"): TIG back-purged heavy-gauge 316L SS shell, machined ANSI/ASME B16.5 Class 150 lap-joint flanges, PTFE envelope gaskets, 2× borosilicate sight-glass ports, 2× bolted NFPA 91 quick-release cleanout doors, and 1\" ceramic-fiber + stainless hot-flue insulation jacket.", nomIn, reqIn)
		if project.VaporGlazeMode == VaporGlazeSalt {
			notes += " Recirculate alkaline NaHCO3 quench spray (pH 8.0-8.5) and inspect back-purged TIG welds every 5-10 salt firings."
		}
		return COTSPart{
			SKU:         fmt.Sprintf("DLX-Q316L-%02.0f", nomIn),
			Name:        fmt.Sprintf(`%.0f" TIG-Welded Heavy-Gauge 316L SS Quench Tower w/ ANSI B16.5 Class 150 Flanges (%.1f ft)`, nomIn, heightFt),
			NominalIn:   nomIn,
			Material:    "Heavy-Gauge 316L Stainless Steel (11 ga / Sch 10S, TIG Back-Purged) + ANSI B16.5 Class 150 Lap-Joint Flanges + PTFE Envelope Gaskets",
			Source:      "Custom TIG Process Fabricator (scrubber.dxf + ANSI B16.5)",
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       notes,
		}
	}

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
	return selectSinglePackedColumn(project, design, EductorAnalysis{})
}

func selectSinglePackedColumn(project StudioProject, design model.ScrubberDesign, eductor EductorAnalysis) (COTSPart, float64, float64) {
	reqIn := design.PackedBed.ColumnDiameter.Inches()
	if eductor.Enabled && eductor.EnlargedBedDiameterIn > reqIn {
		reqIn = eductor.EnlargedBedDiameterIn - 0.5 // snap cleanly to enlarged eductor nominal bore
	}
	heightFt := design.PackedBed.ColumnHeight.Feet()
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP

	if project.IsDeluxe() {
		var actualIDIn, nomIn, cost, weight float64
		var sku string
		if !project.SnapToCOTS {
			if eductor.Enabled && eductor.EnlargedBedDiameterIn > reqIn {
				reqIn = eductor.EnlargedBedDiameterIn
			}
			nomIn = round1(reqIn)
			actualIDIn = reqIn
			sku = fmt.Sprintf("DLX-FAB-FRP-%02.0f", math.Round(reqIn))
			cost = round2(1080.0 + 78.0*reqIn + 60.0*heightFt)
			weight = round2(math.Max(28.0, 1.10*reqIn*heightFt))
		} else {
			var spec pipeSpec
			if eductor.Enabled {
				spec, _ = snapPipeByNominal(packedFRPCatalog, eductor.EnlargedBedDiameterIn)
			} else {
				spec, _ = snapPipe(packedFRPCatalog, reqIn)
			}
			nomIn = spec.nominalIn
			actualIDIn = spec.actualID
			sku = "DLX-FRP-COL-" + spec.skuSuffix
			cost = round2(spec.baseCost*math.Max(0.95, heightFt/8.0) + 460.0) // +structural gusset rings, bolted 12" packing manway & ANSI stub-end flanges
			weight = round2(spec.kgPerFt*heightFt + 9.5)
		}
		part := COTSPart{
			SKU:         sku,
			Name:        fmt.Sprintf(`%.0f" Custom Filament-Wound Derakane Vinyl-Ester FRP Packed Column w/ Bolted Manway & Gusset Rings (%.1f ft)`, nomIn, heightFt),
			NominalIn:   nomIn,
			Material:    "Filament-Wound Derakane Vinyl-Ester FRP (210 F Wet Acid Rated, ASTM E84 Class 1, Structural Gusset Rings, ANSI Class 150 Flanges)",
			Source:      "Ershigs / Beetle Plastics / Custom Industrial FRP Fabricator",
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       fmt.Sprintf("Custom Fab Deluxe absorber column (ID %.1f\", calc %.1f\") with structural gusset rings, bolted 12\" packing access manway, borosilicate bed sight-glass port, and removable demister cassette flange (alternative: heavy-wall welded PP/CPVC with structural gusset rings).", actualIDIn, reqIn),
		}
		if eductor.Enabled {
			part.Notes += fmt.Sprintf(" Enlarged for dry-fan eductor mixed airflow (%.0f ACFM @ %.1f F, +%.2f inWC positive pressure, outdoor-only bolted ANSI PTFE flanges).",
				eductor.TotalMixedACFM, eductor.MixedTempF, eductor.PositivePressureInWC)
			return part, actualIDIn, eductor.EnlargedBedFloodPct
		}
		actualDiamM := float64(model.InchesToMeters(actualIDIn))
		actualArea := math.Pi * actualDiamM * actualDiamM / 4.0
		actualPctFlood := design.PackedBed.PercentOfFlood
		if actualArea > 0 && design.PackedBed.FloodingVelocity > 0 && design.QuenchedGas.VolumetricFlow > 0 {
			vSupActual := float64(design.QuenchedGas.VolumetricFlow) / actualArea
			actualPctFlood = round2((vSupActual / float64(design.PackedBed.FloodingVelocity)) * 100.0)
		}
		return part, actualIDIn, actualPctFlood
	}

	if !project.SnapToCOTS {
		if eductor.Enabled && eductor.EnlargedBedDiameterIn > reqIn {
			reqIn = eductor.EnlargedBedDiameterIn
		}
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
		pctFlood := design.PackedBed.PercentOfFlood
		if eductor.Enabled {
			pctFlood = eductor.EnlargedBedFloodPct
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
		return part, reqIn, pctFlood
	}

	var spec pipeSpec
	var part COTSPart
	if heavy {
		if eductor.Enabled {
			spec, _ = snapPipeByNominal(packedFRPCatalog, eductor.EnlargedBedDiameterIn)
		} else {
			spec, _ = snapPipe(packedFRPCatalog, reqIn)
		}
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
		if eductor.Enabled {
			spec, _ = snapPipeByNominal(packedCPVCCatalog, eductor.EnlargedBedDiameterIn)
		} else {
			spec, _ = snapPipe(packedCPVCCatalog, reqIn)
		}
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
				Notes:       fmt.Sprintf("Vacuum-rated FRP column (true bore %.1f\", snapped from %.1f\") avoids thin-wall HDPE drum collapse; protected by 200 F hardwired snap-disc + 210 F software interlock.", spec.actualID, reqIn),
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

	if eductor.Enabled {
		part.Notes += fmt.Sprintf(" Enlarged for dry-fan eductor mixed airflow (%.0f ACFM @ %.1f F, +%.2f inWC positive pressure, outdoor-only bolted flanges).",
			eductor.TotalMixedACFM, eductor.MixedTempF, eductor.PositivePressureInWC)
		return part, spec.actualID, eductor.EnlargedBedFloodPct
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

func snapPipeByNominal(catalog []pipeSpec, nomIn float64) (pipeSpec, bool) {
	for _, spec := range catalog {
		if spec.nominalIn >= nomIn-1e-6 {
			return spec, true
		}
	}
	return catalog[len(catalog)-1], false
}

func selectSumpTank(design model.ScrubberDesign) COTSPart {
	return selectSumpTankForProject(StudioProject{}, design)
}

func selectSumpTankForProject(project StudioProject, design model.ScrubberDesign) COTSPart {
	sumpGal := float64(design.PackedBed.SumpVolume) / model.CubicMetersPerUSGallon
	if project.IsDeluxe() {
		switch {
		case sumpGal <= 22.0:
			return COTSPart{
				SKU:         "DLX-SUMP-SLOPED-35",
				Name:        `35-Gallon Sloped-Bottom Baffled Chemical Settling Sump w/ 2" Drain Ball Valve & Ash Basket`,
				NominalIn:   20.0,
				Material:    "Welded Heavy-Wall HDPE / Derakane FRP (Sloped-Bottom Baffled, Atmospheric Vented)",
				Source:      "Custom Plastics / US Plastic Industrial Tank",
				UnitCostUSD: 385.0,
				WeightKg:    18.0,
				Notes:       fmt.Sprintf("Custom Fab Deluxe atmospheric vented sloped-bottom settling sump covering %.1f gal required holdup; includes internal weir baffle, removable 316L ash basket, flooded-suction bulkhead, and 2\" full-port drain ball valve.", sumpGal),
			}
		case sumpGal <= 45.0:
			return COTSPart{
				SKU:         "DLX-SUMP-SLOPED-60",
				Name:        `60-Gallon Sloped-Bottom Baffled Chemical Settling Sump w/ 2" Drain Ball Valve & Ash Basket`,
				NominalIn:   24.0,
				Material:    "Welded Heavy-Wall HDPE / Derakane FRP (Sloped-Bottom Baffled, Atmospheric Vented)",
				Source:      "Custom Plastics / US Plastic Industrial Tank",
				UnitCostUSD: 495.0,
				WeightKg:    24.0,
				Notes:       fmt.Sprintf("Custom Fab Deluxe atmospheric vented sloped-bottom settling sump covering %.1f gal required holdup; includes internal weir baffle, removable 316L ash basket, flooded-suction bulkhead, and 2\" full-port drain ball valve.", sumpGal),
			}
		default:
			return COTSPart{
				SKU:         "DLX-SUMP-SLOPED-110",
				Name:        `110-Gallon Sloped-Bottom Baffled Chemical Settling Sump w/ 2" Drain Ball Valve & Ash Basket`,
				NominalIn:   30.0,
				Material:    "Welded Heavy-Wall XLPE / Derakane FRP (Sloped-Bottom Baffled, Atmospheric Vented)",
				Source:      "Custom Industrial FRP / Ace Roto-Mold",
				UnitCostUSD: 745.0,
				WeightKg:    38.0,
				Notes:       fmt.Sprintf("Custom Fab Deluxe atmospheric vented sloped-bottom co-op settling sump covering %.1f gal required holdup; includes internal weir baffle, removable 316L ash basket, flooded-suction bulkhead, and 2\" full-port drain ball valve.", sumpGal),
			}
		}
	}
	switch {
	case sumpGal <= 22.0:
		return COTSPart{
			SKU:         "ULINE-S-16913",
			Name:        "30-Gallon Open-Head HDPE Baffled Chemical Sump Drum",
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
			Name:        "55-Gallon Open-Head HDPE Baffled Chemical Sump Drum",
			NominalIn:   22.5,
			Material:    "High-Density Polyethylene (HDPE)",
			Source:      "Uline / Grainger / McMaster-Carr",
			UnitCostUSD: 96.0,
			WeightKg:    10.5,
			Notes:       fmt.Sprintf("Atmospheric vented sump ONLY (never under blower vacuum); covers %.1f gal required holdup with easy top access for bulk baking soda dosing and ash basket.", sumpGal),
		}
	default:
		return COTSPart{
			SKU:         "USPLASTIC-100T",
			Name:        "100-Gallon Heavy-Duty Poly Baffled Chemical Sump Tank",
			NominalIn:   28.0,
			Material:    "Cross-Linked HDPE (XLPE)",
			Source:      "US Plastic Corp / Ace Roto-Mold",
			UnitCostUSD: 315.0,
			WeightKg:    22.0,
			Notes:       fmt.Sprintf("Atmospheric vented co-op sump for %.1f gal holdup with bulkheads for recirculation & cooled blowdown.", sumpGal),
		}
	}
}

func dualSumpExtraCost(design model.ScrubberDesign) float64 {
	sump := selectSumpTank(design)
	recircGPM := design.PackedBed.RecirculationFlowGPM
	if recircGPM <= 0 {
		recircGPM = 4.0
	}
	extraPump := round2(220.0 + 10.0*recircGPM)
	return round2(sump.UnitCostUSD + extraPump + 45.0)
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
		Source:      "McMaster-Carr / Jaeger / US Plastic",
		UnitCostUSD: 88.0,
		WeightKg:    2.5,
		Notes:       fmt.Sprintf("Order %.0f cu.ft. to fill %.1f\" ID column to %.1f\" bed depth (%.2f m). Select aquaculture bio-balls ($38/cu.ft.) or DIY kiln-fired stoneware rings ($18/cu.ft.) to save.", packCuFt, actualColIDIn, design.PackedBed.PackingHeight.Inches(), float64(design.PackedBed.PackingHeight)),
	}
	return part, packCuFt
}

func vfdSpecForHP(project StudioProject, hp float64) (voltageStr, vfdSKU string) {
	switch {
	case hp <= 0.5:
		return "120V 1-Ph Input VFD in NEMA 4X Enclosure", "GS21-10P5"
	case hp <= 1.0:
		if project.AvailableVoltage >= 240 {
			return "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure", "GS21-21P0"
		}
		return "120V 1-Ph Input (30A) VFD in NEMA 4X Enclosure", "GS21-11P0"
	case hp <= 2.0:
		return "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure", "GS21-22P0"
	case hp <= 3.0:
		return "240V 1-Ph Input -> 3-Ph Motor VFD in NEMA 4X Enclosure", "GS21-23P0"
	default:
		return "240V 3-Ph / Derated 1-Ph VFD in NEMA 4X Enclosure", "GS21-25P0"
	}
}

func selectBlowerVFD(project StudioProject, design model.ScrubberDesign) COTSPart {
	return selectWetBlowerVFD(project, design)
}

func selectWetBlowerVFD(project StudioProject, design model.ScrubberDesign) COTSPart {
	hp := design.Blower.RecommendedMotorHP
	if hp <= 0 {
		hp = 1.0
	}
	heavy := project.CorrosionTier == CorrosionTierHeavyC276FRP
	voltageStr, vfdSKU := vfdSpecForHP(project, hp)

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

func selectDryMotiveBlowerVFD(project StudioProject, eductor EductorAnalysis) COTSPart {
	hp := eductor.DryFanMotorHP
	if hp <= 0 {
		hp = 1.5
	}
	voltageStr, vfdSKU := vfdSpecForHP(project, hp)
	fanCost, vfdCost, weightKg := dryFanUnitCost(hp)
	wheelIn := round1(math.Max(8.0, 6.5+2.0*math.Sqrt(hp)))
	return COTSPart{
		SKU:         fmt.Sprintf("DRY-EDFAN-%.2gHP-%s-N4X", hp, vfdSKU),
		Name:        fmt.Sprintf("%.2g HP Clean Dry-Air Motive Radial Blower + %s", hp, voltageStr),
		NominalIn:   wheelIn,
		Material:    "Carbon Steel / Cast Aluminum Clean-Air Radial Blower (Zero Acid Contact) + IP20 GS21 VFD inside NEMA 4X Enclosure",
		Source:      "Grainger (Dayton/Cincinnati Fan Clean-Air) + AutomationDirect",
		UnitCostUSD: round2(fanCost + vfdCost),
		WeightKg:    weightKg,
		Notes:       fmt.Sprintf("Handles 100%% dry ambient studio air (%.0f CFM @ %.1f inWC motive pressure) pushing downward into vertical static venturi eductor — eliminates corrosive wet-acid blower wheel & FRP scroll completely! %s VFD in NEC 110.28 NEMA 4X enclosure.", eductor.MotiveAirCFM, eductor.DryFanStaticInWC, vfdSKU),
	}
}

func selectSilencer(design model.ScrubberDesign) COTSPart {
	return selectSilencerForProject(StudioProject{}, design)
}

func selectSilencerForProject(project StudioProject, design model.ScrubberDesign) COTSPart {
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()/2.0)*2.0)
	if project.IsDeluxe() {
		cost := round2(540.0 + 34.0*ductIn)
		weight := round2(28.0 + 1.8*ductIn)
		return COTSPart{
			SKU:         fmt.Sprintf("DLX-ACOUSTIC-ENC-%02.0f", ductIn),
			Name:        fmt.Sprintf(`%.0f" Deluxe Acoustic Blower Baffle Enclosure (2" Rockwool + Cooling Labyrinth) & Tuned Absorptive Stack Silencer`, ductIn),
			NominalIn:   ductIn,
			Material:    `Powder-Coated Steel Acoustic Baffle Box (2" Rockwool Mineral Wool + Perforated Metal Liner + Cooling-Air Labyrinth) + Tuned Absorptive Stack Silencer`,
			Source:      "Custom Acoustic Fab / Rockwool Safe'n'Sound / Industrial Silencer",
			UnitCostUSD: cost,
			WeightKg:    weight,
			Notes:       "Custom Fab Deluxe acoustic package: reduces blower & discharge noise by ~22 dBA total (+9 dBA quieter than standard silencer) while maintaining baffled TEFC motor cooling airflow for ultra-quiet residential property-line operation.",
		}
	}
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
		Notes:       fmt.Sprintf("120V (~%.1f A, 0.7-1.2 inWC) mounted DOWNSTREAM of Stage 4 blower (never in place of the %.1f inWC radial blower). Wetted stack mixing tee & %.0f\" high-velocity exit cone (~3,400 FPM) use CPVC/FRP/316L (never galvanized steel, which corrodes in HCl condensate); add optional cold-intake kiln-jacket warm-air shroud for near-zero visible plume.", amps, design.Blower.TotalStaticPressureInWC, coneIn),
	}
}

func selectSensorKit(design model.ScrubberDesign) COTSPart {
	return selectSensorKitForTier(ControlTierRPiModbus, design)
}

func selectSensorKitForTier(tier string, design model.ScrubberDesign) COTSPart {
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()))
	damperCost := round2(265.0 + 22.0*ductIn)
	hardwiredSafetyCost := 228.0 // 200 F manual-reset snap-disc ($35) + 24VDC latching safety relay ($125) + NEMA 4X E-stop/reset/strobe ($68)

	if tier == ControlTierElectromechanicalOnly {
		analogGaugeAndDialCost := 85.0 // Dwyer Magnehelic 0 to -1.0 inWC dial gauge + VFD manual speed potentiometer + mechanical paddle flow switch
		return COTSPart{
			SKU:         "KIT-ELECTROMECH-SAFE-V1",
			Name:        "Electromechanical Hardwired 200 F NFPA 86 Safety Loop, Magnehelic Gauge & Fail-Safe Bypass Kit",
			NominalIn:   ductIn,
			Material:    "316L SS Snap-Disc / NEMA 4X / 24VDC Latching Safety Relay / Dwyer Magnehelic / Spring-Return Damper",
			Source:      "AutomationDirect / Dwyer / Belimo / Honeywell",
			UnitCostUSD: round2(analogGaugeAndDialCost + damperCost + hardwiredSafetyCost),
			WeightKg:    11.0,
			Notes:       fmt.Sprintf(`Zero-software electromechanical kit (saves ~$960 vs Raspberry Pi Modbus): includes mandatory NFPA 86 hardwired 200 F manual-reset bimetallic snap-disc + 0.25 GPM mechanical paddle flow switch + 24VDC de-energize-to-trip latching safety relay + NEMA 4X E-Stop, analog Dwyer Magnehelic draft dial gauge, manual VFD speed knob, and %.0f" spring-return fail-open bypass damper.`, ductIn),
		}
	}

	const rpiModbusAutomationCost = 1045.0 // RPi + DIN enclosure + 24VDC PSU + Modbus RTU I/O + digital Dwyer draft/pH/RTD/TC transmitters + modulating quench valve
	return COTSPart{
		SKU:         "KIT-MODBUS-SAFE-V2",
		Name:        "Raspberry Pi Modbus RTU Sensor, Hardwired 200 F NFPA 86 Safety Relay & Fail-Safe Bypass Kit",
		NominalIn:   ductIn,
		Material:    "316L SS Thermowells / NEMA 4X / RPi Modbus RTU / 24VDC Latching Safety Relay / Spring-Return Damper",
		Source:      "AutomationDirect / Dwyer / Belimo / Honeywell / Atlas Scientific",
		UnitCostUSD: round2(rpiModbusAutomationCost + damperCost + hardwiredSafetyCost),
		WeightKg:    14.5,
		Notes:       fmt.Sprintf(`Includes independent NFPA 86 hardwired 200 F manual-reset bimetallic snap-disc + 24VDC de-energize-to-trip latching safety relay + NEMA 4X E-Stop, Raspberry Pi + RS-485 Modbus RTU I/O, Type-K flue TC, Pt100 quench RTD (210 F software backup), Dwyer +/-2 inWC digital draft transmitter, 0.25 GPM paddle flow switch, pH probe, modulating spray valve, and %.0f" spring-return fail-open bypass damper.`, ductIn),
	}
}

func selectSensorKitForProject(project StudioProject, design model.ScrubberDesign) COTSPart {
	base := selectSensorKitForTier(project.ControlTier, design)
	if !project.IsDeluxe() {
		return base
	}
	ductIn := math.Max(6.0, math.Ceil(design.Envelope.DuctDiameter.Inches()))
	const deluxeUL508AUpgradeCost = 780.0 // UL 508A-style NEMA 4X stainless/polycarbonate pre-wired enclosure, FM/UL 1/16-DIN 200 F high-limit controller, dual Type-K thermowells, inline flow transmitter, flow-through pH probe, pilot stack light & alarm horn
	return COTSPart{
		SKU:         "DLX-UL508A-N4X-PANEL",
		Name:        "UL 508A-Style Pre-Wired NEMA 4X Control/VFD Panel, FM/UL 200 F High-Limit Safety Loop & Fail-Safe Bypass Kit",
		NominalIn:   ductIn,
		Material:    "NEMA 4X 316 SS / Polycarbonate Pre-Wired Panel / FM-UL 1/16-DIN 200 F High-Limit / Dual Type-K Thermowells / Flow Transmitter / Flow-Through pH / Stack Light & Horn",
		Source:      "UL 508A Panel Shop / AutomationDirect / Dwyer / Belimo / Atlas Scientific",
		UnitCostUSD: round2(base.UnitCostUSD + deluxeUL508AUpgradeCost),
		WeightKg:    round2(base.WeightKg + 13.5),
		Notes:       fmt.Sprintf(`Custom Fab Deluxe UL 508A-style pre-wired NEMA 4X panel (control_tier=%s): through-door NEC 430.102 lockable disconnect, 3%% input line reactor, HF-GFCI, GS21 VFD subpanel, independent FM/UL 1/16-DIN 200 F (93.3 C) high-limit controller + NFPA 86 hardwired 24VDC latching safety relay + E-stop, dual 316L Type-K thermowells, inline flow transmitter (0.25 GPM trip), flow-through pH probe, pilot stack light/alarm horn, and %.0f" Belimo spring-return fail-open bypass damper.`, project.ControlTier, ductIn),
	}
}

func estimateSingleVenturiCost(project StudioProject, design model.ScrubberDesign) float64 {
	ventCost := round2(240.0 + 18.0*design.Venturi.InletDiameter.Inches())
	switch project.BuildPath {
	case BuildPathWeldedDXF:
		ventCost = round2(ventCost * 1.15)
	case BuildPathCustomFabDeluxe:
		ventCost = round2(ventCost * 1.65)
	}
	if project.CorrosionTier == CorrosionTierHeavyC276FRP {
		ventCost = round2(ventCost * 2.6)
	}
	return ventCost
}

func selectBlowerForDrive(
	p StudioProject,
	design model.ScrubberDesign,
	eductor EductorAnalysis,
	comp CompressorAnalysis,
	drive BlowerDriveAnalysis,
) COTSPart {
	switch p.BlowerDriveOption {
	case BlowerDriveRegenSideChannel:
		return COTSPart{
			SKU:         fmt.Sprintf("REGEN-RING-%.2gHP", drive.MotorHP),
			Name:        fmt.Sprintf("%.2g HP Cast-Aluminum Side-Channel Regenerative Ring Blower + Bleed Blast-Gate", drive.MotorHP),
			NominalIn:   round1(math.Max(2.0, design.Envelope.DuctDiameter.Inches()*0.5)),
			Material:    "Die-Cast Aluminum Impeller & Housing (High-Static 60+ inWC, Dry Clean-Air Rated)",
			Source:      "Goorui / Gast / Ametek Windjammer / Grainger",
			UnitCostUSD: drive.FanCostUSD,
			WeightKg:    round2(14.0 + 4.5*drive.MotorHP),
			Notes:       fmt.Sprintf("Delivers %.1f inWC static (60+ inWC max headroom) for $%.0f (%+.0f USD vs industrial radial+VFD). %s", drive.RequiredStaticInWC, drive.FanCostUSD, drive.CostDeltaUSD, drive.Caveats[0]),
		}
	case BlowerDriveDualBypassBlowers:
		return COTSPart{
			SKU:         fmt.Sprintf("AMETEK-3STG-2X-%dV", drive.Voltage),
			Name:        fmt.Sprintf("2× Ametek Lamb 3-Stage Tangential Bypass Blowers in Parallel (%dV) + Staging Relay", drive.Voltage),
			NominalIn:   5.7,
			Material:    "Aluminum/Coated 3-Stage Tangential Bypass Fans (Isolated Motor Windings) + Parallel Manifold",
			Source:      "Ametek Lamb / Grainger / Central Vacuum Distributors",
			UnitCostUSD: drive.FanCostUSD,
			WeightKg:    7.2,
			Notes:       fmt.Sprintf("~240–260 CFM @ 25+ inWC ($%.0f total, %+.0f USD vs industrial radial+VFD). Tangential discharge isolates armature/windings from wet exhaust; swap $15 carbon brushes every ~45 firings.", drive.FanCostUSD, drive.CostDeltaUSD),
		}
	case BlowerDriveCompressorAirAtomize:
		_, weightKg := mediumPressureAirAtomizeBlowerCost(p, drive.MotorHP)
		voltageStr, vfdSKU := vfdSpecForHP(p, drive.MotorHP)
		return COTSPart{
			SKU:         fmt.Sprintf("MP-CAST-%.2gHP-%s-TF25", drive.MotorHP, vfdSKU),
			Name:        fmt.Sprintf("%.2g HP Medium-Pressure Cast-Aluminum Radial Blower + %s + $75 Twin-Fluid Air-Atomizing Nozzle", drive.MotorHP, voltageStr),
			NominalIn:   round1(math.Max(8.0, design.Blower.WheelDiameter.Inches())),
			Material:    "Cast-Aluminum Radial Scroll + IP20 GS21 VFD in NEMA 4X + 316L Twin-Fluid 25 um Air-Atomizing Nozzle & 40 PSI Regulator",
			Source:      "Grainger (Dayton/Cincinnati Fan) + Spraying Systems Co. + AutomationDirect",
			UnitCostUSD: drive.FanCostUSD,
			WeightKg:    weightKg,
			Notes:       fmt.Sprintf("Uses %.1f SCFM @ 40 PSI from studio shop compressor to produce 25 um SMD droplets at 48 m/s throat velocity (active twin-fluid Calvert d50=%.2f um parity), slashing static pressure to %.1f inWC and saving $%.0f vs industrial radial+VFD!", comp.AtomizeSCFMNeeded, activeTwinFluidD50Microns(design), drive.RequiredStaticInWC, math.Abs(drive.CostDeltaUSD)),
		}
	case BlowerDriveCompressorEductor:
		return COTSPart{
			SKU:         "AIR-AMP-316L-EDUCT",
			Name:        "316L Stainless Compressed-Air Amplifier / Venturi Eductor + 1/2\" NPT Regulator (Uses Existing Shop Compressor)",
			NominalIn:   round1(math.Max(2.0, design.Envelope.DuctDiameter.Inches()*0.5)),
			Material:    "316L Stainless Steel Air Amplifier (No Moving Parts)",
			Source:      "EXAIR / Vortec / McMaster-Carr",
			UnitCostUSD: drive.FanCostUSD,
			WeightKg:    3.4,
			Notes:       fmt.Sprintf("Feasible=%v: requires %.1f SCFM (~%.1f HP) vs %.1f SCFM shop compressor. %s", drive.Feasible, comp.EductorSCFMNeeded, comp.EductorHPNeeded, comp.AvailableSCFM, comp.EductorVerdict),
		}
	default:
		if eductor.Enabled {
			return selectDryMotiveBlowerVFD(p, eductor)
		}
		return selectWetBlowerVFD(p, design)
	}
}

// BuildCOTSSummary snaps (when requested) the engineered scrubber geometry to standard
// commercial off-the-shelf hardware, recalculates actual % of flooding velocity, reflects
// all Phase 7 & Phase 8 selectable options, and builds a complete COTS Bill of Materials.
func BuildCOTSSummary(project StudioProject, design model.ScrubberDesign) COTSSummary {
	p := project.Normalize()
	if p.IsLocalCeramic() {
		return buildLocalCeramicCOTSSummary(p, design, EvaluateFrugal(p, design))
	}
	baseDesign := resolveBaselineDesign(p, design)
	eductor := analyzeEductor(p, design)
	parallelBed := analyzeParallelBed(p, design, eductor)
	parallelVenturi := analyzeParallelVenturi(p, design)
	compressor, atomizeDesign := analyzeCompressor(p, design, baseDesign)
	blowerDrive := analyzeBlowerDrive(p, design, baseDesign, eductor, compressor, atomizeDesign)
	packMedia, packSpec, packCuFt := analyzePackingMedia(p, design, eductor, parallelBed)

	quenchPart := selectQuenchVessel(p, design)
	sumpPart := selectSumpTankForProject(p, design)
	nozzlePart := selectSprayNozzle(p, design)
	sensorPart := selectSensorKitForProject(p, design)

	var colPart COTSPart
	var actualColIDIn, actualPctFlood float64

	if parallelBed.Enabled {
		geom := resolveParallelPipeGeom(p, eductor)
		baseSmallPackCost := packCuFt * geom.packUnitCost
		manifoldHardwareCost := round2(math.Max(150.0, parallelBed.ManifoldCostUSD-baseSmallPackCost))
		pipePrefix := "PAR-CPVC80"
		if geom.isPVC {
			pipePrefix = "PAR-PVC40"
		}
		colPart = COTSPart{
			SKU:         fmt.Sprintf("%s-%02dX%.0f-UNISEAL", pipePrefix, parallelBed.PipeCount, parallelBed.PipeNominalIn),
			Name:        fmt.Sprintf(`Shell-Less Uniseal Parallel Bundle (%d × %.0f" %s + Flat Collector Plates + %d Uniseals + Spray Header)`, parallelBed.PipeCount, parallelBed.PipeNominalIn, parallelBed.PipeMaterial, 2*parallelBed.PipeCount),
			NominalIn:   parallelBed.PipeNominalIn,
			Material:    parallelBed.PipeMaterial + " + 1/2\" HDPE/CPVC Flat Collector Plates + Rubber Uniseals (NO External Large Pipe)",
			Source:      "Local Plumbing Supply / Home Depot / US Plastic / Uniseal",
			UnitCostUSD: manifoldHardwareCost,
			WeightKg:    round2(float64(parallelBed.PipeCount)*3.8 + 6.5),
			Notes:       fmt.Sprintf("Shell-less bundle of %d bare %.0f\" risers (ID %.3f\", %.1f%% flood, D/d_p=%.2f) held between top/bottom flat collector plates with %d Uniseal rubber pipe seals ($5.50 ea) and balanced %d-port spray header — eliminates specialty large pipe & industrial flanges.", parallelBed.PipeCount, parallelBed.PipeNominalIn, parallelBed.PipeActualIDIn, parallelBed.ActualFloodPct, parallelBed.WallRatioDToDp, 2*parallelBed.PipeCount, parallelBed.PipeCount),
		}
		actualColIDIn = round2(math.Sqrt(float64(parallelBed.PipeCount)) * parallelBed.PipeActualIDIn)
		actualPctFlood = parallelBed.ActualFloodPct
	} else {
		colPart, actualColIDIn, actualPctFlood = selectSinglePackedColumn(p, design, eductor)
	}

	packingPart := COTSPart{
		SKU:         packSpec.sku,
		Name:        fmt.Sprintf("%s (1 cu.ft.)", packSpec.label),
		NominalIn:   packSpec.diameterIn,
		Material:    packSpec.material,
		Source:      packSpec.source,
		UnitCostUSD: packSpec.unitCostUSD,
		WeightKg:    packSpec.weightKgCuFt,
		Notes:       packMedia.Summary,
	}

	blowerPart := selectBlowerForDrive(p, design, eductor, compressor, blowerDrive)

	var silencerPtr *COTSPart
	if p.IncludeSilencer || p.IsDeluxe() {
		s := selectSilencerForProject(p, design)
		silencerPtr = &s
	}

	var diluterPtr *COTSPart
	if p.IncludePlumeDiluter {
		d := selectPlumeDiluter(p, design)
		diluterPtr = &d
	}

	nozzleQty := float64(design.Quench.NozzleCount + 2)
	if nozzleQty < 3 {
		nozzleQty = 4
	}

	// Stage 2 Venturi / Eductor hardware selection.
	ventMat := "316L Stainless Steel"
	ventCost := estimateSingleVenturiCost(p, design)
	ventWeight := round2(7.5 + 0.9*design.Venturi.InletDiameter.Inches())
	ventName := "Bolted Concentric Reducer Venturi & Throat Injection Spool"
	ventSpec := fmt.Sprintf(`No-weld bolted reducers: %.1f" -> %.2f" throat -> %.1f" with 2x NPT throat injection ports`,
		design.Venturi.InletDiameter.Inches(), design.Venturi.ThroatDiameter.Inches(), design.Venturi.OutletDiameter.Inches())
	if p.BuildPath == BuildPathWeldedDXF {
		ventName = "Rolled & Welded Venturi Cones & Throat Spool (from scrubber.dxf)"
		ventSpec = fmt.Sprintf(`DXF flat-pattern cones: %.0fmm -> %.1fmm throat -> %.0fmm + bolted flanges`,
			design.Venturi.InletDiameter.Millimeters(), design.Venturi.ThroatDiameter.Millimeters(), design.Venturi.OutletDiameter.Millimeters())
	} else if p.IsDeluxe() {
		ventName = "TIG-Welded Heavy-Gauge Venturi Cones & Throat w/ ANSI B16.5 Class 150 Lap-Joint Flanges"
		ventSpec = fmt.Sprintf(`Precision-rolled & back-purged TIG venturi: %.0fmm -> %.1fmm throat -> %.0fmm w/ machined ANSI/ASME B16.5 Class 150 lap-joint flanges & PTFE envelope gaskets`,
			design.Venturi.InletDiameter.Millimeters(), design.Venturi.ThroatDiameter.Millimeters(), design.Venturi.OutletDiameter.Millimeters())
		ventWeight = round2(ventWeight * 1.35)
	}
	if p.CorrosionTier == CorrosionTierHeavyC276FRP {
		ventMat = "Alloy C-276 Throat / Silicon Carbide Liner + FRP Diffuser"
		ventWeight = round2(ventWeight * 1.35)
	}

	if eductor.Enabled {
		ventName = fmt.Sprintf(`Vertical Downflow Static Venturi Eductor Tee & %.0f" 90° Cyclonic Sweep Elbow`, eductor.CyclonicSweepElbowIn)
		ventCost = round2(285.0 + 24.0*eductor.CyclonicSweepElbowIn)
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			ventCost = round2(ventCost * 2.2)
		}
		ventWeight = round2(11.0 + 1.1*eductor.CyclonicSweepElbowIn)
		ventSpec = fmt.Sprintf(`Quench exits out & down to %.0f" flanged Tee -> vertical downflow static venturi driven by %.0f CFM dry motive air -> 90° cyclonic sweep elbow with bottom drain -> Stage 3 bottom plenum (+%.2f inWC positive pressure, outdoor-only airtight bolted PTFE joints).`,
			eductor.CyclonicSweepElbowIn, eductor.MotiveAirCFM, eductor.PositivePressureInWC)
	} else if parallelVenturi.Enabled {
		ventName = fmt.Sprintf(`%d× Parallel COTS Sanitary Mini-Venturis (%.0f" -> %.2f" -> %.0f" Tri-Clamp)`,
			parallelVenturi.ThroatCount, parallelVenturi.InletFittingIn, parallelVenturi.ThroatDiameterIn, parallelVenturi.InletFittingIn)
		ventCost = parallelVenturi.MultiCostUSD
		ventWeight = round2(float64(parallelVenturi.ThroatCount)*2.8 + 4.0)
		ventSpec = fmt.Sprintf(`Zero cone rolling: %d parallel sanitary concentric-reducer throats (%.1f inWC dP, plugging risk: %s). %s`,
			parallelVenturi.ThroatCount, parallelVenturi.ThroatDPInWC, parallelVenturi.PluggingRisk, parallelVenturi.TurndownBenefit)
	}

	// Chemical-resistant mag-drive recirculation pump.
	recircGPM := design.PackedBed.RecirculationFlowGPM
	if recircGPM <= 0 {
		recircGPM = math.Max(4.0, design.Quench.TotalSprayWaterGPM*3.0)
	}
	pumpCost := round2(265.0 + 14.0*recircGPM)
	pumpWeight := round2(7.5 + 0.30*recircGPM)
	pumpName := "Seal-Less Mag-Drive Recirculation Pump & Ash Strainer"
	pumpMat := "Polypropylene / EPDM / Ceramic Shaft"
	pumpSpec := fmt.Sprintf("Rated %.1f GPM @ 35 ft TDH (March / Iwaki / Little Giant TE-MD) + 50-mesh inline Y-strainer", recircGPM)
	if p.IsDeluxe() {
		pumpCost = round2(pumpCost + 195.0)
		pumpWeight = round2(pumpWeight + 4.5)
		pumpName = "Heavy-Duty Magnetic-Drive Sealless Chemical Recirculation Pump Station (PVDF/Ceramic)"
		pumpMat = "PVDF / Ceramic Shaft / PTFE O-Rings + CPVC/316L Isolation Valves & Y-Strainer"
		pumpSpec = fmt.Sprintf("[DLX-MAGDRIVE-PUMP] Rated %.1f GPM @ 45 ft TDH (Iwaki MD / March PVDF) + flooded-suction 316L/CPVC Y-strainer, true-union isolation ball valves, check valve & glycerin pressure gauge", recircGPM)
	}

	// 3-pass polypropylene chevron mist eliminator pad sized to actual column ID.
	colAreaSqFt := math.Pi * (actualColIDIn / 12.0) * (actualColIDIn / 12.0) / 4.0
	demisterCost := round2(145.0 + 95.0*colAreaSqFt)
	demisterWeight := round2(math.Max(2.0, 2.8*colAreaSqFt))
	demisterName := "3-Pass Chevron Mist Eliminator Pad & Packing Support Grate"
	demisterSpec := fmt.Sprintf(`Cut to %.1f" equivalent column bore; captures >99%% entrained mist droplets`, actualColIDIn)
	if p.IsDeluxe() {
		demisterCost = round2(demisterCost + 125.0)
		demisterWeight = round2(demisterWeight + 3.2)
		demisterName = "Removable 3-Pass Chevron Demister Cassette & Reinforced FRP Support Grate"
		demisterSpec = fmt.Sprintf(`[DLX-DEMISTER-CASS] Slide-out removable cassette frame sized to %.1f" bore w/ bolted access cover & reinforced FRP support bar grating`, actualColIDIn)
	}

	// Gaskets, clamps, CPVC cement, hardware kit.
	hardwareCost := 195.0
	hardwareName := "No-Weld Couplings, EPDM/PTFE Gaskets, CPVC Cement & 316 SS Clamps"
	hardwareWeight := 8.5
	hardwareSpec := ">=2300 F alumina-silica ceramic fiber rope/gaskets for dry hot flue upstream of Stage 1 (NO silicone/PTFE on 2400 F flue!), expanded PTFE/EPDM gaskets strictly downstream (<=200 F), band clamps/bolts, CPVC Sch 80 primer & cement, drain ball valve, thermometer (IPC 702.5 <=140 F check), brine hydrometer (SG 1.000-1.200), pH test strips"
	if p.BuildPath == BuildPathWeldedDXF {
		hardwareCost = 245.0
		hardwareName = "PTFE Full-Face Flange Gaskets & 316 SS Hex Bolt/Nut Kit"
	} else if p.IsDeluxe() {
		hardwareCost = 485.0
		hardwareWeight = 16.5
		hardwareName = "ANSI/ASME B16.5 Class 150 Lap-Joint Flanges, PTFE Envelope Gaskets, 316 SS Studs & Braided Chemical Hose Kit"
		hardwareSpec = "[DLX-ANSI150-HOSE] >=2300 F alumina-silica ceramic fiber hot-flue gaskets (NO silicone/PTFE on 2400 F flue!), machined ANSI/ASME B16.5 Class 150 lap-joint flange kit, PTFE envelope gaskets, 316 SS studs w/ anti-seize, braided EPDM/PTFE chemical hoses, IPC 702.5 <=140 F thermometer, brine hydrometer (SG 1.000-1.200) & pH test kit"
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
			Name:        demisterName,
			Material:    "Polypropylene (PP) / FRP Grating",
			Quantity:    1,
			Unit:        "assy",
			Spec:        demisterSpec,
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
			Name:        pumpName,
			Material:    pumpMat,
			Quantity:    1,
			Unit:        "ea",
			Spec:        pumpSpec,
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

	if !p.SharedBaffledSump {
		items = append(items, model.BOMItem{
			Category:    "Liquid & Chemical Handling",
			Name:        "Secondary Dedicated Quench-Ash Sump Drum & Second Mag-Drive Pump",
			Material:    "HDPE Drum + Polypropylene Mag-Drive Pump + CPVC Unions",
			Quantity:    1,
			Unit:        "assy",
			Spec:        "[DUAL-SUMP-KIT] Separate Stage 1-2 ash-settling sump & transfer pump (uncheck to save cost with shared baffled sump)",
			UnitCostUSD: dualSumpExtraCost(design),
			WeightKg:    18.0,
		})
	}

	if p.CompactFootprint {
		items = append(items, model.BOMItem{
			Category:    "Hardware & Plumbing",
			Name:        "Compact U-Turn 180° Swept Return Bend & Folded Skid Mounting Frame",
			Material:    "Schedule 80 CPVC / 316L SS Swept Elbow + Steel Frame",
			Quantity:    1,
			Unit:        "assy",
			Spec:        "[UTURN-SKID-180] Folds train length by ~47% for narrow urban side-yards",
			UnitCostUSD: 65.0,
			WeightKg:    6.5,
		})
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

	if p.IsDeluxe() {
		skidLenFt := math.Max(5.5, design.Envelope.Length.Feet())
		items = append(items,
			model.BOMItem{
				Category:    "Stage 1: Quench Vessel",
				Name:        `2× Borosilicate Sight-Glass Ports, 2× Bolted NFPA 91 Cleanout Doors & 1" Ceramic-Fiber/SS Flue Insulation Jacket`,
				Material:    "Borosilicate Glass / 316L SS / 2300 F Ceramic-Fiber Blanket w/ Stainless Cladding",
				Quantity:    1,
				Unit:        "kit",
				Spec:        "[DLX-SIGHT-CLNOUT-INSUL] 2× flanged borosilicate sight-glass ports (quench spray + packing bed), 2× bolted NFPA 91 quick-release cleanout doors, and 1\" ceramic-fiber + stainless steel hot-flue insulation jacket",
				UnitCostUSD: 585.0,
				WeightKg:    14.5,
			},
			model.BOMItem{
				Category:    "Structural Skid & Vibration Isolation",
				Name:        `Welded 2"×2"×3/16" Galvanized/Powder-Coated Steel Forklift Skid, Diagonal K-Bracing, Pipe Saddles, Anchors & Vibration Boots`,
				Material:    `ASTM A500 2"×2"×3/16" Galvanized/Powder-Coated Steel Tube / 316 SS Wedge Anchors / PTFE-EPDM Flex Boots`,
				Quantity:    1,
				Unit:        "assy",
				Spec:        fmt.Sprintf(`[DLX-SKID-KBRACE-VBOOT] Welded forklift-pocket structural steel skid (%.1f ft base) + Stage 3 column cage & diagonal K-bracing + overhead duct pipe saddles + 6× 1/2"×4" 316 SS concrete wedge anchors + flexible PTFE/EPDM vibration isolation boots at blower inlet & outlet`, skidLenFt),
				UnitCostUSD: round2(780.0 + 28.0*skidLenFt),
				WeightKg:    round2(88.0 + 6.5*skidLenFt),
			},
		)
	}

	discSpec := "[NEC430-DISC-LR] Padlockable line-side disconnect within sight (<=50 ft) of VFD & blower + 3% input line reactor for NEC 210.8(F) HF-GFCI stability (Square D / AutomationDirect)"
	discCost := 215.0
	if p.IsDeluxe() {
		discSpec = "[NEC430-DISC-LR] UL 508A through-door lockable NEC 430.102 disconnect handle + 3% copper-wound input line reactor + NEC 210.8(F) HF-GFCI breaker kit (Square D / AutomationDirect)"
		discCost = 310.0
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
			Spec:        discSpec,
			UnitCostUSD: discCost,
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
			Spec:        hardwareSpec,
			UnitCostUSD: hardwareCost,
			WeightKg:    hardwareWeight,
		},
	)

	title := fmt.Sprintf("Studio COTS Hardware Bill of Materials — %s (%.1f cu.ft.)",
		p.StudioName, design.Input.KilnVolumeCuFt)
	bom := model.BOM{
		Title: title,
		Items: items,
		Notes: []string{
			"All line-item WeightKg values represent TOTAL line-item weight (already multiplied by Quantity).",
			"COTS pricing reflects realistic US studio/industrial distributor budgetary estimates (excluding freight and tax).",
			fmt.Sprintf("Architecture: alignment=%s, packed_bed=%s (%.0f\" pipe), venturi=%s, blower_drive=%s, packing_media=%s, compact_u_turn=%v, shared_sump=%v, controls=%s.",
				p.StageAlignment, p.PackedBedLayout, p.ParallelPipeSizeIn, p.VenturiLayout, p.BlowerDriveOption, p.PackingMediaOption, p.CompactFootprint, p.SharedBaffledSump, p.ControlTier),
			fmt.Sprintf("Packed absorber snapped=%v -> %.1f\" equiv ID operating at %.1f%% of Eckert flooding velocity (design limit <= 75%%).", p.SnapToCOTS, actualColIDIn, actualPctFlood),
			"NFPA 86 & ASTM F441: Stage 3 Sch 80 CPVC/FRP components MUST be protected by an independent hardwired 200 F manual-reset snap-disc + 24VDC de-energize-to-trip latching safety relay driving the spring-return fail-open hot-gas bypass damper (ahead of the 210 F / 0.25 GPM software backup).",
			"NEC 430.102 / 430.122 / 210.8(F) / 110.28: Size VFD branch circuit at >=125% VFD rated 1-phase input current, install lockable NEMA 4X sight disconnect (<=50 ft), house IP20 VFD in NEMA 4X enclosure, and use line-side HF-rated GFCI + input line reactor.",
			"IPC 608 / 702.5 / 803.1 & NFPA 91: Isolate potable water via Air Gap (>=1\") or ASSE 1013 RPZ; cool drain/blowdown <=140 F and neutralize to pH 6.5-8.5 (>=5.0 per 40 CFR 403.5) before sanitary discharge; include NFPA 91 duct cleanouts/low-point drains and ASTM E84 Class 1 plastic ducting.",
		},
	}
	bom.RecalculateTotals()

	return COTSSummary{
		Snapped:              p.SnapToCOTS,
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
