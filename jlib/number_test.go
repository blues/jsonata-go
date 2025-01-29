// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib_test

import (
	"fmt"
	"testing"

	"github.com/blues/jsonata-go/jlib"
	"github.com/blues/jsonata-go/jtypes"
)

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		// Binary numbers
		{"binary lowercase", "0b101", 5, false},
		{"binary uppercase", "0B1010", 10, false},
		{"binary long", "0b11111111", 255, false},
		{"binary with zeros", "0b00001010", 10, false},
		{"invalid binary", "0b102", 0, true},
		{"invalid binary chars", "0b1a1", 0, true},
		
		// Octal numbers
		{"octal lowercase", "0o12", 10, false},
		{"octal uppercase", "0O755", 493, false},
		{"octal long", "0o7777", 4095, false},
		{"octal with zeros", "0o0012", 10, false},
		{"invalid octal", "0o8", 0, true},
		{"invalid octal chars", "0o7a7", 0, true},
		
		// Hexadecimal numbers
		{"hex lowercase", "0x12", 18, false},
		{"hex uppercase", "0XFF", 255, false},
		{"hex mixed case", "0xDeadBeef", 3735928559, false},
		{"hex with zeros", "0x0012", 18, false},
		{"hex all letters", "0xabcdef", 11259375, false},
		{"invalid hex", "0xGG", 0, true},
		{"invalid hex chars", "0x12H4", 0, true},
		
		// Edge cases and special values
		{"empty string", "", 0, true},
		{"invalid prefix", "0k123", 0, true},
		{"just prefix binary", "0b", 0, true},
		{"just prefix octal", "0o", 0, true},
		{"just prefix hex", "0x", 0, true},
		{"boolean true", true, 1, false},
		{"boolean false", false, 0, false},
		{"float64", 3.14159, 3.14159, false},
		{"scientific notation positive", "1.23e-4", 0.000123, false},
		{"scientific notation negative", "-1.23e4", -12300, false},
		{"scientific notation uppercase", "1.23E+4", 12300, false},
		{"whitespace", " 42 ", 42, false},
		{"leading zeros", "00042", 42, false},
		{"negative zero", "-0", 0, false},
		{"negative number", "-42", -42, false},
		{"decimal point", "42.0", 42, false},
		{"multiple decimal points", "42.0.0", 0, true},
		{"invalid chars", "42abc", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jlib.Number(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Number() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Number() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatting(t *testing.T) {
	tests := []struct {
		name     string
		fn       string
		input    float64
		args     interface{}
		want     string
		wantErr  bool
	}{
		{"formatBase binary", "FormatBase", 42, jtypes.NewOptionalFloat64(2), "101010", false},
		{"formatBase octal", "FormatBase", 42, jtypes.NewOptionalFloat64(8), "52", false},
		{"formatBase hex", "FormatBase", 42, jtypes.NewOptionalFloat64(16), "2a", false},
		{"formatBase zero", "FormatBase", 0, jtypes.NewOptionalFloat64(2), "0", false},
		{"formatBase negative", "FormatBase", -42, jtypes.NewOptionalFloat64(2), "-101010", false},
		{"formatBase large number", "FormatBase", 65535, jtypes.NewOptionalFloat64(16), "ffff", false},
		{"formatBase invalid base", "FormatBase", 42, jtypes.NewOptionalFloat64(37), "", true},
		{"formatBase base too small", "FormatBase", 42, jtypes.NewOptionalFloat64(1), "", true},
		{"formatInteger basic", "FormatInteger", 42, "", "42", false},
		{"formatInteger negative", "FormatInteger", -42, "", "-42", false},
		{"formatInteger zero", "FormatInteger", 0, "", "0", false},
		{"formatInteger large", "FormatInteger", 1000000, "", "1000000", false},
		{"formatNumber basic", "FormatNumber", 3.14159, "", "3.14159", false},
		{"formatNumber negative", "FormatNumber", -3.14159, "", "-3.14159", false},
		{"formatNumber zero", "FormatNumber", 0, "", "0", false},
		{"formatNumber large", "FormatNumber", 1e6, "", "1000000", false},
		{"formatNumber small", "FormatNumber", 1e-6, "", "0.000001", false},
		{"formatNumber with picture", "FormatNumber", 12345.6789, "#,###.##", "12,345.68", false},
		{"formatNumber with currency", "FormatNumber", 12345.6789, "$#,###.00", "$12,345.68", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			var err error
			
			switch tt.fn {
			case "FormatBase":
				got, err = jlib.FormatBase(tt.input, tt.args.(jtypes.OptionalFloat64))
			case "FormatInteger":
				got, err = jlib.FormatInteger(tt.input, tt.args.(string))
			case "FormatNumber":
				got, err = jlib.FormatNumber(tt.input, tt.args.(string), jtypes.OptionalValue{})
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("%s() error = %v, wantErr %v", tt.fn, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("%s() = %v, want %v", tt.fn, got, tt.want)
			}
		})
	}
}

func TestParseInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		base    jtypes.OptionalFloat64
		want    float64
		wantErr bool
	}{
		{"binary", "101010", jtypes.NewOptionalFloat64(2), 42, false},
		{"binary uppercase", "101010", jtypes.NewOptionalFloat64(2), 42, false},
		{"binary with zeros", "000101", jtypes.NewOptionalFloat64(2), 5, false},
		{"octal", "52", jtypes.NewOptionalFloat64(8), 42, false},
		{"octal uppercase", "52", jtypes.NewOptionalFloat64(8), 42, false},
		{"octal with zeros", "00052", jtypes.NewOptionalFloat64(8), 42, false},
		{"decimal", "42", jtypes.NewOptionalFloat64(10), 42, false},
		{"decimal negative", "-42", jtypes.NewOptionalFloat64(10), -42, false},
		{"decimal with zeros", "00042", jtypes.NewOptionalFloat64(10), 42, false},
		{"hex", "2a", jtypes.NewOptionalFloat64(16), 42, false},
		{"hex uppercase", "2A", jtypes.NewOptionalFloat64(16), 42, false},
		{"hex with zeros", "002a", jtypes.NewOptionalFloat64(16), 42, false},
		{"invalid base", "42", jtypes.NewOptionalFloat64(37), 0, true},
		{"base too small", "42", jtypes.NewOptionalFloat64(1), 0, true},
		{"invalid digit", "2a", jtypes.NewOptionalFloat64(10), 0, true},
		{"empty string", "", jtypes.NewOptionalFloat64(10), 0, true},
		{"whitespace", " 42 ", jtypes.NewOptionalFloat64(10), 42, false},
		{"invalid chars", "42abc", jtypes.NewOptionalFloat64(10), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jlib.ParseInteger(tt.input, tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMathFuncs(t *testing.T) {
	tests := []struct {
		name string
		fn   func(float64) float64
		x    float64
		want float64
	}{
		{"abs positive", jlib.Abs, 42, 42},
		{"abs negative", jlib.Abs, -42, 42},
		{"abs zero", jlib.Abs, 0, 0},
		{"ceil up", jlib.Ceil, 3.14, 4},
		{"ceil whole", jlib.Ceil, 42, 42},
		{"ceil negative", jlib.Ceil, -3.14, -3},
		{"floor down", jlib.Floor, 3.14, 3},
		{"floor whole", jlib.Floor, 42, 42},
		{"floor negative", jlib.Floor, -3.14, -4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.x); got != tt.want {
				t.Errorf("%v(%v) = %v, want %v", tt.name, tt.x, got, tt.want)
			}
		})
	}
}

func TestRound(t *testing.T) {

	data := []struct {
		Value     float64
		Precision jtypes.OptionalInt
		Output    float64
	}{
		{
			Value:  11.5,
			Output: 12,
		},
		{
			Value:  -11.5,
			Output: -12,
		},
		{
			Value:  12.5,
			Output: 12,
		},
		{
			Value:  -12.5,
			Output: -12,
		},
		{
			Value:  594.325,
			Output: 594,
		},
		{
			Value:  -594.325,
			Output: -594,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(1),
			Output:    594.3,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(1),
			Output:    -594.3,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(2),
			Output:    594.32,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(2),
			Output:    -594.32,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(3),
			Output:    594.325,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(3),
			Output:    -594.325,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(4),
			Output:    594.325,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(4),
			Output:    -594.325,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(-1),
			Output:    590,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(-1),
			Output:    -590,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(-2),
			Output:    600,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(-2),
			Output:    -600,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(-3),
			Output:    1000,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(-3),
			Output:    -1000,
		},
		{
			Value:     594.325,
			Precision: jtypes.NewOptionalInt(-4),
			Output:    0,
		},
		{
			Value:     -594.325,
			Precision: jtypes.NewOptionalInt(-4),
			Output:    0,
		},
	}

	for _, test := range data {

		got := jlib.Round(test.Value, test.Precision)

		if got != test.Output {

			s := fmt.Sprintf("round(%g", test.Value)
			if test.Precision.IsSet() {
				s += fmt.Sprintf(", %d", test.Precision.Int)
			}
			s += ")"

			t.Errorf("%s: Expected %g, got %g", s, test.Output, got)
		}
	}
}
