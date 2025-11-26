package airport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://airport-web.appspot.com/_ah/api/airportsapi/v1/airports"

type AirportInfo struct {
	ICAO       string `json:"ICAO"`
	LastUpdate string `json:"last_update"`
	Name       string `json:"name"`
	URL        string `json:"url"`
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Domain  string `json:"domain"`
			Message string `json:"message"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error"`
}

func GetAirportInfo(ctx context.Context, icaoCode string) (*AirportInfo, error) {
	// Validate ICAO code (should be 4 characters, uppercase letters)
	icaoCode = strings.ToUpper(strings.TrimSpace(icaoCode))
	if len(icaoCode) != 4 {
		return nil, fmt.Errorf("invalid ICAO code: must be 4 characters")
	}

	url := fmt.Sprintf("%s/%s", baseURL, icaoCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	if resp.StatusCode == http.StatusNotFound {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("airport not found: %s", icaoCode)
		}
		return nil, fmt.Errorf("airport not found: %s", errResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var airportInfo AirportInfo
	if err := json.Unmarshal(body, &airportInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &airportInfo, nil
}
