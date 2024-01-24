package test_caddy_test

import (
	. "github.com/point-c/caddy/pkg/test-caddy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTestNetOp(t *testing.T) {
	v := NewTestNetOp(t)
	require.NotNil(t, v)
	require.NotEmpty(t, v.ID)
	require.Equal(t, v, v.New())
}
