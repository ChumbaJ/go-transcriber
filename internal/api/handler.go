// Package api

package api

import (
	"context"
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
		return
	}
	defer file.Close()
}
