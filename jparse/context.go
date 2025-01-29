package jparse

type Context struct {
	Input     interface{}
	Parent    *Context
	Position  int
	Variables map[string]interface{}
}

func NewContext(input interface{}, parent *Context) *Context {
	ctx := &Context{
		Input:     input,
		Parent:    parent,
		Position:  -1,
		Variables: make(map[string]interface{}),
	}
	if parent != nil {
		if parent.Position >= 0 {
			ctx.Position = parent.Position
		}
		if parent.Variables != nil {
			for k, v := range parent.Variables {
				ctx.Variables[k] = v
			}
		}
	}
	return ctx
}

func (ctx *Context) WithVariable(name string, value interface{}) *Context {
	newCtx := *ctx
	if newCtx.Variables == nil {
		newCtx.Variables = make(map[string]interface{})
	}
	newCtx.Variables[name] = value
	return &newCtx
}

func (ctx *Context) WithPosition(pos int) *Context {
	newCtx := *ctx
	newCtx.Position = pos
	return &newCtx
}

func (ctx *Context) WithInput(input interface{}) *Context {
	newCtx := *ctx
	newCtx.Input = input
	return &newCtx
}
