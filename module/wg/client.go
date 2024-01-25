package caddy_wg

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/module/wg/internal"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/wgapi/wgconfig"
)

var (
	_ caddy.Module       = (*Client)(nil)
	_ caddy.Provisioner  = (*Client)(nil)
	_ caddy.CleanerUpper = (*Client)(nil)
	_ pointc.Network     = (*Client)(nil)
)

func init() {
	caddy.RegisterModule(new(Client))
}

func (*Client) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "point-c.net.wgclient",
		New: func() caddy.Module { return new(Client) },
	}
}

// Client is a basic wireguard client.
type (
	Client struct {
		Name      configvalues.Hostname     `json:"name"`
		Endpoint  configvalues.UDPAddr      `json:"endpoint"`
		IP        configvalues.IP           `json:"ip"`
		Private   configvalues.PrivateKey   `json:"private"`
		Public    configvalues.PublicKey    `json:"public"`
		Preshared configvalues.PresharedKey `json:"preshared"`
		wg        internal.Wg
	}
)

// UnmarshalCaddyfile unmarshals a config in caddyfile form.
//
//	{
//	  point-c {
//	    wgclient <name> {
//	      ip <tunnel ip>
//	      endpoint <server address/ip>
//	      private <client private key>
//	      public <server public key>
//	      shared <shared key>
//	  }
//	}
func (c *Client) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	return c.wg.UnmarshalCaddyfile(d, c.Name.UnmarshalCaddyfile, internal.CaddyfileKeyValues{
		"ip":       c.IP.UnmarshalCaddyfile,
		"endpoint": c.Endpoint.UnmarshalCaddyfile,
		"private":  c.Private.UnmarshalCaddyfile,
		"public":   c.Public.UnmarshalCaddyfile,
		"shared":   c.Preshared.UnmarshalCaddyfile,
	})
}

func (c *Client) Start(fn pointc.RegisterFunc) error {
	cfg := wgconfig.Client{
		Private:   c.Private.Value(),
		Public:    c.Public.Value(),
		PreShared: c.Preshared.Value(),
		Endpoint:  *c.Endpoint.Value(),
	}
	cfg.DefaultPersistentKeepAlive()
	cfg.AllowAllIPs()
	return c.wg.Start(fn, map[string]*internal.Net{c.Name.Value(): {IP: c.IP.Value()}}, &cfg)
}

func (c *Client) Cleanup() error                    { return c.wg.Cleanup() }
func (c *Client) Provision(ctx caddy.Context) error { return c.wg.Provision(ctx) }
