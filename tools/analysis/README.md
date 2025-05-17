# JSONata-Go Analysis Tools

This directory contains tools for analyzing the compatibility between the Go and JavaScript implementations of JSONata.

## Available Tools

- **dashboard.go**: A general-purpose dashboard showing overall compatibility
- **feature_matrix.go**: A detailed feature-by-feature compatibility matrix
- **incompatibility_analyzer.go**: A tool to analyze specific incompatibilities

## Running the Tools

```bash
# From the tools/analysis directory
go run dashboard.go
go run feature_matrix.go
go run incompatibility_analyzer.go
```

## Generated Reports

These tools generate various reports:

- **compatibility_report.md**: General compatibility overview
- **feature_matrix.md**: Detailed feature-by-feature matrix
- **incompatibility_analysis.md**: Analysis of specific incompatibilities

## Usage Notes

When running these tools:

1. They must be run from the root of the repository or with the working directory set to access the test suite files.
2. Some tools may take a few minutes to complete as they run the entire test suite.
3. Reports are generated in the current working directory.