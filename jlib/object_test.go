// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xiatechs/jsonata-go/jlib"
	"github.com/xiatechs/jsonata-go/jtypes"
)

type eachTest struct {
	Input    interface{}
	Callable jtypes.Callable
	Output   interface{}
	Error    error
}

func TestEach(t *testing.T) {

	// repeatDots is a Callable that takes a number (N)
	// and returns a string of N dots.
	repeatDots := callable1(func(argv []reflect.Value) (reflect.Value, error) {
		n, _ := jtypes.AsNumber(argv[0])

		res := strings.Repeat(".", int(n))
		return reflect.ValueOf(res), nil
	})

	// repeatString is a Callable that takes a number (N)
	// and a string and returns the string repeated N times.
	repeatString := callable2(func(argv []reflect.Value) (reflect.Value, error) {
		n, _ := jtypes.AsNumber(argv[0])
		s, _ := jtypes.AsString(argv[1])

		res := strings.Repeat(s, int(n))
		return reflect.ValueOf(res), nil
	})

	// printLen is a Callable that takes a number, a string,
	// and an object and returns the object length as a string.
	// Note that the object length includes unexported struct
	// fields.
	printLen := callable3(func(argv []reflect.Value) (reflect.Value, error) {
		var len int
		switch argv[2].Kind() {
		case reflect.Map:
			len = argv[2].Len()
		case reflect.Struct:
			len = argv[2].NumField()
		}
		res := strconv.Itoa(len)
		return reflect.ValueOf(res), nil
	})

	testEach(t, []eachTest{
		{
			// Empty map.
			Input:    map[string]interface{}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Empty struct.
			Input:    struct{}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Struct with no exported fields.
			Input: struct {
				a, b, c int
			}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Callable with 1 argument.
			Input: map[string]interface{}{
				"a": 5,
			},
			Callable: repeatDots,
			Output:   ".....",
		},
		{
			Input: struct {
				A int
			}{
				A: 5,
			},
			Callable: repeatDots,
			Output:   ".....",
		},
		{
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
				"e": 5,
			},
			Callable: repeatDots,
			Output: []interface{}{
				".",
				"..",
				"...",
				"....",
				".....",
			},
		},
		{
			Input: struct {
				A, B, C, D, e int
			}{
				A: 1,
				B: 2,
				C: 3,
				D: 4,
				e: 5, // unexported struct fields are ignored.
			},
			Callable: repeatDots,
			Output: []interface{}{
				".",
				"..",
				"...",
				"....",
			},
		},
		{
			// Callable with 2 arguments.
			Input: map[string]interface{}{
				"a": 5,
			},
			Callable: repeatString,
			Output:   "aaaaa",
		},
		{
			Input: struct {
				A int
			}{
				A: 5,
			},
			Callable: repeatString,
			Output:   "AAAAA",
		},
		{
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
				"e": 5,
			},
			Callable: repeatString,
			Output: []interface{}{
				"a",
				"bb",
				"ccc",
				"dddd",
				"eeeee",
			},
		},
		{
			Input: struct {
				A, B, C, D, e int
			}{
				A: 1,
				B: 2,
				C: 3,
				D: 4,
				e: 5, // unexported struct fields are ignored.
			},
			Callable: repeatString,
			Output: []interface{}{
				"A",
				"BB",
				"CCC",
				"DDDD",
			},
		},
		{
			// Callable with 3 arguments.
			Input: map[string]interface{}{
				"a": 5,
			},
			Callable: printLen,
			Output:   "1",
		},
		{
			Input: struct {
				A int
			}{
				A: 5,
			},
			Callable: printLen,
			Output:   "1",
		},
		{
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
				"e": 5,
			},
			Callable: printLen,
			Output: []interface{}{
				"5",
				"5",
				"5",
				"5",
				"5",
			},
		},
		{
			Input: struct {
				A, B, C, D, e int
			}{
				A: 1,
				B: 2,
				C: 3,
				D: 4,
				e: 5, // unexported struct fields are ignored.
			},
			Callable: printLen,
			Output: []interface{}{
				"5",
				"5",
				"5",
				"5",
			},
		},
		{
			// Invalid input. Return an error.
			// Note that we don't even get as far as validating the
			// Callable in this case.
			Input: "hello",
			Error: fmt.Errorf("argument must be an object"),
		},
		{
			// Callable has too few parameters.
			Input:    map[string]interface{}{},
			Callable: paramCountCallable(0),
			Error:    fmt.Errorf("function must take 1, 2 or 3 arguments"),
		},
		{
			// Callable has too many parameters.
			Input:    struct{}{},
			Callable: paramCountCallable(4),
			Error:    fmt.Errorf("function must take 1, 2 or 3 arguments"),
		},
		{
			// If the Callable returns an error, return the error.
			Input: map[string]interface{}{
				"a": 1,
			},
			Callable: paramCountCallable(1),
			Error:    errTest,
		},
		{
			// If the Callable returns an error, return the error.
			Input: struct {
				A int
			}{},
			Callable: paramCountCallable(1),
			Error:    errTest,
		},
	})
}

func testEach(t *testing.T, tests []eachTest) {

	for i, test := range tests {

		output, err := jlib.Each(reflect.ValueOf(test.Input), test.Callable)

		if !equalStringArray(output, test.Output) {
			t.Errorf("Test %d: expected %v, got %v", i+1, test.Output, output)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("Test %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

type siftTest struct {
	Input    interface{}
	Callable jtypes.Callable
	Output   interface{}
	Error    error
}

func TestSift(t *testing.T) {

	// valueIsOdd is a Callable that takes one argument and
	// returns true if it is an odd number.
	valueIsOdd := callable1(func(argv []reflect.Value) (reflect.Value, error) {
		n, ok := jtypes.AsNumber(argv[0])
		res := ok && int(n)&1 == 1
		return reflect.ValueOf(res), nil
	})

	// valuesAreEqual is a Callable that takes two arguments
	// and returns true if they are equal.
	valuesAreEqual := callable2(func(argv []reflect.Value) (reflect.Value, error) {
		res := reflect.DeepEqual(argv[0].Interface(), argv[1].Interface())
		return reflect.ValueOf(res), nil
	})

	// valueIsLen is a Callable that takes three arguments
	// and returns true if the first argument is equal to
	// the number of elements in the third argument.
	valueIsLen := callable3(func(argv []reflect.Value) (reflect.Value, error) {
		var len int
		switch argv[2].Kind() {
		case reflect.Map:
			len = argv[2].Len()
		case reflect.Struct:
			len = argv[2].NumField()
		}
		n, ok := jtypes.AsNumber(argv[0])
		res := ok && int(n) == len
		return reflect.ValueOf(res), nil
	})

	testSift(t, []siftTest{
		{
			// Empty map.
			Input:    map[string]interface{}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Empty struct.
			Input:    struct{}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Struct with no exported fields.
			Input: struct {
				a, b, c int
			}{},
			Callable: paramCountCallable(1),
			Error:    jtypes.ErrUndefined,
		},
		{
			// Callable with 1 argument.
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
			},
			Callable: valueIsOdd,
			Output: map[string]interface{}{
				"a": 1,
				"c": 3,
			},
		},
		{
			Input: struct {
				A, B, C, d int
			}{
				A: 5,
				B: 0,
				C: 4,
				d: 1, // unexported struct fields are ignored.
			},
			Callable: valueIsOdd,
			Output: map[string]interface{}{
				"A": 5,
			},
		},
		{
			// Callable with 2 arguments.
			Input: map[string]interface{}{
				"a": 1,
				"b": "b",
				"c": true,
			},
			Callable: valuesAreEqual,
			Output: map[string]interface{}{
				"b": "b",
			},
		},
		{
			Input: struct {
				A int
				B string
				c string // unexported struct fields are ignored.
			}{
				A: 1,
				B: "B",
				c: "c",
			},
			Callable: valuesAreEqual,
			Output: map[string]interface{}{
				"B": "B",
			},
		},
		{
			// Callable with 3 arguments.
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			Callable: valueIsLen,
			Output: map[string]interface{}{
				"c": 3,
			},
		},
		{
			Input: struct {
				A, B, C, d int
			}{
				A: 4,
				B: 2,
				C: 3,
				d: 4, // unexported struct fields are ignored.
			},
			Callable: valueIsLen,
			Output: map[string]interface{}{
				"A": 4,
			},
		},
		{
			// Invalid input. Return an error.
			// Note that we don't even get as far as validating the
			// Callable in this case.
			Input: 3.141592,
			Error: fmt.Errorf("argument must be an object"),
		},
		{
			// Invalid key type.
			Input: map[bool]string{
				true: "true",
			},
			Callable: paramCountCallable(1),
			Error:    fmt.Errorf("object key must evaluate to a string, got true (bool)"),
		},
		{
			// Callable has too few parameters.
			Input:    map[string]interface{}{},
			Callable: paramCountCallable(0),
			Error:    fmt.Errorf("function must take 1, 2 or 3 arguments"),
		},
		{
			// Callable has too many parameters.
			Input:    struct{}{},
			Callable: paramCountCallable(4),
			Error:    fmt.Errorf("function must take 1, 2 or 3 arguments"),
		},
		{
			// If the Callable returns an error, return the error.
			Input: map[string]interface{}{
				"a": 1,
			},
			Callable: paramCountCallable(1),
			Error:    errTest,
		},
		{
			// If the Callable returns an error, return the error.
			Input: struct {
				A int
			}{},
			Callable: paramCountCallable(1),
			Error:    errTest,
		},
	})
}

func testSift(t *testing.T, tests []siftTest) {

	for i, test := range tests {

		output, err := jlib.Sift(reflect.ValueOf(test.Input), test.Callable)

		if !reflect.DeepEqual(output, test.Output) {
			t.Errorf("Test %d: expected %v, got %v", i+1, test.Output, output)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("Test %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

type keysTest struct {
	Input  interface{}
	Output interface{}
	Error  error
}

func TestKeys(t *testing.T) {
	testKeys(t, []keysTest{
		{
			// Empty map.
			Input: map[string]interface{}{},
			Error: jtypes.ErrUndefined,
		},
		{
			// Empty struct.
			Input: struct{}{},
			Error: jtypes.ErrUndefined,
		},
		{
			// Struct with no exported fields.
			Input: struct {
				a, b, c int
			}{},
			Error: jtypes.ErrUndefined,
		},
		{
			// Empty array.
			Input: []interface{}{},
			Error: jtypes.ErrUndefined,
		},
		{
			Input: map[string]interface{}{
				"a": 1,
			},
			Output: "a",
		},
		{
			Input: map[string]bool{ // non-shortcut version
				"a": true,
			},
			Output: "a",
		},
		{
			Input: struct {
				A int
			}{},
			Output: "A",
		},
		{
			Input: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			Output: []string{
				"a",
				"b",
				"c",
			},
		},
		{
			Input: map[string]float64{ // non-shortcut version
				"a": 1,
				"b": 2,
				"c": 3,
			},
			Output: []string{
				"a",
				"b",
				"c",
			},
		},
		{
			Input: struct {
				A int
				B string
				C bool
				d float64 // unexported struct fields are ignored.
			}{},
			Output: []string{
				"A",
				"B",
				"C",
			},
		},
		{
			Input: []interface{}{
				map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				struct {
					A, B, C, d int // unexported struct fields are ignored.
				}{},
			},
			Output: []string{
				"a",
				"b",
				"c",
				"A",
				"B",
				"C",
			},
		},
		{
			Input: "this isn't an object",
			Error: jtypes.ErrUndefined,
		},
		{
			Input: []interface{}{
				3.141592,
			},
			Error: jtypes.ErrUndefined,
		},
		{
			Input: map[bool]string{
				true: "true",
			},
			Error: fmt.Errorf("object key must evaluate to a string, got true (bool)"),
		},
		{
			Input: []interface{}{
				map[bool]string{
					false: "false",
				},
			},
			Error: fmt.Errorf("object key must evaluate to a string, got false (bool)"),
		},
	})
}

func testKeys(t *testing.T, tests []keysTest) {

	for i, test := range tests {

		output, err := jlib.Keys(reflect.ValueOf(test.Input))

		if !equalStringArray(output, test.Output) {
			t.Errorf("Test %d: expected %v, got %v", i+1, test.Output, output)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("Test %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

type mergeTest struct {
	Input  interface{}
	Output interface{}
	Error  error
}

func TestMerge(t *testing.T) {
	testMerge(t, []mergeTest{
		{
			// Empty map.
			Input:  map[string]interface{}{},
			Output: map[string]interface{}{},
		},
		{
			// Empty struct.
			Input:  struct{}{},
			Output: map[string]interface{}{},
		},
		{
			// Struct with no exported fields.
			Input: struct {
				a, b, c int
			}{},
			Output: map[string]interface{}{},
		},
		{
			// Empty array.
			Input:  []interface{}{},
			Output: map[string]interface{}{},
		},
		{
			Input: map[string]interface{}{
				"a": 1,
			},
			Output: map[string]interface{}{
				"a": 1,
			},
		},
		{
			Input: struct {
				Pi float64
			}{
				Pi: 3.141592,
			},
			Output: map[string]interface{}{
				"Pi": 3.141592,
			},
		},
		{
			Input: []interface{}{
				map[string]int{
					"a": 1,
					"c": 3,
					"e": 5,
					"g": 7,
				},
				map[string]int{
					"b": 2,
					"d": 4,
					"f": 6,
					"h": 8,
				},
			},
			Output: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
				"e": 5,
				"f": 6,
				"g": 7,
				"h": 8,
			},
		},
		{
			Input: []struct {
				A, B, C, d int // unexported struct fields are ignored.
			}{
				{
					A: 1,
				},
				{
					B: 2,
				},
				{
					C: 3,
				},
			},
			Output: map[string]interface{}{
				"A": 0,
				"B": 0,
				"C": 3,
			},
		},
		{
			Input: []interface{}{
				map[string]interface{}{
					"One": 1,
					"Two": 2,
				},
				map[string]interface{}{
					"Three": 3,
				},
				struct {
					Three string
					Four  string
				}{
					Three: "three",
					Four:  "four",
				},
				map[string]float64{
					"Four": 4.0,
					"Five": 5.0,
				},
			},
			Output: map[string]interface{}{
				"One":   1,
				"Two":   2,
				"Three": "three",
				"Four":  4.0,
				"Five":  5.0,
			},
		},
		{
			Input: "this isn't an object",
			Error: fmt.Errorf("argument must be an object or an array of objects"),
		},
		{
			Input: []interface{}{
				3.141592,
			},
			Error: fmt.Errorf("argument must be an object or an array of objects"),
		},
		{
			Input: map[bool]string{
				true: "true",
			},
			Error: fmt.Errorf("object key must evaluate to a string, got true (bool)"),
		},
		{
			Input: []interface{}{
				map[bool]string{
					false: "false",
				},
			},
			Error: fmt.Errorf("object key must evaluate to a string, got false (bool)"),
		},
	})
}

func testMerge(t *testing.T, tests []mergeTest) {

	for i, test := range tests {

		output, err := jlib.Merge(reflect.ValueOf(test.Input))

		if !reflect.DeepEqual(output, test.Output) {
			t.Errorf("Test %d: expected %v, got %v", i+1, test.Output, output)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("Test %d: expected error %v, got %v", i+1, test.Error, err)
		}
	}
}

var errTest = errors.New("paramCountCallable.Call not implemented")

type paramCountCallable int

func (f paramCountCallable) Name() string {
	return "paramCountCallable"
}

func (f paramCountCallable) ParamCount() int {
	return int(f)
}

func (f paramCountCallable) Call([]reflect.Value) (reflect.Value, error) {
	return reflect.Value{}, errTest
}

type callable1 func([]reflect.Value) (reflect.Value, error)

func (f callable1) Name() string    { return "callable1" }
func (f callable1) ParamCount() int { return 1 }

func (f callable1) Call(argv []reflect.Value) (reflect.Value, error) {
	if len(argv) != 1 {
		return reflect.Value{}, fmt.Errorf("callable1.Call expected 1 argument, got %d", len(argv))
	}
	return f(argv)
}

type callable2 func([]reflect.Value) (reflect.Value, error)

func (f callable2) Name() string    { return "callable2" }
func (f callable2) ParamCount() int { return 2 }

func (f callable2) Call(argv []reflect.Value) (reflect.Value, error) {
	if len(argv) != 2 {
		return reflect.Value{}, fmt.Errorf("callable2.Call expected 2 arguments, got %d", len(argv))
	}
	return f(argv)
}

type callable3 func([]reflect.Value) (reflect.Value, error)

func (f callable3) Name() string    { return "callable3" }
func (f callable3) ParamCount() int { return 3 }

func (f callable3) Call(argv []reflect.Value) (reflect.Value, error) {
	if len(argv) != 3 {
		return reflect.Value{}, fmt.Errorf("callable3.Call expected 3 arguments, got %d", len(argv))
	}
	return f(argv)
}

func equalStringArray(v1, v2 interface{}) bool {
	switch v1 := v1.(type) {
	case nil:
		return v2 == nil
	case string:
		v2, ok := v2.(string)
		return ok && v2 == v1
	case []string:
		v2, ok := v2.([]string)
		return ok && reflect.DeepEqual(stringArrayCount(v1), stringArrayCount(v2))
	case []interface{}:
		v2, ok := v2.([]interface{})
		return ok && reflect.DeepEqual(interfaceArrayCount(v1), interfaceArrayCount(v2))
	default:
		return false
	}
}

func stringArrayCount(values []string) map[string]int {
	var m map[string]int
	for _, s := range values {
		if m == nil {
			m = make(map[string]int)
		}
		m[s]++
	}
	return m
}

func interfaceArrayCount(values []interface{}) map[interface{}]int {
	var m map[interface{}]int
	for _, v := range values {
		if m == nil {
			m = make(map[interface{}]int)
		}
		m[v]++
	}
	return m
}
