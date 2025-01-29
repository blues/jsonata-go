package jparse

func parsePath(p *parser, t token, lhs Node) (Node, error) {
	if lhs == nil {
		return nil, newError(ErrPrefix, t)
	}

	rhs := p.parseExpression(p.bp(typeDot))
	if rhs == nil {
		return nil, newError(ErrSyntaxError, t)
	}

	var steps []Node
	if path, ok := lhs.(*PathNode); ok {
		steps = append(steps, path.Steps...)
	} else {
		steps = append(steps, lhs)
	}

	if path, ok := rhs.(*PathNode); ok {
		steps = append(steps, path.Steps...)
	} else {
		steps = append(steps, rhs)
	}

	return &PathNode{Steps: steps}, nil
}
