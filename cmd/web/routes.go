package main

import (
	"net/http"

	"github.com/bmizerany/pat"

	"github.com/MatheusAbdias/go-simple-chat/internal/handlers"
)

func routes() http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsUpgrade))

	return mux
}
