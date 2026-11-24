package studio

import (
	"fmt"
	"math"
	"strings"
)

// GenerateBuildGuide produces a tailored, potter-friendly Markdown fabrication and
// assembly guide adapting to build_path, corrosion_tier, Phase 7 stage-design options,
// recirculation plan, and exact kiln dimensions.
func GenerateBuildGuide(eval StudioEvaluation) string {
	p := eval.Project.Normalize()
	if p.IsLocalCeramic() {
		return generateLocalCeramicBuildGuide(eval)
	}
	d := eval.Design
	c := eval.COTS
	u := eval.Urban
	a := eval.ArchitectureOptions
	r := eval.Recirculation

	var b strings.Builder

	fmt.Fprintf(&b, "# Build & Assembly Guide — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Kiln Profile:** %.1f cu.ft. | **Firing:** %s (%.0f F peak, %.0f hrs; %.1f hr salting) | **Glaze Mode:** `%s` | **Build Path:** `%s` | **Corrosion Tier:** %s (`%s`)\n",
		d.Input.KilnVolumeCuFt, p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours, p.EffectiveSaltingHours(),
		p.VaporGlazeMode, p.BuildPath, CorrosionTierLabel(p.CorrosionTier), p.CorrosionTier)
	fmt.Fprintf(&b, "> **Stage Architecture:** Alignment=`%s` | Packed Bed=`%s` (%.0f\" pipe) | Venturi=`%s` | Blower Drive=`%s` | Packing Media=`%s` | Shop Compressor=`%.0f SCFM` | Compact U-Turn=`%v` | Shared Baffled Sump=`%v` | Controls=`%s`\n\n",
		p.StageAlignment, p.PackedBedLayout, p.ParallelPipeSizeIn, p.VenturiLayout, p.BlowerDriveOption, p.PackingMediaOption, p.ShopCompressorSCFM, p.CompactFootprint, p.SharedBaffledSump, p.ControlTier)

	// 1. Key Dimensions at a Glance (inches + mm).
	b.WriteString("## 1. Key Dimensions at a Glance\n\n")
	b.WriteString("| Stage / Component | Studio Size (Imperial) | Engineering Size (Metric) | Notes |\n")
	b.WriteString("|---|---|---|---|\n")
	fmt.Fprintf(&b, "| Inlet Hot Flue Duct | %.1f\" dia × %.1f ft | %.0f mm × %.2f m | 316L SS / refractory lined |\n",
		d.Envelope.DuctDiameter.Inches(), d.Input.DuctLength.Feet(),
		d.Envelope.DuctDiameter.Millimeters(), float64(d.Input.DuctLength))
	fmt.Fprintf(&b, "| Stage 1: Quench Tower | %.1f\" COTS (calc %.1f\") × %.1f ft | %.0f mm × %.0f mm | Cools %.0f F -> %.1f F |\n",
		c.QuenchVessel.NominalIn, d.Quench.TowerDiameter.Inches(), d.Quench.TowerHeight.Feet(),
		d.Quench.TowerDiameter.Millimeters(), d.Quench.TowerHeight.Millimeters(),
		d.Input.PeakExitTempF, d.Quench.OutletTemp.Fahrenheit())
	if a.Eductor.Enabled {
		fmt.Fprintf(&b, "| Stage 2: Downflow Eductor | %.0f\" Tee -> %.2f\" throat -> %.0f\" 90° sweep | %.0f mm -> %.1f mm -> %.0f mm | Dry motive %.0f CFM, mixed %.1f F |\n",
			a.Eductor.CyclonicSweepElbowIn, d.Venturi.ThroatDiameter.Inches(), a.Eductor.CyclonicSweepElbowIn,
			a.Eductor.CyclonicSweepElbowIn*25.4, d.Venturi.ThroatDiameter.Millimeters(), a.Eductor.CyclonicSweepElbowIn*25.4,
			a.Eductor.MotiveAirCFM, a.Eductor.MixedTempF)
	} else if a.ParallelVenturi.Enabled {
		fmt.Fprintf(&b, "| Stage 2: Parallel Mini-Venturis | %d × (%.0f\" -> %.2f\" -> %.0f\") | %d × (%.0f -> %.1f -> %.0f mm) | Plugging risk: %s, %.1f inWC dP |\n",
			a.ParallelVenturi.ThroatCount, a.ParallelVenturi.InletFittingIn, a.ParallelVenturi.ThroatDiameterIn, a.ParallelVenturi.InletFittingIn,
			a.ParallelVenturi.ThroatCount, a.ParallelVenturi.InletFittingIn*25.4, a.ParallelVenturi.ThroatDiameterIn*25.4, a.ParallelVenturi.InletFittingIn*25.4,
			a.ParallelVenturi.PluggingRisk, a.ParallelVenturi.ThroatDPInWC)
	} else {
		fmt.Fprintf(&b, "| Stage 2: Wet Venturi Throat | %.1f\" -> %.2f\" throat -> %.1f\" | %.0f mm -> %.1f mm -> %.0f mm | %.1f m/s throat, d50=%.2f um |\n",
			d.Venturi.InletDiameter.Inches(), d.Venturi.ThroatDiameter.Inches(), d.Venturi.OutletDiameter.Inches(),
			d.Venturi.InletDiameter.Millimeters(), d.Venturi.ThroatDiameter.Millimeters(), d.Venturi.OutletDiameter.Millimeters(),
			d.Venturi.ThroatVelocity, d.Venturi.CutDiameterD50Microns)
	}
	if a.ParallelBed.Enabled {
		fmt.Fprintf(&b, "| Stage 3: Shell-Less Uniseal Bundle | %d × %.0f\" %s (ID %.3f\") | %d × %.1f mm ID (NO external large pipe) | %.1f%% flood, D/d_p=%.2f, %.2f inWC bed dP |\n",
			a.ParallelBed.PipeCount, a.ParallelBed.PipeNominalIn, a.ParallelBed.PipeMaterial, a.ParallelBed.PipeActualIDIn,
			a.ParallelBed.PipeCount, a.ParallelBed.PipeActualIDIn*25.4,
			a.ParallelBed.ActualFloodPct, a.ParallelBed.WallRatioDToDp, a.ParallelBed.BedDPInWC)
	} else {
		fmt.Fprintf(&b, "| Stage 3: Packed Column | %.1f\" COTS (calc %.1f\") × %.1f ft | %.0f mm × %.0f mm | %.1f%% of flood (limit <=75%%) |\n",
			c.PackedColumn.NominalIn, d.PackedBed.ColumnDiameter.Inches(), d.PackedBed.ColumnHeight.Feet(),
			d.PackedBed.ColumnDiameter.Millimeters(), d.PackedBed.ColumnHeight.Millimeters(),
			c.ActualPercentOfFlood)
	}
	fmt.Fprintf(&b, "| Packing Bed Depth & Media | %.1f\" depth (%s) | %.0f mm (%.2f m) | voidage %.0f%%, %.2f inWC Ergun dP |\n",
		d.PackedBed.PackingHeight.Inches(), a.PackingMedia.Label, d.PackedBed.PackingHeight.Millimeters(), float64(d.PackedBed.PackingHeight),
		a.PackingMedia.Voidage*100.0, a.PackingMedia.BedDPInWC)
	if a.Eductor.Enabled && p.BlowerDriveOption == BlowerDriveIndustrialRadialVFD {
		fmt.Fprintf(&b, "| Stage 4: Dry Motive Fan & Stack | %.2g HP dry fan, %.0f CFM @ %.1f inWC | Clean ambient air (zero acid contact) | +%.2f inWC bed pressure; min stack %.0f ft |\n",
			a.Eductor.DryFanMotorHP, a.Eductor.MotiveAirCFM, a.Eductor.DryFanStaticInWC, a.Eductor.PositivePressureInWC, u.Plume.MinStackHeightFt)
	} else {
		fmt.Fprintf(&b, "| Stage 4: %s | %.2g HP (%dV), %.0f ACFM @ %.1f inWC | $%.0f (%+.0f USD vs baseline) | Feasible=%v; min stack %.0f ft |\n",
			a.BlowerDrive.Label, a.BlowerDrive.MotorHP, a.BlowerDrive.Voltage, d.Blower.FlowACFM, a.BlowerDrive.RequiredStaticInWC,
			a.BlowerDrive.FanCostUSD, a.BlowerDrive.CostDeltaUSD, a.BlowerDrive.Feasible, u.Plume.MinStackHeightFt)
	}
	fpNote := fmt.Sprintf("Estimated COTS total: **$%.2f**", c.EstimatedTotalUSD)
	if a.CostDeltaUSD != 0 {
		fpNote += fmt.Sprintf(" (%+.0f USD vs baseline)", a.CostDeltaUSD)
	}
	if a.CompactFootprint {
		fpNote += fmt.Sprintf("; U-turn cuts length %.1f%%", a.FootprintSavingsPct)
	}
	fmt.Fprintf(&b, "| Overall Skid Footprint | %.1f ft × %.1f ft × %.1f ft H | %.2f m × %.2f m × %.2f m | %s |\n\n",
		a.FootprintLengthFt, a.FootprintWidthFt, d.Envelope.Height.Feet(),
		a.FootprintLengthFt*0.3048, a.FootprintWidthFt*0.3048, float64(d.Envelope.Height),
		fpNote)

	// 2. Tools & Shop Prep (adapts to build_path).
	b.WriteString("## 2. Tools & Shop Preparation\n\n")
	switch p.BuildPath {
	case BuildPathCOTSNoWeld:
		b.WriteString("**Build Path: No-Weld COTS Hardware (`cots_no_weld`)** — designed for a home workshop using bolted flanges, stainless band clamps, and solvent-welded CPVC pipe (no TIG welder needed):\n\n")
		b.WriteString("- Cordless drill/driver with step bits & bi-metal hole saws (`1\"` to `4\"` for bulkhead and nozzle ports)\n")
		b.WriteString("- Socket/wrench set for 316 SS flange bolts & high-torque stainless band clamps\n")
		b.WriteString("- Heavy-duty CPVC primer (ASTM F656) & Schedule 80 CPVC solvent cement (ASTM F493)\n")
		b.WriteString("- `>= 2300 F` alumina-silica ceramic fiber rope gaskets / refractory mortar for dry hot kiln flue joints upstream of Stage 1 (NEVER use 650 F RTV silicone or PTFE on a 2400 F dry flue!), plus expanded PTFE/EPDM flange gaskets strictly downstream of Stage 1 (`<= 200 F`)\n")
		b.WriteString("- Basic multimeter, wire strippers, brine hydrometer (`SG 1.000–1.200`), and a digital manometer or Dwyer Magnehelic gauge (`0 to -1.0 inWC`)\n\n")
	case BuildPathCustomFabDeluxe:
		b.WriteString("**Build Path: Custom Fabrication Deluxe (`custom_fab_deluxe`)** — turnkey industrial-grade TIG / Derakane FRP / welded forklift skid / UL 508A NEMA 4X panel / Rockwool acoustic enclosure build:\n\n")
		b.WriteString("- Send `scrubber.dxf` (units: **millimeters**) and `shop_plan.json` (ANSI/ASME B16.5 Class 150 bolt circles, steel skid cuts & acoustic baffle panel dimensions) to a CNC laser/waterjet & FRP fabricator\n")
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			b.WriteString("- TIG welder with **ERNiCrMo-4 (Alloy C-276)** filler rod, argon back-purge monitor (`< 50 ppm O2`), interpass temp stick (`<= 200 F`), and heavy-gauge `11 ga / Sch 10S` slip rolls\n")
		} else {
			b.WriteString("- TIG welder with **ER316L** filler rod, argon back-purge monitor (`< 50 ppm O2`), stainless pickling/passivation paste, and heavy-gauge `11 ga / Sch 10S` slip rolls\n")
		}
		b.WriteString("- Calibrated torque wrench (`20–100 ft-lb`), nickel anti-seize for 316 SS ANSI B16.5 Class 150 flange studs, and PTFE envelope gaskets\n")
		b.WriteString("- MIG/TIG structural welder for `2\"×2\"×3/16\"` steel forklift skid + diagonal K-bracing, rotary hammer drill (`1/2\"` masonry bit) for stainless concrete wedge anchors\n")
		b.WriteString("- Acoustic fabrication tools for `2\"` Rockwool Safe'n'Sound panels + perforated metal inner liner + baffled TEFC motor cooling-air labyrinth\n")
		b.WriteString("- Digital multimeter, clamp ammeter, megohmmeter (for VFD motor insulation check), dBA sound level meter, and brine hydrometer (`SG 1.000–1.200`)\n\n")
	case BuildPathLocalCeramic:
		b.WriteString("**Build Path: Local Ceramic & Masonry (`local_ceramic_masonry`)** — designed for Nairobi, Mumbai, and local pottery studios using fired firebrick, IS 651 salt-glazed stoneware (SWG) pipes, potter-extruded stoneware Raschig rings, terracotta baffles, an elevated HDPE gravity tank, and manual treadle/hand pumping (or a local 0.5 HP 230V 50Hz domestic pump):\n\n")
		b.WriteString("- Masonry trowel, spirit level, plumb bob, mortar mixing tub, and firebrick / lime-sand refractory mortar\n")
		b.WriteString("- Studio clay extruder (or hollow wooden dowel) + wire cutter to produce `1\"` unglazed grogged stoneware Raschig rings fired in your own kiln to Cone 6–10\n")
		b.WriteString("- Hand saw, drill, wrenches, and M12 through-bolts for the **diagonally X-braced `3.5 m (11.5 ft)` overhead gravity tank stand** (`200–500 kg` water load!)\n")
		b.WriteString("- Basic plumbing cutters & PVC/HDPE pipe fittings for the gravity downcomer, brass/PVC spray rose, foot valve, and drain valve\n")
		b.WriteString("- `$10` brine hydrometer (`SG 1.000–1.200`), litmus pH test strips (`pH 1–14`), and mechanical bimetallic stem thermometer\n\n")
	default:
		b.WriteString("**Build Path: Rolled Sheet Metal & DXF Fabrication (`welded_dxf`)** — uses the exact flat patterns in `scrubber.dxf` and `shop_plan.json` for laser/plasma cutting and TIG welding:\n\n")
		b.WriteString("- Send `scrubber.dxf` (units: **millimeters**) to a local CNC laser/waterjet shop for Stage 1–2 shells, venturi cones, and bolted flanges\n")
		b.WriteString("- Sheet-metal slip rolls for cylindrical shells and frustum venturi cones\n")
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			b.WriteString("- TIG welder with **ERNiCrMo-4 (Alloy C-276)** filler rod + argon back-purge (or castable acid-proof refractory trowel kit) and vinyl-ester FRP layup kit\n")
		} else {
			b.WriteString("- TIG welder with **ER316L** filler rod, argon back-purge for smooth root welds, and stainless pickling/passivation paste\n")
		}
		b.WriteString("- Plastic hot-air extrusion welder or Schedule 80 CPVC primer & solvent cement for Stage 3 absorber internals\n")
		b.WriteString("- Torque wrench, expanded PTFE full-face flange gaskets, brine hydrometer (`SG 1.000–1.200`), and 316 SS hex bolts/nuts\n\n")
	}

	// 3. Corrosion & Materials Note.
	b.WriteString("## 3. Materials, Corrosion Protection & Selected Options\n\n")
	fmt.Fprintf(&b, "- **Selected Tier (%s):** %s\n", CorrosionTierLabel(p.CorrosionTier), u.Corrosion.MaterialSummary)
	fmt.Fprintf(&b, "- **Maintenance Guidance:** %s\n", u.Corrosion.SodaComparison)
	fmt.Fprintf(&b, "- **Inspection Schedule:** %s\n", u.Corrosion.InspectionInterval)
	fmt.Fprintf(&b, "- **Blower / Motive Drive (`%s`):** %s\n", a.BlowerDrive.Option, a.BlowerDrive.Summary)
	for _, cav := range a.BlowerDrive.Caveats {
		fmt.Fprintf(&b, "  - *Blower Note:* %s\n", cav)
	}
	fmt.Fprintf(&b, "- **Packing Media (`%s`):** %s\n", a.PackingMedia.Option, a.PackingMedia.Summary)
	fmt.Fprintf(&b, "- **Shop Compressor Evaluation (`%.0f SCFM @ 90 PSI`):**\n", a.Compressor.AvailableSCFM)
	fmt.Fprintf(&b, "  - *Main Airflow Eductor (`feasible=%v`, needs %.1f SCFM / %.1f HP):* %s\n", a.Compressor.EductorFeasible, a.Compressor.EductorSCFMNeeded, a.Compressor.EductorHPNeeded, a.Compressor.EductorVerdict)
	fmt.Fprintf(&b, "  - *Twin-Fluid Air-Atomizing Nozzle (`feasible=%v`, needs %.1f SCFM @ 40 PSI, saves %.1f inWC & $%.0f):* %s\n", a.Compressor.AtomizeFeasible, a.Compressor.AtomizeSCFMNeeded, a.Compressor.AtomizeDPSavedInWC, a.Compressor.AtomizeCostSavedUSD, a.Compressor.AtomizeVerdict)
	for _, s := range a.SavingsSummary {
		fmt.Fprintf(&b, "- **Selected Option:** %s\n", s)
	}
	for _, w := range a.Warnings {
		fmt.Fprintf(&b, "- **WARNING:** %s\n", w)
	}
	b.WriteString("\n")

	// 4. Step-by-Step Assembly.
	b.WriteString("## 4. Step-by-Step Assembly\n\n")

	b.WriteString("### Step 1 — Stage 1: Hot Flue Duct & Adiabatic Spray Quench Tower\n")
	fmt.Fprintf(&b, "1. Mount the **%s** vertically on a level steel or masonry pad with a bottom drain leading into the sump.\n", c.QuenchVessel.Name)
	fmt.Fprintf(&b, "2. **NFPA 211 High-Heat Flue Clearances:** Connect the %.1f\" (%.0f mm) hot inlet flue from the kiln bagwall/damper slot to the **top** of the quench tower so flue gas flows **co-currently downward** with the spray mist. Seal hot dry flue joints strictly with **`>= 2300 F` ceramic fiber gaskets / refractory mortar** (never RTV silicone or PTFE). Maintain **>= 36\" clearance to unprotected combustibles** (NFPA 211), never run a high-heat flue through a combustible wall/partition, and include an **NFPA 91 cleanout/access hatch** at the flue elbow for ash removal.\n",
		d.Envelope.DuctDiameter.Inches(), d.Envelope.DuctDiameter.Millimeters())
	fmt.Fprintf(&b, "3. Install **%d × %s** evenly spaced near the top spray header (`%.2f GPM` total spray). Aim cones cleanly down the centerline so droplets atomize before hitting the walls — supply at **60–100 psig** (or twin-fluid air-atomizing) to keep droplet size `<= 200 um`.\n",
		maxInt(d.Quench.NozzleCount, 2), c.SprayNozzle.Name, d.Quench.TotalSprayWaterGPM)
	if p.VaporGlazeMode == VaporGlazeSalt && p.CorrosionTier != CorrosionTierHeavyC276FRP {
		b.WriteString("4. **Standard 316L Salt Protection:** Plumb the quench spray manifold from the **alkaline NaHCO3 recirculation pump** (pH 8.0–8.5) rather than unbuffered acidic tap water, with a clean-water backup line for emergency cooling.\n")
	}
	b.WriteString("5. Install the **Type-K thermocouple** (or mechanical stem thermometer in electromechanical mode) at the quench inlet and the hardwired **200 F snap-disc + Pt100 RTD** immediately at the quench exit (upstream of any plastic/CPVC transition).\n\n")

	if a.Eductor.Enabled {
		b.WriteString("### Step 2 — Stage 2: Dry-Fan Vertical-Downflow Static Venturi Eductor & 90° Cyclonic Sweep\n")
		fmt.Fprintf(&b, "1. Route the Stage 1 quench exit out and down into the side branch of a **%.0f\" flanged Tee**. Mount the clean dry-air radial motive fan (**%.2g HP, %.0f CFM @ %.1f inWC**) directly above/beside pushing dry ambient air straight **DOWN** into the top of the vertical static venturi throat.\n",
			a.Eductor.CyclonicSweepElbowIn, a.Eductor.DryFanMotorHP, a.Eductor.MotiveAirCFM, a.Eductor.DryFanStaticInWC)
		fmt.Fprintf(&b, "2. Mount 2 opposed spray injection ports at the vertical throat entry (`%.2f GPM` liquid flow). Mixing ambient motive air cools the Stage 3 inlet from `%.0f F` down to **`%.1f F`** (`+%.1f F` extra thermal margin below the 200 F CPVC limit).\n",
			d.Venturi.LiquidFlowGPM, d.Quench.OutletTemp.Fahrenheit(), a.Eductor.MixedTempF, a.Eductor.ThermalMarginGainF)
		fmt.Fprintf(&b, "3. Attach a **%.0f\" long-radius 90° sweep elbow** directly beneath the vertical diffuser with a `1.5\"` bottom liquid drain loop (`>= 6\"` water seal leg into the sump) so centrifugal force spins out heavy droplets before gas enters the Stage 3 bottom plenum.\n",
			a.Eductor.CyclonicSweepElbowIn)
		fmt.Fprintf(&b, "4. **OUTDOOR-ONLY POSITIVE PRESSURE RULE:** Stage 3 operates at **`+%.2f inWC` positive pressure** — bolt every downstream flange with full-face expanded PTFE gaskets and soap-bubble leak-test under cold-air fan pressure before firing.\n\n",
			a.Eductor.PositivePressureInWC)
	} else if a.ParallelVenturi.Enabled {
		b.WriteString("### Step 2 — Stage 2: Parallel COTS Sanitary Mini-Venturi Manifold\n")
		fmt.Fprintf(&b, "1. Assemble **%d parallel sanitary concentric-reducer spools** (`%.0f\" -> %.2f\" throat -> %.0f\"` tri-clamp, zero custom cone rolling) between inlet and outlet stainless collector headers (`%.1f inWC` dP).\n",
			a.ParallelVenturi.ThroatCount, a.ParallelVenturi.InletFittingIn, a.ParallelVenturi.ThroatDiameterIn, a.ParallelVenturi.InletFittingIn, a.ParallelVenturi.ThroatDPInWC)
		fmt.Fprintf(&b, "2. Split the `%.2f GPM` venturi spray flow evenly across all %d throats. **Turndown Tip:** %s\n",
			d.Venturi.LiquidFlowGPM, a.ParallelVenturi.ThroatCount, a.ParallelVenturi.TurndownBenefit)
		if a.ParallelVenturi.PluggingRisk == "High" {
			fmt.Fprintf(&b, "3. **Ash Plugging Watch (`High`):** Because individual throats (`%.2f\"`) are under `1.0\"` in wood-ash service, install an upstream ember screen and flush tri-clamp throats after each firing.\n\n", a.ParallelVenturi.ThroatDiameterIn)
		} else {
			b.WriteString("3. Bolt the collector outlet into the Stage 3 lower gas plenum using an expanded PTFE full-face gasket.\n\n")
		}
	} else {
		b.WriteString("### Step 2 — Stage 2: Wet Venturi Particulate Scrubber\n")
		if p.BuildPath == BuildPathCOTSNoWeld {
			fmt.Fprintf(&b, "1. Assemble the no-weld concentric bolted reducers transitioning from `%.1f\"` down to the `%.2f\"` (`%.1f mm`) throat spool and smoothly diverging back to `%.1f\"`.\n",
				d.Venturi.InletDiameter.Inches(), d.Venturi.ThroatDiameter.Inches(),
				d.Venturi.ThroatDiameter.Millimeters(), d.Venturi.OutletDiameter.Inches())
		} else {
			fmt.Fprintf(&b, "1. Roll and TIG-weld the converging cone (`%.0f mm` len), throat (`%.0f mm` len, `%.1f mm` ID), and diverging diffuser (`%.0f mm` len) from `scrubber.dxf`, then bolt on both end flanges.\n",
				d.Venturi.ConvergingLength.Millimeters(), d.Venturi.ThroatLength.Millimeters(),
				d.Venturi.ThroatDiameter.Millimeters(), d.Venturi.DivergingLength.Millimeters())
		}
		if p.BlowerDriveOption == BlowerDriveCompressorAirAtomize {
			fmt.Fprintf(&b, "2. **Twin-Fluid Air-Atomizing Venturi Nozzle (`%.1f SCFM @ 40 PSI`):** Mount the `$75` 316L twin-fluid air-atomizing nozzle cross-flow/counter-current at the venturi throat entry (`%.2f GPM` water + `%.1f SCFM` shop air through an inline regulator & moisture trap) to generate **`25 um` SMD droplets** at reduced `%.0f m/s` throat velocity (active twin-fluid jet shear `v_rel ~75 m/s` achieves Calvert `d50 = %.2f um` parity while slashing total static pressure to `%.1f inWC`).\n",
				a.Compressor.AtomizeSCFMNeeded, d.Venturi.LiquidFlowGPM, a.Compressor.AtomizeSCFMNeeded, d.Venturi.ThroatVelocity, activeTwinFluidD50Microns(d), a.BlowerDrive.RequiredStaticInWC)
		} else {
			fmt.Fprintf(&b, "2. Mount 2 opposed spray injection ports right at the venturi throat entry (`%.2f GPM` liquid flow) so high-velocity `%.0f m/s` gas shears the water into fine droplets that capture submicron ash and salt fume (`d50 = %.2f um`).\n",
				d.Venturi.LiquidFlowGPM, d.Venturi.ThroatVelocity, d.Venturi.CutDiameterD50Microns)
		}
		b.WriteString("3. Bolt the venturi exit into the lower gas inlet plenum of Stage 3 using an expanded PTFE full-face gasket.\n\n")
	}

	b.WriteString("### Step 3 — Stage 3: Packed-Bed Absorber Column, Closed-Loop Sump & Demister\n")
	if p.SharedBaffledSump {
		fmt.Fprintf(&b, "1. Position the **%s** (single shared baffled settling + neutralizing sump, ~%.0f gal working fill) beneath/beside **%s** so scrubber liquor drains freely by gravity.\n",
			c.SumpTank.Name, r.SumpVolumeGal, c.PackedColumn.Name)
	} else {
		fmt.Fprintf(&b, "1. Position the dual dedicated sumps (ash-settling sump + absorber neutralizing sump, ~%.0f gal combined fill) beneath Stage 1-2 and **%s**.\n",
			r.SumpVolumeGal, c.PackedColumn.Name)
	}
	if !a.Eductor.Enabled {
		sealDropIn := math.Max(26.0, math.Ceil(d.Blower.TotalStaticPressureInWC+6.0))
		fmt.Fprintf(&b, "   - **MANDATORY Barometric Water-Seal Leg (Negative-Pressure ID Blower):** Route the Stage 3 gravity return pipe into the atmospheric vented sump through a submerged U-trap / barometric seal leg with at least **`%.0f\"` vertical drop** (exceeding the `%.1f inWC` blower suction + 6\" margin) so suction never pulls ambient air backward up the drain or holds liquid up in the column.\n",
			sealDropIn, d.Blower.TotalStaticPressureInWC)
	}
	b.WriteString("2. **CPVC Solvent-Welding Best Practice:** Chamfer pipe edges, wipe clean, apply purple ASTM F656 primer liberally to both pipe and socket, apply heavy-bodied Schedule 80 CPVC cement, seat with a 1/4-turn twist, hold 30 seconds, and allow **24 hours full cure** before hot-water or firing service.\n")
	if a.ParallelBed.Enabled {
		fmt.Fprintf(&b, "3. **Shell-Less Uniseal %.0f\" Parallel Pipe Bundle (NO External Large Pipe):** Hole-saw **%d ports** in the top and bottom `1/2\"` HDPE/CPVC flat collector plates (`$95`), press in **%d rubber Uniseals** (`$5.50` each), slide in **%d bare risers** of `%.0f\"` %s (`ID %.3f\"`), seat packing support screens at the bottom of every pipe, fill each riser to **%.1f\" depth** with **%s** (`D/d_p = %.2f >= 7.5`), and mount the **balanced %d-port calibrated liquid spray manifold** (`$65`, `%.1f GPM` total) above the top plate.\n",
			a.ParallelBed.PipeNominalIn, a.ParallelBed.PipeCount, 2*a.ParallelBed.PipeCount, a.ParallelBed.PipeCount, a.ParallelBed.PipeNominalIn, a.ParallelBed.PipeMaterial, a.ParallelBed.PipeActualIDIn,
			d.PackedBed.PackingHeight.Inches(), c.PackingMedia.Name, a.ParallelBed.WallRatioDToDp, a.ParallelBed.PipeCount, d.PackedBed.RecirculationFlowGPM)
	} else {
		fmt.Fprintf(&b, "3. Install the lower polypropylene/FRP packing support grate at least `12\"` above the gas inlet port, pour in **%s** to a bed depth of **%.1f\" (%.2f m)**, and pin the top hold-down grid so gas surges cannot fluidize the rings.\n",
			c.PackingMedia.Name, d.PackedBed.PackingHeight.Inches(), float64(d.PackedBed.PackingHeight))
		fmt.Fprintf(&b, "4. Mount the top liquid distributor nozzle/header (`%.1f GPM` counter-current recirculation from the mag-drive pump via a 50-mesh inline Y-strainer).\n", d.PackedBed.RecirculationFlowGPM)
	}
	if p.PackingMediaOption == PackingMediaDIYCeramicRings {
		b.WriteString("   - **DIY Stoneware Rings Support Note:** Unglazed kiln-fired stoneware rings (~46 lb/cu.ft., 70% voidage) are acid-proof and 2000 F-proof, but heavy — reinforce the lower packing support screen with FRP bar grating or 316L stainless cross-rods.\n")
	}
	fmt.Fprintf(&b, "5. Seat the **3-pass polypropylene chevron mist eliminator** (`%.1f\"` thick) above the spray distributor so clean exhaust exits mist-free.\n\n", d.PackedBed.DemisterHeight.Inches())

	b.WriteString("### Step 4 — Stage 4: Blower, Electrical & Acoustic Outdoor Stack\n")
	if a.Eductor.Enabled {
		fmt.Fprintf(&b, "1. Mount the **%s** on rubber vibration-isolation pads on the dry ambient intake side feeding the vertical downflow eductor (no wet downstream blower on the packed-bed outlet!). Include **NFPA 91** cleanout/access hatches and low-point condensate drains on horizontal wet ducts, and use **ASTM E84 Class 1 (FSI <= 25 / SDI <= 50)** CPVC/FRP ductwork.\n", c.BlowerVFD.Name)
	} else {
		fmt.Fprintf(&b, "1. Mount the **%s** on rubber vibration-isolation pads downstream of the column demister (induced-draft suction keeps the entire scrubber under slight negative pressure so fumes cannot leak out). Include **NFPA 91** cleanout/access hatches and low-point condensate drains on horizontal wet ducts, and use **ASTM E84 Class 1 (FSI <= 25 / SDI <= 50)** CPVC/FRP ductwork.\n", c.BlowerVFD.Name)
	}
	usesVFD := p.BlowerDriveOption == BlowerDriveIndustrialRadialVFD || p.BlowerDriveOption == BlowerDriveCompressorAirAtomize
	if p.BlowerDriveOption == BlowerDriveCompressorEductor {
		fmt.Fprintf(&b, "2. **Electrical Hookup (No Electric Blower):** Wire the recirculation pump and 24 VDC safety loop to a **%s** (`%.1f A` min ampacity) via a **GFCI breaker (`NEC 210.8(F)`)**. %s\n", u.Electrical.CircuitLabel, u.Electrical.RequiredAmps, u.Electrical.Summary)
	} else if !usesVFD {
		fmt.Fprintf(&b, "2. **NEC Electrical Hookup (Non-VFD Drive):** Wire the blower motor in fixed conduit (**NEC 400.12**) to a **%s** (`%.1f A` min ampacity) via a **lockable NEMA 4X sight disconnect (`<= 50 ft`, NEC 430.102)** and **GFCI breaker (`NEC 210.8(F)`)**. %s\n", u.Electrical.CircuitLabel, u.Electrical.RequiredAmps, u.Electrical.Summary)
	} else {
		fmt.Fprintf(&b, "2. **NEC Electrical Hookup:** Wire the VFD in fixed conduit (**NEC 400.12**) to a **%s** (`%.1f A` min ampacity per **NEC 430.122(A)**) via a **lockable NEMA 4X sight disconnect (`<= 50 ft`, NEC 430.102)**, **line-side HF-rated GFCI breaker (`NEC 210.8(F)`)**, **3%% input line reactor**, and shielded VFD motor cable. %s\n", u.Electrical.CircuitLabel, u.Electrical.RequiredAmps, u.Electrical.Summary)
	}
	if c.Silencer != nil {
		fmt.Fprintf(&b, "3. Install the **%s** and enclose the blower housing with foil-faced Rockwool panels (leaving motor cooling vents unobstructed) — predicted property-line sound: **%.1f dBA at %.0f ft**.\n",
			c.Silencer.Name, u.Noise.PropertyLineDBA, u.Noise.PropertyLineFt)
	} else {
		fmt.Fprintf(&b, "3. Connect a vertical discharge stack rising at least **%.0f ft** above grade (predicted sound: **%.1f dBA at %.0f ft**; wear hearing protection per OSHA 29 CFR 1910.95 near an unsilenced blower or add an acoustic silencer).\n",
			u.Plume.MinStackHeightFt, u.Noise.PropertyLineDBA, u.Noise.PropertyLineFt)
	}
	if p.VaporGlazeMode == VaporGlazeSalt || d.PackedBed.HClOutletPPM > 0 {
		fmt.Fprintf(&b, "4. **Outdoor Stack & IMC 510 / 501.3.1 Hazardous HCl Setbacks:** Salt-firing stack exhaust contains **~%.1f ppm HCl** (above the `5 ppm` **OSHA 29 CFR 1910.1000** ceiling PEL — NEVER discharge indoors!). Terminate the vertical stack strictly outdoors at least **%.0f ft** above grade, **>= 30 ft** from property lines, **>= 10 ft** from operable windows/doors/outdoor air intakes, and **>= 6 ft** from exterior walls/roofs, with a `1/2\"` bottom condensate drain loop returning to the sump.\n",
			d.PackedBed.HClOutletPPM, u.Plume.MinStackHeightFt)
	} else {
		fmt.Fprintf(&b, "4. **Outdoor Stack & IMC 501.3.1 Setbacks (`0 ppm HCl`):** Terminate the vertical discharge stack outdoors at least **%.0f ft** above grade, **>= 10 ft** from property lines and operable windows/air intakes, and **>= 3 ft** from exterior walls/roofs, with a `1/2\"` bottom condensate drain loop returning condensed water vapor to the sump.\n",
			u.Plume.MinStackHeightFt)
	}
	if c.PlumeDiluter != nil {
		fmt.Fprintf(&b, "5. **Optional Squirrel-Cage Plume Dilution Assembly (`%s`):** Install the **%s** strictly **downstream** of the Stage 4 blower (never in place of the %.1f inWC radial blower) via a **CPVC/FRP/316L** Y/tee mixing plenum (never galvanized steel in wet HCl exhaust) topped with the **%.0f\" high-velocity exit cone (`~3,400 FPM`)**. This shortens horizontal winter steam reach from **~%.0f ft bare down to ~%.0f ft** (dissipating before your %.0f ft property line). *Optional Waste-Heat Upgrade:* Mount the squirrel-cage blower on the **cool intake side** blowing ambient air through a sheet-metal shroud around the kiln exterior/hot flue into the mixing tee to recover free kiln heat and shrink the steam cloud even further!\n\n",
			c.PlumeDiluter.SKU, c.PlumeDiluter.Name, d.Blower.TotalStaticPressureInWC, c.PlumeDiluter.NominalIn,
			u.Plume.BarePlumeLengthFt, u.Plume.DilutedPlumeLengthFt, p.PropertyLineFt)
	} else {
		fmt.Fprintf(&b, "5. **Plume Note:** Bare winter steam reach is ~%.0f ft. If your neighbor's fence is close, enable `include_plume_diluter` for a 120V %.0f CFM squirrel-cage dilution blower + CPVC/FRP/316L mixing tee & high-velocity exit cone (~%.0f ft diluted reach).\n\n",
			u.Plume.BarePlumeLengthFt, u.Plume.DilutionCFMNeeded, u.Plume.DilutedPlumeLengthFt)
	}

	b.WriteString("### Step 5 — NFPA 86 Independent Hardwired Safety Loop & Fail-Safe Bypass Damper\n")
	b.WriteString("1. **NEVER skip this step (NFPA 86 / ASTM F441):** Stage 3 Sch 80 CPVC/PP components are rated to `200 F` max continuous. Install the **spring-return fail-open hot-gas bypass damper** on a refractory/316L tee between the kiln flue and an **NFPA 211 high-heat auxiliary bypass chimney** (`>= 36\"` clearance to combustibles).\n")
	if p.ControlTier == ControlTierElectromechanicalOnly {
		b.WriteString("2. **Electromechanical Safety & Manual Control Panel (`electromechanical_only`):** Zero software required — wire the 24 VDC spring-return actuator and blower Run-Enable/contactor directly through the **latching manual-reset safety relay** driven by:\n")
		b.WriteString("   - **Hardwired thermal limit:** Independent `200 F (93.3 C)` manual-reset bimetallic snap-disc thermostat on the quench outlet transition\n")
		b.WriteString("   - **Low-flow switch:** Mechanical paddle quench spray flow switch (`< 0.25 GPM`)\n")
		if usesVFD {
			b.WriteString("   - **E-Stop & Manual Draft Dial:** NEMA 4X twist-to-release E-Stop (NC) + analog Dwyer Magnehelic draft gauge + manual VFD frequency knob\n")
		} else {
			b.WriteString("   - **E-Stop & Manual Blast Gate:** NEMA 4X twist-to-release E-Stop (NC) + analog Dwyer Magnehelic draft gauge + manual intake/bleed damper\n")
		}
	} else {
		b.WriteString("2. **Independent Hardwired 24 VDC Safety Loop (De-Energize to Trip — Overrides Software/Pi):** Wire the 24 VDC spring-return actuator and VFD Run-Enable/STO through a **latching manual-reset safety relay** driven by normally-closed hardwired field switches:\n")
		b.WriteString("   - **Primary hardwired thermal limit:** Independent `200 F (93.3 C)` manual-reset bimetallic snap-disc thermostat on the quench outlet transition (aligned with the `200 F / 93.3 C` software RTD cutoff)\n")
		b.WriteString("   - **Low-flow switch:** Mechanical paddle quench spray flow switch (`< 0.25 GPM`)\n")
		b.WriteString("   - **E-Stop & Exhaust Proving (NFPA 86):** NEMA 4X twist-to-release E-Stop (NC) + optional blower running/airflow proving interlock for gas-fired kilns\n")
		b.WriteString("   - **Power loss or Modbus software trip (`200 F` / `> 5 s` fault):** Keep `bypass_coil_deenergize_to_trip: true` (default) on fail-safe PLC/relay coils so any power loss, RS-485 disconnect, or OS freeze releases the spring-return damper **OPEN**\n")
	}
	b.WriteString("3. **Gravity Backup Tip:** A small 5-gallon elevated emergency water tank with a normally-open solenoid provides 3–5 minutes of gravity quench spray while the bypass damper swings open during a power outage.\n\n")

	// 5. Closed-Loop Sump Recirculation Summary.
	b.WriteString("## 5. Closed-Loop Sump Recirculation & Dump Schedule\n\n")
	fmt.Fprintf(&b, "%s\n\n", r.Summary)
	for _, step := range r.ScheduleSteps {
		fmt.Fprintf(&b, "- %s\n", step)
	}
	b.WriteString("\n")

	// 6. Cold-Water Commissioning & Leak Test Checklist.
	b.WriteString("## 6. Cold-Water Commissioning & Leak Test Checklist\n\n")
	b.WriteString("Complete every box with the kiln **COLD** before your first firing:\n\n")
	b.WriteString("- [ ] **IPC Section 608 Backflow Verification:** Confirm potable water supply is isolated by an engineered **Air Gap (`>= 1\"` vertical separation above break-tank overflow rim, IPC 608.15.1)** OR a tested **ASSE 1013 RPZ backflow preventer** (never a direct hose connection to pressurized chemical piping or a submerged float valve).\n")
	b.WriteString("- [ ] **Hydro & Leak Test:** Fill atmospheric sump tank with clean tap water; run recirculation pump for 15 minutes and verify zero drips at CPVC joints, pump unions, and flange gaskets.\n")
	if a.Eductor.Enabled {
		fmt.Fprintf(&b, "- [ ] **Positive-Pressure Soap-Bubble Test (Eductor Mode):** Run the dry motive fan cold (`+%.2f inWC` in Stage 3) and brush soapy water onto every downstream flange, cleanout hatch, and sump water-seal leg — fix ANY bubbles before hot firing!\n", a.Eductor.PositivePressureInWC)
	}
	fmt.Fprintf(&b, "- [ ] **Nozzle Pattern Verification:** Inspect quench (`%.2f GPM`) and venturi (`%.2f GPM`) sprays through inspection port — confirm uniform fine cone coverage with no plugged orifices.\n",
		d.Quench.TotalSprayWaterGPM, d.Venturi.LiquidFlowGPM)
	if usesVFD {
		b.WriteString("- [ ] **Blower Rotation & VFD Ramp:** Confirm NEC 430.102 lockable sight disconnect operates cleanly; jog VFD at 15 Hz to verify correct wheel rotation direction; ramp smoothly to 60 Hz and confirm vibration-free operation.\n")
	} else if p.BlowerDriveOption != BlowerDriveCompressorEductor {
		b.WriteString("- [ ] **Blower Rotation & Damper Trim:** Confirm NEC 430.102 lockable sight disconnect operates cleanly; start blower with bleed/trim damper partially open and verify smooth, vibration-free airflow.\n")
	}
	fmt.Fprintf(&b, "- [ ] **Draft Sensor Zero & Calibration:** With blower OFF, zero the `+/-2 inWC` draft gauge/transmitter; start blower and confirm smooth tracking at `%.2f inWC` gentle kiln suction.\n",
		d.Input.TargetKilnDraftInWC)
	b.WriteString("- [ ] **Fail-Safe Bypass Trip Drill (Water Loss):** Close the quench water valve until flow drops below `0.25 GPM` — confirm hardwired safety relay latches and spring-return bypass damper snaps **OPEN** within 3 seconds.\n")
	if p.ControlTier == ControlTierElectromechanicalOnly {
		b.WriteString("- [ ] **Fail-Safe Bypass Trip Drill (200 F Hardwired Limit / Power Loss):** Pull 24 VDC control power and test the `200 F` hardwired snap-disc — confirm bypass damper springs **OPEN** immediately and blower contactor drops out.\n")
	} else {
		b.WriteString("- [ ] **Fail-Safe Bypass Trip Drill (200 F Hardwired + Software Limit / Power Loss):** Pull 24 VDC control power (and test both the `200 F` hardwired snap-disc and `200 F` software trip) — confirm bypass damper springs **OPEN** immediately and VFD stops via hardwired STO/Run-Enable.\n")
	}
	fmt.Fprintf(&b, "- [ ] **Water Supply Check:** Verify potable supply (through Air Gap / ASSE 1013 RPZ) delivers at least `%.2f GPM` continuously.\n",
		math.Max(0.5, d.Quench.TotalSprayWaterGPM))

	if p.IsDeluxe() {
		sp := eval.ShopPlan
		b.WriteString("\n## 7. Custom Fabrication Deluxe Build Guide\n\n")
		if sp.DeluxeSpec != nil {
			fmt.Fprintf(&b, "%s\n\n", sp.DeluxeSpec.Summary)
		}
		b.WriteString("### 7.1 TIG Weld Prep, Back-Purging & Hot-Flue Insulation Jacket\n")
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			b.WriteString("- **Alloy C-276 Weld Procedure:** Bevel `11 ga / Sch 10S` edges to 37.5°, degrease with acetone, use **ERNiCrMo-4** filler rod with continuous argon back-purge (`< 50 ppm O2`), and keep interpass temperature strictly **`<= 200 F`** to avoid secondary-phase carbide precipitation.\n")
		} else {
			b.WriteString("- **316L SS Weld Procedure:** Bevel `11 ga / Sch 10S` edges to 37.5°, degrease with acetone, use **ER316L** filler rod with continuous argon back-purge (`< 50 ppm O2`) for sugar-free smooth root welds, keep interpass temperature **`<= 300 F`**, then pickle and citric/nitric passivate every interior weld seam.\n")
		}
		b.WriteString("- **Sight Glasses, Cleanout Doors & Insulation:** TIG-weld the **2× borosilicate sight-glass nozzles** (Stage 1 quench spray + Stage 3 bed level) and **2× bolted NFPA 91 quick-release cleanout frames**, then wrap the dry hot flue upstream of Stage 1 with the **`1\"` ceramic-fiber blanket + stainless steel cladding jacket** (still maintaining NFPA 211 `>= 36\"` clearance to unprotected combustibles).\n\n")

		b.WriteString("### 7.2 ANSI/ASME B16.5 Class 150 Bolt Torque & Gasket Seating\n")
		b.WriteString("- Apply nickel anti-seize to all 316 SS flange studs to prevent galling, seat **PTFE envelope gaskets** concentrically (never use PTFE on the dry `2400 F` kiln inlet — use `>= 2300 F` ceramic fiber there!), and torque in a **3-stage star pattern (30% -> 65% -> 100%)**.\n")
		b.WriteString("- **Torque Limits:** Torque 316L/C-276 metal-to-metal ANSI Class 150 lap-joint flanges to **55–75 ft-lb**, and cap FRP-to-metal stub-end flanges at **25–35 ft-lb** with wide flat washers under backing rings so FRP stub ends never crack.\n\n")

		b.WriteString("### 7.3 Filament-Wound Derakane Vinyl-Ester FRP Curing & Internals\n")
		b.WriteString("- Allow secondary FRP nozzle/gusset-ring laminations to cure **24 hours at >= 70 F** (or post-cure **4 hours at 180 F** for peak wet-HCl resistance) before hydrostatic testing.\n")
		b.WriteString("- Bolt the **12\" packing access manway** and slide-in **removable 3-pass chevron demister cassette** with EPDM/PTFE seals; align the **sloped-bottom baffled settling sump** (`2\"` full-port drain ball valve) for positive flooded suction into the **sealless PVDF magnetic-drive chemical pump**.\n\n")

		b.WriteString("### 7.4 Welded Structural Steel Skid, K-Bracing & Concrete Anchoring\n")
		fmt.Fprintf(&b, "- Position the welded **`2\"×2\"×3/16\"` galvanized/powder-coated steel forklift skid** (`%.1f\" L × %.1f\" W`) on a level `>= 4\"` concrete slab and anchor with **`6× 1/2\"×4\"` 316 SS concrete wedge anchors** (`>= 3\"` embedment, `40 ft-lb` torque).\n", sp.OverallLengthIn, sp.OverallWidthIn)
		b.WriteString("- Bolt/weld the **Stage 3 column cage & diagonal K-bracing**, secure overhead ducts with cushioned pipe saddles (`<= 4 ft` spacing), and install **flexible PTFE/EPDM vibration isolation boots** at both blower inlet and discharge.\n\n")

		b.WriteString("### 7.5 Acoustic Blower Baffle Enclosure & Tuned Stack Silencer Assembly\n")
		b.WriteString("- Line all 6 steel enclosure panels with **`2\"` Rockwool Safe'n'Sound mineral wool** retained behind a **perforated metal inner liner** and mount the blower pedestal on `1\"` ribbed neoprene isolation pads.\n")
		b.WriteString("- **CRITICAL TEFC Motor Cooling Labyrinth:** Install both baffled intake and exhaust cooling-air labyrinth ducts (minimum `2×` motor fan inlet cross-sectional area, offset 180° turns lined with Rockwool) so the TEFC motor receives full cooling airflow without leaking direct line-of-sight noise.\n")
		fmt.Fprintf(&b, "- Mount the **tuned absorptive discharge stack silencer** directly above the discharge vibration boot (`%.1f dBA` bare -> **`%.1f dBA @ 3 ft`** -> **`%.1f dBA @ %.0f ft`** property line, `-%.1f dBA` total reduction).\n\n",
			sp.AcousticSummary.BareDBA3Ft, sp.AcousticSummary.TreatedDBA3Ft, sp.AcousticSummary.PropertyLineDBA, sp.AcousticSummary.PropertyLineFt, sp.AcousticSummary.TotalReductionDBA)

		b.WriteString("### 7.6 UL 508A NEMA 4X Panel Wiring & NEC 110.26 Clearance Verification\n")
		b.WriteString("- Mount the **UL 508A-style NEMA 4X pre-wired panel** with **through-door NEC 430.102 lockable disconnect**, `3%` input line reactor, `HF-GFCI` breaker, GS21 VFD, **FM/UL 1/16-DIN `200 F` independent high-limit controller + hardwired 24VDC latching safety relay**, dual Type-K thermowells, inline flow transmitter, flow-through pH probe, and pilot stack light/alarm horn.\n")
		b.WriteString("- **Segregation & NEC 110.26 Sign-Off:** Route shielded VFD motor cable in dedicated LFMC **`>= 12\"`** away from `24VDC` NFPA 86 safety loop & RS-485 instrumentation conduit, and verify unobstructed **`36\"` depth × `30\"` width × `78\"` height** electrical working clearance (**NEC 110.26(A)**) plus **`24\"`** pump & cleanout door service swing.\n")
	}

	return b.String()
}

func generateLocalCeramicBuildGuide(eval StudioEvaluation) string {
	p := eval.Project.Normalize()
	d := eval.Design
	c := eval.COTS
	u := eval.Urban
	fr := eval.Frugal
	if fr == nil {
		fr = EvaluateFrugal(p, d)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Local Ceramic & Off-Grid Build Guide — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Kiln Profile:** %.1f cu.ft. | **Firing:** %s (%.0f F peak, %.0f hrs; %.1f hr salting) | **Glaze Mode:** `%s` | **Build Path:** `%s`\n",
		d.Input.KilnVolumeCuFt, p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours, p.EffectiveSaltingHours(),
		p.VaporGlazeMode, p.BuildPath)
	fmt.Fprintf(&b, "> **Region:** %s | **Draft Mode:** `%s` (%s) | **Zero Electricity:** `%v` (`%.0f W`) | **Est. Local Cost:** %s | **Alkali:** %s (`%s`)\n\n",
		fr.RegionLabel, fr.DraftMode, fr.DraftModeLabel, fr.ZeroElectricity, fr.PowerWatts, fr.LocalCurrencyEstimate, fr.AlkaliName, fr.AlkaliFormula)

	b.WriteString("## 1. Key Masonry & Gravity Dimensions at a Glance\n\n")
	b.WriteString("| Stage / Component | Studio Size (Imperial) | Engineering Size (Metric) | Notes |\n")
	b.WriteString("|---|---|---|---|\n")
	fmt.Fprintf(&b, "| Inlet Refractory Flue | %.1f\" masonry flue | %.0f mm | NFPA 211 >= 36\" clearance to combustibles |\n",
		math.Max(10.0, d.Envelope.DuctDiameter.Inches()+2.0), math.Max(10.0, d.Envelope.DuctDiameter.Inches()+2.0)*25.4)
	fmt.Fprintf(&b, "| Stage 1: Masonry Quench Chamber | %.1f\" bore | %.0f mm | All-ceramic firebrick (>1,800 F dry-run immune) |\n",
		c.QuenchVessel.NominalIn, c.QuenchVessel.NominalIn*25.4)
	fmt.Fprintf(&b, "| Stage 2: Terracotta Baffles & Gravity Rose | %.2f GPM @ %.2f PSI | %.2f L/min | Gravity-fed from 3.5 m overhead tank |\n",
		fr.SprayFlowGPM, fr.GravityHeadPSI, fr.SprayFlowLPM)
	fmt.Fprintf(&b, "| Stage 3: Wide Stoneware Packed Tower | %.1f\" bore | %.0f mm | v_s=%.2f m/s, %.3f inWC Ergun dP, %.1f%% HCl |\n",
		fr.BedDiameterIn, fr.BedDiameterIn*25.4, fr.BedSuperficialVelMS, fr.BedDPInWC, fr.HClRemovalPct)
	fmt.Fprintf(&b, "| Overhead Gravity Tank | %.0f gal @ %.1f ft | %.0f L @ 3.5 m | Diagonally X-braced stand (~%.0f kg wet load) |\n",
		fr.GravityTankGal, fr.TankElevationFt, fr.GravityTankLiters, fr.GravityTankLiters*1.05+22.0)
	fmt.Fprintf(&b, "| Stage 4: Exhaust Stack / Draft | %.1f ft chimney | %.1f m | Avail %.3f inWC vs %.3f inWC req (feasible=%v) |\n\n",
		fr.ChimneyHeightFt, fr.ChimneyHeightM, fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftFeasible)

	b.WriteString("## 2. Tools & Local Workshop Preparation\n\n")
	b.WriteString("**Build Path: Local Ceramic & Masonry (`local_ceramic_masonry`)** — designed for Nairobi, Mumbai, and local pottery studios using fired firebrick, IS 651 salt-glazed stoneware (SWG) pipes, potter-extruded stoneware Raschig rings, terracotta baffles, an elevated HDPE gravity tank, and manual treadle/hand pumping (or a local 0.5 HP 230V 50Hz domestic pump):\n\n")
	b.WriteString("- Masonry trowel, spirit level, plumb bob, mortar mixing tub, and firebrick / lime-sand refractory mortar\n")
	b.WriteString("- Studio clay extruder (or hollow wooden dowel) + wire cutter to produce `1\"` unglazed grogged stoneware Raschig rings fired in your own kiln to Cone 6–10\n")
	b.WriteString("- Hand saw, drill, wrenches, and M12 through-bolts for the **diagonally X-braced `3.5 m (11.5 ft)` overhead gravity tank stand** (`200–500 kg` water load!)\n")
	b.WriteString("- Basic plumbing cutters & PVC/HDPE pipe fittings for the gravity downcomer, brass/PVC spray rose, foot valve, and drain valve\n")
	b.WriteString("- `$10` brine hydrometer (`SG 1.000–1.200`), litmus pH test strips (`pH 1–14`), and mechanical bimetallic stem thermometer\n\n")

	b.WriteString("## 3. Materials, Corrosion Immunity & Local Chemistry\n\n")
	fmt.Fprintf(&b, "- **All-Ceramic Material Immunity:** %s\n", u.Corrosion.MaterialSummary)
	fmt.Fprintf(&b, "- **Local Alkali Chemistry:** %s\n", u.Corrosion.SodaComparison)
	fmt.Fprintf(&b, "- **Inspection Schedule:** %s\n\n", u.Corrosion.InspectionInterval)

	b.WriteString("## 4. Step-by-Step Masonry & Gravity Assembly\n\n")
	fmt.Fprintf(&b, "1. **NFPA 211 Flue Clearances & Masonry Quench Chamber:** Lay the firebrick flue and quench chamber on a compacted masonry slab with **>= 36\" clearance to unprotected combustibles** (NFPA 211) and an **NFPA 91** brick ash cleanout arch. Seal hot brick joints strictly with `>= 2300 F` fireclay/refractory mortar.\n")
	fmt.Fprintf(&b, "2. **Wide Stoneware Packed Tower (%.0f\" bore):** Construct the wide masonry or IS 651 SWG tower with an open firebrick support arch, fill with `1\"` kiln-fired unglazed stoneware Raschig rings, and set 3 alternating 45° terracotta chevron mist baffles above the gravity spray rose.\n", fr.BedDiameterIn)
	fmt.Fprintf(&b, "3. **Plastered Brick Baffled Sump & Elevated 3.5 m Gravity Tank:** Plaster the 208 L brick settling sump with a central ash-weir baffle; erect the diagonally X-braced `3.5 m` stand for the `%.0f L` overhead HDPE tank and plumb the `1\"` gravity downcomer (`%.2f L/min` @ `%.2f PSI`) with a manual brass ball valve.\n", fr.GravityTankLiters, fr.SprayFlowLPM, fr.GravityHeadPSI)
	fmt.Fprintf(&b, "4. **Outdoor Exhaust Stack & Setbacks:** Terminate the outdoor masonry/SWG chimney at least **%.1f m (%.1f ft)** above grade, guy-wired/wall-tied every `3 m`, keeping hazardous salt-firing HCl exhaust strictly outdoors (`>= 30 ft` from property lines, `>= 10 ft` from operable windows).\n\n", fr.ChimneyHeightM, fr.ChimneyHeightFt)

	b.WriteString("## 5. Cold-Water Commissioning & Smoke Check\n\n")
	b.WriteString("- [ ] **Masonry Smoke & Leak Test:** Burn damp cardboard/newspaper in the cold kiln firebox and verify zero smoke leaks at brick joints before first firing.\n")
	b.WriteString("- [ ] **Overhead Stand Stability Check:** Fill the `3.5 m` overhead gravity tank to capacity and verify zero rack/sway on all 4 diagonally X-braced legs.\n")
	fmt.Fprintf(&b, "- [ ] **Gravity Spray Rose Flow Check:** Open the downcomer valve and verify uniform `%.2f L/min` (`%.2f GPM`) gravity rain over the quench baffles and stoneware rings.\n", fr.SprayFlowLPM, fr.SprayFlowGPM)
	if !fr.ZeroElectricity {
		b.WriteString("- [ ] **Local 230V 50Hz Pump & RCD Check:** Verify the local domestic monoblock pump circuit is protected by a `30 mA` RCD/ELCB and confirm clean closed-loop tank refill.\n")
	}
	b.WriteString("\n## 7. Local Ceramic & Zero-Electricity Build + Manual Firing Guide\n\n")
	fmt.Fprintf(&b, "%s\n\n", fr.Summary)

	b.WriteString("### 7.1 Firing Your Own Stoneware Raschig Rings & Terracotta Mist Baffles\n")
	fmt.Fprintf(&b, "- **Packing Spec:** %s.\n", fr.PackingMaterial)
	b.WriteString("- **Extrusion & Firing:** Wedge coarse grogged stoneware clay (15–20% sand/grog for thermal-shock immunity), extrude `1\" OD × 3/4\" ID` hollow tubes, slice to `1\"` lengths, dry slowly, and fire unglazed to **Cone 6–10** tucked into unused corners of normal studio firings.\n")
	b.WriteString("- **Heavy Ceramic Support Arch:** Stoneware rings weigh `~46 lb/cu.ft.` (`~21 kg/cu.ft.`) — lay an open **firebrick arch grate** (or heavy stoneware kiln shelves with `1/2\"` gas slots) at the bottom of the wide masonry tower so the bed never sags.\n")
	b.WriteString("- **Terracotta Chevron Baffles:** Slab-roll `3/8\"` unglazed terracotta strips and set 3 alternating 45° chevron tiers above the spray rose to knock down mist droplets at `<= 0.020 inWC` pressure drop.\n\n")

	if fr.SteamEnabled {
		b.WriteString("### 7.2 Masonry Mortar Sealing, Steam-Jet Chimney Sizing & Zero-Meltdown Immunity\n")
		fmt.Fprintf(&b, "- **Thermal Immunity:** %s\n", fr.ThermalImmunityNote)
		fmt.Fprintf(&b, "- **Direct Refractory Flue & Compact Steam-Driven Stack:** Because the certified `%.0f psig` steam-jet ejector provides `%.3f inWC` motive draft (total available `%.3f inWC` vs `%.3f inWC` required, margin `%+.3f inWC`, feasible=`%v`), NO tall concentric brick reheat recuperator is required — point every masonry joint tightly with refractory clay/lime-sand mortar and smoke-test cold before firing.\n\n",
			fr.SteamPressurePSIG, fr.EjectorDraftInWC, fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftMarginInWC, fr.DraftFeasible)
	} else if !fr.ZeroElectricity {
		b.WriteString("### 7.2 Masonry Mortar Sealing, Hybrid Draft & Zero-Meltdown Immunity\n")
		fmt.Fprintf(&b, "- **Thermal Immunity:** %s\n", fr.ThermalImmunityNote)
		fmt.Fprintf(&b, "- **Direct Refractory Flue & Local Low-dP Fan:** Local hybrid draft provides `%.3f inWC` available draft vs `%.3f inWC` required (margin `%+.3f inWC`, feasible=`%v`) without requiring a tall concentric thermal-reheat chimney; point every masonry joint tightly and smoke-test cold before firing.\n\n",
			fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftMarginInWC, fr.DraftFeasible)
	} else {
		b.WriteString("### 7.2 Masonry Mortar Sealing, Concentric Flue Reheat & Zero-Meltdown Immunity\n")
		fmt.Fprintf(&b, "- **Thermal Immunity:** %s\n", fr.ThermalImmunityNote)
		fmt.Fprintf(&b, "- **Concentric Brick Flue-Wall Recuperator:** Build a gas-tight `65 mm` split-firebrick partition wall between the `%.0f F` kiln exit flue and the clean chimney riser so waste heat warms scrubbed gas back to **`%.0f F`** (`%.3f inWC` available draft vs `%.3f inWC` required, margin `%+.3f inWC`, feasible=`%v`). Point every masonry joint tightly with refractory clay/lime-sand mortar and smoke-test cold before first firing so raw flue gas never short-circuits past the packed bed.\n\n",
			d.Input.PeakExitTempF, fr.StackReheatTempF, fr.AvailableDraftInWC, fr.RequiredDraftInWC, fr.DraftMarginInWC, fr.DraftFeasible)
	}

	b.WriteString("### 7.3 Elevated Overhead Gravity Tank Stand Bracing (200–500 kg Load)\n")
	fmt.Fprintf(&b, "- Mount the **`%.0f L` (`%.0f gal`)** overhead HDPE gravity tank at **`%.1f ft` (`3.5 m`)** elevation (`%.2f PSI` static gravity head, `%.2f L/min` / `%.2f GPM` downcomer flow).\n",
		fr.GravityTankLiters, fr.GravityTankGal, fr.TankElevationFt, fr.GravityHeadPSI, fr.SprayFlowLPM, fr.SprayFlowGPM)
	fmt.Fprintf(&b, "- **CRITICAL STRUCTURAL BRACING:** Full of water, the tank weighs **`~%.0f kg` (`%.0f lbs`)** — build a 4-post treated hardwood or mild-steel stand on concrete pier footings with **full diagonal X-bracing on all 4 sides** and M12 through-bolts. Never stack loose dry bricks under an overhead water tank!\n\n",
		fr.GravityTankLiters*1.05+22.0, (fr.GravityTankLiters*1.05+22.0)/0.45359237)

	b.WriteString("### 7.4 Chimney Pre-Warming, Manual Pumping Cadence & Local Alkali Dosing\n")
	fmt.Fprintf(&b, "- **Draft & Pumping Mechanism:** %s | **Pump:** %s\n", fr.DraftMechanism, fr.ManualPumpType)
	if fr.ZeroElectricity {
		fmt.Fprintf(&b, "- **Manual Pumping Budget:** Closed-loop recirculation at `%.2f L/min` requires **`%.1f minutes/hour`** on a treadle or cast-iron semi-rotary hand pump (or **`%d × 20L jerrycans/hour`**) during the active salting window.\n",
			fr.SprayFlowLPM, fr.TreadleMinutesPerHr, fr.Jerrycans20LPerHr)
	} else {
		fmt.Fprintf(&b, "- **Local Hybrid Power (`%.0f W`):** Local `0.5 HP` 230V 50Hz domestic monoblock pump automates recirculation; if grid power cuts out mid-firing, the elevated `%.0f L` gravity tank and all-ceramic body continue safely.\n",
			fr.PowerWatts, fr.GravityTankLiters)
	}
	fmt.Fprintf(&b, "- **Local Alkali (`%s` / `%s`):** %s\n", fr.AlkaliName, fr.AlkaliFormula, fr.AlkaliLocalNote)
	if fr.ScalingWarning != "" {
		fmt.Fprintf(&b, "  - *Alkali Caution:* %s\n", fr.ScalingWarning)
	}
	fmt.Fprintf(&b, "- **Honest Capture Performance:** HCl acid gas **`%.1f%%`** | Coarse wood fly ash **`%.1f%%`** | Submicron NaCl fume **`%.0f%%`**. %s\n",
		fr.HClRemovalPct, fr.CoarseAshCapturePct, fr.SubmicronNaClCapturePct, fr.SubmicronHonestNote)
	b.WriteString("- **Manual Firing-Day Sequence:**\n")
	for _, step := range fr.ManualSteps {
		fmt.Fprintf(&b, "  - %s\n", step)
	}
	b.WriteString("\n### 7.5 Critical Safety Warnings (NO DIY Pressurized Steam Boilers!)\n")
	for _, w := range fr.Warnings {
		fmt.Fprintf(&b, "- **SAFETY WARNING:** %s\n", w)
	}

	if fr.SteamEnabled {
		b.WriteString("\n### 7.6 Zero-Electricity Steam-Jet Ejector Startup, PRV Check & Condensation Scrubbing\n\n")
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			fmt.Fprintf(&b, "- **Certified Low-Pressure Steam Source (`6 psig` / `%.3f barg`):** %s\n", fr.SteamPressureBar, fr.SteamSourceType)
			fmt.Fprintf(&b, "- **Simple Drilled Brass Converging Orifice (Unchoked Subsonic `Mach < 1.0`, Zero Shock Screech):** Hand-drilled **`%.2f mm`** brass pipe-cap orifice (NO CNC de Laval bell needed!), exit jet **`%.1f m/s` (`Mach %.2f < 1.0`, `~52 dBA @ 3 ft` via Lighthill `v^8` reduction)**, motive steam **`%.1f kg/hr` (`%.1f L/hr` feedwater)**, entrainment **`%.1f:1`**, ejector draft **`%.3f inWC`** -> shrinks chimney down to **`%.1f m` (`%.1f ft`)** with zero electric blower!\n",
				fr.SteamNozzleThroatMM, fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.SteamFlowKgHr, fr.SteamWaterLHr,
				fr.EjectorEntrainmentRatio, fr.EjectorDraftInWC, fr.ChimneyHeightM, fr.ChimneyHeightFt)
			b.WriteString("- **4.2 m (`13.8 ft` cold / `4.4 m` hot water) Open Water-Column Hydrostatic Relief Standpipe (`S2`):** Physically caps boiler pressure at **`6.0 psig`** with **zero moving parts** — keep standpipe `S2` permanently unvalved and unobstructed, backed up by the factory `10–15 psig` brass PRV & fusible plug.\n")
			fmt.Fprintf(&b, "- **Thermal Balance:** Steam generation load **`%.1f kW`** vs kiln flue waste heat above `400 F` **`%.1f kW`** (`waste_heat_sufficient=%v`).\n",
				fr.SteamThermalKW, fr.KilnWasteHeatKW, fr.WasteHeatSufficient)
			fmt.Fprintf(&b, "- **Condensation Nucleation & Submicron Capture (`%.0f%%` NaCl):** %s\n", fr.SubmicronNaClCapturePct, fr.CondensationGrowthNote)
			b.WriteString("- **Pre-Firing Hydrostatic Standpipe, PRV Pull-Ring & Startup Sequence:**\n")
			b.WriteString("  1. **Verify `4.2 m` (`4.4 m` hot water) Open Hydrostatic Standpipe (`S2`) & Cold PRV Pull-Ring Test:** Confirm standpipe `S2` has zero shutoff valves, clear open top overflow, and clean water prime; pull the manual test ring on the factory `10–15 psig` brass PRV to confirm free snap action.\n")
			b.WriteString("  2. **Verify `1.5× MAWP` (`23 psig`) Hydro-Test, `0–15 psig` Gauge & Feedwater Supply:** Confirm cold hydrostatic proof test record, zeroed low-range glycerin gauge on brass siphon pigtail, working bronze feedwater check valve, and full open atmospheric `0 psig` feed tank + hand/feed pump.\n")
			b.WriteString("  3. **Wet Tower Spray First -> Gradual 6 psig Steam Admission:** Establish overhead gravity spray across the quench baffle & stoneware rings BEFORE warming the catering steamer; smoothly crack the bronze throttling valve on insulated header `S1` to establish steady quiet inward kiln draft.\n")
			b.WriteString("- **Mandatory Subsonic Steamer & Standpipe Safety Checklist (NO Unvented DIY Welded Drum Boilers):**\n")
		} else {
			fmt.Fprintf(&b, "- **Certified Steam Source (`%.0f psig` / `%.2f barg`):** %s\n", fr.SteamPressurePSIG, fr.SteamPressureBar, fr.SteamSourceType)
			fmt.Fprintf(&b, "- **de Laval Steam Nozzle & 1D Ejector Performance:** Choked throat **`%.2f mm`**, exit jet **`%.1f m/s` (`Mach %.2f`)**, motive steam **`%.1f kg/hr` (`%.1f L/hr` feedwater)**, entrainment **`%.1f:1`**, ejector draft **`%.3f inWC`** -> shrinks chimney down to **`%.1f m` (`%.1f ft`)** with zero electric blower!\n",
				fr.SteamNozzleThroatMM, fr.SteamNozzleVelMS, fr.SteamNozzleMach, fr.SteamFlowKgHr, fr.SteamWaterLHr,
				fr.EjectorEntrainmentRatio, fr.EjectorDraftInWC, fr.ChimneyHeightM, fr.ChimneyHeightFt)
			fmt.Fprintf(&b, "- **Thermal Balance:** Steam generation load **`%.1f kW`** vs kiln flue waste heat above `400 F` **`%.1f kW`** (`waste_heat_sufficient=%v`).\n",
				fr.SteamThermalKW, fr.KilnWasteHeatKW, fr.WasteHeatSufficient)
			fmt.Fprintf(&b, "- **Condensation Nucleation & Submicron Capture (`%.0f%%` NaCl):** %s\n", fr.SubmicronNaClCapturePct, fr.CondensationGrowthNote)
			b.WriteString("- **Pre-Firing PRV Pull-Ring & Startup Sequence:**\n")
			b.WriteString("  1. **Cold PRV Pull-Ring Test:** Pull the manual test ring on the stamped ASME/IBR/KEBS spring-loaded PRV to confirm free snap action and unobstructed downward vent pipe.\n")
			b.WriteString("  2. **Verify `1.5× MAWP` Hydro-Test, Gauge & Feedwater Pump:** Confirm cold hydrostatic proof test record, zeroed `0–100 psig` glycerin gauge on siphon pigtail, working bronze feedwater check valve, and manual/positive-displacement boiler feed pump (gravity alone cannot feed a pressurized boiler!).\n")
			b.WriteString("  3. **Wet Tower Spray First -> Gradual Steam Admission:** Establish overhead gravity spray across the quench baffle & stoneware rings BEFORE warming steam pressure; smoothly crack the bronze throttling valve on insulated header `S1` to establish steady inward kiln draft.\n")
			b.WriteString("- **Mandatory Certified-Boiler Safety Checklist (NO DIY Welded Drum Boilers):**\n")
		}
		for _, req := range fr.SteamSafetyRequirements {
			fmt.Fprintf(&b, "  - [ ] %s\n", req)
		}
	}
	return b.String()
}

// GenerateOperationsChecklist produces a friendly firing-day operating, safety,
// closed-loop recirculation, and post-firing disposal checklist customized for the studio project.
func GenerateOperationsChecklist(eval StudioEvaluation) string {
	p := eval.Project.Normalize()
	if p.IsLocalCeramic() {
		return generateLocalCeramicOperationsChecklist(eval)
	}
	d := eval.Design
	u := eval.Urban
	r := eval.Recirculation
	usesVFD := p.BlowerDriveOption == BlowerDriveIndustrialRadialVFD || p.BlowerDriveOption == BlowerDriveCompressorAirAtomize

	var b strings.Builder

	fmt.Fprintf(&b, "# Firing-Day Operations Checklist — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Target Firing:** %s (%.0f F, %.0f hrs; %.1f hr salting) | **Glaze Mode:** `%s` | **Target Kiln Draft:** `%.2f inWC` | **CPVC Hardwired & Software Cutoff:** `200 F`\n\n",
		p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours, p.EffectiveSaltingHours(), p.VaporGlazeMode, d.Input.TargetKilnDraftInWC)

	// 1. Pre-Firing Preparation.
	b.WriteString("## 1. Pre-Firing Preparation (30 Minutes Before Lighting)\n\n")
	if u.Drain.BakingSodaLbsPerFiring > 0 {
		fmt.Fprintf(&b, "- [ ] **Stage Bulk Baking Soda (`NaHCO3`):** Have **%.1f lbs** on hand (**%s**) for this %.1f-hour active salting window.\n",
			u.Drain.BakingSodaLbsPerFiring, u.Drain.BulkPackagingLabel, u.Drain.SaltingHours)
		fmt.Fprintf(&b, "- [ ] **Pre-Charge Alkaline Sump (%.0f gal working fill):** Check sump specific gravity with a brine hydrometer (must be `< %.3f SG` before reuse; dump first if `>= %.3f SG`). Dissolve **%.1f lbs** initial bulk baking soda buffer and verify starting `pH 8.0 - 8.5` with test strips (never start a salt firing on unbuffered tap water!).\n",
			r.SumpVolumeGal, r.DumpSpecificGravity, r.DumpSpecificGravity, r.InitialChargeLbs)
	} else {
		fmt.Fprintf(&b, "- [ ] **Sump Prep (`%s` mode — 0 lbs HCl neutralization):** Verify baffled sump is at ~%.0f gal working level (reusable up to %d firings with ash skimming); optional 1 cup baking soda keeps sump mildly alkaline (`pH ~8.0`).\n",
			p.VaporGlazeMode, r.SumpVolumeGal, r.FiringsBeforeFullDump)
	}
	b.WriteString("- [ ] **Inspect Ash Strainer & Nozzles:** Clean the 50-mesh pump Y-strainer basket and confirm quench + venturi nozzles spray evenly.\n")
	fmt.Fprintf(&b, "- [ ] **Verify Utilities & IPC 608 Backflow Protection:** Confirm `%s` breaker and NEC 430.102 sight disconnect are ON, and potable water supply (`>= %.2f GPM`) feeds through an **IPC 608 Air Gap (`>= 1\"`) or certified ASSE 1013 RPZ backflow preventer**.\n",
		u.Electrical.CircuitLabel, u.Water.RequiredGPM)
	if p.ControlTier == ControlTierElectromechanicalOnly {
		b.WriteString("- [ ] **Mandatory Bypass Trip Test:** Toggle the flow switch (`< 0.25 GPM`) and verify the **hardwired `200 F` snap-disc safety relay** — confirm the spring-return hot-gas bypass damper snaps **OPEN** cleanly, then press manual Reset.\n")
	} else {
		b.WriteString("- [ ] **Mandatory Bypass Trip Test:** Toggle the flow switch (`< 0.25 GPM`) and verify the **hardwired `200 F` snap-disc safety relay** (and aligned `200 F` software cutoff) — confirm the spring-return hot-gas bypass damper snaps **OPEN** cleanly, then press manual Reset.\n")
	}
	if usesVFD {
		b.WriteString("- [ ] **Start Wet Stages BEFORE Lighting Kiln:** Start recirculation pump -> open quench spray -> start blower at low VFD speed (20–25 Hz).\n\n")
	} else {
		b.WriteString("- [ ] **Start Wet Stages BEFORE Lighting Kiln:** Start recirculation pump -> open quench spray -> start draft motive air/blower with trim damper set for gentle inward draft.\n\n")
	}

	// 2. During Firing.
	b.WriteString("## 2. During Firing (Ramp, High-Fire & Glaze Introduction)\n\n")
	fmt.Fprintf(&b, "- [ ] **Hold Gentle Kiln Draft (`%.2f inWC`):** Watch the draft gauge/UI — slight negative suction pulls exhaust cleanly into the scrubber without chilling ware or blowing flame out spyholes.\n",
		d.Input.TargetKilnDraftInWC)
	fmt.Fprintf(&b, "- [ ] **Watch Quench Outlet Temp (`%.0f - 188 F` normal):** Adiabatic saturation sits near **%.1f F** at peak heat. If outlet temp climbs above `192 F`, inspect quench water pressure immediately; both the hardwired snap-disc and software backup latch at **`200 F`** to protect CPVC.\n",
		d.Quench.OutletTemp.Fahrenheit(), d.Quench.OutletTemp.Fahrenheit())
	b.WriteString("- [ ] **Hourly Sump pH & Level Checks:** Dip a pH strip every hour (target **`pH 7.5 - 8.5`**). If pH dips below `7.0`, stir in another 2–3 lbs of bulk baking soda slurry.\n")

	if d.Input.FuelType == "wood_salt" || d.Input.FuelType == "wood" {
		b.WriteString("- [ ] **Wood-Stoking Surge & Cooled Blowdown Tips:** Heavy wood stokes create 20–60 second gas bursts (`1.5x-2.0x` average flow). Feed smaller, drier splits steadily so draft tracks smoothly, and route any hot sump bleed/blowdown (`~184 F`) into a **heat-rated cooling catch tank or cold-water tempering tee (`<= 140 F` per IPC 702.5)** — NEVER discharge hot `184 F` blowdown directly into PVC/ABS building drains!\n")
	}
	if p.VaporGlazeMode == VaporGlazeSalt {
		fmt.Fprintf(&b, "- [ ] **Salt Introduction Pacing (`%.1f lbs/hr` over %.1f hrs):** Introduce damp rock salt in small, evenly spaced burritos/packets (every 10–15 min) while dosing the remaining bulk baking soda steadily so HCl stays within the absorber's 99%% capture envelope.\n",
			d.Input.SaltRateLbsPerHr, u.Drain.SaltingHours)
	} else if p.VaporGlazeMode == VaporGlazeSoda {
		b.WriteString("- [ ] **Soda Spray / Introduction (`0 ppm HCl`):** Spray dissolved soda ash / baking soda solution smoothly in small bursts; watch for normal brief steam surges without any acid fume.\n")
	}
	if p.VaporGlazeMode == VaporGlazeSalt || d.PackedBed.HClOutletPPM > 0 {
		fmt.Fprintf(&b, "- [ ] **Outdoor Stack & Neighbor Check (Salt Mode):** Confirm elevated outdoor stack (~%.0f F, ~%.1f ppm residual HCl — strictly outdoor per OSHA 29 CFR 1910.1000 & IMC 510) disperses cleanly above eaves; property-line noise should remain near **%.1f dBA**.\n\n",
			u.Plume.ExhaustTempF, d.PackedBed.HClOutletPPM, u.Noise.PropertyLineDBA)
	} else {
		fmt.Fprintf(&b, "- [ ] **Neighbor Courtesy Check (`0 ppm HCl`):** Outdoor stack discharge should be odor-free white water vapor (~%.0f F); property-line noise should remain near **%.1f dBA**.\n\n",
			u.Plume.ExhaustTempF, u.Noise.PropertyLineDBA)
	}

	// 3. Emergency / Fault Response.
	b.WriteString("## 3. If the Safety Interlock Trips Mid-Firing\n\n")
	b.WriteString("1. **Do NOT panic or seal the kiln:** The spring-return bypass damper is designed to fail **OPEN**, safely venting hot flue gas up the NFPA 211 auxiliary bypass chimney while protecting the plastic scrubber stages.\n")
	b.WriteString("2. **Check the trip reason:** Most trips are either a kinked supply line / clogged strainer (`QUENCH_WATER_LOSS < 0.25 GPM`) or temporary nozzle fouling (`200 F` hardwired snap-disc / `HIGH_QUENCH_OUTLET_TEMP >= 200 F`).\n")
	b.WriteString("3. **Restore water & cooling first:** Confirm quench spray flow is back above `0.50 GPM` and outlet temp is `< 185 F` before resetting the hardwired safety relay and clicking **Reset Interlock**.\n\n")

	// 4. Post-Firing Cooldown, Recirculation Reuse & Neutralized Drain Disposal.
	b.WriteString("## 4. Post-Firing Cooldown, Sump Reuse & Neutralized Dump Triggers\n\n")
	b.WriteString("- [ ] **Cooldown Run:** Keep quench spray and blower running on low speed until kiln flue temperature drops below **300 F**, then shut down blower and pumps.\n")
	fmt.Fprintf(&b, "- [ ] **Closed-Loop Reuse Check (up to %d firing(s) per fill — saves %.1f%% water):** Test cooled sump with a $12 brine hydrometer (`~%.1f g/L` TDS added per firing). **Dump Triggers:**\n",
		r.FiringsBeforeFullDump, r.WaterSavedPct, r.EndOfFiringTDSGL)
	for _, trig := range r.DumpTriggers {
		fmt.Fprintf(&b, "  - %s\n", trig)
	}
	b.WriteString("- [ ] **MANDATORY Drain Temperature Check (`<= 140 F`, IPC 702.5):** Measure sump liquor temperature with a thermometer before draining — allow full cooldown or dilute with cold water until **`<= 140 F (60 C)`** so hot brine never warps/melts PVC or ABS sanitary piping.\n")
	b.WriteString("- [ ] **Ash Settling & Filtration:** Allow sump solids to settle; decant or pump cooled sump water through a 5-gallon bucket lined with a fine paint/mesh filter bag to trap wood ash and silica sludge (evaluate sludge for hazardous disposal if firing heavy-metal glazes such as Pb, Ba, Cd, Cr, or Co).\n")
	fmt.Fprintf(&b, "- [ ] **pH-Verified Drain Release (`pH %s`, IPC 803.1 & 40 CFR 403.5(b)(2)):** Test cooled clear sump water with pH strips. Federal pretreatment law strictly prohibits **`pH < 5.0`**; if below `6.5`, stir in bulk baking soda until neutralized into **`pH %s`**, then flush down a sanitary drain with fresh dilution water (confirm local POTW brine/chloride rules).\n",
		u.Drain.TargetDrainPH, u.Drain.TargetDrainPH)
	b.WriteString("- [ ] **5-Minute Buffered / Fresh-Water Rinse:** Run clean tap water (or pH 8 buffered rinse in salt mode) through the pump, quench nozzles, and venturi manifold for 5 minutes so dried salt crystals do not seize the mag-drive impeller or clog orifices between firings.\n")
	fmt.Fprintf(&b, "- [ ] **Log Maintenance:** Record firing hours and schedule next corrosion check (**%s**).\n", u.Corrosion.InspectionInterval)

	return b.String()
}

func generateLocalCeramicOperationsChecklist(eval StudioEvaluation) string {
	p := eval.Project.Normalize()
	d := eval.Design
	u := eval.Urban
	r := eval.Recirculation
	fr := eval.Frugal
	if fr == nil {
		fr = EvaluateFrugal(p, d)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Local Ceramic Firing-Day Operations Checklist — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Target Firing:** %s (%.0f F, %.0f hrs; %.1f hr salting) | **Glaze Mode:** `%s` | **Draft Mode:** `%s` | **All-Ceramic Body (`>1,800 F` dry-run immune, zero CPVC meltdown risk)**\n\n",
		p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours, p.EffectiveSaltingHours(), p.VaporGlazeMode, fr.DraftMode)

	b.WriteString("## 1. Pre-Firing Preparation (30 Minutes Before Lighting)\n\n")
	fmt.Fprintf(&b, "- [ ] **Overhead Gravity Tank & Stand Check:** Inspect all diagonal X-braces on the `3.5 m` stand (`~%.0f kg` filled load) and fill the `%.0f L` overhead tank (`%.2f PSI` static head) before lighting.\n",
		fr.GravityTankLiters*1.05+22.0, fr.GravityTankLiters, fr.GravityHeadPSI)
	fmt.Fprintf(&b, "- [ ] **Stage Local Alkali (`%s` / `%s`):** Stage **`%.2f kg` (`%.1f lbs`)** per firing; pre-charge the 208 L brick baffled sump to `pH 8.0 - 8.5` (check with litmus strips) and verify specific gravity `< %.3f SG` with a brine hydrometer.\n",
		fr.AlkaliName, fr.AlkaliFormula, fr.AlkaliKgPerFiring, fr.AlkaliLbsPerFiring, r.DumpSpecificGravity)
	b.WriteString("- [ ] **Inspect Gravity Spray Rose & Foot Valve:** Clear wood-ash grit from the suction foot strainer and verify even gravity rain over the quench baffles and stoneware Raschig rings.\n")
	if fr.SteamEnabled {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			fmt.Fprintf(&b, "- [ ] **Pre-Firing Subsonic Standpipe & PRV Safety Check (`6 psig` / `0.414 barg` / `%.1f L/hr`):** Confirm the `4.2 m` (`4.4 m` hot water) open vertical hydrostatic standpipe (`S2`) is unvalved, unobstructed, and primed (physically caps pressure at `6.0 psig`), pull the factory `10–15 psig` brass PRV test ring, confirm `1.5× MAWP` (`23 psig`) cold hydro-test record, check `0–15 psig` gauge zero & bronze check valve, verify open `0 psig` atmospheric feed tank level, and check insulated header `S1` scald guards.\n",
				fr.SteamWaterLHr)
		} else {
			fmt.Fprintf(&b, "- [ ] **Pre-Firing Certified Steam Safety Check (`%.0f psig` / `%.1f L/hr`):** Pull the certified PRV test ring to verify free snap action, confirm `1.5× MAWP` cold hydro-test record, check pressure gauge zero & bronze feedwater check valve, verify boiler feed pump & open `0 psig` feed tank level, and check insulated header `S1` scald guards.\n",
				fr.SteamPressurePSIG, fr.SteamWaterLHr)
		}
	}
	b.WriteString("\n## 2. During Firing (Draft, Gravity Spray & Alkali Dosing)\n\n")
	if !fr.SteamEnabled && fr.ZeroElectricity {
		fmt.Fprintf(&b, "- [ ] **Chimney Pre-Warming Before Salting:** Establish warm buoyancy draw (`%.1f m` chimney, target `>280 F` stack reheat -> `%.3f inWC` avail vs `%.3f inWC` req) before introducing the first salt charge.\n",
			fr.ChimneyHeightM, fr.AvailableDraftInWC, fr.RequiredDraftInWC)
	} else if fr.SteamEnabled {
		fmt.Fprintf(&b, "- [ ] **Wet Spray First -> Gradual Steam Admission:** Keep gravity spray flowing across the stoneware bed BEFORE opening the steam header `S1` valve (`%.3f inWC` ejector draft on `%.1f m` stack); monitor condensing sump water temperature (`< 120 F` target for maximum steam-nucleation NaCl capture).\n",
			fr.EjectorDraftInWC, fr.ChimneyHeightM)
	} else {
		fmt.Fprintf(&b, "- [ ] **Hybrid Local Fan & Gravity Spray:** Run the 230V 50Hz domestic pump and low-dP draft fan; if power drops out, gravity spray from the `%.0f L` overhead tank continues cooling safely.\n", fr.GravityTankLiters)
	}
	if fr.ZeroElectricity {
		fmt.Fprintf(&b, "- [ ] **Manual Pumping Cadence (0 W Off-Grid):** Maintain `%.2f L/min` gravity spray via **`%.1f min/hr`** on the treadle/hand pump (or **`%d × 20L jerrycans/hr`**) during active salting.\n",
			fr.SprayFlowLPM, fr.TreadleMinutesPerHr, fr.Jerrycans20LPerHr)
	}
	fmt.Fprintf(&b, "- [ ] **Hourly Sump pH & Alkali Checks:** Dip a litmus strip hourly (hold `pH 7.5 - 8.5` with `%s`); pace salt charges gently so outdoor chimney dispersion remains clean.\n\n", fr.AlkaliName)

	b.WriteString("## 3. Safety & Post-Firing Disposal\n\n")
	if fr.SteamEnabled {
		if fr.DraftMode == LocalDraftSteamEjectorSubsonic {
			b.WriteString("- [ ] **Certified-Steamer-Only Rule (Zero Unvented DIY Welded Drum Boilers):** NEVER fabricate an unvented DIY welded drum or scrap-pipe boiler on a kiln flue — operate ONLY a factory-certified `12–15 psig` commercial catering stainless steamer/kettle fitted with BOTH the `4.2 m` (`4.4 m` hot water) open hydrostatic relief standpipe (`S2`) and a stamped `10–15 psig` brass PRV.\n")
		} else if fr.DraftMode == LocalDraftSteamEjectorHP {
			b.WriteString("- [ ] **Certified-Baby-Boiler-Only Rule (`60 psig` — Zero DIY Welded Drum Boilers & NO Monotube Coils):** Operate ONLY a factory-certified non-IBR (`<=25 L`)/KEBS/ASME baby boiler with stamped PRV and fusible plug.\n")
		} else {
			b.WriteString("- [ ] **Certified-Boiler-Only Rule (Zero DIY Welded Drum Boilers):** NEVER fabricate an uncertified DIY welded drum or scrap-pipe boiler on a kiln flue — operate ONLY factory-certified non-IBR (`<=25 L`)/KEBS/ASME baby boilers or code-rated small-bore (`<2 L`) monotube coils with stamped PRV.\n")
		}
	} else {
		b.WriteString("- [ ] **Zero-Steam-Boiler Rule:** NEVER attempt to build a DIY pressurized steam boiler or steam-jet ejector on the kiln flue (lethal boiler explosion hazard).\n")
	}
	b.WriteString("- [ ] **Cooldown, Settling & Neutralized Disposal (`<= 140 F` / `60 C`):** Allow brick sump liquor to cool to `<= 140 F (60 C)`, let wood-ash/lime sludge settle behind the weir baffle, check `pH 6.5 - 8.5` with litmus strips, and flush the gravity spray rose clean after firing.\n")
	fmt.Fprintf(&b, "- [ ] **Log Maintenance:** Record firing hours and schedule next masonry check (**%s**).\n", u.Corrosion.InspectionInterval)

	return b.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
