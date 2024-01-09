package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"testing"
)

var _ pointc.NetLookup = (*TestNetLookup)(nil)

type TestNetLookup struct {
	T        testing.TB
	LookupFn func(string) (pointc.Net, bool)
}

func NewTestNetLookup(t testing.TB) *TestNetLookup {
	t.Helper()
	return &TestNetLookup{T: t}
}

func (tnl *TestNetLookup) Lookup(name string) (pointc.Net, bool) {
	tnl.T.Helper()
	if tnl.LookupFn != nil {
		return tnl.LookupFn(name)
	}
	return nil, false
}
