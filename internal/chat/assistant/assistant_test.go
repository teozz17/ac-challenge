package assistant_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TESTS:
// For Assistant tests, I use real test calls against OpenAI API.

func TestAssistant_Title_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	assist := assistant.New()

	conv := &model.Conversation{
		ID: primitive.NewObjectID(),
		Messages: []*model.Message{
			{
				Role:      model.RoleUser,
				Content:   "What is the weather like in Barcelona today?",
				CreatedAt: time.Now(),
			},
		},
	}

	title, err := assist.Title(ctx, conv)
	if err != nil {
		t.Fatalf("Title() error = %v", err)
	}

	if title == "" {
		t.Error("Title() returned empty string")
	}

	t.Logf("Generated title: %s", title)
}
