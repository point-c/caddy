package test_caddy

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.Network = (*TestNetwork)(nil)

func (t *TestNetwork) CaddyModule() caddy.ModuleInfo {
	t.t.Helper()
	return caddy.ModuleInfo{
		ID:  t.ID(),
		New: func() caddy.Module { t.t.Helper(); return t },
	}
}

func (t *TestNetwork) ID() caddy.ModuleID {
	t.t.Helper()
	return caddy.ModuleID("point-c.net.test-" + t.id.String())
}

func (t *TestNetwork) Id() string { return t.id.String() }

type TestNetwork struct {
	t                    testing.TB
	id                   uuid.UUID
	UnmarshalCaddyfileFn func(*caddyfile.Dispenser) error `json:"-"`
	UnmarshalJSONFn      func([]byte) error               `json:"-"`
	StartFn              func(pointc.RegisterFunc) error  `json:"-"`
	StopFn               func() error                     `json:"-"`
}

func (t *TestNetwork) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	t.t.Helper()
	if t.UnmarshalCaddyfileFn == nil {
		return nil
	}
	return t.UnmarshalCaddyfileFn(d)
}

func (t *TestNetwork) UnmarshalJSON(b []byte) error {
	t.t.Helper()
	if t.UnmarshalJSONFn == nil {
		return nil
	}
	return t.UnmarshalJSONFn(b)
}
func NewTestNetwork(t testing.TB) *TestNetwork {
	t.Helper()
	tn := &TestNetwork{
		t:  t,
		id: uuid.New(),
	}
	caddy.RegisterModule(tn)
	return tn
}

func (t *TestNetwork) Stop() error {
	t.t.Helper()
	if t.StopFn == nil {
		return nil
	}
	return t.StopFn()
}

func (t *TestNetwork) Start(fn pointc.RegisterFunc) error {
	t.t.Helper()
	if t.StartFn == nil {
		return nil
	}
	return t.StartFn(fn)
}
