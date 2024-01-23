package point_c

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"net"
)

var (
	_ caddy.Module          = (*Sysnet)(nil)
	_ caddy.Provisioner     = (*Sysnet)(nil)
	_ caddy.Validator       = (*Sysnet)(nil)
	_ caddy.CleanerUpper    = (*Sysnet)(nil)
	_ caddyfile.Unmarshaler = (*Sysnet)(nil)
	_ Network               = (*Sysnet)(nil)
	_ Net                   = (*Sysnet)(nil)
)

func init() { caddyreg.R[*Sysnet]() }

type Sysnet struct {
	Hostname configvalues.Hostname `json:"hostname"`
	Addr     configvalues.IP       `json:"addr"`
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Sysnet) Provision(c caddy.Context) error {
	s.ctx, s.cancel = context.WithCancel(c)
	return nil
}

func (s *Sysnet) Validate() error {
	ip := s.Addr.Value()
	if net.IPv4zero.Equal(ip) {
		return nil
	}

	addrs, err := net.InterfaceAddrs()
	for _, addr := range addrs {
		a, ok := addr.(*net.IPNet)
		if ok && a.Contains(ip) {
			return nil
		}
	}
	return errors.Join(err, errors.New("not an address associated with this system"))
}

func (s *Sysnet) Cleanup() error { s.cancel(); return nil }

func (s *Sysnet) LocalAddr() net.IP { return s.Addr.Value() }

func (s *Sysnet) Start(fn RegisterFunc) error { return fn(s.Hostname.Value(), s) }

func (s *Sysnet) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Sysnet, *Sysnet]("point-c.net.system")
}

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

func (s *Sysnet) Listen(addr *net.TCPAddr) (net.Listener, error) {
	return CaddyListen[net.Listener](s.ctx, addr)
}

func (s *Sysnet) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	return CaddyListen[net.PacketConn](s.ctx, addr)
}

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
	return ln.(T), nil
}

type SysDialer struct {
	ctx   context.Context
	local net.IP
	port  uint16
}

func (s *Sysnet) Dialer(ip net.IP, port uint16) Dialer {
	return &SysDialer{ctx: s.ctx, local: ip, port: port}
}

func (s *SysDialer) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)
	context.AfterFunc(s.ctx, cancel)
	return (&net.Dialer{LocalAddr: &net.TCPAddr{IP: s.local, Port: int(s.port)}}).DialContext(ctx, "tcp", addr.String())
}

func (s *SysDialer) DialPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	ln, err := (&net.Dialer{LocalAddr: &net.UDPAddr{IP: s.local, Port: int(s.port)}}).DialContext(s.ctx, "udp", addr.String())
	if err != nil {
		return nil, err
	}
	return ln.(net.PacketConn), nil
}
