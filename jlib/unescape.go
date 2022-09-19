package jlib

import (
	"fmt"
	"encoding/json"
)

// Unescape an escaped json string into JSON (once)
func Unescape(input string) (interface{}, error) {
	var output interface{}

	err := json.Unmarshal([]byte(input), &output)
	if err != nil {
		return output, fmt.Errorf("unescape json unmarshal error: %v", err)
	}

	return output, nil
}