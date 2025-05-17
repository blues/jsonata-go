// feature_matrix.go - Generate a detailed feature compatibility matrix for JSONata-Go
// Run with: go run feature_matrix.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type TestCase struct {
	Expr     string      `json:"expr"`
	Data     interface{} `json:"data"`
	Expected interface{} `json:"expected"`
	Dataset  string      `json:"dataset,omitempty"`
	Error    bool        `json:"error,omitempty"`
	Code     string      `json:"code,omitempty"`
}

type Feature struct {
	Name        string
	Description string
	Cases       map[string]bool // maps case ID to pass/fail
	PassCount   int
	TotalCount  int
	PassRate    float64
}

var featureDescriptions = map[string]string{
	"array-constructor":      "Array creation with []",
	"blocks":                 "Code blocks with { }",
	"boolean-expresssions":   "Boolean expressions (and, or, etc.)",
	"closures":               "Function closures",
	"comparison-operators":   "Comparison operators (==, !=, etc.)",
	"conditionals":           "Conditional expressions",
	"context":                "Context variables",
	"descendent-operator":    "Descendent operator (..)",
	"encoding":               "Character encoding",
	"errors":                 "Error handling",
	"fields":                 "Field access",
	"flattening":             "Array flattening",
	"function-abs":           "$abs function",
	"function-append":        "$append function",
	"function-applications":  "Function application",
	"function-average":       "$average function",
	"function-boolean":       "$boolean function",
	"function-ceil":          "$ceil function",
	"function-contains":      "$contains function",
	"function-count":         "$count function",
	"function-each":          "$each function",
	"function-exists":        "$exists function",
	"function-floor":         "$floor function",
	"function-formatBase":    "$formatBase function",
	"function-formatNumber":  "$formatNumber function",
	"function-fromMillis":    "$fromMillis function",
	"function-join":          "$join function",
	"function-keys":          "$keys function",
	"function-length":        "$length function",
	"function-lookup":        "$lookup function",
	"function-lowercase":     "$lowercase function",
	"function-max":           "$max function",
	"function-merge":         "$merge function",
	"function-number":        "$number function",
	"function-pad":           "$pad function",
	"function-power":         "$power function",
	"function-replace":       "$replace function",
	"function-reverse":       "$reverse function",
	"function-round":         "$round function",
	"function-shuffle":       "$shuffle function",
	"function-sift":          "$sift function",
	"function-signatures":    "Function signatures",
	"function-sort":          "$sort function",
	"function-split":         "$split function",
	"function-spread":        "$spread function",
	"function-sqrt":          "$sqrt function",
	"function-string":        "$string function",
	"function-substring":     "$substring function",
	"function-substringAfter": "$substringAfter function",
	"function-substringBefore": "$substringBefore function",
	"function-sum":           "$sum function",
	"function-tomillis":      "$toMillis function",
	"function-trim":          "$trim function",
	"function-uppercase":     "$uppercase function",
	"function-zip":           "$zip function",
	"higher-order-functions": "Higher-order functions",
	"hof-filter":             "$filter function",
	"hof-map":                "$map function",
	"hof-reduce":             "$reduce function",
	"hof-zip-map":            "$zip/$map composition",
	"inclusion-operator":     "Inclusion operator (in)",
	"lambdas":                "Lambda expressions",
	"literals":               "Literal values",
	"matchers":               "Pattern matchers",
	"missing-paths":          "Handling of missing paths",
	"multiple-array-selectors": "Multiple array selectors",
	"null":                   "Null handling",
	"numeric-operators":      "Numeric operators (+, -, *, /)",
	"object-constructor":     "Object creation with { }",
	"parentheses":            "Parenthesized expressions",
}

func main() {
	fmt.Println("Generating JSONata-Go Feature Compatibility Matrix...")

	// Scan test directory to find all test cases
	groups, err := scanTestGroups()
	if err != nil {
		fmt.Println("Error scanning test groups:", err)
		return
	}

	// Run tests and collect results
	testResults, err := runTests()
	if err != nil {
		fmt.Println("Error running tests:", err)
		return
	}

	// Build feature matrix
	features := buildFeatureMatrix(groups, testResults)

	// Generate report
	generateFeatureMatrix(features)
}

func scanTestGroups() (map[string][]string, error) {
	groups := make(map[string][]string)

	// Find all test groups
	groupsDir := "./test-suite/groups"
	
	entries, err := os.ReadDir(groupsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading groups directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			groupName := entry.Name()
			groupPath := filepath.Join(groupsDir, groupName)
			
			// Find all test cases in this group
			cases, err := os.ReadDir(groupPath)
			if err != nil {
				fmt.Printf("Warning: Could not read group directory %s: %v\n", groupPath, err)
				continue
			}
			
			var caseIDs []string
			for _, c := range cases {
				if !c.IsDir() && strings.HasSuffix(c.Name(), ".json") {
					caseID := strings.TrimSuffix(c.Name(), ".json")
					caseIDs = append(caseIDs, caseID)
				}
			}
			
			groups[groupName] = caseIDs
		}
	}

	return groups, nil
}

func runTests() (map[string]bool, error) {
	fmt.Println("Running tests...")

	// Run the tests with verbose output
	cmd := exec.Command("go", "test", "-v", "./...")
	output, _ := cmd.CombinedOutput()
	
	// Parse test results
	results := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "--- PASS: TestSuite/TestCase_") {
			parts := strings.Split(line, "TestCase_")
			if len(parts) > 1 {
				testID := strings.TrimSpace(parts[1])
				results[testID] = true
			}
		} else if strings.Contains(line, "--- FAIL: TestSuite/TestCase_") {
			parts := strings.Split(line, "TestCase_")
			if len(parts) > 1 {
				testID := strings.TrimSpace(parts[1])
				results[testID] = false
			}
		}
	}

	fmt.Printf("Collected results for %d tests\n", len(results))
	return results, nil
}

func buildFeatureMatrix(groups map[string][]string, testResults map[string]bool) []*Feature {
	features := make(map[string]*Feature)
	
	// Initialize features
	for groupName := range groups {
		desc := groupName
		if description, ok := featureDescriptions[groupName]; ok {
			desc = description
		}
		
		features[groupName] = &Feature{
			Name:        groupName,
			Description: desc,
			Cases:       make(map[string]bool),
		}
	}
	
	// Populate test results
	for groupName, cases := range groups {
		feature := features[groupName]
		
		for _, caseID := range cases {
			testID := groupName + "_" + caseID
			passed, exists := testResults[testID]
			if exists {
				feature.Cases[caseID] = passed
				feature.TotalCount++
				if passed {
					feature.PassCount++
				}
			}
		}
		
		// Calculate pass rate
		if feature.TotalCount > 0 {
			feature.PassRate = float64(feature.PassCount) / float64(feature.TotalCount) * 100
		}
	}
	
	// Convert to sorted slice
	var featureList []*Feature
	for _, feature := range features {
		featureList = append(featureList, feature)
	}
	
	sort.Slice(featureList, func(i, j int) bool {
		return featureList[i].PassRate > featureList[j].PassRate
	})
	
	return featureList
}

func generateFeatureMatrix(features []*Feature) {
	// Calculate overall stats
	var totalTests, passedTests int
	for _, feature := range features {
		totalTests += feature.TotalCount
		passedTests += feature.PassCount
	}
	
	overallPassRate := 0.0
	if totalTests > 0 {
		overallPassRate = float64(passedTests) / float64(totalTests) * 100
	}

	// Generate markdown report
	report := fmt.Sprintf(`# JSONata-Go Feature Compatibility Matrix

## Overview

This document provides a detailed view of JSONata-Go's compatibility with the original JavaScript implementation.

- **Total Test Cases:** %d
- **Passing Test Cases:** %d
- **Overall Compatibility:** %.1f%%

## Feature Compatibility

| Feature | Description | Compatibility | Tests | Passed | Failed |
|---------|-------------|--------------|-------|--------|--------|
`, totalTests, passedTests, overallPassRate)

	for _, feature := range features {
		failed := feature.TotalCount - feature.PassCount
		report += fmt.Sprintf("| %s | %s | %.1f%% | %d | %d | %d |\n",
			feature.Name, feature.Description, feature.PassRate, 
			feature.TotalCount, feature.PassCount, failed)
	}
	
	report += `
## Compatibility Categories

### Well-Supported Features (>75% compatibility)

`
	for _, feature := range features {
		if feature.PassRate >= 75 {
			report += fmt.Sprintf("- **%s**: %.1f%% (%d/%d tests pass)\n", 
				feature.Description, feature.PassRate, feature.PassCount, feature.TotalCount)
		}
	}
	
	report += `
### Partially-Supported Features (25-75% compatibility)

`
	for _, feature := range features {
		if feature.PassRate >= 25 && feature.PassRate < 75 {
			report += fmt.Sprintf("- **%s**: %.1f%% (%d/%d tests pass)\n", 
				feature.Description, feature.PassRate, feature.PassCount, feature.TotalCount)
		}
	}
	
	report += `
### Minimally-Supported Features (<25% compatibility)

`
	for _, feature := range features {
		if feature.PassRate < 25 {
			report += fmt.Sprintf("- **%s**: %.1f%% (%d/%d tests pass)\n", 
				feature.Description, feature.PassRate, feature.PassCount, feature.TotalCount)
		}
	}
	
	report += `
## Implementation Priorities

Based on the compatibility matrix, here are suggested priorities for implementation:

1. **High Priority**: Functions with high usage but low compatibility
2. **Medium Priority**: Partially implemented features that need completion
3. **Low Priority**: Specialized features with limited practical use

This report was automatically generated by the JSONata-Go feature_matrix tool.
`

	// Write the report to a file
	err := os.WriteFile("feature_matrix.md", []byte(report), 0644)
	if err != nil {
		fmt.Println("Error writing feature matrix report:", err)
		return
	}

	fmt.Println("Feature matrix report generated: feature_matrix.md")
}