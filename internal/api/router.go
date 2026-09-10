// Package api

package api

import "github.com/go-chi/chi/v5"

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/transcriptions", func(r chi.Router) {
			r.Post("/", h.CreateTranscription)
		})
	})

	return r
}
