# `kilnscrub` Swarm Handoff (Ready for Phase 2 on Debian)

Start a fresh `/swarm` session in Debian and tell the coordinator:
> *"Read `plan-from-chat.md` and `SWARM_HANDOFF.md`, verify `go test -race ./...` passes, and execute Phases 2 & 3."*

---

## 1. Completed & Verified Work (Do NOT Redo)

1. **Phase 1 Foundation (`pkg/model`, `go.mod`, `cmd/kilnscrub/main.go`):**
   - Module: `github.com/AaronFilson/kilnscrub` (`go 1.24.13`).
   - Pure-Go dependencies already pinned in `go.mod` / `go.sum`:
     - `gonum.org/v1/gonum v0.17.0`
     - `go.einride.tech/pid v0.2.0`
     - `github.com/brandondube/pctl v1.3.0`
   - `pkg/model/units.go` & `pkg/model/types.go` hardened after an independent foundation audit:
     - Strict SI types + NIST unit converters (`°F↔K`, `inWC↔Pa`, `CFM↔m³/s`, `GPM↔m³/s`, `lbs/hr↔kg/s`, `in/ft/mm/µm↔m`, `HP↔W`).
     - Finite (`NaN`/`Inf`) validation across all 18 numeric fields in `DesignInput.Validate()`, `TargetKilnDraftInWC ∈ [-2.0, 0]`, `PackingFactor > 0`, `FuelType` constants.
     - Reconciled LPF defaults in `DefaultControlConfig()` (`DraftFilterCutoffHz = 0.5 Hz`, `DraftFilterTimeConst ≈ 318.31 ms`), `AntiWindUpGain` & non-zero `DerivativeFilterTimeConst` in `PIDGains`, plus `ControlConfig.Validate()` / `Sync()` / `SetTargetDraftInWC()`.
     - `SafetyLimits` default CPVC protection threshold: `MaxQuenchOutletTemp = 210°F (372.04 K)`, `MinWaterFlow = 0.25 GPM`.
     - `BOM.RecalculateTotals()` nil-safe and recomputes whenever `UnitCostUSD > 0`.
2. **Phase 2 Stage A — Scientific Oracle Fixtures (Already Generated & Audited):**
   - `scripts/generate_oracle.py` and `testdata/oracle_fixtures.json` are already generated (using `chemicals 1.5.2`, `fluids 1.3.1`, `scipy`) and independently physics-verified with zero formula/unit/sign errors:
     - NIST Shomate `Cp(T)` & `H(T)-H(298.15)` for `N2, CO2, H2O, O2, HCl` (`350–1588.15 K`).
     - `chemicals.viscosity.Lucas_gas` + embedded `species_constants`.
     - `chemicals.iapws.iapws95_Psat` water vapor pressure (`300–373.15 K`).
     - `fluids.packed_bed` Ergun pressure drop (`1"` Pall rings, `voidage=0.90`, `L=1.2 m`, `<3e-16` analytical error).
     - Calvert (1970, 1972) + Davies/Allen-Raabe Cunningham slip `d50` for condensed NaCl (`2160 kg/m³`, `Pt(d50)=0.5` verified to `1e-12`).
   - **You do NOT need Python, `fluids`, or `chemicals` installed on Debian** — Go tests read only `testdata/oracle_fixtures.json`.

---

## 2. Phase 2 Parallel Swarm Layout (Strict Directory Ownership)

Use `workspace:"share"` (isolated git worktrees) for all 4 Phase 2 authors, and `workspace:"inherit"` for verifiers and Phase 3 integration:

| Worker | Exclusive Ownership | Deliverables |
|---|---|---|
| `thermo` | `pkg/thermo/`, `pkg/scrubber/` | `shomate.go`, `quench.go`, `thermo_test.go`, `venturi.go`, `packedbed.go`, `blower.go`, `scrubber.go`, `scrubber_test.go` (`model.Designer`) |
| `cad` | `pkg/cad/` | `stl.go` (binary mm STL), `dxf.go` (R12 unfolded patterns + exact cone developments), `bom.go` (CSV + Markdown), `cad_test.go` (`model.CADExporter`) |
| `control` | `internal/control/` | `filter.go`, `interlock.go`, `controller.go`, `sim_hal.go`, `daemon.go`, `control_test.go` |
| `web` | `internal/web/` | `server.go` (`//go:embed`, `/api/health`, `/api/defaults`, `/api/size`, `/api/export`), `static/index.html` (live sliders, Three.js cutaway, SVG/Canvas performance curves), `server_test.go` |
| `architect` | `go.mod`, `go.sum`, `cmd/kilnscrub/main.go` | Phase 3 merge of all 4 branches, `go mod tidy`, full CLI wiring (`design`, `export`, `daemon`, `ui`) |
| `verifier-eng` | Read-only (`inherit`) | Verify `1e-6` oracle tolerance, zero Python calls in `go test`, `go test -race ./...`, CAD mesh/DXF/BOM checks, 24 cu.ft. kiln CLI run |
| `verifier-ui` | Read-only (`inherit`) | Launch `kilnscrub ui`, verify `/api/*` endpoints, inspect live `http://localhost:8080` Three.js cutaway & sliders in headless browser, flag screenshots |

---

## 3. Critical Cross-Agent Contracts & Library Traps (From Phase 1 Audit)

Pass these exact notes in Phase 2 worker briefs to avoid subtle integration bugs:

### Shared `pkg/model` Conventions
- **Percent vs Fraction scales:**
  - `PackedBedResult.PercentOfFlood` is **`0..100`** (constraint: `<= 75.0`).
  - `HClRemovalEfficiency`, `EfficiencyPoint.WoodAshEfficiency`, `EfficiencyPoint.NaClFumeEfficiency`, `Overall*Efficiency`, and `BlowerResult.Efficiency` are **`0..1`** (multiply by `100` for UI `%` display).
- **Draft sign convention:**
  - `DesignInput.TargetKilnDraftInWC` and `BlowerResult.TargetKilnDraftInWC` are **negative** (`-0.05 inWC` default).
  - `PressureDropBreakdown.KilnDraft` / `KilnDraftInWC` store the **positive draw magnitude** (`+0.05 inWC`), so `Breakdown.TotalPa()` sums directly to `BlowerResult.TotalStaticPressure`.
- **No double-counting demister dP:** `PressureDropBreakdown.PackedBed` is **packing bed only** (`BedPressureDrop`, excluding `DemisterPressureDrop`).
- **Particulate diameter:** `EfficiencyPoint` uses **physical particle diameter** `d_p`; `VenturiResult.CutDiameterD50` is condensed NaCl fume `d_50` (`2160 kg/m³`), while `WoodAshCutDiameterD50` uses wood ash (`2400 kg/m³`).
- **BOM weights:** `BOMItem.WeightKg` is **total line-item weight** (not per-unit). Call `bom.RecalculateTotals()`.

### `thermo` (`pkg/thermo` & `pkg/scrubber`)
- **`gonum v0.17.0` has no 1D Brent root finder** in `optimize` — implement a clean pure-Go Brent 1D root finder in `pkg/thermo/quench.go` (using `gonum` where helpful).
- **N2 Shomate boundary at `T = 500.0 K`:** `scripts/generate_oracle.py` uses lower-range-wins (`tmin <= T <= tmax`), picking `(100, 500)` at `500.0 K` (`H = 5910.754 J/mol` vs `5911.523 J/mol` in `(500, 2000)` — a `1.3e-4` jump > `1e-6`). Match lower-range-wins `<=` in `pkg/thermo/shomate.go`.
- **Calvert `F(Kp, f)` numerical stability:** Use `math.Log1p(x / 0.7)` plus a small-`x` Taylor series fallback for tiny particles to avoid catastrophic cancellation.
- Verify `<= 1e-6` relative tolerance against `testdata/oracle_fixtures.json` for `Cp`, `H`, `mu`, `Psat`, Ergun `dP`, and Calvert `d50`.

### `control` (`internal/control`)
- **`go.einride.tech/pid v0.2.0` sign convention:** `einride/pid` computes `e = ReferenceSignal - ActualSignal` (`SP - PV`). Both **draft control** (higher/less-negative pressure -> increase blower VFD) and **quench cooling** (higher outlet temp -> increase spray valve) are **direct-acting (`PV - SP`)** — pass `ReferenceSignal: actual, ActualSignal: setpoint` (or negate error) so positive gains in `DefaultControlConfig()` don't cause positive-feedback runaway.
- **`pid.AntiWindupController` div-by-zero guard:** `antiwindupcontroller.go` divides by `LowPassTimeConstant.Seconds()`, returning `NaN` if `0`. Always wire `PIDGains.DerivativeFilterTimeConst` (non-zero) and `PIDGains.AntiWindUpGain` from `model.ControlConfig`.
- **`github.com/brandondube/pctl v1.3.0` LPF wrapper:** `pctl.NewLPF(cutoffFreq, dT)` initializes `prev = 0` (startup jump from `0` to `-12.45 Pa`) with no `Reset()`, and divides by zero if `cutoffFreq == 0`. Wrap it with first-sample seeding and a `cutoffFreq > 0` guard. **Do NOT use `pctl.HPF`** (upstream bug ignores `input`).
- **Fail-safe interlocks:** Latch `BypassDamperOpen = true` & `AlarmLatched = true` on `QuenchOutletTemp >= 210°F (372.04 K)`, `WaterFlow < MinWaterFlow`, or HAL sensor fault/NaN/stale. Require fault clearance + explicit `Reset()`.

### `web` (`internal/web`)
- **Partial JSON decode safety:** In `POST /api/size` and `POST /api/export`, initialize `input := model.DefaultSaltKilnInput()` *before* `json.Unmarshal(body, &input)` so omitted fields (e.g., `packing_factor_1_m`, `ambient_temp_k`) inherit valid defaults instead of failing `Validate()`.
- Ensure Three.js (`three@0.170.0` + `OrbitControls`) renders cleanly with cutaway toggle, flow particles toggle, stage inspector pills, and graceful WebGL fallback.

---

## 4. Phase 3 Integration & Verification Gate

1. Merge all 4 Phase 2 branches into `main`, wire `cmd/kilnscrub/main.go` (`design`, `export`, `daemon`, `ui`), and run `go mod tidy`.
2. Verify all checks pass with **zero Python runtime dependency**:
   ```bash
   go vet ./...
   CGO_ENABLED=0 go build -o /tmp/kilnscrub ./cmd/kilnscrub
   CGO_ENABLED=1 go test -race -count=1 -v ./...
   ```
3. Run end-to-end CLI sizing & export for the reference **24 cu.ft. wood/salt kiln (2400°F peak, 2 lbs NaCl/hr)**:
   ```bash
   /tmp/kilnscrub design --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0
   /tmp/kilnscrub export --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0 --out /tmp/kilnscrub-out
   ```
   Verify non-empty `.stl` (valid binary triangle count & normals), `.dxf` (`ENTITIES`/`EOF`), `bom.csv`, and `bom.md`.
4. Start `/tmp/kilnscrub ui --addr 127.0.0.1:8080`, verify `/api/health`, `/api/defaults`, `/api/size`, `/api/export`, and inspect/screenshot `http://127.0.0.1:8080` in a real browser.

---

## 5. Debian WSL Setup Quick Reference

1. **Keep WSL off `C:` (optional if `C:` is tight):** Export/import Debian onto `D:\wsl\debian` (`wsl --export Debian D:\wsl\debian.tar && wsl --import Debian-D D:\wsl\debian D:\wsl\debian.tar`), then clone `kilnscrub` onto Linux ext4 (`~/src/kilnscrub`) rather than working across `/mnt/c` or `/mnt/d`.
2. **Install Go 1.24 + C compiler for `-race`:**
   ```bash
   sudo apt-get update && sudo apt-get install -y build-essential git curl ca-certificates
   # Install Go 1.24.x (matching go.mod >= 1.24) into /usr/local/go and add /usr/local/go/bin to PATH
   ```
3. **Apply Antigravity CLI permissions:** See `debian-permissions-template.json` in this repo (or add those rules via `/permissions` under **Project** scope).
