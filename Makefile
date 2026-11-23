# Makefile for kilnscrub (Pure-Go Kiln Exhaust Scrubber Designer & Active Controller)

SHELL := /usr/bin/env bash
export PATH := $(PATH):/usr/local/go/bin

GO ?= go
BINARY := bin/kilnscrub
DIST_DIR := dist
PKG := ./cmd/kilnscrub
LDFLAGS := -s -w

.PHONY: help build test vet fmt fmt-check smoke release-all clean

help: ## Show available make targets
	@echo "kilnscrub Makefile targets:"
	@echo "  build        Build static pure-Go binary ($(BINARY))"
	@echo "  test         Run full test suite with race detector (CGO_ENABLED=1)"
	@echo "  vet          Run go vet across all packages"
	@echo "  fmt          Format all Go files with gofmt -w"
	@echo "  fmt-check    Verify all Go files are formatted with gofmt"
	@echo "  smoke        Build and run CLI design & export smoke checks"
	@echo "  release-all  Cross-compile static binaries to $(DIST_DIR)/"
	@echo "  clean        Remove $(BINARY), bin/, and $(DIST_DIR)/"

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

smoke: build ## Run end-to-end CLI design and export smoke verification
	@set -euo pipefail; \
	tmp_dir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	./$(BINARY) design --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0; \
	./$(BINARY) export --kiln-volume 24 --peak-temp 2400 --salt-rate 2.0 --out "$$tmp_dir"; \
	for f in scrubber.stl scrubber.dxf bom.csv bom.md; do \
		test -s "$$tmp_dir/$$f" || { echo "Missing or empty export: $$f"; exit 1; }; \
	done; \
	grep -q "EOF" "$$tmp_dir/scrubber.dxf"; \
	echo "Smoke check passed: .stl, .dxf, bom.csv, bom.md verified."

release-all: ## Cross-compile static binaries to dist/
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-linux-amd64       $(PKG)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-linux-arm64       $(PKG)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-darwin-arm64      $(PKG)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/kilnscrub-windows-amd64.exe $(PKG)
	@ls -lh $(DIST_DIR)

clean: ## Remove build and release artifacts
	rm -rf bin $(DIST_DIR)
