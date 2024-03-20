// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jparse_test

import (
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/xiatechs/jsonata-go/jparse"
)

type testCase struct {
	Input  string
	Inputs []string
	Output jparse.Node
	Error  error
}

func TestStringNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  `""`,
			Output: &jparse.StringNode{},
		},
		{
			Inputs: []string{
				`"hello"`,
				`"\u0068\u0065\u006c\u006c\u006f"`,
			},
			Output: &jparse.StringNode{
				Value: "hello",
			},
		},
		{
			Input: `"hello\t"`,
			Output: &jparse.StringNode{
				Value: "hello\t",
			},
		},
		{
			Input: `"C:\\\\windows\\temp"`,
			Output: &jparse.StringNode{
				Value: "C:\\\\windows\\temp",
			},
		},
		{
			// Non-ASCII UTF-8
			Inputs: []string{
				`"hello 超明體繁"`,
				`"hello \u8d85\u660e\u9ad4\u7e41"`,
			},
			Output: &jparse.StringNode{
				Value: "hello 超明體繁",
			},
		},
		{
			// UTF-16 surrogate pair
			Inputs: []string{
				`"😂 emoji"`,
				`"\ud83d\ude02 emoji"`,
			},
			Output: &jparse.StringNode{
				Value: "😂 emoji",
			},
		},
		{
			// Invalid escape sequence
			Input: `"hello\x"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 1,
				Token:    "hello\\x",
				Hint:     "x",
			},
		},
		{
			// Valid escape sequence followed by invalid escape sequence
			Input: `"\r\x"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 1,
				Token:    "\\r\\x",
				Hint:     "x",
			},
		},
		{
			// Missing hexadecimal sequence
			Input: `"hello\u"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "hello\\u",
				Hint:     "u" + strings.Repeat(string(utf8.RuneError), 4),
			},
		},
		{
			// Invalid hexadecimal sequence
			Input: `"hello\u123t"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "hello\\u123t",
				Hint:     "u123t",
			},
		},
		{
			// Invalid hexadecimal sequence
			Input: `"hello\uworld"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "hello\\uworld",
				Hint:     "uworl",
			},
		},
		{
			// Incomplete UTF-16 surrogate pair
			Input: `"\ud83d"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\ud83d",
				Hint:     "ud83d",
			},
		},
		{
			// Incomplete trailing surrogate
			Input: `"\ud83d\u"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\ud83d\\u",
				Hint:     "ud83d",
			},
		},
		{
			// Trailing surrogate outside allowed range
			Input: `"\ud83d\u0068"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\ud83d\\u0068",
				Hint:     "ud83d",
			},
		},
		{
			// Invalid hexadecimal trailing surrogate
			Input: `"\ud83d\u123t"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\ud83d\\u123t",
				Hint:     "ud83d",
			},
		},
		{
			// Reversed UTF-16 surrogate pair
			Input: `"\ude02\ud83d"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\ude02\\ud83d",
				Hint:     "ude02",
			},
		},
		{
			// Non-terminated string
			Input: `"hello`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnterminatedString,
				Position: 1,
				Token:    "hello",
				Hint:     "\", starting from character position 1",
			},
		},
		{
			// Non-terminated string
			Input: `'world`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnterminatedString,
				Position: 1,
				Token:    "world",
				Hint:     "', starting from character position 1",
			},
		},
	})
}

func TestNumberNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "0",
			Output: &jparse.NumberNode{},
		},
		{
			Input: "100",
			Output: &jparse.NumberNode{
				Value: 100,
			},
		},
		{
			Input: "-0.5",
			Output: &jparse.NumberNode{
				Value: -0.5,
			},
		},
		{
			// invalid syntax
			Input: "1e+",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 0,
				Token:    "1e+",
			},
		},
		{
			// overflow
			Input: "1e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Position: 0,
				Token:    "1e1000",
			},
		},
	})
}

func TestBooleanNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "false",
			Output: &jparse.BooleanNode{},
		},
		{
			Input: "true",
			Output: &jparse.BooleanNode{
				Value: true,
			},
		},
	})
}

func TestNull(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "null",
			Output: &jparse.NullNode{},
		},
	})
}

func TestRegexNode(t *testing.T) {

	// well-formed regular expressions
	good := map[string]string{
		`/\s+/`:           `\s+`,
		`/ab+/`:           `ab+`,
		`/(ab)+/`:         `(ab)+`,
		`/(ab)+/i`:        `(?i)(ab)+`,
		`/(ab)+/m`:        `(?m)(ab)+`,
		`/(ab)+/s`:        `(?s)(ab)+`,
		`/^[1-9][0-9]*$/`: `^[1-9][0-9]*$`,
	}

	// malformed regular expressions
	bad := []string{
		`?`,          // repetition operator with no operand
		`\C+`,        // invalid escape sequence
		`[9-0]`,      // invalid character class range
		`[a-z]{1,0}`, // invalid repeat count
	}

	var data []testCase

	found := false
	for input, expr := range good {

		re, err := regexp.Compile(expr)
		if err != nil {
			t.Logf("Good regex %q does not compile. Skipping test...", expr)
			continue
		}

		found = true
		data = append(data, testCase{
			Input: input,
			Output: &jparse.RegexNode{
				Value: re,
			},
		})
	}

	if !found {
		t.Errorf("No compiling regexes found")
	}

	found = false
	for _, expr := range bad {

		_, err := regexp.Compile(expr)
		if err == nil {
			t.Logf("Bad regex %q compiles. Skipping test...", expr)
			continue
		}

		e, ok := err.(*syntax.Error)
		if !ok {
			t.Logf("Bad regex %q throws a %T (expected a *syntax.Error). Skipping test...", expr, err)
			continue
		}

		found = true
		data = append(data, testCase{
			Input: "/" + expr + "/",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidRegex,
				Position: 1,
				Token:    expr,
				Hint:     string(e.Code),
			},
		})
	}

	if !found {
		t.Errorf("No non-compiling regexes found")
	}

	data = append(data, testCase{
		Input: "//",
		Error: &jparse.Error{
			Type:     jparse.ErrEmptyRegex,
			Position: 1,
		},
	})

	testParser(t, data)
}

func TestVariableNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "$",
			Output: &jparse.VariableNode{},
		},
		{
			Input: "$$",
			Output: &jparse.VariableNode{
				Name: "$",
			},
		},
		{
			Input: "$x",
			Output: &jparse.VariableNode{
				Name: "x",
			},
		},
	})
}

func TestAssignmentNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `$greeting := "hello"`,
			Output: &jparse.AssignmentNode{
				Name: "greeting",
				Value: &jparse.StringNode{
					Value: "hello",
				},
			},
		},
		{
			Input: "$trimlower := $trim ~> $lowercase",
			Output: &jparse.AssignmentNode{
				Name: "trimlower",
				Value: &jparse.FunctionApplicationNode{
					LHS: &jparse.VariableNode{
						Name: "trim",
					},
					RHS: &jparse.VariableNode{
						Name: "lowercase",
					},
				},
			},
		},
		{
			Input: "$bignum := 1e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Position: 11,
				Token:    "1e1000",
			},
		},
		{
			Input: "1 := 100",
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalAssignment,
				Position: 2,
				Token:    ":=",
				Hint:     "1",
			},
		},
	})
}

func TestNegationNode(t *testing.T) {
	testParser(t, []testCase{
		{
			// The parser handles negation of number literals.
			Input: "-1",
			Output: &jparse.NumberNode{
				Value: -1,
			},
		},
		{
			Input: "--0.5",
			Output: &jparse.NumberNode{
				Value: 0.5,
			},
		},
		{
			Input: "-$i",
			Output: &jparse.NegationNode{
				RHS: &jparse.VariableNode{
					Name: "i",
				},
			},
		},
		{
			Input: "-1e",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 1,
				Token:    "1e",
			},
		},
		{
			Input: "-",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 1,
			},
		},
	})
}

func TestBlockNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "()",
			Output: &jparse.BlockNode{},
		},
		{
			Inputs: []string{
				"(1; 2; 3)",
				"(1; 2; 3;)", // allow trailing semicolons
			},
			Output: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.NumberNode{
						Value: 2,
					},
					&jparse.NumberNode{
						Value: 3,
					},
				},
			},
		},
		{
			// Invalid value.
			Input: "(1; 2; 3e)",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 7,
				Token:    "3e",
			},
		},
		{
			// Invalid delimiters.
			Input: "(1, 2, 3)",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedToken,
				Position: 2,
				Token:    ",",
				Hint:     ")",
			},
		},
		{
			// No closing bracket.
			Input: "(1; 2; 3",
			Error: &jparse.Error{
				Type:     jparse.ErrMissingToken,
				Position: 8,
				Hint:     ")",
			},
		},
	})
}

func TestWildcardNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "*",
			Output: &jparse.WildcardNode{},
		},
		{
			Input: "*Field",
			Error: &jparse.Error{
				Type:     jparse.ErrSyntaxError,
				Position: 1,
				Token:    "Field",
			},
		},
	})
}

func TestDescendentNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "**",
			Output: &jparse.DescendentNode{},
		},
		{
			Input: "**Field",
			Error: &jparse.Error{
				Type:     jparse.ErrSyntaxError,
				Position: 2,
				Token:    "Field",
			},
		},
	})
}

func TestObjectTransformationNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `|$|{"one":1}|`,
			Output: &jparse.ObjectTransformationNode{
				Pattern: &jparse.VariableNode{},
				Updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "one",
							},
							&jparse.NumberNode{
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			Input: `|$|{}, ["key1", "key2"]|`,
			Output: &jparse.ObjectTransformationNode{
				Pattern: &jparse.VariableNode{},
				Updates: &jparse.ObjectNode{},
				Deletes: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "key1",
						},
						&jparse.StringNode{
							Value: "key2",
						},
					},
				},
			},
		},
		{
			Input: `|$|{"one":1}, ["key1", "key2"]|`,
			Output: &jparse.ObjectTransformationNode{
				Pattern: &jparse.VariableNode{},
				Updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "one",
							},
							&jparse.NumberNode{
								Value: 1,
							},
						},
					},
				},
				Deletes: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "key1",
						},
						&jparse.StringNode{
							Value: "key2",
						},
					},
				},
			},
		},
		{
			// Bad pattern
			Input: "|?|{}|",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 1,
				Token:    "?",
			},
		},
		{
			// Bad update map
			Input: `|$|{"one":}|`,
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 10,
				Token:    "}",
			},
		},
		{
			// Bad deletion array
			Input: `|$|{},["key\1"]|`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 8,
				Token:    "key\\1",
				Hint:     "1",
			},
		},
	})
}

func TestFunctionCallNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "$random()",
			Output: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "random",
				},
			},
		},
		{
			Input: `$uppercase("hello")`,
			Output: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "uppercase",
				},
				Args: []jparse.Node{
					&jparse.StringNode{
						Value: "hello",
					},
				},
			},
		},
		{
			Input: `$substring("hello", -3, 2)`,
			Output: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "substring",
				},
				Args: []jparse.Node{
					&jparse.StringNode{
						Value: "hello",
					},
					&jparse.NumberNode{
						Value: -3,
					},
					&jparse.NumberNode{
						Value: 2,
					},
				},
			},
		},
		{
			// Trailing delimiter.
			Input: `$trim("hello   ",)`,
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    ")",
				Position: 17,
			},
		},
	})
}

func TestLambdaNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "function(){0}",
			Output: &jparse.LambdaNode{
				Body: &jparse.NumberNode{
					Value: 0,
				},
			},
		},
		{
			Input: "function($w, $h){$w * $h}",
			Output: &jparse.LambdaNode{
				ParamNames: []string{
					"w",
					"h",
				},
				Body: &jparse.NumericOperatorNode{
					Type: jparse.NumericMultiply,
					LHS: &jparse.VariableNode{
						Name: "w",
					},
					RHS: &jparse.VariableNode{
						Name: "h",
					},
				},
			},
		},
		{
			// No function body.
			Input: "function()",
			Error: &jparse.Error{
				Type:     jparse.ErrMissingToken,
				Position: 10,
				Hint:     "{",
			},
		},
		{
			// Empty function body.
			Input: "function(){}",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 11,
				Token:    "}",
			},
		},
		{
			// Invalid function body.
			Input: "function(){1e}",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 11,
				Token:    "1e",
			},
		},
		{
			// Illegal parameter.
			Input: "function($w, h){$w * $h}",
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalParam,
				Position: 13,
				Token:    "h",
			},
		},
		{
			// Duplicate parameter.
			Input: "function($x, $x){$x}",
			Error: &jparse.Error{
				Type:     jparse.ErrDuplicateParam,
				Position: 14,
				Token:    "x",
			},
		},
		{
			// Lambdas cannot be partials.
			Input: "function(?, 10){0}",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 9,
				Token:    "?",
			},
		},
		{
			// Trailing delimiter.
			Input: "function($x, $y,){0}",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 16,
				Token:    ")",
			},
		},
	})
}

func TestTypedLambdaNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "function($x, $y)<nn?:n>{0}",
			Output: &jparse.TypedLambdaNode{
				LambdaNode: &jparse.LambdaNode{
					ParamNames: []string{
						"x",
						"y",
					},
					Body: &jparse.NumberNode{
						Value: 0,
					},
				},
				In: []jparse.Param{
					{
						Type: jparse.ParamTypeNumber,
					},
					{
						Type:   jparse.ParamTypeNumber,
						Option: jparse.ParamOptional,
					},
				},
			},
		},
		{
			Input: "function($arr)<a<(ns)>-:a>{[]}",
			Output: &jparse.TypedLambdaNode{
				LambdaNode: &jparse.LambdaNode{
					ParamNames: []string{
						"arr",
					},
					Body: &jparse.ArrayNode{},
				},
				In: []jparse.Param{
					{
						Type:   jparse.ParamTypeArray,
						Option: jparse.ParamContextable,
						SubParams: []jparse.Param{
							{
								Type: jparse.ParamTypeNumber | jparse.ParamTypeString,
							},
						},
					},
				},
			},
		},
		{
			// Mismatched parameter/signature count.
			Input: "λ($x, $y)<nnn?:n>{0}",
			Error: &jparse.Error{
				Type:     jparse.ErrParamCount,
				Token:    "{",
				Position: 18,
			},
		},
		{
			// Unknown parameter type in signature.
			Input: "λ($x, $y)<nz?:n>{0}",
			Error: &jparse.Error{
				// TODO: Add position info.
				Type: jparse.ErrInvalidParamType,
				Hint: "z",
			},
		},
		{
			// Unknown parameter type in union type.
			Input: "λ($x, $y)<n(y)?:n>{0}",
			Error: &jparse.Error{
				// TODO: Add position info.
				Type: jparse.ErrInvalidUnionType,
				Hint: "y",
			},
		},
		{
			// Option without a parameter.
			Input: "λ($x, $y)<+nn?:n>{0}",
			Error: &jparse.Error{
				// TODO: Add position info.
				Type: jparse.ErrUnmatchedOption,
				Hint: "+",
			},
		},
		{
			// Subtype without a parameter.
			Input: "λ($x, $y)<<x>nn?:n>{0}",
			Error: &jparse.Error{
				// TODO: Add position info.
				Type: jparse.ErrUnmatchedSubtype,
			},
		},
		{
			// Subtype on a non-array, non-function parameter.
			Input: "λ($x, $y)<n<x>n?:n>{0}",
			Error: &jparse.Error{
				// TODO: Add position info.
				Type: jparse.ErrInvalidSubtype,
				Hint: "n",
			},
		},
	})
}

func TestPartialApplicationNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "$substring(?, 0, ?)",
			Output: &jparse.PartialNode{
				Func: &jparse.VariableNode{
					Name: "substring",
				},
				Args: []jparse.Node{
					&jparse.PlaceholderNode{},
					&jparse.NumberNode{},
					&jparse.PlaceholderNode{},
				},
			},
		},
	})
}

func TestFunctionApplicationNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `$trim ~> $uppercase`,
			Output: &jparse.FunctionApplicationNode{
				LHS: &jparse.VariableNode{
					Name: "trim",
				},
				RHS: &jparse.VariableNode{
					Name: "uppercase",
				},
			},
		},
		{
			Input: `"hello world" ~> $substringBefore(" ")`,
			Output: &jparse.FunctionApplicationNode{
				LHS: &jparse.StringNode{
					Value: "hello world",
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.VariableNode{
						Name: "substringBefore",
					},
					Args: []jparse.Node{
						&jparse.StringNode{
							Value: " ",
						},
					},
				},
			},
		},
		{
			Input: `"hello world" ~> $substringBefore("\x")`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 35,
				Token:    "\\x",
				Hint:     "x",
			},
		},
		{
			Input: `"hello world" ~>`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 16,
			},
		},
	})
}

func TestPredicateNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "$",
			Output: &jparse.VariableNode{},
		},
		{
			Input: "$[-1]",
			Output: &jparse.PredicateNode{
				Expr: &jparse.VariableNode{},
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: -1,
					},
				},
			},
		},
		{
			Input: "$[-1][0]",
			Output: &jparse.PredicateNode{
				Expr: &jparse.PredicateNode{
					Expr: &jparse.VariableNode{},
					Filters: []jparse.Node{
						&jparse.NumberNode{
							Value: -1,
						},
					},
				},
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: 0,
					},
				},
			},
		},
		{
			Input: "$[-1][0][]",
			Output: &jparse.PathNode{
				KeepArrays: true,
				Steps: []jparse.Node{
					&jparse.PredicateNode{
						Expr: &jparse.PredicateNode{
							Expr: &jparse.VariableNode{},
							Filters: []jparse.Node{
								&jparse.NumberNode{
									Value: -1,
								},
							},
						},
						Filters: []jparse.Node{
							&jparse.NumberNode{
								Value: 0,
							},
						},
					},
				},
			},
		},
		{
			Input: "path",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "path",
					},
				},
			},
		},
		{
			Input: `path[type="home"]`,
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.PredicateNode{
						Expr: &jparse.NameNode{
							Value: "path",
						},
						Filters: []jparse.Node{
							&jparse.ComparisonOperatorNode{
								Type: jparse.ComparisonEqual,
								LHS: &jparse.PathNode{
									Steps: []jparse.Node{
										&jparse.NameNode{
											Value: "type",
										},
									},
								},
								RHS: &jparse.StringNode{
									Value: "home",
								},
							},
						},
					},
				},
			},
		},
		{
			Input: `path[type="home"][0]`,
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.PredicateNode{
						Expr: &jparse.NameNode{
							Value: "path",
						},
						Filters: []jparse.Node{
							&jparse.ComparisonOperatorNode{
								Type: jparse.ComparisonEqual,
								LHS: &jparse.PathNode{
									Steps: []jparse.Node{
										&jparse.NameNode{
											Value: "type",
										},
									},
								},
								RHS: &jparse.StringNode{
									Value: "home",
								},
							},
							&jparse.NumberNode{
								Value: 0,
							},
						},
					},
				},
			},
		},
		{
			Input: `path[type="home"][0][]`,
			Output: &jparse.PathNode{
				KeepArrays: true,
				Steps: []jparse.Node{
					&jparse.PredicateNode{
						Expr: &jparse.NameNode{
							Value: "path",
						},
						Filters: []jparse.Node{
							&jparse.ComparisonOperatorNode{
								Type: jparse.ComparisonEqual,
								LHS: &jparse.PathNode{
									Steps: []jparse.Node{
										&jparse.NameNode{
											Value: "type",
										},
									},
								},
								RHS: &jparse.StringNode{
									Value: "home",
								},
							},
							&jparse.NumberNode{
								Value: 0,
							},
						},
					},
				},
			},
		},
		{
			Input: `path[price<=1e]`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 12,
				Token:    "1e",
			},
		},
		{
			Input: `path{"one": 1}[0]`,
			Error: &jparse.Error{
				// TODO: Get position.
				Type: jparse.ErrGroupPredicate,
			},
		},
	})
}

func TestConditionalNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `true ? "yes"`,
			Output: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: true,
				},
				Then: &jparse.StringNode{
					Value: "yes",
				},
			},
		},
		{
			Input: `true ? "yes" : "no"`,
			Output: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: true,
				},
				Then: &jparse.StringNode{
					Value: "yes",
				},
				Else: &jparse.StringNode{
					Value: "no",
				},
			},
		},
		{
			// Missing truthy expression.
			Input: `true ?`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 6,
			},
		},
		{
			// Bad truthy expression.
			Input: `true ? 1e`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 7,
				Token:    "1e",
			},
		},
		{
			// Missing falsy expression.
			Input: `true ? 1e10 :`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 13,
			},
		},
		{
			// Bad falsy expression.
			Input: `true ? 1e10 : 1e`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 14,
				Token:    "1e",
			},
		},
	})
}

func TestArrayNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "[]",
			Output: &jparse.ArrayNode{},
		},
		{
			Input: "[1,2,3]",
			Output: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.NumberNode{
						Value: 2,
					},
					&jparse.NumberNode{
						Value: 3,
					},
				},
			},
		},
		{
			Input: "[1..3]",
			Output: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.RangeNode{
						LHS: &jparse.NumberNode{
							Value: 1,
						},
						RHS: &jparse.NumberNode{
							Value: 3,
						},
					},
				},
			},
		},
		{
			Input: "[-2..2,3,4,5]",
			Output: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.RangeNode{
						LHS: &jparse.NumberNode{
							Value: -2,
						},
						RHS: &jparse.NumberNode{
							Value: 2,
						},
					},
					&jparse.NumberNode{
						Value: 3,
					},
					&jparse.NumberNode{
						Value: 4,
					},
					&jparse.NumberNode{
						Value: 5,
					},
				},
			},
		},
		{
			Input: `[0, "", false, null, [], {}]`,
			Output: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{},
					&jparse.StringNode{},
					&jparse.BooleanNode{},
					&jparse.NullNode{},
					&jparse.ArrayNode{},
					&jparse.ObjectNode{},
				},
			},
		},
		{
			// Invalid array member.
			Input: `[1, 2, 3e]`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 7,
				Token:    "3e",
			},
		},
		{
			// Invalid range operand.
			Input: `[1..1e]`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 4,
				Token:    "1e",
			},
		},
		{
			// Invalid delimiters.
			Input: `[1; 2; 3]`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedToken,
				Position: 2,
				Token:    ";",
				Hint:     "]",
			},
		},
		{
			// Trailing delimiter.
			Input: `[1, 2, 3,]`,
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 9,
				Token:    "]",
			},
		},
	})
}

func TestObjectNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input:  "{}",
			Output: &jparse.ObjectNode{},
		},
		{
			Input: `{"one": 1, "two": 2, "three": 3}`,
			Output: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.StringNode{
							Value: "one",
						},
						&jparse.NumberNode{
							Value: 1,
						},
					},
					{
						&jparse.StringNode{
							Value: "two",
						},
						&jparse.NumberNode{
							Value: 2,
						},
					},
					{
						&jparse.StringNode{
							Value: "three",
						},
						&jparse.NumberNode{
							Value: 3,
						},
					},
				},
			},
		},
		{
			// Invalid key.
			Input: `{"on\e": 1}`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 2,
				Token:    "on\\e",
				Hint:     "e",
			},
		},
		{
			// Invalid value.
			Input: `{"one": 1e}`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 8,
				Token:    "1e",
			},
		},
		{
			// Invalid delimiter.
			Input: `{"one"; 1}`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedToken,
				Position: 6,
				Token:    ";",
				Hint:     ":",
			},
		},
		{
			// Invalid delimiters.
			Input: `{"one": 1; "two": 2; "three": 3}`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedToken,
				Position: 9,
				Token:    ";",
				Hint:     "}",
			},
		},
		{
			// Trailing delimiter.
			Input: `{"one": 1,}`,
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 10,
				Token:    "}",
			},
		},
	})
}

func TestGroupNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `*{"one": 1}`,
			Output: &jparse.GroupNode{
				Expr: &jparse.WildcardNode{},
				ObjectNode: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "one",
							},
							&jparse.NumberNode{
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			// Invalid value.
			Input: `*{"one": 1e}`,
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 9,
				Token:    "1e",
			},
		},
		{
			Input: `*{"one": 1}{"two": 2}`,
			Error: &jparse.Error{
				// TODO: Get position.
				Type: jparse.ErrGroupGroup,
			},
		},
	})
}

func TestNumericOperatorNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "3.5 + 1",
			Output: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.NumberNode{
					Value: 3.5,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
		},
		{
			Input: "3.5 - 1",
			Output: &jparse.NumericOperatorNode{
				Type: jparse.NumericSubtract,
				LHS: &jparse.NumberNode{
					Value: 3.5,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
		},
		{
			Input: "3.5 * 1",
			Output: &jparse.NumericOperatorNode{
				Type: jparse.NumericMultiply,
				LHS: &jparse.NumberNode{
					Value: 3.5,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
		},
		{
			Input: "3.5 / 1",
			Output: &jparse.NumericOperatorNode{
				Type: jparse.NumericDivide,
				LHS: &jparse.NumberNode{
					Value: 3.5,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
		},
		{
			Input: "3.5 % 1",
			Output: &jparse.NumericOperatorNode{
				Type: jparse.NumericModulo,
				LHS: &jparse.NumberNode{
					Value: 3.5,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
		},
		{
			Input: "3.5e * 1",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 0,
				Token:    "3.5e",
			},
		},
		{
			Input: "3.5 * 1e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Position: 6,
				Token:    "1e1000",
			},
		},
		{
			Input: "+",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "+",
				Position: 0,
			},
		},
		{
			Input: "-",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 1,
			},
		},
		{
			Input:  "*",
			Output: &jparse.WildcardNode{},
		},
		{
			Input: "/",
			Error: &jparse.Error{
				Type:     jparse.ErrUnterminatedRegex,
				Hint:     "/",
				Position: 1,
			},
		},
		{
			Input: "%",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "%",
				Position: 0,
			},
		},
	})
}

func TestComparisonOperatorNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "1 = 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 != 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonNotEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 > 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreater,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 >= 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreaterEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 < 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 <= 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLessEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			Input: "1 in 2",
			Output: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
		},
		{
			// Invalid left hand side.
			Input: "1e = 1",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 0,
				Token:    "1e",
			},
		},
		{
			// Missing right hand side.
			Input: "1 =",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 3,
			},
		},
		{
			// Invalid right hand side.
			Input: "1 = 1e",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 4,
				Token:    "1e",
			},
		},
		{
			Input: "=",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "=",
				Position: 0,
			},
		},
		{
			Input: "!=",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "!=",
				Position: 0,
			},
		},
		{
			Input: ">",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    ">",
				Position: 0,
			},
		},
		{
			Input: ">=",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    ">=",
				Position: 0,
			},
		},
		{
			Input: "<",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "<",
				Position: 0,
			},
		},
		{
			Input: "<=",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Token:    "<=",
				Position: 0,
			},
		},
		{
			// Treat "in" like a name when it appears as a prefix.
			Input: "in",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "in",
					},
				},
			},
		},
	})
}

func TestBooleanOperatorNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "true and false",
			Output: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
		},
		{
			Input: "true or false",
			Output: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
		},
		{
			// Treat "and" like a name when it appears as a prefix.
			Input: "and",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "and",
					},
				},
			},
		},
		{
			// Treat "or" like a name when it appears as a prefix.
			Input: "or",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "or",
					},
				},
			},
		},
		{
			// Invalid left hand side.
			Input: "1e or 1",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Position: 0,
				Token:    "1e",
			},
		},
		{
			// Missing right hand side.
			Input: "true and",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 8,
			},
		},
		{
			// Invalid right hand side.
			Input: "1 and 1e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Position: 6,
				Token:    "1e1000",
			},
		},
	})
}

func TestConcatenationNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: `"hello" & "world"`,
			Output: &jparse.StringConcatenationNode{
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
		},
		{
			Input: `firstName & " " & lastName`,
			Output: &jparse.StringConcatenationNode{
				LHS: &jparse.StringConcatenationNode{
					LHS: &jparse.PathNode{
						Steps: []jparse.Node{
							&jparse.NameNode{
								Value: "firstName",
							},
						},
					},
					RHS: &jparse.StringNode{
						Value: " ",
					},
				},
				RHS: &jparse.PathNode{
					Steps: []jparse.Node{
						&jparse.NameNode{
							Value: "lastName",
						},
					},
				},
			},
		},
		{
			// bad left hand side.
			Input: `"\u000z" & " escape"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\u000z",
				Hint:     "u000z",
			},
		},
		{
			// missing right hand side.
			Input: `"hello" &`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 9,
			},
		},
		{
			// bad right hand side.
			Input: `"escape" & " this\x"`,
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 12,
				Token:    " this\\x",
				Hint:     "x",
			},
		},
	})
}

func TestSortNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "$^(path)",
			Output: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDefault,
						Expr: &jparse.PathNode{
							Steps: []jparse.Node{
								&jparse.NameNode{
									Value: "path",
								},
							},
						},
					},
				},
			},
		},
		{
			Input: "$^(<first, >second)",
			Output: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortAscending,
						Expr: &jparse.PathNode{
							Steps: []jparse.Node{
								&jparse.NameNode{
									Value: "first",
								},
							},
						},
					},
					{
						Dir: jparse.SortDescending,
						Expr: &jparse.PathNode{
							Steps: []jparse.Node{
								&jparse.NameNode{
									Value: "second",
								},
							},
						},
					},
				},
			},
		},
		{
			// Missing sort terms.
			Input: "$^",
			Error: &jparse.Error{
				Type:     jparse.ErrMissingToken,
				Position: 2,
				Hint:     "(",
			},
		},
		{
			// Empty sort terms.
			Input: "$^()",
			Error: &jparse.Error{
				Type:     jparse.ErrPrefix,
				Position: 3,
				Token:    ")",
			},
		},
	})
}

func TestPathNode(t *testing.T) {
	testParser(t, []testCase{
		{
			Input: "path",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "path",
					},
				},
			},
		},
		{
			Input: "path[]",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.NameNode{
						Value: "path",
					},
				},
				KeepArrays: true,
			},
		},
		{
			Input: "path[0]",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.PredicateNode{
						Expr: &jparse.NameNode{
							Value: "path",
						},
						Filters: []jparse.Node{
							&jparse.NumberNode{
								Value: 0,
							},
						},
					},
				},
			},
		},
		{
			Input: "$.path",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.VariableNode{},
					&jparse.NameNode{
						Value: "path",
					},
				},
			},
		},
		{
			Input: "$.path.$uppercase()",
			Output: &jparse.PathNode{
				Steps: []jparse.Node{
					&jparse.VariableNode{},
					&jparse.NameNode{
						Value: "path",
					},
					&jparse.FunctionCallNode{
						Func: &jparse.VariableNode{
							Name: "uppercase",
						},
					},
				},
			},
		},
		{
			// Incomplete path.
			Input: "path.",
			Error: &jparse.Error{
				Type:     jparse.ErrUnexpectedEOF,
				Position: 5,
			},
		},
		{
			// Literal on rhs of dot operator.
			Input: "path.0",
			Error: &jparse.Error{
				// TODO: Need position info.
				Type: jparse.ErrPathLiteral,
				Hint: "0",
			},
		},
		{
			// Literal on lhs of dot operator.
			Input: `"Product Name".$uppercase()`,
			Error: &jparse.Error{
				// TODO: Need position info.
				Type: jparse.ErrPathLiteral,
				Hint: `"Product Name"`,
			},
		},
		/*
			{
				Input: "`escaped path`",
				Output: &jparse.PathNode{
					Steps: []jparse.Node{
						&jparse.NameNode{
							Value:   "escaped path",
							Escaped: true,
						},
					},
				},
			},
		*/
	})
}

func TestStringers(t *testing.T) {

	data := []struct {
		Input  string
		String string
	}{
		{
			Input:  `"hello"`,
			String: `"hello"`,
		},
		{
			Input:  `'hello'`,
			String: `"hello"`,
		},
		{
			Input:  "100",
			String: "100",
		},
		{
			Input:  "3.14159",
			String: "3.14159",
		},
		{
			Input:  "true",
			String: "true",
		},
		{
			Input:  "false",
			String: "false",
		},
		{
			Input:  "null",
			String: "null",
		},
		{
			Input:  "/ab+/",
			String: "/ab+/",
		},
		{
			Input:  "/ab+/i",
			String: "/(?i)ab+/",
		},
		{
			Input:  "$varname",
			String: "$varname",
		},
		{
			Input:  "name",
			String: "name",
		},
		{
			Input:  "`quoted name`",
			String: "`quoted name`",
		},
		{
			Input:  "path.to.name",
			String: "path.to.name",
		},
		{
			Input:  "path.to.name[]",
			String: "path.to.name[]",
		},
		{
			Input:  "path[].to.name",
			String: "path.to.name[]",
		},
		{
			Input:  "path.to[].name",
			String: "path.to.name[]",
		},
		{
			Input:  "-1",
			String: "-1",
		},
		{
			Input:  "--1",
			String: "1",
		},
		{
			Input:  "-(1+2)",
			String: "-(1 + 2)",
		},
		{
			Input:  "[1..5]",
			String: "[1..5]",
		},
		{
			Input:  "[]",
			String: "[]",
		},
		{
			Input:  "[1,2,3]",
			String: "[1, 2, 3]",
		},
		{
			Input:  "[1..3,4,5,6]",
			String: "[1..3, 4, 5, 6]",
		},
		{
			Input:  "{}",
			String: "{}",
		},
		{
			Input: `{
						"one":   1,
						"two":   2,
						'three': 3
					}`,
			String: `{"one": 1, "two": 2, "three": 3}`,
		},
		{
			Input:  `(-1; -2; "three";)`,
			String: `(-1; -2; "three")`,
		},
		{
			Input:  "*",
			String: "*",
		},
		{
			Input:  "**",
			String: "**",
		},
		{
			Input: `| $ | {
				"one": 1,
				"two": 2
			} |`,
			String: `|$|{"one": 1, "two": 2}|`,
		},
		{
			Input: `| $ | {}, [
				"field1",
				"field2"
			] |`,
			String: `|$|{}, ["field1", "field2"]|`,
		},
		{
			Input:  `$substring('hello',-3,2)`,
			String: `$substring("hello", -3, 2)`,
		},
		{
			Input:  `$substring(?, 0, ?)`,
			String: `$substring(?, 0, ?)`,
		},
		{
			Input:  "function(){0}",
			String: "function(){0}",
		},
		{
			Input:  "λ($w,$h) {$w*$h}",
			String: "λ($w, $h){$w * $h}",
		},
		{
			Input:  "λ($x,$y,$z)<a<(ns)>-nf?:a>{$w*$h}",
			String: "λ($x, $y, $z)<a<(ns)>-nf?>{$w * $h}", // TODO: handle output type
		},
		{
			Input:  "$[0]",
			String: "$[0]",
		},
		{
			Input:  "$[Price>9.99]",
			String: "$[Price > 9.99]",
		},
		{
			Input:  "$[0][Price<25]",
			String: `$[0][Price < 25]`,
		},
		{
			Input: `Product{
				"name":   Name,
				"colour": Color,
				"price":  Price
			}`,
			String: `Product{"name": Name, "colour": Color, "price": Price}`,
		},
		{
			Input:  "true ? 'yes'",
			String: `true ? "yes"`,
		},
		{
			Input:  "true ? 'yes' : 'no'",
			String: `true ? "yes" : "no"`,
		},
		{
			Input:  "$x := $x+1",
			String: "$x := $x + 1",
		},
		{
			Input:  "1+2",
			String: "1 + 2",
		},
		{
			Input:  "1-2",
			String: "1 - 2",
		},
		{
			Input:  "1*2",
			String: "1 * 2",
		},
		{
			Input:  "1/2",
			String: "1 / 2",
		},
		{
			Input:  "1%2",
			String: "1 % 2",
		},
		{
			Input:  "1=2",
			String: "1 = 2",
		},
		{
			Input:  "1!=2",
			String: "1 != 2",
		},
		{
			Input:  "1>2",
			String: "1 > 2",
		},
		{
			Input:  "1>=2",
			String: "1 >= 2",
		},
		{
			Input:  "1<2",
			String: "1 < 2",
		},
		{
			Input:  "1<=2",
			String: "1 <= 2",
		},
		{
			Input:  "1 in [1,2]",
			String: "1 in [1, 2]",
		},
		{
			Input:  "true or false",
			String: "true or false",
		},
		{
			Input:  "null and void",
			String: "null and void",
		},
		{
			Input:  "'hello'&'world'",
			String: `"hello" & "world"`,
		},
		{
			Input:  "Product^(Price)",
			String: "Product^(Price)",
		},
		{
			Input:  "Product^(Price, >Name)",
			String: "Product^(Price, >Name)",
		},
		{
			Input:  "'hello' ~> $uppercase",
			String: `"hello" ~> $uppercase`,
		},
	}

	for _, test := range data {

		ast, err := jparse.Parse(test.Input)
		if err != nil {
			t.Errorf("%s: %s", test.Input, err)
			continue
		}

		if got := ast.String(); got != test.String {
			t.Errorf("%s: expected string %q, got %q", test.Input, test.String, got)
		}
	}
}

func testParser(t *testing.T, data []testCase) {

	for _, test := range data {

		inputs := test.Inputs
		if len(inputs) == 0 {
			inputs = []string{test.Input}
		}

		for _, input := range inputs {

			output, err := jparse.Parse(input)
			if err != nil && test.Error != nil {
				assert.EqualError(t, err, fmt.Sprintf("%v", test.Error))
			} else {
				assert.Equal(t, output.String(), test.Output.String())
			}
		}
	}
}
