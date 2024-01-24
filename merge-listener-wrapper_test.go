package point_c_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	point_c "github.com/point-c/caddy"
	test_caddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/point-c/simplewg"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func TestMergeWrapper_WrapListener(t *testing.T) {
	t.Run("closed before accepted", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		ln := test_caddy.NewTestListenerModule[func(net.Listener)](t)
		ln.Register()
		conn := test_caddy.NewTestConn(t)
		closed := make(chan struct{})
		conn.CloseFn = func() error { defer close(closed); return nil }
		c := make(chan net.Conn, 1)
		c <- conn
		ready := make(chan struct{})
		ln.AcceptFn = func() (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, net.ErrClosed
			case <-ready:
			}
			select {
			case <-ctx.Done():
				return nil, net.ErrClosed
			case c := <-c:
				return c, nil
			}
		}
		v, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON[any](t))
		require.NoError(t, err)
		wrapped := v.(caddy.ListenerWrapper).WrapListener(ln)
		require.NoError(t, wrapped.Close())
		close(ready)
		timeout, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		select {
		case <-timeout.Done():
			t.Fatal("timeout")
		case <-closed:
		}
	})

	t.Run("one listener, accept from wrapped", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, wrapped *test_caddy.TestListenerModule[func(net.Listener)], _ []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return []*test_caddy.TestListenerModule[func(net.Listener)]{wrapped}
		})
	})

	t.Run("one listener, accept from merged", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, _ *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return lns
		})
	})

	t.Run("two listeners, accept from wrapped", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, wrapped *test_caddy.TestListenerModule[func(net.Listener)], _ []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return []*test_caddy.TestListenerModule[func(net.Listener)]{wrapped}
		})
	})

	t.Run("two listeners, accept from one random merged", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, _ *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return []*test_caddy.TestListenerModule[func(net.Listener)]{lns[rand.Intn(len(lns))]}
		})
	})

	t.Run("two listeners, accept from both", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, _ *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return lns
		})
	})

	t.Run("three listeners, accept all", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, wrapped *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			lns = append([]*test_caddy.TestListenerModule[func(net.Listener)]{wrapped}, lns...)
			rand.Shuffle(len(lns), func(i, j int) { lns[i], lns[j] = lns[j], lns[i] })
			return lns
		})
	})

	t.Run("three listeners, accept wrapped and one merged", func(t *testing.T) {
		acceptTest(t, 1, func(t testing.TB, wrapped *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)] {
			return []*test_caddy.TestListenerModule[func(net.Listener)]{wrapped, lns[rand.Intn(len(lns))]}
		})
	})
}

func acceptTest(t testing.TB, n int, acceptor func(t testing.TB, wrapped *test_caddy.TestListenerModule[func(net.Listener)], lns []*test_caddy.TestListenerModule[func(net.Listener)]) []*test_caddy.TestListenerModule[func(net.Listener)]) {
	n = max(n, 1)
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
	defer func() {
		cancel()
	}()

	ln, _ := NewTestListeners(t, n+1) // create an extra one to be wrapped
	conn, errs := make(chan net.Conn), make(chan error)
	accept := acceptor(t, ln[0], ln[1:])
	for _, ln := range ln {
		ln.AcceptFn = func() (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, net.ErrClosed
			}
		}
	}

	for _, ln := range ln[1:] {
		ln := ln
		ln.Register()
		ln.StartFn = func(fn func(net.Listener)) error {
			fn(ln)
			return nil
		}
	}

	var wg simplewg.Wg
	var counter sync.WaitGroup
	for _, a := range accept {
		c := make(chan net.Conn)
		wg.Go(func() {
			defer close(c)
			select {
			case <-ctx.Done():
			case c <- test_caddy.NewTestConn(t):
			}
		})
		counter.Add(1)
		done := sync.OnceFunc(counter.Done)
		a.AcceptFn = func() (net.Conn, error) {
			defer done()
			if c, ok := <-c; ok {
				return c, nil
			}
			return nil, net.ErrClosed
		}
	}
	v, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON(t, ln[1:]...))
	require.NoError(t, err)

	wrapped := v.(caddy.ListenerWrapper).WrapListener(ln[0])
	go func() {
		defer wrapped.Close()
		wg.Wait()
		counter.Wait()
	}()
	go func() {
		for {
			c, err := wrapped.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
				case errs <- err:
				}
				return
			}
			select {
			case <-ctx.Done():
			case conn <- c:
			}
		}
	}()

	for i, _ := range append(accept, nil) {
		select {
		case err := <-errs:
			require.ErrorIs(t, err, net.ErrClosed)
		case c := <-conn:
			require.NotNil(t, c, "i = %d", i)
		case <-ctx.Done():
			t.Fatalf("context cancelled i = %d", i)
		case <-time.After(time.Second * 10):
			t.Fatalf("test timed out i = %d", i)
		}
	}
}

func TestMergeWrapper_Cleanup(t *testing.T) {
	t.Run("listener closed", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		ln := test_caddy.NewTestListenerModule[func(net.Listener)](t)
		ln.Register()
		v, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON(t, ln))
		require.NoError(t, err)
		cancel()
		require.Equal(t, &point_c.MergeWrapper{}, v.(*point_c.MergeWrapper))
	})
}

func TestMergeWrapper_Provision(t *testing.T) {
	t.Run("listeners set to null", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge", []byte(`{"listeners": null}`))
		require.NoError(t, err)
	})

	t.Run("listener failed to load", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		ln := test_caddy.NewTestListenerModule[any](t)
		ln.Register()
		ln.ProvisionerFn = func(caddy.Context) error { return errors.New(ln.ID) }
		_, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON(t, ln))
		require.ErrorContains(t, err, ln.ID)
	})

	t.Run("listener fully provisions one listeners", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		ln1 := test_caddy.NewTestListenerModule[func(net.Listener)](t)
		ln1.Register()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON(t, ln1))
		require.NoError(t, err)
	})

	t.Run("listener fully provisions two listeners", func(t *testing.T) {
		ctx, cancel := caddy.NewContext(caddy.Context{Context: context.TODO()})
		defer cancel()
		lns, reg := NewTestListeners(t, 2)
		reg()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge", generateMergedJSON(t, lns...))
		require.NoError(t, err)
	})
}

func TestMergeWrapper_UnmarshalCaddyfile(t *testing.T) {
	lns, reg := NewTestListeners(t, 2)
	ln1, ln2 := lns[0], lns[1]
	reg()

	tests := []struct {
		name      string
		caddyfile string
		json      string
		wantErr   bool
	}{
		{
			name: "basic",
			caddyfile: `merge {
	` + ln1.ID + `
	` + ln2.ID + `
}`,
			json: string(caddyconfig.JSON(map[string]any{
				"listeners": []any{
					map[string]any{
						"listener": ln1.ID,
					},
					map[string]any{
						"listener": ln2.ID,
					},
				},
			}, nil)),
		},
		{
			name: "submodule does not exist",
			caddyfile: `merge {
	foobar
}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pc point_c.MergeWrapper
			if err := pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tt.caddyfile)); tt.wantErr {
				require.Errorf(t, err, "UnmarshalCaddyfile() wantErr %v", tt.wantErr)
				return
			} else {
				require.NoError(t, err, "UnmarshalCaddyfile()")
			}
			require.JSONEq(t, tt.json, string(caddyconfig.JSON(pc, nil)), "caddyfile != json")
		})
	}
}

func generateMergedJSON[T any](t testing.TB, tln ...*test_caddy.TestListenerModule[T]) []byte {
	raw := make([]json.RawMessage, len(tln))
	for i, ln := range tln {
		raw[i] = caddyconfig.JSONModuleObject(struct{}{}, "listener", ln.ID, nil)
	}
	return caddyconfig.JSON(map[string]any{"listeners": raw}, nil)
}

func NewTestListeners(t testing.TB, n int) ([]*test_caddy.TestListenerModule[func(net.Listener)], func()) {
	ln := make([]*test_caddy.TestListenerModule[func(net.Listener)], n)
	for i := 0; i < n; i++ {
		ln[i] = test_caddy.NewTestListenerModule[func(net.Listener)](t)
	}
	return ln, func() {
		for _, cl := range ln {
			cl.Register()
		}
	}
}
