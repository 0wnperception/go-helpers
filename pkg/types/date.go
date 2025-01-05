package types

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

const (
	dateFormat = "2006-01-02"
)

type OptDate struct {
	V       time.Time
	Defined bool
}

func NewDate(v time.Time) OptDate {
	return OptDate{V: v, Defined: true}
}

func (v *OptDate) SetValue(val time.Time) {
	v.V, v.Defined = val, true
}

func (v *OptDate) Undefine() {
	v.V, v.Defined = time.Now(), false
}

func (v *OptDate) decodeText(src []byte) error {
	if src == nil {
		*v = OptDate{Defined: false}

		return nil
	}

	t, err := time.ParseInLocation(dateFormat, string(src), time.UTC)
	if err != nil {
		return err
	}

	*v = OptDate{V: t, Defined: true}

	return nil
}

func (v OptDate) IsDefined() bool {
	return v.Defined
}

func (v *OptDate) Scan(src any) error {
	if src == nil {
		*v = OptDate{Defined: false}

		return nil
	}

	switch src := src.(type) {
	case string:
		return v.decodeText([]byte(src))
	case []byte:
		srcCopy := make([]byte, len(src))
		copy(srcCopy, src)

		return v.decodeText(srcCopy)
	case time.Time:
		*v = OptDate{V: src, Defined: true}

		return nil
	}

	return fmt.Errorf("cannot scan %T, error: %w", src, ErrConvert)
}

func (v OptDate) Value() (driver.Value, error) {
	if v.Defined {
		return v.V, nil
	}

	return nil, nil
}

func (v *OptDate) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Skip()

		v.V, v.Defined = time.Time{}, false
	} else {
		str := lexer.String()

		t, err := time.Parse(dateFormat, str)
		if err != nil {
			lexer.AddError(err)
		}

		v.V = t
		v.Defined = true
	}
}

func (v OptDate) MarshalEasyJSON(writer *jwriter.Writer) {
	if v.Defined {
		const dateLen = 12

		b := make([]byte, 0, dateLen)
		b = append(b, '"')
		b = v.V.AppendFormat(b, dateFormat)
		b = append(b, '"')

		writer.Raw(b, nil)
	} else {
		writer.RawString("null")
	}
}

func (v *OptDate) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptDate) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (v OptDate) Equal(other OptDate) bool {
	if !v.Defined && !other.Defined {
		return true
	}

	if v.Defined && other.Defined {
		return v.V.Equal(other.V)
	}

	return false
}

func (v OptDate) String() string {
	if !v.Defined {
		return undef
	}

	return v.V.Format(dateFormat)
}

// --- pgx

func (v *OptDate) ScanDate(src pgtype.Date) error {
	if !src.Valid {
		*v = OptDate{}

		return nil
	}

	*v = NewDate(src.Time)

	return nil
}

func (v OptDate) DateValue() (pgtype.Date, error) {
	if !v.Defined {
		return pgtype.Date{Valid: false}, nil
	}

	return pgtype.Date{Time: v.V, Valid: true}, nil
}

func (v OptDate) LogValue() slog.Value {
	if v.Defined {
		return slog.StringValue(v.V.Format(dateFormat))
	}

	return slog.AnyValue(nil)
}
