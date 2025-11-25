package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/openai/openai-go/v2"
)

type DateTool struct{}

func (t *DateTool) Name() string {
	return "get_today_date"
}

func (t *DateTool) Description() string {
	return "Get today's date and time in RFC3339 format"
}

func (t *DateTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *DateTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
