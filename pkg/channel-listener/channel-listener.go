// Package channel_listener contains a listener that is able to pass arbitrary connections through a channel.
package channel_listener

import (
	"net"
	"sync/atomic"
)

// Listener is an implementation of a network listener that accepts connections
// from a channel. This allows for a flexible way to handle incoming connections.
type Listener struct {
	addr net.Addr        // addr represents the network address of the listener.
	c    <-chan net.Conn // c is the incoming [net.Conn]s.
	done chan struct{}   // done is used to signal the closing of the listener.

	// closeErr safely stores the error that occurs upon closing the listener.
	// Only the error from the first call to [Listener.CloseWithErr] will be set.
	// [Listener.CloseWithErr] will set the error to [net.ErrClosed] if called with nil.
	closeErr atomic.Pointer[error]
}

// New creates a new instance of Listener.
// [net.Conn] from the channel in are passed to [Listener.Accept] calls.
// addr is passed through to [Listener.Addr].
func New(in <-chan net.Conn, addr net.Addr) *Listener {
	cl := &Listener{
		c:    in,
		done: make(chan struct{}),
		addr: addr,
	}
	return cl
}

// Accept waits for and returns the next connection from the channel.
// If the listener is closed, it returns the error used to close.
func (d *Listener) Accept() (net.Conn, error) {
	for {
		select {
		case <-d.done:
			// If the listener is closed, return the error stored in closeErr.
			return nil, *d.closeErr.Load()
		case c, ok := <-d.c:
			// Retrieve the next connection from the channel.
			if !ok {
				// If the channel is closed, close listener and return the error
				d.Close()
				return nil, *d.closeErr.Load()
			}
			return c, nil
		}
	}
}

// Close closes the listener and signals any goroutines in [Listener.Accept] to stop waiting.
// It calls [Listener.CloseWithErr] with nil.
func (d *Listener) Close() error { return d.CloseWithErr(nil) }

// CloseWithErr allows closing the listener with a specific error.
// This error will be returned by [Listener.Accept] when the listener is closed.
// Only the error from the first time this is called will be returned.
// If the passed err is nil, the error used to close is [net.ErrClosed].
func (d *Listener) CloseWithErr(err error) error {
	if err == nil {
		err = net.ErrClosed
	}
	if d.closeErr.CompareAndSwap(nil, &err) {
		close(d.done)
	}
	return nil
}

// Addr returns the address of the listener
func (d *Listener) Addr() net.Addr { return d.addr }

// Done returns a channel that's closed when the listener is closed.
func (d *Listener) Done() <-chan struct{} { return d.done }
