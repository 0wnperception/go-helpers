package types

import (
	"bytes"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptInt32_NewInt32(t *testing.T) {
	v := NewInt32(5)
	if !(v.Defined && v.V == 5) {
		t.Fatalf("Invalid created OptInt64")
	}
}

func TestOptInt32_SetValue(t *testing.T) {
	v := OptInt32{}

	v.SetValue(10)
	if !(v.Defined && v.V == 10) {
		t.Fatalf("Invalid set OptInt64 value")
	}
}

func TestOptInt32_Undefine(t *testing.T) {
	v := NewInt32(10)

	v.Undefine()

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid Undefined call")
	}
}

func TestOptInt32_MarshalJSON(t *testing.T) {
	s := NewInt32(8765)

	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal JSON error. %v\n", err)
	}

	if bytes.Compare(b, []byte{'8', '7', '6', '5'}) != 0 {
		t.Fatalf("Invalid marshal json. Must be {'8', '7', '6', '5'}, but %+v\n", b)
	}

	s.Undefine()
	b, err = s.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal JSON error. %v\n", err)
	}

	if bytes.Compare(b, []byte{'n', 'u', 'l', 'l'}) != 0 {
		t.Fatalf("Invalid marshal json. Must be {'n', 'u', 'l', 'l'}, but %+v\n", b)
	}
}

func TestOptInt32_UnmarshalJSON(t *testing.T) {
	s := OptInt32{}

	err := s.UnmarshalJSON([]byte{'1', '3', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 135) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}
}

func TestOptInt32_String(t *testing.T) {
	v := OptInt32{}

	if v.String() != undef {
		t.Fatalf("Invalid string value for empty OptInt64\n")
	}

	v.SetValue(44444)
	if v.String() != "44444" {
		t.Fatalf("Invalid String(). Must be '44444', but '%s'\n", v.String())
	}
}

func TestOptInt32_Scan(t *testing.T) {
	v := NewInt32(4)

	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid scan for nil\n")
	}

	if err := v.Scan(int64(2)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 2) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 788956, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int32(2)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 2) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 2, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int16(3)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 3) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 3, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int8(2)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 2) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 2, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int64(math.MaxInt32 + 1)); err == nil {
		t.Fatalf("Scan must return error")
	}

	if err := v.Scan(int64(math.MinInt32 - 1)); err == nil {
		t.Fatalf("Scan must return error")
	}

	var i int32 = 788956
	if err := v.Scan(i); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 788956) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 788956, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan("345678"); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 345678) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 345678, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan([]byte("1345678")); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 1345678) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 1345678, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(time.Now()); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}
}

func TestOptInt32_Value(t *testing.T) {
	v := OptInt32{}

	a, err := v.Value()
	if err != nil {
		t.Fatalf("Value() error - %v\n", err)
	}

	if a != nil {
		t.Fatalf("Invalid value for empty OptInt64\n")
	}

	v.SetValue(678)

	a, err = v.Value()
	if err != nil {
		t.Fatalf("Value() error - %v\n", err)
	}

	if a.(int64) != 678 {
		t.Fatalf("Invalid value for OptInt64\n")
	}
}

func TestOptInt32_Equal(t *testing.T) {
	v1 := NewInt32(2)
	v2 := NewInt32(3)

	if v1.Equal(v2) {
		t.Fatalf("Invalid Equal result")
	}

	v2.SetValue(2)

	if !v1.Equal(v2) {
		t.Fatalf("Invalid Equal result")
	}

	v1.Undefine()
	if v1.Equal(v2) {
		t.Fatalf("Invalid Equal result")
	}

	v2.Undefine()
	if !v1.Equal(v2) {
		t.Fatalf("Invalid Equal result")
	}
}

func TestOptInt32_LogValue(t *testing.T) {
	v := OptInt32{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(3)
	val = v.LogValue()

	assert.Equal(t, slog.KindInt64, val.Kind())
}
