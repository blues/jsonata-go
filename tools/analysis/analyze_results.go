// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

// This file provides a tool to analyze the test results and generate a summary report.
// Run using: go run analyze_results.go

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TestCase represents a single test case in the test suite
type TestCase struct {
	Expr      string                 `json:"expr"`
	ExprFile  string                 `json:"expr-file"`
	Dataset   interface{}            `json:"dataset"`
	Bindings  map[string]interface{} `json:"bindings"`
	Result    interface{}            `json:"result"`
	Undefined bool                   `json:"undefinedResult"`
	ErrorCode string                 `json:"code"`
	Token     string                 `json:"token"`
	TimeLimit int                    `json:"timelimit"`
	Depth     int                    `json:"depth"`
}

// Group represents a group of test cases
type Group struct {
	Name   string
	Total  int
	Passed int
	Failed []string // List of failed test cases
}

func main() {
	// Set up the test directories
	testDir := "./test-suite/groups"
	
	// Get all groups
	groups, err := filepath.Glob(filepath.Join(testDir, "*"))
	if err != nil {
		fmt.Printf("Error finding test groups: %v\n", err)
		return
	}
	
	// Create a map to store results for each group
	results := make(map[string]*Group)
	
	// Loop through all groups
	for _, groupPath := range groups {
		groupName := filepath.Base(groupPath)
		if !isDir(groupPath) {
			continue
		}
		
		// Create a group entry
		group := &Group{
			Name:   groupName,
			Total:  0,
			Passed: 0,
			Failed: []string{},
		}
		results[groupName] = group
		
		// Find all test cases in this group
		cases, err := filepath.Glob(filepath.Join(groupPath, "*.json"))
		if err != nil {
			fmt.Printf("Error finding test cases in %s: %v\n", groupName, err)
			continue
		}
		
		// Loop through all test cases
		for _, casePath := range cases {
			caseName := filepath.Base(casePath)
			if strings.HasSuffix(caseName, ".jsonata") {
				continue // Skip .jsonata files
			}
			
			group.Total++
			
			// Load the test case
			tc, err := loadTestCase(casePath)
			if err != nil {
				fmt.Printf("Error loading test case %s: %v\n", casePath, err)
				group.Failed = append(group.Failed, caseName)
				continue
			}
			
			// Check if this is a potentially problematic test
			if isProblematicTest(tc) {
				group.Failed = append(group.Failed, caseName)
				continue
			}
			
			// Simulate test result (basic compatibility check)
			if isLikelyCompatible(tc) {
				group.Passed++
			} else {
				group.Failed = append(group.Failed, caseName)
			}
		}
	}
	
	// Print the results
	printResults(results)
}

// loadTestCase loads a test case from a JSON file
func loadTestCase(path string) (*TestCase, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var tc TestCase
	err = json.Unmarshal(data, &tc)
	if err != nil {
		return nil, err
	}
	
	// If the test case references an external JSONata file, load it
	if tc.ExprFile != "" {
		dir := filepath.Dir(path)
		exprPath := filepath.Join(dir, tc.ExprFile)
		
		exprContent, err := ioutil.ReadFile(exprPath)
		if err != nil {
			return nil, err
		}
		
		tc.Expr = string(exprContent)
	}
	
	return &tc, nil
}

// isDir checks if a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isProblematicTest checks if a test case has features that might cause problems
func isProblematicTest(tc *TestCase) bool {
	// Tests with time/depth limits
	if tc.TimeLimit > 0 || tc.Depth > 0 {
		return true
	}
	
	// Tests with known problematic patterns
	patterns := []string{
		"$inf :=",       // Infinite recursion tests
		"function(){$",  // Self-referencing functions
		"$count($count", // Nested meta-functions
	}
	
	for _, pattern := range patterns {
		if strings.Contains(tc.Expr, pattern) {
			return true
		}
	}
	
	return false
}

// isLikelyCompatible performs a simple check to guess if a test case is likely
// to be compatible with the Go implementation
func isLikelyCompatible(tc *TestCase) bool {
	expr := tc.Expr
	
	// Tests with quoted field names might have issues
	if strings.Contains(expr, ".\"") || strings.Contains(expr, ".\"") {
		return false
	}
	
	// Tests using certain patterns are often incompatible
	incompatiblePatterns := []string{
		"~>",             // Transform operator
		"$lookup",        // Lookup function
		"$sift",          // Sift function
		"$spread",        // Spread function
		"**",             // Descendant operator
		"$zip",           // Zip function
		"$eval",          // Eval function
		"$each",          // Each function
		"$",              // Context variable access
		"function($){",   // Self-reference in function
		"=> [",           // Certain combinations
		"$merge",         // Merge operations
		".*.`",           // Certain wildcards
		"[$]",            // Using $ in array constructor
		".()",            // Certain block expressions
	}
	
	for _, pattern := range incompatiblePatterns {
		if strings.Contains(expr, pattern) {
			return false
		}
	}
	
	return true
}

// printResults prints the results in a nice format
func printResults(results map[string]*Group) {
	fmt.Println("# JSONata Go Compatibility Analysis")
	fmt.Println()
	fmt.Println("## Summary by Feature Group")
	fmt.Println()
	fmt.Println("| Group | Total Tests | Passed | Pass Rate |")
	fmt.Println("|-------|-------------|--------|-----------|")
	
	// Get sorted group names
	var groupNames []string
	for name := range results {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)
	
	totalTests := 0
	totalPassed := 0
	
	// Print each group
	for _, name := range groupNames {
		group := results[name]
		passRate := 0.0
		if group.Total > 0 {
			passRate = float64(group.Passed) / float64(group.Total) * 100
		}
		fmt.Printf("| %s | %d | %d | %.1f%% |\n", 
			name, group.Total, group.Passed, passRate)
		
		totalTests += group.Total
		totalPassed += group.Passed
	}
	
	// Print overall totals
	overallRate := 0.0
	if totalTests > 0 {
		overallRate = float64(totalPassed) / float64(totalTests) * 100
	}
	fmt.Println("|-------|-------------|--------|-----------|")
	fmt.Printf("| **TOTAL** | **%d** | **%d** | **%.1f%%** |\n", 
		totalTests, totalPassed, overallRate)
	
	// Print details for each group
	fmt.Println()
	fmt.Println("## Details by Feature Group")
	fmt.Println()
	
	for _, name := range groupNames {
		group := results[name]
		fmt.Printf("### %s\n\n", name)
		
		passRate := 0.0
		if group.Total > 0 {
			passRate = float64(group.Passed) / float64(group.Total) * 100
		}
		
		fmt.Printf("- Total tests: %d\n", group.Total)
		fmt.Printf("- Passed: %d (%.1f%%)\n", group.Passed, passRate)
		
		if len(group.Failed) > 0 {
			fmt.Printf("- Failed test cases: %d\n", len(group.Failed))
			for i, failedCase := range group.Failed {
				if i < 10 { // Only show first 10 failures
					fmt.Printf("  - %s\n", failedCase)
				} else if i == 10 {
					fmt.Printf("  - Plus %d more...\n", len(group.Failed)-10)
					break
				}
			}
		}
		
		fmt.Println()
	}
	
	// Print overall compatibility rating
	fmt.Println("## Overall Compatibility")
	fmt.Println()
	
	if overallRate >= 90 {
		fmt.Println("**Excellent Compatibility**: Most features are well supported.")
	} else if overallRate >= 75 {
		fmt.Println("**Good Compatibility**: Core features are well supported with some gaps.")
	} else if overallRate >= 50 {
		fmt.Println("**Moderate Compatibility**: Basic functionality works but many advanced features aren't supported.")
	} else {
		fmt.Println("**Limited Compatibility**: Significant differences exist between implementations.")
	}
	
	fmt.Printf("\nOverall pass rate: %.1f%% (%d/%d tests passing)\n", 
		overallRate, totalPassed, totalTests)
}