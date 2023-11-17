// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib_test

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xiatechs/jsonata-go/jlib"
	"github.com/xiatechs/jsonata-go/jtypes"
)

var typereplaceCallable = reflect.TypeOf((*replaceCallable)(nil)).Elem()
var typeMatchCallable = reflect.TypeOf((*matchCallable)(nil)).Elem()
var typeCallable = reflect.TypeOf((*jtypes.Callable)(nil)).Elem()

func TestMain(m *testing.M) {

	if !typereplaceCallable.Implements(typeCallable) {
		fmt.Fprintln(os.Stderr, "replaceCallable is not a Callable. Aborting...")
		os.Exit(1)
	}

	if !reflect.PtrTo(typeMatchCallable).Implements(typeCallable) {
		fmt.Fprintln(os.Stderr, "*matchCallable is not a Callable. Aborting...")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestString(t *testing.T) {

	data := []struct {
		Input  interface{}
		Output string
		Error  error
	}{
		{
			Input:  "string",
			Output: "string",
		},
		{
			Input:  []byte("string"),
			Output: "string",
		},
		{
			Input:  100,
			Output: "100",
		},
		{
			Input:  3.14159265359,
			Output: "3.14159265359",
		},
		{
			Input:  true,
			Output: "true",
		},
		{
			Input:  false,
			Output: "false",
		},
		{
			Input:  nil,
			Output: "null",
		},
		{
			Input:  []interface{}{},
			Output: "[]",
		},
		{
			Input: []interface{}{
				"hello",
				100,
				3.14159265359,
				false,
			},
			Output: `["hello",100,3.14159265359,false]`,
		},
		{
			Input:  map[string]interface{}{},
			Output: "{}",
		},
		{
			Input: map[string]interface{}{
				"hello":       "world",
				"one hundred": 100,
				"pi":          3.14159265359,
				"bool":        true,
				"null":        nil,
			},
			Output: `{"bool":true,"hello":"world","null":null,"one hundred":100,"pi":3.14159265359}`,
		},
		{
			Input:  replaceCallable(nil),
			Output: "",
		},
		{
			Input:  &matchCallable{},
			Output: "",
		},
		{
			Input: math.Inf(0),
			Error: &jlib.Error{
				Func: "string",
				Type: jlib.ErrNaNInf,
			},
		},
		{
			Input: math.NaN(),
			Error: &jlib.Error{
				Func: "string",
				Type: jlib.ErrNaNInf,
			},
		},
	}

	for _, test := range data {

		got, err := jlib.String(test.Input)

		if got != test.Output {
			t.Errorf("%v: Expected %q, got %q", test.Input, test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%v: Expected error %v, got %v", test.Input, test.Error, err)
		}
	}
}

func TestSubstring(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Start  int
		Length jtypes.OptionalInt
		Output string
	}{
		{
			Start:  2,
			Output: "emoji",
		},
		{
			// Start position greater than string length.
			Start:  7,
			Output: "",
		},
		{
			// Negative start position.
			Start:  -5,
			Output: "emoji",
		},
		{
			// Negative start position beyond start of string.
			Start:  -20,
			Output: "😂 emoji",
		},
		{
			Start:  0,
			Length: jtypes.NewOptionalInt(1),
			Output: "😂",
		},
		{
			Start:  2,
			Length: jtypes.NewOptionalInt(3),
			Output: "emo",
		},
		{
			// Length greater than string length.
			Start:  2,
			Length: jtypes.NewOptionalInt(30),
			Output: "emoji",
		},
		{
			// Zero length.
			Start:  2,
			Length: jtypes.NewOptionalInt(0),
			Output: "",
		},
		{
			// Negative length.
			Start:  2,
			Length: jtypes.NewOptionalInt(-5),
			Output: "",
		},
	}

	for _, test := range data {

		got := jlib.Substring(src, test.Start, test.Length)

		if got != test.Output {

			s := fmt.Sprintf("substring(%q, %d", src, test.Start)
			if test.Length.IsSet() {
				s += fmt.Sprintf(", %d", test.Length.Int)
			}
			s += ")"

			t.Errorf("%s: Expected %q, got %q", s, test.Output, got)
		}
	}
}

func TestSubstringBefore(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Substr string
		Output string
	}{
		{
			Substr: "ji",
			Output: "😂 emo",
		},
		{
			Substr: " ",
			Output: "😂",
		},
		{
			Substr: "😂",
			Output: "",
		},
		{
			// The index of an empty substr is zero, so substringBefore
			// returns an empty string.
			Substr: "",
			Output: "",
		},
		{
			// If substr is not present, return the full string.
			Substr: "x",
			Output: "😂 emoji",
		},
	}

	for _, test := range data {

		got := jlib.SubstringBefore(src, test.Substr)

		if got != test.Output {
			t.Errorf("substringBefore(%q, %q): Expected %q, got %q", src, test.Substr, test.Output, got)
		}
	}
}

func TestSubstringAfter(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Substr string
		Output string
	}{
		{
			Substr: "emo",
			Output: "ji",
		},
		{
			Substr: " ",
			Output: "emoji",
		},
		{
			Substr: "😂",
			Output: " emoji",
		},
		{
			// The index of an empty substr is zero, so substringAfter
			// returns the full string.
			Substr: "",
			Output: "😂 emoji",
		},
		{
			// If substr is not present, return the full string.
			Substr: "x",
			Output: "😂 emoji",
		},
	}

	for _, test := range data {

		got := jlib.SubstringAfter(src, test.Substr)

		if got != test.Output {
			t.Errorf("substringAfter(%q, %q): Expected %q, got %q", src, test.Substr, test.Output, got)
		}
	}
}

func TestPad(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Width  int
		Chars  jtypes.OptionalString
		Output string
	}{
		{
			Width:  10,
			Output: "😂 emoji   ",
		},
		{
			Width:  -10,
			Output: "   😂 emoji",
		},
		{
			// Pad with custom character.
			Width:  10,
			Chars:  jtypes.NewOptionalString("😂"),
			Output: "😂 emoji😂😂😂",
		},
		{
			// Pad with custom character.
			Width:  -10,
			Chars:  jtypes.NewOptionalString("😂"),
			Output: "😂😂😂😂 emoji",
		},
		{
			// Pad with multiple characters.
			Width:  12,
			Chars:  jtypes.NewOptionalString("123"),
			Output: "😂 emoji12312",
		},
		{
			// Pad with multiple characters.
			Width:  -12,
			Chars:  jtypes.NewOptionalString("123"),
			Output: "12312😂 emoji",
		},
		{
			// Width less than length of string.
			Width:  5,
			Output: "😂 emoji",
		},
		{
			// Width less than length of string.
			Width:  -5,
			Output: "😂 emoji",
		},
	}

	for _, test := range data {

		got := jlib.Pad(src, test.Width, test.Chars)

		if got != test.Output {

			s := fmt.Sprintf("pad(%q, %d", src, test.Width)
			if test.Chars.IsSet() {
				s += fmt.Sprintf(", %q", test.Chars.String)
			}
			s += ")"

			t.Errorf("%s: Expected %q, got %q", s, test.Output, got)
		}
	}
}

func TestTrim(t *testing.T) {

	data := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "     hello    world  ",
			Output: "hello world",
		},
		{
			Input:  "hello\r\nworld",
			Output: "hello world",
		},
		{
			Input: `multiline
                        string
                            with
                                tabs`,
			Output: "multiline string with tabs",
		},
	}

	for _, test := range data {

		got := jlib.Trim(test.Input)

		if got != test.Output {
			t.Errorf("trim(%q): Expected %q, got %q", test.Input, test.Output, got)
		}
	}
}

func TestContains(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Pattern interface{} // pattern can be a string or a matching function
		Output  bool
		Error   error
	}{
		{
			Pattern: "moji",
			Output:  true,
		},
		{
			Pattern: "😂",
			Output:  true,
		},
		{
			Pattern: "muji",
			Output:  false,
		},
		{
			// Matches for regex "m.ji".
			Pattern: &matchCallable{
				name: "/m.ji/",
				matches: []match{
					{
						Match: "moji",
						Start: 6,
						End:   10,
					},
				},
			},
			Output: true,
		},
		{
			// Matches for regex "😂".
			Pattern: &matchCallable{
				name: "/😂/",
				matches: []match{
					{
						Match: "😂",
						Start: 0,
						End:   4,
					},
				},
			},
			Output: true,
		},
		{
			// No matches.
			Pattern: &matchCallable{
				name: "/^m.ji/",
			},
			Output: false,
		},
		{
			// Invalid pattern.
			Pattern: 100,
			Error:   fmt.Errorf("function contains takes a string or a regex"),
		},
	}

	for _, test := range data {

		pattern := newStringCallable(test.Pattern)
		got, err := jlib.Contains(src, pattern)

		if got != test.Output {
			t.Errorf("contains(%q, %s): Expected %t, got %t", src, formatStringCallable(pattern), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("contains(%q, %s): Expected error %v, got %v", src, formatStringCallable(pattern), test.Error, err)
		}
	}
}
func TestSplit(t *testing.T) {

	src := "😂 emoji"

	data := []struct {
		Separator interface{} // separator can be a string or a matching function
		Limit     jtypes.OptionalInt
		Output    []string
		Error     error
	}{
		{
			Separator: " ",
			Output: []string{
				"😂",
				"emoji",
			},
		},
		{
			Separator: "",
			Output: []string{
				"😂",
				" ",
				"e",
				"m",
				"o",
				"j",
				"i",
			},
		},
		{
			Separator: "",
			Limit:     jtypes.NewOptionalInt(3),
			Output: []string{
				"😂",
				" ",
				"e",
			},
		},
		{
			Separator: "",
			Limit:     jtypes.NewOptionalInt(0),
			Output:    []string{},
		},
		{
			Separator: "",
			Limit:     jtypes.NewOptionalInt(-1),
			Error:     fmt.Errorf("third argument of the split function must evaluate to a positive number"),
		},
		{
			Separator: "muji",
			Output: []string{
				"😂 emoji",
			},
		},
		{
			// Matches for regex "\\s".
			Separator: &matchCallable{
				name: "/\\s/",
				matches: []match{
					{
						Match: " ",
						Start: 4,
						End:   5,
					},
				},
			},
			Output: []string{
				"😂",
				"emoji",
			},
		},
		{
			// Matches for regex "[aeiou]".
			Separator: &matchCallable{
				name: "/[aeiou]/",
				matches: []match{
					{
						Match: "e",
						Start: 5,
						End:   6,
					},
					{
						Match: "o",
						Start: 7,
						End:   8,
					},
					{
						Match: "i",
						Start: 9,
						End:   10,
					},
				},
			},
			Output: []string{
				"😂 ",
				"m",
				"j",
				"",
			},
		},
		{
			// No match.
			Separator: &matchCallable{
				name: "/muji/",
			},
			Output: []string{
				"😂 emoji",
			},
		},
		{
			// Invalid separator.
			Separator: 100,
			Error:     fmt.Errorf("function split takes a string or a regex"),
		},
	}

	for _, test := range data {

		separator := newStringCallable(test.Separator)

		prefix := func() string {
			s := fmt.Sprintf("split(%q, %s", src, formatStringCallable(separator))
			if test.Limit.IsSet() {
				s += fmt.Sprintf(", %d", test.Limit.Int)
			}
			return s + ")"
		}

		got, err := jlib.Split(src, separator, test.Limit)

		if !reflect.DeepEqual(got, test.Output) {
			t.Errorf("%s: Expected %v, got %v", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

func TestJoin(t *testing.T) {

	data := []struct {
		Values    interface{}
		Separator jtypes.OptionalString
		Output    string
		Error     error
	}{
		{
			// Single values are returned unchanged.
			Values: "😂 emoji",
			Output: "😂 emoji",
		},
		{
			Values: []string{},
			Output: "",
		},
		{
			Values: []string{
				"😂",
				"emoji",
			},
			Output: "😂emoji",
		},
		{
			Values: []interface{}{
				"one",
				"two",
				"three",
				"four",
				"five",
			},
			Separator: jtypes.NewOptionalString("😂"),
			Output:    "one😂two😂three😂four😂five",
		},
		{
			Values: []interface{}{
				"one",
				"two",
				"three",
				"four",
				5,
			},
			Error: fmt.Errorf("function join takes an array of strings"),
		},
	}

	for _, test := range data {

		prefix := func() string {
			s := fmt.Sprintf("join(%v", test.Values)
			if test.Separator.IsSet() {
				s += fmt.Sprintf(", %q", test.Separator.String)
			}
			return s + ")"
		}

		got, err := jlib.Join(reflect.ValueOf(test.Values), test.Separator)

		if got != test.Output {
			t.Errorf("%s: Expected %q, got %q", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

func TestMatch(t *testing.T) {

	src := "abracadabra"

	data := []struct {
		Pattern jtypes.Callable
		Limit   jtypes.OptionalInt
		Output  []map[string]interface{}
		Error   error
	}{
		{
			// Matches for regex "a."
			Pattern: abracadabraMatches0(),
			Output: []map[string]interface{}{
				{
					"match":  "ab",
					"index":  0,
					"groups": []string{},
				},
				{
					"match":  "ac",
					"index":  3,
					"groups": []string{},
				},
				{
					"match":  "ad",
					"index":  5,
					"groups": []string{},
				},
				{
					"match":  "ab",
					"index":  7,
					"groups": []string{},
				},
			},
		},
		{
			// Matches for regex "a(.)"
			Pattern: abracadabraMatches1(),
			Output: []map[string]interface{}{
				{
					"match": "ab",
					"index": 0,
					"groups": []string{
						"b",
					},
				},
				{
					"match": "ac",
					"index": 3,
					"groups": []string{
						"c",
					},
				},
				{
					"match": "ad",
					"index": 5,
					"groups": []string{
						"d",
					},
				},
				{
					"match": "ab",
					"index": 7,
					"groups": []string{
						"b",
					},
				},
			},
		},
		{
			// Matches for regex "(a)(.)"
			Pattern: abracadabraMatches2(),
			Output: []map[string]interface{}{
				{
					"match": "ab",
					"index": 0,
					"groups": []string{
						"a",
						"b",
					},
				},
				{
					"match": "ac",
					"index": 3,
					"groups": []string{
						"a",
						"c",
					},
				},
				{
					"match": "ad",
					"index": 5,
					"groups": []string{
						"a",
						"d",
					},
				},
				{
					"match": "ab",
					"index": 7,
					"groups": []string{
						"a",
						"b",
					},
				},
			},
		},
		{
			Pattern: abracadabraMatches2(),
			Limit:   jtypes.NewOptionalInt(1),
			Output: []map[string]interface{}{
				{
					"match": "ab",
					"index": 0,
					"groups": []string{
						"a",
						"b",
					},
				},
			},
		},
		{
			Pattern: abracadabraMatches2(),
			Limit:   jtypes.NewOptionalInt(0),
			Output:  []map[string]interface{}{},
		},
		{
			Pattern: abracadabraMatches2(),
			Limit:   jtypes.NewOptionalInt(-1),
			Error:   fmt.Errorf("third argument of function match must evaluate to a positive number"),
		},
		{
			Pattern: &matchCallable{
				name: "/muji/",
			},
			Output: []map[string]interface{}{},
		},
	}

	for _, test := range data {

		prefix := func() string {
			s := fmt.Sprintf("match(%q, %s", src, test.Pattern.Name())
			if test.Limit.IsSet() {
				s += fmt.Sprintf(", %d", test.Limit.Int)
			}
			return s + ")"
		}

		got, err := jlib.Match(src, test.Pattern, test.Limit)

		if !reflect.DeepEqual(got, test.Output) {
			t.Errorf("%s: Expected %v, got %v", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

func TestReplace(t *testing.T) {

	src := "abracadabra"

	data := []struct {
		Pattern interface{} // pattern can be a string or a matching function
		Repl    interface{} // repl can be a string or a replacement function
		Limit   jtypes.OptionalInt
		Output  string
		Error   error
	}{

		// String patterns

		{
			Pattern: "a",
			Repl:    "å",
			Output:  "åbråcådåbrå",
		},
		{
			Pattern: "a",
			Repl:    "å",
			Limit:   jtypes.NewOptionalInt(3),
			Output:  "åbråcådabra",
		},
		{
			Pattern: "a",
			Repl:    "å",
			Limit:   jtypes.NewOptionalInt(0),
			Output:  "abracadabra",
		},
		{
			Pattern: "a",
			Repl:    "å",
			Limit:   jtypes.NewOptionalInt(-1),
			Error:   fmt.Errorf("fourth argument of function replace must evaluate to a positive number"),
		},
		{
			Pattern: "a",
			Repl:    "",
			Output:  "brcdbr",
		},
		{
			Pattern: "😂",
			Repl:    "",
			Output:  "abracadabra",
		},
		{
			Pattern: "",
			Repl:    "å",
			Limit:   jtypes.NewOptionalInt(0),
			Error:   fmt.Errorf("second argument of function replace can't be an empty string"),
		},
		{
			Pattern: "a",
			Repl:    replaceCallable(nil),
			Limit:   jtypes.NewOptionalInt(0),
			Error:   fmt.Errorf("third argument of function replace must be a string when pattern is a string"),
		},

		// Matching function patterns

		{
			// Matches for regex "a."
			Pattern: abracadabraMatches0(),
			Repl:    "åå",
			Output:  "åårååååååra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl:    "åå",
			Limit:   jtypes.NewOptionalInt(3),
			Output:  "åårååååabra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl:    "åå",
			Limit:   jtypes.NewOptionalInt(0),
			Output:  "abracadabra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl:    "åå",
			Limit:   jtypes.NewOptionalInt(-1),
			Error:   fmt.Errorf("fourth argument of function replace must evaluate to a positive number"),
		},
		{
			// $0 is replaced by the full matched string.
			Pattern: abracadabraMatches0(),
			Repl:    "$0",
			Output:  "abracadabra",
		},
		{
			// $N is replaced by the Nth captured string.
			// Matches for regex "a(.)"
			Pattern: abracadabraMatches1(),
			Repl:    "$1",
			Output:  "brcdbra",
		},
		{
			// Matches for regex "(a)(.)"
			Pattern: abracadabraMatches2(),
			Repl:    "$2$1",
			Output:  "barcadabara",
		},
		{
			// If N is greater than the number of captured strings,
			// $N evaluates to an empty string...
			Pattern: abracadabraMatches2(),
			Repl:    "$3",
			Output:  "rra",
		},
		{
			// ...unless N has more than one digit, in which case we
			// discard the rightmost digit and retry. Discarded digits
			// are copied to the output.
			Pattern: abracadabraMatches2(),
			Repl:    "$10$200",
			Output:  "a0b00ra0c00a0d00a0b00ra",
		},
		{
			// Trailing dollar signs are treated as literal dollar signs.
			Pattern: abracadabraMatches2(),
			Repl:    "$",
			Output:  "$r$$$ra",
		},
		{
			// Trailing dollar signs are treated as literal dollar signs.
			Pattern: abracadabraMatches2(),
			Repl:    "$1$2$",
			Output:  "ab$rac$ad$ab$ra",
		},
		{
			// Double dollar signs are treated as literal dollar signs.
			Pattern: abracadabraMatches2(),
			Repl:    "$$",
			Output:  "$r$$$ra",
		},
		{
			// Double dollar signs are treated as literal dollar signs.
			Pattern: abracadabraMatches2(),
			Repl:    "$1$$$2",
			Output:  "a$bra$ca$da$bra",
		},
		{
			// Dollar signs followed by anything other than another dollar
			// sign or a digit are treated as normal text.
			Pattern: abracadabraMatches2(),
			Repl:    "$ ",
			Output:  "$ r$ $ $ ra",
		},
		{
			// Dollar signs followed by anything other than another dollar
			// sign or a digit are treated as normal text.
			Pattern: abracadabraMatches2(),
			Repl:    "$😂",
			Output:  "$😂r$😂$😂$😂ra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				match, _ := m["match"].(string)
				return strings.ToUpper(match), nil
			}),
			Output: "ABrACADABra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				match, _ := m["match"].(string)
				return strings.ToUpper(match), nil
			}),
			Limit:  jtypes.NewOptionalInt(3),
			Output: "ABrACADabra",
		},
		{
			Pattern: abracadabraMatches0(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				match, _ := m["match"].(string)
				return strings.ToUpper(match), nil
			}),
			Limit:  jtypes.NewOptionalInt(0),
			Output: "abracadabra",
		},
		{
			Pattern: abracadabraMatches1(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				groups, _ := m["groups"].([]string)
				if len(groups) != 1 {
					return "", fmt.Errorf("replaceCallable expected 1 captured group, got %d", len(groups))
				}
				index, _ := m["index"].(int)
				return strings.Repeat(groups[0], index), nil
			}),
			Output: "rcccdddddbbbbbbbra",
		},
		{
			Pattern: abracadabraMatches2(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				groups, _ := m["groups"].([]string)
				if len(groups) != 2 {
					return "", fmt.Errorf("replaceCallable expected 2 captured groups, got %d", len(groups))
				}
				index, _ := m["index"].(int)
				c, _ := utf8.DecodeRuneInString(groups[1])
				return fmt.Sprintf("%d%c", index, 'ⓐ'+c-'a'), nil
			}),
			Output: "0ⓑr3ⓒ5ⓓ7ⓑra",
		},
		{
			Pattern: abracadabraMatches2(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				return 100, nil
			}),
			Error: fmt.Errorf("third argument of function replace must be a function that returns a string"),
		},
		{
			Pattern: abracadabraMatches2(),
			Repl: replaceCallable(func(m map[string]interface{}) (interface{}, error) {
				return nil, fmt.Errorf("this callable returned an error")
			}),
			Error: fmt.Errorf("this callable returned an error"),
		},
		{
			Pattern: abracadabraMatches2(),
			Repl:    100,
			Error:   fmt.Errorf("third argument of function replace must be a string or a function"),
		},
	}

	for _, test := range data {

		repl := newStringCallable(test.Repl)
		pattern := newStringCallable(test.Pattern)

		prefix := func() string {
			s := fmt.Sprintf("replace(%q, %s, %s", src, formatStringCallable(pattern), formatStringCallable(repl))
			if test.Limit.IsSet() {
				s += fmt.Sprintf(", %d", test.Limit.Int)
			}
			return s + ")"
		}

		got, err := jlib.Replace(src, pattern, repl, test.Limit)

		if got != test.Output {
			t.Errorf("%s: Expected %q, got %q", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

func TestReplaceInvalidPattern(t *testing.T) {

	_, got := jlib.Replace("abracadabra", newStringCallable(100), newStringCallable(""), jtypes.OptionalInt{})
	exp := fmt.Errorf("second argument of function replace must be a string or a regex")

	if !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected error %v, got %v", exp, got)
	}
}

func TestFormatNumber(t *testing.T) {

	data := []struct {
		Value   float64
		Picture string
		Options interface{}
		Output  string
		Error   error
	}{
		{
			Value:   0.0,
			Picture: "0",
			Output:  "0",
		},
		{
			Value:   12345.6,
			Picture: "#,###.00",
			Output:  "12,345.60",
		},
		{
			Value:   12345678.9,
			Picture: "9,999.99",
			Output:  "12,345,678.90",
		},
		{
			Value:   123.9,
			Picture: "9999",
			Output:  "0124",
		},
		{
			Value:   0.14,
			Picture: "01%",
			Output:  "14%",
		},
		{
			Value:   0.14,
			Picture: "01pc",
			// Custom percent symbol.
			Options: map[string]interface{}{
				"percent": "pc",
			},
			Output: "14pc",
		},
		{
			Value:   0.014,
			Picture: "01‰",
			Output:  "14‰",
		},
		{
			Value:   0.014,
			Picture: "01pm",
			// Custom per-mille symbol.
			Options: map[string]interface{}{
				"per-mille": "pm",
			},
			Output: "14pm",
		},
		{
			Value:   -6,
			Picture: "000",
			Output:  "-006",
		},
		{
			Value:   1234.5678,
			Picture: "#ʹ##0·00",
			// Custom grouping and decimal separators.
			Options: map[string]interface{}{
				"grouping-separator": "ʹ",
				"decimal-separator":  "·",
			},
			Output: "1ʹ234·57",
		},
		{
			Value:   1234.5678,
			Picture: "00.000E0",
			// Custom exponent separator.
			Options: map[string]interface{}{
				"exponent-separator": "E",
			},
			Output: "12.346E2",
		},
		{
			Value:   0.234,
			Picture: "0.0E0",
			// Custom exponent separator.
			Options: map[string]interface{}{
				"exponent-separator": "E",
			},
			Output: "2.3E-1",
		},
		{
			Value:   0.234,
			Picture: "#.00E0",
			// Custom exponent separator.
			Options: map[string]interface{}{
				"exponent-separator": "E",
			},
			Output: "0.23E0",
		},
		{
			Value:   0.234,
			Picture: ".00E0",
			// Custom exponent separator.
			Options: map[string]interface{}{
				"exponent-separator": "E",
			},
			Output: ".23E0",
		},
	}

	for _, test := range data {

		prefix := func() string {
			s := fmt.Sprintf("formatNumber(%g, %q", test.Value, test.Picture)
			if test.Options != nil {
				s += fmt.Sprintf(", %v", test.Options)
			}
			return s + ")"
		}

		var options jtypes.OptionalValue
		if test.Options != nil {
			// TODO: This shouldn't require nested ValueOf calls!
			options.Set(reflect.ValueOf(reflect.ValueOf(test.Options)))
		}

		got, err := jlib.FormatNumber(test.Value, test.Picture, options)

		if got != test.Output {
			t.Errorf("%s: Expected %q, got %q", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

func TestFormatBase(t *testing.T) {

	value := float64(100)

	data := []struct {
		Base   jtypes.OptionalFloat64
		Output string
		Error  error
	}{
		{
			Output: "100",
		},
		{
			Base:   jtypes.NewOptionalFloat64(2),
			Output: "1100100",
		},
		{
			Base:   jtypes.NewOptionalFloat64(8),
			Output: "144",
		},
		{
			Base:   jtypes.NewOptionalFloat64(16),
			Output: "64",
		},
		{
			Base:   jtypes.NewOptionalFloat64(20),
			Output: "50",
		},
		{
			Base:   jtypes.NewOptionalFloat64(32),
			Output: "34",
		},
		{
			Base:   jtypes.NewOptionalFloat64(36),
			Output: "2s",
		},
		{
			Base:  jtypes.NewOptionalFloat64(1),
			Error: fmt.Errorf("the second argument to formatBase must be between 2 and 36"),
		},
		{
			Base:  jtypes.NewOptionalFloat64(40),
			Error: fmt.Errorf("the second argument to formatBase must be between 2 and 36"),
		},
	}

	for _, test := range data {

		prefix := func() string {
			s := fmt.Sprintf("formatBase(%g", value)
			if test.Base.IsSet() {
				s += fmt.Sprintf(", %g", test.Base.Float64)
			}
			return s + ")"
		}

		got, err := jlib.FormatBase(value, test.Base)

		if got != test.Output {
			t.Errorf("%s: Expected %q, got %q", prefix(), test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", prefix(), test.Error, err)
		}
	}
}

// Callables

type match struct {
	Match  string
	Start  int
	End    int
	Groups []string
}

type matchCallable struct {
	name    string
	index   int
	matches []match
}

func (f *matchCallable) Name() string {
	return f.name
}

func (f *matchCallable) ParamCount() int {
	return 1
}

func (f *matchCallable) Call(vs []reflect.Value) (reflect.Value, error) {

	if f.index >= len(f.matches) {
		return reflect.Value{}, nil
	}

	obj := map[string]interface{}{
		"match":  f.matches[f.index].Match,
		"start":  f.matches[f.index].Start,
		"end":    f.matches[f.index].End,
		"groups": f.matches[f.index].Groups,
		"next":   f,
	}

	f.index++
	return reflect.ValueOf(obj), nil
}

type replaceCallable func(map[string]interface{}) (interface{}, error)

func (f replaceCallable) Name() string {
	return "<replace>"
}

func (f replaceCallable) ParamCount() int {
	return 1
}

func (f replaceCallable) Call(vs []reflect.Value) (reflect.Value, error) {

	if len(vs) != 1 {
		return reflect.Value{}, fmt.Errorf("replaceCallable: expected 1 argument, got %d", len(vs))
	}

	if !vs[0].CanInterface() {
		return reflect.Value{}, fmt.Errorf("replaceCallable: expected an interfaceable type, got %s", vs[0].Type())
	}

	m, ok := vs[0].Interface().(map[string]interface{})
	if !ok {
		return reflect.Value{}, fmt.Errorf("replaceCallable: expected argument to be a map of string to empty interface, got %s", vs[0].Type())
	}

	res, err := f(m)
	return reflect.ValueOf(res), err
}

// Helpers

func abracadabraMatches0() jtypes.Callable {
	m := abracadabraMatches("/a./")
	for i := range m.matches {
		m.matches[i].Groups = m.matches[i].Groups[:0]
	}
	return m
}

func abracadabraMatches1() jtypes.Callable {
	m := abracadabraMatches("/a(.)/")
	for i := range m.matches {
		m.matches[i].Groups = m.matches[i].Groups[1:]
	}
	return m
}

func abracadabraMatches2() jtypes.Callable {
	return abracadabraMatches("/(a)(.)/")
}

func abracadabraMatches(name string) *matchCallable {
	return &matchCallable{
		name: name,
		matches: []match{
			{
				Match: "ab",
				Start: 0,
				End:   2,
				Groups: []string{
					"a",
					"b",
				},
			},
			{
				Match: "ac",
				Start: 3,
				End:   5,
				Groups: []string{
					"a",
					"c",
				},
			},
			{
				Match: "ad",
				Start: 5,
				End:   7,
				Groups: []string{
					"a",
					"d",
				},
			},
			{
				Match: "ab",
				Start: 7,
				End:   9,
				Groups: []string{
					"a",
					"b",
				},
			},
		},
	}
}

func newStringCallable(v interface{}) jlib.StringCallable {
	return jlib.StringCallable(reflect.ValueOf(v))
}

func formatStringCallable(value jlib.StringCallable) string {

	v := reflect.Value(value).Interface()

	switch v := v.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case jtypes.Callable:
		return v.Name()
	default:
		return fmt.Sprintf("<%s>", reflect.ValueOf(v).Kind())
	}
}
