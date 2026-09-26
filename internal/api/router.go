// Package api

package api

import "github.com/go-chi/chi/v5"

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/jobs", func(r chi.Router) {
			r.Post("/", h.CreateJob)
			r.Get("/{id}", h.GetJob)
		})
	})

	return r
}
