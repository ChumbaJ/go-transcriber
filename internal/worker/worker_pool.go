// Package worker
package worker

import (
	"context"
	"strconv"

	"github.com/ChumbaJ/go-transcriber/internal/job"
)

type queue interface {
	Dequeue(ctx context.Context, consumerName string) (*job.QueueMessage, error)
	Ack(ctx context.Context, messageID string) error
	ClaimStale(ctx context.Context, consumerName string) (*job.QueueMessage, error)
}

type chunkStorage interface {
	Get(ctx context.Context, addr string) ([]byte, error)
}

type (
	jobRepository interface {
		UpdateStatus(ctx context.Context, jobID int64, status job.JobStatus) error
		SetResult(ctx context.Context, jobID int64, result string) error
	}
	chunkRepository interface {
		CountByJobID(ctx context.Context, jobID int64) (int, error)
	}
	transcriptionsRepository interface {
		Create(ctx context.Context, t job.Transcribtion) error
		CountByJobID(ctx context.Context, jobID int64) (int, error)
		ListByJobID(ctx context.Context, jobID int64) ([]job.Transcribtion, error)
	}
)

type WorkerPool struct {
	// Number of workers
	count         uint8
	queue         queue
	storage       chunkStorage
	chunksRepo    chunkRepository
	jobsRepo      jobRepository
	transcripRepo transcriptionsRepository
}

func NewPool(count uint8, queue queue, cs chunkStorage, jobRepo jobRepository, chunksRepo chunkRepository, transRepo transcriptionsRepository) *WorkerPool {
	return &WorkerPool{
		count:         count,
		queue:         queue,
		storage:       cs,
		jobsRepo:      jobRepo,
		chunksRepo:    chunksRepo,
		transcripRepo: transRepo,
	}
}

// Run runs workers goroutines
func (wp *WorkerPool) Run(ctx context.Context) {
	for i := range wp.count {
		wName := "worker-" + strconv.Itoa(int(i))

		go wp.runWorker(ctx, wName)
	}
}
