package jwriter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendInt(t *testing.T) {
	for i, test := range []struct {
		toFormat int64
		want     string
	}{
		{toFormat: 0, want: "0"},
		{toFormat: 0, want: "0"},
		{toFormat: 1, want: "1"},
		{toFormat: -1, want: "-1"},
		{toFormat: 78, want: "78"},
		{toFormat: -78, want: "-78"},
		{toFormat: -101, want: "-101"},
		{toFormat: 101, want: "101"},
		{toFormat: -1021, want: "-1021"},
		{toFormat: 1021, want: "1021"},
		{toFormat: -31021, want: "-31021"},
		{toFormat: 31021, want: "31021"},
	} {
		b := appendInt(nil, test.toFormat)
		require.Equalf(t, test.want, string(b), "test index: %d", i)
	}
}

func TestAppendUint(t *testing.T) {
	for i, test := range []struct {
		toFormat uint64
		want     string
	}{
		{toFormat: 0, want: "0"},
		{toFormat: 0, want: "0"},
		{toFormat: 1, want: "1"},
		{toFormat: 78, want: "78"},
		{toFormat: 101, want: "101"},
		{toFormat: 1021, want: "1021"},
		{toFormat: 31021, want: "31021"},
	} {
		b := appendUint(nil, test.toFormat)
		require.Equalf(t, test.want, string(b), "test index: %d", i)
	}
}
