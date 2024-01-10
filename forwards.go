package point_c

import (
	"encoding/json"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/caddy/pkg/lifecycler"
)

func init() {
	caddy.RegisterModule(new(Forward))
}

var (
	_ NetOp                 = (*Forward)(nil)
	_ caddy.Provisioner     = (*Forward)(nil)
	_ caddy.CleanerUpper    = (*Forward)(nil)
	_ caddy.Module          = (*Forward)(nil)
	_ caddyfile.Unmarshaler = (*Forward)(nil)
)

type (
	Forward struct {
		ForwardsRaw []json.RawMessage `json:"forwards,omitempty" caddy:"namespace=point-c.op.forward inline_key=forward"`
		Host        configvalues.Hostname
		lf          lifecycler.LifeCycler[Net]
	}
	ForwardProto = lifecycler.LifeCyclable[Net]
)

func (f *Forward) Provision(ctx caddy.Context) error {
	return f.lf.Provision(ctx, &lifecycler.ProvisionInfo{
		StructPointer: f,
		FieldName:     "ForwardsRaw",
		Raw:           &f.ForwardsRaw,
	})
}

func (f *Forward) Start(lookup NetLookup) error {
	v, ok := lookup.Lookup(f.Host.Value())
	if !ok {
		return fmt.Errorf("host %q not found", f.Host.Value())
	}
	f.lf.SetValue(v)
	return f.lf.Start()
}

func (f *Forward) Cleanup() error { return f.lf.Cleanup() }

func (f *Forward) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "point-c.ops.forward",
		New: func() caddy.Module { return new(Forward) },
	}
}

func (f *Forward) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return f.lf.UnmarshalCaddyfile(d, &lifecycler.CaddyfileInfo{
		ModuleID:           "point-c.op.forward.",
		Raw:                &f.ForwardsRaw,
		SubModuleSpecifier: "forward",
		ParseVerbLine:      &f.Host,
	})
}
