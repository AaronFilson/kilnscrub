package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AaronFilson/kilnscrub/internal/control"
)

func TestCLIBranchCoverageAndErrorPaths(t *testing.T) {
	// 1. runDesign with warnings (triggers printDesignSummary else/range warnings).
	var out, errBuf bytes.Buffer
	if err := run(context.Background(), []string{"design", "--kiln-volume", "3", "--throat-vel", "40"}, &out, &errBuf); err != nil {
		t.Fatalf("runDesign warnings: %v", err)
	}
	if !strings.Contains(out.String(), "Warnings (") {
		t.Errorf("expected Warnings section in output:\n%s", out.String())
	}

	// 2. openControlHAL: unknown HAL mode & nonexistent modbus serial device.
	dcfg := control.DefaultDaemonConfig()
	dcfg.HALMode = "bluetooth"
	if _, _, err := openControlHAL(dcfg); err == nil {
		t.Error("expected unknown hal mode error")
	}
	dcfg.HALMode = control.HALModeModbus
	dcfg.Modbus.Device = filepath.Join(t.TempDir(), "nonexistent-ttyUSB99")
	if _, _, err := openControlHAL(dcfg); err == nil {
		t.Error("expected modbus nonexistent serial device error")
	}

	// 3. CLI error branches across subcommands.
	errCases := [][]string{
		{"design", "extra-pos"},
		{"design", "--unknown-flag"},
		{"design", "--preset", "no_such_preset"},
		{"design", "--kiln-volume", "-5"},
		{"export", "extra-pos"},
		{"export", "--unknown-flag"},
		{"export", "--out", "   "},
		{"export", "--preset", "no_such_preset"},
		{"presets", "extra-pos"},
		{"presets", "--unknown-flag"},
		{"guide", "extra-pos"},
		{"guide", "--unknown-flag"},
		{"guide", "--preset", "no_such_preset"},
		{"packet", "extra-pos"},
		{"packet", "--unknown-flag"},
		{"packet", "--preset", "no_such_preset"},
		{"daemon", "extra-pos"},
		{"daemon", "--unknown-flag"},
		{"daemon", "--duration", "-1s"},
		{"daemon", "--hal", "invalid"},
		{"daemon", "--config", filepath.Join(t.TempDir(), "missing.json")},
		{"ui", "extra-pos"},
		{"ui", "--unknown-flag"},
	}
	for _, args := range errCases {
		out.Reset()
		errBuf.Reset()
		if err := run(context.Background(), args, &out, &errBuf); err == nil {
			t.Errorf("run(%v): expected error, got nil", args)
		}
	}
}
