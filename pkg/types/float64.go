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

type OptFloat64 struct {
	V       float64
	Defined bool
}

func NewFloat64(v float64) OptFloat64 {
	return OptFloat64{V: v, Defined: true}
}

func (v *OptFloat64) SetValue(val float64) {
	v.V, v.Defined = val, true
}

func (v *OptFloat64) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptFloat64) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Float64(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptFloat64) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptFloat64{}
	} else {
		v.V = l.Float64()
		v.Defined = true
	}
}

func (v *OptFloat64) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptFloat64) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptFloat64) IsDefined() bool {
	return v.Defined
}

func (v OptFloat64) String() string {
	if !v.Defined {
		return undef
	}

	return fmt.Sprint(v.V)
}

//nolint:revive,mnd
func (v *OptFloat64) Scan(src any) error {
	if src == nil {
		*v = OptFloat64{}

		return nil
	}

	switch src := src.(type) {
	case float64:
		*v = NewFloat64(src)

		return nil
	case string:
		n, err := strconv.ParseFloat(src, 64)
		if err != nil {
			return fmt.Errorf("parse float64 error: %w", err)
		}

		*v = NewFloat64(n)

		return nil
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)

		n, err := strconv.ParseFloat(string(src), 64)
		if err != nil {
			return fmt.Errorf("parse float64 error: %w", err)
		}

		*v = NewFloat64(n)

		return nil
	}

	return fmt.Errorf("cannot scan %T: %w", src, ErrDecode)
}

func (v OptFloat64) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return v.V, nil
}

func (v OptFloat64) Equal(other OptFloat64) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptFloat64) ScanFloat64(src pgtype.Float8) error {
	if !src.Valid {
		*v = OptFloat64{}

		return nil
	}

	*v = NewFloat64(src.Float64)

	return nil
}

func (v OptFloat64) Float64Value() (pgtype.Float8, error) {
	if !v.Defined {
		return pgtype.Float8{Valid: false}, nil
	}

	return pgtype.Float8{Float64: v.V, Valid: true}, nil
}

func (v OptFloat64) LogValue() slog.Value {
	if v.Defined {
		return slog.Float64Value(v.V)
	}

	return slog.AnyValue(nil)
}
