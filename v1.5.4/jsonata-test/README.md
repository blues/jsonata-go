# JSONata Test

A CLI tool for running jsonata-go against the [JSONata test suite](https://github.com/jsonata-js/jsonata/tree/master/test/test-suite).

## Install

    go install github.com/blues/jsonata-test

## Usage

1. Clone the [jsonata-js](https://github.com/jsonata-js/jsonata) repository to your local machine.

    git clone https://github.com/jsonata-js/jsonata

2. To access a particular version of the test suite, check out the relevant branch, e.g.

    git checkout v1.8.4

3. Run the test tool, specifying the location of the JSONata `test-suite` directory in the command line, e.g.

    jsonata-test ~/projects/jsonata/test/test-suite

## Known issues

This library was originally developed against jsonata-js 1.5 and has thus far implemented a subset of features from newer version of that library. You can see potential differences by looking at the [jsonata-js changelog](https://github.com/jsonata-js/jsonata/blob/master/CHANGELOG.md).

While most tests pass from jsonata-js 1.5 do pass, currently there are **1598 tests** in the JSONata 1.8.4 test suite. Running against the 1.8.4 test suite results in **307 failing tests**. The failures are mostly related to functionality in newer versions of JSONata that this library does not yet implement. The outstanding issues are summarised below, split into the categories "Won't fix", "To be fixed" and "To be investigated".

### Won't Fix

#### Regex matches on zero-length strings

jsonata-js throws an error if a regular expression matches a zero length string. It does this because repeatedly calling JavaScript's [Regexp.exec](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp/exec) method can cause an infinite loop if it matches a zero length string. Go's regex handling doesn't have this problem so there's no real need to take the precaution.

### To be investigated

#### Null handling

jsonata-go uses `*interface{}(nil)` to represent the JSON value null. This is a nil value but it's distinguishable from `interface{}(nil)` (non-pointer) which indicates that a value does not exist. That's useful inside JSONata but Go's json package does not make that distinction. Some tests fail because jsonata-go returns a differently-typed nil. I don't *think* this will cause any practical problems (because the returned value still equals nil) but I'd like to look into it more.

#### Rounding when converting numbers to strings

JSONata's `$string()` function converts objects to their JSON representation. In jsonata-js, any numbers in the object are rounded to 15 decimal places so that floating point errors are discarded. The rounding takes place in a callback function passed to JavaScript's JSON encoding function. I haven't found a way to replicate this in Go. It would be a nice feature to have though.

### To be fixed
Some functions like `$formatInteger()` and `$parseInteger()` from newer versions of the jsonata-js library are not yet implemented.
