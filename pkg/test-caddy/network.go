package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.Network = (*TestNetwork)(nil)

// TestNetwork is a mock point-c network.
type TestNetwork struct {
	TestModule[pointc.RegisterFunc]
}

// NewTestNetwork creates and initializes a new instance of [TestNetwork].
func NewTestNetwork(t testing.TB) (v *TestNetwork) {
	defer NewTestModule[pointc.RegisterFunc, *TestNetwork](t, &v, func(v *TestNetwork) *TestModule[pointc.RegisterFunc] { return &v.TestModule }, "point-c.net.")
	return &TestNetwork{}
}
