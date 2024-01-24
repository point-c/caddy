package test_caddy_test

import (
	. "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTestNetwork(t *testing.T) {
	v := NewTestNetwork(t)
	require.NotNil(t, v)
	require.NotEmpty(t, v.ID)
}
