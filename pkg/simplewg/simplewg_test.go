package simplewg

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleWg(t *testing.T) {
	var wg Wg
	var i int
	wg.Go(func() { i = 123 })
	wg.Wait()
	require.Equal(t, 123, i)
}
