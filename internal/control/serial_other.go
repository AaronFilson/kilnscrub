//go:build !linux

package control

import (
	"fmt"
	"os"
	"time"
)

type fallbackSerialPort struct {
	file *os.File
}

// OpenSerialPort provides a pure-Go fallback opener on non-Linux platforms so cross-compiled
// binaries build cleanly with CGO_ENABLED=0. On non-Linux hosts, the serial device's baud/8N1
// parameters must be pre-configured by the OS (e.g. via stty/mode).
func OpenSerialPort(device string, baudRate int) (ModbusTransport, error) {
	if device == "" {
		return nil, fmt.Errorf("serial device path must not be empty")
	}
	if !isSupportedBaudRate(baudRate) {
		return nil, fmt.Errorf("unsupported serial baud rate %d", baudRate)
	}
	f, err := os.OpenFile(device, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open serial device %q: %w", device, err)
	}
	return &fallbackSerialPort{file: f}, nil
}

func (s *fallbackSerialPort) Read(p []byte) (int, error) {
	return s.file.Read(p)
}

func (s *fallbackSerialPort) Write(p []byte) (int, error) {
	return s.file.Write(p)
}

func (s *fallbackSerialPort) Close() error {
	return s.file.Close()
}

func (s *fallbackSerialPort) SetDeadline(t time.Time) error {
	_ = s.file.SetDeadline(t)
	return nil
}
