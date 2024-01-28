// Package sysnet is a point-c network for the host network.
package sysnet

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"io"
	"net"
)

var (
	_ caddy.Module          = (*Sysnet)(nil)
	_ caddy.Provisioner     = (*Sysnet)(nil)
	_ caddy.CleanerUpper    = (*Sysnet)(nil)
	_ caddyfile.Unmarshaler = (*Sysnet)(nil)
	_ point_c.Network       = (*Sysnet)(nil)
	_ point_c.Net           = (*Sysnet)(nil)
)

func init() { caddyreg.R[*Sysnet]() }

// Sysnet is a point-c network that can dial and listen on the host system.
type Sysnet struct {
	Hostname configvalues.Hostname `json:"hostname"`
	Addr     configvalues.IP       `json:"addr"`
	ctx      context.Context
	cancel   context.CancelFunc
}

// Provision implements [caddy.Provision].
func (s *Sysnet) Provision(c caddy.Context) error {
	s.ctx, s.cancel = context.WithCancel(c)
	return nil
}

// Cleanup implements [caddy.CleanerUpper].
func (s *Sysnet) Cleanup() error { s.cancel(); return nil }

// LocalAddr returns the address this module is configured with.
func (s *Sysnet) LocalAddr() net.IP { return s.Addr.Value() }

// Start implements [Network]. Is registers this module with the given hostname.
func (s *Sysnet) Start(fn point_c.RegisterFunc) error { return fn(s.Hostname.Value(), s) }

// CaddyModule implements [caddy.Module].
func (s *Sysnet) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Sysnet, *Sysnet]("point-c.net.system")
}

// UnmarshalCaddyfile implements [caddyfule.Unmarshaler].
func (s *Sysnet) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for v := d.Val(); v == "" || v == "system"; v = d.Val() {
			if !d.NextArg() {
				return d.ArgErr()
			}
		}
		err := s.Hostname.UnmarshalCaddyfile(d)
		if !d.NextArg() {
			return err
		}
		return errors.Join(err, s.Addr.UnmarshalCaddyfile(d))
	}
	return nil
}

// Listen listens on the given address using TCP.
func (s *Sysnet) Listen(addr *net.TCPAddr) (net.Listener, error) {
	return CaddyListen[net.Listener](s.ctx, addr)
}

// ListenPacket listens on the given address using UDP.
func (s *Sysnet) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	return CaddyListen[net.PacketConn](s.ctx, addr)
}

// CaddyListen helps with listening on an address using Caddy's method.
// The listener is type asserted to T. An error will be returned if the assertion fails.
func CaddyListen[T any](ctx context.Context, addr net.Addr) (v T, err error) {
	var na caddy.NetworkAddress
	na, err = caddy.ParseNetworkAddress(addr.Network() + "/" + addr.String())
	if err != nil {
		return
	}

	var ln any
	ln, err = na.Listen(ctx, 0, net.ListenConfig{})
	if err != nil {
		return
	}

	l, ok := ln.(T)
	if !ok {
		err = errors.New("invalid listener type")
		if cl, ok := ln.(io.Closer); ok {
			err = errors.Join(err, cl.Close())
		}
		return
	}
	return l, nil
}

// SysDialer allows dialing TCP and UDP connections on the system.
type SysDialer struct {
	ctx   context.Context
	local net.IP
	port  uint16
}

// Dialer returns a [SysDialer] ready to dial on the given address and port.
func (s *Sysnet) Dialer(_ net.IP, port uint16) point_c.Dialer {
	return &SysDialer{ctx: s.ctx, local: s.Addr.Value(), port: port}
}

// Dial dials the given address using TCP.
func (s *SysDialer) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)
	context.AfterFunc(s.ctx, cancel)
	return (&net.Dialer{LocalAddr: &net.TCPAddr{IP: s.local, Port: int(s.port)}}).DialContext(ctx, "tcp", addr.String())
}

// DialPacket dials the given address using UDP.
func (s *SysDialer) DialPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	ln, err := (&net.Dialer{LocalAddr: &net.UDPAddr{IP: s.local, Port: int(s.port)}}).DialContext(s.ctx, "udp", addr.String())
	if err != nil {
		return nil, err
	}
	return ln.(net.PacketConn), nil
}
