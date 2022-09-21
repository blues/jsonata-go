package jlib

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// LJoin (golint)
func LJoin(v, v2 reflect.Value, field1, field2 string) (interface{}, error) {
	output := make([]interface{}, 0)

	i1, ok := v.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	i2, ok := v2.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	for a := range i1 {
		item1, ok := i1[a].(map[string]interface{})
		if !ok {
			continue
		}

		f1 := item1[field1]

		for b := range i2 {
			item2, ok := i2[b].(map[string]interface{})
			if !ok {
				continue
			}
			
			f2 := item2[field2]
			if f1 == f2 {
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
	}

	return output, nil
}
