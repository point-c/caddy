package configvalues

import (
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestPortPairValue_ToCaddyAddr(t *testing.T) {
	t.Run("udp", func(t *testing.T) {
		addr := (&PortPairValue{
			Src:   1,
			Dst:   2,
			IsUDP: true,
		}).ToCaddyAddr()
		require.Equal(t, "udp", addr.Network)
		require.Equal(t, uint(1), addr.StartPort)
		require.Equal(t, "", addr.Host)
	})
	t.Run("udp host", func(t *testing.T) {
		addr := (&PortPairValue{
			Src:   1,
			Dst:   2,
			IsUDP: true,
			Host:  net.IPv4(1, 2, 3, 4),
		}).ToCaddyAddr()
		require.Equal(t, "udp", addr.Network)
		require.Equal(t, uint(1), addr.StartPort)
		require.Equal(t, "1.2.3.4", addr.Host)
	})
	t.Run("tcp", func(t *testing.T) {
		addr := (&PortPairValue{
			Src: 1,
			Dst: 2,
		}).ToCaddyAddr()
		require.Equal(t, "tcp", addr.Network)
		require.Equal(t, uint(1), addr.StartPort)
		require.Equal(t, "", addr.Host)
	})
	t.Run("tcp host", func(t *testing.T) {
		addr := (&PortPairValue{
			Src:  1,
			Dst:  2,
			Host: net.IPv4(1, 2, 3, 4),
		}).ToCaddyAddr()
		require.Equal(t, "tcp", addr.Network)
		require.Equal(t, uint(1), addr.StartPort)
		require.Equal(t, "1.2.3.4", addr.Host)
	})
}
