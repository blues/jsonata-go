package timeparse_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	jsonatatime "github.com/xiatechs/jsonata-go/jlib/timeparse"
)

type TestCase struct {
	TestDesc       string `json:"testDesc"`
	InputSrcTs     string `json:"input_srcTs"`
	InputSrcFormat string `json:"input_srcFormat"`
	InputSrcTz     string `json:"input_srcTz"`
	DateDim        struct {
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
	} `json:"DateDim"`
}

func TestTime(t *testing.T) {
	tests := []TestCase{}
	fileBytes, err := os.ReadFile("testdata.json")
	require.NoError(t, err)
	err = json.Unmarshal(fileBytes, &tests)

	for _, tc := range tests {
		tc := tc // race protection

		tc.

		t.Run(tc.TestDesc, func(t *testing.T) {
			result, err := jsonatatime.TimeDateDimensions(tc.InputSrcTs, tc.InputSrcFormat, tc.InputSrcTz)
			require.NoError(t, err)
			expectedByts, _ := json.Marshal(tc.DateDim)
			expectedDateDim := jsonatatime.DateDim{}
			actualByts, _ := json.Marshal(result)
			actualDateDim := jsonatatime.DateDim{}
			err = json.Unmarshal(actualByts, &actualDateDim)
			err = json.Unmarshal(expectedByts, &expectedDateDim)
			require.Equal(t, expectedDateDim, actualDateDim)
		})
	}
}
