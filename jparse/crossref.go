package jparse

import (
	"fmt"
)

type CrossReferenceNode struct {
	LHS  Node
	RHS  Node
	Path Node
}

func (n *CrossReferenceNode) optimize() (Node, error) {
	var err error
	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}
	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}
	if n.Path != nil {
		n.Path, err = n.Path.optimize()
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (n CrossReferenceNode) String() string {
	if n.Path != nil {
		return fmt.Sprintf("%s@%s.%s", n.LHS, n.RHS, n.Path)
	}
	return fmt.Sprintf("%s@%s", n.LHS, n.RHS)
}

func (n CrossReferenceNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if lhs == nil {
		return nil, nil
	}

	// Handle array inputs by applying cross reference to each element
	if arr, ok := lhs.([]interface{}); ok {
		var results []interface{}
		for i, item := range arr {
			itemCtx := NewContext(item, ctx)
			itemCtx.Position = i
			
			// Extract variable name and bind it
			if varNode, ok := n.RHS.(*VariableNode); ok {
				itemCtx = itemCtx.WithVariable(varNode.Name, item)
				
				// If there's a predicate after the variable, evaluate it
				if pred := varNode.Next; pred != nil {
					predCtx := itemCtx.WithInput(item)
					result, err := pred.Evaluate(predCtx)
					if err != nil {
						return nil, err
					}
					
					// Only include items that match the predicate
					if b, ok := result.(bool); ok && b {
						if n.Path != nil {
							pathResult, err := n.Path.Evaluate(itemCtx)
							if err != nil {
								return nil, err
							}
							if pathResult != nil {
								results = append(results, pathResult)
							}
						} else {
							results = append(results, item)
						}
					}
					continue
				}
				
				if n.Path != nil {
					pathResult, err := n.Path.Evaluate(itemCtx)
					if err != nil {
						return nil, err
					}
					if pathResult != nil {
						results = append(results, pathResult)
					}
				} else {
					results = append(results, item)
				}
				continue
			}
			
			// Handle non-variable RHS expressions
			result, err := n.RHS.Evaluate(itemCtx)
			if err != nil {
				return nil, err
			}
			if result != nil {
				if n.Path != nil {
					pathCtx := NewContext(result, itemCtx)
					pathResult, err := n.Path.Evaluate(pathCtx)
					if err != nil {
						return nil, err
					}
					if pathResult != nil {
						results = append(results, pathResult)
					}
				} else {
					results = append(results, result)
				}
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

	// Create new context with LHS as input and bind variable
	rhsCtx := NewContext(lhs, ctx)
	
	// Handle variable binding
	if varNode, ok := n.RHS.(*VariableNode); ok {
		rhsCtx = rhsCtx.WithVariable(varNode.Name, lhs)
		
		// If there's a predicate after the variable, evaluate it
		if pred := varNode.Next; pred != nil {
			predCtx := rhsCtx.WithInput(lhs)
			result, err := pred.Evaluate(predCtx)
			if err != nil {
				return nil, err
			}
			
			// Only return value if predicate matches
			if b, ok := result.(bool); ok && b {
				if n.Path != nil {
					return n.Path.Evaluate(rhsCtx)
				}
				return lhs, nil
			}
			return nil, nil
		}
		
		if n.Path != nil {
			return n.Path.Evaluate(rhsCtx)
		}
		return lhs, nil
	}
	
	result, err := n.RHS.Evaluate(rhsCtx)
	if err != nil {
		return nil, err
	}

	if n.Path != nil && result != nil {
		pathCtx := NewContext(result, rhsCtx)
		return n.Path.Evaluate(pathCtx)
	}

	return result, nil
}
