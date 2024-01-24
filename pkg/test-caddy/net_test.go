package test_caddy

import (
	pointc "github.com/point-c/caddy"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestNewTestNet(t *testing.T) {
	require.NotNil(t, NewTestNet(t))
}

func TestTestNetListen(t *testing.T) {
	tn := NewTestNet(t)

	_, err := tn.Listen(&net.TCPAddr{})
	require.EqualError(t, err, "listen not implemented")

	testListener := NewTestListener(t)
	tn.ListenFn = func(addr *net.TCPAddr) (net.Listener, error) {
		return testListener, nil
	}
	listener, err := tn.Listen(&net.TCPAddr{})
	require.NoError(t, err)
	require.Equal(t, testListener, listener)
}

func TestTestNetListenPacket(t *testing.T) {
	tn := NewTestNet(t)

	_, err := tn.ListenPacket(&net.UDPAddr{})
	require.EqualError(t, err, "ListenPacket not implemented")

	testPacketConn := &net.UDPConn{}
	tn.ListenPacketFn = func(addr *net.UDPAddr) (net.PacketConn, error) {
		return testPacketConn, nil
	}
	packetConn, err := tn.ListenPacket(&net.UDPAddr{})
	require.NoError(t, err)
	require.Equal(t, testPacketConn, packetConn)
}

func TestTestNetDialer(t *testing.T) {
	tn := NewTestNet(t)

	require.NotNil(t, tn.Dialer(net.IPv4zero, 0))

	testDialer := NewTestDialer(t)
	tn.DialerFn = func(laddr net.IP, port uint16) pointc.Dialer {
		return testDialer
	}
	dialer := tn.Dialer(net.IPv4zero, 0)
	require.Equal(t, testDialer, dialer)
}

func TestTestNetLocalAddr(t *testing.T) {
	tn := NewTestNet(t)

	require.Equal(t, net.IPv4zero, tn.LocalAddr())

	testLocalAddr := net.IPv4allsys
	tn.LocalAddrFn = func() net.IP {
		return testLocalAddr
	}
	require.Equal(t, testLocalAddr, tn.LocalAddr())
}
