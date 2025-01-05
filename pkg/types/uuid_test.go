package types

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUUIDHex  = "010002030405060708090a0b0c0d0e0f"
	testUUIDGUID = "01000203-0405-0607-0809-0a0b0c0d0e0f"
)

var (
	testUUID = []byte{1, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
)

func TestUUID_HexString(t *testing.T) {
	u, err := NewUUID(testUUID)

	require.NoError(t, err)
	require.Equal(t, testUUIDHex, u.HexString())
}

func TestUUID_String(t *testing.T) {
	u, err := NewUUID(testUUID)

	require.NoError(t, err)
	require.Equal(t, testUUIDGUID, u.String())
}

func TestNewUUIDFromString(t *testing.T) {
	u, err := NewUUID(testUUID)

	require.NoError(t, err)

	tests := []struct {
		val string
		u   UUID
		err bool
	}{
		{
			testUUIDGUID,
			u,
			false,
		},
		{
			"{01000203-0405-0607-0809-0a0b0c0d0e0f}",
			u,
			false,
		},
		{
			"{01000203-0405-0607-0809-0a0b0c0d0e0f}",
			u,
			false,
		},
		{
			"urn:uuid:01000203-0405-0607-0809-0a0b0c0d0e0f",
			u,
			false,
		},
		{
			"010002030405060708090a0b0c0d0e0f",
			u,
			false,
		},
		{
			"010002030405060708090a0b0c0d0e0",
			u,
			true,
		},
		{
			"urt:uuid:01000203-0405-0607-0809-0a0b0c0d0e0f",
			u,
			true,
		},
	}

	for i, tt := range tests {
		uuid, err := NewUUIDFromString(tt.val)
		if tt.err {
			require.Errorf(t, err, "test index %d", i)
		} else {
			require.NoErrorf(t, err, "test index %d", i)
			require.Equalf(t, u, uuid, "test index %d", i)
		}
	}
}

func TestUUID_MarshalJSON(t *testing.T) {
	u, err := NewUUID(testUUID)

	require.NoError(t, err)

	b, err := u.MarshalJSON()

	require.NoError(t, err)
	require.Equal(t, "\""+testUUIDGUID+"\"", string(b))
}

func TestUUID_UnmarshalJSON(t *testing.T) {
	u := UUID{}

	err := u.UnmarshalJSON([]byte("\"" + testUUIDGUID + "\""))
	require.NoError(t, err)

	expected, err := NewUUID(testUUID)
	require.NoError(t, err)
	require.Equal(t, expected, u)
}

func TestOptUUID_IsDefined(t *testing.T) {
	u, err := NewUUID(testUUID)

	require.NoError(t, err)

	ou := OptUUID{}

	require.False(t, ou.IsDefined())

	ou.SetValue(u)

	require.True(t, ou.IsDefined())

	ou.Undefine()

	require.False(t, ou.IsDefined())
}

func TestOptUUiD_LogValue(t *testing.T) {
	v := OptUUID{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(UUID(testUUID))
	val = v.LogValue()

	assert.Equal(t, slog.KindString, val.Kind())
}
