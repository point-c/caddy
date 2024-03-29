package sysnet

import (
	"bytes"
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/module/point-c"
	"github.com/point-c/wg/pkg/ipcheck"
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
				"hostname":  "test",
				"dial-addr": "127.0.0.1",
				"local":     "127.0.0.1",
				"type":      "system",
			},
		},
	}, nil))
	require.NoError(t, err)
	pc, ok := a.(point_c.NetLookup)
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
		_, err := CaddyListen[any](nil, &net.TCPAddr{Port: math.MaxUint16 + 1})
		require.ErrorContains(t, err, "invalid start port")
	})
	t.Run("bad listen", func(t *testing.T) {
		ln, err := net.Listen("tcp", "0.0.0.0:0")
		require.NoError(t, err)
		defer ln.Close()
		_, err = CaddyListen[net.Listener](context.Background(), ln.Addr())
		require.Error(t, err)
	})
	t.Run("bad type", func(t *testing.T) {
		_, err := CaddyListen[bool](context.Background(), &net.TCPAddr{IP: net.IPv4zero})
		require.ErrorContains(t, err, "invalid listener type")
	})
	t.Run("ok", func(t *testing.T) {
		ln, err := CaddyListen[net.Listener](context.Background(), &net.TCPAddr{})
		require.NoError(t, err)
		require.NoError(t, ln.Close())
	})
}

func TestSysnet_UnmarshalCaddyfile(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		err := new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser(""))
		require.ErrorContains(t, err, "local address not set")
		require.ErrorContains(t, err, "dial address not set")
		require.ErrorContains(t, err, "hostname not set")
	})
	t.Run("no hostname", func(t *testing.T) {
		err := new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system"))
		require.ErrorContains(t, err, "local address not set")
		require.ErrorContains(t, err, "dial address not set")
		require.ErrorContains(t, err, "hostname not set")
	})
	t.Run("no dial addr", func(t *testing.T) {
		err := new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test"))
		require.ErrorContains(t, err, "local address not set")
		require.ErrorContains(t, err, "dial address not set")
	})
	t.Run("no local addr", func(t *testing.T) {
		err := new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test 1.2.3.4"))
		require.ErrorContains(t, err, "local address not set")
	})
	t.Run("full", func(t *testing.T) {
		var sn Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test 1.2.3.4 4.3.2.1")))
		require.Equal(t, "test", sn.Hostname.Value())
		require.Equal(t, net.IPv4(1, 2, 3, 4), sn.DialAddr.Value())
		require.Equal(t, net.IPv4(4, 3, 2, 1), sn.Local.Value())
	})
	t.Run("resolve dial address", func(t *testing.T) {
		var sn Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test localhost 4.3.2.1")))
		require.Equal(t, "test", sn.Hostname.Value())
		require.True(t, ipcheck.IsLoopback(sn.DialAddr.Value()))
		require.Equal(t, net.IPv4(4, 3, 2, 1), sn.Local.Value())
	})
	t.Run("resolve local address", func(t *testing.T) {
		var sn Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test 1.2.3.4 localhost")))
		require.Equal(t, "test", sn.Hostname.Value())
		require.Equal(t, net.IPv4(1, 2, 3, 4), sn.DialAddr.Value())
		require.True(t, ipcheck.IsLoopback(sn.Local.Value()))
	})
	t.Run("resolve both", func(t *testing.T) {
		var sn Sysnet
		require.NoError(t, sn.UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test localhost localhost")))
		require.Equal(t, "test", sn.Hostname.Value())
		require.True(t, ipcheck.IsLoopback(sn.DialAddr.Value()))
		require.True(t, ipcheck.IsLoopback(sn.Local.Value()))
	})
	t.Run("skip system", func(t *testing.T) {
		require.NoError(t, new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system system test 1.2.3.4 1.2.3.4")))
	})
	t.Run("lookup fail", func(t *testing.T) {
		err := new(Sysnet).UnmarshalCaddyfile(caddyfile.NewTestDispenser("system test foobar localhost"))
		var e *net.DNSError
		require.ErrorAs(t, err, &e)
	})
}
