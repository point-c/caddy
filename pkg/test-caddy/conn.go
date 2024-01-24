package test_caddy

import (
	"errors"
	"net"
	"testing"
	"time"
)

var _ net.Conn = (*TestConn)(nil)

// TestConn is a mock [net.Conn].
type TestConn struct {
	t                  testing.TB
	ReadFn             func([]byte) (int, error)
	WriteFn            func([]byte) (int, error)
	CloseFn            func() error
	LocalAddrFn        func() net.Addr
	RemoteAddrFn       func() net.Addr
	SetDeadlineFn      func(time.Time) error
	SetReadDeadlineFn  func(time.Time) error
	SetWriteDeadlineFn func(time.Time) error
}

// NewTestConn creates and initializes a new instance of [TestConn].
func NewTestConn(t testing.TB) *TestConn {
	return &TestConn{t: t}
}

// Read attempts to call ReadFn. If ReadFn is not set, an error is returned.
func (tc *TestConn) Read(b []byte) (int, error) {
	if tc.ReadFn != nil {
		return tc.ReadFn(b)
	}
	return 0, errors.New("read not implemented")
}

// Write attempts to call WriteFn. If WriteFn is not set, an error is returned.
func (tc *TestConn) Write(b []byte) (int, error) {
	if tc.WriteFn != nil {
		return tc.WriteFn(b)
	}
	return 0, errors.New("write not implemented")
}

// Close attempts to call CloseFn. If CloseFn is not set, nil is returned.
func (tc *TestConn) Close() error {
	if tc.CloseFn != nil {
		return tc.CloseFn()
	}
	return nil
}

// LocalAddr attempts to call LocalAddrFn. If LocalAddrFn is not set, nil is returned.
func (tc *TestConn) LocalAddr() net.Addr {
	if tc.LocalAddrFn != nil {
		return tc.LocalAddrFn()
	}
	return &net.TCPAddr{}
}

// RemoteAddr attempts to call RemoteAddrFn. If RemoteAddrFn is not set, a pointer to a [net.TCPAddr] is returned..
func (tc *TestConn) RemoteAddr() net.Addr {
	if tc.RemoteAddrFn != nil {
		return tc.RemoteAddrFn()
	}
	return &net.TCPAddr{}
}

// SetDeadline attempts to call SetDeadlineFn. If SetDeadlineFn is not set, nil is returned.
func (tc *TestConn) SetDeadline(t time.Time) error {
	if tc.SetDeadlineFn != nil {
		return tc.SetDeadlineFn(t)
	}
	return nil
}

// SetReadDeadline attempts to call SetReadDeadlineFn. If SetReadDeadlineFn is not set, nil is returned.
func (tc *TestConn) SetReadDeadline(t time.Time) error {
	if tc.SetReadDeadlineFn != nil {
		return tc.SetReadDeadlineFn(t)
	}
	return nil
}

// SetWriteDeadline attempts to call SetWriteDeadlineFn. If SetWriteDeadlineFn is not set, nil is returned.
func (tc *TestConn) SetWriteDeadline(t time.Time) error {
	if tc.SetWriteDeadlineFn != nil {
		return tc.SetWriteDeadlineFn(t)
	}
	return nil
}
