package test_caddy

import (
	"context"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestNewTestDialer(t *testing.T) {
	require.NotNil(t, NewTestDialer(t))
}

func TestTestDialerDial(t *testing.T) {
	td := NewTestDialer(t)

	_, err := td.Dial(context.Background(), &net.TCPAddr{})
	require.EqualError(t, err, "dial not implemented")

	testConn := NewTestConn(t)
	td.DialFn = func(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
		return testConn, nil
	}
	conn, err := td.Dial(context.Background(), &net.TCPAddr{})
	require.NoError(t, err)
	require.Equal(t, testConn, conn)
}

func TestTestDialerDialPacket(t *testing.T) {
	td := NewTestDialer(t)

	_, err := td.DialPacket(&net.UDPAddr{})
	require.EqualError(t, err, "dialPacket not implemented")

	testConn := &net.UDPConn{}
	td.DialPacketFn = func(addr *net.UDPAddr) (net.PacketConn, error) {
		return testConn, nil
	}
	packetConn, err := td.DialPacket(&net.UDPAddr{})
	require.NoError(t, err)
	require.Equal(t, testConn, packetConn)
}
