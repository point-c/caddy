package point_c

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/caddy/pkg/simplewg"
	"go.mrchanchal.com/zaphandler"
	"io"
	"log/slog"
	"net"
)

func init() {
	caddy.RegisterModule(new(ForwardTCP))
}

var (
	_ ForwardProto          = (*ForwardTCP)(nil)
	_ caddy.Module          = (*ForwardTCP)(nil)
	_ caddyfile.Unmarshaler = (*ForwardTCP)(nil)
	_ caddy.Provisioner     = (*ForwardTCP)(nil)
	_ caddy.CleanerUpper    = (*ForwardTCP)(nil)
)

type ForwardTCP struct {
	Ports  configvalues.PortPair `json:"ports"`
	logger *slog.Logger
	ctx    context.Context
	cancel func()
	wait   func()
}

func (f *ForwardTCP) Provision(ctx caddy.Context) error {
	f.logger = slog.New(zaphandler.New(ctx.Logger()))
	f.ctx, f.cancel = context.WithCancel(ctx)
	return nil
}

func (f *ForwardTCP) Cleanup() error { return f.Stop() }

func (f *ForwardTCP) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "point-c.op.forward.tcp",
		New: func() caddy.Module { return new(ForwardTCP) },
	}
}

func (f *ForwardTCP) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if err := f.Ports.UnmarshalCaddyfile(d); err != nil {
			return err
		} else if f.Ports.Value().IsUDP {
			return errors.New("cannot forward udp packets with tcp forwarder")
		}
	}
	return nil
}

func (f *ForwardTCP) Start(n Net) error {
	rawLn, err := f.Ports.Value().ToCaddyAddr().Listen(f.ctx, 0, net.ListenConfig{})
	if err != nil {
		return err
	}
	ln := rawLn.(net.Listener)
	context.AfterFunc(f.ctx, func() { ln.Close() })

	var wg simplewg.Wg
	f.wait = wg.Wait

	conns := make(chan net.Conn)
	wg.Go(func() { ListenLoop(ln, conns) })
	pairs := make(chan *ConnPair)
	wg.Go(func() { PrepareConnPairLoop(f.ctx, f.logger, conns, pairs) })
	dialed := make(chan *ConnPair)
	wg.Go(func() { DialRemoteLoop(n, f.Ports.Value().Dst, pairs, dialed) })
	wg.Go(func() { StartCopyLoop(dialed, TcpCopy) })
	return nil
}

func ListenLoop(ln net.Listener, conns chan<- net.Conn) {
	defer close(conns)
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		conns <- c
	}
}

func PrepareConnPairLoop(ctx context.Context, logger *slog.Logger, conns <-chan net.Conn, pairs chan<- *ConnPair) {
	defer close(pairs)
	for c := range conns {
		select {
		case <-ctx.Done():
			c.Close()
		default:
			// Copy c so it can be used in goroutines
			cp := ConnPair{Logger: logger, Remote: c}
			// Prepare connection specific context
			cp.Ctx, cp.Cancel = context.WithCancel(ctx)
			// Prevent leakage?
			context.AfterFunc(ctx, cp.Cancel)
			// Close remote connection when context is canceled
			context.AfterFunc(cp.Ctx, func() { cp.Remote.Close() })
			pairs <- &cp
		}
	}
}

type ConnPair struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Remote net.Conn
	Tunnel net.Conn
	Logger *slog.Logger
}

func (cp *ConnPair) DialTunnel(n Net, dstPort uint16) bool {
	// Prepare dialer that will preserve remote ip
	remote := cp.Remote.RemoteAddr().(*net.TCPAddr).IP
	d := n.Dialer(remote, 0)

	// Dial in the tunnel
	rc, err := d.Dial(cp.Ctx, &net.TCPAddr{IP: n.LocalAddr(), Port: int(dstPort)})
	if err != nil {
		cp.Logger.Error("failed to dial remote in tunnel", "local", remote, "remote", n.LocalAddr(), "port", dstPort)
		// Don't leak context, close remote connection
		cp.Cancel()
		// Remote might be temporarily down, don't kill everything because of one problem
		return false
	}
	cp.Tunnel = rc
	// Close tunnel connection when context is canceled
	context.AfterFunc(cp.Ctx, func() { cp.Tunnel.Close() })
	return true
}

func DialRemoteLoop(n Net, dstPort uint16, pairs <-chan *ConnPair, dialed chan<- *ConnPair) {
	var wg simplewg.Wg
	// Wait for any senders on pairs to finish before closing
	defer func() { defer wg.Wait(); close(dialed) }()
	for c := range pairs {
		select {
		case <-c.Ctx.Done():
			c.Cancel()
		default:
			c := c
			wg.Go(func() {
				if c.DialTunnel(n, dstPort) {
					dialed <- c
				}
			})
		}
	}
}

func StartCopyLoop(pairs <-chan *ConnPair, copyFn func(done func(), dst io.Writer, src io.Reader, logger *slog.Logger)) {
	var wg simplewg.Wg
	defer wg.Wait()
	for p := range pairs {
		select {
		case <-p.Ctx.Done():
			p.Cancel()
		default:
			c := p
			wg.Go(func() { copyFn(c.Cancel, c.Remote, c.Tunnel, c.Logger) })
			wg.Go(func() { copyFn(c.Cancel, c.Tunnel, c.Remote, c.Logger) })
		}
	}
}

func TcpCopy(done func(), dst io.Writer, src io.Reader, logger *slog.Logger) {
	defer done()
	if _, err := io.Copy(dst, src); err != nil {
		logger.Error("error copying data between connections", "error", err)
	}
}

func (f *ForwardTCP) Stop() error {
	if f.cancel != nil {
		f.cancel()
	}
	if f.wait != nil {
		f.wait()
	}
	f.ctx = nil
	f.wait = nil
	f.logger = nil
	f.cancel = nil
	return nil
}
