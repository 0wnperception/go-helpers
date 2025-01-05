package jwriter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	s := "Some string\nAnother row"

	val := escapeIndex(s, false)

	t.Log("escapeIndex: ", val)

	val = escapeIndex("Some string Another row Я", false)

	t.Log("escapeIndex: ", val)
}

func TestTime(t *testing.T) {
	var tt time.Time

	ttype := reflect.TypeOf(tt)

	t.Log(ttype.Name())
	t.Log(ttype.PkgPath())

	t.Log(ttype == reflect.TypeOf((*time.Time)(nil)).Elem())

	timeKind := reflect.TypeOf(time.Time{}).Kind()

	if ttype.Kind() != timeKind {
		t.Error(ttype)
	}
}

func TestEncodeTime(t *testing.T) {
	jw := New()

	now := time.Now()

	jw.Time(now)
	b := jw.Buffer.BuildBytes()

	expected, err := now.UTC().MarshalJSON()

	require.NoError(t, err)
	require.Equal(t, expected, b)
}

func TestWriter_Bool(t *testing.T) {
	w := New()

	w.Bool(true)
	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `true`, string(b))

	w = New()
	w.Bool(false)
	b, err = w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `false`, string(b))
}

func TestWriter_Uint(t *testing.T) {
	w := New()

	w.Uint(1532)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `1532`, string(b))
}

func TestWriter_UintStr(t *testing.T) {
	w := New()

	w.UintStr(71)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"71"`, string(b))
}

func TestWriter_Uint8(t *testing.T) {
	w := New()

	w.Uint8(15)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `15`, string(b))
}

func TestWriter_Uint8Str(t *testing.T) {
	w := New()

	w.Uint8Str(7)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"7"`, string(b))
}

func TestWriter_Uint16(t *testing.T) {
	w := New()

	w.Uint16(1024)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `1024`, string(b))
}

func TestWriter_Uint16Str(t *testing.T) {
	w := New()

	w.Uint16Str(1025)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"1025"`, string(b))
}

func TestWriter_Uint32(t *testing.T) {
	w := New()

	w.Uint32(78123)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `78123`, string(b))
}

func TestWriter_Uint32Str(t *testing.T) {
	w := New()

	w.Uint32Str(78124)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"78124"`, string(b))
}

func TestWriter_Uint64(t *testing.T) {
	w := New()

	w.Uint64(3000000001)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `3000000001`, string(b))
}

func TestWriter_Uint64Str(t *testing.T) {
	w := New()

	w.Uint64Str(3000000021)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"3000000021"`, string(b))
}

func TestWriter_Int(t *testing.T) {
	w := New()

	w.Int(-123)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `-123`, string(b))
}

func TestWriter_IntStr(t *testing.T) {
	w := New()

	w.IntStr(-66784)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"-66784"`, string(b))
}

func TestWriter_Int8(t *testing.T) {
	w := New()

	w.Int8(-123)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `-123`, string(b))
}

func TestWriter_Int8Str(t *testing.T) {
	w := New()

	w.Int8Str(32)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"32"`, string(b))
}

func TestWriter_Int16(t *testing.T) {
	w := New()

	w.Int16(-12378)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `-12378`, string(b))
}

func TestWriter_Int16Str(t *testing.T) {
	w := New()

	w.Int16Str(32)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"32"`, string(b))
}

func TestWriter_Int32(t *testing.T) {
	w := New()

	w.Int32(-12378)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `-12378`, string(b))
}

func TestWriter_Int32Str(t *testing.T) {
	w := New()

	w.Int32Str(33897)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"33897"`, string(b))
}

func TestWriter_Int64(t *testing.T) {
	w := New()

	w.Int64(-3050000001)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `-3050000001`, string(b))
}

func TestWriter_Int64Str(t *testing.T) {
	w := New()

	w.Int64Str(3389712345)

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"3389712345"`, string(b))
}

func TestWriter_String1(t *testing.T) {
	w := New()

	w.String("abc\"d\refg\nWЖОПА")

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"abc\"d\refg\nWЖОПА"`, string(b))
}

func TestWriter_String2(t *testing.T) {
	w := New()

	w.String("abcdefghjklmnbvc")

	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `"abcdefghjklmnbvc"`, string(b))
}

func TestWriter_HexBytes(t *testing.T) {
	w := New()

	w.HexBytes([]byte{1, 2, 10})
	b, err := w.BuildBytes()

	require.NoError(t, err)
	require.Equal(t, `01020a`, string(b))
}

func TestWriter_Base64Bytes(t *testing.T) {
	w := New()

	w.Base64Bytes([]byte("abce"))

	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, `"YWJjZQ=="`, string(b))
}

func TestWriter_StringNoEscape(t *testing.T) {
	w := New()

	w.StringNoEscape("aa\"bbbbb")
	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, "\"aa\"bbbbb\"", string(b))
}

func TestWriter_Float32(t *testing.T) {
	w := New()

	w.Float32(float32(3.14))
	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, `3.14`, string(b))
}

func TestWriter_Float32Str(t *testing.T) {
	w := New()

	w.Float32Str(float32(3.145))
	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, `"3.145"`, string(b))
}

func TestWriter_Float64(t *testing.T) {
	w := New()

	w.Float64(3.14)
	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, `3.14`, string(b))
}

func TestWriter_Float64Str(t *testing.T) {
	w := New()

	w.Float64Str(3.145)
	b, err := w.BuildBytes()
	require.NoError(t, err)
	require.Equal(t, `"3.145"`, string(b))
}
