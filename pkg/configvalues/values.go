package configvalues

import (
	"bytes"
	"encoding/binary"
	"errors"
	"golang.org/x/exp/constraints"
	"net"
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

// Reset sets the value to false.
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

// Reset sets the value to the empty string
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

// Reset sets this value to 0.
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

// Reset sets this value to an empty UDPAddr.
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

// Reset sets this value to nil.
func (ip *ValueIP) Reset() { *ip = nil }

// ValuePair represents a structured combination of `<value>:<value>` pairs.
type ValuePair[V any, T any, TP valueConstraint[V, T]] struct {
	left, right CaddyTextUnmarshaler[V, T, TP]
}

// UnmarshalText unmarshals a `<value>:<value>` pair.
func (pp *ValuePair[V, T, TP]) UnmarshalText(b []byte) error {
	left, right, ok := bytes.Cut(b, []byte{':'})
	if !ok {
		return errors.New("not a pair value")
	}
	return errors.Join(pp.left.UnmarshalText(left), pp.right.UnmarshalText(right))
}

// Value returns the pair's base values.
func (pp *ValuePair[V, T, TP]) Value() *PairValue[V] {
	return &PairValue[V]{
		Left:  pp.left.Value(),
		Right: pp.right.Value(),
	}
}

// Reset resets the pair values to their zero values.
func (pp *ValuePair[V, T, TP]) Reset() {
	for _, e := range []*CaddyTextUnmarshaler[V, T, TP]{&pp.left, &pp.right} {
		*e = CaddyTextUnmarshaler[V, T, TP]{}
	}
}
