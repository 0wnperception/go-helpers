package jwriter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeIndex(t *testing.T) {
	tests := []struct {
		s          string
		escapeHTML bool
		result     int
	}{
		{s: `""`, escapeHTML: false, result: 0},
		{s: "1234", escapeHTML: false, result: -1},
		{s: `1234"`, escapeHTML: false, result: 4},
		{s: `123456789"`, escapeHTML: false, result: 9},
		{s: `123456789>`, escapeHTML: false, result: -1},

		{s: `""`, escapeHTML: true, result: 0},
		{s: "1234", escapeHTML: true, result: -1},
		{s: `1234"`, escapeHTML: true, result: 4},
		{s: `123456789"`, escapeHTML: true, result: 9},
		{s: `123456789>`, escapeHTML: true, result: 9},
		{s: `1234>56789`, escapeHTML: true, result: 4},
	}

	for i, tt := range tests {
		require.Equalf(t, tt.result, escapeIndex(tt.s, tt.escapeHTML), "index: %d", i)
	}
}
