package configvalues

import (
	"net"
)

// These are convenience types that may be used in configurations.
type (
	// Port defines a network port value, which is an unsigned 16-bit integer.
	// The validity of 0 as a port number depends on the specific use case or context.
	Port = CaddyTextUnmarshaler[uint16, ValueUnsigned[uint16], *ValueUnsigned[uint16]]

	// UDPAddr is a type alias for handling UDP network addresses.
	// It wraps the [net.UDPAddr] type and utilizes [CaddyTextUnmarshaler] for parsing
	// and handling UDP addresses in text form.
	UDPAddr = CaddyTextUnmarshaler[*net.UDPAddr, ValueUDPAddr, *ValueUDPAddr]

	// IP is a type alias for handling IP addresses.
	// It wraps the [net.IP] type and uses [CaddyTextUnmarshaler] for converting text-based
	// IPv4 or IPv6 address representations into [net.IP].
	IP = CaddyTextUnmarshaler[net.IP, ValueIP, *ValueIP]

	// Hostname represents a unique hostname string.
	Hostname = String

	// String is a custom type leveraging [CaddyTextUnmarshaler] to handle text-based data.
	// It allows caddy's replacer to replace the string before it is used.
	String = CaddyTextUnmarshaler[string, ValueString, *ValueString]

	// PortPair represents a structured combination of ports formatted as <src>:<dst>.
	PortPair = CaddyTextUnmarshaler[*PairValue[uint16], ValuePair[uint16, ValueUnsigned[uint16], *ValueUnsigned[uint16]], *ValuePair[uint16, ValueUnsigned[uint16], *ValueUnsigned[uint16]]]
	// HostnamePair represents a structured combination of hostnames formatted as <hostname>:<hostname>.
	HostnamePair = CaddyTextUnmarshaler[*PairValue[string], ValuePair[string, ValueString, *ValueString], *ValuePair[string, ValueString, *ValueString]]
)

type PairValue[T any] struct {
	Left, Right T
}
