// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package test

import (
	"testing"

	jsonata "github.com/blues/jsonata-go"
)

// TestRegularExpressions tests the regular expression features of JSONata
// which are implemented differently between JavaScript and Go
func TestRegularExpressions(t *testing.T) {
	t.Run("BasicRegex", func(t *testing.T) {
		expr, err := jsonata.Compile(`/ab/ ("ab")`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check that the result is a map with the expected fields
		match, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if match["match"] != "ab" {
			t.Errorf("Expected match 'ab', got %v", match["match"])
		}

		// Different implementations may use float64 or int for these values
		checkNumericField(t, match, "start", 0)
		checkNumericField(t, match, "end", 2)

		// Check that groups exists and is a slice
		groups, ok := match["groups"]
		if !ok {
			t.Fatalf("Expected groups field to exist")
		}

		// Handle both []interface{} and []string
		switch g := groups.(type) {
		case []interface{}:
			if len(g) != 0 {
				t.Errorf("Expected empty groups, got %v", g)
			}
		case []string:
			if len(g) != 0 {
				t.Errorf("Expected empty groups, got %v", g)
			}
		default:
			t.Errorf("Expected groups to be a slice, got %T", groups)
		}
	})

	t.Run("RegexWithNoInput", func(t *testing.T) {
		expr, err := jsonata.Compile(`/ab/ ()`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		// Some implementations might return an error, others might return nil
		if err == nil && result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("RegexWithPlusQuantifier", func(t *testing.T) {
		expr, err := jsonata.Compile(`/ab+/ ("ababbabbcc")`)
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
		expr, err := jsonata.Compile(`/a(b+)/ ("ababbabbcc")`)
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

		// Check that groups exists
		groups, ok := match["groups"]
		if !ok {
			t.Fatalf("Expected groups field to exist")
		}

		// Handle both []interface{} and []string
		switch g := groups.(type) {
		case []interface{}:
			if len(g) != 1 || g[0] != "b" {
				t.Errorf("Expected groups ['b'], got %v", g)
			}
		case []string:
			if len(g) != 1 || g[0] != "b" {
				t.Errorf("Expected groups ['b'], got %v", g)
			}
		default:
			t.Errorf("Expected groups to be a slice, got %T", groups)
		}
	})

	t.Run("RegexMatchFunction", func(t *testing.T) {
		expr, err := jsonata.Compile(`$match("ababbabbcc", /a(b+)/)`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(nil)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check if result is an array of maps
		matches, ok := result.([]interface{})
		if !ok {
			// Try as []map[string]interface{} instead
			matchMaps, ok := result.([]map[string]interface{})
			if !ok {
				t.Fatalf("Expected array result, got %T", result)
			}

			// Convert to []interface{} for consistency
			matches = make([]interface{}, len(matchMaps))
			for i, m := range matchMaps {
				matches[i] = m
			}
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

		// Different implementations may use different field names
		// Check for either "index" or "start"
		if idx, ok := match1["index"]; ok {
			checkNumericValue(t, "index", idx, 0)
		} else if start, ok := match1["start"]; ok {
			checkNumericValue(t, "start", start, 0)
		} else {
			t.Errorf("Expected either 'index' or 'start' field")
		}
	})
}

// Helper function to check numeric fields which might be float64 or int
func checkNumericField(t *testing.T, m map[string]interface{}, field string, expected int) {
	t.Helper()

	val, ok := m[field]
	if !ok {
		t.Errorf("Expected field %s to exist", field)
		return
	}

	checkNumericValue(t, field, val, expected)
}

// Helper function to check numeric values which might be float64 or int
func checkNumericValue(t *testing.T, field string, val interface{}, expected int) {
	t.Helper()

	switch v := val.(type) {
	case float64:
		if int(v) != expected {
			t.Errorf("Expected %s %d, got %v", field, expected, v)
		}
	case int:
		if v != expected {
			t.Errorf("Expected %s %d, got %v", field, expected, v)
		}
	case int64:
		if int(v) != expected {
			t.Errorf("Expected %s %d, got %v", field, expected, v)
		}
	default:
		t.Errorf("Expected %s to be numeric, got %T", field, val)
	}
}
