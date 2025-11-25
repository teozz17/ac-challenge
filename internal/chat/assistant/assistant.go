package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/weather"
	ics "github.com/arran4/golang-ical"
	"github.com/openai/openai-go/v2"
)

type Assistant struct {
	cli openai.Client
}

func New() *Assistant {
	return &Assistant{cli: openai.NewClient()}
}

func (a *Assistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "An empty conversation", nil
	}

	slog.InfoContext(ctx, "Generating title for conversation", "conversation_id", conv.ID)

	// Build messages array with system instruction first
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a title generator. Create a concise, descriptive title that SUMMARIZES the topic of the user's question. Do NOT answer the question. The title should be 3-8 words maximum, no special characters or emojis. Examples: 'Weather in Barcelona', 'Today's Date', 'Upcoming Holidays'."),
	}

	for _, m := range conv.Messages {
		if m.Role == model.RoleUser {
			msgs = append(msgs, openai.UserMessage(m.Content))
		}
	}

	// We could also change the model to make it quicker (GPT-4o or GPT-3.5-turbo)
	resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelO1,
		Messages: msgs,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return "", errors.New("empty response from OpenAI for title generation")
	}

	title := resp.Choices[0].Message.Content
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.Trim(title, " \t\r\n-\"'")

	if len(title) > 80 {
		title = title[:80]
	}

	return title, nil
}

func (a *Assistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "", errors.New("conversation has no messages")
	}

	slog.InfoContext(ctx, "Generating reply for conversation", "conversation_id", conv.ID)

	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful, concise AI assistant. Provide accurate, safe, and clear responses."),
	}

	for _, m := range conv.Messages {
		switch m.Role {
		case model.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Content))
		case model.RoleAssistant:
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		}
	}

	for i := 0; i < 15; i++ {
		resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModelGPT4_1,
			Messages: msgs,
			Tools: []openai.ChatCompletionToolUnionParam{
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "get_weather",
					Description: openai.String("Get weather at the given location"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]string{
								"type": "string",
							},
						},
						"required": []string{"location"},
					},
				}),
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "get_forecast",
					Description: openai.String("Get weather forecast for a location"),
					Parameters: openai.FunctionParameters{
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
					},
				}),
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "get_today_date",
					Description: openai.String("Get today's date and time in RFC3339 format"),
				}),
				openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
					Name:        "get_holidays",
					Description: openai.String("Gets local bank and public holidays. Each line is a single holiday in the format 'YYYY-MM-DD: Holiday Name'."),
					Parameters: openai.FunctionParameters{
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
					},
				}),
			},
		})

		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", errors.New("no choices returned by OpenAI")
		}

		if message := resp.Choices[0].Message; len(message.ToolCalls) > 0 {
			msgs = append(msgs, message.ToParam())

			for _, call := range message.ToolCalls {
				slog.InfoContext(ctx, "Tool call received", "name", call.Function.Name, "args", call.Function.Arguments)

				switch call.Function.Name {
				case "get_weather":
					var weatherPayload struct {
						Location string `json:"location"`
					}
					if err := json.Unmarshal([]byte(call.Function.Arguments), &weatherPayload); err != nil {
						msgs = append(msgs, openai.ToolMessage("failed to parse location parameter", call.ID))
						break
					}

					weatherData, err := weather.GetCurrentWeather(ctx, weatherPayload.Location)
					if err != nil {
						slog.ErrorContext(ctx, "Failed to get weather", "location", weatherPayload.Location, "error", err)
						msgs = append(msgs, openai.ToolMessage(fmt.Sprintf("failed to get weather: %s", err.Error()), call.ID))
						break
					}

					weatherMsg := fmt.Sprintf("Weather in %s, %s: %s, Temperature: %.1f°C, Feels like: %.1f°C, Wind: %.1f km/h %s, Humidity: %d%%, Cloud coverage: %d%%",
						weatherData.Location.Name,
						weatherData.Location.Country,
						weatherData.Current.Condition.Text,
						weatherData.Current.TempC,
						weatherData.Current.FeelslikeC,
						weatherData.Current.WindKph,
						weatherData.Current.WindDir,
						weatherData.Current.Humidity,
						weatherData.Current.Cloud)

					msgs = append(msgs, openai.ToolMessage(weatherMsg, call.ID))
				case "get_forecast":
					var forecastPayload struct {
						Location string `json:"location"`
						Days     int    `json:"days"`
						Hour     *int   `json:"hour,omitempty"`
						Date     string `json:"date,omitempty"`
					}
					if err := json.Unmarshal([]byte(call.Function.Arguments), &forecastPayload); err != nil {
						msgs = append(msgs, openai.ToolMessage("failed to parse arguments", call.ID))
						break
					}
					if forecastPayload.Days == 0 {
						// Default to 3 days if not specified.
						forecastPayload.Days = 3
					}

					if forecastPayload.Date != "" {
						// TODO: We could validate here if the date is within the next 14 days to avoid an API call
						// but for now we let the API handle the error.
						parsedDate, err := time.Parse("2006-01-02", forecastPayload.Date)
						if err != nil {
							msgs = append(msgs, openai.ToolMessage("invalid date format, use YYYY-MM-DD", call.ID))
							break
						}
						// Just to ensure format is correct, we use the parsed date string back if needed,
						// but forecastPayload.Date is already string.
						_ = parsedDate
					}

					forecast, err := weather.GetForecast(ctx, forecastPayload.Location, forecastPayload.Days, forecastPayload.Hour, forecastPayload.Date)
					if err != nil {
						slog.ErrorContext(ctx, "Failed to get forecast", "location", forecastPayload.Location, "error", err)
						msgs = append(msgs, openai.ToolMessage(fmt.Sprintf("failed to get forecast: %s", err.Error()), call.ID))
						break
					}

					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("Forecast for %s, %s:\n", forecast.Location.Name, forecast.Location.Country))

					for _, day := range forecast.Forecast.ForecastDay {
						// If specific hour requested, show hour details
						if forecastPayload.Hour != nil && len(day.Hour) > 0 {
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
					msgs = append(msgs, openai.ToolMessage(sb.String(), call.ID))
				case "get_today_date":
					msgs = append(msgs, openai.ToolMessage(time.Now().Format(time.RFC3339), call.ID))
				case "get_holidays":
					link := "https://www.officeholidays.com/ics/spain/catalonia"
					if v := os.Getenv("HOLIDAY_CALENDAR_LINK"); v != "" {
						link = v
					}

					events, err := LoadCalendar(ctx, link)
					if err != nil {
						msgs = append(msgs, openai.ToolMessage("failed to load holiday events", call.ID))
						break
					}

					var payload struct {
						BeforeDate time.Time `json:"before_date,omitempty"`
						AfterDate  time.Time `json:"after_date,omitempty"`
						MaxCount   int       `json:"max_count,omitempty"`
					}

					if err := json.Unmarshal([]byte(call.Function.Arguments), &payload); err != nil {
						msgs = append(msgs, openai.ToolMessage("failed to parse tool call arguments: "+err.Error(), call.ID))
						break
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

					msgs = append(msgs, openai.ToolMessage(strings.Join(holidays, "\n"), call.ID))
				default:
					return "", errors.New("unknown tool call: " + call.Function.Name)
				}
			}

			continue
		}

		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("too many tool calls, unable to generate reply")
}
