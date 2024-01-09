package configvalues

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
	"net"
	"testing"
)

func TestValueBool(t *testing.T) {
	t.Run("true values", func(t *testing.T) {
		testParseBools(t, []string{"1", "t", "T", "TRUE", "true", "True"}, true)
	})

	t.Run("false values", func(t *testing.T) {
		testParseBools(t, []string{"0", "f", "F", "FALSE", "false", "False"}, false)
	})

	t.Run("invalid", func(t *testing.T) {
		require.Error(t, testParseBool(t, "", false))
	})
}

func testParseBools(t testing.TB, b []string, expected bool) {
	t.Helper()
	for _, b := range b {
		require.NoError(t, testParseBool(t, b, expected))
	}
}

func testParseBool(t testing.TB, b string, expected bool) error {
	t.Helper()
	var vb ValueBool
	if err := vb.UnmarshalText([]byte(b)); err != nil {
		return err
	}
	require.Exactly(t, expected, vb.Value())
	vb.Reset()
	require.Equal(t, *new(ValueBool), vb)
	return nil
}

func TestValueString(t *testing.T) {
	var vs ValueString
	const testStr = "foobar"
	require.NoError(t, vs.UnmarshalText([]byte(testStr)))
	require.Exactly(t, testStr, vs.Value())
	vs.Reset()
	require.Equal(t, *new(ValueString), vs)
}

func TestValueUnsigned(t *testing.T) {
	testValueUnsigned[uint](t)
	testValueUnsigned[uint8](t)
	testValueUnsigned[uint16](t)
	testValueUnsigned[uint32](t)
	testValueUnsigned[uint64](t)
	testValueUnsigned[uintptr](t)
}

func testValueUnsigned[N constraints.Unsigned](t *testing.T) {
	t.Helper()
	testValueUnsignedInvalid[N](t, "")
	testValueUnsignedInvalid[N](t, "abc")
	testValueUnsignedInvalid[N](t, "+123")
	testValueUnsignedInvalid[N](t, "-123")
	testValueUnsignedValid[N](t, 0)
	testValueUnsignedValid[N](t, 1)
	testValueUnsignedValid[N](t, 10)
	testValueUnsignedValid[N](t, ^(*new(N)))
}

func testValueUnsignedInvalid[N constraints.Unsigned](t *testing.T, b string) {
	t.Helper()
	t.Run(fmt.Sprintf("%T invalid parse %q", *new(N), b), func(t *testing.T) {
		var vu ValueUnsigned[N]
		require.Error(t, vu.UnmarshalText([]byte(b)))
	})
}

func testValueUnsignedValid[N constraints.Unsigned](t *testing.T, n N) {
	t.Helper()
	t.Run(fmt.Sprintf("%[1]T parse %[1]d", n), func(t *testing.T) {
		var vu ValueUnsigned[N]
		require.NoError(t, vu.UnmarshalText([]byte(fmt.Sprintf("%d", n))))
		require.Exactly(t, vu.Value(), n)
		vu.Reset()
		require.Equal(t, *new(ValueUnsigned[N]), vu)
	})
}

func TestValueIP(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		var vip ValueIP
		require.Error(t, vip.UnmarshalText([]byte{'%'}))
	})

	t.Run("ipv4", func(t *testing.T) {
		var vip ValueIP
		v4 := net.IPv4(1, 1, 1, 1)
		b, err := v4.MarshalText()
		require.NoError(t, err)
		require.NoError(t, vip.UnmarshalText(b))
		require.Exactly(t, v4, vip.Value())
		vip.Reset()
		require.Equal(t, *new(ValueIP), vip)
	})

	t.Run("ipv6", func(t *testing.T) {
		var vip ValueIP
		v6 := net.ParseIP("abcd:23::33")
		b, err := v6.MarshalText()
		require.NoError(t, err)
		require.NoError(t, vip.UnmarshalText(b))
		require.Exactly(t, v6, vip.Value())
		vip.Reset()
		require.Equal(t, *new(ValueIP), vip)
	})
}

func TestValueUDPAddr(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		var vu ValueUDPAddr
		require.Error(t, vu.UnmarshalText([]byte{'.'}))
	})

	t.Run("ipv4", func(t *testing.T) {
		var vu ValueUDPAddr
		addr, err := net.ResolveUDPAddr("udp", "1.1.1.1:0")
		require.NoError(t, err)
		require.NoError(t, vu.UnmarshalText([]byte(addr.String())))
		require.Exactly(t, addr, vu.Value())
		vu.Reset()
		require.Equal(t, *new(ValueUDPAddr), vu)
	})
}

func TestValuePortPair(t *testing.T) {
	failTests := func(t *testing.T, strs ...string) {
		t.Helper()
		for _, str := range strs {
			t.Run(str, func(t *testing.T) {
				var pp ValuePortPair
				require.Error(t, pp.UnmarshalText([]byte(str)))
			})
		}
	}

	okTest := func(t *testing.T, str string, exp PortPairValue) {
		t.Helper()
		t.Run(str, func(t *testing.T) {
			var pp ValuePortPair
			require.NoError(t, pp.UnmarshalText([]byte(str)))
			require.Equal(t, exp, *pp.Value())
			pp.Reset()
			require.Equal(t, *new(ValuePortPair), pp)
		})
	}

	t.Run("does not have a ':'", func(t *testing.T) {
		failTests(t, "", "/udp", "0.0.0.0/udp", "0.0.0.0")
	})

	t.Run("bad host", func(t *testing.T) {
		failTests(t, "..:1:2")
	})

	t.Run("bad protocol", func(t *testing.T) {
		failTests(t, "1:2/foo", "0.0.0.0:1:2/foo")
	})

	t.Run("bad src:dst", func(t *testing.T) {
		failTests(t,
			"0.0.0.0:a:2/tcp",
			"0.0.0.0:1:b/tcp",
			"0.0.0.0:a:b/tcp",
			"0.0.0.0:a:2",
			"0.0.0.0:1:b",
			"0.0.0.0:a:b",
			"a:2/tcp",
			"1:b/tcp",
			"a:b/tcp",
			"a:2",
			"1:b",
			"a:b",
		)
	})

	t.Run("valid", func(t *testing.T) {
		okTest(t, "1:2", PortPairValue{
			Src:  1,
			Dst:  2,
			Host: net.IPv4zero,
		})
		okTest(t, "1:2/udp", PortPairValue{
			Src:   1,
			Dst:   2,
			IsUDP: true,
			Host:  net.IPv4zero,
		})
		ip := net.IPv4(1, 2, 3, 4)
		okTest(t, "1.2.3.4:1:2", PortPairValue{
			Src:  1,
			Dst:  2,
			Host: ip,
		})
		okTest(t, "1.2.3.4:1:2/udp", PortPairValue{
			Src:   1,
			Dst:   2,
			IsUDP: true,
			Host:  ip,
		})
	})
}
