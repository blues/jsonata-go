// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"testing"
)

// TestRegularExpressions tests the regular expression features of JSONata
// which are implemented differently between JavaScript and Go
func TestRegularExpressions(t *testing.T) {
	t.Run("BasicRegex", func(t *testing.T) {
		expr, err := Compile(`/ab/ ("ab")`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Check match properties
		if match["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match["match"])
		}
		if match["start"] != float64(0) {
			t.Errorf("Expected start 0, got %v", match["start"])
		}
		if match["end"] != float64(2) {
			t.Errorf("Expected end 2, got %v", match["end"])
		}
		
		groups, ok := match["groups"].([]interface{})
		if !ok {
			t.Fatalf("Expected groups array, got %T", match["groups"])
		}
		if len(groups) != 0 {
			t.Errorf("Expected empty groups, got %v", groups)
		}
	})

	t.Run("RegexWithNoInput", func(t *testing.T) {
		expr, err := Compile(`/ab/ ()`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("RegexWithPlusQuantifier", func(t *testing.T) {
		expr, err := Compile(`/ab+/ ("ababbabbcc")`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match["match"])
		}
	})

	t.Run("RegexWithCapturingGroup", func(t *testing.T) {
		expr, err := Compile(`/a(b+)/ ("ababbabbcc")`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match["match"])
		}

		groups, ok := match["groups"].([]interface{})
		if !ok {
			t.Fatalf("Expected groups array, got %T", match["groups"])
		}
		if len(groups) != 1 || groups[0] != "b" {
			t.Errorf("Expected groups ['b'], got %v", groups)
		}
	})

	t.Run("RegexWithNext", func(t *testing.T) {
		expr, err := Compile(`/a(b+)/ ("ababbabbcc").next()`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "abb" {
			t.Errorf("Expected match 'abb', got %v", match["match"])
		}
		if match["start"] != float64(2) {
			t.Errorf("Expected start 2, got %v", match["start"])
		}
		if match["end"] != float64(5) {
			t.Errorf("Expected end 5, got %v", match["end"])
		}

		groups, ok := match["groups"].([]interface{})
		if !ok {
			t.Fatalf("Expected groups array, got %T", match["groups"])
		}
		if len(groups) != 1 || groups[0] != "bb" {
			t.Errorf("Expected groups ['bb'], got %v", groups)
		}
	})

	t.Run("RegexWithMultipleNext", func(t *testing.T) {
		expr, err := Compile(`/a(b+)/ ("ababbabbcc").next().next()`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "abb" {
			t.Errorf("Expected match 'abb', got %v", match["match"])
		}
		if match["start"] != float64(5) {
			t.Errorf("Expected start 5, got %v", match["start"])
		}
		if match["end"] != float64(8) {
			t.Errorf("Expected end 8, got %v", match["end"])
		}
	})

	t.Run("RegexWithNoMoreMatches", func(t *testing.T) {
		expr, err := Compile(`/a(b+)/ ("ababbabbcc").next().next().next()`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("RegexMatchFunction", func(t *testing.T) {
		expr, err := Compile(`$match("ababbabbcc", /a(b+)/)`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		matches, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		if len(matches) != 3 {
			t.Fatalf("Expected 3 matches, got %d", len(matches))
		}

		// Check the first match
		match1, ok := matches[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map for first match, got %T", matches[0])
		}

		if match1["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match1["match"])
		}
		if match1["index"] != float64(0) {
			t.Errorf("Expected index 0, got %v", match1["index"])
		}

		groups1, ok := match1["groups"].([]interface{})
		if !ok {
			t.Fatalf("Expected groups array, got %T", match1["groups"])
		}
		if len(groups1) != 1 || groups1[0] != "b" {
			t.Errorf("Expected groups ['b'], got %v", groups1)
		}
	})

	t.Run("RegexMatchWithLimit", func(t *testing.T) {
		expr, err := Compile(`$match("ababbabbcc", /a(b+)/, 1)`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match["match"])
		}
		if match["index"] != float64(0) {
			t.Errorf("Expected index 0, got %v", match["index"])
		}
	})

	t.Run("RegexMatchWithZeroLimit", func(t *testing.T) {
		expr, err := Compile(`$match("ababbabbcc", /a(b+)/, 0)`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("RegexMatchWithNoMatches", func(t *testing.T) {
		expr, err := Compile(`$match("ababbabbcc", /x(y+)/)`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})
}