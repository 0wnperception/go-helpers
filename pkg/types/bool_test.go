package types

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBool(t *testing.T) {
	b := NewBool(true)
	assert.Equal(t, b.V, true)
	assert.True(t, b.Defined)
}

func TestOptBool_IsDefined(t *testing.T) {
	v := OptBool{}

	assert.False(t, v.IsDefined())
}

func TestOptBool_MarshalJSON(t *testing.T) {
	v := OptBool{}

	b, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "null")

	v.SetValue(true)
	b, err = v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "true")
}

func TestOptBool_Scan(t *testing.T) {
	v := NewBool(true)

	err := v.Scan(nil)
	assert.NoError(t, err)
	assert.False(t, v.Defined)

	err = v.Scan(true)

	assert.NoError(t, err)
	assert.True(t, v.Defined)
	assert.True(t, v.V)

	v.Undefine()
	err = v.Scan("false")
	assert.NoError(t, err)
	assert.True(t, v.Defined)
	assert.False(t, v.V)

	err = v.Scan("fal")
	assert.Error(t, err)

	v.Undefine()
	err = v.Scan([]byte{'t', 'r', 'u', 'e'})

	assert.NoError(t, err)
	assert.True(t, v.Defined)
	assert.True(t, v.V)

	v.Undefine()
	err = v.Scan([]byte{'t', 'r', 'e'})

	assert.Error(t, err)

	v.Undefine()
	err = v.Scan(1)

	assert.Error(t, err)
}

func TestOptBool_SetValue(t *testing.T) {
	v := OptBool{}

	v.SetValue(false)

	assert.True(t, v.Defined)
	assert.False(t, v.V)
}

func TestOptBool_String(t *testing.T) {
	v := OptBool{}

	assert.Equal(t, v.String(), "<undefined>")

	v.SetValue(false)

	assert.Equal(t, v.String(), "false")
}

func TestOptBool_Undefine(t *testing.T) {
	v := NewBool(false)

	v.Undefine()

	assert.False(t, v.Defined)
}

func TestOptBool_UnmarshalJSON(t *testing.T) {
	v := NewBool(false)

	err := v.UnmarshalJSON([]byte("true"))
	assert.NoError(t, err)
	assert.True(t, v.V)
	assert.True(t, v.IsDefined())

	err = v.UnmarshalJSON([]byte("false"))
	assert.NoError(t, err)
	assert.False(t, v.V)
	assert.True(t, v.IsDefined())

	err = v.UnmarshalJSON([]byte("tre"))
	assert.Error(t, err)
}

func TestOptBool_Value(t *testing.T) {
	v := OptBool{}

	val, err := v.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)

	v.SetValue(true)
	val, err = v.Value()
	assert.NoError(t, err)
	assert.Equal(t, val, true)
}

func TestOptBool_LogValue(t *testing.T) {
	v := OptBool{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(true)
	val = v.LogValue()

	assert.Equal(t, slog.KindBool, val.Kind())
}
