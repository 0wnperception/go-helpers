package types

import (
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	tm := time.Now()

	ot := NewDate(tm)

	assert.True(t, ot.Defined)
	assert.Equal(t, ot.V, tm)
}

func TestOptDate_Equal(t *testing.T) {
	tt := time.Now()

	v1 := NewDate(tt)
	v2 := OptDate{}

	if v1.Equal(v2) {
		t.Fatalf("invalid equal")
	}

	v2.SetValue(tt)

	if !v1.Equal(v2) {
		t.Fatalf("must be equal")
	}

	v1 = OptDate{}
	v2 = OptDate{}

	if !v2.Equal(v1) {
		t.Fatalf("must be equal")
	}
}

func TestOptDate_String(t *testing.T) {
	tt := time.Now()

	v := NewDate(tt)
	assert.Equal(t, v.String(), tt.Format(dateFormat))

	v.Undefine()

	assert.Equal(t, v.String(), undef)
}

func TestOptDate_Value(t *testing.T) {
	od := OptDate{}

	v, err := od.Value()
	assert.NoError(t, err)
	assert.Nil(t, v)

	od.SetValue(time.Now())

	v, err = od.Value()
	assert.NoError(t, err)
	assert.Equal(t, od.V, v)
}

func TestOptDate_MarshalJSON(t *testing.T) {
	od := OptDate{}

	b, err := od.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "null", string(b))

	od.SetValue(time.Now())
	b, err = od.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"`+od.V.Format(dateFormat)+`"`, string(b))
}

func TestOptDate_UnmarshalJSON(t *testing.T) {
	od := OptDate{}
	od.SetValue(time.Now())

	err := od.UnmarshalJSON([]byte("null"))
	assert.NoError(t, err)
	assert.False(t, od.IsDefined())

	od.Undefine()

	err = od.UnmarshalJSON([]byte(`"2021-07-08"`))
	assert.NoError(t, err)
	assert.True(t, od.IsDefined())
	assert.Equal(t, 2021, od.V.Year())
	assert.Equal(t, time.July, od.V.Month())
	assert.Equal(t, 8, od.V.Day())
}

func TestOptDate_Scan(t *testing.T) {
	od := &OptDate{}

	od.SetValue(time.Now())

	err := od.Scan(nil)

	assert.NoError(t, err)
	assert.False(t, od.IsDefined())

	od.Undefine()

	err = od.Scan("2021-07-08")
	assert.NoError(t, err)
	assert.True(t, od.IsDefined())
	assert.Equal(t, 2021, od.V.Year())
	assert.Equal(t, time.July, od.V.Month())
	assert.Equal(t, 8, od.V.Day())

	err = od.Scan("aaa")
	assert.Error(t, err)

	err = od.Scan([]byte("2021-07-08"))
	assert.NoError(t, err)
	assert.True(t, od.IsDefined())
	assert.Equal(t, 2021, od.V.Year())
	assert.Equal(t, time.July, od.V.Month())
	assert.Equal(t, 8, od.V.Day())

	err = od.Scan(12)
	assert.ErrorIs(t, err, ErrConvert)
}

func TestOptDate_DateValue(t *testing.T) {
	od := OptDate{}

	d, err := od.DateValue()
	assert.NoError(t, err)
	assert.False(t, d.Valid)

	od.SetValue(time.Now())

	d, err = od.DateValue()
	assert.NoError(t, err)
	assert.True(t, d.Valid)
	assert.Equal(t, od.V, d.Time)
}

func TestOptDate_ScanDate(t *testing.T) {
	od := &OptDate{V: time.Now(), Defined: true}

	err := od.ScanDate(pgtype.Date{})

	assert.NoError(t, err)
	assert.False(t, od.Defined)

	od.Undefine()

	now := time.Now()
	err = od.ScanDate(pgtype.Date{Valid: true, Time: now})
	assert.NoError(t, err)
	assert.True(t, od.Defined)
	assert.Equal(t, now, od.V)
}

func TestOptDate_LogValue(t *testing.T) {
	v := OptDate{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(time.Now())
	val = v.LogValue()

	assert.Equal(t, slog.KindString, val.Kind())
}
