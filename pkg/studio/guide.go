package studio

import (
	"fmt"
	"math"
	"strings"
)

// GenerateBuildGuide produces a tailored, potter-friendly Markdown fabrication and
// assembly guide adapting to build_path, corrosion_tier, and exact kiln dimensions.
func GenerateBuildGuide(eval StudioEvaluation) string {
	p := eval.Project
	d := eval.Design
	c := eval.COTS
	u := eval.Urban

	var b strings.Builder

	fmt.Fprintf(&b, "# Build & Assembly Guide — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Kiln Profile:** %.1f cu.ft. | **Firing:** %s (%.0f F peak, %.0f hrs) | **Glaze Mode:** `%s` | **Build Path:** `%s` | **Corrosion Tier:** `%s`\n\n",
		d.Input.KilnVolumeCuFt, p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours,
		p.VaporGlazeMode, p.BuildPath, p.CorrosionTier)

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
	fmt.Fprintf(&b, "| Stage 2: Wet Venturi Throat | %.1f\" -> %.2f\" throat -> %.1f\" | %.0f mm -> %.1f mm -> %.0f mm | %.1f m/s throat, d50=%.2f um |\n",
		d.Venturi.InletDiameter.Inches(), d.Venturi.ThroatDiameter.Inches(), d.Venturi.OutletDiameter.Inches(),
		d.Venturi.InletDiameter.Millimeters(), d.Venturi.ThroatDiameter.Millimeters(), d.Venturi.OutletDiameter.Millimeters(),
		d.Venturi.ThroatVelocity, d.Venturi.CutDiameterD50Microns)
	fmt.Fprintf(&b, "| Stage 3: Packed Column | %.1f\" COTS (calc %.1f\") × %.1f ft | %.0f mm × %.0f mm | %.1f%% of flood (limit <=75%%) |\n",
		c.PackedColumn.NominalIn, d.PackedBed.ColumnDiameter.Inches(), d.PackedBed.ColumnHeight.Feet(),
		d.PackedBed.ColumnDiameter.Millimeters(), d.PackedBed.ColumnHeight.Millimeters(),
		c.ActualPercentOfFlood)
	fmt.Fprintf(&b, "| Packing Bed Depth | %.1f\" depth | %.0f mm (%.2f m) | %s |\n",
		d.PackedBed.PackingHeight.Inches(), d.PackedBed.PackingHeight.Millimeters(), float64(d.PackedBed.PackingHeight),
		c.PackingMedia.Name)
	fmt.Fprintf(&b, "| Stage 4: ID Blower & Stack | %.2g HP, %.0f ACFM @ %.1f inWC | Wheel %.0f mm @ %.0f RPM | Min stack height %.0f ft |\n",
		d.Blower.RecommendedMotorHP, d.Blower.FlowACFM, d.Blower.TotalStaticPressureInWC,
		d.Blower.WheelDiameter.Millimeters(), d.Blower.OperatingRPM, u.Plume.MinStackHeightFt)
	fmt.Fprintf(&b, "| Overall Skid Footprint | %.1f ft × %.1f ft × %.1f ft H | %.2f m × %.2f m × %.2f m | Estimated COTS total: **$%.2f** |\n\n",
		d.Envelope.Length.Feet(), d.Envelope.Width.Feet(), d.Envelope.Height.Feet(),
		float64(d.Envelope.Length), float64(d.Envelope.Width), float64(d.Envelope.Height),
		c.EstimatedTotalUSD)

	// 2. Tools & Shop Prep (adapts to build_path).
	b.WriteString("## 2. Tools & Shop Preparation\n\n")
	if p.BuildPath == BuildPathCOTSNoWeld {
		b.WriteString("**Build Path: No-Weld COTS Hardware (`cots_no_weld`)** — designed for a home workshop using bolted flanges, stainless band clamps, and solvent-welded CPVC pipe (no TIG welder needed):\n\n")
		b.WriteString("- Cordless drill/driver with step bits & bi-metal hole saws (`1\"` to `4\"` for bulkhead and nozzle ports)\n")
		b.WriteString("- Socket/wrench set for 316 SS flange bolts & high-torque stainless band clamps\n")
		b.WriteString("- Heavy-duty CPVC primer (ASTM F656) & Schedule 80 CPVC solvent cement (ASTM F493)\n")
		b.WriteString("- PTFE thread tape, high-temp RTV silicone (rated 650 F) for dry flue joints, and expanded PTFE flange gaskets\n")
		b.WriteString("- Basic multimeter, wire strippers, and a digital manometer or Dwyer Magnehelic gauge (`0 to -1.0 inWC`)\n\n")
	} else {
		b.WriteString("**Build Path: Rolled Sheet Metal & DXF Fabrication (`welded_dxf`)** — uses the exact flat patterns in `scrubber.dxf` for laser/plasma cutting and TIG welding:\n\n")
		b.WriteString("- Send `scrubber.dxf` (units: **millimeters**) to a local CNC laser/waterjet shop for Stage 1–2 shells, venturi cones, and bolted flanges\n")
		b.WriteString("- Sheet-metal slip rolls for cylindrical shells and frustum venturi cones\n")
		if p.CorrosionTier == CorrosionTierHeavyC276FRP {
			b.WriteString("- TIG welder with **ERNiCrMo-4 (Alloy C-276)** filler rod + argon back-purge (or castable acid-proof refractory trowel kit) and vinyl-ester FRP layup kit\n")
		} else {
			b.WriteString("- TIG welder with **ER316L** filler rod, argon back-purge for smooth root welds, and stainless pickling/passivation paste\n")
		}
		b.WriteString("- Plastic hot-air extrusion welder or Schedule 80 CPVC primer & solvent cement for Stage 3 absorber internals\n")
		b.WriteString("- Torque wrench, expanded PTFE full-face flange gaskets, and 316 SS hex bolts/nuts\n\n")
	}

	// 3. Corrosion & Materials Note.
	b.WriteString("## 3. Materials & Corrosion Protection\n\n")
	fmt.Fprintf(&b, "- **Selected Tier (`%s`):** %s\n", p.CorrosionTier, u.Corrosion.MaterialSummary)
	fmt.Fprintf(&b, "- **Vapor Glaze Impact:** %s\n", u.Corrosion.SodaComparison)
	fmt.Fprintf(&b, "- **Inspection Schedule:** %s\n\n", u.Corrosion.InspectionInterval)

	// 4. Step-by-Step Assembly.
	b.WriteString("## 4. Step-by-Step Assembly\n\n")

	b.WriteString("### Step 1 — Stage 1: Hot Flue Duct & Adiabatic Spray Quench Tower\n")
	fmt.Fprintf(&b, "1. Mount the **%s** vertically on a level steel or masonry pad with a bottom drain leading into the sump.\n", c.QuenchVessel.Name)
	fmt.Fprintf(&b, "2. Connect the %.1f\" (%.0f mm) hot inlet flue from the kiln bagwall/damper slot to the **top** of the quench tower so flue gas flows **co-currently downward** with the spray mist (gravity assists ash and droplet drainage).\n",
		d.Envelope.DuctDiameter.Inches(), d.Envelope.DuctDiameter.Millimeters())
	fmt.Fprintf(&b, "3. Install **%d × %s** evenly spaced near the top spray header (`%.2f GPM` total spray). Aim cones cleanly down the centerline so droplets atomize before hitting the walls — supply at **60–100 psig** (or twin-fluid air-atomizing) to keep droplet size `<= 200 um`.\n",
		maxInt(d.Quench.NozzleCount, 2), c.SprayNozzle.Name, d.Quench.TotalSprayWaterGPM)
	if p.VaporGlazeMode == VaporGlazeSalt && p.CorrosionTier == CorrosionTierHobby316L {
		b.WriteString("4. **Critical 316L Salt Protection:** Plumb the quench spray manifold from the **alkaline NaHCO3 recirculation pump** (pH 8.0–8.5) rather than unbuffered acidic tap water, with a clean-water backup line for emergency cooling.\n")
	}
	b.WriteString("5. Install the **Type-K thermocouple** at the quench inlet and the **Pt100 RTD** immediately at the quench exit (upstream of any plastic/CPVC transition).\n\n")

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
	fmt.Fprintf(&b, "2. Mount 2 opposed spray injection ports right at the venturi throat entry (`%.2f GPM` liquid flow) so high-velocity `%.0f m/s` gas shears the water into fine droplets that capture submicron ash and salt fume (`d50 = %.2f um`).\n",
		d.Venturi.LiquidFlowGPM, d.Venturi.ThroatVelocity, d.Venturi.CutDiameterD50Microns)
	b.WriteString("3. Bolt the venturi exit into the lower gas inlet plenum of Stage 3 using an expanded PTFE full-face gasket.\n\n")

	b.WriteString("### Step 3 — Stage 3: Packed-Bed Absorber Column, Sump & Demister\n")
	fmt.Fprintf(&b, "1. Position the **%s** directly beneath or beside the **%s** so scrubber liquor drains freely by gravity.\n",
		c.SumpTank.Name, c.PackedColumn.Name)
	b.WriteString("2. **CPVC Solvent-Welding Best Practice:** Chamfer pipe edges, wipe clean, apply purple ASTM F656 primer liberally to both pipe and socket, apply heavy-bodied Schedule 80 CPVC cement, seat with a 1/4-turn twist, hold 30 seconds, and allow **24 hours full cure** before hot-water or firing service.\n")
	fmt.Fprintf(&b, "3. Install the lower polypropylene/FRP packing support grate at least `12\"` above the gas inlet port, pour in **%s** to a bed depth of **%.1f\" (%.2f m)**, and pin the top hold-down grid so gas surges cannot fluidize the rings.\n",
		c.PackingMedia.Name, d.PackedBed.PackingHeight.Inches(), float64(d.PackedBed.PackingHeight))
	fmt.Fprintf(&b, "4. Mount the top liquid distributor nozzle/header (`%.1f GPM` counter-current recirculation from the mag-drive pump via a 50-mesh inline Y-strainer).\n", d.PackedBed.RecirculationFlowGPM)
	fmt.Fprintf(&b, "5. Seat the **3-pass polypropylene chevron mist eliminator** (`%.1f\"` thick) above the spray distributor so clean exhaust enters the blower dry.\n\n", d.PackedBed.DemisterHeight.Inches())

	b.WriteString("### Step 4 — Stage 4: Induced-Draft Blower, Electrical & Acoustic Stack\n")
	fmt.Fprintf(&b, "1. Mount the **%s** on rubber vibration-isolation pads downstream of the column demister (induced-draft suction keeps the entire scrubber under slight negative pressure so fumes cannot leak out).\n", c.BlowerVFD.Name)
	fmt.Fprintf(&b, "2. **Electrical Hookup:** Wire the VFD to a **%s** (~%.1f A). %s\n", u.Electrical.CircuitLabel, u.Electrical.RequiredAmps, u.Electrical.Summary)
	if c.Silencer != nil {
		fmt.Fprintf(&b, "3. Install the **%s** on the blower discharge and enclose the blower housing with foil-faced Rockwool panels (leaving motor cooling vents unobstructed) — predicted property-line sound: **%.1f dBA at %.0f ft**.\n",
			c.Silencer.Name, u.Noise.PropertyLineDBA, u.Noise.PropertyLineFt)
	} else {
		fmt.Fprintf(&b, "3. Connect a vertical discharge stack rising at least **%.0f ft** above grade (predicted sound: **%.1f dBA at %.0f ft**; add an acoustic silencer if firing at night).\n",
			u.Plume.MinStackHeightFt, u.Noise.PropertyLineDBA, u.Noise.PropertyLineFt)
	}
	fmt.Fprintf(&b, "4. Add a `1/2\"` bottom condensate drain loop on the vertical discharge stack (minimum **%.0f ft** tall) returning condensed water vapor to the sump.\n", u.Plume.MinStackHeightFt)
	if c.PlumeDiluter != nil {
		fmt.Fprintf(&b, "5. **Optional Squirrel-Cage Plume Dilution Assembly (`%s`):** Install the **%s** strictly **downstream** of the Stage 4 radial ID blower (never in place of the %.1f inWC radial blower) via a Y/tee mixing plenum topped with the **%.0f\" high-velocity exit cone (`~3,400 FPM`)**. This shortens horizontal winter steam reach from **~%.0f ft bare down to ~%.0f ft** (dissipating before your %.0f ft property line). *Optional Waste-Heat Upgrade:* Mount the squirrel-cage blower on the **cool intake side** blowing ambient air through a sheet-metal shroud around the kiln exterior/hot flue into the mixing tee to recover free kiln heat and shrink the steam cloud even further!\n\n",
			c.PlumeDiluter.SKU, c.PlumeDiluter.Name, d.Blower.TotalStaticPressureInWC, c.PlumeDiluter.NominalIn,
			u.Plume.BarePlumeLengthFt, u.Plume.DilutedPlumeLengthFt, p.PropertyLineFt)
	} else {
		fmt.Fprintf(&b, "5. **Plume Note:** Bare winter steam reach is ~%.0f ft (clean water vapor). If your neighbor's fence is close, enable `include_plume_diluter` for a 120V %.0f CFM squirrel-cage dilution blower + high-velocity exit cone (~%.0f ft diluted reach).\n\n",
			u.Plume.BarePlumeLengthFt, u.Plume.DilutionCFMNeeded, u.Plume.DilutedPlumeLengthFt)
	}

	b.WriteString("### Step 5 — Fail-Safe Bypass Damper & Safety Interlock Rigging\n")
	b.WriteString("1. **NEVER skip this step:** Stage 3 CPVC/HDPE/PP components soften above `200 F`. Install the **spring-return fail-open hot-gas bypass damper** on a tee between the kiln flue and an auxiliary heat-proof bypass chimney.\n")
	b.WriteString("2. Wire the 24 VDC spring-return actuator through **hardwired normally-closed safety relay contacts** driven by:\n")
	b.WriteString("   - Quench outlet RTD high-temp trip (`>= 210 F / 98.9 C`)\n")
	b.WriteString("   - Quench spray water flow switch low-flow trip (`< 0.25 GPM`)\n")
	b.WriteString("   - Power loss or Modbus daemon heartbeat timeout (`> 5 s`)\n")
	b.WriteString("3. **Gravity Backup Tip:** A small 5-gallon elevated emergency water tank with a normally-open solenoid provides 3–5 minutes of gravity quench spray while the bypass damper swings open during a power outage.\n\n")

	// 5. Cold-Water Commissioning & Leak Test Checklist.
	b.WriteString("## 5. Cold-Water Commissioning & Leak Test Checklist\n\n")
	b.WriteString("Complete every box with the kiln **COLD** before your first firing:\n\n")
	b.WriteString("- [ ] **Hydro & Leak Test:** Fill sump tank with clean tap water; run recirculation pump for 15 minutes and verify zero drips at CPVC joints, pump unions, and flange gaskets.\n")
	fmt.Fprintf(&b, "- [ ] **Nozzle Pattern Verification:** Inspect quench (`%.2f GPM`) and venturi (`%.2f GPM`) sprays through inspection port — confirm uniform fine cone coverage with no plugged orifices.\n",
		d.Quench.TotalSprayWaterGPM, d.Venturi.LiquidFlowGPM)
	b.WriteString("- [ ] **Blower Rotation & VFD Ramp:** Jog VFD at 15 Hz to verify correct wheel rotation direction; ramp smoothly to 60 Hz and confirm vibration-free operation.\n")
	fmt.Fprintf(&b, "- [ ] **Draft Sensor Zero & Calibration:** With blower OFF, zero the `+/-2 inWC` draft transmitter; start blower and confirm controller tracks `%.2f inWC` gentle kiln suction.\n",
		d.Input.TargetKilnDraftInWC)
	b.WriteString("- [ ] **Fail-Safe Bypass Trip Drill (Water Loss):** Close the quench water valve until flow drops below `0.25 GPM` — confirm alarm latches and spring-return bypass damper snaps **OPEN** within 3 seconds.\n")
	b.WriteString("- [ ] **Fail-Safe Bypass Trip Drill (High Temp / Power Loss):** Pull 24 VDC control power (or test 210 F trip relay) — confirm bypass damper springs **OPEN** immediately and blower cuts to safe state.\n")
	fmt.Fprintf(&b, "- [ ] **Water Supply Check:** Verify garden hose delivers at least `%.2f GPM` continuously while sump makeup float operates smoothly.\n",
		math.Max(0.5, d.Quench.TotalSprayWaterGPM))

	return b.String()
}

// GenerateOperationsChecklist produces a friendly firing-day operating, safety,
// and post-firing disposal checklist customized for the studio project.
func GenerateOperationsChecklist(eval StudioEvaluation) string {
	p := eval.Project
	d := eval.Design
	u := eval.Urban

	var b strings.Builder

	fmt.Fprintf(&b, "# Firing-Day Operations Checklist — %s\n\n", p.StudioName)
	fmt.Fprintf(&b, "> **Target Firing:** %s (%.0f F, %.0f hrs) | **Glaze Mode:** `%s` | **Target Kiln Draft:** `%.2f inWC` | **CPVC Cutoff:** `210 F`\n\n",
		p.FiringCone, d.Input.PeakExitTempF, p.FiringDurationHours, p.VaporGlazeMode, d.Input.TargetKilnDraftInWC)

	// 1. Pre-Firing Preparation.
	b.WriteString("## 1. Pre-Firing Preparation (30 Minutes Before Lighting)\n\n")
	if u.Drain.BakingSodaLbsPerFiring > 0 {
		fmt.Fprintf(&b, "- [ ] **Stage Baking Soda (`NaHCO3`):** Have **%.1f lbs** on hand (**%d × 4-lb Arm & Hammer boxes**) for this %.0f-hour salt firing.\n",
			u.Drain.BakingSodaLbsPerFiring, u.Drain.BakingSoda4LbBoxes, p.FiringDurationHours)
		b.WriteString("- [ ] **Pre-Charge Alkaline Sump:** Dissolve the first **4-lb box** of baking soda into the fresh sump water and verify starting `pH 8.0 - 8.5` with test strips (never start a salt firing on unbuffered tap water!).\n")
	} else {
		fmt.Fprintf(&b, "- [ ] **Sump Prep (`%s` mode — 0 lbs HCl neutralization):** Fill sump with clean water (~%.0f gal); optional 1 cup baking soda keeps sump mildly alkaline (`pH ~8.0`).\n",
			p.VaporGlazeMode, math.Max(25.0, float64(d.PackedBed.SumpVolume)/0.003785411784))
	}
	b.WriteString("- [ ] **Inspect Ash Strainer & Nozzles:** Clean the 50-mesh pump Y-strainer basket and confirm quench + venturi nozzles spray evenly.\n")
	fmt.Fprintf(&b, "- [ ] **Verify Utilities:** Confirm `%s` breaker is ON and garden hose supply (`>= %.2f GPM`) is connected with float valve active.\n",
		u.Electrical.CircuitLabel, u.Water.RequiredGPM)
	b.WriteString("- [ ] **Mandatory Bypass Trip Test:** Toggle the flow switch (`< 0.25 GPM`) and high-temp test (`210 F`) — verify the spring-return hot-gas bypass damper snaps **OPEN** cleanly, then reset to normal closed position.\n")
	b.WriteString("- [ ] **Start Wet Stages BEFORE Lighting Kiln:** Start recirculation pump -> open quench spray -> start ID blower at low VFD speed (20–25 Hz).\n\n")

	// 2. During Firing.
	b.WriteString("## 2. During Firing (Ramp, High-Fire & Glaze Introduction)\n\n")
	fmt.Fprintf(&b, "- [ ] **Hold Gentle Kiln Draft (`%.2f inWC`):** Watch the draft gauge/UI — slight negative suction pulls exhaust cleanly into the scrubber without chilling ware or blowing flame out spyholes.\n",
		d.Input.TargetKilnDraftInWC)
	fmt.Fprintf(&b, "- [ ] **Watch Quench Outlet Temp (`%.0f - 188 F` normal):** Adiabatic saturation sits near **%.1f F** at peak heat. If outlet temp climbs above `195 F`, inspect quench water pressure immediately; at **`210 F`** the fail-safe bypass opens automatically to protect CPVC.\n",
		d.Quench.OutletTemp.Fahrenheit(), d.Quench.OutletTemp.Fahrenheit())
	b.WriteString("- [ ] **Hourly Sump pH & Level Checks:** Dip a pH strip every hour (target **`pH 7.5 - 8.5`**). If pH dips below `7.0`, stir in another 2–4 lbs of baking soda slurry.\n")

	if d.Input.FuelType == "wood_salt" || d.Input.FuelType == "wood" {
		b.WriteString("- [ ] **Wood-Stoking Surge Tips:** Heavy wood stokes create 20–60 second gas bursts (`1.5x-2.0x` average flow). Feed smaller, drier splits at steady intervals so the VFD PID tracks smoothly without firebox puff-backs, and crack your ash-bleed blowdown (`~0.08 GPM`) during peak stoking to keep fly ash from clogging nozzles.\n")
	}
	if p.VaporGlazeMode == VaporGlazeSalt {
		fmt.Fprintf(&b, "- [ ] **Salt Introduction Pacing (`%.1f lbs/hr`):** Introduce damp rock salt in small, evenly spaced burritos/packets (every 10–15 min) rather than huge dumps so HCl gas stays within the absorber's 99%% capture envelope.\n",
			d.Input.SaltRateLbsPerHr)
	} else if p.VaporGlazeMode == VaporGlazeSoda {
		b.WriteString("- [ ] **Soda Spray / Introduction (`0 ppm HCl`):** Spray dissolved soda ash / baking soda solution smoothly in small bursts; watch for normal brief steam surges without any acid fume.\n")
	}
	fmt.Fprintf(&b, "- [ ] **Neighbor Courtesy Check:** Stack discharge should be odor-free white water vapor (~%.0f F); property-line noise should remain near **%.1f dBA**.\n\n",
		u.Plume.ExhaustTempF, u.Noise.PropertyLineDBA)

	// 3. Emergency / Fault Response.
	b.WriteString("## 3. If the Safety Interlock Trips Mid-Firing\n\n")
	b.WriteString("1. **Do NOT panic or seal the kiln:** The spring-return bypass damper is designed to fail **OPEN**, safely venting hot flue gas up the auxiliary bypass chimney while protecting the plastic scrubber stages.\n")
	b.WriteString("2. **Check the trip reason:** Most trips are either a kinked garden hose / clogged strainer (`QUENCH_WATER_LOSS < 0.25 GPM`) or temporary nozzle fouling (`HIGH_QUENCH_OUTLET_TEMP >= 210 F`).\n")
	b.WriteString("3. **Restore water & cooling first:** Confirm quench spray flow is back above `0.50 GPM` and outlet temp is `< 185 F` before clicking **Reset Interlock**.\n\n")

	// 4. Post-Firing Cooldown & Neutralized Drain Disposal.
	b.WriteString("## 4. Post-Firing Cooldown, Flush & Drain Disposal\n\n")
	b.WriteString("- [ ] **Cooldown Run:** Keep quench spray and blower running on low speed until kiln flue temperature drops below **300 F**, then shut down blower and pumps.\n")
	b.WriteString("- [ ] **Ash Settling & Filtration:** Allow sump solids to settle for 1–2 hours; decant or pump sump water through a 5-gallon bucket lined with a fine paint/mesh filter bag to trap wood ash and silica sludge (dispose of damp ash solids in regular trash).\n")
	fmt.Fprintf(&b, "- [ ] **pH-Verified Drain Release (`pH %s`):** Test clear sump water with pH strips. If below `6.5`, stir in small pinches of baking soda until neutralized into **`pH %s`**, then flush down a sanitary drain with plenty of fresh dilution water (confirm local municipal brine guidelines).\n",
		u.Drain.TargetDrainPH, u.Drain.TargetDrainPH)
	b.WriteString("- [ ] **5-Minute Fresh-Water Rinse:** Run clean tap water through the pump, quench nozzles, and venturi manifold for 5 minutes so dried salt crystals do not seize the mag-drive impeller or clog orifices between firings.\n")
	fmt.Fprintf(&b, "- [ ] **Log Maintenance:** Record firing hours and schedule next corrosion check (**%s**).\n", u.Corrosion.InspectionInterval)

	return b.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
