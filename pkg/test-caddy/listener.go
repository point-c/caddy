package test_caddy

import (
	"errors"
	"net"
	"testing"
)

var _ net.Listener = (*TestListener)(nil)

type TestListener struct {
	T        testing.TB
	AcceptFn func() (net.Conn, error)
	CloseFn  func() error
	AddrFn   func() net.Addr
}

func NewTestListener(t testing.TB) *TestListener {
	t.Helper()
	return &TestListener{T: t}
}

func (tl *TestListener) Accept() (net.Conn, error) {
	tl.T.Helper()
	if tl.AcceptFn != nil {
		return tl.AcceptFn()
	}
	return nil, errors.New("accept not implemented")
}

func (tl *TestListener) Close() error {
	tl.T.Helper()
	if tl.CloseFn != nil {
		return tl.CloseFn()
	}
	return nil
}

func (tl *TestListener) Addr() net.Addr {
	tl.T.Helper()
	if tl.AddrFn != nil {
		return tl.AddrFn()
	}
	return &net.TCPAddr{}
}
