package time

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v2"
)

type TimeInZoneTool struct{}

func (t *TimeInZoneTool) Name() string {
	return "get_time_in_zone"
}

func (t *TimeInZoneTool) Description() string {
	return "Get current time in a specific IANA time zone (e.g., 'America/New_York', 'Europe/London', 'Asia/Tokyo')"
}

func (t *TimeInZoneTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"timezone": map[string]string{
				"type":        "string",
				"description": "IANA time zone name (e.g. America/New_York)",
			},
		},
		"required": []string{"timezone"},
	}
}

func (t *TimeInZoneTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var payload struct {
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return "failed to parse timezone parameter", nil
	}

	loc, err := time.LoadLocation(payload.Timezone)
	if err != nil {
		return fmt.Sprintf("invalid timezone '%s': %v", payload.Timezone, err), nil
	}

	return time.Now().In(loc).Format(time.RFC3339), nil
}
