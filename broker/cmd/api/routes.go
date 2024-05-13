package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)

	mux.Post("/mes", app.HandlePostMessage)
	mux.Get("/mess{hash}", app.GetHandler)
	mux.Get("/", app.HandleGetMainPage)
	return mux
}
