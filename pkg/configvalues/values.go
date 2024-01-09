package configvalues

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"net"
	"slices"
	"strconv"
	"unsafe"
)

// ValueBool handles unmarshalling bool values.
type ValueBool bool

// UnmarshalText parses the bool with [strconv.ParseBool] internally.
func (b *ValueBool) UnmarshalText(text []byte) error {
	bb, err := strconv.ParseBool(string(text))
	if err != nil {
		return err
	}
	*b = ValueBool(bb)
	return nil
}

// Value returns the underlying bool value of ValueBool.
func (b *ValueBool) Value() bool {
	return bool(*b)
}

func (b *ValueBool) Reset() { *b = false }

// ValueString handles unmarshalling string values.
type ValueString string

// UnmarshalText just sets the value to string(b).
func (s *ValueString) UnmarshalText(b []byte) error {
	*s = ValueString(b)
	return nil
}

// Value returns the underlying string value of ValueString.
func (s *ValueString) Value() string {
	return string(*s)
}

func (s *ValueString) Reset() { *s = "" }

// ValueUnsigned is a generic type for unmarshalling an unsigned number.
// N must be an unsigned type (e.g., uint, uint32).
type ValueUnsigned[N constraints.Unsigned] struct{ V N }

// UnmarshalText parses the uint with [strconv.ParseUint] internally.
func (n *ValueUnsigned[N]) UnmarshalText(b []byte) error {
	var size int
	switch any(n.V).(type) {
	// uintptr and uint report -8 with binary.Size
	case uintptr, uint:
		size = int(unsafe.Sizeof(n.V))
	default:
		size = binary.Size(n.V)
	}

	i, err := strconv.ParseUint(string(b), 10, size*8)
	if err != nil {
		return err
	}
	n.V = N(i)
	return nil
}

// Value returns the underlying unsigned number of ValueUnsigned.
func (n *ValueUnsigned[N]) Value() N {
	return n.V
}

func (n *ValueUnsigned[N]) Reset() { n.V = 0 }

// ValueUDPAddr handles unmarshalling a [net.UDPAddr].
type ValueUDPAddr net.UDPAddr

// UnmarshalText implements the unmarshaling of text data into a UDP address.
// It resolves the text using [net.ResolveUDPAddr].
func (addr *ValueUDPAddr) UnmarshalText(text []byte) error {
	a, err := net.ResolveUDPAddr("udp", string(text))
	if err != nil {
		return err
	}
	*addr = (ValueUDPAddr)(*a)
	return nil
}

// Value returns the underlying net.UDPAddr of ValueUDPAddr.
func (addr *ValueUDPAddr) Value() *net.UDPAddr {
	return (*net.UDPAddr)(addr)
}

func (addr *ValueUDPAddr) Reset() { *addr = ValueUDPAddr{} }

// ValueIP handles unmarshalling [net.IP].
type ValueIP net.IP

// UnmarshalText implements the unmarshaling of text data into an IP address.
// It delegates to the [encoding.TextUnmarshaler] implementation of [net.IP].
func (ip *ValueIP) UnmarshalText(text []byte) error {
	return ((*net.IP)(ip)).UnmarshalText(text)
}

// Value returns the underlying net.IP of ValueIP.
func (ip *ValueIP) Value() net.IP {
	return net.IP(*ip)
}

func (ip *ValueIP) Reset() { *ip = nil }

type ValueProtocol string

func (p *ValueProtocol) UnmarshalText(text []byte) error {
	if !slices.Contains([]string{"tcp", "udp"}, string(text)) {
		return fmt.Errorf("unrecognized protocol %q", text)
	}
	*p = ValueProtocol(text)
	return nil
}

func (p *ValueProtocol) Value() string {
	if *p == "" {
		return "tcp"
	}
	return string(*p)
}

func (p *ValueProtocol) Reset() { *p = "" }

type ValuePortPair struct {
	src, dst Port
	proto    Protocol
	host     *IP
}

func (pp *ValuePortPair) UnmarshalText(b []byte) error {
	src, dst, ok := bytes.Cut(b, []byte{':'})
	if !ok {
		return errors.New("not a port:port pair")
	}

	host := src
	src, dst, ok = bytes.Cut(dst, []byte{':'})
	if ok {
		pp.host = new(IP)
		if err := pp.host.UnmarshalText(host); err != nil {
			return err
		}
	} else {
		pp.host = nil
		src, dst = host, src
	}

	dst, proto, ok := bytes.Cut(dst, []byte{'/'})
	if ok {
		if err := pp.proto.UnmarshalText(proto); err != nil {
			return err
		}
	}

	if err := errors.Join(pp.src.UnmarshalText(src), pp.dst.UnmarshalText(dst)); err != nil {
		return err
	}
	return nil
}

func (pp *ValuePortPair) Value() *PortPairValue {
	return &PortPairValue{
		Src:   pp.src.Value(),
		Dst:   pp.dst.Value(),
		IsUDP: pp.proto.Value() == "udp",
		Host: func() net.IP {
			if pp.host == nil {
				return net.IPv4zero
			}
			return pp.host.Value()
		}(),
	}
}

func (pp *ValuePortPair) Reset() {
	pp.host = nil
	pp.proto.original = ""
	pp.proto.value.Reset()
	pp.dst.original = ""
	pp.dst.value.Reset()
	pp.src.original = ""
	pp.src.value.Reset()
}
