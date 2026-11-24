# Project Design Document: `kilnscrub` (Pure-Go Kiln Exhaust Scrubber Designer & Active Controller)

## 1. Objective
Build `kilnscrub`, a zero-dependency Go application (`go 1.24`) that:
1. Calculates scientifically validated engineering dimensions for a 4-stage studio kiln wet scrubber (Adiabatic Spray Quench -> High-Velocity Venturi -> Counter-Current Packed-Bed HCl Absorber -> Induced-Draft Blower).
2. Generates fabrication-ready 3D meshes (`.stl`), 2D unfolded sheet-metal cut templates (`.dxf`), and a Bill of Materials (BOM).
3. Runs an embedded web dashboard (`localhost:8080`) with interactive 3D cutaway visualization (Three.js) and performance curves.
4. Provides a closed-loop hardware daemon (`kilnscrub daemon`) with industrial PID draft control, quench temperature modulation, and fail-safe bypass interlocks.

## 2. Repository Directory & Strict Agent Ownership
To prevent merge conflicts during `/swarm` execution, each agent owns its assigned directory exclusively and communicates only through the defined Go interfaces in `pkg/model`.

```text
kilnscrub/
├── go.mod                          # [Agent 1: Architect] Module & dependency manifest
├── cmd/kilnscrub/main.go           # [Agent 1: Architect] CLI entrypoints (design, export, daemon, ui)
├── pkg/model/types.go              # [Agent 1: Architect] Shared structs, SI units, HAL interfaces
├── scripts/generate_oracle.py      # [Agent 2: Thermo & ChemEng] One-time Python oracle script (`fluids` + `chemicals`)
├── testdata/oracle_fixtures.json   # [Agent 2: Thermo & ChemEng] Ground-truth JSON physics fixtures for Go unit tests
├── pkg/thermo/                     # [Agent 2: Thermo & ChemEng]
│   ├── shomate.go                  # NIST 7-coeff Cp(T) & enthalpy for N2, CO2, H2O, O2, HCl (300-2500K)
│   ├── quench.go                   # Adiabatic saturation solver (gonum root.Brent) & water GPM
│   └── thermo_test.go              # Table-driven tests verified against testdata/oracle_fixtures.json
├── pkg/scrubber/                   # [Agent 2: Thermo & ChemEng]
│   ├── venturi.go                  # Calvert cut-diameter particulate efficiency & throat dP
│   ├── packedbed.go                # Souders-Brown/Eckert flooding diameter, Ergun dP, NTU-HTU HCl removal
│   ├── blower.go                   # Static pressure summation & fan operating point solver
│   └── scrubber_test.go            # Unit tests validating L/G ratios, flooding limits (210F / water loss -> bypass damper)
│   └── control_test.go             # Closed-loop step-response simulation against wood-stoking heat spikes
└── internal/web/                   # [Agent 5: Web UI]
    ├── server.go                   # //go:embed HTTP server & JSON API endpoints (/api/size, /api/export)
    └── static/index.html           # Single-file interactive dashboard + Three.js 3D scrubber viewer
```

## 3. Core Scientific & Engineering Specifications (`pkg/thermo` & `pkg/scrubber`)
- **Stage 1: Adiabatic Quench Tower (316L Stainless Steel)**
  - Cool flue gas from kiln exit temp (up to 2400°F / 1588 K) to adiabatic saturation (~165°F / 347 K) via atomized water spray.
  - Use `gonum.org/v1/gonum/optimize` / Brent root-finding to solve energy balance: `H_gas(T_in) + m_w * h_liquid = H_gas_sat(T_sat)` and verify stack dew point via `://github.com`.
- **Stage 2: Wet Venturi Scrubber (316L Stainless Steel)**
  - Throat velocity target: 50–90 m/s (165–295 ft/s).
  - Compute particulate cut diameter `d_50` and fractional collection efficiency for 0.1–2.0 µm wood ash and condensed NaCl fume using the **Calvert equation**; calculate throat pressure drop (`inWC`).
- **Stage 3: Counter-Current Packed-Bed Absorber (CPVC / Polypropylene)**
  - Sizing for HCl gas neutralization (from salt firing) using recirculating NaHCO3 solution (pH 7.5–8.5).
  - Column diameter: **Souders-Brown / Eckert flooding correlation** constrained so superficial gas velocity `<= 75%` of flooding velocity.
  - Packing height: `Z = NTU * HTU` for >=98% HCl removal; bed pressure drop via **Ergun equation**; include chevron mist eliminator pressure drop.
- **Stage 4: Induced-Draft (ID) Centrifugal Blower**
  - Placed downstream of demister. Total required static pressure = `dP_duct + dP_quench + dP_venturi + dP_packedbed + dP_demister` + target kiln draft (`-0.05 inWC`).

## 4. Scientific Oracle Verification Script (`scripts/generate_oracle.py`)
Agent 2 must create and run this script once during Phase 2 (`pip install fluids chemicals`) to generate `testdata/oracle_fixtures.json`, ensuring the pure-Go math matches peer-reviewed Python chemical engineering libraries to within `1e-6` relative tolerance:

```python
#!/usr/bin/env python3
"""Generates ground-truth thermophysical and scrubber test fixtures for Go unit tests."""
import json
import math
from chemicals import heat_capacity, viscosity, vapor_pressure
from fluids.packed_bed import Ergun
from fluids.separator import Calvert_cut_diameter

def generate_fixtures():
    fixtures = {
        "gas_properties": [],
        "ergun_packed_bed": [],
        "calvert_venturi": []
    }

    # 1. Flue gas species Cp (J/mol/K) & viscosity (Pa*s) across kiln temps (350K to 1600K)
    species = {"N2": "7727-37-9", "CO2": "124-38-9", "H2O": "7732-18-5", "O2": "7782-44-7", "HCl": "7647-01-0"}
    for T in [350.0, 500.0, 900.0, 1300.0, 1588.15]:
        entry = {"T_K": T, "species": {}}
        for name, cas in species.items():
            cp = heat_capacity.Cp_gas_Shomate(T, CASRN=cas) if hasattr(heat_capacity, "Cp_gas_Shomate") else heat_capacity.Cp_gas(T, CASRN=cas)
            mu = viscosity.viscosity_gas(T, CASRN=cas)
            entry["species"][name] = {"Cp_J_mol_K": cp, "mu_Pa_s": mu}
        fixtures["gas_properties"].append(entry)

    # 2. Packed-bed pressure drop (Ergun equation for 1" / 0.0254m Pall rings, voidage=0.90)
    for vs in [0.5, 1.0, 1.5, 2.0]:  # Superficial gas velocity (m/s)
        dp = Ergun(dp=0.0254, voidage=0.90, vs=vs, rho=1.05, mu=2.05e-5, L=1.2)
        fixtures["ergun_packed_bed"].append({
            "dp_m": 0.0254, "voidage": 0.90, "vs_m_s": vs,
            "rho_kg_m3": 1.05, "mu_Pa_s": 2.05e-5, "L_m": 1.2,
            "dP_Pa": dp
        })

    # 3. Venturi particulate cut diameter (Calvert model)
    for v_throat in [50.0, 70.0, 90.0]:  # Throat velocity (m/s)
        for L_G in [0.0008, 0.0015]:     # Liquid-to-gas volumetric ratio (m3/m3)
            d50 = Calvert_cut_diameter(
                mu_g=2.1e-5, d_droplet=50e-6, v_throat=v_throat,
                rho_p=2160.0, L_G=L_G  # rho_p = 2160 kg/m3 (condensed NaCl salt)
            )
            fixtures["calvert_venturi"].append({
                "mu_g_Pa_s": 2.1e-5, "d_droplet_m": 50e-6, "v_throat_m_s": v_throat,
                "rho_p_kg_m3": 2160.0, "L_G": L_G, "d50_m": d50
            })

    with open("testdata/oracle_fixtures.json", "w") as f:
        json.dump(fixtures, f, indent=2)
    print("Wrote testdata/oracle_fixtures.json successfully.")

if __name__ == "__main__":
    generate_fixtures()
```

## 5. Swarm Execution & Verification Plan
- **Phase 1 (Foundation — Agent 1):** Initialize `go.mod`, write `pkg/model/types.go` with strict unit types (`Kelvin`, `Pascal`, `CubicMetersPerSec`) and `://github.com` helpers.
- **Phase 2 (Parallel Implementation — Agents 2, 3, 4, 5):**
  - Agent 2 writes and runs `scripts/generate_oracle.py` to generate `testdata/oracle_fixtures.json`, then implements `pkg/thermo` and `pkg/scrubber` with table-driven tests passing against the oracle fixtures.
  - Agent 3 implements `pkg/cad` generating `.stl`, unfolded `.dxf` patterns, and CSV/Markdown BOMs.
  - Agent 4 implements `internal/control` with `go.einride.tech/pid`, `pctl` low-pass draft filtering, fail-safe interlocks, and simulated kiln transient tests.
  - Agent 5 builds `internal/web` with live parameter sliders and Three.js parametric cylinder/cone rendering.
- **Phase 3 (Integration & Independent Verification):**
  - Wire all commands in `cmd/kilnscrub/main.go`.
  - Run `go test -race ./...` (verifying zero Python runtime dependency in Go tests—only reading the committed `testdata/oracle_fixtures.json`).
  - Execute a CLI sizing run for a 24 cu.ft. wood/salt kiln (2400°F peak, 2 lbs NaCl/hr), verify non-empty `.stl`/`.dxf` outputs, and inspect the live `localhost:8080` UI in a browser.
