// Package job
package job

import (
	"errors"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

var (
	ErrNotFound = errors.New("job not found")
	JobFailed   = errors.New("job marked as failed")
)

type Chunk struct {
	Order int
	Addr  string
}

type Transcribtion struct {
	JobID      int64
	ChunkOrder int
	Text       string
}

type Job struct {
	ID          int64
	Status      JobStatus
	ResultText  string
	CreatedAt   time.Time
	CompletedAt *time.Time
}
