package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v2"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() openai.FunctionParameters
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

// Registry manages the available tools
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) Definitions() []openai.ChatCompletionToolUnionParam {
	var defs []openai.ChatCompletionToolUnionParam
	for _, t := range r.tools {
		defs = append(defs, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        t.Name(),
			Description: openai.String(t.Description()),
			Parameters:  t.Parameters(),
		}))
	}
	return defs
}

func (r *Registry) Execute(ctx context.Context, name string, args string) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return t.Execute(ctx, json.RawMessage(args))
}
