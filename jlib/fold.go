package jlib

import "errors"

func FoldArray(input interface{}) ([][]interface{}, error) {
	inputSlice, ok := input.([]interface{})
	if !ok {
		return nil, errors.New("input for $foldarray was not an []interface type")
	}

	result := make([][]interface{}, len(inputSlice))

	for i := range inputSlice {
		result[i] = make([]interface{}, i+1)
		copy(result[i], inputSlice[:i+1])
	}

	return result, nil
}
