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

type OptUInt struct {
	V       uint
	Defined bool
}

func NewUint(v uint) OptUInt {
	return OptUInt{V: v, Defined: true}
}

func (v *OptUInt) Undefine() {
	v.V, v.Defined = 0, false
}

func (v *OptUInt) SetValue(val uint) {
	v.V, v.Defined = val, true
}

func (v OptUInt) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Uint(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptUInt) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptUInt{}
	} else {
		v.V = l.Uint()
		v.Defined = true
	}
}

func (v *OptUInt) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUInt) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptUInt) IsDefined() bool {
	return v.Defined
}

func (v OptUInt) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatUint(uint64(v.V), base10)
}

func (v *OptUInt) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	s := asString(value)

	//nolint: mnd
	i64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, s, "uint", err)
	}

	v.V = uint(i64)

	return nil
}

func (v OptUInt) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil //nolint:gosec
}

func (v OptUInt) LogValue() slog.Value {
	if v.Defined {
		return slog.Uint64Value(uint64(v.V))
	}

	return slog.AnyValue(nil)
}
