package test_caddy

import (
	pointc "github.com/point-c/caddy/module/point-c"
	"testing"
)

var _ pointc.NetLookup = (*TestNetLookup)(nil)

// TestNetLookup implements [pointc.NetLookup].
type TestNetLookup struct {
	t        testing.TB
	LookupFn func(string) (pointc.Net, bool)
}

// NewTestNetLookup creates and initializes a new instance of [TestNetLookup].
func NewTestNetLookup(t testing.TB) *TestNetLookup {
	return &TestNetLookup{t: t}
}

// Lookup attempts to call LookupFn. If LookupFn is not set (nil, false) is returned.
func (tnl *TestNetLookup) Lookup(name string) (pointc.Net, bool) {
	if tnl.LookupFn != nil {
		return tnl.LookupFn(name)
	}
	return nil, false
}
