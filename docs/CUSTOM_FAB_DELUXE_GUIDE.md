# Custom Fabrication Deluxe — Engineering, Fabrication & Installation Manual (`kilnscrub`)

**Purpose:** Shop-ready mechanical, metallurgical, electrical, hydraulic, structural, acoustic, and commissioning manual for building the **Custom Fabrication Deluxe** configuration of the 4-stage `kilnscrub` ceramic kiln exhaust scrubber. Engineered for production studios, university/community ceramic departments, and urban kiln yards demanding multi-decade corrosion resistance, UL 508A-grade electrical safety, whisper-quiet residential acoustics (`38–42 dBA @ 40 ft`), and clean AHJ inspection sign-off under **NEC (NFPA 70)**, **IMC**, **IPC**, **NFPA 86 / 91 / 211**, **OSHA 29 CFR 1910**, and **40 CFR Part 403**.

**Validated Engineering Baseline (`wood_salt_24` Reference — Scales to `48 cu.ft.` Co-Op Kilns):**
- **Reference Kiln Load:** `24.0 cu.ft.` Cone 11 (`2400°F` / `1588.7 K`, `129.6 kW` peak, `579.2 ACFM` hot flue gas, `2.0 lbs/hr` NaCl over a `3.0-hr` active salting window — `9.8 lbs` bulk `NaHCO₃` per firing from `15 lb` bags / `50 lb` sacks).
- **Stage 1 Quench Exit:** `255.9 ACFM` (`0.1208 m³/s`, `0.0974 kg/s`, `ρ = 0.806 kg/m³`) at **`183.9°F`** (`357.5 K`, `55.7% H₂O`), `1.16 GPM` spray (`0.58 GPM` evaporated).
- **Stage 2 Wet Venturi:** `46.9 mm` (`1.85"`) throat at `70 m/s`, `L/G = 0.0012` (`2.30 GPM`), `d_50 = 0.214 µm` (NaCl) / `0.199 µm` (wood ash), `ΔP = 18.88 inWC`.
- **Stage 3 Packed Bed:** `14"` nominal (`12.50"`–`14.0"` ID) filament-wound Derakane vinyl-ester FRP or heavy-wall Sch 80 CPVC/PP column, `1.05 m` (`41.3"`) bed depth, `5.36 GPM` alkaline recirculation, `<= 55%` of Eckert flood, `99.0%` HCl absorption (`1,048 → 10.0 ppm`).
- **Stage 4 Induced-Draft Prime Mover:** `256 ACFM @ 21.21 inWC`, `1.38 shaft HP` → **`2.0 HP` 3-phase TEFC motor** driven by a `240V` single-phase-in VFD (`30A` circuit, `#10 AWG` per NEC 430.122(A)). *(For `48 cu.ft.` co-op kilns: step to `3.0 HP` motor, `GS21-23P0` VFD, `240V 40A` circuit, `#8 AWG`.)*
- **Design & CAD Sources:** Flat-pattern sheet-metal developments and flange bolt circles come directly from `scrubber.dxf` and `shop_plan.json` ([pkg/studio/plan.go](../pkg/studio/plan.go), [pkg/studio/cots.go](../pkg/studio/cots.go), [pkg/studio/guide.go](../pkg/studio/guide.go)).

> **Code & Safety Notice:** Model codes (NEC, IMC, IPC, NFPA) are enforced with state and local amendments. Submit this manual, your `shop_plan.json` / `scrubber.dxf` packet, and the panel electrical schematic to your local Authority Having Jurisdiction (AHJ) and Publicly Owned Treatment Works (POTW) prior to slab pouring or utility tie-in.

---

## 1. Custom Fabrication Deluxe Architecture & Metallurgy

The Custom Fabrication Deluxe build eliminates light-gauge slip joints, thin-walled drums, and improvised seals in favor of fully TIG-welded corrosion-resistant alloy hot stages (`Stages 1–2`) bolted via **ANSI/ASME B16.5 Class 150** flanges to a vacuum-rated **Derakane vinyl-ester FRP** (or gusseted heavy-wall PP/CPVC) absorber column (`Stage 3`) and an acoustically baffled **Stage 4** radial pressure blower.

```mermaid
flowchart LR
    Kiln["2400°F Kiln Flue\n(Ceramic Fiber + SS Jacket\nNFPA 211 >=36\" Clearance)"] --> Tee["Refractory/316L Bypass Tee\n+ Belimo Fail-Open Damper\n-> NFPA 211 Bypass Chimney"]
    Tee --> S1["Stage 1: TIG-Welded 316L / C-276\nAdiabatic Quench Tower\n(Sight Glass + NFPA 91 Cleanout)"]
    S1 -->|"183.9°F + 200°F Snap-Disc"| S2["Stage 2: TIG-Welded 316L / C-276\nWet Venturi (25° / 7° Cones)\nANSI Class 150 Flanges + ePTFE"]
    S2 -->|"ANSI 150 Flange"| S3["Stage 3: Filament-Wound Derakane FRP\nor Gusseted PP/CPVC Absorber\n(Manway + Removable Chevron Cassette)"]
    S3 -->|"Integral Sloped Baffled Sump"| Pump["Flooded-Suction Mag-Drive Pump\n+ 20-40 Mesh Y-Strainer\n+ 3-Branch Balancing Manifold"]
    Pump -.->|"Alkaline pH 8.0-8.5 Recirc"| S1
    Pump -.->|"Alkaline Recirc"| S2
    Pump -.->|"5.36 GPM"| S3
    S3 -->|"Flex PTFE/EPDM Boot"| S4["Stage 4: 2.0 HP Radial ID Blower\nin 2\" Rockwool Baffled Doghouse\n(Neoprene Mounts + Z-Cooling Ducts)"]
    S4 -->|"Flex Boot"| Sil["Inline Absorptive Stack Silencer\n(Tedlar-Lined Wool, 1/2\" Drain)"]
    Sil --> Stack["Outdoor Vertical Stack\n>=12-15 ft H, IMC 501.3.1\n>=30 ft Property Line (Salt)"]
```

### 1.1 Metallurgy & Corrosion Tier Selection (Stages 1 & 2)

Hot wet hydrogen chloride (`1,000–4,100 ppm` upstream HCl condensing across `184°F` wet-bulb saturation) is aggressive toward austenitic stainless steel whenever pH drops below `6.0` or chloride brine concentrates inside weld crevices. Choose metallurgy by duty cycle:

| Component / Zone | Operating Environment | Standard 316L SS Deluxe *(Occasional Studio Firing + Mandatory Alkaline Recirc)* | Heavy-Duty Alloy C-276 / FRP Deluxe *(Institutional / Frequent Salt Firing)* |
| :--- | :--- | :--- | :--- |
| **Inlet Hot Flue & Bypass Tee** | Dry `1800°F–2400°F` flue gas, thermal cycling, embers | `14 ga` (`0.075"` / `1.90 mm`) **316L SS (UNS S31603)** or **310S / Inconel 600** hot liner + ceramic-fiber blanket & SS jacket | **Inconel 600 / 310S** hot section or castable high-alumina refractory-lined 316L shell + ceramic-fiber & SS jacket |
| **Stage 1 Quench Vessel Shell** | `2400°F → 183.9°F` thermal shock, wet spray, `pH 8.0–8.5` buffered liquor | `14 ga` (`1.90 mm`) or `12 ga` (`2.66 mm`) **316L SS** rolled cylinder; bolted replaceable spool; inspect welds every `5–10` salt firings | `12 ga` (`2.66 mm`) **Alloy C-276 (UNS N10276)** OR acid-proof castable refractory-lined 316L shell; inspect every `20–25` salt firings |
| **Stage 2 Venturi Cones & Throat** | `184°F`, **`70 m/s`** gas + abrasive ash/droplet impingement, `-20 inWC` | `12 ga` (`2.66 mm`) **316L SS** converging/diverging cones (`25°`/`7°` from `scrubber.dxf`) + heavy **Sch 40S 316L throat spool** (`0.154"` wall) | **Alloy C-276 (`1/8"` / `3.2 mm`)** throat & cones (or removable **reaction-bonded SiC ceramic throat sleeve** inside C-276 housing) |
| **Stage 3 Absorber Column & Sump** | `184°F`, wet `1,048 → 10 ppm` HCl, `-21 inWC` vacuum, `pH 7.5–8.5` | Gusseted heavy-wall **Sch 80 CPVC (`ASTM F441`, `200°F`)** or **`1/2"` extrusion-welded PP** with external vacuum stiffening rings | **Filament-wound Derakane 411-350 / 470-300 vinyl-ester FRP** (`100 mil` C-veil barrier, **ASTM E84 Class 1**, rated `-25 inWC` vacuum & `210°F`) |
| **Stage 4 Blower Wheel & Scroll** | `184°F` saturated air, `10 ppm` HCl mist, `3,450 RPM` tip stress | **316L SS** welded radial wheel inside epoxy-phenolic coated steel scroll (rinse `5 min` with pH 8 water after every salt firing) | **Solid Derakane vinyl-ester FRP** scroll & chemically resistant FRP/C-276 wheel (immune to chloride stress-corrosion cracking) |

> **316L Salt-Service Operating Rule:** Whenever Stage 1 and Stage 2 are fabricated in **316L SS** for salt glazing, **always** supply quench and venturi spray from the buffered **`pH 8.0–8.5` `NaHCO₃` recirculation manifold** (never unbuffered acidic tap water) and execute a 5-minute clean buffered rinse during cooldown so wet chloride films never sit stagnant in crevices between firings.

### 1.2 Back-Purged TIG (GTAW) Welding Specification for Chloride Resistance

Standard open-root stainless welds form granular chromium-depleted oxide ("sugaring") on the interior root face. In wet `184°F` HCl/NaCl service, sugared root seams fail by pitting and crevice corrosion within 2–5 firings. Every wetted 316L SS and Alloy C-276 weld in Stages 1 and 2 must follow this shop WPS:

1. **Joint Preparation & Fit-Up:** Shear or laser-cut flat patterns from `scrubber.dxf`. Bevel plate `>= 12 ga` (`> 2.5 mm`) to a `30°` single-V (`60°` included) with `1/16"` (`1.5 mm`) root land. Degrease with acetone and wire-brush strictly with **dedicated unused 316 SS bristles** (never carbon-steel brushes or grinding wheels that embed free iron!).
2. **Mandatory Argon Back-Purging:**
   - Seal vessel ends with water-soluble purge paper or silicone purge dams and flood the interior with **99.995% (Grade 4.5/5.0) pure Argon** at `15–25 CFH` until oxygen drops to **`<= 50 ppm O₂`** (verified with a purge monitor; silver-to-straw interior root color, zero grey/black coking).
   - Maintain `5–10 CFH` positive purge flow with a `1/4"` pressure-relief vent hole throughout the root and hot pass so internal pressure does not cause root concavity.
3. **Filler Metal & Heat Input:**
   - **316L SS base metal:** Use **AWS A5.9 ER316L** (`<= 0.03% C` prevents sensitization) — or upgrade root and cap passes to **ERNiCrMo-3 (Alloy 625)** or **ERNiCrMo-4 (Alloy C-276)** over-alloyed filler for dramatically higher weld-seam pitting resistance (`PREN > 50`).
   - **Alloy C-276 base metal:** Strictly **AWS A5.14 ERNiCrMo-4**.
   - **Interpass Temperature:** Keep **`<= 300°F`** for 316L SS and **`<= 200°F`** for Alloy C-276 (use Tempilstik crayons). Use stringer beads only — never heavy weaving, which promotes sigma-phase and carbide precipitation in the heat-affected zone (HAZ).
4. **Full-Penetration Crevice-Free Geometry:** All internal seams (shell longitudinal seams, venturi cone-to-throat transitions, nozzle bosses, and sight-glass stubs) must be **100% full-penetration flush welds** with zero overlapping lap ledges facing the gas stream. Where interior access permits, blend internal root beads smooth (`<= 63 Ra µin`) with clean aluminum-oxide flap discs.
5. **Post-Weld Pickling & Passivation (ASTM A380 / A967):**
   - Apply commercial nitric-hydrofluoric pickling gel to all heat-tinted weld zones inside and out to dissolve chromium-depleted thermal oxides, scrub with nylon bristle, and flush thoroughly with low-chloride water.
   - Follow with **citric or nitric passivation** per ASTM A967 to regenerate a uniform `20–30 Å` passive `Cr₂O₃` film; confirm absence of embedded carbon-steel contamination with a copper-sulfate swab test.

### 1.3 ANSI/ASME B16.5 Class 150 Flanges, Fasteners & PTFE Envelope Gaskets

All modular spool breaks (Stage 1 top/bottom, Stage 2 inlet/outlet, Stage 3 sump/mid/top, and blower transitions) use **ANSI/ASME B16.5 Class 150** bolt patterns so standard gaskets, blind plates, and replacement spools bolt up identically:

| Nominal Pipe / Shell Size | Flange Type (Metal Stages 1–2) | Flange Type (FRP / CPVC Stage 3) | ANSI Class 150 Bolt Circle | Bolt Count × Size | Recommended Torque (ePTFE / Envelope Gasket) |
| :---: | :--- | :--- | :---: | :---: | :---: |
| **`6"`** *(Inlet/Stack)* | 316L/C-276 Weld-Neck or Lap-Joint Stub End + Backing Ring | Sch 80 CPVC / FRP Van Stone Loose-Ring | `9.50"` (`241.3 mm`) | `8 × 3/4"` *(or `5/8"` FRP)* | `25–35 ft-lbs` *(FRP/CPVC)* / `45 ft-lbs` *(metal)* |
| **`8"`** *(Duct/Stack)* | 316L/C-276 Weld-Neck or Lap-Joint (Van Stone) | Sch 80 CPVC / FRP Van Stone | `11.75"` (`298.5 mm`) | `8 × 3/4"` | `30–40 ft-lbs` *(FRP/CPVC)* / `50 ft-lbs` *(metal)* |
| **`14"`** *(Quench / Col.)* | 316L/C-276 Lap-Joint Stub End + 304SS/Epoxy Backing Ring | Filament-Wound FRP Hand-Layup or Sch 80 CPVC Van Stone | `18.75"` (`476.3 mm`) | `12 × 1"` *(use `5/8"`–`3/4"` studs on plastic)* | `30–45 ft-lbs` *(FRP/CPVC)* / `60 ft-lbs` *(metal)* |
| **`18"`** *(Co-Op / Eductor)* | 316L/C-276 Lap-Joint Stub End | Filament-Wound Derakane FRP Full-Face | `22.75"` (`577.9 mm`) | `16 × 1-1/8"` *(use `3/4"` studs on FRP)* | `35–50 ft-lbs` *(FRP full-face)* |

- **Why Lap-Joint (Van Stone) Flanges Excel on Deluxe Builds:** Rotating backing rings make aligning heavy rolled cones and FRP columns effortless during assembly, allow cheaper 304 SS or epoxy-coated backing rings behind a thin wetted **316L or C-276 lap-joint stub end**, and eliminate torsional twist on fragile FRP nozzles.
- **Fasteners:** Use **ASTM A193 Grade B8M (316 SS)** hex bolts/studs, heavy hex nuts, and **oversized flat washers** under every nut and bolt head bearing on FRP/CPVC flanges. Coat threads with **nickel-based chloride-free anti-seize** (prevents stainless galling).
- **Gaskets (Strict Thermal Boundary!):**
  - **Downstream of Stage 1 Quench (`<= 200°F`, all wet flanged joints):** Use `1/8"` (`3.2 mm`) **PTFE envelope gaskets** (slit virgin PTFE jacket over resilient EPDM/compressed-fiber core — provides PTFE chemical immunity with low seating stress ideal for FRP/CPVC flanges!) or **`1/8"` expanded PTFE (ePTFE) full-face gaskets** (e.g., Gore-Tex / Garlock Gylon 3545). Tighten in a 3-pass star pattern (`30% → 60% → 100%`) and re-torque after 24 hours to account for PTFE cold-flow creep.
  - **Upstream of Stage 1 Quench (`1800°F–2400°F` dry hot flue):** **NEVER use PTFE (`> 500°F` emits toxic perfluoroisobutylene!) or `650°F` RTV silicone.** Use exclusively **`2300°F–2600°F` braided alumina-silica ceramic fiber rope gaskets** (`3/8"`–`1/2"` thick) or high-temperature refractory mortar.

### 1.4 Borosilicate Sight-Glass Ports & Bolted NFPA 91 Cleanout Doors

- **Borosilicate Sight-Glass Ports (2 Locations Minimum):**
  1. **Stage 1 Quench Spray Zone** (mounted `12"`–`18"` below top nozzle header, `90°` offset from spray lances) to confirm uninterrupted fine full-cone atomization.
  2. **Stage 2 Venturi Exit / Stage 3 Lower Plenum** to verify active liquid shearing and check for ash buildup.
  - **Spec:** `3"` or `4"` bolted ANSI weld-pad or sanitary tri-clamp **tempered borosilicate sight glass** (ASTM E438 Type I Class A, rated `>= 450°F` thermal shock & `>= 100 psig`), cushioned on both faces with `1/16"` virgin PTFE ring gaskets (zero glass-to-metal contact!). Add a `1/8"` NPT internal **clean-water lens-wash jet** (`5 sec` push-button valve) so wood fly ash never blinds the glass mid-firing.
- **Bolted NFPA 91 Cleanout & Access Doors:**
  - Per **NFPA 91**, install bolted, gasketed cleanout doors (`6" × 8"` minimum rectangular frame with `1/4"` reinforcement doubler plate, OR `8"` ANSI blind-flanged tee) at:
    1. Every horizontal hot-flue elbow and junction upstream of Stage 1 (ceramic fiber gasket, captive 316 SS studs + heavy hex nuts).
    2. Base of Stage 1 quench vessel (`6"` above bottom drain) for heavy wood-ash/char removal.
    3. Transition elbow between Stage 2 diffuser and Stage 3 column (ePTFE gasket, `316 SS` toggle latches or bolted cover).
  - Never use light sheet-metal snap-tabs or duct tape; every door must hold `-25 inWC` suction airtight during the smoke test.

### 1.5 `2400°F` Hot Inlet Flue Insulation, SS Jacketing & NFPA 211 Clearances

Uninsulated `2400°F` stainless flue pipe glows bright cherry-orange (`1300°F–1600°F` skin temperature), creating severe operator burn hazards, studio heat gain, and structural ignition risk:

1. **Ceramic-Fiber Insulation Blanket:** Wrap the entire hot flue run (kiln damper slot → bypass tee → Stage 1 top flange, plus the vertical bypass riser) with **two staggered `1"`–`1.5"` layers (`2"`–`3"` total)** of **`8 lb/ft³` `2300°F–2600°F` alumina-silica ceramic fiber blanket** (e.g., Unifrax Fiberfrax Durablanket S / Morgan Superwool Plus low-biopersistence alkaline-earth silicate wool). Stagger longitudinal and circumferential butt joints by `>= 6"` and tie with `18 ga` Nichrome/316 SS wire every `12"`. *(Drops outer skin temperature from `~1450°F` down to `<= 150°F–185°F`.)*
2. **Weatherproof Stainless Steel Jacketing:** Encase the ceramic blanket in **`24 ga` or `26 ga` (`0.020"`–`0.024"`) Type 304/316 SS stucco-embossed or smooth metal jacketing** with `2"` overlapped weather-shedding seams facing downward, `1/2"` 316 SS banding straps + wing seals every `12"`, and a **`1/2"` sliding thermal expansion slip-joint** at every `6 ft` of run. Seal exterior outdoor weather laps with high-temp outdoor aluminum/silicone caulk strictly on the **cold outer jacket** (never inside the hot flue seam!).
3. **Mandatory NFPA 211 High-Heat Flue Rules:**
   - Maintain **`>= 36 inches` unreduced clearance** from the hot flue and auxiliary bypass chimney to any unprotected combustible framing, siding, fence, or roof eaves (reducible only via engineered NFPA 211 ventilated masonry/sheet-metal heat shields approved by the AHJ).
   - **Never** route a `> 1000°F` high-heat chimney connector through a combustible wall, roof, or partition.
   - Terminate the **NFPA 211 auxiliary emergency bypass chimney** (refractory-lined masonry or UL-listed high-heat factory chimney) above the roofline per the **3-2-10 ft rule** (`>= 3 ft` above roof penetration and `>= 2 ft` higher than any roof/structure within `10 ft`).

### 1.6 Stage 3 Deluxe Absorber Column: Derakane FRP / Gusseted PP-CPVC Internals

For `14"`–`18"` columns handling `183.9°F` wet HCl under **`-21 inWC` induced-draft suction**, specify either **filament-wound Derakane 411-350 / 470-300 vinyl-ester FRP** (preferred for institutional/heavy salt duty) or **gusseted heavy-wall Schedule 80 CPVC / `1/2"` welded PP**:

- **FRP Laminate Schedule (ASTM D4097 / RTP-1):**
  - Inner corrosion liner (`100 mil` / `2.5 mm` minimum): `1 ×` Nexus/C-glass surfacing veil (`10 mil`, 90% resin) + `2 ×` `1.5 oz/ft²` chopped-strand mat (`90 mil`, 75% resin) using **brominated fire-retardant Derakane 510A-40 or 510C-350** (achieves **ASTM E84 Class 1 `FSI <= 25 / SDI <= 50`** / FM 4922 per IMC 510 & NFPA 91).
  - Structural wall: Filament-wound continuous E-glass roving (`55°` helix angle, minimum `0.25"`–`0.31"` total wall for `14"`–`18"` bore) with external circumferential **FRP box-stiffening rings** (`1.5" × 1.5"` foam-cored ribs every `30"`) guaranteeing buckling safety factor **`>= 5.0` at `-25 inWC` vacuum (`185°F` wet modulus)**.
  - *Alternative Gusseted PP/CPVC Shell:* If fabricating from `1/2"` stress-relieved polypropylene sheet (hot-air extrusion welded) or `14"` Sch 80 CPVC (`0.750"` wall), weld external `1/2" × 2"` circumferential stiffening rings every `24"` and `8` triangular base gussets (`6" × 6" × 1/2"`) tying the shell into the bottom support flange.
- **Bolted Packing Access Manway (`12"`–`16"` Side Port):**
  - Install a flanged **`12"` (on `14"` column) or `16"` (on `18"` column) side access manway** centered `6"` above the lower packing support grate (with a matching `10"`–`12"` upper inspection port above the packing hold-down grid).
  - Fitted with a hinged davit arm (optional), `1/8"` PTFE envelope gasket, and `316 SS` studs so `3–4 cu.ft.` of `1"` PP Pall rings or heavier kiln-fired stoneware rings can be raked out, washed, or re-packed in 20 minutes without unbolting overhead ductwork.
  - **Reinforced Packing Support Grate:** Rest packing on a `1.5"` deep **molded Derakane FRP pultruded bar grating** (`>= 70%` open area, `1.5" × 1.5"` mesh overlaid with `1/2"` PP screen) supported by a `360°` welded internal ledge ring (`1.5"` wide) — sized to carry `>= 300 lbs` wet stoneware ceramic packing without creeping at `184°F`. Pin a matching top hold-down grid so wood-stoking surges cannot fluidize light PP rings.
- **Removable 3-Pass Chevron Mist-Eliminator Cassette:**
  - Fabricate the `3-pass` polypropylene chevron blade pack (`4"`–`6"` deep, `1.0"` blade pitch, `0.26 inWC` pressure drop) inside a rigid bolted/sliding **PP/FRP cartridge frame** seated on an internal perimeter EPDM knife-edge seal.
  - Provide either a **side-withdrawal bolted rectangular cassette door** (with stainless cam-latches and ePTFE gasket) OR a **split top flange spool** right above the liquid distributor. Include an upward/downward `1/4"` NPT **clean-water demister flush nozzle** (`1.5 GPM` intermittent wash) to rinse dried salt bridges without removing the cassette.
- **Integral Sloped-Bottom Baffled Settling Sump:**
  - Mold or weld the base of Stage 3 (or its directly coupled skid sump tank) with a **`>= 1/4"` per foot (`2°–3°`) sloped bottom** pitching cleanly toward a `2"` full-port flush/sludge cleanout valve.
  - Divide the `40–55 gal` working basin with a **removable vertical over/under PP/FRP weir baffle**:
    - **Zone 1 (Dirty Return & Ash Settling — `60%` volume):** Receives gravity drainage from Stage 1 quench, Stage 2 sweep/venturi, and Stage 3 packed bed while wood fly ash settles cleanly to the sloped bottom floor. **Mandatory Hydrostatic / Barometric Seal Sizing:**
      1. *Stage 1 Quench (`-1 to -2 inWC`) & Dry-Fan Eductor Mode (`+2 to +6 inWC`):* A **`>= 8"`** submerged downcomer / U-trap into an atmospheric vented sump provides ample hydrostatic seal margin.
      2. *Stage 2 & Stage 3 Under `-21 inWC` Induced-Draft Vacuum:* Atmospheric pressure pushes sump liquor **`21 inches` vertically up** any open drain pipe (instantly breaking a shallow `6"`–`8"` seal!). Use one of two engineered solutions:
         - **Barometric Seal Leg into Atmospheric Vented Sump:** Elevate the Stage 3 column drain nozzle so the vertical submerged downcomer / U-trap holds **`>= 26"`–`28"` of vertical water column** (`|ΔP| + 5"` safety margin against `-21 inWC` suction) before entering the vented sump; include a `2"` screened atmospheric vent, hinged top `NaHCO₃` dosing hatch (`15 lb` bags / `50 lb` sacks), and IPC 608 overhead Air-Gap makeup standpipe.
         - **Integral Vacuum-Sealed Stage 3 Sump:** Fabricate the sloped baffled sump as a vacuum-rated (`-25 inWC`) sealed extension directly beneath Stage 3 with **no atmospheric vent** (so column liquor drains freely by gravity at equalized `-21 inWC` pressure into the flooded-suction mag-drive pump; dose bulk `NaHCO₃` through a bolted gasketed hatch when shut down and feed makeup from an external IPC 608 air-gapped break tank via a check-valved transfer line).
    - **Zone 2 (Clean Alkaline Suction — `40%` volume):** Clarified `pH 8.0–8.5` liquor overflows the weir through a slide-out **`50-mesh` 316L SS vertical screen** into the clean pump suction chamber (`4"` bulkhead stand-off above floor).
  - Include a side glass/PFA sight-level tube (`min`/`working`/`overflow` marks) rated for full sump vacuum/pressure and verify flooded suction head (`>= 6"`–`12"`) above the mag-drive pump centerline.

---

## 2. Electrical Layout, Conduit Routing & NEC Clearances

Every electrical component in the Custom Fabrication Deluxe build follows **UL 508A** industrial control panel practice and **NEC (NFPA 70)** Articles **110.26, 110.28, 210.8(F), 400.12, 430.22, 430.102, and 430.122(A)**.

### 2.1 NEC 430.122(A) VFD Conductor, Breaker & Disconnect Sizing

Per **NEC 430.122(A)**, line-side supply conductors feeding a Variable Frequency Drive (VFD) must be sized to **at least `125%` of the VFD's rated single-phase input current** (from drive nameplate/manual — **never** motor full-load current alone!), plus continuous auxiliary loads (mag-drive recirculation pump + `24VDC` controls + optional downstream plume diluter). Single-phase diode rectifiers draw high harmonic RMS current (`PF ≈ 0.55–0.65`), so input current is `1.5×–1.9×` higher than naive horsepower estimates:

| Blower Motor HP (3-Ph Output) | Supply Service (1-Ph Input) | VFD Model (`GS21` in NEMA 4X or `GS21X` Washdown) | VFD Rated 1-Ph Input Amps (VT) | NEC 430.122(A) `125%` VFD + Aux Pump/Controls | Min Copper Conductor (THWN-2, `75°C`) | NEC Breaker (Line-Side HF-GFCI) | NEC 430.102 Lockable Sight Disconnect |
| :---: | :---: | :--- | :---: | :---: | :---: | :---: | :---: |
| **`0.5 HP`** | `120V` 1-Ph | `GS21-10P5` | `11.1 A` | `13.9 + 3.0 = 16.9 A` | `#12 AWG` | `20A` 1-Pole HF-GFCI | `30A` NEMA 4X Non-Fused |
| **`1.0 HP`** | `240V` 1-Ph *(rec.)* | `GS21-21P0` | `9.7 A` | `12.1 + 1.5 = 13.6 A` | `#12 AWG` | `20A` 2-Pole HF-GFCI | `30A` NEMA 4X Non-Fused |
| **`2.0 HP`** *(24 cu.ft. Ref.)* | **`240V` 1-Ph** | **`GS21-22P0`** *(or `GS21X-22P0`)* | **`18.5 A`** | **`23.1 + 1.5 = 24.6 A`** | **`#10 AWG`** | **`30A` 2-Pole HF-GFCI** | **`30A` 2-Pole NEMA 4X Through-Door** |
| **`3.0 HP`** *(48 cu.ft. Co-Op)* | **`240V` 1-Ph** | **`GS21-23P0`** *(or `GS21X-23P0`)* | **`27.5 A`** | **`34.4 + 1.5 = 35.9 A`** | **`#8 AWG`** | **`40A` 2-Pole HF-GFCI** | **`60A` 2-Pole NEMA 4X Through-Door** |

> **Plume Diluter Auxiliary Circuit Note:** If installing the optional `120V` squirrel-cage plume dilution blower (`1,000–3,000 CFM`, drawing `4.1–12.5 A @ 120V`), feed it on a dedicated **`120V 20A` GFCI branch circuit** (or via an internal `15A` supplementary control breaker off a 4-wire `120/240V` feeder with neutral) so its startup inrush never crowds the VFD feeder.

### 2.2 HF-GFCI Compatibility & 3% Input Line Reactor (NEC 210.8(F))

Under **2020/2023 NEC 210.8(F)**, outdoor dwelling-unit outlets (including hardwired `<= 50A` / `<= 150V`-to-ground single-phase equipment) require Class A (`5 mA`) GFCI protection. Standard PWM VFDs generate high-frequency common-mode capacitive leakage current across shielded motor cables (`3–15 mA`), tripping ordinary residential GFCI breakers. Build the Deluxe power train with all four mitigation layers:
1. **Line-Side HF-Rated GFCI Breaker:** Install a high-frequency-compatible Class A GFCI breaker (e.g., Square D `QO230GFI` / Siemens QF230A HF-rated series) strictly on the **line side** upstream of the VFD (never on the load side!).
2. **`3%` Impedance Copper-Wound Input Line Reactor:** Wire an AutomationDirect **`LR2-22P0`** (2 HP) or **`LR2-23P0`** (3 HP) `3%` line reactor immediately between the through-door disconnect and VFD `L1/L2` terminals. The reactor attenuates high-frequency rectifier spikes, reduces harmonic heating, protects VFD input diodes from utility surges, and dramatically lowers GFCI nuisance tripping.
3. **Short Shielded VFD Motor Cable (`<= 25 ft`):** Use low-capacitance XLPE-insulated shielded VFD cable (`12/4` copper + foil/85% braid + dual drain wires, e.g., Belden `29501` or AutomationDirect `CF12-4-VFD`). Lower VFD carrier frequency (`P00.17` on GS21) to **`4–6 kHz`** (balance between low capacitive leakage and acceptable motor audible whine inside the acoustic doghouse).
4. **360° Shield Grounding:** Terminate the VFD motor cable shield via metallic EMC grounding clamps at **both** the VFD PE ground plate and the motor terminal box frame.

### 2.3 UL 508A-Style NEMA 4X Panel Layout & Through-Door Disconnect

House all controls inside a corrosion-proof **NEMA 4X 316L SS or fiberglass enclosure** (minimum **`24" H × 20" W × 10" D`** for `GS21` inside panel, or `20" × 16" × 8"` if mounting a washdown `GS21X-22P0` VFD externally beside the box) per **NEC 110.28**. Mount a **lockable rotary through-door disconnect handle** (rated `30A` or `60A`, padlockable in OFF with 3 lock holes for OSHA 1910.147 LOTO) directly on the panel door, sited **in sight and `<= 50 ft`** of both the VFD and Stage 4 blower (**NEC 430.102(A)/(B)**). *(Never open a load-side disconnect between a running VFD and motor without early-break auxiliary contacts wired into VFD STO/Stop, or IGBT failure can result.)*

```text
================================================================================
  UL 508A-STYLE NEMA 4X CONTROL PANEL SUBPANEL LAYOUT (24" H x 20" W x 10" D)
  [HIGH-VOLTAGE 240VAC BAY — LEFT]   |   [LOW-VOLTAGE 24VDC / MODBUS BAY — RIGHT]
================================================================================
+------------------------------------+-----------------------------------------+
| [THROUGH-DOOR LOCKABLE DISCONNECT] | [24VDC DIN PSU: Mean Well NDR-120-24]   |
|  30A 2P (NEC 430.102, <=50 ft)     | [5.1VDC ISOLATED BUCK: DDR-15G-5]       |
|  L1/L2 In -> Fuse/CB Dist. Block   | [F1: 2A Safety CB] [F2: 3A Aux/Pi CB]   |
|                                    |                                         |
| [3% INPUT LINE REACTOR: LR2-22P0]  | [HARDWIRED NFPA 86 SAFETY RELAY K_SAFE] |
|  (Top vented clearance >= 3")      |  Force-Guided Latching (SR22-24 / Pilz) |
|                                    |  * Overrides Pi/PLC via VFD STO +       |
|  ====== 2"x2" SLOTTED DUCT ======  |    Belimo Spring-Return Bypass Damper!  |
|                                    |                                         |
| [GS21-22P0 2 HP VFD (IP20)]        |  ======== 2"x2" LOW-VOLT DUCT ========  |
|  >=4" Top/Bot Cooling Clearance!   |                                         |
|  (Or mount GS21X-22P0 outside N4X  | [RASPBERRY PI 4B / CLICK PLC]           |
|   to remove ~85W heat from box)    |  (OverlayFS RO + HW Watchdog)           |
|  DI1/STO <-- K_SAFE Contact #2     | [ISOLATED USB-RS485 (FT232RL 2500Vrms)] |
|                                    | [MODBUS RTU I/O: 4-20mA / TC / RTD / DO]|
|  ====== GROUNDED PE BUS BAR =====  |  ==== INSTRUMENT SHIELD GROUND BUS ===  |
| [240V Power & VFD Motor Glands]    | [24V Safety Glands] [RS-485/Analog Gl.] |
+-----(CONDUIT A: HV POWER)----------+---(CONDUIT B)---------(CONDUIT C)-------+
       ^                                   ^                       ^
       | >= 6" (150 mm) parallel spacing   |   90° crossings ONLY  |
       +-----------------------------------+-----------------------+

DOOR EXTERIOR (NEMA 4X):
  • Lockable Rotary Disconnect Handle (Padlockable OFF, mechanical door defeat)
  • Red Twist-to-Release E-Stop Mushroom (NC) + Illuminated Blue Manual Reset PB
  • 24VDC Red Trip Strobe / Piezo Horn + Analog Dwyer Magnehelic (-1.0 to +0.2 inWC)
```

> **Panel Thermal Sizing Check:** A `2.0 HP` `GS21-22P0` VFD dissipates `~75–95 W` at full load, plus `~25 W` for the 24VDC PSU and Pi/I/O (`~115 W` total). Inside an unvented `24"×20"×10"` fiberglass box (`~14.2 ft²` surface area), `115 W` raises internal temperature `~28°F–35°F` above ambient. If outdoor summer ambient exceeds `85°F` (`> 115°F` internal), either: (a) mount a **standalone washdown `GS21X-22P0` VFD** directly outside the box (cutting internal heat to `< 30 W`), OR (b) use a **316L SS enclosure** with a NEMA 4X heat exchanger / filtered thermostatic cooling fan (`>= 95°F` thermostat) positioned on the clean side away from acid mist.

### 2.4 Conduit Routing, Physical Separation & EMI Hygiene (NEC 400.12)

Per **NEC 400.12**, flexible extension cords are strictly prohibited for permanent scrubber power. Route all field wiring inside **Schedule 80 PVC rigid conduit** or **Liquid-Tight Flexible Nonmetallic Conduit (`LFNC-B`)** with corrosion-proof nylon/316L NEMA 4X liquid-tight glands entering exclusively through the **bottom** of the enclosure (drip loops prevent water tracking into live terminals!). Maintain **strict three-conduit physical separation**:

1. **Conduit A — High-Voltage Power (`3/4"` Sch 80 PVC / LFNC-B):**
   - Carries `240VAC` `#10 AWG` THWN-2 feeder from studio subpanel to through-door disconnect, and `#12/4` shielded VFD motor cable (`U/V/W/PE`) from VFD to the Stage 4 blower motor peckerhead (finish the last `18"`–`24"` at the motor with a slack loop of `LFNC-B` for vibration decoupling!).
   - Also carries the dedicated `120VAC` recirculation pump circuit (GFCI protected, `14/3` or `#12 AWG` THWN-2).
2. **Conduit B — Hardwired `24 VDC` NFPA 86 Safety Loop (`1/2"` Sch 80 PVC / LFNC-B):**
   - Carries `18 AWG` stranded MTW/THWN conductors for the **de-energize-to-trip safety string**: `E-Stop (NC)` → `200°F` Manual-Reset Snap-Disc (`NC` at quench exit) → `0.25 GPM` Mechanical Paddle Flow Switch (`NC` on quench header) → `K_SAFE` latching coil → `Belimo NFB24-SR` spring-return fail-open bypass damper + VFD hardwired `STO/DI1`.
   - Physically separated **`>= 6 inches` (`150 mm`)** from Conduit A along all parallel runs.
3. **Conduit C — Low-Voltage Instrumentation & RS-485 Modbus (`1/2"` Sch 80 PVC / LFNC-B):**
   - Carries `Belden 3105A` (`22 AWG` shielded twisted pair + `120Ω` terminators at both bus ends) for RS-485 Modbus RTU, shielded `20 AWG` twisted pairs for `4–20 mA` transmitters (Dwyer `±2 inWC` draft, isolated sump pH, Hall flowmeter), Type-K thermocouple extension wire (yellow/red ANSI color code — never splice with copper!), and 3-wire shielded Pt100 RTD cable.
   - **Single-Point Shield Grounding:** Bond RS-485 and `4–20 mA` cable shields **strictly at the control panel instrument PE bus ONLY** (tape/insulate the field-end drain wire). This eliminates ground-loop offsets on the submerged sump pH electrode and prevents VFD switching noise from corrupting Modbus frames.
   - **Crossing Rule:** Where Conduit B or C must cross Conduit A, cross strictly at **`90°` perpendicular angles**.

### 2.5 Mandatory NEC 110.26 `36" × 30"` Electrical Working Clearance

Do not tuck the NEMA 4X control panel behind the Stage 3 column or above the chemical sump drum! **NEC Section 110.26(A)** mandates dedicated, permanently unobstructed electrical working space directly in front of the enclosure:

```text
                         +-------------------------------+
                         |  NEMA 4X PANEL (20"-24" W)    |
                         |  Bottom >= 36", Top <= 78"    |
                         |  Disconnect Handle <= 6'-7"   |
                         +---------------+---------------+
                                         |  Door swings >= 90°
                  |<-------- >= 30" MIN WIDTH ---------->|
                  +--------------------------------------+ ---
                  |                                      |  ^
                  |   MANDATORY NEC 110.26 WORKING ZONE  |  | >= 36" MIN DEPTH
                  |   Paint Yellow/Black Floor Stripes   |  | (42" if grounded
                  |                                      |  |  masonry wall
                  |   NO SUMP DRUM / NO WOOD PILES /     |  |  behind operator)
                  |   NO SODA BAGS / NO PIPING ALLOWED!  |  v
                  +--------------------------------------+ ---
                  * Headroom: >= 6'-6" (78") clear vertical height
                  * Illumination: Permanent >= 10 fc work light above panel
```

| NEC 110.26 Parameter | Mandatory Minimum Requirement | Deluxe Installation Standard |
| :--- | :--- | :--- |
| **Depth of Working Space (`110.26(A)(1)`)** | **`36"` (`3.0 ft`)** for `<= 150V` to ground (`120/240V` split-phase, Condition 1); **`42"`** if exposed grounded concrete/brick wall is directly behind the operator (Condition 2) | Reserve **`42"` clear depth** painted on slab so AHJ approves under either Condition 1 or 2 |
| **Width of Working Space (`110.26(A)(2)`)** | **`30"` (`2.5 ft`)** or width of enclosure, whichever is greater (does not need to be centered, but must allow `>= 90°` door swing) | **`36"` clear width** centered on panel + unobstructed `110°` door swing |
| **Headroom (`110.26(A)(3)`)** | **`6'-6"` (`78"`)** from floor/pad grade | Zero overhead ducts, pipes, or rain gutters intruding below `78"` |
| **Disconnect Handle Height (`NEC 404.8(A)`)** | Center of grip **`<= 6'-7"` (`79"`)** above working platform | Mount panel center at **`54"`–`60"`** eye level (handle at `~60"`–`64"`) |
| **Dedicated Space (`110.26(E)`)** | Footprint of panel extending floor-to-ceiling (or `6 ft` above) reserved strictly for electrical conduit | Zero water pipes, chemical hoses, or drain lines routed above the NEMA 4X panel! |

---

## 3. Recirculation Pump Station, Plumbing & Hosing

On the `24 cu.ft.` Deluxe build, a single closed-loop recirculation station circulates buffered `pH 8.0–8.5` alkaline liquor across all three wet stages (**`8.82 GPM` total**: Stage 1 Quench `1.16 GPM` + Stage 2 Venturi `2.30 GPM` + Stage 3 Packed Bed `5.36 GPM`), cutting water consumption by **`70%–85%`** vs. once-through operation while protecting Stages 1–2 from hot unbuffered HCl pitting.

### 3.1 Flooded-Suction Magnetic-Drive Sealless Pump Siting

Select a **seal-less magnetic-drive chemical pump** (e.g., **March TE-5C-MD**, **Iwaki MD-55R/MD-70R**, or **Little Giant TE-5.5-MD-HC**, polypropylene or PVDF wet-end, alumina ceramic shaft/thrust washers, EPDM/Viton O-rings, TEFC motor rated **`10–14 GPM @ 40–50 ft TDH`**, `1/3–1/2 HP` `120/240V`).

- **Mandatory Flooded Suction Siting:** Mount the pump on a corrosion-coated steel pedestal on the skid floor so the pump suction centerline sits **`>= 6"`–`12"` BELOW** the minimum working liquid level of Sump Zone 2 (`NPSHa >= 10 ft`, vs. typical `NPSHr <= 4–6 ft` at `184°F` — remember hot `184°F` water has **`8.2 psia` vapor pressure**, which cuts atmospheric NPSH margin by **`19 ft`**! Never mount a mag-drive pump *above* a hot `184°F` sump or it will cavitate and seize its ceramic thrust bearing in seconds!).
- **Suction Piping Geometry:**
  - Pull from a **`1-1/2"` Sch 80 CPVC true-union bulkhead** mounted **`4"` above the sloped sump floor** in clean Zone 2 (downstream of the over/under ash weir and `50-mesh` weir screen).
  - Keep suction run **`<= 36"` long**, oversized at **`1-1/2"` ID** (`v <= 2.5 ft/s` friction loss `< 0.2 ft`), with **`>= 5D` (`7.5"`) straight pipe** directly into the pump suction flange and a slight upward pitch (`1/8"`/ft) toward the pump so zero air pockets can trap.
- **Dry-Run Protection:** Install a polypropylene **low-level float switch** in Sump Zone 2 (wired to stop the pump motor relay before suction uncovers) or an inline motor current-sensing under-load relay.

### 3.2 Pump Station Valving, Strainer & 3-Branch Manifold Sequence

Assemble the pump station in this exact hydraulic order using **Schedule 80 CPVC (`ASTM F441`, `200°F`)** and **EPDM/FKM true-union fittings** so any component swaps out in 5 minutes without cutting pipe:

```text
[SUMP ZONE 2] (4" above floor, >=6"-12" flooded head)
     │
     ├─► [1.5" True-Union Isolation Ball Valve (V1)]  (Isolates sump for strainer/pump service)
     │
     ├─► [1.5" CPVC Y-Strainer (20-40 Mesh 316L/PP Screen) + 1/2" Flush Ball Valve]
     │       * Traps coarse wood char/ash chunks before impeller (check delta-P / clean pre-firing!)
     │
     ├─► [Flexible Braided EPDM Vibration Coupling (6" L)] -> [MAG-DRIVE PUMP (10-14 GPM @ 45 ft)]
     │
     ├─► [1" True-Union Spring/Ball Check Valve (CV1)] (Prevents reverse column drain & impeller back-spin)
     │
     ├─► [0-60 psig Glycerin Pressure Gauge + 316L/PTFE Diaphragm Seal] (Reads pump head & strainer fouling)
     │
     ├─► [1" Discharge Isolation Ball Valve (V2)] -> [1" Sch 80 CPVC 3-BRANCH HEADER]
             │
             ├── Branch 1 (Stage 1 Quench):   1/2" CPVC -> 50-Mesh Fine Strainer -> Throttle Valve + Rotameter (1.16 GPM)
             │                                -> Hardwired 0.25 GPM Mechanical Paddle Flow Switch (NFPA 86!)
             │                                -> Convoluted PTFE/SS Hose -> Stage 1 Spray Header (60 psig* / fine cone)
             │                                (*Note: Use dedicated high-pressure booster or twin-fluid air-assist if
             │                                 hydraulic quench nozzles require >=60 psig while recirc header runs 20 psig!)
             │
             ├── Branch 2 (Stage 2 Venturi):  3/4" CPVC -> Throttle Valve + Rotameter (2.30 GPM)
             │                                -> Braided EPDM Hose -> Opposed Venturi Throat Ports
             │
             └── Branch 3 (Stage 3 Absorber): 3/4" CPVC -> Throttle Valve + Rotameter (5.36 GPM)
                                              -> Top Column Liquid Distributor Spray Header
```

> **Strainer Mesh Coordination (`20–40 Mesh` Suction + `50 Mesh` Nozzle Guard):** Install a large-area **`20–40 mesh`** (`840–420 µm` opening) Schedule 80 CPVC Y-strainer on the pump suction to stop wood-char flakes and scale without starving NPSH on hot `184°F` liquor. Add a secondary **`50-mesh`** (`297 µm`) inline strainer strictly on the pressurized discharge branch feeding small-orifice (`<= 0.060"`) Stage 1 quench or twin-fluid nozzles.

### 3.3 Braided EPDM / PTFE Chemical Hose Bend-Radius & Support Rules

Rigid Schedule 80 CPVC becomes ductile and sags between wide supports at `184°F`, while rigid connections to vibrating pumps or removable Stage 1/2 spray lances transmit fatigue cracks. Use industrial chemical hoses correctly:

| Hose Specification | Approved Service Zone | Temp / Pressure Rating | Minimum Bend Radius | Support & Clamping Rules |
| :--- | :--- | :---: | :---: | :--- |
| **Peroxide-Cured Braided EPDM Chemical Hose** (`3/4"` / `1"` ID) | Recirculation pump decoupling, Stage 2 venturi feeds, Stage 3 top distributor jumper | `-40°F to +212°F`, `150–200 psig` | **`>= 6×` to `8×` Hose OD** (`>= 6.0"` for `3/4"` ID; `>= 8.0"` for `1"` ID) | Support horizontal spans every **`24"`–`36"`** with EPDM-cushioned 316 SS P-clamps; never pull taut (leave `2"` slack) |
| **Convoluted PTFE-Lined 316 SS Braided Hose** (`1/2"` / `3/4"` ID) | Stage 1 Quench nozzle lances (close to `2400°F` radiant heat zone!) & high-pressure (`60–100 psig`) quench header | `-65°F to +400°F`, `>= 1,000 psig` | **`>= 5×` to `6×` Hose OD** (never kink PTFE liner — kinking permanently fractures liner!) | Use `45°` or `90°` 316L swivel elbows at manifold/lance ports instead of forcing tight bends; clip every `24"` |
| **Rigid Schedule 80 CPVC Pipe** (`ASTM F441`) | Fixed header, sump manifold, vertical risers | `<= 200°F` continuous (derate pressure `50%` at `180°F`) | Rigid fittings only (no cold bending!) | Support horizontal runs continuously in fiberglass channel or every **`3.0 ft`** (`180°F` ASTM F441 spacing!) with loose sliding hangers for thermal expansion (`~0.5"` per `10 ft` at `ΔT = 115°F`) |

- **Clamps & Fittings:** Use **316L SS** or **glass-filled PP Camlock / barbed couplings** secured with **dual opposed 316 SS T-bolt or smooth-liner band clamps** (rotate screws `180°` apart). Never use galvanized/zinc worm-gear clamps — salt mist snaps them within a season.

### 3.4 IPC Section 608 Potable Water Isolation (Engineered Air Gap or ASSE 1013 RPZ)

Connecting a studio potable water line directly to a pressurized chemical quench manifold or submerged sump float valve creates a **High-Hazard chemical cross-connection** (`HCl`, `NaCl`, `NaHCO₃`, wood ash, ceramic glaze metals) prohibited by **IPC Section 608 / Table 608.1**. The Deluxe plumbing station enforces one of two code-compliant isolation methods:

1. **Method A — Engineered Overhead Air Gap (Preferred Deluxe Standard, IPC 608.15.1 / ASME A112.1.2):**
   - Terminate the `3/4"` potable makeup supply pipe vertically over the atmospheric Sump Zone 2 (or a dedicated `15-gal` clean-water break tank feeding the high-pressure quench booster pump) with a rigid bolted pipe bracket so the spout cannot be deflected downward.
   - Maintain a permanent unobstructed vertical air gap of **`>= 2×` effective supply pipe diameter AND minimum `1.0 inch`** (`>= 1.5"` recommended for `3/4"` pipe) **above the vessel's physical flood-level overflow rim** (install a `2"` screened overflow standpipe `2"` below the tank top rim to establish a guaranteed flood plane).
   - **Advantage:** 100% mechanical fail-safe against both backsiphonage and backpressure, zero moving parts, and zero annual certification fees.
2. **Method B — Certified ASSE 1013 Reduced Pressure Principle (RPZ) Assembly:**
   - Required whenever potable water feeds a **pressurized** emergency quench line or tempering tee directly without an atmospheric break tank.
   - Install a lead-free bronze or 316 SS **ASSE 1013 RPZ** (e.g., Watts `009QT` / Zurn Wilkins `975XL2`) horizontally **`12"`–`60"` above finished grade** in a freeze-protected, accessible location with an upstream strainer, isolation ball valves, and an unobstructed air-gapped floor drain beneath the RPZ relief vent (never in a vault/pit!).
   - **Mandatory Maintenance:** Must be tested upon commissioning and **annually thereafter** by a state-certified backflow tester (log test certificate in the Commissioning Binder).

### 3.5 IPC Section 702.5 `<= 140°F` Cool-Down Sump Drain Manifold

Scrubber sump liquor runs at **`183.9°F`** during firing — **`44°F` above the `140°F` (`60°C`) maximum temperature limit** for residential/commercial PVC and ABS sanitary DWV piping (**IPC Section 702.5**). Never pipe raw sump blowdown directly into a PVC floor drain. Fabricate the Deluxe Cool-Down Drain Manifold:

```text
[SLOPED SUMP LOW-POINT 2" CPVC DRAIN]
       │
       ├─► [2" Lockable Full-Port CPVC Drain Ball Valve (Padlock Closed During Firing!)]
       │
       ├─► [Cool-Down Tempering Manifold]  <-- [3/4" Cold Tap Tempering Line via IPC 608 RPZ/Air-Gap]
       │       (OR route 0.08 GPM live blowdown into a 20-gal CPVC/316L atmospheric cooling catch tank!)
       │
       ├─► [316L Thermowell + Instant-Read Dial Thermometer]  --> VERIFY <= 140°F (60°C) (IPC 702.5)!
       │
       ├─► [5-Gal Removable Ash Filter Basket / 100-Micron Felt Sock]  --> Traps wood ash/silica (40 CFR 403.5(b)(3))
       │
       ├─► [1/2" pH & SG Sampling Valve]  --> VERIFY pH 6.5-8.5 (IPC 803.1; 40 CFR 403.5(b)(2) strictly forbids pH < 5.0!)
       │
       └─► [Sanitary Floor Drain / Approved POTW Connection]
```

- **Dump Cadence & Salt Conservation:** Using closed-loop alkaline recirculation over a `3.0-hr` salting window (`+21.6 g/L` TDS per firing in a `40-gal` working sump), reuse the neutralized sump for **`5–6` firings** and dump strictly when cooled liquor reaches **`SG >= 1.105`** (`150 g/L` TDS on a `$12` brine hydrometer) or Zone 1 ash sludge exceeds `4"`.
- **Glaze Heavy-Metal Caution:** Standard wood/salt/soda ash sludge with iron/alumina/silica clay bodies is typically non-hazardous once dewatered and neutralized, **but** if firing ceramic glazes/slips containing regulated heavy metals (`Pb`, `Ba`, `Cd`, `Cr`, `Co`, `Se`), test dried ash cake per local hazardous-waste rules rather than disposing in municipal trash or sanitary drains.

---

## 4. Structural Skid Frame, Column Bracing & Vibration Mounting

A Deluxe scrubber carries **nearly half a ton of wet operating mass** concentrated in a `10–15 ft` tall cantilevered column and stack, paired with a `3,450 RPM` radial blower and a `2400°F` thermally expanding flue. Rigidity, overturning resistance, and vibration isolation must be engineered — not improvised.

### 4.1 Welded `2" × 2" × 3/16"` Structural Steel Forklift Skid

- **Frame Members:** Fabricate the perimeter base frame and cross-braces from **ASTM A500 Grade B `2" × 2" × 3/16"` square HSS steel tubing** (`4.32 lbs/ft`, `I = 0.722 in⁴`). Place full-width cross-members on **`<= 24"` centers** directly beneath Stage 1 quench legs, Stage 3 column/sump footprint, and Stage 4 blower plinth.
- **Footprint Options:**
  - **Linear Skid:** `~12.9 ft L × 5.4 ft W` (`155" × 65"`), best for long exterior kiln walls.
  - **Compact U-Turn Folded Skid (`-46%` length):** `~7.0 ft L × 6.2 ft W` (`84" × 74"`), placing Stage 1 + Stage 2 on Leg 1, turning `180°` via two long-radius sweep elbows into Stage 3 on Leg 2, and tucking the Stage 4 acoustic doghouse beside Stage 1 over the shared sump.
- **Integrated Forklift Pockets:** Weld two **`3" × 6" × 3/16"` rectangular HSS forklift tunnels** full-width beneath the main longitudinal runners on `36"`–`42"` center-to-center spacing (open on both long sides for pallet-jack or forklift rigging).
- **Weld & Coating Spec:** Seal-weld all tube-to-tube miter and T-joints completely (`AWS D1.1`, `3/16"` fillet welds all around) with `1/4"` seal-welded end caps so acid condensate cannot enter hollow tubes. Blast to **SSPC-SP6 (Commercial Blast)** and coat with **`3 mil` zinc-rich epoxy primer + two `5 mil` coats of chemical-resistant polyamide epoxy / aliphatic polyurethane** (or hot-dip galvanize per ASTM A123 + epoxy topcoat).

### 4.2 Concrete Pad & `1/2" × 4"` Concrete Wedge Anchor Sizing

- **Concrete Foundation:** Pour a level `>= 4"`–`6"` thick reinforced concrete slab (`>= 3,500–4,000 psi`, `#4` rebar on `12"` centers or `6×6 W2.9×W2.9` WWF over `4"` compacted crushed gravel) extending `>= 12"` beyond the skid perimeter and covering the full `36" × 30"` NEC 110.26 electrical working zone. Slope pad `1/8"`/ft away from the electrical panel toward an acid-resistant floor drain/curb.
- **Anchor Bolts (`1/2" × 4"` 316 SS Wedge Anchors):**
  - Weld `3/8"` thick × `4" × 4"` steel anchor tabs with `9/16"` slotted holes at **8 perimeter points** around the skid (`4` anchors directly flanking the Stage 3 tall-column baseplate to resist overturning tension!).
  - Install **`1/2" × 4"` Type 316 Stainless Steel concrete wedge anchors** (e.g., Simpson Strong-Tie Strong-Bolt 2 `STB2-504006SS` or Hilti Kwik Bolt TZ2 316 SS):
    - Drill `1/2"` carbide ANSI bit perpendicular to slab to `>= 3-1/2"` depth; blow & brush hole clean.
    - Minimum effective embedment: **`2-1/4"`–`2-3/4"`** (nominal `3"` embedment leaves `1"` above slab for `3/8"` tab + washer + nut).
    - Minimum edge distance: **`>= 4"`** from slab edge; minimum anchor-to-anchor spacing **`>= 6"`**.
    - Installation torque: **`40–50 ft-lbs`** (verify against manufacturer ICC-ES ESR table; never impact-wrench 316 SS wedge anchors without nickel anti-seize or threads will gall!).

### 4.3 Stage 3 Tall-Column Wet Mass Budget & Diagonal K-Bracing

Size the column frame and bracing for the **heaviest possible wet operating condition** (including optional kiln-fired stoneware ceramic packing + flooded sump + wind/seismic lateral forces):

| Stage 3 Load Component (`14"`–`18"` Deluxe Column) | PP Pall Rings Config (`14"` FRP/CPVC) | Heavy Stoneware Ceramic Rings Config (`14"`–`18"` Column) | Load Engineering Note |
| :--- | :---: | :---: | :--- |
| **Column Shell, Flanges, Manway & Cassette** | `115–190 lbs` | `140–225 lbs` | `14"` Sch 80 CPVC (`~23 lb/ft`) or Derakane FRP (`~12–16 lb/ft`) |
| **Dry Packing Media (`3–4.2 ft³`)** | `18–25 lbs` | **`150–195 lbs`** | Unglazed Cone 6–10 stoneware rings (`~46 lb/ft³`) |
| **Irrigated Dynamic Liquid Holdup (`~10%` Bed)** | `30–45 lbs` | `45–65 lbs` | Continuous `5.36 GPM` irrigation + top distributor pool |
| **Chevron Demister + Top Reducer + `12 ft` Stack** | `45–65 lbs` | `55–75 lbs` | Includes wet condensate film & inline silencer weight |
| **Integral / Base Flooded Sump (`40–55 gal` Brine, `SG 1.10`)** | `370–505 lbs` | `370–505 lbs` | Full working level (`9.17 lb/gal` at `SG 1.10`) |
| **TOTAL WET OPERATING MASS (Stage 3 Assembly)** | **`580–830 lbs`** | **`760–1,065 lbs`** | **Design structural frame & grate for `>= 1,250 lbs` vertical dead/live load!** |

- **Tall-Column Structural Cage & Diagonal K-Bracing:**
  - Erect a **4-post `2" × 2" × 3/16"` HSS vertical tower frame** welded directly to the skid baseplate rising to **`72"`–`96"` elevation** (above the upper column flange).
  - Weld **`1-1/2" × 1-1/2" × 1/8"` steel angle diagonal K-bracing** (`45°` chevron bays) across the **back, left, and right faces** of the cage. Leave the **front face open** (with a removable bolted horizontal cross-tie above the packing manway) so the Stage 3 manway and side-slide demister cassette remain 100% accessible.
  - **Wind & Seismic Lateral Load Check:** At `115 mph` ASCE 7 3-sec gust (`q_z ≈ 25–30 psf`, cylinder shape factor `C_f = 0.7`), a `14"`–`18"` column + `12–15 ft` outdoor stack experiences `~120–185 lbs` lateral wind shear (`~800–1,400 ft-lbs` overturning moment at the base); in Seismic Design Category D (`0.4g` horizontal), a `1,000 lb` wet column generates `~400 lbs` lateral seismic shear (`~1,800 ft-lbs` overturning moment). Four `1/2" × 4"` 316 SS wedge anchors on a `30" × 30"` baseplate provide `> 4×` safety factor against overturning.
  - **Cushioned Column Restraint Collars:** Clamp the FRP/CPVC column to the steel cage at two elevations (mid-height and top collar) using **split `2" × 1/4"` rolled steel bands lined with `1/4"` EPDM neoprene rubber strips** (tightened snug — never crushing the plastic shell — allowing vertical thermal expansion of the heated `184°F` column!).
  - **Tall Outdoor Stack Bracing (`> 6 ft` Above Column Collar):** Brace the `12–15 ft` vertical outdoor exhaust stack with either a rigid `2"` galvanized/316L mast strapped to the K-braced frame OR **3-point `1/8"` 7×19 Type 316 SS aircraft-cable guy wires** anchored at `120°` intervals (`45°` elevation angle) with turnbuckles.

### 4.4 Overhead Duct Pipe-Saddle Supports, Neoprene Vibration Mounts & Flexible Boots

1. **Overhead Hot-Flue Sliding Roller Saddles (`2400°F` Side):**
   - A `10 ft` run of 316L stainless flue heated from `70°F` to `1400°F` expands axially by **`~1.5 inches`** (`α ≈ 9.6 × 10⁻⁶ in/in·°F`). Never rigidly clamp both ends of the hot flue!
   - Support horizontal overhead flue runs every `5–6 ft` on **316L SS pipe-roller hangers or PTFE-pad sliding saddles** (insulated beneath the saddle with high-density calcium-silicate pipe-support blocks so the steel hanger never creates a thermal bridge!) and include a **high-temperature braided ceramic/Inconel bellows expansion joint** between the kiln masonry damper and the Stage 1 quench inlet flange.
2. **Overhead Wet CPVC/FRP Duct Saddles (`<= 200°F` Side):**
   - Support horizontal `6"`–`8"` wet transition ducts every **`3–4 ft`** using **`180°` wide-arc steel pipe saddles lined with `1/4"` EPDM rubber** (`>= 4"` axial saddle width) hung from threaded 316 SS rods (`3/8"`–`1/2"`). Never hang warm `184°F` CPVC from bare narrow U-bolts or wire straps. Slope horizontal wet ducts **`>= 1/4"` per foot** back toward Stage 3 or toward an NFPA 91 low-point condensate drain leg.
3. **Neoprene Blower Vibration Isolators:**
   - Bolt the Stage 4 `2.0–3.0 HP` radial blower and TEFC motor onto a rigid **`1/4"` welded steel inertia sub-base** (adds mass and maintains shaft/housing alignment).
   - Mount the sub-base onto **`4 ×` molded neoprene shear/compression vibration mounts** (e.g., Mason `ND-B` or McMaster-Carr oil/acid-resistant `50–60` Shore A durometer neoprene isolators, rated `100–150 lbs` per mount, yielding `0.15"–0.25"` static deflection).
   - At `3,450 RPM` (`57.5 Hz` forcing frequency), a `0.20"` static deflection gives natural frequency `f_n = 3.13 / √0.20 = 7.0 Hz` (`f / f_n = 8.2`), isolating **`> 98%` of structure-borne motor/impeller vibration** from the steel skid!
4. **Flexible PTFE / EPDM Expansion Boots (Mandatory on BOTH Blower Ports):**
   - Never hard-bolt rigid CPVC/FRP ductwork directly to the blower scroll — it transmits structural drumming (`+3–6 dBA`) and cracks plastic flanges!
   - Install **`6"`–`8"` long reinforced PTFE-lined EPDM flexible expansion boots** (e.g., Garlock / Flexaust chemical-duty bellows, `>= 220°F` wet HCl rated) on **both** the blower suction inlet and discharge outlet, fastened with **dual 316 SS T-bolt clamps** on each cuff.
   - **Suction Vacuum Reinforcement:** Because the blower inlet sits at **`-21 inWC`** vacuum, specify a **wire-helix reinforced or internal 316L/PP support-ring suction boot** so negative suction never collapses the sleeve inward! Support all duct weight independently on adjacent pipe saddles so the flexible boot carries zero dead load.

---

## 5. Sound Baffling & Urban Acoustic Engineering

An unsilenced `2.0–3.0 HP` (`3,450 RPM`) industrial radial pressure blower pushing `256 ACFM @ 21.2 inWC` generates strong tonal blade-pass energy (`f_BPF = 3,450 × 8 / 60 = 460 Hz` fundamental + `920 Hz` harmonic) plus high-velocity discharge turbulence. Near reflective masonry kiln walls (or on `3 HP` co-op / high-RPM bypass blowers), bare sound pressure reaches **`86 dBA @ 3 ft`** — exceeding the **OSHA 29 CFR 1910.95 `85 dBA`** 8-hr TWA Action Level for potters tending the kiln and radiating **`~63.5 dBA @ 40 ft`**, violating `55–60 dBA` daytime and `45–55 dBA` nighttime residential ordinances.

The Deluxe Acoustic System reduces noise across **all three transmission paths** (structure-borne vibration, stack-mouth aerodynamic radiation, and blower-casing/motor breakout) by **`~22 dBA`**, hitting **`~64 dBA @ 3 ft`** and **`~38–42 dBA @ 40 ft` property line** — quieter than a suburban nighttime breeze.

### 5.1 Acoustic Baffle Enclosure Construction ("Deluxe Acoustic Doghouse")

Construct a modular, weatherproof, lift-off or hinged-door enclosure surrounding the Stage 4 blower + motor assembly on the skid:

- **Outer Mass-Barrier Shell (`>= 4–5 lb/ft²` surface density):**
  - **Metal Shop Deluxe:** `16 ga` (`0.060"`, `2.5 lb/ft²`) powder-coated/316L steel shell lined internally with **`1 lb/ft²` mass-loaded vinyl (MLV)** bonded to the outer skin (`3.5 lb/ft²` total).
  - **Studio Composite Deluxe:** `3/4"` exterior marine plywood (`2.2 lb/ft²`) laminated to `1/2"` HardieBacker / cement board (`2.5 lb/ft²`, `4.7 lb/ft²` total — noncombustible interior face, weather-sealed exterior paint + sloped metal drip roof).
- **Inner Acoustic Absorption Layer:**
  - Line **all six interior surfaces** (4 walls, roof, and sub-base baffle) with **`2"` thick Rockwool Safe'n'Sound** (or industrial Roxul `RHT80`, `3.0–4.0 lb/ft³` mineral wool, **NRC `1.05`**, noncombustible ASTM E136, hydrophobic).
  - Retain the mineral wool behind a **`22–24 ga` perforated galvanized or 316L SS inner facing** (`>= 30%–40%` open area, `3/16"` staggered holes on `5/16"` centers) so high-velocity motor cooling air never erodes wool fibers and operators cannot snag the insulation during belt/bearing inspection.
- **Airtight Acoustic Door & Penetration Seals:**
  - Seal the hinged full-front service door with continuous **`1/2" × 3/4"` closed-cell EPDM D-profile compression gasketing** compressed by stainless over-center draw latches.
  - Seal inlet/outlet duct penetrations around the flexible boots with **`1/2"` thick split neoprene/EPDM collar rings** (never let rigid duct touch the enclosure shell!). *Acoustic rule: an unsealed `1%` air gap destroys up to `10 dB` of enclosure transmission loss!*

### 5.2 CRITICAL: Baffled Labyrinth Motor-Cooling Air Intake & Exhaust

> **THERMAL WARNING — NEVER SEAL A TEFC MOTOR IN AN UNVENTED ACOUSTIC BOX!**
> A `2.0 HP` TEFC motor at `85%` efficiency rejects **`~260 W` (`890 BTU/hr`)** of electrical heat, plus `~200–350 W` of radiant/convective heat from the `184°F` wet blower scroll (**`~1,500–2,000 BTU/hr` total**). Running inside an unvented Rockwool box cooks motor winding insulation (`Class F` limit) and trips thermal overloads within 30–60 minutes!

To provide generous TEFC motor cooling while preserving `12–15 dBA` of acoustic attenuation, build **two Rockwool-lined Z-bend acoustic labyrinth ducts** integrated into opposite sides of the enclosure:

```text
                CROSS-SECTION: DELUXE ACOUSTIC BLOWER DOGHOUSE (SIDE VIEW)
========================================================================================
  [SLOPED WEATHERPROOF ROOF — 16 ga Steel + 2" Rockwool Safe'n'Sound + Perf. Metal]
  +----------------------------------------------------------------------+
  |  //////////////////////////// 2" ROCKWOOL /////////////////////////  |
  |  ------------------------------------------------------------------  |  WARM COOLING
  |  |  +--------------------------+                                  |  |  AIR EXHAUST
  |  |  |  2" Rockwool Lined       |    [STAGE 4 RADIAL BLOWER]       |  |  LABYRINTH
  |  |  |  Z-BEND INTAKE LABYRINTH |    [& 2.0 HP TEFC MOTOR]         |  |  (>=72 in²)
  |  |  |  (>=72 in² Free Airway)  |       _______________            +--+------------+
  |  |  |  ZERO Line-of-Sight!     |      |  BLOWER       |           | 2" Rockwool   |
  |  |  |                          |      |  SCROLL  [M]  |=====> warm|  Baffle Turn  |
  |  |  |  cool air   -->   -->    |=====>|  (184°F) TEFC |  air      |   |           |
  |  |  +--------+   +-------------+      |_______________|           |   v   Exhaust |
  |  |           |   |                      |           |             |   +---------->
  |  +-----------+   |                 [1/4" Steel Inertia Base]      |   | (Weather
  |  <--- Cool       |                 [Neoprene Mounts (x4)   ]      +---+  Louver)
  |  (Intake Hood)   |                                                |  |
  +------------------+------------------------------------------------+--+
  ===================== 2"x2"x3/16" STRUCTURAL STEEL SKID ================
  * Ducts enter/exit through EPDM collar seals on flexible PTFE/EPDM boots
  * Add 120V quiet IP55 100 CFM exhaust booster fan in exit labyrinth if ambient >95°F!
```

- **Sizing the Cooling Labyrinths:**
  - Target **`>= 150–250 CFM`** of cooling airflow (`ΔT <= 12°F–18°F` inside enclosure at `2,000 BTU/hr` heat rejection).
  - Size both the low-elevation intake labyrinth (aligned near the TEFC rear fan cowl) and high-elevation exhaust labyrinth (above the motor barrel) with **`>= 72 in²` (`6" × 12"`) clear internal airway** (`v <= 400–500 FPM`, negligible `< 0.05 inWC` static resistance).
  - Include **at least two lined `90°` turns** (`>= 24"`–`36"` total Rockwool-lined path length) with **`2"` Rockwool + perforated metal** on all baffle faces so zero direct sound ray escapes.
  - Fit exterior openings with sloped rain hoods and `1/4"` 316 SS rodent/insect screens. In hot climates (`> 95°F` summer ambient), add a quiet `120V` `100–120 CFM` IP55 axial exhaust booster fan (`< 35 dBA`) inside the exhaust labyrinth wired to the VFD Run relay (`DO1`).

### 5.3 Inline Absorptive Stack Discharge Silencer Placement

Enclosing the blower stops casing breakout noise, **but** half the acoustic energy fires straight up the exhaust duct and out the top of the vertical stack like a megaphone. Install an **inline absorptive tubular stack discharge silencer**:

- **Placement:** Mount vertically immediately downstream of the Stage 4 blower discharge flexible boot and **upstream** of the vertical stack riser (and **upstream** of any optional squirrel-cage plume diluter tee), supported on the K-braced steel mast.
- **Moisture- & Acid-Proof Construction:** Because exhaust gas is `184°F` saturated water vapor (with `~10 ppm` HCl in salt mode), **standard bare fiberglass HVAC silencers will waterlog and rot within 3 firings!** Specify:
  - **Outer Shell:** `8"` (or stack-matched `6"`–`10"`) Schedule 80 CPVC, Derakane FRP, or 316L SS outer jacket (`36"`–`48"` long) with ANSI Class 150 flanged ends.
  - **Inner Perforated Tube:** `316L SS` or `PP` perforated core (`>= 35%` open area) wrapped in a **`1.0 mil` Tedlar (PVF) or expanded PTFE acoustic vapor-barrier film** (acoustically transparent below `4 kHz`, 100% waterproof and HCl-proof!).
  - **Acoustic Fill:** `2"` thick annular bed of hydrophobic **Rockwool mineral wool (`4.0 lb/ft³` / `64 kg/m³`)** packed between the Tedlar-wrapped core and outer shell.
  - **Condensate Drain:** Include a **`1/2"` bottom condensate drain loop** (`>= 6"` water-trap seal) returning stack condensation to the shared sump.

### 5.4 Stage-by-Stage Acoustic Attenuation Budget (`86 dBA` Bare → `38–42 dBA @ 40 ft`)

Acoustic reductions below reflect field-verified industrial blower acoustics (A-weighted `250 Hz – 4 kHz` radial-blade spectrum) and hemispherical inverse-square outdoor spreading (`ΔL = 20 log₁₀(40 / 3) = -22.5 dB`):

| Acoustic Control Stage | Target Noise Path | Incremental Reduction | Sound Level at `3 ft` (Operator Near-Field) | Sound Level at `40 ft` (Property / Fence Line) | Code / Neighbor Status |
| :--- | :--- | :---: | :---: | :---: | :--- |
| **0. Bare Unsilenced Blower** *(2–3 HP Radial / Reflective Wall)* | Casing + Stack Mouth + Skid Drumming | Baseline | **`86 dBA`** | **`~63.5 dBA`** | **FAIL** — Exceeds OSHA `85 dBA` Action Level & `60 dBA` Daytime / `50 dBA` Nighttime |
| **1. Vibration Decoupling** *(Neoprene Mounts + PTFE/EPDM Boots + LFNC Whip)* | Structure-borne skid & duct resonance | `-2 to -3 dBA` | `~83–84 dBA` | `~60.5–61.5 dBA` | Eliminates low-frequency booming/drumming |
| **2. + Inline Absorptive Stack Discharge Silencer** *(Tedlar-Lined Wool, 36"–48" L)* | Stack-mouth blade-pass (`460 Hz`) & jet noise | `-8 to -10 dBA` | `~74–75 dBA` | `~51.5–52.5 dBA` | Passes daytime `60 dBA`; borderline overnight `45–50 dBA` |
| **3. + `2"` Rockwool Safe'n'Sound Baffled Doghouse** *(Perf. Metal + Z-Cooling Labyrinths)* | Blower scroll & TEFC motor fan breakout | `-10 to -11 dBA` | **`~64 dBA`** | **`~41.5 dBA`** | **PASS** — `21 dBA` below OSHA `85 dBA`! |
| **4. + Hemispherical Spreading (`40 ft`) + Yard Ground/Fence Softening** | Distance + soft ground/fence diffraction (`0 to -3.5 dB`) | `-22.5 to -26 dB` *(from 3 ft)* | **`~64 dBA @ 3 ft`** | **`~38–42 dBA @ 40 ft`** | **WHISPER-QUIET PASS** — Passes strictest `45 dBA` overnight residential ordinances! |
| **5. Bonus: VFD Preheat / Reduction Turndown (`45%–65%` Speed)** | Overnight non-salting kiln hours (`50 log₁₀(N₂/N₁)`) | `-6 to -15 dBA` *(non-salt hrs)* | `~49–58 dBA` | **`< 30–36 dBA`** | Virtually inaudible below ambient night background (`~35 dBA`) |

---

## 6. Complete Deluxe Inspection & Commissioning Checklist

Complete **every test in sequence with the kiln COLD** prior to initial lighting, and archive the signed sheet in your studio compliance binder for AHJ/POTW inspection.

**Project / Studio:** `___________________________` **Kiln Size:** `_____ cu.ft.` **Date:** `__________` **Commissioning Lead:** `_____________________`

### 6.1 Mechanical, Metallurgy & Structural Verification

- [ ] **Weld & Passivation Sign-Off (Stages 1–2):** Confirm 100% argon back-purged root welds inside 316L SS / Alloy C-276 quench & venturi spools (zero sugaring/crevices) and completed ASTM A380/A967 pickling & passivation.
- [ ] **Hot-Flue Ceramic Gaskets & NFPA 211 Clearances:** Verify **`>= 2300°F`** alumina-silica ceramic fiber rope gaskets on all hot dry flue flanges upstream of Stage 1 (**zero** silicone or PTFE on hot flue!), `2"`–`3"` ceramic-fiber blanket + SS weather jacketing installed, **`>= 36"`** clearance to all unprotected combustibles, zero combustible wall penetrations, and NFPA 211 auxiliary bypass chimney unobstructed.
- [ ] **Downstream Flange Torque & Sight Glasses:** Verify `1/8"` PTFE envelope / ePTFE gaskets on all Stage 1-exit / Stage 2 / Stage 3 ANSI Class 150 flanges torqued in 3-pass star sequence (`24-hr` re-torque completed); verify borosilicate sight glasses cushioned with PTFE rings and clear view of Stage 1 spray cone.
- [ ] **NFPA 91 Cleanouts & Stage 3 Access:** Confirm bolted gasketed cleanout doors at flue elbows, quench base, and venturi transition; verify Stage 3 packing manway and removable chevron demister cassette latch tightly and slide freely.
- [ ] **Structural Skid, Anchors & Bracing:** Verify all **8× `1/2" × 4"` 316 SS concrete wedge anchors** torqued to **`40–50 ft-lbs`** (`>= 4"` edge distance), diagonal K-bracing bolts tight on Stage 3 cage, EPDM-lined restraint collars snug without crushing FRP/CPVC, overhead flue sliding saddles free to expand, and blower sitting on **4× neoprene vibration mounts** with **flexible PTFE/EPDM boots** (suction boot vacuum-ring reinforced) carrying zero duct weight.

### 6.2 Hydrostatic & Vacuum / Pressure Smoke Leak Tests

- [ ] **Hydrostatic Sump & Piping Leak Test (`30 Minutes`):** Fill sloped baffled sump with clean tap water to the overflow weir mark; confirm `>= 6"` flooded suction head above mag-drive pump; run pump against throttled discharge at **`1.25×` normal operating pressure (`~35–45 psig`) for `30 min`**.
  - *Acceptance Criterion:* **Zero weeping or dripping** at sump bulkheads, `20–40 mesh` Y-strainer, check valve, pressure-gauge diaphragm seal, CPVC solvent joints (cured `>= 24 hrs`), braided EPDM/PTFE hoses (`>= 6×–8× OD` bend radius), and Stage 1–3 flange joints.
- [ ] **Nozzle Spray & Flow-Balance Verification:** Inspect Stage 1 (`1.16 GPM` @ design pressure), Stage 2 (`2.30 GPM`), and Stage 3 (`5.36 GPM`) sprays through borosilicate sight glasses & top port — confirm uniform full-cone coverage with zero plugged orifices and clean `20–40 mesh` / `50-mesh` strainers.
- [ ] **Cold Vacuum Smoke Test (`-21 inWC` Suction):** Run Stage 4 blower cold at `60 Hz` (`100%` speed) with wet liquid seals flooded (`>= 8"` submerged seal on Stage 1 quench at `-1 to -2 inWC`; **`>= 26"`–`28"` vertical barometric seal leg / U-trap** on Stage 2/3 drains into an atmospheric vented sump, **or** bolted vacuum-rated hatch sealed tight on an integral vacuum-sealed Stage 3 sump). Slowly trace a theatrical/HVAC smoke pencil around **every** hot-flue gasket, NFPA 91 cleanout door, borosilicate sight glass, venturi flange, Stage 3 manway, demister cassette seal, sump hatch (if vacuum-sealed), and suction flexible boot.
  - *Acceptance Criterion:* **Zero inward smoke draw** anywhere along the negative-pressure train (confirms airtight seal against `1,000+ ppm` HCl ingress/leakage).
- [ ] **Downstream Positive-Pressure Soap-Bubble Test:** With blower at `60 Hz`, brush soapy water onto the blower discharge boot, inline stack silencer flanges, and condensate drain trap (and all Stage 3 flanges at `+6 inWC` if operating in dry-fan eductor mode).
  - *Acceptance Criterion:* **Zero expanding bubbles** at any positive-pressure seam.

### 6.3 Electrical, NEC 110.26 / 430.102 / 430.122 & Thermal Sign-Off

- [ ] **NEC 110.26 Electrical Working Clearance:** Tape-measure and verify permanently clear **`>= 36"` depth** (`42"` preferred) × **`>= 30"` width** × **`>= 78"` headroom** directly in front of the NEMA 4X panel; yellow/black floor zone striped; door swings `>= 90°`; overhead work light active; zero water piping above panel.
- [ ] **NEC 430.102 Lockable Through-Door Disconnect (`<= 50 ft`):** Confirm disconnect is in unobstructed line-of-sight (`<= 50 ft`) of both VFD and Stage 4 blower, grip center `<= 79"` above grade, mechanical door interlock functional, and accepts 3 LOTO padlocks cleanly in OFF.
- [ ] **NEC 430.122(A) & 210.8(F) Branch Circuit Verification:** Verify **`30A` 2-pole HF-rated GFCI breaker** + **`#10 AWG` THWN-2 copper** conductors (`40A` / `#8 AWG` for `3.0 HP`), **`3%` input line reactor** wired ahead of VFD, fixed Sch 80 PVC / LFNC-B conduit (NEC 400.12), and `>= 6"` parallel spacing (`90°` crossings) separating Conduit A (VFD power) from Conduits B/C (24VDC safety + RS-485/analog).
- [ ] **Ground Continuity & EMI Shielding Check:** Verify `< 0.5 Ω` PE ground continuity from main panel to skid steel frame, VFD ground plate, and blower motor frame; verify shielded VFD motor cable grounded at **both** ends and RS-485 / `4–20 mA` shields grounded at **control panel PE bus ONLY**.
- [ ] **VFD Bump Test & HF-GFCI Stability:** Close disconnect, jog VFD at **`15 Hz`** to verify correct blower wheel rotation direction (swap any two motor phases `U/V/W` if reversed!), then ramp smoothly `15 → 60 Hz` across 5 cycles confirming **zero HF-GFCI nuisance trips** and smooth motor amps `<= FLA`.

### 6.4 Plumbing: IPC 608 Backflow & IPC 702.5 Drain Compliance Sign-Off

- [ ] **IPC Section 608 Potable Water Isolation:**
  - *If Air Gap (IPC 608.15.1):* Measure vertical separation with ruler — confirm **`>= 1.0"` and `>= 2×` supply ID** above tank overflow rim, rigid bracket locked in place, zero submerged hose or float valve.
  - *If ASSE 1013 RPZ:* Verify horizontal installation `12"`–`60"` above grade, unobstructed relief drain, and attach dated **certified backflow test report**.
- [ ] **IPC 702.5 Drain Manifold & Field Kit Check:** Verify sump drain valve padlockable in CLOSED position, cold-water tempering valve and inline thermometer (`<= 140°F` verify mark) functional, ash filter sock staged, and verify calibrated field test kit on hand (`$12` brine hydrometer `SG 1.000–1.220`, wide-range `pH 1–14` strips, instant-read thermometer, and bulk `15 lb` bags / `50 lb` sacks of `NaHCO₃`).

### 6.5 NFPA 86 Hardwired Safety Loop & Fail-Safe Bypass Drop Tests

> **Mandatory Test Condition:** During Tests 1–4 below, intentionally command the supervisory controller (Raspberry Pi / PLC / PID) to hold `VFD = 60 Hz` and `Damper = CLOSED` so you **prove** the independent hardwired `24 VDC` `K_SAFE` relay overrides all software!

- [ ] **Test 1 — `200°F` Manual-Reset Snap-Disc Thermal Trip:** Apply a calibrated hot-air gun or boiling-water bath to the independent quench-exit snap-disc sensor (`200°F ± 5°F`).
  - *Pass Criterion:* `K_SAFE` drops immediately, **Belimo spring-return bypass damper drives 100% OPEN in `< 20 s`**, VFD halts via hardwired `STO/DI1`, 24VDC strobe/horn activates, and system **stays latched OFF** after cooling until deliberate human press of **Manual Reset PB**!
- [ ] **Test 2 — `0.25 GPM` Quench Low-Flow Trip:** Gradually throttle the Stage 1 quench ball valve until flow drops below **`0.25 GPM`**.
  - *Pass Criterion:* Mechanical paddle flow switch opens `K_SAFE` within `<= 2 s`, bypass damper springs **100% OPEN**, VFD stops immediately, alarm latches until flow restored + Manual Reset pressed.
- [ ] **Test 3 — Emergency Stop (E-Stop) Trip:** Depress the NEMA 4X red E-Stop mushroom button while blower and pump are running.
  - *Pass Criterion:* Instantaneous drop of `K_SAFE`, bypass damper springs **100% OPEN**, VFD stops; twisting E-Stop out does **not** restart equipment until Manual Reset PB is pressed.
- [ ] **Test 4 — Total Utility Power-Loss Fail-Safe Drop Test:** While running at `60 Hz`, throw the **NEC 430.102 main disconnect switch to OFF** (simulating blackout).
  - *Pass Criterion:* Under **zero electrical power**, internal mechanical spring drives the bypass damper **100% OPEN in `< 20 s`** (measure & record time: `_____ sec`), and optional gravity emergency quench NO solenoid opens cleanly.
- [ ] **Test 5 — Digital Supervisory Backup & Watchdog Fault (if Pi/PLC installed):** Disconnect the RS-485 cable (or stop `kilnscrub daemon`) while running.
  - *Pass Criterion:* Software/hardware Modbus watchdog drops permissive coil (`bypass_coil_deenergize_to_trip: true`) within **`<= 5 s`**, confirming defense-in-depth backup.

### 6.6 Acoustic dBA Meter Check, TEFC Thermal Run & Final Sign-Off

- [ ] **Ambient Baseline Noise Measurement:** Using a calibrated Type 2 Sound Level Meter (`dBA`, Slow response, windscreen fitted, `4–5 ft` tripod height), record background outdoor noise with scrubber OFF:
  - Baseline at `3 ft`: `_______ dBA` | Baseline at `40 ft` property/fence line: `_______ dBA`
- [ ] **Full-Load `60 Hz` Acoustic Verification:** Run Stage 4 blower at `100%` (`60 Hz`, full `21.2 inWC` static load) with acoustic doghouse door latched and inline stack silencer attached:
  - Measured at `3 ft` from doghouse: `_______ dBA` *(Target: **`<= 65–68 dBA`**, strictly `< 85 dBA` OSHA Action Level)*
  - Measured at `40 ft` fence / property line: `_______ dBA` *(Target: **`38–42 dBA`**, strictly below local daytime & overnight ordinances)*
- [ ] **60-Minute TEFC Motor Thermal Verification:** Keep blower running at `60 Hz` inside the closed acoustic doghouse for **60 continuous minutes**; verify unobstructed cooling airflow exiting the upper Z-baffle labyrinth and measure temperatures:
  - Ambient Outdoor Temp: `_______ °F` | Inside Doghouse Air Temp: `_______ °F` *(Target `ΔT <= 20°F`)* | Motor Frame Temp: `_______ °F` *(Target `<= 140°F`)*
- [ ] **IMC 501.3.1 / OSHA Stack Setback Final Check:** Confirm vertical outdoor exhaust stack terminates **`>= 12–15 ft` above grade**, **`>= 30 ft`** from property lines (`Salt` mode; `>= 10 ft` non-salt), **`>= 10 ft`** from operable windows/doors/air intakes, and **`>= 6 ft`** from exterior walls/roofs (`>= 3 ft` non-salt).

| Sign-Off Role | Printed Name / License # | Signature | Date |
| :--- | :--- | :--- | :--- |
| **Mechanical / Fabrication Lead** | | | |
| **Licensed Electrician (NEC)** | | | |
| **Certified Backflow Tester (IPC 608, if RPZ)** | | | |
| **Studio Owner / Kiln Master** | | | |

---

## 7. Cross-References

- **US National Code Compliance Audit (C1–C14):** [docs/CODE_COMPLIANCE_AUDIT.md](CODE_COMPLIANCE_AUDIT.md)
- **Five Control Architectures & Complete Raspberry Pi Safety BOM:** [docs/CONTROL_ALTERNATIVES_AND_RPI_BOM.md](CONTROL_ALTERNATIVES_AND_RPI_BOM.md)
- **Stage Design Options, Parallel Piping, Eductor & Closed-Loop Sump Study:** [docs/STAGE_DESIGN_OPTIONS.md](STAGE_DESIGN_OPTIONS.md)
- **Budget Engineering Study (Shell-Less Bundles, Affordable Blowers & Compressors):** [docs/BUDGET_ENGINEERING_STUDY.md](BUDGET_ENGINEERING_STUDY.md)
- **Validated Scrubber Physics & Equations:** [docs/ENGINEERING_VALIDATION.md](ENGINEERING_VALIDATION.md)
- **Studio Hardware Catalog, Shop Plan & Build Guide Generators:** [pkg/studio/cots.go](../pkg/studio/cots.go), [pkg/studio/plan.go](../pkg/studio/plan.go), [pkg/studio/guide.go](../pkg/studio/guide.go)
