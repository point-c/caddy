package test_caddy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTestNetwork(t *testing.T) {
	v := NewTestNetwork(t)
	require.NotNil(t, v)
	require.NotEmpty(t, v.ID)
}
