package jparse

import (
	"fmt"
)

type PositionNode struct {
	Input     Node
	Expr      Node
	Variable  Node
	Predicate Node
}

func (n *PositionNode) optimize() (Node, error) {
	var err error
	if n.Input != nil {
		n.Input, err = n.Input.optimize()
		if err != nil {
			return nil, err
		}
	}
	if n.Expr != nil {
		n.Expr, err = n.Expr.optimize()
		if err != nil {
			return nil, err
		}
	}
	if n.Variable != nil {
		n.Variable, err = n.Variable.optimize()
		if err != nil {
			return nil, err
		}
	}
	if n.Predicate != nil {
		n.Predicate, err = n.Predicate.optimize()
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (n PositionNode) String() string {
	if n.Variable != nil {
		if n.Predicate != nil {
			return fmt.Sprintf("%s#%s[%s]", n.Input, n.Variable, n.Predicate)
		}
		return fmt.Sprintf("%s#%s", n.Input, n.Variable)
	}
	if n.Expr != nil {
		return fmt.Sprintf("%s#%s", n.Input, n.Expr)
	}
	return fmt.Sprintf("%s#", n.Input)
}

func (n PositionNode) Evaluate(ctx *Context) (interface{}, error) {
	if ctx.Position < 0 && n.Input == nil {
		return nil, fmt.Errorf("position operator used outside of sequence context")
	}

	var input interface{}
	var err error

	if n.Input != nil {
		input, err = n.Input.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		input = ctx.Input
	}

	if input == nil {
		return nil, nil
	}

	// Handle array inputs
	items, ok := input.([]interface{})
	if !ok {
		items = []interface{}{input}
	}

	// Handle variable binding with optional predicate
	if n.Variable != nil {
		var results []interface{}
		for i, item := range items {
			itemCtx := NewContext(item, ctx)
			itemCtx.Position = i

			if varNode, ok := n.Variable.(*VariableNode); ok {
				itemCtx = itemCtx.WithVariable(varNode.Name, float64(i))
			}

			if n.Predicate != nil {
				match, err := n.Predicate.Evaluate(itemCtx)
				if err != nil {
					return nil, err
				}
				if b, ok := match.(bool); ok && b {
					results = append(results, float64(i))
				}
			} else {
				results = append(results, float64(i))
			}
		}
		if len(results) == 0 {
			return nil, nil
		}
		if len(results) == 1 {
			return results[0], nil
		}
		return results, nil
	}

	// Handle explicit position expression
	if n.Expr != nil {
		pos, err := n.Expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if num, ok := pos.(float64); ok {
			idx := int(num)
			if idx >= 0 && idx < len(items) {
				return items[idx], nil
			}
		}
		return nil, nil
	}

	// Return current position
	if ctx.Position >= 0 {
		return float64(ctx.Position), nil
	}
	return nil, nil
}
