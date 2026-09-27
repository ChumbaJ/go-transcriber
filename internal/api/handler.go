// Package api
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
	logger     *slog.Logger
}

func NewHandler(js jobService, logger *slog.Logger) *Handler {
	return &Handler{
		jobService: js,
		logger:     logger,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, _, err := r.FormFile("audio")
	if err != nil {
		h.logger.Error("error reading file:", "error", err)
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

	// TODO: 400 & 500 codes

	// Create a job with audio chunks and enqueue
	err = h.jobService.Create(ctx, file)
	if err != nil {
		h.logger.Error("error creating job: ", "error", err)
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
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	rawID := chi.URLParam(r, "id")

	jobId, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		h.logger.Error("error parsing jobid from request:", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    strconv.Itoa(http.StatusBadRequest),
				Message: "invalid job id",
			},
		})
		return
	}

	input := dto.GetJobRequest{
		ID: jobId,
	}
	if err := input.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    strconv.Itoa(http.StatusBadRequest),
				Message: err.Error(),
			},
		})
		return
	}

	j, err := h.jobService.Get(ctx, jobId)
	if err != nil {

		if errors.Is(err, job.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "job_not_found",
					Message: "job is not found",
				},
			})
			return
		}

		h.logger.ErrorContext(ctx, "get job failed", "job_id", jobId, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "internal_error",
				Message: "Internal server error",
			},
		})
		return
	}

	json.NewEncoder(w).Encode(dto.GetJobResponse{
		Job: dto.JobResponse{
			ID:         j.ID,
			Status:     j.Status,
			ResultText: j.ResultText,
		},
	})

}
