//go:build linux

package control

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	// Linux ioctl constants for flushing serial queues (uniform across Linux architectures).
	ioctlTCFLSH uintptr = 0x540B
	tcioflush   uintptr = 2
)

// serialPort wraps an *os.File opened in raw 8N1 serial mode and satisfies ModbusTransport.
type serialPort struct {
	file *os.File
}

// OpenSerialPort opens a Linux serial/RS-485 device (e.g. /dev/ttyUSB0, /dev/ttyAMA0)
// and configures raw 8N1 framing at baudRate using stdlib syscall TCGETS/TCSETS ioctls
// with zero external dependencies (CGO_ENABLED=0 compatible).
func OpenSerialPort(device string, baudRate int) (ModbusTransport, error) {
	if device == "" {
		return nil, fmt.Errorf("serial device path must not be empty")
	}
	baudFlag, err := linuxBaudFlag(baudRate)
	if err != nil {
		return nil, err
	}

	fd, err := syscall.Open(device, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("open serial device %q: %w", device, err)
	}

	if err := configureLinuxTermios(fd, baudFlag); err != nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("configure serial device %q (%d baud 8N1): %w", device, baudRate, err)
	}

	// Best-effort flush of stale input/output bytes before starting Modbus RTU framing.
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlTCFLSH, tcioflush)

	file := os.NewFile(uintptr(fd), device)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("create os.File for serial device %q", device)
	}

	return &serialPort{file: file}, nil
}

func configureLinuxTermios(fd int, baudFlag uint32) error {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}

	// Raw 8N1 configuration: no input/output processing, no canonical mode, no echo, no parity, 1 stop bit.
	t.Iflag = 0
	t.Oflag = 0
	t.Lflag = 0
	t.Cflag = baudFlag | syscall.CS8 | syscall.CREAD | syscall.CLOCAL
	t.Ispeed = baudFlag
	t.Ospeed = baudFlag
	t.Cc[syscall.VMIN] = 0
	t.Cc[syscall.VTIME] = 1 // 100 ms decisecond fallback read timeout

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return errno
	}
	return nil
}

func linuxBaudFlag(baudRate int) (uint32, error) {
	switch baudRate {
	case 1200:
		return syscall.B1200, nil
	case 2400:
		return syscall.B2400, nil
	case 4800:
		return syscall.B4800, nil
	case 9600:
		return syscall.B9600, nil
	case 19200:
		return syscall.B19200, nil
	case 38400:
		return syscall.B38400, nil
	case 57600:
		return syscall.B57600, nil
	case 115200:
		return syscall.B115200, nil
	case 230400:
		return syscall.B230400, nil
	default:
		return 0, fmt.Errorf("unsupported serial baud rate %d (supported: 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400)", baudRate)
	}
}

func (s *serialPort) Read(p []byte) (int, error) {
	return s.file.Read(p)
}

func (s *serialPort) Write(p []byte) (int, error) {
	return s.file.Write(p)
}

func (s *serialPort) Close() error {
	return s.file.Close()
}

func (s *serialPort) SetDeadline(t time.Time) error {
	if err := s.file.SetDeadline(t); err != nil {
		// Some character devices do not support runtime netpoller deadlines;
		// readExact also enforces wall-clock deadlines with VTIME=1 fallback.
		return nil
	}
	return nil
}
