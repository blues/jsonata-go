package jsonata

import (
	"github.com/goccy/go-json"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.EqualError(t, err, "cannot call non-function $address, position:25, arguments: ")

	})
	t.Run("Trying to get the maximum of a string field:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$max(employees.firstName)")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "cannot call max on an array with non-number types, position: 1, arguments: number:0 value:[John Anna Peter] ")

	})
	t.Run("Invalid Function Call on a non-array field:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.department.$count()")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "function \"count\" takes 1 argument(s), got 0, position: 22, arguments: ")

	})
	t.Run("Cannot use wildcard on non-object type:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.*.salary")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "no results found")
	})
	t.Run("Indexing on non-array type:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.firstName[1]")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "no results found")

	})
	t.Run("Use of an undefined variable:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$undefinedVariable")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "no results found")

	})
	t.Run("Use of an undefined function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$undefinedFunction()")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "cannot call non-function $undefinedFunction, position:1, arguments: ")

	})
	t.Run("Comparison of incompatible types:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("employees.firstName > employees.salary")

		// Evaluate.
		_, err := e.Eval(data)
		assert.EqualError(t, err, "left side of the \">\" operator must evaluate to a number or string, position:0, arguments: ")

	})
	t.Run("Use of an invalid JSONata operator:", func(t *testing.T) {

		// Create expression.
		_, err := Compile("employees ! employees")
		assert.EqualError(t, err, "syntax error: '', position: 10")
	})
	t.Run("Incorrect use of the reduce function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$reduce(employees.firstName, function($acc, $val) { $acc + $val })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "left side of the \"+\" operator must evaluate to a number, position:57, arguments: ")

	})
	t.Run("Incorrect use of the map function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$map(employees, function($employee) { $employee.firstName + 5 })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "left side of the \"+\" operator must evaluate to a number, position:58")

	})
	t.Run("Incorrect use of the filter function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$filter(employees, function($employee) { $employee.salary.$uppercase() })")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "argument 1 of function \"uppercase\" does not match function signature, position: 1, arguments: number")

	})
	t.Run("Incorrect use of the join function:", func(t *testing.T) {

		// Create expression.
		e := MustCompile("$join(employees.firstName, 5)")

		// Evaluate.
		_, err := e.Eval(data)
		assert.ErrorContains(t, err, "argument 2 of function \"join\" does not match function signature, position: 1, arguments: number:0 value:[John Anna Peter] number:1 value:5 ")

	})
}
