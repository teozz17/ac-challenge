package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestWeatherTool_Name(t *testing.T) {
	tool := &WeatherTool{}
	if tool.Name() != "get_weather" {
		t.Errorf("expected name 'get_weather', got '%s'", tool.Name())
	}
}

func TestWeatherTool_Execute_InvalidJSON(t *testing.T) {
	tool := &WeatherTool{}

	result, err := tool.Execute(context.Background(), []byte("invalid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "failed to parse location parameter" {
		t.Errorf("expected parse error message, got: %s", result)
	}
}

func TestWeatherTool_Execute_MissingAPIKey(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("WEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("WEATHER_API_KEY", originalKey)
		}
	}()

	// Unset the key
	os.Unsetenv("WEATHER_API_KEY")

	tool := &WeatherTool{}
	args, _ := json.Marshal(map[string]string{"location": "Barcelona"})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "WEATHER_API_KEY") {
		t.Errorf("expected API key error, got: %s", result)
	}
}

// Integration test - only runs if WEATHER_API_KEY is set
func TestWeatherTool_Execute_Integration(t *testing.T) {
	if os.Getenv("WEATHER_API_KEY") == "" {
		t.Skip("Skipping integration test: WEATHER_API_KEY not set")
	}

	tool := &WeatherTool{}
	args, _ := json.Marshal(map[string]string{"location": "London"})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the result contains expected information
	if !strings.Contains(result, "London") {
		t.Errorf("expected London in result, got: %s", result)
	}

	if !strings.Contains(result, "Temperature") {
		t.Errorf("expected Temperature in result, got: %s", result)
	}
}

func TestForecastTool_Name(t *testing.T) {
	tool := &ForecastTool{}
	if tool.Name() != "get_forecast" {
		t.Errorf("expected name 'get_forecast', got '%s'", tool.Name())
	}
}

func TestForecastTool_Execute_InvalidJSON(t *testing.T) {
	tool := &ForecastTool{}

	result, err := tool.Execute(context.Background(), []byte("invalid json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "failed to parse arguments" {
		t.Errorf("expected parse error message, got: %s", result)
	}
}

func TestForecastTool_Execute_DefaultDays(t *testing.T) {
	if os.Getenv("WEATHER_API_KEY") == "" {
		t.Skip("Skipping integration test: WEATHER_API_KEY not set")
	}

	tool := &ForecastTool{}
	// Don't specify days - should default to 3
	args, _ := json.Marshal(map[string]string{"location": "Paris"})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Paris") {
		t.Errorf("expected Paris in result, got: %s", result)
	}
}

// Integration test for forecast with specific date
func TestForecastTool_Execute_WithDate_Integration(t *testing.T) {
	if os.Getenv("WEATHER_API_KEY") == "" {
		t.Skip("Skipping integration test: WEATHER_API_KEY not set")
	}

	tool := &ForecastTool{}
	args, _ := json.Marshal(map[string]any{
		"location": "Tokyo",
		"date":     "2025-11-26",
		"days":     1,
	})

	result, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Tokyo") {
		t.Errorf("expected Tokyo in result, got: %s", result)
	}

	if !strings.Contains(result, "2025-11-26") {
		t.Errorf("expected date 2025-11-26 in result, got: %s", result)
	}
}
