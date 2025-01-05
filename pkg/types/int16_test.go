package types

import (
	"bytes"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptInt16_NewInt16(t *testing.T) {
	v := NewInt16(5)
	if !(v.Defined && v.V == 5) {
		t.Fatalf("Invalid created OptInt64")
	}
}

func TestOptInt16_SetValue(t *testing.T) {
	v := OptInt16{}

	v.SetValue(10)
	if !(v.Defined && v.V == 10) {
		t.Fatalf("Invalid set OptInt64 value")
	}
}

func TestOptInt16_Undefine(t *testing.T) {
	v := NewInt16(10)

	v.Undefine()

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid Undefined call")
	}
}

func TestOptInt16_MarshalJSON(t *testing.T) {
	s := NewInt16(8765)

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

func TestOptInt16_UnmarshalJSON(t *testing.T) {
	s := OptInt16{}

	err := s.UnmarshalJSON([]byte{'1', '3', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 135) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}
}

func TestOptInt16_String(t *testing.T) {
	v := OptInt16{}

	if v.String() != undef {
		t.Fatalf("Invalid string value for empty OptInt64\n")
	}

	v.SetValue(4444)
	if v.String() != "4444" {
		t.Fatalf("Invalid String(). Must be '4444', but '%s'\n", v.String())
	}
}

func TestOptInt16_Scan(t *testing.T) {
	v := NewInt16(4)

	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid scan for nil\n")
	}

	var i int32 = 788956
	if err := v.Scan(i); err == nil {
		t.Fatalf("Scan() must return error")
	}

	i = 1025
	if err := v.Scan(i); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 1025) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 1025, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan("58"); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 58) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 58, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan([]byte("57")); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 57) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 57, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(time.Now()); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int64(32)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 32) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 32, but Defined: %v, V: %v", v.Defined, v.V)
	}
	if err := v.Scan(int32(33)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 33) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 33, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int8(34)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 34) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 34, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(uint8(35)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 35) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 35, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(uint16(36)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 36) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 36, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(int64(math.MaxInt16 + 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int64(math.MinInt16 - 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int32(math.MaxInt16 + 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int32(math.MinInt16 - 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}
}

func TestOptInt16_Value(t *testing.T) {
	v := OptInt16{}

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

func TestOptInt16_Equal(t *testing.T) {
	v1 := NewInt16(2)
	v2 := NewInt16(3)

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

func TestOptInt16_LogValue(t *testing.T) {
	v := OptInt16{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(3)
	val = v.LogValue()

	assert.Equal(t, slog.KindInt64, val.Kind())
}
