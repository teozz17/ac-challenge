package assistant

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/chat/tools"
	"github.com/acai-travel/tech-challenge/internal/chat/tools/airport"
	"github.com/acai-travel/tech-challenge/internal/chat/tools/date"
	"github.com/acai-travel/tech-challenge/internal/chat/tools/holidays"
	timetools "github.com/acai-travel/tech-challenge/internal/chat/tools/time"
	"github.com/acai-travel/tech-challenge/internal/chat/tools/weather"
	"github.com/openai/openai-go/v2"
)

type Assistant struct {
	cli      openai.Client
	registry *tools.Registry
}

func New() *Assistant {
	registry := tools.NewRegistry()
	registry.Register(&weather.WeatherTool{})
	registry.Register(&weather.ForecastTool{})
	registry.Register(&date.DateTool{})
	registry.Register(&holidays.HolidaysTool{})
	registry.Register(&timetools.TimeInZoneTool{})
	registry.Register(&airport.AirportTool{})

	return &Assistant{
		cli:      openai.NewClient(),
		registry: registry,
	}
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
		openai.SystemMessage("You are a helpful, concise AI assistant. Provide accurate, safe, and clear responses. IMPORTANT: When users ask about relative dates like 'tomorrow', 'next week', etc., ALWAYS call get_today_date first to get the current date, then calculate the target date from that result. Pay close attention to the year."),
		openai.SystemMessage("You are a helpful, concise AI assistant specialized in Weather, Holidays, and German Airports (ICAO codes). You can answer general questions normally, but when doing so, briefly mention that your primary expertise lies in Weather, Holidays, and German Airports. Provide accurate, safe, and clear responses. IMPORTANT: When users ask about relative dates like 'tomorrow', 'next week', etc., ALWAYS call get_today_date first to get the current date, then calculate the target date from that result. Pay close attention to the year."),
	}

	for _, m := range conv.Messages {
		switch m.Role {
		case model.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Content))
		case model.RoleAssistant:
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		}
	}

	toolDefs := a.registry.Definitions()

	for i := 0; i < 15; i++ {
		resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModelGPT4_1,
			Messages: msgs,
			Tools:    toolDefs,
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

				result, err := a.registry.Execute(ctx, call.Function.Name, call.Function.Arguments)
				if err != nil {
					slog.ErrorContext(ctx, "Tool execution failed", "tool", call.Function.Name, "error", err)
					msgs = append(msgs, openai.ToolMessage("Error executing tool: "+err.Error(), call.ID))
				} else {
					msgs = append(msgs, openai.ToolMessage(result, call.ID))
				}
			}

			continue
		}

		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("too many tool calls, unable to generate reply")
}
