package configvalues_test

import (
	"fmt"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCaddyTextUnmarshaler_UnmarshalText(t *testing.T) {
	t.Run("unmarshal regular text", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		const text = "foobar"
		require.NoError(t, v.UnmarshalText([]byte(text)))
		require.Exactly(t, text, v.Value())
		b, err := v.MarshalText()
		require.NoError(t, err)
		require.Exactly(t, text, string(b))
	})

	t.Run("unmarshal replaced text", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		k := uuid.New().String()
		const text = "foobar"
		t.Setenv(k, text)
		str := fmt.Sprintf("{env.%s}", k)
		require.NoError(t, v.UnmarshalText([]byte(str)))
		require.Exactly(t, text, v.Value())
		b, err := v.MarshalText()
		require.NoError(t, err)
		require.Exactly(t, str, string(b))
	})
}

func TestCaddyTextUnmarshaler_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal json string", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		const text = "foobar"
		require.NoError(t, v.UnmarshalJSON([]byte(`"`+text+`"`)))
		require.Exactly(t, text, v.Value())
		b, err := v.MarshalJSON()
		require.NoError(t, err)
		require.Exactly(t, `"`+text+`"`, string(b))
	})

	t.Run("unmarshal json number", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[uint8, configvalues.ValueUnsigned[uint8], *configvalues.ValueUnsigned[uint8]]
		const text = "123"
		require.NoError(t, v.UnmarshalJSON([]byte(text)))
		require.Exactly(t, uint8(123), v.Value())
		b, err := v.MarshalJSON()
		require.NoError(t, err)
		require.Exactly(t, text, string(b))
	})
	t.Run("null", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		require.NoError(t, v.UnmarshalJSON([]byte("null")))
		require.Equal(t, "", v.Value())
		b, err := v.MarshalJSON()
		require.NoError(t, err)
		require.JSONEq(t, "null", string(b))
	})
}

func TestCaddyTextUnmarshaler_UnmarshalCaddyfile(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		const text = "foobar"
		require.NoError(t, v.UnmarshalCaddyfile(caddyfile.NewTestDispenser(text)))
		require.Exactly(t, text, v.Value())
		b, err := v.MarshalText()
		require.NoError(t, err)
		require.Exactly(t, text, string(b))
	})
	t.Run("invalid", func(t *testing.T) {
		var v configvalues.CaddyTextUnmarshaler[string, configvalues.ValueString, *configvalues.ValueString]
		require.Error(t, v.UnmarshalCaddyfile(caddyfile.NewTestDispenser("{")))
	})
}
