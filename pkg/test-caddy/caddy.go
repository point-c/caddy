package test_caddy

import (
	"context"
	"encoding/json"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/point-c/caddy/pkg/lifecycler"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"unsafe"
)

func NewCaddyContext(t testing.TB, base context.Context, cfg caddy.Config) (caddy.Context, context.CancelFunc) {
	t.Helper()
	ctx := caddy.Context{Context: base}

	// Set unexposed 'cfg' field
	f, ok := reflect.TypeOf(ctx).FieldByName("cfg")
	require.True(t, ok)
	cfgPtr := uintptr(reflect.ValueOf(&ctx).UnsafePointer()) + f.Offset
	*(**caddy.Config)(unsafe.Pointer(cfgPtr)) = &cfg

	// Initialize 'apps' map
	f, ok = reflect.TypeOf(cfg).FieldByName("apps")
	require.True(t, ok)
	appsPtr := uintptr(reflect.ValueOf(&cfg).UnsafePointer()) + f.Offset
	*(*map[string]caddy.App)(unsafe.Pointer(appsPtr)) = map[string]caddy.App{}

	// Return new context
	return caddy.NewContext(ctx)
}

var (
	_ caddy.Module                 = (*TestModule[any])(nil)
	_ caddyfile.Unmarshaler        = (*TestModule[any])(nil)
	_ caddy.Provisioner            = (*TestModule[any])(nil)
	_ caddy.Validator              = (*TestModule[any])(nil)
	_ lifecycler.LifeCyclable[any] = (*TestModule[any])(nil)
	_ caddy.CleanerUpper           = (*TestModule[any])(nil)
	_ json.Unmarshaler             = (*TestModule[any])(nil)
	_ json.Marshaler               = (*TestModule[any])(nil)
)

type TestModule[T any] struct {
	t                    testing.TB
	ID                   string         `json:"-"`
	Module               caddy.ModuleID `json:"-"`
	New                  func() caddy.Module
	UnmarshalCaddyfileFn func(*caddyfile.Dispenser) error `json:"-"`
	ProvisionerFn        func(caddy.Context) error        `json:"-"`
	ValidateFn           func() error                     `json:"-"`
	CleanupFn            func() error                     `json:"-"`
	StartFn              func(T) error                    `json:"-"`
	MarshalJSONFn        func() ([]byte, error)           `json:"-"`
	UnmarshalJSONFn      func(b []byte) error             `json:"-"`
}

func NewTestModule[T any, Parent caddy.Module](t testing.TB, p *Parent, fn func(Parent) *TestModule[T], module string) {
	t.Helper()
	id := "test" + uuid.NewString()
	*fn(*p) = TestModule[T]{
		t:      t,
		ID:     id,
		Module: caddy.ModuleID(module + id),
		New:    func() caddy.Module { t.Helper(); return *p },
	}
}

func (t *TestModule[T]) MarshalJSON() ([]byte, error) {
	if t.MarshalJSONFn != nil {
		return t.MarshalJSONFn()
	}
	return json.Marshal(struct{}{})
}

func (t *TestModule[T]) UnmarshalJSON(b []byte) error {
	if t.UnmarshalJSONFn != nil {
		return t.UnmarshalJSONFn(b)
	}
	return nil
}

func (t *TestModule[T]) Register() { t.t.Helper(); caddy.RegisterModule(t) }

func (t *TestModule[T]) CaddyModule() caddy.ModuleInfo {
	t.t.Helper()
	return caddy.ModuleInfo{
		ID:  t.Module,
		New: func() caddy.Module { t.t.Helper(); return t },
	}
}

func (t *TestModule[T]) Provision(ctx caddy.Context) error {
	if t.ProvisionerFn != nil {
		return t.ProvisionerFn(ctx)
	}
	return nil
}

func (t *TestModule[T]) Validate() error {
	if t.ValidateFn != nil {
		return t.ValidateFn()
	}
	return nil
}

func (t *TestModule[T]) Start(v T) error {
	if t.StartFn != nil {
		return t.StartFn(v)
	}
	return nil
}

func (t *TestModule[T]) Cleanup() error {
	if t.CleanupFn != nil {
		return t.CleanupFn()
	}
	return nil
}

func (t *TestModule[T]) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if t.UnmarshalCaddyfileFn != nil {
		return t.UnmarshalCaddyfileFn(d)
	}
	return nil
}
