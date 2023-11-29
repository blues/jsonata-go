package timeparse

import (
	"fmt"
	"strconv"
	"time"
)

// DateDim is the date dimension object returned from the timeparse function
type DateDim struct {
	// Other
	TimeZone       string `json:"TimeZone"`       // lite
	TimeZoneOffset string `json:"TimeZoneOffset"` // lite
	YearMonth      int    `json:"YearMonth"`      // int
	YearWeek       int    `json:"YearWeek"`       // int
	YearIsoWeek    int    `json:"YearIsoWeek"`    // int
	YearDay        int    `json:"YearDay"`        // int
	DateID         string `json:"DateId"`         // lite
	DateKey        int    `json:"DateKey"`        // lite
	DateTimeKey    int    `json:"DateTimeKey"`    // lite
	HourID         string `json:"HourId"`
	HourKey        int    `json:"HourKey"`
	Millis         int    `json:"Millis"`   // lite
	RawValue       string `json:"RawValue"` // lite

	// UTC
	UTC     string `json:"UTC"`     // lite
	DateUTC string `json:"DateUTC"` // lite
	HourUTC int    `json:"HourUTC"`

	// Local
	Local     string `json:"Local"`     // lite
	DateLocal string `json:"DateLocal"` // lite
	HourLocal int    `json:"HourLocal"`
}

// TimeDateDimensions generates a JSON object dependent on input source timestamp, input source format and input source timezone
// using golang time formats
func TimeDateDimensions(inputSrcTs, inputSrcFormat, inputSrcTz, requiredTz string) (*DateDim, error) {
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

	year, week := localTime.ISOWeek()

	yearDay, err := strconv.Atoi(localTime.Format("2006") + localTime.Format("002"))
	if err != nil {
		return nil, err
	}

	hourKeyStr := localTime.Format("2006010215")

	mondayWeek, err := getWeekOfYearString(localTime)
	if err != nil {
		return nil, err
	}

	yearIsoWeekInt, err := strconv.Atoi(fmt.Sprintf("%d%02d", year, week))
	if err != nil {
		return nil, err
	}

	yearMonthInt, err := strconv.Atoi(localTime.Format("200601"))
	if err != nil {
		return nil, err
	}

	dateKeyInt, err := strconv.Atoi(dateID)
	if err != nil {
		return nil, err
	}

	dateTimeID := localTime.Format("20060102150405000")

	dateTimeKeyInt, err := strconv.Atoi(dateTimeID)
	if err != nil {
		return nil, err
	}

	hourKeyInt, err := strconv.Atoi(hourKeyStr)
	if err != nil {
		return nil, err
	}

	localTimeStamp := localTime.Format("2006-01-02T15:04:05.000-07:00")
	offsetStr := localTime.Format("-07:00")
	// construct the date dimension structure
	dateDim := &DateDim{
		RawValue:       inputSrcTs,
		TimeZoneOffset: offsetStr,
		YearWeek:       mondayWeek,
		YearDay:        yearDay,
		YearIsoWeek:    yearIsoWeekInt,
		YearMonth:      yearMonthInt,
		Millis:         int(localTime.UnixMilli()),
		HourLocal:      localTime.Hour(),
		HourKey:        hourKeyInt,
		HourID:         "Hours_" + hourKeyStr,
		DateLocal:      localTime.Format("2006-01-02"),
		TimeZone:       localTime.Location().String(),
		Local:          localTimeStamp,
		DateKey:        dateKeyInt,
		DateTimeKey:    dateTimeKeyInt,
		DateID:         "Dates_" + dateID,
		DateUTC:        utcAsYearMonthDay,
		UTC:            utcTime.Format("2006-01-02T15:04:05.000Z"),
		HourUTC:        utcTime.Hour(),
	}

	return dateDim, nil
}

func getWeekOfYearString(date time.Time) (int, error) {
	_, week := date.ISOWeek()

	firstWednesday := date.AddDate(0, 0, -int(date.Weekday())+1)
	if firstWednesday.Weekday() != time.Wednesday {
		firstWednesday = firstWednesday.AddDate(0, 0, 7-int(firstWednesday.Weekday())+int(time.Wednesday))
	}

	if date.Weekday() == time.Sunday || date.Before(firstWednesday) {
		week--
	}

	return strconv.Atoi(fmt.Sprintf("%04d%02d", date.Year(), week))
}
