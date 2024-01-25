package internal

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/wg"
	"github.com/point-c/wgapi"
	"github.com/point-c/wgevents"
	"go.mrchanchal.com/zaphandler"
	"io"
	"log/slog"
)

type Wg struct {
	Wg     io.Closer
	Logger *slog.Logger
}

type (
	CaddyfileKeyValues   map[string]UnmarshalCaddyfileFn
	UnmarshalCaddyfileFn func(*caddyfile.Dispenser) error
)

func (w *Wg) UnmarshalCaddyfile(d *caddyfile.Dispenser, name UnmarshalCaddyfileFn, m CaddyfileKeyValues) (err error) {
	for d.Next() {
		err = name(d)
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if err != nil {
				return
			}

			key := d.Val()
			if !d.NextArg() {
				return d.ArgErr()
			}

			value, ok := m[key]
			if !ok {
				return fmt.Errorf("unrecognized option %q", key)
			}
			err = value(d)
		}
	}
	return
}

func UnmarshalCaddyfileNesting[T any](s *[]*T, fn func(*T) (UnmarshalCaddyfileFn, CaddyfileKeyValues)) UnmarshalCaddyfileFn {
	return func(d *caddyfile.Dispenser) (err error) {
		var v T
		name, m := fn(&v)
		err = name(d)
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if err != nil {
				return
			}

			key := d.Val()
			if !d.NextArg() {
				return d.ArgErr()
			}

			value, ok := m[key]
			if !ok {
				return fmt.Errorf("unrecognized option %q", key)
			}
			err = value(d)
		}

		if err == nil {
			*s = append(*s, &v)
		}
		return
	}
}

func (w *Wg) Cleanup() error { return w.Wg.Close() }

func (w *Wg) Provision(ctx caddy.Context) error {
	w.Logger = slog.New(zaphandler.New(ctx.Logger()))
	return nil
}

func (w *Wg) Start(fn pointc.RegisterFunc, m map[string]*Net, cfg wgapi.Configurable) (err error) {
	for k, v := range m {
		if err := fn(k, v); err != nil {
			return err
		}
	}

	var n *wg.Net
	w.Wg, err = wg.New(
		wg.OptionConfig(cfg),
		wg.OptionLogger(wgevents.Events(func(e wgevents.Event) { e.Slog(w.Logger) })),
		wg.OptionNetDevice(&n),
	)

	if err == nil {
		for _, v := range m {
			v.Net = n
		}
	}
	return
}
