// //nolint: revive
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

const (
	pgTimestamptzHourFormat   = "2006-01-02 15:04:05.999999999Z07"
	pgTimestamptzMinuteFormat = "2006-01-02 15:04:05.999999999Z07:00"
	pgTimestamptzSecondFormat = "2006-01-02 15:04:05.999999999Z07:00:00"
)

type OptTime struct {
	V       time.Time
	Defined bool
}

func NewTime(v time.Time) OptTime {
	return OptTime{V: v, Defined: true}
}

func (v *OptTime) SetValue(val time.Time) {
	v.V, v.Defined = val, true
}

func (v *OptTime) Undefine() {
	v.V, v.Defined = time.Now(), false
}

func (v OptTime) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Time(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptTime) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		v.V, v.Defined = time.Time{}, false
	} else {
		str := lexer.String()

		parsedTime, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			parsedTime, err = time.Parse(time.RFC3339, lexer.String())
		}

		if err != nil {
			lexer.AddError(err)
		}

		v.V = normalizePotentialUTC(parsedTime)
		v.Defined = true
	}
}

func (v *OptTime) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptTime) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptTime) String() string {
	if !v.Defined {
		return undef
	}

	return v.V.String()
}

func (v OptTime) IsDefined() bool {
	return v.Defined
}

func (v *OptTime) Scan(value any) error {
	if value == nil {
		v.Defined = false

		return nil
	}

	switch src := value.(type) {
	case string:
		t, err := parseString(src)
		if err != nil {
			return err
		}

		*v = NewTime(t)

		return nil
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)

		t, err := parseText(srcCopy)
		if err != nil {
			return err
		}

		*v = NewTime(t)

		return nil
	case time.Time:
		v.V, v.Defined = src, true

		return nil
	case OptTime:
		v.V, v.Defined = src.V, src.Defined

		return nil
	}

	return fmt.Errorf("cannot scan %T: %w", value, ErrDecode)
}

func parseString(sbuf string) (time.Time, error) {
	bc := false

	if strings.HasSuffix(sbuf, " BC") {
		sbuf = sbuf[:len(sbuf)-3]
		bc = true
	}

	var format string

	// //nolint:gocritic
	if len(sbuf) >= 9 && (sbuf[len(sbuf)-9] == '-' || sbuf[len(sbuf)-9] == '+') {
		format = pgTimestamptzSecondFormat
	} else if len(sbuf) >= 6 && (sbuf[len(sbuf)-6] == '-' || sbuf[len(sbuf)-6] == '+') {
		format = pgTimestamptzMinuteFormat
	} else {
		format = pgTimestamptzHourFormat
	}

	tim, err := time.Parse(format, sbuf)
	if err != nil {
		return time.Now(), err
	}

	if bc {
		year := -tim.Year() + 1
		tim = time.Date(year, tim.Month(), tim.Day(), tim.Hour(), tim.Minute(), tim.Second(), tim.Nanosecond(), tim.Location())
	}

	return tim, nil
}

func parseText(src []byte) (time.Time, error) {
	return parseString(string(src))
}

// Value implements the driver.Valuer interface for database serialization.
func (v OptTime) Value() (driver.Value, error) {
	if !v.IsDefined() {
		return nil, nil
	}

	if v.V.Location().String() == time.UTC.String() {
		return v.V.UTC(), nil
	}

	return v.V, nil
}

func (v OptTime) MarshalText() ([]byte, error) {
	if v.Defined {
		b, err := v.V.MarshalText()
		if err != nil {
			return b, fmt.Errorf("marshal text error: %w", err)
		}

		return b, nil
	}

	return []byte{}, nil
}

func (v *OptTime) UnmarshalText(data []byte) error {
	var err error

	if len(data) > 0 {
		v.Defined = false

		err = v.V.UnmarshalText(data)
		if err == nil {
			v.Defined = true
		}
	}

	v.Defined = false

	return nil
}

func (v OptTime) Equal(other OptTime) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V.Equal(other.V)
	}

	return false
}

func normalizePotentialUTC(timestamp time.Time) time.Time {
	if timestamp.Location().String() != time.UTC.String() {
		return timestamp
	}

	return timestamp.UTC()
}

func (v *OptTime) ScanTimestamptz(src pgtype.Timestamptz) error {
	if !src.Valid {
		*v = OptTime{}

		return nil
	}

	*v = NewTime(src.Time)

	return nil
}

func (v OptTime) TimestamptzValue() (pgtype.Timestamptz, error) {
	if !v.Defined {
		return pgtype.Timestamptz{Valid: false}, nil
	}

	return pgtype.Timestamptz{Time: v.V, Valid: true}, nil
}

func (v OptTime) LogValue() slog.Value {
	if v.Defined {
		return slog.TimeValue(v.V)
	}

	return slog.AnyValue(nil)
}
