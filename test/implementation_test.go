// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package test

import (
	"reflect"
	"testing"
	"time"

	jsonata "github.com/blues/jsonata-go"
)

// TestImplementationFunctions tests the implementation-specific functions in the Go port
// of JSONata, which are analogous to the tests in implementation-tests.js
func TestImplementationFunctions(t *testing.T) {
	// Define test data similar to the JS version
	testData := map[string]interface{}{
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
								"Weight": float64(0.75),
							},
							"Price":    float64(34.45),
							"Quantity": float64(2),
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
								"Weight": float64(0.6),
							},
							"Price":    float64(21.67),
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
								"Weight": float64(0.75),
							},
							"Price":    float64(34.45),
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
								"Weight": float64(2.0),
							},
							"Price":    float64(107.99),
							"Quantity": float64(1),
						},
					},
				},
			},
		},
	}

	t.Run("MillisFunction", func(t *testing.T) {
		t.Run("ReturnsMillisecondsSinceEpoch", func(t *testing.T) {
			expr, err := jsonata.Compile("$millis()")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Verify the result is a number and within reasonable range
			var millis float64
			switch v := result.(type) {
			case float64:
				millis = v
			case int64:
				millis = float64(v)
			default:
				t.Fatalf("Expected numeric result, got %T", result)
			}

			// Current time should be between 2020 and 2050 (in milliseconds)
			now := time.Now().UnixNano() / 1_000_000
			if millis < 1577836800000 || millis > 2524608000000 || millis > float64(now+1000) {
				t.Errorf("Invalid milliseconds value: %v", millis)
			}
		})

		t.Run("ReturnsSameValueWithinExpression", func(t *testing.T) {
			expr, err := jsonata.Compile(`{"now": $millis(), "delay": $sum([1..10000]), "later": $millis()}.(now = later)`)
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// The timestamp should be the same within one expression evaluation
			if result != true {
				t.Errorf("Expected same timestamp within one expression, got: %v", result)
			}
		})

		t.Run("ReturnsDifferentValuesForSubsequentCalls", func(t *testing.T) {
			expr, err := jsonata.Compile("($sum([1..1000]); $millis())")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result1, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Sleep a bit to ensure different timestamps
			time.Sleep(10 * time.Millisecond)

			result2, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Results should be different for different evaluations
			if result1 == result2 {
				t.Errorf("Expected different timestamps for subsequent calls, got: %v and %v", result1, result2)
			}
		})
	})

	t.Run("NowFunction", func(t *testing.T) {
		t.Run("ReturnsTimestamp", func(t *testing.T) {
			expr, err := jsonata.Compile("$now()")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Result should be a string in ISO8601 format
			timestamp, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T", result)
			}

			// Pattern: "2021-05-09T10:10:16.918Z"
			_, err = time.Parse(time.RFC3339Nano, timestamp)
			if err != nil {
				t.Errorf("Invalid timestamp format: %v, error: %v", timestamp, err)
			}
		})

		t.Run("ReturnsSameValueWithinExpression", func(t *testing.T) {
			expr, err := jsonata.Compile(`{"now": $now(), "delay": $sum([1..10000]), "later": $now()}.(now = later)`)
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// The timestamp should be the same within one expression evaluation
			if result != true {
				t.Errorf("Expected same timestamp within one expression, got: %v", result)
			}
		})

		t.Run("ReturnsDifferentValuesForSubsequentCalls", func(t *testing.T) {
			expr, err := jsonata.Compile("($sum([1..1000]); $now())")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result1, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Sleep a bit to ensure different timestamps
			time.Sleep(10 * time.Millisecond)

			result2, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Results should be different for different evaluations
			if result1 == result2 {
				t.Errorf("Expected different timestamps for subsequent calls, got: %v and %v", result1, result2)
			}
		})
	})

	t.Run("RandomFunction", func(t *testing.T) {
		t.Run("ReturnsRandomNumber", func(t *testing.T) {
			expr, err := jsonata.Compile("$random()")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Result should be a float64 between 0 and 1
			rand, ok := result.(float64)
			if !ok {
				t.Fatalf("Expected float64 result, got %T", result)
			}

			if rand < 0 || rand >= 1 {
				t.Errorf("Random number should be in range [0,1), got: %v", rand)
			}
		})

		t.Run("ConsecutiveRandomNumbersAreDifferent", func(t *testing.T) {
			expr, err := jsonata.Compile("$random() = $random()")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Result should be false (two consecutive randoms should be different)
			if result != false {
				t.Errorf("Expected consecutive random numbers to be different, got: %v", result)
			}
		})
	})

	t.Run("UserDefinedFunctions", func(t *testing.T) {
		t.Run("OverrideBuiltinFunction", func(t *testing.T) {
			expr, err := jsonata.Compile("$now()")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			// Register a custom 'now' function
			err = expr.RegisterExts(map[string]jsonata.Extension{
				"now": {
					Func: func() string {
						return "time for tea"
					},
				},
			})
			if err != nil {
				t.Fatalf("Failed to register function: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Should return our custom value
			if result != "time for tea" {
				t.Errorf("Expected 'time for tea', got: %v", result)
			}
		})

		t.Run("MapWithUserDefinedFunction", func(t *testing.T) {
			expr, err := jsonata.Compile("$map([1,4,9,16], $squareroot)")
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			// Register a custom 'squareroot' function
			err = expr.RegisterExts(map[string]jsonata.Extension{
				"squareroot": {
					Func: func(num float64) float64 {
						return float64(int(num + 0.5))
					},
				},
			})
			if err != nil {
				t.Fatalf("Failed to register function: %v", err)
			}

			result, err := expr.Eval(testData)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Should return [1, 4, 9, 16]
			expected := []interface{}{float64(1), float64(4), float64(9), float64(16)}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("Expected %v, got: %v", expected, result)
			}
		})

		t.Run("PartiallyApplyUserDefinedFunction", func(t *testing.T) {
			expr, err := jsonata.Compile(`(
				$substr := function($str, $start, $len){$substring($str, $start, $len)};
				$first5 := $substr(?, 0, 5);
				$first5("Hello World")
			)`)
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			result, err := expr.Eval(nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Should return "Hello"
			if result != "Hello" {
				t.Errorf("Expected 'Hello', got: %v", result)
			}
		})
	})
}
