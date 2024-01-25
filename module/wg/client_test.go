package caddy_wg

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/module/wg/internal"
	"github.com/point-c/wg"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func Test_Client_Start(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := Client{}
		c.wg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		require.NoError(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgclient foo {
	endpoint 1.1.1.1:1
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	public MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
	shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	ip 1.2.3.4
}`)))
		defer func(fn func() wg.Bind) { wg.DefaultBind = fn }(wg.DefaultBind)
		wg.DefaultBind = func() wg.Bind { return internal.TestBind{} }
		require.NoError(t, c.Start(func(string, pointc.Net) error { return nil }))
	})
	t.Run("error", func(t *testing.T) {
		c := Client{}
		c.wg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		require.Error(t, c.Start(func(string, pointc.Net) error { return errors.New("") }))
	})
}

func Test_Client_UnmarshalCaddyfile(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		var c Client
		require.ErrorContains(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgclient foo {
	endpoint xyz.com:51280
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	public MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
	ip 1.2..
	shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
}`)), "1.2..")
	})

	t.Run("valid", func(t *testing.T) {
		var c Client
		require.NoError(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgclient foo {
	endpoint xyz.com:51280
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	public MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
	shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	ip 1.2.3.4
}`)))
	})
	t.Run("invalid key", func(t *testing.T) {
		var c Client
		require.ErrorContains(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgclient foo {
	endpoint xyz.com:51280
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	bar MjsB7gz/BRd5RpYv2NjiquHgloPUpLuQKNQD2epZpGE=
	shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	ip 1.2.3.4
}`)), "bar")
	})
	t.Run("no value for key", func(t *testing.T) {
		var c Client
		require.Error(t, c.UnmarshalCaddyfile(caddyfile.NewTestDispenser(`wgclient foo {
	endpoint xyz.com:51280
	private cITqTRxZEM/w1GkCKEin6yiw/D6Co67rE0jL/nRZtEU=
	public
	shared z2wt6ONPPSDyShUfCHk4dtZrbsmU/cRX21Lls12wfK0=
	ip 1.2.3.4
}`)))
	})
}

func Test_Client_Provision_Cleanup(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	var c Client
	c.wg.Wg = io.NopCloser(nil)
	require.NoError(t, c.Provision(ctx))
	require.NoError(t, c.Cleanup())
}
