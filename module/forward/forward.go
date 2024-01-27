// Package forward manages network forwarders.
package forward

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/caddy/pkg/lifecycler"
)

func init() {
	caddyreg.R[*Forward]()
}

var (
	_ point_c.NetOp         = (*Forward)(nil)
	_ caddy.Provisioner     = (*Forward)(nil)
	_ caddy.CleanerUpper    = (*Forward)(nil)
	_ caddy.Module          = (*Forward)(nil)
	_ caddyfile.Unmarshaler = (*Forward)(nil)
)

type (
	// Forward manages forwarders for internet traffic.
	Forward struct {
		ForwardsRaw []json.RawMessage         `json:"forwards,omitempty" caddy:"namespace=point-c.op.forward inline_key=forward"`
		Hosts       configvalues.HostnamePair `json:"hosts"`
		lf          lifecycler.LifeCycler[*ForwardNetworks]
	}
	// ForwardProto is implemented by modules in the "point-c.op.forward" namespace.
	ForwardProto = lifecycler.LifeCyclable[*ForwardNetworks]
	// ForwardNetworks contains the networks that have their traffic forwarded.
	ForwardNetworks struct{ Src, Dst point_c.Net }
)

// Provision implements [caddy.Provisioner].
func (f *Forward) Provision(ctx caddy.Context) error {
	return f.lf.Provision(ctx, &lifecycler.ProvisionInfo{
		StructPointer: f,
		FieldName:     "ForwardsRaw",
		Raw:           &f.ForwardsRaw,
	})
}

// Start implements [NetOp].
func (f *Forward) Start(lookup point_c.NetLookup) error {
	check := func(name string, n *point_c.Net) error {
		if v, ok := lookup.Lookup(name); ok {
			*n = v
			return nil
		}
		return fmt.Errorf("host %q not found", name)
	}

	var fn ForwardNetworks
	if err := errors.Join(
		check(f.Hosts.Value().Left, &fn.Src),
		check(f.Hosts.Value().Right, &fn.Dst),
	); err != nil {
		return err
	}

	f.lf.SetValue(&fn)
	return f.lf.Start()
}

// Cleanup implements [caddy.CleanerUpper].
func (f *Forward) Cleanup() error { return f.lf.Cleanup() }

// CaddyModule implements [caddy.Module].
func (f *Forward) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Forward, *Forward]("point-c.op.forward")
}

// UnmarshalCaddyfile unmarshals the caddyfile.
// ```
//
//	point-c netops {
//	    forward <src network name>:<dst network name> {
//			    <submodule name> <submodule config>
//	    }
//	}
//
// ```
func (f *Forward) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return f.lf.UnmarshalCaddyfile(d, &lifecycler.CaddyfileInfo{
		ModuleID:           []string{"point-c", "op", "forward"},
		Raw:                &f.ForwardsRaw,
		SubModuleSpecifier: "forward",
		ParseVerbLine:      &f.Hosts,
	})
}
