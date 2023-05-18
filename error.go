// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/xiatechs/jsonata-go/jtypes"
)

// ErrUndefined is returned by the evaluation methods when
// a JSONata expression yields no results. Unlike most errors,
// ErrUndefined does not mean that evaluation failed.
//
// The simplest way to trigger ErrUndefined is to look up a
// field that is not present in the JSON data. Many JSONata
// operators and functions also return ErrUndefined when
// called with undefined inputs.
var ErrUndefined = errors.New("no results found")

// ErrType indicates the reason for an error.
type ErrType uint

// Types of errors that may be encountered by JSONata.
const (
	ErrNonIntegerLHS ErrType = iota
	ErrNonIntegerRHS
	ErrNonNumberLHS
	ErrNonNumberRHS
	ErrNonComparableLHS
	ErrNonComparableRHS
	ErrTypeMismatch
	ErrNonCallable
	ErrNonCallableApply
	ErrNonCallablePartial
	ErrNumberInf
	ErrNumberNaN
	ErrMaxRangeItems
	ErrIllegalKey
	ErrDuplicateKey
	ErrClone
	ErrIllegalUpdate
	ErrIllegalDelete
	ErrNonSortable
	ErrSortMismatch
)

var errmsgs = map[ErrType]string{
	ErrNonIntegerLHS:      `left side of the "{{value}}" operator must evaluate to an integer, position:{{position}}, arguments: {{arguments}}`,
	ErrNonIntegerRHS:      `right side of the "{{value}}" operator must evaluate to an integer, position:{{position}}, arguments: {{arguments}}`,
	ErrNonNumberLHS:       `left side of the "{{value}}" operator must evaluate to a number, position:{{position}}, arguments: {{arguments}}`,
	ErrNonNumberRHS:       `right side of the "{{value}}" operator must evaluate to a number, position:{{position}}, arguments: {{arguments}}`,
	ErrNonComparableLHS:   `left side of the "{{value}}" operator must evaluate to a number or string, position:{{position}}, arguments: {{arguments}}`,
	ErrNonComparableRHS:   `right side of the "{{value}}" operator must evaluate to a number or string, position:{{position}}, arguments: {{arguments}}`,
	ErrTypeMismatch:       `both sides of the "{{value}}" operator must have the same type, position:{{position}}, arguments: {{arguments}}`,
	ErrNonCallable:        `cannot call non-function {{token}}, position:{{position}}, arguments: {{arguments}}`,
	ErrNonCallableApply:   `cannot use function application with non-function {{token}}, position:{{position}}, arguments: {{arguments}}`,
	ErrNonCallablePartial: `cannot partially apply non-function {{token}}, position:{{position}}, arguments: {{arguments}}`,
	ErrNumberInf:          `result of the "{{value}}" operator is out of range, position:{{position}}, arguments: {{arguments}}`,
	ErrNumberNaN:          `result of the "{{value}}" operator is not a valid number, position:{{position}}, arguments: {{arguments}}`,
	ErrMaxRangeItems:      `range operator has too many items, position:{{position}}, arguments: {{arguments}}`,
	ErrIllegalKey:         `object key {{token}} does not evaluate to a string, position:{{position}}, arguments: {{arguments}}`,
	ErrDuplicateKey:       `multiple object keys evaluate to the value "{{value}}", position:{{position}}, arguments: {{arguments}}`,
	ErrClone:              `object transformation: cannot make a copy of the object, position:{{position}}, arguments: {{arguments}}`,
	ErrIllegalUpdate:      `the insert/update clause of an object transformation must evaluate to an object, position:{{position}}, arguments: {{arguments}}`,
	ErrIllegalDelete:      `the delete clause of an object transformation must evaluate to an array of strings, position:{{position}}, arguments: {{arguments}}`,
	ErrNonSortable:        `expressions in a sort term must evaluate to strings or numbers, position:{{position}}, arguments: {{arguments}}`,
	ErrSortMismatch:       `expressions in a sort term must have the same type, position:{{position}}, arguments: {{arguments}}`,
}

var reErrMsg = regexp.MustCompile("{{(token|value|position|arguments)}}")

// An EvalError represents an error during evaluation of a
// JSONata expression.
type EvalError struct {
	Type      ErrType
	Token     string
	Value     string
	Pos       int
	Arguments string
}

func newEvalError(typ ErrType, token interface{}, value interface{}, pos int) *EvalError {

	stringify := func(v interface{}) string {
		switch v := v.(type) {
		case string:
			return v
		case fmt.Stringer:
			return v.String()
		default:
			return ""
		}
	}

	return &EvalError{
		Type:  typ,
		Token: stringify(token),
		Value: stringify(value),
		Pos:   pos,
	}
}

func (e EvalError) Error() string {

	s := errmsgs[e.Type]
	if s == "" {
		return fmt.Sprintf("EvalError: unknown error type %d", e.Type)
	}

	return reErrMsg.ReplaceAllStringFunc(s, func(match string) string {
		switch match {
		case "{{token}}":
			return e.Token
		case "{{value}}":
			return e.Value
		case "{{arguments}}":
			return e.Arguments
		case "{{position}}":
			return fmt.Sprintf("%v", e.Pos)
		default:
			return match
		}
	})
}

// ArgCountError is returned by the evaluation methods when an
// expression contains a function call with the wrong number of
// arguments.
type ArgCountError struct {
	Func     string
	Expected int
	Received int
}

func newArgCountError(f jtypes.Callable, received int) *ArgCountError {
	return &ArgCountError{
		Func:     f.Name(),
		Expected: f.ParamCount(),
		Received: received,
	}
}

func (e ArgCountError) Error() string {
	return fmt.Sprintf("function %q takes %d argument(s), got %d", e.Func, e.Expected, e.Received)
}

// ArgTypeError is returned by the evaluation methods when an
// expression contains a function call with the wrong argument
// type.
type ArgTypeError struct {
	Func      string
	Which     int
	Pos       int
	Arguments string
}

func newArgTypeError(f jtypes.Callable, which int) *ArgTypeError {
	return &ArgTypeError{
		Func:  f.Name(),
		Which: which,
	}
}

func (e ArgTypeError) Error() string {
	return fmt.Sprintf("argument %d of function %q does not match function signature, position: %v, arguments: %v", e.Which, e.Func, e.Pos, e.Arguments)
}
