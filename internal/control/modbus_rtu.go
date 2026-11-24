package control

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// Standard Modbus RTU function codes implemented by ModbusRTUClient.
const (
	FuncReadHoldingRegisters   uint8 = 0x03
	FuncReadInputRegisters     uint8 = 0x04
	FuncWriteSingleCoil        uint8 = 0x05
	FuncWriteSingleRegister    uint8 = 0x06
	FuncWriteMultipleRegisters uint8 = 0x10
)

// Standard Modbus RTU limits and framing constants.
const (
	modbusMinSlaveID      uint8  = 1
	modbusMaxSlaveID      uint8  = 247
	modbusMaxReadRegs     uint16 = 125
	modbusMaxWriteRegs    uint16 = 123
	modbusCoilOn          uint16 = 0xFF00
	modbusCoilOff         uint16 = 0x0000
	modbusExceptionBit    uint8  = 0x80
	modbusExceptionLength        = 5
)

// Sentinel errors returned by ModbusRTUClient frame validation and I/O.
var (
	ErrModbusCRC              = errors.New("modbus: CRC-16 mismatch")
	ErrModbusFrameTooShort    = errors.New("modbus: truncated response frame")
	ErrModbusSlaveMismatch    = errors.New("modbus: response slave ID mismatch")
	ErrModbusFunctionMismatch = errors.New("modbus: response function code mismatch")
	ErrModbusUnexpectedLength = errors.New("modbus: unexpected response byte count")
	ErrModbusEchoMismatch     = errors.New("modbus: write response echo mismatch")
	ErrModbusTimeout          = errors.New("modbus: request timed out")
)

// ModbusExceptionError represents a standard Modbus exception response (FC | 0x80).
type ModbusExceptionError struct {
	SlaveID       uint8
	FunctionCode  uint8
	ExceptionCode uint8
}

// Error formats the Modbus exception code with its standard protocol meaning.
func (e *ModbusExceptionError) Error() string {
	return fmt.Sprintf("modbus: slave %d exception for FC 0x%02X: 0x%02X (%s)",
		e.SlaveID, e.FunctionCode, e.ExceptionCode, modbusExceptionName(e.ExceptionCode))
}

func modbusExceptionName(code uint8) string {
	switch code {
	case 0x01:
		return "illegal function"
	case 0x02:
		return "illegal data address"
	case 0x03:
		return "illegal data value"
	case 0x04:
		return "slave device failure"
	case 0x05:
		return "acknowledge"
	case 0x06:
		return "slave device busy"
	case 0x08:
		return "memory parity error"
	case 0x0A:
		return "gateway path unavailable"
	case 0x0B:
		return "gateway target device failed to respond"
	default:
		return "unknown exception"
	}
}

// ModbusTransport abstracts a byte-stream RTU link (serial port or in-memory test mock).
// Implementations must respect deadlines passed to SetDeadline.
type ModbusTransport interface {
	io.ReadWriteCloser
	SetDeadline(t time.Time) error
}

// ModbusCRC16 computes the standard CRC-16/Modbus checksum (init 0xFFFF, reflected polynomial 0xA001).
// On the wire, Modbus RTU appends the low byte first, followed by the high byte.
func ModbusCRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// appendModbusCRC appends the little-endian CRC-16/Modbus checksum to frame.
func appendModbusCRC(frame []byte) []byte {
	crc := ModbusCRC16(frame)
	return append(frame, byte(crc&0x00FF), byte(crc>>8))
}

// verifyModbusCRC checks that frame has at least 4 bytes and ends with a valid CRC-16/Modbus.
func verifyModbusCRC(frame []byte) bool {
	if len(frame) < 4 {
		return false
	}
	want := ModbusCRC16(frame[:len(frame)-2])
	got := uint16(frame[len(frame)-2]) | uint16(frame[len(frame)-1])<<8
	return want == got
}

// modbusFlusher is an optional interface implemented by real serial transports to flush stale
// RX bytes between Modbus RTU half-duplex frames (preventing frame desynchronization after timeouts).
type modbusFlusher interface {
	FlushInput() error
}

// ModbusRTUClient implements pure-Go Modbus RTU master framing over a ModbusTransport.
// All bus transactions are serialized with a mutex to prevent byte interleaving on half-duplex RS-485.
type ModbusRTUClient struct {
	mu              sync.Mutex
	transport       ModbusTransport
	timeout         time.Duration
	interFrameDelay time.Duration
	lastTxDone      time.Time
}

// NewModbusRTUClient creates a new RTU master over transport with a per-request timeout.
func NewModbusRTUClient(transport ModbusTransport, timeout time.Duration) (*ModbusRTUClient, error) {
	if transport == nil {
		return nil, fmt.Errorf("modbus transport must not be nil")
	}
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}
	var ifd time.Duration
	if _, isSerial := transport.(modbusFlusher); isSerial {
		// Modbus Serial Line Spec V1.02 §2.5.1.1: >= 3.5 char silent interval (>= 1.75 ms at >= 19200 baud,
		// ~4 ms at 9600 baud). Applied on physical serial ports only so mock unit tests stay instantaneous.
		ifd = 4 * time.Millisecond
	}
	return &ModbusRTUClient{
		transport:       transport,
		timeout:         timeout,
		interFrameDelay: ifd,
	}, nil
}

// Close closes the underlying ModbusTransport.
func (c *ModbusRTUClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.transport.Close()
}

// ReadHoldingRegisters executes FC03 (Read Holding Registers) starting at addr for quantity registers.
func (c *ModbusRTUClient) ReadHoldingRegisters(ctx context.Context, slaveID uint8, addr, quantity uint16) ([]uint16, error) {
	return c.readRegisters(ctx, slaveID, FuncReadHoldingRegisters, addr, quantity)
}

// ReadInputRegisters executes FC04 (Read Input Registers) starting at addr for quantity registers.
func (c *ModbusRTUClient) ReadInputRegisters(ctx context.Context, slaveID uint8, addr, quantity uint16) ([]uint16, error) {
	return c.readRegisters(ctx, slaveID, FuncReadInputRegisters, addr, quantity)
}

func (c *ModbusRTUClient) readRegisters(ctx context.Context, slaveID, fc uint8, addr, quantity uint16) ([]uint16, error) {
	if err := validateSlaveID(slaveID); err != nil {
		return nil, err
	}
	if quantity == 0 || quantity > modbusMaxReadRegs {
		return nil, fmt.Errorf("modbus: register read quantity %d out of range [1, %d]", quantity, modbusMaxReadRegs)
	}

	req := make([]byte, 6, 8)
	req[0] = slaveID
	req[1] = fc
	binary.BigEndian.PutUint16(req[2:4], addr)
	binary.BigEndian.PutUint16(req[4:6], quantity)
	req = appendModbusCRC(req)

	resp, err := c.roundTrip(ctx, slaveID, fc, req, int(quantity)*2)
	if err != nil {
		return nil, err
	}

	byteCount := int(resp[2])
	expectedBytes := int(quantity) * 2
	if byteCount != expectedBytes {
		return nil, fmt.Errorf("%w: got %d data bytes, want %d", ErrModbusUnexpectedLength, byteCount, expectedBytes)
	}

	regs := make([]uint16, quantity)
	for i := 0; i < int(quantity); i++ {
		regs[i] = binary.BigEndian.Uint16(resp[3+2*i : 5+2*i])
	}
	return regs, nil
}

// WriteSingleCoil executes FC05 (Write Single Coil) at addr with value (true=0xFF00, false=0x0000).
func (c *ModbusRTUClient) WriteSingleCoil(ctx context.Context, slaveID uint8, addr uint16, value bool) error {
	if err := validateSlaveID(slaveID); err != nil {
		return err
	}
	coilVal := modbusCoilOff
	if value {
		coilVal = modbusCoilOn
	}

	req := make([]byte, 6, 8)
	req[0] = slaveID
	req[1] = FuncWriteSingleCoil
	binary.BigEndian.PutUint16(req[2:4], addr)
	binary.BigEndian.PutUint16(req[4:6], coilVal)
	req = appendModbusCRC(req)

	resp, err := c.roundTrip(ctx, slaveID, FuncWriteSingleCoil, req, 0)
	if err != nil {
		return err
	}
	echoAddr := binary.BigEndian.Uint16(resp[2:4])
	echoVal := binary.BigEndian.Uint16(resp[4:6])
	if echoAddr != addr || echoVal != coilVal {
		return fmt.Errorf("%w: FC05 echoed addr=0x%04X val=0x%04X, want addr=0x%04X val=0x%04X",
			ErrModbusEchoMismatch, echoAddr, echoVal, addr, coilVal)
	}
	return nil
}

// WriteSingleRegister executes FC06 (Write Single Register) at addr with value.
func (c *ModbusRTUClient) WriteSingleRegister(ctx context.Context, slaveID uint8, addr, value uint16) error {
	if err := validateSlaveID(slaveID); err != nil {
		return err
	}

	req := make([]byte, 6, 8)
	req[0] = slaveID
	req[1] = FuncWriteSingleRegister
	binary.BigEndian.PutUint16(req[2:4], addr)
	binary.BigEndian.PutUint16(req[4:6], value)
	req = appendModbusCRC(req)

	resp, err := c.roundTrip(ctx, slaveID, FuncWriteSingleRegister, req, 0)
	if err != nil {
		return err
	}
	echoAddr := binary.BigEndian.Uint16(resp[2:4])
	echoVal := binary.BigEndian.Uint16(resp[4:6])
	if echoAddr != addr || echoVal != value {
		return fmt.Errorf("%w: FC06 echoed addr=0x%04X val=0x%04X, want addr=0x%04X val=0x%04X",
			ErrModbusEchoMismatch, echoAddr, echoVal, addr, value)
	}
	return nil
}

// WriteMultipleRegisters executes FC16 (Write Multiple Registers) starting at addr.
func (c *ModbusRTUClient) WriteMultipleRegisters(ctx context.Context, slaveID uint8, addr uint16, values []uint16) error {
	if err := validateSlaveID(slaveID); err != nil {
		return err
	}
	n := len(values)
	if n == 0 || n > int(modbusMaxWriteRegs) {
		return fmt.Errorf("modbus: register write quantity %d out of range [1, %d]", n, modbusMaxWriteRegs)
	}

	byteCount := n * 2
	req := make([]byte, 7+byteCount, 9+byteCount)
	req[0] = slaveID
	req[1] = FuncWriteMultipleRegisters
	binary.BigEndian.PutUint16(req[2:4], addr)
	binary.BigEndian.PutUint16(req[4:6], uint16(n))
	req[6] = byte(byteCount)
	for i, v := range values {
		binary.BigEndian.PutUint16(req[7+2*i:9+2*i], v)
	}
	req = appendModbusCRC(req)

	resp, err := c.roundTrip(ctx, slaveID, FuncWriteMultipleRegisters, req, 0)
	if err != nil {
		return err
	}
	echoAddr := binary.BigEndian.Uint16(resp[2:4])
	echoQty := binary.BigEndian.Uint16(resp[4:6])
	if echoAddr != addr || echoQty != uint16(n) {
		return fmt.Errorf("%w: FC16 echoed addr=0x%04X qty=%d, want addr=0x%04X qty=%d",
			ErrModbusEchoMismatch, echoAddr, echoQty, addr, n)
	}
	return nil
}

// roundTrip sends req and reads/validates the Modbus RTU response frame.
// For FC03/FC04, expectedDataBytes is 2*quantity; for FC05/FC06/FC16, expectedDataBytes is 0 (fixed 8-byte response).
func (c *ModbusRTUClient) roundTrip(ctx context.Context, slaveID, expectedFC uint8, req []byte, expectedDataBytes int) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Enforce Modbus RTU t3.5 silent interval on physical serial lines and discard any stale RX bytes.
	if c.interFrameDelay > 0 && !c.lastTxDone.IsZero() {
		if wait := c.interFrameDelay - time.Since(c.lastTxDone); wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
	}
	if fl, ok := c.transport.(modbusFlusher); ok {
		_ = fl.FlushInput()
	}
	defer func() {
		c.lastTxDone = time.Now()
	}()

	deadline := time.Now().Add(c.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := c.transport.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("modbus: set deadline: %w", err)
	}

	// Interrupt blocked serial Read/Write immediately if ctx is canceled mid-transaction.
	doneCh := make(chan struct{})
	defer close(doneCh)
	if ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				_ = c.transport.SetDeadline(time.Now())
			case <-doneCh:
			}
		}()
	}

	if err := writeAll(c.transport, req); err != nil {
		return nil, classifyIOError(ctx, err)
	}

	// Read 2-byte header: [slaveID, funcCode].
	header := make([]byte, 2)
	if err := readExact(ctx, c.transport, header, deadline); err != nil {
		return nil, err
	}

	respSlave := header[0]
	respFC := header[1]

	// Exception response frame: 5 bytes total [slaveID, fc|0x80, excCode, crcLo, crcHi].
	if respFC&modbusExceptionBit != 0 {
		rest := make([]byte, 3)
		if err := readExact(ctx, c.transport, rest, deadline); err != nil {
			return nil, err
		}
		frame := append(header, rest...)
		if !verifyModbusCRC(frame) {
			return nil, ErrModbusCRC
		}
		if respSlave != slaveID {
			return nil, fmt.Errorf("%w: got slave %d, want %d", ErrModbusSlaveMismatch, respSlave, slaveID)
		}
		baseFC := respFC &^ modbusExceptionBit
		if baseFC != expectedFC {
			return nil, fmt.Errorf("%w: got exception for FC 0x%02X, want 0x%02X", ErrModbusFunctionMismatch, baseFC, expectedFC)
		}
		return nil, &ModbusExceptionError{
			SlaveID:       respSlave,
			FunctionCode:  baseFC,
			ExceptionCode: rest[0],
		}
	}

	var frame []byte
	switch expectedFC {
	case FuncReadHoldingRegisters, FuncReadInputRegisters:
		// If a stray non-read frame arrived (e.g. a late 8-byte FC05/FC06 echo), reject before
		// misinterpreting its address byte as a variable byte count.
		if respFC != expectedFC {
			if fl, ok := c.transport.(modbusFlusher); ok {
				_ = fl.FlushInput()
			}
			return nil, fmt.Errorf("%w: got FC 0x%02X, want 0x%02X", ErrModbusFunctionMismatch, respFC, expectedFC)
		}
		// Read byte count byte first so we know exact remaining frame size.
		bcBuf := make([]byte, 1)
		if err := readExact(ctx, c.transport, bcBuf, deadline); err != nil {
			return nil, err
		}
		byteCount := int(bcBuf[0])
		if byteCount == 0 || byteCount > int(modbusMaxReadRegs)*2 {
			return nil, fmt.Errorf("%w: invalid byte count %d", ErrModbusUnexpectedLength, byteCount)
		}
		payloadAndCRC := make([]byte, byteCount+2)
		if err := readExact(ctx, c.transport, payloadAndCRC, deadline); err != nil {
			return nil, err
		}
		frame = make([]byte, 0, 3+len(payloadAndCRC))
		frame = append(frame, header...)
		frame = append(frame, bcBuf[0])
		frame = append(frame, payloadAndCRC...)

	case FuncWriteSingleCoil, FuncWriteSingleRegister, FuncWriteMultipleRegisters:
		// Fixed 8-byte response: [slave, fc, addrHi, addrLo, valOrQtyHi, valOrQtyLo, crcLo, crcHi].
		rest := make([]byte, 6)
		if err := readExact(ctx, c.transport, rest, deadline); err != nil {
			return nil, err
		}
		frame = make([]byte, 0, 8)
		frame = append(frame, header...)
		frame = append(frame, rest...)

	default:
		return nil, fmt.Errorf("modbus: unsupported function code 0x%02X", expectedFC)
	}

	if !verifyModbusCRC(frame) {
		return nil, ErrModbusCRC
	}
	if respSlave != slaveID {
		return nil, fmt.Errorf("%w: got slave %d, want %d", ErrModbusSlaveMismatch, respSlave, slaveID)
	}
	if respFC != expectedFC {
		return nil, fmt.Errorf("%w: got FC 0x%02X, want 0x%02X", ErrModbusFunctionMismatch, respFC, expectedFC)
	}
	if expectedDataBytes > 0 && len(frame) != 3+expectedDataBytes+2 {
		return nil, fmt.Errorf("%w: frame length %d, want %d", ErrModbusUnexpectedLength, len(frame), 5+expectedDataBytes)
	}

	return frame, nil
}

func validateSlaveID(slaveID uint8) error {
	if slaveID < modbusMinSlaveID || slaveID > modbusMaxSlaveID {
		return fmt.Errorf("modbus: slave ID %d out of valid unicast range [%d, %d]",
			slaveID, modbusMinSlaveID, modbusMaxSlaveID)
	}
	return nil
}

func writeAll(w io.Writer, data []byte) error {
	written := 0
	for written < len(data) {
		n, err := w.Write(data[written:])
		written += n
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func readExact(ctx context.Context, r io.Reader, buf []byte, deadline time.Time) error {
	read := 0
	for read < len(buf) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return ErrModbusTimeout
		}
		n, err := r.Read(buf[read:])
		read += n
		if err != nil {
			if read == len(buf) && errors.Is(err, io.EOF) {
				return nil
			}
			return classifyIOError(ctx, err)
		}
		if n == 0 {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return ErrModbusTimeout
			}
			// Guard against non-blocking readers returning (0, nil) repeatedly.
			time.Sleep(time.Millisecond)
		}
	}
	return nil
}

func classifyIOError(ctx context.Context, err error) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", ErrModbusTimeout, err)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: %v", ErrModbusTimeout, err)
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("%w: %v", ErrModbusFrameTooShort, err)
	}
	return err
}
