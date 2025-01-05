package types

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFloat32(t *testing.T) {
	f := float32(124.5768)
	v := NewFloat32(f)

	if !(v.V == f && v.Defined) {
		t.Fatalf("invalid OptFloat64.")
	}
}

func TestOptFloat32_Equal(t *testing.T) {
	v1 := NewFloat32(1.2)
	v2 := OptFloat32{}

	if v1.Equal(v2) {
		t.Fatalf("invalid equal")
	}

	v2.SetValue(1.2)

	if !v1.Equal(v2) {
		t.Fatalf("must be equal")
	}

	v1 = OptFloat32{}
	v2 = OptFloat32{}

	if !v2.Equal(v1) {
		t.Fatalf("must be equal")
	}
}

func TestOptFloat32_IsDefined(t *testing.T) {
	v := OptFloat32{}

	assert.Equal(t, v.IsDefined(), false)

	v.SetValue(1.76)

	assert.Equal(t, v.IsDefined(), true)
}

func TestOptFloat32_MarshalJSON(t *testing.T) {
	v := NewFloat32(9.65)

	b, err := v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "9.65")

	v.Undefine()
	b, err = v.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, string(b), "null")
}

func TestOptFloat32_Scan(t *testing.T) {
	v := NewFloat32(4)

	err := v.Scan(nil)
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, false)

	err = v.Scan(5.29)
	assert.Equal(t, v.V, float32(5.29))
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)

	err = v.Scan("1.32")
	assert.Equal(t, v.V, float32(1.32))
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)

	err = v.Scan([]byte("1.54"))
	assert.Equal(t, v.V, float32(1.54))
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)
}

func TestOptFloat32_SetValue(t *testing.T) {
	v := NewFloat32(1.87)

	v.SetValue(2.78)

	assert.Equal(t, v.V, float32(2.78))
	assert.Equal(t, v.Defined, true)
}

func TestOptFloat32_String(t *testing.T) {
	v := OptFloat32{}

	assert.Equal(t, v.String(), "<undefined>")

	v.SetValue(2.78)
	assert.Equal(t, v.String(), "2.78")
}

func TestOptFloat32_Undefine(t *testing.T) {
	v := NewFloat32(45.67)

	v.Undefine()

	assert.Equal(t, v.Defined, false)
	assert.Equal(t, v.V, float32(0))
}

func TestOptFloat32_UnmarshalJSON(t *testing.T) {
	v := NewFloat32(5.43)

	err := v.UnmarshalJSON([]byte("8.21"))
	assert.NoError(t, err)
	assert.Equal(t, v.Defined, true)
	assert.Equal(t, v.V, float32(8.21))

	err = v.UnmarshalJSON([]byte("fdgdshsh"))
	assert.Error(t, err)
}

func TestOptFloat32_Value(t *testing.T) {
	v := OptFloat32{}

	val, err := v.Value()
	assert.Nil(t, val)
	assert.NoError(t, err)

	v.SetValue(9.56)

	val, err = v.Value()

	assert.NoError(t, err)
	assert.Equal(t, val, float32(9.56))
}

func TestOptFloat32_LogValue(t *testing.T) {
	v := OptFloat32{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(float32(5))
	val = v.LogValue()

	assert.Equal(t, slog.KindFloat64, val.Kind())
}
