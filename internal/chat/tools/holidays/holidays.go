package holidays

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/openai/openai-go/v2"
)

type HolidaysTool struct{}

func (t *HolidaysTool) Name() string {
	return "get_holidays"
}

func (t *HolidaysTool) Description() string {
	return "Gets local bank and public holidays. Each line is a single holiday in the format 'YYYY-MM-DD: Holiday Name'."
}

func (t *HolidaysTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"before_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays before this date. If not provided, all holidays will be returned.",
			},
			"after_date": map[string]string{
				"type":        "string",
				"description": "Optional date in RFC3339 format to get holidays after this date. If not provided, all holidays will be returned.",
			},
			"max_count": map[string]string{
				"type":        "integer",
				"description": "Optional maximum number of holidays to return. If not provided, all holidays will be returned.",
			},
		},
	}
}

func (t *HolidaysTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	link := "https://www.officeholidays.com/ics/spain/catalonia"
	if v := os.Getenv("HOLIDAY_CALENDAR_LINK"); v != "" {
		link = v
	}

	events, err := LoadCalendar(ctx, link)
	if err != nil {
		return "failed to load holiday events", nil
	}

	var payload struct {
		BeforeDate time.Time `json:"before_date,omitempty"`
		AfterDate  time.Time `json:"after_date,omitempty"`
		MaxCount   int       `json:"max_count,omitempty"`
	}

	if err := json.Unmarshal(args, &payload); err != nil {
		return "failed to parse tool call arguments: " + err.Error(), nil
	}

	var holidays []string
	for _, event := range events {
		date, err := event.GetAllDayStartAt()
		if err != nil {
			continue
		}

		if payload.MaxCount > 0 && len(holidays) >= payload.MaxCount {
			break
		}

		if !payload.BeforeDate.IsZero() && date.After(payload.BeforeDate) {
			continue
		}

		if !payload.AfterDate.IsZero() && date.Before(payload.AfterDate) {
			continue
		}

		holidays = append(holidays, date.Format(time.DateOnly)+": "+event.GetProperty(ics.ComponentPropertySummary).Value)
	}

	return strings.Join(holidays, "\n"), nil
}
