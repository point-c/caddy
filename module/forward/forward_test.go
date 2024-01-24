package forward

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	test_caddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestForward_Provision(t *testing.T) {
	var f Forward
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	require.NoError(t, f.Provision(ctx))
}

func TestForward_Start(t *testing.T) {
	var f Forward
	require.Error(t, f.Start(test_caddy.NewTestNetLookup(t)))
}

func TestForward_StartCleanup(t *testing.T) {
	var f Forward
	require.NoError(t, f.Hosts.UnmarshalText([]byte("foo:bar")))
	n := test_caddy.NewTestNetLookup(t)
	n.LookupFn = func(string) (pointc.Net, bool) { return nil, true }
	require.NoError(t, f.Start(n))
	require.NoError(t, f.Cleanup())
}

func TestForward_UnmarshalCaddyfile(t *testing.T) {
	var f Forward
	require.NoError(t, f.UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))
}
