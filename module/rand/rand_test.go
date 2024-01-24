package caddy_randhandler

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var r Rand
		require.NoError(t, r.UnmarshalCaddyfile(nil))
	})
	t.Run("adapt", func(t *testing.T) {
		adapter := caddyconfig.GetAdapter("caddyfile")
		require.NotNil(t, adapter)
		b := caddyfile.Format([]byte(`:80 {
	route {
		rand
	}
}
`))
		b, warn, err := adapter.Adapt(b, nil)
		require.NoError(t, err)
		require.Empty(t, warn)
		require.Equal(t, string(caddyconfig.JSON(map[string]any{"apps": map[string]any{"http": map[string]any{"servers": map[string]any{"srv0": map[string]any{"listen": []any{":80"}, "routes": []any{map[string]any{"handle": []any{map[string]any{"handler": "subroute", "routes": []any{map[string]any{"handle": []any{map[string]any{"handler": "rand"}}}}}}}}}}}}}, nil)), string(b))
	})
}

func TestNewRand(t *testing.T) {
	require.IsType(t, rand.New(rand.NewSource(1)), NewRand(1))
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	v, err := ctx.LoadModuleByID("http.handlers.rand", []byte("{}"))
	require.NoError(t, err)
	require.IsType(t, new(Rand), v)
}

func TestNewHeaderValues(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		defer func() { nowSrc = time.Now }()
		nowSrc = func(t time.Time) func() time.Time { return func() time.Time { return t } }(time.Now())
		h := NewHeaderValues(http.Header{})
		require.Equal(t, HeaderSeed{}.Default(), h.Seed())
		require.Equal(t, HeaderSize{}.Default(), h.Size())
	})
	t.Run("parsed", func(t *testing.T) {
		h := NewHeaderValues(http.Header{
			HeaderSeed{}.Key(): []string{"123"},
			HeaderSize{}.Key(): []string{"321"},
		})
		require.Equal(t, int64(321), h.Size())
		require.Equal(t, int64(123), h.Seed())
	})
}

func TestHandler(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://1.1.1.1", nil)
		var r Rand
		resp := httptest.NewRecorder()
		defer func(fn func(int64) io.Reader) { NewRand = fn }(NewRand)
		NewRand = func(int64) io.Reader { return strings.NewReader("test") }
		require.NoError(t, r.ServeHTTP(resp, req, nil))
		require.Equal(t, "test", resp.Body.String())
	})
	t.Run("headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://1.1.1.1", nil)
		req.Header = http.Header{
			HeaderSeed{}.Key(): []string{"123"},
			HeaderSize{}.Key(): []string{"321"},
		}
		var r Rand
		resp := httptest.NewRecorder()
		defer func(fn func(int64) io.Reader) { NewRand = fn }(NewRand)
		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			defer pw.Close()
			for {
				if _, err := pw.Write([]byte{'a'}); err != nil {
					return
				}
			}
		}()
		NewRand = func(int64) io.Reader { return pr }
		require.NoError(t, r.ServeHTTP(resp, req, nil))
		require.Equal(t, strings.Repeat("a", 321), resp.Body.String())
	})
	t.Run("rand", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://1.1.1.1", nil)
		req.Header = http.Header{
			HeaderSeed{}.Key(): []string{"123"},
			HeaderSize{}.Key(): []string{"321"},
		}
		resp := httptest.NewRecorder()
		require.NoError(t, new(Rand).ServeHTTP(resp, req, nil))
		b, err := io.ReadAll(io.LimitReader(rand.New(rand.NewSource(123)), 321))
		require.NoError(t, err)
		require.Equal(t, b, resp.Body.Bytes())
	})
}
