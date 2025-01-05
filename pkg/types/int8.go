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

type OptInt8 struct {
	V       int8
	Defined bool
}

func NewInt8(v int8) OptInt8 {
	return OptInt8{V: v, Defined: true}
}

func (v *OptInt8) SetValue(val int8) {
	v.V, v.Defined = val, true
}

func (v *OptInt8) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptInt8) MarshalEasyJSON(writer *jwriter.Writer) {
	if v.Defined {
		writer.Int8(v.V)
	} else {
		writer.RawString("null")
	}
}

func (v *OptInt8) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		*v = OptInt8{}
	} else {
		v.V, v.Defined = lexer.Int8(), true
	}
}

func (v *OptInt8) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptInt8) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptInt8) IsDefined() bool {
	return v.Defined
}

func (v OptInt8) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatInt(int64(v.V), base10)
}

// //nolint: revive,mnd
func (v *OptInt8) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	strVal := asString(value)

	parseInt, err := strconv.ParseInt(strVal, 10, 8)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, strVal, "int8", err)
	}

	v.V = int8(parseInt) //nolint:gosec

	return nil
}

func (v OptInt8) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptInt8) Equal(other OptInt8) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptInt8) ScanInt64(src pgtype.Int8) error {
	if !src.Valid {
		*v = OptInt8{}

		return nil
	}

	n := src.Int64
	if n < math.MinInt8 {
		return fmt.Errorf("%d is less than maximum value for byte - %w", n, ErrDecode)
	}

	if n > math.MaxInt8 {
		return fmt.Errorf("%d is greater than maximum value for byte - %w", n, ErrDecode)
	}

	*v = NewInt8(int8(n)) //nolint:gosec

	return nil
}

func (v OptInt8) Int64Value() (pgtype.Int8, error) {
	return pgtype.Int8{Int64: int64(v.V), Valid: v.Defined}, nil
}

func (v OptInt8) LogValue() slog.Value {
	if v.Defined {
		return slog.Int64Value(int64(v.V))
	}

	return slog.AnyValue(nil)
}
