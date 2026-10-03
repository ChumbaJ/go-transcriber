// Package job
package job

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

type JobRepository interface {
	Create(ctx context.Context, startedAt time.Time) (*Job, error)
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

type audioProcessor interface {
	Split(ctx context.Context, r io.Reader) ([][]byte, string, error)
}

type JobService struct {
	jobRepo        JobRepository
	chunksRepo     ChunksRepo
	storage        Storage
	Queue          Queue
	audioProcessor audioProcessor
}

var MaxChunkSize int64 = 20 * 1024 * 1024 // 20 MiB, below OpenAI's 25 MB file limit

func NewService(jobRepo JobRepository, chunksRepo ChunksRepo, storage Storage, queue Queue, ap audioProcessor) *JobService {
	return &JobService{
		jobRepo:        jobRepo,
		chunksRepo:     chunksRepo,
		storage:        storage,
		Queue:          queue,
		audioProcessor: ap,
	}
}

func (s *JobService) Create(ctx context.Context, r io.Reader, startedAt time.Time) error {
	parts, format, err := s.audioProcessor.Split(ctx, r)
	if err != nil {
		return fmt.Errorf("audio processor: %w", err)
	}

	chunks, err := s.uploadChunks(ctx, parts)
	if err != nil {
		return fmt.Errorf("upload chunks: %w", err)
	}

	job, err := s.jobRepo.Create(ctx, startedAt)
	if err != nil {
		return fmt.Errorf("error creating job: %w", err)
	}

	// Разобраться как это работает
	if err := s.chunksRepo.CreateBatch(ctx, job.ID, chunks); err != nil {
		return fmt.Errorf("error creating chunks batch: %w", err)
	}

	if err := s.pushChunksToQueue(ctx, job, chunks, format); err != nil {
		return err
	}

	return nil
}

func (s *JobService) uploadChunks(ctx context.Context, parts [][]byte) ([]*Chunk, error) {
	var chunks []*Chunk

	for i, part := range parts {
		addr, err := s.storage.Upload(ctx, part)
		if err != nil {
			return nil, fmt.Errorf("storage upload: %w", err)
		}

		chunks = append(chunks, &Chunk{
			Addr:  addr,
			Order: i,
		})
	}

	return chunks, nil
}

func (s *JobService) pushChunksToQueue(ctx context.Context, job *Job, chunks []*Chunk, format string) error {
	// push to Queue
	for _, c := range chunks {
		qi := &queue.Item{
			JobID:      job.ID,
			ChunkOrder: c.Order,
			Addr:       c.Addr,
			Format:     format,
		}

		if err := s.Queue.Enqueue(ctx, qi); err != nil {
			return fmt.Errorf("error pushing to queue: %w", err)
		}
	}

	return nil
}

func (s *JobService) Get(ctx context.Context, jobID int64) (*Job, error) {
	job, err := s.jobRepo.Get(ctx, jobID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("get job: %w", err)
		}
		return nil, fmt.Errorf("error get job: %w", err)
	}
	return job, nil
}
