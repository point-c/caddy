package test_caddy

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/point-c/caddy/pkg/lifecycler"
	"reflect"
	"testing"
	"unsafe"
)

func NewCaddyContext(t testing.TB, base context.Context, cfg caddy.Config) (caddy.Context, context.CancelFunc) {
	t.Helper()
	ctx := caddy.Context{Context: base}

	// Set unexposed 'cfg' field
	f, ok := reflect.TypeOf(ctx).FieldByName("cfg")
	if !ok {
		t.Fatal("cannot populate config")
	}
	cfgPtr := uintptr(reflect.ValueOf(&ctx).UnsafePointer()) + f.Offset
	*(**caddy.Config)(unsafe.Pointer(cfgPtr)) = &cfg

	// Initialize 'apps' map
	f, ok = reflect.TypeOf(cfg).FieldByName("apps")
	if !ok {
		t.Fatal("cannot initialize apps")
	}
	appsPtr := uintptr(reflect.ValueOf(&cfg).UnsafePointer()) + f.Offset
	*(*map[string]caddy.App)(unsafe.Pointer(appsPtr)) = map[string]caddy.App{}

	// Return new context
	return caddy.NewContext(ctx)
}

var (
	_ caddy.CleanerUpper           = (*TestModule[any])(nil)
	_ caddy.Module                 = (*TestModule[any])(nil)
	_ caddy.Provisioner            = (*TestModule[any])(nil)
	_ caddy.Validator              = (*TestModule[any])(nil)
	_ caddyfile.Unmarshaler        = (*TestModule[any])(nil)
	_ lifecycler.LifeCyclable[any] = (*TestModule[any])(nil)
)

type TestModule[T any] struct {
	t                    testing.TB
	ID                   string                           `json:"-"`
	Module               caddy.ModuleID                   `json:"-"`
	ProvisionerFn        func(caddy.Context) error        `json:"-"`
	ValidateFn           func() error                     `json:"-"`
	UnmarshalCaddyfileFn func(*caddyfile.Dispenser) error `json:"-"`
	StartFn              func(T) error                    `json:"-"`
	CleanupFn            func() error                     `json:"-"`
}

func NewTestModule[T any](t testing.TB, module string) *TestModule[T] {
	t.Helper()
	id := "test" + uuid.NewString()
	return &TestModule[T]{
		t:      t,
		ID:     id,
		Module: caddy.ModuleID(module + id),
	}
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
