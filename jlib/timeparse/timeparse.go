package timeparse

import (
	"fmt"
	"time"
)

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
	Millis         string  `json:"Millis"`
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

	now, err := time.Parse("2006-01-02T15:04:05.000-07:00", `2020-01-01T00:00:00.000+00:00`)
	if err != nil {
		return nil, err
	}

	// TODO clean up, speed up, make more efficient etc etc - but first, get it to work!

	dateDim := DateDim{}
	dateID := parsedTime.Format("20060102")
	dateDim.DateID = "Dates_" + dateID
	dateDim.DateKey = dateID
	dateDim.Utc = parsedTime.UTC().Format("2006-01-02T15:04:05.000Z")
	utcAsYearMonthDay := parsedTime.UTC().Format("2006-01-02")
	dateDim.Local = now.Format("2006-01-02T15:04:05.000-07:00")
	dateDim.DateLocal = now.Format("2006-01-02")
	dateDim.DateUTC = utcAsYearMonthDay
	dateDim.TimeZone = parsedTime.Location().String()
	dateDim.HourID = "Hours_" + parsedTime.Format("2006010215")
	dateDim.HourKey = parsedTime.Format("2006010215")
	dateDim.Hour = fmt.Sprintf("%d",((parsedTime.YearDay()-1)*24 + parsedTime.Hour()))
	dateDim.Millis = fmt.Sprintf("%d",parsedTime.UnixNano() / 1e6)

	dateDim.YearMonth = parsedTime.Format("200601")
	year, week := parsedTime.ISOWeek()
	normalWeek := getNormalWeekNumber(parsedTime)
	dateDim.YearIsoWeek = fmt.Sprintf("%d%02d", year, week)
	dateDim.YearWeek = fmt.Sprintf("%d%02d", year, normalWeek)
	yearDay := parsedTime.Format("2006") + parsedTime.Format("002")
	dateDim.YearDay = yearDay
	offset, _ := getTimeOffsetString(parsedTime, now)
	dateDim.TimeZoneOffset = offset
	return dateDim, nil
}


func getTimeOffsetString(t1, t2 time.Time) (string, error) {
	duration := t1.Sub(t2)

	// Calculate hours and minutes
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	// Format the offset string
	offsetString := fmt.Sprintf("%+03d:%02d", hours, minutes)

	return offsetString, nil
}

func getNormalWeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	weekday := int(t.Weekday())

	// Adjust week number based on the weekday
	if weekday == 0 {
		// Sunday, move to the previous week
		week--
	} else if weekday == 6 {
		// Saturday, move to the next week
		week++
	}

	return week
}
