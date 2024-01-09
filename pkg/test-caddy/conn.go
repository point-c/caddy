package test_caddy

import (
	"errors"
	"net"
	"testing"
	"time"
)

var _ net.Conn = (*TestConn)(nil)

type TestConn struct {
	T                  testing.TB
	ReadFn             func([]byte) (int, error)
	WriteFn            func([]byte) (int, error)
	CloseFn            func() error
	LocalAddrFn        func() net.Addr
	RemoteAddrFn       func() net.Addr
	SetDeadlineFn      func(time.Time) error
	SetReadDeadlineFn  func(time.Time) error
	SetWriteDeadlineFn func(time.Time) error
}

func NewTestConn(t testing.TB) *TestConn {
	t.Helper()
	return &TestConn{T: t}
}

func (tc *TestConn) Read(b []byte) (int, error) {
	tc.T.Helper()
	if tc.ReadFn != nil {
		return tc.ReadFn(b)
	}
	return 0, errors.New("read not implemented")
}

func (tc *TestConn) Write(b []byte) (int, error) {
	tc.T.Helper()
	if tc.WriteFn != nil {
		return tc.WriteFn(b)
	}
	return 0, errors.New("write not implemented")
}

func (tc *TestConn) Close() error {
	tc.T.Helper()
	if tc.CloseFn != nil {
		return tc.CloseFn()
	}
	return nil
}

func (tc *TestConn) LocalAddr() net.Addr {
	tc.T.Helper()
	if tc.LocalAddrFn != nil {
		return tc.LocalAddrFn()
	}
	return &net.TCPAddr{}
}

func (tc *TestConn) RemoteAddr() net.Addr {
	tc.T.Helper()
	if tc.RemoteAddrFn != nil {
		return tc.RemoteAddrFn()
	}
	return &net.TCPAddr{}
}

func (tc *TestConn) SetDeadline(t time.Time) error {
	tc.T.Helper()
	if tc.SetDeadlineFn != nil {
		return tc.SetDeadlineFn(t)
	}
	return nil
}

func (tc *TestConn) SetReadDeadline(t time.Time) error {
	tc.T.Helper()
	if tc.SetReadDeadlineFn != nil {
		return tc.SetReadDeadlineFn(t)
	}
	return nil
}

func (tc *TestConn) SetWriteDeadline(t time.Time) error {
	tc.T.Helper()
	if tc.SetWriteDeadlineFn != nil {
		return tc.SetWriteDeadlineFn(t)
	}
	return nil
}
