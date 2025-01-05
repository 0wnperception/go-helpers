package slice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestByteSlice2String(t *testing.T) {
	b := []byte{'a', 'b', 'c'}

	if ToString(b) != "abc" {
		t.Fatalf("Invalid convert")
	}
}

func TestStringToBytes(t *testing.T) {
	str := "Hello"

	b := ToBytes(str)

	require.Equal(t, 5, len(b))
	require.Equal(t, 5, cap(b))
	require.Equal(t, []byte{'H', 'e', 'l', 'l', 'o'}, b)

	require.Nil(t, ToBytes(""))
}
