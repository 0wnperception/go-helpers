// //nolint: revive
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptUInt64 struct {
	V       uint64
	Defined bool
}

func NewUint64(v uint64) OptUInt64 {
	return OptUInt64{V: v, Defined: true}
}

func (v OptUInt64) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Uint64(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptUInt64) Undefine() {
	v.V, v.Defined = 0, false
}

func (v *OptUInt64) SetValue(val uint64) {
	v.V, v.Defined = val, true
}

func (v *OptUInt64) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptUInt64{}
	} else {
		v.V = l.Uint64()
		v.Defined = true
	}
}

func (v *OptUInt64) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUInt64) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptUInt64) IsDefined() bool {
	return v.Defined
}

func (v OptUInt64) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatUint(v.V, base10)
}

func (v *OptUInt64) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	v.Defined = true
	strVal := asString(value)

	//nolint: mnd
	i64, err := strconv.ParseUint(strVal, 10, 64)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %strVal: %w",
			value, strVal, "uint64", err)
	}

	v.V = i64

	return nil
}

func (v OptUInt64) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil //nolint:gosec
}

func (v OptUInt64) LogValue() slog.Value {
	if v.Defined {
		return slog.Uint64Value(v.V)
	}

	return slog.AnyValue(nil)
}
