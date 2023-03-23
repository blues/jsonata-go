// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jlib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/xiatechs/jsonata-go/jlib/jxpath"
	"github.com/xiatechs/jsonata-go/jtypes"
)

// 2006-01-02T15:04:05.000Z07:00
const defaultFormatTimeLayout = "[Y]-[M01]-[D01]T[H01]:[m]:[s].[f001][Z01:01t]"

const amSuffix = "am"
const pmSuffix = "pm"

var defaultParseTimeLayouts = []string{
	"[Y]-[M01]-[D01]T[H01]:[m]:[s][Z01:01t]",
	"[Y]-[M01]-[D01]T[H01]:[m]:[s][Z0100t]",
	"[Y]-[M01]-[D01]T[H01]:[m]:[s]",
	"[Y0001]-[M01]-[D01]",
	"[Y]-[M01]-[D01]",
	"[Y0001]-[M01]-[D01] [H01]:[m01]:[s01]",
	"[Y0001]-[M01]-[D01] [H01]:[m01]:[s01] [P]",
	"[H01]",
	"[Y]",
}

// FromMillis (golint)
func FromMillis(ms int64, picture jtypes.OptionalString, tz jtypes.OptionalString) (string, error) {
	t := msToTime(ms).UTC()

	if tz.String != "" {
		loc, err := parseTimeZone(tz.String)
		if err != nil {
			return "", err
		}

		t = t.In(loc)
	}

	layout := picture.String
	if layout == "" {
		layout = defaultFormatTimeLayout
	}

	return jxpath.FormatTime(t, layout)
}

// parseTimeZone parses a JSONata timezone.
//
// The format is a "+" or "-" character, followed by four digits, the first two
// denoting the hour offset, and the last two denoting the minute offset.
func parseTimeZone(tz string) (*time.Location, error) {
	// must be exactly 5 characters
	if len(tz) != 5 {
		return nil, fmt.Errorf("invalid timezone")
	}

	plusOrMinus := string(tz[0])

	// the first character must be a literal "+" or "-" character.
	// Any other character will error.
	var offsetMultiplier int
	switch plusOrMinus {
	case "-":
		offsetMultiplier = -1
	case "+":
		offsetMultiplier = 1
	default:
		return nil, fmt.Errorf("invalid timezone")
	}

	// take the first two digits as "HH"
	hours, err := strconv.Atoi(tz[1:3])
	if err != nil {
		return nil, fmt.Errorf("invalid timezone")
	}

	// take the last two digits as "MM"
	minutes, err := strconv.Atoi(tz[3:5])
	if err != nil {
		return nil, fmt.Errorf("invalid timezone")
	}

	// convert to seconds
	offsetSeconds := offsetMultiplier * (60 * ((60 * hours) + minutes))

	// construct a time.Location based on the tz string and the offset in seconds.
	loc := time.FixedZone(tz, offsetSeconds)

	return loc, nil
}

// ToMillis (golint)
func ToMillis(s string, picture jtypes.OptionalString, tz jtypes.OptionalString) (int64, error) {
	var err error
	var t time.Time

	layouts := defaultParseTimeLayouts
	if picture.String != "" {
		layouts = []string{picture.String}
	}

	// TODO: How are timezones used for parsing?
	for _, l := range layouts {
		if t, err = parseTime(s, l); err == nil {
			return timeToMS(t), nil
		}
	}

	return 0, err
}

var reMinus7 = regexp.MustCompile("-(0*7)")

func parseTime(s string, picture string) (time.Time, error) {
	// Go's reference time: Mon Jan 2 15:04:05 MST 2006
	refTime := time.Date(
		2006,
		time.January,
		2,
		15,
		4,
		5,
		0,
		time.FixedZone("MST", -7*60*60),
	)

	layout, err := jxpath.FormatTime(refTime, picture)
	if err != nil {
		return time.Time{}, fmt.Errorf("the second argument of the toMillis function must be a valid date format")
	}

	// Replace -07:00 with Z07:00
	layout = reMinus7.ReplaceAllString(layout, "Z$1")

	// First remove the milliseconds from the date time string as it messes up our layouts
	splitString := strings.Split(s, ".")
	var dateTimeWithoutMilli = splitString[0]

	var formattedTime = dateTimeWithoutMilli
	switch layout {
	case time.DateOnly:
		formattedTime = formattedTime[:len(time.DateOnly)]
	case time.RFC3339:
		// If the layout contains a time zone but the date string doesn't, lets remove it.
		if !strings.Contains(formattedTime, "Z") {
			layout = layout[:len(time.DateTime)]
		}
	}

	// Occasionally date time strings contain a T in the string and the layout doesn't, if that's the
	// case, lets remove it.
	if strings.Contains(formattedTime, "T") && !strings.Contains(layout, "T") {
		formattedTime = strings.ReplaceAll(formattedTime, "T", "")
	} else if !strings.Contains(formattedTime, "T") && strings.Contains(layout, "T") {
		layout = strings.ReplaceAll(layout, "T", "")
	}

	sanitisedLayout := strings.ToLower(stripSpaces(layout))
	sanitisedDateTime := strings.ToLower(stripSpaces(formattedTime))

	sanitisedLayout = addSuffixIfNotExists(sanitisedLayout, sanitisedDateTime)
	sanitisedDateTime = addSuffixIfNotExists(sanitisedDateTime, sanitisedLayout)

	t, err := time.Parse(sanitisedLayout, sanitisedDateTime)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"could not parse time %q due to inconsistency in layout and date time string, date %s layout %s",
			s,
			sanitisedDateTime,
			sanitisedLayout,
		)
	}

	return t, nil
}

// It isn't consistent that both the date time string and format have a PM/AM suffix. If we find the suffix
// on one of the strings, add it to the other. Sometimes we can have conflicting suffixes for example the layout
// is always in PM 2006-01-0215:04:05pm but the actual date time string could be AM 2023-01-3110:44:59am.
// If this is the case, just ignore it as the time will parse correctly.
func addSuffixIfNotExists(s string, target string) string {
	if strings.HasSuffix(target, amSuffix) && !strings.HasSuffix(s, amSuffix) && !strings.HasSuffix(s, pmSuffix) {
		return s + amSuffix
	}

	return s
}

func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

func msToTime(ms int64) time.Time {
	return time.Unix(ms/1000, (ms%1000)*int64(time.Millisecond))
}

func timeToMS(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
