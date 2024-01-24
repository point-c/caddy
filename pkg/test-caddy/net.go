package test_caddy

import (
	"errors"
	pointc "github.com/point-c/caddy/module/point-c"
	"net"
	"testing"
)

var _ pointc.Net = (*TestNet)(nil)

// TestNet is a mock point-c network net module.
type TestNet struct {
	t              testing.TB
	ListenFn       func(*net.TCPAddr) (net.Listener, error)
	ListenPacketFn func(*net.UDPAddr) (net.PacketConn, error)
	DialerFn       func(net.IP, uint16) pointc.Dialer
	LocalAddrFn    func() net.IP
}

// NewTestNet creates and initializes a new instance of [TestNet].
func NewTestNet(t testing.TB) *TestNet {
	return &TestNet{t: t}
}

// Listen attempts to call ListenFn. If ListenFn is not set, an error is returned.
func (tn *TestNet) Listen(addr *net.TCPAddr) (net.Listener, error) {
	if tn.ListenFn != nil {
		return tn.ListenFn(addr)
	}
	return nil, errors.New("listen not implemented")
}

// ListenPacket attempts to call ListenPacketFn. If ListenPacketFn is not set, an error is returned.
func (tn *TestNet) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	if tn.ListenPacketFn != nil {
		return tn.ListenPacketFn(addr)
	}
	return nil, errors.New("ListenPacket not implemented")
}

// Dialer attempts to call DialerFn. If DialerFn is not set, [NewTestDialer] is used to create a value.
func (tn *TestNet) Dialer(laddr net.IP, port uint16) pointc.Dialer {
	if tn.DialerFn != nil {
		return tn.DialerFn(laddr, port)
	}
	return NewTestDialer(tn.t)
}

// LocalAddr attempts to call LocalAddrFn. If LocalAddrFn is not set, an error is returned.
func (tn *TestNet) LocalAddr() net.IP {
	if tn.LocalAddrFn != nil {
		return tn.LocalAddrFn()
	}
	return net.IPv4zero
}
