package types

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptInt64_NewInt64(t *testing.T) {
	v := NewInt64(5)
	if !(v.Defined && v.V == 5) {
		t.Fatalf("Invalid created OptInt64")
	}
}

func TestOptInt64_SetValue(t *testing.T) {
	v := OptInt64{}

	v.SetValue(10)
	if !(v.Defined && v.V == 10) {
		t.Fatalf("Invalid set OptInt64 value")
	}
}

func TestOptInt64_Undefine(t *testing.T) {
	v := NewInt64(10)

	v.Undefine()

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid Undefined call")
	}
}

func TestOptInt64_MarshalJSON(t *testing.T) {
	s := NewInt64(8765)

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

func TestOptInt64_UnmarshalJSON(t *testing.T) {
	s := OptInt64{}

	err := s.UnmarshalJSON([]byte{'1', '3', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 135) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}
}

func TestOptInt64_String(t *testing.T) {
	v := OptInt64{}

	if v.String() != undef {
		t.Fatalf("Invalid string value for empty OptInt64\n")
	}

	v.SetValue(9876543212)
	if v.String() != "9876543212" {
		t.Fatalf("Invalid String(). Must be '9876543212', but '%s'\n", v.String())
	}
}

func TestOptInt64_Scan(t *testing.T) {
	v := NewInt64(4)

	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid scan for nil\n")
	}

	var i int64 = 788956
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

func TestOptInt64_Value(t *testing.T) {
	v := OptInt64{}

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

func TestOptInt64_Equal(t *testing.T) {
	v1 := NewInt64(2)
	v2 := NewInt64(3)

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

func TestOptInt64_LogValue(t *testing.T) {
	v := OptInt64{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(3)
	val = v.LogValue()

	assert.Equal(t, slog.KindInt64, val.Kind())
}
