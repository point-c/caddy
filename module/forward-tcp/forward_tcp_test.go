package forward_tcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	"github.com/point-c/caddy/module/forward"
	pointc "github.com/point-c/caddy/module/point-c"
	channel_listener "github.com/point-c/caddy/pkg/channel-listener"
	"github.com/point-c/caddy/pkg/configvalues"
	testcaddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/point-c/simplewg"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"testing/iotest"
	"time"
)

func TestConnPair_DialTunnel(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		var cancelled bool
		var buf bytes.Buffer
		n := testcaddy.NewTestNet(t)
		n.LocalAddrFn = func() net.IP { return net.IPv4zero }
		n.DialerFn = func(net.IP, uint16) pointc.Dialer {
			d := testcaddy.NewTestDialer(t)
			d.DialFn = func(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
				return nil, errors.New("test")
			}
			return d
		}
		require.False(t, (&ConnPair{
			Cancel: func() { cancelled = true },
			Remote: testcaddy.NewTestConn(t),
			Logger: slog.New(slog.NewTextHandler(&buf, nil)),
		}).DialTunnel(n, 0))
		require.True(t, cancelled)
		require.NotEmpty(t, buf.Bytes())
	})
	t.Run("ok", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		var buf bytes.Buffer
		var cancelled bool
		n := testcaddy.NewTestNet(t)
		closed := make(chan struct{})
		n.LocalAddrFn = func() net.IP { return net.IPv4zero }
		n.DialerFn = func(net.IP, uint16) pointc.Dialer {
			d := testcaddy.NewTestDialer(t)
			d.DialFn = func(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
				c := testcaddy.NewTestConn(t)
				c.CloseFn = func() error { close(closed); return nil }
				return c, nil
			}
			return d
		}
		require.True(t, (&ConnPair{
			Ctx:    ctx,
			Cancel: func() { cancel(); cancelled = true },
			Remote: testcaddy.NewTestConn(t),
			Logger: slog.New(slog.NewTextHandler(&buf, nil)),
		}).DialTunnel(n, 0))
		require.False(t, cancelled)
		require.Empty(t, buf.Bytes())
		cancel()
		_, ok := <-closed
		require.False(t, ok)
	})
}

func TestForwardTCP_Start(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer func() { require.NoError(t, ln.Close()) }()
		port := ln.Addr().(*net.TCPAddr).Port
		var f ForwardTCP
		require.NoError(t, f.Ports.UnmarshalText([]byte(fmt.Sprintf("%d:%d", port, port))))
		require.Error(t, f.Start(&forward.ForwardNetworks{Src: testcaddy.NewTestNet(t), Dst: testcaddy.NewTestNet(t)}))
	})
}

func TestForwardTCP_ProvisionStartCleanup(t *testing.T) {
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer cancel()
	v, err := ctx.LoadModuleByID("point-c.op.forward.tcp", caddyconfig.JSON(map[string]any{"ports": "80:80"}, nil))
	require.NoError(t, err)
	require.NotNil(t, v)
	ftcp, ok := v.(*ForwardTCP)
	require.True(t, ok)

	str1, str2 := "test", "foobar"
	srcDone, dstDone := make(chan struct{}), make(chan struct{})
	srcNet := compareNet(t, str1, str2, srcDone)
	dstNet := compareNet(t, str2, str1, dstDone)

	require.NoError(t, ftcp.Start(&forward.ForwardNetworks{Src: srcNet, Dst: dstNet}))
	select {
	case <-srcDone:
		select {
		case <-dstDone:
		case <-time.After(time.Second * 5):
			t.Fatal("timeout")
		}
	case <-dstDone:
		select {
		case <-srcDone:
		case <-time.After(time.Second * 5):
			t.Fatal("timeout")
		}
	case <-time.After(time.Second * 5):
		t.Fatal("timeout")
	}
}

func compareNet(t *testing.T, readSend, writeExpected string, done chan<- struct{}) *testcaddy.TestNet {
	conn := testcaddy.NewTestConn(t)
	conn.ReadFn = func(b []byte) (int, error) {
		conn.ReadFn = func([]byte) (int, error) { return 0, net.ErrClosed }
		return copy(b, readSend), nil
	}
	conn.WriteFn = func(b []byte) (int, error) {
		defer close(done)
		conn.WriteFn = func([]byte) (int, error) { return 0, net.ErrClosed }
		require.Equal(t, writeExpected, string(b[:len(writeExpected)]))
		return len(writeExpected), nil
	}

	ln := testcaddy.NewTestListener(t)
	ln.AcceptFn = func() (net.Conn, error) {
		ln.AcceptFn = nil
		return conn, nil
	}

	dl := testcaddy.NewTestDialer(t)
	dl.DialFn = func(context.Context, *net.TCPAddr) (net.Conn, error) {
		dl.DialFn = nil
		return conn, nil
	}

	n := testcaddy.NewTestNet(t)
	n.ListenFn = func(*net.TCPAddr) (net.Listener, error) {
		dl.DialFn = nil
		return ln, nil
	}
	n.DialerFn = func(net.IP, uint16) pointc.Dialer {
		n.ListenFn = nil
		return dl
	}
	return n
}

func TestForwardTCP_UnmarshalCaddyfile(t *testing.T) {
	t.Run("invalid port pair", func(t *testing.T) {
		var ftcp ForwardTCP
		require.ErrorContains(t, ftcp.UnmarshalCaddyfile(caddyfile.NewTestDispenser(uuid.New().String())), "not a pair value")
	})
	t.Run("valid", func(t *testing.T) {
		var ftcp ForwardTCP
		require.NoError(t, ftcp.UnmarshalCaddyfile(caddyfile.NewTestDispenser("80:80")))
		require.Equal(t, configvalues.PairValue[uint16]{
			Left: 80, Right: 80,
		}, *ftcp.Ports.Value())
	})
	t.Run("invalid buf size", func(t *testing.T) {
		var ftcp ForwardTCP
		require.Error(t, ftcp.UnmarshalCaddyfile(caddyfile.NewTestDispenser("80:80 a")))
	})
	t.Run("no next", func(t *testing.T) {
		var ftcp ForwardTCP
		require.Error(t, ftcp.UnmarshalCaddyfile(caddyfile.NewTestDispenser("tcp")))

	})
	t.Run("no buf size", func(t *testing.T) {
		b, warn, err := caddyconfig.GetAdapter("caddyfile").Adapt(caddyfile.Format([]byte(`{
	point-c netops {
		forward test:foo {
			tcp 80:80
		}
	}
}`)), nil)
		require.NoError(t, err)
		require.Empty(t, warn)
		require.JSONEq(t, string(caddyconfig.JSON(map[string]any{
			"apps": map[string]any{
				"point-c": map[string]any{
					"net-ops": []any{
						map[string]any{
							"hosts": "test:foo",
							"forwards": []any{
								map[string]any{
									"forward": "tcp",
									"ports":   "80:80",
									"buf":     nil,
								},
							},
							"op": "forward",
						},
					},
				},
			},
		}, nil)), string(b))
	})
	t.Run("full", func(t *testing.T) {
		b, warn, err := caddyconfig.GetAdapter("caddyfile").Adapt(caddyfile.Format([]byte(`{
	point-c netops {
		forward test:foo {
			tcp 80:80 1234
		}
	}
}`)), nil)
		require.NoError(t, err)
		require.Empty(t, warn)
		require.JSONEq(t, string(caddyconfig.JSON(map[string]any{
			"apps": map[string]any{
				"point-c": map[string]any{
					"net-ops": []any{
						map[string]any{
							"hosts": "test:foo",
							"forwards": []any{
								map[string]any{
									"forward": "tcp",
									"ports":   "80:80",
									"buf":     1234,
								},
							},
							"op": "forward",
						},
					},
				},
			},
		}, nil)), string(b))
	})
}

func TestTcpCopy(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var buf bytes.Buffer
		TcpCopy(func() {}, slog.New(slog.NewTextHandler(&buf, nil)), io.Discard, io.LimitReader(nil, 0), []byte{0})
		require.Empty(t, buf.Bytes())
	})
	t.Run("not ok", func(t *testing.T) {
		var buf bytes.Buffer
		TcpCopy(func() {}, slog.New(slog.NewTextHandler(&buf, nil)), io.Discard, iotest.ErrReader(errors.New("test")), []byte{0})
		require.NotEmpty(t, buf.Bytes())
	})
}

func TestListenLoop(t *testing.T) {
	in := make(chan net.Conn)
	out := make(chan net.Conn)
	var conn testcaddy.TestConn

	cln := channel_listener.New(in, &net.TCPAddr{})
	go func() {
		defer cln.Close()
		in <- &conn
	}()

	var wg simplewg.Wg
	wg.Go(func() { ListenLoop(cln, out) })

	v, ok := <-out
	require.True(t, ok)
	require.Equal(t, &conn, v)
	_, ok = <-out
	require.False(t, ok)
	wg.Wait()
}

func TestPrepareConnPairLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	in := make(chan net.Conn)
	out := make(chan *ConnPair)

	var wg simplewg.Wg
	wg.Go(func() {
		PrepareConnPairLoop(ctx, slog.Default(), in, out)
	})

	in <- testcaddy.NewTestConn(t)
	v, ok := <-out
	require.True(t, ok)
	require.Equal(t, slog.Default(), v.Logger)
	require.NotNil(t, v.Ctx)
	require.NotNil(t, v.Cancel)
	v.Cancel()
	cancel()
	in <- testcaddy.NewTestConn(t)
	close(in)
	wg.Wait()
}

func TestStartCopyLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	in := make(chan *ConnPair)
	out := make(chan bool)

	var wg simplewg.Wg
	defer wg.Wait()
	wg.Go(func() {
		defer close(out)
		StartCopyLoop(in, func(done func(), _ *slog.Logger, _ io.Writer, _ io.Reader) { done(); out <- true })
	})

	newConnPair := func() *ConnPair {
		var cp ConnPair
		cp.Ctx, cp.Cancel = context.WithCancel(ctx)
		cp.Logger = slog.Default()
		cp.Remote = new(testcaddy.TestConn)
		cp.Tunnel = new(testcaddy.TestConn)
		return &cp
	}

	cp := newConnPair()
	in <- cp
	v, ok := <-out
	require.True(t, v)
	require.True(t, ok)
	v, ok = <-out
	require.True(t, v)
	require.True(t, ok)
	cp = newConnPair()
	cp.Cancel()
	cp.Cancel = func() { out <- true }
	in <- cp
	v, ok = <-out
	require.True(t, v)
	require.True(t, ok)
	close(in)
	v, ok = <-out
	require.False(t, v)
	require.False(t, ok)
}

func TestDialRemoteLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil))

	in := make(chan *ConnPair)
	out := make(chan *ConnPair)

	var wg simplewg.Wg
	dialerFn := make(chan pointc.Dialer)
	n := testcaddy.NewTestNet(t)
	n.DialerFn = func(net.IP, uint16) pointc.Dialer { return <-dialerFn }
	n.LocalAddrFn = func() net.IP { return net.IPv4zero }

	closeInOnce := sync.OnceFunc(func() { close(in) })
	wg.Go(func() {
		defer closeInOnce()
		DialRemoteLoop(n, 0, in, out)
	})
	go func() {
		defer close(out)
		wg.Wait()
	}()

	{
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		var cancelled bool
		done := make(chan struct{})
		in <- &ConnPair{
			Ctx:    ctx,
			Cancel: func() { defer close(done); cancelled = true },
			Remote: testcaddy.NewTestConn(t),
		}
		select {
		case <-time.After(time.Second * 5):
			t.Fatal("timeout")
		case <-done:
			require.True(t, cancelled)
		}
	}

	{
		buf.Reset()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			d := testcaddy.NewTestDialer(t)
			d.DialFn = func(context.Context, *net.TCPAddr) (net.Conn, error) { return nil, errors.New("test") }
			dialerFn <- d
		}()
		var cancelled bool
		in <- &ConnPair{
			Ctx:    ctx,
			Cancel: func() { cancel(); cancelled = true },
			Remote: testcaddy.NewTestConn(t),
			Logger: l,
		}
		require.NotNil(t, <-ctx.Done())
		require.True(t, cancelled)
		require.NotEmpty(t, buf.Bytes())
	}

	{
		buf.Reset()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		var cancelled bool
		go func() {
			d := testcaddy.NewTestDialer(t)
			d.DialFn = func(context.Context, *net.TCPAddr) (net.Conn, error) { return testcaddy.NewTestConn(t), nil }
			dialerFn <- d
		}()
		pc := &ConnPair{
			Ctx:    ctx,
			Cancel: func() { cancel(); cancelled = true },
			Remote: testcaddy.NewTestConn(t),
			Logger: l,
		}

		in <- pc
		select {
		case <-ctx.Done():
			t.FailNow()
		case v, ok := <-out:
			require.True(t, ok)
			require.Equal(t, pc, v)
		}
		require.False(t, cancelled)
		require.Empty(t, buf.Bytes())
	}

	{
		closeInOnce()
		v, ok := <-out
		require.Nil(t, v)
		require.False(t, ok)
	}
}
