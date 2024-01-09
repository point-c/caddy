package point_c

import (
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/configvalues"
	"net"
)

func init() {
	caddy.RegisterModule(new(Listener))
}

var (
	_ caddy.Provisioner     = (*Listener)(nil)
	_ net.Listener          = (*Listener)(nil)
	_ caddy.Module          = (*Listener)(nil)
	_ caddyfile.Unmarshaler = (*Listener)(nil)
)

// Listener allows a caddy server to listen on a point-c network.
type Listener struct {
	Name configvalues.Hostname `json:"name"`
	Port configvalues.Port     `json:"port"`
	ln   net.Listener
}

func (p *Listener) Provision(ctx caddy.Context) error {
	m, err := ctx.App("point-c")
	if err != nil {
		return err
	}
	n, ok := m.(NetLookup).Lookup(p.Name.Value())
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

func (p *Listener) Accept() (net.Conn, error) { return p.ln.Accept() }
func (p *Listener) Close() error              { return p.ln.Close() }
func (p *Listener) Addr() net.Addr            { return p.ln.Addr() }

func (*Listener) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.listeners.merge.listeners.point-c",
		New: func() caddy.Module { return new(Listener) },
	}
}

// UnmarshalCaddyfile unmarshals the caddyfile.
//
//		{
//		  servers :443 {
//		    listener_wrappers {
//		      merge {
//	         # this is the actual listener definition
//		        point-c <network name> <port to expose>
//		      }
//	       # make sure tls goes after otherwise encryption will be dropped
//		      tls
//		    }
//		  }
//		}
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
