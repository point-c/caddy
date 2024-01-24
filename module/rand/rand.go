package caddy_randhandler

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/point-c/caddy/pkg/caddyreg"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	caddyreg.R[*Rand]()
	httpcaddyfile.RegisterHandlerDirective("rand", func(httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) { return new(Rand), nil })
}

var (
	_ caddy.Module                = (*Rand)(nil)
	_ caddyhttp.MiddlewareHandler = (*Rand)(nil)
	_ caddyfile.Unmarshaler       = (*Rand)(nil)
)

// Rand struct represents the custom Caddy HTTP handler for generating random data.
type Rand struct{}

// UnmarshalCaddyfile implements [caddyfile.Unmarshaler].
func (r *Rand) UnmarshalCaddyfile(*caddyfile.Dispenser) error { return nil }

// CaddyModule implemnts [caddy.Module].
func (r *Rand) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[Rand, *Rand]("http.handlers.rand")
}

// ServeHTTP handles sending random data to the client. The following headers can be optionally set to modify the generation of data:
// - `Rand-Seed`: int64 seed value for the random data generator.
// - `Rand-Size`: int64 size in bytes of data to generate. A value less than zero represents an infinite stream.
func (r *Rand) ServeHTTP(resp http.ResponseWriter, req *http.Request, _ caddyhttp.Handler) (err error) {
	if err = req.Body.Close(); err == nil {
		if h := NewHeaderValues(req.Header); h.Size() < 0 {
			_, err = io.Copy(resp, NewRand(h.Seed()))
		} else {
			_, err = io.Copy(resp, io.LimitReader(NewRand(h.Seed()), h.Size()))
		}
	}
	return
}

// NewRand function creates a new random data generator.
var NewRand = func(seed int64) io.Reader { return rand.New(rand.NewSource(seed)) }

type (
	// HeaderValues struct holds the seed and size values extracted from headers.
	HeaderValues struct {
		seed HeaderValue[HeaderSeed]
		size HeaderValue[HeaderSize]
	}
	HeaderValue[K HeaderKey] int64
)

// NewHeaderValues creates a new [HeaderValues] instance from HTTP headers.
func NewHeaderValues(headers http.Header) *HeaderValues {
	var h HeaderValues
	h.seed.GetValue(headers)
	h.size.GetValue(headers)
	return &h
}

// Seed returns the seed value.
func (h *HeaderValues) Seed() int64 { return int64(h.seed) }

// Size returns the size value
func (h *HeaderValues) Size() int64 { return int64(h.size) }

// GetValue extracts a value from the HTTP headers.
func (h *HeaderValue[K]) GetValue(headers http.Header) {
	var k K
	*h = HeaderValue[K](k.Default())
	n, err := strconv.ParseInt(headers.Get(k.Key()), 10, 64)
	if err == nil {
		*h = HeaderValue[K](n)
	}
}

type (
	HeaderSeed struct{}
	HeaderSize struct{}
	HeaderKey  interface {
		Key() string
		Default() int64
	}
)

var nowSrc = time.Now

// Key returns `Rand-Seed`.
func (HeaderSeed) Key() string { return "Rand-Seed" }

// Default returns the current unix micro time.
func (HeaderSeed) Default() int64 { return nowSrc().UnixMicro() }

// Key return `Rand-Size`.
func (HeaderSize) Key() string { return "Rand-Size" }

// Default returns -1.
func (HeaderSize) Default() int64 { return -1 }
