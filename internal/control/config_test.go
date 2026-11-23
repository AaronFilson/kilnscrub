package control

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func TestLoadExampleDaemonConfig(t *testing.T) {
	examplePath := filepath.Join("..", "..", "configs", "kilnscrub.example.json")
	cfg, err := LoadDaemonConfig(examplePath)
	if err != nil {
		t.Fatalf("LoadDaemonConfig(%q) failed: %v", examplePath, err)
	}

	if cfg.HALMode != HALModeSim {
		t.Fatalf("example config hal_mode = %q, want %q", cfg.HALMode, HALModeSim)
	}
	if err := cfg.Control.Validate(); err != nil {
		t.Fatalf("example control config invalid: %v", err)
	}
	if err := cfg.Modbus.Validate(); err != nil {
		t.Fatalf("example modbus config invalid: %v", err)
	}
	if err := cfg.Sim.Validate(); err != nil {
		t.Fatalf("example sim config invalid: %v", err)
	}
	if cfg.Modbus.Registers.OutputScale != 10000 {
		t.Fatalf("example modbus output_scale = %d, want 10000", cfg.Modbus.Registers.OutputScale)
	}
	if math.Abs(cfg.Control.QuenchSetpoint.Fahrenheit()-188.0) > 1e-4 {
		t.Fatalf("example control quench_setpoint = %.2f F, want 188.0 F", cfg.Control.QuenchSetpoint.Fahrenheit())
	}
	if math.Abs(cfg.Sim.BaselineQuenchTemp.Fahrenheit()-188.0) > 1e-4 {
		t.Fatalf("example sim baseline_quench_temp = %.2f F, want 188.0 F", cfg.Sim.BaselineQuenchTemp.Fahrenheit())
	}
}

func TestParsePartialDaemonConfigInheritsDefaultsAndSyncs(t *testing.T) {
	raw := []byte(`{
		"control": {
			"target_draft_in_wc": -0.08,
			"draft_filter_cutoff_hz": 1.0
		},
		"sim": {
			"initial_sump_ph": 8.4
		}
	}`)

	cfg, err := ParseDaemonConfig(raw)
	if err != nil {
		t.Fatalf("ParseDaemonConfig(partial) failed: %v", err)
	}

	if cfg.HALMode != HALModeSim {
		t.Fatalf("expected default hal_mode %q, got %q", HALModeSim, cfg.HALMode)
	}
	wantPa := model.InWCToPascal(-0.08)
	if math.Abs(float64(cfg.Control.TargetDraftPa-wantPa)) > 1e-6 {
		t.Fatalf("Control.Sync() did not sync TargetDraftPa: got %.6f, want %.6f", cfg.Control.TargetDraftPa, wantPa)
	}
	wantTau := model.CutoffHzToTimeConst(1.0)
	if cfg.Control.DraftFilterTimeConst != wantTau {
		t.Fatalf("Control.Sync() did not sync DraftFilterTimeConst: got %v, want %v", cfg.Control.DraftFilterTimeConst, wantTau)
	}
	// Unspecified fields must inherit defaults.
	if cfg.Control.LoopInterval != 100*time.Millisecond {
		t.Fatalf("expected default LoopInterval 100ms, got %v", cfg.Control.LoopInterval)
	}
	if cfg.Sim.InitialSumpPH != 8.4 {
		t.Fatalf("expected overridden InitialSumpPH 8.4, got %v", cfg.Sim.InitialSumpPH)
	}
	if cfg.Sim.MaxWaterFlowGPM != DefaultSimHALConfig().MaxWaterFlowGPM {
		t.Fatalf("expected default MaxWaterFlowGPM %v, got %v", DefaultSimHALConfig().MaxWaterFlowGPM, cfg.Sim.MaxWaterFlowGPM)
	}

	// Partial modbus config inheriting default register map.
	mbRaw := []byte(`{
		"hal_mode": "MODBUS",
		"modbus": {
			"device": "/dev/ttyAMA0",
			"baud_rate": 19200,
			"slave_id": 3
		}
	}`)
	mbCfg, err := ParseDaemonConfig(mbRaw)
	if err != nil {
		t.Fatalf("ParseDaemonConfig(partial modbus) failed: %v", err)
	}
	if mbCfg.HALMode != HALModeModbus {
		t.Fatalf("expected normalized hal_mode %q, got %q", HALModeModbus, mbCfg.HALMode)
	}
	if mbCfg.Modbus.Device != "/dev/ttyAMA0" || mbCfg.Modbus.BaudRate != 19200 || mbCfg.Modbus.SlaveID != 3 {
		t.Fatalf("unexpected modbus overrides: %+v", mbCfg.Modbus)
	}
	if mbCfg.Modbus.Registers.OutputScale != 10000 {
		t.Fatalf("expected inherited default output_scale 10000, got %d", mbCfg.Modbus.Registers.OutputScale)
	}
}

func TestDaemonConfigInvalidRejection(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "malformed JSON",
			raw:  `{"hal_mode": "sim",`,
		},
		{
			name: "unsupported hal_mode",
			raw:  `{"hal_mode": "spi"}`,
		},
		{
			name: "positive draft setpoint",
			raw:  `{"control": {"target_draft_in_wc": 0.15}}`,
		},
		{
			name: "quench setpoint above CPVC limit",
			raw:  `{"control": {"quench_setpoint_k": 380.0}}`,
		},
		{
			name: "modbus unsupported baud rate",
			raw:  `{"hal_mode": "modbus", "modbus": {"baud_rate": 12345}}`,
		},
		{
			name: "modbus invalid slave id 0",
			raw:  `{"hal_mode": "modbus", "modbus": {"slave_id": 0}}`,
		},
		{
			name: "modbus invalid slave id 248",
			raw:  `{"hal_mode": "modbus", "modbus": {"slave_id": 248}}`,
		},
		{
			name: "modbus empty device",
			raw:  `{"hal_mode": "modbus", "modbus": {"device": ""}}`,
		},
		{
			name: "modbus zero register scale",
			raw:  `{"hal_mode": "modbus", "modbus": {"registers": {"draft_pa": {"address": 0, "scale": 0}}}}`,
		},
		{
			name: "modbus duplicate output register address",
			raw:  `{"hal_mode": "modbus", "modbus": {"registers": {"blower_vfd_addr": 100, "quench_spray_valve_addr": 100}}}`,
		},
		{
			name: "modbus duplicate coil address",
			raw:  `{"hal_mode": "modbus", "modbus": {"registers": {"bypass_damper_coil_addr": 2, "alarm_latch_coil_addr": 2}}}`,
		},
		{
			name: "modbus unsupported temp_unit",
			raw:  `{"hal_mode": "modbus", "modbus": {"registers": {"temp_unit": "R"}}}`,
		},
		{
			name: "sim invalid initial sump pH",
			raw:  `{"hal_mode": "sim", "sim": {"initial_sump_ph": 15.0}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseDaemonConfig([]byte(tc.raw)); err == nil {
				t.Fatalf("expected ParseDaemonConfig to reject %s", tc.name)
			}
		})
	}

	if _, err := LoadDaemonConfig("/nonexistent/path/kilnscrub.json"); err == nil {
		t.Fatalf("expected LoadDaemonConfig to fail for missing file")
	}
}
