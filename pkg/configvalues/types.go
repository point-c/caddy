package configvalues

import (
	"github.com/caddyserver/caddy/v2"
	"net"
	"strconv"
	"strings"
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

	// PortPair represents a structured combination of ports and, optionally, their host and protocol,
	// formatted as [<host>:]<src>:<dst>[/<tcp|udp>].
	PortPair = CaddyTextUnmarshaler[*PortPairValue, ValuePortPair, *ValuePortPair]

	Protocol = CaddyTextUnmarshaler[string, ValueProtocol, *ValueProtocol]
)

type PortPairValue struct {
	Src, Dst uint16
	IsUDP    bool
	Host     net.IP
}

func (pp *PortPairValue) ToCaddyAddr() caddy.NetworkAddress {
	var addrStr strings.Builder
	if pp.IsUDP {
		addrStr.WriteString("udp/")
	}

	if !(pp.Host == nil || pp.Host.Equal(net.IPv4zero) || pp.Host.Equal(net.IPv6unspecified)) {
		addrStr.WriteString(pp.Host.String())
	}
	addrStr.WriteRune(':')
	addrStr.Write(strconv.AppendInt(nil, int64(pp.Src), 10))
	// The string we built will always be parsable
	addr, _ := caddy.ParseNetworkAddress(addrStr.String())
	return addr
}
