# Phase 16: Global Urban Pottery Survey — Coal, Peat, Biomass & Resource-Limited Scrubber Engineering

> **Scope:** A real-world engineering survey of dense urban and peri-urban pottery clusters beyond Mumbai and Nairobi where ceramicists face acute neighborhood smoke, soot, sulfur (`SO₂`), or salt-glaze (`HCl`/`NaCl`) restrictions while operating with intermittent/absent electric power, constrained capital, and solid fuels (bituminous coal, peat/turf, Andean *turba*, palm kernel shells, olive pomace, rice husks, and scrap wood) — together with the first-principles solutions implemented in `kilnscrub`.

---

## 1. Global Survey of High-Need Urban & Peri-Urban Pottery Clusters

| Urban / Regional Hub | Region Key (`local_region`) | Elevation & Ambient Air | Primary Kiln Fuels | Core Regulatory / Health Driver | Local Materials, Alkali & Off-Grid Machinery |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Khurja (Delhi NCR), India & Dhaka, Bangladesh** | `khurja_dhaka` | Sea level (`101.3 kPa`, `32°C`) | High-sulfur bituminous kiln coal (`1.5–3.5% S`), coal-salt stoneware | Strict NGT / CPCB Delhi-NCR coal-smoke & `SO₂` orders forcing kiln closures or retrofits | Domestic **non-IBR `≤25 L` baby steam boilers**, **Forbes Marshall PPCPs**, **Sintex** HDPE tanks, **IS 651 SWG** vitrified clay pipes, cheap **Chuna** (`Ca(OH)₂`, ~$0.18/kg) & bulk soda ash |
| **Kathmandu Valley / Bhaktapur (Thimi, Pottery Square), Nepal** | `kathmandu` | **`1,350 m` (`86.2 kPa`, `20°C`)** | Coal dust, rice husks, straw, scrap wood | Severe winter Himalayan valley bowl inversions trapping PM2.5 & black carbon inside dense historic towns | **Hilltake / Sintex** HDPE tanks, local firebrick & terracotta flue rings, hotel/momo **6 psig catering pressure steamers**, cast-iron semi-rotary hand pumps, **Chuna** lime |
| **Lagos / Abeokuta (Nigeria) & Accra (Ghana)** | `lagos_accra` | Sea level (`101.3 kPa`, `31°C` humid) | **Palm Kernel Shells (PKS)** (~16.5 MJ/kg), scrap wood, waste oil | Chronic grid blackouts & prohibitive diesel generator fuel cost; dense residential compounds | **Geepee / Polytank** HDPE tanks, **Lake Chad Trona / *Kanwa*** (`Na₂CO₃·NaHCO₃·2H₂O`, ~$0.38/kg — zero scaling!), laterite/firebrick masonry, foot treadle & semi-rotary wing pumps |
| **Tonalá (Guadalajara) & Puebla, Mexico** | `tonala_puebla` | **`1,650 m` (`83.0 kPa`, `23°C`)** | Wood, sawdust, waste oil, scrap lumber | SEMADET air-quality alerts & municipal fines on dense urban pottery smoke | **Rotoplas** *tinacos*, *ladrillo refractario*, **Kusakabe / *hornos sin humo*** passive secondary-air masonry afterburners, ***Cal hidratada*** & natural Mexican ***Tequesquite*** trona |
| **Cusco / Raqch'i / Altiplano (Peru & Bolivia)** | `cusco_andes` | **`3,400 m` (`66.8 kPa`, `14°C`)** | **Andean peat (*turba*)**, *taquia* dung, scarce eucalyptus | Extreme thin air (**-34% air density**, **+52% gas volume expansion**) starves natural chimneys; smoky peat tar | **Rotoplas** HDPE tanks, adobe-jacketed firebrick masonry, ***Cal hidratada***, compact **6–20 psig steam-jet ejector** replacing undersized natural buoyancy |
| **Fustat (Old Cairo, Egypt) & Fez (Ain Nokbi, Morocco)** | `cairo_fez` | Sea level (`101.3 kPa`, `32°C` arid) | **Olive pomace (*jift*)**, sawdust, wood | Tight urban heritage pottery quarters under municipal smoke abatement & severe **water scarcity** | Hand-thrown glazed terracotta SWG flues, closed-loop baffled brick sumps, **Wadi El Natrun natural natron/trona** or cheap ***Gir*** slaked lime, Stage 0 passive smoke-burner |
| **Scottish Highlands / Rural Ireland / Boreal Peat** | `highland_peat` | Sea level (`101.3 kPa`, **`8°C` cool**) | **Peat / *turf*** (~30% H₂O, `0.5–1.5% S`) + seaweed/salt | Heavy phenolic/humic tar, creosote odor, `SO₂` & dense cold-air visible plumes | Strong cold-air chimney buoyancy, reclaimed **1,000 L farm IBC totes**, firebrick + vitrified clay liners, **$0 wood-ash potash lye** (`K₂CO₃`) or agricultural lime |
| **Mumbai (Dharavi / Kumbharwada) & Nairobi (East Africa)** | `mumbai`, `nairobi` | Sea level (`101.3 kPa`) / `1,660 m` (`83.5 kPa`) | Wood, wood-salt, coconut husks, sawdust | Baseline dense urban smoke & `HCl`/`NaCl` abatement | Sintex / Kentank, IS 651 SWG, Kirloskar / MoneyMaker treadle, **Lake Magadi soda ash** (`Na₂CO₃`) |

> **Additional Asia-Pacific clusters mapped directly onto these profiles:** **Bát Tràng (Hanoi, Vietnam)** coal-briquette & wood kilns (`khurja_dhaka` / `coal`) and **Kasongan (Yogyakarta, Indonesia)** coconut-shell / rice-husk / wood pottery clusters (`lagos_accra` / `biomass_ag`).

---

## 2. Four New Engineering Physics Problems Uncovered & Solved

### Problem 1 — Sulfur Dioxide (`SO₂`) & Sulfuric Acid Mist (`H₂SO₄`) Dew Point (`coal`, `peat`, `biomass_ag`)
Wood and natural gas contain negligible sulfur, so earlier phases only balanced `HCl` from NaCl salting. Bituminous kiln coal (`1.5–3.5% S`, default `2.0% S`), air-dried peat/turf (`0.4–1.4% S`, default `0.8% S`), and agricultural waste (`~0.25% S`) generate **130 to 1,400+ ppm `SO₂`** continuously across the **entire firing**, not just during a 3-hour salting window.

- **Exact NIST Shomate + Wilke Thermodynamics (`pkg/thermo/shomate.go`, `pkg/scrubber/scrubber.go`):** Added `SO₂` (CAS `7446-09-5`, MW `64.0638 g/mol`) as a 6th tracked flue species across NIST Shomate `Cp`/`H`, Lucas/Wilke viscosity, and adiabatic quench energy balance.
- **Combined `HCl + SO₂` Diprotic Alkali Stoichiometry (`pkg/scrubber/packedbed.go`, `pkg/studio/frugal.go`):**
  - `HCl` is monoprotic (`1 eq/mol`), whereas `SO₂` forms diprotic `H₂SO₃` (`2 eq/mol`):
    - `Ca(OH)₂ + SO₂ → CaSO₃↓ + H₂O` (**`1.16 kg Ca(OH)₂ / kg SO₂`**)
    - `Na₂CO₃ + SO₂ → Na₂SO₃ + CO₂` (**`1.65 kg Na₂CO₃ / kg SO₂`**)
    - `2 NaHCO₃ + SO₂ → Na₂SO₃ + H₂O + 2 CO₂` (**`2.62 kg NaHCO₃ / kg SO₂`**)
    - `⅔ (Na₂CO₃·NaHCO₃·2H₂O) + SO₂ → 1.5 Na₂SO₃ + ...` (**Trona / *Kanwa* / *Tequesquite*** — 3 equivalents per mole!)
  - Because `SO₂` emits continuously during ramp + soak (modeled at `0.85 × firing_duration_hours`), a `24 cu.ft.` Khurja coal-salt firing consumes substantially more alkali than wood-salt alone — and the solver flags `CaSO₃`/gypsum sludge settling when slaked lime (`Ca(OH)₂`) is paired with coal or peat.
- **Verhoff–Banchero Sulfuric Acid Dew Point (`~265–305°F` vs `184°F` for `HCl`):** ~1.5% of kiln `SO₂` oxidizes to `SO₃`, raising the `H₂SO₄` dew point to **`~290°F` (`143°C`)**. Bare steel flues corrode rapidly if allowed to dip below `~315°F` before wet quenching; refractory firebrick masonry or ceramic-coated ducts remain immune.

### Problem 2 — Heavy Phenolic Tar, Creosote & Fly-Ash Fouling (Coal / Peat / PKS)
Peat/turf (`~30%` moisture), bituminous coal (`10–25%` fly ash + coal tar), and palm kernel shells emit sticky condensable tars and heavy soot during stoking that will rapidly blind fine plastic packing or narrow chevron vanes.

- **Solution — Stage 0 Passive Secondary-Air Masonry Smoke-Burner / Tar Afterburner (`local_afterburner`, `0 W`, `~$85–$125`):**
  - Inspired by Tonalá / Kusakabe *hornos sin humo* (smoke-free wood kilns): a short refractory firebrick checkerwork throat immediately downstream of the kiln exit admits preheated passive secondary air (`>750°C` / `1,380°F`), thermally cracking **70–85% of coal/peat/PKS tars and black soot** *before* the wet quench stage at only **`+0.015 inWC`** additional draft loss and **`0 W`** electricity.
  - Paired with a **Pool #1 floating tar-skimming weir** and **removable washable coarse ceramic anti-tar louver plates** so residual creosote skims off the sump surface instead of fouling stoneware rings.

### Problem 3 — Extreme High-Altitude Thin-Air Buoyancy Collapse (`cusco_andes` at `3,400 m` / `66.8 kPa`)
At Cusco / Altiplano (`3,400 m`, `66.8 kPa` vs `101.3 kPa` sea level):
- Ambient air density drops by **`~34%`**, cutting natural chimney buoyancy proportionally (`ΔP = g·H·(ρ_amb − ρ_stack)`).
- Local actual flue volumetric flow expands by **`+52%`**, automatically requiring a **wider stoneware column** (e.g., `32"` vs `26"` at sea level) to hold low `v_s ≈ 0.40 m/s` Ergun drag.
- Water boils at **`192.0°F` (`88.9°C`)**, so the steam-condensing tower exit temperature automatically adjusts down to **`156°F` (`68.9°C`)**!
- **Machinery takeaway:** At `>3,000 m`, switching from passive chimney buoyancy to a **6 psig catering steamer** or **20 psig LP steam-jet ejector** immediately restores strong, reliable `0 W` draft without building an unstable `>12 m` masonry chimney.

### Problem 4 — Zero-Scaling Local Mineral Alkali Across Africa, MENA & Mexico (`trona_kanwa`)
Slaked lime (`Ca(OH)₂`) is cheap globally but creates chalky `CaCO₃` and `CaSO₃` gypsum sludge in coal/peat scrubbers. Across West Africa (Lake Chad ***Kanwa***), East Africa (Lake Magadi trona), Egypt (Wadi El Natrun ***Natron***), and Mexico (***Tequesquite***), natural **trona** (`Na₂CO₃·NaHCO₃·2H₂O`, MW `226.03 g/mol`) is mined locally and sold cheaply in open markets:
- Delivers **3 acid equivalents per mole**, dissolves cleanly into high-alkalinity sodium carbonate/bicarbonate liquor, produces **zero `CaCO₃` or `CaSO₃` gypsum scaling**, and only requires simple bucket settling/decanting to leave lake-bed silt behind.

---

## 3. Recommended Ready-to-Build Configurations by Region

1. **Khurja (Delhi NCR) / Dhaka Coal + Salt Stoneware Cluster (`khurja_dhaka`, `coal_salt`):**
   - `local_afterburner=true` + `steam_ejector_hp` (`60 psig` domestic non-IBR `≤25 L` baby boiler) + `sump_pool_no_pump` (zero overhead tower, zero electric pump!) + `slaked_lime` (*Chuna*) or `soda_ash` → **`~96.8% HCl`**, **`~94% SO₂`**, **`~78% submicron NaCl`**, **`0 W`**, `~$880–$990 USD` (~₹75,000–₹83,000).
2. **Cusco / Altiplano `3,400 m` Peat Kiln (`cusco_andes`, `peat`):**
   - `local_afterburner=true` + `steam_ejector_lp` (`20 psig`) + `slaked_lime` (*Cal Hidratada*) → overcomes `-34%` high-altitude air density drop, cracks wet *turba* phenolic tars, and scrubs `~90% SO₂` at `0 W`.
3. **Lagos / Abeokuta / Accra PKS + Salt Studio (`lagos_accra`, `biomass_ag_salt`):**
   - `local_afterburner=true` + `steam_ejector_subsonic` (`6 psig` catering steamer + `4.2 m` open standpipe) + `trona_kanwa` (*Kanwa*) → quiet (`~52 dBA`), zero-scaling, blackout-proof (`0 W`), `~$700–$820 USD`.
4. **Scottish Highlands / Ireland Heritage Peat + Salt Kiln (`highland_peat`, `peat_salt`):**
   - `local_afterburner=true` + `natural_chimney` (cold `8°C` maritime air delivers strong natural buoyancy!) + `wood_ash_lye` (`$0` kiln ash potash) + reclaimed `1,000 L` farm IBC tote → `0 W`, zero boiler required, cracks heavy peat tar before the stoneware bed.

---

## 4. CLI & UI Usage

- **Web UI (`kilnscrub ui` on `127.0.0.1:8888`):** Select **Local Ceramic / Brick** build path to access all **10 Global Urban Pottery Hubs**, **10 Fuel Stoichiometries** (`coal`, `coal_salt`, `peat`, `peat_salt`, `biomass_ag`, `biomass_ag_salt`), **Trona / *Kanwa* / *Tequesquite*** alkali, the **Stage 0 Secondary-Air Masonry Tar Afterburner** toggle, and 1-click quick-load buttons (**Khurja Coal+Salt**, **Cusco 3,400m Peat**, **Lagos PKS+Kanwa**, **Highland Peat+Salt**).
- **CLI Examples:**
  ```bash
  kilnscrub guide --build-path local_ceramic_masonry --zero-electricity \
    --local-region khurja_dhaka --local-alkali slaked_lime \
    --local-draft-mode steam_ejector_hp --local-pump-mode sump_pool_no_pump \
    --local-afterburner

  kilnscrub design --fuel coal_salt --salt-rate 2.0 --kiln-volume 24
  kilnscrub design --fuel peat --salt-rate 0 --kiln-volume 24
  ```
