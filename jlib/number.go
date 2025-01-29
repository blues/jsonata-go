// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/blues/jsonata-go/jtypes"
)

var (
	reNumber = regexp.MustCompile(`^-?(([0-9]+))(\.[0-9]+)?([Ee][-+]?[0-9]+)?$`)
	reBinary = regexp.MustCompile(`^0[bB][01]+$`)
	reOctal  = regexp.MustCompile(`^0[oO][0-7]+$`)
	reHex    = regexp.MustCompile(`^0[xX][0-9a-fA-F]+$`)
)

// Number converts values to numbers. Numeric values are returned
// unchanged. Strings in legal JSON number format are converted
// to the number they represent. Booleans are converted to 0 or 1.
// All other types trigger an error.
func Number(value interface{}) (float64, error) {
	v := reflect.ValueOf(value)
	if b, ok := jtypes.AsBool(v); ok {
		if b {
			return 1, nil
		}
		return 0, nil
	}

	if n, ok := jtypes.AsNumber(v); ok {
		return n, nil
	}

	s, ok := jtypes.AsString(v)
	if !ok {
		return 0, fmt.Errorf("unable to cast value to a number")
	}
	s = strings.TrimSpace(s)

	if reBinary.MatchString(s) {
		n, _ := strconv.ParseInt(s[2:], 2, 64)
		return float64(n), nil
	}
	if reOctal.MatchString(s) {
		n, _ := strconv.ParseInt(s[2:], 8, 64)
		return float64(n), nil
	}
	if reHex.MatchString(s) {
		n, _ := strconv.ParseInt(s[2:], 16, 64)
		return float64(n), nil
	}
	if reNumber.MatchString(s) {
		if n, err := strconv.ParseFloat(s, 64); err == nil {
			return n, nil
		}
	}

	return 0, fmt.Errorf("unable to cast %q to a number", s)
}

// Round rounds its input to the number of decimal places given
// in the optional second parameter. By default, Round rounds to
// the nearest integer. A negative precision specifies which column
// to round to on the left hand side of the decimal place.
func Round(x float64, prec jtypes.OptionalInt) float64 {
	// Adapted from gonum's floats.RoundEven.
	// https://github.com/gonum/gonum/tree/master/floats

	if x == 0 {
		// Make sure zero is returned
		// without the negative bit set.
		return 0
	}
	// Fast path for positive precision on integers.
	if prec.Int >= 0 && x == math.Trunc(x) {
		return x
	}
	intermed := multByPow10(x, prec.Int)
	if math.IsInf(intermed, 0) {
		return x
	}
	if isHalfway(intermed) {
		correction, _ := math.Modf(math.Mod(intermed, 2))
		intermed += correction
		if intermed > 0 {
			x = math.Floor(intermed)
		} else {
			x = math.Ceil(intermed)
		}
	} else {
		if x < 0 {
			x = math.Ceil(intermed - 0.5)
		} else {
			x = math.Floor(intermed + 0.5)
		}
	}

	if x == 0 {
		return 0
	}

	return multByPow10(x, -prec.Int)
}

// Power returns x to the power of y.
func Power(x, y float64) (float64, error) {
	res := math.Pow(x, y)
	if math.IsInf(res, 0) || math.IsNaN(res) {
		return 0, fmt.Errorf("the power function has resulted in a value that cannot be represented as a JSON number")
	}
	return res, nil
}

// Sqrt returns the square root of a number. It returns an error
// if the number is less than zero.
func Sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, fmt.Errorf("the sqrt function cannot be applied to a negative number")
	}
	return math.Sqrt(x), nil
}

// Random returns a random floating point number between 0 and 1.
func Random() float64 {
	return rand.Float64()
}

// Abs returns the absolute value of x.
func Abs(x float64) float64 {
	return math.Abs(x)
}

// Ceil returns the least integer value greater than or equal to x.
func Ceil(x float64) float64 {
	return math.Ceil(x)
}

// Floor returns the greatest integer value less than or equal to x.
func Floor(x float64) float64 {
	return math.Floor(x)
}

// FormatBase formats a number using the specified base (2-36).
func FormatBase(x float64, base jtypes.OptionalFloat64) (string, error) {
	radix := 10
	if base.IsSet() {
		radix = int(Round(base.Float64, jtypes.OptionalInt{}))
	}

	if radix < 2 || radix > 36 {
		return "", fmt.Errorf("the second argument to formatBase must be between 2 and 36")
	}
	n := int64(Round(x, jtypes.OptionalInt{}))
	return strconv.FormatInt(n, radix), nil
}

// FormatInteger formats an integer using the specified picture string.
func FormatInteger(x float64, picture string) (string, error) {
	if picture == "" {
		return strconv.FormatInt(int64(Round(x, jtypes.OptionalInt{})), 10), nil
	}
	return formatNumberWithPicture(x, picture, jtypes.OptionalValue{})
}

// FormatNumber formats a number using the specified picture string and options.
func FormatNumber(x float64, picture string, options jtypes.OptionalValue) (string, error) {
	if picture == "" {
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	}
	return formatNumberWithPicture(x, picture, options)
}

// ParseInteger parses a string as an integer using the specified base.
func ParseInteger(value interface{}, base jtypes.OptionalFloat64) (float64, error) {
	s, ok := jtypes.AsString(reflect.ValueOf(value))
	if !ok {
		return 0, fmt.Errorf("first argument of parseInteger must be a string")
	}
	s = strings.TrimSpace(s)

	radix := 10
	if base.IsSet() {
		radix = int(Round(base.Float64, jtypes.OptionalInt{}))
	}

	if radix < 0 || radix == 1 || radix > 36 {
		return 0, fmt.Errorf("invalid base: %d", radix)
	}

	n, err := strconv.ParseInt(s, radix, 64)
	if err != nil {
		return 0, err
	}
	return float64(n), nil
}

// multByPow10 multiplies a number by 10 to the power of n.
// It does this by converting back and forth to strings to
// avoid floating point rounding errors, e.g.
//
//	4.525 * math.Pow10(2) returns 452.50000000000006
func multByPow10(x float64, n int) float64 {
	if n == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	s := fmt.Sprintf("%g", x)

	chunks := strings.Split(s, "e")
	switch len(chunks) {
	case 1:
		s = chunks[0] + "e" + strconv.Itoa(n)
	case 2:
		e, _ := strconv.Atoi(chunks[1])
		s = chunks[0] + "e" + strconv.Itoa(e+n)
	default:
		return x
	}

	x, _ = strconv.ParseFloat(s, 64)
	return x
}

func isHalfway(x float64) bool {
	_, frac := math.Modf(x)
	frac = math.Abs(frac)
	return frac == 0.5 || (math.Nextafter(frac, math.Inf(-1)) < 0.5 && math.Nextafter(frac, math.Inf(1)) > 0.5)
}
