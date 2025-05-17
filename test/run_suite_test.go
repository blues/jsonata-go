// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	jsonata "github.com/blues/jsonata-go"
)

// TestRunSimpleCase tests running a simple test case from the test suite
func TestRunSimpleCase(t *testing.T) {
	// Test a simple case directly
	casePath := "./test-suite/groups/function-abs/case000.json"

	// Load the test case
	content, err := ioutil.ReadFile(casePath)
	if err != nil {
		t.Fatalf("Failed to read test case: %v", err)
	}

	var testCase struct {
		Expr     string                 `json:"expr"`
		Dataset  interface{}            `json:"dataset"`
		Bindings map[string]interface{} `json:"bindings"`
		Result   interface{}            `json:"result"`
	}

	if err := json.Unmarshal(content, &testCase); err != nil {
		t.Fatalf("Failed to parse test case: %v", err)
	}

	// Compile the expression
	expr, err := jsonata.Compile(testCase.Expr)
	if err != nil {
		t.Fatalf("Failed to compile expression: %v", err)
	}

	// Register any bindings
	if testCase.Bindings != nil && len(testCase.Bindings) > 0 {
		err = expr.RegisterVars(testCase.Bindings)
		if err != nil {
			t.Fatalf("Failed to register bindings: %v", err)
		}
	}

	// Evaluate the expression
	result, err := expr.Eval(nil)
	if err != nil {
		t.Fatalf("Failed to evaluate expression: %v", err)
	}

	// For this specific case, we expect 3.7
	expected := 3.7
	if result != expected {
		t.Errorf("Result mismatch: expected %v, got %v", expected, result)
	}
}

// TestRunGroupSample tests running a sample of test cases from a specific group
func TestRunGroupSample(t *testing.T) {
	testCases := []struct {
		group    string
		file     string
		expected interface{}
	}{
		{"numeric-operators", "case000.json", float64(140)},
		{"array-constructor", "case000.json", []interface{}{}},
		{"literals", "case001.json", "hello"},
		{"function-boolean", "case000.json", true},
		{"function-string", "case000.json", "5"},
	}

	for _, tc := range testCases {
		t.Run(tc.group+"/"+tc.file, func(t *testing.T) {
			// Load the test case
			casePath := filepath.Join("./test-suite/groups", tc.group, tc.file)
			content, err := ioutil.ReadFile(casePath)
			if err != nil {
				t.Fatalf("Failed to read test case: %v", err)
			}

			var testCase struct {
				Expr     string                 `json:"expr"`
				Dataset  interface{}            `json:"dataset"`
				Bindings map[string]interface{} `json:"bindings"`
				Result   interface{}            `json:"result"`
			}

			if err := json.Unmarshal(content, &testCase); err != nil {
				t.Fatalf("Failed to parse test case: %v", err)
			}

			// Compile the expression
			expr, err := jsonata.Compile(testCase.Expr)
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			// Register any bindings
			if testCase.Bindings != nil && len(testCase.Bindings) > 0 {
				err = expr.RegisterVars(testCase.Bindings)
				if err != nil {
					t.Fatalf("Failed to register bindings: %v", err)
				}
			}

			// If the test case has a dataset, load it
			var data interface{}
			if testCase.Dataset != nil {
				datasetName, ok := testCase.Dataset.(string)
				if ok && datasetName != "" {
					datasetPath := filepath.Join("./test-suite/datasets", datasetName+".json")
					datasetContent, err := ioutil.ReadFile(datasetPath)
					if err != nil {
						t.Fatalf("Failed to read dataset: %v", err)
					}

					if err := json.Unmarshal(datasetContent, &data); err != nil {
						t.Fatalf("Failed to parse dataset: %v", err)
					}
				}
			}

			// Evaluate the expression
			result, err := expr.Eval(data)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Check the result
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Result mismatch: expected %v (%T), got %v (%T)",
					tc.expected, tc.expected, result, result)
			}
		})
	}
}
