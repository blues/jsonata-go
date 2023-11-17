package jlib

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

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

func TestOneToManyJoin(t *testing.T) {
	tests := []struct {
		description    string
		object1        string
		object2        string
		joinStr1       string
		joinStr2       string
		expectedOutput string
		hasError       bool
	}{
		{
			description:    "one to many join on key 'id'",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - left side not an array",
			object1:        `{"id":1,"age":5}`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - right side not an array",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `{"id":1,"name":"Tim"}`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, ["1", "2"]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, [{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"1","Price":29.99},{"ProductID":"2","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"},{\"Price\":29.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":39.99,\"ProductID\":\"2\"}]}]",
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

			output, err := OneToManyJoin(o1, o2, tt.joinStr1, tt.joinStr2, "example")
			assert.Equal(t, err != nil, tt.hasError)
			if err != nil {
				log.Println(tt.description, "|", err)
			}

			bytes, err := json.Marshal(output)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(bytes))
		})
	}
}
