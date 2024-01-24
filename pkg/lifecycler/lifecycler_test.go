package lifecycler_test

import (
	"context"
	"encoding/json"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/point-c/caddy/pkg/lifecycler"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type BaseTestModule struct {
	id  caddy.ModuleID
	new caddy.Module
}

type TestModule[T any] struct {
	BaseTestModule
	start              func(T) error
	stop               func() error
	unmarshalCaddyfile func(d *caddyfile.Dispenser) error
	unmarhsalJSON      func([]byte) error
	SubModules         []json.RawMessage `json:"submodules,omitempty" caddy:"namespace=submodules inline_key=type"`
}

func (t *TestModule[T]) UnmarshalJSON(b []byte) error {
	if t.unmarhsalJSON == nil {
		type rawType TestModule[T]
		var raw rawType
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		t.SubModules = raw.SubModules
		return nil
	}
	return t.unmarhsalJSON(b)
}

func NewTestModule[T any]() *TestModule[T] {
	tm := &TestModule[T]{}
	tm.BaseTestModule = BaseTestModule{
		id:  caddy.ModuleID(uuid.New().String()),
		new: tm,
	}
	return tm
}

func (t *TestModule[T]) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if t.unmarshalCaddyfile != nil {
		return t.unmarshalCaddyfile(d)
	}
	return nil
}

func (t *BaseTestModule) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  t.id,
		New: func() caddy.Module { return t.new },
	}
}

func (t *TestModule[T]) Start(v T) error {
	if t.start != nil {
		return t.start(v)
	}
	return nil
}

func (t *TestModule[T]) Stop() error {
	if t.stop != nil {
		return t.stop()
	}
	return nil
}

func TestLifeCycler_JSON(t *testing.T) {
	var lf lifecycler.LifeCycler[any]
	require.Error(t, lf.UnmarshalJSON(nil))
	_, err := lf.MarshalJSON()
	require.Error(t, err)
}

func TestLifeCycler_SetValue(t *testing.T) {
	var lf lifecycler.LifeCycler[int]
	require.Equal(t, 0, lf.V)
	lf.SetValue(123)
	require.Equal(t, 123, lf.V)
}

func TestLifeCycler_Start(t *testing.T) {
	t.Run("one ok", func(t *testing.T) {
		lf := lifecycler.LifeCycler[int]{
			V: 123,
			Modules: []lifecycler.LifeCyclable[int]{
				&TestModule[int]{},
			},
		}
		require.NoError(t, lf.Start())
	})
	t.Run("two ok", func(t *testing.T) {
		lf := lifecycler.LifeCycler[int]{
			V: 123,
			Modules: []lifecycler.LifeCyclable[int]{
				&TestModule[int]{},
				&TestModule[int]{},
			},
		}
		require.NoError(t, lf.Start())
	})
	t.Run("error", func(t *testing.T) {
		lf := lifecycler.LifeCycler[int]{
			V: 123,
			Modules: []lifecycler.LifeCyclable[int]{
				&TestModule[int]{},
				&TestModule[int]{start: func(int) error { return errors.New("") }},
			},
		}
		require.Error(t, lf.Start())
	})
	t.Run("panic", func(t *testing.T) {
		lf := lifecycler.LifeCycler[int]{
			V: 123,
			Modules: []lifecycler.LifeCyclable[int]{
				&TestModule[int]{start: func(int) error { panic("") }},
				&TestModule[int]{},
			},
		}
		require.Error(t, lf.Start())
	})
}

func TestLifeCycler_UnmarshalCaddyfile(t *testing.T) {
	t.Run("invalid info", func(t *testing.T) {
		t.Run("nil info", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.Error(t, lf.UnmarshalCaddyfile(nil, nil))
		})
		t.Run("nil AppendModule", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.Error(t, lf.UnmarshalCaddyfile(nil, &lifecycler.CaddyfileInfo{
				SubModuleSpecifier: "f",
			}))
		})
		t.Run("empty SubModuleSpecifier", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.Error(t, lf.UnmarshalCaddyfile(nil, &lifecycler.CaddyfileInfo{
				Raw: new([]json.RawMessage),
			}))
		})
	})

	t.Run("verb parse line", func(t *testing.T) {
		t.Run("no next arg", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.Error(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`foo {
}`), &lifecycler.CaddyfileInfo{
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
				ParseVerbLine:      &TestVerb{},
			}))
		})
		t.Run("error unmarshalling", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.Error(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`foo bar {
}`), &lifecycler.CaddyfileInfo{
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
				ParseVerbLine:      &TestVerb{unmarshalErr: errors.New("")},
			}))
		})
		t.Run("valid", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			require.NoError(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`foo bar {
}`), &lifecycler.CaddyfileInfo{
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
				ParseVerbLine:      &TestVerb{},
			}))
		})
	})

	t.Run("parse sub modules", func(t *testing.T) {
		t.Run("fail to unmarshal", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			m := NewTestModule[any]()
			caddy.RegisterModule(m)
			m.unmarshalCaddyfile = func(*caddyfile.Dispenser) error { return errors.New("test error") }
			require.Error(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`foo bar {
	`+string(m.id)+`
}`), &lifecycler.CaddyfileInfo{
				ModuleID:           []string{string(m.id)},
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
				ParseVerbLine:      &TestVerb{},
			}))
		})
		t.Run("one", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			parent := NewTestModule[any]()
			caddy.RegisterModule(parent)
			m := NewTestModule[any]()
			m.id = caddy.ModuleID(strings.Join([]string{string(parent.id), string(m.id)}, "."))
			caddy.RegisterModule(m)
			require.NoError(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(string(parent.id)+` {
	`+m.id.Name()+`
}`), &lifecycler.CaddyfileInfo{
				ModuleID:           []string{string(parent.id)},
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
			}))
		})
		t.Run("two", func(t *testing.T) {
			var lf lifecycler.LifeCycler[any]
			parent := NewTestModule[any]()
			caddy.RegisterModule(parent)
			m1 := NewTestModule[any]()
			m1.id = caddy.ModuleID(strings.Join([]string{string(parent.id), string(m1.id)}, "."))
			caddy.RegisterModule(m1)
			m2 := NewTestModule[any]()
			m2.id = caddy.ModuleID(strings.Join([]string{string(parent.id), string(m2.id)}, "."))
			caddy.RegisterModule(m2)
			require.NoError(t, lf.UnmarshalCaddyfile(caddyfile.NewTestDispenser(string(parent.id)+` {
	`+m1.id.Name()+`
	`+m1.id.Name()+`
}`), &lifecycler.CaddyfileInfo{
				ModuleID:           []string{string(parent.id)},
				Raw:                new([]json.RawMessage),
				SubModuleSpecifier: "f",
			}))
		})
	})
}

func TestLifeCycler_Cleanup(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		lf := lifecycler.LifeCycler[int]{
			V: 123,
			Started: []lifecycler.LifeCyclable[int]{
				&TestModule[int]{},
			},
		}
		require.NoError(t, lf.Cleanup())
		require.Equal(t, lifecycler.LifeCycler[int]{}, lf)
	})
}

func TestLifeCycler_Provision(t *testing.T) {
	t.Run("invalid info", func(t *testing.T) {
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.Error(t, lf.Provision(ctx, nil))
	})
	t.Run("invalid raw", func(t *testing.T) {
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.Error(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{}))
	})
	t.Run("nil raw", func(t *testing.T) {
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.NoError(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{Raw: new([]json.RawMessage)}))
	})
	t.Run("one module", func(t *testing.T) {
		tm := NewTestModule[any]()
		caddy.RegisterModule(tm)
		stm := NewTestModule[any]()
		tm.SubModules = []json.RawMessage{caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm.id), nil)}
		stm.id = "submodules." + stm.id
		caddy.RegisterModule(stm)
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.NoError(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{
			StructPointer: tm,
			FieldName:     "SubModules",
			Raw:           &tm.SubModules,
		}))
	})
	t.Run("two module", func(t *testing.T) {
		tm := NewTestModule[any]()
		caddy.RegisterModule(tm)
		stm1 := NewTestModule[any]()
		stm2 := NewTestModule[any]()
		tm.SubModules = []json.RawMessage{
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm1.id), nil),
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm2.id), nil),
		}
		stm1.id = "submodules." + stm1.id
		stm2.id = "submodules." + stm2.id
		caddy.RegisterModule(stm1)
		caddy.RegisterModule(stm2)
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.NoError(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{
			StructPointer: tm,
			FieldName:     "SubModules",
			Raw:           &tm.SubModules,
		}))
	})
	t.Run("submodule fails to unmarshal", func(t *testing.T) {
		tm := NewTestModule[any]()
		caddy.RegisterModule(tm)
		stm1 := NewTestModule[any]()
		stm2 := NewTestModule[any]()
		stm2.unmarhsalJSON = func([]byte) error { return errors.New("test json fail") }
		tm.SubModules = []json.RawMessage{
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm1.id), nil),
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm2.id), nil),
		}
		stm1.id = "submodules." + stm1.id
		stm2.id = "submodules." + stm2.id
		caddy.RegisterModule(stm1)
		caddy.RegisterModule(stm2)
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.Error(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{
			StructPointer: tm,
			FieldName:     "SubModules",
			Raw:           &tm.SubModules,
		}))
	})
	t.Run("submodule is wrong type interface", func(t *testing.T) {
		tm := NewTestModule[any]()
		caddy.RegisterModule(tm)
		stm1 := NewTestModule[any]()
		stm2 := &BaseTestModule{id: caddy.ModuleID(uuid.New().String())}
		stm2.new = stm2
		tm.SubModules = []json.RawMessage{
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm1.id), nil),
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm2.id), nil),
		}
		stm1.id = "submodules." + stm1.id
		stm2.id = "submodules." + stm2.id
		caddy.RegisterModule(stm1)
		caddy.RegisterModule(stm2)
		var lf lifecycler.LifeCycler[any]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.Error(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{
			StructPointer: tm,
			FieldName:     "SubModules",
			Raw:           &tm.SubModules,
		}))
	})
	t.Run("submodule is wrong type", func(t *testing.T) {
		tm := NewTestModule[int]()
		caddy.RegisterModule(tm)
		stm1 := NewTestModule[string]()
		stm2 := NewTestModule[string]()
		tm.SubModules = []json.RawMessage{
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm1.id), nil),
			caddyconfig.JSONModuleObject(struct{}{}, "type", string(stm2.id), nil),
		}
		stm1.id = "submodules." + stm1.id
		stm2.id = "submodules." + stm2.id
		caddy.RegisterModule(stm1)
		caddy.RegisterModule(stm2)
		var lf lifecycler.LifeCycler[int]
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		require.Error(t, lf.Provision(ctx, &lifecycler.ProvisionInfo{
			StructPointer: tm,
			FieldName:     "SubModules",
			Raw:           &tm.SubModules,
		}))
	})
}

type TestVerb struct {
	unmarshalErr error
}

func (t *TestVerb) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return t.unmarshalErr
}
