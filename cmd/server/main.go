package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/acai-travel/tech-challenge/internal/chat"
	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/httpx"
	"github.com/acai-travel/tech-challenge/internal/mongox"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/acai-travel/tech-challenge/internal/telemetry"
	"github.com/gorilla/mux"
	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	if os.Getenv("WEATHER_API_KEY") == "" {
		slog.Warn("WEATHER_API_KEY is not set. Weather tools will fail.")
	}

	shutdown, err := telemetry.Init(ctx, "acai-chat-service")
	if err != nil {
		slog.Error("Failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown telemetry", "error", err)
		}
	}()

	mongo := mongox.MustConnect()

	repo := model.New(mongo)
	assist := assistant.New()

	server := chat.NewServer(repo, assist)

	// Configure handler
	handler := mux.NewRouter()
	handler.Use(
		httpx.Logger(),
		httpx.Recovery(),
		otelhttp.NewMiddleware("server"),
	)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hi, my name is Clippy!")
	})

	handler.PathPrefix("/twirp/").Handler(pb.NewChatServiceServer(server, twirp.WithServerJSONSkipDefaults(true)))

	// Start the server
	slog.Info("Starting the server...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
