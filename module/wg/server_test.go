package wg

import (
	"context"
	"encoding/json"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func Test_Server_UnmarshalCaddyfile(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		var c Server
		require.ErrorContains(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgserver foo {
	port 51218
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	ip 1.2..
}`)), "1.2..")
	})

	t.Run("valid", func(t *testing.T) {
		var c Server
		require.NoError(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgserver foo {
	ip 1.2.3.4
	port 51820
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	peer test1 {
		ip 1.1.1.2
		public MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
		shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	}
	peer test2 {
		ip 1.1.1.3
		public MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
		shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	}
}`)))
	})
	t.Run("invalid key", func(t *testing.T) {
		var c Server
		require.ErrorContains(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgserver foo {
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	bar MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
	ip 1.2.3.4
}`)), "bar")
	})
	t.Run("no value for key", func(t *testing.T) {
		var c Server
		require.Error(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgserver foo {
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	public
	ip 1.2.3.4
}`)))
	})
}

func Test_Server_Provision_Cleanup(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	var c Server
	c.wg.Wg = io.NopCloser(nil)
	require.NoError(t, c.Provision(ctx))
	require.NoError(t, c.Cleanup())
}

func Test_Server_Start(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		var c Server
		name := uuid.NewString()
		require.NoError(t, json.Unmarshal(caddyconfig.JSON([]any{
			map[string]any{
				"hostname": name,
			},
			map[string]any{
				"hostname": name,
			},
		}, nil), &c.Peers))
		c.wg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		require.ErrorContains(t, c.Start(func(s string, net pointc.Net) error { return nil }), name)
	})
	t.Run("ok", func(t *testing.T) {
		var c Server
		c.Name.UnmarshalText([]byte(uuid.NewString()))
		c.wg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		require.NoError(t, c.Start(func(s string, net pointc.Net) error { return nil }))
	})
}
