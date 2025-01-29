package jparse

import (
	"fmt"
)

type ParentNode struct {
	Expr  Node
	Input Node
}

func (n *ParentNode) optimize() (Node, error) {
	var err error
	if n.Expr != nil {
		n.Expr, err = n.Expr.optimize()
		if err != nil {
			return nil, err
		}
	}
	if n.Input != nil {
		n.Input, err = n.Input.optimize()
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (n ParentNode) String() string {
	if n.Input != nil {
		return fmt.Sprintf("%s.%s", n.Input, n.Expr)
	}
	return fmt.Sprintf("%%%s", n.Expr)
}

func (n ParentNode) Evaluate(ctx *Context) (interface{}, error) {
	// Handle modulo operator case
	if n.Input != nil {
		lhs, err := n.Input.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		if lhs == nil {
			return nil, nil
		}
		
		// Create context for evaluating RHS
		rhsCtx := NewContext(lhs, ctx)
		rhs, err := n.Expr.Evaluate(rhsCtx)
		if err != nil {
			return nil, err
		}
		
		// Handle numeric modulo operation
		if ln, ok := lhs.(float64); ok {
			if rn, ok := rhs.(float64); ok {
				if rn == 0 {
					return nil, fmt.Errorf("division by zero in modulo operation")
				}
				return float64(int(ln) % int(rn)), nil
			}
		}
		
		// Handle path access
		if path, ok := n.Expr.(*NameNode); ok {
			if obj, ok := lhs.(map[string]interface{}); ok {
				return obj[path.Value], nil
			}
		}
		
		return nil, fmt.Errorf("invalid operands for modulo/parent operator")
	}
	
	// Handle parent operator case
	if ctx.Parent == nil {
		return nil, fmt.Errorf("parent operator used in root context")
	}
	
	parentCtx := ctx.Parent
	if n.Expr == nil {
		return parentCtx.Input, nil
	}
	
	// Create new context with parent's input and variables
	newCtx := NewContext(parentCtx.Input, parentCtx)
	newCtx.Position = ctx.Position
	
	// Handle array inputs by applying parent operator to each element
	if arr, ok := parentCtx.Input.([]interface{}); ok {
		var results []interface{}
		for _, item := range arr {
			itemCtx := NewContext(item, parentCtx)
			result, err := n.Expr.Evaluate(itemCtx)
			if err != nil {
				return nil, err
			}
			if result != nil {
				results = append(results, result)
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
	
	result, err := n.Expr.Evaluate(newCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}
