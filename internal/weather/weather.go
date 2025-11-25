package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	baseURL = "https://api.weatherapi.com/v1"
)

func GetCurrentWeather(ctx context.Context, location string) (*WeatherResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEATHER_API_KEY environment variable not set")
	}

	u, err := url.Parse(baseURL + "/current.json")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("key", apiKey)
	q.Set("q", location)
	q.Set("lang", "en")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error %d: %s", errResp.Error.Code, errResp.Error.Message)
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &weatherResp, nil
}
