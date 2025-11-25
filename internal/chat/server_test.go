package chat

import (
	"context"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/testing/protocmp"
)

// TESTS:
// For Server tests, we use a MockAssistant to test the server logic without depending on external services.

// MockAssistant is a mock implementation of the Assistant interface
type MockAssistant struct {
	TitleFunc func(ctx context.Context, conv *model.Conversation) (string, error)
	ReplyFunc func(ctx context.Context, conv *model.Conversation) (string, error)
}

func (m *MockAssistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.TitleFunc != nil {
		return m.TitleFunc(ctx, conv)
	}
	return "Mock Title", nil
}

func (m *MockAssistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.ReplyFunc != nil {
		return m.ReplyFunc(ctx, conv)
	}
	return "Mock Reply", nil
}

func TestServer_StartConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("start conversation successfully", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			TitleFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "Generated Title", nil
			},
			ReplyFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "Generated Reply", nil
			},
		}

		srv := NewServer(model.New(ConnectMongo()), mockAssist)

		req := &pb.StartConversationRequest{
			Message: "Hello, world!",
		}

		resp, err := srv.StartConversation(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Title != "Generated Title" {
			t.Errorf("expected title 'Generated Title', got '%s'", resp.Title)
		}

		if resp.Reply != "Generated Reply" {
			t.Errorf("expected reply 'Generated Reply', got '%s'", resp.Reply)
		}

		// Verify conversation was created in DB
		conv, err := srv.repo.DescribeConversation(ctx, resp.ConversationId)
		if err != nil {
			t.Fatalf("failed to retrieve conversation from DB: %v", err)
		}

		if conv.Title != "Generated Title" {
			t.Errorf("db title mismatch: expected 'Generated Title', got '%s'", conv.Title)
		}

		if len(conv.Messages) != 2 { // User message + Assistant reply
			t.Errorf("expected 2 messages, got %d", len(conv.Messages))
		}
	}))

	t.Run("start conversation with empty message should fail", WithFixture(func(t *testing.T, f *Fixture) {
		srv := NewServer(model.New(ConnectMongo()), &MockAssistant{})

		_, err := srv.StartConversation(ctx, &pb.StartConversationRequest{Message: ""})
		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.InvalidArgument {
			t.Fatalf("expected twirp.InvalidArgument error, got %v", err)
		}
	}))
}

func TestServer_DescribeConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), nil)

	t.Run("describe existing conversation", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()

		out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: c.ID.Hex()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, want := out.GetConversation(), c.Proto()
		if !cmp.Equal(got, want, protocmp.Transform()) {
			t.Errorf("DescribeConversation() mismatch (-got +want):\n%s", cmp.Diff(got, want, protocmp.Transform()))
		}
	}))

	t.Run("describe non existing conversation should return 404", WithFixture(func(t *testing.T, f *Fixture) {
		_, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: "08a59244257c872c5943e2a2"})
		if err == nil {
			t.Fatal("expected error for non-existing conversation, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.NotFound {
			t.Fatalf("expected twirp.NotFound error, got %v", err)
		}
	}))
}
