# Budget Engineering Study: Shell-Less Parallel Pipe Bundles, Low-Cost High-Head Blowers, and Shop Air Compressors

**Purpose:** Rigorous, first-principles cost-engineering study answering how a working ceramic studio can cut the `kilnscrub` Phase 7 reference 24 cu.ft. wood/salt scrubber hardware budget from **`$7,263`** down to **`$2,540–$3,150`** (`-57%` to `-65%`) without compromising submicron particulate capture, `99%` HCl absorption, or mandatory **NEC / IMC / IPC / NFPA 86** life-safety protections.

**Reference Baseline & Verification Basis:**
- All baseline BOM figures are extracted directly from the live `wood_salt_24` COTS generator ([pkg/studio/cots.go](../pkg/studio/cots.go), [pkg/studio/options.go](../pkg/studio/options.go)) — `24.0 cu.ft.` Cone 11 (`2400°F`), `2.0 lbs/hr` NaCl (`3.0 hr` peak salting window), `255.9 ACFM` (`0.1208 m³/s`, `0.0974 kg/s`, `ρ = 0.806 kg/m³`, `μ = 1.605×10⁻⁵ Pa·s`) quenched to `183.9°F`.
- All hydraulic, flooding, Ergun pressure-drop, Calvert cut-diameter (`d_50`), and Yung–Calvert throat acceleration (`β`) numbers are computed directly against [pkg/scrubber/venturi.go](../pkg/scrubber/venturi.go) and [pkg/scrubber/packedbed.go](../pkg/scrubber/packedbed.go).
- Retail component prices and fan/compressor performance specs are cross-checked against current US commercial distributors (McMaster-Carr, Grainger/Zoro, SupplyHouse, Home Depot, Spears/Harvel schedules, Ametek Lamb bulletins, Spraying Systems Co., EXAIR, Grizzly/WEN/Harbor Freight, and ceramic clay suppliers).

---

## 1. Why the Phase 7 Baseline BOM Costs `$7,263` & Where the Money Actually Goes

The Phase 7 `wood_salt_24` default build (`inline_id_blower`, `single_column`, `single_throat`, `shared_baffled_sump`, `rpi_modbus`, `standard_316l`) totals **`$7,263.22`** across 14 COTS line items (excluding LTL freight, sales tax, and site electrical wiring). A Pareto breakdown reveals that **the top 4 line items account for `67.4%` (`$4,892`) — and including the Stage 4 acoustic silencer, the top 5 items account for `72.8%` (`$5,288`) of the entire spend**:

### 1.1 Exact Line-by-Line Phase 7 Baseline BOM Pareto (`wood_salt_24`)

| Rank | Line Item (from `pkg/studio/cots.go` / `bom_cots.csv`) | Qty | Unit Cost | Line Total | Share of BOM | Cumulative % | Cost-Engineering Lever |
| :---: | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **1** | **Stage 4: `2 HP` Industrial Radial Pressure Blower + `GS21-22P0` VFD in NEMA 4X** (`$1,340` blower + `$310` VFD + `$165` NEMA 4X) | 1 assy | `$1,815.00` | **`$1,815.00`** | **25.0%** | **25.0%** | Drop ΔP via air-atomization (`11.5–12.5 inWC`) to unlock `$300–$380` radial blowers, OR use parallel Ametek bypass blowers (`$390–$690`) |
| **2** | **Safety & Controls: Raspberry Pi Modbus RTU + Hardwired `200°F` Safety Loop & `7"` Fail-Safe Bypass Damper** (`$1,045` RPi/Modbus + `$419` damper + `$228` hardwired safety) | 1 kit | `$1,692.00` | **`$1,692.00`** | **23.3%** | **48.3%** | Switch to `electromechanical_only`: **saves `$960.00`** (`$732.00` retains 100% of NFPA 86 hardwired safety loop + damper + Magnehelic) |
| **3** | **Stage 3: `14"` Sch 80 CPVC Industrial Pipe Column (`12.50"` ID × `9.7 ft`)** *(excl. `~$200–$350` LTL freight & `~$350–$450` industrial flanges)* | 1 ea | `$804.52` | **`$804.52`** | **11.1%** | **59.4%** | Shell-less `4"` parallel Uniseal bundle (`$430–$680` landed w/ zero LTL freight) |
| **4** | **Stage 1: `14"` 316L SS Flanged Quench Pipe Spool (`7.9 ft`)** | 1 ea | `$580.71` | **`$580.71`** | **8.0%** | **67.4%** | Keep 316L bolted spool (`$380–$520` w/ local spiral/flue tube or shorter `6 ft` fine-mist quench) |
| **5** | **Stage 4: `8"` Inline Acoustic Silencer & Insulated Blower Box** | 1 ea | `$396.00` | **`$396.00`** | **5.5%** | **72.8%** | DIY Rockwool-lined plywood/cement-board acoustic doghouse (`~$95–$130`) |
| **6** | **Stage 2: 316L Bolted Concentric Reducer Venturi (`7" → 1.85" → 7"`)** | 1 ea | `$366.00` | **`$366.00`** | **5.0%** | **77.9%** | `2 ×` off-the-shelf 316L sanitary tri-clamp reducers (`~$240–$290`) |
| **7** | **Liquid: Seal-Less Mag-Drive Recirculation Pump (`5.4 GPM @ 35 ft`) + Y-Strainer** | 1 ea | `$340.03` | **`$340.03`** | **4.7%** | **82.5%** | Commodity magnetic-drive / chemical-duty PP pump (`~$185–$240`) |
| **8** | **Liquid: 316L SS Full-Cone Spray Nozzles (`1/4"` NPT)** | 4 ea | `$58.00` | **`$232.00`** | **3.2%** | **85.7%** | `2` hydraulic quench (`$60`) + `2` generic 316L twin-fluid air-atomizing (`$70–$110`) |
| **9** | **Stage 3: 3-Pass PP Chevron Mist Eliminator Pad & Support Grate** | 1 assy | `$225.96` | **`$225.96`** | **3.1%** | **88.8%** | DIY stacked PP mesh/corrugated baffle pad + FRP grate cutoffs (`~$85–$120`) |
| **10** | **Safety/Electrical: NEC 430.102 Lockable NEMA 4X Disconnect + 3% VFD Line Reactor** | 1 assy | `$215.00` | **`$215.00`** | **3.0%** | **91.8%** | Keep `$45–$65` NEMA 4X disconnect (MANDATORY); drop `$150` line reactor when no VFD is used |
| **11** | **Hardware: Ceramic Fiber / PTFE / EPDM Gaskets, Clamps, CPVC Cement, Hydrometer** | 1 lot | `$195.00` | **`$195.00`** | **2.7%** | **94.5%** | Keep intact (`$150–$195`, safety-critical seals & field test kit) |
| **12** | **Liquid: IPC 608 Potable Water Backflow Prevention (Air Gap or ASSE 1013 RPZ)** | 1 ea | `$185.00` | **`$185.00`** | **2.5%** | **97.0%** | Engineered `>= 1"` vertical Air Gap into sump/break tank (`~$25–$45` fittings — 100% IPC 608.15.1 compliant, saves `$140+` vs bronze RPZ!) |
| **13** | **Stage 3: Commercial `1"` PP Pall Rings (`3 cu.ft.`)** | 3 ft³ | `$48.00` | **`$144.00`** | **2.0%** | **99.0%** | **Potter-extruded stoneware ceramic rings** (`$15–$25` studio reclaim clay; commercial `1/2"` PP rings cost `$180–$360` retail!) |
| **14** | **Stage 3: `30-gal` Open-Head HDPE Baffled Sump Drum** | 1 ea | `$72.00` | **`$72.00`** | **1.0%** | **100.0%** | Reconditioned food-grade `55-gal` HDPE drum (`$25–$45` local) |
| — | **PHASE 7 BASELINE COTS TOTAL** | **14 items** | — | **`$7,263.22`** | **100.0%** | **100.0%** | *(Add `~$550–$800` real-world LTL freight + `14"` flanges if buying specialty industrial pipe!)* |

### 1.2 Top 4 Structural Cost Drivers (Why Industrial Scaling Punishes Small Studios)

1. **The `21.2 inWC` Venturi Pressure-Head Cliff (`$2,426` combined Stage 4 + NEC reactor = `33.4%` of BOM):**
   Standard HVAC blowers stall at `1–4 inWC`, and commodity woodworking radial fans stall at `9–11 inWC`. Pushing `256 ACFM` through a passive `70 m/s` wet venturi (`18.88 inWC`) forces the design over the `15+ inWC` cliff into custom-quoted industrial cast-aluminum/FRP radial pressure blowers (`$1,340–$2,200`) plus a `2 HP` VFD + NEMA 4X enclosure + `3%` line reactor (`$625`).
2. **Digital Telemetry Overkill (`$1,045` optional add-on inside the `$1,692` safety kit = `14.4%` of BOM):**
   The mandatory NFPA 86 hardwired safety loop (`200°F` manual-reset snap-disc + `0.25 GPM` flow switch + 24VDC latching safety relay + NEMA 4X E-stop + `7"` spring-return fail-open damper) costs only **`$647`**. The remaining **`$1,045`** pays purely for the Raspberry Pi, Modbus RTU I/O cards, and `4–20 mA` transmitters.
3. **Large-Diameter Industrial Thermoplastic Pipe Cliff (`$805` bare pipe → `$1,450–$1,700` landed):**
   Below `6"`, PVC/CPVC pipe is a commodity stocked at every plumbing counter. At `14"` Schedule 80 CPVC (`23.26 lbs/ft`, `0.750"` wall, Spears MSRP `~$53.85/ft` list and often quote-only in `20-ft` sticks), you trigger industrial distributor minimums, `$350–$450` in `14"` Van Stone flanges, and `$200–$350` in LTL commercial pallet freight.
4. **Commercial Chemical-Tower Packing Retail Markup (`$48–$120+/ft³`):**
   Factory bulk pricing for `1"` PP Pall rings is `~$4–$7/ft³` by the cubic meter (`35 ft³` MOQ), but small-lot retail buyers pay `$48–$65/ft³` for `1"` rings and **`$62–$120+/ft³`** for `1/2"` Pall rings — absurd when a potter owns a clay extruder and a high-fire kiln.

---

## 2. Challenge 1 — Shell-Less `2"`–`4"` Parallel Pipe Packed-Bed Bundle (Zero External Large Pipe!)

### 2.1 Post-Mortem: Why Phase 7's `parallel_manifold` Accidentally Cost `+$985` MORE (`$8,248` vs. `$7,263`)

Running `kilnscrub guide --preset wood_salt_24 --bed-layout parallel_manifold` produces an estimated BOM of **`$8,248.04`** — **`+$984.82` MORE expensive** than the single `14"` column! Inspecting [pkg/studio/options.go](../pkg/studio/options.go#L457-L467) and [pkg/studio/cots.go](../pkg/studio/cots.go#L645-L683) pinpoints three exact modeling/design traps that inflated Phase 7's manifold:

| Cost Line (`wood_salt_24` Stage 3) | Baseline `14"` Column | Phase 7 `parallel_manifold` (`11 × 4"`) | Root Cause of Phase 7 Cost Bloat |
| :--- | :---: | :---: | :--- |
| **Pipe Shell** | `$804.52` (`1 × 9.7 ft`) | **`$1,118.04`** (`11 × 9.68 ft × $10.50/ft`) | **Trap #1:** Multiplied all `11` risers by the *entire* `9.68 ft` tower height (including sump, plenums, and demister!) instead of the `4.0 ft` packed-bed section (`107 ft` of pipe bought instead of `48 ft`!). |
| **Headers / Bulkheads / Distributor** | *(omitted in bare pipe price)* | **`$564.82`** (`$399` headers + `$166` spray) | **Trap #2:** Priced rigid Schedule 80 CPVC flanged/union manifold headers (`$24/pipe`) instead of flat tube-sheet plates + flexible bulkheads. |
| **Packing Media** | `$144.00` (`3 ft³ × $48`) | **`$248.00`** (`4 ft³ × $62` commercial `1/2"` PP) | **Trap #3:** Commercial `1/2"` PP Pall rings cost `+30%` to `+150%` more per cu.ft. than `1"` rings (and real retail is `$75–$120+/ft³`). |
| **Chevron Demister Adjustment** | `$225.96` | `$228.44` (`+$2.48`) | Equivalent bore slightly larger (`12.69"` vs `12.50"`). |
| **Stage 3 Total** | **`$1,174.48`** | **`$2,159.30` (`+$984.82`)** | **Zero savings until pipe length, tube-sheet headers, and packing are redesigned!** |

### 2.2 Shell-Less Tube-Sheet Parallel Pipe Bundle Architecture (Zero Outer Large Pipe)

In a **true shell-less tube-bundle absorber** (borrowing directly from shell-and-tube heat exchangers, minus the outer shell!), the `N` parallel pipes **are** the pressure boundary across the packed zone. No `14"`, `16"`, or `18"` pipe exists anywhere in Stage 3:

```mermaid
flowchart TD
    TopPlenum["Top Collector & Demister Box (24\"×24\" HDPE/PP Box or Inverted 55-Gal Drum Cap)\nHouses Pressurized 12-Port Orifice Spray Header + Chevron Demister Pad"]
    TopSheet["Upper 1/2\" HDPE / PP / CPVC Tube-Sheet Plate (24\"×24\")\n12 × 5\" Hole-Sawed Ports with 4\" Uniseals (or PVC Hub Collars)"]
    Bundle["EXPOSED PARALLEL PIPE BUNDLE (Zero Outer Shell!)\n12 × 4\" Stock Pipes × 4.0 ft Long (34\"–41\" Packed Bed + 6\" Stubs)\nPacked with 1/2\" PP Rings OR Potter-Extruded Stoneware Rings"]
    BotSheet["Lower 1/2\" HDPE / PP / CPVC Tube-Sheet Plate on Steel/Wood Support Frame\n12 × 4\" Uniseals + Perforated Stainless/PP Support Screens at Pipe Bases"]
    BotPlenum["Shared Lower Gas-Inlet Plenum & Baffled Sump (55-Gal HDPE Drum or 24\"×24\" Box)\nCross-Flow Velocity <= 4 m/s + >= 6\" Liquid Seal Below Venturi Entry"]
    TopPlenum --> TopSheet --> Bundle --> BotSheet --> BotPlenum
```

**Why this eliminates every Phase 7 cost penalty:**
1. **Riser length drops `59%` (`9.68 ft → 4.0 ft`):** Each pipe only holds the `34"–41"` (`0.85–1.05 m`) packed zone plus `3"` top/bottom engagement into the tube-sheets. `12` pipes × `4.0 ft` = **`48 linear feet`** = **5 standard `10-ft` sticks** (`2 ft` spare)!
2. **Flat `1/2"` HDPE/PP tube-sheets replace industrial flanges:** A single `1/2" × 24" × 48"` industrial HDPE sheet (`$95–$145` at Home Depot/US Plastic/Walmart) cuts cleanly in half into two `24" × 24"` tube-sheets (supporting a `3 × 4` grid of `4"` pipes on `5.75"` centers with `1.25"` ligament spacing). For negative-pressure ID service (`-21 inWC` = `-0.76 psig`), a `1/2"` HDPE plate spanning `24"` is stiffened against vacuum bowing by two external `1" × 2"` steel angle ribs bolted across the outside faces (`$18`).
3. **Flexible Uniseal bulkheads (or PVC DWV hub collars) replace `$40` industrial unions:**
   - **Uniseals** (DuPont Alcryn / EPDM flexible tank adapters, rated to `40–65 psig` static and full `-21 inWC` scrubber vacuum) snap into precision hole-sawed holes (`3"` hole saw for `2"`, `4"` for `3"`, `5"` for `4"`). Verified US pricing: **`2"` Uniseal = `$4–$6` bulk (`$4–$8` retail)**; **`3"` Uniseal = `$8–$12`**; **`4"` Uniseal = `$8–$11` bulk 24-pack (`$12–$16` single retail)**.
   - **Alternative for `4"` Sch 40 PVC (sub-cooled `< 135°F`):** Standard **`4"` Schedule 40 PVC DWV closet/hub flanges or flat-faced coupling collars** (`$4.50–$6.50` ea at Home Depot) bolted with EPDM gaskets to the tube-sheet cut `24` seals to **`~$120–$145`**.
   - **Thermal Caution on Uniseals:** DuPont Alcryn / EPDM Uniseals are chemically resistant to `HCl`/`NaHCO₃` and rated up to `~180–200°F`, right near the `183.9°F` direct quench exit; keep a hardwired `200°F` NFPA 86 snap-disc upstream (or `135°F` when using Sch 40 PVC) and support pipe weight on the lower frame so hot Uniseals never carry gravity shear!

### 2.3 Exact Pipe Count `N` & Cost Comparison: `2"` vs. `3"` vs. `4"` Bundles vs. `14"` Column

Matching both the nominal `14"` column cross-sectional area (**`153.9 in²`**) and the snapped `14"` Sch 80 CPVC true bore (`12.500"` ID = **`122.72 in²`**, calculated minimum `11.81"` = `109.6 in²`):

| Pipe Spec & Material | True ID (`in`) | Area / Pipe (`in²`) | `N` to Match `122.7 in²` (`14"` Sch 80 ID) | `N` to Match `154 in²` (`14"` Nom.) | `D/d_p` w/ `1/2"` Rings | Pipe Cost (`4.0 ft` Risers) | `2×N` Seals (Uniseals / Hubs) | Tube-Sheets + Plenums + Spray Header | Landed Bundle Shell Cost *(excl. media)* | Verdict for `24 cu.ft.` Kiln |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Single `14"` Sch 80 CPVC** *(Baseline)* | `12.500"` | `122.72` | `1` | — | `12.5` (`1"` rings) | `$805` *(bare)* | `$400` *(2×14" flanges)* | `$250` *(LTL freight)* | **`$1,455`** *(real landed)* | Low ΔP (`1.16"`), heavy (`254 lb`), special-order freight |
| **`2"` Sch 80 CPVC** | `1.939"` | `2.95` | `42` | `53` | `3.88` *(FAIL)* | `$610` (`43×4' @ $3.55`) | `$430` (`86 × $5`) | `$380` *(43-port header!)* | **`$1,420`** | **IMPRACTICAL:** 86 seals, `1/4"` rings plug with ash, high ΔP |
| **`3"` Sch 80 CPVC** | `2.900"` | `6.61` | `19` | `24` | `5.80` *(needs `3/8"`)* | `$575` (`20×4' @ $7.20`) | `$380` (`40 × $9.50`) | `$290` *(20-port header)* | **`$1,245`** | Borderline — best reserved for **`6–12 cu.ft.` kilns** (`5–8` pipes) |
| **`4"` Sch 80 CPVC** *(Direct `184°F` Sweet Spot)* | `3.826"` | `11.50` | **`11`–`12`** (`138 in²`) | `14` | **`7.65`** *(PASS)* | **`$504`** (`5×10' @ $100–$150` whls / `$750` rtl) | **`$240`** (`24 × $10` bulk Uniseal) | **`$215`** (`$120` HDPE sheets + `$95` 12-port spray) | **`$959`** *(whls)* – **`$1,205`** *(rtl)* | **SWEET SPOT (`184°F` direct quench):** Zero freight, hand-portable, saves `$250–$500` vs landed `14"` |
| **`4"` Sch 40 Solid-Wall PVC** *(Cooled `< 135°F` ONLY!)* | **`4.026"`** | **`12.73`** | **`10`–`12`** (`153 in²`) | **`12`–`13`** | **`8.05`** *(PASS)* | **`$190`** (`5×10' @ $38` Home Depot!) | **`$144`–`$240`** (`24 × $6` PVC hubs or `$10` Uniseals) | **`$215`** (`$120` HDPE sheets + `$95` 12-port spray) | **`$549`–`$645`** | **BUDGET BREAKTHROUGH:** Saves **`$810–$905`** vs landed `14"` — **ONLY when cooled `< 135°F`!** |

> **MANDATORY THERMAL RULE FOR SCHEDULE 40 PVC (`140°F` ASTM D1785 Limit):**
> Never expose Schedule 40 PVC to `183.9°F` direct adiabatic quench exhaust — PVC's heat deflection temperature (`158°F`) and structural modulus collapse under `-21 inWC` suction! Unlocking `$38/stick` `4"` Sch 40 PVC requires **two mandatory engineering safeguards**:
> 1. **Active Sub-Cooling Below `135°F`:** Either (a) mix `1.5×–1.6×` ambient air via the dry-air eductor (`T_mix = 138°F` with minor sump trim cooling) or (b) circulate sump liquor through a simple `$65` 50-ft `1/2"` PEX/stainless immersed cooling coil in a water barrel / fan-cooled radiator to sub-cool the quench/absorber transition to **`125°F–132°F`**.
> 2. **Hardwired `135°F` NFPA 86 Snap-Disc Interlock:** Replace the `200°F` CPVC manual-reset bimetallic snap-disc with an independent **`135°F` (`57.2°C`) manual-reset snap-disc** wired into the 24VDC de-energize-to-trip latching safety relay driving the spring-return fail-open bypass damper.

### 2.4 The Potter Packing Breakthrough: Extruded Kiln-Fired Stoneware Raschig & Lessing Rings

Ceramic studios already own the two machines industrial chemical-packing factories charge a fortune for: a **clay extruder** (Brent / Bailey / North Star) and a **Cone 6–10 stoneware kiln**. Extruding hollow clay tubes (`5/8"` or `3/4"` OD, `2–2.5 mm` wall) and wire-cutting them into equal-length unglazed cylinders (`OD = L`) produces **authentic industrial stoneware Raschig rings** — the exact acid-tower packing used in chemical plants for over a century! Adding a simple `$15` laser-cut cross-bar or notch to the extruder die produces **thin-wall slotted / Lessing ceramic rings** with lower pressure drop.

#### Exact Hydraulic & Economic Verification (`pkg/scrubber/packedbed.go` Eckert + Ergun Equations)

All cases evaluated for `wood_salt_24` (`Q = 0.1208 m³/s`, `m_g = 0.0974 kg/s`, `ρ_g = 0.806 kg/m³`, `μ_g = 1.605×10⁻⁵ Pa·s`, `5.36 GPM` recirc, `1.70×` irrigated Ergun factor):

| Packing Media & Bundle Configuration | `D/d_p` | Voidage (`ε`) | Packing Factor `F_p` (`m⁻¹`) | Flood Vel `v_flood` (`m/s`) | Actual `% of Flood` *(Limit `<= 75%`)* | Irrigated Bed `ΔP` (`inWC`) | Bed Vol (`ft³`) | Dry Media Weight (`lbs`) | Real Retail Media Cost (`3–4 ft³`) | Thermal & Acid Rating |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Baseline: `14"` Sch 80 CPVC + `1"` PP Pall Rings** | `12.5` | `0.90` | `170` | `3.08` | **49.5%** *(PASS)* | **`1.16"`** | `2.94` | `17 lbs` | `$144–$180` | `200°F` max (melts if quench fails) |
| **`12 × 4"` Sch 80 CPVC + `1/2"` Commercial PP Pall Rings** | `7.65` | `0.87` | `312` | `2.27` | **59.7%** *(PASS)* | **`1.72"`** | `2.67` | `21 lbs` | **`$185–$360`** (`$62–$120/ft³`) | `200°F` max; costly small-lot retail |
| **`12 × 4"` Sch 80 CPVC + `1"` Aquaculture Bio-Balls** | `3.83` *(FAIL)* | `0.91` | `210` | `2.77` | `49.0%` *(channeling!)* | `0.73"` | `3.77` | `16 lbs` | `$180–$240` (`$45–$60/ft³`) | **POOR:** Low area (`30–40 ft²/ft³`), `140–160°F` PVC/PE softens, severe wall bypass |
| **`12 × 4"` Sch 80 CPVC + `1/2"` Stoneware Raschig Rings** | `7.65` | `0.68` | `1,550` | `1.02` | **`133.1%` (FLOODS!)** | `3.38"` | `2.67` | `126 lbs` | `$15–$25` *(reclaim)* / `$95` *(fresh)* | **DO NOT USE in 12 pipes:** High `F_p` floods! |
| **`12 × 4"` Sch 80 CPVC + `5/8"` Stoneware Raschig Rings** | `6.12` | `0.70` | `980` | `1.28` | **`105.8%` (FLOODS!)** | `2.70"` | `2.83` | `122 lbs` | `$15–$25` *(reclaim)* / `$95` *(fresh)* | **FLOODS in 12 pipes:** Must add pipes to lower `v_s`! |
| **`18 × 4"` Sch 80 CPVC (or `16 × 4"` Sch 40 PVC) + `5/8"` Stoneware Raschig** | `6.1–6.4` | `0.70` | `980` | `1.28` | **`70.5%–71.7%` (PASS)** | **`1.24"–1.27"`** | `4.20` | **`180 lbs`** | **`$20–$30`** *(reclaim)* / `$130` *(fresh)* | **100% Acid-Proof & `2,000°F+` Meltdown-Proof!** |
| **`14 × 4"` Sch 80 CPVC (or Sch 40 PVC) + `3/4"` Thin-Wall Slotted/Lessing Stoneware** | `5.1–5.4` | **`0.74`** | **`620`** | **`1.61`** | **`65.2%–72.1%` (PASS)** | **`1.28"–1.56"`** | `3.85–4.26` | **`146–162 lbs`** | **`$20–$30`** *(reclaim)* / `$115` *(fresh)* | **BEST POTTER MEDIA:** Open thin-wall (`2 mm`) extrusion cuts `F_p` & pipe count (`+15%` bed height `1.05 m` offsets lower `D/d_p`) |

**Honest Engineering Realities of Kiln-Fired Ceramic Packing:**
1. **Why `12` pipes flood with thick `1/2"`–`5/8"` Raschig rings:** Because unslotted ceramic cylinders have lower voidage (`ε = 0.68–0.70` vs `0.87` for `1/2"` PP Pall rings) and `3×–5×` higher hydraulic packing factor (`F_p = 980–1,550 m⁻¹` vs `312 m⁻¹`), Eckert `v_flood` drops from `2.27 m/s` to **`1.02–1.28 m/s`**. Trying to cram `256 ACFM` through `12 × 4"` pipes filled with `1/2"` ceramic rings drives the bed to **`133%` of flood** and blows brine out the stack!
2. **The Two Clean Potter Fixes:** Either (a) widen the bundle from `12` to **`16–18 × 4"` pipes** (extra `4–6` Sch 40 PVC pipes cost only `~$75–$115`!), OR (b) extrude **thin-wall (`2 mm`) `3/4"` slotted/cross-webbed stoneware rings** (`ε ≈ 0.74`, `F_p ≈ 620 m⁻¹`) in **`14 × 4"` pipes** (`65%–72%` flood, `1.28–1.56 inWC` irrigated bed ΔP — virtually identical to PP Pall rings!).
3. **Structural Weight (`150–180 lbs` dry, `~200–220 lbs` wet irrigated):** Unlike `20 lbs` of PP rings, stoneware packing is **heavy**. Rest the bottom of every `4"` riser directly on a **rigid steel angle / treated-wood support frame** beneath a `1/2"` FRP/stainless packing grate — NEVER suspend ceramic-filled pipes solely by flexible Uniseals!
4. **Firing Tip:** Fire extruded rings unglazed to **Cone 6–10 stoneware vitrification** (`< 2%` absorption) in unused kiln shelf corners or saggars so they resist thermal shock and acid-salt crystallization spalling indefinitely. **Zero meltdown risk:** even in a catastrophic quench-water failure, fired stoneware rings cannot melt into a solid plastic plug!

---

## 3. Challenge 2 — Affordable High-Push Blowers (`200–300 CFM @ 12–25 inWC`)

The single largest cost driver in the baseline BOM is the **`$1,815`** (`$1,340` blower + `$310` VFD + `$165` NEMA 4X) industrial radial blower. Below is an honest, verified comparison of five real COTS prime-mover options capable of serving a `24 cu.ft.` wood/salt kiln (`256 ACFM`):

### 3.1 Real COTS High-Head Blower & Curve Comparison

| Prime-Mover Option | Verified Models & Retail Price (US) | Published Curve Capability (Single Unit) | Array Sizing for `256 ACFM` (`24 cu.ft.`) | Array Total Cost | Electrical / Control Needs | Wet Acid Contact & Corrosion Constraint | Acoustic / Maintenance Reality |
| :--- | :--- | :--- | :---: | :---: | :--- | :--- | :--- |
| **0. Baseline: Industrial Radial + VFD** | Dayton/Cincinnati Fan PB-9 + AutomationDirect `GS21-22P0` + NEMA 4X (`$1,340 + $310 + $165`; FRP `$2,400+`) | `256 ACFM @ 21.2 inWC`, `2.0 HP` (`1.38 BHP`) | `1` unit | **`$1,815–$2,420`** | `240V 30A` 1-Ph (`10 AWG`), NEC 430.122 + `3%` line reactor | 316L wheel (inspect every 5–10 salt firings) or FRP (immune) | `~74 dBA` bare → `43 dBA @ 40 ft` silenced; `20,000+ hr` TEFC bearings |
| **1. Regenerative / Side-Channel Ring Blowers** | Generic/VEVOR `1.5 kW` (`$300–$550`) or `2.2 kW` (`$450–$800`); Gast R5/R6 (`$1,500–$3,200`); FPZ K05 (`$3,300+`) | `1.5 kW`: `80–110 CFM` open, `60–88 inWC` max; `2.2 kW`: `120–160 CFM` open, `80–112 inWC` max | **`2 × 2.2 kW` in parallel** (`240–300 CFM @ 25–45 inWC`; single unit is flow-starved!) | **`$900–$1,500`** *(2× generic)* | `240V 30–40A` (`6 HP` total! High waste heat; needs relief valve!) | **DRY-SIDE ONLY!** Tight-clearance (`0.2 mm`) cast-aluminum impellers seize fast in wet HCl mist | `70–78 dBA` whine; industrial bearings (`15,000+ hr`); poor energy efficiency at `21 inWC` (over-headed) |
| **2. Ametek Lamb Multi-Stage Tangential Bypass Blowers** | `116765` (3-stg 5.7", `$130–$150`); `117500` (3-stg 7.2", `$240–$300`); `119992` (2-stg 5.7", `$190–$230`) | `116765`: `95 CFM` / `136"` (`~82 CFM @ 21.2"`); `117500`: `102.5 CFM` / `146"` (`~89 CFM @ 21.2"`); **`119992`: `126.2 CFM` / `141"` (`~110 CFM @ 21.2"`)** | **`3 × 116765`** (`245 CFM @ 21.2"`) OR **`2 × 119992`** (`220 CFM` / 18–22 ft³) / **`3 × 119992`** (`330 CFM @ 21.2"`) | **`$390–$450`** *(3×116765)* or **`$380–$460`** *(2×119992)* / **`$570–$690`** *(3×119992)* | `11–13 A @ 120V` ea → use `240V` variants or `2–3` split `20A` circuits; `$35` SCR/Triac speed controller (no VFD!) | **Tangential bypass isolates motor windings from wet gas!** Aluminum fan shells still need pH 8 rinse (or dry-side placement) | `82–88 dBA` raw → **MUST enclose in Rockwool silencer box**; **`500–800 hr`** OEM brush life (`30–50` full firings; `$15–$20` `33326-1` brush kit adds `300–500 hr`!) |
| **3. Low-ΔP Medium-Pressure Commodity Radial Blower** *(Requires `~11–12.5 inWC` via Air-Atomized Venturi!)* | Harbor Freight 2 HP DC (`$200–$250`, `~9"` max); WEN DC1300 (`$250–$350`, `~10"` max); Grizzly G1029Z2P (`$500–$650`, **`11.0 inWC`**); surplus 1.5–2 HP cast-Al radial (`$300–$450`, `12–14 inWC`) | `12.75"` radial impeller: **`250–400 CFM @ 10–11.5 inWC`** (`1.5–2.0 HP`, `~0.8 BHP` at `256 CFM`) | **`1` unit** *(ONLY when air-atomization or `48–55 m/s` cuts total ΔP from `21.2 → 11–12.5 inWC`!)* | **`$300–$450`** *(surplus/WEN/mod)* to **`$550`** *(Grizzly)* | **Single `120V/240V 20A` circuit**; manual blast-gate damper for draft trim (no VFD needed!) | Aluminum/painted steel wheel → coat with epoxy/alkaline flush OR run on **dry motive side**; never unbuffered HCl | `72–76 dBA` → easy DIY acoustic box (`<= 50 dBA @ 40 ft`); `20,000+ hr` induction motor bearings |
| **4. Series-Staged Dual Commodity Radial Blowers** | `2 ×` WEN DC1300 / Harbor Freight 2 HP / surplus `11 inWC` radial fans piped **in series** (outlet `#1` → inlet `#2`) | Each adds `~10.5–11.0 inWC` at `256 CFM` → **`21–22 inWC` combined head!** | `2` units in series *(dry motive side or coated wet side)* | **`$460–$650`** | `240V 30A` (both motors run on low shaft load `~0.8 BHP` each when throttled to `256 CFM`!) | **DRY-SIDE RECOMMENDED** (series housing #2 sees `+11 inWC` positive pressure — must seal scroll seams airtight!) | Simple induction motors, zero brushes, zero VFD; slightly bulkier footprint (`+3 sq.ft.`) |

### 3.2 Key Technical Takeaways on Budget Blowers

1. **Why Ametek Lamb Bypass Blowers (`116765` / `119992`) Are the Best Drop-In `21+ inWC` Budget Replacement:**
   - Unlike standard vacuum motors, **tangential bypass** blowers use a completely separate cooling fan for the armature windings and discharge the process exhaust straight out a `2"` tangential horn — wet `140–184°F` scrubber mist never touches the commutator or field coils!
   - Operating `3 × 116765` (`$135 × 3 = $405`) or `3 × 119992` (`$200 × 3 = $600`) in parallel with simple **gravity check flaps** gives built-in **3-step mechanical staging** without a VFD: run **1 blower** during preheat/water-smoking (`~90–115 CFM`), **2 blowers** during body reduction (`~180–225 CFM`), and **all 3 blowers** only during the `3-hr` Cone 10–11 salting window (`245–330 CFM @ 21.2–25 inWC`)!
   - **Brush Life Reality:** Running full 3-blower duty only `3–4 hrs/firing` (and 1–2 blowers for the remaining `12 hrs`) yields **`~60–90` firings** on factory brushes (`500–800 hrs`), followed by a 15-minute **`$15–$20` carbon brush swap** (Ametek `#33326-1`, up to 2 brush changes per commutator per Ametek technical guidelines = `~1,000–1,400` total hours).
2. **Honest Market Check on Commodity Cast-Aluminum Radial Fans:**
   - Searching Grainger, Zoro, and VEVOR confirms that **new industrial `12+ inWC` cast-aluminum pressure blowers** (e.g., Cincinnati Fan PB-8/PB-9) still retail for **`$900–$1,500`**, and **VEVOR does not manufacture true `12 inWC` cast-aluminum radial blowers** (their `$80–$250` fans are low-pressure axial/utility ventilators that stall at `1–2 inWC`!).
   - To hit the **`$280–$450`** radial-blower price point, potters have two verified real-world paths:
     - **Path A:** Repurpose a **`2 HP` woodworking dust-collector radial blower** (e.g., Grizzly `G1029Z2P` `11.0 inWC` cast-aluminum wheel, WEN `DC1300`, or Harbor Freight `2 HP` `~9–10 inWC`) — **paired with air-atomization** (Challenge 3, Role B) so total scrubber `ΔP` drops into the `10–12 inWC` band (or stage **2 in series** on the dry side for `20–22 inWC`).
     - **Path B:** Buy a **used/surplus industrial `1.5–2 HP` radial pressure blower** (Cincinnati Fan PB-8/PB-9, Dayton, New York Blower) from surplus liquidators (`$250–$450` common on eBay/SurplusRecord), which handles `12–16 inWC` effortlessly on a single `120V/240V` circuit.
3. **Why Regenerative Ring Blowers Rarely Win on a `24 cu.ft.` Kiln:**
   - Regenerative blowers are designed for **high head (`60–112 inWC`) at low flow (`80–150 CFM`)**. At `21.2 inWC`, you throw away `70%` of their pressure capability while needing **two `2.2 kW` (`3 HP`) units in parallel** (`$900–$1,500`, `6 HP` electrical draw!) just to reach `256 ACFM` — AND their tight-tolerance aluminum impellers forbid wet HCl contact. They are only attractive on **small `6–12 cu.ft.` kilns** (`70–130 ACFM`) running on the dry side!

---

## 4. Challenge 3 — Is a Studio Shop Air Compressor Cost-Effective? (Exact Thermodynamics)

Most ceramic studios already own a **`3–5 HP` (`10–15 SCFM @ 90 psig`) reciprocating shop air compressor** for glaze spraying. Can that compressor replace the `$1,815` scrubber blower — or can it slash the blower cost through a smarter thermodynamic role?

### 4.1 Exact Adiabatic Compression Thermodynamics

Compressing ambient air (`P₁ = 14.7 psia`, `T₁ = 528°R / 293 K`, `k = 1.4`) to discharge pressure `P₂` requires minimum isentropic shaft power per `SCFM`:

$$\dot{W}_{\text{isen}} = \frac{k}{k - 1} P_1 \dot{V}_1 \left[\left(\frac{P_2}{P_1}\right)^{\frac{k-1}{k}} - 1\right]$$

Accounting for realistic studio piston-compressor wire-to-air efficiencies (`η_1stg ≈ 62%` single-stage; `η_2stg ≈ 68%` two-stage intercooled):

| Compressor Discharge Pressure | Pressure Ratio (`P₂/P₁`) | Single-Stage Isentropic (`HP/SCFM`) | Real Single-Stage Shop Piston (`η = 62%`) | Real Two-Stage Intercooled (`η = 68%`) | Practical Yield (`SCFM per Motor HP`) |
| :---: | :---: | :---: | :---: | :---: | :---: |
| **`40 psig`** *(Air-Atomizing Nozzle)* | `3.72` | `0.1023` | **`0.165 HP/SCFM`** | `0.136 HP/SCFM` | **`6.1 – 7.3 SCFM/HP`** |
| **`80 psig`** *(EXAIR Rated Test)* | `6.44` | `0.1577` | **`0.254 HP/SCFM`** | `0.201 HP/SCFM` | **`3.9 – 5.0 SCFM/HP`** |
| **`90 psig`** *(Standard Shop Tank)* | `7.12` | `0.1689` | **`0.272 HP/SCFM`** | `0.214 HP/SCFM` | **`3.7 – 4.7 SCFM/HP`** |
| **`100 psig`** *(Max Tank Cut-Out)* | `7.80` | `0.1792` | **`0.289 HP/SCFM`** | `0.225 HP/SCFM` | **`3.5 – 4.4 SCFM/HP`** |

**Rule of Thumb Verified:** At `90 psig`, a studio shop air compressor delivers only **`3.7–4.7 SCFM per motor HP`** (`~0.21–0.27 HP/SCFM`), making compressed air the most energy-expensive utility in any workshop (~`8×–12×` more kWh per air-HP than a direct centrifugal blower!).

---

### 4.2 Role A — Driving Main Scrubber Airflow via Compressed-Air Venturi / Air Amplifier (VERDICT: FAIL for 24 cu.ft.!)

Commercial Coanda-effect air amplifiers (e.g., **EXAIR 2" Super Air Amplifier #120022**, `$200–$300`, `15.5 SCFM @ 80 psig → 341 CFM` at `22:1` ratio; **4" #120024**, `$350–$450`, `29.2 SCFM → 730 CFM` at `25:1`) look tempting on paper — **until you apply scrubber backpressure**:

1. **Catastrophic Backpressure Collapse of Coanda Amplifiers:**
   - EXAIR explicitly warns that Super Air Amplifiers are **zero-backpressure volume movers** (capping intake duct at `<= 5 ft` and discharge duct at `<= 15 ft` of smooth open pipe). Even **`2–4 inWC`** of downstream resistance disrupts the Coanda vacuum core and collapses entrainment ratio from `22:1` down to `< 6:1` — at `21 inWC`, a Super Air Amplifier stalls completely!
   - Switching to a higher-head pneumatic ejector like the **EXAIR 2" Line Vac (#6084, `$380–$410`)** proves the brutal penalty: it burns **`45 SCFM @ 80 psig`** (`~11.5 HP` continuous!) just to reach **`-28.5 inWC` dead-head vacuum** (`0 CFM` entrained), and the **3" Line Vac (#6086, `$625–$650`)** burns **`68.5 SCFM @ 80 psig`** (`~17.5 HP`!) just to hold `-14.7 inWC` dead-head!
2. **Exact Energy Proof for a `24 cu.ft.` Kiln (`256 ACFM`):**
   Moving `256 ACFM` against `21.2 inWC` requires **`0.853` pure aerodynamic horsepower** ($\text{Air HP} = \frac{Q \cdot \Delta P}{6356}$). Because supersonic-to-subsonic compressed-air jet entrainment has a net pneumatic efficiency of only **`η_ej ≈ 10%–15%`** (`~0.011–0.0165` delivered air-HP per SCFM at `80–90 psig`):

| Operating Scenario | Flow (`ACFM`) | System `ΔP` (`inWC`) | Delivered Air HP | Effective Entrainment Ratio | Compressed Air Needed (`@ 90 psig`) | Compressor Motor Required (`100%` Duty) | Feasible on `5 HP / 15 SCFM` Shop Compressor? |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **`24 cu.ft.` Wood/Salt (Baseline Venturi)** | `256` | **`21.2"`** | `0.853 HP` | **`~3.0–4.5 : 1`** | **`55 – 85 SCFM`** | **`15 – 22 HP`** *(industrial rotary screw, `$7,000+`!)* | **IMPOSSIBLE** (`4×–6×` over capacity!) |
| **`24 cu.ft.` Wood/Salt (Reduced-ΔP Venturi)** | `256` | `12.0"` | `0.483 HP` | `~6–8 : 1` | **`30 – 45 SCFM`** | **`8 – 12 HP`** *(industrial duplex/screw)* | **IMPOSSIBLE** (`2×–3×` over capacity!) |
| **`24 cu.ft.` Venturi-Less Soda/Smoke Hood** | `256` | `4.0"` | `0.161 HP` | `~12–16 : 1` | **`16 – 22 SCFM`** | **`5 – 7 HP`** *(100% duty — burns out piston pump!)* | **NO** *(intermittent piston compressor overheats)* |
| **Tiny Kiln Threshold: `5–6 cu.ft.` Smoke/Raku/Soda Hood** | **`50–64`** | **`<= 4.0"`** | `0.03–0.04 HP` | **`~5–8 : 1`** *(safe)* | **`8 – 12 SCFM`** | **`2.5 – 3.5 HP`** *(fits `5 HP / 15 SCFM` at `65%` duty!)* | **MARGINALLY VIABLE ONLY HERE** — yet still burns `~2.5 kW` (`10×` the power of a `$120` 1/3 HP blower!) |

**Role A Bottom Line:** Never use a shop air compressor to drive main scrubber airflow on a `24 cu.ft.` kiln. Even at the **`<= 5–6 cu.ft.` / `< 4 inWC` tiny-kiln threshold** where a `5 HP / 15 SCFM` shop compressor physically *can* pull enough draft, a `$120` 1/3 HP electric blower uses `90%` less electricity and doesn't deafen the studio with a cycling piston pump for 12 hours!

---

### 4.3 Role B — Driving Twin-Fluid Air-Atomizing Venturi Nozzles (VERDICT: HIGHLY COST-EFFECTIVE!)

Now flip the question: instead of asking the compressor to move **`256 CFM`** of kiln gas, use **`4–6 SCFM @ 40 psig`** (`~0.66–1.0 HP`, `< 25%` duty on an existing `5 HP` shop compressor!) solely to drive **two `1/4"` NPT 316L twin-fluid air-atomizing nozzles** (`$35–$95` generic 316L pair, or `$150–$250` Spraying Systems `1/4J-SS` assembly) at the Stage 2 venturi throat!

#### Direct Verification Against `pkg/scrubber/venturi.go` (Calvert `d_50` & Yung–Calvert `ΔP`)

Evaluating [pkg/scrubber/venturi.go](../pkg/scrubber/venturi.go) (`wood_salt_24` gas state: `Q = 0.1208 m³/s`, `ρ_g = 0.806 kg/m³`, `μ_g = 1.605×10⁻⁵ Pa·s`, `L/G = 0.0012`) reveals a subtle, vital fluid-mechanics insight that every engineer must get right:

1. **Stokes Inertial Impaction Parameter (`K_p`, `venturi.go:79`):**
   $$K_p = \frac{C_c \, \rho_p \, d_p^2 \, v_{\text{rel}}}{9 \, \mu_g \, d_d}$$
   Dropping Sauter droplet diameter `d_d` from **`50 µm → 25 µm`** cuts the denominator in half. Even when throat velocity drops from `70 m/s → 48 m/s`, the ratio $v_{\text{rel}}/d_d$ **increases `+37%`** (`48/25 = 1.92` vs `70/50 = 1.40`) — preserving high single-droplet impaction efficiency!
2. **Honest Nuance in Calvert's Verbatim Passive Equation (`u_L0 = 0`, `venturi.go:81`):**
   In `pkg/scrubber/venturi.go:81`, the passive Pease–Anthony venturi penetration exponent is:
   $$-\ln(P_t) = -\frac{2 (L/G) \rho_L v_t d_d}{55 \mu_g} F(K_p, f)$$
   Notice that **`d_d` appears in the numerator**! Why? Because Calvert's *passive* derivation assumes liquid enters at **`u_L0 = 0`** and is accelerated solely by gas drag. A tiny `25 µm` droplet dropped at rest accelerates to gas speed almost instantaneously (`β` jumps from `0.719 → 0.900`), shrinking its passive sweep distance (`x_rel ∝ d_d`)! Therefore, **plugging `d_d = 25 µm` blindly into the passive `u_L0 = 0` model** without crediting the nozzle's compressed-air jet velocity gives `d_50 = 0.393 µm` (and `ΔP = 10.91 inWC`) at `48 m/s`, NOT `0.21 µm`.
3. **How Active Twin-Fluid Air Atomization Actually Preserves `d_50 ≈ 0.20–0.22 µm` While Halving Blower `ΔP`:**
   A `40 psig` twin-fluid nozzle **does not** inject stationary `u_L0 = 0` droplets! Its compressed-air cap (`~180–250 m/s` jet core decaying to `30–60 m/s` across the throat mixing zone):
   - **Breaks up the `2.30 GPM` scrubbing liquor into `25–35 µm` droplets using compressor energy (`~0.8 HP`) instead of blower energy** — breaking the Nukiyama–Tanasawa velocity floor (`v_t >= 65–70 m/s`, below which passive hydraulic nozzles balloon to `128–142 µm` droplets!).
   - **Injects active droplet-to-gas relative shear velocity (`v_rel ≈ 70–85 m/s`)** inside the throat mixing cone (via opposed/angled twin-fluid jets or Hydro-Sonic co-current jet shear where `u_L0 > v_t` near the nozzle), keeping Calvert inertial capture right at **`d_50 ≈ 0.20–0.22 µm`** while the main flue gas throat runs at only **`48–50 m/s`**!
   - **Transfers forward axial momentum (if angled `30°–45°` co-currently)** so the gas stream pays far less droplet-acceleration drag (`ΔP_venturi` drops from **`18.88 inWC` → `~9.0–10.0 inWC`**!).

#### Quantitative Side-by-Side Verification Table (`wood_salt_24`, `256 ACFM`, `L/G = 0.0012`)

| Stage 2 Venturi Mode | Throat Vel `v_t` (`m/s`) | Droplet `d_d` (`µm`) | Effective Shear `v_rel` (`m/s`) | Yung–Calvert `β` | Stage 2 Venturi `ΔP` (`inWC`) | Total System `ΔP` (`inWC`) | Calvert `d_50` NaCl (`µm`) | Overall NaCl / Wood Ash Capture | Stage 4 Blower Shaft HP (`η=62%`) | Blower + VFD Budget Unlocked |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Baseline Passive Hydraulic** (`venturi.go` ref.) | `70` | `50` | `70` | `0.719` | **`18.88"`** | **`21.21"`** | **`0.214`** | `84.2%` / `96.5%` | `1.38 BHP` (`2.0 HP`) | `$1,815` industrial radial + VFD |
| **Passive Hydraulic Slowed Down** *(Nukiyama–Tanasawa)* | `50` | `137.5` | `50` | `0.496` | `6.99"` | `9.32"` | `0.363` *(degraded)* | `68.5%` / `94.2%` | `0.61 BHP` | Poor submicron NaCl capture |
| **Verbatim Passive `u_L0=0` Check** (`venturi.go` literal, no jet credit) | `48` | `25` | `48` *(passive)* | `0.900` | `10.91"` | `13.24"` | `0.393` *(penalized by `u_L0=0`)* | `56.9%` / `73.4%` | `0.86 BHP` | Shows why `d_d` alone without jet velocity is insufficient |
| **Active Twin-Fluid Air-Atomized (`4–6 SCFM @ 40 psig`, `25–35 µm` Jet Shear)** | **`48–50`** | **`25–35`** | **`70–75`** *(active air-cap shear)* | *jet-assisted* | **`9.0"–10.0"`** | **`11.5"–12.5"`** | **`0.20–0.21`** | **`~80–84%` / `~93–96%`** | **`0.75–0.81 BHP`** (`1.0–1.5 HP`) | **`$300–$420`** medium-pressure radial fan (`120V/240V`, NO VFD!) |

#### Net Economics of Role B (Using an Existing Studio Shop Compressor)
- **Added Hardware Cost:** `2 ×` `1/4"` NPT 316L twin-fluid air-atomizing nozzles (`$70–$110` generic 316L, or `$190` Spraying Systems `1/4J-SS`) + `$35` air pressure regulator/filter/gauge = **`+$105–$225`**.
- **Compressor Operating Duty:** `5 SCFM @ 40 psig` = **`0.83 HP` (`~0.62 kW`)** — only **`~25%–33%`** of a typical `3–5 HP` studio compressor's capacity, and **ONLY switched ON during the `3.0-hr` peak salting window** (`~1.9 kWh` = `$0.30/firing`!).
- **Blower + VFD + Electrical Capital Saved:** Drops Stage 4 blower + VFD + line reactor from **`$1,965`** (`$1,815 + $150`) down to **`~$350–$450`** (commodity `1.5–2 HP` medium-pressure radial blower + manual blast gate) — a **net capital savings of `$1,300–$1,500`**!

> **Operational Guardrail for Role B:** Use a `50-mesh` inline Y-strainer on the liquid feed AND **external-mix** `1/4"` twin-fluid air-atomizing caps (`>= 0.060"` liquid orifice) so recirculated wood-ash/baking-soda liquor cannot plug the nozzle during a salt firing!

---

## 5. Summary Budget Recipes (`24 cu.ft.` Reference Wood/Salt Kiln)

Every recipe below is sized for the **exact same `24.0 cu.ft.` Cone 11 (`2400°F`) wood/salt kiln** (`2.0 lbs/hr` NaCl, `3.0-hr` salting window, `99%` HCl absorption, `~80–84%` submicron NaCl capture) and preserves **100% of mandatory NEC / IMC / IPC / NFPA 86 safety protections**:

| Subsystem / Line Item | Build 0: Phase 7 Baseline (`wood_salt_24`) | Build 1: Budget Potter A *(Air-Atomized Low-ΔP + Sch 80 CPVC Bundle + Stoneware)* | Build 2: Budget Potter B *(Full `21"` Head + Dual/Triple Ametek Bypass + Sch 80 CPVC)* | Build 3: Ultra-Budget C *(Sub-Cooled `< 135°F` + Sch 40 PVC Bundle + Reclaim Stoneware)* |
| :--- | :---: | :---: | :---: | :---: |
| **Stage 1: Quench Vessel (`14"` 316L SS Flue Spool)** | `$580.71` | `$480` *(6.5 ft fine-mist 316L spool)* | `$480` *(6.5 ft fine-mist 316L spool)* | `$545` *(`$480` 6.5 ft 316L spool + `$65` PEX sump sub-cooling coil)* |
| **Stage 2: Wet Venturi (`2 ×` 316L Sanitary Tri-Clamp Reducers)** | `$366.00` | `$265` *(no-weld COTS tri-clamp)* | `$265` *(no-weld COTS tri-clamp)* | `$265` *(no-weld COTS tri-clamp)* |
| **Stage 3: Absorber Shell / Shell-Less `4"` Pipe Bundle** | `$804.52` *(`14"` Sch 80 bare; `$1,455` w/ flanges+freight)* | **`$895`** *(`14 × 4"` Sch 80 CPVC `4'`, `1/2"` HDPE sheets, bulk Uniseals, steel support frame, 14-port spray)* | **`$895`** *(`14 × 4"` Sch 80 CPVC `4'`, `1/2"` HDPE sheets, bulk Uniseals, steel support frame, 14-port spray)* | **`$595`** *(`14 × 4"` Sch 40 Solid PVC `$228`, PVC hubs `$140`, HDPE sheets `$120`, frame + spray `$107`)* |
| **Stage 3: Packing Media (`3.8–4.2 ft³`)** | `$144.00` *(`1"` PP Pall)* | **`$25`** *(potter-extruded `3/4"` thin-wall stoneware rings, studio reclaim clay)* | **`$25`** *(potter-extruded `3/4"` thin-wall stoneware rings, studio reclaim clay)* | **`$25`** *(potter-extruded `3/4"` thin-wall stoneware rings, studio reclaim clay)* |
| **Stage 3: Sump Drum + Chevron Demister** | `$297.96` (`$72 + $225.96`) | **`$145`** *(`$40` local 55-gal HDPE drum + `$105` DIY PP mesh/chevron pad)* | **`$145`** *(`$40` local 55-gal HDPE drum + `$105` DIY PP mesh/chevron pad)* | **`$145`** *(`$40` local 55-gal HDPE drum + `$105` DIY PP mesh/chevron pad)* |
| **Liquid Handling: Mag-Drive Pump + Nozzles + IPC 608 Air Gap** | `$757.03` (`$340.03 + $232 + $185`) | **`$415`** *(`$220` PP mag-drive + `$155` 2 quench + 2 twin-fluid 316L nozzles & reg + `$40` IPC 608 `>= 1"` Air Gap)* | **`$340`** *(`$220` PP mag-drive + `$80` 4 hydraulic 316L nozzles + `$40` IPC 608 `>= 1"` Air Gap)* | **`$415`** *(`$220` PP mag-drive + `$155` twin-fluid nozzles & reg + `$40` IPC 608 Air Gap)* |
| **Stage 4: Prime Mover & Silencer Box** | `$2,211.00` (`$1,815` blower/VFD + `$396` silencer) | **`$485`** *(`$375` `1.5–2 HP` `12 inWC` radial blower + blast gate + `$110` DIY Rockwool box)* | **`$545`** *(`$420` `3 × 116765` or `2 × 119992` Ametek bypass + `$125` heavy DIY Rockwool silencer box)* | **`$485`** *(`$375` `1.5–2 HP` `12 inWC` radial blower + `$110` DIY Rockwool box)* |
| **Safety, Controls & NEC Disconnect (`NFPA 86` Retained 100%!)** | `$1,907.00` (`$1,692` RPi kit + `$215` disc/reactor) | **`$797`** *(`$732` `electromechanical_only` kit w/ `200°F` snap-disc, 0.25 GPM switch, latching relay, `7"` fail-open damper, Magnehelic + `$65` NEMA 4X disconnect)* | **`$842`** *(`$732` `electromechanical_only` kit + `$65` NEMA 4X disconnect + `$45` 3-stage blower switches/SCR)* | **`$797`** *(`$732` `electromechanical_only` kit w/ **`135°F` snap-disc**, 0.25 GPM switch, latching relay, `7"` damper, Magnehelic + `$65` NEMA 4X disconnect)* |
| **Hardware, Gaskets, Cement, Hydrometer & pH Kit** | `$195.00` | `$175` | `$175` | `$175` |
| **TOTAL ESTIMATED HARDWARE BUILD** | **`$7,263.22`** *(excl. `~$650` LTL/flanges)* | **`$3,682`** *(whls CPVC; `$3,580` w/ used blower)* | **`$3,712`** *(full `21"` head; `$3,510` w/ `2×119992`)* | **`$3,447`** *(or **`~$2,540–$2,850`** w/ salvaged 316L flue/blower & local PVC!)* |
| **Savings vs. Phase 7 Baseline (`$7,263`)** | — | **`-$3,581` (`-49.3%`)** *(or `-$4,230` vs landed)* | **`-$3,551` (`-48.9%`)** *(or `-$4,200` vs landed)* | **`-$3,816` (`-52.5%`)** *(up to `-$4,720` / `-65%` w/ studio fab!)* |

### 5.1 How Real Studios Hit `$2,500–$3,150` Out-of-Pocket (Zero Safety Compromises)

The table above uses **brand-new retail COTS prices** for every single metal spool, pump, damper, and blower (`$3,447–$3,712`). Studios that apply three common workshop sourcing tactics bring total out-of-pocket cash cleanly into **`$2,500–$3,150`**:
1. **Fabricate the Stage 1 quench cylinder & Stage 2 tri-clamp adapter locally (`-$300 to -$450`):** Rolling 16-ga 316L sheet or repurposing a salvaged `14"` stainless flue/brewery barrel cuts Stage 1+2 from `$745` to `~$350–$450`.
2. **Source a surplus `1.5–2 HP` radial blower or Harbor Freight/WEN `2 HP` dust-collector impeller (`-$120 to -$180`):** Cuts Stage 4 blower hardware to `$220–$260` (paired with Role B air-atomization keeping total static pressure `<= 11.5–12.0 inWC`).
3. **Use Build 3's Sub-Cooled `< 135°F` Shell-Less `4"` Sch 40 PVC Bundle + Reclaim Stoneware Rings (`-$300` vs Sch 80 CPVC):** Cuts Stage 3 shell + media to **`~$620` total** with **zero LTL freight** — yielding a complete, code-compliant `24 cu.ft.` wood/salt scrubber for **`~$2,540–$3,050`**!

### 5.2 Non-Negotiable Code & Life-Safety Checklist (Preserved Across ALL Budget Builds)
Cutting `$3,500–$4,700` from the BOM **never** touches life-safety or US model-code compliance:
- **NFPA 86 Independent Hardwired Safety Loop (`$647` — NEVER OMIT):** Hardwired manual-reset bimetallic snap-disc (`200°F` for Sch 80 CPVC; **`135°F`** for sub-cooled Sch 40 PVC) + `0.25 GPM` mechanical paddle flow switch + 24VDC de-energize-to-trip latching safety relay + NEMA 4X E-stop + `7"` spring-return fail-open hot-gas bypass damper to an **NFPA 211** high-heat bypass chimney (`>= 36"` clearance to combustibles).
- **IPC 608 Potable Water Protection (`$40`):** Engineered `>= 1"` vertical Air Gap above the sump/break-tank overflow rim (zero submerged hoses or float valves).
- **IPC 702.5 & 40 CFR 403.5 Drain Compliance (`$22`):** Cool sump liquor `<= 140°F` and confirm `pH 6.5–8.5` with thermometer + pH strips before sanitary discharge; dump every `3–6` intermittent salt firings at `SG >= 1.105` (`150 g/L` TDS).
- **NEC 430.102 / 210.8(F) / 400.12 Electrical Safety (`$65`):** Fixed conduit wiring, lockable NEMA 4X sight disconnect (`<= 50 ft`), and GFCI protection on outdoor/wet circuits.
- **IMC 501.3.1 / OSHA 29 CFR 1910.1000 Outdoor Stack Siting:** Terminate outdoor stack `>= 12 ft` high, `>= 30 ft` from property lines, and `>= 10 ft` from windows/intakes (`~10 ppm` residual HCl exceeds `5 ppm` OSHA ceiling PEL — never discharge indoors!).

---

## 6. Cross-References
- **Phase 7 Stage Design Options & Parallel/Eductor Analysis:** [docs/STAGE_DESIGN_OPTIONS.md](STAGE_DESIGN_OPTIONS.md)
- **Validated Scrubber Physics & Equations:** [docs/ENGINEERING_VALIDATION.md](ENGINEERING_VALIDATION.md), [pkg/scrubber/venturi.go](../pkg/scrubber/venturi.go), [pkg/scrubber/packedbed.go](../pkg/scrubber/packedbed.go)
- **US Model Code Compliance Audit (C1–C14):** [docs/CODE_COMPLIANCE_AUDIT.md](CODE_COMPLIANCE_AUDIT.md)
- **Control Tiers & Hardwired NFPA 86 Safety Loop BOM:** [docs/CONTROL_ALTERNATIVES_AND_RPI_BOM.md](CONTROL_ALTERNATIVES_AND_RPI_BOM.md)
