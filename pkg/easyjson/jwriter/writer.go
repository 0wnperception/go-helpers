// Package jwriter contains a JSON writer.
// //nolint:funlen,lll,gocognit,gocyclo,gofumpt,revive,unconvert,mnd,gocritic,gochecknoglobals,godot,wsl,nonamedreturns
package jwriter

import (
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/0wnperception/go-helpers/pkg/buffer"
)

// Flags describe various encoding options. The behavior may be actually implemented in the encoder, but
// Flags field in Writer is used to set and pass them around.
type Flags int

const (
	NilMapAsEmpty   Flags = 1 << iota // Encode nil map as '{}' rather than 'null'.
	NilSliceAsEmpty                   // Encode nil slice as '[]' rather than 'null'.

	quote = '"'

	chars = "0123456789abcdef"
)

func New() Writer {
	return Writer{
		Flags:        NilMapAsEmpty + NilSliceAsEmpty,
		NoEscapeHTML: true,
	}
}

// Writer is a JSON writer.
type Writer struct {
	Error        error
	Buffer       buffer.Buffer
	NoEscapeHTML bool
	Flags        Flags
}

// Size returns the size of the data that was written out.
func (w *Writer) Size() int {
	return w.Buffer.Size()
}

// DumpTo outputs the data to given io.Writer, resetting the buffer.
func (w *Writer) DumpTo(out io.Writer) (written int64, err error) {
	return w.Buffer.WriteTo(out)
}

// BuildBytes returns writer data as a single byte slice. You can optionally provide one byte slice
// as argument that it will try to reuse.
func (w *Writer) BuildBytes(reuse ...[]byte) ([]byte, error) {
	if w.Error != nil {
		return nil, w.Error
	}

	return w.Buffer.BuildBytes(reuse...), nil
}

// ReadCloser returns an io.ReadCloser that can be used to read the data.
// ReadCloser also resets the buffer.
func (w *Writer) ReadCloser() (io.ReadCloser, error) {
	if w.Error != nil {
		return nil, w.Error
	}

	return w.Buffer.ReadCloser(), nil
}

// RawByte appends raw binary data to the buffer.
func (w *Writer) RawByte(c byte) {
	w.Buffer.AppendByte(c)
}

// RawString appends raw string data to the buffer.
func (w *Writer) RawString(s string) {
	w.Buffer.AppendString(s)
}

// Raw appends raw binary data to the buffer or sets the error if it is given. Useful for
// calling with results of MarshalJSON-like functions.
func (w *Writer) Raw(data []byte, err error) {
	switch {
	case w.Error != nil:
		return
	case err != nil:
		w.Error = err
	case len(data) > 0:
		w.Buffer.AppendBytes(data)
	default:
		w.RawString("null")
	}
}

// RawText encloses raw binary data in quotes and appends in to the buffer.
// Useful for calling with results of MarshalText-like functions.
func (w *Writer) RawText(data []byte, err error) {
	switch {
	case w.Error != nil:
		return
	case err != nil:
		w.Error = err
	case len(data) > 0:
		w.String(string(data))
	default:
		w.RawString("null")
	}
}

// Base64Bytes appends data to the buffer after base64 encoding it
func (w *Writer) Base64Bytes(data []byte) {
	if data == nil {
		w.Buffer.AppendString("null")

		return
	}

	w.Buffer.AppendByte(quote)
	w.base64(data)
	w.Buffer.AppendByte(quote)
}

func (w *Writer) Uint8(n uint8) {
	w.Buffer.EnsureSpace(3)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
}

func (w *Writer) Uint16(n uint16) {
	w.Buffer.EnsureSpace(5)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
}

func (w *Writer) Uint32(n uint32) {
	w.Buffer.EnsureSpace(10)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
}

func (w *Writer) Uint(n uint) {
	w.Buffer.EnsureSpace(20)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
}

func (w *Writer) Uint64(n uint64) {
	w.Buffer.EnsureSpace(20)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, n)
}

func (w *Writer) Int8(n int8) {
	w.Buffer.EnsureSpace(4)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
}

func (w *Writer) Int16(n int16) {
	w.Buffer.EnsureSpace(6)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
}

func (w *Writer) Int32(n int32) {
	w.Buffer.EnsureSpace(11)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
}

func (w *Writer) Int(n int) {
	w.Buffer.EnsureSpace(21)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
}

func (w *Writer) Int64(n int64) {
	w.Buffer.EnsureSpace(21)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
}

func (w *Writer) Uint8Str(n uint8) {
	w.Buffer.EnsureSpace(5)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Uint16Str(n uint16) {
	w.Buffer.EnsureSpace(7)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Uint32Str(n uint32) {
	w.Buffer.EnsureSpace(12)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) UintStr(n uint) {
	w.Buffer.EnsureSpace(22)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, uint64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Uint64Str(n uint64) {
	w.Buffer.EnsureSpace(22)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendUint(w.Buffer.Buf, n)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) UintptrStr(n uintptr) {
	w.Buffer.EnsureSpace(22)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = strconv.AppendUint(w.Buffer.Buf, uint64(n), 10)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Int8Str(n int8) {
	w.Buffer.EnsureSpace(6)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Int16Str(n int16) {
	w.Buffer.EnsureSpace(8)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Int32Str(n int32) {
	w.Buffer.EnsureSpace(13)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) IntStr(n int) {
	w.Buffer.EnsureSpace(23)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, int64(n))
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Int64Str(n int64) {
	w.Buffer.EnsureSpace(23)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = appendInt(w.Buffer.Buf, n)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Float32(n float32) {
	w.Buffer.EnsureSpace(20)
	w.Buffer.Buf = strconv.AppendFloat(w.Buffer.Buf, float64(n), 'g', -1, 32)
}

func (w *Writer) Float32Str(n float32) {
	w.Buffer.EnsureSpace(23)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = strconv.AppendFloat(w.Buffer.Buf, float64(n), 'g', -1, 32)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Float64(n float64) {
	w.Buffer.EnsureSpace(20)
	w.Buffer.Buf = strconv.AppendFloat(w.Buffer.Buf, n, 'g', -1, 64)
}

func (w *Writer) Float64Str(n float64) {
	w.Buffer.EnsureSpace(23)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = strconv.AppendFloat(w.Buffer.Buf, float64(n), 'g', -1, 64)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Time(t time.Time) {
	w.Buffer.EnsureSpace(len(time.RFC3339Nano) + len(`""`))

	w.Buffer.Buf = append(w.Buffer.Buf, quote)
	w.Buffer.Buf = t.UTC().AppendFormat(w.Buffer.Buf, time.RFC3339Nano)
	w.Buffer.Buf = append(w.Buffer.Buf, quote)
}

func (w *Writer) Bool(v bool) {
	w.Buffer.EnsureSpace(7)
	if v {
		w.Buffer.Buf = append(w.Buffer.Buf, "true"...)
	} else {
		w.Buffer.Buf = append(w.Buffer.Buf, "false"...)
	}
}

func getTable(falseValues ...int) [128]bool {
	table := [128]bool{}

	for i := range 128 {
		table[i] = true
	}

	for _, v := range falseValues {
		table[v] = false
	}

	return table
}

var (
	htmlEscapeTable   = getTable(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, '"', '&', '<', '>', '\\')
	htmlNoEscapeTable = getTable(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, '"', '\\')
)

func (w *Writer) StringNoEscape(s string) {
	l := len(s)
	if l == 0 {
		w.Buffer.AppendTwoBytes(quote, quote)

		return
	}

	w.Buffer.AppendByte(quote)
	w.Buffer.AppendString(s)
	w.Buffer.AppendByte(quote)
}

func (w *Writer) String(s string) {
	l := len(s)
	if l == 0 {
		w.Buffer.AppendTwoBytes(quote, quote)

		return
	}

	w.Buffer.EnsureSpace(l + 2)

	w.Buffer.AppendByte(quote)

	i := 0
	j := 0

	if l >= 8 {
		if j = escapeIndex(s, !w.NoEscapeHTML); j < 0 {
			w.Buffer.AppendString(s)
			w.Buffer.AppendByte(quote)

			return
		}
	}

	escapeTable := &htmlEscapeTable
	if w.NoEscapeHTML {
		escapeTable = &htmlNoEscapeTable
	}

	for j < len(s) && j >= 0 {
		c := s[j]

		if c < utf8.RuneSelf {
			if escapeTable[c] {
				// single-width character, no escaping is required
				j++

				continue
			}

			w.Buffer.AppendString(s[i:j])

			switch c {
			case '\\', '"':
				w.Buffer.AppendTwoBytes('\\', c)
			case '\n':
				w.Buffer.AppendTwoBytes('\\', 'n')
			case '\b':
				w.Buffer.AppendTwoBytes('\\', 'b')
			case '\f':
				w.Buffer.AppendTwoBytes('\\', 'f')
			case '\r':
				w.Buffer.AppendTwoBytes('\\', 'r')
			case '\t':
				w.Buffer.AppendTwoBytes('\\', 't')
			case '<', '>', '&':
				w.Buffer.AppendString(`\u00`)
				w.Buffer.AppendTwoBytes(chars[c>>4], chars[c&0xF])
			default:
				w.Buffer.AppendString(`\u00`)
				w.Buffer.AppendTwoBytes(chars[c>>4], chars[c&0xF])
			}

			i = j + 1
			j = j + 1
		}

		r, size := utf8.DecodeRuneInString(s[j:])
		if r == utf8.RuneError && size == 1 {
			w.Buffer.AppendString(s[i:j])
			w.Buffer.AppendString(`\ufffd`)
			i = j + size
			j = j + size

			continue
		}

		switch r {
		case '\u2028', '\u2029':
			w.Buffer.AppendString(s[i:j])
			w.Buffer.AppendString(`\u202`)
			w.Buffer.AppendByte(chars[r&0xF])

			i = j + size
			j = j + size

			continue
		}

		j += size
	}

	w.Buffer.AppendString(s[i:])
	w.Buffer.AppendByte(quote)
}

const (
	encode   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	padChar  = '='
	hextable = "0123456789abcdef"
)

func (w *Writer) HexBytes(b []byte) {
	if len(b) > 0 {
		w.Buffer.EnsureSpace(2 * len(b))
		for _, v := range b {
			w.Buffer.Buf = append(w.Buffer.Buf, hextable[v>>4])
			w.Buffer.Buf = append(w.Buffer.Buf, hextable[v&0x0f])
		}
	}
}

func (w *Writer) base64(in []byte) {
	if len(in) == 0 {
		return
	}

	w.Buffer.EnsureSpace(((len(in)-1)/3 + 1) * 4)

	si := 0
	n := (len(in) / 3) * 3

	for si < n {
		// Convert 3x 8bit source bytes into 4 bytes
		val := uint(in[si+0])<<16 | uint(in[si+1])<<8 | uint(in[si+2])

		w.Buffer.Buf = append(w.Buffer.Buf, encode[val>>18&0x3F], encode[val>>12&0x3F], encode[val>>6&0x3F], encode[val&0x3F])

		si += 3
	}

	remain := len(in) - si
	if remain == 0 {
		return
	}

	// Add the remaining small block
	val := uint(in[si+0]) << 16
	if remain == 2 {
		val |= uint(in[si+1]) << 8
	}

	w.Buffer.Buf = append(w.Buffer.Buf, encode[val>>18&0x3F], encode[val>>12&0x3F])

	switch remain {
	case 2:
		w.Buffer.Buf = append(w.Buffer.Buf, encode[val>>6&0x3F], byte(padChar))
	case 1:
		w.Buffer.Buf = append(w.Buffer.Buf, byte(padChar), byte(padChar))
	}
}
