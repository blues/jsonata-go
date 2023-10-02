package jsonata

import (
	"fmt"
)

type JsonataProcessor struct {
	tree *Expr
}

func NewProcessor(jsonataString string) (j *JsonataProcessor, err error) {
	defer func() { // go-jsonata uses panic fallthrough design so this is necessary
		if r := recover(); r != nil {
			err = fmt.Errorf("jsonata error: %v", r)
		}
	}()

	jsnt := replaceQuotesAndCommentsInPaths(jsonataString)

	e := MustCompile(jsnt)

	j = &JsonataProcessor{}

	j.tree = e

	return j, err
}

// Execute - helper function that lets you parse and run jsonata scripts against an object
func (j *JsonataProcessor) Execute(input interface{}) (output []map[string]interface{}, err error) {
	defer func() { // go-jsonata uses panic fallthrough design so this is necessary
		if r := recover(); r != nil {
			err = fmt.Errorf("jsonata error: %v", r)
		}
	}()

	output = make([]map[string]interface{}, 0)

	item, err := j.tree.Eval(input)
	if err != nil {
		return nil, err
	}

	if aMap, ok := item.(map[string]interface{}); ok {
		output = append(output, aMap)

		return output, nil
	}

	if aList, ok := item.([]interface{}); ok {
		for index := range aList {
			if aMap, ok := aList[index].(map[string]interface{}); ok {
				output = append(output, aMap)
			}
		}

		return output, nil
	}

	if aList, ok := item.([]map[string]interface{}); ok {
		return aList, nil
	}

	return output, nil
}
