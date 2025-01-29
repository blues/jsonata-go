package jparse

import (
	"testing"
)

// Using existing deepEqual from node.go

func TestOperatorEvaluation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		data    interface{}
		want    interface{}
		wantErr bool
	}{
		// Parent operator tests
		{
			name:  "parent operator basic",
			input: "Account.%.value",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"child": map[string]interface{}{
						"id": 1,
					},
					"value": "parent",
				},
			},
			want: "parent",
		},
		{
			name:  "parent operator no parent",
			input: "Account.%",
			data: map[string]interface{}{
				"test": "value",
			},
			wantErr: true,
		},
		{
			name:  "parent operator nested",
			input: "Account.Order.Product.%.OrderID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": map[string]interface{}{
						"OrderID": "O1",
						"Product": map[string]interface{}{
							"ProductID": "P1",
						},
					},
				},
			},
			want: "O1",
		},

		// Cross reference operator tests
		{
			name:  "cross reference basic",
			input: "Account.Order.Product@$P.ProductID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": map[string]interface{}{
						"Product": map[string]interface{}{
							"ProductID": "P123",
							"Name":      "Widget",
						},
					},
				},
			},
			want: "P123",
		},
		{
			name:  "cross reference with variable binding",
			input: "Account.Order.Product@$P[$P.Name='Widget'].ProductID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": map[string]interface{}{
						"Product": []interface{}{
							map[string]interface{}{
								"ProductID": "P123",
								"Name":      "Widget",
							},
							map[string]interface{}{
								"ProductID": "P456",
								"Name":      "Gadget",
							},
						},
					},
				},
			},
			want: "P123",
		},
		{
			name:  "cross reference with null",
			input: "missing@something",
			data: map[string]interface{}{
				"test": "value",
			},
			want: nil,
		},
		{
			name:  "cross reference with array",
			input: "Account.Orders@$O.Products@$P.ProductID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P1"},
								map[string]interface{}{"ProductID": "P2"},
							},
						},
					},
				},
			},
			want: []interface{}{"P1", "P2"},
		},

		// Position operator tests
		{
			name:  "position operator basic",
			input: "Account.Order[#=1]",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						map[string]interface{}{"id": "1"},
						map[string]interface{}{"id": "2"},
						map[string]interface{}{"id": "3"},
					},
				},
			},
			want: map[string]interface{}{"id": "2"},
		},
		{
			name:  "position operator with expression",
			input: "Account.Order[#.id='2']",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						map[string]interface{}{"id": "1"},
						map[string]interface{}{"id": "2"},
						map[string]interface{}{"id": "3"},
					},
				},
			},
			want: map[string]interface{}{"id": "2"},
		},
		{
			name:    "position operator outside sequence",
			input:   "Account.Order[#]",
			data:    42,
			wantErr: true,
		},
		{
			name:  "position operator in array transformation",
			input: "Account.Order.$each(function($v, $i) { $i })",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						"first",
						"second",
						"third",
					},
				},
			},
			want: []interface{}{0.0, 1.0, 2.0},
		},

		// Combined operator tests
		{
			name:  "combined operators basic",
			input: "Account.Order#$i.Product@$P[%.OrderID]",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "O1",
							"Product": map[string]interface{}{
								"ProductID": "P1",
							},
						},
					},
				},
			},
			want: "O1",
		},
		{
			name:  "combined operators with predicates",
			input: "Account.Order#$i[%@$O.Type='retail'].Product@$P[%.OrderID]",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID": "O1",
							"Type":    "retail",
							"Product": map[string]interface{}{
								"ProductID": "P1",
							},
						},
						map[string]interface{}{
							"OrderID": "O2",
							"Type":    "wholesale",
							"Product": map[string]interface{}{
								"ProductID": "P2",
							},
						},
					},
				},
			},
			want: "O1",
		},
		{
			name:  "nested parent references",
			input: "Account.Order.Product.{name: ProductName, order: %.OrderID, account: %%.AccountID}",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"AccountID": "A1",
					"Order": map[string]interface{}{
						"OrderID": "O1",
						"Product": map[string]interface{}{
							"ProductName": "Widget",
						},
					},
				},
			},
			want: map[string]interface{}{
				"name":    "Widget",
				"order":   "O1",
				"account": "A1",
			},
		},
		{
			name:  "position with parent and cross reference",
			input: "Account.Order#$i[Position=#].Product@$P[%.OrderID]",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": []interface{}{
						map[string]interface{}{
							"OrderID":  "O1",
							"Position": 0,
							"Product": map[string]interface{}{
								"ProductID": "P1",
							},
						},
						map[string]interface{}{
							"OrderID":  "O2",
							"Position": 1,
							"Product": map[string]interface{}{
								"ProductID": "P2",
							},
						},
					},
				},
			},
			want: "O1",
		},
		{
			name:  "modulo operator",
			input: "10 % 3",
			data:  nil,
			want:  float64(1),
		},
		{
			name:  "modulo with variables",
			input: "$a % $b",
			data: map[string]interface{}{
				"a": float64(10),
				"b": float64(3),
			},
			want: float64(1),
		},
		{
			name:  "parent operator vs modulo disambiguation",
			input: "Account.Order.%.OrderID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Order": map[string]interface{}{
						"OrderID": "O1",
						"Product": map[string]interface{}{
							"ProductID": "P1",
						},
					},
				},
			},
			want: "O1",
		},
		{
			name:  "parent operator with array context",
			input: "Account.Orders.Products.%.OrderID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{
							"OrderID": "O1",
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P1"},
								map[string]interface{}{"ProductID": "P2"},
							},
						},
					},
				},
			},
			want: []interface{}{"O1", "O1"},
		},
		{
			name:  "cross reference with array binding",
			input: "Account.Orders@$O.Products@$P[$O.Type='retail'].ProductID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{
							"OrderID": "O1",
							"Type":    "retail",
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P1"},
								map[string]interface{}{"ProductID": "P2"},
							},
						},
						map[string]interface{}{
							"OrderID": "O2",
							"Type":    "wholesale",
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P3"},
							},
						},
					},
				},
			},
			want: []interface{}{"P1", "P2"},
		},
		{
			name:  "combined operators complex",
			input: "Account.Orders@$O#$i.Products@$P[%.OrderID]",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{
							"OrderID": "O1",
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P1"},
							},
						},
						map[string]interface{}{
							"OrderID": "O2",
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P2"},
							},
						},
					},
				},
			},
			want: []interface{}{"O1", "O2"},
		},
		{
			name:  "position operator with array transformation",
			input: "Account.Orders#$i.{id: OrderID, pos: $i}",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{"OrderID": "O1"},
						map[string]interface{}{"OrderID": "O2"},
					},
				},
			},
			want: []interface{}{
				map[string]interface{}{"id": "O1", "pos": float64(0)},
				map[string]interface{}{"id": "O2", "pos": float64(1)},
			},
		},
		{
			name:  "position operator with parent context",
			input: "Account.Orders#$i.Products[%.Position=$i].ProductID",
			data: map[string]interface{}{
				"Account": map[string]interface{}{
					"Orders": []interface{}{
						map[string]interface{}{
							"Position": float64(0),
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P1"},
								map[string]interface{}{"ProductID": "P2"},
							},
						},
						map[string]interface{}{
							"Position": float64(1),
							"Products": []interface{}{
								map[string]interface{}{"ProductID": "P3"},
							},
						},
					},
				},
			},
			want: []interface{}{"P1", "P2", "P3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			ctx := NewContext(tt.data, nil)
			got, err := expr.Evaluate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatorLexer(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []token
	}{
		{
			name:  "parent operator",
			input: "%.name",
			want: []token{
				{Type: typeParent, Value: "%", Position: 0},
				{Type: typeDot, Value: ".", Position: 1},
				{Type: typeName, Value: "name", Position: 2},
			},
		},
		{
			name:  "cross reference operator",
			input: "books@$B",
			want: []token{
				{Type: typeName, Value: "books", Position: 0},
				{Type: typeCrossRef, Value: "@", Position: 5},
				{Type: typeVariable, Value: "B", Position: 6},
			},
		},
		{
			name:  "position operator",
			input: "Order#3",
			want: []token{
				{Type: typeName, Value: "Order", Position: 0},
				{Type: typePosition, Value: "#", Position: 5},
				{Type: typeNumber, Value: "3", Position: 6},
			},
		},
		{
			name:  "combined operators",
			input: "Account.Order#$i.Product@$P[%.OrderID]",
			want: []token{
				{Type: typeName, Value: "Account", Position: 0},
				{Type: typeDot, Value: ".", Position: 7},
				{Type: typeName, Value: "Order", Position: 8},
				{Type: typePosition, Value: "#", Position: 13},
				{Type: typeVariable, Value: "i", Position: 14},
				{Type: typeDot, Value: ".", Position: 15},
				{Type: typeName, Value: "Product", Position: 16},
				{Type: typeCrossRef, Value: "@", Position: 23},
				{Type: typeVariable, Value: "P", Position: 24},
				{Type: typeBracketOpen, Value: "[", Position: 25},
				{Type: typeParent, Value: "%", Position: 26},
				{Type: typeDot, Value: ".", Position: 27},
				{Type: typeName, Value: "OrderID", Position: 28},
				{Type: typeBracketClose, Value: "]", Position: 35},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			l := newLexer(tt.input)
			var got []token

			for {
				tok := l.next(true)
				if tok.Type == typeEOF {
					break
				}
				got = append(got, tok)
			}

			if len(got) != len(tt.want) {
				t.Errorf("got %d tokens, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Type != tt.want[i].Type {
					t.Errorf("token[%d].Type = %v, want %v", i, got[i].Type, tt.want[i].Type)
				}
				if got[i].Value != tt.want[i].Value {
					t.Errorf("token[%d].Value = %q, want %q", i, got[i].Value, tt.want[i].Value)
				}
				if got[i].Position != tt.want[i].Position {
					t.Errorf("token[%d].Position = %d, want %d", i, got[i].Position, tt.want[i].Position)
				}
			}
		})
	}
}
