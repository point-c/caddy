package point_c_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	. "github.com/point-c/caddy/module/point-c"
	test_caddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestPointc_Register(t *testing.T) {
	t.Run("network exists", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.Register()
		testNet1.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test1", n)
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.Register()
		testNet2.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test1", n)
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "`+testNet1.ID+`"}, {"type": "`+testNet2.ID+`"}]}`))
		require.Error(t, err)
	})
}

func TestPointcNet_Listen(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	testNet1 := test_caddy.NewTestNetwork(t)
	testNet1.Register()
	testN := test_caddy.NewTestNet(t)
	testN.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
	testN.ListenFn = func(addr *net.TCPAddr) (net.Listener, error) { return nil, nil }
	testN.ListenPacketFn = func(addr *net.UDPAddr) (net.PacketConn, error) { return nil, nil }
	testNet1.StartFn = func(fn RegisterFunc) error {
		return fn("test1", testN)
	}
	testNet2 := test_caddy.NewTestNetwork(t)
	testNet2.Register()
	testNet2.StartFn = func(fn RegisterFunc) error {
		n := test_caddy.NewTestNet(t)
		n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 2) }
		return fn("test2", n)
	}
	pcm, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "`+testNet1.ID+`"}, {"type": "`+testNet2.ID+`"}]}`))
	require.NoError(t, err)
	pc, ok := pcm.(*Pointc)
	require.True(t, ok)
	n, ok := pc.Lookup("test1")
	require.True(t, ok)

	t.Run("listen", func(t *testing.T) {
		t.Run("tcp", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				defer func() { testN.ListenFn = nil }()
				testN.ListenFn = func(addr *net.TCPAddr) (net.Listener, error) { return nil, nil }
				_, err := n.Listen(&net.TCPAddr{IP: net.IPv4zero})
				require.NoError(t, err)
			})
		})
		t.Run("udp", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				defer func() { testN.ListenPacketFn = nil }()
				testN.ListenPacketFn = func(addr *net.UDPAddr) (net.PacketConn, error) { return nil, nil }
				_, err := n.ListenPacket(&net.UDPAddr{IP: net.IPv4zero})
				require.NoError(t, err)
			})
		})
	})

	t.Run("dialer", func(t *testing.T) {
		t.Run("local net", func(t *testing.T) {
			defer func() { testN.DialerFn = nil }()
			testN.DialerFn = func(ip net.IP, _ uint16) Dialer {
				require.True(t, net.IPv4(192, 168, 0, 2).Equal(ip))
				return nil
			}
			n.Dialer(net.IPv4(192, 168, 0, 2), 0)
		})
		t.Run("other net", func(t *testing.T) {
			defer func() { testN.DialerFn = nil }()
			ipExp := net.IPv4(192, 168, 1, 1)
			testN.DialerFn = func(ip net.IP, _ uint16) Dialer { require.True(t, ipExp.Equal(ip)); return nil }
			n.Dialer(ipExp, 0)
		})
	})
}

func TestPointc_StartStop(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	v, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{}`))
	require.NoError(t, err)
	app, ok := v.(caddy.App)
	require.True(t, ok)
	require.NoError(t, app.Start())
	require.NoError(t, app.Stop())
}

func TestPointc_Lookup(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		v, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{}`))
		require.NoError(t, err)
		lookup, ok := v.(NetLookup)
		require.True(t, ok)
		n, ok := lookup.Lookup("")
		require.False(t, ok)
		require.Nil(t, n)
	})

	t.Run("exists", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		v, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{}`))
		require.NoError(t, err)
		lookup, ok := v.(NetLookup)
		require.True(t, ok)
		n, ok := lookup.Lookup("")
		require.False(t, ok)
		require.Nil(t, n)
	})
}

func TestPointc_Provision(t *testing.T) {
	t.Run("null networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{}`))
		require.NoError(t, err)
	})

	t.Run("empty network slice networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": []}`))
		require.NoError(t, err)
	})

	t.Run("load network with no networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet := test_caddy.NewTestNetwork(t)
		testNet.Register()
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "`+testNet.ID+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with one network", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet := test_caddy.NewTestNetwork(t)
		testNet.Register()
		testNet.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test1", n)
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "`+testNet.ID+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.Register()
		testNet1.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test1", n)
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.Register()
		testNet2.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test2", n)
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "`+testNet1.ID+`"}, {"type": "`+testNet2.ID+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks and two net ops, failing on netops provision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.Register()
		testNet1.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test1", n)
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.Register()
		testNet2.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test2", n)
		}
		testOp1, testOp2 := test_caddy.NewTestNetOp(t), test_caddy.NewTestNetOp(t)
		expErr := errors.New("json unmarshal fail " + uuid.NewString())
		testOp2.UnmarshalJSONFn = func([]byte) error { return expErr }
		caddy.RegisterModule(testOp1)
		caddy.RegisterModule(testOp2)
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string]any{
			"networks": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet1.ID, nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet2.ID, nil),
			},
			"net-ops": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp1.ID, nil),
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp2.ID, nil),
			},
		}, nil))
		require.ErrorContains(t, err, expErr.Error())
	})

	t.Run("load network with two networks and two net ops", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.Register()
		testNet1.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test1", n)
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.Register()
		testNet2.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test2", n)
		}
		testOp1, testOp2 := test_caddy.NewTestNetOp(t), test_caddy.NewTestNetOp(t)
		caddy.RegisterModule(testOp1)
		caddy.RegisterModule(testOp2)
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string]any{
			"networks": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet1.ID, nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet2.ID, nil),
			},
			"net-ops": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp1.ID, nil),
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp2.ID, nil),
			},
		}, nil))
		require.NoError(t, err)
	})

	t.Run("load network fail with name collision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.Register()
		testNet1.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 0) }
			return fn("test1", n)
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.Register()
		testNet2.StartFn = func(fn RegisterFunc) error {
			n := test_caddy.NewTestNet(t)
			n.LocalAddrFn = func() net.IP { return net.IPv4(192, 168, 0, 1) }
			return fn("test1", n)
		}
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string][]json.RawMessage{
			"networks": {
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet1.ID, nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", testNet2.ID, nil),
			},
		}, nil))
		require.Error(t, err)
		require.ErrorContains(t, err, "network \"test1\" already exists")
	})

	t.Run("load network fails", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet := test_caddy.NewTestNetwork(t)
		testNet.UnmarshalJSONFn = func([]byte) error { return errors.New("test") }
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.ID+`"}]}`))
		require.Error(t, err)
	})
}

func TestPointc_UnmarshalCaddyfile(t *testing.T) {
	t.Run("bad verb", func(t *testing.T) {
		var pc Pointc
		require.ErrorContains(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(string(caddyfile.Format([]byte(`point-c foo {
}`))))), "verb")
	})

	t.Run("bad verb", func(t *testing.T) {
		var pc Pointc
		require.NoError(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))
	})

	testNet := test_caddy.NewTestNetwork(t)
	testNet.Register()
	testOp := test_caddy.NewTestNetOp(t)
	caddy.RegisterModule(testOp)

	tests := []struct {
		name      string
		caddyfile string
		json      string
		wantErr   bool
	}{
		{
			name: "2x netop",
			caddyfile: fmt.Sprintf(`{
	point-c netops {
		%[1]s
	}
	point-c netops {
		%[1]s
	}
}`, testOp.ID),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.ID},
					map[string]string{"op": testOp.ID},
				},
			}, nil)),
		},
		{
			name: "netop",
			caddyfile: fmt.Sprintf(`{
point-c netops {
	%[1]s
	%[1]s
}
}
`, testOp.ID),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.ID},
					map[string]string{"op": testOp.ID},
				},
			}, nil)),
		},
		{
			name: "2x point-c & 2x net op",
			caddyfile: fmt.Sprintf(`{
	point-c {
		%[1]s
	}
	point-c netops {
		%[2]s
	}
	point-c {
		%[1]s
	}
	point-c netops {
		%[2]s
	}
}`, testNet.Module.Name(), testOp.ID),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.ID},
					map[string]string{"op": testOp.ID},
				},
				"networks": []any{
					map[string]string{"type": testNet.Module.Name()},
					map[string]string{"type": testNet.Module.Name()},
				},
			}, nil)),
		},
		{
			name: "2x point-c",
			caddyfile: fmt.Sprintf(`{
	point-c {
		%[1]s
	}
	point-c {
		%[1]s
	}
}`, testNet.Module.Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.Module.Name()),
		},
		{
			name: "point-c",
			caddyfile: fmt.Sprintf(`{
point-c {
	%[1]s
	%[1]s
}
}`, testNet.Module.Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.Module.Name()),
		},
		{
			name: "point c submodule does not exist",
			caddyfile: `{
point-c {
	foobar
}
}`,
			wantErr: true,
		},
		{
			name: "net op submodule does not exist",
			caddyfile: `{
netop {
	foobar
}
}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.caddyfile = string(caddyfile.Format([]byte(tt.caddyfile)))
			adapter := caddyconfig.GetAdapter("caddyfile")
			require.NotNil(t, adapter)
			b, warn, err := adapter.Adapt([]byte(tt.caddyfile), nil)
			if tt.wantErr {
				if err == nil && len(warn) == 0 {
					require.Error(t, err)
					require.NotEmpty(t, warn)
				}
			} else {
				require.NoError(t, err)
				require.Empty(t, warn)
				require.JSONEq(t, string(caddyconfig.JSON(map[string]any{"apps": map[string]any{"point-c": json.RawMessage(tt.json)}}, nil)), string(b))
			}
		})
	}

	t.Run("full", func(t *testing.T) {
		b := func() []byte {
			var pc Pointc
			ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
			defer cancel()
			require.NoError(t, pc.Provision(ctx))
			require.NoError(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(fmt.Sprintf(`point-c {
	%[1]s
}`, testNet.Module.Name()))))
			b, err := json.Marshal(pc)
			require.NoError(t, err)
			return b
		}()

		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		_, err := ctx.LoadModuleByID("point-c", b)
		require.NoError(t, err)
	})
}
