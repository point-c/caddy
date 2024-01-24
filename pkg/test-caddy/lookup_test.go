package test_caddy_test

import (
	pointc "github.com/point-c/caddy/module/point-c"
	. "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTestNetLookup(t *testing.T) {
	require.NotNil(t, NewTestNetLookup(t))
}

func TestTestNetLookupLookup(t *testing.T) {
	tnl := NewTestNetLookup(t)

	_, found := tnl.Lookup("test")
	require.False(t, found)

	mockNet := NewTestNet(t)

	tnl.LookupFn = func(name string) (pointc.Net, bool) {
		if name == "test" {
			return mockNet, true
		}
		return nil, false
	}
	net, found := tnl.Lookup("test")
	require.True(t, found)
	require.Equal(t, mockNet, net)

	_, found = tnl.Lookup("non-existing")
	require.False(t, found)
}
