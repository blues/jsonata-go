// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xiatechs/jsonata-go/jlib"
	"github.com/xiatechs/jsonata-go/jparse"
	"github.com/xiatechs/jsonata-go/jtypes"
)

type evalTestCase struct {
	Input  jparse.Node
	Vars   map[string]interface{}
	Exts   map[string]Extension
	Data   interface{}
	Equals func(interface{}, interface{}) bool
	Output interface{}
	Error  error
}

func TestEvalString(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.StringNode{},
			Output: "",
		},
		{
			Input: &jparse.StringNode{
				Value: "hello world",
			},
			Output: "hello world",
		},
	})
}

func TestEvalNumber(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.NumberNode{},
			Output: float64(0),
		},
		{
			Input: &jparse.NumberNode{
				Value: 3.14159,
			},
			Output: 3.14159,
		},
	})
}

func TestEvalBoolean(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.BooleanNode{},
			Output: false,
		},
		{
			Input: &jparse.BooleanNode{
				Value: true,
			},
			Output: true,
		},
	})
}

func TestEvalNull(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.NullNode{},
			Output: null,
		},
	})
}

func TestEvalRegex(t *testing.T) {

	re := regexp.MustCompile("a(b+)")

	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.RegexNode{
				Value: re,
			},
			Output: &regexCallable{
				callableName: callableName{
					name: "a(b+)",
				},
				re: re,
			},
		},
	})
}

func TestEvalVariable(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Undefined variable. Return undefined.
			Input: &jparse.VariableNode{
				Name: "x",
			},
			Output: nil,
		},
		{
			// Defined variable. Return its value.
			Input: &jparse.VariableNode{
				Name: "x",
			},
			Vars: map[string]interface{}{
				"x": "marks the spot",
			},
			Output: "marks the spot",
		},
		{
			// The zero VariableNode ($) returns the evaluation
			// context.
			Input: &jparse.VariableNode{},
			Data: map[string]interface{}{
				"x": "marks the spot",
			},
			Output: map[string]interface{}{
				"x": "marks the spot",
			},
		},
	})
}

func TestEvalNegation(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Negate a number (in practice this happens during
			// parsing).
			Input: &jparse.NegationNode{
				RHS: &jparse.NumberNode{
					Value: 100,
				},
			},
			Output: float64(-100),
		},
		{
			// Negate a field.
			Input: &jparse.NegationNode{
				RHS: &jparse.NameNode{
					Value: "number",
				},
			},
			Data: map[string]interface{}{
				"number": -100,
			},
			Output: float64(100),
		},
		{
			// Negate a variable.
			Input: &jparse.NegationNode{
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Vars: map[string]interface{}{
				"x": 100,
			},
			Output: float64(-100),
		},
		{
			// Negating undefined should return undefined.
			Input: &jparse.NegationNode{
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: nil,
		},
		{
			// Negating a non-number should return an error.
			Input: &jparse.NegationNode{
				RHS: &jparse.BooleanNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// Negating an error should return the error.
			Input: &jparse.NegationNode{
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
	})
}

func TestEvalRange(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Standard case.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: -2,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
			Output: []interface{}{
				float64(-2),
				float64(-1),
				float64(0),
				float64(1),
				float64(2),
			},
		},
		{
			// Right side equals left side. Return a singleton
			// array.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 10,
				},
				RHS: &jparse.NumberNode{
					Value: 10,
				},
			},
			Output: []interface{}{
				float64(10),
			},
		},
		{
			// Left side returns an error. Return the error.
			Input: &jparse.RangeNode{
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// an error on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// a non-number on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.BooleanNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// a non-integer on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.NumberNode{
					Value: 1.5,
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// an undefined right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Left side is not a number. Return an error.
			Input: &jparse.RangeNode{
				LHS: &jparse.BooleanNode{},
				RHS: &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "false",
				Value: "..",
			},
		},
		{
			// A non-number on the left side takes precedence
			// over a non-number on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.BooleanNode{},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "false",
				Value: "..",
			},
		},
		{
			// A non-number on the left side takes precedence
			// over a non-integer on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.BooleanNode{},
				RHS: &jparse.NumberNode{
					Value: 1.5,
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "false",
				Value: "..",
			},
		},
		{
			// A non-number on the left side takes precedence
			// over an undefined right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.BooleanNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "false",
				Value: "..",
			},
		},
		{
			// Left side is not an integer. Return an error.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 1.5,
				},
				RHS: &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// A non-integer on the left side takes precedence
			// over a non-integer on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 1.5,
				},
				RHS: &jparse.NumberNode{
					Value: 0.5,
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// A non-integer on the left side takes precedence
			// over a non-number on the right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 1.5,
				},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// A non-integer on the left side takes precedence
			// over an undefined right side.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 1.5,
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerLHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// Right side returns an error. Return the error.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// An error on the right side takes precedence over
			// a non-number on the left side.
			Input: &jparse.RangeNode{
				LHS: &jparse.NullNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// An error on the right side takes precedence over
			// a non-integer on the left side.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 1.5,
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// An error on the right side takes precedence over
			// an undefined left side.
			Input: &jparse.RangeNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// Right side is not a number. Return an error.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "null",
				Value: "..",
			},
		},
		{
			// A non-number right side takes precedence over
			// an undefined left side.
			Input: &jparse.RangeNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "null",
				Value: "..",
			},
		},
		{
			// Right side is not an integer. Return an error.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.NumberNode{
					Value: 1.5,
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// A non-integer right side takes precedence over
			// an undefined left side.
			Input: &jparse.RangeNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NumberNode{
					Value: 1.5,
				},
			},
			Error: &EvalError{
				Type:  ErrNonIntegerRHS,
				Token: "1.5",
				Value: "..",
			},
		},
		{
			// If left side is undefined, return undefined.
			Input: &jparse.RangeNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NumberNode{},
			},
			Output: nil,
		},
		{
			// If right side is undefined, return undefined.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: nil,
		},
		{
			// If right side is less than left side, return undefined.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 2,
				},
				RHS: &jparse.NumberNode{
					Value: -2,
				},
			},
			Output: nil,
		},
		{
			// If there are too many items in the range, return
			// an error.
			Input: &jparse.RangeNode{
				LHS: &jparse.NumberNode{
					Value: 0,
				},
				RHS: &jparse.NumberNode{
					Value: maxRangeItems,
				},
			},
			Error: &EvalError{
				Type:  ErrMaxRangeItems,
				Token: "..",
			},
		},
	})
}

func TestEvalArray(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.ArrayNode{},
			Output: []interface{}{},
		},
		{
			Input: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.StringNode{
						Value: "two",
					},
					&jparse.BooleanNode{
						Value: true,
					},
					&jparse.NullNode{},
				},
			},
			Output: []interface{}{
				float64(1),
				"two",
				true,
				null,
			},
		},
		{
			// Nested ArrayNodes are added to the containing array
			// as arrays.
			Input: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.ArrayNode{
						Items: []jparse.Node{
							&jparse.StringNode{
								Value: "two",
							},
							&jparse.BooleanNode{
								Value: true,
							},
						},
					},
					&jparse.NullNode{},
				},
			},
			Output: []interface{}{
				float64(1),
				[]interface{}{
					"two",
					true,
				},
				null,
			},
		},
		{
			// Other array data gets flattened into the containing
			// array.
			Input: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.RangeNode{
						LHS: &jparse.NumberNode{
							Value: 2,
						},
						RHS: &jparse.NumberNode{
							Value: 4,
						},
					},
					&jparse.NumberNode{
						Value: 5,
					},
				},
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				float64(5),
			},
		},
		{
			// Undefined values are ignored.
			Input: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.VariableNode{
						Name: "two",
					},
					&jparse.NameNode{
						Value: "three",
					},
					&jparse.StringNode{
						Value: "four",
					},
				},
			},
			Output: []interface{}{
				float64(1),
				"four",
			},
		},
		{
			// If an item returns an error, return that error.
			Input: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.NegationNode{
						RHS: &jparse.BooleanNode{},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
	})
}

func TestEvalObject(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.ObjectNode{},
			Output: map[string]interface{}{},
		},
		{
			Input: &jparse.ObjectNode{
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
			Output: map[string]interface{}{
				"one":   float64(1),
				"two":   float64(2),
				"three": float64(3),
			},
		},
		{
			// If a key evaluates to an error, return that error.
			Input: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.NegationNode{
							RHS: &jparse.StringNode{},
						},
						&jparse.NumberNode{},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: `""`,
				Value: "-",
			},
		},
		{
			// If a key is not a string, return an error.
			Input: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.NullNode{},
						&jparse.NumberNode{},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrIllegalKey,
				Token: "null",
			},
		},
		{
			// If a key is duplicated, return an error.
			Input: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.StringNode{
							Value: "key",
						},
						&jparse.NumberNode{},
					},
					{
						&jparse.StringNode{
							Value: "key",
						},
						&jparse.BooleanNode{},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrDuplicateKey,
				Token: `"key"`,
				Value: "key",
			},
		},
		{
			// Values that evaluate to undefined are ignored.
			Input: &jparse.ObjectNode{
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
						&jparse.VariableNode{
							Name: "x",
						},
					},
					{
						&jparse.StringNode{
							Value: "three",
						},
						&jparse.NameNode{
							Value: "Field",
						},
					},
				},
			},
			Output: map[string]interface{}{
				"one": float64(1),
			},
		},
		{
			// If a value evaluates to an error, return that error.
			Input: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.StringNode{},
						&jparse.NegationNode{
							RHS: &jparse.StringNode{},
						},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: `""`,
				Value: "-",
			},
		},
	})
}

func TestEvalGroup(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.GroupNode{
				Expr: &jparse.VariableNode{},
				ObjectNode: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.NameNode{
								Value: "name",
							},
							&jparse.FunctionCallNode{
								Func: &jparse.VariableNode{
									Name: "sum",
								},
								Args: []jparse.Node{
									&jparse.NameNode{
										Value: "value",
									},
								},
							},
						},
					},
				},
			},
			Data: []interface{}{
				map[string]interface{}{
					"name": "zero",
				},
				map[string]interface{}{
					"name":  "one",
					"value": 1,
				},
				map[string]interface{}{
					"name":  "two",
					"value": 2,
				},
				map[string]interface{}{
					"name":  "two",
					"value": 12,
				},
				map[string]interface{}{
					"name":  "three",
					"value": 3,
				},
				map[string]interface{}{
					"name":  "three",
					"value": 13,
				},
				map[string]interface{}{
					"name":  "three",
					"value": 23,
				},
			},
			Exts: map[string]Extension{
				"sum": {
					Func: jlib.Sum,
					UndefinedHandler: func(argv []reflect.Value) bool {
						return len(argv) > 0 && argv[0] == undefined
					},
				},
			},
			Output: map[string]interface{}{
				"one":   float64(1),
				"two":   float64(14),
				"three": float64(39),
			},
		},
		{
			// Expression being grouped returns an error. Return the error.
			Input: &jparse.GroupNode{
				Expr: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
	})
}

func TestEvalAssignment(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Assignment returns the value being assigned.
			Input: &jparse.AssignmentNode{
				Name: "x",
				Value: &jparse.StringNode{
					Value: "marks the spot",
				},
			},
			Output: "marks the spot",
		},
		{
			// Assignment modifies the environment.
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.AssignmentNode{
						Name: "x",
						Value: &jparse.StringNode{
							Value: "marks the spot",
						},
					},
					&jparse.VariableNode{
						Name: "x",
					},
				},
			},
			Output: "marks the spot",
		},
		{
			// If the value evaluates to an error, return
			// that error.
			Input: &jparse.AssignmentNode{
				Name: "x",
				Value: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
	})
}

func TestEvalBlock(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input:  &jparse.BlockNode{},
			Output: nil,
		},
		{
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.AssignmentNode{
						Name: "pi",
						Value: &jparse.NumberNode{
							Value: 3.14159,
						},
					},
					&jparse.VariableNode{
						Name: "pi",
					},
				},
			},
			Output: 3.14159,
		},
		{
			// Variables defined in a block should be scoped
			// to that block.
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.BlockNode{
						Exprs: []jparse.Node{
							&jparse.AssignmentNode{
								Name: "pi",
								Value: &jparse.NumberNode{
									Value: 100,
								},
							},
						},
					},
					&jparse.VariableNode{
						Name: "pi",
					},
				},
			},
			Output: nil,
		},
		{
			// If an expression evaluates to an error, return
			// that error.
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.FunctionCallNode{
						Func: &jparse.BooleanNode{},
					},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
	})
}

func TestEvalConditional(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Condition is true, no else. Return true value.
			Input: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: true,
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
			},
			Output: "true",
		},
		{
			// Condition is false, no else. Return undefined.
			Input: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: false,
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
			},
			Output: nil,
		},
		{
			// Condition is undefined, no else. Return undefined.
			Input: &jparse.ConditionalNode{
				If: &jparse.VariableNode{
					Name: "x",
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
			},
			Output: nil,
		},
		{
			// Condition is true, else. Return true value.
			Input: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: true,
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
				Else: &jparse.StringNode{
					Value: "false",
				},
			},
			Output: "true",
		},
		{
			// Condition is false, else. Return false value.
			Input: &jparse.ConditionalNode{
				If: &jparse.BooleanNode{
					Value: false,
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
				Else: &jparse.StringNode{
					Value: "false",
				},
			},
			Output: "false",
		},
		{
			// Condition is undefined, else. Return false value.
			Input: &jparse.ConditionalNode{
				If: &jparse.VariableNode{
					Name: "x",
				},
				Then: &jparse.StringNode{
					Value: "true",
				},
				Else: &jparse.StringNode{
					Value: "false",
				},
			},
			Output: "false",
		},
		{
			// Condition evaluates to an error. Return the error.
			Input: &jparse.ConditionalNode{
				If: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{
						Value: false,
					},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
	})
}

func TestEvalWildcard(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// If the input is an empty array, the wildcard
			// operator evaluates to undefined.
			Input:  &jparse.WildcardNode{},
			Data:   []interface{}{},
			Output: nil,
		},
		{
			// If the input is an array with one item, the
			// wildcard operator evaluates to the item.
			Input: &jparse.WildcardNode{},
			Data: []interface{}{
				"one",
			},
			Output: "one",
		},
		{
			// If the input is a multi-item array, the wildcard
			// operator evaluates to the items.
			Input: &jparse.WildcardNode{},
			Data: []interface{}{
				"one",
				"two",
				"three",
			},
			Output: []interface{}{
				"one",
				"two",
				"three",
			},
		},
		{
			// Nested arrays are flattened into a single array.
			Input: &jparse.WildcardNode{},
			Data: []interface{}{
				"one",
				[]interface{}{
					"two",
					[]interface{}{
						"three",
						[]interface{}{
							"four",
							[]interface{}{
								"five",
							},
						},
					},
				},
			},
			Output: []interface{}{
				"one",
				"two",
				"three",
				"four",
				"five",
			},
		},
		{
			// If the input is an empty map, the wildcard
			// operator evaluates to undefined.
			Input:  &jparse.WildcardNode{},
			Data:   map[string]interface{}{},
			Output: nil,
		},
		{
			// If the input is a map with one item, the wildcard
			// operator evaluates to the item's value.
			Input: &jparse.WildcardNode{},
			Data: map[string]interface{}{
				"one": 1,
			},
			Output: 1,
		},
		{
			// If the input is a map with multiple items, the
			// wildcard operator returns the map values in an
			// array. Note that the order is non-deterministic
			// because Go traverses maps in random order.
			Input: &jparse.WildcardNode{},
			Data: map[string]interface{}{
				"one":   1,
				"two":   2,
				"three": 3,
				"fourfive": []int{
					4,
					5,
				},
				"six": 6,
			},
			Equals: equalArraysUnordered,
			Output: []interface{}{
				1,
				2,
				3,
				4,
				5,
				6,
			},
		},
		{
			// If the input is an empty struct, the wildcard
			// operator evaluates to undefined.
			Input:  &jparse.WildcardNode{},
			Data:   struct{}{},
			Output: nil,
		},
		{
			// If the input is a struct with one field, the
			// wildcard operator evaluates to the field's value.
			Input: &jparse.WildcardNode{},
			Data: struct {
				Value string
			}{
				Value: "hello world",
			},
			Output: "hello world",
		},
		{
			// Unexported struct fields are ignored, The reflect
			// package does not allow them to be copied into an
			// array (as the wildcard operator reuires).
			Input: &jparse.WildcardNode{},
			Data: struct {
				value string
			}{
				value: "hello world",
			},
			Output: nil,
		},
		{
			// If the input is a struct with multiple fields, the
			// wildcard operator returns the field values in an
			// array. Unlike with a map, the order of the fields
			// is predictable.
			Input: &jparse.WildcardNode{},
			Data: struct {
				One      int
				Two      int
				Three    int
				FourFive []int
				six      int
			}{
				One:   1,
				Two:   2,
				Three: 3,
				FourFive: []int{
					4,
					5,
				},
				six: 6,
			},
			Output: []interface{}{
				1,
				2,
				3,
				4,
				5,
			},
		},
	})
}

func TestEvalDescendent(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// If the input is an empty array, the descendent
			// operator evaluates to undefined.
			Input:  &jparse.DescendentNode{},
			Data:   []interface{}{},
			Output: nil,
		},
		{
			// If the input is an array with one item, the
			// descendent operator evaluates to the item.
			Input: &jparse.DescendentNode{},
			Data: []interface{}{
				"one",
			},
			Output: "one",
		},
		{
			// If the input is a multi-item array, the descendent
			// operator evaluates to the items.
			Input: &jparse.DescendentNode{},
			Data: []interface{}{
				"one",
				"two",
				"three",
			},
			Output: []interface{}{
				"one",
				"two",
				"three",
			},
		},
		{
			// Nested arrays are flattened into a single array.
			Input: &jparse.DescendentNode{},
			Data: []interface{}{
				"one",
				[]interface{}{
					"two",
					[]interface{}{
						"three",
						[]interface{}{
							"four",
							[]interface{}{
								"five",
							},
						},
					},
				},
			},
			Output: []interface{}{
				"one",
				"two",
				"three",
				"four",
				"five",
			},
		},
		{
			// If the input is an empty map, the descendent
			// operator evaluates to an empty map.
			Input:  &jparse.DescendentNode{},
			Data:   map[string]interface{}{},
			Output: map[string]interface{}{},
		},
		{
			// If the input is a map with one item, the descendent
			// operator evaluates to an array containing the map
			// itself and its single value.
			Input: &jparse.DescendentNode{},
			Data: map[string]interface{}{
				"one": 1,
			},
			Output: []interface{}{
				map[string]interface{}{
					"one": 1,
				},
				1,
			},
		},
		{
			// If the input is a map with multiple items, the
			// descendent operator returns an array containing
			// the map itself and its values. Note that the order
			// of the values is non-deterministic because Go
			// traverses maps in random order.
			Input: &jparse.DescendentNode{},
			Data: map[string]interface{}{
				"one":   1,
				"two":   2,
				"three": 3,
				"fourfive": []int{
					4,
					5,
				},
				"six": 6,
			},
			Equals: equalArraysUnordered,
			Output: []interface{}{
				map[string]interface{}{
					"one":   1,
					"two":   2,
					"three": 3,
					"fourfive": []int{
						4,
						5,
					},
					"six": 6,
				},
				1,
				2,
				3,
				4,
				5,
				6,
			},
		},
		{
			// If the input is an empty struct, the descendent
			// operator evaluates to an empty struct.
			Input:  &jparse.DescendentNode{},
			Data:   struct{}{},
			Output: struct{}{},
		},
		{
			// If the input is a struct with one field, the
			// descendent operator evaluates to an array containing
			// the struct itself and the single field's value.
			Input: &jparse.DescendentNode{},
			Data: struct {
				Value string
			}{
				Value: "hello world",
			},
			Output: []interface{}{
				struct {
					Value string
				}{
					Value: "hello world",
				},
				"hello world",
			},
		},
		{
			// Unexported struct fields appear as part of the struct
			// but are not added to the result array as individual
			// fields.
			Input: &jparse.DescendentNode{},
			Data: struct {
				value string
			}{
				value: "hello world",
			},
			Output: struct {
				value string
			}{
				value: "hello world",
			},
		},
		{
			// If the input is a struct with multiple fields, the
			// descendent operator returns an array containing the
			// struct itself plus the field values. Unlike with a
			// map, the order of the individual values is predictable.
			Input: &jparse.DescendentNode{},
			Data: struct {
				One      int
				Two      int
				Three    int
				FourFive []int
				six      int
			}{
				One:   1,
				Two:   2,
				Three: 3,
				FourFive: []int{
					4,
					5,
				},
				six: 6,
			},
			Output: []interface{}{
				struct {
					One      int
					Two      int
					Three    int
					FourFive []int
					six      int
				}{
					One:   1,
					Two:   2,
					Three: 3,
					FourFive: []int{
						4,
						5,
					},
					six: 6,
				},
				1,
				2,
				3,
				4,
				5,
			},
		},
	})
}

func TestEvalPredicate(t *testing.T) {

	makeRange := func(from, to float64) *jparse.ArrayNode {
		return &jparse.ArrayNode{
			Items: []jparse.Node{
				&jparse.RangeNode{
					LHS: &jparse.NumberNode{
						Value: from,
					},
					RHS: &jparse.NumberNode{
						Value: to,
					},
				},
			},
		}
	}

	testEvalTestCases(t, []evalTestCase{
		{
			// Positive indexes are zero-based.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: 3,
					},
				},
			},
			Output: float64(4),
		},
		{
			// Negative indexes count from the end.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: -3,
					},
				},
			},
			Output: float64(8),
		},
		{
			// Non-integer indexes round down.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: 0.9,
					},
				},
			},
			Output: float64(1),
		},
		{
			// Non-integer indexes round down (i.e. away from
			// zero for negative values).
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: -0.9,
					},
				},
			},
			Output: float64(10),
		},
		{
			// Out of bounds indexes return undefined.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.NumberNode{
						Value: 20,
					},
				},
			},
			Output: nil,
		},
		{
			// Multiple indexes return an array of values.
			// Note that:
			//
			// 1. Values are returned in the same order as
			//    the source array.
			//
			// 2. If an index appears multiple times in the
			//    filter expression, the corresponding value
			//    will appear mutliple times in the results.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.ArrayNode{
						Items: []jparse.Node{
							&jparse.NumberNode{
								Value: -1,
							},
							&jparse.NumberNode{
								Value: 4,
							},
							&jparse.NumberNode{
								Value: 4,
							},
							&jparse.NumberNode{
								Value: 10,
							},
							&jparse.NumberNode{
								Value: 0,
							},
						},
					},
				},
			},
			Output: []interface{}{
				float64(1),
				float64(5),
				float64(5),
				float64(10),
			},
		},
		{
			// Filters that are not numeric indexes are tested
			// for truthiness.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					&jparse.ComparisonOperatorNode{
						Type: jparse.ComparisonEqual,
						LHS: &jparse.NumericOperatorNode{
							Type: jparse.NumericModulo,
							LHS:  &jparse.VariableNode{},
							RHS: &jparse.NumberNode{
								Value: 3,
							},
						},
						RHS: &jparse.NumberNode{
							Value: 0,
						},
					},
				},
			},
			Output: []interface{}{
				float64(3),
				float64(6),
				float64(9),
			},
		},
		{
			// Multiple filters are evaluated in succession.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					makeRange(3, 6),
					&jparse.NumberNode{
						Value: 1,
					},
				},
			},
			Output: float64(5),
		},
		{
			// If the predicate expression evaluates to undefined,
			// the predicate should evaluate to undefined.
			Input: &jparse.PredicateNode{
				Expr: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: nil,
		},
		{
			// If the predicate expression evaluates to an error,
			// the predicate should return the error.
			Input: &jparse.PredicateNode{
				Expr: &jparse.NegationNode{
					RHS: makeRange(1, 10),
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "[1..10]",
				Value: "-",
			},
		},
		{
			// If a filter evaluates to an error, the predicate
			// should return the error.
			Input: &jparse.PredicateNode{
				Expr: makeRange(1, 10),
				Filters: []jparse.Node{
					makeRange(1, 1e10),
				},
			},
			Error: &EvalError{
				Type:  ErrMaxRangeItems,
				Token: "..",
			},
		},
	})
}

func TestEvalSort(t *testing.T) {

	data := []interface{}{
		map[string]interface{}{
			"value": 0,
		},
		map[string]interface{}{
			"value": 1,
			"en":    "one",
		},
		map[string]interface{}{
			"value": 2,
			"en":    "two",
		},
		map[string]interface{}{
			"value": 3,
			"en":    "three",
		},
		map[string]interface{}{
			"value": 4,
			"en":    "four",
		},
		map[string]interface{}{
			"value": 5,
			"en":    "five",
		},
		map[string]interface{}{
			"value": 1,
			"es":    "uno",
		},
		map[string]interface{}{
			"value": 2,
			"es":    "dos",
		},
		map[string]interface{}{
			"value": 3,
			"es":    "tres",
		},
		map[string]interface{}{
			"value": 4,
			"es":    "cuatro",
		},
		map[string]interface{}{
			"value": 5,
			"es":    "cinco",
		},
	}

	testEvalTestCases(t, []evalTestCase{
		{
			// Sorting by "en" should sort the items by their
			// English name. Items without an English name should
			// appear at the end of the list in the same order as
			// the original array.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortAscending,
						Expr: &jparse.NameNode{
							Value: "en",
						},
					},
				},
			},
			Data: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": 5,
					"en":    "five",
				},
				map[string]interface{}{
					"value": 4,
					"en":    "four",
				},
				map[string]interface{}{
					"value": 1,
					"en":    "one",
				},
				map[string]interface{}{
					"value": 3,
					"en":    "three",
				},
				map[string]interface{}{
					"value": 2,
					"en":    "two",
				},
				map[string]interface{}{
					"value": 0,
				},
				map[string]interface{}{
					"value": 1,
					"es":    "uno",
				},
				map[string]interface{}{
					"value": 2,
					"es":    "dos",
				},
				map[string]interface{}{
					"value": 3,
					"es":    "tres",
				},
				map[string]interface{}{
					"value": 4,
					"es":    "cuatro",
				},
				map[string]interface{}{
					"value": 5,
					"es":    "cinco",
				},
			},
		},
		{
			// Sorting by "es" should sort the items by their
			// Spanish name. Items without a Spanish name should
			// appear at the end of the list in the same order as
			// the original array.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDescending,
						Expr: &jparse.NameNode{
							Value: "es",
						},
					},
				},
			},
			Data: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": 1,
					"es":    "uno",
				},
				map[string]interface{}{
					"value": 3,
					"es":    "tres",
				},
				map[string]interface{}{
					"value": 2,
					"es":    "dos",
				},
				map[string]interface{}{
					"value": 4,
					"es":    "cuatro",
				},
				map[string]interface{}{
					"value": 5,
					"es":    "cinco",
				},
				map[string]interface{}{
					"value": 0,
				},
				map[string]interface{}{
					"value": 1,
					"en":    "one",
				},
				map[string]interface{}{
					"value": 2,
					"en":    "two",
				},
				map[string]interface{}{
					"value": 3,
					"en":    "three",
				},
				map[string]interface{}{
					"value": 4,
					"en":    "four",
				},
				map[string]interface{}{
					"value": 5,
					"en":    "five",
				},
			},
		},
		{
			// Sorting by "value" should sort the items by their
			// numeric value. Items with the same numeric value
			// should appear in the same order as the original
			// array (i.e. items with English names should come
			// before items with Spanish names).
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDefault,
						Expr: &jparse.NameNode{
							Value: "value",
						},
					},
				},
			},
			Data: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": 0,
				},
				map[string]interface{}{
					"value": 1,
					"en":    "one",
				},
				map[string]interface{}{
					"value": 1,
					"es":    "uno",
				},
				map[string]interface{}{
					"value": 2,
					"en":    "two",
				},
				map[string]interface{}{
					"value": 2,
					"es":    "dos",
				},
				map[string]interface{}{
					"value": 3,
					"en":    "three",
				},
				map[string]interface{}{
					"value": 3,
					"es":    "tres",
				},
				map[string]interface{}{
					"value": 4,
					"en":    "four",
				},
				map[string]interface{}{
					"value": 4,
					"es":    "cuatro",
				},
				map[string]interface{}{
					"value": 5,
					"en":    "five",
				},
				map[string]interface{}{
					"value": 5,
					"es":    "cinco",
				},
			},
		},
		{
			// Sorting by "value" and "es" should first sort the
			// items by their numeric value. Then items with the
			// same numeric value should be sorted according to
			// their Spanish name. Items without a Spanish name
			// should appear at the end of each internal sort.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDescending,
						Expr: &jparse.NameNode{
							Value: "value",
						},
					},
					{
						Dir: jparse.SortDescending,
						Expr: &jparse.NameNode{
							Value: "es",
						},
					},
				},
			},
			Data: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": 5,
					"es":    "cinco",
				},
				map[string]interface{}{
					"value": 5,
					"en":    "five",
				},
				map[string]interface{}{
					"value": 4,
					"es":    "cuatro",
				},
				map[string]interface{}{
					"value": 4,
					"en":    "four",
				},
				map[string]interface{}{
					"value": 3,
					"es":    "tres",
				},
				map[string]interface{}{
					"value": 3,
					"en":    "three",
				},
				map[string]interface{}{
					"value": 2,
					"es":    "dos",
				},
				map[string]interface{}{
					"value": 2,
					"en":    "two",
				},
				map[string]interface{}{
					"value": 1,
					"es":    "uno",
				},
				map[string]interface{}{
					"value": 1,
					"en":    "one",
				},
				map[string]interface{}{
					"value": 0,
				},
			},
		},
		{
			// Sorting a single item should return that item.
			// TODO: Find out why the latest version of jsonata-js
			// returns an array for this case. If it's because of
			// the recent changes to the sort function, we don't
			// need to replicate the behaviour.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDefault,
						Expr: &jparse.NameNode{
							Value: "value",
						},
					},
				},
			},
			Data: []interface{}{
				map[string]interface{}{
					"value": 0,
				},
			},
			Output: map[string]interface{}{
				"value": 0,
			},
		},
		{
			// Sorting an empty array should return an empty
			// array.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Dir: jparse.SortDefault,
						Expr: &jparse.NameNode{
							Value: "value",
						},
					},
				},
			},
			Data:   []interface{}{},
			Output: []interface{}{},
		},
		{
			// If the expression being sorted evaluates to an
			// error, return the error.
			Input: &jparse.SortNode{
				Expr: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// If the expression being sorted evaluates to
			// undefined, return undefined.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
			},
			Output: nil,
		},
		{
			// If a sort term evaluates to an error, return
			// the error.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Expr: &jparse.NegationNode{
							RHS: &jparse.VariableNode{},
						},
					},
				},
			},
			Data: []interface{}{
				"hello",
				"world",
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "$",
				Value: "-",
			},
		},
		{
			// If a sort term evaluates to anything other than
			// a string or a number, return an error.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Expr: &jparse.VariableNode{},
					},
				},
			},
			Data: []interface{}{
				"hello",
				"world",
				false,
			},
			Error: &EvalError{
				Type:  ErrNonSortable,
				Token: "$",
			},
		},
		{
			// If the sort terms evaluate to different types for
			// different items in the source array, return an error.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Expr: &jparse.VariableNode{},
					},
				},
			},
			Data: []interface{}{
				"hello",
				"world",
				100,
			},
			Error: &EvalError{
				Type:  ErrSortMismatch,
				Token: "$",
			},
		},
		{
			// If the sort terms evaluates to different types for
			// different items in the source array, return an error.
			Input: &jparse.SortNode{
				Expr: &jparse.VariableNode{},
				Terms: []jparse.SortTerm{
					{
						Expr: &jparse.VariableNode{},
					},
				},
			},
			Data: []interface{}{
				100,
				"string",
			},
			Error: &EvalError{
				Type:  ErrSortMismatch,
				Token: "$",
			},
		},
	})
}

func TestEvalLambda(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.LambdaNode{
				Body: &jparse.BooleanNode{
					Value: true,
				},
				ParamNames: []string{
					"x",
					"y",
				},
			},
			Output: &lambdaCallable{
				callableName: callableName{
					name: "lambda",
				},
				paramNames: []string{
					"x",
					"y",
				},
				body: &jparse.BooleanNode{
					Value: true,
				},
				env: newEnvironment(nil, 0),
			},
		},
	})
}

func TestEvalTypedLambda(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.TypedLambdaNode{
				LambdaNode: &jparse.LambdaNode{
					Body: &jparse.BooleanNode{
						Value: true,
					},
					ParamNames: []string{
						"x",
						"y",
					},
				},
				In: []jparse.Param{
					{
						Type: jparse.ParamTypeString,
					},
					{
						Type:   jparse.ParamTypeNumber,
						Option: jparse.ParamOptional,
					},
				},
			},
			Output: &lambdaCallable{
				callableName: callableName{
					name: "lambda",
				},
				paramNames: []string{
					"x",
					"y",
				},
				body: &jparse.BooleanNode{
					Value: true,
				},
				typed: true,
				params: []jparse.Param{
					{
						Type: jparse.ParamTypeString,
					},
					{
						Type:   jparse.ParamTypeNumber,
						Option: jparse.ParamOptional,
					},
				},
				env: newEnvironment(nil, 0),
			},
		},
	})
}

func TestEvalObjectTransformation(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Updates, no deletions.
			Input: &jparse.ObjectTransformationNode{
				Pattern: &jparse.VariableNode{},
				Updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "key",
							},
							&jparse.NameNode{
								Value: "value",
							},
						},
					},
				},
			},
			Output: &transformationCallable{
				callableName: callableName{
					name: "transform",
				},
				pattern: &jparse.VariableNode{},
				updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "key",
							},
							&jparse.NameNode{
								Value: "value",
							},
						},
					},
				},
				env: newEnvironment(nil, 0),
			},
		},
		{
			// Updates and deletions.
			Input: &jparse.ObjectTransformationNode{
				Pattern: &jparse.VariableNode{},
				Updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "key",
							},
							&jparse.NameNode{
								Value: "value",
							},
						},
					},
				},
				Deletes: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "field1",
						},
						&jparse.StringNode{
							Value: "field2",
						},
					},
				},
			},
			Output: &transformationCallable{
				callableName: callableName{
					name: "transform",
				},
				pattern: &jparse.VariableNode{},
				updates: &jparse.ObjectNode{
					Pairs: [][2]jparse.Node{
						{
							&jparse.StringNode{
								Value: "key",
							},
							&jparse.NameNode{
								Value: "value",
							},
						},
					},
				},
				deletes: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "field1",
						},
						&jparse.StringNode{
							Value: "field2",
						},
					},
				},
				env: newEnvironment(nil, 0),
			},
		},
	})
}

func TestEvalPartial(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.PartialNode{
				Func: &jparse.LambdaNode{
					Body: &jparse.NullNode{},
					ParamNames: []string{
						"x",
						"y",
					},
				},
				Args: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.PlaceholderNode{},
				},
			},
			Output: &partialCallable{
				callableName: callableName{
					name: "lambda_partial",
				},
				fn: &lambdaCallable{
					callableName: callableName{
						name: "lambda",
					},
					body: &jparse.NullNode{},
					paramNames: []string{
						"x",
						"y",
					},
					env: newEnvironment(nil, 0),
				},
				args: []jparse.Node{
					&jparse.NumberNode{
						Value: 1,
					},
					&jparse.PlaceholderNode{},
				},
				env: newEnvironment(nil, 0),
			},
		},
		{
			// Error evaluating the embedded function. Return the error.
			Input: &jparse.PartialNode{
				Func: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				Args: []jparse.Node{
					&jparse.PlaceholderNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// Embedded function is not a Callable. Return an error.
			Input: &jparse.PartialNode{
				Func: &jparse.BooleanNode{},
				Args: []jparse.Node{
					&jparse.PlaceholderNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallablePartial,
				Token: "false",
			},
		},
	})
}

func TestEvalFunctionCall(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Call lambda.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.LambdaNode{
					Body: &jparse.NumericOperatorNode{
						Type: jparse.NumericMultiply,
						LHS: &jparse.VariableNode{
							Name: "x",
						},
						RHS: &jparse.VariableNode{
							Name: "y",
						},
					},
					ParamNames: []string{
						"x",
						"y",
					},
				},
				Args: []jparse.Node{
					&jparse.NumberNode{
						Value: 3,
					},
					&jparse.NumberNode{
						Value: 18,
					},
				},
			},
			Output: float64(54),
		},
		{
			// Call Extension.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "repeat",
				},
				Args: []jparse.Node{
					&jparse.StringNode{
						Value: "x",
					},
					&jparse.NumberNode{
						Value: 10,
					},
				},
			},
			Exts: map[string]Extension{
				"repeat": {
					Func: strings.Repeat,
				},
			},
			Output: "xxxxxxxxxx",
		},
		{
			// Call partial.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.PartialNode{
					Func: &jparse.VariableNode{
						Name: "repeat",
					},
					Args: []jparse.Node{
						&jparse.PlaceholderNode{},
						&jparse.NumberNode{
							Value: 5,
						},
					},
				},
				Args: []jparse.Node{
					&jparse.StringNode{
						Value: "😅",
					},
				},
			},
			Exts: map[string]Extension{
				"repeat": {
					Func: strings.Repeat,
				},
			},
			Output: "😅😅😅😅😅",
		},
		{
			// Error evaluating the callee. Return the error.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// Callee is not callable. Return an error.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Argument evaluates to an error. Return the error.
			Input: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "repeat",
				},
				Args: []jparse.Node{
					&jparse.StringNode{
						Value: "x",
					},
					&jparse.NegationNode{
						RHS: &jparse.BooleanNode{},
					},
				},
			},
			Exts: map[string]Extension{
				"repeat": {
					Func: strings.Repeat,
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
	})
}

func TestEvalFunctionApplication(t *testing.T) {

	trimSpace, _ := newGoCallable("trim", Extension{
		Func: strings.TrimSpace,
	})

	toUpper, _ := newGoCallable("uppercase", Extension{
		Func: strings.ToUpper,
	})

	testEvalTestCases(t, []evalTestCase{
		{
			// Pass an argument to a function call.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.StringNode{
					Value: "😂",
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.VariableNode{
						Name: "repeat",
					},
					Args: []jparse.Node{
						&jparse.NumberNode{
							Value: 5,
						},
					},
				},
			},
			Exts: map[string]Extension{
				"repeat": {
					Func: strings.Repeat,
				},
			},
			Output: "😂😂😂😂😂",
		},
		{
			// Pass an argument to a function.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.VariableNode{
					Name: "uppercase",
				},
			},
			Exts: map[string]Extension{
				"uppercase": {
					Func: strings.ToUpper,
				},
			},
			Output: "HELLO",
		},
		{
			// Chain two functions together.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.VariableNode{
					Name: "trim",
				},
				RHS: &jparse.VariableNode{
					Name: "uppercase",
				},
			},
			Exts: map[string]Extension{
				"trim": {
					Func: strings.TrimSpace,
				},
				"uppercase": {
					Func: strings.ToUpper,
				},
			},
			Output: &chainCallable{
				callables: []jtypes.Callable{
					trimSpace,
					toUpper,
				},
			},
		},
		{
			// Left side returns an error. Return the error.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.VariableNode{
					Name: "uppercase",
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// an error on the right side.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// a non-callable right side.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.BooleanNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// Right side returns an error. Return the error.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// Right side is not a Callable. Return an error.
			Input: &jparse.FunctionApplicationNode{
				LHS: &jparse.NumberNode{},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallableApply,
				Token: "null",
				Value: "~>",
			},
		},
	})
}

func TestEvalNumericOperator(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Addition.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.NumberNode{
					Value: 100,
				},
				RHS: &jparse.NumberNode{
					Value: 3.14159,
				},
			},
			Output: 103.14159,
		},
		{
			// Subtraction.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericSubtract,
				LHS: &jparse.NumberNode{
					Value: 100,
				},
				RHS: &jparse.NumberNode{
					Value: 17.5,
				},
			},
			Output: 82.5,
		},
		{
			// Multiplication.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericMultiply,
				LHS: &jparse.NumberNode{
					Value: 10,
				},
				RHS: &jparse.NumberNode{
					Value: 1.25e5,
				},
			},
			Output: float64(1250000),
		},
		{
			// Division.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericDivide,
				LHS: &jparse.NumberNode{
					Value: -99,
				},
				RHS: &jparse.NumberNode{
					Value: 3,
				},
			},
			Output: float64(-33),
		},
		{
			// Modulo.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericModulo,
				LHS: &jparse.NumberNode{
					Value: -99,
				},
				RHS: &jparse.NumberNode{
					Value: 6,
				},
			},
			Output: float64(-3),
		},
		{
			// Expression evaluates to infinity. Return an error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericDivide,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 0,
				},
			},
			Error: &EvalError{
				Type:  ErrNumberInf,
				Value: "/",
			},
		},
		{
			// Expression evaluates to NaN. Return an error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericDivide,
				LHS: &jparse.NumberNode{
					Value: 0,
				},
				RHS: &jparse.NumberNode{
					Value: 0,
				},
			},
			Error: &EvalError{
				Type:  ErrNumberNaN,
				Value: "/",
			},
		},
		{
			// Left side returns an error. Return the error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// an error on the right side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// a non-number on the right side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.BooleanNode{},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the left side takes precedence over
			// an undefined right side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Left side is not a number. Return an error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.BooleanNode{},
				RHS:  &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberLHS,
				Token: "false",
				Value: "+",
			},
		},
		{
			// A non-number on the left side takes precedence
			// over a non-number on the right side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.BooleanNode{},
				RHS:  &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberLHS,
				Token: "false",
				Value: "+",
			},
		},
		{
			// A non-number on the left side takes precedence
			// over an undefined right side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.BooleanNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberLHS,
				Token: "false",
				Value: "+",
			},
		},
		{
			// Right side returns an error. Return the error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.NumberNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// An error on the right side takes precedence over
			// a non-number on the left side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.NullNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// An error on the right side takes precedence over
			// an undefined left side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "false",
			},
		},
		{
			// Right side is not a number. Return an error.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.NumberNode{},
				RHS:  &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "+",
			},
		},
		{
			// A non-number right side rakes precedence over
			// an undefined left side.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "+",
			},
		},
		{
			// Left side is undefined. Return undefined.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NumberNode{},
			},
			Output: nil,
		},
		{
			// Right side is undefined. Return undefined.
			Input: &jparse.NumericOperatorNode{
				Type: jparse.NumericAdd,
				LHS:  &jparse.NumberNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: nil,
		},
	})
}

func TestEvalComparisonOperator(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// Number = Number: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.NumberNode{
					Value: 100,
				},
				RHS: &jparse.NumberNode{
					Value: 1e2,
				},
			},
			Output: true,
		},
		{
			// Number = Number: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.NumberNode{
					Value: 100,
				},
				RHS: &jparse.NumberNode{
					Value: -100,
				},
			},
			Output: false,
		},
		{
			// Number = another type: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.NumberNode{
					Value: 100,
				},
				RHS: &jparse.StringNode{
					Value: "100",
				},
			},
			Output: false,
		},
		{
			// String = String: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "hello",
				},
			},
			Output: true,
		},
		{
			// String = String: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
			Output: false,
		},
		{
			// String = another type: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.StringNode{
					Value: "null",
				},
				RHS: &jparse.NullNode{},
			},
			Output: false,
		},
		{
			// Boolean = Boolean: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.BooleanNode{
					Value: false,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
			Output: true,
		},
		{
			// Boolean = Boolean: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
			Output: false,
		},
		{
			// Boolean = another type: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.NullNode{},
			},
			Output: false,
		},
		{
			// Null = Null: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS:  &jparse.NullNode{},
				RHS:  &jparse.NullNode{},
			},
			Output: true,
		},
		{
			// Null = another type: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS:  &jparse.NullNode{},
				RHS:  &jparse.StringNode{},
			},
			Output: false,
		},
		{
			// Array = Array: true
			// (Note: as of jsonata 1.7 arrays are compared with a deep comparison)
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.AssignmentNode{
						Name:  "x",
						Value: &jparse.ArrayNode{},
					},
					&jparse.ComparisonOperatorNode{
						Type: jparse.ComparisonEqual,
						LHS: &jparse.VariableNode{
							Name: "x",
						},
						RHS: &jparse.VariableNode{
							Name: "x",
						},
					},
				},
			},
			Output: true,
		},
		{
			// Array = Array: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS:  &jparse.ArrayNode{},
				RHS:  &jparse.ArrayNode{},
			},
			Output: true,
		},
		{
			// Object = Object: true
			// (Note: must be the same object in memory)
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.AssignmentNode{
						Name:  "x",
						Value: &jparse.ObjectNode{},
					},
					&jparse.ComparisonOperatorNode{
						Type: jparse.ComparisonEqual,
						LHS: &jparse.VariableNode{
							Name: "x",
						},
						RHS: &jparse.VariableNode{
							Name: "x",
						},
					},
				},
			},
			Output: true,
		},
		{
			// Object = Object: false
			// (Note: as of jsonata 1.7 objects are compared with a deep comparison)
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS:  &jparse.ObjectNode{},
				RHS:  &jparse.ObjectNode{},
			},
			Output: true,
		},
		{
			// Lambda = Lambda: true
			// (Note: must be the same object in memory)
			Input: &jparse.BlockNode{
				Exprs: []jparse.Node{
					&jparse.AssignmentNode{
						Name: "f",
						Value: &jparse.LambdaNode{
							Body: &jparse.NullNode{},
						},
					},
					&jparse.ComparisonOperatorNode{
						Type: jparse.ComparisonEqual,
						LHS: &jparse.VariableNode{
							Name: "f",
						},
						RHS: &jparse.VariableNode{
							Name: "f",
						},
					},
				},
			},
			Output: true,
		},
		{
			// Lambda = Lambda: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.LambdaNode{
					Body: &jparse.NullNode{},
				},
				RHS: &jparse.LambdaNode{
					Body: &jparse.NullNode{},
				},
			},
			Output: false,
		},
		{
			// Go Callable = Go Callable: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.VariableNode{
					Name: "f1",
				},
				RHS: &jparse.VariableNode{
					Name: "f1",
				},
			},
			Exts: map[string]Extension{
				"f1": {
					Func: func() interface{} { return nil },
				},
			},
			Output: true,
		},
		{
			// Go Callable = Go Callable: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.VariableNode{
					Name: "f1",
				},
				RHS: &jparse.VariableNode{
					Name: "f2",
				},
			},
			Exts: map[string]Extension{
				"f1": {
					Func: func() interface{} { return nil },
				},
				"f2": {
					Func: func() interface{} { return nil },
				},
			},
			Output: false,
		},
		{
			// Number < Number: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 2,
				},
			},
			Output: true,
		},
		{
			// Number < Number: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.NumberNode{
					Value: 2,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
			Output: false,
		},
		{
			// String < String: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.StringNode{
					Value: "cats",
				},
				RHS: &jparse.StringNode{
					Value: "dogs",
				},
			},
			Output: true,
		},
		{
			// String < String: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.StringNode{
					Value: "dogs",
				},
				RHS: &jparse.StringNode{
					Value: "cats",
				},
			},
			Output: false,
		},
		{
			// x != y: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonNotEqual,
				LHS:  &jparse.NullNode{},
				RHS:  &jparse.BooleanNode{},
			},
			Output: true,
		},
		{
			// x != y: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonNotEqual,
				LHS:  &jparse.NullNode{},
				RHS:  &jparse.NullNode{},
			},
			Output: false,
		},
		{
			// x > y: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreater,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 0,
				},
			},
			Output: true,
		},
		{
			// x > y: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreater,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "hello",
				},
			},
			Output: false,
		},
		{
			// x >= y: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreaterEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
			Output: true,
		},
		{
			// x >= y: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreaterEqual,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
			Output: false,
		},
		{
			// x <= y: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLessEqual,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "hello",
				},
			},
			Output: true,
		},
		{
			// x <= y: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLessEqual,
				LHS: &jparse.NumberNode{
					Value: 1,
				},
				RHS: &jparse.NumberNode{
					Value: 0,
				},
			},
			Output: false,
		},
		{
			// x in y: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "hello",
				},
			},
			Output: true,
		},
		{
			// x in y: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
			Output: false,
		},
		{
			// x in another type: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.NumberNode{
					Value: 1,
				},
			},
			Output: false,
		},
		{
			// x in [y]: true
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "bonjour",
						},
						&jparse.StringNode{
							Value: "hola",
						},
						&jparse.StringNode{
							Value: "ciao",
						},
						&jparse.StringNode{
							Value: "hello",
						},
					},
				},
			},
			Output: true,
		},
		{
			// x in [y]: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.ArrayNode{
					Items: []jparse.Node{
						&jparse.StringNode{
							Value: "au revoir",
						},
						&jparse.StringNode{
							Value: "adiós",
						},
						&jparse.StringNode{
							Value: "ciao",
						},
						&jparse.StringNode{
							Value: "goodbye",
						},
					},
				},
			},
			Output: false,
		},
		{
			// x in []: false
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.ArrayNode{},
			},
			Output: false,
		},
		{
			// Left side returns an error. Return the error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// an error on the right side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// a non-comparable on the right side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.BooleanNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// an undefined right side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// Left side is non-comparable. Return an error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS:  &jparse.BooleanNode{},
				RHS:  &jparse.NumberNode{},
			},
			Error: &EvalError{
				Type:  ErrNonComparableLHS,
				Token: "false",
				Value: "<",
			},
		},
		{
			// A non-comparable on the left side takes precedence
			// over a non-comparable on the right side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreater,
				LHS:  &jparse.BooleanNode{},
				RHS:  &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonComparableLHS,
				Token: "false",
				Value: ">",
			},
		},
		{
			// A non-comparable on the left side takes precedence
			// over an undefined right side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonGreaterEqual,
				LHS:  &jparse.BooleanNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonComparableLHS,
				Token: "false",
				Value: ">=",
			},
		},
		{
			// Right side returns an error. Return the error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonIn,
				LHS:  &jparse.NumberNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the right side takes precedence over
			// a non-comparable on the left side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS:  &jparse.BooleanNode{},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// An error on the right side takes precedence over
			// an undefined left side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.FunctionCallNode{
					Func: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Right side is non-comparable. Return an error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS:  &jparse.NumberNode{},
				RHS:  &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonComparableRHS,
				Token: "null",
				Value: "<",
			},
		},
		{
			// A non-comparable right side takes precedence over
			// an undefined left side.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLessEqual,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NullNode{},
			},
			Error: &EvalError{
				Type:  ErrNonComparableRHS,
				Token: "null",
				Value: "<=",
			},
		},
		{
			// Type mismatch. Return an error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS:  &jparse.NumberNode{},
				RHS:  &jparse.StringNode{},
			},
			Error: &EvalError{
				Type:  ErrTypeMismatch,
				Value: "<",
			},
		},
		{
			// Type mismatch in a non-comparable operation
			// is not an error.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS:  &jparse.NumberNode{},
				RHS:  &jparse.StringNode{},
			},
			Output: false,
		},
		{
			// Left side is undefined. Return false.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonEqual,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.StringNode{},
			},
			Output: false,
		},
		{
			// Right side is undefined. Return false.
			Input: &jparse.ComparisonOperatorNode{
				Type: jparse.ComparisonLess,
				LHS:  &jparse.NumberNode{},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: false,
		},
	})
}

func TestEvalBooleanOperator(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// true and true: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: true,
		},
		{
			// true and false: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
			Output: false,
		},
		{
			// true and undefined: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: false,
		},
		{
			// false and true: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: false,
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: false,
		},
		{
			// undefined and true: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: false,
		},
		{
			// false and false: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: false,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
			Output: false,
		},
		{
			// undefined and undefined: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: false,
		},
		{
			// true or true: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: true,
		},
		{
			// true or false: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.BooleanNode{
					Value: false,
				},
			},
			Output: true,
		},
		{
			// true or undefined: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: true,
		},
		{
			// false or true: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.BooleanNode{
					Value: false,
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: true,
		},
		{
			// undefined or true: true
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Output: true,
		},
		{
			// undefined or undefined: false
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanOr,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: false,
		},
		{
			// Error on left side. Return the error.
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.BooleanNode{
					Value: true,
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// an error on the right side.
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// An error on the left side takes precedence over
			// an undefined right side.
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "false",
				Value: "-",
			},
		},
		{
			// Error on the right side. Return the error.
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.BooleanNode{
					Value: true,
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An error on the right side takes precedence over
			// an undefined left side.
			Input: &jparse.BooleanOperatorNode{
				Type: jparse.BooleanAnd,
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
	})
}

func TestEvalStringConcatenation(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			// string & string
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
			Output: "helloworld",
		},
		{
			// string & undefined
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.StringNode{
					Value: "hello",
				},
				RHS: &jparse.VariableNode{
					Name: "x",
				},
			},
			Output: "hello",
		},
		{
			// undefined & string
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.StringNode{
					Value: "world",
				},
			},
			Output: "world",
		},
		{
			// undefined & undefined
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.VariableNode{
					Name: "x",
				},
				RHS: &jparse.VariableNode{
					Name: "y",
				},
			},
			Output: "",
		},
		{
			// Left side evaluation error. Return the error.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.StringNode{},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An evaluation error on the left side takes
			// precedence over an evaluation error on the
			// right side.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An evaluation error on the left side takes
			// precedence over a conversion error on the
			// right side.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
				RHS: &jparse.NumberNode{
					Value: math.NaN(),
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// Left side conversion error. Return the error.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.NumberNode{
					Value: math.NaN(),
				},
				RHS: &jparse.StringNode{},
			},
			Error: &jlib.Error{
				Type: jlib.ErrNaNInf,
				Func: "string",
			},
		},
		{
			// Right side evaluation error. Return the error.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.StringNode{},
				RHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// An evaluation error on the right side takes
			// precedence over a conversion error on the left
			// side.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.NumberNode{
					Value: math.NaN(),
				},
				RHS: &jparse.NegationNode{
					RHS: &jparse.NullNode{},
				},
			},
			Error: &EvalError{
				Type:  ErrNonNumberRHS,
				Token: "null",
				Value: "-",
			},
		},
		{
			// Right side conversion error. Return the error.
			Input: &jparse.StringConcatenationNode{
				LHS: &jparse.StringNode{},
				RHS: &jparse.NumberNode{
					Value: math.NaN(),
				},
			},
			Error: &jlib.Error{
				Type: jlib.ErrNaNInf,
				Func: "string",
			},
		},
	})
}

func TestEvalName(t *testing.T) {
	testEvalTestCases(t, []evalTestCase{
		{
			Input: &jparse.NameNode{
				Value: "Field",
			},
			Output: nil,
		},
		{
			Input: &jparse.NameNode{
				Value: "Field",
			},
			Data: map[string]interface{}{
				"Field": 9.99,
			},
			Output: 9.99,
		},
		{
			Input: &jparse.NameNode{
				Value: "Field",
			},
			Data: struct {
				Field float64
			}{
				Field: 9.99,
			},
			Output: 9.99,
		},
		{
			Input: &jparse.NameNode{
				Value: "Field",
			},
			Data: []interface{}{
				map[string]interface{}{
					"Field": "one",
				},
				map[string]interface{}{
					"Field": 2.0,
				},
				map[string]interface{}{
					"Field": true,
				},
			},
			Output: []interface{}{
				"one",
				2.0,
				true,
			},
		},
	})
}

func testEvalTestCases(t *testing.T, tests []evalTestCase) {

	for _, test := range tests {

		env := newEnvironment(nil, len(test.Vars))

		for name, v := range test.Vars {
			env.bind(name, reflect.ValueOf(v))
		}

		for name, ext := range test.Exts {
			f, err := newGoCallable(name, ext)
			if err != nil {
				t.Fatalf("newGoCallable error: %s", err)
			}
			env.bind(name, reflect.ValueOf(f))
		}

		v, err := eval(test.Input, reflect.ValueOf(test.Data), env)
		var output interface{}
		if v.IsValid() && v.CanInterface() {
			output = v.Interface()
		}

		equal := test.Equals
		if test.Equals == nil {
			equal = reflect.DeepEqual
		}

		if !equal(output, test.Output) {
			t.Errorf("%s: Expected: %v, got: %v", test.Input, test.Output, output)
		}

		if err != nil && test.Error != nil {
			assert.EqualError(t, err, test.Error.Error())
		}

	}
}
