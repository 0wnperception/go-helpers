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

type OptBool struct {
	V       bool
	Defined bool
}

func NewBool(v bool) OptBool {
	return OptBool{V: v, Defined: true}
}

func (v *OptBool) SetValue(val bool) {
	v.V, v.Defined = val, true
}

func (v *OptBool) Undefine() {
	v.V, v.Defined = false, false
}

func (v OptBool) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Bool(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptBool) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptBool{}
	} else {
		v.V = l.Bool()
		v.Defined = true
	}
}

func (v *OptBool) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptBool) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}

	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptBool) IsDefined() bool {
	return v.Defined
}

func (v OptBool) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatBool(v.V)
}

//  --- pgx methods

func (v *OptBool) Scan(src any) error {
	if src == nil {
		*v = OptBool{}

		return nil
	}

	switch src := src.(type) {
	case bool:
		*v = NewBool(src)

		return nil
	case string:
		b, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}

		*v = NewBool(b)

		return nil
	case []byte:
		b, err := strconv.ParseBool(string(src))
		if err != nil {
			return err
		}

		*v = NewBool(b)

		return nil
	}

	return fmt.Errorf("cannot scan %T - %w", src, ErrDecode)
}

func (v OptBool) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return v.V, nil
}

func (v *OptBool) ScanBool(src pgtype.Bool) error {
	if !src.Valid {
		*v = OptBool{}

		return nil
	}

	*v = NewBool(src.Bool)

	return nil
}

func (v OptBool) BoolValue() (pgtype.Bool, error) {
	return pgtype.Bool{Bool: v.V, Valid: v.Defined}, nil
}

func (v OptBool) LogValue() slog.Value {
	if v.Defined {
		return slog.BoolValue(v.V)
	}

	return slog.AnyValue(nil)
}
