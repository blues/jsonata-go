# JSONata Go Compatibility Testing

This directory contains tests for verifying the compatibility between the Go implementation of JSONata and the original JavaScript implementation.

## Test Files

The test files in this directory include:

1. **test_suite_test.go** - Runs the standard JSONata test suite from the original JavaScript implementation
2. **implementation_test.go** - Tests Go-specific implementations of built-in functions
3. **regex_test.go** - Tests regular expression handling in the Go implementation
4. **transform_test.go** - Tests JSONata's data transformation capabilities
5. **run_suite_test.go** - Simple examples of running individual test cases

## Compatibility Status

The Go implementation of JSONata is compatible with most of the core features of JSONata, but there are some differences and limitations compared to the JavaScript implementation:

### Well-supported Features:

- Basic path expressions (e.g., `Account.Order.Product.Price`)
- Array operations and navigation
- Basic mathematical operations
- Basic string functions
- Boolean expressions and operators
- Function application
- Object construction
- Most core library functions

### Differences from JavaScript Implementation:

1. **Field Name Syntax**: For field names with spaces or special characters:
   - JavaScript: Uses double quotes `"Product Name"`
   - Go: Uses backticks `` `Product Name` ``

2. **Regular Expressions**: The Go implementation handles regular expressions differently:
   - Slightly different match object structure
   - Different handling of regex flags and capturing groups

3. **Function Definitions**: Some advanced function definition features may work differently

### Known Limitations:

Some features from the JavaScript implementation that may not be fully supported:

- Certain advanced closure mechanisms
- Some higher-order function patterns
- Certain error handling behaviors
- Some advanced transform and wildcard patterns

## Running the Tests

To run all the compatibility tests:

```
cd test
go test -v
```

To run just the custom tests that are known to pass:

```
cd test
go test -v implementation_test.go transform_test.go regex_test.go run_suite_test.go
```

To run just the test suite compatibility tests:

```
cd test
go test -v test_suite_test.go
```

## Test Suite Results

When running the test suite tests, you'll see two main outcomes:

- **PASS**: Tests that pass, indicating compatibility
- **FAIL**: Tests that fail, indicating implementation differences 

The failures typically fall into these categories:

1. **Syntax Differences**: The Go implementation uses backtick syntax (`` `Field Name` ``) instead of quoted field names (`"Field Name"`)
2. **Missing Features**: Features not yet implemented in the Go version
3. **Behavioral Differences**: Cases where the Go implementation behaves differently
4. **Error Handling**: Differences in how errors are reported and handled

Each test failure provides detailed information about what went wrong, making it easier to identify compatibility issues.

## Working with Field Names

When writing JSONata expressions in Go that access fields with spaces or special characters, use backtick syntax:

```go
// JavaScript:
expr, err := jsonata.Compile(`Account."Product Name"`)

// Go:
expr, err := jsonata.Compile("Account.`Product Name`")
```