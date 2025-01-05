//nolint:revive,godot
package easyjson

import (
	"github.com/0wnperception/go-helpers/pkg/easyjson/jlexer"
	"github.com/0wnperception/go-helpers/pkg/easyjson/jwriter"
)

// UnknownFieldsProxy implements UnknownsUnmarshaler and UnknownsMarshaler
// use it as embedded field in your structure to parse and then serialize unknown struct fields
type UnknownFieldsProxy struct {
	unknownFields map[string][]byte
}

func (s *UnknownFieldsProxy) UnmarshalUnknown(in *jlexer.Lexer, key string) {
	if s.unknownFields == nil {
		s.unknownFields = make(map[string][]byte, 1)
	}

	s.unknownFields[key] = in.Raw()
}

func (s UnknownFieldsProxy) MarshalUnknowns(out *jwriter.Writer, first bool) {
	for key, val := range s.unknownFields {
		if first {
			first = false
		} else {
			out.RawByte(',')
		}

		out.String(key)
		out.RawByte(':')
		out.Raw(val, nil)
	}
}
