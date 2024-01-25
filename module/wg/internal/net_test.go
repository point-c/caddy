package internal

import (
	"context"
	"errors"
	test_caddy "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/point-c/wg"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func Test_Bind(t *testing.T) {
	var b TestBind
	require.NoError(t, b.Close())
	require.NoError(t, b.Send(nil, nil))
	require.NoError(t, b.SetMark(0))
	require.Equal(t, 0, b.BatchSize())
	fns, p, err := b.Open(0)
	require.NoError(t, err)
	require.Equal(t, uint16(0), p)
	require.Empty(t, fns)
	v, err := b.ParseEndpoint("")
	require.NoError(t, err)
	require.Nil(t, v)
}

type TestWgDialer struct {
	DialTCPFn func(ctx context.Context, addr *net.TCPAddr) (net.Conn, error)
	DialUDPFn func(addr *net.UDPAddr) (net.PacketConn, error)
}

func (m *TestWgDialer) DialTCP(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	if m.DialTCPFn != nil {
		return m.DialTCPFn(ctx, addr)
	}
	return &net.TCPConn{}, nil
}

func (m *TestWgDialer) DialUDP(addr *net.UDPAddr) (net.PacketConn, error) {
	if m.DialUDPFn != nil {
		return m.DialUDPFn(addr)
	}
	return &net.UDPConn{}, nil
}

func TestClientDialer_Dial(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testWgDialer := &TestWgDialer{}
		clientDialer := &Dialer{testWgDialer}
		_, err := clientDialer.Dial(context.Background(), &net.TCPAddr{})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		testWgDialer := &TestWgDialer{}
		clientDialer := &Dialer{testWgDialer}
		errExp := errors.New("dial error")
		testWgDialer.DialTCPFn = func(context.Context, *net.TCPAddr) (net.Conn, error) { return nil, errExp }
		_, err := clientDialer.Dial(context.Background(), &net.TCPAddr{})
		require.ErrorIs(t, err, errExp)
	})
}

func TestClientDialer_DialPacket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testWgDialer := &TestWgDialer{}
		clientDialer := &Dialer{testWgDialer}
		_, err := clientDialer.DialPacket(&net.UDPAddr{})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		testWgDialer := &TestWgDialer{}
		clientDialer := &Dialer{testWgDialer}
		errExp := errors.New("dial error")
		testWgDialer.DialUDPFn = func(*net.UDPAddr) (net.PacketConn, error) { return nil, errExp }
		_, err := clientDialer.DialPacket(&net.UDPAddr{})
		require.ErrorIs(t, err, errExp)
	})
}

func TestClientNet_Listen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n := NewTestNet(t)
		n.ListenFn = func(*net.TCPAddr) (net.Listener, error) { return nil, nil }
		_, err := (&Net{Net: n}).Listen(&net.TCPAddr{})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		n := NewTestNet(t)
		clientNet := &Net{Net: n}
		errExp := errors.New("listen error")
		n.ListenFn = func(*net.TCPAddr) (net.Listener, error) { return nil, errExp }
		_, err := clientNet.Listen(&net.TCPAddr{})
		require.ErrorIs(t, err, errExp)
	})
}

func TestClientNet_LocalAddr(t *testing.T) {
	n := NewTestNet(t)
	clientNet := &Net{Net: n}
	require.NoError(t, clientNet.IP.UnmarshalText([]byte("1.2.3.4")))
	require.Equal(t, net.IPv4(1, 2, 3, 4), clientNet.LocalAddr())
}

func TestClientNet_ListenPacket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n := NewTestNet(t)
		n.ListenPacketFn = func(*net.UDPAddr) (net.PacketConn, error) { return nil, nil }
		_, err := (&Net{Net: n}).ListenPacket(&net.UDPAddr{})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		n := NewTestNet(t)
		clientNet := &Net{Net: n}
		errExp := errors.New("listen packet error")
		n.ListenPacketFn = func(*net.UDPAddr) (net.PacketConn, error) { return nil, errExp }
		_, err := clientNet.ListenPacket(&net.UDPAddr{})
		require.ErrorIs(t, err, errExp)
	})
}

func TestClientNet_Dialer(t *testing.T) {
	require.IsType(t, &Dialer{}, (&Net{Net: NewTestNet(t)}).Dialer(net.IPv4zero, 0))
}

type TestNet struct {
	test_caddy.TestNet
	DialerFn func(net.IP, uint16) *wg.Dialer
}

func (t *TestNet) Dialer(laddr net.IP, port uint16) *wg.Dialer {
	if t.DialerFn != nil {
		return t.DialerFn(laddr, port)
	}
	return nil
}

func NewTestNet(t *testing.T) *TestNet {
	return &TestNet{TestNet: *test_caddy.NewTestNet(t)}
}
