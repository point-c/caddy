package test_caddy

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.NetOp = (*TestNetOp)(nil)

type TestNetOp struct {
	t                    testing.TB
	StartFn              func(pointc.NetLookup) error     `json:"-"`
	StopFn               func() error                     `json:"-"`
	UnmarshalCaddyfileFn func(*caddyfile.Dispenser) error `json:"-"`
	UnmarshalJSONFn      func(b []byte) error             `json:"-"`
	id                   uuid.UUID
}

func (t *TestNetOp) CaddyModule() caddy.ModuleInfo {
	t.t.Helper()
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID("point-c.op." + t.id.String()),
		New: func() caddy.Module { t.t.Helper(); return t },
	}
}

func NewTestNetOp(t testing.TB) *TestNetOp {
	t.Helper()
	return &TestNetOp{
		t:  t,
		id: uuid.New(),
	}
}

func (t *TestNetOp) Id() string { t.t.Helper(); return t.id.String() }

func (t *TestNetOp) Start(l pointc.NetLookup) error {
	t.t.Helper()
	if t.StartFn != nil {
		return t.StartFn(l)
	}
	return nil
}

func (t *TestNetOp) Stop() error {
	t.t.Helper()
	if t.StopFn != nil {
		return t.StopFn()
	}
	return nil
}

func (t *TestNetOp) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	t.t.Helper()
	if t.UnmarshalCaddyfileFn != nil {
		return t.UnmarshalCaddyfileFn(d)
	}
	return nil
}

func (t *TestNetOp) UnmarshalJSON(b []byte) error {
	t.t.Helper()
	if t.UnmarshalJSONFn != nil {
		return t.UnmarshalJSONFn(b)
	}
	return nil
}
