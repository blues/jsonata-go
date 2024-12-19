// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jxpath

import (
	"reflect"
	"testing"
	"time"
)

func TestFormatYear(t *testing.T) {
	input := time.Date(2018, time.April, 1, 12, 0, 0, 0, time.UTC)

	data := []struct {
		Picture string
		Output  string
		Error   error
	}{
		{
			// Default layout is 1.
			Picture: "[Y]",
			Output:  "2018",
		},
		{
			Picture: "[Y0001] [M01] [D01]",
			Output:  "2018 04 01",
		},
		{
			Picture: "[Y1]",
			Output:  "2018",
		},
		{
			Picture: "[Y01]",
			Output:  "18",
		},
		{
			Picture: "[Y001]",
			Output:  "018",
		},
		{
			Picture: "[Y0001]",
			Output:  "2018",
		},
		{
			Picture: "[Y9,999,*]",
			Output:  "2,018",
		},
		{
			Picture: "[Y1,*-1]",
			Output:  "8",
		},
		{
			Picture: "[Y1,*-2]",
			Output:  "18",
		},
		/*{
			Picture: "[Y1,*-3]",
			Output:  "018",
		},*/
		{
			Picture: "[Y1,*-4]",
			Output:  "2018",
		},
		{
			// Unsupported layouts should fall back to the default.
			Picture: "[YNn]",
			Output:  "2018",
		},
	}

	for _, test := range data {

		got, err := FormatTime(input, test.Picture)

		if got != test.Output {
			t.Errorf("%s: Expected %q, got %q", test.Picture, test.Output, got)
		}

		if !reflect.DeepEqual(err, test.Error) {
			t.Errorf("%s: Expected error %v, got %v", test.Picture, test.Error, err)
		}
	}
}

func TestFormatYearAndTimezone(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Rome")
	input := time.Date(2018, time.April, 1, 12, 0, 0, 0, location)

	picture := "[Y0001]-[M01]-[D01] [H01]:[m01]:[s01] [P]"

	got, err := FormatTime(input, picture)
	if err != nil {
		t.Errorf("unable to format time %+v", err)
	}
	if got != "2018-04-01 12:00:00 pm" {
		t.Errorf("got %s expected %s", got, "2018-04-01 12:00:00 pm")
	}
}

func TestFormatTimezone(t *testing.T) {
	const minutes = 60
	const hours = 60 * minutes

	timezones := []struct {
		Name   string
		Offset int
	}{
		{
			Name:   "HST",
			Offset: -10 * hours,
		},
		{
			Name:   "EST",
			Offset: -5 * hours,
		},
		{
			Name:   "GMT",
			Offset: 0,
		},
		{
			Name:   "IST",
			Offset: 5*hours + 30*minutes,
		},
		{
			Offset: 13 * hours,
		},
	}

	times := make([]time.Time, len(timezones))
	for i, tz := range timezones {
		// We're mostly interested in the timezone for these tests
		// so the exact date used here doesn't matter. But the time
		// must be 12:00 for the final test case (which also outputs
		// the time) to work.
		times[i] = time.Date(2018, time.April, 1, 12, 0, 0, 0, time.FixedZone(tz.Name, tz.Offset))
	}

	data := []struct {
		Picture  string
		Location *time.Location
		Outputs  []string
	}{
		{
			// Default layout is 00:00.
			Picture: "[Z]",
			Outputs: []string{
				"-10:00",
				"-05:00",
				"+00:00",
				"+05:30",
				"+13:00",
			},
		},
		{
			Picture: "[Z0]",
			Outputs: []string{
				"-10",
				"-5",
				"+0",
				"+5:30",
				"+13",
			},
		},
		{
			Picture: "[Z00]",
			Outputs: []string{
				"-10",
				"-05",
				"+00",
				"+05:30",
				"+13",
			},
		},
		{
			Picture: "[Z00t]",
			Outputs: []string{
				"-10",
				"-05",
				"Z",
				"+05:30",
				"+13",
			},
		},
		{
			Picture: "[Z000]",
			Outputs: []string{
				"-1000",
				"-500",
				"+000",
				"+530",
				"+1300",
			},
		},
		{
			Picture: "[Z0000]",
			Outputs: []string{
				"-1000",
				"-0500",
				"+0000",
				"+0530",
				"+1300",
			},
		},
		{
			Picture: "[Z0000t]",
			Outputs: []string{
				"-1000",
				"-0500",
				"Z",
				"+0530",
				"+1300",
			},
		},
		{
			Picture: "[Z0:00]",
			Outputs: []string{
				"-10:00",
				"-5:00",
				"+0:00",
				"+5:30",
				"+13:00",
			},
		},
		{
			Picture: "[Z00:00]",
			Outputs: []string{
				"-10:00",
				"-05:00",
				"+00:00",
				"+05:30",
				"+13:00",
			},
		},
		{
			Picture: "[Z00:00t]",
			Outputs: []string{
				"-10:00",
				"-05:00",
				"Z",
				"+05:30",
				"+13:00",
			},
		},
		{
			Picture: "[z]",
			Outputs: []string{
				"GMT-10:00",
				"GMT-05:00",
				"GMT+00:00",
				"GMT+05:30",
				"GMT+13:00",
			},
		},
		{
			Picture: "[ZZ]",
			Outputs: []string{
				"W",
				"R",
				"Z",
				"+05:30", // military layout only supports whole hour offsets, fall back to the default
				"+13:00", // military layout only supports offsets up to 12 hours, fall back to the default
			},
		},
		{
			Picture: "[ZN]",
			Outputs: []string{
				"HST",
				"EST",
				"GMT",
				"IST",
				"+13:00", // timezone has no name, fall back to the default layout
			},
		},
		{
			Picture:  "[H00]:[m00] [ZN]",
			Location: time.FixedZone("EST", -5*hours),
			Outputs: []string{
				"17:00 EST", // Note: The XPath docs (incorrectly) have this as 06:00 EST
				"12:00 EST",
				"07:00 EST",
				"01:30 EST",
				"18:00 EST",
			},
		},
	}

	for _, test := range data {
		for i, tm := range times {

			if test.Location != nil {
				tm = tm.In(test.Location)
			}

			got, err := FormatTime(tm, test.Picture)

			if got != test.Outputs[i] {
				t.Errorf("%s: Expected %q, got %q", test.Picture, test.Outputs[i], got)
			}

			if err != nil {
				t.Errorf("%s: Expected nil error, got %s", test.Picture, err)
			}
		}
	}
}

func TestFormatDayOfWeek(t *testing.T) {

	startTime := time.Date(2018, time.April, 1, 12, 0, 0, 0, time.UTC)

	var times [7]time.Time
	for i := range times {
		times[i] = startTime.AddDate(0, 0, i)
	}

	data := []struct {
		Picture string
		Outputs [7]string
	}{
		{
			// Default layout is n
			Picture: "[F]",
			Outputs: [7]string{
				"sunday",
				"monday",
				"tuesday",
				"wednesday",
				"thursday",
				"friday",
				"saturday",
			},
		},
		{
			Picture: "[FNn]",
			Outputs: [7]string{
				"Sunday",
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
			},
		},
		{
			Picture: "[FNn,*-6]",
			Outputs: [7]string{
				"Sunday",
				"Monday",
				"Tues",
				"Weds",
				"Thurs",
				"Friday",
				"Sat",
			},
		},
		{
			Picture: "[FNn,6-6]",
			Outputs: [7]string{
				"Sunday",
				"Monday",
				"Tues  ",
				"Weds  ",
				"Thurs ",
				"Friday",
				"Sat   ",
			},
		},
		{
			Picture: "[FNn,*-5]",
			Outputs: [7]string{
				"Sun",
				"Mon",
				"Tues",
				"Weds",
				"Thurs",
				"Fri",
				"Sat",
			},
		},
		{
			Picture: "[FNn,*-4]",
			Outputs: [7]string{
				"Sun",
				"Mon",
				"Tues",
				"Weds",
				"Thur",
				"Fri",
				"Sat",
			},
		},
		{
			Picture: "[FN,*-3]",
			Outputs: [7]string{
				"SUN",
				"MON",
				"TUE",
				"WED",
				"THU",
				"FRI",
				"SAT",
			},
		},
		{
			Picture: "[FNn,*-2]",
			Outputs: [7]string{
				"Su",
				"Mo",
				"Tu",
				"We",
				"Th",
				"Fr",
				"Sa",
			},
		},
		{
			Picture: "Day [F01]: [FNn]",
			Outputs: [7]string{
				"Day 01: Sunday",
				"Day 02: Monday",
				"Day 03: Tuesday",
				"Day 04: Wednesday",
				"Day 05: Thursday",
				"Day 06: Friday",
				"Day 07: Saturday",
			},
		},
		{
			Picture: "[FNn] is the [F1o] day of the week",
			Outputs: [7]string{
				"Sunday is the 1st day of the week",
				"Monday is the 2nd day of the week",
				"Tuesday is the 3rd day of the week",
				"Wednesday is the 4th day of the week",
				"Thursday is the 5th day of the week",
				"Friday is the 6th day of the week",
				"Saturday is the 7th day of the week",
			},
		},
		{
			// Unsupported layouts should fall back to the default.
			Picture: "[FI]",
			Outputs: [7]string{
				"sunday",
				"monday",
				"tuesday",
				"wednesday",
				"thursday",
				"friday",
				"saturday",
			},
		},
	}

	for _, test := range data {
		for i, tm := range times {

			got, err := FormatTime(tm, test.Picture)

			if got != test.Outputs[i] {
				t.Errorf("%s: Expected %q, got %q", test.Picture, test.Outputs[i], got)
			}

			if err != nil {
				t.Errorf("%s: Expected nil error, got %s", test.Picture, err)
			}
		}
	}
}
