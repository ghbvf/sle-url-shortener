package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/user/url-shortener/internal/handler"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handler.Health)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	slog.Info("server_starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server_failed", "error", err)
		os.Exit(1)
	}
}
