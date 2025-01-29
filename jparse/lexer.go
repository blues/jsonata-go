// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jparse

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

type tokenType uint8

const (
	typeEOF tokenType = iota
	typeError

	typeString   // string literal, e.g. "hello"
	typeNumber   // number literal, e.g. 3.14159
	typeBoolean  // true or false
	typeNull     // null
	typeName     // field name, e.g. Price
	typeNameEsc  // escaped field name, e.g. `Product Name`
	typeVariable // variable, e.g. $x
	typeRegex    // regular expression, e.g. /ab+/

	// Symbol operators
	typeBracketOpen
	typeBracketClose
	typeBraceOpen
	typeBraceClose
	typeParenOpen
	typeParenClose
	typeDot
	typeComma
	typeColon
	typeSemicolon
	typeCondition
	typePlus
	typeMinus
	typeMult
	typeDiv
	typeMod
	typeParent   // %
	typeCrossRef // @
	typePosition // #
	typePipe
	typeEqual
	typeNotEqual
	typeLess
	typeLessEqual
	typeGreater
	typeGreaterEqual
	typeApply
	typeSort
	typeConcat
	typeRange
	typeAssign
	typeDescendent

	// Keyword operators
	typeAnd
	typeOr
	typeIn
)

func (tt tokenType) String() string {
	switch tt {
	case typeEOF:
		return "(eof)"
	case typeError:
		return "(error)"
	case typeString:
		return "(string)"
	case typeNumber:
		return "(number)"
	case typeBoolean:
		return "(boolean)"
	case typeName, typeNameEsc:
		return "(name)"
	case typeVariable:
		return "(variable)"
	case typeRegex:
		return "(regex)"
	default:
		if s := symbolsAndKeywords[tt]; s != "" {
			return s
		}
		return "(unknown)"
	}
}

// symbols1 maps 1-character symbols to the corresponding
// token types.
var symbols1 = [...]tokenType{
	'[': typeBracketOpen,
	']': typeBracketClose,
	'{': typeBraceOpen,
	'}': typeBraceClose,
	'(': typeParenOpen,
	')': typeParenClose,
	'.': typeDot,
	',': typeComma,
	';': typeSemicolon,
	':': typeColon,
	'?': typeCondition,
	'+': typePlus,
	'-': typeMinus,
	'*': typeMult,
	'/': typeDiv,
	'%': typeMod,
	'@': typeCrossRef,
	'#': typePosition,
	'|': typePipe,
	'=': typeEqual,
	'<': typeLess,
	'>': typeGreater,
	'^': typeSort,
	'&': typeConcat,
}

type runeTokenType struct {
	r  rune
	tt tokenType
}

// symbols2 maps 2-character symbols to the corresponding
// token types.
var symbols2 = [...][]runeTokenType{
	'%': {{'%', typeMod}},
	'!': {{'=', typeNotEqual}},
	'<': {{'=', typeLessEqual}},
	'>': {{'=', typeGreaterEqual}},
	'.': {{'.', typeRange}},
	'~': {{'>', typeApply}},
	':': {{'=', typeAssign}},
	'*': {{'*', typeDescendent}},
}

const (
	symbol1Count = rune(len(symbols1))
	symbol2Count = rune(len(symbols2))
)

func lookupSymbol1(r rune) tokenType {
	if r < 0 || r >= symbol1Count {
		return 0
	}
	return symbols1[r]
}

func lookupSymbol2(r rune) []runeTokenType {
	if r < 0 || r >= symbol2Count {
		return nil
	}
	return symbols2[r]
}

func lookupKeyword(s string) tokenType {
	switch s {
	case "and":
		return typeAnd
	case "or":
		return typeOr
	case "in":
		return typeIn
	case "true", "false":
		return typeBoolean
	case "null":
		return typeNull
	default:
		return 0
	}
}

// A token represents a discrete part of a JSONata expression
// such as a string, a number, a field name, or an operator.
type token struct {
	Type     tokenType
	Value    string
	Position int
	Flags    string // Used for regex flags
}

// lexer converts a JSONata expression into a sequence of tokens.
// The implmentation is based on the technique described in Rob
// Pike's 'Lexical Scanning in Go' talk.
type lexer struct {
	input   string
	length  int
	start   int
	current int
	width   int
	err     error
}

// newLexer creates a new lexer from the provided input. The
// input is tokenized by successive calls to the next method.
func newLexer(input string) lexer {
	return lexer{
		input:  input,
		length: len(input),
	}
}

// next returns the next token from the provided input. When
// the end of the input is reached, next returns EOF for all
// subsequent calls.
//
// The allowRegex argument determines how the lexer interprets
// a forward slash character. Forward slashes in JSONata can
// either be the start of a regular expression or the division
// operator depending on their position. If allowRegex is true,
// the lexer will treat a forward slash like a regular
// expression.
func (l *lexer) next(allowRegex bool) token {
	l.skipWhitespace()

	pos := l.start
	ch := l.nextRune()
	if ch == eof {
		return l.eof()
	}

	// Handle comments and regex
	if ch == '/' {
		next := l.peek()
		if next == '*' || next == '/' {
			l.backup()
			return l.scanComment(allowRegex)
		}
		if allowRegex {
			l.backup()
			return l.scanRegex(ch)
		}
		return token{Type: typeDiv, Value: "/", Position: pos}
	}

	// Handle two-character operators first
	if rts := lookupSymbol2(ch); rts != nil {
		for _, rt := range rts {
			if l.acceptRune(rt.r) {
				return token{Type: rt.tt, Value: string(ch) + string(rt.r), Position: pos}
			}
		}
		l.backup()
	}

	// Handle single-character operators and special cases
	if tt := lookupSymbol1(ch); tt > 0 {
		switch ch {
		case '%':
			next := l.peek()
			if next == '.' || next == '[' || next == ']' {
				return token{Type: typeParent, Value: "%", Position: pos}
			}
			return token{Type: typeMod, Value: "%", Position: pos}

		case '@':
			if l.acceptRune('$') {
				start := l.current
				for {
					r := l.nextRune()
					if r == eof || isWhitespace(r) || lookupSymbol1(r) > 0 || lookupSymbol2(r) != nil {
						l.backup()
						break
					}
				}
				if l.current > start {
					return token{Type: typeVariable, Value: l.input[start:l.current], Position: pos + 1}
				}
				l.backup() // Remove the $
			}
			return token{Type: typeCrossRef, Value: "@", Position: pos}

		case '#':
			if l.acceptRune('$') {
				start := l.current
				for {
					r := l.nextRune()
					if r == eof || isWhitespace(r) || lookupSymbol1(r) > 0 || lookupSymbol2(r) != nil {
						l.backup()
						break
					}
				}
				if l.current > start {
					return token{Type: typeVariable, Value: l.input[start:l.current], Position: pos + 1}
				}
				l.backup() // Remove the $
			}
			return token{Type: typePosition, Value: "#", Position: pos}

		default:
			return token{Type: tt, Value: string(ch), Position: pos}
		}
	}

	// Handle strings
	if ch == '"' || ch == '\'' {
		l.backup()
		return l.scanString(ch)
	}

	// Handle numbers
	if ch == '-' || ch == '.' || (ch >= '0' && ch <= '9') {
		l.backup()
		return l.scanNumber()
	}

	// Handle escaped names
	if ch == '`' {
		l.backup()
		return l.scanEscapedName(ch)
	}

	// Handle names and variables
	l.backup()
	return l.scanName()
}

// scanRegex reads a regular expression from the current position
// and returns a regex token. The opening delimiter has already
// been consumed.
func (l *lexer) scanRegex(delim rune) token {
	pos := l.start
	l.nextRune() // consume opening '/'

	var pattern strings.Builder
	escaped := false
	for {
		ch := l.nextRune()
		if ch == eof {
			return token{Type: typeError, Value: "unterminated regular expression (no closing '/')", Position: pos}
		}
		if escaped {
			pattern.WriteRune('\\')
			pattern.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '/' && !escaped {
			break
		}
		pattern.WriteRune(ch)
	}

	if pattern.Len() == 0 {
		return token{Type: typeError, Value: "invalid regular expression: expression cannot be empty", Position: pos}
	}

	var flags strings.Builder
	for {
		ch := l.peek()
		if !isRegexFlag(ch) {
			break
		}
		l.nextRune()
		flags.WriteRune(ch)
	}

	flagStr := flags.String()
	if flagStr != "" {
		flagStr = "(?i" + strings.ReplaceAll(flagStr, "i", "") + ")"
	}

	return token{Type: typeRegex, Value: flagStr + pattern.String(), Position: pos}
}

// scanString reads a string literal from the current position
// and returns a string token. The opening quote has already been
// consumed.
func (l *lexer) scanString(quote rune) token {
	pos := l.start
	l.nextRune() // consume opening quote

	var value strings.Builder
	for {
		ch := l.nextRune()
		if ch == eof {
			return l.error(ErrUnterminatedString, string(quote))
		}
		if ch == quote {
			break
		}
		if ch == '\\' {
			ch = l.nextRune()
			if ch == eof {
				return l.error(ErrUnterminatedString, string(quote))
			}
			switch ch {
			case 'n':
				value.WriteRune('\n')
			case 'r':
				value.WriteRune('\r')
			case 't':
				value.WriteRune('\t')
			case '"', '\'', '\\':
				value.WriteRune(ch)
			default:
				value.WriteRune('\\')
				value.WriteRune(ch)
			}
			continue
		}
		value.WriteRune(ch)
	}

	return token{
		Type:     typeString,
		Value:    value.String(),
		Position: pos,
	}
}

// scanNumber reads a number literal from the current position
// and returns a number token.
func (l *lexer) scanNumber() token {
	pos := l.start

	// Handle negative numbers
	isNegative := l.acceptRune('-')
	if isNegative && !isDigit(l.peek()) {
		return token{Type: typeMinus, Value: "-", Position: pos}
	}

	// Handle special number formats (hex, binary, octal)
	if l.acceptRune('0') {
		next := l.peek()
		switch next {
		case 'b', 'B':
			l.nextRune()
			if !l.acceptAll(isBinaryDigit) {
				return token{Type: typeError, Value: "invalid binary number", Position: pos}
			}
			return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
		case 'o', 'O':
			l.nextRune()
			if !l.acceptAll(isOctalDigit) {
				return token{Type: typeError, Value: "invalid octal number", Position: pos}
			}
			return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
		case 'x', 'X':
			l.nextRune()
			if !l.acceptAll(isHexDigit) {
				return token{Type: typeError, Value: "invalid hexadecimal number", Position: pos}
			}
			return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
		case '.':
			l.nextRune()
			if !l.acceptAll(isDigit) {
				l.backup()
				return token{Type: typeNumber, Value: "0", Position: pos}
			}
			if l.acceptRunes2('e', 'E') {
				l.acceptRunes2('+', '-')
				if !l.acceptAll(isDigit) {
					return token{Type: typeError, Value: "invalid number literal", Position: pos}
				}
			}
			return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
		case 'e', 'E':
			l.nextRune()
			l.acceptRunes2('+', '-')
			if !l.acceptAll(isDigit) {
				return token{Type: typeError, Value: "invalid number literal", Position: pos}
			}
			return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
		}

		// Handle single zero or leading zeros
		if !isDigit(next) {
			return token{Type: typeNumber, Value: "0", Position: pos}
		}

		l.acceptAll(isDigit)
		if l.acceptRune('.') {
			if !l.acceptAll(isDigit) {
				l.backup()
				return token{Type: typeNumber, Value: l.input[pos : l.current-1], Position: pos}
			}
		}
		if l.acceptRunes2('e', 'E') {
			l.acceptRunes2('+', '-')
			if !l.acceptAll(isDigit) {
				return token{Type: typeError, Value: "invalid number literal", Position: pos}
			}
		}
		return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
	}

	// Handle regular decimal numbers
	hasDigits := l.acceptAll(isDigit)

	// Handle decimal point
	if l.acceptRune('.') {
		if !l.acceptAll(isDigit) && !hasDigits {
			l.backup()
			return token{Type: typeDot, Value: ".", Position: pos}
		}
	}

	// Handle exponent
	if l.acceptRunes2('e', 'E') {
		l.acceptRunes2('+', '-')
		if !l.acceptAll(isDigit) {
			return token{Type: typeError, Value: "invalid number literal", Position: pos}
		}
	}

	if !hasDigits && !l.acceptAll(isDigit) {
		return token{Type: typeError, Value: "invalid number literal", Position: pos}
	}

	return token{Type: typeNumber, Value: l.input[pos:l.current], Position: pos}
}

// scanEscapedName reads a field name from the current position
// and returns a name token. The opening quote has already been
// consumed.
func (l *lexer) scanEscapedName(quote rune) token {
Loop:
	for {
		switch l.nextRune() {
		case quote:
			break Loop
		case eof, '\n':
			return l.error(ErrUnterminatedName, string(quote))
		}
	}

	l.backup()
	t := l.newToken(typeNameEsc)
	l.acceptRune(quote)
	l.ignore()
	return t
}

// scanName reads from the current position and returns a name,
// variable, or keyword token.
func (l *lexer) scanName() token {
	pos := l.start
	isVar := l.acceptRune('$')
	start := l.current

	for {
		ch := l.nextRune()
		if ch == eof {
			break
		}

		if isWhitespace(ch) || lookupSymbol1(ch) > 0 || lookupSymbol2(ch) != nil {
			l.backup()
			break
		}
	}

	if l.current == start {
		if isVar {
			return token{Type: typeVariable, Value: "", Position: pos}
		}
		return token{Type: typeError, Value: "empty name", Position: pos}
	}

	value := l.input[start:l.current]
	if isVar {
		return token{Type: typeVariable, Value: value, Position: pos + 1}
	}

	if tt := lookupKeyword(value); tt > 0 {
		return token{Type: tt, Value: value, Position: pos}
	}

	return token{Type: typeName, Value: value, Position: pos}
}

func (l *lexer) eof() token {
	return token{
		Type:     typeEOF,
		Position: l.current,
	}
}

func (l *lexer) error(typ ErrType, hint string) token {
	t := l.newToken(typeError)
	l.err = newErrorHint(typ, t, hint)
	return t
}

func (l *lexer) newToken(tt tokenType) token {
	t := token{
		Type:     tt,
		Value:    l.input[l.start:l.current],
		Position: l.start,
	}
	l.width = 0
	l.start = l.current
	return t
}

func (l *lexer) nextRune() rune {

	if l.err != nil || l.current >= l.length {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRuneInString(l.input[l.current:])
	l.width = w
	l.current += w
	/*
		if r == '\n' {
			l.line++
		}
	*/
	return r
}

func (l *lexer) backup() {
	l.current -= l.width
}

func (l *lexer) peek() rune {
	r := l.nextRune()
	l.backup()
	return r
}

func (l *lexer) ignore() {
	l.start = l.current
}

func (l *lexer) acceptRune(r rune) bool {
	return l.accept(func(c rune) bool {
		return c == r
	})
}

func (l *lexer) acceptRunes2(r1, r2 rune) bool {
	return l.accept(func(c rune) bool {
		return c == r1 || c == r2
	})
}

func (l *lexer) accept(isValid func(rune) bool) bool {
	if isValid(l.nextRune()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptAll(isValid func(rune) bool) bool {
	var b bool
	for l.accept(isValid) {
		b = true
	}
	return b
}

func (l *lexer) skipWhitespace() {
	for {
		ch := l.peek()
		if !isWhitespace(ch) {
			break
		}
		l.nextRune()
	}
	l.ignore()
}

func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v':
		return true
	default:
		return false
	}
}

func isRegexFlag(r rune) bool {
	return r == 'i' || r == 'I' || r == 'm' || r == 's'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBinaryDigit(r rune) bool {
	return r == '0' || r == '1'
}

func isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func (l *lexer) scanComment(allowRegex bool) token {
	startPos := l.start
	next := l.peek()
	l.nextRune() // consume first '/'

	if next == '*' {
		l.nextRune() // consume '*'
		depth := 1
		for depth > 0 {
			r := l.nextRune()
			if r == eof {
				return token{Type: typeError, Value: "unterminated comment (no closing */)", Position: startPos}
			}
			if r == '*' && l.peek() == '/' {
				l.nextRune() // consume '/'
				depth--
			} else if r == '/' && l.peek() == '*' {
				l.nextRune() // consume '*'
				depth++
			}
		}
	} else if next == '/' {
		l.nextRune() // consume second '/'
		for {
			r := l.nextRune()
			if r == '\n' || r == eof {
				break
			}
		}
	}

	l.start = startPos
	l.ignore()
	l.skipWhitespace()
	return l.next(allowRegex)
}

// symbolsAndKeywords maps operator token types back to their
// string representations. It's only used by tokenType.String
// (and one test).
var symbolsAndKeywords = func() map[tokenType]string {

	m := map[tokenType]string{
		typeAnd:  "and",
		typeOr:   "or",
		typeIn:   "in",
		typeNull: "null",
	}

	for r, tt := range symbols1 {
		if tt > 0 {
			m[tt] = fmt.Sprintf("%c", r)
		}
	}

	for r, rts := range symbols2 {
		for _, rt := range rts {
			m[rt.tt] = fmt.Sprintf("%c", r) + fmt.Sprintf("%c", rt.r)
		}
	}

	return m
}()
