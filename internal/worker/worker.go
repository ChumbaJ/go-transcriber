// Package worker
package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/infra/transcription"
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

func (wp *WorkerPool) runWorker(ctx context.Context, name string) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := wp.queue.ClaimStale(ctx, name)
		if err != nil {
			wp.logger.Error("error while claimStale", "error", err, "worker", name)
		}

		if msg == nil {
			msg, err = wp.queue.Dequeue(ctx, name)
			if err != nil {
				wp.logger.Error("dequeue", "error", err, "worker", name)
				// Here we can do a revive operation with retry count
				return
			}
		}

		if msg == nil {
			continue
		}

		if err := wp.processChunk(ctx, msg); err != nil {
			if !errors.Is(err, job.JobFailed) {
				wp.logger.Error("processChunk", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
				return
			}
			wp.logger.Warn("job failed", "worker", name, "messageID", msg.ID, "jobID", msg.Item.JobID)
		}

		if err := wp.queue.Confirm(ctx, msg.ID); err != nil {
			wp.logger.Error("queue confirm", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
			return
		}
	}
}

func (wp *WorkerPool) processChunk(ctx context.Context, msg *queue.Message) error {
	chunk := msg.Item
	j, err := wp.jobsRepo.Get(ctx, chunk.JobID)
	if err != nil {
		return fmt.Errorf("get job status: %w", err)
	}
	if j.Status == job.JobStatusFailed {
		return job.JobFailed
	}

	exists, err := wp.transcripRepo.ExistsByJobIDAndChunkOrder(ctx, chunk.JobID, chunk.ChunkOrder)
	if err != nil {
		return fmt.Errorf("check if transription exists: %w", err)
	}

	if !exists {
		if err := wp.transcribeChunk(ctx, msg); err != nil {
			return fmt.Errorf("transcribe chunk: %w", err)
		}
	}

	return wp.tryCompleteJob(ctx, chunk.JobID)
}

func (wp *WorkerPool) transcribeChunk(ctx context.Context, msg *queue.Message) error {
	chunk := msg.Item

	b, err := wp.storage.Get(ctx, chunk.Addr)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	result, err := wp.transcribeWithRetry(ctx, b)
	if err != nil {
		if !errors.Is(err, transcription.ErrServiceUnavaliable) {
			return fmt.Errorf("transcribe: %w", err)
		}
		if markErr := wp.markJobFailed(ctx, msg); markErr != nil {
			return fmt.Errorf("transcribe failed (%v), mark failed: %w", err, markErr)
		}
		return job.JobFailed
	}

	// insert into transcriptions table
	t := job.Transcribtion{
		JobID:      chunk.JobID,
		ChunkOrder: chunk.ChunkOrder,
		Text:       result,
	}
	if err := wp.transcripRepo.Create(ctx, t); err != nil {
		return fmt.Errorf("create transcription: %w", err)
	}

	return nil
}

func (wp *WorkerPool) tryCompleteJob(ctx context.Context, jobID int64) error {
	// count transcriptions where job_id = chunk.JobID = completedChunks
	completedChunks, err := wp.transcripRepo.CountByJobID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("count transcriptions by job id: %w", err)
	}

	// count chunks where job_id = chunk.JobID = totalChunks
	totalChunks, err := wp.chunksRepo.CountByJobID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("count chunks by id: %w", err)
	}

	if completedChunks == totalChunks {
		if err := wp.jobsRepo.UpdateStatus(ctx, jobID, job.JobStatusCompleted); err != nil {
			return fmt.Errorf("update status job: %w", err)
		}

		transcriptions, err := wp.transcripRepo.ListByJobID(ctx, jobID)
		if err != nil {
			return fmt.Errorf("list transcriptions by jobid: %w", err)
		}

		parts := make([]string, 0, len(transcriptions))

		for _, t := range transcriptions {
			parts = append(parts, t.Text)
		}

		result := strings.Join(parts, "")

		if err := wp.jobsRepo.SetResult(ctx, jobID, result); err != nil {
			return fmt.Errorf("set job result text: %w", err)
		}
	}
	return nil
}

func (wp *WorkerPool) markJobFailed(ctx context.Context, msg *queue.Message) error {
	if err := wp.jobsRepo.MarkFailed(ctx, msg.Item.JobID); err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	return nil
}

const (
	transcribeMaxAttempts = 3 // Initial request and two retries.
	transcribeRetryDelay  = time.Second
)

func (wp *WorkerPool) transcribeWithRetry(ctx context.Context, b []byte) (string, error) {
	for attempt := 1; attempt <= transcribeMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		result, err := wp.transcriberClient.Transcribe(ctx, b)
		if err == nil {
			return result, nil
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if !errors.Is(err, transcription.ErrServiceUnavaliable) || attempt == transcribeMaxAttempts {
			return "", fmt.Errorf("attempt %d: %w", attempt, err)
		}
		timer := time.NewTimer(transcribeRetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	return "", fmt.Errorf("transcription attempts exhausted: %w", transcription.ErrServiceUnavaliable)
}
