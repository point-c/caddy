package caddyreg

import "github.com/caddyserver/caddy/v2"

// R registers the provided caddy module
func R[T caddy.Module]() { caddy.RegisterModule(*new(T)) }

func Info[T any, I interface {
	*T
	caddy.Module
}](id string) caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: caddy.ModuleID(id), New: func() caddy.Module { return (any(new(T))).(caddy.Module) }}
}
