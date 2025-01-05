// //nolint: revive
package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

type OptUInt8 struct {
	V       uint8
	Defined bool
}

func NewUint8(v uint8) OptUInt8 {
	return OptUInt8{V: v, Defined: true}
}

func (v *OptUInt8) Undefine() {
	v.V, v.Defined = 0, false
}

func (v *OptUInt8) SetValue(val uint8) {
	v.V, v.Defined = val, true
}

func (v OptUInt8) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.Uint8(v.V)
	} else {
		w.RawString("null")
	}
}

func (v *OptUInt8) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptUInt8{}
	} else {
		v.V = l.Uint8()
		v.Defined = true
	}
}

func (v *OptUInt8) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUInt8) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptUInt8) IsDefined() bool {
	return v.Defined
}

func (v OptUInt8) String() string {
	if !v.Defined {
		return undef
	}

	return strconv.FormatUint(uint64(v.V), base10)
}

func (v *OptUInt8) Scan(value any) error {
	if value == nil {
		v.V, v.Defined = 0, false

		return nil
	}

	var s string

	switch src := value.(type) {
	case int64:
		if src < 0 {
			return fmt.Errorf("%d is greater than maximum value for Uint8: %w", value, ErrConvert)
		}

		if src > math.MaxUint8 {
			return fmt.Errorf("%d is greater than maximum value for uint8: %w", value, ErrConvert)
		}

		v.V, v.Defined = uint8(src), true //nolint:gosec

		return nil
	case string:
		s = src
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)
		s = string(srcCopy)
	default:
		s = asString(src)
	}

	v.Defined = true

	//nolint: mnd
	ui8, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		err = strconvErr(err)

		return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %w",
			value, s, "uint8", err)
	}

	v.V = uint8(ui8) //nolint:gosec

	return nil
}

func (v OptUInt8) Value() (driver.Value, error) {
	if !v.Defined {
		return nil, nil
	}

	return int64(v.V), nil
}

func (v OptUInt8) LogValue() slog.Value {
	if v.Defined {
		return slog.Uint64Value(uint64(v.V))
	}

	return slog.AnyValue(nil)
}
