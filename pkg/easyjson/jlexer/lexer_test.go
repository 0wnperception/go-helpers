package jlexer

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestU4(t *testing.T) {
	toParse := `"\ud83d\ude00"`

	want := "😀"

	l := Lexer{Data: []byte(toParse)}
	got := l.String()

	if got != want {
		t.Errorf("[%q] String() = %v; want %v", toParse, got, want)
	}
}
func TestString(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: `"simple string"`, want: "simple string"},
		{toParse: `"Жопа"`, want: "Жопа"},
		{toParse: " \r\r\n\t  " + `"test"`, want: "test"},
		{toParse: `"\n\t\"\/\\\f\r"`, want: "\n\t\"/\\\f\r"},
		{toParse: `"\u0020"`, want: " "},
		{toParse: `"\u0020-\t"`, want: " -\t"},
		{toParse: `"\ufffd\uFFFD"`, want: "\ufffd\ufffd"},
		{toParse: `"\ud83d\ude00"`, want: "😀"},
		{toParse: `"\ud83d\ude08"`, want: "😈"},
		{toParse: `"\ud8"`, wantError: true},

		{toParse: `"test"junk`, want: "test"},

		{toParse: `5`, wantError: true},    // not a string
		{toParse: `"\x"`, wantError: true}, // invalid escape
		{toParse: `"\ud800"`, want: "�"},   // invalid utf-8 char; return replacement char
	} {
		{
			l := Lexer{Data: []byte(test.toParse)}

			got := l.String()
			if got != test.want {
				t.Errorf("[%d, %q] String() = %v; want %v", i, test.toParse, got, test.want)
			}
			err := l.Error()
			if err != nil && !test.wantError {
				t.Errorf("[%d, %q] String() error: %v", i, test.toParse, err)
			} else if err == nil && test.wantError {
				t.Errorf("[%d, %q] String() ok; want error", i, test.toParse)
			}
		}
		{
			l := Lexer{Data: []byte(test.toParse)}

			got := l.StringIntern()
			if got != test.want {
				t.Errorf("[%d, %q] String() = %v; want %v", i, test.toParse, got, test.want)
			}
			err := l.Error()
			if err != nil && !test.wantError {
				t.Errorf("[%d, %q] String() error: %v", i, test.toParse, err)
			} else if err == nil && test.wantError {
				t.Errorf("[%d, %q] String() ok; want error", i, test.toParse)
			}
		}
	}
}

func TestStringIntern(t *testing.T) {
	data := []byte(`"string interning test"`)
	var l Lexer

	allocsPerRun := testing.AllocsPerRun(1000, func() {
		l = Lexer{Data: data}
		_ = l.StringIntern()
	})
	if allocsPerRun != 0 {
		t.Fatalf("expected 0 allocs, got %f", allocsPerRun)
	}

	allocsPerRun = testing.AllocsPerRun(1000, func() {
		l = Lexer{Data: data}
		_ = l.String()
	})
	if allocsPerRun != 1 {
		t.Fatalf("expected 1 allocs, got %f", allocsPerRun)
	}
}

func TestBytes(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: `"c2ltcGxlIHN0cmluZw=="`, want: "simple string"},
		{toParse: " \r\r\n\t  " + `"dGVzdA=="`, want: "test"},
		{toParse: `"c3ViamVjdHM\/X2Q9MQ=="`, want: "subjects?_d=1"}, // base64 with forward slash escaped

		{toParse: `5`, wantError: true},                     // not a JSON string
		{toParse: `"foobar"`, wantError: true},              // not base64 encoded
		{toParse: `"c2ltcGxlIHN0cmluZw="`, wantError: true}, // invalid base64 padding
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Bytes()
		if bytes.Compare(got, []byte(test.want)) != 0 {
			t.Errorf("[%d, %q] Bytes() = %v; want: %v", i, test.toParse, got, []byte(test.want))
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Bytes() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Bytes() ok; want error", i, test.toParse)
		}
	}
}

func TestInt8(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      int8
		wantError bool
	}{
		{toParse: "123", want: 123, wantError: false},
		{toParse: "\n\r123", want: 123, wantError: false},
		{toParse: "-123", want: -123, wantError: false},
		{toParse: "0", want: 0, wantError: false},
		{toParse: "1234", want: 0, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Int8()
		if got != test.want {
			t.Errorf("[%d, %q] number() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] number() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] number() ok; want error", i, test.toParse)
		}
	}
}

func TestInt8Str(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      int8
		wantError bool
	}{
		{toParse: `"123"`, want: 123, wantError: false},
		{toParse: `"-123"`, want: -123, wantError: false},
		{toParse: `"0"`, want: 0, wantError: false},
		{toParse: `"1234"`, want: 0, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Int8Str()
		if got != test.want {
			t.Errorf("[%d, %q] number() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] number() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] number() ok; want error", i, test.toParse)
		}
	}
}

func TestInt16(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      int16
		wantError bool
	}{
		{toParse: "123", want: 123, wantError: false},
		{toParse: "\n-123", want: -123, wantError: false},
		{toParse: "0", want: 0, wantError: false},
		{toParse: "123400", want: 0, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Int16()
		if got != test.want {
			t.Errorf("[%d, %q] number() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] number() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] number() ok; want error", i, test.toParse)
		}
	}
}

func TestInt32(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      int32
		wantError bool
	}{
		{toParse: "123", want: 123, wantError: false},
		{toParse: "\n-123", want: -123, wantError: false},
		{toParse: "0", want: 0, wantError: false},
		{toParse: "12340092738019734103947", want: 0, wantError: true},
		{toParse: "123B40", want: 0, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Int32()
		if got != test.want {
			t.Errorf("[%d, %q] number() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] number() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] number() ok; want error", i, test.toParse)
		}
	}
}

func TestNumber(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: "123", want: "123"},
		{toParse: "-123", want: "-123"},
		{toParse: "\r\n12.35", want: "12.35"},
		{toParse: "12.35e+1", want: "12.35e+1"},
		{toParse: "12.35e-15", want: "12.35e-15"},
		{toParse: "12.35E-15", want: "12.35E-15"},
		{toParse: "12.35E15", want: "12.35E15"},

		{toParse: `"a"`, wantError: true},
		{toParse: "123junk", wantError: true},
		{toParse: "1.2.3", wantError: true},
		{toParse: "1e2e3", wantError: true},
		{toParse: "1e2.3", wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.number()
		if got != test.want {
			t.Errorf("[%d, %q] number() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] number() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] number() ok; want error", i, test.toParse)
		}
	}
}

func TestNull(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		wantError bool
	}{
		{toParse: "null", wantError: false},
		{toParse: "nul", wantError: true},
		{toParse: "nullaaa", wantError: true},
		{toParse: "nlul", wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		l.Null()

		err := l.Error()
		if test.wantError {
			require.Error(t, err, "[%d, %q] null error", i, test.toParse)
		} else {
			require.NoError(t, err, "[%d, %q] null error", i, test.toParse)
		}
	}
}

func TestBool(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      bool
		wantError bool
	}{
		{toParse: "true", want: true},
		{toParse: "false", want: false},

		{toParse: "1", wantError: true},
		{toParse: "tr", wantError: true},
		{toParse: "tre", wantError: true},
		{toParse: "fa", wantError: true},
		{toParse: "fales", wantError: true},
		{toParse: "treu", wantError: true},
		{toParse: "truejunk", wantError: true},
		{toParse: "falsejunk", wantError: true},
		{toParse: `false"junk"`, wantError: true},
		{toParse: `true"junk"`, wantError: true},
		{toParse: "True", wantError: true},
		{toParse: "False", wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Bool()
		if got != test.want {
			t.Errorf("[%d, %q] Bool() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Bool() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Bool() ok; want error", i, test.toParse)
		}
	}
}

func TestSkipRecursive(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		left      string
		wantError bool
	}{
		{toParse: "5, 4", left: ", 4"},
		{toParse: "[5, 6], 4", left: ", 4"},
		{toParse: "[5, [7,8]]: 4", left: ": 4"},

		{toParse: `{"a":1}, 4`, left: ", 4"},
		{toParse: `{"a":1, "b":{"c": 5}, "e":[12,15]}, 4`, left: ", 4"},

		// array start/end chars in a string
		{toParse: `[5, "]"], 4`, left: ", 4"},
		{toParse: `[5, "\"]"], 4`, left: ", 4"},
		{toParse: `[5, "["], 4`, left: ", 4"},
		{toParse: `[5, "\"["], 4`, left: ", 4"},

		// object start/end chars in a string
		{toParse: `{"a}":1}, 4`, left: ", 4"},
		{toParse: `{"a\"}":1}, 4`, left: ", 4"},
		{toParse: `{"a{":1}, 4`, left: ", 4"},
		{toParse: `{"a\"{":1}, 4`, left: ", 4"},

		// object with double slashes at the end of string
		{toParse: `{"a":"hey\\"}, 4`, left: ", 4"},

		// make sure skipping an invalid json results in an error
		{toParse: `{"a": [ ##invalid json## ]}, 4`, wantError: true},
		{toParse: `{"a": [ [1], [ ##invalid json## ]]}, 4`, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		l.SkipRecursive()

		got := string(l.Data[l.pos:])
		if got != test.left {
			t.Errorf("[%d, %q] SkipRecursive() left = %v; want %v", i, test.toParse, got, test.left)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] SkipRecursive() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] SkipRecursive() ok; want error", i, test.toParse)
		}
	}
}

func TestInterface(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      any
		wantError bool
	}{
		{toParse: "null", want: nil},
		{toParse: "true", want: true},
		{toParse: `"a"`, want: "a"},
		{toParse: "5", want: float64(5)},

		{toParse: `{}`, want: map[string]any{}},
		{toParse: `[]`, want: []any{}},

		{toParse: `{"a": "b"}`, want: map[string]any{"a": "b"}},
		{toParse: `[5]`, want: []any{float64(5)}},

		{toParse: `{"a":5 , "b" : "string"}`, want: map[string]any{"a": float64(5), "b": "string"}},
		{toParse: `["a", 5 , null, true]`, want: []any{"a", float64(5), nil, true}},

		{toParse: `{"a" "b"}`, wantError: true},
		{toParse: `{"a": "b",}`, wantError: true},
		{toParse: `{"a":"b","c" "b"}`, wantError: true},
		{toParse: `{"a": "b","c":"d",}`, wantError: true},
		{toParse: `{,}`, wantError: true},

		{toParse: `[1, 2,]`, wantError: true},
		{toParse: `[1  2]`, wantError: true},
		{toParse: `[,]`, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Interface()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("[%d, %q] Interface() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Interface() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Interface() ok; want error", i, test.toParse)
		}
	}
}

func TestConsumed(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		wantError bool
	}{
		{toParse: "", wantError: false},
		{toParse: "   ", wantError: false},
		{toParse: "\r\n", wantError: false},
		{toParse: "\t\t ", wantError: false},

		{toParse: "{", wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}
		l.Consumed()

		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Consumed() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Consumed() ok; want error", i, test.toParse)
		}
	}
}

func TestJsonNumber(t *testing.T) {
	for i, test := range []struct {
		toParse        string
		want           json.Number
		wantLexerError bool
		wantValue      any
		wantValueError bool
	}{
		{toParse: `10`, want: json.Number("10"), wantValue: int64(10)},
		{toParse: `0`, want: json.Number("0"), wantValue: int64(0)},
		{toParse: `0.12`, want: json.Number("0.12"), wantValue: 0.12},
		{toParse: `25E-4`, want: json.Number("25E-4"), wantValue: 25e-4},

		{toParse: `"10"`, want: json.Number("10"), wantValue: int64(10)},
		{toParse: `"0"`, want: json.Number("0"), wantValue: int64(0)},
		{toParse: `"0.12"`, want: json.Number("0.12"), wantValue: 0.12},
		{toParse: `"25E-4"`, want: json.Number("25E-4"), wantValue: 25e-4},

		{toParse: `"foo"`, want: json.Number("foo"), wantValueError: true},
		{toParse: `null`, want: json.Number(""), wantValueError: true},

		{toParse: `"a""`, want: json.Number("a"), wantValueError: true},

		{toParse: `[1]`, want: json.Number(""), wantLexerError: true, wantValueError: true},
		{toParse: `{}`, want: json.Number(""), wantLexerError: true, wantValueError: true},
		{toParse: `a`, want: json.Number(""), wantLexerError: true, wantValueError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.JSONNumber()
		if got != test.want {
			t.Errorf("[%d, %q] JSONNumber() = %v; want %v", i, test.toParse, got, test.want)
		}

		err := l.Error()
		if err != nil && !test.wantLexerError {
			t.Errorf("[%d, %q] JSONNumber() lexer error: %v", i, test.toParse, err)
		} else if err == nil && test.wantLexerError {
			t.Errorf("[%d, %q] JSONNumber() ok; want lexer error", i, test.toParse)
		}

		var valueErr error
		var gotValue any
		switch test.wantValue.(type) {
		case float64:
			gotValue, valueErr = got.Float64()
		default:
			gotValue, valueErr = got.Int64()
		}

		if !reflect.DeepEqual(gotValue, test.wantValue) && !test.wantLexerError && !test.wantValueError {
			t.Errorf("[%d, %q] JSONNumber() = %v; want %v", i, test.toParse, gotValue, test.wantValue)
		}

		if valueErr != nil && !test.wantValueError {
			t.Errorf("[%d, %q] JSONNumber() value error: %v", i, test.toParse, valueErr)
		} else if valueErr == nil && test.wantValueError {
			t.Errorf("[%d, %q] JSONNumber() ok; want value error", i, test.toParse)
		}
	}
}

func TestFetchStringUnterminatedString(t *testing.T) {
	for _, test := range []struct {
		data []byte
	}{
		{data: []byte(`"sting without trailing quote`)},
		{data: []byte(`"\"`)},
		{data: []byte{'"'}},
	} {
		l := Lexer{Data: test.data}
		l.fetchString()
		if l.pos > len(l.Data) {
			t.Errorf("fetchString(%s): pos=%v should not be greater than length of Data = %v", test.data, l.pos, len(l.Data))
		}
		if l.Error() == nil {
			t.Errorf("fetchString(%s): should add parsing error", test.data)
		}
	}
}
