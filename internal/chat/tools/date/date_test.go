package date

import (
	"context"
	"testing"
	"time"
)

func TestDateTool_Name(t *testing.T) {
	tool := &DateTool{}
	if tool.Name() != "get_today_date" {
		t.Errorf("expected name 'get_today_date', got '%s'", tool.Name())
	}
}

func TestDateTool_Execute(t *testing.T) {
	tool := &DateTool{}

	result, err := tool.Execute(context.Background(), []byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it returns a valid RFC3339 timestamp
	parsedTime, parseErr := time.Parse(time.RFC3339, result)
	if parseErr != nil {
		t.Errorf("result is not valid RFC3339: %s, error: %v", result, parseErr)
	}

	// Verify the date is reasonably close to now (within 1 minute)
	now := time.Now()
	diff := now.Sub(parsedTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Minute {
		t.Errorf("returned time is too far from current time: %s (now: %s)", result, now.Format(time.RFC3339))
	}
}
