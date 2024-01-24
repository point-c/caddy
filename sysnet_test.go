package point_c_test

import (
	"bytes"
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	pointc "github.com/point-c/caddy"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSysnet(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()
	a, err := ctx.LoadModuleByID("point-c", caddyconfig.JSON(map[string]any{
		"networks": []any{
			map[string]any{
				"hostname": "test",
				"addr":     "127.0.0.1",
				"type":     "system",
			},
		},
	}, nil))
	require.NoError(t, err)
	pc, ok := a.(pointc.NetLookup)
	require.True(t, ok)
	sys, ok := pc.Lookup("test")
	require.True(t, ok)
	dial := sys.Dialer(sys.LocalAddr(), 0)

	t.Run("local addr", func(t *testing.T) {
		require.Equal(t, net.IPv4(127, 0, 0, 1), sys.LocalAddr())
	})

	const testMsg = "foobar"
	t.Run("tcp", func(t *testing.T) {
		ln, err := sys.Listen(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
		require.NoError(t, err)
		defer ln.Close()
		go func() {
			cn, err := ln.Accept()
			require.NoError(t, err)
			defer cn.Close()
			n, err := io.Copy(cn, strings.NewReader(testMsg))
			require.NoError(t, err)
			require.Equal(t, int64(len(testMsg)), n)
		}()
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		cn, err := dial.Dial(ctx, ln.Addr().(*net.TCPAddr))
		require.NoError(t, err)
		defer cn.Close()
		var buf bytes.Buffer
		n, err := io.Copy(&buf, cn)
		require.NoError(t, err)
		require.Equal(t, int64(len(testMsg)), n)
		require.Equal(t, testMsg, buf.String())
	})

	t.Run("udp", func(t *testing.T) {
		cn1, err := sys.ListenPacket(&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		require.NoError(t, err)
		defer cn1.Close()
		cn2, err := dial.DialPacket(cn1.LocalAddr().(*net.UDPAddr))
		require.NoError(t, err)
		defer cn2.Close()
		go func() {
			n, err := cn1.WriteTo([]byte(testMsg), cn2.LocalAddr())
			require.NoError(t, err)
			require.Equal(t, len(testMsg), n)
		}()
		var buf [len(testMsg)]byte
		n, addr, err := cn2.ReadFrom(buf[:])
		require.NoError(t, err)
		require.Equal(t, addr.String(), cn1.LocalAddr().String())
		require.Equal(t, addr.Network(), cn1.LocalAddr().Network())
		require.Equal(t, len(testMsg), n)
		require.Equal(t, testMsg, string(buf[:]))
	})

	t.Run("bad udp dial", func(t *testing.T) {
		_, err := dial.DialPacket(nil)
		require.Error(t, err)
	})
}

func TestCaddyListen(t *testing.T) {
	t.Run("bad addr", func(t *testing.T) {
		_, err := pointc.CaddyListen[any](nil, &net.TCPAddr{Port: math.MaxUint16 + 1})
		require.ErrorContains(t, err, "invalid start port")
	})
	t.Run("bad listen", func(t *testing.T) {
		ln, err := net.Listen("tcp", "0.0.0.0:0")
		require.NoError(t, err)
		defer ln.Close()
		_, err = pointc.CaddyListen[net.Listener](context.Background(), ln.Addr())
		require.ErrorContains(t, err, "address already in use")
	})
	t.Run("bad type", func(t *testing.T) {
		_, err := pointc.CaddyListen[bool](context.Background(), &net.TCPAddr{IP: net.IPv4zero})
		require.ErrorContains(t, err, "invalid listener type")
	})
	t.Run("ok", func(t *testing.T) {
		ln, err := pointc.CaddyListen[net.Listener](context.Background(), &net.TCPAddr{})
		require.NoError(t, err)
		require.NoError(t, ln.Close())
	})
}

func TestSysnet_UnmarshalCaddyfile(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		require.NoError(t, new(pointc.Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("")))
	})
	t.Run("no hostname", func(t *testing.T) {
		require.Error(t, new(pointc.Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system")))
	})
	t.Run("no addr", func(t *testing.T) {
		var sn pointc.Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test")))
		require.Equal(t, "test", sn.Hostname.Value())
	})
	t.Run("full", func(t *testing.T) {
		var sn pointc.Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test 1.2.3.4")))
		require.Equal(t, "test", sn.Hostname.Value())
		require.Equal(t, net.IPv4(1, 2, 3, 4), sn.Addr.Value())
	})
}
