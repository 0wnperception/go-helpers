package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNil(t *testing.T) {
	assert.Equal(t, true, IsNil(nil))

	var c any
	var p *int = nil
	c = p

	assert.Equal(t, false, c == nil)
	assert.Equal(t, true, IsNil(c))
}

func TestIsPointer(t *testing.T) {
	assert.Equal(t, false, IsPointer(2))

	i := 2
	assert.Equal(t, true, IsPointer(&i))
}

func TestDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan string)

	assert.False(t, Done(ctx, ch))

	cancel()

	assert.True(t, Done(ctx, ch))

	ctx, cancel = context.WithCancel(context.Background())
	close(ch)

	assert.True(t, Done(ctx, ch))

	cancel()
}

func TestCheckBase64URL(t *testing.T) {
	tests := []struct {
		s string
		r bool
	}{
		{"123", true},
		{"", true},
		{"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_", true},
		{"DFжопа", false},
		{",[]", false},
	}

	for i, v := range tests {
		require.Equalf(t, v.r, CheckBase64URL(v.s), "test index: %d", i)
	}
}
