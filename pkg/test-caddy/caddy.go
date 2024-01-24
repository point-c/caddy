// Package test_caddy provides mock types and utilities for testing caddy modules.
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

// NewCaddyContext returns an instantiated caddy context.
// This differs from [caddy.NewContext] in that the context's config is set and the `apps` map is initialized.
// This is required to load other apps in provision for some tests.
//
// UNSTABLE: Heavy use of reflect and unsafe pointers to access unexported struct fields.
func NewCaddyContext(t testing.TB, base context.Context, cfg caddy.Config) (caddy.Context, context.CancelFunc) {
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

// TestModule is a mock base caddy module. It implements the functions in the caddy lifecycle.
// It also implements [lifecycler.LifeCyclable] and the json marshalling methods.
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

// NewTestModule creates and initializes a new instance of [TestModule].
func NewTestModule[T any, Parent caddy.Module](t testing.TB, p *Parent, fn func(Parent) *TestModule[T], module string) {
	id := "test" + uuid.NewString()
	*fn(*p) = TestModule[T]{
		t:      t,
		ID:     id,
		Module: caddy.ModuleID(module + id),
		New:    func() caddy.Module { return *p },
	}
}

// MarshalJSON attempts to call MarshalJSONFn. If MarshalJSONFn is not set, an empty struct is marshalled and returned instead.
func (t *TestModule[T]) MarshalJSON() ([]byte, error) {
	if t.MarshalJSONFn != nil {
		return t.MarshalJSONFn()
	}
	return json.Marshal(struct{}{})
}

// UnmarshalJSON attempts to call UnmarshalJSONFn. If UnmarshalJSONFn is not set, nil is returned.
func (t *TestModule[T]) UnmarshalJSON(b []byte) error {
	if t.UnmarshalJSONFn != nil {
		return t.UnmarshalJSONFn(b)
	}
	return nil
}

// Register registers this module with caddy.
func (t *TestModule[T]) Register() { caddy.RegisterModule(t) }

// CaddyModule returns the [caddy.ModuleInfo] for this module.
// The New method returns this instance everytime.
func (t *TestModule[T]) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  t.Module,
		New: func() caddy.Module { return t },
	}
}

// Provision attempts to call ProvisionerFn. If ProvisionerFn is not set, nil is returned.
func (t *TestModule[T]) Provision(ctx caddy.Context) error {
	if t.ProvisionerFn != nil {
		return t.ProvisionerFn(ctx)
	}
	return nil
}

// Validate attempts to call ValidateFn. If ValidateFn is not set, nil is returned.
func (t *TestModule[T]) Validate() error {
	if t.ValidateFn != nil {
		return t.ValidateFn()
	}
	return nil
}

// Start attempts to call StartFn. If StartFn is not set, nil is returned.
func (t *TestModule[T]) Start(v T) error {
	if t.StartFn != nil {
		return t.StartFn(v)
	}
	return nil
}

// Cleanup attempts to call CleanupFn. If CleanupFn is not set, nil is returned.
func (t *TestModule[T]) Cleanup() error {
	if t.CleanupFn != nil {
		return t.CleanupFn()
	}
	return nil
}

// UnmarshalCaddyfile attempts to call UnmarshalCaddyfileFn. If UnmarshalCaddyfileFn is not set, nil is returned.
func (t *TestModule[T]) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if t.UnmarshalCaddyfileFn != nil {
		return t.UnmarshalCaddyfileFn(d)
	}
	return nil
}
