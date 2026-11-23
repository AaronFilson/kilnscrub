# Independent Chemical Engineering & Physics Validation Report (`kilnscrub`)

**Scope:** Independent first-principles audit of the thermophysical, aerodynamic, mass-transfer, hydraulic, and materials-engineering models implemented in `pkg/model/`, `pkg/thermo/`, `pkg/scrubber/`, and `scripts/generate_oracle.py`, cross-checked against `testdata/oracle_fixtures.json`, peer-reviewed chemical engineering literature, and proven industrial/studio wet scrubber designs.

**Reference Baseline Audited:** Default 24 cu.ft. wood/salt studio ceramic kiln (`PeakExitTempF = 2400°F`, `SaltRateLbsPerHr = 2.0 lbs/hr`, `ExcessAirFraction = 0.30`, `VenturiThroatVelocity = 70 m/s`, `VenturiLGRatio = 0.0012`, `TargetHClRemoval = 0.99`, `1" PP Pall Rings`, `TargetKilnDraftInWC = -0.05 inWC`), plus parametric verification across `12`, `24`, and `48 cu.ft.` kilns and all four fuel modes (`wood_salt`, `wood`, `gas_salt`, `gas`).

---

## Executive Summary

The core thermophysical kernel in `pkg/thermo/` (NIST 7-coefficient Shomate $C_p/H$, Lucas gas viscosity with Wilke mixing, IAPWS-95 water vapor pressure, and Brent adiabatic saturation closure) is **mathematically rigorous, conservation-exact (energy residual $< 10^{-4}\text{ W}$), and matches peer-reviewed Python oracles (`chemicals` 1.5.2, `fluids` 1.3.1) to within $< 10^{-6}$ relative error**. Likewise, the Calvert (1970, 1972) cut-diameter integral (`CalvertF` small-$x$ Taylor expansion + `CalvertCutDiameter`), Eckert GPDC flooding polynomial, Haaland Darcy-Weisbach duct friction, and static-pressure accounting (no demister double-counting, clean signed draft convention) are numerically well-engineered.

However, this independent audit identified **four Category A code/physics defects** and several **critical Category B engineering & materials-safety limitations** that materially affect blower sizing, pressure drop prediction, active control behavior, and corrosion survival in real salt-kiln service:

| Severity | Finding | Exact Location | Quantitative Impact (24 cu.ft. Ref. Case) |
| :--- | :--- | :--- | :--- |
| **Category A (Physics Bug)** | **Packed-bed Ergun $\Delta P$ uses nominal ring size ($25.4\text{ mm}$) instead of Sauter equivalent sphere diameter ($2.90\text{ mm}$) and omits irrigated holdup** | `pkg/scrubber/packedbed.go:183` | Predicts **$0.094\text{ inWC}$ ($23.5\text{ Pa}$)** vs. **$0.85\text{ inWC}$** (dry Sauter Ergun) and **$1.20\text{–}2.07\text{ inWC}$ ($300\text{–}514\text{ Pa}$)** irrigated GPDC — a **13×–22× underprediction** of bed $\Delta P$. |
| **Category A (Physics Bug)** | **Venturi acceleration $\Delta P$ conflates Calvert's impaction factor $f = 0.5$ with droplet momentum acceleration fraction $\beta$** | `pkg/scrubber/venturi.go:303` | Predicts **$11.78\text{ inWC}$** acceleration ($13.76\text{ inWC}$ total) vs. **$16.77\text{–}16.90\text{ inWC}$** (Boll 1973 ODE / Yung–Calvert 1977 for $d_d=50\,\mu\text{m}$, $L_t=80\text{ mm}$) and **$23.56\text{ inWC}$** (Calvert 1970 $\beta=1.0$). Combined with packed-bed error, realistic total system $\Delta P$ is **$20.5\text{–}28.5\text{ inWC}$** vs. coded **$14.75\text{ inWC}$** ($1.35\text{–}1.85\text{ shaft HP}$ vs. $0.96\text{ HP}$). |
| **Category A (Thermo/Control Bug)** | **Default PID `QuenchSetpoint = 165.0°F` is below the thermodynamic adiabatic saturation floor ($183.9°F$ at $2400°F$; $177.5°F$ at $1800°F$)** | `pkg/model/types.go:585` (`internal/control/sim_hal.go:54`) | Impossible for an adiabatic spray quench to reach $165°F$ without large cold once-through subcooling; real hardware PID spray valve will saturate permanently at 100% during high-fire. |
| **Category A (Minor Constant)** | **`refPeakDeltaTK` uses `1588.15 K` ($2399.0°F$) instead of `1588.71 K` ($2400.0°F$)** | `pkg/scrubber/scrubber.go:26` | $+0.043\%$ thermal scaling bias at $2400°F$ (negligible engineering impact, trivial fix). |
| **Category B (Critical Materials Safety)** | **316L Stainless Steel in hot wet HCl quench/venturi ($183.9°F$, $1047\text{–}2055\text{ ppm HCl}$)** | `pkg/thermo/quench.go:118`, `pkg/scrubber/venturi.go:16`, `pkg/scrubber/blower.go:18` | 316L SS suffers rapid passive-film breakdown, chloride pitting, and Cl-SCC above $140°F$ in acidic wet chlorides; requires **Alloy C-276 / refractory-lined quench / vinyl-ester FRP** or alkaline-buffered quench spray with sacrificial inspection. |

---

## 1. Parametric Benchmark Outputs (`kilnscrub design`)

Read-only execution of the compiled `kilnscrub` binary across studio kiln sizes (`12`, `24`, `48 cu.ft.`) and fuel stoichiometries (`wood_salt`, `wood`, `gas_salt`, `gas`) at `2400°F`, `30%` excess air, and `v_throat = 70 m/s`, `L/G = 0.0012` (`8.98 gal/1000 acf`) yields the baseline dataset audited below:

| Parameter | 12 cu.ft. `wood_salt` | **24 cu.ft. `wood_salt` (Ref.)** | 48 cu.ft. `wood_salt` | 24 cu.ft. `wood` (0 salt) | 24 cu.ft. `gas_salt` | 24 cu.ft. `gas` (0 salt) |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| Firing Load (LHV) [kW] | 64.8 | **129.6** | 259.2 | 129.6 | 129.6 | 129.6 |
| Inlet Flue Flow [ACFM / mol/s] | 290.3 / 1.051 | **579.5 / 2.098** | 1157.8 / 4.191 | 578.3 / 2.094 | 598.5 / 2.167 | 597.3 / 2.162 |
| Inlet $\text{H}_2\text{O}$ / $\text{HCl}$ | 13.03% / 4102 ppm | **13.05% / 2055 ppm** | 13.07% / 1029 ppm | 13.08% / 0 ppm | 14.92% / 1990 ppm | 14.95% / 0 ppm |
| Quench $T_{\text{sat}}$ [°F / K] | 183.84 / 357.51 | **183.86 / 357.52** | 183.87 / 357.52 | 183.88 / 357.53 | 184.10 / 357.65 | 184.12 / 357.66 |
| Quench Evap / Total Spray [GPM] | 0.289 / 0.579 | **0.578 / 1.156** | 1.155 / 2.309 | 0.577 / 1.154 | 0.579 / 1.158 | 0.578 / 1.156 |
| Quench Gas Out [ACFM / $y_{\text{H2O}}$] | 128.2 / 55.69% | **256.0 / 55.71%** | 511.5 / 55.72% | 255.5 / 55.73% | 260.6 / 56.00% | 260.1 / 56.02% |
| Quench Tower $D \times H$ [m] | $0.250 \times 2.35$ | **$0.350 \times 2.40$** | $0.500 \times 2.35$ | $0.350 \times 2.40$ | $0.350 \times 2.45$ | $0.350 \times 2.45$ |
| Venturi Throat $D_t \times L_t$ [mm] | $33.2 \times 80.0$ | **$46.9 \times 80.0$** | $66.3 \times 99.4$ | $46.8 \times 80.0$ | $47.3 \times 80.0$ | $47.3 \times 80.0$ |
| Venturi $d_{50}$ NaCl / Ash [$\mu$m] | 0.214 / 0.199 | **0.214 / 0.199** | 0.214 / 0.199 | 0.214 / 0.199 | 0.214 / 0.199 | 0.214 / 0.199 |
| Venturi $\eta_{\text{NaCl}}$ / $\eta_{\text{Ash}}$ / $\Delta P$ | 84.2% / 96.5% / 13.77" | **84.2% / 96.5% / 13.76"** | 84.2% / 96.5% / 13.76" | 84.2% / 96.5% / 13.76" | 84.1% / 96.5% / 13.72" | 84.1% / 96.5% / 13.72" |
| Packed Bed $D \times Z$ [m] (% Flood) | $0.25 \times 1.05$ (40.1%) | **$0.30 \times 1.05$ (55.5%)** | $0.40 \times 1.05$ (62.4%) | $0.30 \times 1.05$ (55.4%) | $0.30 \times 1.05$ (56.1%) | $0.30 \times 1.05$ (55.9%) |
| Packed Bed HTU [m] / HCl Out [ppm] | 0.222 / 18.5 | **0.226 / 10.0** | 0.227 / 5.2 | 0.226 / 0.0 | 0.226 / 9.9 | 0.226 / 0.0 |
| Packed Bed $\Delta P_{\text{bed}}$ / $\Delta P_{\text{dem}}$ [inWC] | 0.049 / 0.141 | **0.094 / 0.260** | 0.119 / 0.328 | 0.094 / 0.259 | 0.095 / 0.263 | 0.095 / 0.262 |
| $\text{NaHCO}_3$ Consumption [kg/hr] | 1.486 | **1.485** | 1.485 | 0.000 | 1.485 | 0.000 |
| Blower Duty [ACFM @ inWC] | 128.0 @ 14.42" | **255.7 @ 14.75"** | 511.2 @ 14.98" | 255.5 @ 14.74" | 260.3 @ 14.71" | 260.1 @ 14.71" |
| Blower Shaft / Motor HP | 0.47 / 0.75 HP | **0.96 / 1.50 HP** | 1.95 / 3.00 HP | 0.96 / 1.50 HP | 0.97 / 1.50 HP | 0.97 / 1.50 HP |
| Blower Wheel $D_w$ [mm] @ RPM | 690 @ 3524 | **700 @ 3513** | 700 @ 3541 | 700 @ 3512 | 700 @ 3549 | 700 @ 3548 |

---

## 2. First-Principles Audit of Implemented Models

### 2.1 Combustion Stoichiometry & Flue Gas Flow (`pkg/scrubber/scrubber.go`)

1. **Heat Release & Kiln Volume Scaling (`scrubber.go:25-65`):**
   - Reference volumetric heat load is `refHeatLoadWattsPerCuFt = 5400.0 W/cu.ft.` ($18{,}425\text{ BTU/hr}\cdot\text{ft}^{-3}$ at $2400°F$), yielding $129.6\text{ kW}$ ($442{,}214\text{ BTU/hr}$) for a $24\text{ cu.ft.}$ kiln.
   - **Validation against ceramic kiln practice:** Ceramic burner/kiln engineering texts (Olsen, *The Kiln Book*; commercial stoneware kiln specs) specify $8{,}000\text{–}12{,}000\text{ BTU/hr}\cdot\text{ft}^{-3}$ for insulated softbrick/fiber kilns and $15{,}000\text{–}20{,}000\text{ BTU/hr}\cdot\text{ft}^{-3}$ for heavy hardbrick/wood-fired cross-draft kilns at Cone 10 ($2350\text{–}2400°F$). The implied hardwood burn rate at $16.0\text{ MJ/kg}$ is $29.16\text{ kg/hr}$ ($64.3\text{ lbs/hr}$), which is realistic for peak Cone 10 stoking of a $24\text{ cu.ft.}$ wood kiln.
   - **Minor Constant Bug (`scrubber.go:26`):** `refPeakDeltaTK = 1588.15 - 293.15` treats $2400°F$ as $1588.15\text{ K}$, whereas exact conversion $(2400 - 32)\times 5/9 + 273.15 = 1588.7056\text{ K}$ ($1588.15\text{ K}$ is $2399.0°F$). This introduces a negligible $+0.043\%$ offset in `tempScale`.

2. **Fuel Stoichiometry (`scrubber.go:28-85`):**
   - **Wood (`C1 H1.44 O0.66`, 15% wet-basis moisture):** Dry unit molar mass $= 12.011 + 1.44(1.008) + 0.66(15.999) = 24.022\text{ g/mol C}$; as-fired mass $= 24.022 / 0.85 = 28.261\text{ g/mol C}$; moisture $= (28.261 - 24.022)/18.01528 = 0.2353\text{ mol }\text{H}_2\text{O/mol C}$ (`woodMoistH2OPerMolC = 0.235`); combustion water $= 1.44/2 = 0.72$; stoichiometric $\text{O}_2 = 1 + 1.44/4 - 0.66/2 = 1.03$. At $15\%$ moisture, seasoned hardwood LHV $\approx 0.85(18.8) - 0.15(2.44) = 15.6\text{–}16.0\text{ MJ/kg}$. **Verified exact.**
   - **Natural Gas ($\text{CH}_4$):** $\text{LHV} = 802.3\text{ kJ/mol}$, stoichiometric $\text{O}_2 = 2.0$, $\text{H}_2\text{O} = 2.0$, $\text{CO}_2 = 1.0$. **Verified exact.**
   - **Air ratio:** `airN2PerO2 = 0.79 / 0.21 = 3.7619` (dry standard air; neglects ambient humidity of combustion air, which adds $\sim 1\text{–}2\%$ $\text{H}_2\text{O}$ mole fraction in summer).

3. **First-Law Energy Balance Closure & Excess-Air Limit:**
   - For the 24 cu.ft. `wood_salt` reference case at $e = 0.30$: fuel LHV input is **$129.60\text{ kW}$**, while flue gas sensible enthalpy leaving at $1588.71\text{ K}$ is $2.0978\text{ mol/s} \times 46{,}493.9\text{ J/mol} = \mathbf{97.54\text{ kW}}$ (**$75.3\%$ of LHV**), implying $32.06\text{ kW}$ ($24.7\%$) kiln wall/ware heat loss — physically realistic for a hot Cone 10 soak.
   - **Adiabatic Flame Temperature Caveat:** Because `thermalWatts` is held fixed independent of `ExcessAirFraction` while `PeakExitTempF` is user-prescribed, specifying excess air above **$e_{\max} = 84.8\%$** at $2400°F$ (for 15%-moisture wood) produces flue-gas sensible enthalpy exceeding the fuel LHV (violating the First Law, since adiabatic flame temperature falls below $2400°F$ above $e = 84.8\%$). `DesignInput.Validate()` permits `ExcessAirFraction` up to `5.0`, so extreme combinations $(T_{\text{peak}}, e)$ are not checked against adiabatic flame temperature.

4. **Salt Hydrolysis Stoichiometry (`scrubber.go:42-43, 88-93`):**
   - Reaction: $2\,\text{NaCl} + \text{H}_2\text{O} + \text{SiO}_2 \to \text{Na}_2\text{SiO}_3 + 2\,\text{HCl}$ ($1\text{ mol HCl}$ per $\text{mol NaCl}$).
   - At $2.0\text{ lbs/hr}$ NaCl ($0.0002520\text{ kg/s}$, $MW = 0.0584428\text{ kg/mol}$), $n_{\text{HCl}} = 0.004312\text{ mol/s}$ ($2055\text{ ppm}$ dry-inlet basis; $1047\text{ ppm}$ post-quench).
   - Assuming **100% conversion to HCl** is an intentionally conservative upper bound for Stage 3 absorber and $\text{NaHCO}_3$ sizing (in reality, $15\text{–}40\%$ of NaCl escapes as unreacted salt fume or deposits on brickwork). Omitting the $0.5\,n_{\text{HCl}}$ water consumed by hydrolysis ($0.00216\text{ mol/s}$ vs. $0.2737\text{ mol/s}$ flue $\text{H}_2\text{O}$) is $< 0.8\%$ and negligible.

---

### 2.2 Stage 1: Adiabatic Spray Quench (`pkg/thermo/shomate.go`, `pkg/thermo/quench.go`)

1. **Thermophysical Properties (`shomate.go:53-482`):**
   - **Shomate $C_p(T)$ & $H(T)-H(298.15\text{ K})$:** Official NIST Chemistry WebBook 7-coefficient parameters for $\text{N}_2, \text{CO}_2, \text{H}_2\text{O}, \text{O}_2, \text{HCl}$. All 25 $(T, \text{species})$ oracle fixture points pass within $< 10^{-6}$ relative error. Note that NIST's $\text{H}_2\text{O}$ polynomial (calibrated $500\text{–}1700\text{ K}$) is extrapolated down to $298\text{ K}$; our audit verified $H_{\text{H2O}}(298.15\text{ K}) = +0.53\text{ J/mol}$ ($< 0.002\%$ of quench enthalpy — completely negligible).
   - **Viscosity:** Pure-gas `LucasGasViscosity` (`shomate.go:190-231`) matches Poling et al. (5th ed., Eqs. 9-4.15–18) and `chemicals.viscosity.Lucas_gas` to $< 10^{-6}$; `MixtureViscosity` (`shomate.go:396-437`) implements Wilke's (1950) mixing rule exactly.
   - **Water Vapor Pressure (`WaterPsat`, `shomate.go:252-297`):** Uses Wagner & Pruss (1993) Eq. 2.5 plus a $C^1$-tapered Helmholtz residual polynomial over $300\text{–}373.15\text{ K}$, matching `chemicals.iapws.iapws95_Psat` to $\le 2 \times 10^{-7}$ across all 10 oracle points while preserving strict monotonicity up to $T_c = 647.096\text{ K}$.

2. **Enthalpy Reference State & Latent Heat Consistency (`shomate.go:329-343`):**
   - Shomate gas-phase $\text{H}_2\text{O}$ sets $H_v(298.15\text{ K}) = 0$. `LiquidWaterEnthalpyJPerMol(T)` places liquid water on the **identical datum**:
     $$h_l(T) = -43{,}987.9 + 75.34\,(T - 298.15)\quad[\text{J/mol}]$$
   - At $T_{\text{sat}} = 357.52\text{ K}$ ($84.37°C$), the implied enthalpy of vaporization is $h_{fg} = (H_v(357.52) - h_l(357.52))/MW_{\text{H2O}} = \mathbf{2304.66\text{ kJ/kg}}$, compared to exact IAPWS-95 $h_{fg}(84.37°C) = 2297.6\text{ kJ/kg}$ (**$+0.31\%$ agreement**). **Verified thermodynamically consistent.**

3. **Brent Energy & Mass Balance Closure (`quench.go:127-246`):**
   - Independent check on the 24 cu.ft. reference case confirms:
     - Mass balance: $\dot{m}_{\text{in}} + \dot{m}_{\text{evap}} = 0.06097 + 0.03643 = 0.097400\text{ kg/s} \equiv \dot{m}_{\text{quenched}} = 0.097400\text{ kg/s}$.
     - Energy balance: $\text{LHS} = \dot{n}_{\text{in}} H_{\text{in}} + \dot{n}_{\text{evap}} h_l(293.15\text{ K}) = 7912.76\text{ W} \equiv \text{RHS} = \dot{n}_{\text{sat}} H_{\text{sat}} = 7912.76\text{ W}$ (residual $= 0.0000\text{ W}$).
   - **Recirculating vs. Once-Through Overspray Nuance (`quench.go:108, 244-246`):** The solver assumes `overspray = 2.0` ($1.156\text{ GPM}$ total spray vs. $0.578\text{ GPM}$ evaporated), meaning $0.578\text{ GPM}$ drains un-evaporated. The energy equation `hInTotal + nEvap*hLiq = hSatTotal` is **exact for a recirculating quench sump** (where $0.578\text{ GPM}$ of $20°C$ makeup water enters and hot sump water at $T_{\text{sat}}$ recirculates with zero net sensible transfer). If operated **once-through** with $20°C$ city water draining at $84.4°C$, heating the un-evaporated $1\times$ overspray absorbs an additional **$9.80\text{ kW}$** ($11.0\%$ of evaporation latent heat), lowering $T_{\text{sat}}$ by $\sim 1.5°C$.

4. **Residence Time & Droplet Evaporation Feasibility (`quench.go:249-265`):**
   - Tower diameter ($0.350\text{ m} = 13.8\text{ in}$) is sized from the geometric mean flow $Q_{\text{mean}} = \sqrt{Q_{\text{in}} Q_{\text{out}}}$ targeting $2.0\text{ m/s}$ (giving $v_{\text{in}} = 2.84\text{ m/s}$, $v_{\text{mean}} = 1.89\text{ m/s}$, $v_{\text{out}} = 1.26\text{ m/s}$). Height $H = 2.40\text{ m}$ provides $\tau = \mathbf{1.27\text{ s}}$.
   - **$d^2$-Law Evaporation Audit:** Evaluating droplet evaporation lifetime $\tau_{\text{evap}} = \rho_L h_{fg} d_0^2 / (4 k_g \text{Nu} \ln(1 + B_T))$ with $B_T = c_{p,g}\Delta T / h_{fg}$ across the quench thermal gradient:
     - Mean zone ($\Delta T = 340\text{ K}$, stagnant $\text{Nu} = 2$): $d_{\max}(\tau = 1.27\text{ s}) = \mathbf{252\,\mu\text{m}}$ ($200\,\mu\text{m}$ evaporates in $0.80\text{ s}$; $400\,\mu\text{m}$ requires $3.20\text{ s}$).
     - Mean zone with convective slip ($\Delta T = 340\text{ K}$, $\text{Nu} = 4$): $d_{\max} = \mathbf{357\,\mu\text{m}}$.
     - Conservative exit zone ($\Delta T = 120\text{ K}$, $\text{Nu} = 2$): $d_{\max} = \mathbf{154\,\mu\text{m}}$.
   - **Engineering Consequence:** Standard low-pressure ($20\text{–}40\text{ psig}$) $0.75\text{ GPM}$ hydraulic full-cone nozzles produce Sauter mean diameters of $350\text{–}600\,\mu\text{m}$ ($D_{\max} \approx 900\text{–}1200\,\mu\text{m}$), which will **not** fully evaporate within $1.27\text{ s}$ and will impinge on the $0.35\text{ m}$ wall! Complete evaporation within $1.27\text{ s}$ requires **air-atomizing (twin-fluid) nozzles** or **high-pressure ($80\text{–}150\text{ psig}$) hollow/full-cone fog nozzles** producing $\text{SMD} \le 150\text{–}200\,\mu\text{m}$, or increasing quench residence time to $2.0\text{–}3.0\text{ s}$ (standard EPA/industrial quench practice for hydraulic nozzles).

5. **Outlet Temperature vs. CPVC/PP Thermal Limit & Control Setpoint Bug (`types.go:456, 585`):**
   - At $2400°F$ kiln exhaust, **$T_{\text{sat}} = 183.86°F$ ($84.37°C$)** — leaving only **$26.1°F$** below the $210°F$ bypass interlock (`MaxQuenchOutletTemp`).
   - Standard ASTM Schedule 80 **CPVC** maximum continuous temperature is **$200°F$ ($93°C$)** (derated at pressure), and homopolymer **Polypropylene (PP)** continuous atmospheric limit is **$180\text{–}200°F$ ($82\text{–}93°C$)**, with thin-walled PP Pall rings softening under bed load above $190\text{–}200°F$. Setting the interlock at $210°F$ leaves zero margin above CPVC/PP continuous service ratings.
   - **Category A Bug (`pkg/model/types.go:585`, `internal/control/sim_hal.go:54`):** `DefaultControlConfig().QuenchSetpoint` is **`165.0°F` ($347.04\text{ K}$)** — **$18.9°F$ below the thermodynamic adiabatic saturation temperature at $2400°F$ ($183.9°F$)**, and still $12.5°F$ below $T_{\text{sat}}$ at $1800°F$ ($177.5°F$) and $2.7°F$ below $T_{\text{sat}}$ even at $1200°F$ ($167.7°F$). An adiabatic evaporative quench is physically bounded below by $T_{\text{sat}}$; commanding a PID loop to hold $165°F$ causes permanent integral windup / 100% spray valve saturation during high-fire unless excess cold once-through subcooling water is supplied to condense vapor.

---

### 2.3 Stage 2: Wet Venturi Scrubber (`pkg/scrubber/venturi.go`)

1. **Calvert (1970, 1972) Impaction Integral & Cut Diameter (`venturi.go:24-151`):**
   - **Cunningham Slip (`venturi.go:26-32`):** Implements Davies (1945) / Allen–Raabe: $C_c = 1 + \text{Kn}\,(1.257 + 0.4\,e^{-1.1/\text{Kn}})$ with $\text{Kn} = 2\lambda/d_p$. **Verified exact.** Using fixed $\lambda = 66.0\text{ nm}$ (20°C air) instead of the actual quenched-gas mean free path $\lambda(357.5\text{ K}) = (\mu_g/P)\sqrt{\pi R T / (2 M)} = 70.3\text{ nm}$ shifts $d_{50}$ by only $2.3\%$ ($0.214 \to 0.209\,\mu\text{m}$).
   - **Calvert $F(K_p, f)$ (`venturi.go:40-60`):**
     $$F(K_p, f) = \frac{1}{K_p}\left[-0.7 - K_p f + 1.4\ln\!\left(1 + \frac{K_p f}{0.7}\right) + \frac{0.49}{0.7 + K_p f}\right]$$
     For $u = K_p f / 0.7 \le 0.05$, the code switches to the exact alternating Taylor series $F(K_p, f) = f \sum_{k=3}^{14} (-1)^k \frac{k-2}{k} u^{k-1}$. Independent algebraic expansion confirms every coefficient and sign ($k=3$ yields $-u^2/3 < 0$, eliminating catastrophic cancellation for submicron particles). Matches all 6 `calvert_venturi` oracle points to $< 10^{-6}$.
   - **Penetration (`venturi.go:64-93`):** $P_t = \exp\!\left(\frac{2 (L/G) \rho_L v_t d_d}{55 \mu_g} F(K_p, f)\right)$ with $K_p = \frac{C_c \rho_p d_p^2 v_t}{9 \mu_g d_d}$. **Matches Calvert (1970, 1972) Eq. 12 identically.**

2. **Droplet SMD: Hardcoded $50\,\mu\text{m}$ Default vs. Nukiyama–Tanasawa (`venturi.go:155-172`, `types.go:68`):**
   - `NukiyamaTanasawaDropletDiameter` implements the standard Nukiyama–Tanasawa (1939) pneumatic atomization correlation accurately (CGS property units: $\sigma = 72.8\text{ dyn/cm}$, $\rho_L = 1.0\text{ g/cm}^3$, $\mu_L = 0.01\text{ P}$ at $20°C$).
   - However, `DefaultSaltKilnInput()` (`pkg/model/types.go:68`) hardcodes `VenturiDropletDiameter = 50 μm` rather than `0` (auto-calculate). At $v_t = 70\text{ m/s}$ and $L/G = 0.0012$, Nukiyama–Tanasawa yields **$d_{32} = 109.0\,\mu\text{m}$** at $20°C$ water properties (and **$92.9\,\mu\text{m}$** at $74°C$ saturated water properties: $\sigma = 63.7\text{ dyn/cm}$, $\mu_L = 0.00383\text{ P}$).
   - **Sensitivity Analysis (24 cu.ft. Reference Case):**

| Venturi Modeling Assumption | $d_d$ [$\mu$m] | $f$ | $d_{50,\text{NaCl}}$ [$\mu$m] | Overall $\eta_{\text{NaCl}}$ (MMAD $0.55\,\mu\text{m}$) | Overall $\eta_{\text{Ash}}$ (MMAD $1.5\,\mu\text{m}$) |
| :--- | :---: | :---: | :---: | :---: | :---: |
| Coded default (`types.go:68`) | 50.0 | 0.50 | 0.214 | 84.2% | 96.5% |
| Actual quenched $\lambda = 70.3\text{ nm}$ | 50.0 | 0.50 | 0.209 | 84.5% | 96.6% |
| Auto Nukiyama–Tanasawa ($74°C$ water) | 92.9 | 0.50 | 0.225 | 86.9% | 98.6% |
| Auto Nukiyama–Tanasawa ($20°C$ water) | 109.0 | 0.50 | 0.236 | 86.4% | 98.7% |
| Hydrophobic dust limit ($f = 0.25$) | 109.0 | 0.25 | **0.497** | **54.9%** | **88.8%** |

   - **Key Takeaway:** Calvert's empirical factor $f$ dominates submicron NaCl capture far more than $d_d$ or $\lambda$. While $f = 0.50$ is appropriate for strongly hydrophilic/hygroscopic NaCl aerosol (Calvert 1972 recommends $0.4\text{–}0.5$ for hydrophilic particulates), partially carbonaceous wood smoke / soot is hydrophobic ($f \approx 0.25$), where fine-fume capture drops sharply.

3. **Category A Physics Bug — Venturi Pressure Drop (`venturi.go:302-305`):**
   - Coded formula (`pkg/scrubber/venturi.go:303-305`):
     ```go
     dpAccel := DefaultCalvertFactor * rhoL * vThroat * vThroat * lgRatio
     dpDiffuser := venturiDiffuserLossK * 0.5 * rhoG * vThroat * vThroat
     ```
   - **Defect:** Line 303 multiplies the droplet acceleration momentum balance $\rho_L v_t^2 (L/G)$ by **`DefaultCalvertFactor = 0.5`**. From first-principles 1D axial momentum conservation across a venturi throat accelerating liquid from $u_{L,0} \approx 0$ to $u_{L,\text{exit}} = \beta v_t$:
     $$\Delta P_{\text{accel}} = \frac{\dot{m}_L u_{L,\text{exit}}}{A_t} = \beta\,\rho_L v_t^2 \left(\frac{L}{G}\right)$$
     Calvert's empirical particulate impaction constant $f \in [0.25, 0.50]$ is **physically unrelated** to the droplet velocity ratio $\beta = u_{L,\text{exit}}/v_t$.
   - **Independent Quantitative Check (Boll 1973 ODE & Yung–Calvert 1977):** Integrating Boll's (1973) exact 1D droplet drag ODE $\frac{dv_d}{dx} = \frac{3}{4}\frac{\rho_g}{\rho_L}\frac{C_D}{d_d}\frac{(v_g - v_d)^2}{v_d}$ over the coded throat length $L_t = 80.0\text{ mm}$ ($v_t = 70\text{ m/s}$, $\rho_g = 0.806\text{ kg/m}^3$, $L/G = 0.0012$):

| Venturi $\Delta P$ Model | Effective $\beta$ | $\Delta P_{\text{accel}}$ [Pa / inWC] | Total $\Delta P_{\text{venturi}}$ [Pa / inWC] | Relative to `kilnscrub` |
| :--- | :---: | :---: | :---: | :---: |
| **`kilnscrub` (`venturi.go:303-305`)** | 0.500 (uses $f$) | 2935 / 11.78 | **3429 / 13.76** | Baseline |
| **Boll (1973) ODE** ($d_d = 50\,\mu\text{m}, L_t = 80\text{ mm}$) | 0.712 | 4176 / 16.77 | **4670 / 18.75** | **+36.2%** |
| **Yung, Barbarika & Calvert (1977)** ($d_d = 50\,\mu\text{m}, L_t = 80\text{ mm}$) | 0.717 | 4209 / 16.90 | **4703 / 18.88** | **+37.2%** |
| **Boll (1973) ODE** ($d_d = 50\,\mu\text{m}, L_t = 150\text{ mm}$) | 0.815 | 4781 / 19.19 | **5275 / 21.18** | **+53.8%** |
| **Calvert (1972) typical industrial** ($\beta = 0.85$) | 0.850 | 4989 / 20.03 | **5483 / 22.01** | **+59.9%** |
| **Calvert (1968, 1970) full accel.** ($\beta = 1.0$) | 1.000 | 5869 / 23.56 | **5869 / 23.56** *(no separate diffuser)* | **+71.2%** |
| **Boll (1973) ODE** ($d_d = 109\,\mu\text{m}, L_t = 80\text{ mm}$) | 0.514 | 3017 / 12.11 | **3511 / 14.10** | +2.4% *(coincidental)* |
| **Hesketh (1974)** ($A_t = 0.0186\text{ ft}^2$) | — | — | **1433 / 5.75** *(extrapolated below industrial $A_t$)* | -58.2% |

   - **Verdict:** For the coded $d_d = 50\,\mu\text{m}$ default, `venturi.go:303` **underpredicts venturi $\Delta P$ by $36\%\text{–}71\%$** relative to Boll (1973), Yung–Calvert (1977), and Calvert (1970). (Note: Hesketh 1974 scales as $A_t^{0.133}$ calibrated to large industrial throats of $1\text{–}10\text{ ft}^2$ and is known to underpredict small $1.85\text{"}$ studio throats.) Also note that `SizeVenturi` uses hot `ductDiameter` ($177.8\text{ mm}$, sized for $580\text{ ACFM}$ hot gas) as the venturi inlet diameter even though quenched gas is only $256\text{ ACFM}$, which slightly lengthens the converging/diverging cones conservatively.

---

### 2.4 Stage 3: Counter-Current Packed-Bed HCl Absorber (`pkg/scrubber/packedbed.go`)

1. **Eckert GPDC Flooding Correlation (`packedbed.go:48-82`):**
   - Flow parameter $X = (L/G)_{\text{mass}} \sqrt{\rho_g/\rho_L}$ (`X = 0.0985` for the 24 cu.ft. reference case).
   - Polynomial $\ln Y_{\text{flood}} = -3.7121 - 1.0371 \ln X - 0.1501 (\ln X)^2 - 0.007544 (\ln X)^3$ is the standard Aspen/literature cubic fit to the Sherwood–Lobo–Leva–Eckert GPDC flooding line ($Y_{\text{flood}} = 0.1325$ at $X = 0.0985$).
   - Dimensional check: $Y = \frac{v_f^2 F_p (\mu_L/\mu_w)^{0.2} (\rho_g/\rho_L)}{g}$ is strictly dimensionless when $F_p$ is in $\text{m}^{-1}$ ($170\text{ m}^{-1} \approx 52\text{ ft}^{-1}$ for 1" PP Pall rings; manufacturer range $52\text{–}55\text{ ft}^{-1}$ / $170\text{–}180\text{ m}^{-1}$).
   - Output: $v_{\text{flood}} = 3.077\text{ m/s}$, $D_{\text{col}} = 0.300\text{ m}$ ($11.8\text{ in}$), $v_s = 1.709\text{ m/s}$ (**$55.5\%$ of flood**, well within the $\le 75\%$ design constraint). Column-to-packing diameter ratio $D/d_p = 11.8$ (and $9.8$ at 12 cu.ft., $15.7$ at 48 cu.ft.), satisfying the Strigle/Perry rule of thumb $D/d_p \ge 8\text{–}10$ to limit wall channeling. **Verified sound.**

2. **Category A Physics Bug — Dry Nominal-Diameter Ergun Pressure Drop (`packedbed.go:35-44, 183`):**
   - `ErgunPressureDrop(dp, voidage, vs, rho, mu, length)` (`packedbed.go:35-44`) reproduces the textbook Ergun equation and `fluids.packed_bed.dP_packed_bed(..., Method="Ergun")` to machine precision ($< 10^{-12}$).
   - **Defect (`packedbed.go:183`):** `SizePackedBed` calls `ErgunPressureDrop` with `dp = float64(input.PackingDiameter) = 0.0254 m` (the nominal outer diameter of a hollow 1" Pall ring) and applies **dry-bed** Ergun to an **irrigated** counter-current tower.
   - In Ergun's derivation, $d_p$ is the **Sauter equivalent spherical diameter** $d_{p,\text{eq}} = 6(1 - \varepsilon)/a_p$. For 1" PP Pall rings ($\varepsilon = 0.90$, $a_p = 207\text{ m}^2/\text{m}^3$):
     $$d_{p,\text{eq}} = \frac{6 \times (1 - 0.90)}{207} = \mathbf{0.00290\text{ m} = 2.90\text{ mm}}\quad(\text{8.76× smaller than } 25.4\text{ mm!})$$
   - **Quantitative Impact (24 cu.ft. Reference Case, $v_s = 1.709\text{ m/s}$, $Z = 1.05\text{ m}$, $55.5\%$ of flood, $L = 7.05\text{ GPM/ft}^2$):**
     - Coded dry Ergun ($d_p = 25.4\text{ mm}$): $22.3\text{ Pa/m} \implies \mathbf{23.5\text{ Pa}\ (0.094\text{ inWC}, 0.027\text{ inWC/ft})}$.
     - Dry Ergun with Sauter $d_{p,\text{eq}} = 2.90\text{ mm}$: $201.8\text{ Pa/m} \implies \mathbf{211.8\text{ Pa}\ (0.850\text{ inWC}, 0.247\text{ inWC/ft})}$ (**9.0× higher**).
     - Irrigated Eckert/Strigle GPDC at $X = 0.0985$, $Y_{\text{op}} = 0.0409$ ($55.5\%$ flood): $0.35\text{–}0.60\text{ inWC/ft}$ ($285\text{–}490\text{ Pa/m}$) $\implies \mathbf{300\text{–}514\text{ Pa}\ (1.20\text{–}2.07\text{ inWC})}$ (**13×–22× higher**).
   - Because `testdata/oracle_fixtures.json` explicitly tests `Ergun(dp=0.0254, voidage=0.90)` against `fluids.packed_bed`, the unit test passed while masking the physical misuse of nominal ring size inside `SizePackedBed`.

3. **NTU–HTU Mass-Transfer Sizing (`packedbed.go:155-173`):**
   - For irreversible rapid chemical neutralization of HCl in alkaline $\text{NaHCO}_3$ ($m \approx 0$), overall mass transfer is gas-film controlled ($H_{OG} \approx H_G$) and $N_{OG} = -\ln(1 - \eta)$. At $\eta = 0.99$, $N_{OG} = 4.605$.
   - Coded empirical correlation (`packedbed.go:159`) yields `HTU = 0.226 m` ($0.74\text{ ft}$) and $Z = 1.05\text{ m}$ ($3.44\text{ ft}$), predicting $99.04\%$ removal ($1047 \to 10.0\text{ ppm}$).
   - **Independent Onda (1968) Mechanistic Verification:** Evaluating Onda's wetted-area and gas-film correlations for HCl ($D_G \approx 2.35 \times 10^{-5}\text{ m}^2/\text{s}$ at $357.5\text{ K}$) on hydrophobic 1" PP Pall rings ($\sigma_c = 0.033\text{ N/m}$, $a_t = 207\text{ m}^2/\text{m}^3$):
     - At hot sump ($84°C$, $\sigma_L = 0.062\text{ N/m}$, $\mu_L = 0.34\text{ cP}$): $a_w/a_t = 0.434 \implies H_G = \mathbf{0.321\text{ m}}$ ($0.434\text{ m}$ with $1.35\times$ maldistribution factor).
     - At cool sump ($25°C$, $\sigma_L = 0.072\text{ N/m}$, $\mu_L = 1.0\text{ cP}$): $a_w/a_t = 0.358 \implies H_G = \mathbf{0.389\text{ m}}$ ($0.526\text{ m}$ with $1.35\times$ maldistribution factor).
   - **Assessment (Category B):** Coded `HTU = 0.226 m` is **$1.4\times\text{–}2.3\times$ optimistic** relative to Onda (1968) and industrial HCl packed-tower design rules ($0.35\text{–}0.60\text{ m}$, Strigle / EPA Cost Manual). At $H_{OG} = 0.43\text{ m}$, a $1.05\text{ m}$ bed alone achieves $91.3\%$ removal ($Z \approx 2.0\text{ m}$ needed for $99\%$). In the complete 4-stage train, however, **uncredited upstream HCl absorption in the Stage 1 quench and Stage 2 venturi** (typically $30\%\text{–}60\%$ for highly soluble HCl) partially offsets the short Stage 3 bed, bringing overall train removal into the $\sim 95\%\text{–}98\%$ band.

4. **Wetting Rate, Reagent Stoichiometry & Demister (`packedbed.go:175-188, 238-254`):**
   - **Minimum Wetting Rate (MWR):** At $5.36\text{ GPM}$ over $A = 0.0707\text{ m}^2$, superficial liquid flux is $0.00479\text{ m}^3/(\text{m}^2\cdot\text{s})$ ($7.05\text{ GPM/ft}^2$; $L/G = 20.9\text{ gal/1000 acf}$), giving $\text{MWR} = L_{\text{flux}}/a_t = \mathbf{2.31 \times 10^{-5}\text{ m}^2/\text{s}}$. Literature MWR for plastic packings is $2.2\text{–}3.3 \times 10^{-5}\text{ m}^2/\text{s}$ (Schmidt / Norton), so the 24 cu.ft. design sits right at the lower edge of full wetting ($0.70\text{–}1.05\times$ MWR); note that `packedBedMinWettingFlux = 0.0014 m3/(m2*s)` ($2.0\text{ GPM/ft}^2$) is only $0.68 \times 10^{-5}\text{ m}^2/\text{s}$ — too low for hydrophobic PP if triggered on small turndown columns.
   - **$\text{NaHCO}_3$ Stoichiometry:** $\text{HCl} + \text{NaHCO}_3 \to \text{NaCl} + \text{H}_2\text{O} + \text{CO}_2$. Stoichiometric demand is $0.004312 \times 0.9904 \times 0.0840066 \times 3600 = 1.292\text{ kg/hr}$; with $1.15\times$ buffer $= \mathbf{1.485\text{ kg/hr}}$ ($3.27\text{ lbs/hr}$). **Verified exact.** Minor omission: `outletGas` (`packedbed.go:238-254`) subtracts $n_{\text{HCl,rem}}$ from $n_{\text{out}}$ without adding the $1:1$ liberated $\text{CO}_2$ ($0.10\%$ molar effect — negligible).
   - **Chevron Mist Eliminator (`packedbed.go:24, 187`):** $\Delta P_{\text{dem}} = \max(35, 55 \times \frac{1}{2}\rho_g v_s^2) = \mathbf{64.8\text{ Pa}\ (0.260\text{ inWC})}$ at $v_s = 1.71\text{ m/s}$, well aligned with industrial 3-pass polypropylene chevron blades ($0.2\text{–}0.5\text{ inWC}$ at $1.5\text{–}2.5\text{ m/s}$).

---

### 2.5 Stage 4: Induced-Draft Centrifugal Blower (`pkg/scrubber/blower.go`)

1. **Static Pressure Accounting & Duct Friction (`blower.go:58-124`):**
   - `DuctPressureDrop` uses the exact **Haaland (1983)** explicit Darcy friction factor ($\varepsilon = 0.15\text{ mm}$, $\sum K = 2.5$). **Verified exact.** However, only the **hot inlet duct** (`3.0 m`) is accounted for; no downstream cold exhaust duct or discharge stack friction/exit loss ($\sim 0.15\text{–}0.40\text{ inWC}$) is included.
   - `PressureDropBreakdown` separates `PackedBed` (`BedPressureDrop` only) from `Demister` (zero double-counting) and adds `abs(TargetKilnDraftInWC)` ($+0.050\text{ inWC}$) while echoing signed `-0.050 inWC` in `TargetKilnDraftInWC`. **Verified exact.**

2. **Vacuum Suction Density & Power (`blower.go:125-131`):**
   - Because the ID blower is downstream of the venturi and packed bed, its actual inlet static pressure is $P_{\text{in}} = 101{,}325 - \Delta P_{\text{tot}}$. At coded $\Delta P = 3673\text{ Pa}$, $P_{\text{in}} = 97{,}652\text{ Pa}$ ($96.4\%$ atm), increasing actual blower inlet ACFM and shaft power by **$+3.76\%$** (and **$+6.0\%\text{–}7.5\%$** under realistic $20.5\text{–}28.5\text{ inWC}$ total $\Delta P$).
   - Including the Category A Venturi + Irrigated Packed-Bed corrections, realistic duty for the 24 cu.ft. kiln is **$265\text{–}275\text{ ACFM}$ @ $20.5\text{–}28.5\text{ inWC}$**, requiring **$1.35\text{–}1.85\text{ shaft HP}$** (a **$2.0\text{–}3.0\text{ HP}$ NEMA motor**, rather than the coded $0.96\text{ shaft HP} \to 1.5\text{ HP}$ motor).

3. **Turbomachinery Specific Speed & Impeller Geometry (`blower.go:134-144`):**
   - At $Q = 255.7\text{ ACFM}$, $\Delta P = 14.75\text{ inWC}$ (equivalent to **$21.95\text{ inWC}$** at standard air $\rho = 1.20\text{ kg/m}^3$), and $N = 3513\text{ RPM}$, the blower's specific speed is:
     - Dimensional US specific speed: $N_s = N\sqrt{Q}/\Delta P_{\text{std}}^{0.75} = \mathbf{5{,}539}$
     - Cordier dimensionless specific speed: $\omega_s = \mathbf{0.230}$ ($D_s = 16.55$)
   - **Physical Finding:** Backward-inclined centrifugal fans (`blowerStaticEfficiency = 0.62`, $\psi = 0.55$) operate optimally at $N_s = 15{,}000\text{–}40{,}000$ ($\omega_s = 0.6\text{–}1.5$). Forcing $\psi = 0.55$ at $N_s = 5{,}539$ drives `blower.go:141-144` to expand the wheel to **$700\text{ mm}$ ($27.6\text{ in}$) at $3513\text{ RPM}$** (tip speed $U_{\text{tip}} = \mathbf{128.8\text{ m/s} = 25{,}345\text{ ft/min}}$ — near structural/acoustic limits for a small fabricator-built wheel!). In industrial practice, low-flow / high-head duty ($250\text{ ACFM @ } 15\text{–}28\text{ inWC}$) requires a **narrow radial-blade pressure blower** ($N_s = 5{,}000\text{–}15{,}000$, $\psi \approx 0.9\text{–}1.1$, wheel $\sim 19\text{–}22\text{ in}$ @ $3550\text{ RPM}$, $\eta_s \approx 0.50\text{–}0.58$) or a **multistage / regenerative blower**.

4. **System Curve & Negative Kiln Draft (`blower.go:157-189`):**
   - `P_sys(Q) = P_draft + K_sys * Q^2` properly includes the nonzero static intercept $P_{\text{draft}} = +0.05\text{ inWC}$ at $Q = 0$ and intersects the fan curve cleanly at `OperatingPoint`. Note that if venturi liquid flow $Q_L$ is held constant by a fixed-speed pump while VFD gas flow $Q$ varies, $L/G \propto 1/Q$ makes $\Delta P_{\text{venturi}} \propto Q$ (linear) rather than $Q^2$; the quadratic curve applies when $L/G$ is ratio-controlled.

---

## 3. Rigorous First-Principles Determinism vs. Irreducible Studio-Kiln Uncertainty

A rigorous engineering tool must separate what is **fixed by conservation laws** from what depends on **unmeasured studio firing dynamics**:

| Category | Quantity / Phenomenon | Governing Physics or Source of Uncertainty | Practical Uncertainty Band |
| :--- | :--- | :--- | :--- |
| **Deterministically Rigorous** | **Adiabatic saturation floor $T_{\text{sat}}$ & evap. water rate** | First Law + IAPWS-95 $P_{\text{sat}}(T)$ + NIST Shomate $H(T)$ given $(T_{\text{in}}, \dot{n}, y_{\text{H2O}})$ | **$\pm 1\text{–}2°F$** ($\pm 1\text{–}3\%$ evap. GPM) |
| **Deterministically Rigorous** | **Ideal-gas flue density, ACFM, & viscosity** | Ideal gas law ($Z > 0.998$ at $1\text{ atm}$), Lucas + Wilke mixing | **$\pm 2\text{–}3\%$** |
| **Deterministically Rigorous** | **Upper-bound $\text{NaHCO}_3$ reagent demand** | $1:1$ molar stoichiometry $\text{NaCl} \to \text{HCl} \to \text{NaHCO}_3$ | Exact upper bound (real demand $40\text{–}85\%$ of upper bound) |
| **Deterministically Rigorous** | **Eckert GPDC flooding velocity $v_f$** | Validated empirical packing factor $F_p$ & GPDC curve | **$\pm 10\text{–}15\%$** |
| **Deterministically Rigorous** | **Calvert $d_{50}$ given $(v_t, L/G, d_d, f)$** | Exact inertial impaction integral | Exact for given $(d_d, f)$ |
| **Site / Empirical Uncertainty** | **Peak stoking flue-gas surge vs. steady average** | Wood stoking creates 20–90 s pyrolysis/combustion bursts ($1.5\times\text{–}2.5\times$ average flow & heat load) and positive firebox puffs | **$-40\%$ to $+150\%$** transient flow/heat |
| **Site / Empirical Uncertainty** | **Fraction of thrown NaCl hydrolyzing to HCl vs. NaCl fume vs. kiln glaze** | Depends on kiln atmosphere (oxidizing vs. reducing), moisture content of salt/wood, silica surface area, and firing temp | **$40\text{–}70\%$ HCl**, **$15\text{–}40\%$ NaCl fume**, **$10\text{–}30\%$** retained on pots/brick |
| **Site / Empirical Uncertainty** | **Submicron NaCl fume & wood-smoke PSD + wettability $f$** | Nucleation/condensation during rapid quench shifts MMAD ($0.2\text{–}0.8\,\mu\text{m}$); soot hydrophobicity shifts Calvert $f \in [0.25, 0.50]$ | $\eta_{\text{NaCl}}$ spans **$55\%\text{–}90\%$** at fixed throat $\Delta P$ |
| **Site / Empirical Uncertainty** | **Actual nozzle droplet SMD & quench wall impingement** | Hydraulic nozzle wear, supply pressure drift, and spray-cone overlap in a narrow $13.8\text{"}$ tower | **$\pm 40\text{–}80\%$** in evaporation length |
| **Site / Empirical Uncertainty** | **Irrigated packing maldistribution & fouling** | Wood fly ash / NaCl scaling / $\text{Na}_2\text{CO}_3$ crusting on PP rings + liquid channeling in $12\text{"}$ column | **$\pm 30\text{–}50\%$** on effective HTU and $\Delta P$ |
| **Site / Empirical Uncertainty** | **Stack plume condensation & buoyancy** | Saturated $184°F$ ($56\%\text{ H}_2\text{O}$) plume exits visibly condensing in cool ambient air; chimney buoyancy fights or aids $-0.05\text{ inWC}$ draft | Site/weather dependent; requires stack drain & demister |

---

## 4. Comparison Against Proven Industrial & Studio Scrubber Designs

### 4.1 Quantitative Benchmark Table (24 cu.ft. Wood/Salt Reference vs. Literature Standards)

| Design Parameter | `kilnscrub` (24 cu.ft. Ref.) | EPA Cost Manual (EPA/452/B-02-001, Sec. 5.2 & Sec. 6 Ch. 2) | Perry's / Treybal / Strigle Packed-Tower HCl Rules | Proven Studio / University / Commercial Acid-Gas Scrubbers | Audit Verdict |
| :--- | :---: | :---: | :---: | :---: | :--- |
| **Quench Residence Time** | $1.27\text{ s}$ ($v = 1.89\text{ m/s}$) | $1.0\text{–}3.0\text{ s}$ (air-atomized $\sim 1\text{ s}$; hydraulic $2\text{–}3\text{ s}$) | $1.5\text{–}2.5\text{ s}$ | $1.5\text{–}3.0\text{ s}$ (or refractory-lined wetted-wall quencher) | **Borderline tight** — requires fine-atomizing nozzles ($\le 200\,\mu\text{m}$) |
| **Quench Outlet Temp** | $183.9°F$ ($84.4°C$) | Adiabatic saturation ($160\text{–}190°F$ for hot combustion gas) | $T_{\text{sat}}$ ($175\text{–}190°F$) | $165\text{–}185°F$ (often with upstream dilution air to drop $T_{\text{sat}} < 165°F$ for PVC/FRP) | **Physically exact**, but tight against $200°F$ CPVC/PP limit |
| **Venturi Throat Velocity** | $70.0\text{ m/s}$ ($230\text{ ft/s}$) | $50\text{–}120\text{ m/s}$ ($150\text{–}400\text{ ft/s}$; submicron fume $\ge 75\text{–}100\text{ m/s}$) | $60\text{–}100\text{ m/s}$ | $60\text{–}90\text{ m/s}$ (submicron fume scrubbers) | **In range** (mid-energy venturi) |
| **Venturi $L/G$ Ratio** | $0.0012$ ($8.98\text{ gal/1000 acf}$) | $2\text{–}20\text{ gal/1000 acf}$ (typ. $7\text{–}15$) | $5\text{–}15\text{ gal/1000 acf}$ | $8\text{–}15\text{ gal/1000 acf}$ | **Optimal** |
| **Venturi $\Delta P$** | **$13.76\text{ inWC}$** *(coded)* | $10\text{–}60\text{ inWC}$ (submicron salt/metal fume typically $20\text{–}45\text{ inWC}$) | $15\text{–}40\text{ inWC}$ | $18\text{–}35\text{ inWC}$ for submicron fume | **Underpredicted** (true Boll/Calvert is $18.8\text{–}23.6\text{ inWC}$) |
| **Packed-Bed Superficial Vel.** | $1.71\text{ m/s}$ ($5.61\text{ ft/s}$) | $1.0\text{–}3.0\text{ m/s}$ ($3\text{–}10\text{ ft/s}$) | $1.2\text{–}2.5\text{ m/s}$ ($4\text{–}8\text{ ft/s}$) | $1.2\text{–}2.0\text{ m/s}$ ($4\text{–}6.5\text{ ft/s}$) | **In range** |
| **Packed-Bed % of Flood** | $55.5\%$ ($\le 75\%$ cap) | $50\%\text{–}70\%$ (cap $75\%$) | $50\%\text{–}70\%$ | $50\%\text{–}65\%$ | **Optimal** |
| **Packed-Bed $L/G$ & Flux** | $20.9\text{ gal/1000 acf}$ ($7.05\text{ GPM/ft}^2$) | $5\text{–}20\text{ gal/1000 acf}$; $2\text{–}10\text{ GPM/ft}^2$ | $\ge 5\text{–}8\text{ GPM/ft}^2$ for plastic Pall rings | $10\text{–}25\text{ gal/1000 acf}$; $5\text{–}10\text{ GPM/ft}^2$ | **Good** (at lower bound of PP MWR) |
| **Packed-Bed $\text{HTU}_{OG}$ & Depth** | $0.226\text{ m}$ ($0.74\text{ ft}$); $Z = 1.05\text{ m}$ | $0.30\text{–}0.75\text{ m}$ ($1.0\text{–}2.5\text{ ft}$); $Z = 1.5\text{–}3.0\text{ m}$ | $0.30\text{–}0.55\text{ m}$ ($1.0\text{–}1.8\text{ ft}$) for 1" PP Pall rings | $1.2\text{–}2.4\text{ m}$ ($4\text{–}8\text{ ft}$) for $99\%$ HCl | **Optimistic HTU** ($1.4\times\text{–}2.3\times$ lower than Onda/Strigle) |
| **Packed-Bed $\Delta P$/Depth** | **$0.027\text{ inWC/ft}$** ($0.094\text{ inWC}$ total) | $0.25\text{–}0.75\text{ inWC/ft}$ ($0.5\text{–}1.0\text{ inWC/ft}$ max) | $0.35\text{–}0.60\text{ inWC/ft}$ at $55\text{–}65\%$ flood | $0.4\text{–}0.6\text{ inWC/ft}$ ($1.5\text{–}2.5\text{ inWC}$ total) | **Severe underprediction (13×–22× too low)** |
| **Neutralization Reagent** | $\text{NaHCO}_3$ (pH 7.5–8.5) | $\text{NaOH}$ (industry standard) or $\text{Na}_2\text{CO}_3$/$\text{NaHCO}_3$ | $\text{NaOH}$ / $\text{Na}_2\text{CO}_3$ | $\text{NaHCO}_3$ or $\text{Na}_2\text{CO}_3$ preferred in studios (non-caustic handling safety) | **Excellent studio safety choice** |

### 4.2 Critical Materials-of-Construction & Corrosion Evaluation

1. **Why 316L Stainless Steel Fails in Hot Wet HCl Service (`quench.go:118`, `venturi.go:16`, `blower.go:18`):**
   - In a salt-kiln quench and venturi, $2400°F$ flue gas containing $1{,}000\text{–}4{,}100\text{ ppm}$ $\text{HCl}$ meets water spray and rapidly cools across the acid dew point to $183.9°F$ ($84.4°C$). Unbuffered water droplets absorb HCl down to $\text{pH } 1\text{–}2$ with thousands of ppm dissolved $\text{Cl}^-$.
   - Austenitic **316L SS (UNS S31603, ~2–3% Mo)** loses passivity rapidly in dilute hot HCl ($> 100\text{ mpy}$ general corrosion in $\text{pH} < 2$ at $84°C$), suffers severe **chloride pitting and crevice corrosion** at flange gaskets and wet/dry deposit zones, and undergoes **transgranular chloride stress-corrosion cracking (Cl-SCC)** above $140°F$ ($60°C$) in welded shells and high-stress rotating blower impellers ($U_{\text{tip}} = 129\text{ m/s}$!).
   - **Proven Industrial & Academic Alternatives:**
     - *Gold-standard metallic quench/venturi:* **Alloy C-276 (UNS N10276)** or **Alloy C-22** (~16% Mo, high Ni — essentially immune to HCl pitting and Cl-SCC), or **graphite / silicon carbide (SiC) venturi throat**.
     - *Cost-effective industrial/studio quench:* **Acid-brick / castable refractory-lined steel quench** with a flooded alloy-C276 or SiC spray ring/weir, transitioning at $< 185°F$ directly into **Derakane / Hetron vinyl-ester FRP** (rated to $200\text{–}220°F$ wet acid service, immune to chlorides) for the venturi diffuser, packed tower, and blower housing + **FRP or Alloy C-276 / Titanium Gr. 7 blower wheel**.
     - *Budget studio mitigation if 316L SS is retained:* Never spray unbuffered tap water into 316L during salt firing — **recirculate alkaline $\text{NaHCO}_3/\text{Na}_2\text{CO}_3$ liquor ($\text{pH } 8.0\text{–}8.5$) to BOTH the quench and venturi**, keep sump chloride concentration controlled via blowdown, treat Stage 1–2 as **replaceable bolted/flanged sacrificial spools**, and inspect weld seams/wall thickness every 5–10 firings.
2. **Upstream Dilution Air & Thermal-Shock Protection for CPVC/PP:**
   - Many university and commercial studio salt-kiln scrubbers (and packaged units from manufacturers such as Tri-Mer, Croll-Reynolds, and Duall) incorporate a **barometric/controlled ambient dilution-air damper** upstream of the quench (e.g., diluting $2400°F$ flue gas to $1000\text{–}1400°F$). Although dilution increases volumetric flow ($+30\%\text{–}50\%$), it drops $T_{\text{sat}}$ from $184°F$ to **$145\text{–}160°F$** — providing a wide $40\text{–}55°F$ thermal safety margin for CPVC/PP/FRP and eliminating the $184°F$ vs. $200°F$ continuous-rating squeeze.
   - Hardware bypass interlocks protecting CPVC/PP **must** be backed by a **hardwired fail-open spring-return / gravity-weighted bypass damper** (power-loss and PLC-hang safe) plus a **gravity-fed emergency quench water tank** (3–5 min holdup) to bridge the $2\text{–}5\text{ s}$ damper travel time during power/pump failure.

---

## 5. Categorized Findings & Actionable Recommendations

### Category A — Code / Physics Bugs

1. **A1 — Packed-Bed Ergun $\Delta P$ Uses Nominal Ring Size ($25.4\text{ mm}$) & Dry-Bed Equation (`pkg/scrubber/packedbed.go:183`):**
   - **Issue:** `SizePackedBed` passes `input.PackingDiameter` ($0.0254\text{ m}$) directly into `ErgunPressureDrop`, violating Ergun's equivalent-sphere definition $d_{p,\text{eq}} = 6(1-\varepsilon)/a_p \approx 2.90\text{ mm}$ for 1" PP Pall rings and omitting irrigated liquid holdup.
   - **Impact:** Underpredicts packed-bed $\Delta P$ by **13×–22×** ($0.094\text{ inWC}$ coded vs. $1.20\text{–}2.07\text{ inWC}$ irrigated GPDC).
   - **Fix:** Preserve `ErgunPressureDrop` for oracle compatibility, but compute Stage 3 `BedPressureDrop` using an irrigated GPDC pressure-drop correlation (e.g., Robbins 1991 / Strigle GPDC / Billet–Schultes) or Sauter $d_{p,\text{eq}} = 6(1-\varepsilon)/a_p$ with an irrigated holdup multiplier $f_{\text{wet}}(L_{\text{flux}}, \%_{\text{flood}})$.

2. **A2 — Venturi Acceleration $\Delta P$ Conflates Calvert Impaction Factor $f$ with Momentum Factor $\beta$ (`pkg/scrubber/venturi.go:303`):**
   - **Issue:** `dpAccel := DefaultCalvertFactor * rhoL * vThroat * vThroat * lgRatio` uses Calvert's empirical particulate impaction constant $f = 0.5$ as the liquid momentum multiplier.
   - **Impact:** For default $d_d = 50\,\mu\text{m}$ and $L_t = 80\text{ mm}$, underpredicts venturi $\Delta P$ by **$36\%\text{–}71\%$** ($13.76\text{ inWC}$ coded vs. $18.75\text{–}18.88\text{ inWC}$ Boll/Yung–Calvert and $23.56\text{ inWC}$ Calvert 1970). Combined with A1, total 24 cu.ft. static pressure is underpredicted by **$6\text{–}14\text{ inWC}$**, undersizing the recommended blower motor ($1.5\text{ HP}$ coded vs. $2.0\text{–}3.0\text{ HP}$ required).
   - **Fix:** Replace `DefaultCalvertFactor` in `dpAccel` with the Yung, Barbarika & Calvert (1977) analytical throat acceleration factor $\beta(L_t, d_d, v_t, \rho_g, \rho_L)$ (or conservative Calvert $\beta = 0.85\text{–}1.0$).

3. **A3 — Default `QuenchSetpoint = 165.0°F` Is Below Adiabatic Saturation Floor (`pkg/model/types.go:585`, `internal/control/sim_hal.go:54`):**
   - **Issue:** `DefaultControlConfig()` sets `QuenchSetpoint = 165.0°F`, whereas `SolveAdiabaticQuench` proves $T_{\text{sat}} = 183.9°F$ at $2400°F$ (and $177.5°F$ at $1800°F$).
   - **Impact:** Physically unattainable in adiabatic spray quench mode; causes permanent 100% spray-valve saturation on real hardware.
   - **Fix:** Default `QuenchSetpoint` to $T_{\text{sat}} + 5°F$ ($\approx 188\text{–}190°F$ for undiluted $2400°F$ firing, or $165°F$ only when upstream dilution air is specified), and calibrate `SimHAL` baseline equilibrium against `SolveAdiabaticQuench`.

4. **A4 — Minor Kelvin Conversion Constant Offset (`pkg/scrubber/scrubber.go:26`):**
   - **Issue:** `refPeakDeltaTK = 1588.15 - 293.15` uses $1588.15\text{ K}$ ($2399.0°F$) instead of `float64(model.FahrenheitToKelvin(2400.0) - model.CelsiusToKelvin(20.0))` ($1588.7056 - 293.15\text{ K}$).
   - **Impact:** $+0.043\%$ thermal scaling bias (negligible).

---

### Category B — Engineering Limitations & Conservative Margins

1. **B1 — Critical Corrosion Hazard of 316L SS in Hot Wet HCl (`quench.go:118`, `venturi.go:16`, `blower.go:18`):** Emit an explicit design warning whenever `SaltRateLbsPerHr > 0` detailing 316L pitting/Cl-SCC risk above $140°F$ and recommending Alloy C-276, refractory-lined quench + vinyl-ester FRP, or alkaline-buffered quench/venturi spray.
2. **B2 — Tight Thermal Margin for CPVC/PP Packing at $183.9°F$ vs. $210°F$ Interlock (`types.go:456`):** Lower default `MaxQuenchOutletTemp` to **$195\text{–}200°F$** (matching ASTM CPVC/PP continuous ratings) or specify Glass-Filled PP / PVDF / vinyl-ester FRP internals.
3. **B3 — Optimistic Stage 3 HTU (`packedbed.go:159`):** Coded `HTU = 0.226 m` is $1.4\times\text{–}2.3\times$ below Onda (1968) with maldistribution ($0.43\text{–}0.53\text{ m}$). Add a `1.35×` maldistribution margin (increasing $Z$ from $1.05\text{ m}$ to $\sim 1.5\text{–}1.8\text{ m}$) or explicitly credit Stage 1–2 HCl pre-scrubbing.
4. **B4 — Low-Specific-Speed Blower Wheel & Suction Vacuum (`blower.go:125-144`):** At $N_s = 5{,}539$, switch impeller correlation from backward-inclined ($\psi = 0.55$) to radial-blade pressure blower ($\psi \approx 0.95$, $\eta_s \approx 0.55$), evaluate suction density at $P_{\text{amb}} - \Delta P_{\text{tot}}$ ($+4\%\text{–}7\%$ ACFM/HP), include downstream exhaust duct/stack losses, and add **$1.3\times\text{–}1.5\times$ stoking-surge flow headroom** for wood kilns.
5. **B5 — Default Venturi Droplet Diameter Override (`types.go:68`):** Defaulting `VenturiDropletDiameter = 0` (auto Nukiyama–Tanasawa using hot-water properties) avoids masking droplet-size coupling to $(v_t, L/G)$.
6. **B6 — Sump Blowdown & Dissolved Salt Buildup:** At $2\text{ lbs/hr}$ NaCl neutralized by $\text{NaHCO}_3$, the $21.4\text{ gal}$ ($81\text{ L}$) sump accumulates $\sim 1.03\text{ kg/hr}$ dissolved $\text{NaCl}$ plus residual $\text{NaHCO}_3$/wood ash ($\sim 1.3\%\text{ wt/hr}$ TDS). A continuous/batch **blowdown stream** ($0.05\text{–}0.15\text{ GPM}$) and mesh strainer are required to prevent nozzle plugging and salt crusting during 12–48 hr wood firings.
7. **B7 — Unchecked Adiabatic Flame Limit at High Excess Air (`scrubber.go:56-85`):** Add a validation warning if `inletGas` sensible enthalpy exceeds fuel LHV (occurs at $e > 84.8\%$ for $2400°F$ wood firing).

---

### Category C — Recommended Next Engineering Upgrades

1. **Mechanistic Mass-Transfer & Irrigated-Hydraulic Stage 3 Model:** Replace empirical `HTU` and dry nominal Ergun with **Onda (1968)** or **Billet–Schultes (1993)** wetted-area/mass-transfer equations coupled to **Robbins (1991) / Strigle GPDC** irrigated pressure drop, and add packing specific surface area $a_p$ and critical surface tension $\sigma_c$ to `DesignInput`.
2. **Boll (1973) / Yung–Calvert (1977) Coupled Venturi Solver:** Integrate droplet acceleration and gas-droplet drag across finite throat length $L_t$ with hot-water Nukiyama–Tanasawa SMD so pressure drop, throat length, and cut diameter $d_{50}$ are physically self-consistent.
3. **Optional Upstream Dilution-Air Quench Mode:** Add a `DilutionAirFraction` parameter to `DesignInput` so studio builders can trade $+30\%$ blower flow for a $25\text{–}35°F$ reduction in $T_{\text{sat}}$ (bringing CPVC/FRP safely into $150\text{–}160°F$ service).
4. **Transient Wood-Stoking Surge & Sump Chemistry Module:** Size blower motor, quench spray peak capacity, and sump blowdown for a configurable `StokingSurgeFactor` (default $1.5\times$ for wood kilns, $1.15\times$ for gas kilns) and compute sump TDS accumulation and pH buffering curves over firing duration.
5. **Expanded Materials-Selection Matrix:** Support selectable metallurgy/polymer tiers in `DesignInput` and BOM export: *Tier 1 Industrial* (Alloy C-276 quench/venturi + vinyl-ester FRP tower/blower), *Tier 2 Refractory/FRP* (castable-lined quench + SiC throat + FRP), and *Tier 3 Budget Sacrificial* (alkaline-sprayed 316L + CPVC with corrosion-inspection schedule).

---

## References

1. **U.S. EPA (2002).** *EPA Air Pollution Control Cost Manual*, 6th Ed., EPA/452/B-02-001 — Section 5.2, Ch. 1 (*Wet Scrubbers for Acid Gas*) & Section 6, Ch. 2 (*Wet Scrubbers for Particulate Matter*).
2. **Calvert, S. (1970).** "Venturi and Other Atomizing Scrubbers Efficiency and Pressure Drop," *AIChE Journal*, 16(3), 392–396.
3. **Calvert, S., Lundgren, D., & Mehta, D. S. (1972).** "Venturi Scrubber Performance," *Journal of the Air Pollution Control Association*, 22(7), 529–532.
4. **Boll, R. H. (1973).** "Particle Collection and Pressure Drop in Venturi Scrubbers," *Industrial & Engineering Chemistry Fundamentals*, 12(1), 40–50.
5. **Yung, S.-C., Barbarika, H. F., & Calvert, S. (1977).** "Pressure Loss in Venturi Scrubbers," *J. Air Pollution Control Assoc.*, 27(4), 348–351 (and EPA-600/2-77-172).
6. **Hesketh, H. E. (1974).** "Fine Particle Collection Efficiency Related to Pressure Drop, Scrubbant and Particle Properties, and Contact Mechanism," *J. Air Pollution Control Assoc.*, 24(10), 939–942.
7. **Nukiyama, S., & Tanasawa, Y. (1939).** "Experiments on the Atomization of Liquids in an Air Stream," *Trans. JSME*, 5(18), 62–67.
8. **Eckert, J. S. (1970).** "Selecting the Proper Distillation Column Packing," *Chemical Engineering Progress*, 66(3), 39–44.
9. **Strigle, R. F. (1994).** *Packed Tower Design and Applications*, 2nd Ed., Gulf Publishing; **Treybal, R. E. (1980).** *Mass-Transfer Operations*, 3rd Ed., McGraw-Hill.
10. **Onda, K., Takeuchi, H., & Okumoto, Y. (1968).** "Mass Transfer Coefficients Between Gas and Liquid Phases in Packed Columns," *J. Chem. Eng. Japan*, 1(1), 56–62.
11. **Ergun, S. (1952).** "Fluid Flow Through Packed Columns," *Chemical Engineering Progress*, 48(2), 89–94.
12. **Wagner, W., & Pruss, A. (1993, 2002).** *IAPWS Formulation 1995 for the Thermodynamic Properties of Ordinary Water Substance*, *J. Phys. Chem. Ref. Data*, 31(2), 387–535.
13. **Poling, B. E., Prausnitz, J. M., & O'Connell, J. P. (2001).** *The Properties of Gases and Liquids*, 5th Ed., McGraw-Hill (Lucas viscosity; Wilke 1950 mixture rule).
14. **NIST Chemistry WebBook** (SRD 69), Gas-Phase Shomate Thermochemistry Tables ($\text{N}_2, \text{CO}_2, \text{H}_2\text{O}, \text{O}_2, \text{HCl}$).
15. **Haaland, S. E. (1983).** "Simple and Explicit Formulas for the Friction Factor in Turbulent Pipe Flow," *J. Fluids Engineering*, 105(1), 89–90.
16. **Green, D. W., & Southard, M. Z. (2019).** *Perry's Chemical Engineers' Handbook*, 9th Ed., McGraw-Hill (Sections 14, 17, 25: Packed Towers, Scrubbers, Corrosion / Alloy C-276 vs. 316L SS in HCl).
17. **Jorgensen, R. (1999).** *Fan Engineering*, 9th Ed., Howden Buffalo; **Bleier, F. P. (1998).** *Fan Handbook*, McGraw-Hill.

---

## Post-Audit Resolution (A1–A4 Fixed)

All four Category A physics and control defects identified in this audit have been resolved while preserving exact `<= 1e-6` relative agreement against all committed `testdata/oracle_fixtures.json` benchmarks (`ErgunPressureDrop` and `CalvertCutDiameter` remain unchanged):

- **A1 (`pkg/scrubber/packedbed.go`):** `SizePackedBed` now evaluates packing bed $\Delta P$ using the Sauter equivalent sphere diameter $d_{p,\text{eq}} = 6(1-\varepsilon)/a_p$ (with specific surface area $a_p = 207.0 \times (0.0254 / d_p)\text{ m}^2/\text{m}^3$ scaled from 1" PP Pall rings) multiplied by the Eckert GPDC irrigated liquid-holdup factor (`1.70`).
- **A2 (`pkg/scrubber/venturi.go`):** Replaced `DefaultCalvertFactor` (`f = 0.5`) in `dpAccel` with the Yung, Barbarika & Calvert (1977) analytical throat droplet-acceleration factor `VenturiDropletAccelFactor(L_t, d_d, v_t, rho_g, rho_L, mu_g)` ($\beta \approx 0.717$ for $d_d = 50\,\mu\text{m}$, $L_t = 80\text{ mm}$), clamped to `(0, 1.0]`.
- **A3 (`pkg/model/types.go`, `internal/control/sim_hal.go`, `configs/kilnscrub.example.json`, `internal/web/static/index.html`):** Raised default `QuenchSetpoint` and `SimHAL` `BaselineQuenchTemp` from the physically unreachable `165.0°F` to `188.0°F` (`359.82 K`), safely above the `183.9°F` adiabatic saturation floor at `2400°F` and below the `210.0°F` CPVC trip limit. Added a load-dependent adiabatic saturation floor (`177.5°F–183.9°F`) in `SimHAL` so closed-loop PID modulates cleanly around `188.0°F` without saturating at 100%, while still tripping `>= 210°F` on water loss or severe unmitigated thermal spikes.
- **A4 (`pkg/scrubber/scrubber.go`):** Corrected `refPeakDeltaTK` to `float64(model.FahrenheitToKelvin(2400.0)) - 293.15` (`1588.7056 - 293.15 K`), eliminating the $+0.043\%$ thermal scaling bias.

### Verified Before / After Metrics (24 cu.ft. Reference Wood/Salt Kiln, 2400°F, 2.0 lbs/hr NaCl)

| Metric | Before (Pre-Fix Baseline) | After (A1–A4 Fixed) | Engineering Validation Target |
| :--- | :---: | :---: | :---: |
| **Inlet / Quenched Gas Flow** | 579.5 / 256.0 ACFM | **579.2 / 255.9 ACFM** | Exact `2400°F` scaling (A4) |
| **Venturi Throat $\Delta P$** | 13.76 inWC (3429 Pa, $\beta=0.500$) | **18.88 inWC (4703 Pa, $\beta=0.717$)** | 18.75–18.88 inWC (Boll / Yung–Calvert) |
| **Packed-Bed $\Delta P_{\text{bed}}$** | 0.094 inWC (23.5 Pa) | **1.44 inWC (360 Pa)** | 1.20–2.07 inWC (Irrigated GPDC) |
| **Stage 3 Total $\Delta P$ (Bed + Demister)** | 0.35 inWC (88 Pa) | **1.70 inWC (425 Pa)** | Realistic irrigated tower + chevron |
| **Blower Total Static Pressure** | 14.75 inWC (3673 Pa) | **21.21 inWC (5284 Pa)** | 20.5–22.5 inWC |
| **Blower Shaft / NEMA Motor HP** | 0.96 / 1.50 HP | **1.38 / 2.00 HP** | 2.0–3.0 HP motor |
| **Default Control `QuenchSetpoint`** | 165.0°F (below $183.9°F$ $T_{\text{sat}}$) | **188.0°F (359.82 K)** | $> 183.9°F$ $T_{\text{sat}}$, $< 210°F$ trip |
| **Oracle Fixtures (`Ergun`, `Calvert d50`, `Shomate`, `Lucas`, `IAPWS-95`)** | Pass ($\le 10^{-6}$) | **Pass ($\le 10^{-6}$)** | Unchanged ground-truth accuracy |
