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
	OutputSrcTz    string `json:"output_srcTz"`
	DateDim        struct {
		// Other
		TimeZone       string `json:"TimeZone"`       // lite
		TimeZoneOffset string `json:"TimeZoneOffset"` // lite
		YearMonth      int    `json:"YearMonth"`      // int
		YearWeek       int    `json:"YearWeek"`       // int
		YearIsoWeek    int    `json:"YearIsoWeek"`    // int
		YearDay        int    `json:"YearDay"`        // int
		DateID         string `json:"DateId"`         // lite
		DateKey        string `json:"DateKey"`        // lite
		HourID         string `json:"HourId"`
		HourKey        string `json:"HourKey"`
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
			result, err := jsonatatime.TimeDateDimensions(tc.InputSrcTs, tc.InputSrcFormat, tc.InputSrcTz, tc.OutputSrcTz)
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

func TestTimeLite(t *testing.T) {
	tests := []TestCase{}
	fileBytes, err := os.ReadFile("testdata_lite.json")
	require.NoError(t, err)
	err = json.Unmarshal(fileBytes, &tests)
	require.NoError(t, err)

	output := make([]interface{}, 0)

	for _, tc := range tests {
		tc := tc // race protection

		t.Run(tc.TestDesc, func(t *testing.T) {
			result, err := jsonatatime.TimeDateDimensionsLite(tc.InputSrcTs, tc.InputSrcFormat, tc.InputSrcTz, tc.OutputSrcTz)
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
	_ = os.WriteFile("outputdata_lite.json", outputbytes, os.ModePerm)
}
