// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata_test

import (
	"fmt"
	"github.com/goccy/go-json"
	"log"

	jsonata "github.com/xiatechs/jsonata-go"
)

const jsonString = `
    {
        "orders": [
            {"price": 10, "quantity": 3},
            {"price": 0.5, "quantity": 10},
            {"price": 100, "quantity": 1}
        ]
    }
`

func ExampleExpr_Eval() {

	var data interface{}

	// Decode JSON.
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		log.Fatal(err)
	}

	// Create expression.
	e := jsonata.MustCompile("$sum(orders.(price*quantity))")

	// Evaluate.
	res, err := e.Eval(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
	// Output: 135
}
