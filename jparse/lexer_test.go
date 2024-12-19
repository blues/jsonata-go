// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jparse

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type lexerTestCase struct {
	Input      string
	AllowRegex bool
	Tokens     []token
	Error      error
}

func TestLexerWhitespace(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: "",
		},
		{
			Input: "       ",
		},
		{
			Input: "\v\t\r\n",
		},
		{
			Input: `


			`,
		},
	})
}

func TestLexerRegex(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: `//`,
			Tokens: []token{
				tok(typeDiv, "/", 0),
				tok(typeDiv, "/", 1),
			},
		},
		{
			Input:      `//`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "", 1),
			},
		},
		{
			Input:      `/ab+/`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "ab+", 1),
			},
		},
		{
			Input:      `/(ab+/)/`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "(ab+/)", 1),
			},
		},
		{
			Input:      `/ab+/i`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "(?i)ab+", 1),
			},
		},
		{
			Input:      `/ab+/ i`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "ab+", 1),
				tok(typeName, "i", 6),
			},
		},
		{
			Input:      `/ab+/I`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeRegex, "ab+", 1),
				tok(typeName, "I", 5),
			},
		},
		{
			Input:      `/ab+`,
			AllowRegex: true,
			Tokens: []token{
				tok(typeError, "ab+", 1),
			},
			Error: &Error{
				Type:     ErrUnterminatedRegex,
				Token:    "ab+",
				Hint:     "/",
				Position: 1,
			},
		},
	})
}

func TestLexerStrings(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: `""`,
			Tokens: []token{
				tok(typeString, "", 1),
			},
		},
		{
			Input: `''`,
			Tokens: []token{
				tok(typeString, "", 1),
			},
		},
		{
			Input: `"double quotes"`,
			Tokens: []token{
				tok(typeString, "double quotes", 1),
			},
		},
		{
			Input: "'single quotes'",
			Tokens: []token{
				tok(typeString, "single quotes", 1),
			},
		},
		{
			Input: `"escape\t"`,
			Tokens: []token{
				tok(typeString, "escape\\t", 1),
			},
		},
		{
			Input: `'escape\u0036'`,
			Tokens: []token{
				tok(typeString, "escape\\u0036", 1),
			},
		},
		{
			Input: `"超明體繁"`,
			Tokens: []token{
				tok(typeString, "超明體繁", 1),
			},
		},
		{
			Input: `'日本語'`,
			Tokens: []token{
				tok(typeString, "日本語", 1),
			},
		},
		{
			Input: `"No closing quote...`,
			Tokens: []token{
				tok(typeError, "No closing quote...", 1),
			},
			Error: &Error{
				Type:     ErrUnterminatedString,
				Token:    "No closing quote...",
				Hint:     "\", starting from character position 1",
				Position: 1,
			},
		},
		{
			Input: `'No closing quote...`,
			Tokens: []token{
				tok(typeError, "No closing quote...", 1),
			},
			Error: &Error{
				Type:     ErrUnterminatedString,
				Token:    "No closing quote...",
				Hint:     "', starting from character position 1",
				Position: 1,
			},
		},
	})
}

func TestLexerNumbers(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: "1",
			Tokens: []token{
				tok(typeNumber, "1", 0),
			},
		},
		{
			Input: "3.14159",
			Tokens: []token{
				tok(typeNumber, "3.14159", 0),
			},
		},
		{
			Input: "1e10",
			Tokens: []token{
				tok(typeNumber, "1e10", 0),
			},
		},
		{
			Input: "1E-10",
			Tokens: []token{
				tok(typeNumber, "1E-10", 0),
			},
		},
		{
			// Signs are separate tokens.
			Input: "-100",
			Tokens: []token{
				tok(typeMinus, "-", 0),
				tok(typeNumber, "100", 1),
			},
		},
		{
			// Leading zeroes are not supported.
			Input: "007",
			Tokens: []token{
				tok(typeNumber, "0", 0),
				tok(typeNumber, "0", 1),
				tok(typeNumber, "7", 2),
			},
		},
		{
			// Leading decimal points are not supported.
			Input: ".5",
			Tokens: []token{
				tok(typeDot, ".", 0),
				tok(typeNumber, "5", 1),
			},
		},
		{
			// Trailing decimal points are not supported.
			// TODO: Why does this require a character following the decimal point?
			Input: "5. ",
			Tokens: []token{
				tok(typeNumber, "5", 0),
				tok(typeDot, ".", 1),
			},
		},
	})
}

func TestLexerNames(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: "hello",
			Tokens: []token{
				tok(typeName, "hello", 0),
			},
		},
		{
			// Names break at whitespace...
			Input: "hello world",
			Tokens: []token{
				tok(typeName, "hello", 0),
				tok(typeName, "world", 6),
			},
		},
		{
			// ...and anything that looks like a symbol.
			Input: "hello, world.",
			Tokens: []token{
				tok(typeName, "hello", 0),
				tok(typeComma, ",", 5),
				tok(typeName, "world", 7),
				tok(typeDot, ".", 12),
			},
		},
		{
			// Exclamation marks are not symbols but the != operator
			// begins with one so it has the same effect on a name.
			Input: "HELLO!",
			Tokens: []token{
				tok(typeName, "HELLO", 0),
				tok(typeName, "!", 5),
			},
		},
		{
			// Escaped names can contain whitespace, symbols...
			Input: "`hello, world.`",
			Tokens: []token{
				tok(typeNameEsc, "hello, world.", 1),
			},
		},
		{
			// ...and keywords.
			Input: "`true or false`",
			Tokens: []token{
				tok(typeNameEsc, "true or false", 1),
			},
		},
		{
			Input: "`no closing quote...",
			Tokens: []token{
				tok(typeError, "no closing quote...", 1),
			},
			Error: &Error{
				Type:     ErrUnterminatedName,
				Token:    "no closing quote...",
				Hint:     "`",
				Position: 1,
			},
		},
	})
}

func TestLexerVariables(t *testing.T) {
	testLexer(t, []lexerTestCase{
		{
			Input: "$",
			Tokens: []token{
				tok(typeVariable, "", 1),
			},
		},
		{
			Input: "$$",
			Tokens: []token{
				tok(typeVariable, "$", 1),
			},
		},
		{
			Input: "$var",
			Tokens: []token{
				tok(typeVariable, "var", 1),
			},
		},
		{
			Input: "$uppercase",
			Tokens: []token{
				tok(typeVariable, "uppercase", 1),
			},
		},
	})
}

func TestLexerSymbolsAndKeywords(t *testing.T) {

	var tests []lexerTestCase

	for tt, s := range symbolsAndKeywords {
		tests = append(tests, lexerTestCase{
			Input: s,
			Tokens: []token{
				tok(tt, s, 0),
			},
		})
	}

	testLexer(t, tests)
}

func testLexer(t *testing.T, data []lexerTestCase) {

	for _, test := range data {

		l := newLexer(test.Input)
		eof := tok(typeEOF, "", len(test.Input))

		for _, exp := range test.Tokens {
			compareTokens(t, test.Input, exp, l.next(test.AllowRegex))
		}

		compareErrors(t, test.Input, test.Error, l.err)

		// The lexer should keep returning EOF after exhausting
		// the input. Call next() a few times to make sure that
		// repeated calls return EOF as expected.
		for i := 0; i < 3; i++ {
			compareTokens(t, test.Input, eof, l.next(test.AllowRegex))
		}
	}
}

func compareTokens(t *testing.T, prefix string, exp, got token) {

	if got.Type != exp.Type {
		t.Errorf("%s: expected token with Type '%s', got '%s'", prefix, exp.Type, got.Type)
	}

	if got.Value != exp.Value {
		t.Errorf("%s: expected token with Value %q, got %q", prefix, exp.Value, got.Value)
	}

	if got.Position != exp.Position {
		t.Errorf("%s: expected token with Position %d, got %d", prefix, exp.Position, got.Position)
	}
}

func compareErrors(t *testing.T, prefix string, exp, got error) {
	if exp != nil && got != nil {
		assert.EqualError(t, exp, fmt.Sprintf("%v", got))
	}

}

func tok(typ tokenType, value string, position int) token {
	return token{
		Type:     typ,
		Value:    value,
		Position: position,
	}
}
