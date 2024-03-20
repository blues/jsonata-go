// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"bytes"
	"github.com/goccy/go-json"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	jsonata "github.com/xiatechs/jsonata-go"
	"github.com/xiatechs/jsonata-go/jtypes"
)

func init() {

	argUndefined0 := jtypes.ArgUndefined(0)

	exts := map[string]jsonata.Extension{
		"formatTime": {
			Func:             formatTime,
			UndefinedHandler: argUndefined0,
		},
		"parseTime": {
			Func:             parseTime,
			UndefinedHandler: argUndefined0,
		},
	}

	if err := jsonata.RegisterExts(exts); err != nil {
		panic(err)
	}
}

func main() {

	port := flag.Uint("port", 8080, "The port `number` to serve on")
	flag.Parse()

	http.HandleFunc("/eval", evaluate)
	http.HandleFunc("/bench", benchmark)
	http.Handle("/", http.FileServer(http.Dir("site")))

	log.Printf("Starting JSONata Server on port %d:\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func evaluate(w http.ResponseWriter, r *http.Request) {
	input := strings.TrimSpace(r.FormValue("json"))
	if input == "" {
		http.Error(w, "Input is empty", http.StatusBadRequest)
		return
	}

	expression := strings.TrimSpace(r.FormValue("expr"))
	if expression == "" {
		http.Error(w, "Expression is empty", http.StatusBadRequest)
		return
	}

	b, status, err := eval(input, expression)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), status)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Fatal(err)
	}
}

func eval(input, expression string) (b []byte, status int, err error) {
	defer func() {
		if r := recover(); r != nil {
			b = nil
			status = http.StatusInternalServerError
			err = fmt.Errorf("PANIC: %v", r)
			return
		}
	}()

	// Decode the JSON.
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("input error: %s", err)
	}

	// Compile the JSONata expression.
	expr, err := jsonata.Compile(expression)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("compile error: %s", err)
	}

	// Evaluate the JSONata expression.
	result, err := expr.Eval(data)
	if err != nil {
		if err == jsonata.ErrUndefined {
			// Don't treat not finding any results as an error.
			return []byte("No results found"), http.StatusOK, nil
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("eval error: %s", err)
	}

	// Return the JSONified results.
	b, err = jsonify(result)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("encode error: %s", err)
	}

	return b, http.StatusOK, nil
}

func jsonify(v interface{}) ([]byte, error) {

	b := bytes.Buffer{}
	e := json.NewEncoder(&b)
	e.SetIndent("", "    ")
	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
