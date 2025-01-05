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

type OptUInt32 struct {
	V       uint32
	Defined bool
}

func NewUint32(v uint32) OptUInt32 {
	return OptUInt32{V: v, Defined: true}
}

func (v *OptUInt32) Undefine() {
	v.V, v.Defined = 0, false
}

func (v *OptUInt32) SetValue(val uint32) {
	v.V, v.Defined = val, true
}

func (v OptUInt32) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Uint32(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptUInt32) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		*v = OptUInt32{}
	} else {
		v.V = lexer.Uint32()
		v.Defined = true
	}
}

func (v *OptUInt32) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUInt32) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptUInt32) IsDefined() bool {
	return v.Defined
}

func (v OptUInt32) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatUint(uint64(v.V), base10)
}

func (v *OptUInt32) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	s := asString(value)

	//nolint: mnd
	i32, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, s, "uint32", err)
	}

	v.V = uint32(i32) //nolint:gosec

	return nil
}

func (v OptUInt32) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptUInt32) LogValue() slog.Value {
	if v.Defined {
		return slog.Uint64Value(uint64(v.V))
	}

	return slog.AnyValue(nil)
}
