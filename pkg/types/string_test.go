package types

import (
	"bytes"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewString(t *testing.T) {
	s := NewString("abc")

	if s.V != "abc" {
		t.Fatalf("Invalid OptString value. Need 'abc', but '%s", s.V)
	}

	if !s.IsDefined() {
		t.Fatalf("Invalid Defined flag. Need true but %v", s.IsDefined())
	}
}

func TestStringSetValue(t *testing.T) {
	s := NewString("bcd")

	s.SetValue("dcn")

	if s.V != "dcn" {
		t.Fatalf("Invalid OptString value. Need 'dcn', but '%s", s.V)
	}

	if !s.IsDefined() {
		t.Fatalf("Invalid Defined flag. Need true but %v", s.IsDefined())
	}
}

func TestStringUndefine(t *testing.T) {
	s := NewString("321")

	s.Undefine()

	if s.V != "" {
		t.Fatalf("Invalid OptString value. Need '', but '%s", s.V)
	}

	if s.IsDefined() {
		t.Fatalf("Invalid Defined flag. Need false but %v", s.IsDefined())
	}
}

func TestStringGet(t *testing.T) {
	s := NewString("321")

	v := s.Get()
	str, ok := v.(string)
	if !ok {
		t.Fatalf("Value is not string")
	}

	if str != "321" {
		t.Fatalf("Invalid OptString value. Need '321', but '%s", s.V)
	}

	s.Undefine()
	v = s.Get()
	if v != nil {
		t.Fatalf("Not nil for emptry string")
	}
}

func TestStringMarshalJSON(t *testing.T) {
	s := NewString("8765")

	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal JSON error. %v\n", err)
	}

	if bytes.Compare(b, []byte{'"', '8', '7', '6', '5', '"'}) != 0 {
		t.Fatalf("Invalid marshal json. Must be {'\"', '8', '7', '6', '5', '\"'}, but %+v\n", b)
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

func TestStringUnmarshalJSON(t *testing.T) {
	s := OptString{}

	err := s.UnmarshalJSON([]byte{'"', 'r', '1', 't', '"'})
	if err != nil {
		t.Fatalf("Unmarshal JSON error. %v\n", err)
	}

	if !(s.Defined && s.V == "r1t") {
		t.Fatalf("Invalid unmarshalled value. Must be Defined: true, V: 'r1t', but Defined: %v, V: %s\n", s.Defined, s.V)
	}
}

func TestStringScan(t *testing.T) {
	s := OptString{}

	str := "123456"
	err := s.Scan(str)
	if err != nil {
		t.Fatalf("Scan error - %v", err)
	}

	if !(s.Defined && s.V == str) {
		t.Fatalf("Invalid scan. Must be Defined: true, V: '123456', but Defined: %v, V: %s\n", s.Defined, s.V)
	}

	err = s.Scan(nil)
	if err != nil {
		t.Fatalf("Scan error - %v", err)
	}

	if !(!s.Defined && s.V == "") {
		t.Fatalf("Invalid scan. Must be Defined: false, V: '', but Defined: %v, V: %s\n", s.Defined, s.V)
	}
}

func TestStringEqual(t *testing.T) {
	s1 := NewString("123")
	s2 := NewString("123")

	if !s1.Equal(s2) {
		t.Fatalf("Strings must be equal: %v, %v\n", s1, s2)
	}

	s1.Undefine()

	if s1.Equal(s2) {
		t.Fatalf("Strings must be different: %v, %v\n", s1, s2)
	}

	s2.Undefine()

	if !s1.Equal(s2) {
		t.Fatalf("Strings must be equal: %v, %v\n", s1, s2)
	}
}

func TestOptString_String(t *testing.T) {
	s := NewString("765")

	if s.String() != "765" {
		t.Fatalf("Invalid string value")
	}

	s.Undefine()

	if s.String() != "<undefined>" {
		t.Fatalf("Invalid indefined value")
	}
}

func TestOptString_SetValue(t *testing.T) {
	s := NewString("123")

	v, err := s.Value()
	if err != nil {
		t.Fatalf("Value error - %v\n", err)
	}

	str, ok := v.(string)
	if !ok {
		t.Fatalf("Value() - invalid type\n")
	}

	if str != "123" {
		t.Fatalf("Invalid string: must be '123', but '%s'\n", str)
	}

	s = OptString{}

	v, err = s.Value()
	if err != nil {
		t.Fatalf("Value error - %v", err)
	}

	if v != nil {
		t.Fatalf("Value() error, must be nil, but %v\n", v)
	}
}

func TestOptString_LogValue(t *testing.T) {
	v := OptString{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue("aa")
	val = v.LogValue()

	assert.Equal(t, slog.KindString, val.Kind())
}

func TestOptString_TextValue(t *testing.T) {
	v := NewString("123")
	ov, err := v.TextValue()

	require.NoError(t, err)
	require.Equal(t, "123", ov.String)
	require.True(t, ov.Valid)

	v = OptString{}

	ov, err = v.TextValue()
	require.NoError(t, err)
	require.Equal(t, "", ov.String)
	require.False(t, ov.Valid)
}

func TestOptString_ScanText(t *testing.T) {
	v := &OptString{}

	err := v.ScanText(pgtype.Text{
		String: "321",
		Valid:  true,
	})

	require.NoError(t, err)
	require.True(t, v.Defined)
	require.Equal(t, "321", v.V)

	err = v.ScanText(pgtype.Text{})
	require.NoError(t, err)
	require.False(t, v.Defined)
}

func TestOptString_MarshalText(t *testing.T) {
	v := NewString("123")

	b, err := v.MarshalText()

	require.NoError(t, err)
	require.Equal(t, []byte("123"), b)

	b, err = OptString{}.MarshalText()
	require.NoError(t, err)
	require.Nil(t, b)
}

func TestOptString_UnmarshalText(t *testing.T) {
	v := &OptString{}

	err := v.UnmarshalText([]byte("123"))
	require.NoError(t, err)
	require.True(t, v.Defined)
	require.Equal(t, "123", v.V)

	err = v.UnmarshalText(nil)
	require.NoError(t, err)
	require.False(t, v.Defined)
}
