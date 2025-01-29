// Copyright 2018 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package jparse

import (
	"sort"
	"fmt"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Node represents an individual node in a syntax tree.
type Node interface {
	String() string
	optimize() (Node, error)
	Evaluate(ctx *Context) (interface{}, error)
}

// Operator parsing functions moved to parser_operators.go

// A StringNode represents a string literal.
type StringNode struct {
	Value string
}

func parseString(p *parser, t token) (Node, error) {
	if t.Value == "" {
		return nil, newError(ErrSyntaxError, t)
	}

	s, ok := unescape(t.Value)
	if !ok {
		typ := ErrIllegalEscape
		if len(s) > 0 && s[0] == 'u' {
			typ = ErrIllegalEscapeHex
		}
		return nil, newErrorHint(typ, t, s)
	}

	return &StringNode{Value: s}, nil
}

func (n *StringNode) optimize() (Node, error) {
	return n, nil
}

func (n StringNode) String() string {
	return fmt.Sprintf("%q", n.Value)
}

func (n StringNode) Evaluate(ctx *Context) (interface{}, error) {
	return n.Value, nil
}

// A NumberNode represents a number literal.
type NumberNode struct {
	Value float64
}

func parseNumber(p *parser, t token) (Node, error) {

	// Number literals are promoted to type float64.
	n, err := strconv.ParseFloat(t.Value, 64)
	if err != nil {
		typ := ErrInvalidNumber
		if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
			typ = ErrNumberRange
		}
		return nil, newError(typ, t)
	}

	return &NumberNode{
		Value: n,
	}, nil
}

func (n *NumberNode) optimize() (Node, error) {
	return n, nil
}

func (n NumberNode) String() string {
	return fmt.Sprintf("%g", n.Value)
}

func (n NumberNode) Evaluate(ctx *Context) (interface{}, error) {
	return n.Value, nil
}

// A BooleanNode represents the boolean constant true or false.
type BooleanNode struct {
	Value bool
}

func parseBoolean(p *parser, t token) (Node, error) {

	var b bool

	switch t.Value {
	case "true":
		b = true
	case "false":
		b = false
	default: // should be unreachable
		panicf("parseBoolean: unexpected value %q", t.Value)
	}

	return &BooleanNode{
		Value: b,
	}, nil
}

func (n *BooleanNode) optimize() (Node, error) {
	return n, nil
}

func (n BooleanNode) String() string {
	return fmt.Sprintf("%t", n.Value)
}

func (n BooleanNode) Evaluate(ctx *Context) (interface{}, error) {
	return n.Value, nil
}

// A NullNode represents the JSON null value.
type NullNode struct{}

func parseNull(p *parser, t token) (Node, error) {
	return &NullNode{}, nil
}

func (n *NullNode) optimize() (Node, error) {
	return n, nil
}

func (NullNode) String() string {
	return "null"
}

func (n NullNode) Evaluate(ctx *Context) (interface{}, error) {
	return nil, nil
}

// A RegexNode represents a regular expression.
type RegexNode struct {
	Value *regexp.Regexp
}

func parseRegex(p *parser, t token) (Node, error) {

	if t.Value == "" {
		return nil, newError(ErrEmptyRegex, t)
	}

	re, err := regexp.Compile(t.Value)
	if err != nil {
		hint := "unknown error"
		if e, ok := err.(*syntax.Error); ok {
			hint = string(e.Code)
		}

		return nil, newErrorHint(ErrInvalidRegex, t, hint)
	}

	return &RegexNode{
		Value: re,
	}, nil
}

func (n *RegexNode) optimize() (Node, error) {
	return n, nil
}

func (n RegexNode) String() string {
	var expr string
	if n.Value != nil {
		expr = n.Value.String()
	}
	return fmt.Sprintf("/%s/", expr)
}

func (n RegexNode) Evaluate(ctx *Context) (interface{}, error) {
	return n.Value, nil
}

// A VariableNode represents a JSONata variable.
type VariableNode struct {
	Name string
	Next Node
}

func parseVariable(p *parser, t token) (Node, error) {
	return &VariableNode{
		Name: t.Value,
	}, nil
}

func (n *VariableNode) optimize() (Node, error) {
	return n, nil
}

func (n VariableNode) String() string {
	return "$" + n.Name
}

func (n VariableNode) Evaluate(ctx *Context) (interface{}, error) {
	if n.Name == "" {
		return ctx.Input, nil
	}
	return nil, fmt.Errorf("variable %s not found", n.Name)
}

// A NameNode represents a JSON field name.
type NameNode struct {
	Value   string
	escaped bool
}

func parseName(p *parser, t token) (Node, error) {
	return &NameNode{
		Value: t.Value,
	}, nil
}

func parseEscapedName(p *parser, t token) (Node, error) {
	return &NameNode{
		Value:   t.Value,
		escaped: true,
	}, nil
}

func (n *NameNode) optimize() (Node, error) {
	return &PathNode{
		Steps: []Node{n},
	}, nil
}

func (n NameNode) String() string {
	if n.escaped {
		return fmt.Sprintf("`%s`", n.Value)
	}
	return n.Value
}

// Escaped returns true for names enclosed in backticks (e.g.
// `Product Name`), and false otherwise. This doesn't affect
// evaluation but may be useful when recreating a JSONata
// expression from its AST.
func (n NameNode) Escaped() bool {
	return n.escaped
}

func (n NameNode) Evaluate(ctx *Context) (interface{}, error) {
	if ctx.Input == nil {
		return nil, nil
	}
	
	switch v := ctx.Input.(type) {
	case map[string]interface{}:
		return v[n.Value], nil
	case []interface{}:
		var results []interface{}
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if val, exists := m[n.Value]; exists {
					results = append(results, val)
				}
			}
		}
		if len(results) == 1 {
			return results[0], nil
		}
		return results, nil
	default:
		return nil, nil
	}
}

// A PathNode represents a JSON object path. It consists of one
// or more 'steps' or Nodes (most commonly NameNode objects).
type PathNode struct {
	Steps      []Node
	KeepArrays bool
}

func (n *PathNode) optimize() (Node, error) {
	return n, nil
}

func (n PathNode) String() string {
	s := joinNodes(n.Steps, ".")
	if n.KeepArrays {
		s += "[]"
	}
	return s
}

func (n PathNode) Evaluate(ctx *Context) (interface{}, error) {
	if len(n.Steps) == 0 {
		return nil, nil
	}

	var current interface{} = ctx.Input
	for i, step := range n.Steps {
		nextCtx := &Context{
			Parent:   ctx,
			Position: -1,
			Input:    current,
		}

		var err error
		current, err = step.Evaluate(nextCtx)
		if err != nil {
			return nil, err
		}

		if current == nil {
			return nil, nil
		}
	}

	if n.KeepArrays {
		if arr, ok := current.([]interface{}); ok {
			return arr, nil
		}
	}

	return current, nil
}

// A NegationNode represents a numeric negation operation.
type NegationNode struct {
	RHS Node
}

func parseNegation(p *parser, t token) (Node, error) {
	return &NegationNode{
		RHS: p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *NegationNode) optimize() (Node, error) {
	var err error
	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	if number, ok := n.RHS.(*NumberNode); ok {
		return &NumberNode{
			Value: -number.Value,
		}, nil
	}

	return n, nil
}

func (n NegationNode) String() string {
	return fmt.Sprintf("-%s", n.RHS)
}

func (n NegationNode) Evaluate(ctx *Context) (interface{}, error) {
	val, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	num, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("cannot negate non-numeric value")
	}
	
	return -num, nil
}

// A RangeNode represents the range operator.
type RangeNode struct {
	LHS Node
	RHS Node
}

func (n *RangeNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n RangeNode) String() string {
	return fmt.Sprintf("%s..%s", n.LHS, n.RHS)
}

func (n RangeNode) Evaluate(ctx *Context) (interface{}, error) {
	start, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	end, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	startNum, ok := start.(float64)
	if !ok {
		return nil, fmt.Errorf("range start must be numeric")
	}
	
	endNum, ok := end.(float64)
	if !ok {
		return nil, fmt.Errorf("range end must be numeric")
	}
	
	var result []interface{}
	for i := startNum; i <= endNum; i++ {
		result = append(result, i)
	}
	return result, nil
}

// An ArrayNode represents an array of items.
type ArrayNode struct {
	Items []Node
}

func parseArray(p *parser, t token) (Node, error) {

	var items []Node

	for hasItems := p.token.Type != typeBracketClose; hasItems; { // disallow trailing commas

		item := p.parseExpression(0)

		if p.token.Type == typeRange {

			p.consume(typeRange, true)

			item = &RangeNode{
				LHS: item,
				RHS: p.parseExpression(0),
			}
		}

		items = append(items, item)

		if p.token.Type != typeComma {
			break
		}
		p.consume(typeComma, true)
	}

	p.consume(typeBracketClose, false)

	return &ArrayNode{
		Items: items,
	}, nil
}

func (n *ArrayNode) optimize() (Node, error) {

	var err error

	for i := range n.Items {
		n.Items[i], err = n.Items[i].optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n ArrayNode) String() string {
	return fmt.Sprintf("[%s]", joinNodes(n.Items, ", "))
}

func (n ArrayNode) Evaluate(ctx *Context) (interface{}, error) {
	result := make([]interface{}, len(n.Items))
	for i, item := range n.Items {
		val, err := item.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// An ObjectNode represents an object, an unordered list of
// key-value pairs.
type ObjectNode struct {
	Pairs [][2]Node
}

func parseObject(p *parser, t token) (Node, error) {

	var pairs [][2]Node

	for hasItems := p.token.Type != typeBraceClose; hasItems; { // disallow trailing commas

		key := p.parseExpression(0)
		p.consume(typeColon, true)
		value := p.parseExpression(0)

		pairs = append(pairs, [2]Node{key, value})

		if p.token.Type != typeComma {
			break
		}
		p.consume(typeComma, true)
	}

	p.consume(typeBraceClose, false)

	return &ObjectNode{
		Pairs: pairs,
	}, nil
}

func (n *ObjectNode) optimize() (Node, error) {

	var err error

	for i := range n.Pairs {
		for j := 0; j < 2; j++ {
			n.Pairs[i][j], err = n.Pairs[i][j].optimize()
			if err != nil {
				return nil, err
			}
		}
	}

	return n, nil
}

func (n ObjectNode) String() string {
	values := make([]string, len(n.Pairs))

	for i, pair := range n.Pairs {
		values[i] = fmt.Sprintf("%s: %s", pair[0], pair[1])
	}

	return fmt.Sprintf("{%s}", strings.Join(values, ", "))
}

func (n ObjectNode) Evaluate(ctx *Context) (interface{}, error) {
	result := make(map[string]interface{})
	
	for _, pair := range n.Pairs {
		key, err := pair[0].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		
		keyStr, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("object key must evaluate to string")
		}
		
		value, err := pair[1].Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		
		result[keyStr] = value
	}
	
	return result, nil
}

// A BlockNode represents a block expression.
type BlockNode struct {
	Exprs []Node
}

func parseBlock(p *parser, t token) (Node, error) {

	var exprs []Node

	for p.token.Type != typeParenClose { // allow trailing semicolons

		exprs = append(exprs, p.parseExpression(0))

		if p.token.Type != typeSemicolon {
			break
		}
		p.consume(typeSemicolon, true)
	}

	p.consume(typeParenClose, false)

	return &BlockNode{
		Exprs: exprs,
	}, nil
}

func (n *BlockNode) optimize() (Node, error) {

	var err error

	for i := range n.Exprs {
		n.Exprs[i], err = n.Exprs[i].optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n BlockNode) String() string {
	return fmt.Sprintf("(%s)", joinNodes(n.Exprs, "; "))
}

func (n BlockNode) Evaluate(ctx *Context) (interface{}, error) {
	var result interface{}
	for _, expr := range n.Exprs {
		var err error
		result, err = expr.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// A WildcardNode represents the wildcard operator.
type WildcardNode struct{}

func parseWildcard(p *parser, t token) (Node, error) {
	return &WildcardNode{}, nil
}

func (n *WildcardNode) optimize() (Node, error) {
	return n, nil
}

func (WildcardNode) String() string {
	return "*"
}

func (n WildcardNode) Evaluate(ctx *Context) (interface{}, error) {
	if ctx.Input == nil {
		return nil, nil
	}
	
	switch v := ctx.Input.(type) {
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, val := range v {
			result = append(result, val)
		}
		return result, nil
	case []interface{}:
		return v, nil
	default:
		return nil, nil
	}
}

// A DescendentNode represents the descendent operator.
type DescendentNode struct{}

func parseDescendent(p *parser, t token) (Node, error) {
	return &DescendentNode{}, nil
}

func (n *DescendentNode) optimize() (Node, error) {
	return n, nil
}

func (DescendentNode) String() string {
	return "**"
}

func (n DescendentNode) Evaluate(ctx *Context) (interface{}, error) {
	if ctx.Input == nil {
		return nil, nil
	}

	var results []interface{}
	
	switch v := ctx.Input.(type) {
	case map[string]interface{}:
		for _, val := range v {
			results = append(results, val)
			if nested, err := n.Evaluate(&Context{Input: val}); err == nil && nested != nil {
				if arr, ok := nested.([]interface{}); ok {
					results = append(results, arr...)
				} else {
					results = append(results, nested)
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			results = append(results, item)
			if nested, err := n.Evaluate(&Context{Input: item}); err == nil && nested != nil {
				if arr, ok := nested.([]interface{}); ok {
					results = append(results, arr...)
				} else {
					results = append(results, nested)
				}
			}
		}
	}

	return results, nil
}

// An ObjectTransformationNode represents the object transformation
// operator.
type ObjectTransformationNode struct {
	Pattern Node
	Updates Node
	Deletes Node
}

func parseObjectTransformation(p *parser, t token) (Node, error) {

	var deletes Node

	pattern := p.parseExpression(0)
	p.consume(typePipe, true)
	updates := p.parseExpression(0)
	if p.token.Type == typeComma {
		p.consume(typeComma, true)
		deletes = p.parseExpression(0)
	}
	p.consume(typePipe, true)

	return &ObjectTransformationNode{
		Pattern: pattern,
		Updates: updates,
		Deletes: deletes,
	}, nil
}

func (n *ObjectTransformationNode) optimize() (Node, error) {

	var err error

	n.Pattern, err = n.Pattern.optimize()
	if err != nil {
		return nil, err
	}

	n.Updates, err = n.Updates.optimize()
	if err != nil {
		return nil, err
	}

	if n.Deletes != nil {
		n.Deletes, err = n.Deletes.optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n ObjectTransformationNode) String() string {
	s := fmt.Sprintf("|%s|%s", n.Pattern, n.Updates)
	if n.Deletes != nil {
		s += fmt.Sprintf(", %s", n.Deletes)
	}
	s += "|"
	return s
}

func (n ObjectTransformationNode) Evaluate(ctx *Context) (interface{}, error) {
	pattern, err := n.Pattern.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if pattern == nil {
		return nil, nil
	}

	obj, ok := pattern.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("pattern must evaluate to object")
	}

	updates, err := n.Updates.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	updateObj, ok := updates.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("updates must evaluate to object")
	}

	result := make(map[string]interface{})
	for k, v := range obj {
		result[k] = v
	}

	for k, v := range updateObj {
		result[k] = v
	}

	if n.Deletes != nil {
		deletes, err := n.Deletes.Evaluate(ctx)
		if err != nil {
			return nil, err
		}

		deleteArr, ok := deletes.([]interface{})
		if !ok {
			return nil, fmt.Errorf("deletes must evaluate to array")
		}

		for _, key := range deleteArr {
			keyStr, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("delete key must be string")
			}
			delete(result, keyStr)
		}
	}

	return result, nil
}

// A ParamType represents the type of a parameter in a lambda
// function signature.
type ParamType uint

// Supported parameter types.
const (
	ParamTypeNumber ParamType = 1 << iota
	ParamTypeString
	ParamTypeBool
	ParamTypeNull
	ParamTypeArray
	ParamTypeObject
	ParamTypeFunc
	ParamTypeJSON
	ParamTypeAny
)

func parseParamType(r rune) (ParamType, bool) {

	var typ ParamType

	switch r {
	case 'n':
		typ = ParamTypeNumber
	case 's':
		typ = ParamTypeString
	case 'b':
		typ = ParamTypeBool
	case 'l':
		typ = ParamTypeNull
	case 'a':
		typ = ParamTypeArray
	case 'o':
		typ = ParamTypeObject
	case 'f':
		typ = ParamTypeFunc
	case 'j':
		typ = ParamTypeJSON
	case 'x':
		typ = ParamTypeAny
	default:
		return 0, false
	}

	return typ, true
}

func (typ ParamType) String() string {

	var s string

	if typ&ParamTypeNumber != 0 {
		s += "n"
	}
	if typ&ParamTypeString != 0 {
		s += "s"
	}
	if typ&ParamTypeBool != 0 {
		s += "b"
	}
	if typ&ParamTypeNull != 0 {
		s += "l"
	}
	if typ&ParamTypeArray != 0 {
		s += "a"
	}
	if typ&ParamTypeObject != 0 {
		s += "o"
	}
	if typ&ParamTypeFunc != 0 {
		s += "f"
	}
	if typ&ParamTypeJSON != 0 {
		s += "j"
	}
	if typ&ParamTypeAny != 0 {
		s += "x"
	}

	if len(s) > 1 {
		s = "(" + s + ")"
	}

	return s
}

// A ParamOpt represents the options on a parameter in a lambda
// function signature.
type ParamOpt uint8

const (
	_ ParamOpt = iota

	// ParamOptional denotes an optional parameter.
	ParamOptional

	// ParamVariadic denotes a variadic parameter.
	ParamVariadic

	// ParamContextable denotes a parameter that can be
	// replaced by the evaluation context if no value is
	// provided by the caller.
	ParamContextable
)

func parseParamOpt(r rune) (ParamOpt, bool) {

	var opt ParamOpt

	switch r {
	case '?':
		opt = ParamOptional
	case '+':
		opt = ParamVariadic
	case '-':
		opt = ParamContextable
	default:
		return 0, false
	}

	return opt, true
}

func (opt ParamOpt) String() string {
	switch opt {
	case ParamOptional:
		return "?"
	case ParamVariadic:
		return "+"
	case ParamContextable:
		return "-"
	default:
		return ""
	}
}

// A Param represents a parameter in a lambda function signature.
type Param struct {
	Type      ParamType
	Option    ParamOpt
	SubParams []Param
}

func (p Param) String() string {

	s := p.Type.String()

	if p.SubParams != nil {
		s += "<"
		for _, sub := range p.SubParams {
			s += sub.String()
		}
		s += ">"
	}

	s += p.Option.String()
	return s
}

func parseParams(s string) ([]Param, error) {

	params := []Param{}

	for len(s) > 0 {

		r, w := utf8.DecodeRuneInString(s)

		if r == ':' {
			break
		}

		if typ, ok := parseParamType(r); ok {
			params = append(params, Param{
				Type: typ,
			})
			s = s[w:]
			continue
		}

		if r == '(' {
			part := getBracketedString(s, '(', ')')
			var types ParamType
			for _, c := range part {
				typ, ok := parseParamType(c)
				if !ok {
					// TODO: Add position to this error.
					return nil, &Error{
						Type: ErrInvalidUnionType,
						Hint: string(c),
					}
				}
				types |= typ
			}
			params = append(params, Param{
				Type: types,
			})
			s = s[len(part)+2:]
			continue
		}

		if opt, ok := parseParamOpt(r); ok {
			if len(params) == 0 {
				// TODO: Add position to this error.
				return nil, &Error{
					Type: ErrUnmatchedOption,
					Hint: string(r),
				}
			}
			params[len(params)-1].Option = opt
			s = s[w:]
			continue
		}

		if r == '<' {
			if len(params) == 0 {
				// TODO: Add position to this error.
				return nil, &Error{
					Type: ErrUnmatchedSubtype,
				}
			}
			n := len(params) - 1
			if params[n].Type != ParamTypeArray && params[n].Type != ParamTypeFunc {
				// TODO: Add position to this error.
				return nil, &Error{
					Type: ErrInvalidSubtype,
					Hint: params[n].Type.String(),
				}
			}
			part := getBracketedString(s, '<', '>')
			sub, err := parseParams(part)
			if err != nil {
				return nil, err
			}
			params[n].SubParams = sub
			s = s[len(part)+2:]
			continue
		}

		// TODO: Add position to this error.
		return nil, &Error{
			Type: ErrInvalidParamType,
			Hint: string(r),
		}
	}

	return params, nil
}

func getBracketedString(s string, open, close rune) string {

	var depth int

	for pos, c := range s {

		if pos == 0 && c != open {
			break
		}

		if c == open {
			depth++
			continue
		}

		if c == close {
			depth--
			if depth == 0 {
				return s[utf8.RuneLen(open):pos]
			}
		}
	}

	return ""
}

// A LambdaNode represents a user-defined JSONata function.
type LambdaNode struct {
	Body       Node
	ParamNames []string
	shorthand  bool
}

func (n *LambdaNode) optimize() (Node, error) {
	var err error
	n.Body, err = n.Body.optimize()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n LambdaNode) String() string {
	name := "function"
	if n.shorthand {
		name = "λ"
	}

	params := make([]string, len(n.ParamNames))
	for i, s := range n.ParamNames {
		params[i] = "$" + s
	}

	return fmt.Sprintf("%s(%s){%s}", name, strings.Join(params, ", "), n.Body)
}

func (n LambdaNode) Evaluate(ctx *Context) (interface{}, error) {
	return map[string]interface{}{
		"__lambda": true,
		"params":   n.ParamNames,
		"body":     n.Body,
		"context":  ctx,
	}, nil
}

// Shorthand returns true if the lambda function was defined
// with the shorthand symbol "λ", and false otherwise. This
// doesn't affect evaluation but may be useful when recreating
// a JSONata expression from its AST.
func (n LambdaNode) Shorthand() bool {
	return n.shorthand
}

// A TypedLambdaNode represents a user-defined JSONata function
// with a type signature.
type TypedLambdaNode struct {
	*LambdaNode
	In  []Param
	Out []Param
}

func (n *TypedLambdaNode) optimize() (Node, error) {
	node, err := n.LambdaNode.optimize()
	if err != nil {
		return nil, err
	}
	n.LambdaNode = node.(*LambdaNode)
	return n, nil
}

func (n TypedLambdaNode) String() string {
	name := "function"
	if n.shorthand {
		name = "λ"
	}

	params := make([]string, len(n.ParamNames))
	for i, s := range n.ParamNames {
		params[i] = "$" + s
	}

	inputs := make([]string, len(n.In))
	for i, p := range n.In {
		inputs[i] = p.String()
	}

	return fmt.Sprintf("%s(%s)<%s>{%s}", name, strings.Join(params, ", "), strings.Join(inputs, ""), n.Body)
}

func (n TypedLambdaNode) Evaluate(ctx *Context) (interface{}, error) {
	return map[string]interface{}{
		"__lambda": true,
		"params":   n.ParamNames,
		"body":     n.Body,
		"context":  ctx,
		"in":       n.In,
		"out":      n.Out,
	}, nil
}

// A PartialNode represents a partially applied function.
type PartialNode struct {
	Func Node
	Args []Node
}

func (n PartialNode) Evaluate(ctx *Context) (interface{}, error) {
	fn, err := n.Func.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	args := make([]interface{}, len(n.Args))
	for i, arg := range n.Args {
		if _, ok := arg.(*PlaceholderNode); !ok {
			val, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			args[i] = val
		}
	}

	return map[string]interface{}{
		"function": fn,
		"args":     args,
	}, nil
}

func (n *PartialNode) optimize() (Node, error) {

	var err error

	n.Func, err = n.Func.optimize()
	if err != nil {
		return nil, err
	}

	for i := range n.Args {
		n.Args[i], err = n.Args[i].optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n PartialNode) String() string {
	return fmt.Sprintf("%s(%s)", n.Func, joinNodes(n.Args, ", "))
}

// A PlaceholderNode represents a placeholder argument
// in a partially applied function.
type PlaceholderNode struct{}

func (n PlaceholderNode) Evaluate(ctx *Context) (interface{}, error) {
	return nil, fmt.Errorf("placeholder cannot be evaluated")
}

func (n PlaceholderNode) optimize() (Node, error) {
	return n, nil
}

func (PlaceholderNode) String() string {
	return "?"
}

// A FunctionCallNode represents a call to a function.
type FunctionCallNode struct {
	Func Node
	Args []Node
}

func (n FunctionCallNode) Evaluate(ctx *Context) (interface{}, error) {
	fn, err := n.Func.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	args := make([]interface{}, len(n.Args))
	for i, arg := range n.Args {
		val, err := arg.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return map[string]interface{}{
		"function": fn,
		"args":     args,
	}, nil
}

const typePlaceholder = typeCondition

func parseFunctionCall(p *parser, t token, lhs Node) (Node, error) {

	if isLambda, shorthand := isLambdaName(lhs); isLambda {
		return parseLambdaDefinition(p, shorthand)
	}

	var args []Node
	var isPartial bool

	for hasArgs := p.token.Type != typeParenClose; hasArgs; { // disallow trailing commas

		var arg Node

		if p.token.Type == typePlaceholder {
			isPartial = true
			arg = &PlaceholderNode{}
			p.consume(typePlaceholder, true)
		} else {
			arg = p.parseExpression(0)
		}

		args = append(args, arg)

		if p.token.Type != typeComma {
			break
		}
		p.consume(typeComma, true)
	}

	p.consume(typeParenClose, false)

	if isPartial {
		return &PartialNode{
			Func: lhs,
			Args: args,
		}, nil
	}

	return &FunctionCallNode{
		Func: lhs,
		Args: args,
	}, nil
}

func (n *FunctionCallNode) optimize() (Node, error) {

	var err error

	n.Func, err = n.Func.optimize()
	if err != nil {
		return nil, err
	}

	for i := range n.Args {
		n.Args[i], err = n.Args[i].optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n FunctionCallNode) String() string {
	return fmt.Sprintf("%s(%s)", n.Func, joinNodes(n.Args, ", "))
}

func isLambdaName(n Node) (bool, bool) {
	switch n := n.(type) {
	case *NameNode:
		return n.Value == "function" || n.Value == "λ", n.Value == "λ"
	default:
		return false, false
	}
}

func parseLambdaDefinition(p *parser, shorthand bool) (Node, error) {

	var params []Param

	paramNames, err := extractParamNames(p)
	if err != nil {
		return nil, err
	}

	sig, isTyped := extractSignature(p)
	if isTyped {
		params, err = parseParams(sig)
		if err != nil {
			return nil, err
		}
		if len(params) != len(paramNames) {
			return nil, newError(ErrParamCount, p.token)
		}
	}

	p.consume(typeBraceOpen, true)
	body := p.parseExpression(0)
	p.consume(typeBraceClose, true)

	lambda := &LambdaNode{
		Body:       body,
		ParamNames: paramNames,
		shorthand:  shorthand,
	}

	if !isTyped {
		return lambda, nil
	}

	return &TypedLambdaNode{
		LambdaNode: lambda,
		In:         params,
	}, nil
}

func extractParamNames(p *parser) ([]string, error) {

	var names []string
	usedNames := map[string]bool{}

	currToken := p.token
	for hasArgs := p.token.Type != typeParenClose; hasArgs; { // disallow trailing commas

		arg := p.parseExpression(0)

		v, ok := arg.(*VariableNode)
		if !ok {
			return nil, newError(ErrIllegalParam, currToken)
		}

		if usedNames[v.Name] {
			return nil, newError(ErrDuplicateParam, currToken)
		}

		usedNames[v.Name] = true
		names = append(names, v.Name)

		if p.token.Type != typeComma {
			break
		}
		p.consume(typeComma, true)

		currToken = p.token
	}

	p.consume(typeParenClose, false)

	return names, nil
}

func extractSignature(p *parser) (string, bool) {

	const (
		typeSigStart = typeLess
		typeSigEnd   = typeGreater
	)

	if p.token.Type != typeSigStart {
		return "", false
	}

	sig := ""
	depth := 1

Loop:
	for p.token.Type != typeBraceOpen && p.token.Type != typeEOF {

		p.advance(true)

		switch p.token.Type {
		case typeSigEnd:
			depth--
			if depth == 0 {
				break Loop
			}
		case typeSigStart:
			depth++
		}

		sig += p.token.Value
	}

	p.consume(typeSigEnd, true)
	return sig, true
}

// A PredicateNode represents a predicate expression.
type PredicateNode struct {
	Expr    Node
	Filters []Node
}

func (n *PredicateNode) optimize() (Node, error) {
	return n, nil
}

func (n PredicateNode) String() string {
	return fmt.Sprintf("%s[%s]", n.Expr, joinNodes(n.Filters, ", "))
}

func (n PredicateNode) Evaluate(ctx *Context) (interface{}, error) {
	expr, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if expr == nil {
		return nil, nil
	}

	var results []interface{}
	var items []interface{}

	switch v := expr.(type) {
	case []interface{}:
		items = v
	default:
		items = []interface{}{v}
	}

	for i, item := range items {
		itemCtx := &Context{
			Input:    item,
			Parent:   ctx,
			Position: i,
		}

		match := true
		for _, filter := range n.Filters {
			result, err := filter.Evaluate(itemCtx)
			if err != nil {
				return nil, err
			}

			switch v := result.(type) {
			case bool:
				if !v {
					match = false
					break
				}
			case float64:
				if int(v) != i {
					match = false
					break
				}
			default:
				match = false
			}
		}

		if match {
			results = append(results, item)
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// A GroupNode represents a group expression.
type GroupNode struct {
	Expr Node
	*ObjectNode
}

func parseGroup(p *parser, t token, lhs Node) (Node, error) {

	obj, err := parseObject(p, t)
	if err != nil {
		return nil, err
	}

	return &GroupNode{
		Expr:       lhs,
		ObjectNode: obj.(*ObjectNode),
	}, nil
}

func (n *GroupNode) optimize() (Node, error) {

	var err error

	n.Expr, err = n.Expr.optimize()
	if err != nil {
		return nil, err
	}

	if _, isGroup := n.Expr.(*GroupNode); isGroup {
		// TODO: Add position info.
		return nil, &Error{
			Type: ErrGroupGroup,
		}
	}

	obj, err := n.ObjectNode.optimize()
	if err != nil {
		return nil, err
	}
	n.ObjectNode = obj.(*ObjectNode)

	return n, nil
}

func (n GroupNode) String() string {
	return fmt.Sprintf("%s%s", n.Expr, n.ObjectNode)
}

// A ConditionalNode represents an if-then-else expression.
type ConditionalNode struct {
	If   Node
	Then Node
	Else Node
}

func parseConditional(p *parser, t token, lhs Node) (Node, error) {

	var els Node
	rhs := p.parseExpression(0)

	if p.token.Type == typeColon {
		p.consume(typeColon, true)
		els = p.parseExpression(0)
	}

	return &ConditionalNode{
		If:   lhs,
		Then: rhs,
		Else: els,
	}, nil
}

func (n *ConditionalNode) optimize() (Node, error) {

	var err error

	n.If, err = n.If.optimize()
	if err != nil {
		return nil, err
	}

	n.Then, err = n.Then.optimize()
	if err != nil {
		return nil, err
	}

	if n.Else != nil {
		n.Else, err = n.Else.optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n ConditionalNode) String() string {
	s := fmt.Sprintf("%s ? %s", n.If, n.Then)
	if n.Else != nil {
		s += fmt.Sprintf(" : %s", n.Else)
	}
	return s
}

func (n ConditionalNode) Evaluate(ctx *Context) (interface{}, error) {
	cond, err := n.If.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	cbool, ok := cond.(bool)
	if !ok {
		return nil, fmt.Errorf("condition must evaluate to boolean")
	}
	
	if cbool {
		return n.Then.Evaluate(ctx)
	}
	if n.Else != nil {
		return n.Else.Evaluate(ctx)
	}
	return nil, nil
}

// An AssignmentNode represents a variable assignment.
type AssignmentNode struct {
	Name  string
	Value Node
}

func parseAssignment(p *parser, t token, lhs Node) (Node, error) {

	v, ok := lhs.(*VariableNode)
	if !ok {
		return nil, newErrorHint(ErrIllegalAssignment, t, lhs.String())
	}

	return &AssignmentNode{
		Name:  v.Name,
		Value: p.parseExpression(p.bp(t.Type) - 1), // right-associative
	}, nil
}

func (n *AssignmentNode) optimize() (Node, error) {

	var err error

	n.Value, err = n.Value.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n AssignmentNode) String() string {
	return fmt.Sprintf("$%s := %s", n.Name, n.Value)
}

func (n AssignmentNode) Evaluate(ctx *Context) (interface{}, error) {
	value, err := n.Value.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	// Store in context's variable map (to be implemented)
	return value, nil
}

// A NumericOperator is a mathematical operation between two
// numeric values.
type NumericOperator uint8

// Numeric operations supported by JSONata.
const (
	_ NumericOperator = iota
	NumericAdd
	NumericSubtract
	NumericMultiply
	NumericDivide
	NumericModulo
)

func (op NumericOperator) String() string {
	switch op {
	case NumericAdd:
		return "+"
	case NumericSubtract:
		return "-"
	case NumericMultiply:
		return "*"
	case NumericDivide:
		return "/"
	case NumericModulo:
		return "%"
	default:
		return ""
	}
}

// A NumericOperatorNode represents a numeric operation.
type NumericOperatorNode struct {
	Type NumericOperator
	LHS  Node
	RHS  Node
}

func parseNumericOperator(p *parser, t token, lhs Node) (Node, error) {

	var op NumericOperator

	switch t.Type {
	case typePlus:
		op = NumericAdd
	case typeMinus:
		op = NumericSubtract
	case typeMult:
		op = NumericMultiply
	case typeDiv:
		op = NumericDivide
	case typeMod:
		op = NumericModulo
	default: // should be unreachable
		panicf("parseNumericOperator: unexpected operator %q", t.Value)
	}

	return &NumericOperatorNode{
		Type: op,
		LHS:  lhs,
		RHS:  p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *NumericOperatorNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n NumericOperatorNode) String() string {
	return fmt.Sprintf("%s %s %s", n.LHS, n.Type, n.RHS)
}

func (n NumericOperatorNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	rhs, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	lnum, ok := lhs.(float64)
	if !ok {
		return nil, fmt.Errorf("left operand must be numeric")
	}
	
	rnum, ok := rhs.(float64)
	if !ok {
		return nil, fmt.Errorf("right operand must be numeric")
	}
	
	switch n.Type {
	case NumericAdd:
		return lnum + rnum, nil
	case NumericSubtract:
		return lnum - rnum, nil
	case NumericMultiply:
		return lnum * rnum, nil
	case NumericDivide:
		if rnum == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return lnum / rnum, nil
	case NumericModulo:
		if rnum == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return float64(int64(lnum) % int64(rnum)), nil
	default:
		return nil, fmt.Errorf("unknown numeric operator")
	}
}

// A ComparisonOperator is an operation that compares two values.
type ComparisonOperator uint8

// Comparison operations supported by JSONata.
const (
	_ ComparisonOperator = iota
	ComparisonEqual
	ComparisonNotEqual
	ComparisonLess
	ComparisonLessEqual
	ComparisonGreater
	ComparisonGreaterEqual
	ComparisonIn
)

func (op ComparisonOperator) String() string {
	switch op {
	case ComparisonEqual:
		return "="
	case ComparisonNotEqual:
		return "!="
	case ComparisonLess:
		return "<"
	case ComparisonLessEqual:
		return "<="
	case ComparisonGreater:
		return ">"
	case ComparisonGreaterEqual:
		return ">="
	case ComparisonIn:
		return "in"
	default:
		return ""
	}
}

// A ComparisonOperatorNode represents a comparison operation.
type ComparisonOperatorNode struct {
	Type ComparisonOperator
	LHS  Node
	RHS  Node
}

func parseComparisonOperator(p *parser, t token, lhs Node) (Node, error) {

	var op ComparisonOperator

	switch t.Type {
	case typeEqual:
		op = ComparisonEqual
	case typeNotEqual:
		op = ComparisonNotEqual
	case typeLess:
		op = ComparisonLess
	case typeLessEqual:
		op = ComparisonLessEqual
	case typeGreater:
		op = ComparisonGreater
	case typeGreaterEqual:
		op = ComparisonGreaterEqual
	case typeIn:
		op = ComparisonIn
	default: // should be unreachable
		panicf("parseComparisonOperator: unexpected operator %q", t.Value)
	}

	return &ComparisonOperatorNode{
		Type: op,
		LHS:  lhs,
		RHS:  p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *ComparisonOperatorNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n ComparisonOperatorNode) String() string {
	return fmt.Sprintf("%s %s %s", n.LHS, n.Type, n.RHS)
}

func (n ComparisonOperatorNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	rhs, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	switch n.Type {
	case ComparisonEqual:
		return deepEqual(lhs, rhs), nil
	case ComparisonNotEqual:
		return !deepEqual(lhs, rhs), nil
	case ComparisonLess:
		return compareValues(lhs, rhs) < 0, nil
	case ComparisonLessEqual:
		return compareValues(lhs, rhs) <= 0, nil
	case ComparisonGreater:
		return compareValues(lhs, rhs) > 0, nil
	case ComparisonGreaterEqual:
		return compareValues(lhs, rhs) >= 0, nil
	case ComparisonIn:
		return isValueIn(lhs, rhs), nil
	default:
		return nil, fmt.Errorf("unknown comparison operator")
	}
}

// A BinaryNode represents a binary operation between two nodes.
type BinaryNode struct {
    Op    tokenType
    Left  Node
    Right Node
}

func (n *BinaryNode) optimize() (Node, error) {
    var err error
    n.Left, err = n.Left.optimize()
    if err != nil {
        return nil, err
    }
    n.Right, err = n.Right.optimize()
    if err != nil {
        return nil, err
    }
    return n, nil
}

func (n BinaryNode) String() string {
    return fmt.Sprintf("%s %s %s", n.Left, n.Op, n.Right)
}

func (n BinaryNode) Evaluate(ctx *Context) (interface{}, error) {
    lhs, err := n.Left.Evaluate(ctx)
    if err != nil {
        return nil, err
    }
    
    rhs, err := n.Right.Evaluate(ctx)
    if err != nil {
        return nil, err
    }
    
    switch n.Op {
    case typeMod:
        lnum, ok := lhs.(float64)
        if !ok {
            return nil, fmt.Errorf("left operand must be numeric")
        }
        rnum, ok := rhs.(float64)
        if !ok {
            return nil, fmt.Errorf("right operand must be numeric")
        }
        if rnum == 0 {
            return nil, fmt.Errorf("modulo by zero")
        }
        return float64(int64(lnum) % int64(rnum)), nil
    default:
        return nil, fmt.Errorf("unsupported binary operator: %v", n.Op)
    }
}

// A BooleanOperator is a logical AND or OR operation between
// two values.
type BooleanOperator uint8

// Boolean operations supported by JSONata.
const (
	_ BooleanOperator = iota
	BooleanAnd
	BooleanOr
)

func (op BooleanOperator) String() string {
	switch op {
	case BooleanAnd:
		return "and"
	case BooleanOr:
		return "or"
	default:
		return ""
	}
}

// A BooleanOperatorNode represents a boolean operation.
type BooleanOperatorNode struct {
	Type BooleanOperator
	LHS  Node
	RHS  Node
}

func parseBooleanOperator(p *parser, t token, lhs Node) (Node, error) {

	var op BooleanOperator

	switch t.Type {
	case typeAnd:
		op = BooleanAnd
	case typeOr:
		op = BooleanOr
	default: // should be unreachable
		panicf("parseBooleanOperator: unexpected operator %q", t.Value)
	}

	return &BooleanOperatorNode{
		Type: op,
		LHS:  lhs,
		RHS:  p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *BooleanOperatorNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n BooleanOperatorNode) String() string {
	return fmt.Sprintf("%s %s %s", n.LHS, n.Type, n.RHS)
}

func (n BooleanOperatorNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	// Short-circuit evaluation for AND/OR
	lbool, ok := lhs.(bool)
	if !ok {
		return nil, fmt.Errorf("left operand must be boolean")
	}
	
	if n.Type == BooleanAnd && !lbool {
		return false, nil
	}
	if n.Type == BooleanOr && lbool {
		return true, nil
	}
	
	rhs, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	rbool, ok := rhs.(bool)
	if !ok {
		return nil, fmt.Errorf("right operand must be boolean")
	}
	
	switch n.Type {
	case BooleanAnd:
		return lbool && rbool, nil
	case BooleanOr:
		return lbool || rbool, nil
	default:
		return nil, fmt.Errorf("unknown boolean operator")
	}
}

// A StringConcatenationNode represents a string concatenation
// operation.
type StringConcatenationNode struct {
	LHS Node
	RHS Node
}

func (n StringConcatenationNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	rhs, err := n.RHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	lstr, ok := lhs.(string)
	if !ok {
		return nil, fmt.Errorf("left operand must be string")
	}

	rstr, ok := rhs.(string)
	if !ok {
		return nil, fmt.Errorf("right operand must be string")
	}

	return lstr + rstr, nil
}

func parseStringConcatenation(p *parser, t token, lhs Node) (Node, error) {
	return &StringConcatenationNode{
		LHS: lhs,
		RHS: p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *StringConcatenationNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n StringConcatenationNode) String() string {
	return fmt.Sprintf("%s & %s", n.LHS, n.RHS)
}

// SortDir describes the sort order of a sort operation.
type SortDir uint8

// Sort orders supported by JSONata.
const (
	_ SortDir = iota
	SortDefault
	SortAscending
	SortDescending
)

// A SortTerm defines a JSONata sort term.
type SortTerm struct {
	Dir  SortDir
	Expr Node
}

// A SortNode represents a sort clause on a JSONata path step.
type SortNode struct {
	Expr  Node
	Terms []SortTerm
}

func parseSort(p *parser, t token, lhs Node) (Node, error) {

	var terms []SortTerm

	p.consume(typeParenOpen, true)

	for {
		dir := SortDefault

		switch typ := p.token.Type; typ {
		case typeLess:
			dir = SortAscending
			p.consume(typ, true)
		case typeGreater:
			dir = SortDescending
			p.consume(typ, true)
		}

		terms = append(terms, SortTerm{
			Dir:  dir,
			Expr: p.parseExpression(0),
		})

		if p.token.Type != typeComma {
			break
		}
		p.consume(typeComma, true)
	}

	p.consume(typeParenClose, true)

	return &SortNode{
		Expr:  lhs,
		Terms: terms,
	}, nil
}

func (n *SortNode) optimize() (Node, error) {

	var err error

	n.Expr, err = n.Expr.optimize()
	if err != nil {
		return nil, err
	}

	for i := range n.Terms {
		n.Terms[i].Expr, err = n.Terms[i].Expr.optimize()
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (n SortNode) String() string {
	terms := make([]string, len(n.Terms))

	for i, t := range n.Terms {
		var sym string
		switch t.Dir {
		case SortAscending:
			sym = "<"
		case SortDescending:
			sym = ">"
		}
		terms[i] = sym + t.Expr.String()
	}

	return fmt.Sprintf("%s^(%s)", n.Expr, strings.Join(terms, ", "))
}

func (n SortNode) Evaluate(ctx *Context) (interface{}, error) {
	expr, err := n.Expr.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if expr == nil {
		return nil, nil
	}

	items, ok := expr.([]interface{})
	if !ok {
		return expr, nil
	}

	sorted := make([]interface{}, len(items))
	copy(sorted, items)

	sort.SliceStable(sorted, func(i, j int) bool {
		for _, term := range n.Terms {
			iCtx := &Context{Input: sorted[i], Parent: ctx}
			jCtx := &Context{Input: sorted[j], Parent: ctx}

			iVal, err := term.Expr.Evaluate(iCtx)
			if err != nil {
				return false
			}

			jVal, err := term.Expr.Evaluate(jCtx)
			if err != nil {
				return false
			}

			cmp := compareValues(iVal, jVal)
			if cmp == 0 {
				continue
			}

			if term.Dir == SortDescending {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})

	return sorted, nil
}

// A FunctionApplicationNode represents a function application
// operation.
type FunctionApplicationNode struct {
	LHS Node
	RHS Node
}

func parseFunctionApplication(p *parser, t token, lhs Node) (Node, error) {
	return &FunctionApplicationNode{
		LHS: lhs,
		RHS: p.parseExpression(p.bp(t.Type)),
	}, nil
}

func (n *FunctionApplicationNode) optimize() (Node, error) {

	var err error

	n.LHS, err = n.LHS.optimize()
	if err != nil {
		return nil, err
	}

	n.RHS, err = n.RHS.optimize()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (n FunctionApplicationNode) String() string {
	return fmt.Sprintf("%s ~> %s", n.LHS, n.RHS)
}

func (n FunctionApplicationNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.LHS.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	rhsCtx := &Context{
		Input:    lhs,
		Parent:   ctx,
		Position: -1,
	}

	return n.RHS.Evaluate(rhsCtx)
}

// A dotNode is an interim structure used to process JSONata path
// expressions. It is deliberately unexported and creates a PathNode
// during its optimize phase.



// A singletonArrayNode is an interim data structure used when
// processing path expressions. It is deliberately unexported
// and gets converted into a PathNode during optimization.
type singletonArrayNode struct {
	lhs Node
}

func (n *singletonArrayNode) optimize() (Node, error) {
	lhs, err := n.lhs.optimize()
	if err != nil {
		return nil, err
	}

	switch lhs := lhs.(type) {
	case *PathNode:
		lhs.KeepArrays = true
		return lhs, nil
	default:
		return &PathNode{
			Steps:      []Node{lhs},
			KeepArrays: true,
		}, nil
	}
}

func (n singletonArrayNode) String() string {
	return fmt.Sprintf("%s[]", n.lhs)
}

func (n singletonArrayNode) Evaluate(ctx *Context) (interface{}, error) {
	result, err := n.lhs.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	
	if result == nil {
		return nil, nil
	}
	
	// If result is already an array, return it as-is
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	
	// Otherwise wrap the single value in an array
	return []interface{}{result}, nil
}

// A predicateNode is an interim data structure used when processing
// predicate expressions. It is deliberately unexported and gets
// converted into a PredicateNode during optimization.
type predicateNode struct {
	lhs Node // the context for this predicate
	rhs Node // the predicate expression
}

func parsePredicate(p *parser, t token, lhs Node) (Node, error) {
	if p.token.Type == typeBracketClose {
		p.consume(typeBracketClose, false)
		return &singletonArrayNode{
			lhs: lhs,
		}, nil
	}

	rhs := p.parseExpression(0)
	p.consume(typeBracketClose, false)

	return &predicateNode{
		lhs: lhs,
		rhs: rhs,
	}, nil
}

func (n *predicateNode) optimize() (Node, error) {
	lhs, err := n.lhs.optimize()
	if err != nil {
		return nil, err
	}

	rhs, err := n.rhs.optimize()
	if err != nil {
		return nil, err
	}

	switch lhs := lhs.(type) {
	case *GroupNode:
		return nil, &Error{
			Type: ErrGroupPredicate,
		}
	case *PathNode:
		i := len(lhs.Steps) - 1
		switch last := lhs.Steps[i].(type) {
		case *PredicateNode:
			last.Filters = append(last.Filters, rhs)
		default:
			step := &PredicateNode{
				Expr:    last,
				Filters: []Node{rhs},
			}
			lhs.Steps = append(lhs.Steps[:i], step)
		}
		return lhs, nil
	default:
		return &PredicateNode{
			Expr:    lhs,
			Filters: []Node{rhs},
		}, nil
	}
}

func (n *predicateNode) String() string {
	return fmt.Sprintf("%s[%s]", n.lhs, n.rhs)
}

func (n predicateNode) Evaluate(ctx *Context) (interface{}, error) {
	lhs, err := n.lhs.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	if lhs == nil {
		return nil, nil
	}

	var items []interface{}
	switch v := lhs.(type) {
	case []interface{}:
		items = v
	default:
		items = []interface{}{v}
	}

	var results []interface{}
	for i, item := range items {
		itemCtx := &Context{
			Input:    item,
			Parent:   ctx,
			Position: i,
		}

		match, err := n.rhs.Evaluate(itemCtx)
		if err != nil {
			return nil, err
		}

		switch v := match.(type) {
		case bool:
			if v {
				results = append(results, item)
			}
		case float64:
			if int(v) == i {
				results = append(results, item)
			}
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// Helpers

func joinNodes(nodes []Node, sep string) string {

	values := make([]string, len(nodes))

	for i, n := range nodes {
		values[i] = n.String()
	}

	return strings.Join(values, sep)
}

var jsonEscapes = map[rune]string{
	'"':  "\"",
	'\\': "\\",
	'/':  "/",
	'b':  "\b",
	'f':  "\f",
	'n':  "\n",
	'r':  "\r",
	't':  "\t",
}

// unescape replaces JSON escape sequences in a string with their
// unescaped equivalents. Valid escape sequences are:
//
// \X, where X is a character from jsonEscapes
// \uXXXX, where XXXX is a 4-digit hexadecimal Unicode code point.
//
// unescape returns the unescaped string and true if successful,
// otherwise it returns the invalid escape sequence and false.
func unescape(src string) (string, bool) {

	pos := strings.IndexRune(src, '\\')
	if pos < 0 {
		return src, true
	}

	prefix := src[:pos]
	pos++

	esc, w := utf8.DecodeRuneInString(src[pos:])
	pos += w

	repl := jsonEscapes[esc]

	switch {
	case repl != "":
	case esc == 'u':
		hex, w := decodeRunes(src[pos:], 4)
		pos += w

		r := parseRune(hex)

		switch {
		case utf8.ValidRune(r):
		case utf16.IsSurrogate(r):
			hex2, w := decodeRunes(src[pos:], 6)
			pos += w

			if strings.HasPrefix(hex2, "\\u") {
				r = utf16.DecodeRune(r, parseRune(hex2[2:]))
				if r != utf8.RuneError {
					break
				}
			}
			fallthrough
		default:
			return "u" + hex, false
		}
		repl = string(r)
	default:
		return string(esc), false
	}

	rest, ok := unescape(src[pos:])
	if !ok {
		return rest, ok
	}

	return prefix + repl + rest, true
}

// decodeRunes reads n runes from the string s and returns them
// as a string along with the number of bytes read. The returned
// string will always be n runes long, padded with the unicode
// replacement character if the source string contains fewer
// than n runes.
func decodeRunes(s string, n int) (string, int) {

	pos := 0
	runes := make([]rune, n)

	for i := range runes {
		r, w := utf8.DecodeRuneInString(s[pos:])
		runes[i] = r
		pos += w
	}

	return string(runes), pos
}

// parseRune converts a string of hexadecimal digits into the
// equivalent rune. It returns an invalid rune if the input is
// not valid hex.
func parseRune(hex string) rune {
	n, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		return -1
	}
	return rune(n)
}

func deepEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch v1 := a.(type) {
	case float64:
		if v2, ok := b.(float64); ok {
			return v1 == v2
		}
	case string:
		if v2, ok := b.(string); ok {
			return v1 == v2
		}
	case bool:
		if v2, ok := b.(bool); ok {
			return v1 == v2
		}
	case []interface{}:
		if v2, ok := b.([]interface{}); ok {
			if len(v1) != len(v2) {
				return false
			}
			for i := range v1 {
				if !deepEqual(v1[i], v2[i]) {
					return false
				}
			}
			return true
		}
	case map[string]interface{}:
		if v2, ok := b.(map[string]interface{}); ok {
			if len(v1) != len(v2) {
				return false
			}
			for k, val1 := range v1 {
				val2, exists := v2[k]
				if !exists || !deepEqual(val1, val2) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func compareValues(a, b interface{}) int {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return 0
		}
		if a == nil {
			return -1
		}
		return 1
	}

	switch v1 := a.(type) {
	case float64:
		if v2, ok := b.(float64); ok {
			if v1 < v2 {
				return -1
			}
			if v1 > v2 {
				return 1
			}
			return 0
		}
	case string:
		if v2, ok := b.(string); ok {
			return strings.Compare(v1, v2)
		}
	}
	return 0
}

func isValueIn(needle, haystack interface{}) bool {
	switch h := haystack.(type) {
	case []interface{}:
		for _, item := range h {
			if deepEqual(needle, item) {
				return true
			}
		}
	case map[string]interface{}:
		key, ok := needle.(string)
		if !ok {
			return false
		}
		_, exists := h[key]
		return exists
	}
	return false
}
