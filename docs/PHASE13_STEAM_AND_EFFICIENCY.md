# Phase 13 & 14 Engineering Study: Steam Water Pumps, Kiln/Stage-1 Waste-Heat Steam Generation, High-Capture Zero-Pump Sump-Pool Scrubbing & Ceramic-Lined Deluxe Eco Upgrades

This document summarizes the first-principles thermodynamics, corrosion metallurgy, urban industrial sourcing (Mumbai/Pune & Nairobi), patent freedom-to-operate (FTO) clearance, and ASME/IBR safety boundaries for the **Phase 13 / Phase 14** options in `kilnscrub`.

---

## 1. Steam-Driven Water Pumps & High-Capture Zero-Pump Sump Pools (`--local-pump-mode`)

When a zero-electricity ceramic studio already operates a certified `20 psig` (LP) or `60 psig` (HP) steam source for the ejector, steam can eliminate manual water pumping (and electric pumps draw **`0 W`** even if `zero_electricity` is toggled off):

| Mode Key | Hardware | Min Motive Steam | Handles Hot Alkaline Sump? | Urban Availability (Mumbai / Nairobi) | Budgetary Cost | Steam Load & Engineering Verdict |
|---|---|---|---|---|---|---|
| `manual_treadle` | MoneyMaker treadle / cast-iron semi-rotary wing pump + `3.5 m` Kentank/Sintex | `0 psig` | Yes | Universal | `$45–$85` | Baseline (~6–12 min/hr labor); works on all tiers (`0 / 6 / 20 / 60 psig`). |
| `steam_ppcp` | **Mechanical Pressure-Powered Condensate Pump (PPCP)** — *Forbes Marshall / Spirax Sarco MFP / Thermax* | **≥20 psig** *(requires ≥5 psig above `3.5 m` / `5.1 psi` static head)* | **Yes (up to >200°F)** — float-actuated displacement, zero steam-water mixing | **High** — Forbes Marshall factories in Pune/Mumbai & branch in Nairobi (Babadogo Rd); Spirax Sarco stockists in Nairobi Industrial Area | `~$680` new *(~$250–$450 tea/laundry salvage)* | **Best technical fit** for `0 W` automatic hot-sump recirculation at `20` or `60 psig`. Consumes only **`~0.45–1.0 kg/hr`** steam (<2% of ejector load) and vents spent steam into the scrubber. |
| `steam_jet_syphon` | **Gunmetal/316L Steam-Jet Cold Makeup Syphon + Penberthy-style Bronze Boiler Feed Injector** | **≥20 psig** *(optimal `60 psig`)* | **NO for hot sump!** Condenses steam into water; stalls ("breaks") above `120°F (49°C)` and adds `+30–45°F` per pass | High in Mumbai/Pune (IndiaMART/local fabricators); specialty order in Nairobi | `~$245–$295` | **Low-cost `0 W` lift strictly from a dedicated COLD (`<120°F`) fresh makeup drum** and **automatic pressurized boiler feedwater**. Must NOT pump hot recirculating sump liquor. |
| `sump_pool_no_pump` | **High-Capture Zero-Pump Flooded Sump-Pool + `1.4 m` Co-Current Drenched Stoneware Trough + 2-Stage Ceramic Sieve Tray** | **Strictly `60 psig` HP** *(requires `~4.3 inWC` ejector draft; rejected at `20 psig` LP)* | **Yes** — dual ground-level baffled brick sump pools (~95 gal / 360 L), zero pumps & zero `3.5 m` water tower! | 100% local masonry + fired ceramic sieve plates | **Saves `~$120–$170`** *(eliminates Kentank, 3.5 m braced stand & pump!)* | **Solved for High Capture (`~96.8%` HCl, `~96.0%` coarse ash, `~78%` submicron NaCl, `0 W`):** Wide `2×` barometric surge plenum + shallow `0.50"` S-vane skimmer (`0.65 inWC`, damps kiln spyhole pulsing) → **gravity-flooded sump-level `60 psig` ejector at Pool #2 waterline (`0"` suction lift — zero dip-tube lift!)** → `1.4 m` co-current drenched stoneware Raschig-ring trough (`0.15 inWC`) → 2-stage shallow (`0.5" + 0.5"`) perforated ceramic sieve tray (`1.15 inWC`) → total required draft **`~2.02 inWC` vs `~4.30 inWC` available (`+2.28 inWC` margin)**. |

---

## 2. Boiler Heat-Source Options & Metallurgy (`--local-boiler-source`)

In salt-glaze kiln exhaust (`2400°F` / `1316°C`), `HCl` reacts with bare metal oxides (`Cr₂O₃`, `Fe₂O₃`) to form volatile metal chlorides (`FeCl₂`, `CrCl₃`), and molten `NaCl` (`mp = 801°C / 1474°F`) fluxes passivation layers. `kilnscrub` offers **four** strictly code-checked heat-source modes:

1. **`dedicated_hearth` (Baseline — `6 / 20 / 60 psig`):** Certified steamer/boiler fired externally on wood/LPG; zero flue corrosion risk.
2. **`stage1_feedwater_preheater` (`0 psig` Vented Atmospheric Stage 1 Economizer — Safe on `6 / 20 / 60 psig`):**
   - Unpressurized `0 psig` vented copper/316L economizer coil wrapped on Stage 1's dry upper exterior preheating cold boiler feedwater from `~22°C → 85°C` before the boiler feed pump (`+$140`).
   - **Zero pressure-vessel code risk** (open vented `0 psig`), saves **`~13%` boiler fuel** and **`~6–10%` Stage 1 evaporative quench water**.
3. **`stage1_water_jacket` (`ASME/IBR-Certified ≤25 L` Shop-Fabricated Stage 1 Water-Jacket Boiler — `20 / 60 psig` ONLY):**
   - Double-walled annular pressurized jacket around the dry upper Stage 1 collar (*above* spray nozzles) locks the inner wall at **`~120–160°C`** (above the `~84°C` HCl dew point yet below `400°C` hot chloride attack) and **replaces the standalone baby boiler** (no double billing; saves energy-conserved **`~18–38%` Stage 1 evap water**).
   - **MANDATORY ASME Section VIII UG-28 Rule:** Because annular pressure puts the inner flue tube under *external collapsing pressure*, this jacket MUST be shop-fabricated by an ASME/IBR/KEBS vessel shop with **UG-28 external-pressure stiffener rings**, `≤25 L` water volume, stamped PRV, armored sight glass, and low-water fusible plug (NEVER DIY field-welded!).
4. **`kiln_exhaust_coil` (SiC/Firebrick-Muffled `<2 L` Monotube Flue Coil — Strictly `20 psig` LP ONLY):**
   - Seamless `<2 L` water-flooded monotube behind a SiC / split-firebrick radiation muffle (or Alloy 625 clad), **replacing the standalone LP firebox** (no double billing; **banned at `60 psig`** per the Certified-Baby-Boiler-Only rule!).

---

## 3. Custom Fabrication Deluxe Energy & Water Saving Upgrades

Three optional duty-scaled upgrades for institutional / production studios running `build_path = "custom_fab_deluxe"` (all propagated into COTS BOM, ShopPlan schedules, Recirculation, and Urban Plume/Water checks):

| Upgrade Flag / Option | Duty-Scaled CapEx (24 cu.ft. Ref) | Water Savings | Energy / Heat Savings | Engineering & Material Spec |
|---|---|---|---|---|
| **Ceramic-Coated / Sintered-SiC Condensing Plume Abater & Acid-Condensate Reclaimer** (`--deluxe-condensing-reclaimer`) | `~$2,550–$2,750` (`$1,450 + $24/kW`) | Reclaims **`~52–65%`** of remaining Stage 1 evap water (coupled with Stage 1 jacket: **`~70–78%`** total!) | Exact IAPWS-95 psychrometric condensing duty **`~45–55 kW`** (`184°F → 150°F`); shrinks winter plume reach cleanly in `Urban.Plume` | **Vitreous ceramic-enamel coated steel / sintered SiC condensing tubes in a ceramic-lined Derakane FRP shell** (**ZERO bare 316L SS** in warm `pH 2.5–3.5` chloride-bearing HCl condensate!) + **outdoor closed-loop dry-cooler** heat sink + **Sch 80 CPVC** condensate trap draining directly into the alkaline neutralizing sump. |
| **Auto Variable-Throat Venturi + VFD Turndown** (`--deluxe-variable-throat-vfd`) | `~$1,460` (`$1,100 + $180/HP`) | — | **`~55–68%` blower electricity (`kWh`) saved** + **`-10 dBA`** off-peak noise | Strictly requires `inline_id_blower` + `single_throat` + VFD drive. 24VDC Alloy C-276 / ceramic-coated flooded-disc plug widens throat during non-salting hours (`ΔP ~19 → ~3.8 inWC`, VFD `~48%` RPM). |
| **Stage 1 Water-Jacketed Quench Pre-Cooler** (`--deluxe-stage1-heat-recovery`) | `~$1,765` (`$1,150 + $22/kW`) | **`~30–40%`** less Stage 1 evap water loss | Reclaims **`~18–28 kW`** studio hydronic heat + **`~6%`** blower power savings | Mounted strictly *below* hot inlet flue nozzle; includes **ASME 30 psig relief valve, expansion tank, and hydronic flow switch wired into the E3 NFPA 86 fail-safe bypass loop** to prevent dry-out boiling during a pump outage. |

---

## 4. Patent & Prior-Art Freedom-to-Operate (FTO) Summary (Phase 14)

- **Public Domain / Expired Foundations:** Steam-jet / Hydro-Sonic scrubbers (`US3852408A`, `US3852409A`, `US3912469A`, `US4141701A`), condensation nucleation scrubbing (`US4401444A`, `US5176723A`, `US3803805A`), Rotoclone submerged-orifice impingers (`US3104961A`, `US4300924A`), variable-throat venturi scrubbers (`US3638925A`, `US3898308A`, `US4023942A`, `US4144041A`), and condensing plume abatement (`US3619983A`, `US8721771B2` fee-expired) are all expired / public domain.
- **Active Patent Clearance:** `US7842264B2` (Ewan et al., active ~Apr 2029) claims *multi-stage* aerodynamic reactors/aero-coalescers; `kilnscrub` cleanly avoids it by pairing a single-stage *de Laval* steam-jet ejector with classical ceramic Raschig-ring packing and perforated ceramic sieve trays.
