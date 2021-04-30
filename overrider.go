package localredis

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type ConnOverride struct {
	buffer *bytes.Buffer
	closed bool
	AddrOverride
	deadline time.Time
}

func (m *ConnOverride) Write(p []byte) (int, error) {
	return m.buffer.Write(p)
}

func (m *ConnOverride) Read(p []byte) (int, error) {
	return m.buffer.Read(p)
}

func (m *ConnOverride) Close() error {
	if !m.closed {
		return nil
	}
	return fmt.Errorf("already closed")
}

type AddrOverride struct{}

func (m *AddrOverride) Network() string {
	return "tcp"
}
func (m *AddrOverride) String() string {
	return "127.0.0.1"
}

func (m *ConnOverride) LocalAddr() net.Addr {
	return &m.AddrOverride
}

func (m *ConnOverride) RemoteAddr() net.Addr {
	return &m.AddrOverride
}

func (m *ConnOverride) SetDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func (m *ConnOverride) SetReadDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func (m *ConnOverride) SetWriteDeadline(t time.Time) error {
	m.deadline = t
	return nil
}

func NewConnOverride() *ConnOverride {
	m := new(ConnOverride)
	m.buffer = new(bytes.Buffer)
	return m
}

func CommandOverride(cmd string, exec CommandExecutioner) {
	commandMap[cmd] = exec
}
