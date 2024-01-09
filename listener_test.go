package point_c_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/google/uuid"
	pointc "github.com/point-c/caddy"
	"github.com/stretchr/testify/require"
	"net"
	"reflect"
	"testing"
	"time"
	"unsafe"
	_ "unsafe"
)

func TestListener_UnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name      string
		caddyfile string
		json      string
		wantErr   bool
	}{
		{
			name:      "basic",
			caddyfile: "point-c remote 80",
			json:      `{"name": "remote", "port": 80}`,
		},
		{
			name:      "no port",
			caddyfile: "point-c remote",
			wantErr:   true,
		},
		{
			name:      "no network name",
			caddyfile: "point-c",
			wantErr:   true,
		},
		{
			name:      "non numeric port",
			caddyfile: "point-c remote aa",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pc pointc.Listener
			if err := pc.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tt.caddyfile)); tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				b, err := json.Marshal(pc)
				require.NoError(t, err)
				require.JSONEq(t, tt.json, string(b))
			}
		})
	}
}

func TestListener_Provision(t *testing.T) {
	t.Run("point-c not configured correctly", func(t *testing.T) {
		testNet := NewTestNet(t)
		testNet.UnmarshalErr = errors.New("test")
		ctx, cancel := NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{"networks": [{"type": "test-` + testNet.Id() + `"}]}`),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.listeners.point-c", json.RawMessage("{}"))
		require.Error(t, err)
	})

	t.Run("network not found", func(t *testing.T) {
		ctx, cancel := NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{}`),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.listeners.point-c", json.RawMessage("{}"))
		require.Error(t, err)
	})

	t.Run("network errors on listen", func(t *testing.T) {
		tn := NewTestNet(t)
		errExp := errors.New("test err " + uuid.New().String())
		tn.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test", &TestNet{
				ListenFn:    func(*net.TCPAddr) (net.Listener, error) { return nil, errExp },
				LocalAddrFn: func() net.IP { return nil },
			})
		}

		ctx, cancel := NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": caddyconfig.JSON(map[string]any{
					"networks": []any{
						caddyconfig.JSONModuleObject(struct{}{}, "type", "test-"+tn.Id(), nil),
					},
				}, nil),
			},
		})
		defer cancel()
		_, err := ctx.LoadModuleByID("caddy.listeners.merge.listeners.point-c", json.RawMessage(`{"name": "test", "port": 80}`))
		require.ErrorContains(t, err, errExp.Error())
	})

	t.Run("valid", func(t *testing.T) {
		tn := NewTestNet(t)
		tl := NewStaticTestListener()
		tn.StartFn = func(fn pointc.RegisterFunc) error {
			return fn("test", &TestNet{
				ListenFn:    func(*net.TCPAddr) (net.Listener, error) { return tl, nil },
				LocalAddrFn: func() net.IP { return nil },
			})
		}

		ctx, cancel := NewCaddyContext(t, context.TODO(), caddy.Config{
			AppsRaw: map[string]json.RawMessage{
				"point-c": json.RawMessage(`{"networks": [{"type": "test-` + tn.Id() + `"}]}`),
			},
		})
		defer cancel()
		pcr, err := ctx.App("point-c")
		require.NoError(t, err)
		require.IsType(t, new(pointc.Pointc), pcr)
		pc := pcr.(*pointc.Pointc)
		require.NoError(t, pc.Start())
		defer func() { require.NoError(t, pc.Stop()) }()

		a, err := ctx.LoadModuleByID("caddy.listeners.merge.listeners.point-c", json.RawMessage(`{"name": "test", "port": 80}`))
		require.NoError(t, err)
		_, ok := a.(net.Listener)
		require.True(t, ok)
		require.Equal(t, tl.Close(), a.(net.Listener).Close())
		require.Equal(t, tl.Addr(), a.(net.Listener).Addr())
		ec, ee := tl.Accept()
		gc, ge := a.(net.Listener).Accept()
		require.Equal(t, ec, gc)
		require.Equal(t, ee, ge)
	})
}

type TestNet struct {
	ListenFn       func(addr *net.TCPAddr) (net.Listener, error)
	ListenPacketFn func(addr *net.UDPAddr) (net.PacketConn, error)
	DialerFn       func(laddr net.IP, port uint16) pointc.Dialer
	LocalAddrFn    func() net.IP
}

func (t *TestNet) Listen(addr *net.TCPAddr) (net.Listener, error)      { return t.ListenFn(addr) }
func (t *TestNet) ListenPacket(a *net.UDPAddr) (net.PacketConn, error) { return t.ListenPacketFn(a) }
func (t *TestNet) Dialer(laddr net.IP, port uint16) pointc.Dialer      { return t.DialerFn(laddr, port) }
func (t *TestNet) LocalAddr() net.IP                                   { return t.LocalAddrFn() }

type StaticTestListener struct {
	closeErr  error
	acceptErr error
	addr      net.Addr
	conn      net.Conn
}

func NewStaticTestListener() *StaticTestListener {
	return &StaticTestListener{
		closeErr:  errors.New(""),
		addr:      &net.TCPAddr{},
		acceptErr: errors.New(""),
		conn:      new(StaticTestConn),
	}
}

func (s *StaticTestListener) Accept() (net.Conn, error) { return s.conn, s.acceptErr }
func (s *StaticTestListener) Close() error              { return s.closeErr }
func (s *StaticTestListener) Addr() net.Addr            { return s.addr }

type StaticTestConn struct {
	remoteAddr net.Addr
	closeFn    func() error
}

func (*StaticTestConn) Read([]byte) (int, error)         { return 0, nil }
func (*StaticTestConn) Write([]byte) (int, error)        { return 0, nil }
func (*StaticTestConn) LocalAddr() net.Addr              { return nil }
func (stc *StaticTestConn) RemoteAddr() net.Addr         { return stc.remoteAddr }
func (*StaticTestConn) SetDeadline(time.Time) error      { return nil }
func (*StaticTestConn) SetReadDeadline(time.Time) error  { return nil }
func (*StaticTestConn) SetWriteDeadline(time.Time) error { return nil }
func (stc *StaticTestConn) Close() error {
	if stc.closeFn != nil {
		return stc.closeFn()
	}
	return nil
}

func NewCaddyContext(t testing.TB, base context.Context, cfg caddy.Config) (caddy.Context, context.CancelFunc) {
	t.Helper()
	ctx := caddy.Context{Context: base}

	// Set unexposed 'cfg' field
	f, ok := reflect.TypeOf(ctx).FieldByName("cfg")
	if !ok {
		t.Fatal("cannot populate config")
	}
	cfgPtr := uintptr(reflect.ValueOf(&ctx).UnsafePointer()) + f.Offset
	*(**caddy.Config)(unsafe.Pointer(cfgPtr)) = &cfg

	// Initialize 'apps' map
	f, ok = reflect.TypeOf(cfg).FieldByName("apps")
	if !ok {
		t.Fatal("cannot initialize apps")
	}
	appsPtr := uintptr(reflect.ValueOf(&cfg).UnsafePointer()) + f.Offset
	*(*map[string]caddy.App)(unsafe.Pointer(appsPtr)) = map[string]caddy.App{}

	// Return new context
	return caddy.NewContext(ctx)
}
