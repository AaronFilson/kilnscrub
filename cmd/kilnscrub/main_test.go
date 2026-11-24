package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/studio"
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
		"--quench-temp-f", "188",
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

func TestRunDaemonWithConfigFile(t *testing.T) {
	cfgPath := filepath.Join("..", "..", "configs", "kilnscrub.example.json")
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"daemon",
		"--config", cfgPath,
		"--duration", "200ms",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("daemon --config failed: %v (stderr: %s)", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "hal=sim") || !strings.Contains(out, "quench_sp=188.0 F") {
		t.Errorf("daemon --config output missing expected hal/setpoint:\n%s", out)
	}
	if !strings.Contains(out, "tick=1") || !strings.Contains(out, "stopped cleanly") {
		t.Errorf("daemon --config output missing ticks or clean stop:\n%s", out)
	}
}

func TestRunDaemonModbusUnavailableDeviceReturnsCleanError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"daemon",
		"--hal", "modbus",
		"--duration", "100ms",
	}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error when modbus serial device is unavailable, got nil")
	}
	if !strings.Contains(err.Error(), "modbus") {
		t.Errorf("expected error mentioning modbus, got: %v", err)
	}
}

func reserveEphemeralAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen ephemeral: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func TestRunUILiveControlWiringAndNoControl(t *testing.T) {
	t.Run("default_live_sim_control_and_sse", func(t *testing.T) {
		addr := reserveEphemeralAddr(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var stdout, stderr bytes.Buffer
		errCh := make(chan error, 1)
		go func() {
			errCh <- run(ctx, []string{"ui", "--addr", addr}, &stdout, &stderr)
		}()

		client := &http.Client{Timeout: 2 * time.Second}
		statusURL := "http://" + addr + "/api/control/status"
		var statusBody struct {
			Enabled                bool                `json:"enabled"`
			HALMode                string              `json:"hal_mode"`
			SupportsFaultInjection bool                `json:"supports_fault_injection"`
			Config                 model.ControlConfig `json:"config"`
		}

		deadline := time.Now().Add(3 * time.Second)
		ready := false
		for time.Now().Before(deadline) {
			resp, err := client.Get(statusURL)
			if err == nil {
				err = json.NewDecoder(resp.Body).Decode(&statusBody)
				resp.Body.Close()
				if err == nil && resp.StatusCode == http.StatusOK {
					ready = true
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !ready {
			cancel()
			t.Fatalf("ui server did not become ready on %s (stderr=%q)", addr, stderr.String())
		}
		if !statusBody.Enabled || statusBody.HALMode != "sim" || !statusBody.SupportsFaultInjection {
			t.Fatalf("unexpected live control status: %+v", statusBody)
		}
		if gotF := statusBody.Config.QuenchSetpoint.Fahrenheit(); gotF < 187.9 || gotF > 188.1 {
			t.Errorf("live UI control QuenchSetpoint = %.2f F, want 188.0 F", gotF)
		}

		// Verify SSE /api/control/stream emits a valid JSON telemetry frame.
		sseCtx, sseCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer sseCancel()
		req, _ := http.NewRequestWithContext(sseCtx, http.MethodGet, "http://"+addr+"/api/control/stream", nil)
		sseResp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			t.Fatalf("GET /api/control/stream failed: %v", err)
		}
		scanner := bufio.NewScanner(sseResp.Body)
		gotEvent := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				var st model.ControlStatus
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &st); err != nil {
					t.Errorf("invalid SSE JSON payload: %v", err)
				}
				gotEvent = true
				break
			}
		}
		sseResp.Body.Close()
		if !gotEvent {
			t.Errorf("did not receive SSE data event from /api/control/stream")
		}

		cancel()
		select {
		case runErr := <-errCh:
			if runErr != nil {
				t.Fatalf("ui returned unexpected error on shutdown: %v", runErr)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("ui did not shut down cleanly after context cancel")
		}
	})

	t.Run("no_control_flag_disables_backend", func(t *testing.T) {
		addr := reserveEphemeralAddr(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var stdout, stderr bytes.Buffer
		errCh := make(chan error, 1)
		go func() {
			errCh <- run(ctx, []string{"ui", "--addr", addr, "--no-control"}, &stdout, &stderr)
		}()

		client := &http.Client{Timeout: 2 * time.Second}
		var statusBody struct {
			Enabled bool `json:"enabled"`
		}
		deadline := time.Now().Add(3 * time.Second)
		ready := false
		for time.Now().Before(deadline) {
			resp, err := client.Get("http://" + addr + "/api/control/status")
			if err == nil {
				err = json.NewDecoder(resp.Body).Decode(&statusBody)
				resp.Body.Close()
				if err == nil && resp.StatusCode == http.StatusOK {
					ready = true
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !ready {
			cancel()
			t.Fatalf("ui --no-control did not become ready")
		}
		if statusBody.Enabled {
			t.Errorf("expected enabled=false with --no-control, got true")
		}

		cancel()
		if runErr := <-errCh; runErr != nil {
			t.Fatalf("ui --no-control shutdown error: %v", runErr)
		}
	})
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
		{name: "presets unexpected arg", args: []string{"presets", "extra"}},
		{name: "guide unknown preset", args: []string{"guide", "--preset", "nonexistent"}},
		{name: "guide both preset and project", args: []string{"guide", "--preset", "wood_salt_24", "--project", "p.json"}},
		{name: "guide empty out dir", args: []string{"guide", "--out", "   "}},
		{name: "guide invalid alignment", args: []string{"guide", "--alignment", "bogus"}},
		{name: "guide invalid bed layout", args: []string{"guide", "--bed-layout", "bogus"}},
		{name: "guide invalid venturi layout", args: []string{"guide", "--venturi-layout", "bogus"}},
		{name: "guide invalid controls", args: []string{"guide", "--controls", "bogus"}},
		{name: "guide negative salting hours", args: []string{"guide", "--salting-hours", "-1"}},
		{name: "guide invalid blower drive", args: []string{"guide", "--blower-drive", "bogus"}},
		{name: "guide invalid parallel pipe size", args: []string{"guide", "--parallel-pipe-in", "6"}},
		{name: "guide invalid packing media", args: []string{"guide", "--packing-media", "bogus"}},
		{name: "guide invalid compressor scfm", args: []string{"guide", "--compressor-scfm", "-5"}},
		{name: "guide invalid build path", args: []string{"guide", "--build-path", "bogus"}},
		{name: "guide empty build path", args: []string{"guide", "--build-path", ""}},
		{name: "guide invalid local region", args: []string{"guide", "--build-path", "local_ceramic_masonry", "--local-region", "bogus"}},
		{name: "guide invalid local alkali", args: []string{"guide", "--build-path", "local_ceramic_masonry", "--local-alkali", "bogus"}},
		{name: "packet unknown preset", args: []string{"packet", "--preset", "nonexistent"}},
		{name: "packet empty out path", args: []string{"packet", "--out", "   "}},
		{name: "packet invalid alignment", args: []string{"packet", "--alignment", "bogus"}},
		{name: "packet invalid controls", args: []string{"packet", "--controls", "bogus"}},
		{name: "packet invalid blower drive", args: []string{"packet", "--blower-drive", "bogus"}},
		{name: "packet invalid parallel pipe size", args: []string{"packet", "--parallel-pipe-in", "5"}},
		{name: "packet invalid packing media", args: []string{"packet", "--packing-media", "bogus"}},
		{name: "packet invalid compressor scfm", args: []string{"packet", "--compressor-scfm", "250"}},
		{name: "packet invalid build path", args: []string{"packet", "--build-path", "bogus"}},
		{name: "packet invalid local region", args: []string{"packet", "--build-path", "local_ceramic_masonry", "--local-region", "bogus"}},
		{name: "packet invalid local alkali", args: []string{"packet", "--build-path", "local_ceramic_masonry", "--local-alkali", "bogus"}},
		{name: "design unknown preset", args: []string{"design", "--preset", "nonexistent"}},
		{name: "export unknown preset", args: []string{"export", "--preset", "nonexistent"}},
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
		out := stdout.String()
		if !strings.Contains(out, "Usage:") {
			t.Errorf("args %v did not print usage: %q", args, out)
		}
		for _, cmd := range []string{"presets", "guide", "packet", "version"} {
			if !strings.Contains(out, cmd) {
				t.Errorf("usage missing subcommand %q:\n%s", cmd, out)
			}
		}
	}
}

func TestRunVersion(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}, {"-version"}} {
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), args, &stdout, &stderr); err != nil {
			t.Fatalf("args %v failed: %v", args, err)
		}
		if !strings.Contains(stdout.String(), "kilnscrub "+version) {
			t.Errorf("args %v output = %q, want prefix %q", args, stdout.String(), "kilnscrub "+version)
		}
	}
}

func TestRunPresetsTextAndJSON(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), []string{"presets"}, &stdout, &stderr); err != nil {
			t.Fatalf("presets failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, id := range []string{"electric_raku_6", "gas_soda_14", "wood_salt_24", "coop_wood_48", "foundry_glass_16"} {
			if !strings.Contains(out, id) {
				t.Errorf("presets output missing %q;\n%s", id, out)
			}
		}
	})

	t.Run("json", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), []string{"presets", "--json"}, &stdout, &stderr); err != nil {
			t.Fatalf("presets --json failed: %v (stderr: %s)", err, stderr.String())
		}
		var payload struct {
			Presets []studio.StudioPreset `json:"presets"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("unmarshal presets --json: %v", err)
		}
		if len(payload.Presets) != 5 {
			t.Fatalf("got %d presets, want 5", len(payload.Presets))
		}
	})
}

func TestRunGuideStdoutAndOutDirAndProject(t *testing.T) {
	t.Run("stdout_default_and_preset", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), []string{"guide", "--preset", "gas_soda_14"}, &stdout, &stderr); err != nil {
			t.Fatalf("guide failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"Kiln Scrub Studio Readiness & COTS Summary",
			"gas_soda_14",
			"Urban Readiness",
			"Winter Plume",
			"# Build & Assembly Guide",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("guide stdout missing %q;\n%s", want, out)
			}
		}
	})

	t.Run("write_guides_from_project_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		projPath := filepath.Join(tmpDir, "project.json")
		proj := studio.DefaultProject()
		proj.StudioName = "Cedar Creek Wood-Salt"
		proj.IncludePlumeDiluter = true
		if err := studio.SaveProject(projPath, proj); err != nil {
			t.Fatalf("save project: %v", err)
		}

		outDir := filepath.Join(tmpDir, "guides")
		var stdout, stderr bytes.Buffer
		if err := run(context.Background(), []string{"guide", "--project", projPath, "--out", outDir}, &stdout, &stderr); err != nil {
			t.Fatalf("guide --project --out failed: %v (stderr: %s)", err, stderr.String())
		}
		for _, fname := range []string{"BUILD_GUIDE.md", "OPERATIONS_CHECKLIST.md"} {
			data, err := os.ReadFile(filepath.Join(outDir, fname))
			if err != nil {
				t.Fatalf("read %s: %v", fname, err)
			}
			if !strings.Contains(string(data), "Cedar Creek Wood-Salt") {
				t.Errorf("%s missing studio name", fname)
			}
		}
	})
	t.Run("phase7_stage_design_flags", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--salting-hours", "2.5",
			"--alignment", "eductor_dry_fan",
			"--bed-layout", "parallel_manifold",
			"--venturi-layout", "parallel_mini_venturis",
			"--compact",
			"--controls", "electromechanical_only",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide with phase 7 flags failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"Stage Options",
			"alignment=eductor_dry_fan",
			"bed=parallel_manifold",
			"venturi=parallel_mini_venturis",
			"compact=true",
			"controls=electromechanical_only",
			"Recirculation",
			"SG >= 1.105",
			"Shop Plan",
			"9 parts",
			"15 lb",
			"50 lb",
			"2.5 hr salting",
			"Standard 316L SS (Light / Occasional)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("guide stdout missing %q;\n%s", want, out)
			}
		}
		for _, bad := range []string{"4-lb", "Hobby 316L"} {
			if strings.Contains(out, bad) {
				t.Errorf("guide stdout must not contain deprecated string %q", bad)
			}
		}
	})

	t.Run("phase8_budget_combo_flags", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--blower-drive", "compressor_air_atomize",
			"--bed-layout", "parallel_manifold",
			"--parallel-pipe-in", "4",
			"--packing-media", "diy_ceramic_rings",
			"--compressor-scfm", "15",
			"--controls", "electromechanical_only",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide with phase 8 flags failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"Stage Options",
			"bed=parallel_manifold (4\" pipe)",
			"controls=electromechanical_only",
			"Blower Drive",
			"compressor_air_atomize",
			"Packing Media",
			"diy_ceramic_rings",
			"Shop Compressor (Eductor vs Air-Atomize)",
			"15.0 SCFM avail",
			"Eductor:",
			"Air-Atomize:",
			"feasible=true",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 8 guide stdout missing %q;\n%s", want, out)
			}
		}
		for _, bad := range []string{"4-lb", "Hobby 316L"} {
			if strings.Contains(out, bad) {
				t.Errorf("phase 8 guide stdout must not contain deprecated string %q", bad)
			}
		}
	})

	t.Run("phase8_compressor_eductor_only_feasibility", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--blower-drive", "compressor_eductor_only",
			"--compressor-scfm", "15",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide --blower-drive compressor_eductor_only failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"Blower Drive",
			"compressor_eductor_only",
			"feasible=false",
			"Shop Compressor (Eductor vs Air-Atomize)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("compressor_eductor_only stdout missing %q;\n%s", want, out)
			}
		}
	})

	t.Run("phase9_custom_fab_deluxe_flag", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--build-path", "custom_fab_deluxe",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide --build-path custom_fab_deluxe failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"path=custom_fab_deluxe",
			"Build Path / Deluxe : custom_fab_deluxe",
			"delta=+",
			"dBA @ 3ft",
			"Shop Plan Systems :",
			"electrical circuits",
			"plumbing lines",
			"bracing members",
			"acoustic zones",
			`NEC 110.26 36"x30" panel zone`,
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 9 guide stdout missing %q;\n%s", want, out)
			}
		}
		for _, bad := range []string{"4-lb", "Hobby 316L"} {
			if strings.Contains(out, bad) {
				t.Errorf("phase 9 guide stdout must not contain deprecated string %q", bad)
			}
		}
	})

	t.Run("phase10_nairobi_zero_electricity_soda_ash", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--zero-electricity",
			"--local-region", "nairobi",
			"--local-alkali", "soda_ash",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide nairobi zero-electricity failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"path=local_ceramic_masonry",
			"Local Ceramic / Off-Grid :",
			"Nairobi / East Africa (KES)",
			"KES",
			"Power: 0 W Off-Grid",
			"Chimney & Draft :",
			"inWC (feasible=true)",
			"Ceramic Bed     :",
			"Ergun dP=",
			"Gravity & Labor :",
			"jerrycans/hr",
			"treadle min/hr",
			"Local Alkali    :",
			"Na2CO3",
			"Honest Capture  :",
			"HCl ",
			"Coarse Ash ",
			"Submicron NaCl ",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 10 nairobi guide stdout missing %q;\n%s", want, out)
			}
		}
		for _, bad := range []string{"4-lb box", "Hobby 316L", "Switching from Salt (NaCl) to Soda"} {
			if strings.Contains(out, bad) {
				t.Errorf("phase 10 nairobi guide stdout must not contain deprecated string %q", bad)
			}
		}
	})

	t.Run("phase10_mumbai_zero_electricity_slaked_lime", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--build-path", "local_ceramic_masonry",
			"--zero-electricity",
			"--local-region", "mumbai",
			"--local-alkali", "slaked_lime",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide mumbai slaked_lime failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"path=local_ceramic_masonry",
			"Mumbai / Western India (INR)",
			"INR",
			"Power: 0 W Off-Grid",
			"Ca(OH)2",
			"jerrycans/hr",
			"treadle min/hr",
			"inWC (feasible=true)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 10 mumbai guide stdout missing %q;\n%s", want, out)
			}
		}
	})

	t.Run("phase10_wood_ash_lye_zero_electricity", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--zero-electricity",
			"--local-region", "general_local",
			"--local-alkali", "wood_ash_lye",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide wood_ash_lye failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"path=local_ceramic_masonry",
			"Power: 0 W Off-Grid",
			"K2CO3",
			"$0.00",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 10 wood_ash_lye guide stdout missing %q;\n%s", want, out)
			}
		}
	})

	t.Run("phase10_local_ceramic_hybrid_without_zero_electricity", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := run(context.Background(), []string{
			"guide",
			"--preset", "wood_salt_24",
			"--build-path", "local_ceramic_masonry",
			"--local-region", "mumbai",
			"--local-alkali", "soda_ash",
		}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("guide local_ceramic_masonry hybrid failed: %v (stderr: %s)", err, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{
			"path=local_ceramic_masonry",
			"Mumbai / Western India (INR)",
			"W Local Hybrid",
			"0.5 HP 230V",
			"Na2CO3",
			"inWC (feasible=true)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("phase 10 hybrid guide stdout missing %q;\n%s", want, out)
			}
		}
		if strings.Contains(out, "0 W Off-Grid") {
			t.Errorf("phase 10 hybrid guide must not report 0 W Off-Grid:\n%s", out)
		}
	})
}

func TestRunPacketCreatesValid11FileZip(t *testing.T) {
	outZip := filepath.Join(t.TempDir(), "nested", "ci-packet.zip")
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{
		"packet",
		"--preset", "wood_salt_24",
		"--out", outZip,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("packet failed: %v (stderr: %s)", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "11 files") {
		t.Errorf("packet stdout missing '11 files': %q", stdout.String())
	}

	zipBytes, err := os.ReadFile(outZip)
	if err != nil {
		t.Fatalf("read zip %s: %v", outZip, err)
	}
	assertValid11FilePacketZip(t, zipBytes)

	t.Run("phase7_flags_propagate_to_packet", func(t *testing.T) {
		optZip := filepath.Join(t.TempDir(), "phase7-packet.zip")
		var oStdout, oStderr bytes.Buffer
		err := run(context.Background(), []string{
			"packet",
			"--preset", "wood_salt_24",
			"--salting-hours", "2.5",
			"--alignment", "eductor_dry_fan",
			"--bed-layout", "parallel_manifold",
			"--venturi-layout", "parallel_mini_venturis",
			"--compact",
			"--controls", "electromechanical_only",
			"--out", optZip,
		}, &oStdout, &oStderr)
		if err != nil {
			t.Fatalf("packet with phase 7 flags failed: %v (stderr: %s)", err, oStderr.String())
		}
		data, err := os.ReadFile(optZip)
		if err != nil {
			t.Fatalf("read zip %s: %v", optZip, err)
		}
		files := assertValid11FilePacketZip(t, data)

		var proj studio.StudioProject
		if err := json.Unmarshal(files["project.json"], &proj); err != nil {
			t.Fatalf("unmarshal project.json: %v", err)
		}
		if proj.SaltingHours != 2.5 ||
			proj.StageAlignment != studio.StageAlignmentEductorDryFan ||
			proj.PackedBedLayout != studio.PackedBedLayoutParallelManifold ||
			proj.VenturiLayout != studio.VenturiLayoutParallelMiniVenturis ||
			!proj.CompactFootprint ||
			proj.ControlTier != studio.ControlTierElectromechanicalOnly {
			t.Errorf("project.json did not preserve Phase 7 CLI flags: %+v", proj)
		}

		var plan studio.ShopPlan
		if err := json.Unmarshal(files["shop_plan.json"], &plan); err != nil {
			t.Fatalf("unmarshal shop_plan.json: %v", err)
		}
		if plan.StageAlignment != studio.StageAlignmentEductorDryFan ||
			plan.PackedBedLayout != studio.PackedBedLayoutParallelManifold ||
			!plan.CompactFootprint ||
			len(plan.Parts) < 9 ||
			len(plan.CutList) == 0 {
			t.Errorf("shop_plan.json did not reflect Phase 7 CLI flags: %+v", plan)
		}
	})

	t.Run("phase8_flags_propagate_to_packet", func(t *testing.T) {
		optZip := filepath.Join(t.TempDir(), "phase8-packet.zip")
		var oStdout, oStderr bytes.Buffer
		err := run(context.Background(), []string{
			"packet",
			"--preset", "wood_salt_24",
			"--blower-drive", "compressor_air_atomize",
			"--bed-layout", "parallel_manifold",
			"--parallel-pipe-in", "4",
			"--packing-media", "diy_ceramic_rings",
			"--compressor-scfm", "15",
			"--controls", "electromechanical_only",
			"--out", optZip,
		}, &oStdout, &oStderr)
		if err != nil {
			t.Fatalf("packet with phase 8 flags failed: %v (stderr: %s)", err, oStderr.String())
		}
		data, err := os.ReadFile(optZip)
		if err != nil {
			t.Fatalf("read zip %s: %v", optZip, err)
		}
		files := assertValid11FilePacketZip(t, data)

		var proj studio.StudioProject
		if err := json.Unmarshal(files["project.json"], &proj); err != nil {
			t.Fatalf("unmarshal project.json: %v", err)
		}
		if proj.BlowerDriveOption != studio.BlowerDriveCompressorAirAtomize ||
			proj.PackedBedLayout != studio.PackedBedLayoutParallelManifold ||
			proj.ParallelPipeSizeIn != 4.0 ||
			proj.PackingMediaOption != studio.PackingMediaDIYCeramicRings ||
			proj.ShopCompressorSCFM != 15.0 ||
			proj.ControlTier != studio.ControlTierElectromechanicalOnly {
			t.Errorf("project.json did not preserve Phase 8 CLI flags: %+v", proj)
		}
	})

	t.Run("phase9_custom_fab_deluxe_propagates_to_packet", func(t *testing.T) {
		optZip := filepath.Join(t.TempDir(), "phase9-deluxe-packet.zip")
		var oStdout, oStderr bytes.Buffer
		err := run(context.Background(), []string{
			"packet",
			"--preset", "wood_salt_24",
			"--build-path", "custom_fab_deluxe",
			"--out", optZip,
		}, &oStdout, &oStderr)
		if err != nil {
			t.Fatalf("packet --build-path custom_fab_deluxe failed: %v (stderr: %s)", err, oStderr.String())
		}
		data, err := os.ReadFile(optZip)
		if err != nil {
			t.Fatalf("read zip %s: %v", optZip, err)
		}
		files := assertValid11FilePacketZip(t, data)

		var proj studio.StudioProject
		if err := json.Unmarshal(files["project.json"], &proj); err != nil {
			t.Fatalf("unmarshal project.json: %v", err)
		}
		if proj.BuildPath != studio.BuildPathCustomFabDeluxe {
			t.Errorf("project.json BuildPath = %q, want %q", proj.BuildPath, studio.BuildPathCustomFabDeluxe)
		}

		var rawPlan map[string]json.RawMessage
		if err := json.Unmarshal(files["shop_plan.json"], &rawPlan); err != nil {
			t.Fatalf("unmarshal raw shop_plan.json: %v", err)
		}
		for _, key := range []string{"is_deluxe", "electrical", "plumbing", "bracing", "acoustics", "deluxe_spec"} {
			if _, ok := rawPlan[key]; !ok {
				t.Errorf("shop_plan.json missing required Phase 9 key %q", key)
			}
		}

		var plan studio.ShopPlan
		if err := json.Unmarshal(files["shop_plan.json"], &plan); err != nil {
			t.Fatalf("unmarshal shop_plan.json: %v", err)
		}
		if !plan.IsDeluxe ||
			plan.BuildPath != studio.BuildPathCustomFabDeluxe ||
			len(plan.Electrical) == 0 ||
			len(plan.Plumbing) == 0 ||
			len(plan.Bracing) == 0 ||
			len(plan.Acoustics) == 0 ||
			plan.DeluxeSpec == nil ||
			!plan.DeluxeSpec.Enabled ||
			plan.DeluxeSpec.CostDeltaUSD <= 0 ||
			plan.Clearances.NECPanelDepthIn != 36.0 ||
			plan.Clearances.NECPanelWidthIn != 30.0 {
			t.Errorf("shop_plan.json did not reflect Phase 9 Custom Fab Deluxe systems: %+v", plan)
		}
	})

	t.Run("phase10_nairobi_mumbai_wood_ash_and_hybrid_packets", func(t *testing.T) {
		cases := []struct {
			name            string
			args            []string
			wantZeroElec    bool
			wantRegion      string
			wantAlkali      string
			wantGuideSubstr string
		}{
			{
				name: "nairobi_soda_ash_zero_elec",
				args: []string{
					"--build-path", "local_ceramic_masonry",
					"--zero-electricity",
					"--local-region", "nairobi",
					"--local-alkali", "soda_ash",
				},
				wantZeroElec:    true,
				wantRegion:      studio.LocalRegionNairobi,
				wantAlkali:      studio.LocalAlkaliSodaAsh,
				wantGuideSubstr: "Nairobi",
			},
			{
				name: "mumbai_slaked_lime_zero_elec",
				args: []string{
					"--zero-electricity",
					"--local-region", "mumbai",
					"--local-alkali", "slaked_lime",
				},
				wantZeroElec:    true,
				wantRegion:      studio.LocalRegionMumbai,
				wantAlkali:      studio.LocalAlkaliSlakedLime,
				wantGuideSubstr: "Mumbai",
			},
			{
				name: "wood_ash_lye_zero_elec",
				args: []string{
					"--zero-electricity",
					"--local-region", "nairobi",
					"--local-alkali", "wood_ash_lye",
				},
				wantZeroElec:    true,
				wantRegion:      studio.LocalRegionNairobi,
				wantAlkali:      studio.LocalAlkaliWoodAshLye,
				wantGuideSubstr: "K2CO3",
			},
			{
				name: "hybrid_local_ceramic",
				args: []string{
					"--build-path", "local_ceramic_masonry",
					"--local-region", "mumbai",
					"--local-alkali", "soda_ash",
				},
				wantZeroElec:    false,
				wantRegion:      studio.LocalRegionMumbai,
				wantAlkali:      studio.LocalAlkaliSodaAsh,
				wantGuideSubstr: "0.5 HP 230V",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				optZip := filepath.Join(t.TempDir(), tc.name+".zip")
				cmdArgs := append([]string{"packet", "--preset", "wood_salt_24"}, tc.args...)
				cmdArgs = append(cmdArgs, "--out", optZip)
				var oStdout, oStderr bytes.Buffer
				if err := run(context.Background(), cmdArgs, &oStdout, &oStderr); err != nil {
					t.Fatalf("packet %s failed: %v (stderr: %s)", tc.name, err, oStderr.String())
				}
				data, err := os.ReadFile(optZip)
				if err != nil {
					t.Fatalf("read zip %s: %v", optZip, err)
				}
				files := assertValid11FilePacketZip(t, data)

				var proj studio.StudioProject
				if err := json.Unmarshal(files["project.json"], &proj); err != nil {
					t.Fatalf("unmarshal project.json: %v", err)
				}
				if proj.BuildPath != studio.BuildPathLocalCeramic ||
					proj.ZeroElectricity != tc.wantZeroElec ||
					proj.LocalRegion != tc.wantRegion ||
					proj.LocalAlkali != tc.wantAlkali {
					t.Errorf("project.json mismatch: got build=%q zero=%v region=%q alkali=%q",
						proj.BuildPath, proj.ZeroElectricity, proj.LocalRegion, proj.LocalAlkali)
				}
				if !strings.Contains(string(files["BUILD_GUIDE.md"]), tc.wantGuideSubstr) {
					t.Errorf("BUILD_GUIDE.md missing %q", tc.wantGuideSubstr)
				}
				var plan studio.ShopPlan
				if err := json.Unmarshal(files["shop_plan.json"], &plan); err != nil {
					t.Fatalf("unmarshal shop_plan.json: %v", err)
				}
				if plan.BuildPath != studio.BuildPathLocalCeramic || len(plan.Parts) == 0 {
					t.Errorf("shop_plan.json did not reflect local_ceramic_masonry: build=%q parts=%d",
						plan.BuildPath, len(plan.Parts))
				}
			})
		}
	})
}

func TestRunDesignAndExportWithPreset(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run(context.Background(), []string{"design", "--preset", "gas_soda_14", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("design --preset failed: %v (stderr: %s)", err, stderr.String())
	}
	var design model.ScrubberDesign
	if err := json.Unmarshal(stdout.Bytes(), &design); err != nil {
		t.Fatalf("unmarshal design --preset json: %v", err)
	}
	if design.Input.KilnVolumeCuFt != 14.0 || design.Input.SaltRateLbsPerHr != 0.0 || design.Input.FuelType != model.FuelGas {
		t.Errorf("unexpected preset design input: %+v", design.Input)
	}

	// Verify explicit flag overrides preset base.
	stdout.Reset()
	stderr.Reset()
	if err := run(context.Background(), []string{"design", "--preset", "gas_soda_14", "--kiln-volume", "18", "--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("design --preset with override failed: %v", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), &design); err != nil {
		t.Fatalf("unmarshal override json: %v", err)
	}
	if design.Input.KilnVolumeCuFt != 18.0 || design.Input.FuelType != model.FuelGas {
		t.Errorf("override did not preserve preset fuel or apply kiln-volume: %+v", design.Input)
	}

	// Verify export --preset succeeds.
	outDir := filepath.Join(t.TempDir(), "preset-export")
	stdout.Reset()
	stderr.Reset()
	if err := run(context.Background(), []string{"export", "--preset", "electric_raku_6", "--out", outDir}, &stdout, &stderr); err != nil {
		t.Fatalf("export --preset failed: %v (stderr: %s)", err, stderr.String())
	}
	for _, f := range []string{"scrubber.stl", "scrubber.dxf", "bom.csv", "bom.md"} {
		info, err := os.Stat(filepath.Join(outDir, f))
		if err != nil || info.Size() == 0 {
			t.Errorf("export --preset missing or empty %s: err=%v", f, err)
		}
	}
}

func TestRunUIStudioBackendRoutes(t *testing.T) {
	addr := reserveEphemeralAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdout, stderr bytes.Buffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, []string{"ui", "--addr", addr, "--no-control"}, &stdout, &stderr)
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := "http://" + addr

	// Wait until server is ready by polling /api/studio/presets.
	var presetsPayload struct {
		Presets []studio.StudioPreset `json:"presets"`
	}
	deadline := time.Now().Add(3 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/api/studio/presets")
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				err = json.NewDecoder(resp.Body).Decode(&presetsPayload)
				resp.Body.Close()
				if err == nil {
					ready = true
					break
				}
			} else {
				resp.Body.Close()
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		cancel()
		t.Fatalf("ui studio server did not become ready on %s (stderr=%q)", addr, stderr.String())
	}
	if len(presetsPayload.Presets) != 5 {
		t.Fatalf("GET /api/studio/presets returned %d presets, want 5", len(presetsPayload.Presets))
	}

	// POST /api/studio/evaluate
	evalReq := `{"preset_id":"wood_salt_24","include_plume_diluter":true,"property_line_ft":35,"ambient_winter_f":35}`
	evalResp, err := client.Post(baseURL+"/api/studio/evaluate", "application/json", strings.NewReader(evalReq))
	if err != nil {
		cancel()
		t.Fatalf("POST /api/studio/evaluate failed: %v", err)
	}
	evalBody, _ := io.ReadAll(evalResp.Body)
	evalResp.Body.Close()
	if evalResp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("POST /api/studio/evaluate status = %d, want 200; body=%s", evalResp.StatusCode, string(evalBody))
	}
	var evalOut studio.StudioEvaluation
	if err := json.Unmarshal(evalBody, &evalOut); err != nil {
		cancel()
		t.Fatalf("unmarshal StudioEvaluation: %v", err)
	}
	if evalOut.COTS.EstimatedTotalUSD <= 0 || evalOut.COTS.PlumeDiluter == nil {
		t.Errorf("unexpected COTS evaluation: total=%.2f diluter=%v", evalOut.COTS.EstimatedTotalUSD, evalOut.COTS.PlumeDiluter)
	}
	if evalOut.Urban.Plume.BarePlumeLengthFt <= 0 || evalOut.Urban.Plume.DilutedPlumeLengthFt <= 0 ||
		evalOut.Urban.Plume.DilutedPlumeLengthFt >= evalOut.Urban.Plume.BarePlumeLengthFt {
		t.Errorf("unexpected plume reach lengths: bare=%.1f diluted=%.1f",
			evalOut.Urban.Plume.BarePlumeLengthFt, evalOut.Urban.Plume.DilutedPlumeLengthFt)
	}
	if !strings.Contains(evalOut.BuildGuideMD, "# Build & Assembly Guide") || !strings.Contains(evalOut.OperationsChecklistMD, "# Firing-Day Operations Checklist") {
		t.Errorf("missing guide or checklist markdown in evaluation response")
	}

	// POST /api/studio/packet
	pktResp, err := client.Post(baseURL+"/api/studio/packet", "application/json", strings.NewReader(evalReq))
	if err != nil {
		cancel()
		t.Fatalf("POST /api/studio/packet failed: %v", err)
	}
	pktBytes, _ := io.ReadAll(pktResp.Body)
	pktResp.Body.Close()
	if pktResp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("POST /api/studio/packet status = %d, want 200; body=%s", pktResp.StatusCode, string(pktBytes))
	}
	if ct := pktResp.Header.Get("Content-Type"); ct != "application/zip" {
		t.Errorf("packet Content-Type = %q, want application/zip", ct)
	}
	if cd := pktResp.Header.Get("Content-Disposition"); !strings.Contains(cd, "kilnscrub-wood_salt_24-packet.zip") {
		t.Errorf("packet Content-Disposition = %q, want kilnscrub-wood_salt_24-packet.zip", cd)
	}
	assertValid11FilePacketZip(t, pktBytes)

	cancel()
	select {
	case runErr := <-errCh:
		if runErr != nil {
			t.Fatalf("ui returned error on shutdown: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("ui did not shut down cleanly")
	}
}

func assertValid11FilePacketZip(t *testing.T, zipBytes []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("invalid zip archive: %v", err)
	}
	expected := map[string]bool{
		"scrubber.stl":            false,
		"scrubber.dxf":            false,
		"shop_plan.json":          false,
		"bom_cots.csv":            false,
		"bom_cots.md":             false,
		"bom_engineered.csv":      false,
		"BUILD_GUIDE.md":          false,
		"OPERATIONS_CHECKLIST.md": false,
		"project.json":            false,
		"daemon.json":             false,
		"README_PACKET.txt":       false,
	}
	if len(zr.File) != len(expected) {
		t.Fatalf("zip contains %d files, want %d", len(zr.File), len(expected))
	}
	contents := make(map[string][]byte, len(expected))
	for _, f := range zr.File {
		if _, ok := expected[f.Name]; !ok {
			t.Errorf("unexpected zip entry %q", f.Name)
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Errorf("open zip entry %q: %v", f.Name, err)
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil || len(data) == 0 {
			t.Errorf("zip entry %q empty or unreadable: len=%d err=%v", f.Name, len(data), err)
		}
		expected[f.Name] = true
		contents[f.Name] = data
	}
	for name, seen := range expected {
		if !seen {
			t.Errorf("zip missing expected entry %q", name)
		}
	}
	return contents
}
