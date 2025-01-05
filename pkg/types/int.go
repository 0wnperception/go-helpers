// //nolint: revive
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

const (
	base10 = 10
)

type OptInt struct {
	V       int
	Defined bool
}

func NewInt(v int) OptInt {
	return OptInt{V: v, Defined: true}
}

func (v *OptInt) SetValue(val int) {
	v.V, v.Defined = val, true
}

func (v *OptInt) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptInt) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Int(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptInt) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptInt{}
	} else {
		v.V = l.Int()
		v.Defined = true
	}
}

func (v *OptInt) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptInt) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptInt) IsDefined() bool {
	return v.Defined
}

func (v OptInt) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.Itoa(v.V)
}

func (v *OptInt) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	strVal := asString(value)

	//nolint: mnd
	i64, err := strconv.Atoi(strVal)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, strVal, "int", err)
	}

	v.V = i64

	return nil
}

func (v OptInt) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptInt) Equal(other OptInt) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptInt) ScanInt64(src pgtype.Int8) error {
	if !src.Valid {
		*v = OptInt{}

		return nil
	}

	n := src.Int64
	if n < math.MinInt {
		return fmt.Errorf("%d is less than maximum value for int, %w", n, ErrDecode)
	}

	if n > math.MaxInt {
		return fmt.Errorf("%d is greater than maximum value for int, - %w", n, ErrDecode)
	}

	*v = NewInt(int(n))

	return nil
}

func (v OptInt) Int64Value() (pgtype.Int8, error) {
	return pgtype.Int8{Int64: int64(v.V), Valid: v.Defined}, nil
}

func (v OptInt) LogValue() slog.Value {
	if v.Defined {
		return slog.IntValue(v.V)
	}

	return slog.AnyValue(nil)
}
