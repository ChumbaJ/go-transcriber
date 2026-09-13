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

type Transcribtions struct {
	JobID      int
	ChunkOrder int
	Text       string
}

type Job struct {
	ID         int
	Status     JobStatus
	ResultText string
}
