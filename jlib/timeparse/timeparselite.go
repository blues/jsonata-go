package timeparse

import (
	"strconv"
	"strings"
	"time"
)

/*
	RawValue:       inputSrcTs,
	TimeZoneOffset: getOffsetString(localTimeStamp),
	Millis:         int(localTime.UnixMilli()),
	DateLocal:      localTime.Format("2006-01-02"),
	TimeZone:       localTime.Location().String(),
	Local:          localTimeStamp,
	DateKey:        dateID,
	DateID:         "Dates_" + dateID,
	DateUTC:        utcAsYearMonthDay,
	UTC:            utcTime.Format("2006-01-02T15:04:05.000Z"),
*/

// DateDimLite is the date dimension object returned from the timeparse function (light version)
type DateDimLite struct {
	// Other
	TimeZone       string `json:"TimeZone"`       // lite
	TimeZoneOffset string `json:"TimeZoneOffset"` // lite
	DateID         string `json:"DateId"`         // lite
	DateKey        int    `json:"DateKey"`        // lite
	DateTimeKey    int    `json:"DateTimeKey"`    // lite
	Millis         int    `json:"Millis"`         // lite
	RawValue       string `json:"RawValue"`       // lite

	// UTC
	UTC     string `json:"UTC"`     // lite
	DateUTC string `json:"DateUTC"` // lite

	// Local
	Local     string `json:"Local"`     // lite
	DateLocal string `json:"DateLocal"` // lite
}

// TimeDateDimensionsLite generates a JSON object dependent on input source timestamp, input source format and input source timezone
// using golang time formats
func TimeDateDimensionsLite(inputSrcTs, inputSrcFormat, inputSrcTz, requiredTz string) (*DateDimLite, error) {
	inputLocation, err := time.LoadLocation(inputSrcTz)
	if err != nil {
		return nil, err
	}

	// Since the source timestamp is implied to be in local time ("Europe/London"),
	// we parse it with the location set to Europe/London
	inputTime, err := time.ParseInLocation(inputSrcFormat, inputSrcTs, inputLocation)
	if err != nil {
		return nil, err
	}

	outputLocation, err := time.LoadLocation(requiredTz)
	if err != nil {
		return nil, err
	}

	localTime := inputTime.In(outputLocation)

	// convert the parsed time into a UTC time for UTC calculations
	utcTime := localTime.UTC()

	// UTC TIME values

	utcAsYearMonthDay := utcTime.Format("2006-01-02")

	// Input time stamp TIME values (we confirmed there need to be a seperate set of UTC values)
	dateID := localTime.Format("20060102")

	dateKeyInt, err := strconv.Atoi(dateID)
	if err != nil {
		return nil, err
	}

	// Input time stamp TIME values (we confirmed there need to be a seperate set of UTC values)
	dateTimeID := localTime.Format("20060102150405.000")

	dateTimeID = strings.ReplaceAll(dateTimeID, ".", "")

	dateTimeKeyInt, err := strconv.Atoi(dateTimeID)
	if err != nil {
		return nil, err
	}

	localTimeStamp := localTime.Format("2006-01-02T15:04:05.000-07:00")
	offsetStr := localTime.Format("-07:00")
	// construct the date dimension structure
	dateDim := &DateDimLite{
		RawValue:       inputSrcTs,
		TimeZoneOffset: offsetStr,
		Millis:         int(localTime.UnixMilli()),
		DateLocal:      localTime.Format("2006-01-02"),
		TimeZone:       localTime.Location().String(),
		Local:          localTimeStamp,
		DateKey:        dateKeyInt,
		DateTimeKey:    dateTimeKeyInt,
		DateID:         "Dates_" + dateID,
		DateUTC:        utcAsYearMonthDay,
		UTC:            utcTime.Format("2006-01-02T15:04:05.000Z"),
	}

	return dateDim, nil
}
