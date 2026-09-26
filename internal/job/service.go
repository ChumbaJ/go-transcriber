// Package job
package job

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ChumbaJ/go-transcriber/internal/infra/s3"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

type JobRepository interface {
	Create(ctx context.Context) (*Job, error)
	Get(ctx context.Context, jobID int64) (*Job, error)
}

type ChunksRepo interface {
	CreateBatch(ctx context.Context, jobID int64, chunks []*Chunk) error
}

type Storage interface {
	Upload(ctx context.Context, b []byte) (addr string, err error)
}

type Queue interface {
	Enqueue(ctx context.Context, item *queue.Item) error
}

type JobService struct {
	jobRepo    JobRepository
	chunksRepo ChunksRepo
	storage    Storage
	Queue      Queue
}

var maxChunkSize int64 = 20 * 1024 * 1024 // 20MB

func NewService(jobRepo JobRepository, chunksRepo ChunksRepo, storage Storage, queue Queue) *JobService {
	return &JobService{
		jobRepo:    jobRepo,
		chunksRepo: chunksRepo,
		storage:    storage,
		Queue:      queue,
	}
}

func (s *JobService) Create(ctx context.Context, r io.Reader) error {
	order := 0
	var chunks []*Chunk

	for {
		lr := io.LimitReader(r, maxChunkSize)
		part, err := io.ReadAll(lr)

		addr, err := s.storage.Upload(ctx, part)
		// There is nothing to read
		if errors.Is(err, s3.ErrEOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read: %w", err)
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

	// push to Queue
	for _, c := range chunks {
		qi := &queue.Item{
			JobID:      job.ID,
			ChunkOrder: c.Order,
			Addr:       c.Addr,
		}

		if err := s.Queue.Enqueue(ctx, qi); err != nil {
			return fmt.Errorf("push to queue: %w", err)
		}
	}

	return nil
}

func (s *JobService) Get(ctx context.Context, jobID int64) (*Job, error) {
	job, err := s.jobRepo.Get(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("error get job: %w", err)
	}
	return job, nil
}
