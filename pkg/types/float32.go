// //nolint: revive
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

type OptFloat32 struct {
	V       float32
	Defined bool
}

func NewFloat32(v float32) OptFloat32 {
	return OptFloat32{V: v, Defined: true}
}

func (v *OptFloat32) SetValue(val float32) {
	v.V, v.Defined = val, true
}

func (v *OptFloat32) Undefine() {
	v.V, v.Defined = 0, false
}

func (v OptFloat32) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Float32(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptFloat32) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptFloat32{}
	} else {
		v.V = l.Float32()
		v.Defined = true
	}
}

func (v *OptFloat32) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptFloat32) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptFloat32) IsDefined() bool {
	return v.Defined
}

func (v OptFloat32) String() string {
	if !v.Defined {
		return undef
	}

	return fmt.Sprint(v.V)
}

func (v *OptFloat32) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true

	strVal := asString(value)
	//nolint: mnd
	f64, err := strconv.ParseFloat(strVal, 32)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, strVal, "float32", err)
	}

	v.V = float32(f64)

	return nil
}

func (v OptFloat32) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return v.V, nil
}

func (v OptFloat32) Equal(other OptFloat32) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V == other.V
	}

	return false
}

func (v *OptFloat32) ScanFloat64(src pgtype.Float8) error {
	if !src.Valid {
		*v = OptFloat32{}

		return nil
	}

	*v = NewFloat32(float32(src.Float64))

	return nil
}

func (v OptFloat32) Float64Value() (pgtype.Float8, error) {
	if !v.Defined {
		return pgtype.Float8{Valid: false}, nil
	}

	return pgtype.Float8{Float64: float64(v.V), Valid: true}, nil
}

func (v OptFloat32) LogValue() slog.Value {
	if v.Defined {
		return slog.Float64Value(float64(v.V))
	}

	return slog.AnyValue(nil)
}
