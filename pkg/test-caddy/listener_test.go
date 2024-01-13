package test_caddy

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestNewTestListenerModule(t *testing.T) {
	v := NewTestListenerModule[any](t)
	require.NotNil(t, v)
	require.NotEmpty(t, v.ID)
}

func TestNewTestListener(t *testing.T) {
	require.NotNil(t, NewTestListener(t))
}

func TestTestListenerAccept(t *testing.T) {
	tl := NewTestListener(t)

	_, err := tl.Accept()
	require.EqualError(t, err, "accept not implemented")

	testConn := NewTestConn(t)
	tl.AcceptFn = func() (net.Conn, error) {
		return testConn, nil
	}
	conn, err := tl.Accept()
	require.NoError(t, err)
	require.Equal(t, testConn, conn)
}

func TestTestListenerClose(t *testing.T) {
	tl := NewTestListener(t)

	require.NoError(t, tl.Close())

	tl.CloseFn = func() error {
		return errors.New("close error")
	}
	require.EqualError(t, tl.Close(), "close error")
}

func TestTestListenerAddr(t *testing.T) {
	tl := NewTestListener(t)

	require.IsType(t, &net.TCPAddr{}, tl.Addr())

	mockAddr := &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 8080}
	tl.AddrFn = func() net.Addr {
		return mockAddr
	}
	require.Equal(t, mockAddr, tl.Addr())
}
