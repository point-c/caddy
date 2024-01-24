package test_caddy

import (
	"errors"
	"net"
	"testing"
)

var _ net.Listener = (*TestListener)(nil)

// TestListener is a mock [net.Listener].
type TestListener struct {
	t        testing.TB
	AcceptFn func() (net.Conn, error) `json:"-"`
	CloseFn  func() error             `json:"-"`
	AddrFn   func() net.Addr          `json:"-"`
}

// NewTestListener creates and initializes a new instance of [TestListener].
func NewTestListener(t testing.TB) *TestListener {
	return &TestListener{t: t}
}

// Accept attempts to call AcceptFn. If AcceptFn is not set, an error is returned.
func (tl *TestListener) Accept() (net.Conn, error) {
	if tl.AcceptFn != nil {
		return tl.AcceptFn()
	}
	return nil, errors.New("accept not implemented")
}

// Close attempts to call CloseFn. If CloseFn is not set, nil is returned.
func (tl *TestListener) Close() error {
	if tl.CloseFn != nil {
		return tl.CloseFn()
	}
	return nil
}

// Addr attempts to call AddrFn. If AddrFn is not set, a pointer to a [net.TCPAddr] is returned.
func (tl *TestListener) Addr() net.Addr {
	if tl.AddrFn != nil {
		return tl.AddrFn()
	}
	return &net.TCPAddr{}
}

// TestListenerModule is a mock [net.Listener] that can also be used as a Caddy module.
type TestListenerModule[T any] struct {
	TestListener
	TestModule[T]
}

// NewTestListenerModule creates and initializes a new instance of [TestListenerModule].
func NewTestListenerModule[T any](t testing.TB) (v *TestListenerModule[T]) {
	defer NewTestModule[T, *TestListenerModule[T]](t, &v, func(v *TestListenerModule[T]) *TestModule[T] { return &v.TestModule }, "caddy.listeners.merge.")
	return &TestListenerModule[T]{TestListener: *NewTestListener(t)}
}
