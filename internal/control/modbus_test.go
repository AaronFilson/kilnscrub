package control

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// mockModbusSlave is an in-memory Modbus RTU slave implementing ModbusTransport for deterministic testing.
type mockModbusSlave struct {
	mu           sync.Mutex
	slaveID      uint8
	inputRegs    map[uint16]uint16
	holdingRegs  map[uint16]uint16
	coils        map[uint16]bool
	respBuf      bytes.Buffer
	deadline     time.Time
	closed       bool
	writeHistory []uint8 // recorded function codes executed

	// Fault injection knobs.
	corruptCRC    bool
	exceptionCode uint8 // non-zero -> reply with Modbus exception
	truncateTo    int   // >0 -> truncate response frame to this many bytes
	wrongSlaveID  uint8 // non-zero -> override response slave ID
	wrongFuncCode uint8 // non-zero -> override response function code
	forceTimeout  bool  // true -> sleep until deadline and return os.ErrDeadlineExceeded
	faultReads    bool  // when true, only apply faults to read function codes (FC03/FC04)
}

func newMockModbusSlave(slaveID uint8) *mockModbusSlave {
	return &mockModbusSlave{
		slaveID:     slaveID,
		inputRegs:   make(map[uint16]uint16),
		holdingRegs: make(map[uint16]uint16),
		coils:       make(map[uint16]bool),
	}
}

func (m *mockModbusSlave) SetDeadline(t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deadline = t
	return nil
}

func (m *mockModbusSlave) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockModbusSlave) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	m.respBuf.Reset()

	if len(p) < 4 || !verifyModbusCRC(p) {
		// Real Modbus RTU slaves silently discard frames with bad CRC.
		return len(p), nil
	}

	reqSlave := p[0]
	fc := p[1]
	if reqSlave != m.slaveID {
		return len(p), nil
	}

	isRead := fc == FuncReadHoldingRegisters || fc == FuncReadInputRegisters
	applyFault := !m.faultReads || isRead

	if applyFault && m.forceTimeout {
		// Leave respBuf empty; Read() will honor deadline and return timeout.
		return len(p), nil
	}

	if applyFault && m.exceptionCode != 0 {
		resp := appendModbusCRC([]byte{m.slaveID, fc | modbusExceptionBit, m.exceptionCode})
		m.respBuf.Write(resp)
		return len(p), nil
	}

	var resp []byte
	switch fc {
	case FuncReadHoldingRegisters, FuncReadInputRegisters:
		addr := binary.BigEndian.Uint16(p[2:4])
		qty := binary.BigEndian.Uint16(p[4:6])
		store := m.inputRegs
		if fc == FuncReadHoldingRegisters {
			store = m.holdingRegs
		}
		resp = make([]byte, 3+int(qty)*2)
		resp[0] = m.slaveID
		resp[1] = fc
		resp[2] = byte(qty * 2)
		for i := uint16(0); i < qty; i++ {
			binary.BigEndian.PutUint16(resp[3+2*i:5+2*i], store[addr+i])
		}

	case FuncWriteSingleCoil:
		addr := binary.BigEndian.Uint16(p[2:4])
		val := binary.BigEndian.Uint16(p[4:6])
		m.coils[addr] = val == modbusCoilOn
		m.writeHistory = append(m.writeHistory, fc)
		resp = append([]byte(nil), p[:6]...)

	case FuncWriteSingleRegister:
		addr := binary.BigEndian.Uint16(p[2:4])
		val := binary.BigEndian.Uint16(p[4:6])
		m.holdingRegs[addr] = val
		m.writeHistory = append(m.writeHistory, fc)
		resp = append([]byte(nil), p[:6]...)

	case FuncWriteMultipleRegisters:
		addr := binary.BigEndian.Uint16(p[2:4])
		qty := binary.BigEndian.Uint16(p[4:6])
		for i := uint16(0); i < qty; i++ {
			m.holdingRegs[addr+i] = binary.BigEndian.Uint16(p[7+2*i : 9+2*i])
		}
		m.writeHistory = append(m.writeHistory, fc)
		resp = append([]byte(nil), p[:6]...)

	default:
		resp = []byte{m.slaveID, fc | modbusExceptionBit, 0x01}
	}

	if applyFault && m.wrongSlaveID != 0 {
		resp[0] = m.wrongSlaveID
	}
	if applyFault && m.wrongFuncCode != 0 {
		resp[1] = m.wrongFuncCode
	}
	resp = appendModbusCRC(resp)
	if applyFault && m.corruptCRC {
		resp[len(resp)-1] ^= 0xFF
	}
	if applyFault && m.truncateTo > 0 && m.truncateTo < len(resp) {
		resp = resp[:m.truncateTo]
	}

	m.respBuf.Write(resp)
	return len(p), nil
}

func (m *mockModbusSlave) Read(p []byte) (int, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	if m.respBuf.Len() > 0 {
		n, err := m.respBuf.Read(p)
		m.mu.Unlock()
		return n, err
	}
	forceTimeout := m.forceTimeout
	deadline := m.deadline
	m.mu.Unlock()

	if forceTimeout {
		if !deadline.IsZero() {
			if wait := time.Until(deadline); wait > 0 {
				time.Sleep(wait)
			}
		}
		return 0, os.ErrDeadlineExceeded
	}
	return 0, io.EOF
}

// --- 1. CRC-16/Modbus Known Test Vectors ---

func TestModbusCRC16Vectors(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantCRC  uint16
		wantLoHi [2]byte
	}{
		{
			name:     "standard ASCII check string 123456789",
			data:     []byte("123456789"),
			wantCRC:  0x4B37,
			wantLoHi: [2]byte{0x37, 0x4B},
		},
		{
			name:     "Modbus RTU FC03 read 2 regs at 0x0000 slave 1",
			data:     []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x02},
			wantCRC:  0x0BC4,
			wantLoHi: [2]byte{0xC4, 0x0B},
		},
		{
			name:     "Modbus RTU FC05 write coil 0 ON slave 1",
			data:     []byte{0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
			wantCRC:  0x3A8C,
			wantLoHi: [2]byte{0x8C, 0x3A},
		},
		{
			name:     "Modbus RTU FC04 read 6 regs at 0x0000 slave 1",
			data:     []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x06},
			wantCRC:  0x0870,
			wantLoHi: [2]byte{0x70, 0x08},
		},
		{
			name:     "empty slice init value",
			data:     nil,
			wantCRC:  0xFFFF,
			wantLoHi: [2]byte{0xFF, 0xFF},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ModbusCRC16(tc.data)
			if got != tc.wantCRC {
				t.Fatalf("ModbusCRC16(%x) = 0x%04X, want 0x%04X", tc.data, got, tc.wantCRC)
			}
			framed := appendModbusCRC(append([]byte(nil), tc.data...))
			if framed[len(framed)-2] != tc.wantLoHi[0] || framed[len(framed)-1] != tc.wantLoHi[1] {
				t.Fatalf("wire CRC bytes = [%02X %02X], want [%02X %02X]",
					framed[len(framed)-2], framed[len(framed)-1], tc.wantLoHi[0], tc.wantLoHi[1])
			}
			if len(framed) >= 4 && !verifyModbusCRC(framed) {
				t.Fatalf("verifyModbusCRC failed for valid frame %x", framed)
			}
		})
	}
}

// --- 2. All Modbus Function Codes (FC03, FC04, FC05, FC06, FC16) ---

func TestModbusRTUClientFunctionCodes(t *testing.T) {
	slave := newMockModbusSlave(1)
	slave.inputRegs[0] = 0x1234
	slave.inputRegs[1] = 0xABCD
	slave.holdingRegs[10] = 555

	client, err := NewModbusRTUClient(slave, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("NewModbusRTUClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// FC04: Read Input Registers.
	inRegs, err := client.ReadInputRegisters(ctx, 1, 0, 2)
	if err != nil {
		t.Fatalf("ReadInputRegisters: %v", err)
	}
	if len(inRegs) != 2 || inRegs[0] != 0x1234 || inRegs[1] != 0xABCD {
		t.Fatalf("FC04 got %x, want [1234 abcd]", inRegs)
	}

	// FC03: Read Holding Registers.
	hRegs, err := client.ReadHoldingRegisters(ctx, 1, 10, 1)
	if err != nil {
		t.Fatalf("ReadHoldingRegisters: %v", err)
	}
	if len(hRegs) != 1 || hRegs[0] != 555 {
		t.Fatalf("FC03 got %v, want [555]", hRegs)
	}

	// FC05: Write Single Coil (ON then OFF).
	if err := client.WriteSingleCoil(ctx, 1, 0, true); err != nil {
		t.Fatalf("WriteSingleCoil(true): %v", err)
	}
	if !slave.coils[0] {
		t.Fatalf("expected coil 0 == true")
	}
	if err := client.WriteSingleCoil(ctx, 1, 0, false); err != nil {
		t.Fatalf("WriteSingleCoil(false): %v", err)
	}
	if slave.coils[0] {
		t.Fatalf("expected coil 0 == false")
	}

	// FC06: Write Single Register.
	if err := client.WriteSingleRegister(ctx, 1, 20, 4321); err != nil {
		t.Fatalf("WriteSingleRegister: %v", err)
	}
	if slave.holdingRegs[20] != 4321 {
		t.Fatalf("expected holding reg 20 == 4321, got %d", slave.holdingRegs[20])
	}

	// FC16: Write Multiple Registers.
	if err := client.WriteMultipleRegisters(ctx, 1, 100, []uint16{4500, 6200}); err != nil {
		t.Fatalf("WriteMultipleRegisters: %v", err)
	}
	if slave.holdingRegs[100] != 4500 || slave.holdingRegs[101] != 6200 {
		t.Fatalf("FC16 regs = [%d, %d], want [4500, 6200]", slave.holdingRegs[100], slave.holdingRegs[101])
	}
}

// --- 3. ModbusHAL Normal ReadSensors & ApplyCommands Round-Trips ---

func TestModbusHALReadSensorsAndApplyCommands(t *testing.T) {
	cfg := DefaultModbusConfig()
	slave := newMockModbusSlave(cfg.SlaveID)

	// Populate nominal kiln scrubber sensor registers:
	// DraftPa = -12.45 Pa -> signed int16(-1245) = 0xFB23
	negDraft := int16(-1245)
	slave.inputRegs[0] = uint16(negDraft)
	slave.inputRegs[1] = 18000 // 1800.0 F inlet
	slave.inputRegs[2] = 1650  // 165.0 F outlet
	slave.inputRegs[3] = 100   // 1.00 GPM
	slave.inputRegs[4] = 800   // 8.00 pH
	slave.inputRegs[5] = 4200  // 0.42 (42%) blower speed

	hal, err := NewModbusHAL(cfg, slave)
	if err != nil {
		t.Fatalf("NewModbusHAL: %v", err)
	}
	defer hal.Close()

	fixedNow := time.Date(2026, 11, 22, 18, 0, 0, 0, time.UTC)
	hal.SetClock(func() time.Time { return fixedNow })

	ctx := context.Background()
	readings, err := hal.ReadSensors(ctx)
	if err != nil {
		t.Fatalf("ReadSensors: %v", err)
	}

	if math.Abs(float64(readings.DraftPa)-(-12.45)) > 1e-6 {
		t.Fatalf("DraftPa = %.4f, want -12.45", readings.DraftPa)
	}
	if math.Abs(readings.DraftInWC-model.Pascal(-12.45).InWC()) > 1e-6 {
		t.Fatalf("DraftInWC mismatch: got %.6f", readings.DraftInWC)
	}
	if math.Abs(readings.QuenchInletTemp.Fahrenheit()-1800.0) > 1e-4 {
		t.Fatalf("QuenchInletTemp = %.2f F, want 1800.0 F", readings.QuenchInletTemp.Fahrenheit())
	}
	if math.Abs(readings.QuenchOutletTemp.Fahrenheit()-165.0) > 1e-4 {
		t.Fatalf("QuenchOutletTemp = %.2f F, want 165.0 F", readings.QuenchOutletTemp.Fahrenheit())
	}
	if math.Abs(readings.WaterFlowGPM-1.00) > 1e-6 {
		t.Fatalf("WaterFlowGPM = %.4f, want 1.00", readings.WaterFlowGPM)
	}
	if math.Abs(readings.WaterFlow.GPM()-1.00) > 1e-6 {
		t.Fatalf("WaterFlow SI = %.4f GPM, want 1.00", readings.WaterFlow.GPM())
	}
	if math.Abs(readings.SumpPH-8.00) > 1e-6 {
		t.Fatalf("SumpPH = %.4f, want 8.00", readings.SumpPH)
	}
	if math.Abs(readings.BlowerSpeedFraction-0.42) > 1e-6 {
		t.Fatalf("BlowerSpeedFraction = %.4f, want 0.42", readings.BlowerSpeedFraction)
	}
	if !readings.Timestamp.Equal(fixedNow) {
		t.Fatalf("Timestamp = %v, want %v", readings.Timestamp, fixedNow)
	}

	// Normal ApplyCommands (contiguous VFD=100, Valve=101 -> FC16 + FC05 coils).
	cmd := model.ActuatorCommands{
		BlowerVFD:        0.55,
		QuenchSprayValve: 0.725,
		BypassDamperOpen: false,
		AlarmLatched:     false,
	}
	if err := hal.ApplyCommands(ctx, cmd); err != nil {
		t.Fatalf("ApplyCommands(normal): %v", err)
	}
	if slave.holdingRegs[100] != 5500 || slave.holdingRegs[101] != 7250 {
		t.Fatalf("analog outputs = [%d, %d], want [5500, 7250]", slave.holdingRegs[100], slave.holdingRegs[101])
	}
	if slave.coils[0] || slave.coils[1] {
		t.Fatalf("expected fail-safe coils OFF in normal mode, got damper=%v alarm=%v", slave.coils[0], slave.coils[1])
	}

	// Tripped ApplyCommands -> fail-safe coils MUST be written BEFORE analog outputs.
	slave.mu.Lock()
	slave.writeHistory = nil
	slave.mu.Unlock()

	tripCmd := model.ActuatorCommands{
		BlowerVFD:        0.0,
		QuenchSprayValve: 1.0,
		BypassDamperOpen: true,
		AlarmLatched:     true,
		TripReason:       model.TripHighQuenchTemp,
	}
	if err := hal.ApplyCommands(ctx, tripCmd); err != nil {
		t.Fatalf("ApplyCommands(tripped): %v", err)
	}
	if !slave.coils[0] || !slave.coils[1] {
		t.Fatalf("expected bypass damper and alarm latch coils ON when tripped")
	}
	if slave.holdingRegs[100] != 0 || slave.holdingRegs[101] != 10000 {
		t.Fatalf("tripped analog outputs = [%d, %d], want [0, 10000]", slave.holdingRegs[100], slave.holdingRegs[101])
	}
	if len(slave.writeHistory) < 3 || slave.writeHistory[0] != FuncWriteSingleCoil || slave.writeHistory[1] != FuncWriteSingleCoil {
		t.Fatalf("expected fail-safe coils (FC05) written first on trip, got history %x", slave.writeHistory)
	}
}

func TestModbusHALNonContiguousAndCelsiusMode(t *testing.T) {
	cfg := DefaultModbusConfig()
	cfg.UseHoldingRegisters = true
	cfg.Registers.TempUnit = ModbusTempUnitC
	cfg.Registers.DraftPa.Address = 10
	cfg.Registers.QuenchInletTemp.Address = 200 // > 125 span -> individual reads
	cfg.Registers.QuenchOutletTemp.Address = 201
	cfg.Registers.WaterFlowGPM.Address = 300
	cfg.Registers.SumpPH.Address = 301
	cfg.Registers.BlowerSpeedFeedback.Address = 302
	cfg.Registers.BlowerVFDAddr = 400
	cfg.Registers.QuenchSprayValveAddr = 410 // non-contiguous -> FC06 writes

	slave := newMockModbusSlave(cfg.SlaveID)
	negDraft := int16(-1500)
	slave.holdingRegs[10] = uint16(negDraft)
	slave.holdingRegs[200] = 10000 // 1000.0 C
	slave.holdingRegs[201] = 750   // 75.0 C
	slave.holdingRegs[300] = 120   // 1.20 GPM
	slave.holdingRegs[301] = 780   // 7.80 pH
	slave.holdingRegs[302] = 5000  // 0.50

	hal, err := NewModbusHAL(cfg, slave)
	if err != nil {
		t.Fatalf("NewModbusHAL: %v", err)
	}
	defer hal.Close()

	readings, err := hal.ReadSensors(context.Background())
	if err != nil {
		t.Fatalf("ReadSensors: %v", err)
	}
	if math.Abs(readings.QuenchOutletTemp.Celsius()-75.0) > 1e-4 {
		t.Fatalf("QuenchOutletTemp = %.2f C, want 75.0 C", readings.QuenchOutletTemp.Celsius())
	}

	if err := hal.ApplyCommands(context.Background(), model.ActuatorCommands{
		BlowerVFD:        0.6,
		QuenchSprayValve: 0.4,
	}); err != nil {
		t.Fatalf("ApplyCommands(non-contiguous FC06): %v", err)
	}
	if slave.holdingRegs[400] != 6000 || slave.holdingRegs[410] != 4000 {
		t.Fatalf("FC06 outputs = [%d, %d], want [6000, 4000]", slave.holdingRegs[400], slave.holdingRegs[410])
	}
}

// --- 4. Modbus Faults -> Controller Latches TripSensorFault & Opens Bypass Damper ---

func TestModbusHALFaultsLatchTripSensorFault(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(s *mockModbusSlave)
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "corrupted CRC",
			setup: func(s *mockModbusSlave) {
				s.corruptCRC = true
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, ErrModbusCRC) {
					t.Fatalf("expected ErrModbusCRC, got %v", err)
				}
			},
		},
		{
			name: "Modbus exception 0x04 slave device failure",
			setup: func(s *mockModbusSlave) {
				s.exceptionCode = 0x04
			},
			checkErr: func(t *testing.T, err error) {
				var exc *ModbusExceptionError
				if !errors.As(err, &exc) || exc.ExceptionCode != 0x04 || exc.FunctionCode != FuncReadInputRegisters {
					t.Fatalf("expected ModbusExceptionError(0x04, FC04), got %v", err)
				}
			},
		},
		{
			name: "truncated frame",
			setup: func(s *mockModbusSlave) {
				s.truncateTo = 4
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, ErrModbusFrameTooShort) {
					t.Fatalf("expected ErrModbusFrameTooShort, got %v", err)
				}
			},
		},
		{
			name: "slave timeout",
			setup: func(s *mockModbusSlave) {
				s.forceTimeout = true
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, ErrModbusTimeout) {
					t.Fatalf("expected ErrModbusTimeout, got %v", err)
				}
			},
		},
		{
			name: "wrong slave ID",
			setup: func(s *mockModbusSlave) {
				s.wrongSlaveID = 9
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, ErrModbusSlaveMismatch) {
					t.Fatalf("expected ErrModbusSlaveMismatch, got %v", err)
				}
			},
		},
		{
			name: "wrong function code",
			setup: func(s *mockModbusSlave) {
				s.wrongFuncCode = FuncReadHoldingRegisters
			},
			checkErr: func(t *testing.T, err error) {
				if !errors.Is(err, ErrModbusFunctionMismatch) {
					t.Fatalf("expected ErrModbusFunctionMismatch, got %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mbCfg := DefaultModbusConfig()
			mbCfg.Timeout = 40 * time.Millisecond

			slave := newMockModbusSlave(mbCfg.SlaveID)
			// Nominal valid sensor registers (used once fault clears).
			negDraft := int16(-1245)
			slave.inputRegs[0] = uint16(negDraft)
			slave.inputRegs[1] = 18000
			slave.inputRegs[2] = 1650
			slave.inputRegs[3] = 100
			slave.inputRegs[4] = 800
			slave.inputRegs[5] = 4200
			// Fault only reads so we also verify Controller drives fail-safe coils onto the PLC!
			slave.faultReads = true
			tc.setup(slave)

			hal, err := NewModbusHAL(mbCfg, slave)
			if err != nil {
				t.Fatalf("NewModbusHAL: %v", err)
			}
			defer hal.Close()

			ctrlCfg := model.DefaultControlConfig()
			ctrl, err := NewController(ctrlCfg)
			if err != nil {
				t.Fatalf("NewController: %v", err)
			}

			status, tickErr := ctrl.Tick(context.Background(), hal, ctrlCfg.LoopInterval)
			if tickErr == nil {
				t.Fatalf("expected error from Controller.Tick on %s", tc.name)
			}
			tc.checkErr(t, tickErr)

			if !status.InterlockActive || !status.Commands.BypassDamperOpen || !status.Commands.AlarmLatched {
				t.Fatalf("expected InterlockActive & BypassDamperOpen & AlarmLatched on %s, got %+v", tc.name, status)
			}
			if status.Commands.TripReason != model.TripSensorFault {
				t.Fatalf("TripReason = %q, want %q", status.Commands.TripReason, model.TripSensorFault)
			}
			if status.Commands.BlowerVFD != 0.0 || status.Commands.QuenchSprayValve != ctrlCfg.QuenchPID.MaxOutput {
				t.Fatalf("unexpected fail-safe outputs on %s: %+v", tc.name, status.Commands)
			}
			// Verify fail-safe commands were actually written to the Modbus coils & holding registers!
			if !slave.coils[mbCfg.Registers.BypassDamperCoilAddr] || !slave.coils[mbCfg.Registers.AlarmLatchCoilAddr] {
				t.Fatalf("expected PLC fail-safe coils energized after %s", tc.name)
			}

			// Clear bus fault, execute healthy tick, verify interlock stays latched until explicit ResetInterlock.
			slave.mu.Lock()
			slave.corruptCRC = false
			slave.exceptionCode = 0
			slave.truncateTo = 0
			slave.forceTimeout = false
			slave.wrongSlaveID = 0
			slave.wrongFuncCode = 0
			slave.mu.Unlock()

			stAfterClear, err := ctrl.Tick(context.Background(), hal, ctrlCfg.LoopInterval)
			if err != nil {
				t.Fatalf("healthy recovery Tick failed: %v", err)
			}
			if !stAfterClear.InterlockActive || !stAfterClear.Commands.BypassDamperOpen {
				t.Fatalf("interlock must remain latched after bus recovery until explicit ResetInterlock()")
			}
			if err := ctrl.ResetInterlock(); err != nil {
				t.Fatalf("ResetInterlock failed after bus recovery: %v", err)
			}
			if ctrl.Status().InterlockActive {
				t.Fatalf("expected InterlockActive == false after ResetInterlock")
			}
		})
	}
}

func TestModbusHALDeenergizeToTripPolarity(t *testing.T) {
	cfg := DefaultModbusConfig()
	cfg.BypassCoilDeenergizeToTrip = true
	slave := newMockModbusSlave(cfg.SlaveID)
	negDraft := int16(-1245)
	slave.inputRegs[0] = uint16(negDraft)
	slave.inputRegs[1] = 18000
	slave.inputRegs[2] = 1650
	slave.inputRegs[3] = 100
	slave.inputRegs[4] = 800
	slave.inputRegs[5] = 4200

	hal, err := NewModbusHAL(cfg, slave)
	if err != nil {
		t.Fatalf("NewModbusHAL: %v", err)
	}
	defer hal.Close()

	ctx := context.Background()

	// Normal healthy command: BypassDamperOpen=false -> coil MUST be energized (true) to hold spring-return damper closed.
	if err := hal.ApplyCommands(ctx, model.ActuatorCommands{
		BlowerVFD:        0.5,
		QuenchSprayValve: 0.6,
		BypassDamperOpen: false,
		AlarmLatched:     false,
	}); err != nil {
		t.Fatalf("ApplyCommands normal: %v", err)
	}
	if !slave.coils[cfg.Registers.BypassDamperCoilAddr] {
		t.Fatal("expected bypass damper coil ON (energized hold-closed) during normal operation when BypassCoilDeenergizeToTrip=true")
	}
	if slave.coils[cfg.Registers.AlarmLatchCoilAddr] {
		t.Fatal("expected alarm latch coil OFF during normal operation")
	}

	// Safety trip command: BypassDamperOpen=true -> coil MUST be de-energized (false) first, so spring opens damper.
	slave.mu.Lock()
	slave.writeHistory = nil
	slave.mu.Unlock()
	if err := hal.ApplyCommands(ctx, model.ActuatorCommands{
		BlowerVFD:        0.0,
		QuenchSprayValve: 1.0,
		BypassDamperOpen: true,
		AlarmLatched:     true,
		TripReason:       model.TripHighQuenchTemp,
	}); err != nil {
		t.Fatalf("ApplyCommands tripped: %v", err)
	}
	if slave.coils[cfg.Registers.BypassDamperCoilAddr] {
		t.Fatal("expected bypass damper coil OFF (de-energized fail-open) when BypassDamperOpen=true")
	}
	if !slave.coils[cfg.Registers.AlarmLatchCoilAddr] {
		t.Fatal("expected alarm latch coil ON when tripped")
	}
	if len(slave.writeHistory) < 3 || slave.writeHistory[0] != FuncWriteSingleCoil {
		t.Fatalf("expected fail-safe coil de-energized first on trip, got history %x", slave.writeHistory)
	}

	// Verify Controller fault triggers de-energization on the PLC coil.
	slave.mu.Lock()
	slave.coils[cfg.Registers.BypassDamperCoilAddr] = true
	slave.faultReads = true
	slave.corruptCRC = true
	slave.mu.Unlock()

	ctrl, err := NewController(model.DefaultControlConfig())
	if err != nil {
		t.Fatalf("NewController: %v", err)
	}
	st, tickErr := ctrl.Tick(ctx, hal, 100*time.Millisecond)
	if tickErr == nil || !st.InterlockActive {
		t.Fatalf("expected fault trip, got st=%+v err=%v", st, tickErr)
	}
	if slave.coils[cfg.Registers.BypassDamperCoilAddr] {
		t.Fatal("expected bypass coil de-energized (false) after Controller sensor fault")
	}
}
