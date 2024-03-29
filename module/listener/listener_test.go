package listener

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/point-c/caddy/module/point-c"
	test_caddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	_ "unsafe"
)

func TestListener_UnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name      string
		caddyfile string
		json      string
		wantErr   bool
	}{
		{
			name:      "basic",
			caddyfile: "point-c remote 80",
			json:      `{"name": "remote", "port": 80}`,
		},
		{
			name:      "no port",
			caddyfile: "point-c remote",
			wantErr:   true,
		},
		{
			name:      "no network name",
			caddyfile: "point-c",
			wantErr:   true,
		},
		{
			name:      "non numeric port",
			caddyfile: "point-c remote aa",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pc Listener
			if err := pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tt.caddyfile)); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				b, err := json.Marshal(pc)
				require.NoError(t, err)
				require.JSONEq(t, tt.json, string(b))
			}
		})
	}
}

func TestListener_Provision(t *testing.T) {
	t.Run("point-c not configured correctly", func(t *testing.T) {
		testNet := test_caddy.NewTestNetwork(t)
		testNet.UnmarshalJSONFn = func([]byte) error { return errors.New("test") }
		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{"networks": [{"type": "` + testNet.ID + `"}]}`),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.point-c", json.RawMessage("{}"))
		require.Error(t, err)
	})

	t.Run("network not found", func(t *testing.T) {
		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{}`),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.point-c", json.RawMessage("{}"))
		require.Error(t, err)
	})

	t.Run("network errors on listen", func(t *testing.T) {
		tn := test_caddy.NewTestNetwork(t)
		tn.Register()
		tn.UnmarshalJSONFn = func([]byte) error { return nil }
		errExp := errors.New("test err " + uuid.New().String())
		tn.StartFn = func(fn point_c.RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.ListenFn = func(*net.TCPAddr) (net.Listener, error) { return nil, errExp }
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test", n)
		}

		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": caddyconfig.JSON(map[string]any{
					"networks": []any{
						caddyconfig.JSONModuleObject(struct{}{}, "type", tn.ID, nil),
					},
				}, nil),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.point-c", json.RawMessage(`{"name": "test", "port": 80}`))
		require.ErrorContains(t, err, errExp.Error())
	})

	t.Run("valid", func(t *testing.T) {
		tn := test_caddy.NewTestNetwork(t)
		tn.Register()
		tl := test_caddy.NewTestListener(t)
		tn.StartFn = func(fn point_c.RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.ListenFn = func(*net.TCPAddr) (net.Listener, error) { return tl, nil }
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test", n)
		}

		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{"networks": [{"type": "` + tn.ID + `"}]}`),
			},
		})
		defer cancel()
		pcr, err := ctx.App("point-c")
		require.NoError(t, err)
		require.IsType(t, new(point_c.Pointc), pcr)
		pc := pcr.(*point_c.Pointc)
		require.NoError(t, pc.Start())
		defer func() { require.NoError(t, pc.Stop()) }()

		a, err := ctx.LoadModuleByID("caddy.listeners.merge.point-c", json.RawMessage(`{"name": "test", "port": 80}`))
		require.NoError(t, err)
		_, ok := a.(net.Listener)
		require.True(t, ok)
		require.Equal(t, tl.Close(), a.(net.Listener).Close())
		require.Equal(t, tl.Addr(), a.(net.Listener).Addr())
		ec, ee := tl.Accept()
		gc, ge := a.(net.Listener).Accept()
		require.Equal(t, ec, gc)
		require.Equal(t, ee, ge)
	})
}

func TestListener_Start(t *testing.T) {
	var l Listener
	require.NoError(t, l.Start(func(ln net.Listener) {
		require.Equal(t, &l, ln)
	}))
}
