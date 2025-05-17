// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jsonata

import (
	"encoding/json"
	"testing"
)

// TestTransformFunctions tests the transformation capabilities of JSONata
func TestTransformFunctions(t *testing.T) {
	// Define test data similar to the account.json file
	accountJson := `{
		"Account": {
			"Account Name": "Firefly",
			"Order": [
				{
					"OrderID": "order103",
					"Product": [
						{
							"Product Name": "Bowler Hat",
							"ProductID": 858383,
							"SKU": "0406654608",
							"Description": {
								"Colour": "Purple",
								"Width": 300,
								"Height": 200,
								"Depth": 210,
								"Weight": 0.75
							},
							"Price": 34.45,
							"Quantity": 2
						},
						{
							"Product Name": "Trilby hat",
							"ProductID": 858236,
							"SKU": "0406634348",
							"Description": {
								"Colour": "Orange",
								"Width": 300,
								"Height": 200,
								"Depth": 210,
								"Weight": 0.6
							},
							"Price": 21.67,
							"Quantity": 1
						}
					]
				},
				{
					"OrderID": "order104",
					"Product": [
						{
							"Product Name": "Bowler Hat",
							"ProductID": 858383,
							"SKU": "040657863",
							"Description": {
								"Colour": "Purple",
								"Width": 300,
								"Height": 200,
								"Depth": 210,
								"Weight": 0.75
							},
							"Price": 34.45,
							"Quantity": 4
						},
						{
							"ProductID": 345664,
							"SKU": "0406654603",
							"Product Name": "Cloak",
							"Description": {
								"Colour": "Black",
								"Width": 30,
								"Height": 20,
								"Depth": 210,
								"Weight": 2.0
							},
							"Price": 107.99,
							"Quantity": 1
						}
					]
				}
			]
		}
	}`

	var testData interface{}
	err := json.Unmarshal([]byte(accountJson), &testData)
	if err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	t.Run("SimpleTransform", func(t *testing.T) {
		expr, err := Compile(`Account.Order.Product.{"name": "Product Name", "id": ProductID}`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(testData)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check result is as expected
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		if len(arr) != 4 {
			t.Fatalf("Expected 4 items, got %d", len(arr))
		}

		// Check the first item
		item, ok := arr[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", arr[0])
		}

		if item["name"] != "Bowler Hat" {
			t.Errorf("Expected name 'Bowler Hat', got %v", item["name"])
		}
		if item["id"] != float64(858383) {
			t.Errorf("Expected id 858383, got %v", item["id"])
		}
	})

	t.Run("TransformWithMap", func(t *testing.T) {
		expr, err := Compile(`Account.Order.Product ~> $map(function($p) { { "name": $p."Product Name", "price": $p.Price } })`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(testData)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check result is as expected
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		if len(arr) != 4 {
			t.Fatalf("Expected 4 items, got %d", len(arr))
		}

		// Check the first item
		item, ok := arr[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", arr[0])
		}

		if item["name"] != "Bowler Hat" {
			t.Errorf("Expected name 'Bowler Hat', got %v", item["name"])
		}
		if item["price"] != float64(34.45) {
			t.Errorf("Expected price 34.45, got %v", item["price"])
		}
	})

	t.Run("GroupByTransform", func(t *testing.T) {
		expr, err := Compile(`Account.Order.{
			"id": OrderID,
			"products": Product.{
				"name": "Product Name",
				"price": Price
			}
		}`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(testData)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check result is as expected
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		if len(arr) != 2 {
			t.Fatalf("Expected 2 items, got %d", len(arr))
		}

		// Check the first order
		order, ok := arr[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map, got %T", arr[0])
		}

		if order["id"] != "order103" {
			t.Errorf("Expected id 'order103', got %v", order["id"])
		}

		products, ok := order["products"].([]interface{})
		if !ok {
			t.Fatalf("Expected array for products, got %T", order["products"])
		}

		if len(products) != 2 {
			t.Fatalf("Expected 2 products in first order, got %d", len(products))
		}
	})

	t.Run("TransformWithFilter", func(t *testing.T) {
		expr, err := Compile(`Account.Order.Product[Price > 30].{
			"name": "Product Name",
			"price": Price
		}`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(testData)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check result is as expected
		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("Expected array result, got %T", result)
		}

		// Should only have products with price > 30
		for _, item := range arr {
			product, ok := item.(map[string]interface{})
			if !ok {
				t.Fatalf("Expected map, got %T", item)
			}

			price, ok := product["price"].(float64)
			if !ok {
				t.Fatalf("Expected float64 for price, got %T", product["price"])
			}

			if price <= 30 {
				t.Errorf("Expected price > 30, got %v", price)
			}
		}
	})

	t.Run("ComplexTransform", func(t *testing.T) {
		expr, err := Compile(`{
			"CustomerName": Account."Account Name",
			"Orders": Account.Order.{
				"OrderID": OrderID,
				"TotalPrice": $sum(Product.(Price * Quantity)),
				"Items": Product.{
					"Product": "Product Name",
					"Quantity": Quantity,
					"Price": Price,
					"Total": Price * Quantity
				}
			}
		}`)
		if err != nil {
			t.Fatalf("Failed to compile expression: %v", err)
		}

		result, err := expr.Eval(testData)
		if err != nil {
			t.Fatalf("Failed to evaluate expression: %v", err)
		}

		// Check result is as expected
		obj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		if obj["CustomerName"] != "Firefly" {
			t.Errorf("Expected CustomerName 'Firefly', got %v", obj["CustomerName"])
		}

		orders, ok := obj["Orders"].([]interface{})
		if !ok {
			t.Fatalf("Expected array for Orders, got %T", obj["Orders"])
		}

		if len(orders) != 2 {
			t.Fatalf("Expected 2 orders, got %d", len(orders))
		}

		// Check the first order's total price
		order1, ok := orders[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map for first order, got %T", orders[0])
		}

		// Total price for first order should be 2*34.45 + 1*21.67 = 90.57
		totalPrice, ok := order1["TotalPrice"].(float64)
		if !ok {
			t.Fatalf("Expected float64 for TotalPrice, got %T", order1["TotalPrice"])
		}

		expectedTotal := 2*34.45 + 1*21.67
		if totalPrice < expectedTotal-0.01 || totalPrice > expectedTotal+0.01 {
			t.Errorf("Expected TotalPrice around %.2f, got %.2f", expectedTotal, totalPrice)
		}
	})
}