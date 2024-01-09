package configvalues_test

import (
	"errors"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/stretchr/testify/require"
	"testing"
)

type testCm[T caddyfile.Unmarshaler] []T

func (t *testCm[T]) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var v T
	if err := v.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	*t = append(*t, v)
	return nil
}

type testECm struct{}

func (t *testECm) UnmarshalCaddyfile(*caddyfile.Dispenser) error { return nil }

func (t *testECm) UnmarshalJSON([]byte) error { return errors.New("") }

type testM struct{}

func (testM) UnmarshalCaddyfile(*caddyfile.Dispenser) error { return nil }

type testEM struct{}

func (testEM) UnmarshalCaddyfile(*caddyfile.Dispenser) error { return errors.New("") }

func TestUnmarshaler(t *testing.T) {
	const name = "test"
	fn := configvalues.CaddyfileUnmarshaler[testCm[testM], *testCm[testM]](name)

	t.Run("unmarshal one", func(t *testing.T) {
		_, err := fn(caddyfile.NewTestDispenser(``), nil)
		require.NoError(t, err)
	})

	t.Run("unmarshal two", func(t *testing.T) {
		a, err := fn(caddyfile.NewTestDispenser(`foo`), nil)
		require.NoError(t, err)
		a, err = fn(caddyfile.NewTestDispenser(`foo`), a)
		require.NoError(t, err)
		v, ok := a.(httpcaddyfile.App)
		require.Exactly(t, ok, true, "expected app got %T", v)
		require.Exactly(t, v.Name, name)
		require.JSONEq(t, string(v.Value), "[{}, {}]")
	})

	t.Run("incorrect resume", func(t *testing.T) {
		v, err := fn(nil, httpcaddyfile.App{})
		require.Error(t, err)
		require.Exactly(t, v, nil)
	})

	t.Run("incorrect name", func(t *testing.T) {
		v, err := fn(nil, &httpcaddyfile.App{Name: "foo"})
		require.Error(t, err)
		require.Exactly(t, v, nil)
	})

	t.Run("bad caddy unmarshalling", func(t *testing.T) {
		fn := configvalues.CaddyfileUnmarshaler[testCm[testEM], *testCm[testEM]]("")
		v, err := fn(nil, nil)
		require.Error(t, err)
		require.Exactly(t, v, nil)
	})

	t.Run("bad json unmarshalling", func(t *testing.T) {
		fn := configvalues.CaddyfileUnmarshaler[testECm, *testECm]("")
		v, err := fn(nil, httpcaddyfile.App{})
		require.Error(t, err)
		require.Exactly(t, v, nil)
	})
}
