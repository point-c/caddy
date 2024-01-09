package configvalues

import (
	"encoding/json"
	"fmt"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

type caddyfileUnmarshaler[T any] interface {
	*T
	caddyfile.Unmarshaler
}

// CaddyfileUnmarshaler returns a function that unmarshals Caddyfile configuration into a specific type.
// It works with types T whose pointer implements the [caddyfile.Unmarshaler] interface.
// The function handles both the initialization of a new configuration object and the resumption
// of a partially-unmarshaled configuration. This is useful if the [caddyfile.App] manages a list of submodules
// and whose [caddyfile.Unmarshaler] will mainly just append to that list.
func CaddyfileUnmarshaler[T any, TP caddyfileUnmarshaler[T]](name string) func(d *caddyfile.Dispenser, resume any) (any, error) {
	return func(d *caddyfile.Dispenser, resume any) (any, error) {
		var v T
		// If there is an existing configuration to resume from, try to unmarshal it.
		if resume != nil {
			j, ok := resume.(httpcaddyfile.App)
			if !ok {
				// Type assertion failed
				return nil, fmt.Errorf("not a %T", j)
			} else if j.Name != name {
				// Name mismatch, somehow resuming mismatched configs
				return nil, fmt.Errorf("expected app with name %q, got %q", name, j.Name)
			}

			// Unmarshal the JSON data into the configuration object.
			if err := json.Unmarshal(j.Value, &v); err != nil {
				return nil, err
			}
		}

		// Unmarshal the Caddyfile data into the configuration object.
		// Generic parameters ensure that the pointer of T implements [caddyfile.CaddyfileUnmarshaler].
		if err := any(&v).(caddyfile.Unmarshaler).UnmarshalCaddyfile(d); err != nil {
			return nil, err
		}

		// Return the configuration wrapped in a [httpcaddyfile.App].
		return httpcaddyfile.App{Name: name, Value: caddyconfig.JSON(&v, nil)}, nil
	}
}
