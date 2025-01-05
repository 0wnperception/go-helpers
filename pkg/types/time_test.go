package types

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTime(t *testing.T) {
	tm := time.Now()

	ot := NewTime(tm)

	assert.True(t, ot.Defined)
	assert.Equal(t, ot.V, tm)
}

func TestOptTime_Equal(t *testing.T) {
	tt := time.Now()

	v1 := NewTime(tt)
	v2 := OptTime{}

	if v1.Equal(v2) {
		t.Fatalf("invalid equal")
	}

	v2.SetValue(tt)

	if !v1.Equal(v2) {
		t.Fatalf("must be equal")
	}

	v1 = OptTime{}
	v2 = OptTime{}

	if !v2.Equal(v1) {
		t.Fatalf("must be equal")
	}
}

func TestOptTime_String(t *testing.T) {
	tt := time.Now().UTC()

	v := NewTime(tt)
	assert.Equal(t, v.String(), fmt.Sprint(tt))

	v.Undefine()

	assert.Equal(t, v.String(), undef)
}

func TestOptTime_IsDefined(t *testing.T) {
	tm := OptTime{}

	assert.False(t, tm.IsDefined())

	tm.SetValue(time.Now())

	assert.True(t, tm.IsDefined())
}

func TestOptTime_Undefine(t *testing.T) {
	tm := NewTime(time.Now().UTC())

	assert.True(t, tm.IsDefined())

	tm.Undefine()

	assert.False(t, tm.IsDefined())
}

func TestOptTime_MarshalJSON(t *testing.T) {
	now := time.Now().UTC()

	expected, err := now.MarshalJSON()

	assert.NoError(t, err)

	tt := NewTime(now)

	tb, err := tt.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, expected, tb)
}

func TestOptTime_UnmarshalJSON(t *testing.T) {
	now := time.Now().UTC().Round(time.Microsecond)

	src, err := now.MarshalJSON()

	fmt.Printf("%s\n", string(src))

	assert.NoError(t, err)

	tt := OptTime{}

	err = tt.UnmarshalJSON(src)

	assert.NoError(t, err)
	assert.Equal(t, now, tt.V)
}

func TestOptTime_Scan(t *testing.T) {
	now := time.Now().UTC().Round(time.Microsecond)

	/*
		pgTimestamptzHourFormat   = "2006-01-02 15:04:05.999999999Z07"
			pgTimestamptzMinuteFormat = "2006-01-02 15:04:05.999999999Z07:00"
			pgTimestamptzSecondFormat = "2006-01-02 15:04:05.999999999Z07:00:00"
	*/
	str := now.Format(pgTimestamptzHourFormat)

	tests := []struct {
		val      any
		expected time.Time
		err      bool
	}{
		{
			val:      now,
			expected: now,
			err:      false,
		},
		{
			val:      NewTime(now),
			expected: now,
			err:      false,
		},
		{
			val:      str,
			expected: now,
			err:      false,
		},
		{
			val:      []byte(str),
			expected: now,
			err:      false,
		},
		{
			val:      now.Format(pgTimestamptzMinuteFormat),
			expected: now,
			err:      false,
		},
		{
			val:      now.Format(pgTimestamptzSecondFormat),
			expected: now,
			err:      false,
		},
	}

	for i, tt := range tests {
		v := OptTime{}

		err := v.Scan(tt.val)
		if tt.err {
			assert.Errorf(t, err, "test index %d", i)
		} else {
			assert.NoErrorf(t, err, "test index %d", i)
			assert.Equalf(t, tt.expected, v.V, "test index %d", i)
		}
	}
}

func TestOptTime_LogValue(t *testing.T) {
	v := OptTime{}

	val := v.LogValue()

	assert.Equal(t, slog.KindAny, val.Kind())

	v.SetValue(time.Now())
	val = v.LogValue()

	assert.Equal(t, slog.KindTime, val.Kind())
}
