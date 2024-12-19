// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"github.com/xiatechs/jsonata-go/jlib"
	"github.com/xiatechs/jsonata-go/jtypes"
)

// Default format for dates: e.g. 2006-01-02 15:04 MST
const defaultDateFormat = "[Y]-[M01]-[D01] [H01]:[m] [ZN]"

// formatTime converts a unix time in seconds to a string with the
// given layout. If a time zone is provided, formatTime returns a
// timestamp with that time zone. Otherwise, it returns UTC time.
func formatTime(secs int64, picture jtypes.OptionalString, tz jtypes.OptionalString) (string, error) {

	if picture.String == "" {
		picture = jtypes.NewOptionalString(defaultDateFormat)
	}

	return jlib.FromMillis(secs*1000, picture, tz)
}

// parseTime converts a timestamp string with the given layout to
// a unix time in seconds.
func parseTime(value string, picture jtypes.OptionalString, tz jtypes.OptionalString) (int64, error) {

	if picture.String == "" {
		picture = jtypes.NewOptionalString(defaultDateFormat)
	}

	ms, err := jlib.ToMillis(value, picture, tz)
	if err != nil {
		return 0, err
	}

	return ms / 1000, nil
}
