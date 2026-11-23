package control

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

type stubCustomHAL struct {
	readErr  error
	applyErr error
}

func (s *stubCustomHAL) ReadSensors(ctx context.Context) (model.SensorReadings, error) {
	if s.readErr != nil {
		return model.SensorReadings{}, s.readErr
	}
	return model.SensorReadings{
		DraftPa:             model.InWCToPascal(-0.05),
		DraftInWC:           -0.05,
		QuenchInletTemp:     model.FahrenheitToKelvin(1800),
		QuenchOutletTemp:    model.FahrenheitToKelvin(188),
		WaterFlow:           model.GPMToCubicMetersPerSec(1.0),
		WaterFlowGPM:        1.0,
		SumpPH:              8.0,
		BlowerSpeedFraction: 0.5,
		Timestamp:           time.Now(),
	}, nil
}

func (s *stubCustomHAL) ApplyCommands(ctx context.Context, cmd model.ActuatorCommands) error {
	return s.applyErr
}

func TestModbusExceptionNamesAndClientGuards(t *testing.T) {
	codes := []uint8{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x08, 0x0A, 0x0B, 0x7F}
	for _, code := range codes {
		exc := &ModbusExceptionError{SlaveID: 1, FunctionCode: FuncReadInputRegisters, ExceptionCode: code}
		if !strings.Contains(exc.Error(), "modbus: slave 1") {
			t.Errorf("unexpected error string for 0x%02X: %s", code, exc.Error())
		}
	}

	if _, err := NewModbusRTUClient(nil, 0); err == nil {
		t.Error("expected error for nil transport")
	}

	slave := newMockModbusSlave(1)
	client, err := NewModbusRTUClient(slave, 0) // default timeout branch
	if err != nil {
		t.Fatalf("NewModbusRTUClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if _, err := client.ReadInputRegisters(ctx, 0, 0, 1); err == nil {
		t.Error("expected invalid slaveID=0 error")
	}
	if _, err := client.ReadInputRegisters(ctx, 248, 0, 1); err == nil {
		t.Error("expected invalid slaveID=248 error")
	}
	if _, err := client.ReadInputRegisters(ctx, 1, 0, 0); err == nil {
		t.Error("expected quantity=0 error")
	}
	if _, err := client.ReadInputRegisters(ctx, 1, 0, 126); err == nil {
		t.Error("expected quantity=126 error")
	}
	if err := client.WriteSingleCoil(ctx, 0, 0, true); err == nil {
		t.Error("expected invalid slaveID on FC05")
	}
	if err := client.WriteSingleRegister(ctx, 0, 0, 1); err == nil {
		t.Error("expected invalid slaveID on FC06")
	}
	if err := client.WriteMultipleRegisters(ctx, 0, 0, []uint16{1}); err == nil {
		t.Error("expected invalid slaveID on FC16")
	}
	if err := client.WriteMultipleRegisters(ctx, 1, 0, nil); err == nil {
		t.Error("expected empty slice error on FC16")
	}
	if err := client.WriteMultipleRegisters(ctx, 1, 0, make([]uint16, 124)); err == nil {
		t.Error("expected >123 regs error on FC16")
	}
	if verifyModbusCRC([]byte{1, 2}) {
		t.Error("expected short frame CRC false")
	}
}

func TestModbusHALEdgeBranchesAndValidation(t *testing.T) {
	// Kelvin unit + reversed contiguous analog outputs (QuenchSprayValveAddr+1 == BlowerVFDAddr) + empty TempUnit + SetClock(nil).
	cfg := DefaultModbusConfig()
	cfg.Registers.TempUnit = ModbusTempUnitK
	cfg.Registers.QuenchSprayValveAddr = 200
	cfg.Registers.BlowerVFDAddr = 201
	slave := newMockModbusSlave(cfg.SlaveID)
	slave.inputRegs[0] = 0
	slave.inputRegs[1] = 15000 // 1500.0 K
	slave.inputRegs[2] = 3580  // 358.0 K
	slave.inputRegs[3] = 100
	slave.inputRegs[4] = 800
	slave.inputRegs[5] = 4000

	hal, err := NewModbusHAL(cfg, slave)
	if err != nil {
		t.Fatalf("NewModbusHAL: %v", err)
	}
	hal.SetClock(nil)
	if hal.Config().SlaveID != cfg.SlaveID {
		t.Error("Config() mismatch")
	}
	r, err := hal.ReadSensors(context.Background())
	if err != nil {
		t.Fatalf("ReadSensors Kelvin: %v", err)
	}
	if math.Abs(float64(r.QuenchOutletTemp)-358.0) > 1e-3 {
		t.Errorf("QuenchOutletTemp = %v K, want 358 K", r.QuenchOutletTemp)
	}
	if err := hal.ApplyCommands(context.Background(), model.ActuatorCommands{
		BlowerVFD:        1.5,        // clamp >1
		QuenchSprayValve: math.NaN(), // clamp NaN -> 0
	}); err != nil {
		t.Fatalf("ApplyCommands reversed contiguous: %v", err)
	}
	if slave.holdingRegs[200] != 0 || slave.holdingRegs[201] != cfg.Registers.OutputScale {
		t.Errorf("reversed contiguous outputs = [%d, %d]", slave.holdingRegs[200], slave.holdingRegs[201])
	}
	_ = hal.Close()

	// Empty TempUnit defaults to F.
	cfgEmpty := DefaultModbusConfig()
	cfgEmpty.Registers.TempUnit = ""
	hal2, err := NewModbusHAL(cfgEmpty, newMockModbusSlave(1))
	if err != nil || hal2.Config().Registers.TempUnit != ModbusTempUnitF {
		t.Errorf("expected default TempUnit F, got %q, err=%v", hal2.Config().Registers.TempUnit, err)
	}

	// Validate all ModbusConfig & ModbusRegisterMap error branches.
	badCfgs := []func(*ModbusConfig){
		func(c *ModbusConfig) { c.Device = " " },
		func(c *ModbusConfig) { c.BaudRate = 12345 },
		func(c *ModbusConfig) { c.SlaveID = 0 },
		func(c *ModbusConfig) { c.Timeout = 0 },
		func(c *ModbusConfig) { c.Registers.TempUnit = "R" },
		func(c *ModbusConfig) { c.Registers.DraftPa.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.DraftPa.Offset = math.NaN() },
		func(c *ModbusConfig) { c.Registers.QuenchInletTemp.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.QuenchOutletTemp.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.WaterFlowGPM.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.SumpPH.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.BlowerSpeedFeedback.Scale = 0 },
		func(c *ModbusConfig) { c.Registers.OutputScale = 0 },
		func(c *ModbusConfig) { c.Registers.BlowerVFDAddr = c.Registers.QuenchSprayValveAddr },
		func(c *ModbusConfig) { c.Registers.BypassDamperCoilAddr = c.Registers.AlarmLatchCoilAddr },
	}
	for i, mutate := range badCfgs {
		c := DefaultModbusConfig()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("badCfgs[%d]: expected validation error", i)
		}
		if _, err := NewModbusHAL(c, slave); err == nil {
			t.Errorf("badCfgs[%d]: NewModbusHAL expected error", i)
		}
	}
}

func TestDaemonHubInterlockAndConfigBranches(t *testing.T) {
	// Custom HAL type-switch default + callbacks + error handling + unsupported fault injection.
	if _, err := NewDaemon(model.DefaultControlConfig(), nil); err == nil {
		t.Error("NewDaemon(nil) expected error")
	}

	stub := &stubCustomHAL{readErr: errors.New("boom")}
	var gotErr error
	var gotTel int
	cfg := model.DefaultControlConfig()
	cfg.LoopInterval = 15 * time.Millisecond
	d, err := NewDaemon(cfg, stub,
		WithHALMode("custom-plc"),
		WithErrorCallback(func(err error) { gotErr = err }),
		WithTelemetryCallback(func(model.ControlStatus) { gotTel++ }),
	)
	if err != nil {
		t.Fatalf("NewDaemon: %v", err)
	}
	if d.HALMode() != "custom-plc" {
		t.Errorf("HALMode = %q, want custom-plc", d.HALMode())
	}
	if d.SupportsFaultInjection() {
		t.Error("stubCustomHAL should not support fault injection")
	}
	if !errors.Is(d.InjectFault("water_loss", 1), ErrFaultInjectionUnsupported) {
		t.Error("expected ErrFaultInjectionUnsupported")
	}
	if d.Hub() == nil || d.Controller() == nil || d.Running() {
		t.Error("unexpected initial Daemon accessors")
	}
	if err := d.SetTargetDraftInWC(-0.08); err != nil {
		t.Errorf("SetTargetDraftInWC: %v", err)
	}
	if err := d.SetQuenchSetpointF(190.0); err != nil {
		t.Errorf("SetQuenchSetpointF: %v", err)
	}
	newDraft := -0.06
	if _, err := d.SetSetpoints(&newDraft, nil); err != nil {
		t.Errorf("SetSetpoints: %v", err)
	}

	// Pre-canceled context returns nil immediately.
	canceledCtx, cancel0 := context.WithCancel(context.Background())
	cancel0()
	if err := d.Run(canceledCtx); err != nil {
		t.Errorf("Run(canceled) = %v", err)
	}

	// Active run invoking error & telemetry callbacks + double-Run guard.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()
	time.Sleep(10 * time.Millisecond)
	if d.Running() {
		if err := d.Run(context.Background()); err == nil {
			t.Error("expected concurrent Run error")
		}
	}
	if err := <-done; err != nil {
		t.Errorf("Run returned %v", err)
	}
	if gotErr == nil || gotTel == 0 || d.Ticks() == 0 {
		t.Errorf("expected callbacks & ticks, gotErr=%v gotTel=%d ticks=%d", gotErr, gotTel, d.Ticks())
	}

	// Set callbacks via setters and test ResetInterlock failure while fault active.
	d.SetErrorCallback(nil)
	d.SetTelemetryCallback(nil)
	if err := d.ResetInterlock(); err == nil {
		t.Error("expected ResetInterlock error while fault active")
	}

	// TelemetryHub edge branches: default buffer size, eviction, closed hub.
	hub := NewTelemetryHub(0)
	if hub.Subscribers() != 0 {
		t.Error("expected 0 subscribers")
	}
	subCh, unsub := hub.Subscribe()
	for i := 0; i < 30; i++ {
		hub.Publish(model.ControlStatus{State: "NORMAL"})
	}
	if hub.Dropped() == 0 {
		t.Error("expected dropped frames when buffer overflows")
	}
	unsub()
	unsub() // idempotent
	<-subCh
	hub.Close()
	hub.Close() // idempotent
	hub.Publish(model.ControlStatus{State: "NORMAL"})
	closedCh, unsubClosed := hub.Subscribe()
	unsubClosed()
	if _, ok := <-closedCh; ok {
		t.Error("expected closed channel when subscribing to closed hub")
	}

	// Interlock edge branches.
	badLim := model.DefaultSafetyLimits()
	badLim.SensorFaultTimeout = 0
	if _, err := NewInterlock(badLim); err == nil {
		t.Error("NewInterlock bad limits expected error")
	}
	il, err := NewInterlock(model.DefaultSafetyLimits())
	if err != nil {
		t.Fatalf("NewInterlock: %v", err)
	}
	if il.Limits().MinSumpPH != 6.5 {
		t.Error("Limits() mismatch")
	}
	if err := il.Reset(); err != nil {
		t.Errorf("Reset when unlatched should succeed: %v", err)
	}
	// Evaluate with zero timestamp + high pH warning.
	st := il.Evaluate(model.SensorReadings{
		DraftPa:          -12,
		QuenchInletTemp:  1200,
		QuenchOutletTemp: 355,
		WaterFlow:        model.GPMToCubicMetersPerSec(1.0),
		WaterFlowGPM:     1.0,
		SumpPH:           10.5,
		Timestamp:        time.Now(),
	}, nil, time.Time{})
	if !st.PHWarning || st.BypassDamperOpen {
		t.Errorf("expected high pH warning without bypass, got %+v", st)
	}

	// DaemonConfig & SimHALConfig validation branches.
	if _, err := ParseDaemonConfig([]byte(`{"hal_mode":""}`)); err != nil {
		t.Errorf("ParseDaemonConfig empty hal_mode should default to sim: %v", err)
	}
	if _, err := ParseDaemonConfig([]byte(`{bad`)); err == nil {
		t.Error("ParseDaemonConfig bad JSON expected error")
	}
	if _, err := LoadDaemonConfig(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("LoadDaemonConfig missing expected error")
	}
	badSim := []func(*SimHALConfig){
		func(s *SimHALConfig) { s.NominalInletTemp = 100 },
		func(s *SimHALConfig) { s.BaselineQuenchTemp = 200 },
		func(s *SimHALConfig) { s.MaxWaterFlowGPM = 0 },
		func(s *SimHALConfig) { s.NominalWaterFlowGPM = 100 },
		func(s *SimHALConfig) { s.BlowerMaxSuctionPa = 0 },
		func(s *SimHALConfig) { s.BlowerTimeConst = 0 },
		func(s *SimHALConfig) { s.DryGasOutletTemp = 300 },
		func(s *SimHALConfig) { s.InitialSumpPH = 20 },
		func(s *SimHALConfig) { s.DraftNoiseAmplitudePa = -1 },
		func(s *SimHALConfig) { s.DraftNoiseFreqHz = -1 },
		func(s *SimHALConfig) { s.AutoStepInterval = -1 },
	}
	for i, mutate := range badSim {
		dc := DefaultDaemonConfig()
		mutate(&dc.Sim)
		if err := dc.Validate(); err == nil {
			t.Errorf("badSim[%d]: expected validation error", i)
		}
	}
	badDC := DefaultDaemonConfig()
	badDC.HALMode = "unknown"
	if err := badDC.Validate(); err == nil {
		t.Error("expected unknown hal_mode error")
	}
	badMB := DefaultDaemonConfig()
	badMB.HALMode = HALModeModbus
	badMB.Modbus.Device = ""
	if err := badMB.Validate(); err == nil {
		t.Error("expected invalid modbus config error")
	}
}
