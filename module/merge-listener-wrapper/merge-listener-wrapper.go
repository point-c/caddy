package merge_listener_wrapper

import (
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/channel-listener"
	"github.com/point-c/caddy/pkg/lifecycler"
	"net"
)

func init() {
	caddyreg.R[*MergeWrapper]()
}

var (
	_ caddy.Provisioner     = (*MergeWrapper)(nil)
	_ caddy.CleanerUpper    = (*MergeWrapper)(nil)
	_ caddy.ListenerWrapper = (*MergeWrapper)(nil)
	_ caddy.Module          = (*MergeWrapper)(nil)
	_ caddyfile.Unmarshaler = (*MergeWrapper)(nil)
)

type (
	// MergeWrapper loads multiple [net.Listener]s and aggregates their [net.Conn]s into a single [net.Listener].
	// It allows caddy to accept connections from multiple sources.
	MergeWrapper struct {
		// ListenerRaw is a slice of JSON-encoded data representing listener configurations.
		// These configurations are used to create the actual net.Listener instances.
		// Listeners should implement [net.Listener] and be in the 'caddy.listeners.merge.listeners' namespace.
		ListenerRaw []json.RawMessage `json:"listeners" caddy:"namespace=caddy.listeners.merge inline_key=listener"`

		// listeners is a slice of net.Listener instances created based on the configurations
		// provided in ListenerRaw. These listeners are the actual network listeners that
		// will be accepting connections.
		listeners []net.Listener

		// conns is a channel for net.Conn instances. Connections accepted by any of the
		// listeners in the 'listeners' slice are sent through this channel.
		// This channel is passed to the constructor of [channel_listener.Listener].
		conns chan net.Conn

		lf lifecycler.LifeCycler[func(net.Listener)]
	}
	// ListenerProvider is implemented by modules in the "caddy.listeners.merge" namespace.
	ListenerProvider lifecycler.LifeCyclable[func(net.Listener)]
)

// CaddyModule implements [caddy.Module].
func (p *MergeWrapper) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[MergeWrapper, *MergeWrapper]("caddy.listeners.merge")
}

// Provision implements [caddy.Provisioner].
// It loads the listeners from their configs and asserts them to [net.Listener].
// Any failed assertions will cause a panic.
func (p *MergeWrapper) Provision(ctx caddy.Context) error {
	p.conns = make(chan net.Conn)
	p.lf.SetValue(func(ln net.Listener) { p.listeners = append(p.listeners, ln) })
	if err := p.lf.Provision(ctx, &lifecycler.ProvisionInfo{
		StructPointer: p,
		FieldName:     "ListenerRaw",
		Raw:           &p.ListenerRaw,
	}); err != nil {
		return err
	}
	return p.lf.Start()
}

// Cleanup implements [caddy.CleanerUpper].
// All wrapped listeners are closed and the struct is cleared.
func (p *MergeWrapper) Cleanup() (err error) {
	for _, ln := range p.listeners {
		err = errors.Join(err, ln.Close())
	}
	err = errors.Join(err, p.lf.Cleanup())
	*p = MergeWrapper{}
	return
}

// WrapListener implements [caddy.ListenerWrapper].
// The listener passed in is closed by [MergeWrapper] during cleanup.
func (p *MergeWrapper) WrapListener(ls net.Listener) net.Listener {
	p.listeners = append(p.listeners, ls)
	cl := channel_listener.New(p.conns, ls.Addr())
	for _, ls := range p.listeners {
		go listen(ls, p.conns, cl.Done(), cl.CloseWithErr)
	}
	return cl
}

// listen manages incoming network connections on a given listener.
// It sends accepted connections to the 'conns' channel. When a
// signal is sent to the 'done' channel any accepted connections not passed on are closed and ignored.
// In case of an error during accepting a connection, it calls the 'finish' function with the error.
func listen(ls net.Listener, conns chan<- net.Conn, done <-chan struct{}, finish func(error) error) {
	for {
		c, err := ls.Accept()
		if err != nil {
			// If one connection errors on Accept, pass the error on and close all other connections.
			// Only the first error from an Accept will be passed on.
			_ = finish(err)
			return
		}

		select {
		case <-done:
			// The connection has been closed, close the received connection and ignore it.
			_ = c.Close()
			continue
		case conns <- c:
			// Connection has been accepted
		}
	}
}

// UnmarshalCaddyfile implements [caddyfile.Unmarshaler].
// Must have at least one listener to aggregate with the wrapped listener.
// `tls` should come specifically after any `merge` directives.
//
// ```
//
//	 http caddyfile:
//		{
//		  servers :443 {
//		    listener_wrappers {
//		      merge {
//		        <submodule name> <submodule config>
//		      }
//		      tls
//		    }
//		  }
//		}
//
// ```
func (p *MergeWrapper) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return p.lf.UnmarshalCaddyfile(d, &lifecycler.CaddyfileInfo{
		ModuleID:           []string{"caddy", "listeners", "merge"},
		Raw:                &p.ListenerRaw,
		SubModuleSpecifier: "listener",
	})
}
