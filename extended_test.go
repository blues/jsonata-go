package jsonata

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCasesPath = "extendedTestFiles"

const (
	expectedInputFile    = "input.json"
	expectedOutputFile   = "output.json"
	expectedInputJsonata = "input.jsonata"
)

func TestChassis(t *testing.T) {
	entries, err := os.ReadDir(testCasesPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			testCase := entry.Name()

			testCasePath := filepath.Join(testCasesPath, testCase)

			t.Run(testCasePath, func(t *testing.T) {
				runTest(t,
					filepath.Join(testCasePath, expectedInputFile),
					filepath.Join(testCasePath, expectedOutputFile),
					filepath.Join(testCasePath, expectedInputJsonata),
				)
			})
		}
	}
}

func runTest(t *testing.T, inputfile, outputfile, jsonatafile string) {
	inputBytes, err := os.ReadFile(inputfile)
	require.NoError(t, err)

	outputBytes, err := os.ReadFile(outputfile)
	require.NoError(t, err)

	jsonataBytes, err := os.ReadFile(jsonatafile)
	require.NoError(t, err)

	expr, err := Compile(string(jsonataBytes))
	require.NoError(t, err)

	var input, output interface{}

	err = json.Unmarshal(inputBytes, &input)
	require.NoError(t, err)

	result, err := expr.Eval(input)
	require.NoError(t, err)

	err = json.Unmarshal(outputBytes, &output)
	require.NoError(t, err)

	assert.Equal(t, result, output)
}
