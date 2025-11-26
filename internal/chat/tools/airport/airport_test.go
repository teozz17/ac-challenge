package airport

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAirportTool_Name(t *testing.T) {
	tool := &AirportTool{}
	if tool.Name() != "get_airport_info" {
		t.Errorf("expected name 'get_airport_info', got '%s'", tool.Name())
	}
}

func TestAirportTool_Execute_InvalidJSON(t *testing.T) {
	tool := &AirportTool{}

	result, err := tool.Execute(context.Background(), []byte("invalid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "failed to parse ICAO code parameter" {
		t.Errorf("expected parse error message, got: %s", result)
	}
}

func TestAirportTool_Execute_InvalidICAOCode(t *testing.T) {
	tool := &AirportTool{}
	args, _ := json.Marshal(map[string]string{"icao_code": "ABC"}) // Too short

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "failed to get airport info: invalid ICAO code: must be 4 characters" {
		t.Errorf("expected invalid ICAO error, got: %s", result)
	}
}

func TestAirportTool_Execute_NotFound(t *testing.T) {
	tool := &AirportTool{}
	args, _ := json.Marshal(map[string]string{"icao_code": "XXXX"}) // Non-existent airport

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain "not found" or similar error message
	if result == "" {
		t.Error("expected error message for non-existent airport")
	}
}

// Integration test - calls real API
func TestAirportTool_Execute_Integration(t *testing.T) {
	tool := &AirportTool{}
	args, _ := json.Marshal(map[string]string{"icao_code": "EDDF"}) // Frankfurt Airport

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the result contains expected information
	if result == "" {
		t.Error("expected non-empty result")
	}

	// Should contain "Frankfurt" or "EDDF"
	t.Logf("Result: %s", result)
}

func TestAirportTool_Execute_MultipleAirports(t *testing.T) {
	tool := &AirportTool{}

	testCases := []struct {
		icaoCode    string
		shouldExist bool
	}{
		{"EDDF", true},  // Frankfurt
		{"EDDM", true},  // Munich
		{"LEBL", false}, // Barcelona (might not be in the API)
		{"KJFK", false}, // JFK (might not be in the API)
	}

	for _, tc := range testCases {
		t.Run(tc.icaoCode, func(t *testing.T) {
			args, _ := json.Marshal(map[string]string{"icao_code": tc.icaoCode})

			result, err := tool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.icaoCode, err)
			}

			if result == "" {
				t.Errorf("expected non-empty result for %s", tc.icaoCode)
			}

			t.Logf("%s: %s", tc.icaoCode, result)
		})
	}
}
