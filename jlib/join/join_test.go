package join

import (
	"github.com/goccy/go-json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneToManyJoin(t *testing.T) {
	tests := []struct {
		description    string
		object1        string
		object2        string
		joinStr1       string
		joinStr2       string
		joinType       string
		expectedOutput string
		hasError       bool
	}{
		{
			description:    "one to many join on key 'id'",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - left side not an array",
			object1:        `{"id":1,"age":5}`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - right side not an array",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `{"id":1,"name":"Tim"}`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, ["1", "2"]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, [{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"1","Price":29.99},{"ProductID":"2","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "left",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"},{\"Price\":29.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":39.99,\"ProductID\":\"2\"}]}]",
		},
		{
			description:    "one to many left join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "left",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ID\":4,\"Name\":\"Item2\"}]",
		},
		{
			description:    "one to many right join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "right",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ProductID\":\"3\",\"example\":[{\"Price\":24.99,\"ProductID\":\"3\"},{\"Price\":39.99,\"ProductID\":\"3\"}]}]",
		},
		{
			description:    "one to many full join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "full",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ID\":4,\"Name\":\"Item2\"},{\"ProductID\":\"3\",\"example\":[{\"Price\":24.99,\"ProductID\":\"3\"},{\"Price\":39.99,\"ProductID\":\"3\"}]}]",
		},
		{
			description:    "one to many inner join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "inner",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]}]",
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

			output, err := OneToManyJoin(o1, o2, tt.joinStr1, tt.joinStr2, "example", tt.joinType)
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

func TestOneToManyJoinNew(t *testing.T) {
	tests := []struct {
		description    string
		object1        string
		object2        string
		joinStr1       string
		joinStr2       string
		joinType       string
		expectedOutput string
		hasError       bool
	}{
		{
			description:    "one to many join on key 'id'",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - left side not an array",
			object1:        `{"id":1,"age":5}`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - right side not an array",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `{"id":1,"name":"Tim"}`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "null",
			hasError:       true,
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, ["1", "2"]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join on key 'id' - has a nested different type - should ignore",
			object1:        `[{"id":1,"age":5}]`,
			object2:        `[{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}, [{"id":1,"name":"Tim"},{"id":1,"name":"Tam"}]]`,
			joinStr1:       "id",
			joinStr2:       "id",
			joinType:       "left",
			expectedOutput: "[{\"age\":5,\"example\":[{\"id\":1,\"name\":\"Tim\"},{\"id\":1,\"name\":\"Tam\"}],\"id\":1}]",
		},
		{
			description:    "one to many join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"1","Price":29.99},{"ProductID":"2","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "left",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"},{\"Price\":29.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":39.99,\"ProductID\":\"2\"}]}]",
		},
		{
			description:    "one to many left join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "left",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ID\":4,\"Name\":\"Item2\"}]",
		},
		{
			description:    "one to many right join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "right",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ProductID\":\"3\",\"example\":[{\"Price\":24.99,\"ProductID\":\"3\"},{\"Price\":39.99,\"ProductID\":\"3\"}]}]",
		},
		{
			description:    "one to many full join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "full",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]},{\"ID\":4,\"Name\":\"Item2\"},{\"ProductID\":\"3\",\"example\":[{\"Price\":24.99,\"ProductID\":\"3\"},{\"Price\":39.99,\"ProductID\":\"3\"}]}]",
		},
		{
			description:    "one to many inner join - complex",
			object1:        `[{"ID":1,"Name":"Item1"},{"ID":2,"Name":"Item2"},{"ID":4,"Name":"Item2"}]`,
			object2:        `[{"ProductID":"1","Price":19.99},{"ProductID":"2","Price":29.99},{"ProductID":"2","Price":12.99},{"ProductID":"3","Price":24.99},{"ProductID":"3","Price":39.99}]`,
			joinStr1:       "ID",
			joinStr2:       "ProductID",
			joinType:       "inner",
			expectedOutput: "[{\"ID\":1,\"Name\":\"Item1\",\"example\":[{\"Price\":19.99,\"ProductID\":\"1\"}]},{\"ID\":2,\"Name\":\"Item2\",\"example\":[{\"Price\":29.99,\"ProductID\":\"2\"},{\"Price\":12.99,\"ProductID\":\"2\"}]}]",
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

			output, err := OneToManyJoin2(o1, o2, tt.joinStr1, tt.joinStr2, "example", tt.joinType)
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
