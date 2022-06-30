package jsonata

import (
	"encoding/json"
	"fmt"
)

// Distinct - a function that you can pass an array of json objects through to remove exact duplicates
func Distinct(input interface{}) (interface{}, error) {
	if _, ok := input.([]interface{}); !ok {
		return nil, fmt.Errorf("distinct can only be applied to an array of JSON objects")
	}

	jsonArray := input.([]interface{})

	output := make([]interface{}, 0)

	deduper := make(map[string]struct{})

	for key := range jsonArray {
		bytes, err := json.Marshal(jsonArray[key])
		if err != nil {
			return nil, err
		}

		blobStr := string(bytes)

		if _, ok := deduper[blobStr]; !ok {
			deduper[blobStr] = struct{}{}

			output = append(output, jsonArray[key])
		}
	}

	return output, nil
}
