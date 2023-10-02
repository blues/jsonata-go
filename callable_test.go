// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"errors"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xiatechs/jsonata-go/jparse"
	"github.com/xiatechs/jsonata-go/jtypes"
)

var (
	typeStringRegex    = reflect.TypeOf((*stringRegex)(nil)).Elem()
	typeOptionalString = reflect.TypeOf((*jtypes.OptionalString)(nil)).Elem()
)

// stringRegex is a Variant type that accepts string
// or regex arguments.
type stringRegex reflect.Value

func (sr stringRegex) ValidTypes() []reflect.Type {
	return []reflect.Type{
		typeString,
		typeRegexPtr,
	}
}

// badVariant1 is a Variant that is not derived from
// reflect.Value.
type badVariant1 struct{}

func (badVariant1) ValidTypes() []reflect.Type {
	return nil
}

// badVariant2 is a Variant that has no valid jtypes.
type badVariant2 reflect.Value

func (badVariant2) ValidTypes() []reflect.Type {
	return nil
}

// badVariant3 is a Variant that has an Optional valid type.
type badVariant3 reflect.Value

func (badVariant3) ValidTypes() []reflect.Type {
	return []reflect.Type{
		typeString,
		typeOptionalString,
	}
}

// badVariant4 is a Variant that has a Variant valid type.
type badVariant4 reflect.Value

func (badVariant4) ValidTypes() []reflect.Type {
	return []reflect.Type{
		typeString,
		typeStringRegex,
	}
}

// badOptional1 is a type that implements the Optional and
// Variant interfaces.
type badOptional1 struct{}

func (badOptional1) IsSet() bool                { return false }
func (badOptional1) Set(reflect.Value)          {}
func (badOptional1) Type() reflect.Type         { return typeString }
func (badOptional1) ValidTypes() []reflect.Type { return nil }

// badOptional2 is an Optional with an Optional underlying type.
type badOptional2 struct{}

func (badOptional2) IsSet() bool        { return false }
func (badOptional2) Set(reflect.Value)  {}
func (badOptional2) Type() reflect.Type { return typeOptionalString }

type newGoCallableTest struct {
	Name   string
	Func   interface{}
	Result *goCallable
	Fail   bool
}

func TestNewGoCallable(t *testing.T) {

	typeInt := reflect.TypeOf((*int)(nil)).Elem()

	testNewGoCallable(t, []newGoCallableTest{
		{
			// Error: Func is nil.
			Name: "nil",
			Fail: true,
		},
		{
			// Error: Func is not a function.
			Name: "int",
			Func: 100,
			Fail: true,
		},
		{
			// Error: Function has 0 return values.
			Name: "return0",
			Func: func() {},
			Fail: true,
		},
		{
			// Error: Function has 2 return values but the second
			// value is not an error.
			Name: "nonerror",
			Func: func() (int, int) { return 0, 0 },
			Fail: true,
		},
		{
			// Error: Function has 3 return values.
			Name: "return3",
			Func: func() (int, int, int) { return 0, 0, 0 },
			Fail: true,
		},
		{
			// Error: Parameter is an Optional and a Variant.
			Name: "badOptional1",
			Func: func(badOptional1) int { return 0 },
			Fail: true,
		},
		{
			// Error: Optional parameter has an Optional subtype.
			Name: "badOptional2",
			Func: func(badOptional2) int { return 0 },
			Fail: true,
		},
		{
			// Error: Optional parameter followed by a non-optional
			// parameter.
			Name: "optional_nonoptional",
			Func: func(jtypes.OptionalString, string) int { return 0 },
			Fail: true,
		},
		{
			// Error: Variadic optional parameter.
			Name: "variadic_optional",
			Func: func(...jtypes.OptionalString) int { return 0 },
			Fail: true,
		},
		{
			// Error: Variant type not derived from reflect.Value.
			Name: "badvariant1",
			Func: func(badVariant1) int { return 0 },
			Fail: true,
		},
		{
			// Error: Variant does not return any valid jtypes.
			Name: "badvariant2",
			Func: func(badVariant2) int { return 0 },
			Fail: true,
		},
		{
			// Error: Variant has an Optional valid type.
			Name: "badvariant3",
			Func: func(badVariant3) int { return 0 },
			Fail: true,
		},
		{
			// Error: Variant has a Variant valid type.
			Name: "badvariant4",
			Func: func(badVariant4) int { return 0 },
			Fail: true,
		},
		{
			// Function with 1 return value.
			Name: "return1",
			Func: func() int { return 0 },
			Result: &goCallable{
				callableName: callableName{
					name: "return1",
				},
			},
		},
		{
			// Function with 2 return values.
			Name: "return2",
			Func: func() (int, error) { return 0, nil },
			Result: &goCallable{
				callableName: callableName{
					name: "return2",
				},
			},
		},
		{
			// Standard function.
			Name: "standard",
			Func: func(string, int) int { return 0 },
			Result: &goCallable{
				callableName: callableName{
					name: "standard",
				},
				params: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeInt,
					},
				},
			},
		},
		{
			// Variadic function.
			Name: "variadic",
			Func: func(string, ...int) int { return 0 },
			Result: &goCallable{
				callableName: callableName{
					name: "variadic",
				},
				params: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeInt,
					},
				},
				isVariadic: true,
			},
		},
		{
			// Function with an Optional parameter.
			Name: "optional",
			Func: func(string, jtypes.OptionalString) int { return 0 },
			Result: &goCallable{
				callableName: callableName{
					name: "optional",
				},
				params: []goCallableParam{
					{
						t: typeString,
					},
					{
						t:     typeOptionalString,
						isOpt: true,
						optType: &goCallableParam{
							t: typeString,
						},
					},
				},
			},
		},
		{
			// Func with a Variant parameter.
			Name: "variant",
			Func: func(string, stringRegex) int { return 0 },
			Result: &goCallable{
				callableName: callableName{
					name: "variant",
				},
				params: []goCallableParam{
					{
						t: typeString,
					},
					{
						t:     typeStringRegex,
						isVar: true,
						varTypes: []goCallableParam{
							{
								t: typeString,
							},
							{
								t: typeRegexPtr,
							},
						},
					},
				},
			},
		},
	})
}

func testNewGoCallable(t *testing.T, tests []newGoCallableTest) {

	for _, test := range tests {

		res, err := newGoCallable(test.Name, Extension{
			Func: test.Func,
		})

		if res != nil {
			res.fn = undefined
		}

		if (err != nil) != test.Fail {
			t.Errorf("%s: expected error %v, got %v", test.Name, test.Fail, err)
		}

		if !reflect.DeepEqual(res, test.Result) {
			t.Errorf("%s: expected %v, got %v", test.Name, test.Result, res)
		}
	}
}

type goCallableTest struct {
	Name      string
	Ext       Extension
	Context   interface{}
	Args      []interface{}
	Output    interface{}
	Error     error
	Undefined bool
}

func TestGoCallable(t *testing.T) {
	testGoCallable(t, []goCallableTest{
		{
			// Error: Not enough arguments
			Name: "argCount1",
			Ext: Extension{
				Func: func(string, int) int { return 0 },
			},
			Args: []interface{}{
				"hello",
			},
			Error: &ArgCountError{
				Func:     "argCount1",
				Expected: 2,
				Received: 1,
			},
		},
		{
			// Error: Not enough arguments (variadic)
			Name: "argCount2",
			Ext: Extension{
				Func: func(string, ...int) int { return 0 },
			},
			Error: &ArgCountError{
				Func:     "argCount2",
				Expected: 2,
				Received: 0,
			},
		},
		{
			// Error: Too many arguments
			Name: "argCount3",
			Ext: Extension{
				Func: func(string) int { return 0 },
			},
			Args: []interface{}{
				"hello",
				"world",
			},
			Error: &ArgCountError{
				Func:     "argCount3",
				Expected: 1,
				Received: 2,
			},
		},
		{
			// Error: Bad type
			Name: "argType1",
			Ext: Extension{
				Func: func(string) int { return 0 },
			},
			Args: []interface{}{
				65,
			},
			Error: &ArgTypeError{
				Func:  "argType1",
				Which: 1,
			},
		},
		{
			// Error: Bad type (variadic)
			Name: "argType2",
			Ext: Extension{
				Func: func(...string) int { return 0 },
			},
			Args: []interface{}{
				"hello",
				"world",
				3.14159,
			},
			Error: &ArgTypeError{
				Func:  "argType2",
				Which: 3,
			},
		},
		{
			// Function returns an error
			Name: "error",
			Ext: Extension{
				Func: func() (int, error) {
					return 0, errors.New("test error")
				},
			},
			Error: errors.New("test error"),
		},
		{
			// Function returns jtypes.ErrUndefined
			Name: "errUndefined",
			Ext: Extension{
				Func: func() (int, error) {
					return 0, jtypes.ErrUndefined
				},
			},
			Undefined: true,
		},
		{
			// Extension with UndefinedHandler
			Name: "undefinedHandler",
			Ext: Extension{
				Func: func() int { return 0 },
				UndefinedHandler: func([]reflect.Value) bool {
					return true
				},
			},
			Undefined: true,
		},
		{
			// Standard Extension
			Name: "repeat",
			Ext: Extension{
				Func: func(s string, n int) string {
					return strings.Repeat(s, n)
				},
			},
			Args: []interface{}{
				"*",
				5,
			},
			Output: "*****",
		},
		{
			// Extension with ContextHandler
			Name: "contextHandler",
			Ext: Extension{
				Func: func(s string, n int) string {
					return strings.Repeat(s, n)
				},
				EvalContextHandler: func(argv []reflect.Value) bool {
					return len(argv) < 2
				},
			},
			Context: "x",
			Args: []interface{}{
				3,
			},
			Output: "xxx",
		},
		{
			// Variadic function
			Name: "variadic1",
			Ext: Extension{
				Func: func(nums ...int) []int {
					sort.Ints(nums)
					return nums
				},
			},
			Args: []interface{}{
				3,
				0,
				2,
				-1,
			},
			Output: []int{
				-1,
				0,
				2,
				3,
			},
		},
		{
			// Variadic function (no variadic arguments)
			Name: "variadic2",
			Ext: Extension{
				Func: func(nums ...int) int {
					return len(nums)
				},
			},
			Output: 0,
		},
		{
			// Optional parameter (set)
			Name: "optional_set",
			Ext: Extension{
				Func: func(in jtypes.OptionalInt) interface{} {
					return map[string]interface{}{
						"set":   in.IsSet(),
						"value": in.Int,
					}
				},
			},
			Args: []interface{}{
				100.0,
			},
			Output: map[string]interface{}{
				"set":   true,
				"value": 100,
			},
		},
		{
			// Optional parameter (not set)
			Name: "optional_notset",
			Ext: Extension{
				Func: func(in jtypes.OptionalInt) interface{} {
					return map[string]interface{}{
						"set":   in.IsSet(),
						"value": in.Int,
					}
				},
			},
			Output: map[string]interface{}{
				"set":   false,
				"value": 0,
			},
		},
		{
			// Callable parameter
			Name: "callable",
			Ext: Extension{
				Func: func(f jtypes.Callable) string {
					return f.Name()
				},
			},
			Args: []interface{}{
				&undefinedCallable{
					callableName: callableName{
						name: "test",
					},
				},
			},
			Output: "test",
		},
	})
}

func testGoCallable(t *testing.T, tests []goCallableTest) {

	for _, test := range tests {

		var output interface{}
		var argv []reflect.Value

		fn, err := newGoCallable(test.Name, test.Ext)
		if err != nil {
			t.Errorf("%s: newGoCallable returned %v", test.Name, err)
			continue
		}

		if test.Context != nil {
			fn.SetContext(reflect.ValueOf(test.Context))
		}

		if argc := len(test.Args); argc > 0 {
			argv = make([]reflect.Value, argc)
			for i := range argv {
				argv[i] = reflect.ValueOf(test.Args[i])
			}
		}

		res, err := fn.Call(argv)

		if res.IsValid() && res.CanInterface() {
			output = res.Interface()
		}

		if test.Undefined {
			if res != undefined {
				t.Errorf("%s: expected undefined, got %v", test.Name, res)
			}
		} else {
			if !reflect.DeepEqual(output, test.Output) {
				t.Errorf("%s: expected %v, got %v", test.Name, test.Output, output)
			}
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: expected error %v, got %v", test.Name, test.Error, err)
		}
	}
}

type goCallableArgTest struct {
	Input   interface{}
	Param   goCallableParam
	Output  interface{}
	Compare func(interface{}, interface{}) bool
	Fail    bool
}

func TestProcessGoCallableArg(t *testing.T) {

	re := regexp.MustCompile("ab+")

	typeInt := reflect.TypeOf((*int)(nil)).Elem()
	typeFloat64 := reflect.TypeOf((*float64)(nil)).Elem()
	typeBool := reflect.TypeOf((*bool)(nil)).Elem()
	typeString := reflect.TypeOf((*string)(nil)).Elem()
	typeByteSlice := reflect.TypeOf((*[]byte)(nil)).Elem()

	testProcessGoCallableArgs(t, []goCallableArgTest{

		// int

		{
			// int to int
			Input: -1,
			Param: goCallableParam{
				t: typeInt,
			},
			Output: -1,
		},
		{
			// uint to int
			Input: uint(100),
			Param: goCallableParam{
				t: typeInt,
			},
			Output: 100,
		},
		{
			// int64 to int
			Input: int64(1e6),
			Param: goCallableParam{
				t: typeInt,
			},
			Output: 1000000,
		},
		{
			// float64 to int
			Input: float64(1e12),
			Param: goCallableParam{
				t: typeInt,
			},
			Output: 1000000000000,
		},
		{
			// float64 to int (truncated)
			Input: 3.141592,
			Param: goCallableParam{
				t: typeInt,
			},
			Output: 3,
		},
		{
			// undefined to int
			Param: goCallableParam{
				t: typeInt,
			},
			Fail: true,
		},
		{
			// invalid type to int
			Input: struct{}{},
			Param: goCallableParam{
				t: typeInt,
			},
			Fail: true,
		},

		// float64

		{
			// float64 to float64
			Input: 3.141592,
			Param: goCallableParam{
				t: typeFloat64,
			},
			Output: 3.141592,
		},
		{
			// int to float64
			Input: 100,
			Param: goCallableParam{
				t: typeFloat64,
			},
			Output: 100.0,
		},
		{
			// undefined to float64
			Param: goCallableParam{
				t: typeFloat64,
			},
			Fail: true,
		},
		{
			// invalid type to float64
			Input: struct{}{},
			Param: goCallableParam{
				t: typeFloat64,
			},
			Fail: true,
		},

		// bool

		{
			// bool to bool
			Input: true,
			Param: goCallableParam{
				t: typeBool,
			},
			Output: true,
		},
		{
			// undefined to bool
			Param: goCallableParam{
				t: typeBool,
			},
			Fail: true,
		},
		{
			// invalid type to bool
			Input: struct{}{},
			Param: goCallableParam{
				t: typeBool,
			},
			Fail: true,
		},

		// string

		{
			// string to string
			Input: "hello",
			Param: goCallableParam{
				t: typeString,
			},
			Output: "hello",
		},
		{
			// byte slice to string
			Input: []byte("hello"),
			Param: goCallableParam{
				t: typeString,
			},
			Output: "hello",
		},
		{
			// int to string
			// Note: Conversion from int to string (e.g. string(65))
			// is valid in Go but it isn't permitted in JSONata.
			Input: 65,
			Param: goCallableParam{
				t: typeString,
			},
			Fail: true,
		},
		{
			// rune to string
			// Note: Conversion from rune to string (e.g. string('a'))
			// is valid in Go but it isn't permitted in JSONata.
			Input: 'a',
			Param: goCallableParam{
				t: typeString,
			},
			Fail: true,
		},
		{
			// undefined to string
			Param: goCallableParam{
				t: typeString,
			},
			Fail: true,
		},
		{
			// invalid type to string
			Input: struct{}{},
			Param: goCallableParam{
				t: typeString,
			},
			Fail: true,
		},

		// byte slice

		{
			// byte slice to byte slice
			Input: []byte("hello"),
			Param: goCallableParam{
				t: typeByteSlice,
			},
			Output: []byte("hello"),
		},
		{
			// string to byte slice
			Input: "hello",
			Param: goCallableParam{
				t: typeByteSlice,
			},
			Output: []byte("hello"),
		},
		{
			// undefined to byte slice
			Param: goCallableParam{
				t: typeByteSlice,
			},
			Fail: true,
		},
		{
			// invalid type to byte slice
			Input: struct{}{},
			Param: goCallableParam{
				t: typeByteSlice,
			},
			Fail: true,
		},

		// reflect.Value
		// A Value parameter can accept any type of argument.

		{
			Input: "hello",
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf("hello"),
		},
		{
			Input: -100,
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf(-100),
		},
		{
			Input: 3.141592,
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf(3.141592),
		},
		{
			Input: false,
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf(false),
		},
		{
			Input: []interface{}{},
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf([]interface{}{}),
		},
		{
			Input: map[string]interface{}{},
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Compare: equalReflectValue,
			Output:  reflect.ValueOf(map[string]interface{}{}),
		},
		{
			Param: goCallableParam{
				t: jtypes.TypeValue,
			},
			Output: undefined,
		},

		// jtypes.Callable

		{
			// Callable to Callable
			Input: newRegexCallable(re),
			Param: goCallableParam{
				t: jtypes.TypeCallable,
			},
			Output: &regexCallable{
				callableName: callableName{
					name: "ab+",
				},
				re: re,
			},
		},

		// jtypes.Optional

		{
			// underlying type to Optional
			Input: "hello",
			Param: goCallableParam{
				t:     typeOptionalString,
				isOpt: true,
				optType: &goCallableParam{
					t: typeString,
				},
			},
			Output: jtypes.NewOptionalString("hello"),
		},
		{
			// undefined to Optional
			Param: goCallableParam{
				t:     typeOptionalString,
				isOpt: true,
				optType: &goCallableParam{
					t: typeString,
				},
			},
			Output: jtypes.OptionalString{},
		},
		{
			// invalid type to Optional
			Input: struct{}{},
			Param: goCallableParam{
				t:     typeOptionalString,
				isOpt: true,
				optType: &goCallableParam{
					t: typeString,
				},
			},
			Fail: true,
		},

		// jtypes.Variant

		{
			// supported type to Variant
			Input: "hello",
			Param: goCallableParam{
				t:     typeStringRegex,
				isVar: true,
				varTypes: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeRegexPtr,
					},
				},
			},
			Compare: equalStringRegex,
			Output:  stringRegex(reflect.ValueOf("hello")),
		},
		{
			// supported type to Variant
			Input: re,
			Param: goCallableParam{
				t:     typeStringRegex,
				isVar: true,
				varTypes: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeRegexPtr,
					},
				},
			},
			Compare: equalStringRegex,
			Output:  stringRegex(reflect.ValueOf(re)),
		},
		{
			// unsupported type to Variant
			Input: 65,
			Param: goCallableParam{
				t:     typeStringRegex,
				isVar: true,
				varTypes: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeRegexPtr,
					},
				},
			},
			Fail: true,
		},
		{
			// undefined to Variant
			Param: goCallableParam{
				t:     typeStringRegex,
				isVar: true,
				varTypes: []goCallableParam{
					{
						t: typeString,
					},
					{
						t: typeRegexPtr,
					},
				},
			},
			Fail: true,
		},

		// jtypes.Convertible

		{
			// Convertible to supported type
			Input: newRegexCallable(re),
			Param: goCallableParam{
				t: typeRegexPtr,
			},
			Output: re,
		},
		{
			// Convertible to unsupported type
			Input: newRegexCallable(re),
			Param: goCallableParam{
				t: typeString,
			},
			Fail: true,
		},
	})
}

func testProcessGoCallableArgs(t *testing.T, tests []goCallableArgTest) {

	for _, test := range tests {

		res, ok := processGoCallableArg(reflect.ValueOf(test.Input), test.Param)

		var output interface{}
		if res.IsValid() && res.CanInterface() {
			output = res.Interface()
		}

		isEqual := test.Compare
		if isEqual == nil {
			isEqual = reflect.DeepEqual
		}

		if !isEqual(output, test.Output) {
			t.Errorf("%v => %s: expected %v, got %v", test.Input, test.Param.t, test.Output, res)
		}

		if ok == test.Fail {
			t.Errorf("%v => %s: expected OK %v, got %v", test.Input, test.Param.t, !test.Fail, ok)
		}
	}
}

func equalStringRegex(in1, in2 interface{}) bool {

	v1, ok := in1.(stringRegex)
	if !ok {
		return false
	}

	v2, ok := in2.(stringRegex)
	if !ok {
		return false
	}

	return reflect.DeepEqual(reflect.Value(v1).Interface(), reflect.Value(v2).Interface())
}

func equalReflectValue(in1, in2 interface{}) bool {

	v1, ok := in1.(reflect.Value)
	if !ok {
		return false
	}

	v2, ok := in2.(reflect.Value)
	if !ok {
		return false
	}

	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

type lambdaCallableTest struct {
	Name       string
	Typed      bool
	Params     []jparse.Param
	Body       jparse.Node
	ParamNames []string
	Args       []interface{}
	Vars       map[string]interface{}
	Context    interface{}
	Output     interface{}
	Error      error
	Undefined  bool
}

func TestLambdaCallable(t *testing.T) {
	testLambdaCallable(t, []lambdaCallableTest{
		{
			Name: "multiply",
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
			Args: []interface{}{
				3.6,
				100,
			},
			Output: float64(360),
		},
		{
			// Lambdas can access values from the containing scope.
			Name: "global",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			Vars: map[string]interface{}{
				"x": "marks the spot",
			},
			Output: "marks the spot",
		},
		{
			// Lambdas also have their own local scope. Lambda
			// arguments shadow values from the containing scope.
			Name: "local",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Args: []interface{}{
				100,
			},
			Vars: map[string]interface{}{
				"x": "marks the spot",
			},
			Output: 100,
		},
		{
			// Shadowing occurs even when the argument is not
			// defined by the caller.
			Name: "undefined",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Vars: map[string]interface{}{
				"x": "marks the spot",
			},
			Undefined: true,
		},
		{
			// Typed lambda, no params, no args.
			Name:   "typed",
			Body:   &jparse.NumberNode{},
			Typed:  true,
			Output: float64(0),
		},
		{
			// Typed lambda, no params, some args.
			Name:  "typed",
			Body:  &jparse.NumberNode{},
			Typed: true,
			Args: []interface{}{
				1,
				2,
			},
			Error: &ArgCountError{
				Func:     "typed",
				Expected: 0,
				Received: 2,
			},
		},
		{
			// Typed lambda, contextable first argument.
			Name: "context",
			Body: &jparse.StringConcatenationNode{
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
			Typed: true,
			Params: []jparse.Param{
				{
					Type:   jparse.ParamTypeAny,
					Option: jparse.ParamContextable,
				},
				{
					Type: jparse.ParamTypeAny,
				},
			},
			Args: []interface{}{
				"world",
			},
			Context: "hello",
			Output:  "helloworld",
		},
		{
			// Typed lambda, optional argument.
			Name: "argcount",
			Body: &jparse.VariableNode{
				Name: "y",
			},
			ParamNames: []string{
				"x",
				"y",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeAny,
				},
				{
					Type:   jparse.ParamTypeAny,
					Option: jparse.ParamOptional,
				},
			},
			Args: []interface{}{
				100,
			},
			Undefined: true,
		},
		{
			// Typed lambda, not enough arguments.
			Name: "fewer",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
				"y",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeAny,
				},
				{
					Type: jparse.ParamTypeAny,
				},
			},
			Args: []interface{}{
				0,
			},
			Error: &ArgCountError{
				Func:     "fewer",
				Expected: 2,
				Received: 1,
			},
		},
		{
			// Typed lambda, too many arguments.
			Name: "greater",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
				"y",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeAny,
				},
				{
					Type: jparse.ParamTypeAny,
				},
			},
			Args: []interface{}{
				0,
				1,
				2,
			},
			Error: &ArgCountError{
				Func:     "greater",
				Expected: 2,
				Received: 3,
			},
		},
		{
			// Typed lambda, valid string argument.
			Name: "string1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeString,
				},
			},
			Args: []interface{}{
				"hello",
			},
			Output: "hello",
		},
		{
			// Typed lambda, invalid string argument.
			Name: "string2",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeString,
				},
			},
			Args: []interface{}{
				0,
			},
			Error: &ArgTypeError{
				Func:  "string2",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid number argument.
			Name: "number1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeNumber,
				},
			},
			Args: []interface{}{
				100,
			},
			Output: 100,
		},
		{
			// Typed lambda, invalid number argument.
			Name: "number2",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeNumber,
				},
			},
			Args: []interface{}{
				false,
			},
			Error: &ArgTypeError{
				Func:  "number2",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid boolean argument.
			Name: "boolean1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeBool,
				},
			},
			Args: []interface{}{
				true,
			},
			Output: true,
		},
		{
			// Typed lambda, invalid boolean argument.
			Name: "boolean2",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeBool,
				},
			},
			Args: []interface{}{
				null,
			},
			Error: &ArgTypeError{
				Func:  "boolean2",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid Callable argument.
			Name: "callable1",
			Body: &jparse.FunctionCallNode{
				Func: &jparse.VariableNode{
					Name: "x",
				},
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeFunc,
				},
			},
			Args: []interface{}{
				&undefinedCallable{},
			},
			Undefined: true,
		},
		{
			// Typed lambda, invalid Callable argument.
			Name: "callable2",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeFunc,
				},
			},
			Args: []interface{}{
				"hello",
			},
			Error: &ArgTypeError{
				Func:  "callable2",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid array argument.
			Name: "array1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeArray,
				},
			},
			Args: []interface{}{
				[]interface{}{
					1,
					2,
					3,
				},
			},
			Output: []interface{}{
				1,
				2,
				3,
			},
		},
		{
			// Typed lambda, non-array argument converted to
			// an array.
			Name: "array2",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeArray,
				},
			},
			Args: []interface{}{
				"hello",
			},
			Output: []interface{}{
				"hello",
			},
		},
		{
			// Typed lambda, valid typed array argument.
			Name: "array3",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeArray,
					SubParams: []jparse.Param{
						{
							Type: jparse.ParamTypeString,
						},
					},
				},
			},
			Args: []interface{}{
				[]interface{}{
					"hello",
					"world",
				},
			},
			Output: []interface{}{
				"hello",
				"world",
			},
		},
		{
			// Typed lambda, invalid typed array argument.
			Name: "array4",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeArray,
					SubParams: []jparse.Param{
						{
							Type: jparse.ParamTypeString,
						},
					},
				},
			},
			Args: []interface{}{
				[]interface{}{
					0,
					1,
				},
			},
			Error: &ArgTypeError{
				Func:  "array4",
				Which: 1,
			},
		},
		{
			// Typed lambda, invalid array argument.
			Name: "array5",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeString,
				},
			},
			Args: []interface{}{
				[]interface{}{
					"hello",
					"world",
				},
			},
			Error: &ArgTypeError{
				Func:  "array5",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid object (map) argument.
			Name: "object1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeObject,
				},
			},
			Args: []interface{}{
				map[string]interface{}{
					"x": "marks the spot",
				},
			},
			Output: map[string]interface{}{
				"x": "marks the spot",
			},
		},
		{
			// Typed lambda, valid object (struct) argument.
			Name: "object2",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeObject,
				},
			},
			Args: []interface{}{
				struct {
					x string
					y int
					z bool
				}{
					x: "hello",
					y: 100,
					z: true,
				},
			},
			Output: struct {
				x string
				y int
				z bool
			}{
				x: "hello",
				y: 100,
				z: true,
			},
		},
		{
			// Typed lambda, invalid object argument.
			Name: "object3",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeObject,
				},
			},
			Args: []interface{}{
				"hello",
			},
			Error: &ArgTypeError{
				Func:  "object3",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid JSON arguments.
			Name: "json1",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"s",
				"n",
				"b",
				"a",
				"o",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeJSON,
				},
				{
					Type: jparse.ParamTypeJSON,
				},
				{
					Type: jparse.ParamTypeJSON,
				},
				{
					Type: jparse.ParamTypeJSON,
				},
				{
					Type: jparse.ParamTypeJSON,
				},
			},
			Args: []interface{}{
				"hello",
				100,
				true,
				[]interface{}{},
				map[string]interface{}{},
			},
			Output: float64(0),
		},
		{
			// Typed lambda, invalid JSON argument.
			Name: "json2",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type: jparse.ParamTypeJSON,
				},
			},
			Args: []interface{}{
				undefinedCallable{},
			},
			Error: &ArgTypeError{
				Func:  "json2",
				Which: 1,
			},
		},
		{
			// Typed lambda, valid variadic argument.
			Name: "variadic1",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type:   jparse.ParamTypeString,
					Option: jparse.ParamVariadic,
				},
			},
			Args: []interface{}{
				"john",
				"paul",
				"ringo",
				"george",
			},
			Output: []interface{}{
				"john",
				"paul",
				"ringo",
				"george",
			},
		},
		{
			// Typed lambda, valid variadic argument.
			Name: "variadic2",
			Body: &jparse.VariableNode{
				Name: "x",
			},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type:   jparse.ParamTypeNumber,
					Option: jparse.ParamVariadic,
				},
			},
			Args: []interface{}{
				100,
			},
			Output: []interface{}{
				100,
			},
		},
		{
			// Typed lambda, invalid variadic argument.
			Name: "variadic3",
			Body: &jparse.NumberNode{},
			ParamNames: []string{
				"x",
			},
			Typed: true,
			Params: []jparse.Param{
				{
					Type:   jparse.ParamTypeString,
					Option: jparse.ParamVariadic,
				},
			},
			Args: []interface{}{
				"john",
				"paul",
				"ringo",
				false,
			},
			Error: &ArgTypeError{
				Func:  "variadic3",
				Which: 4,
			},
		},
	})
}

func testLambdaCallable(t *testing.T, tests []lambdaCallableTest) {

	for i, test := range tests {

		env := newEnvironment(nil, len(test.Vars))
		for name, v := range test.Vars {
			env.bind(name, reflect.ValueOf(v))
		}

		f := &lambdaCallable{
			callableName: callableName{
				name: test.Name,
			},
			body:       test.Body,
			paramNames: test.ParamNames,
			typed:      test.Typed,
			params:     test.Params,
			env:        env,
			context:    reflect.ValueOf(test.Context),
		}

		var args []reflect.Value
		for _, arg := range test.Args {
			args = append(args, reflect.ValueOf(arg))
		}

		v, err := f.Call(args)

		var output interface{}
		if v.IsValid() && v.CanInterface() {
			output = v.Interface()
		}

		if test.Undefined {
			if v != undefined {
				t.Errorf("lambda %d: expected undefined, got %v", i+1, v)
			}
		} else {
			if !reflect.DeepEqual(test.Output, output) {
				t.Errorf("lambda %d: expected %v, got %v", i+1, test.Output, output)
			}
		}

		if !reflect.DeepEqual(test.Error, err) {
			t.Errorf("lambda %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

type partialCallableTest struct {
	Name     string
	Func     jtypes.Callable
	FuncArgs []jparse.Node
	Args     []reflect.Value
	Output   interface{}
	Error    error
}

func TestPartialCallableTest(t *testing.T) {
	testPartialCallableTest(t, []partialCallableTest{
		{
			Func: &lambdaCallable{
				body: &jparse.NumericOperatorNode{
					Type: jparse.NumericMultiply,
					LHS: &jparse.VariableNode{
						Name: "x",
					},
					RHS: &jparse.VariableNode{
						Name: "y",
					},
				},
				paramNames: []string{
					"x",
					"y",
				},
			},
			FuncArgs: []jparse.Node{
				&jparse.NumberNode{
					Value: 2,
				},
				&jparse.PlaceholderNode{},
			},
			Args: []reflect.Value{
				reflect.ValueOf(6),
			},
			Output: float64(12),
		},
		{
			// Error evaluating argument in partial definition.
			// Return the error.
			Func: &lambdaCallable{
				body: &jparse.NumberNode{},
				paramNames: []string{
					"x",
					"y",
				},
			},
			FuncArgs: []jparse.Node{
				&jparse.PlaceholderNode{},
				&jparse.NegationNode{
					RHS: &jparse.BooleanNode{},
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

func testPartialCallableTest(t *testing.T, tests []partialCallableTest) {

	for i, test := range tests {

		name := test.Name
		if name == "" {
			name = "partial"
		}

		f := &partialCallable{
			callableName: callableName{
				name: name,
			},
			fn:   test.Func,
			args: test.FuncArgs,
		}

		v, err := f.Call(test.Args)

		var output interface{}
		if v.IsValid() && v.CanInterface() {
			output = v.Interface()
		}

		if !reflect.DeepEqual(output, test.Output) {
			t.Errorf("partial %d: expected %v, got %v", i+1, test.Output, v)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("partial %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

type transformationCallableTest struct {
	Name      string
	Pattern   jparse.Node
	Updates   jparse.Node
	Deletes   jparse.Node
	Input     interface{}
	Output    interface{}
	Error     error
	Undefined bool
}

func TestTransformationCallable(t *testing.T) {

	data := []map[string]interface{}{
		{
			"value": 1,
			"en":    "one",
			"es":    "uno",
		},
		{
			"value": 2,
			"en":    "two",
			"es":    "dos",
		},
		{
			"value": 3,
			"en":    "three",
			"es":    "tres",
		},
		{
			"value": 4,
			"en":    "four",
			"es":    "cuatro",
		},
		{
			"value": 5,
			"en":    "five",
			"es":    "cinco",
		},
	}

	testTransformationCallable(t, []transformationCallableTest{
		{
			// Update only.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.StringNode{
							Value: "en",
						},
						&jparse.NameNode{
							Value: "es",
						},
					},
					{
						&jparse.StringNode{
							Value: "es",
						},
						&jparse.NameNode{
							Value: "en",
						},
					},
				},
			},
			Input: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": float64(1),
					"es":    "one",
					"en":    "uno",
				},
				map[string]interface{}{
					"value": float64(2),
					"es":    "two",
					"en":    "dos",
				},
				map[string]interface{}{
					"value": float64(3),
					"es":    "three",
					"en":    "tres",
				},
				map[string]interface{}{
					"value": float64(4),
					"es":    "four",
					"en":    "cuatro",
				},
				map[string]interface{}{
					"value": float64(5),
					"es":    "five",
					"en":    "cinco",
				},
			},
		},
		{
			// Delete only (single value).
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Deletes: &jparse.StringNode{
				Value: "es",
			},
			Input: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": float64(1),
					"en":    "one",
				},
				map[string]interface{}{
					"value": float64(2),
					"en":    "two",
				},
				map[string]interface{}{
					"value": float64(3),
					"en":    "three",
				},
				map[string]interface{}{
					"value": float64(4),
					"en":    "four",
				},
				map[string]interface{}{
					"value": float64(5),
					"en":    "five",
				},
			},
		},
		{
			// Delete only (multiple values).
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Deletes: &jparse.ArrayNode{
				Items: []jparse.Node{
					&jparse.StringNode{
						Value: "en",
					},
					&jparse.StringNode{
						Value: "es",
					},
				},
			},
			Input: data,
			Output: []interface{}{
				map[string]interface{}{
					"value": float64(1),
				},
				map[string]interface{}{
					"value": float64(2),
				},
				map[string]interface{}{
					"value": float64(3),
				},
				map[string]interface{}{
					"value": float64(4),
				},
				map[string]interface{}{
					"value": float64(5),
				},
			},
		},
		{
			// Update and delete.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.NameNode{
							Value: "en",
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
						Value: "en",
					},
					&jparse.StringNode{
						Value: "es",
					},
					&jparse.StringNode{
						Value: "value",
					},
				},
			},
			Input: data,
			Output: []interface{}{
				map[string]interface{}{
					"one": float64(1),
				},
				map[string]interface{}{
					"two": float64(2),
				},
				map[string]interface{}{
					"three": float64(3),
				},
				map[string]interface{}{
					"four": float64(4),
				},
				map[string]interface{}{
					"five": float64(5),
				},
			},
		},
		{
			// Non-transformable input.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{
				Pairs: [][2]jparse.Node{
					{
						&jparse.StringNode{
							Value: "key",
						},
						&jparse.StringNode{
							Value: "value",
						},
					},
				},
			},
			Input: []int{
				1,
				2,
				3,
			},
			Output: []interface{}{
				float64(1),
				float64(2),
				float64(3),
			},
		},
		{
			// Arg count <> 1. Return error.
			Name:    "noinput",
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Error: &ArgCountError{
				Func:     "noinput",
				Expected: 1,
				Received: 0,
			},
		},
		{
			// Non-cloneable input. Return error.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Input:   []float64{math.NaN()},
			Error: &EvalError{
				Type: ErrClone,
			},
		},
		{
			// Non-object/array input.
			Name:    "nonobject",
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Input:   "hello world",
			Error: &ArgTypeError{
				Func:  "nonobject",
				Which: 1,
			},
		},
		{
			// Undefined input. Return undefined.
			Pattern:   &jparse.VariableNode{},
			Updates:   &jparse.ObjectNode{},
			Input:     undefined,
			Undefined: true,
		},
		{
			// Error evaluating pattern. Return the error.
			Pattern: &jparse.FunctionCallNode{
				Func: &jparse.NullNode{},
			},
			Updates: &jparse.ObjectNode{},
			Input:   data,
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Error evaluating updates. Return the error.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.FunctionCallNode{
				Func: &jparse.NullNode{},
			},
			Input: data,
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Non-object updates. Return an error.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ArrayNode{},
			Input:   data,
			Error: &EvalError{
				Type:  ErrIllegalUpdate,
				Token: "[]",
			},
		},
		{
			// Error evaluating deletes. Return the error.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Deletes: &jparse.FunctionCallNode{
				Func: &jparse.NullNode{},
			},
			Input: data,
			Error: &EvalError{
				Type:  ErrNonCallable,
				Token: "null",
			},
		},
		{
			// Non-string deletes. Return an error.
			Pattern: &jparse.VariableNode{},
			Updates: &jparse.ObjectNode{},
			Deletes: &jparse.NullNode{},
			Input:   data,
			Error: &EvalError{
				Type:  ErrIllegalDelete,
				Token: "null",
			},
		},
	})
}

func testTransformationCallable(t *testing.T, tests []transformationCallableTest) {

	for i, test := range tests {

		name := test.Name
		if name == "" {
			name = "transform"
		}

		f := &transformationCallable{
			callableName: callableName{
				name: name,
			},
			pattern: test.Pattern,
			updates: test.Updates,
			deletes: test.Deletes,
			env:     newEnvironment(nil, 0),
		}

		var argv []reflect.Value
		switch v := test.Input.(type) {
		case nil:
		case reflect.Value:
			argv = append(argv, v)
		default:
			argv = append(argv, reflect.ValueOf(v))
		}

		v, err := f.Call(argv)

		var output interface{}
		if v.IsValid() && v.CanInterface() {
			output = v.Interface()
		}

		if test.Undefined {
			if v != undefined {
				t.Errorf("transform %d: expected undefined, got %v", i+1, v)
			}
		} else {
			if !reflect.DeepEqual(output, test.Output) {
				t.Errorf("transform %d: expected %v, got %v", i+1, test.Output, v)
			}
		}

		if err != nil && test.Error != nil {
			assert.EqualError(t, err, test.Error.Error())
		}
	}
}

type regexCallableTest struct {
	Expr      string
	Input     interface{}
	Results   interface{}
	Undefined bool
}

func TestRegexCallable(t *testing.T) {
	testRegexCallable(t, []regexCallableTest{
		{
			// No input. Return undefined.
			Expr:      "a.",
			Undefined: true,
		},
		{
			// Non-string input. Return undefined.
			Expr:      "a.",
			Input:     100,
			Undefined: true,
		},
		{
			// No matches. Return undefined.
			Expr:      "a.",
			Input:     "hello world",
			Undefined: true,
		},
		{
			// Matches with no capturing groups.
			Expr:  "a.?",
			Input: "abracadabra",
			Results: map[string]interface{}{
				"match":  "ab",
				"start":  0,
				"end":    2,
				"groups": []string{},
				"next": &matchCallable{
					callableName: callableName{
						sync.Mutex{},
						"next",
					},
					match:  "ac",
					start:  3,
					end:    5,
					groups: []string{},
					next: &matchCallable{
						callableName: callableName{
							sync.Mutex{},
							"next",
						},
						match:  "ad",
						start:  5,
						end:    7,
						groups: []string{},
						next: &matchCallable{
							callableName: callableName{
								sync.Mutex{},
								"next",
							},
							match:  "ab",
							start:  7,
							end:    9,
							groups: []string{},
							next: &matchCallable{
								callableName: callableName{
									sync.Mutex{},
									"next",
								},
								match:  "a",
								start:  10,
								end:    11,
								groups: []string{},
								next: &undefinedCallable{
									callableName: callableName{
										name: "next",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Matches with capturing groups.
			Expr:  "a(.?)",
			Input: "abracadabra",
			Results: map[string]interface{}{
				"match": "ab",
				"start": 0,
				"end":   2,
				"groups": []string{
					"b",
				},
				"next": &matchCallable{
					callableName: callableName{
						sync.Mutex{},
						"next",
					},
					match: "ac",
					start: 3,
					end:   5,
					groups: []string{
						"c",
					},
					next: &matchCallable{
						callableName: callableName{
							sync.Mutex{},
							"next",
						},
						match: "ad",
						start: 5,
						end:   7,
						groups: []string{
							"d",
						},
						next: &matchCallable{
							callableName: callableName{
								sync.Mutex{},
								"next",
							},
							match: "ab",
							start: 7,
							end:   9,
							groups: []string{
								"b",
							},
							next: &matchCallable{
								callableName: callableName{
									sync.Mutex{},
									"next",
								},
								match: "a",
								start: 10,
								end:   11,
								groups: []string{
									"",
								},
								next: &undefinedCallable{
									callableName: callableName{
										name: "next",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Matches with capturing groups (some unmatched).
			// Note that capturing groups that don't match any text
			// are represented by empty strings (unlike jsonata-js
			// which uses undefined).
			Expr:  "(a.)|(a)",
			Input: "abracadabra",
			Results: map[string]interface{}{
				"match": "ab",
				"start": 0,
				"end":   2,
				"groups": []string{
					"ab",
					"", // undefined in jsonata-js
				},
				"next": &matchCallable{
					callableName: callableName{
						sync.Mutex{},
						"next",
					},
					match: "ac",
					start: 3,
					end:   5,
					groups: []string{
						"ac",
						"", // undefined in jsonata-js
					},
					next: &matchCallable{
						callableName: callableName{
							sync.Mutex{},
							"next",
						},
						match: "ad",
						start: 5,
						end:   7,
						groups: []string{
							"ad",
							"", // undefined in jsonata-js
						},
						next: &matchCallable{
							callableName: callableName{
								sync.Mutex{},
								"next",
							},
							match: "ab",
							start: 7,
							end:   9,
							groups: []string{
								"ab",
								"", // undefined in jsonata-js
							},
							next: &matchCallable{
								callableName: callableName{
									sync.Mutex{},
									"next",
								},
								match: "a",
								start: 10,
								end:   11,
								groups: []string{
									"", // undefined in jsonata-js
									"a",
								},
								next: &undefinedCallable{
									callableName: callableName{
										name: "next",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Match on a non-ASCII string.
			// Note that the start and end values are byte offsets.
			// This means that a) they won't necessarily match the
			// jsonata-js offsets (e.g. smiley face emoji are only
			// 2 bytes long in JavaScript) and b) they won't play
			// well with JSONata functions that use rune offsets
			// such as $substring.
			Expr:  "😀",
			Input: "😂😁😀",
			Results: map[string]interface{}{
				"match":  "😀",
				"start":  8,  // 4 in jsonata-js
				"end":    12, // 6 in jsonata-js
				"groups": []string{},
				"next": &undefinedCallable{
					callableName: callableName{
						name: "next",
					},
				},
			},
		},
	})
}

func testRegexCallable(t *testing.T, tests []regexCallableTest) {

	for _, test := range tests {

		var argv []reflect.Value
		if test.Input != nil {
			argv = append(argv, reflect.ValueOf(test.Input))
		}

		re := regexp.MustCompile(test.Expr)
		v, err := newRegexCallable(re).Call(argv)
		if err != nil {
			t.Errorf("%s (%q): %s", test.Expr, test.Input, err)
		}

		if test.Undefined {
			if v != undefined {
				t.Errorf("%s: expected undefined result, got %v", test.Expr, v)
			}
			continue
		}

		var results interface{}
		if v.IsValid() && v.CanInterface() {
			results = v.Interface()
		}

		if !reflect.DeepEqual(results, test.Results) {
			t.Errorf("%s: expected results %v, got %v", test.Expr, test.Results, results)
		}
	}
}

func TestCallableParamCount(t *testing.T) {

	typeInt := reflect.TypeOf((*int)(nil)).Elem()

	tests := []struct {
		Callable jtypes.Callable
		Count    int
	}{
		{
			// goCallable, 0 parameters.
			Callable: &goCallable{
				callableName: callableName{
					name: "goCallable0",
				},
			},
			Count: 0,
		},
		{
			// goCallable, 1 parameter.
			Callable: &goCallable{
				callableName: callableName{
					name: "goCallable1",
				},
				params: []goCallableParam{
					{
						t: typeInt,
					},
				},
			},
			Count: 1,
		},
		{
			// lambdaCallable, 0 parameters.
			Callable: &lambdaCallable{
				callableName: callableName{
					name: "lambdaCallable0",
				},
				body: &jparse.NumberNode{},
			},
			Count: 0,
		},
		{
			// lambdaCallable, 2 parameters.
			Callable: &lambdaCallable{
				callableName: callableName{
					name: "lambdaCallable2",
				},
				body: &jparse.NumberNode{},
				paramNames: []string{
					"x",
					"y",
				},
			},
			Count: 2,
		},
		{
			// partialCallable, 0 parameters.
			Callable: &partialCallable{
				callableName: callableName{
					name: "partialCallable0",
				},
				fn: &undefinedCallable{},
			},
			Count: 0,
		},
		{
			// partialCallable, 1 (placeholder) parameter.
			Callable: &partialCallable{
				callableName: callableName{
					name: "partialCallable1",
				},
				fn: &undefinedCallable{},
				args: []jparse.Node{
					&jparse.StringNode{},
					&jparse.PlaceholderNode{},
					&jparse.NumberNode{},
				},
			},
			Count: 1,
		},
		{
			// All transformationCallables take 1 parameter.
			Callable: &transformationCallable{
				callableName: callableName{
					name: "transformationCallable",
				},
				pattern: &jparse.VariableNode{},
				updates: &jparse.ObjectNode{},
			},
			Count: 1,
		},
		{
			// All regexCallables take 1 parameter.
			Callable: &regexCallable{
				callableName: callableName{
					name: "regexCallable",
				},
				re: regexp.MustCompile("ab"),
			},
			Count: 1,
		},
		{
			// All matchCallables take 0 parameters.
			Callable: &matchCallable{
				callableName: callableName{
					name: "matchCallable",
				},
				match: "ab",
				start: 0,
				end:   2,
				next: &undefinedCallable{
					callableName: callableName{
						name: "next",
					},
				},
			},
			Count: 0,
		},
		{
			// All undefinedCallables take 0 parameters.
			Callable: &undefinedCallable{
				callableName: callableName{
					name: "undefinedCallable",
				},
			},
			Count: 0,
		},
		{
			// All chainCallables take 1 parameter.
			Callable: &chainCallable{
				callableName: callableName{
					name: "chainCallable",
				},
				callables: []jtypes.Callable{
					&undefinedCallable{},
					&undefinedCallable{},
				},
			},
			Count: 1,
		},
	}

	for _, test := range tests {
		if count := test.Callable.ParamCount(); count != test.Count {
			t.Errorf("%s: expected ParamCount %d, got %d", test.Callable.Name(), test.Count, count)
		}
	}
}
