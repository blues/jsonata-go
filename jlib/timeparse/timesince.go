package timeparse

import "time"

func Since(time1, time1format, time1location, time2, time2format, time2location string) (float64, error) {
	inputLocation, err := time.LoadLocation(time1location)
	if err != nil {
		return 0, err
	}

	firstTime, err := time.ParseInLocation(time1format, time1, inputLocation)
	if err != nil {
		return 0, err
	}

	outputLocation, err := time.LoadLocation(time2location)
	if err != nil {
		return 0, err
	}

	secondTime, err := time.ParseInLocation(time2format, time2, outputLocation)
	if err != nil {
		return 0, err
	}

	return firstTime.Sub(secondTime).Seconds(), nil
}
