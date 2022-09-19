package jlib

import (
	"fmt"
	"encoding/json"
)

// Unescape a string into JSON - simple but powerful
func Unescape(input string) (interface{}, error) {
	var output interface{}

	err := json.Unmarshal([]byte(input), &output)
	if err != nil {
		return output, fmt.Errorf("unescape json unmarshal error: %v", err)
	}

	return output, nil
}