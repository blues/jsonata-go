package jsonata

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	// your JSON data
	data1 := `{
  "employees": [
    {
      "firstName": "John",
      "lastName": "Doe",
      "department": "Sales",
      "salary": 50000,
      "joining_date": "2020-05-12",
      "details": {
        "address": "123 Main St",
        "city": "New York",
        "state": "NY"
      }
    },
    {
      "firstName": "Anna",
      "lastName": "Smith",
      "department": "Marketing",
      "salary": 60000,
      "joining_date": "2019-07-01",
      "details": {
        "address": "456 Market St",
        "city": "San Francisco",
        "state": "CA"
      }
    },
    {
      "firstName": "Peter",
      "lastName": "Jones",
      "department": "Sales",
      "salary": 70000,
      "joining_date": "2021-01-20",
      "details": {
        "address": "789 Broad St",
        "city": "Los Angeles",
        "state": "CA"
      }
    }
  ]
}`
	var data interface{}

	// Decode JSON.
	err := json.Unmarshal([]byte(data1), &data)
	assert.NoError(t, err)
	t.Run("wrong arithmetic errors", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.firstName + 5")

		// Evaluate.
		_, err := e.Eval(data)
		assert.Error(t, err, "left side of the \"value:+, position: 20\" operator must evaluate to a number")
	})
	t.Run("Cannot call non-function token:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.details.state.$address()")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "cannot call non-function token:$address, position: 25")

	})
	t.Run("Trying to get the maximum of a string field:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$max(employees.firstName)")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "err: cannot call max on an array with non-number types, possition: 1, arguments: number:0 value:[John Anna Peter] ")

	})
	t.Run("Invalid Function Call on a non-array field:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.department.$count()")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "err: function \"count\" takes 1 argument(s), got 0, possition: 22, arguments: ")

	})
	t.Run("Cannot use wildcard on non-object type:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.*.salary")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "err:no results found, token: employees.*.salary, possition: 0")
	})
	t.Run("Indexing on non-array type:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.firstName[1]")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "err:no results found, token: employees.firstName[1], possition: 0")

	})
	t.Run("Use of an undefined variable:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$undefinedVariable")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "err:no results found, token: $undefinedVariable, possition: 1")

	})
	t.Run("Use of an undefined function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$undefinedFunction()")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "cannot call non-function token:$undefinedFunction, position: 1")

	})
	t.Run("Comparison of incompatible types:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.firstName > employees.salary")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "left side of the \"value:>, position: 0\" operator must evaluate to a number or string")

	})
	t.Run("Use of an invalid JSONata operator:", func(t *testing.T) {

		// Create expression.
		_, err := Compile("employees ! employees")
		assert.EqualError(t, err, "syntax error: '', character position 10")
	})
	t.Run("Incorrect use of the reduce function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$reduce(employees.firstName, function($acc, $val) { $acc + $val })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "err: left side of the \"value:+, position: 57\" operator must evaluate to a number, possition: 1, arguments: number:0 value:[John Anna Peter]")

	})
	t.Run("Incorrect use of the map function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$map(employees, function($employee) { $employee.firstName + 5 })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "err: left side of the \"value:+, position: 58\" operator must evaluate to a number, possition: 1, arguments: number:0 value:[map[department:Sales details:map[address:123 Main St city:New York state:NY] firstName:John joining_date:2020-05-12")

	})
	t.Run("Incorrect use of the filter function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$filter(employees, function($employee) { $employee.salary.$uppercase() })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "err: err: argument 1 of function \"uppercase\" does not match function signature, possition: 59, arguments: , possition: 1, arguments: number:0 value:[map[department:Sales details:map[address:123 Main St city:New York state:NY] firstName:John joining_date:2020-05-12 lastName:Doe salary:50000")

	})
	t.Run("Incorrect use of the join function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$join(employees.firstName, 5)")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "err: argument 2 of function \"join\" does not match function signature, possition: 1, arguments: number:0 value:[John Anna Peter] number:1 value:5")

	})
}
