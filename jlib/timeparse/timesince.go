package timeparse

import "time"

func Since(time1, time1format, time1location, time2, time2format, time2location string) (float64, error) {
	firstTime, err := getTimeWithLocation(time1, time1format, time1location)
	if err != nil {
		return 0, err
	}

	secondTime, err := getTimeWithLocation(time2, time2format, time2location)
	if err != nil {
		return 0, err
	}

	return firstTime.Sub(secondTime).Seconds(), nil
}

func getTimeWithLocation(inputSrcTs, inputSrcFormat, inputSrcTz string) (time.Time, error) {
	location, err := time.LoadLocation(inputSrcTz)
	if err != nil {
		return time.Time{}, err
	}

	return parseDateTimeLocation(inputSrcTs, inputSrcFormat, location)
}
