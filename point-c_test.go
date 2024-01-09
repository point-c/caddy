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
	"github.com/stretchr/testify/require"
	"testing"
)

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
		testNet := NewTestNet(t)
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with one network", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet := NewTestNet(t)
		testNet.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := NewTestNet(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
		}
		testNet2 := NewTestNet(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", nil)
		}
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet1.Id()+`"}, {"type": "test-`+testNet2.Id()+`"}]}`))
		require.NoError(t, err)
	})

	t.Run("load network with two networks and two net ops, failing on netops provision", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		testNet1 := NewTestNet(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
		}
		testNet2 := NewTestNet(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", nil)
		}
		testOp1, testOp2 := NewTestNetOp(t), NewTestNetOp(t)
		expErr := errors.New("json unmarshal fail " + uuid.NewString())
		testOp2.unmarshalJSON = func([]byte) error { return expErr }
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
		testNet1 := NewTestNet(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
		}
		testNet2 := NewTestNet(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test2", nil)
		}
		testOp1, testOp2 := NewTestNetOp(t), NewTestNetOp(t)
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
		testNet1 := NewTestNet(t)
		testNet1.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
		}
		testNet2 := NewTestNet(t)
		testNet2.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test1", nil)
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
		testNet := NewTestNet(t)
		testNet.UnmarshalErr = errors.New("test")
		_, err := ctx.LoadModuleByID("point-c", json.RawMessage(`{"networks": [{"type": "test-`+testNet.Id()+`"}]}`))
		require.Error(t, err)
	})
}

func TestPointc_UnmarshalCaddyfile(t *testing.T) {
	testNet := NewTestNet(t)
	testOp := NewTestNetOp(t)
	caddy.RegisterModule(testOp)

	tests := []struct {
		name      string
		caddyfile string
		json      string
		wantErr   bool
	}{
		{name: "nothing", json: "{}"},
		{
			name:      "bad name",
			caddyfile: "foo {\n}\n",
			wantErr:   true,
		},
		{
			name: "2x netop",
			caddyfile: fmt.Sprintf(`netop {
	%[1]s
}
netop {
	%[1]s
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
			name: "netop",
			caddyfile: fmt.Sprintf(`netop {
	%[1]s
	%[1]s
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
			caddyfile: fmt.Sprintf(`point-c {
	%[1]s
}
netop {
	%[2]s
}
point-c {
	%[1]s
}
netop {
	%[2]s
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
			caddyfile: fmt.Sprintf(`point-c {
	%[1]s
}
point-c {
	%[1]s
}`, testNet.ID().Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.ID().Name()),
		},
		{
			name: "point-c",
			caddyfile: fmt.Sprintf(`point-c {
	%[1]s
	%[1]s
}`, testNet.ID().Name()),
			json: fmt.Sprintf(`{"networks": [{"type": "%[1]s"}, {"type": "%[1]s"}]}`, testNet.ID().Name()),
		},
		{
			name: "point c submodule does not exist",
			caddyfile: `point-c {
	foobar
}`,
			wantErr: true,
		},
		{
			name: "net op submodule does not exist",
			caddyfile: `netop {
	foobar
}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pc pointc.Pointc
			ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
			defer cancel()
			require.NoError(t, pc.Provision(ctx))
			if err := pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tt.caddyfile)); tt.wantErr {
				require.Errorf(t, err, "UnmarshalCaddyfile() wantErr %v", tt.wantErr)
				return
			} else {
				require.NoError(t, err, "UnmarshalCaddyfile()")
			}
			require.JSONEq(t, tt.json, string(caddyconfig.JSON(&pc, nil)), "caddyfile != json")
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

type TestNetOp struct {
	t                  *testing.T
	startFn            func(pointc.NetLookup) error
	stopFn             func() error
	unmarshalCaddyfile func(*caddyfile.Dispenser) error
	unmarshalJSON      func(b []byte) error
	id                 uuid.UUID
}

func (t *TestNetOp) CaddyModule() caddy.ModuleInfo {
	t.t.Helper()
	return caddy.ModuleInfo{
		ID:  caddy.ModuleID("point-c.op." + t.id.String()),
		New: func() caddy.Module { t.t.Helper(); return t },
	}
}

func NewTestNetOp(t *testing.T) *TestNetOp {
	t.Helper()
	return &TestNetOp{
		t:  t,
		id: uuid.New(),
	}
}

func (t *TestNetOp) Id() string { t.t.Helper(); return t.id.String() }

func (t *TestNetOp) Start(l pointc.NetLookup) error {
	t.t.Helper()
	if t.startFn != nil {
		return t.startFn(l)
	}
	return nil
}

func (t *TestNetOp) Stop() error {
	t.t.Helper()
	if t.stopFn != nil {
		return t.stopFn()
	}
	return nil
}

func (t *TestNetOp) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	t.t.Helper()
	if t.unmarshalCaddyfile != nil {
		return t.unmarshalCaddyfile(d)
	}
	return nil
}

func (t *TestNetOp) UnmarshalJSON(b []byte) error {
	t.t.Helper()
	if t.unmarshalJSON != nil {
		return t.unmarshalJSON(b)
	}
	return nil
}

func (t *TestNetwork) CaddyModule() caddy.ModuleInfo {
	t.t.Helper()
	return caddy.ModuleInfo{
		ID:  t.ID(),
		New: func() caddy.Module { t.t.Helper(); return t },
	}
}

func (t *TestNetwork) ID() caddy.ModuleID {
	t.t.Helper()
	return caddy.ModuleID("point-c.net.test-" + t.id.String())
}

func (t *TestNetwork) Id() string { return t.id.String() }

type TestNetwork struct {
	t                     testing.TB
	id                    uuid.UUID
	UnmarshalCaddyfileErr error                           `json:"-"`
	UnmarshalErr          error                           `json:"-"`
	StartFn               func(pointc.RegisterFunc) error `json:"-"`
	StopFn                func() error                    `json:"-"`
}

func (t *TestNetwork) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	t.t.Helper()
	return t.UnmarshalCaddyfileErr
}

func (t *TestNetwork) UnmarshalJSON([]byte) error {
	t.t.Helper()
	return t.UnmarshalErr
}

func NewTestNet(t testing.TB) *TestNetwork {
	t.Helper()
	tn := &TestNetwork{
		t:  t,
		id: uuid.New(),
	}
	caddy.RegisterModule(tn)
	return tn
}

func (t *TestNetwork) Stop() error {
	if t.StopFn != nil {
		return t.StopFn()
	}
	return nil
}

func (t *TestNetwork) Start(fn pointc.RegisterFunc) error {
	if t.StartFn != nil {
		return t.StartFn(fn)
	}
	return nil
}
