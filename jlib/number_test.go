// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib_test

import (
	"fmt"
	"testing"

	"github.com/xiatechs/jsonata-go/jlib"
	"github.com/xiatechs/jsonata-go/jtypes"
)

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
