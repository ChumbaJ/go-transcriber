// Package worker
package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
			continue
		}

		if msg == nil {
			msg, err = wp.queue.Dequeue(ctx, name)
			if err != nil {
				wp.logger.Error("dequeue", "error", err, "worker", name)
				continue
			}
		}

		if msg == nil {
			continue
		}

		if err := wp.processChunk(ctx, msg); err != nil {
			if !errors.Is(err, job.JobFailed) {
				wp.logger.Error("processChunk", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
				if markErr := wp.markJobFailed(ctx, msg); markErr != nil {
					wp.logger.Error("mark job failed", "error", markErr, "messageID", msg.ID, "jobID", msg.Item.JobID)
					continue
				}
			} else {
				wp.logger.Warn("job failed", "error", err, "worker", name, "messageID", msg.ID, "jobID", msg.Item.JobID)
			}
		}

		if err := wp.queue.Confirm(ctx, msg.ID); err != nil {
			wp.logger.Error("queue confirm", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
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

	result, err := wp.transcriberClient.Transcribe(ctx, b, chunk.Format)
	if err != nil {
		if markErr := wp.markJobFailed(ctx, msg); markErr != nil {
			return fmt.Errorf("transcribe failed (%v), mark failed: %w", err, markErr)
		}
		return fmt.Errorf("%w: transcription: %v", job.JobFailed, err)
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
		transcriptions, err := wp.transcripRepo.ListByJobID(ctx, jobID)
		if err != nil {
			return fmt.Errorf("list transcriptions by jobid: %w", err)
		}

		parts := make([]string, 0, len(transcriptions))

		for _, t := range transcriptions {
			parts = append(parts, t.Text)
		}

		result := strings.Join(parts, "")

		duration, completed, err := wp.jobsRepo.Complete(ctx, jobID, result)
		if err != nil {
			return fmt.Errorf("complete job: %w", err)
		}
		if completed {
			fmt.Printf("job %d completed in %s\n", jobID, duration)
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
