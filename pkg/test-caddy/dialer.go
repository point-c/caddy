package test_caddy

import (
	"context"
	"errors"
	pointc "github.com/point-c/caddy"
	"net"
	"testing"
)

var _ pointc.Dialer = (*TestDialer)(nil)

// TestDialer is a mock [pointc.Dialer].
type TestDialer struct {
	t            testing.TB
	DialFn       func(context.Context, *net.TCPAddr) (net.Conn, error)
	DialPacketFn func(*net.UDPAddr) (net.PacketConn, error)
}

// NewTestDialer creates and initializes a new instance of [TestDialer].
func NewTestDialer(t testing.TB) *TestDialer {
	return &TestDialer{t: t}
}

// Dial attempts to call DialFn. If DialFn is not set, an error is returned.
func (td *TestDialer) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	if td.DialFn != nil {
		return td.DialFn(ctx, addr)
	}
	return nil, errors.New("dial not implemented")
}

// DialPacket attempts to call DialPacketFn. If DialPacketFn is not set, an error is returned.
func (td *TestDialer) DialPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	if td.DialPacketFn != nil {
		return td.DialPacketFn(addr)
	}
	return nil, errors.New("dialPacket not implemented")
}
