package jsonata_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xiatechs/jsonata-go"
)

const timezoneString = `
    {
    "CompanyID": "MOS",
    "DateCreated": "2023-01-31T00:00:00"
    }
`

func TestTimezoneConversion(t *testing.T) {
	var data interface{}

	// Decode JSON.
	err := json.Unmarshal([]byte(timezoneString), &data)
	if err != nil {
		t.Fatal(err)
	}

	// Create expression.
	e := jsonata.MustCompile("$fromMillis($toMillis(DateCreated),\"[Y0001][M01][D01]\")")

	// Evaluate.
	res, err := e.Eval(data)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
