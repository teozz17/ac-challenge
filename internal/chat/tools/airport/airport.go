package airport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/acai-travel/tech-challenge/internal/airport"
	"github.com/openai/openai-go/v2"
)

type AirportTool struct{}

func (t *AirportTool) Name() string {
	return "get_airport_info"
}

func (t *AirportTool) Description() string {
	return "Get airport information by ICAO code (4-letter airport code). NOTE: This API works best for German airports (e.g., EDDF for Frankfurt, EDDM for Munich, EDDB for Berlin). If an airport is not found, suggest trying a German airport code instead."
}

func (t *AirportTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"icao_code": map[string]string{
				"type":        "string",
				"description": "4-letter ICAO airport code (e.g., EDDF for Frankfurt, EDDM for Munich)",
			},
		},
		"required": []string{"icao_code"},
	}
}

func (t *AirportTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var payload struct {
		ICAOCode string `json:"icao_code"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return "failed to parse ICAO code parameter", nil
	}

	airportInfo, err := airport.GetAirportInfo(ctx, payload.ICAOCode)
	if err != nil {
		slog.Error("Failed to get airport info", "error", err, "icao_code", payload.ICAOCode)
		return fmt.Sprintf("failed to get airport info: %s", err.Error()), nil
	}

	return fmt.Sprintf("Airport: %s (ICAO: %s)\nWebsite: %s",
		airportInfo.Name,
		airportInfo.ICAO,
		airportInfo.URL), nil
}
