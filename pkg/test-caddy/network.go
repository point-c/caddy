package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.Network = (*TestNetwork)(nil)

type TestNetwork struct {
	TestModule[pointc.RegisterFunc]
}

func NewTestNetwork(t testing.TB) (v *TestNetwork) {
	defer NewTestModule[pointc.RegisterFunc, *TestNetwork](t, &v, func(v *TestNetwork) *TestModule[pointc.RegisterFunc] { t.Helper(); return &v.TestModule }, "point-c.net.")
	return &TestNetwork{}
}
