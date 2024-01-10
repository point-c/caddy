package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.NetLookup = (*TestNetLookup)(nil)

type TestNetLookup struct {
	t        testing.TB
	LookupFn func(string) (pointc.Net, bool)
}

func NewTestNetLookup(t testing.TB) *TestNetLookup {
	t.Helper()
	return &TestNetLookup{t: t}
}

func (tnl *TestNetLookup) Lookup(name string) (pointc.Net, bool) {
	tnl.t.Helper()
	if tnl.LookupFn != nil {
		return tnl.LookupFn(name)
	}
	return nil, false
}
