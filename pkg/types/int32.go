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

type OptInt32 struct {
	V       int32
	Defined bool
}

func NewInt32(v int32) OptInt32 {
	return OptInt32{V: v, Defined: true}
}

func (v *OptInt32) SetValue(val int32) {
	v.V, v.Defined = val, true
}

func (v *OptInt32) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptInt32) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Int32(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptInt32) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		v.V, v.Defined = 0, false
	} else {
		v.V = l.Int32()
		v.Defined = true
	}
}

func (v *OptInt32) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptInt32) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptInt32) IsDefined() bool {
	return v.Defined
}

func (v OptInt32) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatInt(int64(v.V), base10)
}

// //nolint: gocyclo, revive, mnd
func (v *OptInt32) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true

	switch val := value.(type) {
	case int64:
		if val < math.MinInt32 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		if val > math.MaxInt32 {
			return fmt.Errorf("%d is greater than maximum value for Int32: %w", value, ErrConvert)
		}

		v.V = int32(val) //nolint:gosec

		return nil
	case int32:
		v.V = val

		return nil
	case int16:
		v.V = int32(val)

		return nil
	case int8:
		v.V = int32(val) //nolint:gosec

		return nil
	case uint8:
		v.V = int32(val)

		return nil
	case uint16:
		v.V = int32(val)

		return nil
	}

	strVal := asString(value)

	//nolint: mnd
	i32, err := strconv.ParseInt(strVal, 10, 32)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, strVal, "int32", err)
	}

	v.V = int32(i32) //nolint:gosec

	return nil
}

func (v OptInt32) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptInt32) Equal(v2 OptInt32) bool {
	if !v.Defined && !v2.Defined {
		return true
	}

	if v.Defined && v2.Defined {
		return v.V == v2.V
	}

	return false
}

func (v *OptInt32) ScanInt64(src pgtype.Int8) error {
	if !src.Valid {
		*v = OptInt32{}

		return nil
	}

	n := src.Int64
	if n < math.MinInt32 {
		return fmt.Errorf("%d is less than maximum value for int32 - %w", n, ErrDecode)
	}

	if n > math.MaxInt32 {
		return fmt.Errorf("%d is greater than maximum value for int32 - %w", n, ErrDecode)
	}

	*v = NewInt32(int32(n)) //nolint:gosec

	return nil
}

func (v OptInt32) Int64Value() (pgtype.Int8, error) {
	return pgtype.Int8{Int64: int64(v.V), Valid: v.Defined}, nil
}

func (v OptInt32) LogValue() slog.Value {
	if v.Defined {
		return slog.Int64Value(int64(v.V))
	}

	return slog.AnyValue(nil)
}
