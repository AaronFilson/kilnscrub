package cad

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"strings"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

const (
	density316LSS     = 8000.0 // kg/m^3
	densityCPVC       = 1550.0 // kg/m^3
	densityPP         = 920.0  // kg/m^3
	bulkDensityPallPP = 88.0   // kg/m^3 for 1" polypropylene Pall rings
)

func round2(v float64) float64 {
	return math.Round(v*100.0) / 100.0
}

func frustumLateralArea(d1, d2, axialLen float64) float64 {
	r1, r2 := 0.5*d1, 0.5*d2
	slant := math.Hypot(axialLen, r2-r1)
	return math.Pi * (r1 + r2) * slant
}

func flangeRingWeightKg(f flangeSpec, density float64) float64 {
	ringArea := 0.25 * math.Pi * (f.OD*f.OD - f.ID*f.ID)
	holeArea := float64(f.BoltCount) * 0.25 * math.Pi * f.BoltHoleD * f.BoltHoleD
	netArea := math.Max(0.0, ringArea-holeArea)
	return netArea * f.Thickness * density
}

// GenerateBOM generates a realistic fabrication Bill of Materials for the 4-stage scrubber.
// Every BOMItem.WeightKg represents the TOTAL weight of that line item (for all Quantity units),
// and bom.RecalculateTotals() is always called before returning.
func (e *Exporter) GenerateBOM(design model.ScrubberDesign) (model.BOM, error) {
	if err := validateDesign(design); err != nil {
		return model.BOM{}, err
	}
	g := deriveGeometry(design)

	quenchMat := design.Quench.Material
	if quenchMat == "" {
		quenchMat = "316L Stainless Steel"
	}
	venturiMat := design.Venturi.Material
	if venturiMat == "" {
		venturiMat = "316L Stainless Steel"
	}
	colMat := design.PackedBed.Material
	if colMat == "" {
		colMat = "CPVC / Polypropylene"
	}
	packingSpec := design.PackedBed.PackingSpec
	if packingSpec == "" {
		packingSpec = design.Input.PackingName
	}
	if packingSpec == "" {
		packingSpec = `1" PP Pall Rings`
	}
	blowerMat := design.Blower.Material
	if blowerMat == "" {
		blowerMat = "FRP / Chemical-Resistant Coated Steel"
	}

	// --- Stage 1: Quench Tower + Inlet Flue Duct (316L SS) ---
	quenchShellArea := math.Pi * g.quenchDia * g.quenchH
	quenchCapsArea := 2.0 * 0.25 * math.Pi * g.quenchDia * g.quenchDia
	ductArea := math.Pi * g.ductDia * g.ductTotalLen
	quenchSheetM2 := round2((quenchShellArea + quenchCapsArea + ductArea) * 1.15) // 15% scrap/nesting allowance
	quenchSheetWeightKg := round2(quenchSheetM2 * g.quenchWall * density316LSS)

	quenchFlangeQty := 4.0 // 2 tower end flanges + 2 inlet duct flanges
	quenchFlangeWeightKg := round2(2.0*flangeRingWeightKg(g.flangeQuench, density316LSS) +
		2.0*flangeRingWeightKg(g.flangeDuct, density316LSS))

	// --- Stage 2: Wet Venturi Scrubber (316L SS) ---
	ventConvArea := frustumLateralArea(g.ventInDia, g.ventThDia, g.ventConvLen)
	ventThArea := math.Pi * g.ventThDia * g.ventThLen
	ventDivArea := frustumLateralArea(g.ventThDia, g.ventOutDia, g.ventDivLen)
	ventSheetM2 := round2((ventConvArea + ventThArea + ventDivArea) * 1.15)
	ventSheetWeightKg := round2(ventSheetM2 * g.ventWall * density316LSS)

	ventFlangeQty := 4.0 // 2 inlet + 2 outlet flanges
	ventFlangeWeightKg := round2(2.0*flangeRingWeightKg(g.flangeVentIn, density316LSS) +
		2.0*flangeRingWeightKg(g.flangeVentOut, density316LSS))

	// --- Stage 3: Packed-Bed Absorber Column + Sump + Packing + Demister (CPVC/PP) ---
	// ColumnHeight already includes sumpH, so subtract sumpH from the upper shell to avoid double-counting.
	upperColH := math.Max(0.25, g.colH-g.sumpH)
	colShellArea := math.Pi * g.colDia * upperColH
	reducerArea := frustumLateralArea(g.colDia, g.stackDia, g.reducerH)
	stackArea := math.Pi * g.stackDia * g.stackRiserH
	colSheetM2 := round2((colShellArea + reducerArea + stackArea) * 1.12)
	colSheetWeightKg := round2(colSheetM2 * g.colWall * densityCPVC)

	sumpArea := math.Pi*g.colDia*g.sumpH + 0.25*math.Pi*g.colDia*g.colDia
	sumpWeightKg := round2(sumpArea * g.colWall * densityCPVC * 1.25) // includes baffles & support plate
	sumpVolGal := g.sumpVol / model.CubicMetersPerUSGallon

	colFlangeQty := 4.0 // sump, mid-service, top column, stack
	colFlangeWeightKg := round2(3.0*flangeRingWeightKg(g.flangeColumn, densityCPVC) +
		1.0*flangeRingWeightKg(g.flangeStack, densityCPVC))

	packingVolM3 := math.Round(g.packVol*1000.0) / 1000.0
	if packingVolM3 < 0.001 {
		packingVolM3 = math.Max(0.001, g.packVol)
	}
	packingWeightKg := round2(g.packVol * bulkDensityPallPP)

	colAreaM2 := 0.25 * math.Pi * g.colDia * g.colDia
	demisterWeightKg := round2(math.Max(2.5, colAreaM2*g.demisterH*0.13*densityPP))

	// --- Pumps, Spray Nozzles, Dosing ---
	sprayNozzleQty := float64(g.quenchNozz + 2) // quench nozzles + 2 venturi throat injection nozzles
	sprayNozzleWeightKg := round2(sprayNozzleQty * 0.45)

	recircGPM := design.PackedBed.RecirculationFlowGPM
	if recircGPM <= 0 {
		recircGPM = math.Max(5.0, design.Quench.TotalSprayWaterGPM*3.0)
	}
	recircPumpCost := round2(420.0 + 18.0*recircGPM)
	recircPumpWeightKg := round2(12.0 + 0.35*recircGPM)

	// --- Stage 4: ID Blower & VFD ---
	motorHP := design.Blower.RecommendedMotorHP
	if motorHP <= 0 {
		motorHP = 1.0
	}
	blowerCost := round2(1150.0 + 380.0*motorHP + 650.0*g.wheelDia)
	blowerWeightKg := round2(28.0 + 14.0*motorHP + 35.0*g.wheelDia)

	vfdCost := round2(340.0 + 145.0*motorHP)
	vfdWeightKg := round2(4.5 + 1.2*motorHP)

	// --- Safety Bypass Damper & Instrumentation ---
	damperCost := round2(480.0 + 420.0*g.ductDia)
	damperWeightKg := round2(8.0 + 18.0*g.ductDia)

	// --- Gaskets & 316 SS Fasteners ---
	// Count one bolt set and one full-face gasket per mated flange joint (not per individual face).
	totalFlangePairs := 7.0 // duct, quench top/bot, venturi in/out, column mid, stack
	gasketWeightKg := round2(totalFlangePairs * 0.35)

	totalBolts := g.flangeDuct.BoltCount + g.flangeQuench.BoltCount +
		g.flangeVentIn.BoltCount + g.flangeVentOut.BoltCount +
		2*g.flangeColumn.BoltCount + g.flangeStack.BoltCount
	fastenerWeightKg := round2(float64(totalBolts) * 0.085)

	items := []model.BOMItem{
		{
			Category:    "Stage 1: Quench Tower",
			Name:        "Quench Tower & Inlet Flue Duct Sheet",
			Material:    quenchMat,
			Quantity:    quenchSheetM2,
			Unit:        "m2",
			Spec:        fmt.Sprintf("D=%.0fmm x H=%.0fmm tower + D=%.0fmm x L=%.2fm duct, t=%.2fmm (incl. 15%% allowance)", g.quenchDia*1000, g.quenchH*1000, g.ductDia*1000, g.ductTotalLen, g.quenchWall*1000),
			UnitCostUSD: 145.0,
			WeightKg:    quenchSheetWeightKg,
		},
		{
			Category:    "Stage 1: Quench Tower",
			Name:        "Quench Tower & Inlet Duct Flanges",
			Material:    quenchMat,
			Quantity:    quenchFlangeQty,
			Unit:        "ea",
			Spec:        fmt.Sprintf("2x D%.0fmm tower flanges + 2x D%.0fmm duct flanges, t=%.1fmm", g.quenchDia*1000, g.ductDia*1000, g.flangeQuench.Thickness*1000),
			UnitCostUSD: 95.0,
			WeightKg:    quenchFlangeWeightKg,
		},
		{
			Category:    "Stage 2: Wet Venturi",
			Name:        "Venturi Converging/Throat/Diverging Sheet",
			Material:    venturiMat,
			Quantity:    ventSheetM2,
			Unit:        "m2",
			Spec:        fmt.Sprintf("D_in=%.0fmm -> D_th=%.0fmm -> D_out=%.0fmm, t=%.2fmm (incl. 15%% allowance)", g.ventInDia*1000, g.ventThDia*1000, g.ventOutDia*1000, g.ventWall*1000),
			UnitCostUSD: 175.0,
			WeightKg:    ventSheetWeightKg,
		},
		{
			Category:    "Stage 2: Wet Venturi",
			Name:        "Venturi Bolted Flanges & Injection Collar",
			Material:    venturiMat,
			Quantity:    ventFlangeQty,
			Unit:        "ea",
			Spec:        fmt.Sprintf("2x D%.0fmm inlet + 2x D%.0fmm outlet bolted flanges + throat collar", g.ventInDia*1000, g.ventOutDia*1000),
			UnitCostUSD: 85.0,
			WeightKg:    ventFlangeWeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Absorber",
			Name:        "Absorber Column Shell, Reducer & Stack",
			Material:    colMat,
			Quantity:    colSheetM2,
			Unit:        "m2",
			Spec:        fmt.Sprintf("D=%.0fmm x H=%.0fmm column + D=%.0fmm stack, t=%.2fmm (incl. 12%% allowance)", g.colDia*1000, g.colH*1000, g.stackDia*1000, g.colWall*1000),
			UnitCostUSD: 110.0,
			WeightKg:    colSheetWeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Absorber",
			Name:        "Chemical Sump Tank & Packing Support Grids",
			Material:    colMat,
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("D=%.0fmm x H=%.0fmm sump (%.1f US gal) + FRP/PP support & hold-down grids", g.colDia*1000, g.sumpH*1000, sumpVolGal),
			UnitCostUSD: round2(360.0 + 4.5*sumpVolGal),
			WeightKg:    sumpWeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Absorber",
			Name:        "Absorber Column Flanges",
			Material:    colMat,
			Quantity:    colFlangeQty,
			Unit:        "ea",
			Spec:        fmt.Sprintf("3x D%.0fmm column flanges + 1x D%.0fmm stack flange, t=%.1fmm", g.colDia*1000, g.stackDia*1000, g.flangeColumn.Thickness*1000),
			UnitCostUSD: 72.0,
			WeightKg:    colFlangeWeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Absorber",
			Name:        "Random Tower Packing",
			Material:    "Polypropylene (PP)",
			Quantity:    packingVolM3,
			Unit:        "m3",
			Spec:        fmt.Sprintf("%s, bed height %.2fm, volume %.3f m3 (%.2f cu.ft.)", packingSpec, g.packH, g.packVol, model.CubicMeters(g.packVol).CuFt()),
			UnitCostUSD: 950.0,
			WeightKg:    packingWeightKg,
		},
		{
			Category:    "Stage 3: Packed-Bed Absorber",
			Name:        "Chevron-Blade Mist Eliminator (Demister)",
			Material:    "Polypropylene (PP)",
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("3-pass chevron profile, D=%.0fmm x H=%.0fmm", g.colDia*1000, g.demisterH*1000),
			UnitCostUSD: round2(260.0 + 480.0*colAreaM2),
			WeightKg:    demisterWeightKg,
		},
		{
			Category:    "Liquid & Chemical Handling",
			Name:        "Atomizing Quench & Venturi Spray Nozzles",
			Material:    "316L SS / PVDF",
			Quantity:    sprayNozzleQty,
			Unit:        "ea",
			Spec:        fmt.Sprintf("%d quench full-cone (%.2f GPM total) + 2 venturi throat nozzles", g.quenchNozz, design.Quench.TotalSprayWaterGPM),
			UnitCostUSD: 68.0,
			WeightKg:    sprayNozzleWeightKg,
		},
		{
			Category:    "Liquid & Chemical Handling",
			Name:        "Chemical-Resistant Recirculation Pump",
			Material:    "CPVC / PVDF Mag-Drive",
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("Rated %.1f GPM @ 40 ft TDH, seal-less acid/alkali service", recircGPM),
			UnitCostUSD: recircPumpCost,
			WeightKg:    recircPumpWeightKg,
		},
		{
			Category:    "Liquid & Chemical Handling",
			Name:        "NaHCO3 Dosing System & Inline pH Controller",
			Material:    "HDPE / PVDF / Glass Electrode",
			Quantity:    1,
			Unit:        "assy",
			Spec:        fmt.Sprintf("Peristaltic metering pump, 15-gal day tank, pH probe (target %.1f, %.2f kg/hr NaHCO3)", design.PackedBed.SumpPH, design.PackedBed.NaHCO3ConsumptionKgPerHr),
			UnitCostUSD: 640.0,
			WeightKg:    14.5,
		},
		{
			Category:    "Stage 4: Induced-Draft Blower",
			Name:        "ID Centrifugal Blower & TEFC Motor",
			Material:    blowerMat,
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("%.1f HP motor, wheel D=%.0fmm, %.0f ACFM @ %.2f inWC static", motorHP, g.wheelDia*1000, design.Blower.FlowACFM, design.Blower.TotalStaticPressureInWC),
			UnitCostUSD: blowerCost,
			WeightKg:    blowerWeightKg,
		},
		{
			Category:    "Stage 4: Induced-Draft Blower",
			Name:        "Variable Frequency Drive (VFD)",
			Material:    "NEMA 4X Enclosure",
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("Rated %.1f HP, 0-10V / 4-20mA PID draft speed reference", motorHP),
			UnitCostUSD: vfdCost,
			WeightKg:    vfdWeightKg,
		},
		{
			Category:    "Safety & Instrumentation",
			Name:        "Fail-Safe Spring-Return Hot-Gas Bypass Damper",
			Material:    "316L SS / 24VDC Spring-Return Actuator",
			Quantity:    1,
			Unit:        "ea",
			Spec:        fmt.Sprintf("D=%.0fmm butterfly damper, fails OPEN on 200 F hardwired snap-disc & 200 F software limit / <0.25 GPM / power loss", g.ductDia*1000),
			UnitCostUSD: damperCost,
			WeightKg:    damperWeightKg,
		},
		{
			Category:    "Safety & Instrumentation",
			Name:        "Process Instrumentation & Interlock Panel",
			Material:    "316 SS / NEMA 4X",
			Quantity:    1,
			Unit:        "lot",
			Spec:        "NFPA 86 independent 200 F manual-reset snap-disc + 24VDC de-energize-to-trip latching safety relay, Type-K inlet TC, Pt100 quench RTD, +/-2 inWC draft transmitter, 0.25 GPM flow switch",
			UnitCostUSD: 890.0,
			WeightKg:    9.5,
		},
		{
			Category:    "Hardware & Seals",
			Name:        "Chemical-Resistant Full-Face Flange Gaskets",
			Material:    "Expanded PTFE / Viton",
			Quantity:    totalFlangePairs,
			Unit:        "ea",
			Spec:        "3.2mm thick acid/high-temp service gaskets matching all stage flanges",
			UnitCostUSD: 28.0,
			WeightKg:    gasketWeightKg,
		},
		{
			Category:    "Hardware & Seals",
			Name:        "Flange Fastener Kit (Hex Bolts, Nuts & Washers)",
			Material:    "316 Stainless Steel",
			Quantity:    float64(totalBolts),
			Unit:        "ea",
			Spec:        fmt.Sprintf("%d M10/M12 bolt-nut-washer sets for all bolted flanges", totalBolts),
			UnitCostUSD: 2.25,
			WeightKg:    fastenerWeightKg,
		},
	}

	title := fmt.Sprintf("4-Stage Kiln Exhaust Scrubber BOM (%.1f cu.ft. Kiln, %.0f F Peak)",
		design.Input.KilnVolumeCuFt, design.Input.PeakExitTempF)
	if design.Input.KilnVolumeCuFt <= 0 {
		title = "4-Stage Kiln Exhaust Scrubber Bill of Materials"
	}

	bom := model.BOM{
		Title: title,
		Items: items,
		Notes: []string{
			"All line-item WeightKg values represent TOTAL line-item weight (accounting for Quantity).",
			"Sheet-metal quantities include fabrication nesting/scrap allowance (15% for 316L SS, 12% for CPVC/PP).",
			"Costs are budgetary USD estimates for studio-scale fabrication and exclude labor, shipping, and taxes.",
			"NFPA 86 & ASTM F441: Stage 3 CPVC/PP components MUST be protected by an independent hardwired 200 F manual-reset snap-disc + 24VDC latching safety relay driving the spring-return fail-open bypass damper (ahead of the 210 F / 98.9 C software backup).",
			"US Code Compliance: Include NEC 430.102 lockable NEMA 4X sight disconnect (<=50 ft) & NEC 430.122 125% VFD ampacity, IPC 608 Air Gap/ASSE 1013 RPZ backflow preventer, IPC 702.5 <=140 F drain cooldown, NFPA 211 >=2300 F ceramic flue gaskets & 36\" clearances, and NFPA 91 duct cleanouts/low-point drains & ASTM E84 Class 1 plastic ductwork.",
		},
	}
	bom.RecalculateTotals()
	return bom, nil
}

// FormatBOMCSV serializes a Bill of Materials to RFC 4180 CSV bytes.
func (e *Exporter) FormatBOMCSV(bom model.BOM) ([]byte, error) {
	// Work on a shallow copy with cloned slice so RecalculateTotals does not mutate caller state unexpectedly.
	local := bom
	local.Items = append([]model.BOMItem(nil), bom.Items...)
	local.RecalculateTotals()

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"Category", "Name", "Material", "Quantity", "Unit",
		"Spec", "UnitCostUSD", "TotalCostUSD", "WeightKg",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("cad: write BOM CSV header: %w", err)
	}

	for _, it := range local.Items {
		row := []string{
			it.Category,
			it.Name,
			it.Material,
			fmt.Sprintf("%.3f", it.Quantity),
			it.Unit,
			it.Spec,
			fmt.Sprintf("%.2f", it.UnitCostUSD),
			fmt.Sprintf("%.2f", it.TotalCostUSD),
			fmt.Sprintf("%.2f", it.WeightKg),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("cad: write BOM CSV row: %w", err)
		}
	}

	totalRow := []string{
		"TOTAL",
		fmt.Sprintf("%d line items", len(local.Items)),
		"",
		"",
		"",
		"",
		"",
		fmt.Sprintf("%.2f", local.TotalCostUSD),
		fmt.Sprintf("%.2f", local.TotalWeightKg),
	}
	if err := w.Write(totalRow); err != nil {
		return nil, fmt.Errorf("cad: write BOM CSV total row: %w", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("cad: flush BOM CSV: %w", err)
	}
	return buf.Bytes(), nil
}

func escapeMD(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// FormatBOMMarkdown renders a Bill of Materials as a GitHub-flavored Markdown document.
func (e *Exporter) FormatBOMMarkdown(bom model.BOM) ([]byte, error) {
	local := bom
	local.Items = append([]model.BOMItem(nil), bom.Items...)
	local.RecalculateTotals()

	title := local.Title
	if title == "" {
		title = "Kiln Exhaust Scrubber Bill of Materials"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s\n\n", escapeMD(title))
	fmt.Fprintf(&buf, "- **Line Items:** %d\n", len(local.Items))
	fmt.Fprintf(&buf, "- **Estimated Material Cost:** $%.2f USD\n", local.TotalCostUSD)
	fmt.Fprintf(&buf, "- **Estimated Total Weight:** %.2f kg (%.1f lbs)\n\n",
		local.TotalWeightKg, local.TotalWeightKg/model.KgPerLb)

	buf.WriteString("| Category | Item | Material | Qty | Unit | Specification | Unit Cost (USD) | Total Cost (USD) | Weight (kg) |\n")
	buf.WriteString("|---|---|---|---:|---|---|---:|---:|---:|\n")

	for _, it := range local.Items {
		fmt.Fprintf(&buf, "| %s | %s | %s | %.3g | %s | %s | $%.2f | $%.2f | %.2f |\n",
			escapeMD(it.Category),
			escapeMD(it.Name),
			escapeMD(it.Material),
			it.Quantity,
			escapeMD(it.Unit),
			escapeMD(it.Spec),
			it.UnitCostUSD,
			it.TotalCostUSD,
			it.WeightKg,
		)
	}

	fmt.Fprintf(&buf, "| **TOTAL** | **%d items** | | | | | | **$%.2f** | **%.2f** |\n",
		len(local.Items), local.TotalCostUSD, local.TotalWeightKg)

	if len(local.Notes) > 0 {
		buf.WriteString("\n## Fabrication Notes\n\n")
		for _, note := range local.Notes {
			fmt.Fprintf(&buf, "- %s\n", note)
		}
	}

	return buf.Bytes(), nil
}
