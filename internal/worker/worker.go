// Package worker
package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/job"
)

func (wp *WorkerPool) runWorker(ctx context.Context, name string) {
	for {
		msg, err := wp.queue.Dequeue(ctx, name)
		if err != nil {
			fmt.Println("dequeue: %w", err.Error())
			return
		}

		wp.processChunk(ctx, &msg.Item)

		// XAck
		if err := wp.queue.Ack(ctx, msg.ID); err != nil {
			fmt.Println("queue ack: ", err.Error())
			return
		}
	}
}

func (wp *WorkerPool) processChunk(ctx context.Context, chunk *job.QueueItem) {
	_, err := wp.storage.Get(ctx, chunk.Addr)
	if err != nil {
		fmt.Println("get storage: ", err.Error())
		return
	}

	// send to LLM
	time.Sleep(time.Second * 15)

	// save response to databse and check at the same time

	// insert into transcriptions table
	t := job.Transcribtion{
		JobID:      chunk.JobID,
		ChunkOrder: chunk.ChunkOrder,
		Text:       "hello world",
	}
	if err := wp.transcripRepo.Create(ctx, t); err != nil {
		fmt.Println("create transcription: ", err.Error())
		return
	}

	// count transcriptions where job_id = chunk.JobID = completedChunks
	completedChunks, err := wp.transcripRepo.CountByJobID(ctx, chunk.JobID)
	if err != nil {
		fmt.Println("count transcriptions by job id: ", err.Error())
		return
	}

	// count chunks where job_id = chunk.JobID = totalChunks
	totalChunks, err := wp.chunksRepo.CountByJobID(ctx, chunk.JobID)
	if err != nil {
		fmt.Println("count chunks by id: ", err.Error())
		return

	}

	if completedChunks == totalChunks {
		if err := wp.jobsRepo.UpdateStatus(ctx, chunk.JobID, job.JobStatusCompleted); err != nil {
			fmt.Println("update status job: ", err.Error())
			return
		}

		transcriptions, err := wp.transcripRepo.ListByJobID(ctx, chunk.JobID)
		if err != nil {
			fmt.Println("list transcriptions by jobid: ", err.Error())
			return
		}

		parts := make([]string, 0, len(transcriptions))

		for _, t := range transcriptions {
			parts = append(parts, t.Text)
		}

		result := strings.Join(parts, "")

		if err := wp.jobsRepo.SetResult(ctx, chunk.JobID, result); err != nil {
			fmt.Println("set job result text: ", err.Error())
			return
		}
	}
}
