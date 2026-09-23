// Package job
package job

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
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
	ID         int64
	Status     JobStatus
	ResultText string
}
