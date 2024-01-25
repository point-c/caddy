package caddy_wg

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/module/wg/internal"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/wgapi/wgconfig"
)

var (
	_ caddy.Module       = (*Server)(nil)
	_ caddy.Provisioner  = (*Server)(nil)
	_ caddy.CleanerUpper = (*Server)(nil)
	_ pointc.Network     = (*Server)(nil)
)

func init() {
	caddy.RegisterModule(new(Server))
}

func (*Server) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "point-c.net.wgserver",
		New: func() caddy.Module { return new(Server) },
	}
}

// Server is a basic wireguard server.
type (
	Server struct {
		Name       configvalues.Hostname   `json:"hostname"`
		IP         configvalues.IP         `json:"ip"`
		ListenPort configvalues.Port       `json:"listen-port"`
		Private    configvalues.PrivateKey `json:"private"`
		Peers      []*ServerPeer           `json:"peers"`
		wg         internal.Wg
	}
	ServerPeer struct {
		Name         configvalues.Hostname     `json:"hostname"`
		Public       configvalues.PublicKey    `json:"public"`
		PresharedKey configvalues.PresharedKey `json:"preshared"`
		IP           configvalues.IP           `json:"ip"`
	}
)

// UnmarshalCaddyfile unmarshals a config in caddyfile form.
//
//		{
//		  point-c {
//		    wgserver <name> {
//		      ip <server ip>
//		      port <server port to listen on>
//		      private <server private key>
//	          peer <name> {
//	              ip <client ip>
//		          public <client public key>
//		          shared <shared key>
//	            }
//		    }
//		}
func (c *Server) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	return c.wg.UnmarshalCaddyfile(d, c.Name.UnmarshalCaddyfile, internal.CaddyfileKeyValues{
		"ip":      c.IP.UnmarshalCaddyfile,
		"port":    c.ListenPort.UnmarshalCaddyfile,
		"private": c.Private.UnmarshalCaddyfile,
		"peer": internal.UnmarshalCaddyfileNesting[ServerPeer](&c.Peers, func(s *ServerPeer) (internal.UnmarshalCaddyfileFn, internal.CaddyfileKeyValues) {
			return s.Name.UnmarshalCaddyfile, internal.CaddyfileKeyValues{
				"ip":     s.IP.UnmarshalCaddyfile,
				"public": s.Public.UnmarshalCaddyfile,
				"shared": s.PresharedKey.UnmarshalCaddyfile,
			}
		}),
	})
}

func (c *Server) Start(fn pointc.RegisterFunc) error {
	m := map[string]*internal.Net{c.Name.Value(): {IP: c.IP.Value()}}

	cfg := wgconfig.Server{
		Private:    c.Private.Value(),
		ListenPort: c.ListenPort.Value(),
	}

	for _, peer := range c.Peers {
		if _, ok := m[peer.Name.Value()]; ok {
			return fmt.Errorf("hostname %q already declared in config", peer.Name.Value())
		}
		cfg.AddPeer(peer.Public.Value(), peer.PresharedKey.Value(), peer.IP.Value())
		m[peer.Name.Value()] = &internal.Net{IP: peer.IP.Value()}
	}
	return c.wg.Start(fn, m, &cfg)
}

func (c *Server) Cleanup() error                    { return c.wg.Cleanup() }
func (c *Server) Provision(ctx caddy.Context) error { return c.wg.Provision(ctx) }
