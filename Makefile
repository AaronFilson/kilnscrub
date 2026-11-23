# Makefile for kilnscrub (Pure-Go Kiln Exhaust Scrubber Designer & Active Controller)

SHELL := /usr/bin/env bash
export PATH := $(PATH):/usr/local/go/bin

GO ?= go
BINARY := bin/kilnscrub
DIST_DIR := dist
PKG := ./cmd/kilnscrub
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: help build test vet fmt fmt-check check coverage coverage-branch smoke packet-smoke release-all release-archives clean

help: ## Show available make targets
	@echo "kilnscrub Makefile targets:"
	@echo "  build             Build static pure-Go binary ($(BINARY))"
	@echo "  test              Run full test suite with race detector (CGO_ENABLED=1)"
	@echo "  vet               Run go vet across all packages"
	@echo "  fmt               Format all Go files with gofmt -w"
	@echo "  fmt-check         Verify all Go files are formatted with gofmt"
	@echo "  check             Run fmt-check, vet, and race tests"
	@echo "  coverage          Run race tests with coverage profile and print total coverage"
	@echo "  coverage-branch   Run atomic race tests and enforce Line (>=85%) & AST Branch (>=72%) coverage"
	@echo "  smoke             Build and run CLI design, export, and daemon smoke checks"
	@echo "  packet-smoke      Build and verify CLI presets & Build Packet .zip export"
	@echo "  release-all       Cross-compile all 5 static binaries and SHA256SUMS.txt to $(DIST_DIR)/"
	@echo "  release-archives  Build packaged release archives (.tar.gz/.zip) + SHA256SUMS.txt"
	@echo "  clean             Remove $(BINARY), bin/, $(DIST_DIR)/, and coverage.out"

build: ## Build static pure-Go binary for current host
	@mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) $(PKG)

test: ## Run full test suite with race detector
	CGO_ENABLED=1 $(GO) test -race -count=1 ./...

vet: ## Run go vet across all packages
	$(GO) vet ./...

fmt: ## Format all Go source files with gofmt
	gofmt -w .

fmt-check: ## Check that all Go source files are gofmt-clean
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted Go files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

check: fmt-check vet test ## Run formatting check, go vet, and race test suite

coverage: ## Run race tests with coverage profile and print total coverage
	CGO_ENABLED=1 $(GO) test -race -count=1 -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out | tail -n 1

coverage-branch: ## Run atomic race tests and report Statement, Line & AST Branch coverage
	CGO_ENABLED=1 $(GO) test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	@$(GO) run ./scripts/branch_coverage.go -profile coverage.out -min-line 85 -min-branch 72

smoke: build ## Run end-to-end CLI design, export, and daemon smoke verification
	@set -euo pipefail; \
	tmp_dir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	./$(BINARY) design --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0; \
	./$(BINARY) export --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0 --out "$$tmp_dir"; \
	for f in scrubber.stl scrubber.dxf bom.csv bom.md; do \
		test -s "$$tmp_dir/$$f" || { echo "Missing or empty export: $$f"; exit 1; }; \
	done; \
	grep -q "EOF" "$$tmp_dir/scrubber.dxf"; \
	./$(BINARY) daemon --config configs/kilnscrub.example.json --duration 300ms --interval 50ms; \
	echo "Smoke check passed: design, export (.stl, .dxf, bom.csv, bom.md), and daemon verified."

packet-smoke: build ## Run CLI presets and Build Packet zip smoke verification
	@set -euo pipefail; \
	tmp_dir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	./$(BINARY) presets; \
	./$(BINARY) packet --preset wood_salt_24 --out "$$tmp_dir/packet.zip"; \
	test -s "$$tmp_dir/packet.zip" || { echo "Missing or empty packet zip"; exit 1; }; \
	if command -v unzip >/dev/null 2>&1; then \
		unzip -t "$$tmp_dir/packet.zip"; \
	else \
		python3 -m zipfile -t "$$tmp_dir/packet.zip"; \
	fi; \
	echo "Packet smoke check passed: presets and packet.zip verified."

release-all: ## Cross-compile static binaries for all 5 targets to dist/
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-linux-amd64       $(PKG)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-linux-arm64       $(PKG)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-darwin-amd64      $(PKG)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-darwin-arm64      $(PKG)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-windows-amd64.exe $(PKG)
	@(cd $(DIST_DIR) && sha256sum kilnscrub-linux-amd64 kilnscrub-linux-arm64 kilnscrub-darwin-amd64 kilnscrub-darwin-arm64 kilnscrub-windows-amd64.exe > SHA256SUMS.txt)
	@ls -lh $(DIST_DIR)

release-archives: ## Build packaged release archives (.tar.gz/.zip) + SHA256SUMS.txt
	./packaging/release.sh "$(VERSION)" "$(DIST_DIR)/release"

clean: ## Remove build, coverage, and release artifacts
	rm -rf bin $(DIST_DIR) coverage.out
