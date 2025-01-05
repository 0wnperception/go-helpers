// //nolint:gofumpt,revive,mnd
package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

const (
	VariantNCS byte = iota
	VariantRFC4122
	VariantMicrosoft
	VariantFuture
)

type UUID [16]byte

var (
	//nolint:gochecknoglobals
	ZeroUUID = UUID{}

	ErrInvalidUUIDFormat = errors.New("invalid UUID format")
	ErrCanNotScanUUID    = errors.New("cannot scan NULL into *id.UUID")

	//nolint:gochecknoglobals
	xvalues = [256]byte{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255, 255, 255, 255, 255, 255,
		255, 10, 11, 12, 13, 14, 15, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 10, 11, 12, 13, 14, 15, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	}
)

func NewUUID(b []byte) (UUID, error) {
	if len(b) != 16 {
		return UUID{}, ErrInvalidUUIDFormat
	}

	r := UUID{}

	copy(r[:], b)

	return r, nil
}

func NewUUIDFromString(s string) (UUID, error) {
	var uuid UUID

	switch len(s) {
	// like xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36:

	// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36 + 9:
		if strings.ToLower(s[:9]) != "urn:uuid:" {
			return uuid, ErrInvalidUUIDFormat
		}

		s = s[9:]

	// like {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
	case 36 + 2:
		s = s[1:]

	// like xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	case 32:
		var ok bool

		for i := range uuid {
			uuid[i], ok = xtob(s[i*2], s[i*2+1])
			if !ok {
				return uuid, ErrInvalidUUIDFormat
			}
		}

		return uuid, nil
	default:
		return uuid, ErrInvalidUUIDFormat
	}
	// s is now at least 36 bytes long
	// it must be of the form  xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return uuid, ErrInvalidUUIDFormat
	}

	for i, x := range [16]int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34} {
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return uuid, ErrInvalidUUIDFormat
		}

		uuid[i] = v
	}

	return uuid, nil
}

func (u UUID) String() string {
	var buf [36]byte

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf[:])
}

func (u UUID) HexString() string {
	var buf [32]byte

	hex.Encode(buf[:], u[:])

	return string(buf[:])
}

func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

func (u *UUID) SetVariant(v byte) {
	// //nolint:gocritic
	switch v {
	case VariantNCS:
		u[8] = u[8]&(0xff>>1) | (0x00 << 7)
	case VariantRFC4122:
		u[8] = u[8]&(0xff>>2) | (0x02 << 6)
	case VariantMicrosoft:
		u[8] = u[8]&(0xff>>3) | (0x06 << 5)
	case VariantFuture:
		fallthrough
	default:
		u[8] = u[8]&(0xff>>3) | (0x07 << 5)
	}
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return u[:], nil
}

func (u *UUID) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return ErrInvalidUUIDFormat
	}

	copy(u[:], data)

	return nil
}

func (u UUID) MarshalEasyJSON(w *jwriter.Writer) {
	w.Buffer.AppendByte('"')
	w.HexBytes(u[0:4])
	w.Buffer.AppendByte('-')
	w.HexBytes(u[4:6])
	w.Buffer.AppendByte('-')
	w.HexBytes(u[6:8])
	w.Buffer.AppendByte('-')
	w.HexBytes(u[8:10])
	w.Buffer.AppendByte('-')
	w.HexBytes(u[10:])
	w.Buffer.AppendByte('"')
}

func (u *UUID) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	u.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	u.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (u *UUID) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.AddError(ErrInvalidUUIDFormat)

		*u = ZeroUUID
	} else {
		var err error

		*u, err = NewUUIDFromString(l.String())

		if err != nil {
			l.AddError(err)
		}
	}
}

func (u UUID) LogValue() slog.Value {
	return slog.StringValue(u.String())
}

type OptUUID struct {
	V       UUID
	Defined bool
}

func (v *OptUUID) Undefine() {
	v.V, v.Defined = ZeroUUID, false
}

func (v *OptUUID) SetValue(val UUID) {
	v.V, v.Defined = val, true
}

func (v OptUUID) IsDefined() bool {
	return v.Defined
}

func (v OptUUID) String() string {
	if !v.Defined {
		return undef
	}

	return v.V.String()
}

func (v OptUUID) MarshalEasyJSON(w *jwriter.Writer) {
	if v.Defined {
		w.String(v.V.String())
	} else {
		w.RawString("null")
	}
}

func (v *OptUUID) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()

		*v = OptUUID{}
	} else {
		var err error

		v.V, err = NewUUIDFromString(l.String())
		if err != nil {
			l.AddError(err)
		} else {
			v.Defined = true
		}
	}
}

func (v *OptUUID) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)

	return w.Buffer.BuildBytes(), w.Error
}

func (v *OptUUID) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)

	if err := l.Error(); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func xtob(x1, x2 byte) (byte, bool) {
	b1 := xvalues[x1]
	b2 := xvalues[x2]

	return (b1 << 4) | b2, b1 != 255 && b2 != 255
}

func (u *UUID) ScanUUID(v pgtype.UUID) error {
	if !v.Valid {
		return ErrCanNotScanUUID
	}

	*u = v.Bytes

	return nil
}

func (u UUID) UUIDValue() (pgtype.UUID, error) {
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func (v *OptUUID) ScanUUID(val pgtype.UUID) error {
	*v = OptUUID{V: val.Bytes, Defined: val.Valid}

	return nil
}

func (v OptUUID) UUIDValue() (pgtype.UUID, error) {
	return pgtype.UUID{Bytes: v.V, Valid: v.Defined}, nil
}

func (v OptUUID) LogValue() slog.Value {
	if v.Defined {
		return slog.StringValue(v.V.String())
	}

	return slog.AnyValue(nil)
}
