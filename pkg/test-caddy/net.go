package test_caddy

import (
	"errors"
	pointc "github.com/point-c/caddy"
	"net"
	"testing"
)

var _ pointc.Net = (*TestNet)(nil)

type TestNet struct {
	t              testing.TB
	ListenFn       func(*net.TCPAddr) (net.Listener, error)
	ListenPacketFn func(*net.UDPAddr) (net.PacketConn, error)
	DialerFn       func(net.IP, uint16) pointc.Dialer
	LocalAddrFn    func() net.IP
}

func NewTestNet(t testing.TB) *TestNet {
	t.Helper()
	return &TestNet{t: t}
}

func (tn *TestNet) Listen(addr *net.TCPAddr) (net.Listener, error) {
	tn.t.Helper()
	if tn.ListenFn != nil {
		return tn.ListenFn(addr)
	}
	return nil, errors.New("Listen not implemented")
}

func (tn *TestNet) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	tn.t.Helper()
	if tn.ListenPacketFn != nil {
		return tn.ListenPacketFn(addr)
	}
	return nil, errors.New("ListenPacket not implemented")
}

func (tn *TestNet) Dialer(laddr net.IP, port uint16) pointc.Dialer {
	tn.t.Helper()
	if tn.DialerFn != nil {
		return tn.DialerFn(laddr, port)
	}
	return NewTestDialer(tn.t)
}

func (tn *TestNet) LocalAddr() net.IP {
	tn.t.Helper()
	if tn.LocalAddrFn != nil {
		return tn.LocalAddrFn()
	}
	return net.IPv4zero
}
