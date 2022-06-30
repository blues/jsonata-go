package jsonata

import (
	"testing"
  "encoding/json"

	"github.com/stretchr/testify/assert"
)

var (
  jsonArr = `[
  {
    "code": "123141141",
    "eans": [
      {
        "ean": 123124,
        "name": "jim"
      },
      {
        "ean": 123128,
        "name": "juliet"
      }
    ]
  },
  {
    "code": "123141141",
    "eans": [
      {
        "ean": 123124,
        "name": "jim"
      },
      {
        "ean": 123128,
        "name": "juliet"
      }
    ]
  },
  {
    "code": "123142142",
    "eans": {
      "ean": 123124,
      "name": "toby"
    }
  }
]`
  notArray = `{"i": "am not an array"}`
)

/*
  the above array has two duplicate (unnecessary) objects
  and the below method will produce a distinct output
*/
func Test_Distinct(t *testing.T) {
  t.Run("an array of data with duplicates in it - will be made distinct", func(t *testing.T) {
	var (
		data               interface{}
		expectedOutputData = []interface{}([]interface{}{map[string]interface{}{"code": "123141141",
			"eans": []interface{}{map[string]interface{}{"ean": float64(123124), "name": "jim"},
				map[string]interface{}{"ean": float64(123128), "name": "juliet"}}}, map[string]interface{}{"code": "123142142",
			"eans": map[string]interface{}{"ean": float64(123124), "name": "toby"}}})
	)

	// Decode JSON.
	err := json.Unmarshal([]byte(jsonArr), &data)
	if err != nil {
		t.Fail()
	}

	distinctData, err := Distinct(data)
	if err != nil {
    t.Fail()
	}

	assert.Equal(t, distinctData, expectedOutputData)
  })

  t.Run("NOT an array - error will be raised", func(t *testing.T) {
    var (
      data               interface{}
    )
  
    // Decode JSON.
    err := json.Unmarshal([]byte(notArray), &data)
    if err != nil {
      t.Fail()
    }
  
    distinctData, err := Distinct(data)
    if err == nil {
      t.Fail()
    }
  
    assert.Equal(t, distinctData, nil)
    })
}
