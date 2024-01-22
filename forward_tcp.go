package point_c

import (
	"context"
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/point-c/caddy/pkg/configvalues"
	"github.com/point-c/simplewg"
	"go.mrchanchal.com/zaphandler"
	"io"
	"log/slog"
	"net"
	"sync"
)

func init() {
	caddyreg.R[*ForwardTCP]()
}

var (
	_ ForwardProto          = (*ForwardTCP)(nil)
	_ caddy.Module          = (*ForwardTCP)(nil)
	_ caddyfile.Unmarshaler = (*ForwardTCP)(nil)
	_ caddy.Provisioner     = (*ForwardTCP)(nil)
	_ caddy.CleanerUpper    = (*ForwardTCP)(nil)
)

type (
	// ForwardTCP is able to forward TCP traffic through networks.
	ForwardTCP struct {
		Ports   configvalues.PortPair `json:"ports"`
		BufSize BufSize               `json:"buf"`
		logger  *slog.Logger
		ctx     context.Context
		cancel  func()
		wait    func()
	}
	BufSize = configvalues.CaddyTextUnmarshaler[uint16, configvalues.ValueUnsigned[uint16], *configvalues.ValueUnsigned[uint16]]
)

// Provision implements [caddy.Provisioner].
func (f *ForwardTCP) Provision(ctx caddy.Context) error {
	f.logger = slog.New(zaphandler.New(ctx.Logger()))
	f.ctx, f.cancel = context.WithCancel(ctx)
	return nil
}

// Cleanup implements [caddy.CleanerUpper].
func (f *ForwardTCP) Cleanup() error {
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

// CaddyModule implements [caddy.Module].
func (f *ForwardTCP) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[ForwardTCP, *ForwardTCP]("point-c.op.forward.tcp")
}

// UnmarshalCaddyfile unmarshals the caddyfile.
// Buffer size is the size of the buffer to use per stream direction.
// Buffer size will be double the specified amount per connection.
// ```
//
//	point-c netops {
//	    forward <src network name>:<dst network name> {
//			    tcp <src port>:<dst port> [buffer size]
//	    }
//	}
//
// ```
func (f *ForwardTCP) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.Val() == "tcp" {
			if !d.NextArg() {
				return d.ArgErr()
			}
		}

		if err := f.Ports.UnmarshalCaddyfile(d); err != nil {
			return err
		}

		if d.NextArg() {
			if err := f.BufSize.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start implements [ForwardProto]. It is responsible for starting the forwarding of network traffic.
func (f *ForwardTCP) Start(n *ForwardNetworks) error {
	ln, err := n.Src.Listen(&net.TCPAddr{IP: n.Src.LocalAddr(), Port: int(f.Ports.Value().Left)})
	if err != nil {
		return err
	}
	context.AfterFunc(f.ctx, func() { ln.Close() })

	var wg simplewg.Wg
	f.wait = wg.Wait

	bufSize := f.BufSize.Value()
	if bufSize == 0 {
		bufSize = 4096
	}
	copyBufs := sync.Pool{New: func() any { return make([]byte, bufSize) }}

	conns := make(chan net.Conn)
	wg.Go(func() { ListenLoop(ln, conns) })
	pairs := make(chan *ConnPair)
	var pairsListeners sync.WaitGroup
	dialed := make(chan *ConnPair)
	var dialedListeners sync.WaitGroup
	for i := 0; i < 10; i++ {
		pairsListeners.Add(1)
		dialedListeners.Add(1)
		wg.Go(func() { defer pairsListeners.Done(); PrepareConnPairLoop(f.ctx, f.logger, conns, pairs) })
		wg.Go(func() { defer dialedListeners.Done(); DialRemoteLoop(n.Dst, f.Ports.Value().Right, pairs, dialed) })
		wg.Go(func() {
			StartCopyLoop(dialed, func(done func(), logger *slog.Logger, dst io.Writer, src io.Reader) {
				buf := copyBufs.Get().([]byte)
				defer copyBufs.Put(buf)
				TcpCopy(done, logger, dst, src, buf)
			})
		})
	}
	return nil
}

// ListenLoop accepts connections and sends them to the next operation.
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

var connPairPool = sync.Pool{New: func() any { return new(ConnPair) }}

// PrepareConnPairLoop initializes the forwarding session.
func PrepareConnPairLoop(ctx context.Context, logger *slog.Logger, conns <-chan net.Conn, pairs chan<- *ConnPair) {
	defer close(pairs)
	for c := range conns {
		select {
		case <-ctx.Done():
			c.Close()
		default:
			cp := connPairPool.Get().(*ConnPair)
			cp.Tunnel = nil
			cp.Logger = logger
			// Copy c so it can be used in goroutines
			cp.Remote = c
			// Prepare connection specific context
			cp.Ctx, cp.Cancel = context.WithCancel(ctx)
			// Prevent leakage?
			context.AfterFunc(ctx, cp.Cancel)
			// Close remote connection when context is canceled
			context.AfterFunc(cp.Ctx, func() { cp.Remote.Close() })
			pairs <- cp
		}
	}
}

// ConnPair helps manage the state of a forwarding session.
type ConnPair struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Remote net.Conn
	Tunnel net.Conn
	Logger *slog.Logger
}

// DialTunnel does the actual remote dialing.
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

// DialRemoteLoop is responsible for dialing the receiver.
func DialRemoteLoop(n Net, dstPort uint16, pairs <-chan *ConnPair, dialed chan<- *ConnPair) {
	var wg simplewg.Wg
	// Wait for any senders on pairs to finish before closing
	defer func() { defer wg.Wait(); close(dialed) }()
	for c := range pairs {
		select {
		case <-c.Ctx.Done():
			connPairPool.Put(c)
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

// StartCopyLoop manages starting the copy for both TCP stream directions.
func StartCopyLoop(pairs <-chan *ConnPair, copyFn func(done func(), logger *slog.Logger, dst io.Writer, src io.Reader)) {
	var wg simplewg.Wg
	defer wg.Wait()
	for p := range pairs {
		select {
		case <-p.Ctx.Done():
			connPairPool.Put(p)
			p.Cancel()
		default:
			c := p
			var pwg sync.WaitGroup
			pwg.Add(2)
			wg.Go(func() { defer pwg.Done(); copyFn(c.Cancel, c.Logger, c.Remote, c.Tunnel) })
			wg.Go(func() { defer pwg.Done(); copyFn(c.Cancel, c.Logger, c.Tunnel, c.Remote) })
			go func() {
				defer connPairPool.Put(c)
				pwg.Wait()
			}()
		}
	}
}

// TcpCopy is the low level function that does the actual copying of TCP traffic. It only copies the stream in one direction e.g. src->dst or dst->src.
func TcpCopy(done func(), logger *slog.Logger, dst io.Writer, src io.Reader, buf []byte) {
	defer done()
	if _, err := io.CopyBuffer(dst, src, buf); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
		logger.Error("error copying data between connections", "error", err)
	}
}
