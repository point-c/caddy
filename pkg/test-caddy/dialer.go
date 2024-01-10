package test_caddy

import (
	"context"
	"errors"
	pointc "github.com/point-c/caddy"
	"net"
	"testing"
)

var _ pointc.Dialer = (*TestDialer)(nil)

type TestDialer struct {
	t            testing.TB
	DialFn       func(context.Context, *net.TCPAddr) (net.Conn, error)
	DialPacketFn func(*net.UDPAddr) (net.PacketConn, error)
}

func NewTestDialer(t testing.TB) *TestDialer {
	t.Helper()
	return &TestDialer{t: t}
}

func (td *TestDialer) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	td.t.Helper()
	if td.DialFn != nil {
		return td.DialFn(ctx, addr)
	}
	return nil, errors.New("dial not implemented")
}

func (td *TestDialer) DialPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	td.t.Helper()
	if td.DialPacketFn != nil {
		return td.DialPacketFn(addr)
	}
	return nil, errors.New("dialPacket not implemented")
}
