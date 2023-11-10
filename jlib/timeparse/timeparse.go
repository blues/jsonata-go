package timeparse

import (
	"fmt"
	"strconv"
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
	Millis         string `json:"Millis"`
	Hour           string `json:"Hour"`
	TimeZone       string `json:"TimeZone"`
	TimeZoneOffset string `json:"TimeZoneOffset"` // skip for now TODO
	YearMonth      string `json:"YearMonth"`
	YearWeek       string `json:"YearWeek"` // skip for now TODO
	YearIsoWeek    string `json:"YearIsoWeek"`
	YearDay        string `json:"YearDay"`
}

// TimeDateDimensions generates a JSON object dependent on input source timestamp, input source format and input source timezone
// using golang time formats
func TimeDateDimensions(inputSrcTs, inputSrcFormat, inputSrcTz string) (interface{}, error) {
	location, err := time.LoadLocation(inputSrcTz)
	if err != nil {
		return nil, err
	}

	// TODO clean up, speed up, make more efficient etc etc - but first, get it to work!

	parsedTime, err := parseDateTimeLocation(inputSrcTs, inputSrcFormat, location)
	if err != nil {
		return nil, err
	}

	// convert the parsed time into a UTC time for UTC calculations
	parsedTimeUTC := parsedTime.UTC()

	dateDim := DateDim{}

	// UTC TIME values

	dateDim.Utc = parsedTimeUTC.Format("2006-01-02T15:04:05.000Z")

	utcAsYearMonthDay := parsedTimeUTC.Format("2006-01-02")

	dateDim.DateUTC = utcAsYearMonthDay

	// LOCAL TIME values (we confirmed there need to be a seperate set of UTC values)

	dateID := parsedTime.Format("20060102")

	dateDim.DateID = "Dates_" + dateID

	dateDim.DateKey = dateID

	dateDim.Local = parsedTime.Format("2006-01-02T15:04:05.000-07:00")

	dateDim.DateLocal = parsedTime.Format("2006-01-02")

	dateDim.TimeZone = parsedTime.Location().String()

	dateDim.HourID = "Hours_" + parsedTime.Format("2006010215")

	dateDim.HourKey = parsedTime.Format("2006010215")

	dateDim.Hour = strconv.Itoa(parsedTime.Hour())

	dateDim.Millis = strconv.Itoa(int(parsedTime.UnixMilli()))

	dateDim.YearMonth = parsedTime.Format("200601")

	year, week := parsedTime.ISOWeek()

	dateDim.YearIsoWeek = fmt.Sprintf("%d%02d", year, week)

	dateDim.YearWeek = "" // TODO

	yearDay := parsedTime.Format("2006") + parsedTime.Format("002")
	
	dateDim.YearDay = yearDay

	dateDim.TimeZoneOffset = "" // TODO

	return dateDim, nil
}

func parseDateTimeLocation(d string, layout string, location *time.Location) (time.Time, error) {
	date, err := time.Parse(layout, d)
	if err != nil {
		return date, err
	}

	return date.In(location), nil
}
