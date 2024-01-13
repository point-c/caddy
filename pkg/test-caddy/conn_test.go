package test_caddy

import (
	"errors"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func TestNewTestConn(t *testing.T) {
	require.NotNil(t, NewTestConn(t))
}

func TestTestConnRead(t *testing.T) {
	conn := NewTestConn(t)

	_, err := conn.Read(nil)
	require.EqualError(t, err, "read not implemented")

	conn.ReadFn = func(b []byte) (int, error) { return len(b), nil }
	n, err := conn.Read(make([]byte, 10))
	require.NoError(t, err)
	require.Equal(t, 10, n)
}

func TestTestConnWrite(t *testing.T) {
	conn := NewTestConn(t)

	_, err := conn.Write(nil)
	require.EqualError(t, err, "write not implemented")

	conn.WriteFn = func(b []byte) (int, error) { return len(b), nil }
	n, err := conn.Write(make([]byte, 10))
	require.NoError(t, err)
	require.Equal(t, 10, n)
}

func TestTestConnClose(t *testing.T) {
	conn := NewTestConn(t)

	require.NoError(t, conn.Close())

	conn.CloseFn = func() error { return errors.New("close error") }
	require.EqualError(t, conn.Close(), "close error")
}

func TestTestConnLocalAddr(t *testing.T) {
	conn := NewTestConn(t)

	require.NotNil(t, conn.LocalAddr())

	conn.LocalAddrFn = func() net.Addr { return &net.TCPAddr{} }
	require.IsType(t, &net.TCPAddr{}, conn.LocalAddr())
}

func TestTestConnRemoteAddr(t *testing.T) {
	conn := NewTestConn(t)

	require.NotNil(t, conn.RemoteAddr())

	conn.RemoteAddrFn = func() net.Addr { return &net.TCPAddr{} }
	require.IsType(t, &net.TCPAddr{}, conn.RemoteAddr())
}

func TestTestConnSetDeadline(t *testing.T) {
	conn := NewTestConn(t)
	deadline := time.Now()

	require.NoError(t, conn.SetDeadline(deadline))

	conn.SetDeadlineFn = func(time.Time) error { return errors.New("deadline error") }
	require.EqualError(t, conn.SetDeadline(deadline), "deadline error")
}

func TestTestConnSetReadDeadline(t *testing.T) {
	conn := NewTestConn(t)
	deadline := time.Now()

	require.NoError(t, conn.SetReadDeadline(deadline))

	conn.SetReadDeadlineFn = func(time.Time) error { return errors.New("read deadline error") }
	require.EqualError(t, conn.SetReadDeadline(deadline), "read deadline error")
}

func TestTestConnSetWriteDeadline(t *testing.T) {
	conn := NewTestConn(t)
	deadline := time.Now()

	require.NoError(t, conn.SetWriteDeadline(deadline))

	conn.SetWriteDeadlineFn = func(time.Time) error { return errors.New("write deadline error") }
	require.EqualError(t, conn.SetWriteDeadline(deadline), "write deadline error")
}
