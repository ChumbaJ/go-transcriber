// Package worker
package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

func (wp *WorkerPool) runWorker(ctx context.Context, name string) {
	for {
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
			wp.logger.Error("processChunk", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
			return
		}

		if err := wp.queue.Confirm(ctx, msg.ID); err != nil {
			wp.logger.Error("queue confirm", "error", err, "messageID", msg.ID, "jobID", msg.Item.JobID)
			return
		}
	}
}

func (wp *WorkerPool) processChunk(ctx context.Context, msg *queue.Message) error {
	chunk := msg.Item

	exists, err := wp.transcripRepo.ExistsByJobIDAndChunkOrder(ctx, chunk.JobID, chunk.ChunkOrder)
	if err != nil {
		return fmt.Errorf("check if transription exists: %w", err)
	}

	if !exists {
		if err := wp.transcribeChunk(ctx, &chunk); err != nil {
			wp.logger.Error("transcribe chunk", "error", err)
			if err := wp.markJobFailed(ctx, msg); err != nil {
				return fmt.Errorf("mark job failed: %w", err)
			}
		}
	}

	return wp.tryCompleteJob(ctx, chunk.JobID)
}

func (wp *WorkerPool) transcribeChunk(ctx context.Context, chunk *queue.Item) error {
	b, err := wp.storage.Get(ctx, chunk.Addr)
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	result, err := wp.transcriberClient.Transcribe(ctx, b)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
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
	chunk := msg.Item

	// remove chunks from storage
	n, err := wp.storage.DeleteChunksByJobID(ctx, chunk.JobID)
	if err != nil {
		return fmt.Errorf("delete chunks by jobID: %w", err)
	}
	wp.logger.Info("chunks deleted from storage", "count", n)

	// delete job
	if err := wp.jobsRepo.DeleteByID(ctx, chunk.JobID); err != nil {
		return fmt.Errorf("delete job by ID: %w", err)
	}
	// confirm jobQueue
	if err := wp.queue.Confirm(ctx, msg.ID); err != nil {
		return fmt.Errorf("queue confirm: %w", err)
	}

	return nil
}
