package caddyreg_test

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/google/uuid"
	"github.com/point-c/caddy/pkg/caddyreg"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestR(t *testing.T) {
	caddyreg.R[*TestModule]()
}

func TestInfo(t *testing.T) {
	m := new(TestModule).CaddyModule()
	require.Equal(t, caddy.ModuleID(strings.Join(moduleId, ".")), m.ID)
	require.NotNil(t, m.New)
	v := m.New()
	require.NotNil(t, v)
	require.IsType(t, new(TestModule), v)
}

var moduleId = []string{uuid.NewString(), uuid.NewString()}

type TestModule struct{}

func (*TestModule) CaddyModule() caddy.ModuleInfo {
	return caddyreg.Info[TestModule, *TestModule](moduleId...)
}
