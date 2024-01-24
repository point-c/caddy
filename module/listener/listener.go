package listener

import (
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/module/merge-listener-wrapper"
	point_c2 "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"net"
)

func init() {
	caddyreg.R[*Listener]()
}

var (
	_ caddy.Provisioner                       = (*Listener)(nil)
	_ net.Listener                            = (*Listener)(nil)
	_ caddy.Module                            = (*Listener)(nil)
	_ caddyfile.Unmarshaler                   = (*Listener)(nil)
	_ merge_listener_wrapper.ListenerProvider = (*Listener)(nil)
)

// Listener allows a caddy server to listen on a point-c network.
type Listener struct {
	Name configvalues.Hostname `json:"name"`
	Port configvalues.Port     `json:"port"`
	ln   net.Listener
}

// Provision implements [caddy.Provisioner].
func (p *Listener) Provision(ctx caddy.Context) error {
	m, err := ctx.App("point-c")
	if err != nil {
		return err
	}
	n, ok := m.(point_c2.NetLookup).Lookup(p.Name.Value())
	if !ok {
		return fmt.Errorf("point-c net %q does not exist", p.Name.Value())
	}

	ln, err := n.Listen(&net.TCPAddr{IP: n.LocalAddr(), Port: int(p.Port.Value())})
	if err != nil {
		return err
	}
	p.ln = ln
	return nil
}

// Accept implements [net.Listener].
func (p *Listener) Accept() (net.Conn, error) { return p.ln.Accept() }

// Close implements [net.Listener].
func (p *Listener) Close() error { return p.ln.Close() }

// Addr implements [net.Listener].
func (p *Listener) Addr() net.Addr { return p.ln.Addr() }

// CaddyModule implements [caddy.Module].
func (*Listener) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Listener, *Listener]("caddy.listeners.merge.point-c")
}

// Start implement [ListenerProvider].
func (p *Listener) Start(fn func(net.Listener)) error { fn(p); return nil }

// UnmarshalCaddyfile unmarshals the caddyfile.
// ```
//
//		{
//		  servers :443 {
//		    listener_wrappers {
//		      merge {
//	            # this is the actual listener definition
//		        point-c <network name> <port to expose>
//		      }
//	          # make sure tls goes after otherwise encryption will be dropped
//		      tls
//		    }
//		  }
//		}
//
// ```
func (p *Listener) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		var name, port string
		if !d.Args(&name, &port) {
			return d.ArgErr()
		}
		if err := errors.Join(
			p.Name.UnmarshalText([]byte(name)),
			p.Port.UnmarshalText([]byte(port)),
		); err != nil {
			return err
		}
	}
	return nil
}
