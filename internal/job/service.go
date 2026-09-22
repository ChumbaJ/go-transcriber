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

type QueueItem struct {
	jobID      int64
	ChunkOrder int
	Addr       string
}

type Queue interface {
	ProduceJob(ctx context.Context, item *QueueItem) error
	ConsumeJob(ctx context.Context) error
	Log(ctx context.Context)
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

	// push to Queue
	for _, c := range chunks {
		qi := &QueueItem{
			jobID:      job.ID,
			ChunkOrder: c.Order,
			Addr:       c.Addr,
		}

		if err := s.Queue.ProduceJob(ctx, qi); err != nil {
			return fmt.Errorf("push to queue: %w", err)
		}
	}

	// log what we pushed
	if err := s.Queue.ConsumeJob(ctx); err != nil {
		return fmt.Errorf("consume job: %w", err)
	}

	return nil
}
