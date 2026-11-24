# Test Coverage Report (Statement, Line & AST Branch Coverage)

Generated automatically by [`scripts/branch_coverage.go`](../scripts/branch_coverage.go) from `go test -race -covermode=atomic -coverprofile=coverage.out ./...`.

## Executive Summary

| Metric | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| **Statements** (matches `go tool cover -func`) | 6892 | 7434 | **92.7%** |
| **Executable Lines** (token-mapped) | 10912 | 11463 | **95.2%** |
| **AST Control-Flow Branches** (C1 decision) | 2890 | 3409 | **84.8%** |

## Per-Package Coverage Summary

| Package | Statements | Statement % | Exec Lines | Line % | AST Branches | Branch % | Status (≥85% Line / ≥72% Branch CI Gate) |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: | :---: |
| `cmd/kilnscrub` | 489 / 535 | 91.4% | 614 / 660 | 93.0% | 236 / 285 | 82.8% | PASS |
| `internal/control` | 936 / 1126 | 83.1% | 1147 / 1345 | 85.3% | 422 / 570 | 74.0% | PASS |
| `internal/web` | 465 / 539 | 86.3% | 477 / 551 | 86.6% | 199 / 260 | 76.5% | PASS |
| `pkg/cad` | 750 / 777 | 96.5% | 1142 / 1169 | 97.7% | 259 / 288 | 89.9% | PASS |
| `pkg/model` | 168 / 170 | 98.8% | 223 / 225 | 99.1% | 102 / 104 | 98.1% | PASS |
| `pkg/scrubber` | 386 / 427 | 90.4% | 534 / 575 | 92.9% | 142 / 181 | 78.5% | PASS |
| `pkg/studio` | 3368 / 3512 | 95.9% | 6407 / 6552 | 97.8% | 1383 / 1553 | 89.1% | PASS |
| `pkg/thermo` | 330 / 348 | 94.8% | 368 / 386 | 95.3% | 147 / 168 | 87.5% | PASS |
| **TOTAL** | **6892 / 7434** | **92.7%** | **10912 / 11463** | **95.2%** | **2890 / 3409** | **84.8%** | — |

## Per-File Breakdown

| File | Statements | Exec Lines | AST Branches |
| :--- | ---: | ---: | ---: |
| `cmd/kilnscrub/main.go` | 489/535 (91.4%) | 614/660 (93.0%) | 236/285 (82.8%) |
| `internal/control/config.go` | 52/53 (98.1%) | 56/57 (98.2%) | 41/42 (97.6%) |
| `internal/control/controller.go` | 127/145 (87.6%) | 187/210 (89.0%) | 36/50 (72.0%) |
| `internal/control/daemon.go` | 82/89 (92.1%) | 87/94 (92.6%) | 33/40 (82.5%) |
| `internal/control/filter.go` | 37/40 (92.5%) | 41/44 (93.2%) | 10/10 (100.0%) |
| `internal/control/hub.go` | 42/43 (97.7%) | 45/46 (97.8%) | 18/20 (90.0%) |
| `internal/control/interlock.go` | 74/76 (97.4%) | 97/99 (98.0%) | 34/37 (91.9%) |
| `internal/control/modbus_hal.go` | 140/161 (87.0%) | 205/226 (90.7%) | 82/101 (81.2%) |
| `internal/control/modbus_rtu.go` | 197/242 (81.4%) | 206/254 (81.1%) | 102/147 (69.4%) |
| `internal/control/serial_linux.go` | 0/52 (0.0%) | 0/52 (0.0%) | 0/28 (0.0%) |
| `internal/control/sim_hal.go` | 185/225 (82.2%) | 223/263 (84.8%) | 66/95 (69.5%) |
| `internal/web/server.go` | 237/282 (84.0%) | 245/290 (84.5%) | 90/128 (70.3%) |
| `internal/web/studio.go` | 85/94 (90.4%) | 85/94 (90.4%) | 39/46 (84.8%) |
| `internal/web/telemetry.go` | 143/163 (87.7%) | 147/167 (88.0%) | 70/86 (81.4%) |
| `pkg/cad/bom.go` | 125/130 (96.2%) | 334/339 (98.5%) | 34/40 (85.0%) |
| `pkg/cad/dxf.go` | 195/206 (94.7%) | 296/307 (96.4%) | 62/72 (86.1%) |
| `pkg/cad/exporter.go` | 131/133 (98.5%) | 173/175 (98.9%) | 91/92 (98.9%) |
| `pkg/cad/stl.go` | 299/308 (97.1%) | 339/348 (97.4%) | 72/84 (85.7%) |
| `pkg/model/types.go` | 128/128 (100.0%) | 183/183 (100.0%) | 102/104 (98.1%) |
| `pkg/model/units.go` | 40/42 (95.2%) | 40/42 (95.2%) | 0/0 (100.0%) |
| `pkg/scrubber/blower.go` | 74/77 (96.1%) | 115/118 (97.5%) | 24/28 (85.7%) |
| `pkg/scrubber/packedbed.go` | 106/119 (89.1%) | 141/154 (91.6%) | 32/42 (76.2%) |
| `pkg/scrubber/scrubber.go` | 68/78 (87.2%) | 102/112 (91.1%) | 31/41 (75.6%) |
| `pkg/scrubber/venturi.go` | 138/153 (90.2%) | 176/191 (92.1%) | 55/70 (78.6%) |
| `pkg/studio/cots.go` | 410/419 (97.9%) | 887/896 (99.0%) | 167/177 (94.4%) |
| `pkg/studio/frugal.go` | 498/513 (97.1%) | 1010/1025 (98.5%) | 198/222 (89.2%) |
| `pkg/studio/guide.go` | 394/404 (97.5%) | 499/509 (98.0%) | 134/140 (95.7%) |
| `pkg/studio/options.go` | 547/595 (91.9%) | 921/969 (95.0%) | 210/263 (79.8%) |
| `pkg/studio/packet.go` | 165/184 (89.7%) | 264/283 (93.3%) | 56/73 (76.7%) |
| `pkg/studio/plan.go` | 554/565 (98.1%) | 1585/1596 (99.3%) | 230/241 (95.4%) |
| `pkg/studio/presets.go` | 234/245 (95.5%) | 417/429 (97.2%) | 144/157 (91.7%) |
| `pkg/studio/recirc.go` | 75/81 (92.6%) | 126/132 (95.5%) | 33/40 (82.5%) |
| `pkg/studio/svg.go` | 82/87 (94.3%) | 115/120 (95.8%) | 29/34 (85.3%) |
| `pkg/studio/urban.go` | 409/419 (97.6%) | 583/593 (98.3%) | 182/206 (88.3%) |
| `pkg/thermo/quench.go` | 148/160 (92.5%) | 174/186 (93.5%) | 58/72 (80.6%) |
| `pkg/thermo/shomate.go` | 182/188 (96.8%) | 194/200 (97.0%) | 89/96 (92.7%) |

## Branch Coverage by Construct Type

| Construct Kind | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| `for-body` | 32 | 34 | 94.1% |
| `for-exit` | 31 | 32 | 96.9% |
| `if-else` | 120 | 124 | 96.8% |
| `if-else-if` | 65 | 67 | 97.0% |
| `if-implicit-else` | 1166 | 1259 | 92.6% |
| `if-then` | 1073 | 1450 | 74.0% |
| `range-body` | 61 | 61 | 100.0% |
| `range-exit` | 56 | 61 | 91.8% |
| `select-case` | 9 | 14 | 64.3% |
| `select-default` | 1 | 3 | 33.3% |
| `switch-case` | 203 | 223 | 91.0% |
| `switch-default` | 62 | 67 | 92.5% |
| `switch-implicit-default` | 8 | 11 | 72.7% |
| `type-switch-case` | 2 | 2 | 100.0% |
| `type-switch-implicit-default` | 1 | 1 | 100.0% |

## Highest-Value Uncovered Branches (Prioritized Test Targets)

Packages below the **85% executable line** or **80% AST branch** target are listed first, followed by remaining packages. Within each package, explicit `switch`/`select` arms, `else`/`else if` paths, and loop entry/exit branches are prioritized ahead of defensive error guards.

### `internal/control` (Line 85.3%, Branch 74.0%, 148 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `internal/control/hub.go:90` | `select-default` | `(*TelemetryHub).Publish` | `default` |
| `internal/control/hub.go:94` | `select-default` | `(*TelemetryHub).Publish` | `default` |
| `internal/control/modbus_rtu.go:325` | `select-case` | `(*ModbusRTUClient).roundTrip` | `case <-ctx.Done()` |
| `internal/control/modbus_rtu.go:328` | `select-case` | `(*ModbusRTUClient).roundTrip` | `case <-timer.C` |
| `internal/control/modbus_rtu.go:353` | `select-case` | `(*ModbusRTUClient).roundTrip.func@L351` | `case <-ctx.Done()` |
| `internal/control/modbus_rtu.go:355` | `select-case` | `(*ModbusRTUClient).roundTrip.func@L351` | `case <-doneCh` |
| `internal/control/modbus_rtu.go:436` | `switch-default` | `(*ModbusRTUClient).roundTrip` | `default` |
| `internal/control/serial_linux.go:83` | `switch-case` | `linuxBaudFlag` | `case 1200` |
| `internal/control/serial_linux.go:85` | `switch-case` | `linuxBaudFlag` | `case 2400` |
| `internal/control/serial_linux.go:87` | `switch-case` | `linuxBaudFlag` | `case 4800` |
| `internal/control/serial_linux.go:89` | `switch-case` | `linuxBaudFlag` | `case 9600` |
| `internal/control/serial_linux.go:91` | `switch-case` | `linuxBaudFlag` | `case 19200` |

*...and 136 additional uncovered branches in `internal/control` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `internal/web` (Line 86.6%, Branch 76.5%, 61 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `internal/web/server.go:179` | `select-case` | `(*Server).ListenAndServe` | `case err := <-errCh` |
| `internal/web/server.go:245` | `if-else` | `(*Server).handleIndex` | `else of (strings.HasPrefix(cleanRel, "_next/static/"))` |
| `internal/web/server.go:55` | `if-then` | `NewServer` | `err != nil` |
| `internal/web/server.go:60` | `if-then` | `NewServer` | `err != nil` |
| `internal/web/server.go:133` | `if-then` | `(*Server).SetSSEKeepaliveInterval` | `d <= 0` |
| `internal/web/server.go:145` | `if-then` | `(*Server).getSSEKeepaliveInterval` | `d <= 0` |
| `internal/web/server.go:160` | `if-then` | `(*Server).ListenAndServe` | `err != nil` |
| `internal/web/server.go:180` | `if-then` | `(*Server).ListenAndServe` | `errors.Is(err, http.ErrServerClosed)` |
| `internal/web/server.go:189` | `if-then` | `(*Server).ListenAndServe` | `shutdownErr != nil` |
| `internal/web/server.go:192` | `if-then` | `(*Server).ListenAndServe` | `serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed)` |
| `internal/web/server.go:221` | `if-then` | `(*Server).handleIndex` | `cleanRel == "" \|\| cleanRel == "." \|\| strings.HasPrefix(cleanRel, "..")` |
| `internal/web/server.go:231` | `if-then` | `(*Server).handleIndex` | `err != nil` |

*...and 49 additional uncovered branches in `internal/web` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/scrubber` (Line 92.9%, Branch 78.5%, 39 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/scrubber/packedbed.go:109` | `for-exit` | `SizePackedBed` | `exit/skip (iter < 8)` |
| `pkg/scrubber/packedbed.go:143` | `for-body` | `SizePackedBed` | `step < 200 && (qGas/(math.Pi*colDiam*colDiam/4.0))/vFlood*100.0 > pac...` |
| `pkg/scrubber/blower.go:128` | `if-then` | `SizeBlower` | `pAmb <= 0` |
| `pkg/scrubber/blower.go:167` | `if-then` | `SizeBlower` | `kSys < 0` |
| `pkg/scrubber/blower.go:188` | `if-then` | `SizeBlower` | `pFanVal < 0` |
| `pkg/scrubber/packedbed.go:81` | `if-then` | `EckertFloodingVelocity` | `vFloodSq <= 0 \|\| math.IsNaN(vFloodSq) \|\| math.IsInf(vFloodSq, 0)` |
| `pkg/scrubber/packedbed.go:112` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:139` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:148` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:190` | `if-then` | `SizePackedBed` | `input.PackingVoidage <= 0.78` |
| `pkg/scrubber/packedbed.go:196` | `if-then` | `SizePackedBed` | `err != nil` |
| `pkg/scrubber/packedbed.go:268` | `if-then` | `SizePackedBed` | `err != nil` |

*...and 27 additional uncovered branches in `pkg/scrubber` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `cmd/kilnscrub` (Line 93.0%, Branch 82.8%, 49 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `cmd/kilnscrub/main.go:126` | `if-else-if` | `(sizingFlags).applyExplicit` | `else of (input.SaltRateLbsPerHr > 0)` |
| `cmd/kilnscrub/main.go:129` | `if-else-if` | `(sizingFlags).applyExplicit` | `else of (input.FuelType == model.FuelGasSalt)` |
| `cmd/kilnscrub/main.go:29` | `if-then` | `main` | `err != nil` |
| `cmd/kilnscrub/main.go:121` | `if-then` | `(sizingFlags).applyExplicit` | `input.FuelType == model.FuelGas` |
| `cmd/kilnscrub/main.go:123` | `if-then` | `(sizingFlags).applyExplicit` | `input.FuelType == model.FuelWood` |
| `cmd/kilnscrub/main.go:126` | `if-then` | `(sizingFlags).applyExplicit` | `input.SaltRateLbsPerHr == 0` |
| `cmd/kilnscrub/main.go:127` | `if-then` | `(sizingFlags).applyExplicit` | `input.FuelType == model.FuelGasSalt` |
| `cmd/kilnscrub/main.go:129` | `if-then` | `(sizingFlags).applyExplicit` | `input.FuelType == model.FuelWoodSalt` |
| `cmd/kilnscrub/main.go:214` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:220` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:339` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:345` | `if-then` | `runExport` | `err != nil` |

*...and 37 additional uncovered branches in `cmd/kilnscrub` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

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

### `pkg/studio` (Line 97.8%, Branch 89.1%, 170 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/studio/guide.go:115` | `switch-case` | `GenerateBuildGuide` | `case BuildPathLocalCeramic` |
| `pkg/studio/options.go:841` | `switch-case` | `eductorEntrainmentRatio` | `case dpInWC <= 4.0` |
| `pkg/studio/plan.go:266` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 6.5` |
| `pkg/studio/plan.go:268` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 8.5` |
| `pkg/studio/plan.go:272` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 12.5` |
| `pkg/studio/plan.go:276` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 16.5` |
| `pkg/studio/plan.go:278` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 18.5` |
| `pkg/studio/plan.go:280` | `switch-case` | `ansi150FlangeOD` | `case nomIn <= 20.5` |
| `pkg/studio/plan.go:284` | `switch-default` | `ansi150FlangeOD` | `default` |
| `pkg/studio/presets.go:440` | `switch-default` | `(StudioProject).EffectiveLocalDraftMode` | `default` |
| `pkg/studio/recirc.go:42` | `switch-case` | `workingSumpVolumeGal` | `case "USPLASTIC-100T"` |
| `pkg/studio/recirc.go:52` | `switch-default` | `workingSumpVolumeGal` | `default` |

*...and 158 additional uncovered branches in `pkg/studio` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/cad` (Line 97.7%, Branch 89.9%, 29 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/cad/bom.go:110` | `if-then` | `(*Exporter).GenerateBOM` | `packingVolM3 < 0.001` |
| `pkg/cad/bom.go:372` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:388` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:404` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/bom.go:409` | `if-then` | `(*Exporter).FormatBOMCSV` | `err != nil` |
| `pkg/cad/dxf.go:44` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:47` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:50` | `if-then` | `ComputeFrustumDevelopment` | `err != nil` |
| `pkg/cad/dxf.go:108` | `if-then` | `(*dxfBuilder).trackPoint` | `!isFinite(x) \|\| !isFinite(y)` |
| `pkg/cad/dxf.go:225` | `if-then` | `(*dxfBuilder).addFrustumPattern` | `dev.IsCylinder` |
| `pkg/cad/dxf.go:303` | `if-then` | `(*dxfBuilder).build` | `!b.hasPts` |
| `pkg/cad/dxf.go:452` | `if-then` | `(*Exporter).ExportDXF` | `err != nil` |

*...and 17 additional uncovered branches in `pkg/cad` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/model` (Line 99.1%, Branch 98.1%, 2 uncovered branches)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/model/types.go:615` | `if-implicit-else` | `(*ControlConfig).Sync` | `!(c.DraftFilterTimeConst > 0)` |
| `pkg/model/types.go:640` | `if-implicit-else` | `(ControlConfig).Validate` | `!(expectedTau > 0)` |

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
go run ./scripts/branch_coverage.go -profile coverage.out -min-line 85 -min-branch 72
```
