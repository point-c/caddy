package point_c

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/point-c/channel-listener"
	"net"
)

const NetworkStubName = "stub:"

func init() {
	caddy.RegisterNetwork(NetworkStubName, StubListener)
}

// StubListener creates a stub network listener. This listener does not accept
// actual network connections but instead blocks on Accept calls until Close is called.
// It can be used as a base listener when only tunnel listeners are required.
func StubListener(_ context.Context, _, addr string, _ net.ListenConfig) (any, error) {
	return channel_listener.New(make(<-chan net.Conn), stubAddr(addr)), nil
}

// stubAddr implements [net.Addr] for [StubListener].
type stubAddr string

// Network always returns "stub".
func (stubAddr) Network() string { return "stub" }

// String return [stubAddr] as a string.
func (d stubAddr) String() string { return string(d) }
