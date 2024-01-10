package point_c_test

import (
	"context"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/point-c/caddy"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
)

func TestStubListener(t *testing.T) {
	ln, err := point_c.StubListener(context.TODO(), "", "test", net.ListenConfig{})
	if err != nil {
		t.Fail()
		return
	}
	defer ln.(io.Closer).Close()
}

func TestStubAddr(t *testing.T) {
	ln, err := point_c.StubListener(context.TODO(), "", "test", net.ListenConfig{})
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
