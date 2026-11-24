package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Supported HALMode values for DaemonConfig.
const (
	HALModeSim    = "sim"
	HALModeModbus = "modbus"
)

// DaemonConfig is the complete JSON configuration for running the kilnscrub control daemon
// with either the physics-based simulation HAL ("sim") or a physical Modbus RTU PLC/VFD ("modbus").
type DaemonConfig struct {
	Description string              `json:"description,omitempty"`
	HALMode     string              `json:"hal_mode"`
	Control     model.ControlConfig `json:"control"`
	Modbus      ModbusConfig        `json:"modbus"`
	Sim         SimHALConfig        `json:"sim"`
}

// DefaultDaemonConfig returns a fully valid baseline configuration defaulting to simulation mode.
func DefaultDaemonConfig() DaemonConfig {
	return DaemonConfig{
		HALMode: HALModeSim,
		Control: model.DefaultControlConfig(),
		Modbus:  DefaultModbusConfig(),
		Sim:     DefaultSimHALConfig(),
	}
}

// ParseDaemonConfig decodes JSON configuration seeded from DefaultDaemonConfig so partial
// JSON files inherit valid defaults, rejects unknown JSON keys so typos in safety parameters
// fail fast, reconciles derived control fields via Control.Sync(), and validates DaemonConfig.
func ParseDaemonConfig(data []byte) (DaemonConfig, error) {
	cfg := DefaultDaemonConfig()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return DaemonConfig{}, fmt.Errorf("parse daemon config JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return DaemonConfig{}, fmt.Errorf("parse daemon config JSON: unexpected trailing tokens")
	}
	cfg.HALMode = strings.ToLower(strings.TrimSpace(cfg.HALMode))
	if cfg.HALMode == "" {
		cfg.HALMode = HALModeSim
	}
	cfg.Control.Sync()
	if err := cfg.Validate(); err != nil {
		return DaemonConfig{}, fmt.Errorf("invalid daemon config: %w", err)
	}
	return cfg, nil
}

// LoadDaemonConfig reads a JSON configuration file from path, seeds defaults, syncs derived
// control parameters, and validates all fields.
func LoadDaemonConfig(path string) (DaemonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DaemonConfig{}, fmt.Errorf("read daemon config %q: %w", path, err)
	}
	return ParseDaemonConfig(data)
}

// Validate verifies HALMode, ControlConfig, and the active HAL's configuration block.
func (c DaemonConfig) Validate() error {
	switch c.HALMode {
	case HALModeSim, HALModeModbus:
	default:
		return fmt.Errorf("hal_mode must be %q or %q, got %q", HALModeSim, HALModeModbus, c.HALMode)
	}

	if err := c.Control.Validate(); err != nil {
		return fmt.Errorf("control: %w", err)
	}

	if c.HALMode == HALModeModbus {
		if err := c.Modbus.Validate(); err != nil {
			return err
		}
	} else {
		if err := c.Sim.Validate(); err != nil {
			return fmt.Errorf("sim: %w", err)
		}
	}
	return nil
}

// Validate checks that SimHALConfig parameters are finite and physically consistent.
func (s SimHALConfig) Validate() error {
	if !isFinite(float64(s.NominalInletTemp)) || s.NominalInletTemp <= 300 || s.NominalInletTemp > 2500 {
		return fmt.Errorf("nominal_inlet_temp_k must be finite and in (300, 2500], got %v", s.NominalInletTemp)
	}
	if !isFinite(float64(s.BaselineQuenchTemp)) || s.BaselineQuenchTemp <= 273.15 || s.BaselineQuenchTemp >= s.NominalInletTemp {
		return fmt.Errorf("baseline_quench_temp_k must be finite and in (273.15, nominal_inlet_temp_k), got %v", s.BaselineQuenchTemp)
	}
	if !isFinite(s.MaxWaterFlowGPM) || s.MaxWaterFlowGPM <= 0 {
		return fmt.Errorf("max_water_flow_gpm must be finite and > 0, got %v", s.MaxWaterFlowGPM)
	}
	if !isFinite(s.NominalWaterFlowGPM) || s.NominalWaterFlowGPM <= 0 || s.NominalWaterFlowGPM > s.MaxWaterFlowGPM {
		return fmt.Errorf("nominal_water_flow_gpm must be finite and in (0, max_water_flow_gpm], got %v", s.NominalWaterFlowGPM)
	}
	if !isFinite(s.BlowerMaxSuctionPa) || s.BlowerMaxSuctionPa <= 0 {
		return fmt.Errorf("blower_max_suction_pa must be finite and > 0, got %v", s.BlowerMaxSuctionPa)
	}
	if s.BlowerTimeConst <= 0 || s.DraftTimeConst <= 0 || s.ValveTimeConst <= 0 || s.QuenchTimeConst <= 0 {
		return fmt.Errorf("sim time constants must all be > 0")
	}
	if !isFinite(float64(s.DryGasOutletTemp)) || s.DryGasOutletTemp <= s.BaselineQuenchTemp {
		return fmt.Errorf("dry_gas_outlet_temp_k must be finite and > baseline_quench_temp_k, got %v", s.DryGasOutletTemp)
	}
	if !isFinite(s.InitialSumpPH) || s.InitialSumpPH < 0 || s.InitialSumpPH > 14.0 {
		return fmt.Errorf("initial_sump_ph must be finite and in [0, 14], got %v", s.InitialSumpPH)
	}
	if !isFinite(s.DraftNoiseAmplitudePa) || s.DraftNoiseAmplitudePa < 0 {
		return fmt.Errorf("draft_noise_amplitude_pa must be finite and >= 0, got %v", s.DraftNoiseAmplitudePa)
	}
	if !isFinite(s.DraftNoiseFreqHz) || s.DraftNoiseFreqHz < 0 {
		return fmt.Errorf("draft_noise_freq_hz must be finite and >= 0, got %v", s.DraftNoiseFreqHz)
	}
	if s.AutoStepInterval < 0 {
		return fmt.Errorf("auto_step_interval_ns must be >= 0, got %v", s.AutoStepInterval)
	}
	return nil
}
