package stub_listener

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
	"time"
)

func TestStubListener(t *testing.T) {
	t.Run("call", func(t *testing.T) {
		ln, err := StubListener(context.TODO(), "", "test", net.ListenConfig{})
		if err != nil {
			t.Fail()
			return
		}
		defer ln.(io.Closer).Close()
	})

	t.Run("caddy call", func(t *testing.T) {
		addr, err := caddy.ParseNetworkAddress("stub://0.0.0.0")
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		ctx, cancel = caddy.NewContext(caddy.Context{Context: ctx})
		defer cancel()
		ln, err := addr.Listen(ctx, 0, net.ListenConfig{})
		require.NoError(t, err)
		require.NotNil(t, ln)
	})
}

func TestStubAddr(t *testing.T) {
	ln, err := StubListener(context.TODO(), "", "test", net.ListenConfig{})
	if err != nil {
		t.Fail()
		return
	}
	defer ln.(io.Closer).Close()
	addr := ln.(net.Listener).Addr()
	if addr.Network() != "stub" {
		t.Fail()
	}
	if addr.String() != "test" {
		t.Fail()
	}
}

func TestConfig(t *testing.T) {
	adapter := caddyconfig.GetAdapter("caddyfile")
	require.NotNil(t, adapter)
	b, warn, err := adapter.Adapt([]byte(`{
	default_bind stub://1.2.3.4
}

:80 {
}

foo.bar.com {
}
`), nil)
	require.NoError(t, err)
	require.Empty(t, warn)
	require.JSONEq(t, `{
        "apps": {
            "http": {
                "servers": {
                    "srv0": {
                        "listen": ["stub://1.2.3.4:443"],
						"routes": [
							{
								"match": [{"host": ["foo.bar.com"]}],
								"terminal": true
							}
						]
                    },
					"srv1": {
                        "listen": ["stub://1.2.3.4:80"]
					}
                }
            }
        }
    }`, string(b))
}
