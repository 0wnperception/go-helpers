// //nolint:dupl,revive
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptUInt16 struct {
	V       uint16
	Defined bool
}

func NewUint16(v uint16) OptUInt16 {
	return OptUInt16{V: v, Defined: true}
}

func (v *OptUInt16) Undefine() {
	v.V, v.Defined = 0, false
}

func (v *OptUInt16) SetValue(val uint16) {
	v.V, v.Defined = val, true
}

func (v OptUInt16) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Uint16(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptUInt16) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptUInt16{}
	} else {
		v.V = l.Uint16()
		v.Defined = true
	}
}

func (v *OptUInt16) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUInt16) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptUInt16) IsDefined() bool {
	return v.Defined
}

func (v OptUInt16) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatUint(uint64(v.V), base10)
}

func (v *OptUInt16) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	strVal := asString(value)

	//nolint: mnd
	ui16, err := strconv.ParseUint(strVal, 10, 16)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %strVal: %w",
			value, strVal, "uint16", err)
	}

	v.V = uint16(ui16) //nolint:gosec

	return nil
}

func (v OptUInt16) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptUInt16) LogValue() slog.Value {
	if v.Defined {
		return slog.Uint64Value(uint64(v.V))
	}

	return slog.AnyValue(nil)
}
