package test_caddy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCaddyContext(t *testing.T) {
	ctx, cancel := NewCaddyContext(t, context.TODO(), caddy.Config{})
	cancel()
	require.NotNil(t, ctx)
	require.NotNil(t, <-ctx.Done())
}

func TestRegister(t *testing.T) {
	module := &TestModule[int]{t: t, Module: caddy.ModuleID(uuid.NewString())}
	module.Register()
}

func TestNewTestModule(t *testing.T) {
	var parent caddy.Module
	NewTestModule[int, caddy.Module](t, &parent, func(p caddy.Module) *TestModule[int] {
		return &TestModule[int]{}
	}, "example.module.")
}

func TestTestModuleMarshalJSON(t *testing.T) {
	module := &TestModule[int]{}

	bytes, err := module.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, []byte("{}"), bytes)

	mockJSON := []byte(`{"key":"value"}`)
	module.MarshalJSONFn = func() ([]byte, error) {
		return mockJSON, nil
	}
	bytes, err = module.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, mockJSON, bytes)
}

func TestTestModuleUnmarshalJSON(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.UnmarshalJSON([]byte(`{"key":"value"}`)))

	module.UnmarshalJSONFn = func(b []byte) error {
		var data map[string]string
		if err := json.Unmarshal(b, &data); err != nil {
			return err
		}
		if data["key"] != "value" {
			return json.Unmarshal(b, &data)
		}
		return nil
	}
	require.NoError(t, module.UnmarshalJSON([]byte(`{"key":"value"}`)))
}

func TestTestModuleProvision(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.Provision(caddy.Context{}))

	module.ProvisionerFn = func(caddy.Context) error {
		return nil
	}
	require.NoError(t, module.Provision(caddy.Context{}))
}

func TestTestModuleValidate(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.Validate())

	module.ValidateFn = func() error {
		return nil
	}
	require.NoError(t, module.Validate())
}

func TestTestModuleStart(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.Start(0))

	module.StartFn = func(v int) error {
		return nil
	}
	require.NoError(t, module.Start(0))
}

func TestTestModuleCleanup(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.Cleanup())

	module.CleanupFn = func() error {
		return nil
	}
	require.NoError(t, module.Cleanup())
}

func TestTestModuleUnmarshalCaddyfile(t *testing.T) {
	module := &TestModule[int]{}

	require.NoError(t, module.UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))

	module.UnmarshalCaddyfileFn = func(d *caddyfile.Dispenser) error {
		return errors.New("test")
	}
	require.Error(t, module.UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))
}
