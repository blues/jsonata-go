# JSONata Go Compatibility Analysis

## Summary by Feature Group

| Group | Total Tests | Passed | Pass Rate |
|-------|-------------|--------|-----------|
| array-constructor | 21 | 15 | 71.4% |
| blocks | 7 | 3 | 42.9% |
| boolean-expresssions | 17 | 14 | 82.4% |
| closures | 2 | 0 | 0.0% |
| comparison-operators | 25 | 24 | 96.0% |
| conditionals | 9 | 3 | 33.3% |
| context | 4 | 0 | 0.0% |
| descendent-operator | 18 | 0 | 0.0% |
| encoding | 4 | 0 | 0.0% |
| errors | 25 | 20 | 80.0% |
| fields | 8 | 8 | 100.0% |
| flattening | 42 | 35 | 83.3% |
| function-abs | 4 | 0 | 0.0% |
| function-append | 6 | 0 | 0.0% |
| function-applications | 22 | 0 | 0.0% |
| function-average | 13 | 0 | 0.0% |
| function-boolean | 24 | 0 | 0.0% |
| function-ceil | 4 | 0 | 0.0% |
| function-contains | 7 | 0 | 0.0% |
| function-count | 14 | 0 | 0.0% |
| function-each | 1 | 0 | 0.0% |
| function-exists | 25 | 0 | 0.0% |
| function-floor | 4 | 0 | 0.0% |
| function-formatBase | 8 | 0 | 0.0% |
| function-formatNumber | 37 | 0 | 0.0% |
| function-fromMillis | 3 | 0 | 0.0% |
| function-join | 12 | 0 | 0.0% |
| function-keys | 7 | 0 | 0.0% |
| function-length | 17 | 0 | 0.0% |
| function-lookup | 4 | 0 | 0.0% |
| function-lowercase | 2 | 0 | 0.0% |
| function-max | 27 | 0 | 0.0% |
| function-merge | 5 | 0 | 0.0% |
| function-number | 27 | 0 | 0.0% |
| function-pad | 11 | 0 | 0.0% |
| function-power | 7 | 0 | 0.0% |
| function-replace | 12 | 0 | 0.0% |
| function-reverse | 4 | 0 | 0.0% |
| function-round | 18 | 0 | 0.0% |
| function-shuffle | 4 | 0 | 0.0% |
| function-sift | 3 | 0 | 0.0% |
| function-signatures | 35 | 0 | 0.0% |
| function-sort | 11 | 0 | 0.0% |
| function-split | 19 | 0 | 0.0% |
| function-spread | 4 | 0 | 0.0% |
| function-sqrt | 4 | 0 | 0.0% |
| function-string | 23 | 0 | 0.0% |
| function-substring | 19 | 0 | 0.0% |
| function-substringAfter | 5 | 0 | 0.0% |
| function-substringBefore | 5 | 0 | 0.0% |
| function-sum | 7 | 0 | 0.0% |
| function-tomillis | 9 | 0 | 0.0% |
| function-trim | 3 | 0 | 0.0% |
| function-uppercase | 2 | 0 | 0.0% |
| function-zip | 6 | 0 | 0.0% |
| higher-order-functions | 3 | 0 | 0.0% |
| hof-filter | 3 | 0 | 0.0% |
| hof-map | 9 | 0 | 0.0% |
| hof-reduce | 9 | 0 | 0.0% |
| hof-zip-map | 4 | 0 | 0.0% |
| inclusion-operator | 9 | 8 | 88.9% |
| lambdas | 13 | 0 | 0.0% |
| literals | 18 | 18 | 100.0% |
| matchers | 2 | 0 | 0.0% |
| missing-paths | 6 | 5 | 83.3% |
| multiple-array-selectors | 3 | 3 | 100.0% |
| null | 7 | 6 | 85.7% |
| numeric-operators | 17 | 17 | 100.0% |
| object-constructor | 24 | 11 | 45.8% |
| parentheses | 8 | 8 | 100.0% |
| partial-application | 5 | 2 | 40.0% |
| predicates | 4 | 2 | 50.0% |
| quoted-selectors | 8 | 4 | 50.0% |
| range-operator | 11 | 9 | 81.8% |
| regex | 37 | 0 | 0.0% |
| simple-array-selectors | 23 | 17 | 73.9% |
| sorting | 20 | 11 | 55.0% |
| string-concat | 12 | 12 | 100.0% |
| tail-recursion | 10 | 0 | 0.0% |
| token-conversion | 4 | 0 | 0.0% |
| transform | 104 | 23 | 22.1% |
| transforms | 12 | 0 | 0.0% |
| variables | 13 | 0 | 0.0% |
| wildcards | 10 | 8 | 80.0% |
|-------|-------------|--------|-----------|
| **TOTAL** | **1074** | **286** | **26.6%** |

## Overall Compatibility Status

The Go implementation of JSONata currently has **limited compatibility** with the JavaScript version, passing about 26.6% of the test suite. However, this percentage is slightly misleading as the analysis shows certain core features are well-supported while others are not implemented yet.

## Well-Supported Features (75-100%)

The following features have high compatibility:

- Literals (100%)
- Multiple Array Selectors (100%)
- Numeric Operators (100%)
- Parentheses (100%)
- String Concatenation (100%)
- Comparison Operators (96%)
- Inclusion Operator (89%)
- Null Handling (86%)
- Flattening (83%)
- Boolean Expressions (82%)
- Range Operator (82%)
- Wildcards (80%)
- Error Handling (80%)

## Moderate Support (40-75%)

These features have moderate compatibility:

- Simple Array Selectors (74%)
- Array Constructor (71%)
- Sorting (55%)
- Quoted Selectors (50%)
- Predicates (50%)
- Object Constructor (46%)
- Blocks (43%)
- Partial Application (40%)

## Limited or No Support (0-40%)

These areas have limited or no compatibility:

- Conditionals (33%)
- Transform (22%)
- Functions (most functions at 0%)
- Higher-Order Functions (0%)
- Lambdas (0%)
- Closures (0%)
- Context Variables (0%)
- Regular Expressions (0%)
- Descendant Operator (0%)

## Key Compatibility Challenges

1. **Function Support**: Most of the JSONata functions are not implemented yet in the Go version
2. **Advanced Features**: Higher-order functions, lambdas, and transforms have limited support
3. **Regular Expressions**: Regular expression handling differs between Go and JavaScript
4. **Syntax Differences**: Handling of quoted field names and other syntax elements varies
5. **Context Variables**: The $ context variable and context handling work differently

## Recommendations for Improving Compatibility

1. **Implement Core Functions**: Add implementations for commonly used JSONata functions
2. **Regular Expression Support**: Improve regex compatibility with JavaScript behavior
3. **Quoted Field Names**: Enhanced handling of quoted field names in path expressions
4. **Context Variables**: Better support for context variables and manipulations
5. **Transform Operations**: Improve transform operator (~>) implementations

## Note on Compatibility Testing

The compatibility score is based on automated testing of the JSONata test suite. Some tests may fail due to syntactic differences rather than functional differences, and the actual compatibility in real-world usage may be higher for common use cases.