// Package jlexer contains a JSON lexer implementation.
//
// It is expected that it is mostly used with generated parser code, so the interface is tuned
// for a parser that knows what kind of data is expected.
// //nolint:funlen,mnd,exhaustive,wsl,err113,revive,nakedret,lll,gocognit,whitespace,gocritic,godot,goconst,gocyclo,gochecknoglobals,misspell,nlreturn,nonamedreturns
package jlexer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/josharian/intern"

	"github.com/0wnperception/go-helpers/pkg/slice"
)

// tokenKind determines type of a token.
type tokenKind byte

const (
	tokenUndef  tokenKind = iota // No token.
	tokenDelim                   // Delimiter: one of '{', '}', '[' or ']'.
	tokenString                  // A string literal, e.g. "abc\u1234"
	tokenNumber                  // Number literal, e.g. 1.5e5
	tokenBool                    // Boolean literal: true or false.
	tokenNull                    // null keyword.
)

// token describes a single token: type, position in the input and value.
type token struct {
	byteValue []byte    // Raw value of a token.
	kind      tokenKind // Type of a token.

	boolValue       bool // Value if a boolean literal token.
	byteValueCloned bool // true if byteValue was allocated and does not refer to original json body
	delimValue      byte
}

// Lexer is a JSON lexer: it iterates over JSON tokens in a byte slice.
type Lexer struct {
	fatalError error  // Fatal error occurred during lexing. It is usually a syntax error.
	Data       []byte // Input data given to the lexer.

	multipleErrors []*LexerError // Semantic errors occurred during lexing. Marshalling will be continued after finding this errors.
	token          token         // Last scanned token, if token.kind != tokenUndef.

	start int // Start of the current token.
	pos   int // Current unscanned position in the input stream.

	firstElement bool // Whether current element is the first in array or an object.
	wantSep      byte // A comma or a colon character, which need to occur before a token.

	UseMultipleErrors bool // If we want to use multiple errors.
}

var whitespaceOrCommaOrColon = [256]bool{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
	':':  true,
	',':  true,
}

func tokenQuote(r *Lexer, _ byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.token.kind = tokenString
	r.fetchString()
}

func tokenOpenDelim(r *Lexer, c byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.firstElement = true
	r.token.kind = tokenDelim
	r.token.delimValue = c
	r.pos++
}

func tokenCloseDelim(r *Lexer, c byte) {
	if !r.firstElement && (r.wantSep != ',') {
		r.errSyntax()
	}

	r.wantSep = 0
	r.token.kind = tokenDelim
	r.token.delimValue = c
	r.pos++
}

func tokenNum(r *Lexer, _ byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.token.kind = tokenNumber
	r.fetchNumber()
}

func tokenNullVal(r *Lexer, _ byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.token.kind = tokenNull
	r.fetchNull()
}

func tokenTrue(r *Lexer, _ byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.token.kind = tokenBool
	r.token.boolValue = true
	r.fetchTrue()
}

func tokenFalse(r *Lexer, _ byte) {
	if r.wantSep != 0 {
		r.errSyntax()
	}

	r.token.kind = tokenBool
	r.token.boolValue = false
	r.fetchFalse()
}

func tokenErr(r *Lexer, _ byte) {
	r.errSyntax()
}

var tokenProcessFuncs [256]func(r *Lexer, c byte)

var decodeEscapeFuncs [256]func(data []byte, c byte) (rune, int, error)

// //nolint: gochecknoinits
func init() {
	for i := range 256 {
		decodeEscapeFuncs[i] = decodeEscapeErr
	}

	decodeEscapeFuncs['"'] = decodeEscapeQuota
	decodeEscapeFuncs['/'] = decodeEscapeQuota
	decodeEscapeFuncs['\\'] = decodeEscapeQuota
	decodeEscapeFuncs['b'] = decodeEscapeB
	decodeEscapeFuncs['f'] = decodeEscapeF
	decodeEscapeFuncs['n'] = decodeEscapeN
	decodeEscapeFuncs['r'] = decodeEscapeR
	decodeEscapeFuncs['t'] = decodeEscapeT
	decodeEscapeFuncs['u'] = decodeEscapeU

	for i := range 256 {
		tokenProcessFuncs[i] = tokenErr
	}

	tokenProcessFuncs['"'] = tokenQuote
	tokenProcessFuncs['{'] = tokenOpenDelim
	tokenProcessFuncs['['] = tokenOpenDelim
	tokenProcessFuncs['}'] = tokenCloseDelim
	tokenProcessFuncs[']'] = tokenCloseDelim
	tokenProcessFuncs['0'] = tokenNum
	tokenProcessFuncs['1'] = tokenNum
	tokenProcessFuncs['2'] = tokenNum
	tokenProcessFuncs['3'] = tokenNum
	tokenProcessFuncs['4'] = tokenNum
	tokenProcessFuncs['5'] = tokenNum
	tokenProcessFuncs['6'] = tokenNum
	tokenProcessFuncs['7'] = tokenNum
	tokenProcessFuncs['8'] = tokenNum
	tokenProcessFuncs['9'] = tokenNum
	tokenProcessFuncs['-'] = tokenNum
	tokenProcessFuncs['n'] = tokenNullVal
	tokenProcessFuncs['t'] = tokenTrue
	tokenProcessFuncs['f'] = tokenFalse

}

// FetchToken scans the input for the next token.
func (r *Lexer) FetchToken() {
	r.token.kind = tokenUndef
	r.start = r.pos

	// Check if r.Data has r.pos element
	// If it doesn't, it mean corrupted input data
	if len(r.Data) < r.pos {
		r.errParse("Unexpected end of data")

		return
	}
	// Determine the type of token by skipping whitespace and reading the
	// first character.
	for _, c := range r.Data[r.pos:] {
		// skip whitespaces
		if whitespaceOrCommaOrColon[c] {
			if c == ':' || c == ',' {
				if r.wantSep == c {
					r.pos++
					r.start++
					r.wantSep = 0
				} else {
					r.errSyntax()
				}
			} else {
				r.pos++
				r.start++
			}

			continue
		}

		f := tokenProcessFuncs[c]

		f(r, c)

		return
	}

	r.fatalError = io.EOF
}

var tokenEndMap = [256]bool{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
	'{':  true,
	'}':  true,
	'[':  true,
	']':  true,
	',':  true,
	':':  true,
}

// isTokenEnd returns true if the char can follow a non-delimiter token
func isTokenEnd(c byte) bool {
	return tokenEndMap[c]
}

// fetchNull fetches and checks remaining bytes of null keyword.
func (r *Lexer) fetchNull() {
	if r.pos+4 > len(r.Data) {
		r.errSyntax()

		return
	}

	pos := r.pos + 4

	lastIndex := pos

	if lastIndex < len(r.Data) {
		lastIndex++
	}

	d := r.Data[r.pos:lastIndex]

	if pos != len(r.Data) && !isTokenEnd(d[4]) {
		r.errSyntax()

		return
	}

	if d[1] != 'u' || d[2] != 'l' || d[3] != 'l' {
		r.errSyntax()

		return
	}

	r.pos = pos
}

// fetchTrue fetches and checks remaining bytes of true keyword.
func (r *Lexer) fetchTrue() {
	if r.pos+4 > len(r.Data) {
		r.errSyntax()

		return
	}

	pos := r.pos + 4

	lastIndex := pos

	if lastIndex < len(r.Data) {
		lastIndex++
	}

	d := r.Data[r.pos:lastIndex]

	if pos != len(r.Data) && !isTokenEnd(d[4]) {
		r.errSyntax()

		return
	}

	if d[1] != 'r' || d[2] != 'u' || d[3] != 'e' {
		r.errSyntax()

		return
	}

	r.pos = pos
}

// fetchFalse fetches and checks remaining bytes of false keyword.
func (r *Lexer) fetchFalse() {
	if r.pos+4 > len(r.Data) {
		r.errSyntax()

		return
	}

	pos := r.pos + 5

	lastIndex := pos

	if lastIndex < len(r.Data) {
		lastIndex++
	}

	d := r.Data[r.pos:lastIndex]

	if pos != len(r.Data) && !isTokenEnd(d[5]) {
		r.errSyntax()

		return
	}

	if d[1] != 'a' || d[2] != 'l' || d[3] != 's' || d[4] != 'e' {
		r.errSyntax()

		return
	}

	r.pos = pos
}

// fetchNumber scans a number literal token.
func (r *Lexer) fetchNumber() {
	hasE := false
	afterE := false
	hasDot := false

	pos := r.pos + 1

	for i, c := range r.Data[pos:] {
		switch {
		case c >= '0' && c <= '9':
			afterE = false
		case c == '.' && !hasDot:
			hasDot = true
		case (c == 'e' || c == 'E') && !hasE:
			hasE = true
			hasDot = true
			afterE = true
		case (c == '+' || c == '-') && afterE:
			afterE = false
		default:
			pos += i

			r.pos = pos

			if !isTokenEnd(c) {
				r.errSyntax()
			} else {
				r.token.byteValue = r.Data[r.start:pos]
			}

			return
		}
	}

	r.pos = len(r.Data)
	r.token.byteValue = r.Data[r.start:]
}

// findStringLen tries to scan into the string literal for ending quote char to determine required size.
// The size will be exact if no escapes are present and may be inexact if there are escaped chars.
func findStringLen(data []byte) (bool, int) {
	length := 0

	for {
		idx := bytes.IndexByte(data, '"')
		if idx == -1 {
			return false, len(data)
		}

		if idx == 0 || (idx > 0 && data[idx-1] != '\\') {
			return true, length + idx
		}

		// count \\\\\\\ sequences. even number of slashes means quote is not really escaped
		cnt := 1
		for idx-cnt-1 >= 0 && data[idx-cnt-1] == '\\' {
			cnt++
		}

		if cnt%2 == 0 {
			return true, length + idx
		}

		length += idx + 1
		data = data[idx+1:]
	}
}

// unescapeStringToken performs unescaping of string token.
// if no escaping is needed, original string is returned, otherwise - a new one allocated
func (r *Lexer) unescapeStringToken() error {
	data := r.token.byteValue
	var unescapedData []byte

	for {
		i := bytes.IndexByte(data, '\\')
		if i == -1 {
			break
		}

		escapedRune, escapedBytes, err := decodeEscape(data[i:])
		if err != nil {
			r.errParse(err.Error())

			return err
		}

		if unescapedData == nil {
			unescapedData = make([]byte, 0, len(r.token.byteValue))
		}

		unescapedData = append(unescapedData, data[:i]...)
		unescapedData = utf8.AppendRune(unescapedData, escapedRune)

		data = data[i+escapedBytes:]
	}

	if unescapedData != nil {
		r.token.byteValue = append(unescapedData, data...)
		r.token.byteValueCloned = true
	}

	return nil
}

var u4 = [256]byte{
	'0': '0',
	'1': '0',
	'2': '0',
	'3': '0',
	'4': '0',
	'5': '0',
	'6': '0',
	'7': '0',
	'8': '0',
	'9': '0',
	'a': 'a' - 10,
	'b': 'a' - 10,
	'c': 'a' - 10,
	'd': 'a' - 10,
	'e': 'a' - 10,
	'f': 'a' - 10,
	'A': 'A' - 10,
	'B': 'A' - 10,
	'C': 'A' - 10,
	'D': 'A' - 10,
	'E': 'A' - 10,
	'F': 'A' - 10,
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}

	var val rune

	for i := 2; i < len(s) && i < 6; i++ {
		var v byte
		c := s[i]

		r := u4[c]
		if r == 0 {
			return -1
		}

		v = c - r
		val <<= 4
		val |= rune(v)
	}

	return val
}

func decodeEscapeQuota(_ []byte, c byte) (rune, int, error) {
	return rune(c), 2, nil
}

func decodeEscapeB(_ []byte, _ byte) (rune, int, error) {
	return '\b', 2, nil
}

func decodeEscapeF(_ []byte, _ byte) (rune, int, error) {
	return '\f', 2, nil
}

func decodeEscapeN(_ []byte, _ byte) (rune, int, error) {
	return '\n', 2, nil
}

func decodeEscapeR(_ []byte, _ byte) (rune, int, error) {
	return '\r', 2, nil
}

func decodeEscapeT(_ []byte, _ byte) (rune, int, error) {
	return '\t', 2, nil
}

func decodeEscapeU(data []byte, _ byte) (rune, int, error) {
	var rr rune

	if len(data) < 6 || data[0] != '\\' || data[1] != 'u' {
		rr = -1
	} else {
		var val rune

		for i := 2; i < len(data) && i < 6; i++ {
			var v byte
			c := data[i]

			r := u4[c]
			if r == 0 {
				rr = -1

				break
			}

			v = c - r
			val <<= 4
			val |= rune(v)

			rr = val
		}
	}

	if rr < 0 {
		return 0, 0, errors.New("incorrectly escaped \\uXXXX sequence")
	}

	read := 6
	if utf16.IsSurrogate(rr) {
		rr1 := getu4(data[read:])
		if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
			read += 6
			rr = dec
		} else {
			rr = unicode.ReplacementChar
		}
	}

	return rr, read, nil
}

func decodeEscapeErr(_ []byte, _ byte) (rune, int, error) {
	return 0, 0, errors.New("incorrectly escaped bytes")
}

// decodeEscape processes a single escape sequence and returns number of bytes processed.
func decodeEscape(data []byte) (decoded rune, bytesProcessed int, err error) {
	if len(data) < 2 {
		return 0, 0, errors.New("incorrect escape symbol \\ at the end of token")
	}

	c := data[1]

	f := decodeEscapeFuncs[c]

	return f(data, c)
}

// fetchString scans a string literal token.
func (r *Lexer) fetchString() {
	pos := r.pos + 1

	data := r.Data[pos:]

	isValid, length := findStringLen(data)
	if !isValid {
		pos += length
		r.pos = pos
		r.errParse("unterminated string literal")

		return
	}

	r.token.byteValue = data[:length]
	r.pos = pos + length + 1 // skip closing '"' as well
}

// scanToken scans the next token if no token is currently available in the lexer.
func (r *Lexer) scanToken() {
	if r.token.kind != tokenUndef || r.fatalError != nil {
		return
	}

	r.FetchToken()
}

// consume resets the current token to allow scanning the next one.
func (r *Lexer) consume() {
	r.token.kind = tokenUndef
	r.token.byteValueCloned = false
	r.token.delimValue = 0
}

// Ok returns true if no error (including io.EOF) was encountered during scanning.
func (r *Lexer) Ok() bool {
	return r.fatalError == nil
}

const maxErrorContextLen = 13

func (r *Lexer) errParse(what string) {
	if r.fatalError == nil {
		var str string

		if len(r.Data)-r.pos <= maxErrorContextLen {
			str = string(r.Data)
		} else {
			str = string(r.Data[r.pos:r.pos+maxErrorContextLen-3]) + "..."
		}

		r.fatalError = &LexerError{
			Reason: what,
			Offset: r.pos,
			Data:   str,
		}
	}
}

func (r *Lexer) errSyntax() {
	r.errParse("syntax error")
}

func (r *Lexer) errInvalidToken(expected string) {
	if r.fatalError != nil {
		return
	}

	if r.UseMultipleErrors {
		r.pos = r.start
		r.consume()
		r.SkipRecursive()

		switch expected {
		case "[":
			r.token.delimValue = ']'
			r.token.kind = tokenDelim
		case "{":
			r.token.delimValue = '}'
			r.token.kind = tokenDelim
		}
		r.addNonfatalError(&LexerError{
			Reason: "expected " + expected,
			Offset: r.start,
			Data:   string(r.Data[r.start:r.pos]),
		})

		return
	}

	var str string

	if len(r.token.byteValue) <= maxErrorContextLen {
		str = string(r.token.byteValue)
	} else {
		str = string(r.token.byteValue[:maxErrorContextLen-3]) + "..."
	}

	r.fatalError = &LexerError{
		Reason: "expected " + expected,
		Offset: r.pos,
		Data:   str,
	}
}

func (r *Lexer) GetPos() int {
	return r.pos
}

// Delim consumes a token and verifies that it is the given delimiter.
func (r *Lexer) Delim(c byte) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.delimValue != c {
		r.consume() // errInvalidToken can change token if UseMultipleErrors is enabled.
		r.errInvalidToken(string([]byte{c}))
	} else {
		r.consume()
	}
}

// IsDelim returns true if there was no scanning error and next token is the given delimiter.
func (r *Lexer) IsDelim(c byte) bool {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	return !r.Ok() || r.token.delimValue == c
}

// Null verifies that the next token is null and consumes it.
func (r *Lexer) Null() {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenNull {
		r.errInvalidToken("null")
	}

	r.consume()
}

// IsNull returns true if the next token is a null keyword.
func (r *Lexer) IsNull() bool {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	return r.Ok() && r.token.kind == tokenNull
}

// Skip skips a single token.
func (r *Lexer) Skip() {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	r.consume()
}

// SkipRecursive skips next array or object completely, or just skips a single token if not
// an array/object.
//
// Note: no syntax validation is performed on the skipped data.
func (r *Lexer) SkipRecursive() {
	r.scanToken()
	var start, end byte
	startPos := r.start

	switch r.token.delimValue {
	case '{':
		start, end = '{', '}'
	case '[':
		start, end = '[', ']'
	default:
		r.consume()

		return
	}

	r.consume()

	level := 1
	inQuotes := false
	wasEscape := false

	for i, c := range r.Data[r.pos:] {
		switch {
		case c == start && !inQuotes:
			level++
		case c == end && !inQuotes:
			level--
			if level == 0 {
				r.pos += i + 1
				if !json.Valid(r.Data[startPos:r.pos]) {
					r.pos = len(r.Data)
					r.fatalError = &LexerError{
						Reason: "skipped array/object json value is invalid",
						Offset: r.pos,
						Data:   string(r.Data[r.pos:]),
					}
				}

				return
			}
		case c == '\\' && inQuotes:
			wasEscape = !wasEscape

			continue
		case c == '"' && inQuotes:
			inQuotes = wasEscape
		case c == '"':
			inQuotes = true
		}

		wasEscape = false
	}

	r.pos = len(r.Data)
	r.fatalError = &LexerError{
		Reason: "EOF reached while skipping array/object or token",
		Offset: r.pos,
		Data:   string(r.Data[r.pos:]),
	}
}

// Raw fetches the next item recursively as a data slice
func (r *Lexer) Raw() []byte {
	r.SkipRecursive()

	if !r.Ok() {
		return nil
	}

	return r.Data[r.start:r.pos]
}

// IsStart returns whether the lexer is positioned at the start
// of an input string.
func (r *Lexer) IsStart() bool {
	return r.pos == 0
}

var consumed = [256]bool{
	' ':  true,
	'\t': true,
	'\r': true,
	'\n': true,
}

// Consumed reads all remaining bytes from the input, publishing an error if
// there is anything but whitespace remaining.
func (r *Lexer) Consumed() {
	if r.pos > len(r.Data) || !r.Ok() {
		return
	}

	pos := r.pos
	start := r.start

	for _, c := range r.Data[pos:] {
		if !consumed[c] {
			r.AddError(&LexerError{
				Reason: "invalid character '" + string(c) + "' after top-level value",
				Offset: r.pos,
				Data:   string(r.Data[r.pos:]),
			})

			r.pos = pos
			r.start = start

			return
		}

		pos++
		start++
	}

	r.pos = pos
	r.start = start
}

func (r *Lexer) unsafeString(skipUnescape bool) (string, []byte) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")

		return "", nil
	}

	if !skipUnescape {
		if err := r.unescapeStringToken(); err != nil {
			r.errInvalidToken("string")

			return "", nil
		}
	}

	bs := r.token.byteValue

	ret := slice.ToString(r.token.byteValue)

	r.consume()

	return ret, bs
}

// UnsafeString returns the string value if the token is a string literal.
//
// Warning: returned string may point to the input buffer, so the string should not outlive
// the input buffer. Intended pattern of usage is as an argument to a switch statement.
func (r *Lexer) UnsafeString() string {
	ret, _ := r.unsafeString(false)

	return ret
}

// UnsafeBytes returns the byte slice if the token is a string literal.
func (r *Lexer) UnsafeBytes() []byte {
	_, ret := r.unsafeString(false)

	return ret
}

// UnsafeFieldName returns current member name string token
func (r *Lexer) UnsafeFieldName(skipUnescape bool) string {
	ret, _ := r.unsafeString(skipUnescape)

	return ret
}

// String reads a string literal.
func (r *Lexer) String() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")

		return ""
	}

	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")

		return ""
	}

	var ret string

	if r.token.byteValueCloned {
		ret = slice.ToString(r.token.byteValue)
	} else {
		ret = string(r.token.byteValue)
	}

	r.consume()

	return ret
}

// StringIntern reads a string literal, and performs string interning on it.
func (r *Lexer) StringIntern() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")

		return ""
	}

	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")

		return ""
	}

	ret := intern.Bytes(r.token.byteValue)
	r.consume()

	return ret
}

// Bytes reads a string literal and base64 decodes it into a byte slice.
func (r *Lexer) Bytes() []byte {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")

		return nil
	}

	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")

		return nil
	}

	ret := make([]byte, base64.StdEncoding.DecodedLen(len(r.token.byteValue)))

	n, err := base64.StdEncoding.Decode(ret, r.token.byteValue)
	if err != nil {
		r.fatalError = &LexerError{
			Reason: err.Error(),
		}

		return nil
	}

	r.consume()

	return ret[:n]
}

// Bool reads a true or false boolean keyword.
func (r *Lexer) Bool() bool {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenBool {
		r.errInvalidToken("bool")

		return false
	}

	ret := r.token.boolValue
	r.consume()

	return ret
}

func (r *Lexer) number() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.kind != tokenNumber {
		r.errInvalidToken("number")

		return ""
	}

	ret := slice.ToString(r.token.byteValue)
	r.consume()

	return ret
}

func (r *Lexer) uintBits(bits int) uint64 {
	s := r.number()

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   s,
		})

		return 0
	}

	return n
}

func (r *Lexer) Uint8() uint8 {
	return uint8(r.uintBits(8)) //nolint:gosec
}

func (r *Lexer) Uint16() uint16 {
	return uint16(r.uintBits(16)) //nolint:gosec
}

func (r *Lexer) Uint32() uint32 {
	return uint32(r.uintBits(32)) //nolint:gosec
}

func (r *Lexer) Uint64() uint64 {
	return r.uintBits(64)
}

func (r *Lexer) Uint() uint {
	return uint(r.Uint64())
}

func (r *Lexer) intBits(bits int) int64 {
	s := r.number()
	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   s,
		})

		return 0
	}

	return n
}

func (r *Lexer) Int8() int8 {
	return int8(r.intBits(8)) //nolint:gosec
}

func (r *Lexer) Int16() int16 {
	return int16(r.intBits(16)) //nolint:gosec
}

func (r *Lexer) Int32() int32 {
	return int32(r.intBits(32)) //nolint:gosec
}

func (r *Lexer) Int64() int64 {
	return r.intBits(64)
}

func (r *Lexer) Int() int {
	return int(r.Int64())
}

func (r *Lexer) uintStrBits(bits int) uint64 {
	s, b := r.unsafeString(false)

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   string(b),
		})

		return 0
	}

	return n
}

func (r *Lexer) Uint8Str() uint8 {
	return uint8(r.uintStrBits(8)) //nolint:gosec
}

func (r *Lexer) Uint16Str() uint16 {
	return uint16(r.uintStrBits(16)) //nolint:gosec
}

func (r *Lexer) Uint32Str() uint32 {
	return uint32(r.uintStrBits(32)) //nolint:gosec
}

func (r *Lexer) Uint64Str() uint64 {
	return r.uintStrBits(64)
}

func (r *Lexer) UintStr() uint {
	return uint(r.Uint64Str())
}

func (r *Lexer) UintptrStr() uintptr {
	return uintptr(r.Uint64Str())
}

func (r *Lexer) intStr(bitSize int) int64 {
	s, b := r.unsafeString(false)

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   string(b),
		})

		return 0
	}

	return n
}

func (r *Lexer) Int8Str() int8 {
	return int8(r.intStr(8)) //nolint:gosec
}

func (r *Lexer) Int16Str() int16 {
	return int16(r.intStr(16)) //nolint:gosec
}

func (r *Lexer) Int32Str() int32 {
	return int32(r.intStr(32)) //nolint:gosec
}

func (r *Lexer) Int64Str() int64 {
	return r.intStr(64)
}

func (r *Lexer) IntStr() int {
	return int(r.Int64Str())
}

func (r *Lexer) Float32() float32 {
	s := r.number()

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseFloat(s, 32)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   s,
		})
	}

	return float32(n)
}

func (r *Lexer) Float32Str() float32 {
	s, b := r.unsafeString(false)

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseFloat(s, 32)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   string(b),
		})
	}

	return float32(n)
}

func (r *Lexer) Float64() float64 {
	s := r.number()

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   s,
		})
	}

	return n
}

func (r *Lexer) Float64Str() float64 {
	s, b := r.unsafeString(false)

	if !r.Ok() {
		return 0
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		r.addNonfatalError(&LexerError{
			Offset: r.start,
			Reason: err.Error(),
			Data:   string(b),
		})
	}

	return n
}

func (r *Lexer) Error() error {
	return r.fatalError
}

func (r *Lexer) AddError(e error) {
	if r.fatalError == nil {
		r.fatalError = e
	}
}

func (r *Lexer) AddNonFatalError(e error) {
	r.addNonfatalError(&LexerError{
		Offset: r.start,
		Data:   string(r.Data[r.start:r.pos]),
		Reason: e.Error(),
	})
}

func (r *Lexer) addNonfatalError(err *LexerError) {
	if r.UseMultipleErrors {
		// We don't want to add errors with the same offset.
		if len(r.multipleErrors) != 0 && r.multipleErrors[len(r.multipleErrors)-1].Offset == err.Offset {
			return
		}

		r.multipleErrors = append(r.multipleErrors, err)

		return
	}

	r.fatalError = err
}

func (r *Lexer) GetNonFatalErrors() []*LexerError {
	return r.multipleErrors
}

// JSONNumber fetches and json.Number from 'encoding/json' package.
// Both int, float or string, contains them are valid values
func (r *Lexer) JSONNumber() json.Number {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() {
		r.errInvalidToken("json.Number")

		return ""
	}

	switch r.token.kind {
	case tokenString:
		return json.Number(r.String())
	case tokenNumber:
		return json.Number(r.Raw())
	case tokenNull:
		r.Null()

		return ""
	default:
		r.errSyntax()

		return ""
	}
}

// Interface fetches an any analogous to the 'encoding/json' package.
func (r *Lexer) Interface() any {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() {
		return nil
	}

	switch r.token.kind {
	case tokenString:
		return r.String()
	case tokenNumber:
		return r.Float64()
	case tokenBool:
		return r.Bool()
	case tokenNull:
		r.Null()

		return nil
	}

	if r.token.delimValue == '{' {
		r.consume()

		ret := map[string]any{}

		for !r.IsDelim('}') {
			key := r.String()
			r.WantColon()
			ret[key] = r.Interface()
			r.WantComma()
		}

		r.Delim('}')

		if r.Ok() {
			return ret
		} else {
			return nil
		}
	} else if r.token.delimValue == '[' {
		r.consume()

		ret := []any{}

		for !r.IsDelim(']') {
			ret = append(ret, r.Interface())
			r.WantComma()
		}

		r.Delim(']')

		if r.Ok() {
			return ret
		} else {
			return nil
		}
	}

	r.errSyntax()

	return nil
}

// WantComma requires a comma to be present before fetching next token.
func (r *Lexer) WantComma() {
	r.wantSep = ','
	r.firstElement = false
}

// WantColon requires a colon to be present before fetching next token.
func (r *Lexer) WantColon() {
	r.wantSep = ':'
	r.firstElement = false
}
