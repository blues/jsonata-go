package jlib

import (
	"reflect"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
)

func TestSJoin(t *testing.T) {
	tests := []struct {
		description    string
		object1        string
		object2        string
		joinStr1       string
		joinStr2       string
		expectedOutput string
	}{
		{
			description: "simple join",
			object1: `[{"test": {
				"id": 1,
				"age": 5
				}}]`,
			object2: `[{"test": {
					"id": 1,
					"name": "Tim"
					}}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "[{\"test\":{\"age\":5,\"id\":1}}]",
		},
		{
			description: "nested join",
			object1: `[
				{
				 "age": 5,
				 "id": 1
				}
			   ]`,
			object2: `[
				{
				 "test": {
				  "id": 1,
				  "name": "Tim"
				 }
				}
			   ]`,
			joinStr1:       "id",
			joinStr2:       "test¬id",
			expectedOutput: "[{\"age\":5,\"id\":1,\"test\":{\"id\":1,\"name\":\"Tim\"}}]",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.description, func(t *testing.T) {
			var o1, o2 interface{}

			err := json.Unmarshal([]byte(tt.object1), &o1)
			assert.NoError(t, err)
			err = json.Unmarshal([]byte(tt.object2), &o2)
			assert.NoError(t, err)

			i1 := reflect.ValueOf(o1)
			i2 := reflect.ValueOf(o2)

			output, err := SimpleJoin(i1, i2, tt.joinStr1, tt.joinStr2)
			assert.NoError(t, err)

			bytes, err := json.Marshal(output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(bytes))
		})
	}
}

func TestRenameKeys(t *testing.T) {
	tests := []struct {
		description    string
		object1        string
		object2        string
		expectedOutput string
	}{
		{
			description: "Rename Keys",
			object1: `{
				"itemLineId": "1",
				"unitPrice": 104.5,
				"percentageDiscountValue": 5,
				"discountedLinePrice": 104.5,
				"name": "LD Wrong Price",
				"discountAmount": 5.5,
				"discountType": "AMOUNT",
				"discountReasonCode": "9901",
				"discountReasonName": "LD Wrong Price"
			}`,
			object2:        `[["ReasonCode","reasonCode"],["ReasonName","reasonName"],["DiscountValue","value"],["Amount","amount"],["Type","type"],["LinePrice","linePrice"]]`,
			expectedOutput: "{\"amount\":5.5,\"itemLineId\":\"1\",\"linePrice\":104.5,\"name\":\"LD Wrong Price\",\"reasonCode\":\"9901\",\"reasonName\":\"LD Wrong Price\",\"type\":\"AMOUNT\",\"unitPrice\":104.5,\"value\":5}",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.description, func(t *testing.T) {
			var o1, o2 interface{}

			err := json.Unmarshal([]byte(tt.object1), &o1)
			assert.NoError(t, err)
			err = json.Unmarshal([]byte(tt.object2), &o2)
			assert.NoError(t, err)

			output, err := RenameKeys(o1, o2)
			assert.NoError(t, err)

			bytes, err := json.Marshal(output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(bytes))
		})
	}
}

func TestSetValue_ArrayIndexing(t *testing.T) {
	tests := []struct {
		description    string
		input          map[string]interface{}
		code           string
		value          interface{}
		expectedOutput map[string]interface{}
	}{
		{
			description: "Set simple value at array index",
			input:       map[string]interface{}{},
			code:        "employees[0].name",
			value:       "Alice",
			expectedOutput: map[string]interface{}{
				"employees": []interface{}{
					map[string]interface{}{"name": "Alice"},
				},
			},
		},
		{
			description: "Extend array to meet required index",
			input:       map[string]interface{}{},
			code:        "employees[2].name",
			value:       "Charlie",
			expectedOutput: map[string]interface{}{
				"employees": []interface{}{
					map[string]interface{}{},
					map[string]interface{}{},
					map[string]interface{}{"name": "Charlie"},
				},
			},
		},
		{
			description: "Nested arrays and objects",
			input:       map[string]interface{}{},
			code:        "company.departments[1].staff[0].role",
			value:       "Manager",
			expectedOutput: map[string]interface{}{
				"company": map[string]interface{}{
					"departments": []interface{}{
						map[string]interface{}{},
						map[string]interface{}{
							"staff": []interface{}{
								map[string]interface{}{"role": "Manager"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			setValue(tt.input, tt.code, tt.value)
			assert.Equal(t, tt.expectedOutput, tt.input)
		})
	}
}

func TestObjectsToDocument_ValPriority(t *testing.T) {
	tests := []struct {
		description    string
		input          string
		expectedOutput string
	}{
		{
			description: "Use Val when present",
			input: `[
				{"Code":"person.name","Val":"Alice","Value":"ShouldNotUse"},
				{"Code":"person.age","Val":30}
			]`,
			expectedOutput: `{"person":{"age":30,"name":"Alice"}}`,
		},
		{
			description: "Use Value when Val not present",
			input: `[
				{"Code":"person.name","Value":"Bob"},
				{"Code":"person.age","Val":25}
			]`,
			// "person.name" should come from Value since Val is not present
			expectedOutput: `{"person":{"age":25,"name":"Bob"}}`,
		},
		{
			description: "Array indexing in Code with Val",
			input: `[
				{"Code":"people[0].name","Val":"Carol"},
				{"Code":"people[1].name","Value":"Dave"}
			]`,
			expectedOutput: `{"people":[{"name":"Carol"},{"name":"Dave"}]}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			var inputData interface{}
			err := json.Unmarshal([]byte(tt.input), &inputData)
			assert.NoError(t, err)

			output, err := ObjectsToDocument(inputData)
			assert.NoError(t, err)

			bytes, err := json.Marshal(output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(bytes))
		})
	}
}

func TestObjectsToDocument_ComplexArrayPaths(t *testing.T) {
	tests := []struct {
		description    string
		input          string
		expectedOutput string
	}{
		{
			description: "Set multiple nested array values",
			input: `[
				{"Code":"root.list[0].items[0].value","Val":"Item0-0"},
				{"Code":"root.list[0].items[1].value","Value":"Item0-1"},
				{"Code":"root.list[1].items[0].value","Val":"Item1-0"}
			]`,
			expectedOutput: `{"root":{"list":[{"items":[{"value":"Item0-0"},{"value":"Item0-1"}]},{"items":[{"value":"Item1-0"}]}]}}`,
		},
		{
			description: "No Val or Value",
			input: `[
				{"Code":"topKey","Whatever":"none"},
				{"Code":"anotherKey","Val":null}
			]`,
			// No Val or Value means keys are not set
			expectedOutput: `{}`,
		},
		{
			description: "Val is nil, use Value",
			input: `[
				{"Code":"testKey","Val":null,"Value":"RealValue"}
			]`,
			expectedOutput: `{"testKey":"RealValue"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.description, func(t *testing.T) {
			var inputData interface{}
			err := json.Unmarshal([]byte(tt.input), &inputData)
			assert.NoError(t, err)

			output, err := ObjectsToDocument(inputData)
			assert.NoError(t, err)

			bytes, err := json.Marshal(output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(bytes))
		})
	}
}
