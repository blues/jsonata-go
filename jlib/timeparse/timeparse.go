package timeparse

import (
	"fmt"
	"strconv"
	"time"
)

// Now - for tests
var Now = func() time.Time {
	return time.Now()
}

// DateDim is the date dimension object returned from the timeparse function
type DateDim struct {
	DateID         string `json:"DateId"`
	DateKey        string `json:"DateKey"`
	Utc            string `json:"UTC"`
	DateUTC        string `json:"DateUTC"`
	Local          string `json:"Local"`
	DateLocal      string `json:"DateLocal"`
	HourID         string `json:"HourId"`
	HourKey        string `json:"HourKey"`
	Millis         string `json:"Millis"`
	Hour           string `json:"Hour"`
	TimeZone       string `json:"TimeZone"`
	TimeZoneOffset string `json:"TimeZoneOffset"`
	YearMonth      string `json:"YearMonth"`
	YearWeek       string `json:"YearWeek"`
	YearIsoWeek    string `json:"YearIsoWeek"`
	YearDay        string `json:"YearDay"`
}

// TimeDateDimensions generates a JSON object dependent on input source timestamp, input source format and input source timezone
// using golang time formats
func TimeDateDimensions(inputSrcTs, inputSrcFormat, inputSrcTz string) (interface{}, error) {
	parsedTime, err := time.Parse(inputSrcFormat, inputSrcTs)
	if err != nil {
		return nil, err
	}

	//localLocation, err := time.LoadLocation(inputSrcTz)
	//if err != nil {
	//	return nil, err
	//}

	now := Now()

	// TODO clean up, speed up, make more efficient etc etc - but first, get it to work!

	// convert the parsed time into a UTC time for UTC calculations
	parsedTimeUTC := parsedTime.UTC()

	dateDim := DateDim{}

	dateID := parsedTime.Format("20060102")

	dateDim.DateID = "Dates_" + dateID

	dateDim.DateKey = dateID

	dateDim.Utc = parsedTimeUTC.Format("2006-01-02T15:04:05.000Z")

	dateDim.Local = now.Format("2006-01-02T15:04:05.000-07:00")

	dateDim.DateLocal = now.Format("2006-01-02")

	utcAsYearMonthDay := parsedTimeUTC.Format("2006-01-02")

	dateDim.DateUTC = utcAsYearMonthDay

	dateDim.TimeZone = parsedTime.Location().String()

	dateDim.HourID = "Hours_" + parsedTimeUTC.Format("2006010215")

	dateDim.HourKey = parsedTimeUTC.Format("2006010215")

	dateDim.Hour = strconv.Itoa(parsedTimeUTC.Hour())

	dateDim.Millis = strconv.Itoa(int(parsedTimeUTC.UnixMilli()))

	dateDim.YearMonth = parsedTimeUTC.Format("200601")

	year, week := parsedTimeUTC.ISOWeek()

	dateDim.YearIsoWeek = fmt.Sprintf("%d%02d", year, week)

	mondayWeek := getWeekOfYearString(parsedTimeUTC)

	dateDim.YearWeek = mondayWeek

	yearDay := parsedTimeUTC.Format("2006") + parsedTimeUTC.Format("002")
	dateDim.YearDay = yearDay

	offset, err := getTimeOffsetString(parsedTimeUTC, parsedTime)
	if err != nil {
		return nil, err
	}

	dateDim.TimeZoneOffset = offset

	return dateDim, nil
}

func getTimeOffsetString(t1, t2 time.Time) (string, error) {
	duration := t1.Sub(t2)

	offsetString := formatOffset(duration)

	return offsetString, nil
}

func formatOffset(diff time.Duration) string {
	sign := "+"

	if diff < 0 {
		sign = "-"
		diff = -diff
	}

	hours := diff / time.Hour

	minutes := (diff % time.Hour) / time.Minute

	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

func getWeekOfYearString(date time.Time) string {
	_, week := date.ISOWeek()

	// Find the start of the ISO week containing the first Wednesday
	firstWednesday := date.AddDate(0, 0, -int(date.Weekday())+1)
	if firstWednesday.Weekday() != time.Wednesday {
		firstWednesday = firstWednesday.AddDate(0, 0, 7-int(firstWednesday.Weekday())+int(time.Wednesday))
	}

	// Adjust week number for weeks starting from Monday and the first week containing a Wednesday
	if date.Weekday() == time.Sunday || date.Before(firstWednesday) {
		week--
	}

	return fmt.Sprintf("%04d%02d", date.Year(), week)
}
