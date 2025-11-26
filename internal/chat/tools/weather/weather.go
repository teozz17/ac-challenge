package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/acai-travel/tech-challenge/internal/weather"
	"github.com/openai/openai-go/v2"
)

type WeatherTool struct{}

func (t *WeatherTool) Name() string {
	return "get_weather"
}

func (t *WeatherTool) Description() string {
	return "Get weather at the given location"
}

func (t *WeatherTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]string{
				"type": "string",
			},
		},
		"required": []string{"location"},
	}
}

func (t *WeatherTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var payload struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return "failed to parse location parameter", nil
	}

	weatherData, err := weather.GetCurrentWeather(ctx, payload.Location)
	if err != nil {
		slog.Error("Failed to get weather", "error", err, "location", payload.Location)
		return fmt.Sprintf("failed to get weather: %s", err.Error()), nil
	}

	return fmt.Sprintf("Weather in %s, %s: %s, Temperature: %.1f°C, Feels like: %.1f°C, Wind: %.1f km/h %s, Humidity: %d%%, Cloud coverage: %d%%",
		weatherData.Location.Name,
		weatherData.Location.Country,
		weatherData.Current.Condition.Text,
		weatherData.Current.TempC,
		weatherData.Current.FeelslikeC,
		weatherData.Current.WindKph,
		weatherData.Current.WindDir,
		weatherData.Current.Humidity,
		weatherData.Current.Cloud), nil
}

type ForecastTool struct{}

func (t *ForecastTool) Name() string {
	return "get_forecast"
}

func (t *ForecastTool) Description() string {
	return "Get weather forecast for a location"
}

func (t *ForecastTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]string{
				"type":        "string",
				"description": "City name or coordinates",
			},
			"days": map[string]any{
				"type":        "integer",
				"description": "Number of days (1-14)",
			},
			"hour": map[string]any{
				"type":        "integer",
				"description": "Specific hour (0-23) to get forecast for",
			},
			"date": map[string]any{
				"type":        "string",
				"description": "Specific date in YYYY-MM-DD format. Must be within next 14 days.",
			},
		},
		"required": []string{"location"},
	}
}

func (t *ForecastTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var payload struct {
		Location string `json:"location"`
		Days     int    `json:"days"`
		Hour     *int   `json:"hour,omitempty"`
		Date     string `json:"date,omitempty"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return "failed to parse arguments", nil
	}
	if payload.Days == 0 {
		payload.Days = 3
	}

	// Note: Client-side date validation was removed as per user request.
	// API handles validation for dates > 14 days.

	forecast, err := weather.GetForecast(ctx, payload.Location, payload.Days, payload.Hour, payload.Date)
	if err != nil {
		slog.Error("Failed to get forecast", "error", err, "location", payload.Location)
		return fmt.Sprintf("failed to get forecast: %s", err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Forecast for %s, %s:\n", forecast.Location.Name, forecast.Location.Country))

	for _, day := range forecast.Forecast.ForecastDay {
		// If specific hour requested, show hour details
		if payload.Hour != nil && len(day.Hour) > 0 {
			// The API returns only the requested hour in the hour array
			h := day.Hour[0]
			sb.WriteString(fmt.Sprintf("- %s %s: %s, Temp: %.1f°C, Rain: %d%%\n",
				day.Date,
				strings.Split(h.Time, " ")[1], // Extract time part
				h.Condition.Text,
				h.TempC,
				h.ChanceOfRain))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: %s, Max: %.1f°C, Min: %.1f°C, Rain: %d%%\n",
				day.Date,
				day.Day.Condition.Text,
				day.Day.MaxTempC,
				day.Day.MinTempC,
				day.Day.DailyChanceOfRain))
		}
	}
	return sb.String(), nil
}
