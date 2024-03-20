// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/goccy/go-json"

	jsonata "github.com/xiatechs/jsonata-go"
)

var (
	benchData = []byte(`
{
    "req":"note.add",
    "device":"sim32-1232353453453452346",
    "app":"test1",
    "file":"geiger.q",
    "note":"abc123",
    "by":"1",
    "when":1512335179,
    "where":"87JFH688+2GP",
    "payload":"SGVsbG8sIHdvcmxkLg==",
    "body":
    {
        "loc_olc":"87JFH688+2GP",
        "env_temp":9.407184,
        "env_humid":77.071495,
        "env_press":1016.25323,
        "bat_voltage":3.866328,
        "bat_current":0.078125,
        "bat_charge":64.42578,
        "lnd_7318u":27.6,
        "lnd_7318c":23.1,
        "lnd_7128ec":9.3,
        "pms_pm01_0":0,
        "pms_pm02_5":0,
        "pms_pm10_0":1,
        "pms_c00_30":11076,
        "pms_c00_50":3242,
        "pms_c01_00":246,
        "pms_c02_50":44,
        "pms_c05_00":10,
        "pms_c10_00":10,
        "pms_csecs":118,
        "opc_pm01_0":1.9840136,
        "opc_pm02_5":3.9194343,
        "opc_pm10_0":9.284608,
        "opc_c00_38":139,
        "opc_c00_54":154,
        "opc_c01_00":121,
        "opc_c02_10":30,
        "opc_c05_00":3,
        "opc_c10_00":0,
        "opc_csecs":120
    }
}`)

	benchExpression = `
(
	$values := {
        "device_uid": device,
        "when_captured": $formatTime(when),
        "loc_lat": $latitudeFromOLC(body.loc_olc),
        "loc_lon": $longitudeFromOLC(body.loc_olc)
    };

    req = "note.add" and when ? $merge([body, $values]) : $error("unexpected req/when")
)`
)

// Decode the JSON.
var data interface{}

func init() {
	if err := json.Unmarshal(benchData, &data); err != nil {
		panic(err)
	}
}

func benchmark(w http.ResponseWriter, r *http.Request) {

	// Compile the JSONata expression.
	expr, err := jsonata.Compile(benchExpression)
	if err != nil {
		bencherr(w, err)
	}

	// Evaluate the JSONata expression.
	_, err = expr.Eval(data)
	if err != nil {
		bencherr(w, err)
	}

	if _, err := w.Write([]byte("success")); err != nil {
		log.Fatal(err)
	}
}

func bencherr(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)

}
