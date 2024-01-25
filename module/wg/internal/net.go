package internal

import (
	"context"
	pointc "github.com/point-c/caddy/module/point-c"
	"github.com/point-c/wg"
	"golang.zx2c4.com/wireguard/conn"
	"net"
)

type (
	Dialer   struct{ WgDialer }
	WgDialer interface {
		DialTCP(ctx context.Context, addr *net.TCPAddr) (net.Conn, error)
		DialUDP(addr *net.UDPAddr) (net.PacketConn, error)
	}
	WgNet interface {
		Listen(addr *net.TCPAddr) (net.Listener, error)
		ListenPacket(addr *net.UDPAddr) (net.PacketConn, error)
		Dialer(laddr net.IP, port uint16) *wg.Dialer
	}
)

func (c *Dialer) Dial(ctx context.Context, addr *net.TCPAddr) (net.Conn, error) {
	return c.DialTCP(ctx, addr)
}

func (c *Dialer) DialPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	return c.DialUDP(addr)
}

type Net struct {
	Net WgNet
	IP  net.IP
}

func (c *Net) Listen(addr *net.TCPAddr) (net.Listener, error) { return c.Net.Listen(addr) }
func (c *Net) LocalAddr() net.IP                              { return c.IP }
func (c *Net) ListenPacket(addr *net.UDPAddr) (net.PacketConn, error) {
	return c.Net.ListenPacket(addr)
}
func (c *Net) Dialer(laddr net.IP, port uint16) pointc.Dialer {
	return &Dialer{c.Net.Dialer(laddr, port)}
}

type TestBind struct{}

func (TestBind) Send(bufs [][]byte, ep conn.Endpoint) error    { return nil }
func (TestBind) ParseEndpoint(s string) (conn.Endpoint, error) { return nil, nil }
func (TestBind) BatchSize() int                                { return 0 }
func (TestBind) Close() error                                  { return nil }
func (TestBind) SetMark(uint32) error                          { return nil }
func (TestBind) Open(uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	return nil, 0, nil
}
