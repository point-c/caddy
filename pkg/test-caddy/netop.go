package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.NetOp = (*TestNetOp)(nil)

// TestNetOp is a mock point-c network operation.
type TestNetOp struct{ TestModule[pointc.NetLookup] }

// NewTestNetOp creates and initializes a new instance of [TestNetOp].
func NewTestNetOp(t testing.TB) (v *TestNetOp) {
	defer NewTestModule[pointc.NetLookup, *TestNetOp](t, &v, func(v *TestNetOp) *TestModule[pointc.NetLookup] { return &v.TestModule }, "point-c.op.")
	return &TestNetOp{}
}
