## Summary

<!-- Briefly describe what this PR changes and why. -->

## Verification Checklist

- [ ] **Go Quality & Race Tests**: `make check` passes (`gofmt -l .`, `go vet ./...`, and `CGO_ENABLED=1 go test -race -count=1 ./...`), `go.mod`/`go.sum` are tidy, and zero Python runtime dependencies exist in `*.go`.
- [ ] **Physics & Oracle Verification**: Thermochemical/hydraulic changes preserve SI-internal units, pass `testdata/oracle_fixtures.json` parity tests, and uphold CPVC 210°F thermal safety interlock margins (or document updated oracle rationale).
- [ ] **COTS & Potter UX Impact**: Studio presets, COTS part snapping, urban site readiness checks, Build Packet `.zip` exports, and Web UI workflows remain intuitive, accurate, and backward-compatible.
- [ ] **Documentation & Packaging**: Updated relevant docs, examples, or packaging files (`README.md`, `docs/`, `configs/kilnscrub.example.json`, `packaging/`) to reflect any CLI, API, or engineering changes.

## Test Evidence

<!-- Paste key command outputs (e.g., go test -race, smoke checks, UI screenshots) if applicable. -->
