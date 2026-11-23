package studio

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// PlanPart represents a numbered assembly part in the interactive Shop Plan Viewer.
type PlanPart struct {
	Num          int     `json:"num"`
	ID           string  `json:"id"`
	Label        string  `json:"label"`
	Stage        string  `json:"stage"`
	Material     string  `json:"material"`
	DimensionsIn string  `json:"dimensions_in"`
	DimensionsMM string  `json:"dimensions_mm"`
	SKU          string  `json:"sku"`
	Qty          float64 `json:"qty"`
	CostUSD      float64 `json:"cost_usd"`
	Notes        string  `json:"notes"`
}

// CutSheetItem describes a single pipe cut length or exact unfolded sheet-metal
// flat pattern (cylinder rectangle or frustum annular sector from pkg/cad math).
type CutSheetItem struct {
	PartNum       int     `json:"part_num"`
	Name          string  `json:"name"`
	Shape         string  `json:"shape"`
	Material      string  `json:"material"`
	FlatWidthIn   float64 `json:"flat_width_in"`
	FlatHeightIn  float64 `json:"flat_height_in"`
	OuterRadiusIn float64 `json:"outer_radius_in"`
	InnerRadiusIn float64 `json:"inner_radius_in"`
	SweepDeg      float64 `json:"sweep_deg"`
	Qty           int     `json:"qty"`
	Notes         string  `json:"notes"`
}

// PlanClearances specifies mandatory code setbacks and shop maintenance aisles.
type PlanClearances struct {
	FlueToCombustiblesIn float64 `json:"flue_to_combustibles_in"`
	PropertyLineMinFt    float64 `json:"property_line_min_ft"`
	StackAboveRoofFt     float64 `json:"stack_above_roof_ft"`
	ServiceAisleIn       float64 `json:"service_aisle_in"`
}

// ShopPlan provides complete, self-contained in-browser assembly plans, numbered
// parts legend, flat-pattern cut sheets, and clearances without requiring CAD software.
type ShopPlan struct {
	Title            string         `json:"title"`
	StageAlignment   string         `json:"stage_alignment"`
	PackedBedLayout  string         `json:"packed_bed_layout"`
	VenturiLayout    string         `json:"venturi_layout"`
	CompactFootprint bool           `json:"compact_footprint"`
	OverallLengthIn  float64        `json:"overall_length_in"`
	OverallWidthIn   float64        `json:"overall_width_in"`
	OverallHeightIn  float64        `json:"overall_height_in"`
	Parts            []PlanPart     `json:"parts"`
	CutList          []CutSheetItem `json:"cut_list"`
	Clearances       PlanClearances `json:"clearances"`
}

func frustumBoundingBoxIn(rIn, rOut, thetaRad float64) (w, h float64) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	angles := []float64{0, thetaRad}
	for _, card := range []float64{0.5 * math.Pi, math.Pi, 1.5 * math.Pi} {
		if card > 0 && card < thetaRad {
			angles = append(angles, card)
		}
	}
	for _, ang := range angles {
		ca, sa := math.Cos(ang), math.Sin(ang)
		for _, r := range []float64{rIn, rOut} {
			px, py := r*ca, r*sa
			if px < minX {
				minX = px
			}
			if py < minY {
				minY = py
			}
			if px > maxX {
				maxX = px
			}
			if py > maxY {
				maxY = py
			}
		}
	}
	return round2(maxX - minX), round2(maxY - minY)
}

func buildFrustumCutItem(partNum int, name, material string, d1In, d2In, axialIn float64, qty int, notes string) CutSheetItem {
	dev, err := cad.ComputeFrustumDevelopment(d1In, d2In, axialIn)
	if err != nil || dev.IsCylinder {
		circ := round2(math.Pi * math.Max(d1In, d2In))
		return CutSheetItem{
			PartNum:      partNum,
			Name:         name,
			Shape:        "cylinder",
			Material:     material,
			FlatWidthIn:  circ,
			FlatHeightIn: round2(axialIn),
			Qty:          qty,
			Notes:        notes,
		}
	}
	w, h := frustumBoundingBoxIn(dev.InnerRadius, dev.OuterRadius, dev.ThetaRad)
	return CutSheetItem{
		PartNum:       partNum,
		Name:          name,
		Shape:         "frustum",
		Material:      material,
		FlatWidthIn:   w,
		FlatHeightIn:  h,
		OuterRadiusIn: round2(dev.OuterRadius),
		InnerRadiusIn: round2(dev.InnerRadius),
		SweepDeg:      round2(dev.ThetaDeg),
		Qty:           qty,
		Notes:         notes,
	}
}

// BuildShopPlan generates the structured ShopPlan (parts callouts, exact cut list
// with pkg/cad frustum flat-pattern developments, and code clearances).
func BuildShopPlan(
	project StudioProject,
	design model.ScrubberDesign,
	cots COTSSummary,
	urban UrbanAssessment,
	arch ArchitectureOptionsResult,
) ShopPlan {
	p := project.Normalize()

	ductIn := round1(design.Envelope.DuctDiameter.Inches())
	ductLenIn := round1(design.Input.DuctLength.Inches())
	quenchDiaIn := round1(cots.QuenchVessel.NominalIn)
	quenchHIn := round1(design.Quench.TowerHeight.Inches())
	colDiaIn := round1(cots.PackedColumn.NominalIn)
	colHIn := round1(design.PackedBed.ColumnHeight.Inches())
	stackHIn := round1(urban.Plume.MinStackHeightFt * 12.0)

	ventInIn := round1(design.Venturi.InletDiameter.Inches())
	ventThIn := round2(design.Venturi.ThroatDiameter.Inches())
	ventOutIn := round1(design.Venturi.OutletDiameter.Inches())
	ventConvIn := round1(design.Venturi.ConvergingLength.Inches())
	ventThLenIn := round1(design.Venturi.ThroatLength.Inches())
	ventDivIn := round1(design.Venturi.DivergingLength.Inches())

	overallLenIn := round1(arch.FootprintLengthFt * 12.0)
	overallWidthIn := round1(arch.FootprintWidthFt * 12.0)
	overallHeightIn := round1(math.Max(design.Envelope.Height.Inches(), math.Max(quenchHIn, colHIn)+18.0))

	// Find venturi, pump, and demister costs from COTS BOM for exact traceability.
	ventCostUSD := estimateSingleVenturiCost(p, design)
	if arch.ParallelVenturi.Enabled && !arch.Eductor.Enabled {
		ventCostUSD = arch.ParallelVenturi.MultiCostUSD
	}
	if arch.Eductor.Enabled {
		ventCostUSD = round2(285.0 + 24.0*arch.Eductor.CyclonicSweepElbowIn)
	}

	// Build numbered Parts Legend [1]..[9].
	stage2Label := "Stage 2: Wet Venturi Particulate Scrubber"
	stage2SKU := "VENT-COTS-RED"
	stage2DimsIn := fmt.Sprintf(`%.1f" -> %.2f" throat -> %.1f" (L %.1f")`, ventInIn, ventThIn, ventOutIn, ventConvIn+ventThLenIn+ventDivIn)
	stage2DimsMM := fmt.Sprintf("%.0f -> %.1f -> %.0f mm (L %.0f mm)", ventInIn*25.4, ventThIn*25.4, ventOutIn*25.4, (ventConvIn+ventThLenIn+ventDivIn)*25.4)
	stage2Qty := 1.0
	stage2Notes := fmt.Sprintf("%.0f m/s throat, d50=%.2f um NaCl fume; 2x opposed throat spray ports.", design.Venturi.ThroatVelocity, design.Venturi.CutDiameterD50Microns)
	if arch.Eductor.Enabled {
		stage2Label = "Stage 2: Vertical Downflow Static Venturi Eductor & 90° Cyclonic Sweep Elbow"
		stage2SKU = fmt.Sprintf("EDUCT-VSTAT-%.0f", arch.Eductor.CyclonicSweepElbowIn)
		stage2DimsIn = fmt.Sprintf(`%.0f" Tee -> %.2f" downflow throat -> %.0f" 90° sweep elbow`, arch.Eductor.CyclonicSweepElbowIn, ventThIn, arch.Eductor.CyclonicSweepElbowIn)
		stage2DimsMM = fmt.Sprintf("%.0f mm Tee -> %.1f mm throat -> %.0f mm sweep", arch.Eductor.CyclonicSweepElbowIn*25.4, ventThIn*25.4, arch.Eductor.CyclonicSweepElbowIn*25.4)
		stage2Notes = fmt.Sprintf("Dry ambient motive air (%.0f CFM) enters top vertically, entrains quenched gas (%.0f ACFM), and turns 90° into Stage 3 bottom plenum (+%.2f inWC positive pressure).",
			arch.Eductor.MotiveAirCFM, arch.Eductor.QuenchedACFM, arch.Eductor.PositivePressureInWC)
	} else if arch.ParallelVenturi.Enabled {
		stage2Label = fmt.Sprintf("Stage 2: %d× Parallel COTS Sanitary Mini-Venturis", arch.ParallelVenturi.ThroatCount)
		stage2SKU = fmt.Sprintf("SAN-MINI-VENT-%dX", arch.ParallelVenturi.ThroatCount)
		stage2DimsIn = fmt.Sprintf(`%d × (%.0f" -> %.2f" -> %.0f" tri-clamp)`, arch.ParallelVenturi.ThroatCount, arch.ParallelVenturi.InletFittingIn, arch.ParallelVenturi.ThroatDiameterIn, arch.ParallelVenturi.InletFittingIn)
		stage2DimsMM = fmt.Sprintf("%d × (%.0f -> %.1f -> %.0f mm)", arch.ParallelVenturi.ThroatCount, arch.ParallelVenturi.InletFittingIn*25.4, arch.ParallelVenturi.ThroatDiameterIn*25.4, arch.ParallelVenturi.InletFittingIn*25.4)
		stage2Qty = float64(arch.ParallelVenturi.ThroatCount)
		stage2Notes = fmt.Sprintf("Zero cone rolling; plugging risk: %s. %s", arch.ParallelVenturi.PluggingRisk, arch.ParallelVenturi.TurndownBenefit)
	}

	stage3Label := fmt.Sprintf("Stage 3: %s", cots.PackedColumn.Name)
	stage3DimsIn := fmt.Sprintf(`%.1f" nom × %.1f" H (%.1f" packing)`, colDiaIn, colHIn, design.PackedBed.PackingHeight.Inches())
	stage3DimsMM := fmt.Sprintf("%.0f mm × %.0f mm H (%.0f mm packing)", colDiaIn*25.4, colHIn*25.4, design.PackedBed.PackingHeight.Millimeters())
	stage3Qty := 1.0
	stage3Notes := fmt.Sprintf("%.1f%% of Eckert flooding velocity (limit <= 75%%); includes %s & 3-pass chevron demister.", cots.ActualPercentOfFlood, cots.PackingMedia.Name)
	if arch.ParallelBed.Enabled {
		stage3Label = fmt.Sprintf(`Stage 3: Parallel Manifold (%d × %.0f" Sch 80 CPVC Pipes)`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn)
		stage3DimsIn = fmt.Sprintf(`%d × %.0f" (ID %.3f") × %.1f" H`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn, arch.ParallelBed.PipeActualIDIn, colHIn)
		stage3DimsMM = fmt.Sprintf("%d × %.1f mm ID × %.0f mm H", arch.ParallelBed.PipeCount, arch.ParallelBed.PipeActualIDIn*25.4, colHIn*25.4)
		stage3Qty = float64(arch.ParallelBed.PipeCount)
		stage3Notes = fmt.Sprintf("%.1f%% flood, D/d_p=%.2f with %.3g\" PP packing, %.2f inWC bed dP; requires balanced %d-port top liquid distributor.",
			arch.ParallelBed.ActualFloodPct, arch.ParallelBed.WallRatioDToDp, arch.ParallelBed.PackingDiameterIn, arch.ParallelBed.BedDPInWC, arch.ParallelBed.PipeCount)
	}

	stage4Label := fmt.Sprintf("Stage 4: %s", cots.BlowerVFD.Name)
	stage4DimsIn := fmt.Sprintf(`%.1f" wheel, %.0f ACFM @ %.1f inWC`, cots.BlowerVFD.NominalIn, design.Blower.FlowACFM, design.Blower.TotalStaticPressureInWC)
	stage4DimsMM := fmt.Sprintf("%.0f mm wheel, %.2g HP VFD", cots.BlowerVFD.NominalIn*25.4, design.Blower.RecommendedMotorHP)
	stage4Notes := cots.BlowerVFD.Notes
	if arch.Eductor.Enabled {
		stage4Label = fmt.Sprintf("Stage 4 (Upstream Dry Motive): %s", cots.BlowerVFD.Name)
		stage4DimsIn = fmt.Sprintf(`%.0f CFM dry motive @ %.1f inWC (%.2g HP)`, arch.Eductor.MotiveAirCFM, arch.Eductor.DryFanStaticInWC, arch.Eductor.DryFanMotorHP)
		stage4DimsMM = fmt.Sprintf("%.0f m3/hr dry clean air, %.2g HP", arch.Eductor.MotiveAirCFM*1.699, arch.Eductor.DryFanMotorHP)
	}

	sumpLabel := cots.SumpTank.Name + " + Mag-Drive Pump"
	sumpQty := 1.0
	if !p.SharedBaffledSump {
		sumpLabel = "Dual Dedicated Sumps (Ash Settling + Absorber) + 2× Mag-Drive Pumps"
		sumpQty = 2.0
	}

	parts := []PlanPart{
		{
			Num:          1,
			ID:           "inlet_flue",
			Label:        "Hot Flue Inlet Duct & NFPA 86 Fail-Safe Bypass Tee",
			Stage:        "Inlet",
			Material:     cots.QuenchVessel.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" dia × %.1f" L`, ductIn, ductLenIn),
			DimensionsMM: fmt.Sprintf("%.0f mm dia × %.0f mm L", ductIn*25.4, ductLenIn*25.4),
			SKU:          fmt.Sprintf("FLUE-TEE-%02.0f", ductIn),
			Qty:          1,
			CostUSD:      round2(145.0 + 18.0*ductIn),
			Notes:        "Seal with >=2300 F ceramic fiber gaskets (NO silicone/PTFE); keep >=36\" clearance to combustibles (NFPA 211).",
		},
		{
			Num:          2,
			ID:           "quench_tower",
			Label:        "Stage 1: " + cots.QuenchVessel.Name,
			Stage:        "Stage 1",
			Material:     cots.QuenchVessel.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" nom × %.1f" H`, quenchDiaIn, quenchHIn),
			DimensionsMM: fmt.Sprintf("%.0f mm × %.0f mm H", quenchDiaIn*25.4, quenchHIn*25.4),
			SKU:          cots.QuenchVessel.SKU,
			Qty:          1,
			CostUSD:      cots.QuenchVessel.UnitCostUSD,
			Notes:        fmt.Sprintf("Co-current downward spray cools %.0f F -> %.1f F via %d × %s (%.2f GPM).", design.Input.PeakExitTempF, design.Quench.OutletTemp.Fahrenheit(), maxInt(design.Quench.NozzleCount, 2), cots.SprayNozzle.Name, design.Quench.TotalSprayWaterGPM),
		},
		{
			Num:          3,
			ID:           "venturi_stage",
			Label:        stage2Label,
			Stage:        "Stage 2",
			Material:     design.Venturi.Material,
			DimensionsIn: stage2DimsIn,
			DimensionsMM: stage2DimsMM,
			SKU:          stage2SKU,
			Qty:          stage2Qty,
			CostUSD:      ventCostUSD,
			Notes:        stage2Notes,
		},
		{
			Num:          4,
			ID:           "packed_bed",
			Label:        stage3Label,
			Stage:        "Stage 3",
			Material:     cots.PackedColumn.Material,
			DimensionsIn: stage3DimsIn,
			DimensionsMM: stage3DimsMM,
			SKU:          cots.PackedColumn.SKU,
			Qty:          stage3Qty,
			CostUSD:      cots.PackedColumn.UnitCostUSD,
			Notes:        stage3Notes,
		},
		{
			Num:          5,
			ID:           "sump_pump",
			Label:        sumpLabel,
			Stage:        "Stage 3 Sump",
			Material:     cots.SumpTank.Material,
			DimensionsIn: fmt.Sprintf(`%.1f" dia vented atmospheric drum`, cots.SumpTank.NominalIn),
			DimensionsMM: fmt.Sprintf("%.0f mm dia atmospheric sump", cots.SumpTank.NominalIn*25.4),
			SKU:          cots.SumpTank.SKU,
			Qty:          sumpQty,
			CostUSD:      round2(cots.SumpTank.UnitCostUSD + arch.SumpDeltaUSD),
			Notes:        "Atmospheric vented sump ONLY (never under blower vacuum); baffled ash settling zone + 50-mesh pump Y-strainer.",
		},
		{
			Num:          6,
			ID:           "blower_vfd",
			Label:        stage4Label,
			Stage:        "Stage 4",
			Material:     cots.BlowerVFD.Material,
			DimensionsIn: stage4DimsIn,
			DimensionsMM: stage4DimsMM,
			SKU:          cots.BlowerVFD.SKU,
			Qty:          1,
			CostUSD:      cots.BlowerVFD.UnitCostUSD,
			Notes:        stage4Notes,
		},
		{
			Num:          7,
			ID:           "safety_controls",
			Label:        cots.SensorKit.Name,
			Stage:        "Safety & Controls",
			Material:     cots.SensorKit.Material,
			DimensionsIn: fmt.Sprintf(`%.0f" fail-open bypass damper + NEMA 4X panel`, cots.SensorKit.NominalIn),
			DimensionsMM: fmt.Sprintf("%.0f mm bypass damper + 24VDC safety loop", cots.SensorKit.NominalIn*25.4),
			SKU:          cots.SensorKit.SKU,
			Qty:          1,
			CostUSD:      cots.SensorKit.UnitCostUSD,
			Notes:        fmt.Sprintf("Control tier: %s. Includes mandatory NFPA 86 hardwired 200 F snap-disc, 0.25 GPM flow switch & latching safety relay.", p.ControlTier),
		},
		{
			Num:          8,
			ID:           "exhaust_stack",
			Label:        "Outdoor Vertical Exhaust Stack (+ Silencer / Diluter)",
			Stage:        "Exhaust",
			Material:     "Sch 80 CPVC / Vinyl-Ester FRP (ASTM E84 Class 1)",
			DimensionsIn: fmt.Sprintf(`%.1f" dia × %.0f" H (%.0f ft min)`, ductIn, stackHIn, urban.Plume.MinStackHeightFt),
			DimensionsMM: fmt.Sprintf("%.0f mm dia × %.0f mm H", ductIn*25.4, stackHIn*25.4),
			SKU:          fmt.Sprintf("STACK-CPVC-%02.0f", ductIn),
			Qty:          1,
			CostUSD: func() float64 {
				c := round2(110.0 + 12.0*urban.Plume.MinStackHeightFt)
				if cots.Silencer != nil {
					c += cots.Silencer.UnitCostUSD
				}
				if cots.PlumeDiluter != nil {
					c += cots.PlumeDiluter.UnitCostUSD
				}
				return round2(c)
			}(),
			Notes: fmt.Sprintf("Terminate outdoors >=%.0f ft above grade with 1/2\" bottom condensate drain loop returning to sump.", urban.Plume.MinStackHeightFt),
		},
		{
			Num:          9,
			ID:           "code_utilities",
			Label:        "IPC 608 Backflow Preventer & NEC 430.102 Lockable Sight Disconnect",
			Stage:        "Utilities",
			Material:     "Lead-Free Bronze RPZ / Air-Gap + NEMA 4X Disconnect + 3% Line Reactor",
			DimensionsIn: `3/4" water isolation + NEMA 4X local disconnect <=50 ft`,
			DimensionsMM: "20 mm potable isolation + NEMA 4X sight disconnect",
			SKU:          "CODE-IPC608-NEC430",
			Qty:          1,
			CostUSD:      400.0,
			Notes:        fmt.Sprintf("Mandatory code utility kit: %s + IPC 608 Air Gap (>=1\") or ASSE 1013 RPZ.", urban.Electrical.CircuitLabel),
		},
	}

	// Build exact Cut Sheet (pipe cut lengths + unfolded cylinder/frustum flat patterns via pkg/cad).
	cutList := []CutSheetItem{
		{
			PartNum:      1,
			Name:         "Inlet Hot Flue Duct Pipe",
			Shape:        "pipe",
			Material:     cots.QuenchVessel.Material,
			FlatWidthIn:  round2(math.Pi * ductIn),
			FlatHeightIn: ductLenIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`Cut %.1f" dia pipe to %.1f" length (flat wrap %.2f" × %.1f"); add NFPA 91 cleanout hatch.`, ductIn, ductLenIn, math.Pi*ductIn, ductLenIn),
		},
		{
			PartNum:      2,
			Name:         "Stage 1 Quench Tower Shell",
			Shape:        "cylinder",
			Material:     cots.QuenchVessel.Material,
			FlatWidthIn:  round2(math.Pi * quenchDiaIn),
			FlatHeightIn: quenchHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`%.1f" dia × %.1f" tall shell (flat sheet %.2f" × %.1f"). Drill %d top spray ports.`, quenchDiaIn, quenchHIn, math.Pi*quenchDiaIn, quenchHIn, maxInt(design.Quench.NozzleCount, 2)),
		},
	}

	if arch.Eductor.Enabled {
		edIn := arch.Eductor.CyclonicSweepElbowIn
		cutList = append(cutList,
			buildFrustumCutItem(3, "Downflow Eductor Converging Cone", design.Venturi.Material, edIn, ventThIn, ventConvIn, 1,
				fmt.Sprintf(`Vertical downflow cone %.1f" -> %.2f" throat (unfolded annular sector).`, edIn, ventThIn)),
			buildFrustumCutItem(3, "Downflow Eductor Diffuser Cone", design.Venturi.Material, ventThIn, edIn, ventDivIn, 1,
				fmt.Sprintf(`Vertical downflow diffuser %.2f" -> %.1f" into 90° cyclonic sweep elbow.`, ventThIn, edIn)),
			CutSheetItem{
				PartNum:      3,
				Name:         `90° Long-Radius Cyclonic Sweep Elbow`,
				Shape:        "pipe",
				Material:     "Sch 80 CPVC / 316L SS",
				FlatWidthIn:  round2(math.Pi * edIn),
				FlatHeightIn: round2(1.5 * edIn * 0.5 * math.Pi),
				Qty:          1,
				Notes:        fmt.Sprintf(`%.0f" long-radius 90° sweep elbow with bottom 1.5" drain stub to sump (cyclonic droplet separation).`, edIn),
			},
		)
	} else if arch.ParallelVenturi.Enabled {
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      3,
				Name:         fmt.Sprintf(`Parallel Sanitary Mini-Venturi Spools (%.0f"->%.2f"->%.0f")`, arch.ParallelVenturi.InletFittingIn, arch.ParallelVenturi.ThroatDiameterIn, arch.ParallelVenturi.InletFittingIn),
				Shape:        "pipe",
				Material:     design.Venturi.Material,
				FlatWidthIn:  round2(math.Pi * arch.ParallelVenturi.ThroatDiameterIn),
				FlatHeightIn: round2(math.Max(3.0, 2.5*arch.ParallelVenturi.ThroatDiameterIn)),
				Qty:          arch.ParallelVenturi.ThroatCount,
				Notes:        fmt.Sprintf("Assemble %d identical COTS sanitary concentric reducer spools with tri-clamp tees (zero cone rolling).", arch.ParallelVenturi.ThroatCount),
			},
		)
	} else {
		cutList = append(cutList,
			buildFrustumCutItem(3, "Stage 2 Venturi Converging Cone", design.Venturi.Material, ventInIn, ventThIn, ventConvIn, 1,
				fmt.Sprintf(`Unfolded 25° included converging cone (%.1f" -> %.2f", axial %.1f").`, ventInIn, ventThIn, ventConvIn)),
			CutSheetItem{
				PartNum:      3,
				Name:         "Stage 2 Venturi Throat Spool",
				Shape:        "cylinder",
				Material:     design.Venturi.Material,
				FlatWidthIn:  round2(math.Pi * ventThIn),
				FlatHeightIn: ventThLenIn,
				Qty:          1,
				Notes:        fmt.Sprintf(`%.2f" ID × %.1f" L throat spool with 2× opposed NPT injection bosses.`, ventThIn, ventThLenIn),
			},
			buildFrustumCutItem(3, "Stage 2 Venturi Diverging Diffuser Cone", design.Venturi.Material, ventThIn, ventOutIn, ventDivIn, 1,
				fmt.Sprintf(`Unfolded 7° included pressure-recovery diffuser (%.2f" -> %.1f", axial %.1f").`, ventThIn, ventOutIn, ventDivIn)),
		)
	}

	if arch.ParallelBed.Enabled {
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      4,
				Name:         fmt.Sprintf(`Parallel %.0f" Sch 80 CPVC Absorber Risers`, arch.ParallelBed.PipeNominalIn),
				Shape:        "pipe",
				Material:     "Schedule 80 CPVC (ASTM F441)",
				FlatWidthIn:  round2(math.Pi * arch.ParallelBed.PipeActualIDIn),
				FlatHeightIn: colHIn,
				Qty:          arch.ParallelBed.PipeCount,
				Notes:        fmt.Sprintf(`Cut %d equal lengths of %.0f" Sch 80 CPVC pipe to %.1f" each; fill %.1f" depth with %.3g" PP packing.`, arch.ParallelBed.PipeCount, arch.ParallelBed.PipeNominalIn, colHIn, design.PackedBed.PackingHeight.Inches(), arch.ParallelBed.PackingDiameterIn),
			},
		)
	} else {
		topRedH := round1(math.Max(6.0, (colDiaIn-ductIn)*0.9))
		cutList = append(cutList,
			CutSheetItem{
				PartNum:      4,
				Name:         "Stage 3 Packed-Bed Column Shell",
				Shape:        "cylinder",
				Material:     cots.PackedColumn.Material,
				FlatWidthIn:  round2(math.Pi * colDiaIn),
				FlatHeightIn: colHIn,
				Qty:          1,
				Notes:        fmt.Sprintf(`%.1f" nom × %.1f" tall absorber shell; support grate at 12" above inlet, packing depth %.1f".`, colDiaIn, colHIn, design.PackedBed.PackingHeight.Inches()),
			},
			buildFrustumCutItem(4, "Stage 3 Top Column-to-Stack Reducer Cone", cots.PackedColumn.Material, colDiaIn, ductIn, topRedH, 1,
				fmt.Sprintf(`Transitions %.1f" absorber column to %.1f" outlet duct/stack.`, colDiaIn, ductIn)),
		)
	}

	cutList = append(cutList,
		CutSheetItem{
			PartNum:      8,
			Name:         "Outdoor Exhaust Stack Pipe",
			Shape:        "pipe",
			Material:     "Sch 80 CPVC / FRP (ASTM E84 Class 1)",
			FlatWidthIn:  round2(math.Pi * ductIn),
			FlatHeightIn: stackHIn,
			Qty:          1,
			Notes:        fmt.Sprintf(`%.1f" dia × %.0f" (%.0f ft) vertical outdoor stack with 1/2" condensate drain loop.`, ductIn, stackHIn, urban.Plume.MinStackHeightFt),
		},
	)

	propMinFt := 10.0
	if p.VaporGlazeMode == VaporGlazeSalt || design.PackedBed.HClOutletPPM > 0 {
		propMinFt = 30.0 // IMC 501.3.1 hazardous HCl exhaust setback
	}

	return ShopPlan{
		Title:            fmt.Sprintf("Complete Shop Plan — %s (%.1f cu.ft.)", p.StudioName, design.Input.KilnVolumeCuFt),
		StageAlignment:   p.StageAlignment,
		PackedBedLayout:  p.PackedBedLayout,
		VenturiLayout:    p.VenturiLayout,
		CompactFootprint: p.CompactFootprint,
		OverallLengthIn:  overallLenIn,
		OverallWidthIn:   overallWidthIn,
		OverallHeightIn:  overallHeightIn,
		Parts:            parts,
		CutList:          cutList,
		Clearances: PlanClearances{
			FlueToCombustiblesIn: 36.0,
			PropertyLineMinFt:    propMinFt,
			StackAboveRoofFt:     6.0,
			ServiceAisleIn:       36.0,
		},
	}
}
