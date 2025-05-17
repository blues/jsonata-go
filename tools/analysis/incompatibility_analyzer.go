// incompatibility_analyzer.go - Tool to analyze specific incompatibilities in detail
// Run with: go run incompatibility_analyzer.go
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

type TestCase struct {
	Expr     string      `json:"expr"`
	Data     interface{} `json:"data"`
	Expected interface{} `json:"expected"`
	Dataset  string      `json:"dataset,omitempty"`
	Error    bool        `json:"error,omitempty"`
	Code     string      `json:"code,omitempty"`
}

type IncompatibilityIssue struct {
	Group       string
	Case        string
	Expression  string
	Error       string
	ErrorType   string
	Severity    string
	Suggestion  string
}

func main() {
	fmt.Println("JSONata-Go Incompatibility Analyzer")
	fmt.Println("==================================")

	// Load and parse all test cases
	testCases, err := loadTestCases()
	if err != nil {
		fmt.Println("Error loading test cases:", err)
		return
	}

	// Run Go test suite and capture detailed failures
	issues, err := analyzeFailures(testCases)
	if err != nil {
		fmt.Println("Error analyzing failures:", err)
		return
	}

	// Group and categorize issues
	categorizedIssues := categorizeIssues(issues)

	// Generate reports
	printIssueSummary(categorizedIssues)
	generateDetailedReport(categorizedIssues)
}

func loadTestCases() (map[string]TestCase, error) {
	fmt.Println("Loading test cases...")
	testCases := make(map[string]TestCase)

	// Find all test groups
	groupsDir := "./test-suite/groups"
	groups, err := os.ReadDir(groupsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading groups directory: %v", err)
	}

	for _, group := range groups {
		if !group.IsDir() {
			continue
		}

		groupName := group.Name()
		groupDir := filepath.Join(groupsDir, groupName)
		
		cases, err := os.ReadDir(groupDir)
		if err != nil {
			fmt.Printf("Warning: Could not read group directory %s: %v\n", groupDir, err)
			continue
		}

		for _, testCase := range cases {
			if !testCase.IsDir() && strings.HasSuffix(testCase.Name(), ".json") {
				caseName := strings.TrimSuffix(testCase.Name(), ".json")
				casePath := filepath.Join(groupDir, testCase.Name())
				
				data, err := os.ReadFile(casePath)
				if err != nil {
					fmt.Printf("Warning: Could not read test case %s: %v\n", casePath, err)
					continue
				}

				var tc TestCase
				if err := json.Unmarshal(data, &tc); err != nil {
					fmt.Printf("Warning: Could not parse test case %s: %v\n", casePath, err)
					continue
				}

				key := groupName + "_" + caseName
				testCases[key] = tc
			}
		}
	}

	fmt.Printf("Loaded %d test cases\n", len(testCases))
	return testCases, nil
}

func analyzeFailures(testCases map[string]TestCase) ([]IncompatibilityIssue, error) {
	fmt.Println("Analyzing test failures...")
	var issues []IncompatibilityIssue

	for testID, tc := range testCases {
		parts := strings.SplitN(testID, "_", 2)
		if len(parts) != 2 {
			continue
		}
		
		groupName := parts[0]
		caseName := parts[1]
		
		// Prepare test case file for analysis
		tmpDir, err := os.MkdirTemp("", "jsonata-analysis")
		if err != nil {
			return nil, fmt.Errorf("error creating temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		
		// Create a simplified test file that just runs this one test
		testFile := filepath.Join(tmpDir, "analyze_test.go")
		testCode := fmt.Sprintf(`package main

import (
	"testing"
	"path/filepath"
	jsonata "github.com/blues/jsonata-go"
)

func TestAnalyzeCase(t *testing.T) {
	// This test runs a single test case to analyze its failure
	expr := %q
	
	// Parse the expression
	e, err := jsonata.Compile(expr)
	if err != nil {
		t.Fatalf("Failed to compile expression: %%v", err)
	}
	
	// Try to evaluate (will typically fail for incompatible cases)
	_, err = e.Evaluate(nil)
	if err != nil {
		t.Logf("ANALYSIS: Expression evaluation failed: %%v", err)
	} else {
		t.Log("ANALYSIS: Expression evaluated successfully")
	}
}
`, tc.Expr)

		if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
			return nil, fmt.Errorf("error writing analysis test file: %v", err)
		}
		
		// Now run this test to see the detailed error
		errorOutput := runAnalysisTest(tmpDir)
		
		// Parse the error to create an incompatibility issue
		issue := IncompatibilityIssue{
			Group:      groupName,
			Case:       caseName,
			Expression: tc.Expr,
			Error:      extractErrorMessage(errorOutput),
			ErrorType:  categorizeError(errorOutput),
		}
		
		// Assign severity and suggestions
		classifyIssue(&issue)
		
		issues = append(issues, issue)
	}

	return issues, nil
}

func runAnalysisTest(testDir string) string {
	cmd := fmt.Sprintf("cd %s && go test -v", testDir)
	output, _ := exec.Command("bash", "-c", cmd).CombinedOutput()
	return string(output)
}

func extractErrorMessage(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ANALYSIS:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "ANALYSIS:"))
		}
	}
	return "Unknown error"
}

func categorizeError(output string) string {
	// Identify common error patterns
	if strings.Contains(output, "not implemented") || strings.Contains(output, "not found") {
		return "Missing Implementation"
	} else if strings.Contains(output, "syntax error") {
		return "Syntax Difference"
	} else if strings.Contains(output, "type") && strings.Contains(output, "mismatch") {
		return "Type Handling"
	} else if strings.Contains(output, "nil pointer") || strings.Contains(output, "null pointer") {
		return "Null Handling"
	} else if strings.Contains(output, "regex") || strings.Contains(output, "regular expression") {
		return "Regex Difference"
	} else {
		return "Other"
	}
}

func classifyIssue(issue *IncompatibilityIssue) {
	// Assign severity based on error type
	switch issue.ErrorType {
	case "Missing Implementation":
		issue.Severity = "Medium"
		issue.Suggestion = "Implement the missing function or operator"
	case "Syntax Difference":
		issue.Severity = "High"
		issue.Suggestion = "Update parser to handle this syntax pattern"
	case "Type Handling":
		issue.Severity = "Medium"
		issue.Suggestion = "Align type conversion/handling with JavaScript implementation"
	case "Null Handling":
		issue.Severity = "High"
		issue.Suggestion = "Fix null pointer issue and align null handling with spec"
	case "Regex Difference":
		issue.Severity = "Medium"
		issue.Suggestion = "Update regex implementation to match JavaScript behavior"
	default:
		issue.Severity = "Low"
		issue.Suggestion = "Investigate specific cause of failure"
	}
	
	// Check for high-priority functions
	if strings.Contains(issue.Group, "function-") {
		funcName := strings.TrimPrefix(issue.Group, "function-")
		if isCommonFunction(funcName) {
			issue.Severity = "High"
		}
	}
}

func isCommonFunction(funcName string) bool {
	commonFunctions := map[string]bool{
		"sum": true, "count": true, "max": true, "min": true,
		"average": true, "string": true, "number": true, "boolean": true,
		"exists": true, "length": true, "keys": true, "join": true,
		"map": true, "filter": true, "reduce": true,
	}
	return commonFunctions[funcName]
}

func categorizeIssues(issues []IncompatibilityIssue) map[string][]IncompatibilityIssue {
	categorized := make(map[string][]IncompatibilityIssue)
	
	for _, issue := range issues {
		categorized[issue.ErrorType] = append(categorized[issue.ErrorType], issue)
	}
	
	return categorized
}

func printIssueSummary(categorizedIssues map[string][]IncompatibilityIssue) {
	fmt.Println("\nIncompatibility Summary")
	fmt.Println("----------------------")
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Issue Type\tCount\tSeverity")
	fmt.Fprintln(w, "----------\t-----\t--------")
	
	// Get sorted keys
	var categories []string
	for category := range categorizedIssues {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	
	for _, category := range categories {
		issues := categorizedIssues[category]
		
		// Calculate common severity
		severityCount := map[string]int{
			"High": 0, "Medium": 0, "Low": 0,
		}
		
		for _, issue := range issues {
			severityCount[issue.Severity]++
		}
		
		var primarySeverity string
		if severityCount["High"] > 0 {
			primarySeverity = "High"
		} else if severityCount["Medium"] > 0 {
			primarySeverity = "Medium"
		} else {
			primarySeverity = "Low"
		}
		
		fmt.Fprintf(w, "%s\t%d\t%s\n", category, len(issues), primarySeverity)
	}
	
	w.Flush()
}

func generateDetailedReport(categorizedIssues map[string][]IncompatibilityIssue) {
	// Generate markdown report
	report := `# JSONata-Go Incompatibility Analysis

This report provides a detailed analysis of incompatibilities between JSONata-Go and the original JavaScript implementation.

## Summary of Incompatibility Types

| Issue Type | Count | Priority | Description |
|------------|-------|----------|-------------|
`
	
	// Get sorted keys
	var categories []string
	for category := range categorizedIssues {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	
	for _, category := range categories {
		issues := categorizedIssues[category]
		
		// Determine highest severity
		highestSeverity := "Low"
		for _, issue := range issues {
			if issue.Severity == "High" {
				highestSeverity = "High"
				break
			} else if issue.Severity == "Medium" && highestSeverity != "High" {
				highestSeverity = "Medium"
			}
		}
		
		description := getIssueTypeDescription(category)
		report += fmt.Sprintf("| %s | %d | %s | %s |\n", 
			category, len(issues), highestSeverity, description)
	}
	
	report += `
## High Priority Issues

These issues should be addressed first as they affect core functionality or commonly used features:

`
	
	for _, category := range categories {
		issues := categorizedIssues[category]
		
		// Find high severity issues in this category
		var highPriorityIssues []IncompatibilityIssue
		for _, issue := range issues {
			if issue.Severity == "High" {
				highPriorityIssues = append(highPriorityIssues, issue)
			}
		}
		
		if len(highPriorityIssues) > 0 {
			report += fmt.Sprintf("### %s\n\n", category)
			
			for i, issue := range highPriorityIssues {
				if i < 5 { // Limit to 5 examples per category to keep report manageable
					report += fmt.Sprintf("- **%s_%s**: `%s`\n  - %s\n  - Suggestion: %s\n\n", 
						issue.Group, issue.Case, issue.Expression, issue.Error, issue.Suggestion)
				} else if i == 5 {
					report += fmt.Sprintf("- _%d more issues in this category..._\n\n", len(highPriorityIssues)-5)
					break
				}
			}
		}
	}
	
	report += `
## Implementation Recommendations

Based on the analysis of incompatibilities, here are the recommended steps for improving compatibility:

1. **Address Function Implementation Gaps**
   - Implement missing core functions (particularly higher-order functions)
   - Ensure function signatures match the JavaScript implementation

2. **Fix Syntax Handling Differences**
   - Update parser to handle all valid JSONata syntax
   - Pay special attention to quoted field names and path handling

3. **Align Type Handling**
   - Ensure consistent type conversion behavior
   - Address issues with null/nil handling

4. **Improve Regular Expression Support**
   - Make regex behavior match JavaScript implementation
   - Fix pattern matching differences

This report was automatically generated by the JSONata-Go incompatibility_analyzer tool.
`

	// Write the report to a file
	err := os.WriteFile("incompatibility_analysis.md", []byte(report), 0644)
	if err != nil {
		fmt.Println("Error writing incompatibility report:", err)
		return
	}

	fmt.Println("\nIncompatibility analysis report generated: incompatibility_analysis.md")
}

func getIssueTypeDescription(category string) string {
	descriptions := map[string]string{
		"Missing Implementation": "Function or feature not yet implemented in Go",
		"Syntax Difference": "Syntax accepted in JavaScript but not in Go",
		"Type Handling": "Differences in how types are handled or converted",
		"Null Handling": "Differences in handling null/nil values",
		"Regex Difference": "Regular expression behavior differences",
		"Other": "Miscellaneous incompatibilities",
	}
	
	if desc, ok := descriptions[category]; ok {
		return desc
	}
	return "Unclassified issue type"
}