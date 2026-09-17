// Package job
package job

import (
	"context"
	"fmt"
	"io"
)

type JobRepository interface {
	Create(ctx context.Context) (*Job, error)
	Get(ctx context.Context, jobID string) *Job
}

type ChunksRepo interface {
	CreateBatch(ctx context.Context, jobID int64, chunks []*Chunk) error
}

type Storage interface {
	Upload(ctx context.Context, r io.Reader) (addr string, n int64, err error)
}

type JobService struct {
	jobRepo    JobRepository
	chunksRepo ChunksRepo
	storage    Storage
}

var maxChunkSize int64 = 20 * 1024 * 1024 // 20MB

func NewService(jobRepo JobRepository, chunksRepo ChunksRepo, storage Storage) *JobService {
	return &JobService{
		jobRepo:    jobRepo,
		chunksRepo: chunksRepo,
		storage:    storage,
	}
}

func (s *JobService) Create(ctx context.Context, r io.Reader) error {
	order := 0
	var chunks []*Chunk

	for {
		lr := io.LimitReader(r, maxChunkSize)
		addr, n, err := s.storage.Upload(ctx, lr)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		if n == 0 {
			break
		}

		chunks = append(chunks, &Chunk{
			Addr:  addr,
			Order: order,
		})
		order++
	}

	job, err := s.jobRepo.Create(ctx)
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	if err := s.chunksRepo.CreateBatch(ctx, job.ID, chunks); err != nil {
		return fmt.Errorf("creating chunks batch: %w", err)
	}

	return nil
}
