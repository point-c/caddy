// Package configvalues contains types to help with configuring point-c.
package configvalues

import (
	"github.com/point-c/wgapi"
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
	// Hosts are resolved to the IP they represent if possible.
	UDPAddr = CaddyTextUnmarshaler[*net.UDPAddr, ValueUDPAddr, *ValueUDPAddr]

	// IP is a type alias for handling IP addresses.
	// It wraps the [net.IP] type and uses [CaddyTextUnmarshaler] for converting text-based
	// IPv4 or IPv6 address representations into [net.IP].
	IP = CaddyTextUnmarshaler[net.IP, ValueIP, *ValueIP]

	// ResolvedIP is a type alias for handling hostnames that can be resolved into IP addresses.
	// If wraps the [net.IP] type and resolves the IP address when [CaddyTextUnmarshaler] is unmarshalling.
	ResolvedIP = CaddyTextUnmarshaler[net.IP, ValueResolvedIP, *ValueResolvedIP]

	// Hostname represents a unique hostname string.
	Hostname = String

	// String is a custom type leveraging [CaddyTextUnmarshaler] to handle text-based data.
	// It allows caddy's replacer to replace the string before it is used.
	String = CaddyTextUnmarshaler[string, ValueString, *ValueString]

	// PortPair represents a structured combination of ports formatted as <src>:<dst>.
	PortPair = CaddyTextUnmarshaler[*PairValue[uint16], ValuePair[uint16, ValueUnsigned[uint16], *ValueUnsigned[uint16]], *ValuePair[uint16, ValueUnsigned[uint16], *ValueUnsigned[uint16]]]
	// HostnamePair represents a structured combination of hostnames formatted as <hostname>:<hostname>.
	HostnamePair = CaddyTextUnmarshaler[*PairValue[string], ValuePair[string, ValueString, *ValueString], *ValuePair[string, ValueString, *ValueString]]

	// PrivateKey is a wireguard private key in base64 format.
	PrivateKey = CaddyTextUnmarshaler[wgapi.PrivateKey, valueKey[wgapi.PrivateKey], *valueKey[wgapi.PrivateKey]]
	// PublicKey is a wireguard public key in base64 format.
	PublicKey = CaddyTextUnmarshaler[wgapi.PublicKey, valueKey[wgapi.PublicKey], *valueKey[wgapi.PublicKey]]
	// PresharedKey is a wireguard preshared key in base64 format.
	PresharedKey = CaddyTextUnmarshaler[wgapi.PresharedKey, valueKey[wgapi.PresharedKey], *valueKey[wgapi.PresharedKey]]
)

// PairValue is used to store the base values of a parsed pair value.
type PairValue[T any] struct {
	Left, Right T
}
