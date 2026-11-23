# Test Coverage Report (Statement, Line & AST Branch Coverage)

Generated automatically by [`scripts/branch_coverage.go`](../scripts/branch_coverage.go) from `go test -race -covermode=atomic -coverprofile=coverage.out ./...`.

## Executive Summary

| Metric | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| **Statements** (matches `go tool cover -func`) | 3675 | 4389 | **83.7%** |
| **Executable Lines** (token-mapped) | 5164 | 5925 | **87.2%** |
| **AST Control-Flow Branches** (C1 decision) | 1345 | 1978 | **68.0%** |

## Per-Package Coverage Summary

| Package | Statements | Statement % | Exec Lines | Line % | AST Branches | Branch % | Status (≥85% Line / ≥80% Branch) |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: | :---: |
| `cmd/kilnscrub` | 378 / 436 | 86.7% | 462 / 520 | 88.8% | 160 / 219 | 73.1% | Below Branch |
| `internal/control` | 794 / 1030 | 77.1% | 989 / 1228 | 80.5% | 346 / 530 | 65.3% | Below Line & Branch |
| `internal/web` | 409 / 459 | 89.1% | 421 / 471 | 89.4% | 169 / 210 | 80.5% | PASS |
| `pkg/cad` | 672 / 761 | 88.3% | 1063 / 1152 | 92.3% | 192 / 280 | 68.6% | Below Branch |
| `pkg/model` | 145 / 170 | 85.3% | 200 / 225 | 88.9% | 79 / 104 | 76.0% | Below Branch |
| `pkg/scrubber` | 340 / 418 | 81.3% | 488 / 566 | 86.2% | 99 / 177 | 55.9% | Below Branch |
| `pkg/studio` | 652 / 767 | 85.0% | 1218 / 1377 | 88.5% | 197 / 290 | 67.9% | Below Branch |
| `pkg/thermo` | 285 / 348 | 81.9% | 323 / 386 | 83.7% | 103 / 168 | 61.3% | Below Line & Branch |
| **TOTAL** | **3675 / 4389** | **83.7%** | **5164 / 5925** | **87.2%** | **1345 / 1978** | **68.0%** | — |

## Per-File Breakdown

| File | Statements | Exec Lines | AST Branches |
| :--- | ---: | ---: | ---: |
| `cmd/kilnscrub/main.go` | 378/436 (86.7%) | 462/520 (88.8%) | 160/219 (73.1%) |
| `internal/control/config.go` | 37/48 (77.1%) | 41/52 (78.8%) | 29/40 (72.5%) |
| `internal/control/controller.go` | 96/109 (88.1%) | 150/163 (92.0%) | 31/44 (70.5%) |
| `internal/control/daemon.go` | 63/81 (77.8%) | 68/86 (79.1%) | 24/36 (66.7%) |
| `internal/control/filter.go` | 37/40 (92.5%) | 41/44 (93.2%) | 10/10 (100.0%) |
| `internal/control/hub.go` | 38/43 (88.4%) | 40/45 (88.9%) | 15/20 (75.0%) |
| `internal/control/interlock.go` | 52/59 (88.1%) | 68/75 (90.7%) | 25/29 (86.2%) |
| `internal/control/modbus_hal.go` | 118/158 (74.7%) | 182/222 (82.0%) | 68/101 (67.3%) |
| `internal/control/modbus_rtu.go` | 168/219 (76.7%) | 176/230 (76.5%) | 78/129 (60.5%) |
| `internal/control/serial_linux.go` | 0/48 (0.0%) | 0/48 (0.0%) | 0/26 (0.0%) |
| `internal/control/sim_hal.go` | 185/225 (82.2%) | 223/263 (84.8%) | 66/95 (69.5%) |
| `internal/web/server.go` | 193/220 (87.7%) | 200/227 (88.1%) | 68/92 (73.9%) |
| `internal/web/studio.go` | 82/88 (93.2%) | 82/88 (93.2%) | 36/40 (90.0%) |
| `internal/web/telemetry.go` | 134/151 (88.7%) | 139/156 (89.1%) | 65/78 (83.3%) |
| `pkg/cad/bom.go` | 114/129 (88.4%) | 322/337 (95.5%) | 24/40 (60.0%) |
| `pkg/cad/dxf.go` | 183/198 (92.4%) | 284/299 (95.0%) | 54/68 (79.4%) |
| `pkg/cad/exporter.go` | 95/133 (71.4%) | 137/175 (78.3%) | 60/92 (65.2%) |
| `pkg/cad/stl.go` | 280/301 (93.0%) | 320/341 (93.8%) | 54/80 (67.5%) |
| `pkg/model/types.go` | 105/128 (82.0%) | 160/183 (87.4%) | 79/104 (76.0%) |
| `pkg/model/units.go` | 40/42 (95.2%) | 40/42 (95.2%) | 0/0 (100.0%) |
| `pkg/scrubber/blower.go` | 62/71 (87.3%) | 103/112 (92.0%) | 15/26 (57.7%) |
| `pkg/scrubber/packedbed.go` | 92/116 (79.3%) | 127/151 (84.1%) | 19/40 (47.5%) |
| `pkg/scrubber/scrubber.go` | 62/78 (79.5%) | 96/112 (85.7%) | 24/41 (58.5%) |
| `pkg/scrubber/venturi.go` | 124/153 (81.0%) | 162/191 (84.8%) | 41/70 (58.6%) |
| `pkg/studio/cots.go` | 164/213 (77.0%) | 412/501 (82.2%) | 53/77 (68.8%) |
| `pkg/studio/guide.go` | 123/124 (99.2%) | 169/170 (99.4%) | 21/22 (95.5%) |
| `pkg/studio/packet.go` | 97/120 (80.8%) | 170/193 (88.1%) | 29/51 (56.9%) |
| `pkg/studio/presets.go` | 83/103 (80.6%) | 195/217 (89.9%) | 32/54 (59.3%) |
| `pkg/studio/urban.go` | 185/207 (89.4%) | 272/296 (91.9%) | 62/86 (72.1%) |
| `pkg/thermo/quench.go` | 133/160 (83.1%) | 159/186 (85.5%) | 44/72 (61.1%) |
| `pkg/thermo/shomate.go` | 152/188 (80.9%) | 164/200 (82.0%) | 59/96 (61.5%) |

## Branch Coverage by Construct Type

| Construct Kind | Covered | Total | Coverage |
| :--- | ---: | ---: | ---: |
| `for-body` | 25 | 28 | 89.3% |
| `for-exit` | 24 | 26 | 92.3% |
| `if-else` | 31 | 40 | 77.5% |
| `if-else-if` | 6 | 8 | 75.0% |
| `if-implicit-else` | 733 | 797 | 92.0% |
| `if-then` | 354 | 845 | 41.9% |
| `range-body` | 40 | 41 | 97.6% |
| `range-exit` | 35 | 41 | 85.4% |
| `select-case` | 9 | 10 | 90.0% |
| `select-default` | 1 | 3 | 33.3% |
| `switch-case` | 65 | 102 | 63.7% |
| `switch-default` | 17 | 28 | 60.7% |
| `switch-implicit-default` | 3 | 6 | 50.0% |
| `type-switch-case` | 2 | 2 | 100.0% |
| `type-switch-implicit-default` | 0 | 1 | 0.0% |

## Highest-Value Uncovered Branches (Prioritized Test Targets)

Packages below the **85% executable line** or **80% AST branch** target are listed first, followed by remaining packages. Within each package, explicit `switch`/`select` arms, `else`/`else if` paths, and loop entry/exit branches are prioritized ahead of defensive error guards.

### `pkg/scrubber` (Line 86.2%, Branch 55.9%, 78 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/scrubber/scrubber.go:85` | `switch-default` | `computeInletFlueGas` | `default` |
| `pkg/scrubber/blower.go:80` | `if-else` | `DuctPressureDrop` | `else of (mu > 0 && v > 0)` |
| `pkg/scrubber/blower.go:33` | `range-exit` | `selectStandardMotorHP` | `exit/skip (range standardMotorHPs)` |
| `pkg/scrubber/packedbed.go:109` | `for-exit` | `SizePackedBed` | `exit/skip (iter < 8)` |
| `pkg/scrubber/packedbed.go:143` | `for-body` | `SizePackedBed` | `(qGas/(math.Pi*colDiam*colDiam/4.0))/vFlood*100.0 > packedBedMaxFloodPct` |
| `pkg/scrubber/venturi.go:139` | `for-body` | `CalvertCutDiameter` | `residual(low) < 0 && low > 1e-12` |
| `pkg/scrubber/venturi.go:142` | `for-body` | `CalvertCutDiameter` | `residual(high) > 0 && high < 1.0` |
| `pkg/scrubber/blower.go:44` | `if-then` | `AutoSizeDuctDiameter` | `inletGas.VolumetricFlow <= 0` |
| `pkg/scrubber/blower.go:50` | `if-then` | `AutoSizeDuctDiameter` | `nom < ductMinDiamM` |
| `pkg/scrubber/blower.go:59` | `if-then` | `DuctPressureDrop` | `ductDiameter <= 0 \|\| gas.VolumetricFlow <= 0 \|\| gas.Density <= 0` |
| `pkg/scrubber/blower.go:72` | `if-then` | `DuctPressureDrop` | `re < 2300.0` |
| `pkg/scrubber/blower.go:100` | `if-then` | `SizeBlower` | `outletGas.VolumetricFlow <= 0 \|\| outletGas.Density <= 0` |

*...and 66 additional uncovered branches in `pkg/scrubber` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/thermo` (Line 83.7%, Branch 61.3%, 65 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/thermo/shomate.go:217` | `switch-case` | `LucasGasViscosity` | `case "7440-59-7"` |
| `pkg/thermo/shomate.go:219` | `switch-case` | `LucasGasViscosity` | `case "1333-74-0"` |
| `pkg/thermo/shomate.go:221` | `switch-case` | `LucasGasViscosity` | `case "7782-39-0"` |
| `pkg/thermo/quench.go:86` | `if-else` | `BrentRoot` | `else of (math.Abs(e) >= tol1 && math.Abs(fa) > math.Abs(fb))` |
| `pkg/thermo/quench.go:44` | `for-exit` | `BrentRoot` | `exit/skip (iter < maxIter)` |
| `pkg/thermo/shomate.go:148` | `range-exit` | `selectShomateRange` | `exit/skip (range ranges)` |
| `pkg/thermo/quench.go:14` | `if-then` | `BrentRoot` | `math.IsNaN(a) \|\| math.IsNaN(b) \|\| math.IsInf(a, 0) \|\| math.IsInf(b, 0)` |
| `pkg/thermo/quench.go:17` | `if-then` | `BrentRoot` | `tol <= 0` |
| `pkg/thermo/quench.go:20` | `if-then` | `BrentRoot` | `maxIter <= 0` |
| `pkg/thermo/quench.go:26` | `if-then` | `BrentRoot` | `math.IsNaN(fa) \|\| math.IsNaN(fb)` |
| `pkg/thermo/quench.go:29` | `if-then` | `BrentRoot` | `fa == 0` |
| `pkg/thermo/quench.go:32` | `if-then` | `BrentRoot` | `fb == 0` |

*...and 53 additional uncovered branches in `pkg/thermo` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `internal/control` (Line 80.5%, Branch 65.3%, 184 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `internal/control/hub.go:84` | `select-default` | `(*TelemetryHub).Publish` | `default` |
| `internal/control/hub.go:88` | `select-default` | `(*TelemetryHub).Publish` | `default` |
| `internal/control/modbus_hal.go:435` | `switch-case` | `convertModbusTempToKelvin` | `case ModbusTempUnitK` |
| `internal/control/modbus_rtu.go:62` | `switch-case` | `modbusExceptionName` | `case 0x01` |
| `internal/control/modbus_rtu.go:64` | `switch-case` | `modbusExceptionName` | `case 0x02` |
| `internal/control/modbus_rtu.go:66` | `switch-case` | `modbusExceptionName` | `case 0x03` |
| `internal/control/modbus_rtu.go:70` | `switch-case` | `modbusExceptionName` | `case 0x05` |
| `internal/control/modbus_rtu.go:72` | `switch-case` | `modbusExceptionName` | `case 0x06` |
| `internal/control/modbus_rtu.go:74` | `switch-case` | `modbusExceptionName` | `case 0x08` |
| `internal/control/modbus_rtu.go:76` | `switch-case` | `modbusExceptionName` | `case 0x0A` |
| `internal/control/modbus_rtu.go:78` | `switch-case` | `modbusExceptionName` | `case 0x0B` |
| `internal/control/modbus_rtu.go:80` | `switch-default` | `modbusExceptionName` | `default` |

*...and 172 additional uncovered branches in `internal/control` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/studio` (Line 88.5%, Branch 67.9%, 93 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/studio/cots.go:307` | `switch-default` | `selectSumpTank` | `default` |
| `pkg/studio/cots.go:411` | `switch-case` | `selectBlowerVFD` | `case hp <= 3.0` |
| `pkg/studio/cots.go:468` | `switch-case` | `RecommendedDilutionCFM` | `case target <= 2250.0` |
| `pkg/studio/cots.go:479` | `switch-case` | `selectPlumeDiluter` | `case cfm <= 500.0` |
| `pkg/studio/cots.go:481` | `switch-case` | `selectPlumeDiluter` | `case cfm <= 1000.0` |
| `pkg/studio/cots.go:485` | `switch-case` | `selectPlumeDiluter` | `case cfm <= 2000.0` |
| `pkg/studio/cots.go:487` | `switch-default` | `selectPlumeDiluter` | `default` |
| `pkg/studio/packet.go:91` | `switch-case` | `PacketFilename` | `case r == ' '` |
| `pkg/studio/presets.go:261` | `switch-case` | `(StudioProject).EffectiveDesignInput` | `case "", model.FuelGasSalt` |
| `pkg/studio/presets.go:268` | `switch-case` | `(StudioProject).EffectiveDesignInput` | `case ""` |
| `pkg/studio/presets.go:270` | `switch-case` | `(StudioProject).EffectiveDesignInput` | `case model.FuelWood` |
| `pkg/studio/presets.go:272` | `switch-case` | `(StudioProject).EffectiveDesignInput` | `case model.FuelGas` |

*...and 81 additional uncovered branches in `pkg/studio` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/cad` (Line 92.3%, Branch 68.6%, 88 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/cad/exporter.go:303` | `if-else` | `deriveGeometry` | `else of (sumpVol > 0)` |
| `pkg/cad/stl.go:127` | `if-else-if` | `(*stlMesh).addFrustum` | `else of (r0 > 1e-6 && r1 > 1e-6)` |
| `pkg/cad/stl.go:129` | `if-else` | `(*stlMesh).addFrustum` | `else of (r0 > 1e-6)` |
| `pkg/cad/bom.go:47` | `if-then` | `(*Exporter).GenerateBOM` | `quenchMat == ""` |
| `pkg/cad/bom.go:51` | `if-then` | `(*Exporter).GenerateBOM` | `venturiMat == ""` |
| `pkg/cad/bom.go:55` | `if-then` | `(*Exporter).GenerateBOM` | `colMat == ""` |
| `pkg/cad/bom.go:59` | `if-then` | `(*Exporter).GenerateBOM` | `packingSpec == ""` |
| `pkg/cad/bom.go:62` | `if-then` | `(*Exporter).GenerateBOM` | `packingSpec == ""` |
| `pkg/cad/bom.go:66` | `if-then` | `(*Exporter).GenerateBOM` | `blowerMat == ""` |
| `pkg/cad/bom.go:108` | `if-then` | `(*Exporter).GenerateBOM` | `packingVolM3 < 0.001` |
| `pkg/cad/bom.go:121` | `if-then` | `(*Exporter).GenerateBOM` | `recircGPM <= 0` |
| `pkg/cad/bom.go:129` | `if-then` | `(*Exporter).GenerateBOM` | `motorHP <= 0` |

*...and 76 additional uncovered branches in `pkg/cad` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `cmd/kilnscrub` (Line 88.8%, Branch 73.1%, 59 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `cmd/kilnscrub/main.go:647` | `switch-default` | `openControlHAL` | `default` |
| `cmd/kilnscrub/main.go:276` | `if-else` | `printDesignSummary` | `else of (len(d.Warnings) == 0)` |
| `cmd/kilnscrub/main.go:278` | `range-body` | `printDesignSummary` | `range d.Warnings` |
| `cmd/kilnscrub/main.go:278` | `range-exit` | `printDesignSummary` | `exit/skip (range d.Warnings)` |
| `cmd/kilnscrub/main.go:29` | `if-then` | `main` | `err != nil` |
| `cmd/kilnscrub/main.go:172` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:199` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:205` | `if-then` | `runDesign` | `err != nil` |
| `cmd/kilnscrub/main.go:294` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:297` | `if-then` | `runExport` | `fs.NArg() > 0` |
| `cmd/kilnscrub/main.go:324` | `if-then` | `runExport` | `err != nil` |
| `cmd/kilnscrub/main.go:330` | `if-then` | `runExport` | `err != nil` |

*...and 47 additional uncovered branches in `cmd/kilnscrub` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

### `pkg/model` (Line 88.9%, Branch 76.0%, 25 uncovered branches — Priority Target)

| Location | Construct | Function | Condition / Arm |
| :--- | :--- | :--- | :--- |
| `pkg/model/types.go:609` | `if-else` | `(*ControlConfig).Sync` | `else of (c.TargetDraftInWC != 0 \|\| c.TargetDraftPa == 0)` |
| `pkg/model/types.go:614` | `if-else-if` | `(*ControlConfig).Sync` | `else of (c.DraftFilterCutoffHz > 0)` |
| `pkg/model/types.go:96` | `if-then` | `(DesignInput).Validate` | `peakK <= d.AmbientTempK \|\| peakK > 2500.0` |
| `pkg/model/types.go:467` | `if-then` | `(SafetyLimits).Validate` | `!isFinite(float64(s.MaxQuenchOutletTemp)) \|\| s.MaxQuenchOutletTemp <=...` |
| `pkg/model/types.go:470` | `if-then` | `(SafetyLimits).Validate` | `!isFinite(float64(s.MinWaterFlow)) \|\| s.MinWaterFlow < 0` |
| `pkg/model/types.go:473` | `if-then` | `(SafetyLimits).Validate` | `!isFinite(s.MinSumpPH) \|\| !isFinite(s.MaxSumpPH) \|\| s.MinSumpPH < 0 \|...` |
| `pkg/model/types.go:476` | `if-then` | `(SafetyLimits).Validate` | `s.SensorFaultTimeout <= 0` |
| `pkg/model/types.go:479` | `if-then` | `(SafetyLimits).Validate` | `!isFinite(float64(s.MaxDraftMagnitudePa)) \|\| s.MaxDraftMagnitudePa <= 0` |
| `pkg/model/types.go:505` | `if-then` | `(PIDGains).Validate` | `!isFinite(g.Kp) \|\| g.Kp < 0` |
| `pkg/model/types.go:508` | `if-then` | `(PIDGains).Validate` | `!isFinite(g.Ki) \|\| g.Ki < 0` |
| `pkg/model/types.go:511` | `if-then` | `(PIDGains).Validate` | `!isFinite(g.Kd) \|\| g.Kd < 0` |
| `pkg/model/types.go:514` | `if-then` | `(PIDGains).Validate` | `!isFinite(g.AntiWindUpGain) \|\| g.AntiWindUpGain < 0` |

*...and 13 additional uncovered branches in `pkg/model` (run `go run ./scripts/branch_coverage.go -profile coverage.out -top 100` for the complete list).*

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
