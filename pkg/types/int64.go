// //nolint: revive,gocognit
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptInt64 struct {
	V       int64
	Defined bool
}

func NewInt64(v int64) OptInt64 {
	return OptInt64{V: v, Defined: true}
}

func (v *OptInt64) SetValue(val int64) {
	v.V, v.Defined = val, true
}

func (v *OptInt64) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptInt64) MarshalEasyJSON(writer *jwriter.Writer) {
	if v.Defined {
		writer.Int64(v.V)
	} else {
		writer.RawString("null")
	}
}

func (v *OptInt64) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		v.V, v.Defined = 0, false
	} else {
		v.V = lexer.Int64()
		v.Defined = true
	}
}

func (v *OptInt64) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptInt64) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptInt64) IsDefined() bool {
	return v.Defined
}

func (v OptInt64) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatInt(v.V, base10)
}

// //nolint: revive,mnd
func (v *OptInt64) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	switch src := value.(type) {
	case int64:
		v.V, v.Defined = src, true

		return nil
	case string:
		n, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return fmt.Errorf("decode int64 error: %w", err)
		}

		v.V, v.Defined = n, true

		return nil
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)

		n, err := strconv.ParseInt(string(src), 10, 64)
		if err != nil {
			return fmt.Errorf("decode int64 error: %w", err)
		}

		v.V, v.Defined = n, true

		return nil
	}

	return fmt.Errorf("cannot scan %T: %w", value, ErrDecode)
}

func (v OptInt64) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return v.V, nil
}

func (v OptInt64) Equal(other OptInt64) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptInt64) ScanInt64(src pgtype.Int8) error {
	if !src.Valid {
		*v = OptInt64{}

		return nil
	}

	*v = NewInt64(src.Int64)

	return nil
}

func (v OptInt64) Int64Value() (pgtype.Int8, error) {
	return pgtype.Int8{Int64: v.V, Valid: v.Defined}, nil
}

func (v OptInt64) LogValue() slog.Value {
	if v.Defined {
		return slog.Int64Value(v.V)
	}

	return slog.AnyValue(nil)
}
