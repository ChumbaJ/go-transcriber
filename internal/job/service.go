// Package job
package job

import (
	"context"
	"io"
)

type Repository interface {
	Create(ctx context.Context, j Job) error
	Get(ctx context.Context, jobID string) *Job
}

type Storage interface {
	Upload(ctx context.Context, r io.Reader) (addr string, err error)
}

type JobService struct {
	repo    Repository
	storage Storage
}

var maxChunkSize int64 = 20 * 1024 * 1024 // 20MB

func NewService(jobRepo Repository) *JobService {
	return &JobService{
		repo: jobRepo,
	}
}

func (s *JobService) Create(ctx context.Context, r io.Reader) error {
	order := 0
	var chunks []*Chunk

	for {
		lr := io.LimitReader(r, maxChunkSize)
		addr, err := s.storage.Upload(ctx, lr)

		if err == io.EOF {
			return nil
		}

		chunks = append(chunks, &Chunk{
			Addr:  addr,
			Order: order,
		})
		order++
	}
}
