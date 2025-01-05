// Package easyjson contains marshaler/unmarshaler interfaces and helper functions.
// //nolint: revive,godot
package easyjson

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/0wnperception/go-helpers/pkg/checks"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

var ErrMarshal = errors.New("marshal easyjson error")

// Marshaler is an easyjson-compatible marshaler interface.
type Marshaler interface {
	MarshalEasyJSON(w *jwriter.Writer)
}

// Unmarshaler is an easyjson-compatible unmarshaler interface.
type Unmarshaler interface {
	UnmarshalEasyJSON(w *jlexer.Lexer)
}

// MarshalerUnmarshaler is an easyjson-compatible marshaler/unmarshaler interface.
type MarshalerUnmarshaler interface {
	Marshaler
	Unmarshaler
}

// Optional defines an undefined-test method for a type to integrate with 'omitempty' logic.
type Optional interface {
	IsDefined() bool
}

// UnknownsUnmarshaler provides a method to unmarshal unknown struct fileds and save them as you want.
type UnknownsUnmarshaler interface {
	UnmarshalUnknown(in *jlexer.Lexer, key string)
}

// UnknownsMarshaler provides a method to write additional struct fields.
type UnknownsMarshaler interface {
	MarshalUnknowns(w *jwriter.Writer, first bool)
}

func isNilInterface(i any) bool {
	return checks.IsNil(i)
}

// Marshal returns data as a single byte slice. Method is suboptimal as the data is likely to be copied
// from a chain of smaller chunks.
func Marshal(v Marshaler) ([]byte, error) {
	if isNilInterface(v) {
		return nullBytes, nil
	}

	w := jwriter.Writer{
		Flags:        jwriter.NilMapAsEmpty + jwriter.NilSliceAsEmpty,
		NoEscapeHTML: true,
	}

	v.MarshalEasyJSON(&w)

	return w.BuildBytes()
}

// MarshalToWriter marshals the data to an io.Writer.
func MarshalToWriter(v Marshaler, w io.Writer) (int64, error) {
	if isNilInterface(v) {
		bb, err := w.Write(nullBytes)

		return int64(bb), err
	}

	jw := jwriter.Writer{
		Flags:        jwriter.NilMapAsEmpty + jwriter.NilSliceAsEmpty,
		NoEscapeHTML: true,
	}

	v.MarshalEasyJSON(&jw)

	return jw.DumpTo(w)
}

func MarshalToResponseWriter(statusCode int, v Marshaler, w http.ResponseWriter) (int64, error) {
	if isNilInterface(v) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(nullBytes)))

		w.WriteHeader(statusCode)

		bb, e := w.Write(nullBytes)

		return int64(bb), e
	}

	jw := jwriter.Writer{
		Flags:        jwriter.NilMapAsEmpty + jwriter.NilSliceAsEmpty,
		NoEscapeHTML: true,
	}

	v.MarshalEasyJSON(&jw)

	if jw.Error != nil {
		//nolint:errorlint
		return 0, fmt.Errorf("%s: %w", jw.Error.Error(), ErrMarshal)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(jw.Size()))

	w.WriteHeader(statusCode)

	written, err := jw.DumpTo(w)

	return written, err
}

// MarshalToHTTPResponseWriter sets Content-Length and Content-Type headers for the
// http.ResponseWriter, and send the data to the writer. started will be equal to
// false if an error occurred before any http.ResponseWriter methods were actually
// invoked (in this case a 500 reply is possible).
//
// Deprecated: use MarshalToResponseWriter
// //nolint: nonamedreturns
func MarshalToHTTPResponseWriter(v Marshaler, w http.ResponseWriter) (started bool, written int64, err error) {
	if isNilInterface(v) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(nullBytes)))

		bb, e := w.Write(nullBytes)

		return true, int64(bb), e
	}

	jw := jwriter.Writer{
		Flags:        jwriter.NilMapAsEmpty + jwriter.NilSliceAsEmpty,
		NoEscapeHTML: true,
	}

	v.MarshalEasyJSON(&jw)

	if jw.Error != nil {
		return false, 0, jw.Error
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(jw.Size()))

	written, err = jw.DumpTo(w)

	return true, written, err
}

// Unmarshal decodes the JSON in data into the object.
func Unmarshal(data []byte, v Unmarshaler) error {
	l := jlexer.Lexer{Data: data}

	v.UnmarshalEasyJSON(&l)

	return l.Error()
}

// UnmarshalFromReader reads all the data in the reader and decodes as JSON into the object.
func UnmarshalFromReader(r io.Reader, v Unmarshaler) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return Unmarshal(data, v)
}
