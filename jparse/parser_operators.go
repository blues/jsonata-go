package jparse

func parseParentPrefix(p *parser, t token) (Node, error) {
	// Parent operator should not be used as prefix
	return nil, newError(ErrPrefix, t)
}

func parsePositionPrefix(p *parser, t token) (Node, error) {
	expr := p.parseExpression(p.bp(typePosition))
	if expr == nil {
		return &PositionNode{}, nil
	}
	return &PositionNode{Expr: expr}, nil
}

func parseParentOperator(p *parser, t token, lhs Node) (Node, error) {
	if lhs == nil {
		return nil, newError(ErrPrefix, t)
	}

	// Check if this is a modulo operator in numeric context
	if p.token.Type == typeNumber || p.token.Type == typeVariable || p.token.Type == typeName {
		rhs := p.parseExpression(p.bp(typeMod))
		if rhs == nil {
			return nil, newError(ErrSyntaxError, t)
		}
		return &BinaryNode{
			Op:    typeMod,
			Left:  lhs,
			Right: rhs,
		}, nil
	}

	// Handle parent operator with dot notation
	if p.token.Type == typeDot {
		p.advance(false)
		if p.token.Type != typeName {
			return nil, newError(ErrSyntaxError, t)
		}
		expr := &NameNode{Value: p.token.Value}
		p.advance(false)
		return &ParentNode{
			Expr:  expr,
			Input: lhs,
		}, nil
	}

	// Handle parent operator with bracket notation
	if p.token.Type == typeBracketOpen {
		p.advance(false)
		expr := p.parseExpression(0)
		if expr == nil {
			return nil, newError(ErrSyntaxError, t)
		}
		if p.token.Type != typeBracketClose {
			return nil, newError(ErrSyntaxError, p.token)
		}
		p.advance(false)
		return &ParentNode{
			Expr:  expr,
			Input: lhs,
		}, nil
	}

	return &ParentNode{Input: lhs}, nil
}

func parseCrossReferenceOperator(p *parser, t token, lhs Node) (Node, error) {
	if lhs == nil {
		return nil, newError(ErrPrefix, t)
	}

	// Parse variable after @ operator
	if p.token.Type != typeVariable {
		return nil, newError(ErrSyntaxError, p.token)
	}

	varName := p.token.Value
	p.advance(false)

	var path Node
	if p.token.Type == typeBracketOpen {
		p.advance(false)
		expr := p.parseExpression(0)
		if expr == nil {
			return nil, newError(ErrSyntaxError, t)
		}
		if p.token.Type != typeBracketClose {
			return nil, newError(ErrSyntaxError, p.token)
		}
		p.advance(false)
		path = expr
	} else if p.token.Type == typeDot {
		p.advance(false)
		if p.token.Type != typeName {
			return nil, newError(ErrSyntaxError, p.token)
		}
		path = &NameNode{Value: p.token.Value}
		p.advance(false)
	}

	return &CrossReferenceNode{
		LHS:  lhs,
		RHS:  &VariableNode{Name: varName},
		Path: path,
	}, nil
}

func parsePositionOperator(p *parser, t token, lhs Node) (Node, error) {
	if lhs == nil {
		return nil, newError(ErrPrefix, t)
	}

	// Handle position operator with variable binding
	if p.token.Type == typeVariable {
		varName := p.token.Value
		p.advance(false)

		// Parse optional predicate
		var predicate Node
		if p.token.Type == typeBracketOpen {
			p.advance(false)
			expr := p.parseExpression(0)
			if expr == nil {
				return nil, newError(ErrSyntaxError, t)
			}
			if p.token.Type != typeBracketClose {
				return nil, newError(ErrSyntaxError, p.token)
			}
			p.advance(false)
			predicate = expr
		}

		return &PositionNode{
			Input:     lhs,
			Variable:  &VariableNode{Name: varName},
			Predicate: predicate,
		}, nil
	}

	// Parse optional expression after #
	var expr Node
	if p.token.Type == typeNumber || p.token.Type == typeParenOpen {
		expr = p.parseExpression(p.bp(typePosition))
		if expr == nil {
			return nil, newError(ErrSyntaxError, t)
		}
	}

	return &PositionNode{
		Input: lhs,
		Expr:  expr,
	}, nil
}
