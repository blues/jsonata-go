// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jparse

import (
	"fmt"
	"regexp"
)

// ErrType describes the type of an error.
type ErrType uint

// Error types returned by the parser.
const (
	_ ErrType = iota
	ErrSyntaxError
	ErrUnexpectedEOF
	ErrUnexpectedToken
	ErrMissingToken
	ErrPrefix
	ErrInfix
	ErrUnterminatedString
	ErrUnterminatedRegex
	ErrUnterminatedName
	ErrIllegalEscape
	ErrIllegalEscapeHex
	ErrInvalidNumber
	ErrNumberRange
	ErrEmptyRegex
	ErrInvalidRegex
	ErrGroupPredicate
	ErrGroupGroup
	ErrPathLiteral
	ErrIllegalAssignment
	ErrIllegalParam
	ErrDuplicateParam
	ErrParamCount
	ErrInvalidUnionType
	ErrUnmatchedOption
	ErrUnmatchedSubtype
	ErrInvalidSubtype
	ErrInvalidParamType
)

var errmsgs = map[ErrType]string{
	ErrSyntaxError:        "syntax error: '{{token}}', position: {{position}}",
	ErrUnexpectedEOF:      "unexpected end of expression, position: {{position}}",
	ErrUnexpectedToken:    "expected token '{{hint}}', got '{{token}}', position: {{position}}",
	ErrMissingToken:       "expected token '{{hint}}' before end of expression, position: {{position}}",
	ErrPrefix:             "the symbol '{{token}}' cannot be used as a prefix operator, position: {{position}}",
	ErrInfix:              "the symbol '{{token}}' cannot be used as an infix operator, position: {{position}}",
	ErrUnterminatedString: "unterminated string literal (no closing '{{hint}}'), position: {{position}}",
	ErrUnterminatedRegex:  "unterminated regular expression (no closing '{{hint}}'), position: {{position}}",
	ErrUnterminatedName:   "unterminated name (no closing '{{hint}}'), position: {{position}}",
	ErrIllegalEscape:      "illegal escape sequence \\{{hint}}, position: {{position}}",
	ErrIllegalEscapeHex:   "illegal escape sequence \\{{hint}}: \\u must be followed by a 4-digit hexadecimal code point, position: {{position}}",
	ErrInvalidNumber:      "invalid number literal {{token}}, {{position}}, position: {{position}}",
	ErrNumberRange:        "invalid number literal {{token}}: value out of range, position: {{position}}",
	ErrEmptyRegex:         "invalid regular expression: expression cannot be empty, position: {{position}}",
	ErrInvalidRegex:       "invalid regular expression {{token}}: {{hint}}, position: {{position}}",
	ErrGroupPredicate:     "a predicate cannot follow a grouping expression in a path step, position: {{position}}",
	ErrGroupGroup:         "a path step can only have one grouping expression, position: {{position}}",
	ErrPathLiteral:        "invalid path step {{hint}}: paths cannot contain nulls, strings, numbers or booleans, position: {{position}}",
	ErrIllegalAssignment:  "illegal assignment: {{hint}} is not a variable, position: {{position}}",
	ErrIllegalParam:       "illegal function parameter: {{token}} is not a variable, position: {{position}}",
	ErrDuplicateParam:     "duplicate function parameter: {{token}}, position: {{position}}",
	ErrParamCount:         "invalid type signature: number of types must match number of function parameters, position: {{position}}",
	ErrInvalidUnionType:   "invalid type signature: unsupported union type '{{hint}}', position: {{position}}",
	ErrUnmatchedOption:    "invalid type signature: option '{{hint}}' must follow a parameter, position: {{position}}",
	ErrUnmatchedSubtype:   "invalid type signature: subtypes must follow a parameter, position: {{position}}",
	ErrInvalidSubtype:     "invalid type signature: parameter type {{hint}} does not support subtypes, position: {{position}}",
	ErrInvalidParamType:   "invalid type signature: unknown parameter type '{{hint}}', position: {{position}}",
}

var reErrMsg = regexp.MustCompile("{{(token|hint|position)}}")

// Error describes an error during parsing.
type Error struct {
	Type     ErrType
	Token    string
	Hint     string
	Position int
}

func newError(typ ErrType, tok token) error {
	return newErrorHint(typ, tok, "")

}

func newErrorHint(typ ErrType, tok token, hint string) error {
	return &Error{
		Type:     typ,
		Token:    tok.Value,
		Position: tok.Position,
		Hint:     hint,
	}
}

func (e Error) Error() string {

	s := errmsgs[e.Type]
	if s == "" {
		return fmt.Sprintf("parser.Error: unknown error type %d", e.Type)
	}

	return reErrMsg.ReplaceAllStringFunc(s, func(match string) string {
		switch match {
		case "{{token}}":
			return e.Token
		case "{{hint}}":
			return e.Hint
		case "{{position}}":
			return fmt.Sprintf("%v", e.Position)
		default:
			return match
		}
	})
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
