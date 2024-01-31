package main

import (
	"log/slog"
	"net/http"

	"github.com/MatheusAbdias/go-simple-chat/internal/handlers"
)

func main() {

	mux := routes()

	slog.Info("Stating channel listener")
	go handlers.ListenToWsChannel()

	slog.Info("Starting web server on port  8080")

	_ = http.ListenAndServe(":8080", mux)
}
