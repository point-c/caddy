// Package lifecycler helps Caddy modules manage the life cycle of submodules.
package lifecycler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"reflect"
	"strings"
)

var (
	_ json.Unmarshaler   = (*LifeCycler[any])(nil)
	_ json.Marshaler     = (*LifeCycler[any])(nil)
	_ caddy.CleanerUpper = (*LifeCycler[any])(nil)
)

type (
	// LifeCycler assists with loading submodules and maintaining their lifecycle.
	LifeCycler[T any] struct {
		V       T                 // Value passed to submodules
		Modules []LifeCyclable[T] // Unstarted submodules
		Started []LifeCyclable[T] // Started submodules
	}
	// ProvisionInfo is used to help with caddy provisioning.
	ProvisionInfo struct {
		StructPointer any                // Pointer to the struct to load config into
		FieldName     string             // Name of the field that config will be loaded into
		Raw           *[]json.RawMessage // Pointer to the array that config will be loaded into
	}
	// CaddyfileInfo is used to help unmarshal a caddyfile.
	CaddyfileInfo struct {
		ModuleID           []string              // module.id.path. with dot at end if required
		Raw                *[]json.RawMessage    // Where the raw config of submodules will be appended to
		SubModuleSpecifier string                // The specifier used by caddy to determine the submodule.
		ParseVerbLine      caddyfile.Unmarshaler // An optional [caddyfile.Unmarshaler]. If specified will be used to decode the second item on the first line of the config.
	}
	// LifeCyclable is a submodule that can be started.
	LifeCyclable[T any] interface {
		caddy.Module
		Start(T) error
	}
)

// UnmarshalJSON returns an error preventing [LifeCycler] from being used in configuration.
func (*LifeCycler[T]) UnmarshalJSON([]byte) error { return errors.New("do not unmarshal") }

// MarshalJSON returns an error preventing [LifeCycler] from being used in configuration.
func (*LifeCycler[T]) MarshalJSON() ([]byte, error) { return nil, errors.New("do not marshal") }

// Provision loads submodules from the config. `info` must not be nil and the `Raw` field must be set.
func (l *LifeCycler[T]) Provision(ctx caddy.Context, info *ProvisionInfo) error {
	if info == nil || info.Raw == nil {
		return errors.New("no info")
	}

	if *info.Raw != nil {
		// Load submodules
		val, err := ctx.LoadModule(info.StructPointer, info.FieldName)
		if err != nil {
			return fmt.Errorf("failed to provision field %q: %w", info.FieldName, err)
		}
		*info.Raw = nil

		raw := val.([]any)
		l.Modules = make([]LifeCyclable[T], len(raw))
		l.Started = make([]LifeCyclable[T], 0, len(raw))
		for i, v := range raw {
			// Assert submodule to correct type
			vv, ok := v.(LifeCyclable[T])
			if !ok {
				return fmt.Errorf("expected type LifeCyclable[%s], got type %T", func() (s string) {
					defer func() {
						// Swallow reflect panics
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

// SetValue sets the internal value that will be passed to submodules.
func (l *LifeCycler[T]) SetValue(v T) { l.V = v }

// Start starts submodules and passes the stored value into them.
func (l *LifeCycler[T]) Start() error {
	// Loop through each module. Use standard iterator to capture module number.
	for i := 0; len(l.Modules) > 0; i++ {
		// Start the module in a separate function to capture panics.
		if err := func(op LifeCyclable[T]) (err error) {
			defer func() {
				// Capture revcover
				if r := recover(); r != nil {
					err = errors.Join(err, fmt.Errorf("recovered panic starting %[2]s[%[1]d]: %[3]v", i, op.CaddyModule().ID, r))
				}
				// Add information to errors
				if err != nil {
					err = errors.Join(err, fmt.Errorf("failed to start %[2]s[%[1]d]", i, op.CaddyModule().ID))
				}
			}()
			// Try to start
			return op.Start(l.V)
		}(l.Modules[0]); err != nil {
			return err
		}
		// Module started, keep track of it.
		l.Started = append(l.Started, l.Modules[0])
		l.Modules = l.Modules[1:]
	}
	// All modules are started
	l.Modules = nil
	// Don't need this value anymore
	l.V = *new(T)
	return nil
}

// Cleanup clears the lifecycler struct to help the GC.
// Any caddy submodules will be cleaned up by caddy.
func (l *LifeCycler[T]) Cleanup() (err error) {
	*l = LifeCycler[T]{}
	return
}

// UnmarshalCaddyfile unmarshals a module from a caddyfile.
// `info` must not be nil and must have the `raw` and `SubModuleSpecifier` fields set.
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
			v, err := caddyfile.UnmarshalModule(d, strings.Join(info.ModuleID, ".")+"."+modName)
			if err != nil {
				return err
			}
			*info.Raw = append(*info.Raw, caddyconfig.JSONModuleObject(v, info.SubModuleSpecifier, modName, nil))
		}
	}
	return nil
}
