package timeparse_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		UTC            string `json:"UTC"`
		DateUTC        string `json:"DateUTC"`
		Parsed         string `json:"Parsed"`
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
	} `json:"DateDim"`
}

func TestTime(t *testing.T) {
	tests := []TestCase{}
	fileBytes, err := os.ReadFile("testdata.json")
	require.NoError(t, err)
	err = json.Unmarshal(fileBytes, &tests)
	require.NoError(t, err)

	output := make([]interface{}, 0)

	for _, tc := range tests {
		tc := tc // race protection

		t.Run(tc.TestDesc, func(t *testing.T) {
			result, err := jsonatatime.TimeDateDimensions(tc.InputSrcTs, tc.InputSrcFormat, tc.InputSrcTz)
			require.NoError(t, err)

			testObj := tc

			expectedByts, err := json.Marshal(tc.DateDim)
			require.NoError(t, err)

			expectedDateDim := jsonatatime.DateDim{}

			actualByts, err := json.Marshal(result)
			require.NoError(t, err)

			actualDateDim := jsonatatime.DateDim{}

			err = json.Unmarshal(actualByts, &actualDateDim)
			require.NoError(t, err)

			testObj.DateDim = actualDateDim
			output = append(output, testObj)
			err = json.Unmarshal(expectedByts, &expectedDateDim)
			assert.Equal(t, expectedDateDim, actualDateDim)
		})
	}

	outputbytes, _ := json.MarshalIndent(output, "", " ")
	_ = os.WriteFile("outputdata.json", outputbytes, os.ModePerm)
}
