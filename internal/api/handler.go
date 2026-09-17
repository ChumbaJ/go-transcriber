// Package api
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type jobService interface {
	Create(ctx context.Context, r io.Reader) error
}

type Handler struct {
	jobService jobService
}

func NewHandler(js jobService) *Handler {
	return &Handler{
		jobService: js,
	}
}

func (h *Handler) CreateTranscription(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("audio")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"msg": "error while reading file", "error": err.Error()})
		return
	}
	defer file.Close()

	err = h.jobService.Create(r.Context(), file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
}
