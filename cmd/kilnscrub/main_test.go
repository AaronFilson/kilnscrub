package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func TestRunDesignSummary(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"design",
		"--kiln-volume", "24",
		"--peak-temp", "2400",
		"--salt-rate", "2.0",
		"--fuel", "wood_salt",
		"--throat-vel", "70",
		"--lg", "0.0012",
		"--hcl-removal", "0.99",
		"--draft", "-0.05",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("design failed: %v (stderr: %s)", err, stderr.String())
	}

	out := stdout.String()
	requiredSubstrings := []string{
		"Stage 1 - Adiabatic Quench Tower",
		"Stage 2 - Wet Venturi Scrubber",
		"Stage 3 - Packed-Bed HCl Absorber",
		"Stage 4 - Induced-Draft Blower",
		"Skid Envelope",
		"ACFM",
		"inWC",
		"HP",
	}
	for _, sub := range requiredSubstrings {
		if !strings.Contains(out, sub) {
			t.Errorf("design summary missing %q;\nfull output:\n%s", sub, out)
		}
	}
}

func TestRunDesignJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"design",
		"--kiln-volume", "24",
		"--peak-temp", "2400",
		"--salt-rate", "2.0",
		"--json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("design --json failed: %v (stderr: %s)", err, stderr.String())
	}

	var design model.ScrubberDesign
	if err := json.Unmarshal(stdout.Bytes(), &design); err != nil {
		t.Fatalf("design --json output is not valid ScrubberDesign JSON: %v", err)
	}
	if design.Quench.AdiabaticSatTemp <= 300 || design.Quench.TotalSprayWaterGPM <= 0 {
		t.Errorf("unexpected quench result: Tsat=%v K, GPM=%v", design.Quench.AdiabaticSatTemp, design.Quench.TotalSprayWaterGPM)
	}
	if design.Venturi.CutDiameterD50Microns <= 0 || design.Venturi.PressureDropInWC <= 0 {
		t.Errorf("unexpected venturi result: d50=%v um, dP=%v inWC", design.Venturi.CutDiameterD50Microns, design.Venturi.PressureDropInWC)
	}
	if design.PackedBed.PercentOfFlood <= 0 || design.PackedBed.PercentOfFlood > 75.0 {
		t.Errorf("packed bed percent of flood out of (0, 75]: %v", design.PackedBed.PercentOfFlood)
	}
	if design.PackedBed.HClRemovalEfficiency < 0.98 {
		t.Errorf("HCl removal efficiency < 0.98: %v", design.PackedBed.HClRemovalEfficiency)
	}
	if design.Blower.FlowACFM <= 0 || design.Blower.TotalStaticPressureInWC <= 0 || design.Blower.RecommendedMotorHP <= 0 {
		t.Errorf("unexpected blower result: ACFM=%v, dP=%v inWC, motor=%v HP",
			design.Blower.FlowACFM, design.Blower.TotalStaticPressureInWC, design.Blower.RecommendedMotorHP)
	}
}

func TestRunExport(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "nested-export")
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"export",
		"--kiln-volume", "24",
		"--peak-temp", "2400",
		"--salt-rate", "2.0",
		"--out", outDir,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("export failed: %v (stderr: %s)", err, stderr.String())
	}

	// Verify STL binary format: 80-byte header + 4-byte uint32 count + 50 bytes/triangle.
	stlPath := filepath.Join(outDir, "scrubber.stl")
	stlBytes, err := os.ReadFile(stlPath)
	if err != nil {
		t.Fatalf("read scrubber.stl: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("scrubber.stl too small: %d bytes", len(stlBytes))
	}
	triCount := binary.LittleEndian.Uint32(stlBytes[80:84])
	expectedLen := 84 + int(triCount)*50
	if triCount == 0 || len(stlBytes) != expectedLen {
		t.Errorf("invalid binary STL: triangles=%d, len=%d, expected=%d", triCount, len(stlBytes), expectedLen)
	}

	// Verify DXF R12 structure.
	dxfPath := filepath.Join(outDir, "scrubber.dxf")
	dxfBytes, err := os.ReadFile(dxfPath)
	if err != nil {
		t.Fatalf("read scrubber.dxf: %v", err)
	}
	dxfStr := string(dxfBytes)
	if !strings.Contains(dxfStr, "ENTITIES") || !strings.Contains(dxfStr, "EOF") {
		t.Errorf("scrubber.dxf missing expected ENTITIES/EOF markers")
	}

	// Verify BOM CSV & Markdown.
	csvPath := filepath.Join(outDir, "bom.csv")
	csvBytes, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read bom.csv: %v", err)
	}
	if len(csvBytes) == 0 || !strings.Contains(string(csvBytes), "Category") {
		t.Errorf("bom.csv appears empty or malformed")
	}

	mdPath := filepath.Join(outDir, "bom.md")
	mdBytes, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read bom.md: %v", err)
	}
	if len(mdBytes) == 0 || !strings.Contains(string(mdBytes), "|") {
		t.Errorf("bom.md appears empty or missing markdown table")
	}

	out := stdout.String()
	for _, name := range []string{"scrubber.stl", "scrubber.dxf", "bom.csv", "bom.md"} {
		if !strings.Contains(out, name) {
			t.Errorf("export stdout missing %q;\noutput:\n%s", name, out)
		}
	}
}

func TestRunDaemonBoundedDuration(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	err := run(context.Background(), []string{
		"daemon",
		"--draft", "-0.05",
		"--quench-temp-f", "165",
		"--interval", "50ms",
		"--duration", "250ms",
	}, &stdout, &stderr)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("daemon failed: %v (stderr: %s)", err, stderr.String())
	}
	if elapsed > 3*time.Second {
		t.Errorf("daemon --duration 250ms took too long: %v", elapsed)
	}

	out := stdout.String()
	if !strings.Contains(out, "tick=1") {
		t.Errorf("daemon output missing telemetry ticks;\noutput:\n%s", out)
	}
	if !strings.Contains(out, "stopped cleanly") {
		t.Errorf("daemon output missing clean shutdown message;\noutput:\n%s", out)
	}
}

func TestRunUIStartsAndStopsCleanly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	var stdout, stderr bytes.Buffer
	err := run(ctx, []string{"ui", "--addr", "127.0.0.1:0"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("ui failed: %v (stderr: %s)", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kilnscrub ui listening on") {
		t.Errorf("ui output missing listening banner: %q", stdout.String())
	}
}

func TestRunErrorCases(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "unknown subcommand", args: []string{"bogus"}},
		{name: "design invalid volume", args: []string{"design", "--kiln-volume", "-5"}},
		{name: "design positive draft", args: []string{"design", "--draft", "0.10"}},
		{name: "design invalid fuel", args: []string{"design", "--fuel", "coal"}},
		{name: "design invalid hcl removal", args: []string{"design", "--hcl-removal", "1.5"}},
		{name: "design unexpected arg", args: []string{"design", "extra"}},
		{name: "export invalid peak temp", args: []string{"export", "--peak-temp", "50"}},
		{name: "export empty out dir", args: []string{"export", "--out", "   "}},
		{name: "daemon positive draft", args: []string{"daemon", "--draft", "0.2"}},
		{name: "daemon quench temp above interlock", args: []string{"daemon", "--quench-temp-f", "220"}},
		{name: "daemon negative duration", args: []string{"daemon", "--duration", "-1s"}},
		{name: "daemon zero interval", args: []string{"daemon", "--interval", "0s"}},
		{name: "ui empty addr", args: []string{"ui", "--addr", ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(context.Background(), tc.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil (stdout=%q)", tc.args, stdout.String())
			}
		})
	}
}

func TestRunHelpAndNoArgs(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("args %v returned unexpected error: %v", args, err)
		}
		if !strings.Contains(stdout.String(), "Usage:") {
			t.Errorf("args %v did not print usage: %q", args, stdout.String())
		}
	}
}
