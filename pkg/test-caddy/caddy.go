package test_caddy

import (
	"context"
	"github.com/caddyserver/caddy/v2"
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
