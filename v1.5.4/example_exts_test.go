// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata_test

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	jsonata "github.com/blues/jsonata-go/v1.5.4"
)

//
// This example demonstrates how to extend JSONata with
// custom functions.
//

// titleCase returns a copy of the string s with all Unicode letters that begin words
// mapped to their Unicode title case.
func titleCase(s string) string {
	// Split the string into words
	words := strings.Fields(s)
	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToTitle(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// exts defines a function named "titlecase" which maps to
// our custom titleCase function. Any function,
// from the standard library or otherwise, can be used to
// extend JSONata, as long as it returns either one or two
// arguments (the second argument must be an error).
var exts = map[string]jsonata.Extension{
	"titlecase": {
		Func: titleCase,
	},
}

func ExampleExpr_RegisterExts() {

	// Create an expression that uses the titlecase function.
	e := jsonata.MustCompile(`$titlecase("beneath the underdog")`)

	// Register the titlecase function.
	err := e.RegisterExts(exts)
	if err != nil {
		log.Fatal(err)
	}

	// Evaluate.
	res, err := e.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
	// Output: Beneath The Underdog
}
