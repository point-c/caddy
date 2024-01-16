package point_c_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	pointc "github.com/point-c/caddy"
	"github.com/point-c/test-caddy"
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
			var pc pointc.Listener
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
				"point-c": json.RawMessage(`{"networks": [{"type": "test-` + testNet.Id() + `"}]}`),
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
		tn.UnmarshalJSONFn = func([]byte) error { return nil }
		errExp := errors.New("test err " + uuid.New().String())
		tn.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test", &test_caddy.TestNet{
				T:           t,
				ListenFn:    func(*net.TCPAddr) (net.Listener, error) { return nil, errExp },
				LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) },
			})
		}

		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": caddyconfig.JSON(map[string]any{
					"networks": []any{
						caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+tn.Id(), nil),
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
		tl := test_caddy.NewTestListener(t)
		tn.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test", &test_caddy.TestNet{
				T:           t,
				ListenFn:    func(*net.TCPAddr) (net.Listener, error) { return tl, nil },
				LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) },
			})
		}

		ctx, cancel := test_caddy.NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{"networks": [{"type": "test-` + tn.Id() + `"}]}`),
			},
		})
		defer cancel()
		pcr, err := ctx.App("point-c")
		require.NoError(t, err)
		require.IsType(t, new(pointc.Pointc), pcr)
		pc := pcr.(*pointc.Pointc)
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
