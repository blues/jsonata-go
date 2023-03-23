// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/xiatechs/jsonata-go/jparse"
	"github.com/xiatechs/jsonata-go/jtypes"
)

type testCase struct {

	// Expression is either a single JSONata expression (type: string)
	// or a slice of JSONata expressions (type: []string) that produce
	// the same results.
	Expression interface{}

	// Vars is a map of variables to use when evaluating the given
	// expression(s).
	Vars map[string]interface{}

	// Exts is a map of custom extensions to use when evaluating the
	// given expression(s).
	Exts map[string]Extension

	// Output is the expected output for the given expression(s).
	Output interface{}

	// Error is the expected error for the given expression(s).
	Error error

	// Skip indicates whether or not this test case should be included
	// in the test run. Set to true to exclude a test case. Run "go test"
	// with the verbose flag to see which test cases were skipped.
	Skip bool
}

var testdata struct {
	account interface{}
	address interface{}
	library interface{}
	foobar  interface{}
}

func TestMain(m *testing.M) {

	// Decode and cache frequently used JSON.
	testdata.account = readJSON("account.json")
	testdata.address = readJSON("address.json")
	testdata.library = readJSON("library.json")
	testdata.foobar = readJSON("foobar.json")

	os.Exit(m.Run())
}

func TestLiterals(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`"Hello"`,
				`'Hello'`,
			},
			Output: "Hello",
		},
		{
			Expression: `"Wayne's World"`,
			Output:     "Wayne's World",
		},
		{
			Expression: "42",
			Output:     float64(42),
		},
		{
			Expression: "-42",
			Output:     float64(-42),
		},
		{
			Expression: "3.14159",
			Output:     3.14159,
		},
		{
			Expression: "6.022e23",
			Output:     6.022e23,
		},
		{
			Expression: "1.602E-19",
			Output:     1.602e-19,
		},
		{
			Expression: "1.602E+19",
			Output:     1.602e+19,
		},
		{
			Expression: "10e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Token:    "10e1000",
				Position: 0,
			},
		},
		{
			Expression: "-10e1000",
			Error: &jparse.Error{
				Type:     jparse.ErrNumberRange,
				Token:    "10e1000",
				Position: 1,
			},
		},
		{
			Expression: "1e",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Token:    "1e",
				Position: 0,
			},
		},
		{
			Expression: "-1e",
			Error: &jparse.Error{
				Type:     jparse.ErrInvalidNumber,
				Token:    "1e",
				Position: 1,
			},
		},
		{
			Expression: "-false",
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
	})
}

func TestStringLiterals(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`"hello\tworld"`,
				"'hello\\tworld'",
				"\"hello\\tworld\"",
			},
			Output: "hello\tworld",
		},
		{
			Expression: []string{
				`"hello\nworld"`,
				"'hello\\nworld'",
				"\"hello\\nworld\"",
			},
			Output: "hello\nworld",
		},
		{
			Expression: []string{
				`"hello \"world\""`,
				"\"hello \\\"world\\\"\"",
			},
			Output: `hello "world"`,
		},
		{
			Expression: []string{
				`"C:\\Test\\test.txt"`,
			},
			Output: "C:\\Test\\test.txt",
		},
		{
			Expression: []string{
				`"\u03BB-calculus rocks"`,
			},
			Output: `λ-calculus rocks`,
		},
		{
			Expression: []string{
				`"\uD834\uDD1E"`,
			},
			Output: "𝄞", // U+1D11E treble clef
		},
		{
			Expression: []string{
				`"\y"`,
				"'\\y'",
				"\"\\y\"",
			},
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscape,
				Position: 1,
				Token:    "\\y",
				Hint:     "y",
			},
		},
		{
			Expression: []string{
				`"\u"`,
				"'\\u'",
				"\"\\u\"",
			},
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\u",
				Hint:     "u" + strings.Repeat(string(utf8.RuneError), 4),
			},
		},
		{
			Expression: []string{
				`"\u123t"`,
				"'\\u123t'",
				"\"\\u123t\"",
			},
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalEscapeHex,
				Position: 1,
				Token:    "\\u123t",
				Hint:     "u123t",
			},
		},
	})
}

func TestPaths(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.bar",
			Output:     float64(42),
		},
		{
			Expression: "foo.blah",
			Output: []interface{}{
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "hello",
					},
				},
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "world",
					},
				},
				map[string]interface{}{
					"bazz": "gotcha",
				},
			},
		},
		{
			Expression: "foo.blah.baz",
			Output: []interface{}{
				map[string]interface{}{
					"fud": "hello",
				},
				map[string]interface{}{
					"fud": "world",
				},
			},
		},
		{
			Expression: "foo.blah.baz.fud",
			Output: []interface{}{
				"hello",
				"world",
			},
		},
		{
			Expression: "foo.blah.bazz",
			Output:     "gotcha",
		},
	})
}

func TestPaths2(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: "Other.Misc",
			Output:     nil,
			Skip:       true, // returns ErrUndefined
		},
	})
}

func TestPaths3(t *testing.T) {

	data := []interface{}{
		[]interface{}{
			map[string]interface{}{
				"baz": map[string]interface{}{
					"fud": "hello",
				},
			},
			map[string]interface{}{
				"baz": map[string]interface{}{
					"fud": "hello",
				},
			},
			map[string]interface{}{
				"bazz": "gotcha",
			},
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: "bazz",
			Output:     "gotcha",
		},
	})
}

func TestPaths4(t *testing.T) {

	data := []interface{}{
		42,
		[]interface{}{
			map[string]interface{}{
				"baz": map[string]interface{}{
					"fud": "hello",
				},
			},
			map[string]interface{}{
				"baz": map[string]interface{}{
					"fud": "hello",
				},
			},
			map[string]interface{}{
				"bazz": "gotcha",
			},
		},
		"here",
		map[string]interface{}{
			"fud": "hello",
		},
		"hello",
		map[string]interface{}{
			"fud": "world",
		},
		"world",
		"gotcha",
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: "fud",
			Output: []interface{}{
				"hello",
				"world",
			},
		},
	})
}

func TestSingletonArrays(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: `Phone[type="mobile"].number`,
			Output:     "077 7700 1234",
		},
		{
			Expression: []string{
				`Phone[type="mobile"][].number`,
				`Phone[][type="mobile"].number`,
			},
			Output: []interface{}{
				"077 7700 1234",
			},
		},
		{
			Expression: `Phone[type="office"][].number`,
			Output: []interface{}{
				"01962 001234",
				"01962 001235",
			},
		},
		{
			Expression: `Phone{type: number}`,
			Output: map[string]interface{}{
				"home": "0203 544 1234",
				"office": []interface{}{
					"01962 001234",
					"01962 001235",
				},
				"mobile": "077 7700 1234",
			},
		},
		{
			Expression: `Phone{type: number[]}`,
			Output: map[string]interface{}{
				"home": []interface{}{
					"0203 544 1234",
				},
				"office": []interface{}{
					"01962 001234",
					"01962 001235",
				},
				"mobile": []interface{}{
					"077 7700 1234",
				},
			},
		},
	})
}

func TestArraySelectors(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.blah[0]",
			Output: map[string]interface{}{
				"baz": map[string]interface{}{
					"fud": "hello",
				},
			},
		},
		{
			Expression: []string{
				"foo.blah[0].baz.fud",
				"foo.blah[0][0].baz.fud",
				"foo.blah[0][0][0].baz.fud",
			},
			Output: "hello",
		},
		{
			Expression: []string{
				"foo.blah[1].baz.fud",
				"(foo.blah)[1].baz.fud",
			},
			Output: "world",
		},
		{
			Expression: "foo.blah[-1].bazz",
			Output:     "gotcha",
		},
		{
			Expression: []string{
				"foo.blah.baz.fud[0]",
				"foo.blah.baz.fud[-1]",
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
		{
			Expression: []string{
				"(foo.blah.baz.fud)[0]",
				"(foo.blah.baz.fud)[-2]",
				"(foo.blah.baz.fud)[2-4]",
				"(foo.blah.baz.fud)[-(4-2)]",
			},
			Output: "hello",
		},
		{
			Expression: []string{
				"(foo.blah.baz.fud)[1]",
				"(foo.blah.baz.fud)[2 *0.5]",
				"(foo.blah.baz.fud)[0.25 * 4]",
				"(foo.blah.baz.fud)[-1]",
				"(foo.blah.baz.fud)[$$.foo.bar / 30]",
			},
			Output: "world",
		},
		{
			Expression: "foo.blah[0].baz",
			Output: map[string]interface{}{
				"fud": "hello",
			},
		},
		{
			Expression: "foo.blah.baz[0]",
			Output: []interface{}{
				map[string]interface{}{
					"fud": "hello",
				},
				map[string]interface{}{
					"fud": "world",
				},
			},
		},
		{
			Expression: []string{
				"(foo.blah.baz)[0]",
				"(foo.blah.baz)[0.5]",
				"(foo.blah.baz)[0.99]",
				"(foo.blah.baz)[-1.5]",
			},
			Output: map[string]interface{}{
				"fud": "hello",
			},
		},
		{
			Expression: []string{
				"(foo.blah.baz)[1]",
				"(foo.blah.baz)[-1]",
				"(foo.blah.baz)[-0.01]",
				"(foo.blah.baz)[-0.5]",
			},
			Output: map[string]interface{}{
				"fud": "world",
			},
		},
	})
}

func TestArraySelectors2(t *testing.T) {

	data := []interface{}{
		[]interface{}{
			1,
			2,
		},
		[]interface{}{
			3,
			4,
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: []string{
				"$[0]",
				"$[-2]",
			},
			Output: []interface{}{
				1,
				2,
			},
		},
		{
			Expression: []string{
				"$[1]",
				"$[-1]",
			},
			Output: []interface{}{
				3,
				4,
			},
		},
		{
			Expression: []string{
				"$[1][0]",
				"$[1.1][0.9]",
			},
			Output: 3,
		},
	})
}

func TestArraySelectors3(t *testing.T) {

	runTestCases(t, readJSON("nest3.json"), []*testCase{
		{
			Expression: "nest0.nest1[0]",
			Output: []interface{}{
				float64(1),
				float64(3),
				float64(5),
				float64(6),
			},
		},
	})
}

func TestArraySelectors4(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "[1..10][[1..3,8,-1]]",
			Output: []interface{}{
				float64(2),
				float64(3),
				float64(4),
				float64(9),
				float64(10),
			},
		},
		{
			Expression: "[1..10][[1..3,8,5]]",
			Output: []interface{}{
				float64(2),
				float64(3),
				float64(4),
				float64(6),
				float64(9),
			},
		},
		{
			Expression: "[1..10][[1..3,8,false]]",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
				float64(10),
			},
		},
	})

}

func TestQuotedSelectors(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.`blah`",
			Output: []interface{}{
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "hello",
					},
				},
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "world",
					},
				},
				map[string]interface{}{
					"bazz": "gotcha",
				},
			},
		},
		{
			Expression: []string{
				"`foo`.blah.baz.fud",
				"foo.`blah`.baz.fud",
				"foo.blah.`baz`.fud",
				"foo.blah.baz.`fud`",
				"`foo`.`blah`.`baz`.`fud`",
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
		{
			Expression: "foo.`blah.baz`",
			Output:     "here",
		},
	})
}

func TestNumericOperators(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: []string{
				"foo.bar + bar",
				"bar + foo.bar",
			},
			Output: float64(140),
		},
		{
			Expression: []string{
				"foo.bar * bar",
				"bar * foo.bar",
			},
			Output: float64(4116),
		},
		{
			Expression: "foo.bar - bar",
			Output:     float64(-56),
		},
		{
			Expression: "bar - foo.bar",
			Output:     float64(56),
		},
		{
			Expression: "foo.bar / bar",
			Output:     0.42857143,
		},
		{
			Expression: "bar / foo.bar",
			Output:     2.33333334,
		},
		{
			Expression: "foo.bar % bar",
			Output:     float64(42),
		},
		{
			Expression: "bar % foo.bar",
			Output:     float64(14),
		},
		{
			Expression: []string{
				"bar + foo.bar * bar",
				"foo.bar * bar + bar",
			},
			Output: float64(4214),
		},

		// If either operand returns no results, all arithmetic
		// operators return no results.

		{
			Expression: []string{
				"nothing + 3",
				"nothing - 3",
				"0.5 * nothing",
				"0.5 / nothing",
				"nothing % nothing",
			},
			Error: ErrUndefined,
		},

		// If either operand is a non-number type, return an error.

		{
			Expression: "'5' + 5",
			Error: &EvalError{
				Type:  ErrNonNumberLHS,
				Token: `"5"`,
				Value: "+",
			},
		},
		{
			Expression: "5 - '5'",
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: `"5"`,
				Value: "-",
			},
		},
		{
			Expression: "'5' * '5'",
			Error: &EvalError{
				Type:  ErrNonNumberLHS, // LHS is evaluated first
				Token: `"5"`,
				Value: "*",
			},
		},

		// If the result is out of range, return an error.

		{
			Expression: "10e300 * 10e100",
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "*",
			},
		},
		{
			Expression: "-10e300 * 10e100",
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "*",
			},
		},
		{
			Expression: "1 / (10e300 * 10e100)",
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "*",
			},
		},

		// If the result is NaN, return an error.

		{
			Expression: "0/0",
			Error: &EvalError{
				Type:  ErrNumberNaN,
				Value: "/",
			},
		},
	})
}

func TestComparisonOperators(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"3 > -3",
				"3 >= 3",
				"3 <= 3",
				"3 = 3",
				"'3' = '3'",
				"1 / 4 = 0.25",
				`"hello" = 'hello'`,
				"'hello' != 'world'",
				"'hello' < 'world'",
				"'hello' <= 'world'",
				"true = true",
				"false = false",
				"true != false",
			},
			Output: true,
		},
		{
			Expression: []string{
				"3 > 3",
				"3 < 3",
				"'3' = 3",
				"'hello' > 'world'",
				"'hello' >= 'world'",
				"true = 'true'",
				"false = 0",
			},
			Output: false,
		},
		{
			Expression: "null = null",
			Output:     true,
		},

		// Less/greater than operators require number or string
		// operands.

		{
			Expression: "null <= 'world'",
			Error: &EvalError{
				Type:  ErrNonComparableLHS,
				Token: "null",
				Value: "<=",
			},
		},
		{
			Expression: "3 >= true",
			Error: &EvalError{
				Type:  ErrNonComparableRHS,
				Token: "true",
				Value: ">=",
			},
		},

		// Less/greater than operators require operands of the
		// same type.

		{
			Expression: "'32' < 42",
			Error: &EvalError{
				Type:  ErrTypeMismatch,
				Value: "<",
			},
		},
	})
}

func TestComparisonOperators2(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: []string{
				"bar = 98",
				"foo.bar = 42",
				"foo.bar < bar",
				"foo.bar <= bar",
				"foo.bar != bar",
				"bar > foo.bar",
				"bar = foo.bar + 56",
			},
			Output: true,
		},
		{
			Expression: []string{
				"foo.bar > bar",
				"foo.bar >= bar",
				"foo.bar = bar",
				"bar < foo.bar",
				"bar <= foo.bar",
				"bar != foo.bar + 56",
			},
			Output: false,
		},
		{
			Expression: []string{
				`foo.blah.baz[fud = "hello"]`,
				`foo.blah.baz[fud != "world"]`,
			},
			Output: map[string]interface{}{
				"fud": "hello",
			},
		},

		// If either operand evaluates to no results, all
		// comparison operators return false.

		{
			Expression: []string{
				"bar = nothing",
				"nothing != bar",
				"bar > nothing",
				"nothing >= bar",
				"nothing < bar",
				"bar <= nothing",
			},
			Output: false,
		},
	})
}

func TestComparisonOperators3(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order.Product[Price > 30].Price",
			Output: []interface{}{
				34.45,
				34.45,
				107.99,
			},
		},
		{
			Expression: "Account.Order.Product.Price[$<=35]",
			Output: []interface{}{
				34.45,
				21.67,
				34.45,
			},
		},
	})
}

func TestIncludeOperator(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"1 in [1,2]",
				`"world" in ["hello", "world"]`,
				`"hello" in "hello"`,
			},
			Output: true,
		},
		{
			Expression: []string{
				"3 in [1,2]",
				`"hello" in [1,2]`,
				`in in ["hello", "world"]`,
				`"world" in in`,
			},
			Output: false,
		},
	})
}

func TestIncludeOperator2(t *testing.T) {

	runTestCases(t, testdata.library, []*testCase{
		{
			Expression: `library.books["Aho" in authors].title`,
			Output: []interface{}{
				"The AWK Programming Language",
				"Compilers: Principles, Techniques, and Tools",
			},
		},
	})
}

func TestIncludeOperator3(t *testing.T) {

	data := []interface{}{
		map[string]interface{}{
			"content": map[string]interface{}{
				"integration": map[string]interface{}{
					"name": "fakeIntegrationName",
				},
			},
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: `content.integration.$lowercase(name)`,
			Output:     "fakeintegrationname",
		},
	})
}

func TestParens(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "(4 + 2) / 2",
			Output:     float64(3),
		},
		{
			Expression: []string{
				"foo.blah.baz.fud",
				"(foo).blah.baz.fud",
				"foo.(blah).baz.fud",
				"foo.blah.(baz).fud",
				"foo.blah.baz.(fud)",

				"(foo.blah).baz.fud",
				"foo.(blah.baz).fud",
				"foo.blah.(baz.fud)",

				"(foo.blah.baz).fud",
				"foo.(blah.baz.fud)",

				"(foo.blah.baz.fud)",

				"(foo).blah.baz.(fud)",
				"(foo).(blah).baz.(fud)",
				"(foo).(blah).(baz).(fud)",

				"(foo.(blah).baz.fud)",
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
	})
}

func TestStringConcat(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "'hello' & ' ' & 'world'",
			Output:     "hello world",
		},

		// Non-string operands are converted to strings.

		{
			Expression: "'Hawaii' & ' ' & 5 & '-' & 0",
			Output:     "Hawaii 5-0",
		},
		{
			Expression: "10.0 & 'hello'",
			Output:     "10hello",
		},
		{
			Expression: "3 + 1 & 2.5",
			Output:     "42.5",
		},
		{
			Expression: "true & ' or ' & false",
			Output:     "true or false",
		},
		{
			Expression: "null & ' and void'",
			Output:     "null and void",
		},
		{
			Expression: "[1,2]&[3,4]",
			Output:     "[1,2][3,4]",
		},
		{
			Expression: "[1,2]&3",
			Output:     "[1,2]3",
		},
		{
			Expression: "1&3",
			Output:     "13",
		},
		{
			Expression: "1&[3]",
			Output:     "1[3]",
		},

		// Operands that return no results become blank strings.

		{
			Expression: "'Hello' & nothing",
			Output:     "Hello",
		},
		{
			Expression: "nothing & 'World'",
			Output:     "World",
		},
		{
			Expression: "nothing & nothing",
			Output:     "",
		},
	})
}

func TestStringConcat2(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: []string{
				"foo.blah[0].baz.fud & foo.blah[1].baz.fud",
				"foo.(blah[0].baz.fud & blah[1].baz.fud)",
			},
			Output: "helloworld",
		},
		{
			Expression: "foo.(blah[0].baz.fud & none)",
			Output:     "hello",
		},
		{
			Expression: "foo.(none.here & blah[1].baz.fud)",
			Output:     "world",
		},
	})
}

func TestStringConcat3(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "'Prices: ' & Account.Order.Product.Price",
			Output:     "Prices: [34.45,21.67,34.45,107.99]",
		},
	})
}

func TestArrayFlattening(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order.[Product.Price]",
			Output: []interface{}{
				[]interface{}{
					34.45,
					21.67,
				},
				[]interface{}{
					34.45,
					107.99,
				},
			},
		},
	})
}

func TestArrayFlattening2(t *testing.T) {

	runTestCases(t, readJSON("nest2.json"), []*testCase{
		{
			Expression: []string{
				"nest0",
				"$.nest0",
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
			},
		},
		{
			Expression: []string{
				"$[0]",
				"$[-2]",
			},
			Output: map[string]interface{}{
				"nest0": []interface{}{
					float64(1),
					float64(2),
				},
			},
		},
		{
			Expression: []string{
				"$[1]",
				"$[-1]",
			},
			Output: map[string]interface{}{
				"nest0": []interface{}{
					float64(3),
					float64(4),
				},
			},
		},
		{
			Expression: "$[0].nest0",
			Output: []interface{}{
				float64(1),
				float64(2),
			},
		},
		{
			Expression: "$[1].nest0",
			Output: []interface{}{
				float64(3),
				float64(4),
			},
		},
		{
			Expression: "$[0].nest0[0]",
			Output:     float64(1),
		},
		{
			Expression: "$[1].nest0[1]",
			Output:     float64(4),
		},
	})
}

func TestArrayFlattening3(t *testing.T) {

	runTestCases(t, readJSON("nest1.json"), []*testCase{
		{
			Expression: "nest0.nest1.nest2.nest3",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
			},
		},
		{
			Expression: "nest0.nest1.nest2.[nest3]",
			Output: []interface{}{
				[]interface{}{
					float64(1),
				},
				[]interface{}{
					float64(2),
				},
				[]interface{}{
					float64(3),
				},
				[]interface{}{
					float64(4),
				},
				[]interface{}{
					float64(5),
				},
				[]interface{}{
					float64(6),
				},
				[]interface{}{
					float64(7),
				},
				[]interface{}{
					float64(8),
				},
			},
		},
		{
			Expression: "nest0.nest1.[nest2.nest3]",
			Output: []interface{}{
				[]interface{}{
					float64(1),
					float64(2),
				},
				[]interface{}{
					float64(3),
					float64(4),
				},
				[]interface{}{
					float64(5),
					float64(6),
				},
				[]interface{}{
					float64(7),
					float64(8),
				},
			},
		},
		{
			Expression: "nest0.[nest1.nest2.nest3]",
			Output: []interface{}{
				[]interface{}{
					float64(1),
					float64(2),
					float64(3),
					float64(4),
				},
				[]interface{}{
					float64(5),
					float64(6),
					float64(7),
					float64(8),
				},
			},
		},
		{
			Expression: "nest0.[nest1.[nest2.nest3]]",
			Output: []interface{}{
				[]interface{}{
					[]interface{}{
						float64(1),
						float64(2),
					},
					[]interface{}{
						float64(3),
						float64(4),
					},
				},
				[]interface{}{
					[]interface{}{
						float64(5),
						float64(6),
					},
					[]interface{}{
						float64(7),
						float64(8),
					},
				},
			},
		},
		{
			Expression: "nest0.[nest1.nest2.[nest3]]",
			Output: []interface{}{
				[]interface{}{
					[]interface{}{
						float64(1),
					},
					[]interface{}{
						float64(2),
					},
					[]interface{}{
						float64(3),
					},
					[]interface{}{
						float64(4),
					},
				},
				[]interface{}{
					[]interface{}{
						float64(5),
					},
					[]interface{}{
						float64(6),
					},
					[]interface{}{
						float64(7),
					},
					[]interface{}{
						float64(8),
					},
				},
			},
		},
		{
			Expression: "nest0.nest1.[nest2.[nest3]]",
			Output: []interface{}{
				[]interface{}{
					[]interface{}{
						float64(1),
					},
					[]interface{}{
						float64(2),
					},
				},
				[]interface{}{
					[]interface{}{
						float64(3),
					},
					[]interface{}{
						float64(4),
					},
				},
				[]interface{}{
					[]interface{}{
						float64(5),
					},
					[]interface{}{
						float64(6),
					},
				},
				[]interface{}{
					[]interface{}{
						float64(7),
					},
					[]interface{}{
						float64(8),
					},
				},
			},
		},
		{
			Expression: "nest0.[nest1.[nest2.[nest3]]]",
			Output: []interface{}{
				[]interface{}{
					[]interface{}{
						[]interface{}{
							float64(1),
						},
						[]interface{}{
							float64(2),
						},
					},
					[]interface{}{
						[]interface{}{
							float64(3),
						},
						[]interface{}{
							float64(4),
						},
					},
				},
				[]interface{}{
					[]interface{}{
						[]interface{}{
							float64(5),
						},
						[]interface{}{
							float64(6),
						},
					},
					[]interface{}{
						[]interface{}{
							float64(7),
						},
						[]interface{}{
							float64(8),
						},
					},
				},
			},
		},
	})
}

func TestOperatorPrecedence(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "4 + 2 / 2",
			Output:     float64(5),
		},
		{
			Expression: "(4 + 2) / 2",
			Output:     float64(3),
		},
		{
			Expression: "1 / 2 * 2",
			Output:     1.0,
		},
		{
			Expression: "1 / (2 * 2)",
			Output:     0.25,
		},
	})
}

func TestPredicates(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "nothing[x=6][y=3].number",
			Error:      ErrUndefined,
		},
	})
}

func TestPredicates2(t *testing.T) {

	data := map[string]interface{}{
		"clues": []interface{}{
			map[string]interface{}{
				"x":      6,
				"y":      3,
				"number": 7,
			},
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: "clues[x=6][y=3].number",
			Output:     7,
		},
	})
}

func TestPredicates3(t *testing.T) {

	data := []interface{}{
		map[string]interface{}{
			"x":      6,
			"y":      2,
			"number": 7,
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: "$[x=6][y=2].number",
			Output:     7,
		},
		{
			Expression: "$[x=6][y=3].number",
			Error:      ErrUndefined,
		},
	})
}

func TestPredicates4(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`Account.Order.Product[Description.Colour = "Purple"][0].Price`,
				`Account.Order.Product[$lowercase(Description.Colour) = "purple"][0].Price`,
			},
			Output: []interface{}{
				34.45,
				34.45,
			},
		},
	})
}

func TestNotFound(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: []string{
				"fdf",
				"fdf.ett",
				"fdf.ett[10]",
				"fdf.ett[vc > 10]",
				"fdf.ett + 27",
				"$fdsd",
			},
			Error: ErrUndefined,
		},
	})
}

func TestSortOperator(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`Account.Order.Product.Price^($)`,
				`Account.Order.Product.Price^(<$)`,
			},
			Output: []interface{}{
				21.67,
				34.45,
				34.45,
				107.99,
			},
		},
		{
			Expression: `Account.Order.Product.Price^(>$)`,
			Output: []interface{}{
				107.99,
				34.45,
				34.45,
				21.67,
			},
		},
		{
			Expression: `Account.Order.Product^(Price).Description.Colour`,
			Output: []interface{}{
				"Orange",
				"Purple",
				"Purple",
				"Black",
			},
		},
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Output: []interface{}{
				"0406634348",
				"0406654608",
				"040657863",
				"0406654603",
			},
		},
		{
			Expression: `Account.Order.Product^(Price * Quantity).Description.Colour`,
			Output: []interface{}{
				"Orange",
				"Purple",
				"Black",
				"Purple",
			},
		},
		{
			Expression: `Account.Order.Product^(Quantity, Description.Colour).Description.Colour`,
			Output: []interface{}{
				"Black",
				"Orange",
				"Purple",
				"Purple",
			},
		},
		{
			Expression: `Account.Order.Product^(Quantity, >Description.Colour).Description.Colour`,
			Output: []interface{}{
				"Orange",
				"Black",
				"Purple",
				"Purple",
			},
		},
	})
}

func TestSortOperator2(t *testing.T) {

	runTestCases(t, readJSON("account2.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Output: []interface{}{
				"0406634348",
				"040657863",
				"0406654603",
				"0406654608",
			},
		},
	})
}

func TestSortOperator3(t *testing.T) {

	runTestCases(t, readJSON("account3.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Output: []interface{}{
				"0406654608",
				"040657863",
				"0406654603",
				"0406634348",
			},
		},
	})
}

func TestSortOperator4(t *testing.T) {

	runTestCases(t, readJSON("account4.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Output: []interface{}{
				"040657863",
				"0406654603",
				"0406654608",
				"0406634348",
			},
		},
	})
}

func TestSortOperator5(t *testing.T) {

	runTestCases(t, readJSON("account5.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Error: &EvalError{
				Type:  ErrSortMismatch,
				Token: "Price",
			},
		},
	})
}

func TestSortOperator6(t *testing.T) {

	runTestCases(t, readJSON("account6.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Error: &EvalError{
				Type:  ErrNonSortable,
				Token: "Price",
			},
		},
	})
}

func TestSortOperator7(t *testing.T) {

	runTestCases(t, readJSON("account7.json"), []*testCase{
		{
			Expression: `Account.Order.Product^(Price).SKU`,
			Error:      fmt.Errorf("The expressions within an order-by clause must evaluate to numeric or string values"), // TODO: use a proper error
			Skip:       true,                                                                                              // returns ErrUndefined
		},
	})
}

func TestWildcards(t *testing.T) {

	runTestCasesFunc(t, equalArraysUnordered, testdata.foobar, []*testCase{
		{
			Expression: "foo.*",
			Output: []interface{}{
				float64(42),
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "hello",
					},
				},
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "world",
					},
				},
				map[string]interface{}{
					"bazz": "gotcha",
				},
				"here",
			},
		},
		{
			Expression: "foo.*[0]",
			Output:     float64(42),
			Skip:       true, // We can't predict the order of the items in "foo.*"
		},
	})
}

func TestWildcards2(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.*.baz",
			Output: []interface{}{
				map[string]interface{}{
					"fud": "hello",
				},
				map[string]interface{}{
					"fud": "world",
				},
			},
		},
		{
			Expression: "foo.*.bazz",
			Output:     "gotcha",
		},
		{
			Expression: []string{
				"foo.*.baz.*",
				"*.*.baz.fud",
				"*.*.*.fud",
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
	})
}

func TestWildcards3(t *testing.T) {

	runTestCasesFunc(t, equalArraysUnordered, testdata.address, []*testCase{
		{
			Expression: `*[type="home"]`,
			Output: []interface{}{
				map[string]interface{}{
					"type":   "home",
					"number": "0203 544 1234",
				},
				map[string]interface{}{
					"type": "home",
					"address": []interface{}{
						"freddy@my-social.com",
						"frederic.smith@very-serious.com",
					},
				},
			},
		},
	})
}

func TestWildcards4(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account[$$.Account.`Account Name` = 'Firefly'].*[OrderID='order104'].Product.Price",
			Output: []interface{}{
				34.45,
				107.99,
			},
		},
	})
}

func TestDescendents(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.**.blah",
			Output: []interface{}{
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "hello",
					},
				},
				map[string]interface{}{
					"baz": map[string]interface{}{
						"fud": "world",
					},
				},
				map[string]interface{}{
					"bazz": "gotcha",
				},
			},
		},
		{
			Expression: "foo.**.baz",
			Output: []interface{}{
				map[string]interface{}{
					"fud": "hello",
				},
				map[string]interface{}{
					"fud": "world",
				},
			},
		},
		{
			Expression: []string{
				"foo.**.fud",
				"`foo`.**.fud",
				"foo.**.`fud`",
				"`foo`.**.`fud`",
				"foo.*.**.fud",
				"foo.**.*.fud",
				"foo.**.fud[0]",
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
		{
			Expression: []string{
				"(foo.**.fud)[0]",
				"(**.fud)[0]",
			},
			Output: "hello",
		},
	})
}

func TestDescendents2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order.**.Colour",
			Output: []interface{}{
				"Purple",
				"Orange",
				"Purple",
				"Black",
			},
		},
		{
			Expression: []string{
				"**.Price",
				"**.Price[0]",
			},
			Output: []interface{}{
				34.45,
				21.67,
				34.45,
				107.99,
			},
		},
		{
			Expression: "(**.Price)[0]",
			Output:     34.45,
		},
		{
			Expression: "**[2]",
			Output:     "Firefly",
			Skip:       true, // We can't guarantee the order of object sub-items!
		},
		{
			Expression: []string{
				"Account.Order.blah",
				"Account.Order.blah.**",
			},
			Error: ErrUndefined,
		},
	})
}

func TestDescendents3(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "**",
			Error:      ErrUndefined,
		},
	})
}

func TestBlockExpressions(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "()",
			Error:      ErrUndefined,
		},
		{
			Expression: []string{
				"(1; 2; 3)",
				"(1; 2; 3;)",
			},
			Output: float64(3),
		},
		{
			Expression: "($a:=1; $b:=2; $c:=($a:=4; $a+$b); $a+$c)",
			Output:     float64(7),
		},
	})
}

func TestBlockExpressions2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order.Product.($var1 := Price ; $var2:=Quantity; $var1 * $var2)",
			Output: []interface{}{
				68.9,
				21.67,
				137.8,
				107.99,
			},
		},
		{
			Expression: []string{
				`(
					$func := function($arg) {$arg.Account.Order[0].OrderID};
					$func($)
				)`,
				`(
					$func := function($arg) {$arg.Account.Order[0]};
					$func($).OrderID
				)`,
			},
			Output: "order103",
		},
	})
}

func TestArrayConstructor(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "[]",
			Output:     []interface{}{},
		},
		{
			Expression: "[1]",
			Output: []interface{}{
				float64(1),
			},
		},
		{
			Expression: "[1, 2]",
			Output: []interface{}{
				float64(1),
				float64(2),
			},
		},
		{
			Expression: "[1, 2,3]",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
		{
			Expression: "[1, 2, [3, 4]]",
			Output: []interface{}{
				float64(1),
				float64(2),
				[]interface{}{
					float64(3),
					float64(4),
				},
			},
		},
		{
			Expression: `[1, "two", ["three", 4]]`,
			Output: []interface{}{
				float64(1),
				"two",
				[]interface{}{
					"three",
					float64(4),
				},
			},
		},
		{
			Expression: `[1, $two, ["three", $four]]`,
			Vars: map[string]interface{}{
				"two":  float64(2),
				"four": "four",
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				[]interface{}{
					"three",
					"four",
				},
			},
		},
		{
			Expression: "[1, 2, 3][0]",
			Output:     float64(1),
		},
		{
			Expression: "[1, 2, [3, 4]][-1]",
			Output: []interface{}{
				float64(3),
				float64(4),
			},
		},
		{
			Expression: "[1, 2, [3, 4]][-1][-1]",
			Output:     float64(4),
		},
		{
			Expression: "[1..5][-1]",
			Output:     float64(5),
		},
		{
			Expression: "[1, 2, 3].$",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
	})
}

func TestArrayConstructor2(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "foo.blah.baz.[fud]",
			Output: []interface{}{
				[]interface{}{
					"hello",
				},
				[]interface{}{
					"world",
				},
			},
		},
		{
			Expression: "foo.blah.baz.[fud, fud]",
			Output: []interface{}{
				[]interface{}{
					"hello",
					"hello",
				},
				[]interface{}{
					"world",
					"world",
				},
			},
		},
		{
			Expression: "foo.blah.baz.[[fud, fud]]",
			Output: []interface{}{
				[]interface{}{
					[]interface{}{
						"hello",
						"hello",
					},
				},
				[]interface{}{
					[]interface{}{
						"world",
						"world",
					},
				},
			},
		},
		{
			Expression: `["foo.bar", foo.bar, ["foo.baz", foo.blah.baz]]`,
			Output: []interface{}{
				"foo.bar",
				float64(42),
				[]interface{}{
					"foo.baz",
					map[string]interface{}{
						"fud": "hello",
					},
					map[string]interface{}{
						"fud": "world",
					},
				},
			},
		},
	})
}

func TestArrayConstructor3(t *testing.T) {

	data := readJSON("foobar2.json")

	runTestCases(t, data, []*testCase{
		{
			Expression: "foo.blah.[baz].fud",
			Output:     "hello",
		},
		{
			Expression: "foo.blah.[baz, buz].fud",
			Output: []interface{}{
				"hello",
				"world",
			},
		},
	})
}

func TestArrayConstructor4(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: "[Address, Other.`Alternative.Address`].City",
			Output: []interface{}{
				"Winchester",
				"London",
			},
		},
	})
}

func TestArrayConstructor5(t *testing.T) {

	data := []interface{}{}

	runTestCases(t, data, []*testCase{
		{
			Expression: "[1, 2, 3].$",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
	})
}

func TestArrayConstructor6(t *testing.T) {

	data := []interface{}{4, 5, 6}

	runTestCases(t, data, []*testCase{
		{
			Expression: "[1, 2, 3].$",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
	})
}

func TestObjectConstructor(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "{}",
			Output:     map[string]interface{}{},
		},
		{
			Expression: `{"key":"value"}`,
			Output: map[string]interface{}{
				"key": "value",
			},
		},
		{
			Expression: `{"one": 1, "two": 2}`,
			Output: map[string]interface{}{
				"one": float64(1),
				"two": float64(2),
			},
		},
		{
			Expression: `{"one": 1, "two": 2}.two`,
			Output:     float64(2),
		},
		{
			Expression: `{"one": 1, "two": {"three": 3, "four": "4"}}`,
			Output: map[string]interface{}{
				"one": float64(1),
				"two": map[string]interface{}{
					"three": float64(3),
					"four":  "4",
				},
			},
		},
		{
			Expression: `{"one": 1, "two": [3, "four"]}`,
			Output: map[string]interface{}{
				"one": float64(1),
				"two": []interface{}{
					float64(3),
					"four",
				},
			},
		},
		{
			Expression: `{"test": ()}`,
			Output:     map[string]interface{}{},
		},
		{
			Expression: `blah.{}`,
			Error:      ErrUndefined,
		},
	})
}

func TestObjectConstructor2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order{OrderID: Product.`Product Name`}",
			Output: map[string]interface{}{
				"order103": []interface{}{
					"Bowler Hat",
					"Trilby hat",
				},
				"order104": []interface{}{
					"Bowler Hat",
					"Cloak",
				},
			},
		},
		{
			Expression: "Account.Order.{OrderID: Product.`Product Name`}",
			Output: []interface{}{
				map[string]interface{}{
					"order103": []interface{}{
						"Bowler Hat",
						"Trilby hat",
					},
				},
				map[string]interface{}{
					"order104": []interface{}{
						"Bowler Hat",
						"Cloak",
					},
				},
			},
		},
		{
			Expression: "Account.Order.Product{$string(ProductID): Price}",
			Output: map[string]interface{}{
				"345664": 107.99,
				"858236": 21.67,
				"858383": []interface{}{
					34.45,
					34.45,
				},
			},
		},
		{
			Expression: "Account.Order.Product{$string(ProductID): (Price)[0]}",
			Output: map[string]interface{}{
				"345664": 107.99,
				"858236": 21.67,
				"858383": 34.45,
			},
		},
		{
			Expression: "Account.Order.Product.{$string(ProductID): Price}",
			Output: []interface{}{
				map[string]interface{}{
					"858383": 34.45,
				},
				map[string]interface{}{
					"858236": 21.67,
				},
				map[string]interface{}{
					"858383": 34.45,
				},
				map[string]interface{}{
					"345664": 107.99,
				},
			},
		},
		{
			Expression: "Account.Order.Product{ProductID: `Product Name`}",
			Error: &EvalError{
				Type:  ErrIllegalKey,
				Token: "ProductID",
			},
		},
		{
			Expression: "Account.Order.Product.{ProductID: `Product Name`}",
			Error: &EvalError{
				Type:  ErrIllegalKey,
				Token: "ProductID",
			},
		},
		{
			Expression: "Account.Order{OrderID: $sum(Product.(Price*Quantity))}",
			Output: map[string]interface{}{
				"order103": 90.57,
				"order104": 245.79,
			},
		},
		{
			Expression: "Account.Order.{OrderID: $sum(Product.(Price*Quantity))}",
			Output: []interface{}{
				map[string]interface{}{
					"order103": 90.57,
				}, map[string]interface{}{
					"order104": 245.79,
				},
			},
		},
		{
			Expression: "Account.Order.Product{`Product Name`: Price, `Product Name`: Price}",
			Error: &EvalError{
				Type:  ErrDuplicateKey,
				Token: "`Product Name`",
				Value: "Bowler Hat",
			},
		},
		{
			Expression: `
				Account.Order{
					OrderID: {
						"TotalPrice": $sum(Product.(Price * Quantity)),
						"Items": Product.` + "`Product Name`" + `
					}
				}`,
			Output: map[string]interface{}{
				"order103": map[string]interface{}{
					"TotalPrice": 90.57,
					"Items": []interface{}{
						"Bowler Hat",
						"Trilby hat",
					},
				},
				"order104": map[string]interface{}{
					"TotalPrice": 245.79,
					"Items": []interface{}{
						"Bowler Hat",
						"Cloak",
					},
				},
			},
		},
		{
			Expression: `
				{
					"Order": Account.Order.{
						"ID": OrderID,
						"Product": Product.{
							"Name": ` + "`Product Name`" + `,
							"SKU": ProductID,
							"Details": {
								"Weight": Description.Weight,
								"Dimensions": Description.(Width & " x " & Height & " x " & Depth)
							}
						},
						"Total Price": $sum(Product.(Price * Quantity))
					}
				}`,
			Output: map[string]interface{}{
				"Order": []interface{}{
					map[string]interface{}{
						"ID": "order103",
						"Product": []interface{}{
							map[string]interface{}{
								"Name": "Bowler Hat",
								"SKU":  float64(858383),
								"Details": map[string]interface{}{
									"Weight":     0.75,
									"Dimensions": "300 x 200 x 210",
								},
							},
							map[string]interface{}{
								"Name": "Trilby hat",
								"SKU":  float64(858236),
								"Details": map[string]interface{}{
									"Weight":     0.6,
									"Dimensions": "300 x 200 x 210",
								},
							},
						},
						"Total Price": 90.57,
					},
					map[string]interface{}{
						"ID": "order104",
						"Product": []interface{}{
							map[string]interface{}{
								"Name": "Bowler Hat",
								"SKU":  float64(858383),
								"Details": map[string]interface{}{
									"Weight":     0.75,
									"Dimensions": "300 x 200 x 210",
								},
							},
							map[string]interface{}{
								"Name": "Cloak",
								"SKU":  float64(345664),
								"Details": map[string]interface{}{
									"Weight":     float64(2),
									"Dimensions": "30 x 20 x 210",
								},
							},
						},
						"Total Price": 245.79,
					},
				},
			},
		},
	})
}

func TestObjectConstructor3(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: `Phone{type: $join(number, ", "), "phone":number}`,
			Output: map[string]interface{}{
				"home": "0203 544 1234",
				"phone": []interface{}{
					"0203 544 1234",
					"01962 001234",
					"01962 001235",
					"077 7700 1234",
				},
				"office": "01962 001234, 01962 001235",
				"mobile": "077 7700 1234",
			},
		},
	})
}

func TestRangeOperator(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "[0..9]",
			Output: []interface{}{
				float64(0),
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
			},
		},
		{
			Expression: "[0..9][$ % 2 = 0]",
			Output: []interface{}{
				float64(0),
				float64(2),
				float64(4),
				float64(6),
				float64(8),
			},
		},

		{
			Expression: "[0, 4..9, 20, 22]",
			Output: []interface{}{
				float64(0),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
				float64(20),
				float64(22),
			},
		},
		{
			Expression: "[5..5]",
			Output: []interface{}{
				float64(5),
			},
		},
		{
			Expression: "[5..2]",
			Output:     []interface{}{},
		},
		{
			Expression: "[5..2, 2..5]",
			Output: []interface{}{
				float64(2),
				float64(3),
				float64(4),
				float64(5),
			},
		},
		{
			Expression: "[-2..2]",
			Output: []interface{}{
				float64(-2),
				float64(-1),
				float64(0),
				float64(1),
				float64(2),
			},
		},
		{
			Expression: "[-2..2].($*$)",
			Output: []interface{}{
				float64(4),
				float64(1),
				float64(0),
				float64(1),
				float64(4),
			},
		},
		{
			Expression: "[2*4..3*4-1]",
			Output: []interface{}{
				float64(8),
				float64(9),
				float64(10),
				float64(11),
			},
		},
		{
			Expression: "[-2..notfound]",
			Output:     []interface{}{},
		},
		{
			Expression: "[notfound..5, 3, -2..notfound]",
			Output: []interface{}{
				float64(3),
			},
		},
		{
			Expression: []string{
				"['1'..5]",
				"['1'..'5']", // LHS is evaluated first
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: `"1"`,
				Value: "..",
			},
		},
		{
			Expression: []string{
				"[1.1..5]",
				"[1.1..'5']", // LHS is evaluated first
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "1.1",
				Value: "..",
			},
		},
		{
			Expression: []string{
				"[true..5]",
				"[true..'5']", // LHS is evaluated first
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "true",
				Value: "..",
			},
		},

		{
			Expression: "[1..'5']",
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: `"5"`,
				Value: "..",
			},
		},
		{
			Expression: "[1..5.5]",
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "5.5",
				Value: "..",
			},
		},
		{
			Expression: "[1..false]",
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "false",
				Value: "..",
			},
		},
	})
}

func TestConditionals(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"true ? true",
				"true ? true : false",
				"1 > 0 ? true : false",
			},
			Output: true,
		},
		{
			Expression: []string{
				"false ? true : false",
				"'hello' = 'world' ? true : false",
			},
			Output: false,
		},
		{
			Expression: "false ? true",
			Error:      ErrUndefined,
		},
	})
}

func TestConditionals2(t *testing.T) {

	runTestCases(t, "Bus", []*testCase{
		{
			Expression: []string{
				`["Red"[$$="Bus"], "White"[$$="Police Car"]][0]`,
				`$lookup({"Bus": "Red", "Police Car": "White"}, $$)`,
			},
			Output: "Red",
		},
	})
}

func TestConditionals3(t *testing.T) {

	runTestCases(t, "Police Car", []*testCase{
		{
			Expression: []string{
				`["Red"[$$="Bus"], "White"[$$="Police Car"]][0]`,
				`$lookup({"Bus": "Red", "Police Car": "White"}, $$)`,
			},
			Output: "White",
		},
	})
}

func TestConditionals4(t *testing.T) {

	runTestCases(t, "Tuk tuk", []*testCase{
		{
			Expression: []string{
				`["Red"[$$="Bus"], "White"[$$="Police Car"]][0]`,
				//`$lookup({"Bus": "Red", "Police Car": "White"}, $$)`, // returns nil
			},
			Error: ErrUndefined,
		},
	})
}

func TestConditionals5(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `Account.Order.Product.(Price < 30 ? "Cheap")`,
			Output:     "Cheap",
		},
		{
			Expression: `Account.Order.Product.(Price < 30 ? "Cheap" : "Expensive")`,
			Output: []interface{}{
				"Expensive",
				"Cheap",
				"Expensive",
				"Expensive",
			},
		},
		{
			Expression: `Account.Order.Product.(Price < 30 ? "Cheap" : Price < 100 ? "Expensive" : "Rip off")`,
			Output: []interface{}{
				"Expensive",
				"Cheap",
				"Expensive",
				"Rip off",
			},
		},
	})
}

func TestBooleanExpressions(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"true",
				"true or true",
				"true and true",
				"true or false",
				"false or true",
				"true or nothing",
				"$not(false)",
			},
			Output: true,
		},
		{
			Expression: []string{
				"false",
				"false or false",
				"false and false",
				"false and true",
				"true and false",
				"nothing and false",
				"$not(true)",
			},
			Output: false,
		},
	})
}

func TestBooleanExpressions2(t *testing.T) {

	data := map[string]interface{}{
		"and": 1,
		"or":  2,
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: []string{
				"and=1 and or=2",
				"and>1 or or<=2",
				"and and and",
			},
			Output: true,
		},
		{
			Expression: []string{
				"and>1 or or!=2",
			},
			Output: false,
		},
	})
}

func TestBooleanExpressions3(t *testing.T) {

	data := []interface{}{
		map[string]interface{}{
			"content": map[string]interface{}{
				"origin": map[string]interface{}{
					"name": "fakeIntegrationName",
				},
			},
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: `$[].content.origin.$lowercase(name)`,
			Output: []interface{}{
				"fakeintegrationname",
			},
		},
	})
}

func TestNullExpressions(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "null",
			Output:     nil,
		},
		{
			Expression: "[null]",
			Output: []interface{}{
				nil,
			},
			Skip: true, // uses the wrong kind of nil (*interface{}(nil) instead of plain nil)?
		},
		{
			Expression: "[null, null]",
			Output: []interface{}{
				nil,
				nil,
			},
			Skip: true, // uses the wrong kind of nil (*interface{}(nil) instead of plain nil)?
		},
		{
			Expression: "$not(null)",
			Output:     true,
		},
		{
			Expression: "null = null",
			Output:     true,
		},
		{
			Expression: "null != null",
			Output:     false,
		},
		{
			Expression: `{"true": true, "false":false, "null": null}`,
			Output: map[string]interface{}{
				"true":  true,
				"false": false,
				"null":  nil,
			},
			Skip: true, // uses the wrong kind of nil (*interface{}(nil) instead of plain nil)?
		},
	})
}

func TestVariables(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$price.foo.bar",
			Vars: map[string]interface{}{
				"price": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": 45,
					},
				},
			},
			Output: 45,
		},
		{
			Expression: "$var[1]",
			Vars: map[string]interface{}{
				"var": []interface{}{
					1,
					2,
					3,
				},
			},
			Output: 2,
		},
		{
			Expression: "[1,2,3].$v",
			Vars: map[string]interface{}{
				"v": []interface{}{
					nil,
				},
			},
			Output: nil,
			Skip:   true, // jsonata-js passes in [undefined]. Not sure how to do that in Go.
		},
		{
			Expression: []string{
				"$a := 5",
				"$a := $b := 5",
				"($a := $b := 5; $a)",
				"($a := $b := 5; $b)",
			},
			Output: float64(5),
		},
		{
			Expression: "($a := 5; $a := $a + 2; $a)",
			Output:     float64(7),
		},
		{
			Expression: "5 := 5",
			Error: &jparse.Error{
				Type:     jparse.ErrIllegalAssignment,
				Token:    ":=",
				Hint:     "5",
				Position: 2,
			},
		},
	})
}

func TestVariables2(t *testing.T) {

	runTestCases(t, testdata.foobar, []*testCase{
		{
			Expression: "$.foo.bar",
			Vars: map[string]interface{}{
				"price": map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": 45,
					},
				},
			},
			Output: float64(42),
		},
	})
}

func TestVariableScope(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `( $foo := "defined"; ( $foo := nothing ); $foo )`,
			Output:     "defined",
		},
		{
			Expression: `( $foo := "defined"; ( $foo := nothing; $foo ) )`,
			Error:      ErrUndefined,
		},
	})
}

func TestLambdas(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `function($x){$x*$x}(5)`,
			Output:     float64(25),
		},
		{
			Expression: `function($x){$x>5 ? "foo"}(6)`,
			Output:     "foo",
		},
		{
			Expression: `function($x){$x>5 ? "foo"}(3)`,
			Error:      ErrUndefined,
		},
		{
			Expression: `($factorial:= function($x){$x <= 1 ? 1 : $x * $factorial($x-1)}; $factorial(4))`,
			Output:     float64(24),
		},
		{
			Expression: `($fibonacci := function($x){$x <= 1 ? $x : $fibonacci($x-1) + $fibonacci($x-2)}; [1,2,3,4,5,6,7,8,9].$fibonacci($))`,
			Output: []interface{}{
				float64(1),
				float64(1),
				float64(2),
				float64(3),
				float64(5),
				float64(8),
				float64(13),
				float64(21),
				float64(34),
			},
		},
		{
			Expression: `
				(
					$even := function($n) { $n = 0 ? true : $odd($n-1) };
					$odd := function($n) { $n = 0 ? false : $even($n-1) };
					$even(82)
				)`,
			Output: true,
		},
		{
			Expression: `
				(
					$even := function($n) { $n = 0 ? true : $odd($n-1) };
					$odd := function($n) { $n = 0 ? false : $even($n-1) };
					$even(65)
				)`,
			Output: false,
		},
		{
			Expression: `
				(
					$even := function($n) { $n = 0 ? true : $odd($n-1) };
					$odd := function($n) { $n = 0 ? false : $even($n-1) };
					$odd(65)
				)`,
			Output: true,
		},
		{
			Expression: `
				(
					$gcd := λ($a, $b){$b = 0 ? $a : $gcd($b, $a%$b) };
					[$gcd(8,12), $gcd(9,12)]
				)`,
			Output: []interface{}{
				float64(4),
				float64(3),
			},
		},
		{
			Expression: `
				(
					$range := function($start, $end, $step) { (
						$step:=($step?$step:1);
						$start+$step > $end ? $start : $append($start, $range($start+$step, $end, $step))
					)};
					$range(0,15)
				)`,
			Output: []interface{}{
				float64(0),
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
				float64(10),
				float64(11),
				float64(12),
				float64(13),
				float64(14),
				float64(15),
			},
		},
		{
			Expression: `
				(
					$range := function($start, $end, $step) { (
						$step:=($step?$step:1);
						$start+$step > $end ? $start : $append($start, $range($start+$step, $end, $step))
					)};
					$range(0,15,2)
				)`,
			Output: []interface{}{
				float64(0),
				float64(2),
				float64(4),
				float64(6),
				float64(8),
				float64(10),
				float64(12),
				float64(14),
			},
		},
	})
}

func TestLambdas2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `($nth_price := function($n) { (Account.Order.Product.Price)[$n] }; $nth_price(1) )`,
			Output:     21.67,
		},
	})
}

func TestPartials(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `
				(
					$add := function($x, $y){$x+$y};
					$add2 := $add(?, 2);
					$add2(3)
				)`,
			Output: float64(5),
		},
		{
			Expression: `
				(
					$add := function($x, $y){$x+$y};
					$add2 := $add(2, ?);
					$add2(4)
				)`,
			Output: float64(6),
		},
		{
			Expression: `
				(
					$last_letter := $substring(?, -1);
					$last_letter("Hello World")
				)`,
			Output: "d",
		},
		{
			Expression: `
				(
					$firstn := $substring(?, 0, ?);
					$first5 := $firstn(?, 5);
					$first5("Hello World")
				)`,
			Output: "Hello",
		},
		{
			Expression: `
				(
					$firstn := $substr(?, 0, ?);
					$first5 := $firstn(?, 5);
					$first5("Hello World")
				)`,
			Exts: map[string]Extension{
				"substr": {
					Func: func(s string, start, length int) string {
						if length <= 0 || start >= len(s) {
							return ""
						}

						if start < 0 {
							start += len(s)
						}

						if start > 0 {
							s = s[start:]
						}

						if length < len(s) {
							s = s[:length]
						}

						return s
					},
				},
			},
			Output: "Hello",
		},
		{
			Expression: `substring(?, 0, ?)`,
			Error: &EvalError{
				Type:  ErrNonCallablePartial,
				Token: "substring",
			},
		},
		{
			Expression: `nothing(?)`,
			Error: &EvalError{
				Type:  ErrNonCallablePartial,
				Token: "nothing",
			},
		},
	})
}

func TestFuncBoolean(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$boolean("Hello World")`,
				`$boolean(true)`,
				`$boolean(10)`,
				`$boolean(10.5)`,
				`$boolean(-0.5)`,
				`$boolean([1])`,
				`$boolean([true])`,
				`$boolean([1,2,3])`,
				`$boolean([[[true]]])`,
				`$boolean({"hello":"world"})`,
			},
			Output: true,
		},
		{
			Expression: []string{
				`$boolean("")`,
				`$boolean(false)`,
				`$boolean(0)`,
				`$boolean(0.0)`,
				`$boolean(-0)`,
				`$boolean(null)`,
				`$boolean([])`,
				`$boolean([null])`,
				`$boolean([false])`,
				`$boolean([0])`,
				`$boolean([0,0])`,
				`$boolean([[]])`,
				`$boolean([[null]])`,
				`$boolean({})`,
				`$boolean($boolean)`,
				`$boolean(function(){true})`,
			},
			Output: false,
		},
		{
			Expression: `$boolean(2,3)`,
			Error: &ArgCountError{
				Func:     "boolean",
				Expected: 1,
				Received: 2,
			},
		},
	})
}

func TestFuncBoolean2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`$boolean(Account)`,
				`$boolean(Account.Order.Product.Price)`,
			},
			Output: true,
		},
		{
			Expression: []string{
				`$boolean(Account.blah)`,
			},
			Error: ErrUndefined,
		},
	})
}

func TestFuncExists(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$boolean("Hello World")`,
				`$exists("")`,
				`$exists(true)`,
				`$exists(false)`,
				`$exists(0)`,
				`$exists(-0.5)`,
				`$exists(null)`,
				`$exists([])`,
				`$exists([0])`,
				`$exists([1,2,3])`,
				`$exists([[]])`,
				`$exists([[null]])`,
				`$exists([[[true]]])`,
				`$exists({})`,
				`$exists({"hello":"world"})`,
				`$exists($exists)`,
				`$exists(function(){true})`,
			},
			Output: true,
		},
		{
			Expression: []string{
				`$exists(nothing)`,
				`$exists($ha)`,
			},
			Output: false,
		},
		{
			Expression: "$exists()",
			Error: &ArgCountError{
				Func:     "exists",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$exists(2,3)",
			Error: &ArgCountError{
				Func:     "exists",
				Expected: 1,
				Received: 2,
			},
		},
	})
}

func TestFuncExists2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`$exists(Account)`,
				`$exists(Account.Order.Product.Price)`,
			},
			Output: true,
		},
		{
			Expression: []string{
				`$exists(blah)`,
				`$exists(Account.blah)`,
				`$exists(Account.Order[2])`,
				`$exists(Account.Order[0].blah)`,
			},
			Output: false,
		},
	})
}

func TestFuncCount(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$count([])",
				"$count([nothing])",
				"$count([nothing,nada,now't])",
			},
			Output: 0,
		},
		{
			Expression: []string{
				"$count(0)",
				"$count(false)",
				"$count(null)",
				`$count("")`,
			},
			Output: 1,
		},
		{
			Expression: []string{
				"$count([1,2,3])",
				"$count([1,2,3,nothing,nada,nichts])",
				"$count([1..3])",
				`$count(["1","2","3"])`,
				`$count(["1","2",3])`,
				`$count([0.5,true,{"one":1}])`,
			},
			Output: 3,
		},
		{
			Expression: "$count()",
			Error: &ArgCountError{
				Func:     "count",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: []string{
				"$count([],[])",
				"$count([1,2,3],[])",
			},
			Error: &ArgCountError{
				Func:     "count",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: []string{
				"$count(1,2,3)",
				"$count([],[],[])",
				"$count([1,2],[],[])",
			},
			Error: &ArgCountError{
				Func:     "count",
				Expected: 1,
				Received: 3,
			},
		},
		{
			Expression: "$count(nothing)",
			Output:     0,
		},
		{
			Expression: "$count([1,2,3,4]) / 2",
			Output:     float64(2),
		},
	})
}

func TestFuncCount2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$count(Account.Order.Product.(Price * Quantity))",
			Output:     4,
		},
		{
			Expression: "Account.Order.$count(Product.(Price * Quantity))",
			Output: []interface{}{
				2,
				2,
			},
		},
		{
			Expression: `Account.Order.(OrderID & ": " & $count(Product.(Price*Quantity)))`,
			Output: []interface{}{
				"order103: 2",
				"order104: 2",
			},
		},
	})
}

func TestFuncAppend(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$append([], [])",
				"$append([nothing], [nothing])",
			},
			Output: []interface{}{},
		},
		{
			Expression: []string{
				"$append(1, 2)",
				"$append([1], [2])",
			},
			Output: []interface{}{
				float64(1),
				float64(2),
			},
		},
		{
			Expression: "$append([1,2], [3,4])",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
			},
		},
		{
			Expression: []string{
				"$append(1, [3,4])",
				"$append([1,3], 4)",
				"$append([1,3,4], [])",
				"$append([1,3,4], nothing)",
				"$append([1,3,4], [nothing])",
				"$append([1,3,4], [nothing,nada])",
			},
			Output: []interface{}{
				float64(1),
				float64(3),
				float64(4),
			},
		},
		{
			Expression: []string{
				"$append(1, nothing)",
				"$append(nothing, 1)",
			},
			Output: float64(1),
		},
		{
			Expression: []string{
				"$append([1], nothing)",
				"$append([1,nothing], nothing)",
				"$append(nothing, [1])",
				"$append(nothing, [1,nothing])",
			},
			Output: []interface{}{
				float64(1),
			},
		},
		{
			Expression: []string{
				"$append([2,3,4], nothing)",
				"$append([2,3], [4,nothing])",
				"$append([2], [3,4,nothing])",
				"$append(2, [3,4,nothing])",
			},
			Output: []interface{}{
				float64(2),
				float64(3),
				float64(4),
			},
		},
		{
			Expression: "$append(nothing, nothing)",
			Error:      ErrUndefined,
		},
		{
			Expression: "$append()",
			Error: &ArgCountError{
				Func:     "append",
				Expected: 2,
				Received: 0,
			},
			Skip: true, // returns ErrUndefined
		},
		{
			Expression: "$append([])",
			Error: &ArgCountError{
				Func:     "append",
				Expected: 2,
				Received: 1,
			},
		},
		{
			Expression: "$append([],[],[])",
			Error: &ArgCountError{
				Func:     "append",
				Expected: 2,
				Received: 3,
			},
		},
	})
}

func TestFuncReverse(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$reverse([])",
				"$reverse([nothing])",
			},
			Output: []interface{}{},
		},
		{
			Expression: []string{
				"$reverse(1)",
				"$reverse([1])",
			},
			Output: []interface{}{
				float64(1),
			},
		},
		{
			Expression: []string{
				"$reverse([1..5])",
				"$reverse([1,2,3,4,nothing,5,nada])",
			},
			Output: []interface{}{
				float64(5),
				float64(4),
				float64(3),
				float64(2),
				float64(1),
			},
		},
		{
			Expression: `$reverse(["hello","world"])`,
			Output: []interface{}{
				"world",
				"hello",
			},
		},
		{
			Expression: `$reverse([true,0.5,"hello",{"one":1},[1,2,3]])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
				map[string]interface{}{
					"one": float64(1),
				},
				"hello",
				0.5,
				true,
			},
		},
		{
			Expression: "$reverse(nothing)",
			Error:      ErrUndefined,
		},
		{
			Expression: "$reverse()",
			Error: &ArgCountError{
				Func:     "reverse",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$reverse([],[])",
			Error: &ArgCountError{
				Func:     "reverse",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$reverse(1,2,3)",
			Error: &ArgCountError{
				Func:     "reverse",
				Expected: 1,
				Received: 3,
			},
		},
	})
}

func TestFuncReverse2(t *testing.T) {

	data := []interface{}{1, 2, 3}

	runTestCases(t, data, []*testCase{
		{
			Expression: "[$, $reverse($), $]",
			Output: []interface{}{
				[]interface{}{
					1,
					2,
					3,
				},
				[]interface{}{
					3,
					2,
					1,
				},
				[]interface{}{
					1,
					2,
					3,
				},
			},
			Skip: true, // sub-arrays are wrapped in extra brackets
		},
	})
}

func TestFuncSort(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$sort(nothing)",
			Error:      ErrUndefined,
		},
		{
			Expression: []string{
				"$sort(1)",
				"$sort([1])",
			},
			Output: []interface{}{
				float64(1),
			},
		},
		{
			Expression: "$sort([1,3,2])",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
		{
			Expression: "$sort([1,3,22,11])",
			Output: []interface{}{
				float64(1),
				float64(3),
				float64(11),
				float64(22),
			},
		},
	})
}

func TestFuncSort2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$sort(Account.Order.Product.Price)",
			Output: []interface{}{
				21.67,
				34.45,
				34.45,
				107.99,
			},
		},
		{
			Expression: "$sort(Account.Order.Product.`Product Name`)",
			Output: []interface{}{
				"Bowler Hat",
				"Bowler Hat",
				"Cloak",
				"Trilby hat",
			},
		},
		{
			Expression: `$sort(Account.Order.Product, function($a, $b) { $a.(Price * Quantity) > $b.(Price * Quantity) }).(Price & " x " & Quantity)`,
			Output: []interface{}{
				"21.67 x 1",
				"34.45 x 2",
				"107.99 x 1",
				"34.45 x 4",
			},
		},
		{
			Expression: `$sort(Account.Order.Product, function($a, $b) { $a.Price > $b.Price }).SKU`,
			Output: []interface{}{
				"0406634348",
				"0406654608",
				"040657863",
				"0406654603",
			},
		},
		{
			Expression: `
				(Account.Order.Product
					~> $sort(λ($a,$b){$a.Quantity < $b.Quantity})
					~> $sort(λ($a,$b){$a.Price > $b.Price})
				).SKU`,
			Output: []interface{}{
				"0406634348",
				"040657863",
				"0406654608",
				"0406654603",
			},
		},
		{
			Expression: "$sort(Account.Order.Product)",
			Error:      fmt.Errorf("argument 1 of function sort must be an array of strings or numbers"), // TODO: Use a proper error
		},
	})
}

func TestFuncSort3(t *testing.T) {

	data := []interface{}{1, 3, 2}

	runTestCases(t, data, []*testCase{
		{
			Expression: "[[$], [$sort($)], [$]]",
			Output: []interface{}{
				[]float64{
					1,
					3,
					2,
				},
				[]float64{
					1,
					2,
					3,
				},
				[]float64{
					1,
					3,
					2,
				},
			},
			Skip: true, // inner arrays are wrapped in extra brackets
		},
	})
}

func TestFuncShuffle(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$shuffle([])",
				"$shuffle([nothing])",
			},
			Output: []interface{}{},
		},
		{
			Expression: []string{
				"$shuffle(1)",
				"$shuffle([1])",
			},
			Output: []interface{}{
				float64(1),
			},
		},
		{
			Expression: "$count($shuffle([1..10]))",
			Output:     10,
		},
		{
			Expression: "$sort($shuffle([1..10]))",
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
				float64(10),
			},
		},
		{
			Expression: "$shuffle(nothing)",
			Error:      ErrUndefined,
		},
		{
			Expression: "$shuffle()",
			Error: &ArgCountError{
				Func:     "shuffle",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$shuffle([],[])",
			Error: &ArgCountError{
				Func:     "shuffle",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$shuffle(1,2,3)",
			Error: &ArgCountError{
				Func:     "shuffle",
				Expected: 1,
				Received: 3,
			},
		},
	})
}

func TestFuncShuffle2(t *testing.T) {

	// TODO: These tests don't actually verify that the values
	// have been shuffled.
	runTestCasesFunc(t, equalArraysUnordered, nil, []*testCase{
		{
			Expression: []string{
				"$shuffle([1..10])",
				"$shuffle($reverse([1..10]))",
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
				float64(6),
				float64(7),
				float64(8),
				float64(9),
				float64(10),
			},
		},
		{
			Expression: `$shuffle([true,-0.5,"hello",[1,2,3],{"one":1}])`,
			Output: []interface{}{
				true,
				-0.5,
				"hello",
				[]interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
				map[string]interface{}{
					"one": float64(1),
				},
			},
		},
	})
}

func TestFuncZip(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$zip(1,2,3)`,
				`$zip(1,2,[3])`,
				`$zip([1],[2],[3])`,
				`$zip([1],[2,3],[3,4,5])`,
			},
			Output: []interface{}{
				[]interface{}{
					float64(1),
					float64(2),
					float64(3),
				},
			},
		},
		{
			Expression: `$zip([1,2,3])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
				},
				[]interface{}{
					float64(2),
				},
				[]interface{}{
					float64(3),
				},
			},
		},
		{
			Expression: `$zip([1,2,3],["one","two","three"])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
					"one",
				},
				[]interface{}{
					float64(2),
					"two",
				},
				[]interface{}{
					float64(3),
					"three",
				},
			},
		},
		{
			Expression: `$zip([1,2,3],["one","two","three"],[true,false,true])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
					"one",
					true,
				},
				[]interface{}{
					float64(2),
					"two",
					false,
				},
				[]interface{}{
					float64(3),
					"three",
					true,
				},
			},
		},
		{
			Expression: `$zip([1,2,3],["one","two"],[true,false,true])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
					"one",
					true,
				},
				[]interface{}{
					float64(2),
					"two",
					false,
				},
			},
		},
		{
			Expression: []string{
				`$zip(nothing)`,
				`$zip(nothing,nada,now't)`,
				`$zip(1,2,3,nothing)`,
				`$zip([1,2,3],nothing)`,
				`$zip([1,2,3],[nothing])`,
				`$zip([1,2,3],[nothing,nada,now't])`,
			},
			Output: []interface{}{},
		},
		{
			Expression: `$zip()`,
			Error:      fmt.Errorf("cannot call zip with no arguments"),
		},
	})
}

func TestFuncSum(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$sum(0)",
				"$sum([])",
				"$sum([0])",
				"$sum([nothing])",
				"$sum([0,nothing])",
			},
			Output: float64(0),
		},
		{
			Expression: "$sum(1)",
			Output:     float64(1),
		},
		{
			Expression: []string{
				"$sum(15)",
				"$sum([1..5])",
				"$sum([1..5])",
				"$sum([1..4,5])",
				"$sum([1,2,3,4,5])",
				"$sum([1,2,3,4,nothing,nada,5])",
			},
			Output: float64(15),
		},
		{
			Expression: []string{
				`$sum("")`,
				"$sum(true)",
				`$sum({"one":1})`,
			},
			Error: errors.New("cannot call sum on a non-array type"), // TODO: Don't rely on error strings
		},
		{
			Expression: []string{
				`$sum([1,2,"3"])`,
				"$sum([1,2,true])",
			},
			Error: errors.New("cannot call sum on an array with non-number types"), // TODO: Don't rely on error strings
		},
		{
			Expression: "$sum()",
			Error: &ArgCountError{
				Func:     "sum",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$sum(1,2)",
			Error: &ArgCountError{
				Func:     "sum",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$sum(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSum2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$sum(Account.Order.Product.(Price * Quantity))",
			Output:     336.36,
		},
		{
			Expression: "Account.Order.$sum(Product.(Price * Quantity))",
			Output: []interface{}{
				90.57,
				245.79,
			},
		},
		{
			Expression: `Account.Order.(OrderID & ": " & $sum(Product.(Price*Quantity)))`,
			Output: []interface{}{
				// TODO: Why does jsonata-js only display to 2dp?
				"order103: 90.57",
				"order104: 245.79",
			},
		},
		{
			Expression: "$sum(Account.Order)",
			Error:      fmt.Errorf("cannot call sum on an array with non-number types"), // TODO: relying on error strings is bad
		},
	})
}

func TestFuncMax(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$max(1)",
				"$max([1,0])",
				"$max([1,0,-1.5])",
			},
			Output: float64(1),
		},
		{
			Expression: []string{
				"$max(-1)",
				"$max([-1,-5])",
				"$max([-1,-5,nothing])",
			},
			Output: float64(-1),
		},
		{
			Expression: []string{
				"$max(3)",
				"$max([1,2,3])",
				"$max([1..3])",
				"$max([1..3,nothing])",
			},
			Output: float64(3),
		},
		{
			Expression: []string{
				`$max("")`,
				`$max(true)`,
				`$max({"one":1})`,
			},
			Error: fmt.Errorf("cannot call max on a non-array type"), // TODO: Don't rely on the error string
		},
		{
			Expression: []string{
				`$max(["1","2","3"])`,
				`$max(["1","2",3])`,
			},
			Error: fmt.Errorf("cannot call max on an array with non-number types"), // TODO: Don't rely on the error string
		},
		{
			Expression: "$max()",
			Error: &ArgCountError{
				Func:     "max",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$max([],[])",
			Error: &ArgCountError{
				Func:     "max",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$max(1,2,3)",
			Error: &ArgCountError{
				Func:     "max",
				Expected: 1,
				Received: 3,
			},
		},
		{
			Expression: []string{
				"$max(nothing)",
				"$max([])",
				"$max([nothing])",
				"$max([nothing,nada,nichts])",
			},
			Error: ErrUndefined,
		},
	})
}

func TestFuncMax2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$max(Account.Order.Product.(Price * Quantity))",
			Output:     137.8,
		},
		{
			Expression: "$max(Account.Order.Product.(Price * Quantity))",
			Output:     137.8,
		},
		{
			Expression: "Account.Order.$max(Product.(Price * Quantity))",
			Output: []interface{}{
				68.9,
				137.8,
			},
		},
	})
}

func TestFuncMin(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$min(1)",
				"$min([1,2,3])",
				"$min([1..3])",
				"$min([1..3,nothing])",
			},
			Output: float64(1),
		},
		{
			Expression: []string{
				"$min(-5)",
				"$min([-1,-2,-3,-4,-5])",
				"$min([-5..0])",
				"$min([-5,nothing])",
			},
			Output: float64(-5),
		},
		{
			Expression: []string{
				`$min("")`,
				`$min(true)`,
				`$min({"one":1})`,
			},
			Error: fmt.Errorf("cannot call min on a non-array type"), // TODO: Don't rely on error strings
		},
		{
			Expression: []string{
				`$min(["1","2","3"])`,
				`$min(["1","2",3])`,
			},
			Error: fmt.Errorf("cannot call min on an array with non-number types"), // TODO: Don't rely on error strings
		},
		{
			Expression: "$min()",
			Error: &ArgCountError{
				Func:     "min",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$min([],[])",
			Error: &ArgCountError{
				Func:     "min",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$min(1,2,3)",
			Error: &ArgCountError{
				Func:     "min",
				Expected: 1,
				Received: 3,
			},
		},
		{
			Expression: []string{
				"$min([])",
				"$min(nothing)",
				"$min([nothing,nada,now't])",
			},
			Error: ErrUndefined,
		},
	})
}

func TestFuncMin2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$min(Account.Order.Product.(Price * Quantity))",
			Output:     21.67,
		},
		{
			Expression: "Account.Order.$min(Product.(Price * Quantity))",
			Output: []interface{}{
				21.67,
				107.99,
			},
		},
		{
			Expression: `Account.Order.(OrderID & ": " & $min(Product.(Price*Quantity)))`,
			Output: []interface{}{
				"order103: 21.67",
				"order104: 107.99",
			},
		},
	})
}

func TestFuncAverage(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$average(1)",
			Output:     float64(1),
		},
		{
			Expression: []string{
				"$average(2)",
				"$average([1,2,3])",
				"$average([1..3])",
				"$average([1..3,nothing])",
			},
			Output: float64(2),
		},
		{
			Expression: []string{
				`$average("")`,
				`$average(true)`,
				`$average({"one":1})`,
			},
			Error: fmt.Errorf("cannot call average on a non-array type"), // TODO: Don't rely on the error string
		},
		{
			Expression: []string{
				`$average(["1","2","3"])`,
				`$average(["1","2",3])`,
			},
			Error: fmt.Errorf("cannot call average on an array with non-number types"), // TODO: Don't rely on the error string
		},
		{
			Expression: "$average()",
			Error: &ArgCountError{
				Func:     "average",
				Expected: 1,
				Received: 0,
			},
		},
		{
			Expression: "$average([],[])",
			Error: &ArgCountError{
				Func:     "average",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: "$average(1,2,3)",
			Error: &ArgCountError{
				Func:     "average",
				Expected: 1,
				Received: 3,
			},
		},
		{
			Expression: []string{
				"$average([])",
				"$average(nothing)",
				"$average([nothing,nada,now't])"},
			Error: ErrUndefined,
		},
	})
}

func TestFuncAverage2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "$average(Account.Order.Product.(Price * Quantity))",
			Output:     84.09,
		},
		{
			Expression: "Account.Order.$average(Product.(Price * Quantity))",
			Output: []interface{}{
				45.285,
				122.895,
			},
		},
		{
			Expression: `Account.Order.(OrderID & ": " & $average(Product.(Price*Quantity)))`,
			Output: []interface{}{
				// TODO: Why does jsonata-js only display to 3dp?
				"order103: 45.285",
				"order104: 122.895",
			},
		},
	})
}

func TestFuncSpread(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$spread("Hello World")`,
			Output:     "Hello World",
		},
		{
			Expression: `$spread([1,2,3])`,
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
		{
			Expression: `$string($spread(function($x){$x*$x}))`,
			Output:     "",
		},
		{
			Expression: "$spread(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSpread2(t *testing.T) {

	data := []map[string]interface{}{
		{
			"one": 1,
		},
		{
			"two": 2,
		},
		{
			"three": 3,
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: `$spread($[0])`,
			Output: []interface{}{
				map[string]interface{}{
					"one": 1,
				},
			},
		},
		{
			Expression: `$spread($)`,
			Output: []interface{}{
				map[string]interface{}{
					"one": 1,
				},
				map[string]interface{}{
					"two": 2,
				},
				map[string]interface{}{
					"three": 3,
				},
			},
		},
	})
}

func TestFuncSpread3(t *testing.T) {

	runTestCasesFunc(t, equalArraysUnordered, testdata.account, []*testCase{
		{
			Expression: "$spread((Account.Order.Product.Description))",
			Output: []interface{}{
				map[string]interface{}{
					"Colour": "Purple",
				},
				map[string]interface{}{
					"Width": float64(300),
				},
				map[string]interface{}{
					"Height": float64(200),
				},
				map[string]interface{}{
					"Depth": float64(210),
				},
				map[string]interface{}{
					"Weight": 0.75,
				},
				map[string]interface{}{
					"Colour": "Orange",
				},
				map[string]interface{}{
					"Width": float64(300),
				},
				map[string]interface{}{
					"Height": float64(200),
				},
				map[string]interface{}{
					"Depth": float64(210),
				},
				map[string]interface{}{
					"Weight": 0.6,
				},
				map[string]interface{}{
					"Colour": "Purple",
				},
				map[string]interface{}{
					"Width": float64(300),
				},
				map[string]interface{}{
					"Height": float64(200),
				},
				map[string]interface{}{
					"Depth": float64(210),
				},
				map[string]interface{}{
					"Weight": 0.75,
				},
				map[string]interface{}{
					"Colour": "Black",
				},
				map[string]interface{}{
					"Width": float64(30),
				},
				map[string]interface{}{
					"Height": float64(20),
				},
				map[string]interface{}{
					"Depth": float64(210),
				},
				map[string]interface{}{
					"Weight": float64(2),
				},
			},
		},
	})
}

/*
func TestFuncSpread4(t *testing.T) {

	t.SkipNow()

	data := []struct {
		Int        int
		Bool       bool
		String     string
		Interface  interface{}
		unexported bool
	}{
		{
			Int:    1,
			Bool:   true,
			String: "string",
		},
		{
			Int:    0,
			Bool:   false,
			String: "",
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: `$spread($[0])`,
			Output: []interface{}{
				map[string]interface{}{
					"Int": 1,
				},
				map[string]interface{}{
					"Bool": true,
				},
				map[string]interface{}{
					"String": "string",
				},
				map[string]interface{}{
					"Interface": nil,
				},
			},
		},
		{
			Expression: `$spread($[1])`,
			Output: []interface{}{
				map[string]interface{}{
					"Int": 0,
				},
				map[string]interface{}{
					"Bool": false,
				},
				map[string]interface{}{
					"String": "",
				},
				map[string]interface{}{
					"Interface": nil,
				},
			},
		},
		{
			Expression: `$spread($)`,
			Output: []interface{}{
				map[string]interface{}{
					"Int": 1,
				},
				map[string]interface{}{
					"Bool": true,
				},
				map[string]interface{}{
					"String": "string",
				},
				map[string]interface{}{
					"Interface": nil,
				},
				map[string]interface{}{
					"Int": 0,
				},
				map[string]interface{}{
					"Bool": false,
				},
				map[string]interface{}{
					"String": "",
				},
				map[string]interface{}{
					"Interface": nil,
				},
			},
		},
	})
}
*/

func TestFuncMerge(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$merge({"a":1})`,
			Output: map[string]interface{}{
				"a": float64(1),
			},
		},
		{
			Expression: `$merge([{"a":1}, {"b":2}])`,
			Output: map[string]interface{}{
				"a": float64(1),
				"b": float64(2),
			},
		},
		{
			Expression: `$merge([{"a": 1}, {"b": 2, "c": 3}])`,
			Output: map[string]interface{}{
				"a": float64(1),
				"b": float64(2),
				"c": float64(3),
			},
		},
		{
			Expression: `$merge([{"a": 1}, {"b": 2, "a": 3}])`,
			Output: map[string]interface{}{
				"a": float64(3),
				"b": float64(2),
			},
		},
		{
			Expression: []string{
				`$merge([])`,
				`$merge({})`,
			},
			Output: map[string]interface{}{},
		},
		{
			Expression: `$merge(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncEach(t *testing.T) {

	runTestCasesFunc(t, equalArraysUnordered, testdata.address, []*testCase{
		{
			Expression: `$each(Address, λ($v, $k) {$k & ": " & $v})`,
			Output: []interface{}{
				"Street: Hursley Park",
				"City: Winchester",
				"Postcode: SO21 2JN",
			},
		},
	})
}

func TestFuncMap(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$map([1,2,3], $string)`,
			Output: []interface{}{
				"1",
				"2",
				"3",
			},
		},
		{
			Expression: `$map([1,4,9,16], $squareroot)`,
			Exts: map[string]Extension{
				"squareroot": {
					Func: func(x float64) float64 {
						return math.Sqrt(x)
					},
				},
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
			},
		},
		{
			Expression: `
				(
					$data := {
						"one": [1,2,3,4,5],
						"two": [5,4,3,2,1]
					};
					$add := function($x){$x*$x};
					$map($data.one, $add)
				)`,
			Output: []interface{}{
				float64(1),
				float64(4),
				float64(9),
				float64(16),
				float64(25),
			},
		},
		{
			Expression: "$map($string)",
			Error: &ArgCountError{
				Func:     "map",
				Expected: 2,
				Received: 1,
			},
		},
	})
}

func TestFuncMap2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `Account.Order.Product ~> $map(λ($prod, $index) { $index+1 & ": " & $prod.` + "`Product Name`" + ` })`,
			Output: []interface{}{
				"1: Bowler Hat",
				"2: Trilby hat",
				"3: Bowler Hat",
				"4: Cloak",
			},
		},
		{
			Expression: `Account.Order.Product ~> $map(λ($prod, $index, $arr) { $index+1 & "/" & $count($arr) & ": " & $prod.` + "`Product Name`" + ` })`,
			Output: []interface{}{
				"1/4: Bowler Hat",
				"2/4: Trilby hat",
				"3/4: Bowler Hat",
				"4/4: Cloak",
			},
		},
		{
			Expression: `$map(Phone, function($v, $i) {$v.type="office" ? $i: null})`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncMap3(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: []string{
				`$map(Phone, function($v, $i) {$i[$v.type="office"]})`,
				`$map(Phone, function($v, $i) {$v.type="office" ? $i})`,
			},
			Output: []interface{}{
				1,
				2,
			},
		},
		{
			Expression: `$map(Phone, function($v, $i) {$v.type="office" ? $i: null})`,
			Output: []interface{}{
				nil,
				1,
				2,
				nil,
			},
			Skip: true, // works with non-null else value. Returning the wrong kind of nils?
		},
	})
}

func TestFuncMapZip(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`(
					$data := {
						"one": [1,2,3,4,5],
						"two": [5,4,3,2,1]
					};
					$map($zip($data.one, $data.two), $sum)
				)`,
				`(
					$data := {
						"one": [1,2,3,4,5],
						"two": [5,4,3,2,1]
					};
					$data.$zip(one, two) ~> $map($sum)
				)`,
			},
			Output: []interface{}{
				float64(6),
				float64(6),
				float64(6),
				float64(6),
				float64(6),
			},
		},
		{
			Expression: []string{
				`(
					$data := {
						"one": [1],
						"two": [5]
					};
					$data[].$zip(one, two) ~> $map($sum)
				)`,
				`(
					$data := {
						"one": 1,
						"two": 5
					};
					$data[].$zip(one, two) ~> $map($sum)
				)`,
			},
			Output: []interface{}{
				float64(6),
			},
		},
	})
}

func TestFuncFilter(t *testing.T) {

	runTestCases(t, testdata.library, []*testCase{
		{
			Expression: `$filter([1..10], function($v) {$v % 2})`,
			Output: []interface{}{
				float64(1),
				float64(3),
				float64(5),
				float64(7),
				float64(9),
			},
		},
	})
}

func TestFuncFilter2(t *testing.T) {

	runTestCases(t, testdata.library, []*testCase{
		{
			Expression: `(library.books~>$filter(λ($v, $i, $a) {$v.price = $max($a.price)})).isbn`,
			Output:     "9780262510875",
		},
		{
			Expression: `(nothing~>$filter(λ($v, $i, $a) {$v.price = $max($a.price)})).isbn`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncReduce(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `
				(
					$seq := [1,2,3,4,5];
					$reduce($seq, function($x, $y){$x+$y})
				)`,
			Output: float64(15),
		},
		{
			Expression: `
				(
					$concat := function($s){function($a, $b){$string($a) & $s & $string($b)}};
					$comma_join := $concat(' ... ');
					$reduce([1,2,3,4,5], $comma_join)
				)`,
			Output: "1 ... 2 ... 3 ... 4 ... 5",
		},
		{
			Expression: `
				(
					$seq := [1,2,3,4,5];
					$reduce($seq, function($x, $y){$x+$y}, 2)
				)`,
			Output: float64(17),
		},
		{
			Expression: `
				(
					$seq := 1;
					$reduce($seq, function($x, $y){$x+$y})
				)`,
			Output: float64(1),
		},
		{
			Expression: `
				(
					$product := function($a, $b) { $a * $b };
					$power := function($x, $n) { $n = 0 ? 1 : $reduce([1..$n].($x), $product) };
					[0..5].$power(2, $)
				)`,
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(4),
				float64(8),
				float64(16),
				float64(32),
			},
		},
		{
			Expression: `
				(
					$seq := 1;
					$reduce($seq, function($x){$x})
				)`,
			Error: fmt.Errorf("second argument of function \"reduce\" must be a function that takes two arguments"),
		},
	})
}

func TestFuncReduce2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `$reduce(Account.Order.Product.Quantity, $append)`,
			Output: []interface{}{
				float64(2),
				float64(1),
				float64(4),
				float64(1),
			},
		},
	})
}

func TestFuncReduce3(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: `$reduce(Account.Order.Product.Quantity, $append)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSift(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `
				(
					$data := {
						"one": 1,
						"two": 2,
						"three": 3,
						"four": 4,
						"five": 5,
						"six": 6,
						"seven": 7,
						"eight": 8,
						"nine": 9,
						"ten": 10
					};
					$sift($data, function($v){$v % 2})
				)`,
			Output: map[string]interface{}{
				"one":   float64(1),
				"three": float64(3),
				"five":  float64(5),
				"seven": float64(7),
				"nine":  float64(9),
			},
		},
		{
			Expression: `
				(
					$data := {
						"one": 1,
						"two": 2,
						"three": 3,
						"four": 4,
						"five": 5,
						"six": 6,
						"seven": 7,
						"eight": 8,
						"nine": 9,
						"ten": 10
					};
					$sift($data, function($v,$k){$contains($k,"o")})
				)`,
			Output: map[string]interface{}{
				"one":  float64(1),
				"two":  float64(2),
				"four": float64(4),
			},
		},
		{
			Expression: `
				(
					$data := {
						"one": 1,
						"two": 2,
						"three": 3,
						"four": 4,
						"five": 5,
						"six": 6,
						"seven": 7,
						"eight": 8,
						"nine": 9,
						"ten": 10
					};
					$sift($data, function($v,$k){$length($k) >= $v})
				)`,
			Output: map[string]interface{}{
				"one":   float64(1),
				"two":   float64(2),
				"three": float64(3),
				"four":  float64(4),
			},
		},
	})
}

func TestFuncSift2(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: `$sift(λ($v){$v.**.Postcode})`,
			Output: map[string]interface{}{
				"Address": map[string]interface{}{
					"Street":   "Hursley Park",
					"City":     "Winchester",
					"Postcode": "SO21 2JN",
				},
				"Other": map[string]interface{}{
					"Over 18 ?": true,
					"Misc":      nil,
					"Alternative.Address": map[string]interface{}{
						"Street":   "Brick Lane",
						"City":     "London",
						"Postcode": "E1 6RF",
					},
				},
			},
		},
		{
			Expression: `**[*].$sift(λ($v){$v.Postcode})`,
			Output: []interface{}{
				map[string]interface{}{
					"Address": map[string]interface{}{
						"Street":   "Hursley Park",
						"City":     "Winchester",
						"Postcode": "SO21 2JN",
					},
				},
				map[string]interface{}{
					"Alternative.Address": map[string]interface{}{
						"Street":   "Brick Lane",
						"City":     "London",
						"Postcode": "E1 6RF",
					},
				},
			},
		},
		{
			Expression: []string{
				`$sift(λ($v, $k){$match($k, /^A/)})`,
				`$sift(λ($v, $k){$k ~> /^A/})`,
			},
			Output: map[string]interface{}{
				"Age": float64(28),
				"Address": map[string]interface{}{
					"Street":   "Hursley Park",
					"City":     "Winchester",
					"Postcode": "SO21 2JN",
				},
			},
		},
	})
}

func TestHigherOrderFunctions(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `
				(
					$twice:=function($f){function($x){$f($f($x))}};
					$add3:=function($y){$y+3};
					$add6:=$twice($add3);
					$add6(7)
				)`,
			Output: float64(13),
		},
		{
			Expression: `λ($f) { λ($x) { $x($x) }( λ($g) { $f( (λ($a) {$g($g)($a)}))})}(λ($f) { λ($n) { $n < 2 ? 1 : $n * $f($n - 1) } })(6)`,
			Output:     float64(720),
		},
		{
			Expression: `λ($f) { λ($x) { $x($x) }( λ($g) { $f( (λ($a) {$g($g)($a)}))})}(λ($f) { λ($n) { $n <= 1 ? $n : $f($n-1) + $f($n-2) } })(6)`,
			Output:     float64(8),
		},
	})
}

func TestClosures(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `
				Account.(
					$AccName := function() { $.` + "`Account Name`" + `};
					Order[OrderID = "order104"].Product{
						"Account": $AccName(),
						"SKU-" & $string(ProductID): $.` + "`Product Name`" + `
					}
				)`,
			Output: map[string]interface{}{
				"Account":    "Firefly",
				"SKU-858383": "Bowler Hat",
				"SKU-345664": "Cloak",
			},
		},
	})
}

func TestFuncString(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$string(5)`,
			Output:     "5",
		},
		{
			Expression: `$string(22/7)`,
			Output:     "3.14285715", // TODO: jsonata-js returns "3.142857142857"
		},
		{
			Expression: `$string(1e100)`,
			Output:     "1e+100",
		},
		{
			Expression: `$string(1e-100)`,
			Output:     "1e-100",
		},
		{
			Expression: `$string(1e-6)`,
			Output:     "0.000001",
		},
		{
			Expression: `$string(1e-7)`,
			Output:     "1e-7",
		},
		{
			Expression: `$string(1e+20)`,
			Output:     "100000000000000000000",
		},
		{
			Expression: `$string(1e+21)`,
			Output:     "1e+21",
		},
		{
			Expression: `$string(true)`,
			Output:     "true",
		},
		{
			Expression: `$string(false)`,
			Output:     "false",
		},
		{
			Expression: `$string(null)`,
			Output:     "null",
		},
		{
			Expression: `$string(blah)`,
			Error:      ErrUndefined,
		},
		{
			Expression: []string{
				`$string($string)`,
				`$string(/hat/)`,
				`$string(function(){true})`,
				`$string(function(){1})`,
			},
			Output: "",
		},
		{
			Expression: `$string({"string": "hello"})`,
			Output:     `{"string":"hello"}`,
		},
		{
			Expression: `$string(["string", 5])`,
			Output:     `["string",5]`,
		},
		{
			Expression: `
				$string({
					"string": "hello",
					"number": 78.8 / 2,
					"null":null,
					"boolean": false,
					"function": $sum,
					"lambda": function(){true},
					"object": {
						"str": "another",
						"lambda2": function($n){$n}
					},
					"array": []
				})`,
			// TODO: Can we get this to print in field order like jsonata-js?
			Output: `{"array":[],"boolean":false,"function":"","lambda":"","null":null,"number":39.4,"object":{"lambda2":"","str":"another"},"string":"hello"}`,
			//Output: `{"string":"hello","number":39.4,"null":null,"boolean":false,"function":"","lambda":"","object":{"str":"another","lambda2":""},"array":[]}`,
		},
		{
			Expression: `$string(1/0)`,
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "/",
			},
		},
		{
			Expression: `$string({"inf": 1/0})`,
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "/",
			},
		},
		{
			Expression: `$string(2,3)`,
			Error: &ArgCountError{
				Func:     "string",
				Expected: 1,
				Received: 2,
			},
		},
	})
}

func TestFuncString2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `Account.Order.$string($sum(Product.(Price* Quantity)))`,
			// TODO: jsonata-js rounds to "90.57" and "245.79"
			Output: []interface{}{
				"90.57",
				"245.79",
			},
		},
	})
}

func TestFuncSubstring(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$substring("hello world", 0, 5)`,
			Output:     "hello",
		},
		{
			Expression: []string{
				`$substring("hello world", -5, 5)`,
				`$substring("hello world", 6)`,
			},
			Output: "world",
		},
		{
			Expression: `$substring("hello world", -100, 4)`,
			Output:     "hell",
		},
		{
			Expression: []string{
				`$substring("hello world", 100)`,
				`$substring("hello world", 100, 5)`,
				`$substring("hello world", 0, 0)`,
				`$substring("hello world", 0, -100)`,
				`$substring("超明體繁", 2, 0)`,
			},
			Output: "",
		},
		{
			Expression: []string{
				`$substring("超明體繁", 2)`,
				`$substring("超明體繁", -2)`,
				`$substring("超明體繁", -2, 2)`,
			},
			Output: "體繁",
		},
		{
			Expression: `$substring(nothing, 6)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSubstringBefore(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$substringBefore("Hello World", " ")`,
			Output:     "Hello",
		},
		{
			Expression: `$substringBefore("Hello World", "l")`,
			Output:     "He",
		},
		{
			Expression: `$substringBefore("Hello World", "f")`,
			Output:     "Hello World",
		},
		{
			Expression: `$substringBefore("Hello World", "He")`,
			Output:     "",
		},
		{
			Expression: `$substringBefore("Hello World", "")`,
			Output:     "",
		},
		{
			Expression: `$substringBefore("超明體繁", "體")`,
			Output:     "超明",
		},
		{
			Expression: `$substringBefore(nothing, "He")`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSubstringAfter(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$substringAfter("Hello World", " ")`,
			Output:     "World",
		},
		{
			Expression: `$substringAfter("Hello World", "l")`,
			Output:     "lo World",
		},
		{
			Expression: `$substringAfter("Hello World", "f")`,
			Output:     "Hello World",
		},
		{
			Expression: `$substringAfter("Hello World", "ld")`,
			Output:     "",
		},
		{
			Expression: `$substringAfter("Hello World", "")`,
			Output:     "Hello World",
		},
		{
			Expression: `$substringAfter("超明體繁", "明")`,
			Output:     "體繁",
		},
		{
			Expression: `$substringAfter(nothing, "ld")`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncLowercase(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$lowercase("Hello World")`,
			Output:     "hello world",
		},
		{
			Expression: `$lowercase("Étude in Black")`,
			Output:     "étude in black",
		},
		{
			Expression: `$lowercase(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncUppercase(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$uppercase("Hello World")`,
			Output:     "HELLO WORLD",
		},
		{
			Expression: `$uppercase("étude in black")`,
			Output:     "ÉTUDE IN BLACK",
		},
		{
			Expression: `$uppercase(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncLength(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$length("")`,
			Output:     0,
		},
		{
			Expression: `$length("hello")`,
			Output:     5,
		},
		{
			Expression: `$length(nothing)`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$length("\u03BB-calculus")`,
			Output:     10,
		},
		{
			Expression: `$length("\uD834\uDD1E")`,
			Output:     1,
		},
		{
			Expression: `$length("𝄞")`,
			Output:     1,
		},
		{
			Expression: `$length("超明體繁")`,
			Output:     4,
		},
		{
			Expression: []string{
				`$length("\t")`,
				`$length("\n")`,
			},
			Output: 1,
		},
		{
			Expression: []string{
				`$length(1234)`,
				`$length(true)`,
				`$length(false)`,
				`$length(null)`,
				`$length(1.0)`,
				`$length(["hello"])`,
			},
			Error: &ArgTypeError{
				Func:  "length",
				Which: 1,
			},
		},
		{
			Expression: `$length("hello", "world")`,
			Error: &ArgCountError{
				Func:     "length",
				Expected: 1,
				Received: 2,
			},
		},
	})
}

func TestFuncTrim(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$trim("Hello World")`,
				`$trim("   Hello  \n  \t World  \t ")`,
			},
			Output: "Hello World",
		},
		{
			Expression: "$trim()",
			Error: &ArgCountError{
				Func:     "trim",
				Expected: 1,
				Received: 0,
			},
			Skip: true, // returns ErrUndefined (is it using context?)
		},
	})
}

func TestFuncPad(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$pad("foo", 5)`,
				`$pad("foo", 5, "")`,
				`$pad("foo", 5, " ")`,
			},
			Output: "foo  ",
		},
		{
			Expression: `$pad("foo", -5)`,
			Output:     "  foo",
		},
		{
			Expression: `$pad("foo", -5, ".")`,
			Output:     "..foo",
		},
		{
			Expression: `$pad("foo", 5, "超")`,
			Output:     "foo超超",
		},
		{
			Expression: []string{
				`$pad("foo", 1)`,
				`$pad("foo", -1)`,
			},
			Output: "foo",
		},
		{
			Expression: `$pad("foo", 8, "-+")`,
			Output:     "foo-+-+-",
		},
		{
			Expression: `$pad("超明體繁", 5)`,
			Output:     "超明體繁 ",
		},
		{
			Expression: `$pad("", 6, "超明體繁")`,
			Output:     "超明體繁超明",
		},
		{
			Expression: `$pad(nothing, -1)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncContains(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$contains("Hello World", "lo")`,
				`$contains("Hello World", "World")`,
			},
			Output: true,
		},
		{
			Expression: []string{
				`$contains("Hello World", "Word")`,
				`$contains("Hello World", "world")`,
			},
			Output: false,
		},
		{
			Expression: `$contains("超明體繁", "明體")`,
			Output:     true,
		},
		{
			Expression: `$contains("超明體繁", "體明")`,
			Output:     false,
		},
		{
			Expression: `$contains(nothing, "World")`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$contains(23, 3)`,
			Error: &ArgTypeError{
				Func:  "contains",
				Which: 1,
			},
		},
		{
			Expression: `$contains("23", 3)`,
			Error: &ArgTypeError{
				Func:  "contains",
				Which: 2,
			},
		},
	})
}

func TestFuncSplit(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$split("Hello World", " ")`,
			Output: []string{
				"Hello",
				"World",
			},
		},
		{
			Expression: `$split("Hello  World", " ")`,
			Output: []string{
				"Hello",
				"",
				"World",
			},
		},
		{
			Expression: `$split("Hello", " ")`,
			Output: []string{
				"Hello",
			},
		},
		{
			Expression: `$split("Hello", "")`,
			Output: []string{
				"H",
				"e",
				"l",
				"l",
				"o",
			},
		},
		{
			Expression: `$split("超明體繁", "")`,
			Output: []string{
				"超",
				"明",
				"體",
				"繁",
			},
		},
		{
			Expression: `$sum($split("12345", "").$number($))`,
			Output:     float64(15),
		},
		{
			Expression: []string{
				`$split("a, b, c, d", ", ")`,
				`$split("a, b, c, d", ", ", 10)`,
				//`$split("a, b, c, d", ",").$trim()`,	// returns ErrUndefined
			},
			Output: []string{
				"a",
				"b",
				"c",
				"d",
			},
		},
		{
			Expression: []string{
				`$split("a, b, c, d", ", ", 2)`,
				`$split("a, b, c, d", ", ", 2.5)`,
			},
			Output: []string{
				"a",
				"b",
			},
		},
		{
			Expression: `$split("a, b, c, d", ", ", 0)`,
			Output:     []string{},
		},
		{
			Expression: `$split(nothing, " ")`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$split("a, b, c, d", ", ", -3)`,
			Error:      fmt.Errorf("third argument of the split function must evaluate to a positive number"), // TODO: Use a proper error for this
		},
		{
			Expression: []string{
				`$split("a, b, c, d", ", ", null)`,
				`$split("a, b, c, d", ", ", "2")`,
				`$split("a, b, c, d", ", ", true)`,
			},
			Error: &ArgTypeError{
				Func:  "split",
				Which: 3,
			},
		},
		{
			Expression: `$split(12345, 3)`,
			Error: &ArgTypeError{
				Func:  "split",
				Which: 1,
			},
		},
		{
			Expression: `$split(12345)`,
			Error: &ArgCountError{
				Func:     "split",
				Expected: 3,
				Received: 1,
			},
		},
	})
}

func TestFuncJoin(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				`$join("hello", "")`,
				`$join(["hello"], "")`,
			},
			Output: "hello",
		},
		{
			Expression: `$join(["hello", "world"], "")`,
			Output:     "helloworld",
		},
		{
			Expression: `$join(["hello", "world"], ", ")`,
			Output:     "hello, world",
		},
		{
			Expression: `$join(["超","明","體","繁"])`,
			Output:     "超明體繁",
		},
		{
			Expression: `$join([], ", ")`,
			Output:     "",
		},
		{
			Expression: `$join(true, ", ")`,
			Error:      fmt.Errorf("function join takes an array of strings"), // TODO: Use a proper error
		},
		{
			Expression: `$join([1,2,3], ", ")`,
			Error:      fmt.Errorf("function join takes an array of strings"), // TODO: Use a proper error
		},
		{
			Expression: `$join("hello", 3)`,
			Error: &ArgTypeError{
				Func:  "join",
				Which: 2,
			},
		},
		{
			Expression: `$join()`,
			Error: &ArgCountError{
				Func:     "join",
				Expected: 2,
				Received: 0,
			},
		},
	})
}

func TestFuncJoin2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `$join(Account.Order.Product.Description.Colour, ", ")`,
			Output:     "Purple, Orange, Purple, Black",
		},
		{
			Expression: `$join(Account.Order.Product.Description.Colour, "")`,
			Output:     "PurpleOrangePurpleBlack",
		},
		{
			Expression: `$join(Account.blah.Product.Description.Colour, ", ")`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncReplace(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$replace("Hello World", "World", "Everyone")`,
			Output:     "Hello Everyone",
		},
		{
			Expression: `$replace("the cat sat on the mat", "at", "it")`,
			Output:     "the cit sit on the mit",
		},
		{
			Expression: `$replace("the cat sat on the mat", "at", "it", 0)`,
			Output:     "the cat sat on the mat",
		},
		{
			Expression: `$replace("the cat sat on the mat", "at", "it", 2)`,
			Output:     "the cit sit on the mat",
		},
		{
			Expression: `$replace(nothing, "at", "it", 2)`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$replace("hello")`,
			Error: &ArgCountError{
				Func:     "replace",
				Expected: 4,
				Received: 1,
			},
		},
		{
			Expression: `$replace("hello", 1)`,
			Error: &ArgCountError{
				Func:     "replace",
				Expected: 4,
				Received: 2,
			},
		},
		{
			Expression: `$replace("hello", "l", "1", null)`,
			Error: &ArgTypeError{
				Func:  "replace",
				Which: 4,
			},
		},
		{
			Expression: `$replace(123, 2, 1)`,
			Error: &ArgTypeError{
				Func:  "replace",
				Which: 1,
			},
		},
		{
			Expression: `$replace("hello", 2, 1)`,
			Error: &ArgTypeError{
				Func:  "replace",
				Which: 2,
			},
		},
		{
			Expression: `$replace("hello", "l", "1", -2)`,
			Error:      fmt.Errorf("fourth argument of function replace must evaluate to a positive number"), // TODO: Use a proper error
		},
		{
			Expression: `$replace("hello", "", "bye")`,
			Error:      fmt.Errorf("second argument of function replace can't be an empty string"), // TODO: Use a proper error
		},
	})
}

func TestFormatNumber(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$formatNumber(12345.6, "#,###.00")`,
			Output:     "12,345.60",
		},
		{
			Expression: `$formatNumber(12345678.9, "9,999.99")`,
			Output:     "12,345,678.90",
		},
		{
			Expression: `$formatNumber(123412345678.9, "9,9,99.99")`,
			Output:     "123412345,6,78.90",
		},
		{
			Expression: `$formatNumber(1234.56789, "9,999.999,999")`,
			Output:     "1,234.567,890",
		},
		{
			Expression: `$formatNumber(123.9, "9999")`,
			Output:     "0124",
		},
		{
			Expression: `$formatNumber(0.14, "01%")`,
			Output:     "14%",
		},
		{
			Expression: `$formatNumber(0.4857,"###.###‰")`,
			Output:     "485.7‰",
		},
		{
			Expression: `$formatNumber(0.14, "###pm", {"per-mille": "pm"})`,
			Output:     "140pm",
		},
		{
			Expression: `$formatNumber(-6, "000")`,
			Output:     "-006",
		},
		{
			Expression: `$formatNumber(1234.5678, "00.000e0")`,
			Output:     "12.346e2",
		},
		{
			Expression: `$formatNumber(1234.5678, "00.000e000")`,
			Output:     "12.346e002",
		},
		{
			Expression: `$formatNumber(1234.5678, "①①.①①①e①", {"zero-digit": "\u245f"})`,
			Output:     "①②.③④⑥e②",
		},
		{
			Expression: []string{
				`$formatNumber(1234.5678, "𝟎𝟎.𝟎𝟎𝟎e𝟎", {"zero-digit": "𝟎"})`,
				`$formatNumber(1234.5678, "𝟎𝟎.𝟎𝟎𝟎e𝟎", {"zero-digit": "\ud835\udfce"})`,
			},
			Output: "𝟏𝟐.𝟑𝟒𝟔e𝟐",
		},
		{
			Expression: `$formatNumber(0.234, "0.0e0")`,
			Output:     "2.3e-1",
		},
		{
			Expression: `$formatNumber(0.234, "#.00e0")`,
			Output:     "0.23e0",
		},
		{
			Expression: `$formatNumber(0.123, "#.e9")`,
			Output:     "0.1e0",
		},
		{
			Expression: `$formatNumber(0.234, ".00e0")`,
			Output:     ".23e0",
		},
		{
			Expression: `$formatNumber(2392.14*(-36.58), "000,000.000###;###,###.000###")`,
			Output:     "87,504.4812",
		},
		{
			Expression: `$formatNumber(2.14*86.58,"PREFIX##00.000###SUFFIX")`,
			Output:     "PREFIX185.2812SUFFIX",
		},
		{
			Expression: `$formatNumber(1E20,"#,######")`,
			Output:     "100,000000,000000,000000",
		},

		// TODO: Make proper errors for these.

		{
			Expression: `$formatNumber(20,"#;#;#")`,
			Error:      fmt.Errorf("picture string must contain 1 or 2 subpictures"),
		},
		{
			Expression: `$formatNumber(20,"#.0.0")`,
			Error:      fmt.Errorf("a subpicture cannot contain more than one decimal separator"),
		},
		{
			Expression: `$formatNumber(20,"#0%%")`,
			Error:      fmt.Errorf("a subpicture cannot contain more than one percent character"),
		},
		{
			Expression: `$formatNumber(20,"#0‰‰")`,
			Error:      fmt.Errorf("a subpicture cannot contain more than one per-mille character"),
		},
		{
			Expression: `$formatNumber(20,"#0%‰")`,
			Error:      fmt.Errorf("a subpicture cannot contain both percent and per-mille characters"),
		},
		{
			Expression: `$formatNumber(20,".e0")`,
			Error:      fmt.Errorf("a mantissa part must contain at least one decimal or optional digit"),
		},
		{
			Expression: `$formatNumber(20,"0+.e0")`,
			Error:      fmt.Errorf("a subpicture cannot contain a passive character that is both preceded by and followed by an active character"),
		},
		{
			Expression: `$formatNumber(20,"0,.e0")`,
			Error:      fmt.Errorf("a group separator cannot be adjacent to a decimal separator"),
		},
		{
			Expression: `$formatNumber(20,"0,")`,
			Error:      fmt.Errorf("an integer part cannot end with a group separator"),
		},
		{
			Expression: `$formatNumber(20,"0,,0")`,
			Error:      fmt.Errorf("a subpicture cannot contain adjacent group separators"),
		},
		{
			Expression: `$formatNumber(20,"0#.e0")`,
			Error:      fmt.Errorf("an integer part cannot contain a decimal digit followed by an optional digit"),
		},
		{
			Expression: `$formatNumber(20,"#0.#0e0")`,
			Error:      fmt.Errorf("a fractional part cannot contain an optional digit followed by a decimal digit"),
		},
		{
			Expression: `$formatNumber(20,"#0.0e0%")`,
			Error:      fmt.Errorf("a subpicture cannot contain a percent/per-mille character and an exponent separator"),
		},
		{
			Expression: `$formatNumber(20,"#0.0e0,0")`,
			Error:      fmt.Errorf("an exponent part must consist solely of one or more decimal digits"),
		},
	})
}

func TestFuncFormatBase(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$formatBase(100)",
			Output:     "100",
		},
		{
			Expression: "$formatBase(nothing)",
			Error:      ErrUndefined,
		},
		{
			Expression: []string{
				"$formatBase(100, 2)",
				"$formatBase(99.5, 2.5)",
			},
			Output: "1100100",
		},
		{
			Expression: "$formatBase(-100, 2)",
			Output:     "-1100100",
		},
		{
			Expression: "$formatBase(100, 1)",
			Error:      fmt.Errorf("the second argument to formatBase must be between 2 and 36"),
			/*Error: &EvalError1{
				Errno:    ErrInvalidBase,
				Position: -3,
				Value:    "1",
			},*/
		},
		{
			Expression: "$formatBase(100, 37)",
			Error:      fmt.Errorf("the second argument to formatBase must be between 2 and 36"),
			/*Error: &EvalError1{
				Errno:    ErrInvalidBase,
				Position: -3,
				Value:    "37",
			},*/
		},
	})
}

func TestFuncBase64Encode(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$base64encode("hello:world")`,
			Output:     "aGVsbG86d29ybGQ=",
		},
		{
			Expression: `$base64encode(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncBase64Decode(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$base64decode("aGVsbG86d29ybGQ=")`,
			Output:     "hello:world",
		},
		{
			Expression: `$base64decode(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncNumber(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$number(0)",
				`$number("0")`,
			},
			Output: float64(0),
		},
		{
			Expression: []string{
				"$number(10)",
				`$number("10")`,
			},
			Output: float64(10),
		},
		{
			Expression: []string{
				"$number(-0.05)",
				`$number("-0.05")`,
			},
			Output: -0.05,
		},
		{
			Expression: `$number("1e2")`,
			Output:     float64(100),
		},
		{
			Expression: `$number("-1e2")`,
			Output:     float64(-100),
		},
		{
			Expression: `$number("1.0e-2")`,
			Output:     0.01,
		},
		{
			Expression: `$number("1e0")`,
			Output:     float64(1),
		},
		{
			Expression: `$number("10e500")`,
			Error:      fmt.Errorf("unable to cast %q to a number", "10e500"),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "10e500",
			},*/
		},
		{
			Expression: `$number("Hello world")`,
			Error:      fmt.Errorf("unable to cast %q to a number", "Hello world"),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "Hello world",
			},*/
		},
		{
			Expression: `$number("1/2")`,
			Error:      fmt.Errorf("unable to cast %q to a number", "1/2"),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "1/2",
			},*/
		},
		{
			Expression: `$number("1234 hello")`,
			Error:      fmt.Errorf("unable to cast %q to a number", "1234 hello"),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "1234 hello",
			},*/
		},
		{
			Expression: `$number("")`,
			Error:      fmt.Errorf("unable to cast %q to a number", ""),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "",
			},*/
		},
		{
			Expression: `$number("[1]")`,
			Error:      fmt.Errorf("unable to cast %q to a number", "[1]"),
			/*Error: &EvalError1{
				Errno:    ErrCastNumber,
				Position: -10,
				Value:    "[1]",
			},*/
		},

		{
			Expression: `$number(true)`,
			Output:     1.,
		},
		{
			Expression: `$number(false)`,
			Output:     0.,
		},
		{
			Expression: `$number(null)`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number([])`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number([1,2])`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number(["hello"])`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number(["2"])`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number({})`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number({"hello":"world"})`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number($number)`,
			Error: &ArgTypeError{
				Func:  "number",
				Which: 1,
			},
		},
		{
			Expression: `$number(1,2)`,
			Error: &ArgCountError{
				Func:     "number",
				Expected: 1,
				Received: 2,
			},
		},
		{
			Expression: `$number(nothing)`,
			Error:      ErrUndefined,
		},
	})
}

func TestFuncAbs(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: []string{
				"$abs(3.7)",
				"$abs(-3.7)",
			},
			Output: 3.7,
		},
		{
			Expression: "$abs(0)",
			Output:     float64(0),
		},
		{
			Expression: "$abs(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncFloor(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$floor(3.7)",
			Output:     float64(3),
		},
		{
			Expression: "$floor(-3.7)",
			Output:     float64(-4),
		},
		{
			Expression: "$floor(0)",
			Output:     float64(0),
		},
		{
			Expression: "$floor(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncCeil(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$ceil(3.7)",
			Output:     float64(4),
		},
		{
			Expression: "$ceil(-3.7)",
			Output:     float64(-3),
		},
		{
			Expression: "$ceil(0)",
			Output:     float64(0),
		},
		{
			Expression: "$ceil(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncRound(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$round(4)",
			Output:     float64(4),
		},
		{
			Expression: "$round(2.3)",
			Output:     float64(2),
		},
		{
			Expression: "$round(2.7)",
			Output:     float64(3),
		},
		{
			Expression: "$round(2.5)",
			Output:     float64(2),
		},
		{
			Expression: "$round(3.5)",
			Output:     float64(4),
		},
		{
			Expression: []string{
				"$round(-0.5)",
				"$round(-0.3)",
				"$round(0.5)",
			},
			Output: float64(0),
		},
		{
			Expression: []string{
				"$round(-7.5)",
				"$round(-8.5)",
			},
			Output: float64(-8),
		},
		{
			Expression: "$round(4.49, 1)",
			Output:     float64(4.5),
		},
		{
			Expression: "$round(4.525, 2)",
			Output:     float64(4.52),
		},
		{
			Expression: "$round(4.515, 2)",
			Output:     float64(4.52),
		},
		{
			Expression: "$round(12345, -2)",
			Output:     float64(12300),
		},
		{
			Expression: []string{
				"$round(12450, -2)",
				"$round(12350, -2)",
			},
			Output: float64(12400),
		},
		{
			Expression: "$round(6.022e-23, 24)",
			Output:     6.0e-23,
		},
		{
			Expression: "$round(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSqrt(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$sqrt(4)",
			Output:     float64(2),
		},
		{
			Expression: "$sqrt(2)",
			Output:     math.Sqrt2,
		},
		{
			Expression: "$sqrt(-2)",
			Error:      fmt.Errorf("the sqrt function cannot be applied to a negative number"),
		},
		{
			Expression: "$sqrt(nothing)",
			Error:      ErrUndefined,
		},
	})
}

func TestFuncSqrt2(t *testing.T) {

	runTestCasesFunc(t, equalFloats(1e-13), nil, []*testCase{
		{
			Expression: "$sqrt(10) * $sqrt(10)",
			Output:     float64(10),
		},
	})
}

func TestFuncPower(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "$power(4,2)",
			Output:     float64(16),
		},
		{
			Expression: "$power(4,0.5)",
			Output:     float64(2),
		},
		{
			Expression: "$power(10,-2)",
			Output:     0.01,
		},
		{
			Expression: "$power(-2,3)",
			Output:     float64(-8),
		},
		{
			Expression: "$power(nothing,3)",
			Error:      ErrUndefined,
		},
		{
			Expression: "$power(-2,1/3)",
			Error:      fmt.Errorf("the power function has resulted in a value that cannot be represented as a JSON number"),
		},
		{
			Expression: "$power(100,1000)",
			Error:      fmt.Errorf("the power function has resulted in a value that cannot be represented as a JSON number"),
		},
	})
}

func TestFuncRandom(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "($x := $random(); $x >= 0 and $x < 1)",
			Output:     true,
		},
		{
			Expression: "$random() = $random()",
			Output:     false,
		},
	})
}

func TestFuncKeys(t *testing.T) {

	runTestCasesFunc(t, equalArraysUnordered, testdata.account, []*testCase{
		{
			Expression: "$keys(Account)",
			Output: []string{
				"Account Name",
				"Order",
			},
		},
		{
			Expression: "$keys(Account.Order.Product)",
			Output: []string{
				"Product Name",
				"ProductID",
				"SKU",
				"Description",
				"Price",
				"Quantity",
			},
		},
	})
}

func TestFuncKeys2(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$keys({"foo":{}})`,
			Output:     "foo",
			/*Output: []string{
				"foo",
			},*/
		},
		{
			Expression: []string{
				"$keys({})",
				`$keys("foo")`,
				`$keys(function(){1})`,
				`$keys(["foo", "bar"])`,
			},
			Error: ErrUndefined,
		},
	})
}

func TestFuncLookup(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `$lookup(Account, "Account Name")`,
			Output:     "Firefly",
		},
		{
			Expression: `$lookup(Account.Order.Product, "Product Name")`,
			Output: []interface{}{
				"Bowler Hat",
				"Trilby hat",
				"Bowler Hat",
				"Cloak",
			},
		},
		{
			Expression: `$lookup(Account.Order.Product.ProductID, "Product Name")`,
			Error:      ErrUndefined,
			Skip:       true, // returns a type error instead of ErrUndefined
		},
	})
}

func TestFuncLookup2(t *testing.T) {

	data := map[string]interface{}{
		"temp":      22.7,
		"wind":      7,
		"gust":      nil,
		"timestamp": 1508971317377,
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: []string{
				`$lookup($, "gust")`,
				`$lookup($$, "gust")`,
			},
			Output: nil,
		},
	})
}

func TestDefaultContext(t *testing.T) {

	runTestCases(t, "5", []*testCase{
		{
			Expression: "$number()",
			Output:     float64(5),
		},
	})
}

func TestDefaultContext2(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: "[1..5].$string()",
			Output: []interface{}{
				"1",
				"2",
				"3",
				"4",
				"5",
			},
		},
		{
			Expression: `[1..5].("Item " & $string())`,
			Output: []interface{}{
				"Item 1",
				"Item 2",
				"Item 3",
				"Item 4",
				"Item 5",
			},
		},
	})
}

func TestDefaultContext3(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `Account.Order.Product.` + "`Product Name`" + `.$uppercase().$substringBefore(" ")`,
			Output: []interface{}{
				"BOWLER",
				"TRILBY",
				"BOWLER",
				"CLOAK",
			},
		},
	})
}

func TestApplyOperator(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `
				(
					$uppertrim := $trim ~> $uppercase;
					$uppertrim("   Hello    World   ")
				)`,
			Output: "HELLO WORLD",
		},
		{
			Expression: `"john@example.com" ~> $substringAfter("@") ~> $substringBefore(".")`,
			Output:     "example",
		},
		{
			Expression: `
				(
					$domain := $substringAfter(?,"@") ~> $substringBefore(?,".");
					$domain("john@example.com")
				)`,
			Output: "example",
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					[1..5] ~> $map($square)
				)`,
			Output: []interface{}{
				float64(1),
				float64(4),
				float64(9),
				float64(16),
				float64(25),
			},
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					[1..5] ~> $map($square) ~> $sum()
				)`,
			Output: float64(55),
		},
		{
			Expression: `
				(
					$betweenBackets := $substringAfter(?, "(") ~> $substringBefore(?, ")");
					$betweenBackets("test(foo)bar")
				)`,
			Output: "foo",
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					$chain := λ($f, $g){λ($x){$g($f($x))}};
					$instructions := [$sum, $square];
					$sumsq := $instructions ~> $reduce($chain);
					[1..5] ~> $sumsq()
				)`,
			Output: float64(225),
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					$chain := λ($f, $g){λ($x){ $x ~> $f ~> $g }};
					$instructions := [$sum, $square, $string];
					$sumsq := $instructions ~> $reduce($chain);
					[1..5] ~> $sumsq()
				)`,
			Output: "225",
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					$instructions := $sum ~> $square;
					[1..5] ~> $instructions()
				)`,
			Output: float64(225),
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					$sum_of_squares := $map(?, $square) ~> $sum;
					[1..5] ~> $sum_of_squares()
				)`,
			Output: float64(55),
		},
		{
			Expression: `
				(
					$times := λ($x, $y) { $x * $y };
					$product := $reduce(?, $times);
					$square := function($x){$x*$x};
					$product_of_squares := $map(?, $square) ~> $product;
					[1..5] ~> $product_of_squares()
				)`,
			Output: float64(14400),
		},
		{
			Expression: `
				(
					$square := function($x){$x*$x};
					[1..5] ~> $map($square) ~> $reduce(λ($x, $y) { $x * $y });
				)`,
			Output: float64(14400),
		},
		{
			Expression: `"" ~> $substringAfter("@") ~> $substringBefore(".")`,
			Output:     "",
		},
		{
			Expression: `foo ~> $substringAfter("@") ~> $substringBefore(".")`,
			Error:      ErrUndefined,
		},
	})
}

func TestApplyOperator2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order[0].OrderID ~> $uppercase()",
			Output:     "ORDER103",
		},
		{
			Expression: "Account.Order[0].OrderID ~> $uppercase() ~> $lowercase()",
			Output:     "order103",
		},
		{
			Expression: "Account.Order.OrderID ~> $join()",
			Output:     "order103order104",
		},
		{
			Expression: `Account.Order.OrderID ~> $join(", ")`,
			Output:     "order103, order104",
		},
		{
			Expression: "Account.Order.Product.(Price * Quantity) ~> $sum()",
			Output:     336.36,
		},
		{
			Expression: `
				(
					$prices := Account.Order.Product.Price;
					$quantities := Account.Order.Product.Quantity;
					$product := λ($arr) { $arr[0] * $arr[1] };
					$zip($prices, $quantities) ~> $map($product) ~> $sum()
				)`,
			Output: 336.36,
		},
		{
			Expression: `42 ~> "hello"`,
			Error: &EvalError{
				Type:  ErrNonCallableApply,
				Token: `"hello"`,
				Value: "~>",
			},
		},
	})
}

func TestTransformOperator(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `$ ~> |Account.Order.Product|{"Total":Price*Quantity},["Description", "SKU"]|`,
			Output: map[string]interface{}{
				"Account": map[string]interface{}{
					"Account Name": "Firefly",
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "order103",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"Price":        34.45,
									"Quantity":     float64(2),
									"Total":        68.9,
								},
								map[string]interface{}{
									"Product Name": "Trilby hat",
									"ProductID":    float64(858236),
									"Price":        21.67,
									"Quantity":     float64(1),
									"Total":        21.67,
								},
							},
						},
						map[string]interface{}{
							"OrderID": "order104",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"Price":        34.45,
									"Quantity":     float64(4),
									"Total":        137.8,
								},
								map[string]interface{}{
									"ProductID":    float64(345664),
									"Product Name": "Cloak",
									"Price":        107.99,
									"Quantity":     float64(1),
									"Total":        107.99,
								},
							},
						},
					},
				},
			},
		},
		{
			Expression: `Account.Order ~> |Product|{"Total":Price*Quantity},["Description", "SKU"]|`,
			Output: []interface{}{
				map[string]interface{}{
					"OrderID": "order103",
					"Product": []interface{}{
						map[string]interface{}{
							"Product Name": "Bowler Hat",
							"ProductID":    float64(858383),
							"Price":        34.45,
							"Quantity":     float64(2),
							"Total":        68.9,
						},
						map[string]interface{}{
							"Product Name": "Trilby hat",
							"ProductID":    float64(858236),
							"Price":        21.67,
							"Quantity":     float64(1),
							"Total":        21.67,
						},
					},
				},
				map[string]interface{}{
					"OrderID": "order104",
					"Product": []interface{}{
						map[string]interface{}{
							"Product Name": "Bowler Hat",
							"ProductID":    float64(858383),
							"Price":        34.45,
							"Quantity":     float64(4),
							"Total":        137.8,
						},
						map[string]interface{}{
							"ProductID":    float64(345664),
							"Product Name": "Cloak",
							"Price":        107.99,
							"Quantity":     float64(1),
							"Total":        107.99,
						},
					},
				},
			},
		},
		{
			Expression: `$ ~> |Account.Order.Product|{"Total":Price*Quantity, "Price": Price * 1.2}|`,
			Output: map[string]interface{}{
				"Account": map[string]interface{}{
					"Account Name": "Firefly",
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "order103",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "0406654608",
									"Description": map[string]interface{}{
										"Colour": "Purple",
										"Width":  float64(300),
										"Height": float64(200),
										"Depth":  float64(210),
										"Weight": 0.75,
									},
									"Price":    41.34,
									"Quantity": float64(2),
									"Total":    68.9,
								},
								map[string]interface{}{
									"Product Name": "Trilby hat",
									"ProductID":    float64(858236),
									"SKU":          "0406634348",
									"Description": map[string]interface{}{
										"Colour": "Orange",
										"Width":  float64(300),
										"Height": float64(200),
										"Depth":  float64(210),
										"Weight": 0.6,
									},
									"Price":    26.004,
									"Quantity": float64(1),
									"Total":    21.67,
								},
							},
						},
						map[string]interface{}{
							"OrderID": "order104",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "040657863",
									"Description": map[string]interface{}{
										"Colour": "Purple",
										"Width":  float64(300),
										"Height": float64(200),
										"Depth":  float64(210),
										"Weight": 0.75,
									},
									"Price":    41.34,
									"Quantity": float64(4),
									"Total":    137.8,
								},
								map[string]interface{}{
									"ProductID":    float64(345664),
									"SKU":          "0406654603",
									"Product Name": "Cloak",
									"Description": map[string]interface{}{
										"Colour": "Black",
										"Width":  float64(30),
										"Height": float64(20),
										"Depth":  float64(210),
										"Weight": float64(2),
									},
									"Price":    129.588,
									"Quantity": float64(1),
									"Total":    107.99,
								},
							},
						},
					},
				},
			},
		},
		{
			Expression: []string{
				`$ ~> |Account.Order.Product|{},"Description"|`,
				`$ ~> |Account.Order.Product|nomatch,"Description"|`,
			},
			Output: map[string]interface{}{
				"Account": map[string]interface{}{
					"Account Name": "Firefly",
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "order103",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "0406654608",
									"Price":        34.45,
									"Quantity":     float64(2),
								},
								map[string]interface{}{
									"Product Name": "Trilby hat",
									"ProductID":    float64(858236),
									"SKU":          "0406634348",
									"Price":        21.67,
									"Quantity":     float64(1),
								},
							},
						},
						map[string]interface{}{
							"OrderID": "order104",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "040657863",
									"Price":        34.45,
									"Quantity":     float64(4),
								},
								map[string]interface{}{
									"ProductID":    float64(345664),
									"SKU":          "0406654603",
									"Product Name": "Cloak",
									"Price":        107.99,
									"Quantity":     float64(1),
								},
							},
						},
					},
				},
			},
		},
		{
			Expression: `$ ~> |(Account.Order.Product)[0]|{"Description":"blah"}|`,
			Output: map[string]interface{}{
				"Account": map[string]interface{}{
					"Account Name": "Firefly",
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "order103",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "0406654608",
									"Description":  "blah",
									"Price":        34.45,
									"Quantity":     float64(2),
								},
								map[string]interface{}{
									"Product Name": "Trilby hat",
									"ProductID":    float64(858236),
									"SKU":          "0406634348",
									"Description": map[string]interface{}{
										"Colour": "Orange",
										"Width":  float64(300),
										"Height": float64(200),
										"Depth":  float64(210),
										"Weight": 0.6,
									},
									"Price":    21.67,
									"Quantity": float64(1),
								},
							},
						},
						map[string]interface{}{
							"OrderID": "order104",
							"Product": []interface{}{
								map[string]interface{}{
									"Product Name": "Bowler Hat",
									"ProductID":    float64(858383),
									"SKU":          "040657863",
									"Description": map[string]interface{}{
										"Colour": "Purple",
										"Width":  float64(300),
										"Height": float64(200),
										"Depth":  float64(210),
										"Weight": 0.75,
									},
									"Price":    34.45,
									"Quantity": float64(4),
								},
								map[string]interface{}{
									"ProductID":    float64(345664),
									"SKU":          "0406654603",
									"Product Name": "Cloak",
									"Description": map[string]interface{}{
										"Colour": "Black",
										"Width":  float64(30),
										"Height": float64(20),
										"Depth":  float64(210),
										"Weight": float64(2),
									},
									"Price":    107.99,
									"Quantity": float64(1),
								},
							},
						},
					},
				},
			},
		},
		{
			Expression: `Account ~> |Order|{"Product":"blah"},nomatch|`,
			Output: map[string]interface{}{
				"Account Name": "Firefly",
				"Order": []interface{}{
					map[string]interface{}{
						"OrderID": "order103",
						"Product": "blah",
					},
					map[string]interface{}{
						"OrderID": "order104",
						"Product": "blah",
					},
				},
			},
		},
		{
			Expression: `$ ~> |foo.bar|{"Description":"blah"}|`,
			Output:     testdata.account,
		},
		{
			Expression: `foo ~> |foo.bar|{"Description":"blah"}|`,
			Error:      ErrUndefined,
		},
		{
			Expression: `Account ~> |Order|5|`,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: "5",
			},
		},
		{
			Expression: `Account ~> |Order|"blah"|`,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: `"blah"`,
			},
		},
		{
			Expression: `Account ~> |Order|[]|`,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: "[]",
			},
		},
		{
			Expression: `Account ~> |Order|null|`,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: "null",
			},
		},
		{
			Expression: `Account ~> |Order|false|`,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: "false",
			},
		},
		{
			Expression: `Account ~> |Order|{},5|`,
			Error: &EvalError{
				Type:  ErrIllegalDelete,
				Token: "5",
			},
		},
		{
			Expression: `Account ~> |Order|{},{}|`,
			Error: &EvalError{
				Type:  ErrIllegalDelete,
				Token: "{}",
			},
		},
		{
			Expression: `Account ~> |Order|{},null|`,
			Error: &EvalError{
				Type:  ErrIllegalDelete,
				Token: "null",
			},
		},
		{
			Expression: `Account ~> |Order|{},[1,2,3]|`,
			Error: &EvalError{
				Type:  ErrIllegalDelete,
				Token: "[1, 2, 3]",
			},
		},
	})
}

func TestRegex(t *testing.T) {

	runTestCasesFunc(t, equalRegexMatches, nil, []*testCase{
		{
			Expression: `/ab/ ("ab")`,
			Output: map[string]interface{}{
				"match":  "ab",
				"start":  0,
				"end":    2,
				"groups": []string{},
			},
		},
		{
			Expression: `/ab/ ()`,
			Error:      ErrUndefined,
		},
		{
			Expression: `/ab+/ ("ababbabbcc")`,
			Output: map[string]interface{}{
				"match":  "ab",
				"start":  0,
				"end":    2,
				"groups": []string{},
			},
		},
		{
			Expression: `/a(b+)/ ("ababbabbcc")`,
			Output: map[string]interface{}{
				"match": "ab",
				"start": 0,
				"end":   2,
				"groups": []string{
					"b",
				},
			},
		},
		{
			Expression: `/a(b+)/ ("ababbabbcc").next()`,
			Output: map[string]interface{}{
				"match": "abb",
				"start": 2,
				"end":   5,
				"groups": []string{
					"bb",
				},
			},
		},
		{
			Expression: `/a(b+)/ ("ababbabbcc").next().next()`,
			Output: map[string]interface{}{
				"match": "abb",
				"start": 5,
				"end":   8,
				"groups": []string{
					"bb",
				},
			},
		},
		{
			Expression: `/a(b+)/ ("ababbabbcc").next().next().next()`,
			Error:      ErrUndefined,
		},
		{
			Expression: []string{
				`/a(b+)/i ("Ababbabbcc")`,
				`/(?i)a(b+)/ ("Ababbabbcc")`,
			},
			Output: map[string]interface{}{
				"match": "Ab",
				"start": 0,
				"end":   2,
				"groups": []string{
					"b",
				},
			},
		},
		{
			Expression: `//`,
			Error: &jparse.Error{
				Type:     jparse.ErrEmptyRegex,
				Position: 1,
			},
		},
		{
			Expression: `/`,
			Error: &jparse.Error{
				Type:     jparse.ErrUnterminatedRegex,
				Position: 1,
				Hint:     "/",
			},
		},
	})
}

func TestRegex2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`Account.Order.Product[$.` + "`Product Name`" + ` ~> /hat/i].ProductID`,
				`Account.Order.Product[$.` + "`Product Name`" + ` ~> /(?i)hat/].ProductID`,
			},
			Output: []interface{}{
				float64(858383),
				float64(858236),
				float64(858383),
			},
		},
	})
}

func TestRegexMatch(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$match("ababbabbcc",/ab/)`,
			Output: []map[string]interface{}{
				{
					"match":  "ab",
					"index":  0,
					"groups": []string{},
				},
				{
					"match":  "ab",
					"index":  2,
					"groups": []string{},
				},
				{
					"match":  "ab",
					"index":  5,
					"groups": []string{},
				},
			},
		},
		{
			Expression: `$match("ababbabbcc",/a(b+)/)`,
			Output: []map[string]interface{}{
				{
					"match": "ab",
					"index": 0,
					"groups": []string{
						"b",
					},
				},
				{
					"match": "abb",
					"index": 2,
					"groups": []string{
						"bb",
					},
				},
				{
					"match": "abb",
					"index": 5,
					"groups": []string{
						"bb",
					},
				},
			},
		},
		{
			Expression: `$match("ababbabbcc",/a(b+)/, 1)`,
			Output: []map[string]interface{}{
				{
					"match": "ab",
					"index": 0,
					"groups": []string{
						"b",
					},
				},
			},
		},
		{
			Expression: []string{
				`$match("ababbabbcc",/a(b+)/, 0)`,
				`$match("ababbabbcc",/a(xb+)/)`,
			},
			Output: []map[string]interface{}{},
		},
		{
			Expression: `$match(nothing,/a(xb+)/)`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$match("a, b, c, d", /ab/, -3)`,
			Error:      fmt.Errorf("third argument of function match must evaluate to a positive number"), // TODO: use a proper error
		},
		{
			Expression: `$match(12345, 3)`,
			Error: &ArgTypeError{
				Func:  "match",
				Which: 1,
			},
		},
		{
			Expression: []string{
				`$match("a, b, c, d", "ab")`,
				`$match("a, b, c, d", true)`,
			},
			Error: &ArgTypeError{
				Func:  "match",
				Which: 2,
			},
		},
		{
			Expression: []string{
				`$match("a, b, c, d", /ab/, null)`,
				`$match("a, b, c, d", /ab/, "2")`,
			},
			Error: &ArgTypeError{
				Func:  "match",
				Which: 3,
			},
		},
		{
			Expression: `$match(12345)`,
			Error: &ArgCountError{
				Func:     "match",
				Expected: 3,
				Received: 1,
			},
		},
	})
}

func TestRegexReplace(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$replace("ababbxabbcc",/b+/, "yy")`,
			Output:     "ayyayyxayycc",
		},
		{
			Expression: `$replace("ababbxabbcc",/b+/, "yy", 2)`,
			Output:     "ayyayyxabbcc",
		},
		{
			Expression: `$replace("ababbxabbcc",/b+/, "yy", 0)`,
			Output:     "ababbxabbcc",
		},
		{
			Expression: `$replace("ababbxabbcc",/d+/, "yy")`,
			Output:     "ababbxabbcc",
		},
		{
			Expression: `$replace("John Smith", /(\w+)\s(\w+)/, "$2, $1")`,
			Output:     "Smith, John",
		},
		{
			Expression: `$replace("265USD", /([0-9]+)USD/, "$$$1")`,
			Output:     "$265",
		},
		{
			Expression: `$replace("265USD", /([0-9]+)USD/, "$w")`,
			Output:     "$w",
		},
		{
			Expression: `$replace("265USD", /([0-9]+)USD/, "$0 -> $$$1")`,
			Output:     "265USD -> $265",
		},
		{
			Expression: `$replace("265USD", /([0-9]+)USD/, "$0$1$2")`,
			Output:     "265USD265",
		},
		{
			Expression: `$replace("abcd", /(ab)|(a)/, "[1=$1][2=$2]")`,
			Output:     "[1=ab][2=]cd",
		},
		{
			Expression: `$replace("abracadabra", /bra/, "*")`,
			Output:     "a*cada*",
		},
		{
			Expression: `$replace("abracadabra", /a.*a/, "*")`,
			Output:     "*",
		},
		{
			Expression: `$replace("abracadabra", /a.*?a/, "*")`,
			Output:     "*c*bra",
		},
		{
			Expression: `$replace("abracadabra", /a/, "")`,
			Output:     "brcdbr",
		},
		{
			Expression: `$replace("abracadabra", /a(.)/, "a$1$1")`,
			Output:     "abbraccaddabbra",
		},
		{
			Expression: `$replace("abracadabra", /.*?/, "$1")`,
			Skip:       true, // jsonata-js throws error D1004
		},
		{
			Expression: `$replace("AAAA", /A+/, "b")`,
			Output:     "b",
		},
		{
			Expression: `$replace("AAAA", /A+?/, "b")`,
			Output:     "bbbb",
		},
		{
			Expression: `$replace("darted", /^(.*?)d(.*)$/, "$1c$2")`,
			Output:     "carted",
		},
		{
			Expression: `$replace("abcdefghijklmno", /(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)(k)(l)(m)/, "$8$5$12$12$18$123")`,
			Output:     "hella8l3no",
		},
		{
			Expression: `$replace("abcdefghijklmno", /xyz/, "$8$5$12$12$18$123")`,
			Output:     "abcdefghijklmno",
		},
		{
			Expression: `$replace("abcdefghijklmno", /ijk/, "$8$5$12$12$18$123")`,
			Output:     "abcdefgh22823lmno",
		},
		{
			Expression: `$replace("abcdefghijklmno", /(ijk)/, "$8$5$12$12$18$123")`,
			Output:     "abcdefghijk2ijk2ijk8ijk23lmno",
		},
		{
			Expression: `$replace("abcdefghijklmno", /ijk/, "$x")`,
			Output:     "abcdefgh$xlmno",
		},
		{
			Expression: `$replace("abcdefghijklmno", /(ijk)/, "$x$")`,
			Output:     "abcdefgh$x$lmno",
		},
	})
}

func TestRegexReplace2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: []string{
				`Account.Order.Product.$replace($.` + "`Product Name`" + `, /hat/i, function($match) { "foo" })`,
				`Account.Order.Product.$replace($.` + "`Product Name`" + `, /(?i)hat/, function($match) { "foo" })`,
			},
			Output: []interface{}{
				"Bowler foo",
				"Trilby foo",
				"Bowler foo",
				"Cloak",
			},
		},
		{
			Expression: []string{
				`Account.Order.Product.$replace($.` + "`Product Name`" + `, /(h)(at)/i, function($match) { $uppercase($match.match) })`,
				`Account.Order.Product.$replace($.` + "`Product Name`" + `, /(?i)(h)(at)/, function($match) { $uppercase($match.match) })`,
			},
			Output: []interface{}{
				"Bowler HAT",
				"Trilby HAT",
				"Bowler HAT",
				"Cloak",
			},
		},
		{
			Expression: `Account.Order.Product.$replace($.` + "`Product Name`" + `, /(?i)hat/,
				function($match) { true })`,
			Error: fmt.Errorf("third argument of function replace must be a function that returns a string"),
		},
		{
			Expression: `Account.Order.Product.$replace($.` + "`Product Name`" + `, /(?i)hat/,
				function($match) { 42 })`,
			Error: fmt.Errorf("third argument of function replace must be a function that returns a string"),
		},
	})
}

func TestRegexReplace3(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$replace("temperature = 68F today", /(-?\d+(?:\.\d*)?)F\b/,
				function($m) { ($number($m.groups[0]) - 32) * 5/9 & "C" })`,
			Output: "temperature = 20C today",
		},
	})
}

func TestRegexContains(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$contains("ababbxabbcc", /ab+/)`,
			Output:     true,
		},
		{
			Expression: `$contains("ababbxabbcc", /ax+/)`,
			Output:     false,
		},
	})
}

func TestRegexContains2(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: "Account.Order.Product[$contains(`Product Name`, /hat/)].ProductID",
			Output:     float64(858236),
		},
		{
			Expression: []string{
				"Account.Order.Product[$contains(`Product Name`, /hat/i)].ProductID",
				"Account.Order.Product[$contains(`Product Name`, /(?i)hat/)].ProductID",
			},
			Output: []interface{}{
				float64(858383),
				float64(858236),
				float64(858383),
			},
		},
	})
}

func TestRegexSplit(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$split("ababbxabbcc",/b+/)`,
			Output: []string{
				"a",
				"a",
				"xa",
				"cc",
			},
		},
		{
			Expression: `$split("ababbxabbcc",/b+/, 2)`,
			Output: []string{
				"a",
				"a",
			},
		},
		{
			Expression: `$split("ababbxabbcc",/d+/)`,
			Output: []string{
				"ababbxabbcc",
			},
		},
	})
}

var reNow = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d.\d\d\dZ$`)

func TestFuncNow(t *testing.T) {

	expr, err := Compile("$now()")
	if err != nil {
		t.Fatalf("Compile failed: %s", err)
	}

	var results [2]string

	for i := range results {

		output, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Eval failed: %s", err)
		}

		results[i] = output.(string)
		// $now() returns a timestamp that includes milliseconds, so
		// sleeping for 1ms should be enough to reliably produce a
		// different result.
		time.Sleep(1 * time.Millisecond)
	}

	for _, s := range results {
		if !reNow.MatchString(s) {
			t.Errorf("Timestamp %q does not match expected regex %q", s, reNow)
		}
	}

	if results[0] == results[1] {
		t.Errorf("calling $now() %d times returned identical timestamps: %q", len(results), results[0])
	}
}

func TestFuncNow2(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `{"now": $now(), "delay": $sum([1..10000]), "later": $now()}.(now = later)`,
			Output:     true,
		},
		{
			Expression: `$now()`,
			Exts: map[string]Extension{
				"now": {
					Func: func() string {
						return "time for tea"
					},
				},
			},
			Output: "time for tea",
		},
	})
}

func TestFuncMillis(t *testing.T) {

	expr, err := Compile("$millis()")
	if err != nil {
		t.Fatalf("Compile failed: %s", err)
	}

	var results [2]int64

	for i := range results {

		output, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Eval failed: %s", err)
		}

		results[i] = output.(int64)
		// $millis() returns the unix time in milliseconds, so
		// sleeping for 1ms should be enough to reliably produce
		// a different result.
		time.Sleep(1 * time.Millisecond)
	}

	for _, ms := range results {
		if ms <= 1502264152715 || ms >= 2000000000000 {
			t.Errorf("Unix time %d does not fall between expected values 1502264152715 and 2000000000000", ms)
		}
	}

	if results[0] == results[1] {
		t.Errorf("calling $millis() %d times returned identical unix times: %d", len(results), results[0])
	}
}

func TestFuncMillis2(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `{"now": $millis(), "delay": $sum([1..10000]), "later": $millis()}.(now = later)`,
			Output:     true,
		},
	})
}

func TestFuncToMillis(t *testing.T) {
	defer func() { // added this to help with the test as it panics and that is annoying
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$toMillis("1970-01-01T00:00:00.001Z")`,
			Output:     int64(1),
		},
		{
			Expression: `$toMillis("2017-10-30T16:25:32.935Z")`,
			Output:     int64(1509380732935),
		},
		{
			Expression: `$toMillis(foo)`,
			Error:      ErrUndefined,
		},
		{
			Expression: `$toMillis("foo")`,
			Error:      fmt.Errorf(`could not parse time "foo" due to inconsistency in layout and date time string, date foo layout 2006`),
		},
	})
}

func TestFuncFromMillis(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `$fromMillis(1)`,
			Output:     "1970-01-01T00:00:00.001Z",
		},
		{
			Expression: `$fromMillis(1509380732935)`,
			Output:     "2017-10-30T16:25:32.935Z",
		},
		{
			Expression: `$fromMillis(foo)`,
			Error:      ErrUndefined,
		},
	})
}

func TestLambdaSignatures(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `λ($arg)<b:b>{$not($arg)}(true)`,
			Output:     false,
		},
		{
			Expression: `λ($arg)<b:b>{$not($arg)}(foo)`,
			Output:     true,
		},
		{
			Expression: `λ($arg)<x:b>{$not($arg)}(null)`,
			Output:     true,
		},
		{
			Expression: `function($x,$y)<n-n:n>{$x+$y}(2, 6)`,
			Output:     float64(8),
		},
		{
			Expression: `[1..5].function($x,$y)<n-n:n>{$x+$y}(2, 6)`,
			Output: []interface{}{
				float64(8),
				float64(8),
				float64(8),
				float64(8),
				float64(8),
			},
		},
		{
			Expression: `[1..5].function($x,$y)<n-n:n>{$x+$y}(6)`,
			Output: []interface{}{
				float64(7),
				float64(8),
				float64(9),
				float64(10),
				float64(11),
			},
		},
		{
			Expression: `λ($str)<s->{$uppercase($str)}("hello")`,
			Output:     "HELLO",
		},
		{
			Expression: `λ($str, $prefix)<s-s>{$prefix & $str}("World", "Hello ")`,
			Output:     "Hello World",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}("a")`,
			Output:     "a",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}("a", "-")`,
			Output:     "a",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}(["a"], "-")`,
			Output:     "a",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}(["a", "b"], "-")`,
			Output:     "a-b",
		},
		{
			Expression: `λ($arr, $sep)<as?:s>{$join($arr, $sep)}(["a", "b"], "-")`,
			Output:     "a-b",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}([], "-")`,
			Output:     "",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}(foo, "-")`,
			Error:      ErrUndefined,
		},
		{
			Expression: `λ($obj)<o>{$obj}({"hello": "world"})`,
			Output: map[string]interface{}{
				"hello": "world",
			},
		},
		{
			Expression: `λ($arr)<a<a<n>>>{$arr}([[1]])`,
			Output: []interface{}{
				[]interface{}{
					float64(1),
				},
			},
		},
		{
			Expression: `λ($num)<(ns)-:n>{$number($num)}(5)`,
			Output:     float64(5),
		},
		{
			Expression: `λ($num)<(ns)-:n>{$number($num)}("5")`,
			Output:     float64(5),
		},
		{
			Expression: `[1..5].λ($num)<(ns)-:n>{$number($num)}()`,
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
			},
		},
		{
			Expression: `
				(
					$twice := function($f)<f:f>{function($x)<n:n>{$f($f($x))}};
					$add2 := function($x)<n:n>{$x+2};
					$add4 := $twice($add2);
					$add4(5)
				)`,
			Output: float64(9),
		},
		{
			Expression: `
				(
					$twice := function($f)<f<n:n>:f<n:n>>{function($x)<n:n>{$f($f($x))}};
					$add2 := function($x)<n:n>{$x+2};
					$add4 := $twice($add2);
					$add4(5)
				)`,
			Output: float64(9),
		},
		{
			Expression: `λ($arg)<n<n>>{$arg}(5)`,
			Error: &jparse.Error{
				// TODO: Get position info.
				Type: jparse.ErrInvalidSubtype,
				Hint: "n",
			},
		},
	})
}

func TestLambdaSignatures2(t *testing.T) {

	runTestCases(t, testdata.address, []*testCase{
		{
			Expression: `Age.function($x,$y)<n-n:n>{$x+$y}(6)`,
			Output:     float64(34),
		},
		{
			Expression: `FirstName.λ($str, $prefix)<s-s>{$prefix & $str}("Hello ")`,
			Output:     "Hello Fred",
		},
		{
			Expression: `λ($arr, $sep)<a<s>s?:s>{$join($arr, $sep)}(["a"])`,
			Output:     "a",
		},
	})
}

func TestLambdaSignatures3(t *testing.T) {

	runTestCases(t, testdata.account, []*testCase{
		{
			Expression: `Account.Order.Product.Description.Colour.λ($str)<s->{$uppercase($str)}()`,
			Output: []interface{}{
				"PURPLE",
				"ORANGE",
				"PURPLE",
				"BLACK",
			},
		},
	})
}

func TestLambdaSignatureViolations(t *testing.T) {

	runTestCases(t, nil, []*testCase{
		{
			Expression: `λ($arg1, $arg2)<nn:a>{[$arg1, $arg2]}(1,"2")`,
			Error: &ArgTypeError{
				Func:  "lambda",
				Which: 2,
			},
		},
		{
			Expression: `λ($arg1, $arg2)<nn:a>{[$arg1, $arg2]}(1,3,"2")`,
			Error: &ArgCountError{
				Func:     "lambda",
				Expected: 2,
				Received: 3,
			},
		},
		{
			Expression: `λ($arg1, $arg2)<nn+:a>{[$arg1, $arg2]}(1,3, 2,"g")`,
			Error: &ArgTypeError{
				Func:  "lambda",
				Which: 4,
			},
		},
		{
			Expression: `λ($arr)<a<n>>{$arr}(["3"]) `,
			Error: &ArgTypeError{
				Func:  "lambda",
				Which: 1,
			},
		},
		{
			Expression: `λ($arr)<a<n>>{$arr}([1, 2, "3"]) `,
			Error: &ArgTypeError{
				Func:  "lambda",
				Which: 1,
			},
		},
		{
			Expression: `λ($arr)<a<n>>{$arr}("f")`,
			Error: &ArgTypeError{
				Func:  "lambda",
				Which: 1,
			},
		},
		{
			Expression: `
				(
					$fun := λ($arr)<a<n>>{$arr};
					$fun("f")
				)`,
			Error: &ArgTypeError{
				Func:  "fun",
				Which: 1,
			},
		},
		{
			Expression: `λ($arr)<(sa<n>)>>{$arr}([[1]])`,
			Error: &jparse.Error{
				// TODO: Get position info.
				Type: jparse.ErrInvalidUnionType,
				Hint: "<",
			},
		},
	})
}

func TestTransform(t *testing.T) {

	data := map[string]interface{}{
		"state": map[string]interface{}{
			"tempReadings": []float64{
				27.2,
				28.9,
				28,
				28.2,
				28.4,
			},
			"readingsCount":   5,
			"sumTemperatures": 140.7,
			"avgTemperature":  28.14,
			"maxTemperature":  28.9,
			"minTemperature":  27.2,
		},
		"event": map[string]interface{}{
			"t": 28.4,
		},
	}

	runTestCases(t, data, []*testCase{
		{
			Expression: `
				(
					$tempReadings := $count(state.tempReadings) = 5 ?
						[state.tempReadings[[1..4]], event.t] :
						[state.tempReadings, event.t];

					{
						"tempReadings": $tempReadings,
						"sumTemperatures": $sum($tempReadings),
						"avgTemperature": $average($tempReadings) ~> $round(2),
						"maxTemperature": $max($tempReadings),
						"minTemperature": $min($tempReadings)
					}
				)`,
			Output: map[string]interface{}{
				"tempReadings": []interface{}{
					28.9,
					float64(28),
					28.2,
					28.4,
					28.4,
				},
				"sumTemperatures": 141.9,
				"avgTemperature":  28.38,
				"maxTemperature":  28.9,
				"minTemperature":  float64(28),
			},
		},
	})
}

// Helper functions

type compareFunc func(interface{}, interface{}) bool

func runTestCases(t *testing.T, input interface{}, tests []*testCase) {
	runTestCasesFunc(t, reflect.DeepEqual, input, tests)
}

func runTestCasesFunc(t *testing.T, compare compareFunc, input interface{}, tests []*testCase) {

	for _, test := range tests {
		if test.Skip {
			t.Logf("Skipping: %q", test.Expression)
			continue
		}
		runTestCase(t, compare, input, test)
	}
}

func runTestCase(t *testing.T, equal compareFunc, input interface{}, test *testCase) {

	var exps []string

	switch e := test.Expression.(type) {
	case string:
		exps = append(exps, e)
	case []string:
		exps = append(exps, e...)
	default:
		t.Fatalf("Bad expression: %T %v", e, e)
	}

	var output interface{}

	for _, exp := range exps {

		expr, err := Compile(exp)
		if err == nil {
			must(t, "Vars", expr.RegisterVars(test.Vars))
			must(t, "Exts", expr.RegisterExts(test.Exts))
			output, err = expr.Eval(input)
		}

		if !equal(output, test.Output) {
			t.Errorf("\nExpression: %s\nExp. Value: %v [%T]\nAct. Value: %v [%T]", exp, test.Output, test.Output, output, output)
		}
		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("\nExpression: %s\nExp. Error: %v [%T]\nAct. Error: %v [%T]", exp, test.Error, test.Error, err, err)
		}
	}
}

func equalRegexMatches(v1 interface{}, v2 interface{}) bool {

	makeMap := func(in interface{}) map[string]interface{} {
		m := in.(map[string]interface{})
		res := map[string]interface{}{}
		for k, v := range m {
			if k != "next" {
				res[k] = v
			}
		}
		return res
	}

	switch {
	case v1 == nil && v2 == nil:
		return true
	case v1 == nil, v2 == nil:
		return false
	default:
		return reflect.DeepEqual(makeMap(v1), makeMap(v2))
	}
}

func equalFloats(tolerance float64) func(interface{}, interface{}) bool {
	return func(v1, v2 interface{}) bool {

		n1, ok := v1.(float64)
		if !ok {
			return false
		}

		n2, ok := v2.(float64)
		if !ok {
			return false
		}

		return math.Abs(n1-n2) <= tolerance
	}
}

func equalArraysUnordered(a1, a2 interface{}) bool {

	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)

	if !v1.IsValid() || !v2.IsValid() {
		return false
	}

	if v1.Type() != v2.Type() {
		return false
	}

	if v1.Len() != v2.Len() {
		return false
	}

	matched := map[int]bool{}

	for i := 0; i < v1.Len(); i++ {

		found := false
		i1 := jtypes.Resolve(v1.Index(i)).Interface()

		for j := 0; j < v2.Len(); j++ {

			if matched[j] {
				continue
			}

			i2 := jtypes.Resolve(v2.Index(j)).Interface()

			if !reflect.DeepEqual(i1, i2) {
				continue
			}

			found = true
			matched[j] = true
			break
		}

		if !found {
			return false
		}
	}

	return true
}

func must(t *testing.T, prefix string, err error) {
	if err != nil {
		t.Fatalf("%s: %s", prefix, err)
	}
}

func readJSON(filename string) interface{} {

	data, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		panicf("ioutil.ReadFile error: %s", err)
	}

	var dest interface{}
	if err = json.Unmarshal(data, &dest); err != nil {
		panicf("json.Unmarshal error: %s", err)
	}

	return dest
}
