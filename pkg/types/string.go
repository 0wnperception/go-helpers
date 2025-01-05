// //nolint: revive
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptString struct {
	V       string
	Defined bool
}

func NewString(v string) OptString {
	return OptString{V: v, Defined: true}
}

func (v *OptString) SetValue(val string) {
	v.V, v.Defined = val, true
}

func (v *OptString) Undefine() {
	v.V, v.Defined = "", false
}

func (v *OptString) Get() any {
	if v.Defined {
		return v.V
	}

	return nil
}

func (v OptString) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.String(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptString) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		*v = OptString{}
	} else {
		v.V = lexer.String()
		v.Defined = true
	}
}

func (v *OptString) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptString) UnmarshalJSON(data []byte) error {
	lexer := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&lexer)

	if err := lexer.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptString) IsDefined() bool {
	return v.Defined
}

func (v OptString) String() string {
	if !v.Defined {
		return undef
	}

	return v.V
}

func (v *OptString) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = "", false

		return nil
	}

	var ok bool

	v.V, ok = value.(string)
	if !ok {
		return fmt.Errorf("error scan OptString: %w", ErrConvert)
	}

	v.Defined = true

	return nil
}

func (v OptString) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return v.V, nil
}

func (v OptString) Equal(v2 OptString) bool {
	if !v.Defined && !v2.Defined {
		return true
	}

	if v.Defined && v2.Defined {
		return v.V == v2.V
	}

	return false
}

func (v *OptString) ScanText(src pgtype.Text) error {
	if !src.Valid {
		*v = OptString{}

		return nil
	}

	*v = NewString(src.String)

	return nil
}

func (v OptString) TextValue() (pgtype.Text, error) {
	if !v.Defined {
		return pgtype.Text{Valid: false}, nil
	}

	return pgtype.Text{String: v.V, Valid: true}, nil
}

func (v OptString) LogValue() slog.Value {
	if v.Defined {
		return slog.StringValue(v.V)
	}

	return slog.AnyValue(nil)
}

func (v OptString) MarshalText() (text []byte, err error) {
	if v.IsDefined() {
		return []byte(v.V), nil
	}

	return nil, nil
}

func (v *OptString) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		v.Undefine()
	} else {
		v.SetValue(string(text))
	}

	return nil
}
