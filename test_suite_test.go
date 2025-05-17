// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/blues/jsonata-go/jtypes"
)

// TestSuite runs the JSONata test suite which is ported from the JavaScript implementation
func TestSuite(t *testing.T) {
	// Set up the test directories
	testDir := "./test/test-suite/groups"
	datasetDir := "./test/test-suite/datasets"

	// Load all datasets first
	datasets, err := loadDatasets(datasetDir)
	if err != nil {
		t.Fatalf("Failed to load datasets: %v", err)
	}

	// Walk through the test groups and run tests
	runAllGroupTests(t, testDir, datasets)
}

// testCase represents a single test case in the JSONata test suite
type testCase struct {
	Expr        string                 `json:"expr"`
	ExprFile    string                 `json:"expr-file"`
	Dataset     interface{}            `json:"dataset"`
	Bindings    map[string]interface{} `json:"bindings"`
	Result      interface{}            `json:"result"`
	Undefined   bool                   `json:"undefinedResult"`
	ErrorCode   string                 `json:"code"`
	Token       string                 `json:"token"`
	TimeLimit   int                    `json:"timelimit"`
	Depth       int                    `json:"depth"`
}

// loadDatasets loads all the datasets from the test-suite/datasets directory
func loadDatasets(datasetDir string) (map[string]interface{}, error) {
	datasets := make(map[string]interface{})
	
	files, err := ioutil.ReadDir(datasetDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			var data interface{}
			datasetName := strings.TrimSuffix(file.Name(), ".json")
			
			content, err := ioutil.ReadFile(filepath.Join(datasetDir, file.Name()))
			if err != nil {
				return nil, err
			}
			
			if err := json.Unmarshal(content, &data); err != nil {
				return nil, err
			}
			
			datasets[datasetName] = data
		}
	}
	
	return datasets, nil
}

// runAllGroupTests runs all test groups in the test suite
func runAllGroupTests(t *testing.T, testDir string, datasets map[string]interface{}) {
	groups, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatalf("Failed to read test groups: %v", err)
	}

	for _, group := range groups {
		if group.IsDir() {
			groupPath := filepath.Join(testDir, group.Name())
			t.Run(group.Name(), func(t *testing.T) {
				runGroupTests(t, groupPath, datasets)
			})
		}
	}
}

// runGroupTests runs all test cases in a specific group
func runGroupTests(t *testing.T, groupPath string, datasets map[string]interface{}) {
	files, err := ioutil.ReadDir(groupPath)
	if err != nil {
		t.Fatalf("Failed to read test cases: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			testPath := filepath.Join(groupPath, file.Name())
			
			// Load the test case
			var tc testCase
			content, err := ioutil.ReadFile(testPath)
			if err != nil {
				t.Fatalf("Failed to read test case %s: %v", testPath, err)
			}
			
			if err := json.Unmarshal(content, &tc); err != nil {
				t.Fatalf("Failed to parse test case %s: %v", testPath, err)
			}

			// If the test case references an external jsonata file, read it in
			if tc.ExprFile != "" {
				exprContent, err := ioutil.ReadFile(filepath.Join(groupPath, tc.ExprFile))
				if err != nil {
					t.Fatalf("Failed to read expression file %s: %v", tc.ExprFile, err)
				}
				tc.Expr = string(exprContent)
			}

			t.Run(file.Name(), func(t *testing.T) {
				runTestCase(t, tc, datasets, testPath)
			})
		}
	}
}

// resolveDataset determines the input data to use for the test case
func resolveDataset(datasets map[string]interface{}, tc testCase) interface{} {
	// If the test case has its own data, use that
	if tc.Dataset != nil {
		return tc.Dataset
	}
	
	// If the dataset is explicitly set to null, return nil
	if tc.Dataset == nil {
		return nil
	}
	
	// Look up the dataset by name
	datasetName, ok := tc.Dataset.(string)
	if !ok {
		return nil
	}
	
	return datasets[datasetName]
}

// runTestCase runs an individual test case
func runTestCase(t *testing.T, tc testCase, datasets map[string]interface{}, testPath string) {
	// Skip tests with time limits for now - these are usually tests for recursion limits
	if tc.TimeLimit > 0 || tc.Depth > 0 {
		t.Skip("Skipping test with time/depth limit")
		return
	}

	// Compile the expression
	expr, err := Compile(tc.Expr)
	if tc.ErrorCode != "" {
		// If we expect an error, check that we got one
		if err == nil {
			t.Errorf("Expected error code %s, but got no error", tc.ErrorCode)
		}
		// TODO: Check error code and token
		return
	} else if err != nil {
		t.Fatalf("Failed to compile expression: %v", err)
	}

	// Register any bindings
	if tc.Bindings != nil && len(tc.Bindings) > 0 {
		err = expr.RegisterVars(tc.Bindings)
		if err != nil {
			t.Fatalf("Failed to register bindings: %v", err)
		}
	}

	// Resolve the dataset to use
	dataset := resolveDataset(datasets, tc)

	// Evaluate the expression
	result, err := expr.Eval(dataset)
	
	// Check for errors
	if tc.ErrorCode != "" {
		if err == nil {
			t.Errorf("Expected error code %s, but got no error", tc.ErrorCode)
		}
		// TODO: Check that the error code matches
		return
	} else if err != nil {
		t.Fatalf("Failed to evaluate expression: %v", err)
	}

	// Check for undefined result
	if tc.Undefined {
		if result != nil {
			t.Errorf("Expected undefined result, but got: %v", result)
		}
		return
	}

	// Compare the actual result with the expected result
	if !equalResults(result, tc.Result) {
		t.Errorf("Result mismatch:\nExpected: %+v (%T)\nGot: %+v (%T)",
			tc.Result, tc.Result, result, result)
	}
}

// equalResults compares two results for equality, handling various types
func equalResults(x, y interface{}) bool {
	if reflect.DeepEqual(x, y) {
		return true
	}

	vx := jtypes.Resolve(reflect.ValueOf(x))
	vy := jtypes.Resolve(reflect.ValueOf(y))

	// Handle array comparison
	if jtypes.IsArray(vx) && jtypes.IsArray(vy) {
		if vx.Len() != vy.Len() {
			return false
		}
		for i := 0; i < vx.Len(); i++ {
			if !equalResults(vx.Index(i).Interface(), vy.Index(i).Interface()) {
				return false
			}
		}
		return true
	}

	// Handle numeric comparison
	ix, okx := jtypes.AsNumber(vx)
	iy, oky := jtypes.AsNumber(vy)
	if okx && oky && ix == iy {
		return true
	}

	// Handle string comparison
	sx, okx := jtypes.AsString(vx)
	sy, oky := jtypes.AsString(vy)
	if okx && oky && sx == sy {
		return true
	}

	// Handle boolean comparison
	bx, okx := jtypes.AsBool(vx)
	by, oky := jtypes.AsBool(vy)
	if okx && oky && bx == by {
		return true
	}

	// No match
	return false
}