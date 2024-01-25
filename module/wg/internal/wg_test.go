package internal

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/wg"
	"github.com/point-c/wgapi"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/conn/bindtest"
	"io"
	"log/slog"
	"testing"
)

func Test_Wg_Provision(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	var w Wg
	require.NoError(t, w.Provision(ctx))
	require.NotNil(t, w.Logger)
}

func Test_Wg_Cleanup(t *testing.T) {
	w := Wg{Wg: io.NopCloser(nil)}
	require.NoError(t, w.Cleanup())
}

func Test_Wg_Start(t *testing.T) {
	defer func(fn func() conn.Bind) { wg.DefaultBind = fn }(wg.DefaultBind)
	binds := bindtest.NewChannelBinds()
	wg.DefaultBind = func() wg.Bind { return binds[0] }

	t.Run("fail on fn", func(t *testing.T) {
		w := Wg{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
		err := errors.New("test")
		require.ErrorIs(t, w.Start(func(string, pointc.Net) error { return err }, map[string]*Net{"": nil}, nil), err)
	})
	t.Run("fail to init", func(t *testing.T) {
		w := Wg{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
		require.ErrorContains(t, w.Start(func(string, pointc.Net) error { return nil }, map[string]*Net{}, wgapi.IPC{wgapi.ErrnoIO}), "invalid UAPI device key: errno")
	})
	t.Run("valid", func(t *testing.T) {
		w := Wg{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
		defer func() { require.NoError(t, w.Cleanup()) }()
		require.NoError(t, w.Start(func(string, pointc.Net) error { return nil }, map[string]*Net{"": new(Net)}, nil))
	})
}

func TestWg_UnmarshalCaddyfile(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		var w Wg
		var name, test configvalues.String
		require.NoError(t, w.UnmarshalCaddyfile(caddyfile.NewTestDispenser("foo bar {\n\ttest value\n}\n"), name.UnmarshalCaddyfile, CaddyfileKeyValues{
			"test": test.UnmarshalCaddyfile,
		}))
		require.Equal(t, "bar", name.Value())
		require.Equal(t, "value", test.Value())
	})
	t.Run("no next arg", func(t *testing.T) {
		var w Wg
		var name, test configvalues.String
		require.Error(t, w.UnmarshalCaddyfile(caddyfile.NewTestDispenser("foo bar {\n\ttest \n}\n"), name.UnmarshalCaddyfile, CaddyfileKeyValues{
			"test": test.UnmarshalCaddyfile,
		}))
	})
	t.Run("return error in loop", func(t *testing.T) {
		var w Wg
		errExp := errors.New("test")
		require.ErrorIs(t, w.UnmarshalCaddyfile(caddyfile.NewTestDispenser("test foo {\nfoo bar\n}\n"), func(*caddyfile.Dispenser) error {
			return errExp
		}, nil), errExp)
	})
	t.Run("unrecognized option", func(t *testing.T) {
		var w Wg
		var name configvalues.String
		require.ErrorContains(t, w.UnmarshalCaddyfile(caddyfile.NewTestDispenser("foo bar {\n\ttest value\n}\n"), name.UnmarshalCaddyfile, CaddyfileKeyValues{}), "unrecognized")
	})
}

func Test_UnmarshalCaddyfileNesting(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		var name, test configvalues.String
		var a []*int
		d := caddyfile.NewTestDispenser("foo bar {\n\ttest value\n}\n")
		require.True(t, d.NextArg())
		require.True(t, d.NextArg())
		require.NoError(t, UnmarshalCaddyfileNesting[int](&a, func(i *int) (UnmarshalCaddyfileFn, CaddyfileKeyValues) {
			*i = 1
			return name.UnmarshalCaddyfile, CaddyfileKeyValues{
				"test": test.UnmarshalCaddyfile,
			}
		})(d))
		require.Len(t, a, 1)
		require.NotNil(t, a[0])
		require.Equal(t, 1, *a[0])
		require.Equal(t, "bar", name.Value())
		require.Equal(t, "value", test.Value())
	})
	t.Run("no next arg", func(t *testing.T) {
		var name, test configvalues.String
		var a []*int
		d := caddyfile.NewTestDispenser("foo bar {\n\ttest \n}\n")
		require.True(t, d.NextArg())
		require.True(t, d.NextArg())
		require.Error(t, UnmarshalCaddyfileNesting[int](&a, func(i *int) (UnmarshalCaddyfileFn, CaddyfileKeyValues) {
			return name.UnmarshalCaddyfile, CaddyfileKeyValues{
				"test": test.UnmarshalCaddyfile,
			}
		})(d))
		require.Empty(t, a)
	})
	t.Run("return error in loop", func(t *testing.T) {
		var a []*int
		errExp := errors.New("test")
		d := caddyfile.NewTestDispenser("foo bar {\n\ttest value\n}\n")
		require.True(t, d.NextArg())
		require.True(t, d.NextArg())
		require.ErrorIs(t, UnmarshalCaddyfileNesting[int](&a, func(i *int) (UnmarshalCaddyfileFn, CaddyfileKeyValues) {
			return func(*caddyfile.Dispenser) error {
				return errExp
			}, nil
		})(d), errExp)
		require.Empty(t, a)
	})
	t.Run("unrecognized option", func(t *testing.T) {
		var a []*int
		var name configvalues.String
		d := caddyfile.NewTestDispenser("foo bar {\n\ttest value\n}\n")
		require.True(t, d.NextArg())
		require.True(t, d.NextArg())
		require.ErrorContains(t, UnmarshalCaddyfileNesting[int](&a, func(*int) (UnmarshalCaddyfileFn, CaddyfileKeyValues) {
			return name.UnmarshalCaddyfile, nil
		})(d), "unrecognized")
		require.Empty(t, a)
	})
}
