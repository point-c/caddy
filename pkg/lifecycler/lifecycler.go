package lifecycler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"reflect"
)

var (
	_ json.Unmarshaler   = (*LifeCycler[any])(nil)
	_ json.Marshaler     = (*LifeCycler[any])(nil)
	_ caddy.CleanerUpper = (*LifeCycler[any])(nil)
)

type (
	LifeCycler[T any] struct {
		V       T
		Modules []LifeCyclable[T]
		Started []LifeCyclable[T]
	}
	ProvisionInfo struct {
		StructPointer any
		FieldName     string
		Raw           *[]json.RawMessage
	}
	CaddyfileInfo struct {
		ModuleID           string // module.id.path. with dot at end if required
		Raw                *[]json.RawMessage
		SubModuleSpecifier string
		ParseVerbLine      caddyfile.Unmarshaler
	}
	LifeCyclable[T any] interface {
		caddy.Module
		Start(T) error
	}
)

func (*LifeCycler[T]) UnmarshalJSON([]byte) error   { return errors.New("do not unmarshal") }
func (*LifeCycler[T]) MarshalJSON() ([]byte, error) { return nil, errors.New("do not marshal") }

func (l *LifeCycler[T]) Provision(ctx caddy.Context, info *ProvisionInfo) error {
	if info == nil || info.Raw == nil {
		return errors.New("no info")
	}

	if *info.Raw != nil {
		val, err := ctx.LoadModule(info.StructPointer, info.FieldName)
		if err != nil {
			return fmt.Errorf("failed to provision field %q: %w", info.FieldName, err)
		}
		*info.Raw = nil

		raw := val.([]any)
		l.Modules = make([]LifeCyclable[T], len(raw))
		l.Started = make([]LifeCyclable[T], 0, len(raw))
		for i, v := range raw {
			vv, ok := v.(LifeCyclable[T])
			if !ok {
				return fmt.Errorf("expected type LifeCyclable[%s], got type %T", func() (s string) {
					defer func() {
						recover()
						if s == "" {
							s = "<unknown type>"
						}
					}()
					return reflect.TypeOf(new(T)).Elem().Name()
				}(), v)
			}
			l.Modules[i] = vv
		}
	}
	return nil
}

func (l *LifeCycler[T]) SetValue(v T) { l.V = v }

func (l *LifeCycler[T]) Start() error {
	for i := 0; len(l.Modules) > 0; i++ {
		if err := func(op LifeCyclable[T]) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = errors.Join(err, fmt.Errorf("recovered panic starting %[2]s[%[1]d]: %[3]v", i, op.CaddyModule().ID, r))
				}
				if err != nil {
					err = fmt.Errorf("failed to start %[2]s[%[1]d]: %[3]w", i, op.CaddyModule().ID, err)
				}
			}()
			return op.Start(l.V)
		}(l.Modules[0]); err != nil {
			return err
		}
		l.Started = append(l.Started, l.Modules[0])
		l.Modules = l.Modules[1:]
	}
	l.Modules = nil
	l.V = *new(T)
	return nil
}

func (l *LifeCycler[T]) Cleanup() (err error) {
	*l = LifeCycler[T]{}
	return
}

// UnmarshalCaddyfile unmarshals a module from a caddyfile.
//
//	{
//	  <verb> [rest of line] {
//	    <submodule name> <submodule config>
//	  }
//	}
func (l *LifeCycler[T]) UnmarshalCaddyfile(d *caddyfile.Dispenser, info *CaddyfileInfo) error {
	if info == nil || info.Raw == nil || info.SubModuleSpecifier == "" {
		return errors.New("not enough information to unmarshal caddyfile")
	}

	for d.Next() {
		if info.ParseVerbLine != nil {
			if !d.NextArg() {
				return d.ArgErr()
			}
			if err := info.ParseVerbLine.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			modName := d.Val()
			v, err := caddyfile.UnmarshalModule(d, info.ModuleID+modName)
			if err != nil {
				return err
			}
			var warn []caddyconfig.Warning
			*info.Raw = append(*info.Raw, caddyconfig.JSONModuleObject(v, info.SubModuleSpecifier, modName, &warn))
			if len(warn) != 0 {
				return fmt.Errorf("%s", warn)
			}
		}
	}
	return nil
}
