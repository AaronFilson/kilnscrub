package control

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Supported temperature unit identifiers for ModbusRegisterMap.TempUnit.
const (
	ModbusTempUnitF = "F"
	ModbusTempUnitC = "C"
	ModbusTempUnitK = "K"
)

// ModbusRegisterSpec defines how a single 16-bit Modbus register maps to an engineering value:
//
//	value = raw * Scale + Offset
//
// where raw is interpreted as int16 when Signed is true, or uint16 when Signed is false.
type ModbusRegisterSpec struct {
	Address uint16  `json:"address"`
	Scale   float64 `json:"scale"`
	Offset  float64 `json:"offset,omitempty"`
	Signed  bool    `json:"signed"`
}

// Decode converts a raw 16-bit Modbus register value into an engineering float64.
func (s ModbusRegisterSpec) Decode(raw uint16) float64 {
	if s.Signed {
		return float64(int16(raw))*s.Scale + s.Offset
	}
	return float64(raw)*s.Scale + s.Offset
}

// Validate checks that Scale and Offset are finite and Scale is non-zero.
func (s ModbusRegisterSpec) Validate(name string) error {
	if !isFinite(s.Scale) || s.Scale == 0 {
		return fmt.Errorf("%s.scale must be finite and non-zero, got %v", name, s.Scale)
	}
	if !isFinite(s.Offset) {
		return fmt.Errorf("%s.offset must be finite, got %v", name, s.Offset)
	}
	return nil
}

// ModbusRegisterMap configures the input sensor registers, analog output holding registers,
// and fail-safe digital coils on the studio PLC / RS-485 field I/O module.
type ModbusRegisterMap struct {
	// TempUnit specifies the engineering unit of QuenchInletTemp and QuenchOutletTemp registers:
	// "F" (default, degrees Fahrenheit), "C" (degrees Celsius), or "K" (Kelvin).
	TempUnit string `json:"temp_unit"`

	// Input sensor registers (read via FC04 by default, or FC03 if UseHoldingRegisters is true).
	DraftPa             ModbusRegisterSpec `json:"draft_pa"`
	QuenchInletTemp     ModbusRegisterSpec `json:"quench_inlet_temp"`
	QuenchOutletTemp    ModbusRegisterSpec `json:"quench_outlet_temp"`
	WaterFlowGPM        ModbusRegisterSpec `json:"water_flow_gpm"`
	SumpPH              ModbusRegisterSpec `json:"sump_ph"`
	BlowerSpeedFeedback ModbusRegisterSpec `json:"blower_speed_feedback"`

	// Analog output holding registers (0..OutputScale maps to 0.0..1.0 / 0-100%).
	BlowerVFDAddr        uint16 `json:"blower_vfd_addr"`
	QuenchSprayValveAddr uint16 `json:"quench_spray_valve_addr"`
	OutputScale          uint16 `json:"output_scale"`

	// Fail-safe digital output coils (written via FC05).
	BypassDamperCoilAddr uint16 `json:"bypass_damper_coil_addr"`
	AlarmLatchCoilAddr   uint16 `json:"alarm_latch_coil_addr"`
}

// ModbusConfig configures serial connection parameters, timeouts, and the register map for ModbusHAL.
type ModbusConfig struct {
	Device              string        `json:"device"`
	BaudRate            int           `json:"baud_rate"`
	SlaveID             uint8         `json:"slave_id"`
	Timeout             time.Duration `json:"timeout_ns"`
	UseHoldingRegisters bool          `json:"use_holding_registers"`
	// BypassCoilDeenergizeToTrip inverts BypassDamperCoilAddr polarity for fail-safe
	// de-energize-to-trip relay coils: when true, normal healthy operation writes ON (true)
	// to hold the spring-return damper closed, and a safety trip (BypassDamperOpen == true)
	// writes OFF (false) so coil de-energization or power loss opens the bypass damper.
	BypassCoilDeenergizeToTrip bool              `json:"bypass_coil_deenergize_to_trip,omitempty"`
	Registers                  ModbusRegisterMap `json:"registers"`
}

// DefaultModbusConfig returns standard RS-485 9600-8N1 defaults and a contiguous register map
// suitable for industrial PLC / Modbus RTU studio kiln I/O modules.
func DefaultModbusConfig() ModbusConfig {
	return ModbusConfig{
		Device:              "/dev/ttyUSB0",
		BaudRate:            9600,
		SlaveID:             1,
		Timeout:             500 * time.Millisecond,
		UseHoldingRegisters: false, // FC04 Input Registers for sensors
		Registers: ModbusRegisterMap{
			TempUnit: ModbusTempUnitF,
			DraftPa: ModbusRegisterSpec{
				Address: 0,
				Scale:   0.01, // -1245 -> -12.45 Pa (-0.05 inWC)
				Signed:  true,
			},
			QuenchInletTemp: ModbusRegisterSpec{
				Address: 1,
				Scale:   0.1, // 18000 -> 1800.0 F
				Signed:  true,
			},
			QuenchOutletTemp: ModbusRegisterSpec{
				Address: 2,
				Scale:   0.1, // 1880 -> 188.0 F
				Signed:  true,
			},
			WaterFlowGPM: ModbusRegisterSpec{
				Address: 3,
				Scale:   0.01, // 100 -> 1.00 GPM
				Signed:  false,
			},
			SumpPH: ModbusRegisterSpec{
				Address: 4,
				Scale:   0.01, // 800 -> 8.00 pH
				Signed:  false,
			},
			BlowerSpeedFeedback: ModbusRegisterSpec{
				Address: 5,
				Scale:   0.0001, // 4200 -> 0.42 (42.0%)
				Signed:  false,
			},
			BlowerVFDAddr:        100,
			QuenchSprayValveAddr: 101,
			OutputScale:          10000, // 0..10000 -> 0..100%
			BypassDamperCoilAddr: 0,
			AlarmLatchCoilAddr:   1,
		},
	}
}

// Validate checks serial settings, slave ID, timeout, and register scaling configuration.
func (c ModbusConfig) Validate() error {
	if strings.TrimSpace(c.Device) == "" {
		return fmt.Errorf("modbus.device must not be empty")
	}
	if !isSupportedBaudRate(c.BaudRate) {
		return fmt.Errorf("modbus.baud_rate %d is not supported", c.BaudRate)
	}
	if err := validateSlaveID(c.SlaveID); err != nil {
		return fmt.Errorf("modbus.slave_id: %w", err)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("modbus.timeout_ns must be > 0, got %v", c.Timeout)
	}
	if err := c.Registers.Validate(); err != nil {
		return fmt.Errorf("modbus.registers: %w", err)
	}
	return nil
}

// Validate checks all register specs, temperature unit, and output scale.
func (m ModbusRegisterMap) Validate() error {
	switch strings.ToUpper(strings.TrimSpace(m.TempUnit)) {
	case "", ModbusTempUnitF, ModbusTempUnitC, ModbusTempUnitK:
	default:
		return fmt.Errorf("temp_unit must be %q, %q, or %q, got %q",
			ModbusTempUnitF, ModbusTempUnitC, ModbusTempUnitK, m.TempUnit)
	}
	if err := m.DraftPa.Validate("draft_pa"); err != nil {
		return err
	}
	if err := m.QuenchInletTemp.Validate("quench_inlet_temp"); err != nil {
		return err
	}
	if err := m.QuenchOutletTemp.Validate("quench_outlet_temp"); err != nil {
		return err
	}
	if err := m.WaterFlowGPM.Validate("water_flow_gpm"); err != nil {
		return err
	}
	if err := m.SumpPH.Validate("sump_ph"); err != nil {
		return err
	}
	if err := m.BlowerSpeedFeedback.Validate("blower_speed_feedback"); err != nil {
		return err
	}
	if m.OutputScale == 0 {
		return fmt.Errorf("output_scale must be > 0")
	}
	if m.BlowerVFDAddr == m.QuenchSprayValveAddr {
		return fmt.Errorf("blower_vfd_addr and quench_spray_valve_addr must be distinct (both %d)", m.BlowerVFDAddr)
	}
	if m.BypassDamperCoilAddr == m.AlarmLatchCoilAddr {
		return fmt.Errorf("bypass_damper_coil_addr and alarm_latch_coil_addr must be distinct (both %d)", m.BypassDamperCoilAddr)
	}
	return nil
}

func isSupportedBaudRate(baud int) bool {
	switch baud {
	case 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400:
		return true
	default:
		return false
	}
}

// ModbusHAL implements model.HardwareHAL over Modbus RTU (RS-485 / serial or mock transport).
// Any CRC, framing, timeout, exception, or transport error in ReadSensors is returned directly
// so Controller latches model.TripSensorFault and drives the fail-safe bypass damper open.
type ModbusHAL struct {
	mu      sync.RWMutex
	cfg     ModbusConfig
	client  *ModbusRTUClient
	clockFn func() time.Time
}

var _ model.HardwareHAL = (*ModbusHAL)(nil)

// NewModbusHAL validates cfg and binds a ModbusHAL to an existing ModbusTransport.
func NewModbusHAL(cfg ModbusConfig, transport ModbusTransport) (*ModbusHAL, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid modbus config: %w", err)
	}
	if strings.TrimSpace(cfg.Registers.TempUnit) == "" {
		cfg.Registers.TempUnit = ModbusTempUnitF
	} else {
		cfg.Registers.TempUnit = strings.ToUpper(strings.TrimSpace(cfg.Registers.TempUnit))
	}
	client, err := NewModbusRTUClient(transport, cfg.Timeout)
	if err != nil {
		return nil, err
	}
	return &ModbusHAL{
		cfg:     cfg,
		client:  client,
		clockFn: time.Now,
	}, nil
}

// OpenModbusHAL opens the configured serial port and returns a ready-to-use ModbusHAL.
func OpenModbusHAL(cfg ModbusConfig) (*ModbusHAL, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid modbus config: %w", err)
	}
	transport, err := OpenSerialPort(cfg.Device, cfg.BaudRate)
	if err != nil {
		return nil, err
	}
	hal, err := NewModbusHAL(cfg, transport)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return hal, nil
}

// Close closes the underlying Modbus RTU transport.
func (h *ModbusHAL) Close() error {
	return h.client.Close()
}

// Config returns a copy of the ModbusConfig used by this HAL.
func (h *ModbusHAL) Config() ModbusConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg
}

// SetClock overrides the timestamp source for SensorReadings (useful for deterministic testing).
func (h *ModbusHAL) SetClock(fn func() time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if fn == nil {
		h.clockFn = time.Now
		return
	}
	h.clockFn = fn
}

// ReadSensors polls all configured sensor registers over Modbus RTU and scales them into
// model.SensorReadings. When all sensor addresses fit within a single 125-register window,
// they are fetched in one batched FC04 (or FC03) transaction; otherwise each register is polled
// individually. Any transport, CRC, timeout, or Modbus exception error is returned immediately.
func (h *ModbusHAL) ReadSensors(ctx context.Context) (model.SensorReadings, error) {
	h.mu.RLock()
	cfg := h.cfg
	clockFn := h.clockFn
	h.mu.RUnlock()

	regMap := cfg.Registers
	specs := []ModbusRegisterSpec{
		regMap.DraftPa,
		regMap.QuenchInletTemp,
		regMap.QuenchOutletTemp,
		regMap.WaterFlowGPM,
		regMap.SumpPH,
		regMap.BlowerSpeedFeedback,
	}

	minAddr := specs[0].Address
	maxAddr := specs[0].Address
	for _, sp := range specs[1:] {
		if sp.Address < minAddr {
			minAddr = sp.Address
		}
		if sp.Address > maxAddr {
			maxAddr = sp.Address
		}
	}

	readFn := h.client.ReadInputRegisters
	if cfg.UseHoldingRegisters {
		readFn = h.client.ReadHoldingRegisters
	}

	var rawVals [6]uint16
	span := uint32(maxAddr) - uint32(minAddr) + 1
	if span <= uint32(modbusMaxReadRegs) {
		regs, err := readFn(ctx, cfg.SlaveID, minAddr, uint16(span))
		if err != nil {
			return model.SensorReadings{}, err
		}
		for i, sp := range specs {
			rawVals[i] = regs[sp.Address-minAddr]
		}
	} else {
		for i, sp := range specs {
			regs, err := readFn(ctx, cfg.SlaveID, sp.Address, 1)
			if err != nil {
				return model.SensorReadings{}, err
			}
			rawVals[i] = regs[0]
		}
	}

	draftPa := model.Pascal(regMap.DraftPa.Decode(rawVals[0]))
	inletTempK := convertModbusTempToKelvin(regMap.QuenchInletTemp.Decode(rawVals[1]), regMap.TempUnit)
	outletTempK := convertModbusTempToKelvin(regMap.QuenchOutletTemp.Decode(rawVals[2]), regMap.TempUnit)
	waterFlowGPM := regMap.WaterFlowGPM.Decode(rawVals[3])
	sumpPH := regMap.SumpPH.Decode(rawVals[4])
	blowerSpeed := regMap.BlowerSpeedFeedback.Decode(rawVals[5])

	return model.SensorReadings{
		DraftPa:             draftPa,
		DraftInWC:           draftPa.InWC(),
		QuenchInletTemp:     inletTempK,
		QuenchOutletTemp:    outletTempK,
		WaterFlow:           model.GPMToCubicMetersPerSec(waterFlowGPM),
		WaterFlowGPM:        waterFlowGPM,
		SumpPH:              sumpPH,
		BlowerSpeedFraction: blowerSpeed,
		Timestamp:           clockFn(),
	}, nil
}

// ApplyCommands writes actuator commands to the PLC/VFD holding registers (0..OutputScale)
// and fail-safe coils (FC05). When BypassDamperOpen is true, fail-safe coils are driven first
// and all outputs are attempted even if an individual write fails, returning any joined errors.
func (h *ModbusHAL) ApplyCommands(ctx context.Context, cmd model.ActuatorCommands) error {
	h.mu.RLock()
	cfg := h.cfg
	h.mu.RUnlock()

	regMap := cfg.Registers
	vfdRaw := scaleUnitToRegister(cmd.BlowerVFD, regMap.OutputScale)
	valveRaw := scaleUnitToRegister(cmd.QuenchSprayValve, regMap.OutputScale)

	writeCoils := func() error {
		var errs []error
		bypassCoilVal := cmd.BypassDamperOpen
		if cfg.BypassCoilDeenergizeToTrip {
			// De-energize-to-trip fail-safe polarity: coil ON = healthy hold-closed;
			// coil OFF (trip or power loss) = spring-return damper fails OPEN.
			bypassCoilVal = !cmd.BypassDamperOpen
		}
		if err := h.client.WriteSingleCoil(ctx, cfg.SlaveID, regMap.BypassDamperCoilAddr, bypassCoilVal); err != nil {
			errs = append(errs, fmt.Errorf("bypass damper coil %d: %w", regMap.BypassDamperCoilAddr, err))
		}
		if err := h.client.WriteSingleCoil(ctx, cfg.SlaveID, regMap.AlarmLatchCoilAddr, cmd.AlarmLatched); err != nil {
			errs = append(errs, fmt.Errorf("alarm latch coil %d: %w", regMap.AlarmLatchCoilAddr, err))
		}
		return errors.Join(errs...)
	}

	writeAnalogs := func() error {
		// Use FC16 when blower VFD and quench spray valve registers are contiguous.
		if uint32(regMap.BlowerVFDAddr)+1 == uint32(regMap.QuenchSprayValveAddr) {
			if err := h.client.WriteMultipleRegisters(ctx, cfg.SlaveID, regMap.BlowerVFDAddr, []uint16{vfdRaw, valveRaw}); err != nil {
				return fmt.Errorf("analog outputs [%d..%d]: %w", regMap.BlowerVFDAddr, regMap.QuenchSprayValveAddr, err)
			}
			return nil
		}
		if uint32(regMap.QuenchSprayValveAddr)+1 == uint32(regMap.BlowerVFDAddr) {
			if err := h.client.WriteMultipleRegisters(ctx, cfg.SlaveID, regMap.QuenchSprayValveAddr, []uint16{valveRaw, vfdRaw}); err != nil {
				return fmt.Errorf("analog outputs [%d..%d]: %w", regMap.QuenchSprayValveAddr, regMap.BlowerVFDAddr, err)
			}
			return nil
		}
		var errs []error
		if err := h.client.WriteSingleRegister(ctx, cfg.SlaveID, regMap.BlowerVFDAddr, vfdRaw); err != nil {
			errs = append(errs, fmt.Errorf("blower VFD register %d: %w", regMap.BlowerVFDAddr, err))
		}
		if err := h.client.WriteSingleRegister(ctx, cfg.SlaveID, regMap.QuenchSprayValveAddr, valveRaw); err != nil {
			errs = append(errs, fmt.Errorf("quench spray valve register %d: %w", regMap.QuenchSprayValveAddr, err))
		}
		return errors.Join(errs...)
	}

	if cmd.BypassDamperOpen {
		// Fail-safe priority: energize/open bypass damper & alarm latch coils first, then zero VFD / open spray valve.
		coilErr := writeCoils()
		analogErr := writeAnalogs()
		return errors.Join(coilErr, analogErr)
	}

	if err := writeAnalogs(); err != nil {
		return err
	}
	return writeCoils()
}

func scaleUnitToRegister(fraction float64, fullScale uint16) uint16 {
	if !isFinite(fraction) || fraction <= 0 {
		return 0
	}
	if fraction >= 1.0 {
		return fullScale
	}
	scaled := math.Round(fraction * float64(fullScale))
	if scaled < 0 {
		return 0
	}
	if scaled > float64(fullScale) {
		return fullScale
	}
	return uint16(scaled)
}

func convertModbusTempToKelvin(val float64, unit string) model.Kelvin {
	switch unit {
	case ModbusTempUnitC:
		return model.CelsiusToKelvin(val)
	case ModbusTempUnitK:
		return model.Kelvin(val)
	default:
		return model.FahrenheitToKelvin(val)
	}
}
