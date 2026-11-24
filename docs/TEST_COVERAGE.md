# Test Coverage Report (Statement, Line & AST Branch Coverage)

Generated automatically by [`scripts/branch_coverage.go`](../scripts/branch_coverage.go) from `go test -race -covermode=atomic -coverprofile=coverage.out ./...`.

## Executive Summary

| Metric | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| **Statements** (matches `go tool cover -func`) | 6414 | 6868 | **93.4%** |
| **Executable Lines** (token-mapped) | 10344 | 10801 | **95.8%** |
| **AST Control-Flow Branches** (C1 decision) | 2689 | 3134 | **85.8%** |

## Per-Package Coverage Summary

| Package | Statements | Statement % | Exec Lines | Line % | AST Branches | Branch % | Status (≥85% Line / ≥80% Branch) |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: | :---: |
| `cmd/kilnscrub` | 469 / 506 | 92.7% | 594 / 631 | 94.1% | 225 / 263 | 85.6% | PASS |
| `internal/control` | 869 / 1033 | 84.1% | 1064 / 1231 | 86.4% | 404 / 532 | 75.9% | Below Branch |
| `internal/web` | 409 / 459 | 89.1% | 421 / 471 | 89.4% | 169 / 210 | 80.5% | PASS |
| `pkg/cad` | 734 / 761 | 96.5% | 1126 / 1153 | 97.7% | 251 / 280 | 89.6% | PASS |
| `pkg/model` | 168 / 170 | 98.8% | 223 / 225 | 99.1% | 102 / 104 | 98.1% | PASS |
| `pkg/scrubber` | 379 / 418 | 90.7% | 527 / 566 | 93.1% | 140 / 177 | 79.1% | Below Branch |
| `pkg/studio` | 3056 / 3173 | 96.3% | 6021 / 6138 | 98.1% | 1251 / 1400 | 89.4% | PASS |
| `pkg/thermo` | 330 / 348 | 94.8% | 368 / 386 | 95.3% | 147 / 168 | 87.5% | PASS |
| **TOTAL** | **6414 / 6868** | **93.4%** | **10344 / 10801** | **95.8%** | **2689 / 3134** | **85.8%** | — |

## Per-File Breakdown

| File | Statements | Exec Lines | AST Branches |
| :--- | ---: | ---: | ---: |
| `cmd/kilnscrub/main.go` | 469/506 (92.7%) | 594/631 (94.1%) | 225/263 (85.6%) |
| `internal/control/config.go` | 48/48 (100.0%) | 52/52 (100.0%) | 40/40 (100.0%) |
| `internal/control/controller.go` | 96/109 (88.1%) | 150/163 (92.0%) | 31/44 (70.5%) |
| `internal/control/daemon.go` | 77/81 (95.1%) | 82/86 (95.3%) | 30/36 (83.3%) |
| `internal/control/filter.go` | 37/40 (92.5%) | 41/44 (93.2%) | 10/10 (100.0%) |
| `internal/control/hub.go` | 43/43 (100.0%) | 45/45 (100.0%) | 19/20 (95.0%) |
| `internal/control/interlock.go` | 58/59 (98.3%) | 74/75 (98.7%) | 28/29 (96.6%) |
| `internal/control/modbus_hal.go` | 139/161 (86.3%) | 203/225 (90.2%) | 84/103 (81.6%) |
| `internal/control/modbus_rtu.go` | 186/219 (84.9%) | 194/230 (84.3%) | 96/129 (74.4%) |
| `internal/control/serial_linux.go` | 0/48 (0.0%) | 0/48 (0.0%) | 0/26 (0.0%) |
| `internal/control/sim_hal.go` | 185/225 (82.2%) | 223/263 (84.8%) | 66/95 (69.5%) |
| `internal/web/server.go` | 193/220 (87.7%) | 200/227 (88.1%) | 68/92 (73.9%) |
| `internal/web/studio.go` | 82/88 (93.2%) | 82/88 (93.2%) | 36/40 (90.0%) |
| `internal/web/telemetry.go` | 134/151 (88.7%) | 139/156 (89.1%) | 65/78 (83.3%) |
| `pkg/cad/bom.go` | 124/129 (96.1%) | 333/338 (98.5%) | 34/40 (85.0%) |
| `pkg/cad/dxf.go` | 186/198 (93.9%) | 287/299 (96.0%) | 57/68 (83.8%) |
| `pkg/cad/exporter.go` | 131/133 (98.5%) | 173/175 (98.9%) | 91/92 (98.9%) |
| `pkg/cad/stl.go` | 293/301 (97.3%) | 333/341 (97.7%) | 69/80 (86.2%) |
| `pkg/model/types.go` | 128/128 (100.0%) | 183/183 (100.0%) | 102/104 (98.1%) |
| `pkg/model/units.go` | 40/42 (95.2%) | 40/42 (95.2%) | 0/0 (100.0%) |
| `pkg/scrubber/blower.go` | 69/71 (97.2%) | 110/112 (98.2%) | 23/26 (88.5%) |
| `pkg/scrubber/packedbed.go` | 104/116 (89.7%) | 139/151 (92.1%) | 31/40 (77.5%) |
| `pkg/scrubber/scrubber.go` | 68/78 (87.2%) | 102/112 (91.1%) | 31/41 (75.6%) |
| `pkg/scrubber/venturi.go` | 138/153 (90.2%) | 176/191 (92.1%) | 55/70 (78.6%) |
| `pkg/studio/cots.go` | 402/411 (97.8%) | 879/888 (99.0%) | 165/175 (94.3%) |
| `pkg/studio/frugal.go` | 487/500 (97.4%) | 994/1007 (98.7%) | 191/212 (90.1%) |
| `pkg/studio/guide.go` | 310/313 (99.0%) | 402/405 (99.3%) | 102/106 (96.2%) |
| `pkg/studio/options.go` | 546/595 (91.8%) | 920/969 (94.9%) | 209/263 (79.5%) |
| `pkg/studio/packet.go` | 126/145 (86.9%) | 219/238 (92.0%) | 42/59 (71.2%) |
| `pkg/studio/plan.go` | 514/518 (99.2%) | 1544/1548 (99.7%) | 215/219 (98.2%) |
| `pkg/studio/presets.go` | 215/224 (96.0%) | 397/406 (97.8%) | 129/140 (92.1%) |
| `pkg/studio/recirc.go` | 71/76 (93.4%) | 122/127 (96.1%) | 30/36 (83.3%) |
| `pkg/studio/urban.go` | 385/391 (98.5%) | 544/550 (98.9%) | 168/190 (88.4%) |
| `pkg/thermo/quench.go` | 148/160 (92.5%) | 174/186 (93.5%) | 58/72 (80.6%) |
| `pkg/thermo/shomate.go` | 182/188 (96.8%) | 194/200 (97.0%) | 89/96 (92.7%) |

## Branch Coverage by Construct Type

| Construct Kind | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| `for-body` | 31 | 33 | 93.9% |
| `for-exit` | 30 | 31 | 96.8% |
| `if-else` | 107 | 109 | 98.2% |
| `if-else-if` | 51 | 51 | 100.0% |
| `if-implicit-else` | 1087 | 1167 | 93.1% |
| `if-then` | 996 | 1327 | 75.1% |
| `range-body` | 57 | 57 | 100.0% |
| `range-exit` | 52 | 57 | 91.2% |
| `select-case` | 9 | 10 | 90.0% |
| `select-default` | 2 | 3 | 66.7% |
| `switch-case` | 196 | 211 | 92.9% |
| `switch-default` | 62 | 66 | 93.9% |
| `switch-implicit-default` | 6 | 9 | 66.7% |
| `type-switch-case` | 2 | 2 | 100.0% |
| `type-switch-implicit-default` | 1 | 1 | 100.0% |

## Highest-Value Uncovered Branches (Prioritized Test Targets)

Packages below the **85% executable line** or **80% AST branch** target are listed first, followed by remaining packages. Within each package, explicit `switch`/`select` arms, `else`/`else if` paths, and loop entry/exit branches are prioritized ahead of defensive error guards.

### `internal/control` (Line 86.4%, Branch 75.9%, 128 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `internal/control/hub.go:84` | `select-default` | `(*TelemetryHub).Publish` | `default` |
| `internal/control/modbus_rtu.go:381` | `switch-default` | `(*ModbusRTUClient).roundTrip` | `default` |
| `internal/control/serial_linux.go:82` | `switch-case` | `linuxBaudFlag` | `case 1200` |
| `internal/control/serial_linux.go:84` | `switch-case` | `linuxBaudFlag` | `case 2400` |
| `internal/control/serial_linux.go:86` | `switch-case` | `linuxBaudFlag` | `case 4800` |
| `internal/control/serial_linux.go:88` | `switch-case` | `linuxBaudFlag` | `case 9600` |
| `internal/control/serial_linux.go:90` | `switch-case` | `linuxBaudFlag` | `case 19200` |
| `internal/control/serial_linux.go:92` | `switch-case` | `linuxBaudFlag` | `case 38400` |
| `internal/control/serial_linux.go:94` | `switch-case` | `linuxBaudFlag` | `case 57600` |
| `internal/control/serial_linux.go:96` | `switch-case` | `linuxBaudFlag` | `case 115200` |
| `internal/control/serial_linux.go:98` | `switch-case` | `linuxBaudFlag` | `case 230400` |
| `internal/control/serial_linux.go:100` | `switch-default` | `linuxBaudFlag` | `default` |

*...and 116 additional uncovered branches in `internal/control` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/scrubber` (Line 93.1%, Branch 79.1%, 37 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/scrubber/packedbed.go:109` | `for-exit` | `SizePackedBed` | `exit/skip (iter < 8)` |
| `pkg/scrubber/packedbed.go:143` | `for-body` | `SizePackedBed` | `(qGas/(math.Pi*colDiam*colDiam/4.0))/vFlood*100.0 > packedBedMaxFloodPct` |
| `pkg/scrubber/blower.go:158` | `if-then` | `SizeBlower` | `kSys < 0` |
| `pkg/scrubber/blower.go:179` | `if-then` | `SizeBlower` | `pFanVal < 0` |
| `pkg/scrubber/packedbed.go:81` | `if-then` | `EckertFloodingVelocity` | `vFloodSq <= 0 \|\| math.IsNaN(vFloodSq) \|\| math.IsInf(vFloodSq, 0)` |
| `pkg/scrubber/packedbed.go:112` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:139` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:148` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:192` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:264` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/scrubber.go:64` | `if-then` | `computeInletFlueGas` | `tempScale > 1.80` |
| `pkg/scrubber/scrubber.go:98` | `if-then` | `computeInletFlueGas` | `nTotal <= 0` |

*...and 25 additional uncovered branches in `pkg/scrubber` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `internal/web` (Line 89.4%, Branch 80.5%, 41 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `internal/web/server.go:167` | `select-case` | `(*Server).ListenAndServe` | `case err := <-errCh` |
| `internal/web/server.go:48` | `if-then` | `NewServer` | `err != nil` |
| `internal/web/server.go:121` | `if-then` | `(*Server).SetSSEKeepaliveInterval` | `d <= 0` |
| `internal/web/server.go:133` | `if-then` | `(*Server).getSSEKeepaliveInterval` | `d <= 0` |
| `internal/web/server.go:148` | `if-then` | `(*Server).ListenAndServe` | `err != nil` |
| `internal/web/server.go:168` | `if-then` | `(*Server).ListenAndServe` | `errors.Is(err, http.ErrServerClosed)` |
| `internal/web/server.go:177` | `if-then` | `(*Server).ListenAndServe` | `shutdownErr != nil` |
| `internal/web/server.go:180` | `if-then` | `(*Server).ListenAndServe` | `serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed)` |
| `internal/web/server.go:247` | `if-then` | `(*Server).handleSize` | `s.designer == nil` |
| `internal/web/server.go:253` | `if-then` | `(*Server).handleSize` | `err != nil` |
| `internal/web/server.go:279` | `if-then` | `(*Server).handleExport` | `r.Method != http.MethodPost` |
| `internal/web/server.go:284` | `if-then` | `(*Server).handleExport` | `s.designer == nil \|\| s.exporter == nil` |

*...and 29 additional uncovered branches in `internal/web` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `cmd/kilnscrub` (Line 94.1%, Branch 85.6%, 38 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `cmd/kilnscrub/main.go:29` | `if-then` | `main` | `err != nil` |
| `cmd/kilnscrub/main.go:199` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:205` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:324` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:330` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:334` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:338` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:342` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:346` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:350` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:366` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:394` | `if-then` | `runPresets` | `err != nil` |

*...and 26 additional uncovered branches in `cmd/kilnscrub` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/thermo` (Line 95.3%, Branch 87.5%, 21 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/thermo/quench.go:143` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:171` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:176` | `if-then` | `SolveAdiabaticQuench` | `tLow >= tHigh` |
| `pkg/thermo/quench.go:182` | `if-then` | `SolveAdiabaticQuench.func@L180` | `e != nil` |
| `pkg/thermo/quench.go:186` | `if-then` | `SolveAdiabaticQuench.func@L180` | `ywSat >= 0.9999` |
| `pkg/thermo/quench.go:198` | `if-then` | `SolveAdiabaticQuench.func@L180` | `e1 != nil \|\| e2 != nil \|\| e3 != nil \|\| e4 != nil \|\| e5 != nil` |
| `pkg/thermo/quench.go:206` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:212` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:230` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:235` | `if-then` | `SolveAdiabaticQuench` | `err != nil` |
| `pkg/thermo/quench.go:254` | `if-then` | `SolveAdiabaticQuench` | `towerDiam < minQuenchTowerDiameterM` |
| `pkg/thermo/quench.go:273` | `if-then` | `SolveAdiabaticQuench` | `nozzleCount < minQuenchNozzleCount` |

*...and 9 additional uncovered branches in `pkg/thermo` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/studio` (Line 98.1%, Branch 89.4%, 149 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/studio/options.go:210` | `switch-case` | `mediumPressureAirAtomizeBlowerCost` | `case hp <= 2.0` |
| `pkg/studio/options.go:841` | `switch-case` | `eductorEntrainmentRatio` | `case dpInWC <= 4.0` |
| `pkg/studio/presets.go:440` | `switch-default` | `(StudioProject).EffectiveLocalDraftMode` | `default` |
| `pkg/studio/recirc.go:42` | `switch-case` | `workingSumpVolumeGal` | `case "USPLASTIC-100T"` |
| `pkg/studio/recirc.go:52` | `switch-default` | `workingSumpVolumeGal` | `default` |
| `pkg/studio/urban.go:342` | `switch-case` | `assessElectrical` | `case reqAmps <= 50.0` |
| `pkg/studio/guide.go:125` | `if-else` | `GenerateBuildGuide` | `else of (p.CorrosionTier == CorrosionTierHeavyC276FRP)` |
| `pkg/studio/options.go:897` | `if-else` | `analyzeCompressor` | `else of (err == nil)` |
| `pkg/studio/cots.go:108` | `range-exit` | `snapPipe` | `exit/skip (range catalog)` |
| `pkg/studio/cots.go:394` | `range-exit` | `snapPipeByNominal` | `exit/skip (range catalog)` |
| `pkg/studio/options.go:369` | `range-exit` | `analyzeEductor` | `exit/skip (range catalog)` |
| `pkg/studio/options.go:619` | `for-body` | `analyzeParallelBed` | `(qGas/(float64(pipeCount)*pipeAreaM2))/vFlood*100.0 > 75.0` |

*...and 137 additional uncovered branches in `pkg/studio` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/cad` (Line 97.7%, Branch 89.6%, 29 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/cad/bom.go:108` | `if-then` | `(*Exporter).GenerateBOM` | `packingVolM3 < 0.001` |
| `pkg/cad/bom.go:369` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:385` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:401` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:406` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/dxf.go:44` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:47` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:50` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:108` | `if-then` | `(*dxfBuilder).trackPoint` | `!isFinite(x) \|\| !isFinite(y)` |
| `pkg/cad/dxf.go:225` | `if-then` | `(*dxfBuilder).addFrustumPattern` | `dev.IsCylinder` |
| `pkg/cad/dxf.go:303` | `if-then` | `(*dxfBuilder).build` | `!b.hasPts` |
| `pkg/cad/dxf.go:444` | `if-then` | `(*Exporter).ExportDXF` | `err != nil` |

*...and 17 additional uncovered branches in `pkg/cad` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/model` (Line 99.1%, Branch 98.1%, 2 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/model/types.go:614` | `if-implicit-else` | `(*ControlConfig).Sync` | `!(c.DraftFilterTimeConst > 0)` |
| `pkg/model/types.go:639` | `if-implicit-else` | `(ControlConfig).Validate` | `!(expectedTau > 0)` |

## Methodology & Reproducibility

### How Each Metric Is Computed

1. **Statement Coverage (`go tool cover` Exact Match):**
   Parses `coverage.out` blocks (`file:startLine.startCol,endLine.endCol numStmt count`), merges duplicate block ranges using `mode` semantics (`atomic`/`count` sum, `set` max), and computes `sum(numStmt where count > 0) / sum(numStmt)`. The total statement percentage matches `go tool cover -func=coverage.out` identically.

2. **Executable Line Coverage (`go/scanner` Token Mapping):**
   Tokenizes each profiled Go source file with stdlib `go/scanner` (ignoring comments, semicolons, and standalone `{`/`}` braces) and maps every executable token into the exact byte span `[startOffset, endOffset)` of `numStmt > 0` coverage blocks. A source line is counted as an executable line if it contains at least one executable token inside a statement block, and is covered if any covering block on that line executed at least once (`count > 0`).

3. **AST Control-Flow Branch Coverage (Deterministic C1 Decision Coverage):**
   Walks each profiled Go source file's AST (`go/parser` + `go/ast` + `go/token`) and maps Go `cmd/cover` exact basic-block boundaries (matching Go's stdlib `cmd/cover` instrumentation rules) onto real control-flow branches:
   - **`if` statements (`if-then`, `if-else`, `if-else-if`, `if-implicit-else`):** The `then` branch maps to the block starting at `Body.Lbrace`; explicit `else`/`else if` maps to the block starting right after the `else` keyword (`elseOffset + 4`); implicit `else` (no `else` clause) uses the exact atomic/count execution balance `max(0, entryCount - thenCount)`.
   - **`switch` & type-switch statements (`switch-case`, `switch-default`, `switch-implicit-default`, `type-switch-*`):** Each `case`/`default` arm maps to the `cmd/cover` block starting at `clause.Colon + 1` (including empty case arms, which `cmd/cover` emits as `numStmt=0` counter blocks). Switches without an explicit `default:` also track `switch-implicit-default` via `max(0, entryCount - (sum(caseCounts) - fallthroughHits))`.
   - **`select` statements (`select-case`, `select-default`):** Each `case` and optional `default` communication arm maps to its `clause.Colon + 1` block. (`select` without `default` blocks until ready and has no implicit default branch.)
   - **`for` & `range` loops (`for-body`, `for-exit`, `range-body`, `range-exit`):** Body entry maps to `Body.Lbrace`. Conditional `for` loops and `range` loops also track normal loop exit/skip (`for-exit` / `range-exit`) by accounting for post-loop block hits minus `break` statements targeting the loop (or `entryCount - earlyExits` when the loop is the final statement in its block). Unconditional `for { ... }` loops have no header condition to evaluate false and track `for-body` only.
   - **Compound boolean expressions (`&&`, `||`):** Go's native `go test -covermode=atomic` instruments statement basic blocks and does not emit sub-expression short-circuit counters. Compound conditions are therefore measured as a single decision at their enclosing `if`/`for` statement (strict C1 branch coverage) rather than unmeasurable synthetic MC/DC sub-conditions.
   - **Build-constrained files:** Only files compiled for the host target OS/arch appear in `coverage.out` (e.g., `internal/control/serial_other.go` with `//go:build !linux` is excluded on Linux builds).

### Reproducing Locally & CI Thresholds

```bash
# Via Makefile target (wired by repo build tooling):
make coverage-branch

# Or directly with Go:
CGO_ENABLED=1 go test -race -covermode=atomic -coverprofile=coverage.out ./...
go run ./scripts/branch_coverage.go -profile coverage.out -markdown docs/TEST_COVERAGE.md

# Enforce CI minimum thresholds (exits 1 if below threshold):
go run ./scripts/branch_coverage.go -profile coverage.out -min-line 80 -min-branch 70
```
