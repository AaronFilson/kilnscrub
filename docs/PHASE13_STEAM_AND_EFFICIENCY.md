# Phase 13 Engineering Study: Steam Water Pumps, Kiln/Stage-1 Waste-Heat Steam Generation, Zero-Pump Sump-Pool Scrubbing & Deluxe Eco Upgrades

This document summarizes the first-principles thermodynamics, corrosion metallurgy, urban industrial sourcing (Mumbai/Pune & Nairobi), and safety boundaries for the **Phase 13** options added to `kilnscrub`.

---

## 1. Steam-Driven Water Pumps in Urban Off-Grid Markets (`--local-pump-mode`)

When a zero-electricity ceramic studio already operates a certified `20 psig` (LP) or `60 psig` (HP) steam source for the ejector, steam can also eliminate or reduce manual water pumping:

| Mode Key | Hardware | Min Motive Steam | Handles Hot Alkaline Sump? | Urban Availability (Mumbai / Nairobi) | Budgetary Cost | Steam Load & Engineering Verdict |
|---|---|---|---|---|---|---|
| `manual_treadle` | MoneyMaker treadle / cast-iron semi-rotary wing pump + `3.5 m` Kentank/Sintex | `0 psig` | Yes | Universal | `$45–$85` | Baseline (~6–12 min/hr labor); works on all tiers (`0 / 6 / 20 / 60 psig`). |
| `steam_ppcp` | **Mechanical Pressure-Powered Condensate Pump (PPCP)** — *Forbes Marshall / Spirax Sarco MFP / Thermax* | **≥10–15 psig** *(requires ≥5 psig above `3.5 m` / `5.1 psi` static head)* | **Yes (up to >200°F)** — float-actuated displacement, zero steam-water mixing | **High** — Forbes Marshall factories in Pune/Mumbai & direct branch in Nairobi (Babadogo Rd); Spirax Sarco stockists in Nairobi Industrial Area | `~$680` new *(~$250–$450 tea/laundry salvage)* | **Best technical fit** for 0 W automatic hot-sump recirculation at `20` or `60 psig`. Consumes only **`~0.45–1.0 kg/hr`** steam (<2% of ejector load) and vents spent steam into the scrubber. |
| `steam_jet_syphon` | **Gunmetal/316L Steam-Jet Liquid Syphon + Penberthy-style Bronze Boiler Feed Injector** | **≥20–30 psig** *(optimal `60 psig`)* | **NO for hot sump!** Condenses steam into water; stalls ("breaks") above `120°F (49°C)` and adds `+30–45°F` per pass | High in Mumbai/Pune (IndiaMART/local fabricators); specialty order in Nairobi | `~$245–$295` | **Great low-cost 0 W lift for COLD fresh makeup water (`<120°F`)** and **automatic pressurized boiler feedwater** (pushes cold water into the `60 psig` boiler with zero moving parts). Must NOT pump hot recirculating sump liquor. |
| `sump_pool_no_pump` | **Zero-Pump, Zero-Water-Tower Sump-Pool Self-Induced Scrubber** | **≥20 psig** *(optimal `60 psig`)* | **Yes** — ground-level dual baffled brick sump pools, zero pumps & zero overhead tank! | 100% local masonry + `1/2"` brass/316L dip-tube | **Saves `~$150–$220`** *(eliminates Kentank, 3.5 m timber stand & pump!)* | Uses steam-jet suction to pull gas `0.75"–1.5"` beneath Brick Sump Pool #1 (self-induced S-baffle droplet curtain) while Bernoulli throat vacuum (`~15–35 inWC`) passively aspirates cool alkaline liquor up a `1/2"` dip-tube from Sump Pool #2 into the ejector throat (carburetor atomization -> gravity drain back into the sump!). |

> **Why `6 psig` Subsonic Catering-Steamer Mode Stays Manual:** `3.5 m` of water elevation is `5.1 psi` static head alone, leaving `<0.9 psi` margin at `6.0 psig` — insufficient to actuate a PPCP (`>=5 psi` differential above head), fire a condensing jet syphon (`>=20–30 psig`), or pull a `1.5"` submerged sump-pool curtain (`~1.85 inWC`). Keep the simple MoneyMaker treadle / hand wing pump for `6 psig` mode.

---

## 2. Steam Coil Inside the Kiln Exhaust Stream vs. Stage 1 Water Jacket (`--local-boiler-source`)

### A. Hot NaCl / HCl Metallurgy Limits (Why Bare 316L in `2400°F` Salt Vapor Fails)
In salt-glaze kiln exhaust (`2400°F` / `1316°C`), `HCl` gas reacts with metal oxide scales (`Cr₂O₃`, `Fe₂O₃`) to form volatile metal chlorides (`FeCl₂`, `CrCl₃`), and molten `NaCl` (`mp = 801°C / 1474°F`) fluxes protective passivation layers:

| Material | Max Safe Metal Skin Temp in `NaCl + HCl` Exhaust | Role in `kilnscrub` |
|---|---|---|
| **316L Stainless Steel** | **`< 350–400°C` (`< 660–750°F`)** | Only safe when **continuously water-flooded** (boiling wall `~115–160°C`) or in zero-salt wood/soda firings. Rapidly pits/cracks if dry-fired! |
| **Inconel 625 / Alloy C-276** | **`550–650°C` (`1020–1200°F`)** | Weld-overlay or coil cladding for aggressive salt exhaust heat recovery. |
| **Silicon Carbide (SiC) / Firebrick Muffle** | **`1000–1250°C` (`1830–2280°F`)** | Protective radiation shield separating molten `NaCl`/ash from pressurized water tubes. |

### B. Three Boiler Heat-Source Architectures
1. **`dedicated_hearth` (Baseline):** Certified steamer/boiler fired externally on wood/LPG or indirect flue tap; zero flue corrosion risk.
2. **`kiln_exhaust_coil` (SiC/Firebrick-Muffled Flue Steam Coil):** Pressurized water-flooded coil behind a thin SiC / split-firebrick radiation muffle (or Alloy 625 clad). Recovers up to `~72%` of available flue waste heat (`13–39 kW`), eliminating separate boiler fuel during high fire and cutting Stage 1 evaporative quench water loss by **`~32%`**. **CRITICAL:** Never dry-fire; requires certified PRV, check valve, and continuous water flood so tube skin stays `<215°C`.
3. **`stage1_water_jacket` (Stage 1 Water-Jacketed Dry-Wall Quench Collar — Recommended Waste-Heat Upgrade):**
   - **Thermodynamic Reality:** Once Stage 1 sprays adiabatically quench gas to `184°F (84°C)`, the gas is *below* `212°F (100°C)` and **cannot boil water**.
   - **Engineering Solution:** Place a **double-walled pressurized water jacket around the upper dry-wall entry collar of Stage 1 (`0.6–0.9 m` *above* the spray nozzles)** where `1400–2400°F` gas enters. Boiling water inside the jacket locks the smooth vertical inner wall at **`~120–160°C (250–320°F)`** — **safely ABOVE the `~84°C` `HCl` acid dew point (zero wet acid pitting) yet well BELOW `400°C` hot chloride corrosion** — while shedding fly-ash/salt crust far more cleanly than tight coils!
   - **Payoff:** Simultaneously generates `100%` of ejector steam (`13–39 kW`) with zero extra wood fuel AND cuts Stage 1 evaporative quench water consumption by **`~42%`**.

---

## 3. Custom Fabrication Deluxe Energy & Water Saving Upgrades

Three optional turnkey upgrades for institutional / production studios running `build_path = "custom_fab_deluxe"`:

| Upgrade Flag / Option | CapEx Adder | Water Savings | Energy / Heat Savings | Engineering Mechanism |
|---|---|---|---|---|
| **Closed-Loop Condensing Plume Abater & Water Reclaimer** (`--deluxe-condensing-reclaimer`) | `+$1,850` | **`~70%` of evaporated Stage 1 water reclaimed** (~140–220 gal/firing) | Reclaims **`~25–50 kW`** low-grade studio hydronic heat; cuts winter plume **`>75%`** (eliminating `120V` dilution blower) | Indirect 316L/Titanium finned condensing coil cools saturated `184°F` Stage 3 exhaust down to `~120°F`, returning pure distilled condensate straight to the sump. |
| **Auto Variable-Throat Venturi + VFD Turndown** (`--deluxe-variable-throat-vfd`) | `+$1,450` | — | **`~55–68%` blower electricity (`kWh`) saved** + **`-10 dBA`** off-peak noise | 24VDC actuated flooded-disc/damper throat widens during non-salting ramp & cooldown (`75%` of firing), slashing venturi ΔP (`~19 → ~3.8 inWC`) so VFD throttles to `~48%` RPM (`P ∝ N³`), then locks to `70 m/s` during salting. |
| **Stage 1 Water-Jacketed Quench Pre-Cooler** (`--deluxe-stage1-heat-recovery`) | `+$1,650` | **`~40%` less Stage 1 evaporative water loss** | Reclaims **`~18–35 kW`** studio hot-water/hydronic heat + **`~7%`** blower power savings | Double-walled Alloy C-276 / 316L pressurized cooling collar on the Stage 1 hot inlet extracts sensible heat *before* spray nozzles, reducing vapor volume and makeup water demand. |
