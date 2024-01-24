package configvalues_test

import (
	"fmt"
	"github.com/point-c/caddy/pkg/configvalues"
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
	for _, b := range b {
		require.NoError(t, testParseBool(t, b, expected))
	}
}

func testParseBool(t testing.TB, b string, expected bool) error {
	var vb configvalues.ValueBool
	if err := vb.UnmarshalText([]byte(b)); err != nil {
		return err
	}
	require.Exactly(t, expected, vb.Value())
	vb.Reset()
	require.Equal(t, *new(configvalues.ValueBool), vb)
	return nil
}

func TestValueString(t *testing.T) {
	var vs configvalues.ValueString
	const testStr = "foobar"
	require.NoError(t, vs.UnmarshalText([]byte(testStr)))
	require.Exactly(t, testStr, vs.Value())
	vs.Reset()
	require.Equal(t, *new(configvalues.ValueString), vs)
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
	t.Run(fmt.Sprintf("%T invalid parse %q", *new(N), b), func(t *testing.T) {
		var vu configvalues.ValueUnsigned[N]
		require.Error(t, vu.UnmarshalText([]byte(b)))
	})
}

func testValueUnsignedValid[N constraints.Unsigned](t *testing.T, n N) {
	t.Run(fmt.Sprintf("%[1]T parse %[1]d", n), func(t *testing.T) {
		var vu configvalues.ValueUnsigned[N]
		require.NoError(t, vu.UnmarshalText([]byte(fmt.Sprintf("%d", n))))
		require.Exactly(t, vu.Value(), n)
		vu.Reset()
		require.Equal(t, *new(configvalues.ValueUnsigned[N]), vu)
	})
}

func TestValueIP(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		var vip configvalues.ValueIP
		require.Error(t, vip.UnmarshalText([]byte{'%'}))
	})

	t.Run("ipv4", func(t *testing.T) {
		var vip configvalues.ValueIP
		v4 := net.IPv4(1, 1, 1, 1)
		b, err := v4.MarshalText()
		require.NoError(t, err)
		require.NoError(t, vip.UnmarshalText(b))
		require.Exactly(t, v4, vip.Value())
		vip.Reset()
		require.Equal(t, *new(configvalues.ValueIP), vip)
	})

	t.Run("ipv6", func(t *testing.T) {
		var vip configvalues.ValueIP
		v6 := net.ParseIP("abcd:23::33")
		b, err := v6.MarshalText()
		require.NoError(t, err)
		require.NoError(t, vip.UnmarshalText(b))
		require.Exactly(t, v6, vip.Value())
		vip.Reset()
		require.Equal(t, *new(configvalues.ValueIP), vip)
	})
}

func TestValueUDPAddr(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		var vu configvalues.ValueUDPAddr
		require.Error(t, vu.UnmarshalText([]byte{'.'}))
	})

	t.Run("ipv4", func(t *testing.T) {
		var vu configvalues.ValueUDPAddr
		addr, err := net.ResolveUDPAddr("udp", "1.1.1.1:0")
		require.NoError(t, err)
		require.NoError(t, vu.UnmarshalText([]byte(addr.String())))
		require.Exactly(t, addr, vu.Value())
		vu.Reset()
		require.Equal(t, *new(configvalues.ValueUDPAddr), vu)
	})
}

func TestValuePair(t *testing.T) {
	type testType = configvalues.ValuePair[uint16, configvalues.ValueUnsigned[uint16], *configvalues.ValueUnsigned[uint16]]
	failTests := func(t *testing.T, strs ...string) {
		for _, str := range strs {
			t.Run(str, func(t *testing.T) {
				var pp testType
				require.Error(t, pp.UnmarshalText([]byte(str)))
			})
		}
	}

	okTest := func(t *testing.T, str string, exp configvalues.PairValue[uint16]) {
		t.Run(str, func(t *testing.T) {
			var pp testType
			if err := pp.UnmarshalText([]byte(str)); err != nil {
				t.Fatal(err.Error())
			}
			if exp != *pp.Value() {
				t.Fatal("not equal")
			}
			pp.Reset()
			require.Equal(t, *new(testType), pp)
		})
	}

	t.Run("does not have a ':'", func(t *testing.T) {
		failTests(t, "", "/udp", "0.0.0.0/udp", "0.0.0.0")
	})

	t.Run("bad left", func(t *testing.T) {
		failTests(t, "..:1")
	})

	t.Run("bad right", func(t *testing.T) {
		failTests(t, "1:foo")
	})

	t.Run("bad src:dst", func(t *testing.T) {
		failTests(t, "a:b")
	})

	t.Run("valid", func(t *testing.T) {
		okTest(t, "1:2", configvalues.PairValue[uint16]{
			Left:  1,
			Right: 2,
		})
	})
}
