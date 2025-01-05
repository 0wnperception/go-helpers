package types

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUint(t *testing.T) {
	v := NewUint(50)

	if !(v.Defined && v.V == 50) {
		t.Fatalf("Invalid new value\n")
	}
}

func TestOptUInt_IsDefined(t *testing.T) {
	v := OptUInt{}

	if v.IsDefined() {
		t.Fatalf("Invalid Defined value")
	}
}

func TestOptUInt_MarshalJSON(t *testing.T) {
	s := NewUint(87)

	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal JSON error. %v\n", err)
	}

	if bytes.Compare(b, []byte{'8', '7'}) != 0 {
		t.Fatalf("Invalid marshal json. Must be {'8', '7'}, but %+v\n", b)
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

func TestOptUInt_Scan(t *testing.T) {
	v := NewUint(4)

	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid scan for nil\n")
	}

	var i int64 = 78
	if err := v.Scan(i); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 78) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 78, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan("34"); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 34) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 34, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan([]byte("134")); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 134) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 134, but Defined: %v, V: %v", v.Defined, v.V)
	}

	if err := v.Scan(time.Now()); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}
}

func TestOptUInt_String(t *testing.T) {
	v := OptUInt{}

	if v.String() != undef {
		t.Fatalf("Invalid string value for empty OptInt64\n")
	}

	v.SetValue(98)
	if v.String() != "98" {
		t.Fatalf("Invalid String(). Must be '98', but '%s'\n", v.String())
	}
}

func TestOptUInt_UnmarshalJSON(t *testing.T) {
	s := OptUInt{}

	err := s.UnmarshalJSON([]byte{'1', '3', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 135) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}
}

func TestOptUInt_Value(t *testing.T) {
	v := OptUInt{}

	a, err := v.Value()
	if err != nil {
		t.Fatalf("Value() error - %v\n", err)
	}

	if a != nil {
		t.Fatalf("Invalid value for empty OptInt64\n")
	}

	v.SetValue(67)

	a, err = v.Value()
	if err != nil {
		t.Fatalf("Value() error - %v\n", err)
	}

	if a.(int64) != 67 {
		t.Fatalf("Invalid value for OptInt64\n")
	}
}

func TestOptUInt_LogValue(t *testing.T) {
	v := OptUInt{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(2)
	val = v.LogValue()

	assert.Equal(t, slog.KindUint64, val.Kind())
}
