// Package api
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/ChumbaJ/go-transcriber/internal/api/dto"
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/go-chi/chi/v5"
)

type jobService interface {
	Create(ctx context.Context, r io.Reader) error
	Get(ctx context.Context, jobId int64) (*job.Job, error)
}

type Handler struct {
	jobService jobService
}

func NewHandler(js jobService) *Handler {
	return &Handler{
		jobService: js,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, _, err := r.FormFile("audio")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    strconv.Itoa(http.StatusBadRequest),
				Message: err.Error(),
			},
		})
		return
	}
	defer file.Close()

	// Create a job with audio chunks and enqueue
	err = h.jobService.Create(ctx, file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    strconv.Itoa(http.StatusBadRequest),
				Message: err.Error(),
			},
		})
		return
	}

}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawID := chi.URLParam(r, "id")

	jobId, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
	}

	input := dto.GetJobRequest{
		ID: jobId,
	}
	if err := input.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.jobService.Get(ctx, jobId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    strconv.Itoa(http.StatusBadRequest),
				Message: err.Error(),
			},
		})
		return
	}

	json.NewEncoder(w).Encode(dto.GetJobResponse{
		Job: dto.JobResponse{
			ID:         job.ID,
			Status:     job.Status,
			ResultText: job.ResultText,
		},
	})

}
