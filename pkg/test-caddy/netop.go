package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.NetOp = (*TestNetOp)(nil)

type TestNetOp struct{ TestModule[pointc.NetLookup] }

func NewTestNetOp(t testing.TB) (v *TestNetOp) {
	t.Helper()
	defer NewTestModule[pointc.NetLookup, *TestNetOp](t, &v, func(v *TestNetOp) *TestModule[pointc.NetLookup] { t.Helper(); return &v.TestModule }, "point-c.op.")
	return &TestNetOp{}
}
