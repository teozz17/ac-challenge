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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Assistant struct {
	cli      openai.Client
	registry *tools.Registry
	tracer   trace.Tracer
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
		tracer:   otel.Tracer("assistant"),
	}
}

func (a *Assistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	ctx, span := a.tracer.Start(ctx, "Assistant.Title")
	defer span.End()

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
	ctx, span := a.tracer.Start(ctx, "Assistant.Reply")
	defer span.End()

	if len(conv.Messages) == 0 {
		return "", errors.New("conversation has no messages")
	}

	slog.InfoContext(ctx, "Generating reply for conversation", "conversation_id", conv.ID)

	// Build message history with system prompt
	msgs := []openai.ChatCompletionMessageParamUnion{
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
	iteration := 0

	for {
		response, err := a.callGPT4(ctx, msgs, toolDefs)
		if err != nil {
			return "", err
		}

		shouldContinue, finalAnswer := a.shouldContinue(ctx, response, iteration)
		if !shouldContinue {
			return finalAnswer, nil
		}

		msgs = a.executeTools(ctx, msgs, response)
		iteration++
	}
}

func (a *Assistant) shouldContinue(ctx context.Context, response *openai.ChatCompletion, iteration int) (bool, string) {
	if len(response.Choices[0].Message.ToolCalls) == 0 {
		slog.InfoContext(ctx, "Agent completed", "iterations", iteration)
		return false, response.Choices[0].Message.Content
	}

	const maxIterations = 15
	if iteration >= maxIterations {
		slog.WarnContext(ctx, "Max iterations reached", "iteration", iteration)
		return false, "I apologize, but I've reached the maximum number of processing steps. Please try rephrasing your question or breaking it into smaller parts."
	}

	slog.InfoContext(ctx, "Continuing agent loop", "iteration", iteration, "tool_calls", len(response.Choices[0].Message.ToolCalls))
	return true, ""
}

func (a *Assistant) callGPT4(ctx context.Context, msgs []openai.ChatCompletionMessageParamUnion, toolDefs []openai.ChatCompletionToolUnionParam) (*openai.ChatCompletion, error) {
	resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelGPT4_1,
		Messages: msgs,
		Tools:    toolDefs,
	})

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned by OpenAI")
	}

	return resp, nil
}

func (a *Assistant) executeTools(ctx context.Context, msgs []openai.ChatCompletionMessageParamUnion, response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	message := response.Choices[0].Message

	msgs = append(msgs, message.ToParam())

	for _, call := range message.ToolCalls {
		result := a.executeSingleTool(ctx, call)
		msgs = append(msgs, openai.ToolMessage(result, call.ID))
	}

	return msgs
}

func (a *Assistant) executeSingleTool(ctx context.Context, call any) string {
	type toolCall interface {
		GetID() string
		GetFunction() struct {
			Name      string
			Arguments string
		}
	}

	tc, ok := call.(toolCall)
	if !ok {
		slog.WarnContext(ctx, "Unexpected tool call type")
		return "Error: unexpected tool call type"
	}

	fn := tc.GetFunction()
	slog.InfoContext(ctx, "Tool call received", "name", fn.Name, "args", fn.Arguments)

	toolCtx, toolSpan := a.tracer.Start(ctx, "Tool.Execute", trace.WithAttributes(
		attribute.String("tool.name", fn.Name),
		attribute.String("tool.args", fn.Arguments),
	))
	defer toolSpan.End()

	result, err := a.registry.Execute(toolCtx, fn.Name, fn.Arguments)
	if err != nil {
		slog.ErrorContext(ctx, "Tool execution failed", "tool", fn.Name, "error", err)
		toolSpan.RecordError(err)
		toolSpan.SetStatus(codes.Error, err.Error())
		return "Error executing tool: " + err.Error()
	}

	return result
}
