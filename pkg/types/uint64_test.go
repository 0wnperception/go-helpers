package types

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUint64(t *testing.T) {
	v := NewUint64(50)

	if !(v.Defined && v.V == 50) {
		t.Fatalf("Invalid new value\n")
	}
}

func TestOptUInt64_IsDefined(t *testing.T) {
	v := OptUInt64{}

	if v.IsDefined() {
		t.Fatalf("Invalid Defined value")
	}
}

func TestOptUInt64_MarshalJSON(t *testing.T) {
	s := NewUint64(8765)

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

func TestOptUInt64_Scan(t *testing.T) {
	v := NewUint64(4)

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

func TestOptUInt64_String(t *testing.T) {
	v := OptUInt64{}

	if v.String() != undef {
		t.Fatalf("Invalid string value for empty OptInt64\n")
	}

	v.SetValue(9876543212)
	if v.String() != "9876543212" {
		t.Fatalf("Invalid String(). Must be '9876543212', but '%s'\n", v.String())
	}
}

func TestOptUInt64_UnmarshalJSON(t *testing.T) {
	s := OptUInt64{}

	err := s.UnmarshalJSON([]byte{'1', '3', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 135) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}
}

func TestOptUInt64_Value(t *testing.T) {
	v := OptUInt64{}

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

func TestOptUInt64_LogValue(t *testing.T) {
	v := OptUInt64{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(2)
	val = v.LogValue()

	assert.Equal(t, slog.KindUint64, val.Kind())
}
