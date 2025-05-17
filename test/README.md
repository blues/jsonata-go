# JSONata Go Test Suite

This directory contains tests for the Go implementation of JSONata, including compatibility tests with the original JavaScript implementation.

## Overview

The tests in this directory serve several purposes:

1. Test the compatibility between the Go and JavaScript implementations
2. Test Go-specific implementation details
3. Provide examples of using JSONata in Go
4. Help identify areas for improvement in the Go implementation

## Test Organization

The test files are organized as follows:

- **test_suite_test.go**: Runs the standard JSONata test suite which was ported from the JavaScript implementation
- **implementation_test.go**: Tests Go-specific implementations of functions like `$millis()`, `$now()`, `$random()`, etc.
- **regex_test.go**: Tests regular expression handling in the Go implementation
- **transform_test.go**: Tests JSONata's data transformation capabilities
- **run_suite_test.go**: Simple examples of running individual test cases from the test suite

The `test-suite` directory contains the original test cases from the JavaScript implementation, organized by feature area.

## Analysis Tools

Analysis tools have been moved to `../tools/analysis` directory. These include:

- **dashboard.go**: A general-purpose dashboard showing overall compatibility
- **feature_matrix.go**: A detailed feature-by-feature compatibility matrix
- **incompatibility_analyzer.go**: A tool to analyze specific incompatibilities
- **analyze_results.go**: A simple tool to analyze test results

### Running an Analysis Tool

```bash
cd ../tools/analysis
go run dashboard.go
```

## Running the Tests

To run all tests:

```bash
cd test
go test -v
```

To run specific test files:

```bash
cd test
go test -v implementation_test.go transform_test.go regex_test.go
```

To run specific test functions:

```bash
cd test
go test -v -run TestImplementationFunctions
```

## Test Suite Results

When running the full test suite, you'll see a mix of passing and failing tests. This is expected, as the Go implementation has some differences from the JavaScript implementation:

- **PASS**: Tests that work correctly in the Go implementation
- **FAIL**: Tests where the Go implementation produces different results

See the COMPATIBILITY.md file for more details on compatibility status.

## Analyzing Compatibility

The repository includes several analysis tools to help identify compatibility issues:

```bash
cd ../tools/analysis
go run dashboard.go
go run feature_matrix.go
go run incompatibility_analyzer.go
```

These will generate reports showing:
- Overall compatibility percentage
- Feature-by-feature analysis
- Specific test cases that are failing
- Common patterns in compatibility issues
- Prioritized recommendations for implementation

The generated reports include:
- **compatibility_report.md**: General compatibility overview
- **feature_matrix.md**: Detailed feature-by-feature matrix
- **incompatibility_analysis.md**: Analysis of specific incompatibilities

## Adding New Tests

When adding new tests:

1. For general Go implementation testing, add test cases to the appropriate test file
2. For compatibility testing, add a Go test that runs against the test suite cases
3. For specific compatibility issues, create a targeted test case

### Adding a New Test Case to the Test Suite

To add a new test case to an existing group:
1. Add a new JSON file to the appropriate group in `test-suite/groups`
2. Follow the format of existing test cases:

```json
{
  "expr": "JSONata expression to evaluate",
  "data": "Input data (object or null)",
  "expected": "Expected result",
  "dataset": "Optional dataset name to use instead of data",
  "error": true/false,
  "code": "Optional error code if error is true"
}
```

## Test Data

The test suite uses JSON data files in the `test-suite/datasets` directory. When writing new tests, you can use this existing data or add new test data as needed.