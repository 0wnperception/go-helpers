package types

import (
	"bytes"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInt8(t *testing.T) {
	v := NewInt8(2)

	if !(v.Defined && v.V == 2) {
		t.Fatalf("Invalid created OptInt8\n")
	}
}

func TestOptInt8_Equal(t *testing.T) {
	v1 := NewInt8(2)
	v2 := NewInt8(3)

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

func TestOptInt8_IsDefined(t *testing.T) {
	v := OptInt8{}

	if v.Defined {
		t.Fatalf("Invalid Defined value")
	}
}

func TestOptInt8_MarshalJSON(t *testing.T) {
	s := NewInt8(87)

	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal JSON error. %v\n", err)
	}

	if bytes.Compare(b, []byte{'8', '7'}) != 0 {
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

func TestOptInt8_Scan(t *testing.T) {
	v := NewInt8(4)

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

	i = 5
	if err := v.Scan(i); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 5) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 5, but Defined: %v, V: %v", v.Defined, v.V)
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

	if err := v.Scan(int64(3)); err != nil {
		t.Fatalf("Scan error - %v\n", err)
	}

	if !(v.Defined && v.V == 3) {
		t.Fatalf("Invalid scan value - Must be Defined: true, V: 3, but Defined: %v, V: %v", v.Defined, v.V)
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

	if err := v.Scan(int64(math.MaxInt8 + 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int64(math.MinInt8 - 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int32(math.MaxInt8 + 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}

	if err := v.Scan(int32(math.MinInt8 - 1)); err == nil {
		t.Fatalf("Invalid scan result, must be error\n")
	}
}

func TestOptInt8_SetValue(t *testing.T) {
	v := OptInt8{}

	v.SetValue(12)

	if !(v.Defined && v.V == 12) {
		t.Fatalf("Invalid set value")
	}
}

func TestOptInt8_String(t *testing.T) {
	v := OptInt8{}

	if s := v.String(); s != undef {
		t.Fatalf("Invalid String() - %v\n", s)
	}

	v = NewInt8(43)
	if s := v.String(); s != "43" {
		t.Fatalf("Invalid String() - %v\n", s)
	}
}

func TestOptInt8_Undefine(t *testing.T) {
	v := NewInt8(1)

	v.Undefine()
	if !(!v.Defined && v.V == 0) {
		t.Fatalf("Invalid Defined value")
	}
}

func TestOptInt8_UnmarshalJSON(t *testing.T) {
	s := OptInt8{}

	err := s.UnmarshalJSON([]byte{'1', '1', '5'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == 115) {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 135, but Defined: %v, V: %v\n", s.Defined, s.V)
	}

	err = s.UnmarshalJSON([]byte("null"))
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(!s.Defined && s.V == 0) {
		t.Fatalf("Invalid unmarshall\n")
	}
}

func TestOptInt8_Value(t *testing.T) {
	v := OptInt8{}

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

func TestOptInt8_LogValue(t *testing.T) {
	v := OptInt8{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(3)
	val = v.LogValue()

	assert.Equal(t, slog.KindInt64, val.Kind())
}
