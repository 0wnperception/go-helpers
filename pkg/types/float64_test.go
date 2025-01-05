package types

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat64(t *testing.T) {
	f := 124.5768
	v := NewFloat64(f)

	if !(v.V == f && v.Defined) {
		t.Fatalf("invalid OptFloat64.")
	}
}

func TestOptFloat64_Equal(t *testing.T) {
	v1 := NewFloat64(1.2)
	v2 := OptFloat64{}

	if v1.Equal(v2) {
		t.Fatalf("invalid equal")
	}

	v2.SetValue(1.2)

	if !v1.Equal(v2) {
		t.Fatalf("must be equal")
	}

	v1 = OptFloat64{}
	v2 = OptFloat64{}

	if !v2.Equal(v1) {
		t.Fatalf("must be equal")
	}
}

func TestOptFloat64_IsDefined(t *testing.T) {
	v := OptFloat64{}

	assert.Equal(t, v.IsDefined(), false)

	v.SetValue(1.76)

	assert.Equal(t, v.IsDefined(), true)
}

func TestOptFloat64_MarshalJSON(t *testing.T) {
	v := NewFloat64(9.65)

	b, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "9.65")

	v.Undefine()
	b, err = v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "null")
}

func TestOptFloat64_Scan(t *testing.T) {
	v := NewFloat64(4)

	err := v.Scan(nil)
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, false)

	err = v.Scan(5.29)
	assert.Equal(t, v.V, 5.29)
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)

	err = v.Scan("1.32")
	assert.Equal(t, v.V, 1.32)
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)

	err = v.Scan([]byte("1.54"))
	assert.Equal(t, v.V, 1.54)
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)
}

func TestOptFloat64_SetValue(t *testing.T) {
	v := NewFloat64(1.87)

	v.SetValue(2.78)

	assert.Equal(t, v.V, 2.78)
	assert.Equal(t, v.Defined, true)
}

func TestOptFloat64_String(t *testing.T) {
	v := OptFloat64{}

	assert.Equal(t, v.String(), "<undefined>")

	v.SetValue(2.78)
	assert.Equal(t, v.String(), "2.78")
}

func TestOptFloat64_Undefine(t *testing.T) {
	v := NewFloat64(45.67)

	v.Undefine()

	assert.Equal(t, v.Defined, false)
	assert.Equal(t, v.V, float64(0))
}

func TestOptFloat64_UnmarshalJSON(t *testing.T) {
	v := NewFloat64(5.43)

	err := v.UnmarshalJSON([]byte("8.21"))
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)
	assert.Equal(t, v.V, 8.21)

	err = v.UnmarshalJSON([]byte("fdgdshsh"))
	assert.Error(t, err)
}

func TestOptFloat64_Value(t *testing.T) {
	v := OptFloat64{}

	val, err := v.Value()
	assert.Nil(t, val)
	assert.NoError(t, err)

	v.SetValue(9.56)

	val, err = v.Value()

	assert.NoError(t, err)
	assert.Equal(t, val, 9.56)
}

func TestOptFloat64_LogValue(t *testing.T) {
	v := OptFloat64{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(float64(3))
	val = v.LogValue()

	assert.Equal(t, slog.KindFloat64, val.Kind())
}
