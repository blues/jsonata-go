package jlib

import (
	"github.com/goccy/go-json"
	"errors"
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

const (
	arrDelimiter = "|"
	keyDelimiter = "¬"
)

// SimpleJoin - a multi-key multi-level full OR join - very simple and useful in certain circumstances
func SimpleJoin(v, v2 reflect.Value, field1, field2 string) (interface{}, error) {
	if !(v.IsValid() && v.CanInterface() && v2.IsValid() && v2.CanInterface()) {
		return nil, nil
	}

	i1, ok := v.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	i2, ok := v2.Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("both objects must be slice of objects")
	}

	field1Arr := strings.Split(field1, arrDelimiter) // todo: only works as an OR atm

	field2Arr := strings.Split(field2, arrDelimiter)

	if len(field1Arr) != len(field2Arr) {
		return nil, fmt.Errorf("field arrays must be same length")
	}

	relationMap := make(map[string]*relation)

	for index := range field1Arr {
		addItems(relationMap, i1, i2, field1Arr[index], field2Arr[index])
	}

	output := make([]interface{}, 0)

	for index := range relationMap {
		output = append(output, relationMap[index].generateItem())
	}

	return output, nil
}

type relation struct {
	object  map[string]interface{}
	related []interface{}
}

func newRelation(input map[string]interface{}) *relation {
	return &relation{
		object:  input,
		related: make([]interface{}, 0),
	}
}

func (r *relation) generateItem() map[string]interface{} {
	newitem := make(map[string]interface{})

	for key := range r.object {
		newitem[key] = r.object[key]

		for index := range r.related {
			if val, ok := r.related[index].(map[string]interface{}); ok {
				for key := range val {
					newitem[key] = val[key]
				}
			}
		}

	}

	return newitem
}

func addItems(relationMap map[string]*relation, i1, i2 []interface{}, field1, field2 string) {
	for a := range i1 {
		item1, ok := i1[a].(map[string]interface{})
		if !ok {
			continue
		}

		key := fmt.Sprintf("%v", item1)

		if _, ok := relationMap[key]; !ok {
			relationMap[key] = newRelation(item1)
		}

		rel := relationMap[key]

		f1 := getMapStringValue(strings.Split(field1, keyDelimiter), 0, item1)
		if f1 == nil {
			continue
		}

		for b := range i2 {
			f2 := getMapStringValue(strings.Split(field2, keyDelimiter), 0, i2[b])
			if f2 == nil {
				continue
			}

			if f1 == f2 {
				rel.related = append(rel.related, i2[b])
			}
		}

		relationMap[key] = rel
	}
}

func outsideRange(fieldArr []string, index int) bool {
	return index > len(fieldArr)-1
}

func getMapStringValue(fieldArr []string, index int, item interface{}) interface{} {
	if outsideRange(fieldArr, index) {
		return nil
	}

	if obj, ok := item.(map[string]interface{}); ok {
		for key := range obj {
			if key == fieldArr[index] {
				if len(fieldArr)-1 == index {
					return obj[key]
				} else {
					index++
					new := getMapStringValue(fieldArr, index, obj[key])
					if new != nil {
						return new
					}
				}
			}
		}
	}

	return getArrayValue(fieldArr, index, item)
}

func getArrayValue(fieldArr []string, index int, item interface{}) interface{} {
	if outsideRange(fieldArr, index) {
		return nil
	}

	if obj, ok := item.([]interface{}); ok {
		for value := range obj {
			a := fmt.Sprintf("%v", fieldArr[index])
			b := fmt.Sprintf("%v", obj[value])
			if a == b {
				if len(fieldArr)-1 == index {
					return item
				} else {
					index++
					new := getMapStringValue(fieldArr, index, obj)
					if new != nil {
						return new
					}
				}
			}
		}
	}

	return getSingleValue(fieldArr, index, item)
}

func getSingleValue(fieldArr []string, index int, item interface{}) interface{} {
	if outsideRange(fieldArr, index) {
		return nil
	}

	a := fmt.Sprintf("%v", fieldArr[index])
	b := fmt.Sprintf("%v", item)
	if a == b {
		if len(fieldArr)-1 == index {
			return item
		} else {
			index++
			new := getMapStringValue(fieldArr, index, item)
			if new != nil {
				return new
			}
		}
	}

	return nil
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

// setValue sets the value in the obj map at the specified dot notation path.
func setValue(obj map[string]interface{}, path string, value interface{}) {
	paths := strings.Split(path, ".") // Split the path into parts

	// Iterate through path parts to navigate/create nested maps
	for i := 0; i < len(paths)-1; i++ {
		// If the key does not exist, create a new map at the key
		_, ok := obj[paths[i]]
		if !ok {
			obj[paths[i]] = make(map[string]interface{})
		}

		obj, ok = obj[paths[i]].(map[string]interface{})
		if !ok {
			continue
		}
	}

	obj[paths[len(paths)-1]] = value
}

// objectsToDocument converts an array of Items to a nested map according to the Code paths.
func ObjectsToDocument(input interface{}) (interface{}, error) {
	trueInput, ok := input.([]interface{})
	if !ok {
		return nil, errors.New("$objectsToDocument input must be an array of objects")
	}

	output := make(map[string]interface{}) // Initialize the output map
	// Iterate through each item in the input
	for _, itemToInterface := range trueInput {
		item, ok := itemToInterface.(map[string]interface{})
		if !ok {
			return nil, errors.New("$objectsToDocument input must be an array of objects with Code and Value fields")
		}
		// Call setValue for each item to set the value in the output map
		code, ok := item["Code"].(string)
		if !ok {
			continue
		}
		value := item["Value"]
		setValue(output, code, value)
	}

	return output, nil // Return the output map
}

// TransformRule defines a transformation rule with a search substring and a new name.
type TransformRule struct {
	SearchSubstring string
	NewName         string
}

// RenameKeys applies a series of transformations to the keys in a JSON-compatible data structure.
// 'data' is the original data where keys need to be transformed.
// 'rulesInterface' is expected to be a slice of interface{}, where each element is a slice containing two strings:
// the substring to search for in the keys, and the new name to replace the key with.
func RenameKeys(data interface{}, rulesInterface interface{}) (interface{}, error) {
	// Attempt to assert rulesInterface as a slice of interface{}
	rulesRaw, ok := rulesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("rules must be a slice of interface{}")
	}

	// Process each rule, converting it into a TransformRule
	var rules []TransformRule
	for _, r := range rulesRaw {
		rule, ok := r.([]interface{})
		if !ok || len(rule) != 2 {
			return nil, fmt.Errorf("each rule must be an array of two strings")
		}

		searchSubstring, ok1 := rule[0].(string)
		newName, ok2 := rule[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("each rule must be an array of two strings")
		}

		rules = append(rules, TransformRule{SearchSubstring: searchSubstring, NewName: newName})
	}

	// Marshal the original data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a map for easy manipulation
	var mapData map[string]interface{}
	err = json.Unmarshal(jsonData, &mapData)
	if err != nil {
		return nil, err
	}

	// Create a new map to store the modified data
	newData := make(map[string]interface{})
	for key, value := range mapData {
		newKey := key // Default to the original key
		// Apply transformation rules
		for _, rule := range rules {
			if strings.Contains(key, rule.SearchSubstring) {
				newKey = rule.NewName // Update the key if rule matches
				break
			}
		}
		newData[newKey] = value // Store the value with the new key
	}

	// Re-marshal to JSON to maintain the same data type as the input
	resultJSON, err := json.Marshal(newData)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON back into an interface{} for the return value
	var result interface{}
	err = json.Unmarshal(resultJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
