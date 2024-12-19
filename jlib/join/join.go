package join

import (
	"errors"
	"fmt"
	"reflect"
)

// OneToManyJoin performs a join operation between two slices of maps/structs based on specified keys.
// It supports different types of joins: left, right, inner, and full.
func OneToManyJoin(leftArr, rightArr interface{}, leftKey, rightKey, rightArrayName, joinType string) (interface{}, error) {
	// Convert input to slices of interfaces
	trueLeftArr, ok := leftArr.([]interface{})
	if !ok {
		return nil, errors.New("left input must be an array of Objects")
	}

	trueRightArr, ok := rightArr.([]interface{})
	if !ok {
		return nil, errors.New("right input must be an array of Objects")
	}

	// Maps for tracking processed items
	alreadyProcessed := make(map[string]bool)
	rightProcessed := make(map[string]bool)

	// Create a map for faster lookup of rightArr elements based on the key
	rightMap := make(map[string][]interface{})
	for _, item := range trueRightArr {
		itemMap, ok := item.(map[string]interface{})
		if ok {
			if itemKey, ok := itemMap[rightKey]; ok {
				strVal := fmt.Sprintf("%v", itemKey)
				rightMap[strVal] = append(rightMap[strVal], item)
			}
		}
	}

	// Slice to store the merged results
	var result []map[string]interface{}
	leftMatched := make(map[string]interface{})

	// Iterate through the left array and perform the join
	for _, leftItem := range trueLeftArr {
		itemMap, ok := leftItem.(map[string]interface{})
		if ok {
			if itemKey, ok := itemMap[leftKey]; ok {
				strVal := fmt.Sprintf("%v", itemKey)

				// Determine the right items to join
				rightItems := rightMap[strVal]

				// Perform the join based on the join type
				if joinType == "left" || joinType == "full" || (joinType == "inner" && len(rightItems) > 0) {
					mergedItem := mergeItems(leftItem, rightItems, rightArrayName)
					result = append(result, mergedItem)
				}

				// Mark items as processed
				leftMatched[strVal] = leftItem
				alreadyProcessed[strVal] = true
			}
		}
	}

	// Add items from the right array for right or full join
	if joinType == "right" || joinType == "full" {
		for _, rightItem := range trueRightArr {
			itemMap, ok := rightItem.(map[string]interface{})
			if ok {
				if itemKey, ok := itemMap[rightKey]; ok {
					strVal := fmt.Sprintf("%v", itemKey)

					// Determine the left item to merge with
					var leftItemToMerge interface{}
					if leftMatch, ok := leftMatched[strVal]; ok {
						leftItemToMerge = leftMatch
					} else {
						leftItemToMerge = map[string]interface{}{rightKey: itemKey}
					}

					// Handle right and full join separately to avoid duplication
					if joinType == "right" && !rightProcessed[strVal] {
						mergedItem := mergeItems(leftItemToMerge, rightMap[strVal], rightArrayName)
						result = append(result, mergedItem)
						rightProcessed[strVal] = true
					} else if joinType == "full" && !rightProcessed[strVal] && !alreadyProcessed[strVal] {
						mergedItem := mergeItems(leftItemToMerge, rightMap[strVal], rightArrayName)
						result = append(result, mergedItem)
						rightProcessed[strVal] = true
					}
				}
			}
		}
	}

	return result, nil
}

func mergeItems(leftItem interface{}, rightItems []interface{}, rightArrayName string) map[string]interface{} {
	mergedItem := make(map[string]interface{})

	// Check if leftItem is a map or a struct and merge accordingly
	leftVal := reflect.ValueOf(leftItem)
	if leftVal.Kind() == reflect.Map {
		// Merge fields from the map
		for _, key := range leftVal.MapKeys() {
			mergedItem[key.String()] = leftVal.MapIndex(key).Interface()
		}
	} else {
		// Merge fields from the struct
		leftType := leftVal.Type()
		for i := 0; i < leftVal.NumField(); i++ {
			fieldName := leftType.Field(i).Name
			fieldValue := leftVal.Field(i).Interface()
			mergedItem[fieldName] = fieldValue
		}
	}

	// If there are matching items in the right array, add them under the specified name
	if len(rightItems) > 0 {
		mergedItem[rightArrayName] = rightItems
	}

	return mergedItem
}

// OneToManyJoin2 performs a join operation between two slices of maps/structs based on specified keys.
// It supports different types of joins: left, right, inner, and full.
func OneToManyJoin2(leftArr, rightArr interface{}, leftKey, rightKey, rightArrayName, joinType string) (interface{}, error) {
	// Convert input to slices of interfaces
	trueLeftArr, ok := leftArr.([]interface{})
	if !ok {
		return nil, errors.New("left input must be an array of Objects")
	}

	trueRightArr, ok := rightArr.([]interface{})
	if !ok {
		return nil, errors.New("right input must be an array of Objects")
	}

	// Maps for tracking processed items
	alreadyProcessed := make(map[string]bool)
	rightProcessed := make(map[string]bool)

	// Create a map for faster lookup of rightArr elements based on the key
	rightMap := make(map[string][]interface{})
	for _, item := range trueRightArr {
		itemMap, ok := item.(map[string]interface{})
		if ok {
			if itemKey, ok := itemMap[rightKey]; ok {
				strVal := fmt.Sprintf("%v", itemKey)
				rightMap[strVal] = append(rightMap[strVal], item)
			}
		}
	}

	// Slice to store the merged results
	var result []map[string]interface{}
	leftMatched := make(map[string]interface{})

	// Iterate through the left array and perform the join
	for _, leftItem := range trueLeftArr {
		itemMap, ok := leftItem.(map[string]interface{})
		if ok {
			if itemKey, ok := itemMap[leftKey]; ok {
				strVal := fmt.Sprintf("%v", itemKey)

				// Determine the right items to join
				rightItems := rightMap[strVal]

				// Perform the join based on the join type
				if joinType == "left" || joinType == "full" || (joinType == "inner" && len(rightItems) > 0) {
					mergedItem := mergeItemsNew(leftItem, rightItems, rightArrayName)
					result = append(result, mergedItem)
				}

				// Mark items as processed
				leftMatched[strVal] = leftItem
				alreadyProcessed[strVal] = true
			}
		}
	}

	// Add items from the right array for right or full join
	if joinType == "right" || joinType == "full" {
		for _, rightItem := range trueRightArr {
			itemMap, ok := rightItem.(map[string]interface{})
			if ok {
				if itemKey, ok := itemMap[rightKey]; ok {
					strVal := fmt.Sprintf("%v", itemKey)

					// Determine the left item to merge with
					var leftItemToMerge interface{}
					if leftMatch, ok := leftMatched[strVal]; ok {
						leftItemToMerge = leftMatch
					} else {
						leftItemToMerge = map[string]interface{}{rightKey: itemKey}
					}

					// Handle right and full join separately to avoid duplication
					if joinType == "right" && !rightProcessed[strVal] {
						mergedItem := mergeItemsNew(leftItemToMerge, rightMap[strVal], rightArrayName)
						result = append(result, mergedItem)
						rightProcessed[strVal] = true
					} else if joinType == "full" && !rightProcessed[strVal] && !alreadyProcessed[strVal] {
						mergedItem := mergeItemsNew(leftItemToMerge, rightMap[strVal], rightArrayName)
						result = append(result, mergedItem)
						rightProcessed[strVal] = true
					}
				}
			}
		}
	}

	return result, nil
}

// use reflect sparingly to avoid performance issues
func mergeStruct(item interface{}) map[string]interface{} {
	mergedItem := make(map[string]interface{})
	val := reflect.ValueOf(item)

	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface()
		mergedItem[fieldName] = fieldValue
	}

	return mergedItem
}

func mergeItemsNew(leftItem interface{}, rightItems []interface{}, rightArrayName string) map[string]interface{} {
	mergedItem := make(map[string]interface{})

	switch left := leftItem.(type) {
	case map[string]interface{}:
		for key, value := range left {
			mergedItem[key] = value
		}
	case nil:
		// skip
	default:
		structFields := mergeStruct(left)
		for key, value := range structFields {
			mergedItem[key] = value
		}
	}

	if len(rightItems) > 0 {
		mergedItem[rightArrayName] = rightItems
	}

	return mergedItem
}
