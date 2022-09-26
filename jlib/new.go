package jlib

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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

func getVal(input interface{}) string {
	return fmt.Sprintf("%v", input)
}

// SimpleJoin - a 1 key left join very simple and useful in certain circumstances
func SimpleJoin(v, v2 reflect.Value, field1, field2 string) (interface{}, error) {
	output := make([]interface{}, 0)

	i1, ok := v.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	i2, ok := v2.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	field1Arr := strings.Split(field1, "|") // todo: only works as an OR atm and only 1 dimension deep

	field2Arr := strings.Split(field2, "|")

	if len(field1Arr) != len(field2Arr) {
		return nil, fmt.Errorf("field arrays must be same length")
	}

	for index := range field1Arr {
		output = append(output, addItems(i1, i2, field1Arr[index], field2Arr[index])...)
	}

	return output, nil
}

func addItems(i1, i2 []interface{}, field1, field2 string) []interface{} {
	output := make([]interface{}, 0)

	for a := range i1 {
		item1, ok := i1[a].(map[string]interface{})
		if !ok {
			continue
		}

		var exists bool

		f1 := item1[field1]

		for b := range i2 {
			item2, ok := i2[b].(map[string]interface{})
			if !ok {
				continue
			}

			f2 := item2[field2]
			if f1 == f2 {
				exists = true
				newitem := make(map[string]interface{})
				for key := range item1 {
					newitem[key] = item1[key]
				}
				for key := range item2 {
					newitem[key] = item2[key]
				}

				output = append(output, newitem)
			}
		}

		if !exists {
            output = append(output, item1)
		}
	}

	return output
}

// ObjMerge - merge two map[string]interface{} objects together - if they have unique keys
func ObjMerge(i1, i2 interface{}) interface{} {
	output := make(map[string]interface{})

	merge1, ok1 := i1.(map[string]interface{})
	merge2, ok2 := i2.(map[string]interface{})
	if !ok1 || !ok2 {
		return output
	}

	for key := range merge1 {
		output[key] = merge1[key]
	}

	for key := range merge2 {
		output[key] = merge2[key]
	}

	return output
}