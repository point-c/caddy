package caddyreg

import (
	"github.com/caddyserver/caddy/v2"
	"strings"
)

// R registers the provided caddy module.
func R[T caddy.Module]() { caddy.RegisterModule(*new(T)) }

// Info helps generate the [caddy.ModuleInfo] by automatically filling in the [caddy.ModuleInfo.New] field.
// The id parameter is set to the ID field of the [caddy.ModuleInfo].
func Info[T any, I interface {
	*T
	caddy.Module
}](id ...string) caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: caddy.ModuleID(strings.Join(id, ".")), New: func() caddy.Module { return (any(new(T))).(caddy.Module) }}
}
