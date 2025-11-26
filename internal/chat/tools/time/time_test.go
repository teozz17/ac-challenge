package time

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTimeInZoneTool_Name(t *testing.T) {
	tool := &TimeInZoneTool{}
	if tool.Name() != "get_time_in_zone" {
		t.Errorf("expected name 'get_time_in_zone', got '%s'", tool.Name())
	}
}

func TestTimeInZoneTool_Execute_ValidTimezone(t *testing.T) {
	tool := &TimeInZoneTool{}
	args, _ := json.Marshal(map[string]string{"timezone": "America/New_York"})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's a valid RFC3339 timestamp
	parsedTime, parseErr := time.Parse(time.RFC3339, result)
	if parseErr != nil {
		t.Errorf("result is not valid RFC3339: %s, error: %v", result, parseErr)
	}

	// Verify the location is America/New_York by checking the offset
	loc, _ := time.LoadLocation("America/New_York")
	expectedTime := time.Now().In(loc)

	// Check that the times are within 1 minute of each other
	diff := expectedTime.Sub(parsedTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Minute {
		t.Errorf("time difference too large: %v", diff)
	}
}

func TestTimeInZoneTool_Execute_InvalidTimezone(t *testing.T) {
	tool := &TimeInZoneTool{}
	args, _ := json.Marshal(map[string]string{"timezone": "Invalid/Timezone"})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "invalid timezone") {
		t.Errorf("expected timezone error message, got: %s", result)
	}
}

func TestTimeInZoneTool_Execute_InvalidJSON(t *testing.T) {
	tool := &TimeInZoneTool{}

	result, err := tool.Execute(context.Background(), []byte("invalid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "failed to parse timezone parameter" {
		t.Errorf("expected parse error message, got: %s", result)
	}
}

func TestTimeInZoneTool_Execute_MultipleTimezones(t *testing.T) {
	tool := &TimeInZoneTool{}

	testCases := []struct {
		timezone string
	}{
		{"Europe/London"},
		{"Asia/Tokyo"},
		{"Australia/Sydney"},
		{"UTC"},
	}

	for _, tc := range testCases {
		t.Run(tc.timezone, func(t *testing.T) {
			args, _ := json.Marshal(map[string]string{"timezone": tc.timezone})

			result, err := tool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.timezone, err)
			}

			_, parseErr := time.Parse(time.RFC3339, result)
			if parseErr != nil {
				t.Errorf("result is not valid RFC3339 for %s: %s", tc.timezone, result)
			}
		})
	}
}
