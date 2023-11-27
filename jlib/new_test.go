package jlib

import (
	"encoding/json"
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
