package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptInt16 struct {
	V       int16
	Defined bool
}

func NewInt16(v int16) OptInt16 {
	return OptInt16{V: v, Defined: true}
}

func (v *OptInt16) SetValue(val int16) {
	v.V, v.Defined = val, true
}

func (v *OptInt16) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptInt16) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Int16(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptInt16) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptInt16{}
	} else {
		v.V = l.Int16()
		v.Defined = true
	}
}

func (v *OptInt16) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptInt16) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptInt16) IsDefined() bool {
	return v.Defined
}

func (v OptInt16) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatInt(int64(v.V), base10)
}

// //nolint: funlen,gocyclo,revive,mnd
func (v *OptInt16) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true

	switch val := value.(type) {
	case int64:
		if val < math.MinInt16 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		if val > math.MaxInt16 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		v.V = int16(val) //nolint:gosec

		return nil
	case int32:
		if val < math.MinInt16 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		if val > math.MaxInt16 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		v.V = int16(val) //nolint:gosec

		return nil
	case int16:
		v.V = val

		return nil
	case int8:
		v.V = int16(val) //nolint:gosec

		return nil
	case uint8:
		v.V = int16(val) //nolint:gosec

		return nil
	case uint16:
		v.V = int16(val) //nolint:gosec

		return nil
	}

	strVal := asString(value)

	//nolint: mnd
	i16, err := strconv.ParseInt(strVal, 10, 16)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %strVal: %w",
			value, strVal, "int16", err)
	}

	v.V = int16(i16) //nolint:gosec

	return nil
}

func (v OptInt16) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptInt16) Equal(other OptInt16) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptInt16) ScanInt64(src pgtype.Int8) error {
	if !src.Valid {
		*v = OptInt16{}

		return nil
	}

	n := src.Int64
	if n < math.MinInt16 {
		return fmt.Errorf("%d is less than maximum value for short - %w", n, ErrDecode)
	}

	if n > math.MaxInt16 {
		return fmt.Errorf("%d is greater than maximum value for short - %w", n, ErrDecode)
	}

	*v = NewInt16(int16(n)) //nolint:gosec

	return nil
}

func (v OptInt16) Int64Value() (pgtype.Int8, error) {
	return pgtype.Int8{Int64: int64(v.V), Valid: v.Defined}, nil
}

func (v OptInt16) LogValue() slog.Value {
	if v.Defined {
		return slog.Int64Value(int64(v.V))
	}

	return slog.AnyValue(nil)
}
