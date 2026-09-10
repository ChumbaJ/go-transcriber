// Package job
package job

import (
	"context"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, j Job) error
	Get(ctx context.Context, jobID string) *Job
}

type JobService struct {
	repo Repository
}

func NewService(jobRepo Repository) *JobService {
	return &JobService{
		repo: jobRepo,
	}
}

func (s *JobService) Create(ctx context.Context, j Job) {
	if err := s.repo.Create(ctx, j); err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}
