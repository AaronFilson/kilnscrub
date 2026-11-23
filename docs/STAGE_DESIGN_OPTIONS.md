# Stage Design Options, Parallel Piping, Dry-Fan Eductor & Closed-Loop Recirculation Trade Study

**Purpose:** Practical, first-principles engineering evaluation of the six major studio-builder design optimizations for `kilnscrub`—answering how to cut fabrication cost, build time, physical footprint, reagent consumption, and water waste while staying strictly honest about hydraulic, thermodynamic, and safety trade-offs.

**Reference Baseline:** All quantitative comparisons below are normalized to the validated `kilnscrub` reference studio kiln (`wood_salt_24` preset, post-audit A1–A4 physics in `pkg/scrubber/` and [docs/ENGINEERING_VALIDATION.md](ENGINEERING_VALIDATION.md), post-Phase 6 US code fixes C1–C14 in [docs/CODE_COMPLIANCE_AUDIT.md](CODE_COMPLIANCE_AUDIT.md)):
- **Kiln:** `24 cu.ft.` Cone 11 (`2400°F` / `1588.7 K`), `129.6 kW` peak heat load, `579.2 ACFM` hot flue gas (`30%` excess air).
- **Stage 1 Quench Exit:** `255.9 ACFM` (`0.1208 m³/s`, `0.0974 kg/s`, `ρ = 0.806 kg/m³`) at **`183.9°F`** (`357.5 K`, `55.7% H₂O`); `0.58 GPM` evaporated, `1.16 GPM` spray (`2.0×` overspray).
- **Stage 2 Wet Venturi:** `46.9 mm` (`1.85"`) single throat at `70 m/s`, `L/G = 0.0012` (`2.30 GPM`), `d_50 = 0.214 µm` (NaCl) / `0.199 µm` (wood ash), `ΔP = 18.88 inWC`.
- **Stage 3 Packed Bed:** Calculated `0.300 m` (`11.81"`) → snapped COTS **`14"` Sch 80 CPVC** (`12.500"` true ID, `122.7 in²`), `1"` PP Pall rings (`ε = 0.90`, `F_p = 170 m⁻¹`), `Z = 1.05 m` (`41.3"`), `5.36 GPM` recirc (`7.05 GPM/ft²`), `49.6%` of Eckert flood (snapped; `55.5%` calc), irrigated bed `ΔP = 1.16–1.44 inWC` + chevron demister `0.26 inWC`, `99.0%` HCl capture (`1,048 → 10.0 ppm`).
- **Stage 4 Inline ID Blower:** `256 ACFM @ 21.21 inWC`, `1.38 shaft HP` → **`2.0 HP` 240V 30A** circuit (`GS21-22P0` in NEMA 4X enclosure per NEC 110.28 / 430.122(A)).
- **Baseline Linear Skid Footprint:** `12.9 ft L × 5.4 ft W × 10.7 ft H` (`3.94 m × 1.66 m × 3.25 m`, `70.4 sq.ft.` / `6.54 m²`).

---

## Executive Summary of Phase 7 Design Options

| Option | Core Idea | Capital / Operating Impact (24 cu.ft. Ref.) | Biggest Engineering Win | Honest Trade-Off / Hard Limit |
| :--- | :--- | :--- | :--- | :--- |
| **1. Intermittent Peak Salting** (`salting_hours = 2–4 hr`) | Dose `NaHCO₃` & run full venturi speed only during Cone 10–11 salting window | **-81% `NaHCO₃`** (`9.8 lbs` / `1 × 15 lb` bag vs `52.4 lbs`); **-35–50% electricity** | Slashes chemical cost, TDS buildup (`5–6` firings/dump), and overnight blower noise (`-6 to -9 dBA`) | Must verify `-0.05 inWC` draft authority during preheat turndown and ramp VFD to `100%` *before* first salt throw |
| **2. Dry-Fan Downflow Static Eductor** (`eductor_dry_fan`) | Dry ambient fan pushes air down into vertical static venturi → `90°` sweep elbow → bottom of packed bed | **-$600 to -$1,500** on blower alone; **+$400 to +$700** for larger `18"` column *(net `-$200 to -$900`)* | Zero rotating fan in wet HCl; cools gas `184°F → 138–145°F` (`+40°F` CPVC margin); gravity sweep separates mist | Low jet efficiency (`15–25%`); adds `1.2–1.6×` motive air requiring `18"` bed; **positive pressure (`+2 to +5 inWC`) inside bed → outdoor-only & airtight flanges!** |
| **3. Parallel 4" Pipe Packed-Bed Manifold** (`parallel_manifold`) | Replace single `14"` industrial pipe with `11–12 × 4"` Sch 80 CPVC plumbing pipes | **`~$900–$1,350`** vs **`~$1,150–$1,600`** landed freighted `14"` spool *(saves `$0–$450`)* | Local plumbing-supply stock, zero LTL freight, hand-portable, modular pipe swap | `D/d_p >= 8` forces `1/2"` PP packing (`+$0.6–1.4 inWC` ΔP, pricier media); **demands pressurized orifice/nozzle top distribution** to prevent dry/flooded pipe imbalance |
| **4. Parallel COTS Mini-Venturis** (`parallel_mini_venturis`) | Replace custom-rolled cone with `2–3` parallel 316L sanitary `2" → 1"` / `1.5"` reducers | **`~$280–$390`** vs **`~$450–$850`** rolled/welded cone *(saves `$150–$450` + zero welding)* | Off-the-shelf tri-clamp parts, 30-sec unclamp inspection, **stageable turndown** via blind caps | **Sub-1" (`< 25 mm`) throats plug rapidly with wood ash** (cap at `2–3` throats on `24 cu.ft.`); steeper sanitary diffuser + splitter adds `+1.5–3.0 inWC` ΔP |
| **5. Closed-Loop Baffled Sump Recirculation** (`shared_baffled_sump`) | Recirculate buffered `pH 8.0–8.5` liquor across all 3 wet stages; dump only at TDS/pH trigger | **70–85% water savings** (`~595–670 gal` vs `3,315+ gal`); **5–6 firings per dump** | Cuts water waste & sewer volume drastically; alkaline spray protects 316L SS in Stages 1–2 | Requires `$12` brine hydrometer (`SG >= 1.105` dump trigger), `50-mesh` Y-strainer, ash weir, and **IPC 702.5 `<= 140°F` drain cooldown** |
| **6. U-Turn Skid + Pure Electromechanical Controls** | Fold skid `180°` + replace Raspberry Pi Modbus with hardwired `200°F` snap-disc & manual VFD | **-46% skid length** (`12.9 → 7.0 ft`); **-$900 to -$1,200** controls savings | Fits narrow urban side-yards; zero software/SD/EMI failure modes; retains full NFPA 86 safety | Manual VFD knob adjustment during wood stokes; no remote web dashboard or automated firing CSV logs |

---

## 1. Intermittent Peak-Only Salting Schedules (`salting_hours` vs. Full Firing Duration)

### 1.1 Why Studio Salt Kilns Do Not Emit HCl for 16 Hours
Industrial waste incinerators feed chlorinated waste continuously 24/7; a studio ceramic salt kiln does **not**. A typical 16-hour Cone 10–11 wood-salt or gas-salt firing breaks into three distinct atmospheric phases:
1. **Preheat & Water-Smoking (`Hours 0–6`, ambient → `1100°F`):** Low fuel burn (`20%–40%` of peak flue flow), zero salt, zero HCl, light steam and modest startup smoke.
2. **Body Reduction & Climb to Cone 9 (`Hours 6–12`, `1100°F → 2300°F`):** Moderate-to-high fuel burn (`50%–85%` of peak flow), wood fly ash/soot present, **still `0 ppm` HCl**.
3. **Peak Salting & Soak Window (`Hours 12–15/16`, Cone 10–11 `2350–2400°F`):** The potter introduces `NaCl` (damp rock salt burritos or brine spray) strictly during a **`2–4 hr` window** (`salting_hours = 3.0 hr` typical on `24 cu.ft.`, `4.0 hr` on `48 cu.ft.`), followed by a brief `30–60 min` post-salt soak/cleanup.

Modeling reagent consumption over all `16 hrs` (`FiringDurationHours`) overstates baking soda demand by **4×–8×**.

### 1.2 Quantitative Impact on Bulk `NaHCO₃`, Sump TDS & VFD Turndown

From `pkg/scrubber/packedbed.go`, neutralizing `2.0 lbs/hr` of NaCl (`0.004312 mol/s` HCl at worst-case 100% hydrolysis, `99.04%` capture, `1.15×` alkalinity buffer) consumes **`1.485 kg/hr` (`3.274 lbs/hr`) of `NaHCO₃`** and generates **`0.899 kg/hr` (`1.981 lbs/hr`) of dissolved `NaCl`** plus `0.194 kg/hr` residual buffer (`1.093 kg/hr` total dissolved salts):

| Operating Assumption (`24 cu.ft.`, `2.0 lbs/hr` NaCl) | Active Salting Window | `NaHCO₃` Consumed per Firing | Bulk Packaging Needed (`15 lb` Bag / `50 lb` Sack) | Sump TDS Added per Firing (`40 gal` / `151 L` sump) | Firings Before Full Sump Dump (`150 g/L` / `SG 1.105`) |
| :--- | :---: | :---: | :--- | :---: | :---: |
| **Light Salt Glaze** | `2.0 hrs` | **`6.5 lbs`** | `1 × 15 lb` bag *(covers `2` firings)* / `50 lb` sack (`~7` firings) | `+14.4 g/L` | **`9–10` firings** |
| **Standard Studio Wood-Salt** *(default)* | **`3.0 hrs`** | **`9.8 lbs`** | **`1 × 15 lb` bag** / **`50 lb` sack (`~5` firings)** | **`+21.6 g/L`** | **`5–6` firings** |
| **Heavy Orange-Peel Salt / Co-op** | `4.0 hrs` | **`13.1 lbs`** | `1 × 15 lb` bag / `50 lb` sack (`~3–4` firings) | `+28.9 g/L` | **`4–5` firings** |
| *Old Continuous-Firing Assumption* | *`16.0 hrs`* | *`52.4 lbs`* | *`4 × 15 lb` bags / `>1 × 50 lb` sack per firing* | *`+115.4 g/L`* | *`1` firing (dump every time)* |

> **Bulk Sourcing Note:** Commercial food/feed/pool-grade sodium bicarbonate (`NaHCO₃`) sells in **`15 lb` resealable bulk bags** (`~$12–$18`, warehouse clubs/pool supply) and **`50 lb` industrial kraft sacks** (`~$28–$42`, agricultural feed stores, ceramic suppliers, or Univar/Brenntag)—cutting reagent cost from `~$55–$70/firing` down to **`~$6–$10/firing`**.

### 1.3 VFD Turndown Power & Nighttime Noise Savings
Because fan shaft power obeys the affinity cube law ($P \propto N^3$) and aerodynamic broadband noise scales approximately as $\Delta L_p \approx 50 \log_{10}(N_2 / N_1)$:
- **Preheat & Early Ramp (`40%–50%` VFD speed, `24–30 Hz`):** Blower shaft power drops from `1.38 HP` to **`0.09–0.17 HP`** (**`-88%` to `-93%`** electricity!), and aerodynamic fan noise falls by **`15–20 dBA`**—turning an overnight `43 dBA` silenced scrubber into a virtually inaudible `< 30 dBA` hum.
- **Body Reduction (`60%–75%` VFD speed, `36–45 Hz`):** Shaft power drops to **`0.30–0.58 HP`** (**`-58%` to `-78%`** power) and noise drops **`6–11 dBA`**, while keeping enough velocity (`~45–53 m/s`) to knock down coarse wood ash.
- **Peak Salting Window (`100%` VFD speed, `60 Hz`, `2–4 hrs`):** Full `70 m/s` throat velocity (`21.2 inWC`, `1.38 shaft HP`) runs **only while submicron NaCl fume and HCl are actively generated**.
- **Net 16-hr Electricity Cut:** Running `6 hrs @ 45%`, `6 hrs @ 65%`, and `4 hrs @ 100%` reduces total blower energy from `~22.1 kWh` down to **`~9.6 kWh`** (**`-57%` power**), while keeping the kiln cleanly under `-0.05 inWC` draft throughout.

> **Operational Rule:** Always ramp the VFD back to `95%–100%` (`v_t >= 65–70 m/s`) **at least 5 minutes before** throwing the first salt packet and confirm sump `pH 8.0–8.5`, so submicron NaCl fume never slips past an under-speed venturi throat.

---

## 2. Dry-Air Driven Downflow Static Eductor Alignment (`eductor_dry_fan`)

### 2.1 Proposed Mechanical Layout
Instead of placing an expensive corrosion-resistant induced-draft (ID) blower at the wet, acidic tail end of Stage 3, the **Dry-Air Driven Downflow Static Eductor** moves the fan into **100% clean, dry ambient air** between Stage 1 and Stage 3:

```mermaid
flowchart TD
    Kiln["Kiln Flue (2400°F, 579 ACFM)"] --> Quench["Stage 1: Vertical Adiabatic Quench (Exits 183.9°F, 256 ACFM)"]
    Quench -->|"Out & Down to Flanged Tee"| Tee["Mixing Tee (Suction Port: -0.6 to -1.0 inWC)"]
    DryFan["DRY Ambient Radial Blower (Carbon Steel / Aluminum — Clean 68°F Air, 307–409 CFM @ 14–28 inWC)"] -->|"Pushes Dry Motive Jet Straight Down"| Tee
    Tee --> Eductor["Vertical Downward-Pointing Static Venturi Eductor (Jet Entrainment + Water Shear)"]
    Eductor --> Sweep["90° Long-Radius Bottom Sweep Elbow (Gravity + Centrifugal Droplet Separation)"]
    Sweep -->|"Condensed Mist & Spray Drain"| Sump["Shared Baffled Alkaline Sump (Atmospheric Vented)"]
    Sweep -->|"Cooled Mixed Gas (138–145°F, 586–694 ACFM, +2 to +5 inWC)"| Bed["Stage 3: Enlarged Packed Bed (18\" ID, <=70% Flood) + Chevron Demister"]
    Bed --> Stack["Outdoor Vertical Exhaust Stack (NO Wet Stage 4 Blower!)"]
```

### 2.2 Momentum, Entrainment & Psychrometric Engineering Analysis

#### A. Genuine Engineering Advantages
1. **Zero Rotating Parts in Wet Acid Gas (`$600–$1,500` Blower Savings):**
   - A `2.0–3.0 HP` acid-duty 316L SS or Derakane FRP radial pressure blower costs **`$1,340–$2,490`** (`pkg/studio/cots.go:417–430`) and carries finite chloride stress-corrosion cracking (Cl-SCC) risk on a 316L wheel spinning at `3,500 RPM` in wet `10–18 ppm` HCl.
   - A **clean-air cast-aluminum or carbon-steel radial blower** (e.g., Dayton/Cincinnati Fan/AMETEK) handles only dry `68°F` outdoor air, lasts indefinitely without acid pitting, and costs **`$550–$950`**—saving **`$600–$1,500`** on the prime mover alone.
2. **Psychrometric Cooling Gives CPVC a Wide Thermal Margin (`183.9°F → 138–145°F`):**
   - Mixing `255.9 ACFM` of saturated `183.9°F` (`55.7% H₂O`) quenched gas with `1.2×–1.6×` (`307–409 ACFM`) dry `68°F` (`50% RH`) ambient motive air lowers the equilibrium mixed gas temperature (accounting for latent heat of minor `0.03–0.04 GPM` condensation) to **`138°F–145°F`**:
     - At `1.2×` motive air (`307 CFM`): **`T_mix = 145.2°F`** (`54.8°F` below CPVC's `200°F` ASTM F441 continuous limit!)
     - At `1.4×` motive air (`358 CFM`): **`T_mix = 141.4°F`** (`58.6°F` margin!)
     - At `1.6×` motive air (`409 CFM`): **`T_mix = 138.1°F`** (`61.9°F` margin!)
   - Compared to the baseline's tight `16.1°F` margin (`183.9°F` vs `200°F`), this dramatically relaxes thermoplastic creep/softening risk in Stage 3 (though the independent hardwired `200°F` NFPA 86 snap-disc + fail-open bypass damper remain mandatory if quench water fails).
3. **Gravity-Assisted Downflow + `90°` Sweep Elbow Droplet Separation:**
   - Pointing the static venturi **vertically downward** aligns droplet momentum and gravity (`g`) instead of fighting gravity in an upward diffuser, eliminating liquid pool-back in the throat.
   - Turning the high-velocity two-phase exit jet through a **long-radius `90°` sweep elbow** into the bottom plenum acts as a passive **cyclonic/inertial knockout**: droplets `> 50–100 µm` fling onto the outer elbow wall and drain cleanly into the shared sump below, cutting carryover load on the Stage 3 packing support grate.
4. **No Vacuum Implosion Stress on Stage 3:**
   - Because Stage 3 operates at slight positive pressure (`+2 to +5 inWC`) instead of `-19 to -21 inWC` suction, column walls are in mild hoop tension rather than external vacuum buckling.

#### B. Honest Engineering Trade-Offs (Why Eductors Are Not a Free Lunch)

| Performance Parameter (`24 cu.ft.` Ref.) | Baseline Inline ID Blower | Eductor — Low-ΔP Duty *(Gas/Soda, No High-Energy Throat Spray)* | Eductor — High-Energy Salt Duty *(Submicron Throat Shear + Bed)* |
| :--- | :---: | :---: | :---: |
| **Quenched Flue Gas Flow** | `256 ACFM @ 183.9°F` | `256 ACFM @ 183.9°F` | `256 ACFM @ 183.9°F` |
| **Dry Motive Airflow Added** | `0 CFM` | `307–358 CFM` (`1.2–1.4×`) | `358–409 CFM` (`1.4–1.6×`) |
| **Total Mixed Gas Entering Stage 3** | **`256 ACFM`** | **`586–640 ACFM`** (`2.3–2.5×`) | **`640–694 ACFM`** (`2.5–2.7×`) |
| **Mixed Gas Temp Entering Stage 3** | `183.9°F` (`16°F` to CPVC limit) | **`141–145°F`** (`55–59°F` margin) | **`138–141°F`** (`59–62°F` margin) |
| **Stage 3 Column Bore Needed (`<= 75%` Flood)** | `14"` Sch 80 CPVC (`49.6%` flood) | **`18"` FRP** (`59–65%` flood; `14"` hits `122–134%`!) | **`18"`–`20"` FRP** (`65–70%` flood; `14"` hits `134–146%`!) |
| **Jet-Eductor Mechanical Efficiency (`η_ej`)** | `N/A` (direct fan `62%`) | **`~18–25%`** (turbulent Borda–Carnot mixing loss) | **`~15–20%`** (mixing + droplet drag) |
| **Prime-Mover Duty & Motor HP** | `256 CFM @ 21.2"` (`1.38 BHP` → **`2.0 HP`**) | `310–360 CFM @ 14–18"` (`1.1–1.6 BHP` → **`1.5–2.0 HP`**) | `360–410 CFM @ 22–28"` (`2.2–2.9 BHP` → **`3.0 HP`**, `240V 40A`) |
| **Blower Hardware Cost** | `$1,340` (316L) / `$1,943` (FRP) | **`$600–$850`** (dry Al/steel) | **`$750–$1,050`** (dry high-head Al/steel) |
| **Stage 3 Column Hardware Cost** | `$665` (`14"` Sch 80 CPVC) | **`$1,120`** (`18"` FRP, `+$455`) | **`$1,120–$1,360`** (`18"` FRP, `+$455–$695`) |
| **Packed-Bed Gauge Pressure** | **`-19 to -21 inWC`** *(vacuum — leaks inward)* | **`+2.0 to +3.5 inWC`** *(positive!)* | **`+3.0 to +5.0 inWC`** *(positive!)* |

1. **Low Jet-Eductor Mechanical Efficiency (`~15%–25%`):**
   - Accelerating stagnant flue gas via turbulent shear from a high-speed air jet inherently dissipates `75%–85%` of the motive jet kinetic energy as irreversible mixing entropy ($\eta_{\text{ej}} = \frac{Q_s \Delta P_{\text{lift}}}{Q_m \Delta P_m} \approx 0.15\text{–}0.25$).
   - Consequently, for **high-energy wood-salt duty** (needing both submicron droplet shear and `+3 to +5 inWC` packed-bed push), the dry motive fan must deliver `360–410 CFM @ 22–28 inWC` (`2.2–2.9 shaft HP` → `3.0 HP` on a **`240V 40A`** circuit per NEC 430.122(A)), using **more electricity** than the `2.0 HP` inline ID blower.
   - For **gas-soda / lower-ΔP scrubbing** (no `19 inWC` submicron throat), the eductor shines: a `1.5–2.0 HP` dry blower easily pulls `-0.7 inWC` kiln/quench suction and pushes `+3 inWC` through the packed bed while keeping the fan 100% dry.
2. **Enlarged Packed Column (`18"` vs. `14"`):**
   - Because `1.2×–1.6×` ambient air expands when warmed from `68°F` to `141°F`, total gas volume entering Stage 3 jumps to **`586–694 ACFM`** (`2.3×–2.7×` quenched ACFM). A `14"` Sch 80 CPVC column (`12.50"` ID) would run at **`122%–146%` of Eckert flooding** and blow brine straight out the stack! Staying `<= 75%` of flood mandates stepping up to an **`18"` ID column** (`59%–70%` flood, `+$455–$695`), which offsets roughly half of the blower hardware savings.
3. **CRITICAL SAFETY CONSTRAINT — Positive Pressure (`+2 to +5 inWC`) Inside Stage 3:**
   - Downstream of the eductor throat, the `90°` sweep elbow, lower plenum, packed column, and sump seal sit at **`+2.0 to +5.0 inWC` positive gauge pressure**.
   - Upstream of the packing bed, the gas still carries **`~400–600 ppm` HCl** (diluted from `1,048 ppm` by motive air — still **80×–120× the `5 ppm` OSHA 29 CFR 1910.1000 ceiling PEL** and **8×–12× the `50 ppm` NIOSH IDLH**!).
   - **Mandatory Design Rules for `eductor_dry_fan`:**
     - **Outdoor-Only Siting:** NEVER install a positive-pressure eductor scrubber indoors, in an enclosed garage, or under an attached carport roof overhang.
     - **Airtight Bolted Flanges & Sealed Liquid Loop:** Use full-face EPDM/PTFE bolted flanges (no slip-fit duct tape or loose press-fit lids!), add a **`>= 8"` deep U-trap / liquid seal** on the bottom sweep-elbow sump drain so `+5 inWC` gas cannot blow out through the vented sump drum, and pass a **positive-pressure soap-bubble leak test** at `+6 inWC` during cold commissioning!
     - **Draft Authority Check:** Keep the kiln-to-Tee run short and low-loss (`<= 0.6 inWC`) so the eductor reliably holds `-0.05 inWC` suction at the kiln bagwall during wood-stoking surges.

---

## 3. Parallel `2"`–`4"` Pipe Packed-Bed Manifolds (`parallel_manifold`)

### 3.1 Exact Cross-Sectional Area Equivalence & The `D/d_p >= 8` Wall-Effect Constraint
Large industrial CPVC pipe (`12"`–`16"` Sch 80) is special-order, heavy, and costly to freight. Potters naturally ask: *"Why not bundle cheap, locally stocked `2"`–`4"` Sch 80 CPVC pipes in parallel?"*

Matching the cross-sectional area of the validated `24 cu.ft.` Stage 3 column (`11.81"` calculated bore = `109.6 in²`; snapped COTS **`14"` Sch 80 CPVC** `12.500"` true ID = **`122.72 in²`**) using ASTM F441 Schedule 80 CPVC pipe:

| Pipe Option (Sch 80 CPVC) | True Bore ID (`in`) | Area per Pipe (`in²`) | Pipes Needed to Match `14"` Single Column (`122.7 in²`) | `D/d_p` with `1"` Pall Rings *(Min `>= 8`)* | Required Packing Size (`D/d_p >= 7.5–8`) | Practical Verdict for `24 cu.ft.` Kiln |
| :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **Single `14"` Sch 80 CPVC** *(Baseline)* | `12.500"` | `122.72` | `1` | **`12.5`** *(PASS)* | `1"` PP Pall rings (`F_p = 170 m⁻¹`) | Lowest ΔP (`1.16 inWC`), simplest spray distributor |
| **Parallel `4"` Sch 80 CPVC** *(Sweet Spot)* | `3.826"` | `11.50` | **`11` (`126.5 in²`) – `12` (`138.0 in²`)** *(10 = `115 in²` vs calc)* | `3.83` *(FAIL — severe channeling)* | **`1/2"` (`12.7 mm`) PP Pall rings** (`D/d_p = 7.65`, `F_p ≈ 312 m⁻¹`) | **Viable DIY parallel manifold**; locally stocked pipe, modular |
| **Parallel `3"` Sch 80 CPVC** | `2.900"` | `6.61` | **`19` pipes** (`125.5 in²`) | `2.90` *(FAIL)* | `3/8"` packing (`D/d_p = 7.73`) | Too many fittings for `24 cu.ft.` *(viable only on `6–10 cu.ft.` kilns using `4–6` pipes)* |
| **Parallel `2"` Sch 80 CPVC** | `1.939"` | `2.95` | **`42` pipes** (`123.9 in²`) | `1.94` *(FAIL)* | `1/4"` saddles (`F_p > 600 m⁻¹`) | **Impractical:** extreme fitting cost, high ΔP, ash plugging |

### 3.2 Hydraulic Penalty of Stepping Down to `1/2"` Packing
Because `4"` Sch 80 pipe (`3.826"` ID) violates the Strigle/Perry wall-channeling rule (`D/d_p >= 8`) with `1"` rings (`3.83`), switching to **`1/2"` (`12.7 mm`) polypropylene Pall rings or saddles** (`ε ≈ 0.87`, `F_p ≈ 312 m⁻¹` / `95 ft⁻¹`) changes column hydraulics directly under the `pkg/scrubber/packedbed.go` equations:
- **Eckert Flooding Velocity Drops:** `v_flood` decreases from `3.08 m/s` (`1"` rings) to **`2.27 m/s`** (`1/2"` rings).
  - With `10 × 4"` pipes: `v_s = 1.63 m/s` → `71.7%` of flood (tight against `75%` cap).
  - With **`11 × 4"` pipes**: `v_s = 1.48 m/s` → **`65.1%` of flood** (optimal target!).
  - With **`12 × 4"` pipes**: `v_s = 1.36 m/s` → **`59.7%` of flood** (generous surge margin!).
- **Mass Transfer Improves Slightly, But Irrigated Ergun ΔP Roughly Doubles:**
  - Smaller `1/2"` rings increase specific surface area (`a_p ≈ 380–414 m²/m³`) and cut `HTU` by `~21%` (`(0.5/1.0)^0.35 = 0.785`), shortening required packing depth for `99%` HCl capture from `1.05 m` (`41"`) to **`0.85 m` (`33.5"`)**.
  - Even at the shorter `0.85 m` depth, the smaller Sauter diameter (`d_p,eq ≈ 1.9 mm`) raises irrigated bed pressure drop from **`1.16 inWC`** (single `14"` column) to **`2.03 inWC`** (`11 × 4"` pipes) or **`1.72 inWC`** (`12 × 4"` pipes) — or `2.13–2.51 inWC` if kept at full `1.05 m` depth. On a `21.2 inWC` wood-salt venturi train, `+0.6 to +1.0 inWC` is a modest `+4%` blower load; on a `3.5 inWC` venturi-less soda kiln, it is a `~25%` ΔP jump.

### 3.3 Honest COTS Pricing Comparison (`4"` Plumbing Pipe vs. `14"` Industrial Column)

| Cost Component (`24 cu.ft.` Stage 3 Column Shell + Packing) | Single `14"` Sch 80 CPVC Industrial Column | Parallel `12 × 4"` Sch 80 CPVC Pipe Manifold |
| :--- | :---: | :---: |
| **CPVC Pipe Shell** | `$665` (`8 ft` of `14"` Sch 80 industrial pipe) | **`$410`** (`54 ft` of `4"` Sch 80 stock pipe @ `$7.60/ft`) |
| **End Flanges / Collector Manifolds / Bulkheads** | `$450` (`2 × 14"` Sch 80 Van Stone flanges + blind plate) | `$520` (`24 × 4"` CPVC bulkhead/flange adapters + top/bottom collector plenums) |
| **LTL Freight Shipping** | `$200` *(oversize commercial freight)* | **`$0`** *(local plumbing counter / car pickup)* |
| **Packing Media (`99%` HCl Bed Volume)** | `$144` (`3 cu.ft.` of `1"` PP Pall rings @ `$48/cu.ft.`) | `$255` (`3 cu.ft.` of `1/2"` PP Pall rings @ `$85/cu.ft.`) |
| **Total Landed Column + Media Cost** | **`~$1,459`** | **`~$1,185`** *(saves `~$275`; range `$0–$450`)* |

**Takeaway:** Parallel `4"` CPVC pipes **do not** make the column 5× cheaper once you count `24` manifold couplings and pricier `1/2"` packing, **but** they save `$200–$450` in freight/industrial flanges, source entirely from local plumbing counters, and let one person carry and assemble every piece by hand.

### 3.4 Critical Liquid-Distribution Manifold Engineering (Preventing Dry vs. Flooded Pipes)
Parallel packed columns suffer from a classic positive-feedback instability if liquid is fed passively:
- If Pipe #3 gets `30%` less liquid than Pipe #7, Pipe #3's irrigated ΔP drops → **more gas rushes up dry Pipe #3** (causing HCl breakthrough) while Pipe #7 holds up extra liquid and **floods**!
- **Mandatory Engineering Solution (Choose One):**
  1. **Pressurized Calibrated-Orifice / Individual Nozzle Header (Recommended):** Run a `1"` CPVC ring/ladder header at **`15–25 psig`** feeding **one dedicated `0.45–0.50 GPM` `1/8"`/`1/4"` NPT full-cone nozzle (or precision `3.2 mm` drilled orifice)** centered over the top of each of the `11–12` pipes (`5.4 GPM` total). Because `20 psig` (`554 inWC`) across each nozzle is **`> 200×` larger** than gas pressure differences between pipes, every pipe is hydraulically locked to identical liquid flow (`±3%`). Always protect the header with a **`50-mesh` inline Y-strainer**.
  2. **Leveled V-Notch Overflow Weir Box:** Mount all `12` pipe tops flush inside a sealed, precision-leveled (`<= 1/16"` across span) top trough with identical `60°` V-notch cut-outs feeding each pipe—no small orifices to plug, at the cost of careful leveling.
- **Gas Plenum Symmetry:** Size the bottom gas-distribution plenum so cross-flow velocity is **`<= 4 m/s`** (`>= 10"–12"` tall chamber beneath the `12` pipe inlets) so static pressure is uniform across every pipe entry.

---

## 4. Parallel Smaller Wet Venturis (`parallel_mini_venturis`)

### 4.1 Area-Equivalent Sanitary Reducer Sizing
In the `24 cu.ft.` baseline, Stage 2 requires a single `46.9 mm` (`1.85"`) throat (`1,725 mm²` / `2.68 in²`) running at `70 m/s` (`256 ACFM`). Rolling and TIG-welding custom `25°` converging and `7°` diverging 316L stainless cones from `scrubber.dxf` is the most labor-intensive metalwork step in the entire scrubber.

Splitting `A_t = 1,725 mm²` into **`N` parallel throats** at equal `70 m/s` scales each throat diameter as $D_{t,N} = D_{t,1} / \sqrt{N}$:

| Venturi Layout (`24 cu.ft.`, `256 ACFM @ 70 m/s`) | Throat Count (`N`) | Exact Throat Bore per Unit | Closest COTS 316L Sanitary Tri-Clamp Reducer | Estimated Fab Cost (All Throats + Tees) | Wood-Ash Plugging Risk | Turndown Staging Capability |
| :--- | :---: | :---: | :--- | :---: | :---: | :--- |
| **Single Rolled Cone** *(Baseline)* | `1` | `46.9 mm` (`1.85"`) | Custom rolled (`scrubber.dxf`) or `6" → 2"` pipe | `$450–$850` | **Low** | None (fixed throat) |
| **2 Parallel Mini-Venturis** *(Wood-Safe Sweet Spot)* | **`2`** | **`33.1 mm` (`1.30"`)** | **`2" → 1.5"` sanitary tri-clamp** (`1.37"` ID, `v_t ≈ 63 m/s`) or `1.25"` Sch 40 | **`$260–$360`** | **Low–Med** (`> 1.0"`) | **2-Step** (`50%` / `100%` flow at full `v_t`!) |
| **3 Parallel Mini-Venturis** *(Gas-Salt Sweet Spot)* | **`3`** | **`27.1 mm` (`1.07"`)** | **`2" → 1"` Sch 40 / sanitary concentric** (`1.05"` ID, `v_t ≈ 72 m/s`) | **`$300–$420`** | **Moderate** *(borderline for heavy wood ash)* | **3-Step** (`33%` / `67%` / `100%`) |
| **4 Parallel Mini-Venturis** | `4` | **`23.4 mm` (`0.92"`)** | `2" → 1"` tube (`0.87"` ID) | `$380–$520` | **HIGH (Not Recommended for Wood!)** | 4-Step *(excessive plugging)* |

### 4.2 Pros vs. Honest Trade-Offs
- **No-Weld Speed & Cost Savings:** Off-the-shelf 316L stainless brewery/dairy **tri-clamp concentric reducers (`2" → 1.5"` or `2" → 1"`)** cost `$18–$35` each. Bolting/clamping `2` or `3` parallel spools between `6"`→`2"` manifold headers cuts Stage 2 cost by **`$150–$450`**, eliminates TIG welding and slip-rolling entirely, and lets a potter unclamp any corroded throat in 30 seconds.
- **Stageable Turndown (`d_50` Preservation):** In a single fixed venturi, turning a kiln down to `50%` firing flow drops throat velocity from `70 m/s` to `35 m/s`, collapsing Calvert NaCl fume capture (`d_50` degrades from `0.21 µm` to `> 0.50 µm`). With `2–3` parallel tri-clamp throats, installing a `$6` sanitary blind cap (or quarter-turn isolation butterfly valve) on 1–2 throats during smaller test firings or lower-flow phases keeps the active throat(s) locked at **`65–70 m/s` (`d_50 ≈ 0.21 µm`)**!
- **Honest Trade-Off #1 — Sub-1" (`< 25 mm`) Wood-Ash Plugging Risk:** In wood-fired kilns, wet char fragments and sodium-silicate scale bridge quickly across throats smaller than **`1.0"` (`25 mm`)**—especially where liquid injection ports intrude into the bore. Keep throat ID **`>= 1.05"` (`27 mm`)** for wood kilns (`N <= 2–3` on `24 cu.ft.`; `N = 1` single `1.3"` throat on `12 cu.ft.`), and use tangential/annular wetted-wall injection instead of protruding center-pipe lances.
- **Honest Trade-Off #2 — Steeper Diffuser & Splitter Plenum ΔP Penalty:** Standard sanitary concentric reducers diverge at `14°–22°` included angle rather than the optimal `7°` Pease–Anthony diffuser (`venturiDiffuserLossK = 0.25`), raising diffuser irreversible loss `K_diff` to `~0.40–0.55` (`+1.2 to +2.4 inWC`), plus **`+0.5 to +1.0 inWC`** for the upstream splitter and downstream collector headers. Budget **`+1.5 to +3.0 inWC`** extra blower head (total system `~22.7–24.2 inWC` — still within the `2.0 HP` blower margin if `v_t` is trimmed from `70` to `67 m/s`, or step to `2.5–3.0 HP` for heavy surges).

---

## 5. Closed-Loop Sump Recirculation, Water/Soda Conservation & Exact Dump Point

### 5.1 Water Savings (`70%–85%` Reduction vs. Once-Through)
On the `24 cu.ft.` reference kiln, peak wet spray circulation across all three wet stages totals **`8.82 GPM`**:
- Stage 1 Quench spray: `1.16 GPM` (`0.58 GPM` evaporated + `0.58 GPM` unevaporated overspray)
- Stage 2 Venturi throat spray: `2.30 GPM` (`L/G = 0.0012`)
- Stage 3 Packed-bed liquor recirculation: `5.36 GPM` (`7.05 GPM/ft²`)

Over a `16-hr` firing:
- **Fully once-through (all 3 wet stages):** `8.82 GPM × 960 min = 8,467 gal` *(completely impractical on a residential hose/sewer!)*
- **Semi-once-through (Stage 3 bed recirculates, Stages 1+2 run fresh tap water):** `(1.16 + 2.30) GPM × 960 min = 3,315 gal` *(plus unbuffered tap water corrodes 316L SS in Stages 1–2!)*
- **Full Closed-Loop Recirculation (Shared Baffled Sump + IPC 608 Air-Gap Makeup):** Only **evaporated water (`0.58 GPM` peak; `~0.42–0.58 GPM` firing average)** plus amortized sump fill (`40 gal`) leaves the loop:
  - Peak-conservative `16-hr` makeup: `0.58 × 960 + 40 =` **`595 gal`** (or `672 gal` with optional `0.08 GPM` cooled ash bleed) → **`79.7%–82.1%` water saved** vs. semi-once-through (and **`92%–93%`** vs. fully once-through)!
  - Realistic ramp-averaged `16-hr` makeup (`~70%` average thermal load): **`~430–490 gal`** (**`~85%` water saved**).

### 5.2 Closed-Loop Stoichiometric Chemistry & Exact Sump Dump Point
During the `3.0-hr` active salting window (`2.0 lbs/hr` NaCl, `99.04%` HCl capture), the neutralization reaction $\text{HCl} + \text{NaHCO}_3 \to \text{NaCl} + \text{H}_2\text{O} + \text{CO}_2$ converts airborne HCl into soluble table salt (`NaCl`) right inside the sump:
- **Dissolved Solids Added per 3-hr Salt Firing:** `2.70 kg` (`5.94 lbs`) `NaCl` formed + `0.58 kg` (`1.28 lbs`) residual excess `NaHCO₃` buffer = **`3.28 kg` (`7.22 lbs`) total dissolved salts**.
- **In a `55-gal` Open-Head HDPE Baffled Sump (`40 gal` / `151.4 L` working liquor volume):**
  - Initial `pH 8.3–8.5` pre-charge (`3–4 lbs` `NaHCO₃`): `~10–12 g/L` baseline TDS (`SG ≈ 1.008`).
  - Each `3-hr` salt firing adds **`+21.6 g/L`** TDS (`+1.9% wt` salinity).
  - **Safe Operating Ceiling (`150 g/L` TDS ≈ `13.6% wt` NaCl):** Leaves `> 200 g/L` headroom below `68°F` NaCl crystal saturation (`358 g/L`), preventing salt encrustation on pump ceramic shafts, `0.06"` spray orifices, and PP Pall rings (`docs/ENGINEERING_VALIDATION.md` B6).
  - **Resulting Dump Cadence:** **`5 to 6` intermittent salt firings per full sump dump** (`40 gal` working sump; `~4` firings in a `25 gal` sump), topping up only **`~9.8 lbs` (`~2/3` of a `15 lb` bulk bag)** of `NaHCO₃` per firing to restore alkalinity!

### 5.3 Baffled Settling Drum & Simple Potter Field Tests (No Lab Equipment Required)
Install a removable vertical **over/under HDPE weir baffle** inside the atmospheric `55-gal` sump drum:
- **Zone 1 (Dirty Return Basin — 60% volume):** Receives hot drainage from Stage 1 quench, Stage 2 sweep/venturi, and Stage 3 column. Low velocity lets heavy wood fly ash (`ρ ≈ 2,400 kg/m³`) settle cleanly to the bottom sludge zone.
- **Zone 2 (Clean Alkaline Suction Basin — 40% volume):** Clear liquor overflows the mid-height weir through a removable `50-mesh` stainless screen into the mag-drive pump suction port (with a `50-mesh` inline Y-strainer downstream of the pump).
- **Mandatory IPC 608 Potable Isolation:** Feed makeup water strictly through an **engineered `>= 1"` vertical Air Gap** above the drum's overflow rim (or a certified **ASSE 1013 RPZ** backflow preventer) — never submerge a garden hose or float valve directly in chemical brine.

**Four Plain-English Field Triggers (When to Top Up vs. Full Dump):**

| Field Check | Tool & Cost | Normal Operating Target | Action / Dump Trigger | Code / Safety Reference |
| :--- | :--- | :---: | :--- | :--- |
| **1. Dissolved Salt (TDS)** | **Brine / Salinity Hydrometer** (`$12`, `SG 1.000–1.220`) | `SG 1.010 – 1.095` (`10–135 g/L`) | **`SG >= 1.105`** *(at `68°F`, `>= 150 g/L` TDS)* → **Full Sump Dump & Fresh Recharge** | Prevents pump seal & nozzle salt scaling (`ENGINEERING_VALIDATION.md` B6) |
| **2. Alkalinity Buffer (pH)** | **Wide-Range pH Strips** (`$10`/roll, `pH 1–14`) | **`pH 7.5 – 8.5`** *(check hourly during salting)* | **`pH < 7.0`** → stir in `2–3 lbs` `NaHCO₃` slurry; if `pH` drops rapidly or stays `< 7.0` near high SG → **Dump** | Protects 316L SS from acid pitting & keeps HCl absorption irreversible |
| **3. Settled Wood Ash (TSS)** | **Wooden Dipstick** in Zone 1 settling basin | `< 2"` settled ash layer | **`> 4"` ash sludge depth** → decant clear liquor, scoop ash cake, rinse basin | Prevents Y-strainer clogging & sanitary sewer obstruction (`40 CFR 403.5(b)(3)`) |
| **4. Drain Temperature & Release pH** | **Instant-Read Thermometer** (`$12`) + **pH Strips** | Drain only when **`<= 140°F`** AND **`pH 6.5 – 8.5`** | **NEVER drain `184°F` hot sump water directly into PVC/ABS pipes!** Cool overnight or temper with cold water first. | **IPC 702.5** (`<= 140°F` PVC/ABS limit), **IPC 803.1**, **40 CFR 403.5(b)(2)** (`pH >= 5.0` federal floor); verify non-hazardous if no Pb/Ba/Cd/Cr glazes fired |

---

## 6. Additional Cost, Time & Footprint Savers

### 6.1 Compact U-Turn Folded Skid (`compact_footprint = true`)
- **Problem:** The linear 4-stage `24 cu.ft.` layout (`pkg/scrubber/scrubber.go:164`) spans **`12.9 ft` long × `5.4 ft` wide** (`70.4 sq.ft.` / `6.54 m²`), which rarely fits down a `6–8 ft` urban side-yard beside a kiln shed while preserving a `36"` service aisle.
- **Solution:** Mount Stage 1 Quench + vertical/inclined Stage 2 Venturi on Leg 1, turn `180°` via two long-radius `90°` sweep elbows (`ΔP` penalty only **`~0.15–0.25 inWC`**) directly into the base of Stage 3 Packed Column on Leg 2, and tuck Stage 4 ID Blower beside Stage 1 over the shared sump.
- **Quantitative Impact:** Cuts overall length from **`12.9 ft` down to `~7.0 ft`** (**`-46%` length reduction**), with width increasing modestly from `5.4 ft` to `~6.2 ft` — shrinking total pad footprint from `70.4 sq.ft.` to **`~43.4 sq.ft.` (`-38%` area)** so the entire scrubber sits on one compact `7.5 ft × 6.5 ft` concrete pad while maintaining **NFPA 211 `36"` hot-flue clearance** and a **`36"` maintenance aisle**.

### 6.2 Shared Baffled Single Sump + Single Chemical Pump (`shared_baffled_sump = true`)
- **Savings:** Replaces two separate catch basins + dual chemical pumps with **one `55-gal` baffled open-head HDPE drum (`$96`)** and **one `10–12 GPM @ 40–50 ft TDH` seal-less PP/EPDM mag-drive pump (`~$360–$420`)** feeding a 3-branch CPVC manifold (Stage 1 quench `1.2 GPM`, Stage 2 venturi `2.3 GPM`, Stage 3 bed `5.4 GPM` via throttling ball valves + rotameters) — saving **`~$400–$650`** and halving plumbing joints.
- **Corrosion Bonus:** Automatically supplies buffered `pH 8.0–8.5` `NaHCO₃` liquor to Stage 1 quench and Stage 2 venturi, which is essential to protect **standard 316L SS (light/occasional)** shells from hot acidic HCl pitting.

### 6.3 Pure Electromechanical Safety Controls vs. Raspberry Pi Modbus (`control_tier = "electromechanical_only"`)
Per [docs/CONTROL_ALTERNATIVES_AND_RPI_BOM.md](CONTROL_ALTERNATIVES_AND_RPI_BOM.md), **every** `kilnscrub` installation—digital or analog—must include the **independent NFPA 86 hardwired 24VDC de-energize-to-trip safety loop** (`200°F` manual-reset bimetallic snap-disc + `0.25 GPM` mechanical paddle flow switch + latching safety relay + E-stop + Belimo spring-return fail-open bypass damper + VFD hardwired STO).

Choosing **Pure Electromechanical Controls** omits the Raspberry Pi, isolated Modbus RTU I/O cards, `4–20 mA` draft/pH transmitters, and RTD/TC analog converters, replacing them with a **`$95` mechanical Dwyer Magnehelic (`-1.0 to +0.2 inWC`) dial gauge** and the **VFD's built-in front-panel speed knob**:
- **Hardware Savings:** Drops control electronics from **`$1,486`** (`$2,501` turnkey w/ VFD/damper/NEC) down to **`~$350–$550`** (`~$1,365–$1,565` turnkey) — saving **`~$935–$1,135`** (**`~$900–$1,200`**) and 1–2 weekends of Linux/Modbus wiring.
- **Reliability & Safety:** **Zero** SD-card wear, OS freezes, RS-485 EMI glitches, or firmware maintenance; 100% identical NFPA 86 `200°F`/low-flow/power-loss fail-safe protection.
- **Honest Trade-Off:** Requires manual VFD knob tweaks during wood-stoking bursts (or running slightly higher steady suction) and gives up browser telemetry charts and automated firing CSV logs. You can always bolt a Raspberry Pi + Modbus layer on later without rewiring the hardwired safety loop.

---

## 7. Quick-Reference Option Selector for Studio Builders

| Studio Priority | Recommended Phase 7 Combination | Expected Savings vs. Baseline (`wood_salt_24`) | Key Compliance / Engineering Reminder |
| :--- | :--- | :--- | :--- |
| **Lowest Cost, Bulletproof Backyard Salt Kiln** | Inline ID Blower + `salting_hours=3` + Shared Baffled Sump + U-Turn Skid + **Electromechanical Controls** + `2` Parallel `1.5"` Mini-Venturis | **`~$1,450–$1,900` CapEx saved**; `-81%` `NaHCO₃`; `-80%` water; `-46%` length | Retain full NFPA 86 hardwired `200°F` snap-disc + NEC 430.122 `240V 30A` + IPC 608 air gap + IMC 501.3.1 `30 ft` salt setback |
| **Zero Acid-Exposed Fan (Outdoor Gas-Soda / Occasional Salt)** | **Dry-Fan Downflow Eductor** + `18"` FRP Column (or `16–18` parallel `4"` pipes) + Shared Baffled Sump + Electromechanical Controls | **`~$1,100–$1,600` CapEx saved**; zero wet fan corrosion; `141°F` CPVC margin | **Outdoor-only!** Pressure-test Stage 3 flanges at `+6 inWC`, include `>= 8"` sump U-trap seal, size `240V 40A` if `3.0 HP` salt duty |
| **No-Freight, Local Plumbing-Supply DIY Build** | Inline ID Blower + **Parallel `12 × 4"` Sch 80 CPVC Bed (`1/2"` PP rings)** + **`2` Parallel `1.5"` Tri-Clamp Venturis** + U-Turn Skid | **`~$450–$900` CapEx saved**; zero welding; zero LTL freight | Install pressurized `20 psig` individual nozzle header (`12 × 0.45 GPM`) + `50-mesh` Y-strainer so pipes cannot dry out or flood |

---

## 8. Cross-References
- **Overall Stage-Count Trade Study & Public-Domain Patents:** [docs/ARCHITECTURE_TRADE_STUDY.md](ARCHITECTURE_TRADE_STUDY.md)
- **Validated Scrubber Physics & Audit:** [docs/ENGINEERING_VALIDATION.md](ENGINEERING_VALIDATION.md)
- **US Model Code Compliance (NEC / IMC / IPC / NFPA / OSHA):** [docs/CODE_COMPLIANCE_AUDIT.md](CODE_COMPLIANCE_AUDIT.md)
- **Five Control Architectures & Full Raspberry Pi BOM:** [docs/CONTROL_ALTERNATIVES_AND_RPI_BOM.md](CONTROL_ALTERNATIVES_AND_RPI_BOM.md)
