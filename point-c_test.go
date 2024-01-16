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
	pointc "github.com/point-c/caddy"
	"github.com/point-c/simplewg"
	test_caddy "github.com/point-c/test-caddy"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
	"time"
)

func TestPointc_Register(t *testing.T) {
	t.Run("not private network", func(t *testing.T) {
		var pc pointc.Pointc
		n := test_caddy.NewTestNet(t)
		n.LocalAddrFn = func() net.IP { return net.IPv4(1, 1, 1, 1) }
		require.ErrorContains(t, pc.Register("", n), "address is not private network")
	})

	t.Run("ip address collision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet1.Id()+`"}, {"type": "test-`+testNet2.Id()+`"}]}`))
		require.ErrorContains(t, err, "share same address")
	})

	t.Run("network exists", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet1.Id()+`"}, {"type": "test-`+testNet2.Id()+`"}]}`))
		require.ErrorContains(t, err, "share same address")
	})
}

func TestPointcNet_Listen(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	testNet1 := test_caddy.NewTestNetwork(t)
	testN := &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }}
	testN.ListenFn = func(addr *net.TCPAddr) (net.Listener, error) { return nil, nil }
	testN.ListenPacketFn = func(addr *net.UDPAddr) (net.PacketConn, error) { return nil, nil }
	testNet1.StartFn = func(fn pointc.RegisterFunc) error {
		return fn("test1", testN)
	}
	testNet2 := test_caddy.NewTestNetwork(t)
	testNet2.StartFn = func(fn pointc.RegisterFunc) error {
		return fn("test2", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 2) }})
	}
	pcm, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet1.Id()+`"}, {"type": "test-`+testNet2.Id()+`"}]}`))
	require.NoError(t, err)
	pc, ok := pcm.(*pointc.Pointc)
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
			testN.DialerFn = func(ip net.IP, _ uint16) pointc.Dialer {
				require.True(t, net.IPv4(192, 168, 0, 1).Equal(ip))
				return nil
			}
			n.Dialer(net.IPv4(192, 168, 0, 2), 0)
		})
		t.Run("other net", func(t *testing.T) {
			defer func() { testN.DialerFn = nil }()
			ipExp := net.IPv4(192, 168, 1, 1)
			testN.DialerFn = func(ip net.IP, _ uint16) pointc.Dialer { require.True(t, ipExp.Equal(ip)); return nil }
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
		lookup, ok := v.(pointc.NetLookup)
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
		lookup, ok := v.(pointc.NetLookup)
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
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with one network", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet := test_caddy.NewTestNetwork(t)
		testNet.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) }})
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet1.Id()+`"}, {"type": "test-`+testNet2.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks and two net ops, failing on netops provision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		testOp1, testOp2 := test_caddy.NewTestNetOp(t), test_caddy.NewTestNetOp(t)
		expErr := errors.New("json unmarshal fail " + uuid.NewString())
		testOp2.UnmarshalJSONFn = func([]byte) error { return expErr }
		caddy.RegisterModule(testOp1)
		caddy.RegisterModule(testOp2)
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string]any{
			"networks": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet1.Id(), nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet2.Id(), nil),
			},
			"net-ops": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp1.Id(), nil),
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp2.Id(), nil),
			},
		}, nil))
		require.ErrorContains(t, err, expErr.Error())
	})

	t.Run("load network with two networks and two net ops", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		testOp1, testOp2 := test_caddy.NewTestNetOp(t), test_caddy.NewTestNetOp(t)
		caddy.RegisterModule(testOp1)
		caddy.RegisterModule(testOp2)
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string]any{
			"networks": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet1.Id(), nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet2.Id(), nil),
			},
			"net-ops": []any{
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp1.Id(), nil),
				caddyconfig.JSONModuleObject(struct{}{}, "op", testOp2.Id(), nil),
			},
		}, nil))
		require.NoError(t, err)
	})

	t.Run("load network fail with name collision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := test_caddy.NewTestNetwork(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 0) }})
		}
		testNet2 := test_caddy.NewTestNetwork(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", &test_caddy.TestNet{T: t, LocalAddrFn: func() net.IP { return net.IPv4(192, 168, 0, 1) }})
		}
		_, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string][]json.RawMessage{
			"networks": {
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet1.Id(), nil),
				caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+testNet2.Id(), nil),
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
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.Error(t, err)
	})
}

func TestPointc_UnmarshalCaddyfile(t *testing.T) {
	t.Run("bad verb", func(t *testing.T) {
		var pc pointc.Pointc
		require.ErrorContains(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(string(caddyfile.Format([]byte(`point-c foo {
}`))))), "verb")
	})

	t.Run("bad verb", func(t *testing.T) {
		var pc pointc.Pointc
		require.NoError(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))
	})

	testNet := test_caddy.NewTestNetwork(t)
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
}`, testOp.Id()),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.Id()},
					map[string]string{"op": testOp.Id()},
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
`, testOp.Id()),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.Id()},
					map[string]string{"op": testOp.Id()},
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
}`, testNet.ID().Name(), testOp.Id()),
			json: string(caddyconfig.JSON(map[string]any{
				"net-ops": []any{
					map[string]string{"op": testOp.Id()},
					map[string]string{"op": testOp.Id()},
				},
				"networks": []any{
					map[string]string{"type": testNet.ID().Name()},
					map[string]string{"type": testNet.ID().Name()},
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
}`, testNet.ID().Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.ID().Name()),
		},
		{
			name: "point-c",
			caddyfile: fmt.Sprintf(`{
point-c {
	%[1]s
	%[1]s
}
}`, testNet.ID().Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.ID().Name()),
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
			var pc pointc.Pointc
			ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
			defer cancel()
			require.NoError(t, pc.Provision(ctx))
			require.NoError(t, pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(fmt.Sprintf(`point-c {
	%[1]s
}`, testNet.ID().Name()))))
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

func TestSystemNet(t *testing.T) {
	n := pointc.NewSystemNet()
	d := n.Dialer(net.IPv4zero, 0)
	t.Run("tcp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var w simplewg.Wg
		ln, err := n.Listen(&net.TCPAddr{})
		require.NoError(t, err)
		closed := make(chan struct{})
		w.Go(func() {
			defer cancel()
			cn, err := ln.Accept()
			require.NoError(t, err)
			require.NoError(t, cn.Close())
			close(closed)
			_, err = ln.Accept()
			require.Error(t, err)
		})
		w.Go(func() {
			cn, err := d.Dial(ctx, ln.Addr().(*net.TCPAddr))
			require.NoError(t, err)
			<-closed
			require.NoError(t, cn.Close())
			require.NoError(t, ln.Close())
		})
		w.Wait()
	})
	t.Run("udp", func(t *testing.T) {
		var w simplewg.Wg
		ln, err := n.ListenPacket(&net.UDPAddr{})
		require.NoError(t, err)
		closed := make(chan struct{})
		w.Go(func() {
			_, _, err := ln.ReadFrom([]byte{})
			require.NoError(t, err)
			close(closed)
		})
		w.Go(func() {
			cn, err := d.DialPacket(ln.LocalAddr().(*net.UDPAddr))
			require.NoError(t, err)
			_, err = cn.(io.Writer).Write([]byte{})
			require.NoError(t, err)
			<-closed
			require.NoError(t, cn.Close())
			require.NoError(t, ln.Close())
		})
		w.Wait()
	})
}
