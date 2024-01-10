package test_caddy

import (
	"errors"
	"net"
	"testing"
)

var _ net.Listener = (*TestListener)(nil)

type TestListener struct {
	t        testing.TB
	AcceptFn func() (net.Conn, error) `json:"-"`
	CloseFn  func() error             `json:"-"`
	AddrFn   func() net.Addr          `json:"-"`
}

func NewTestListener(t testing.TB) *TestListener {
	t.Helper()
	return &TestListener{t: t}
}

func (tl *TestListener) Accept() (net.Conn, error) {
	tl.t.Helper()
	if tl.AcceptFn != nil {
		return tl.AcceptFn()
	}
	return nil, errors.New("accept not implemented")
}

func (tl *TestListener) Close() error {
	tl.t.Helper()
	if tl.CloseFn != nil {
		return tl.CloseFn()
	}
	return nil
}

func (tl *TestListener) Addr() net.Addr {
	tl.t.Helper()
	if tl.AddrFn != nil {
		return tl.AddrFn()
	}
	return &net.TCPAddr{}
}

type TestListenerModule[T any] struct {
	TestListener
	TestModule[T]
}

func NewTestListenerModule[T any](t testing.TB) *TestListenerModule[T] {
	return &TestListenerModule[T]{
		TestListener: *NewTestListener(t),
		TestModule:   *NewTestModule[T](t, "caddy.listeners.merge."),
	}
}
