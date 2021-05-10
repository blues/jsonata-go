// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jxpath

import (
	"reflect"
	"testing"
)

type formatNumberTest struct {
	Value   float64
	Picture string
	Output  string
	Error   error
}

func TestExamples(t *testing.T) {

	tests := []formatNumberTest{
		{
			Value:   12345.6,
			Picture: "#,###.00",
			Output:  "12,345.60",
		},
		{
			Value:   12345678.9,
			Picture: "9,999.99",
			Output:  "12,345,678.90",
		},
		{
			Value:   123.9,
			Picture: "9999",
			Output:  "0124",
		},
		{
			Value:   0.14,
			Picture: "01%",
			Output:  "14%",
		},
		{
			Value:   0.14,
			Picture: "001‰",
			Output:  "140‰",
		},
		{
			Value:   -6,
			Picture: "000",
			Output:  "-006",
		},
		{
			Value:   1234.5678,
			Picture: "#,##0.00",
			Output:  "1,234.57",
		},
		{
			Value:   1234.5678,
			Picture: "00.000e0",
			Output:  "12.346e2",
		},
		{
			Value:   0.234,
			Picture: "0.0e0",
			Output:  "2.3e-1",
		},
		{
			Value:   0.234,
			Picture: "#.00e0",
			Output:  "0.23e0",
		},
		{
			Value:   0.234,
			Picture: ".00e0",
			Output:  ".23e0",
		},
	}

	testFormatNumber(t, tests)
}

func testFormatNumber(t *testing.T, tests []formatNumberTest) {

	df := NewDecimalFormat()

	for i, test := range tests {

		output, err := FormatNumber(test.Value, test.Picture, df)

		if output != test.Output {
			t.Errorf("%d. FormatNumber(%v, %q): expected %s, got %s", i+1, test.Value, test.Picture, test.Output, output)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%d. FormatNumber(%v, %q): expected error %v, got %v", i+1, test.Value, test.Picture, test.Error, err)
		}
	}
}
