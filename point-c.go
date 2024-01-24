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
	"net"
)

const (
	CaddyfilePointCName = "point-c"
)

func init() {
	caddyreg.R[*Pointc]()
	httpcaddyfile.RegisterGlobalOption(CaddyfilePointCName, configvalues.CaddyfileUnmarshaler[Pointc, *Pointc](CaddyfilePointCName))
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
	// This also helps with configuration, so that users don't need to remember ip addresses.
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
	// Network is implemented by modules in the "point-c.net" namespace.
	Network = lifecycler.LifeCyclable[RegisterFunc]
	// NetOp is implemented by modules in the "point-c.op" namespace.
	NetOp = lifecycler.LifeCyclable[NetLookup]
)

// NetLookup is implemented by [Pointc].
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

// CaddyModule implements [caddy.Module].
func (*Pointc) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Pointc, *Pointc]("point-c")
}

// Register adds a new network to the [Pointc] instance.
// The 'key' parameter is a unique identifier for the network.
// On success, the network is registered with the [Pointc] instance.
func (pc *Pointc) Register(key string, n Net) error {
	if _, ok := pc.net[key]; ok {
		return fmt.Errorf("network %q already exists", key)
	}
	pc.net[key] = n
	return nil
}

// Provision implements [caddy.Provisioner].
func (pc *Pointc) Provision(ctx caddy.Context) error {
	pc.net = map[string]Net{}
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

// Start implements [caddy.App].
func (pc *Pointc) Start() error { return nil }

// Stop implements [caddy.App].
func (pc *Pointc) Stop() error { return nil }

// Cleanup implements [caddy.CleanerUpper].
func (pc *Pointc) Cleanup() error { return errors.Join(pc.lf.Cleanup(), pc.ops.Cleanup()) }

// Lookup gets a [Net] by its declared name.
func (pc *Pointc) Lookup(name string) (Net, bool) {
	n, ok := pc.net[name]
	return n, ok
}

// UnmarshalCaddyfile unmarshals a submodules from a caddyfile.
// The `netops` modifier causes the modules to be loaded as netops.
//
//	 ```
//		{
//		  point-c [netops] {
//		    <submodule name> <submodule config>
//		  }
//		}
//	 ```
func (pc *Pointc) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if v := d.Val(); !d.NextArg() && v == "point-c" {
			return pc.lf.UnmarshalCaddyfile(d.NewFromNextSegment(), &lifecycler.CaddyfileInfo{
				ModuleID:           []string{"point-c", "net"},
				Raw:                &pc.NetworksRaw,
				SubModuleSpecifier: "type",
			})
		} else if v := d.Val(); v == "netops" {
			return pc.ops.UnmarshalCaddyfile(d.NewFromNextSegment(), &lifecycler.CaddyfileInfo{
				ModuleID:           []string{"point-c", "op"},
				Raw:                &pc.NetOps,
				SubModuleSpecifier: "op",
			})
		} else {
			return fmt.Errorf("unrecognized verb %q", v)
		}
	}
	return nil
}
