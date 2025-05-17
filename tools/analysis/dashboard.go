// dashboard.go - A tool to analyze and display JSONata-Go test results
// Run with: go run dashboard.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

type TestResult struct {
	Pass     bool   `json:"pass"`
	TestFile string `json:"test_file"`
	TestName string `json:"test_name"`
	Group    string `json:"group"`
	Case     string `json:"case"`
	Error    string `json:"error,omitempty"`
}

type TestSummary struct {
	TotalTests  int
	PassedTests int
	FailedTests int
	PassRate    float64
}

type GroupSummary struct {
	Name        string
	TotalTests  int
	PassedTests int
	FailedTests int
	PassRate    float64
}

func main() {
	fmt.Println("JSONata-Go Compatibility Dashboard")
	fmt.Println("=================================")

	// Run tests with JSON output and capture results
	results := runTests()
	if results == nil {
		return
	}

	// Generate summary data
	summary := summarizeResults(results)
	groupSummaries := summarizeByGroup(results)

	// Display overall summary
	printSummary(summary)

	// Display group summaries
	printGroupSummaries(groupSummaries)

	// Generate detailed failure analysis
	analyzeFailures(results)

	// Generate compatibility report
	generateCompatibilityReport(groupSummaries, summary)
}

func runTests() []TestResult {
	fmt.Println("Running tests...")

	// Create a temporary directory for test results
	tmpDir, err := os.MkdirTemp("", "jsonata-tests")
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return nil
	}
	defer os.RemoveAll(tmpDir)

	jsonFile := filepath.Join(tmpDir, "test-results.json")

	// Run the tests with JSON output
	cmd := exec.Command("go", "test", "./...", "-v", "-json")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Tests failed:", err)
		// Continue to process results anyway
	}

	// Save the raw output
	err = os.WriteFile(jsonFile, outputBytes, 0644)
	if err != nil {
		fmt.Println("Error saving test output:", err)
		return nil
	}

	// Parse the results
	lines := strings.Split(string(outputBytes), "\n")
	var results []TestResult

	for _, line := range lines {
		if line == "" {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Only process test results
		if event["Action"] != "fail" && event["Action"] != "pass" {
			continue
		}

		testName, ok := event["Test"].(string)
		if !ok {
			continue
		}

		// Extract group and case from test name
		group, testCase := extractGroupAndCase(testName)

		result := TestResult{
			Pass:     event["Action"] == "pass",
			TestFile: event["Package"].(string),
			TestName: testName,
			Group:    group,
			Case:     testCase,
		}

		if !result.Pass {
			if output, ok := event["Output"].(string); ok {
				result.Error = output
			}
		}

		results = append(results, result)
	}

	fmt.Printf("Processed %d test results\n", len(results))
	return results
}

func extractGroupAndCase(testName string) (string, string) {
	// Example test name: "TestSuite/TestCase_flattening_case004"
	parts := strings.Split(testName, "/")
	if len(parts) < 2 {
		return "unknown", "unknown"
	}

	casePart := parts[len(parts)-1]
	if !strings.HasPrefix(casePart, "TestCase_") {
		return "unknown", "unknown"
	}

	// Remove "TestCase_" prefix
	casePart = strings.TrimPrefix(casePart, "TestCase_")

	// Split into group and case number
	lastUnderscore := strings.LastIndex(casePart, "_")
	if lastUnderscore == -1 {
		return casePart, "unknown"
	}

	group := casePart[:lastUnderscore]
	caseNumber := casePart[lastUnderscore+1:]

	return group, caseNumber
}

func summarizeResults(results []TestResult) TestSummary {
	var summary TestSummary
	summary.TotalTests = len(results)

	for _, result := range results {
		if result.Pass {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
	}

	if summary.TotalTests > 0 {
		summary.PassRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}

	return summary
}

func summarizeByGroup(results []TestResult) []GroupSummary {
	groups := make(map[string]*GroupSummary)

	for _, result := range results {
		if _, exists := groups[result.Group]; !exists {
			groups[result.Group] = &GroupSummary{Name: result.Group}
		}

		group := groups[result.Group]
		group.TotalTests++

		if result.Pass {
			group.PassedTests++
		} else {
			group.FailedTests++
		}
	}

	// Convert map to slice for sorting
	var summaries []GroupSummary
	for _, group := range groups {
		if group.TotalTests > 0 {
			group.PassRate = float64(group.PassedTests) / float64(group.TotalTests) * 100
		}
		summaries = append(summaries, *group)
	}

	// Sort by pass rate in descending order
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].PassRate > summaries[j].PassRate
	})

	return summaries
}

func printSummary(summary TestSummary) {
	fmt.Println("\nOverall Summary")
	fmt.Println("---------------")
	fmt.Printf("Total Tests:   %d\n", summary.TotalTests)
	fmt.Printf("Passed:        %d\n", summary.PassedTests)
	fmt.Printf("Failed:        %d\n", summary.FailedTests)
	fmt.Printf("Success Rate:  %.1f%%\n", summary.PassRate)
}

func printGroupSummaries(groups []GroupSummary) {
	fmt.Println("\nResults by Feature Group")
	fmt.Println("----------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Group\tTests\tPassed\tFailed\tPass Rate")
	fmt.Fprintln(w, "-----\t-----\t------\t------\t---------")

	for _, group := range groups {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%.1f%%\n",
			group.Name, group.TotalTests, group.PassedTests, group.FailedTests, group.PassRate)
	}
	w.Flush()
}

func analyzeFailures(results []TestResult) {
	fmt.Println("\nCommon Failure Patterns")
	fmt.Println("----------------------")

	patterns := map[string]int{}
	
	for _, result := range results {
		if !result.Pass && result.Error != "" {
			errorType := categorizeError(result.Error)
			patterns[errorType]++
		}
	}

	// Convert to slice for sorting
	type patternCount struct {
		pattern string
		count   int
	}
	
	var sortedPatterns []patternCount
	for pattern, count := range patterns {
		sortedPatterns = append(sortedPatterns, patternCount{pattern, count})
	}
	
	sort.Slice(sortedPatterns, func(i, j int) bool {
		return sortedPatterns[i].count > sortedPatterns[j].count
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Error Type\tCount")
	fmt.Fprintln(w, "----------\t-----")
	
	for _, pc := range sortedPatterns {
		fmt.Fprintf(w, "%s\t%d\n", pc.pattern, pc.count)
	}
	w.Flush()
}

func categorizeError(errorText string) string {
	// Extract common error patterns
	if strings.Contains(errorText, "expected but got") {
		return "Result Mismatch"
	} else if strings.Contains(errorText, "not implemented") || strings.Contains(errorText, "function not found") {
		return "Missing Implementation"
	} else if strings.Contains(errorText, "syntax error") {
		return "Syntax Error"
	} else if strings.Contains(errorText, "runtime error") || strings.Contains(errorText, "panic") {
		return "Runtime Error"
	} else if strings.Contains(errorText, "timeout") {
		return "Timeout"
	} else if strings.Contains(errorText, "null pointer") || strings.Contains(errorText, "nil pointer") {
		return "Nil Pointer"
	} else if strings.Contains(errorText, "type assertion") {
		return "Type Error"
	} else {
		return "Other Error"
	}
}

func generateCompatibilityReport(groups []GroupSummary, summary TestSummary) {
	// Generate compatibility report markdown
	reportContent := fmt.Sprintf(`# JSONata-Go Compatibility Report

## Summary

- **Total Tests:** %d
- **Passed:** %d
- **Failed:** %d
- **Overall Compatibility:** %.1f%%

## Compatibility by Feature Group

| Feature Group | Compatibility | Tests | Passed | Failed |
|---------------|---------------|-------|--------|--------|
`, summary.TotalTests, summary.PassedTests, summary.FailedTests, summary.PassRate)

	for _, group := range groups {
		reportContent += fmt.Sprintf("| %s | %.1f%% | %d | %d | %d |\n",
			group.Name, group.PassRate, group.TotalTests, group.PassedTests, group.FailedTests)
	}

	reportContent += `
## Compatibility Status

### Well-Supported Features (>75% compatibility)
`
	for _, group := range groups {
		if group.PassRate >= 75 {
			reportContent += fmt.Sprintf("- %s (%.1f%%)\n", group.Name, group.PassRate)
		}
	}

	reportContent += `
### Partially-Supported Features (25-75% compatibility)
`
	for _, group := range groups {
		if group.PassRate >= 25 && group.PassRate < 75 {
			reportContent += fmt.Sprintf("- %s (%.1f%%)\n", group.Name, group.PassRate)
		}
	}

	reportContent += `
### Minimally-Supported Features (<25% compatibility)
`
	for _, group := range groups {
		if group.PassRate < 25 {
			reportContent += fmt.Sprintf("- %s (%.1f%%)\n", group.Name, group.PassRate)
		}
	}

	reportContent += `
## Next Steps

1. Focus on implementing missing functions in the high-impact feature groups
2. Address syntax differences between Go and JavaScript implementations
3. Fix common error patterns identified in the failure analysis

This report was automatically generated by the JSONata-Go dashboard tool.
`

	// Write the report to a file
	err := os.WriteFile("compatibility_report.md", []byte(reportContent), 0644)
	if err != nil {
		fmt.Println("Error writing compatibility report:", err)
		return
	}

	fmt.Println("\nCompatibility report generated: compatibility_report.md")
}