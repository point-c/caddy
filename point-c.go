package point_c

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/caddy/pkg/lifecycler"
	"github.com/point-c/ipcheck"
	"net"
)

const (
	CaddyfilePointCName = "point-c"
	CaddyfileNetOpName  = "netop"
)

func init() {
	caddyreg.R[*Pointc]()
	httpcaddyfile.RegisterGlobalOption(CaddyfilePointCName, configvalues.CaddyfileUnmarshaler[Pointc, *Pointc](CaddyfilePointCName))
	httpcaddyfile.RegisterGlobalOption(CaddyfileNetOpName, configvalues.CaddyfileUnmarshaler[Pointc, *Pointc](CaddyfileNetOpName))
}

var (
	_ caddy.Provisioner     = (*Pointc)(nil)
	_ caddy.Module          = (*Pointc)(nil)
	_ caddy.App             = (*Pointc)(nil)
	_ caddyfile.Unmarshaler = (*Pointc)(nil)
	_ caddy.CleanerUpper    = (*Pointc)(nil)
	_ NetLookup             = (*Pointc)(nil)
)

type (
	// RegisterFunc registers a unique name to a [Net] tunnel.
	// Since ip addresses may be arbitrary depending on what the application is doing in the tunnel, names are used as lookup.
	// This allows helps with configuration, so that users don't need to remember ip addresses.
	RegisterFunc func(string, Net) error
	// Net is a peer in the networking stack. If it has a local address [Net.LocalAddress] should return a non-nil value.
	Net interface {
		// Listen listens on the given address with the TCP protocol.
		Listen(addr *net.TCPAddr) (net.Listener, error)
		// ListenPacket listens on the given address with the UDP protocol.
		ListenPacket(addr *net.UDPAddr) (net.PacketConn, error)
		// Dialer returns a [Dialer] with a given local address. If the network does not support arbitrary remote addresses this value can be ignored.
		Dialer(laddr net.IP, port uint16) Dialer
		// LocalAddr is the local address of the net interface. If it does not have one, return nil.
		LocalAddr() net.IP
	}
	Dialer interface {
		// Dial dials a remote address with the TCP protocol.
		Dial(context.Context, *net.TCPAddr) (net.Conn, error)
		// DialPacket dials a remote address with the UDP protocol.
		DialPacket(*net.UDPAddr) (net.PacketConn, error)
	}
	Network = lifecycler.LifeCyclable[RegisterFunc]
	NetOp   = lifecycler.LifeCyclable[NetLookup]
)

type NetLookup interface {
	Lookup(string) (Net, bool)
}

// Pointc allows usage of networks through a [net]-ish interface.
type Pointc struct {
	NetworksRaw []json.RawMessage `json:"networks,omitempty" caddy:"namespace=point-c.net inline_key=type"`
	NetOps      []json.RawMessage `json:"net-ops,omitempty" caddy:"namespace=point-c.op inline_key=op"`
	lf          lifecycler.LifeCycler[RegisterFunc]
	ops         lifecycler.LifeCycler[NetLookup]
	net         map[string]Net
}

func (*Pointc) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Pointc, *Pointc]("point-c")
}

type PointcNet struct {
	pc *Pointc
	n  Net
}

func (p *PointcNet) ValidLocalAddr(ip net.IP) bool {
	if ipcheck.IsPrivateNetwork(ip) {
		for _, n := range p.pc.net {
			if n.LocalAddr().Equal(ip) {
				return false
			}
		}
	}
	return true
}

func (p *PointcNet) Listen(addr *net.TCPAddr) (net.Listener, error) {
	// Restrict listening to either the specified ip or all ips
	if !(addr.IP.Equal(p.LocalAddr()) || net.IPv4zero.Equal(addr.IP)) {
		return nil, ipcheck.ErrInvalidLocalIP
	}
	return p.n.Listen(addr)
}

func (p *PointcNet) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	// If equal to registered address but not this network's registered address
	if !(addr.IP.Equal(p.LocalAddr()) || net.IPv4zero.Equal(addr.IP)) {
		return nil, ipcheck.ErrInvalidLocalIP
	}
	return p.n.ListenPacket(addr)
}

func (p *PointcNet) Dialer(laddr net.IP, port uint16) Dialer {
	// External network is probably similar to internal, change ip
	// TODO: something better?
	if !p.ValidLocalAddr(laddr) {
		laddr = p.LocalAddr()
	}
	return p.n.Dialer(laddr, port)
}

func (p *PointcNet) LocalAddr() net.IP { return p.n.LocalAddr() }

func (pc *Pointc) Register(key string, n Net) error {
	if !ipcheck.IsPrivateNetwork(n.LocalAddr()) {
		return errors.New("address is not private network")
	}

	for name, nv := range pc.net {
		if nv.LocalAddr().Equal(n.LocalAddr()) {
			return fmt.Errorf("network %q and %q share same address %s", name, key, nv.LocalAddr().String())
		}
	}

	if _, ok := pc.net[key]; ok {
		return fmt.Errorf("network %q already exists", key)
	}
	pc.net[key] = &PointcNet{
		pc: pc,
		n:  n,
	}
	return nil
}

func (pc *Pointc) Provision(ctx caddy.Context) error {
	pc.net = make(map[string]Net)
	pc.ops.SetValue(pc)
	pc.lf.SetValue(pc.Register)

	if err := pc.lf.Provision(ctx, &lifecycler.ProvisionInfo{
		StructPointer: pc,
		FieldName:     "NetworksRaw",
		Raw:           &pc.NetworksRaw,
	}); err != nil {
		return err
	}

	if err := pc.lf.Start(); err != nil {
		return err
	}

	if err := pc.ops.Provision(ctx, &lifecycler.ProvisionInfo{
		StructPointer: pc,
		FieldName:     "NetOps",
		Raw:           &pc.NetOps,
	}); err != nil {
		return err
	}

	return pc.ops.Start()
}

func (pc *Pointc) Start() error   { return nil }
func (pc *Pointc) Stop() error    { return nil }
func (pc *Pointc) Cleanup() error { return errors.Join(pc.lf.Cleanup(), pc.ops.Cleanup()) }

// Lookup gets a [Net] by its declared name.
func (pc *Pointc) Lookup(name string) (Net, bool) {
	n, ok := pc.net[name]
	return n, ok
}

// UnmarshalCaddyfile unmarshals a submodules from a caddyfile.
//
//	{
//	  point-c {
//	    <submodule name> <submodule config>
//	  }
//	  netop {
//	    <submodule name> <submodule config>
//	  }
//	}
func (pc *Pointc) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		switch v := d.Val(); v {
		case "point-c":
			if err := pc.lf.UnmarshalCaddyfile(d.NewFromNextSegment(), &lifecycler.CaddyfileInfo{
				ModuleID:           "point-c.net.",
				Raw:                &pc.NetworksRaw,
				SubModuleSpecifier: "type",
			}); err != nil {
				return err
			}
		case "netop":
			if err := pc.ops.UnmarshalCaddyfile(d.NewFromNextSegment(), &lifecycler.CaddyfileInfo{
				ModuleID:           "point-c.op.",
				Raw:                &pc.NetOps,
				SubModuleSpecifier: "op",
			}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unrecognized verb %q", v)
		}
	}
	return nil
}
