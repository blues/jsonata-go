package timeparse

import (
	"time"
)

// TimeDateDimensionsLite generates a JSON object dependent on input source timestamp, input source format and input source timezone
// using golang time formats
func TimeDateDimensionsLite(inputSrcTs, inputSrcFormat, inputSrcTz, requiredTz string) (*DateDim, error) {
	inputLocation, err := time.LoadLocation(inputSrcTz)
	if err != nil {
		return nil, err
	}

	inputTime, err := parseDateTimeLocation(inputSrcTs, inputSrcFormat, inputLocation)
	if err != nil {
		return nil, err
	}

	outputLocation, err := time.LoadLocation(requiredTz)
	if err != nil {
		return nil, err
	}

	// Convert the time to the output time zone
	localTime := inputTime.In(outputLocation)

	// convert the parsed time into a UTC time for UTC calculations
	utcTime := localTime.UTC()

	// UTC TIME values

	utcAsYearMonthDay := utcTime.Format("2006-01-02")

	// Input time stamp TIME values (we confirmed there need to be a seperate set of UTC values)
	dateID := localTime.Format("20060102")

	localTimeStamp := localTime.Format("2006-01-02T15:04:05.000Z-07:00")
	// construct the date dimension structure
	dateDim := &DateDim{
		TimeZoneOffset: getOffsetString(localTimeStamp),
		Millis:         int(localTime.UnixMilli()),
		DateLocal:      localTime.Format("2006-01-02"),
		TimeZone:       localTime.Location().String(),
		Local:          localTimeStamp,
		DateKey:        dateID,
		DateID:         "Dates_" + dateID,
		DateUTC:        utcAsYearMonthDay,
		UTC:            utcTime.Format("2006-01-02T15:04:05.000Z"),
	}

	return dateDim, nil
}
