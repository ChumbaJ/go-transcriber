// Package worker
package worker

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

type Queue interface {
	Dequeue(ctx context.Context, consumerName string) (*queue.Message, error)
	Confirm(ctx context.Context, messageID string) error
	ClaimStale(ctx context.Context, consumerName string) (*queue.Message, error)
}

type (
	jobRepository interface {
		UpdateStatus(ctx context.Context, jobID int64, status job.JobStatus) error
		SetResult(ctx context.Context, jobID int64, result string) error
		DeleteByID(ctx context.Context, jobID int64) error
	}
	chunkRepository interface {
		CountByJobID(ctx context.Context, jobID int64) (int, error)
	}
	transcriptionsRepository interface {
		Create(ctx context.Context, t job.Transcribtion) error
		CountByJobID(ctx context.Context, jobID int64) (int, error)
		ListByJobID(ctx context.Context, jobID int64) ([]job.Transcribtion, error)
		ExistsByJobIDAndChunkOrder(ctx context.Context, jobID int64, chunkOrder int) (bool, error)
	}
	chunkStorage interface {
		Get(ctx context.Context, addr string) ([]byte, error)
		DeleteChunksByJobID(ctx context.Context, jobID int64) error
	}
	transcriberClient interface {
		Transcribe(ctx context.Context, b []byte) (result string, err error)
	}
)

type WorkerPool struct {
	// Number of workers
	count uint8

	queue Queue

	storage chunkStorage

	chunksRepo    chunkRepository
	jobsRepo      jobRepository
	transcripRepo transcriptionsRepository

	transcriberClient transcriberClient

	logger *slog.Logger
}

func NewPool(
	count uint8,
	queue Queue,
	cs chunkStorage,
	jobRepo jobRepository,
	chunksRepo chunkRepository,
	transRepo transcriptionsRepository,
	transcriberClient transcriberClient,
	logger *slog.Logger,
) *WorkerPool {
	return &WorkerPool{
		count:             count,
		queue:             queue,
		storage:           cs,
		jobsRepo:          jobRepo,
		chunksRepo:        chunksRepo,
		transcripRepo:     transRepo,
		transcriberClient: transcriberClient,
		logger:            logger,
	}
}

// Run runs workers goroutines
func (wp *WorkerPool) Run(ctx context.Context) {
	for i := range wp.count {
		wName := "worker-" + strconv.Itoa(int(i))

		go wp.runWorker(ctx, wName)
	}
}
