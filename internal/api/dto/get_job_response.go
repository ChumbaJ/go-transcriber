package dto

import "github.com/ChumbaJ/go-transcriber/internal/job"

type GetJobResponse struct {
	Job JobResponse `json:"job"`
}

type JobResponse struct {
	ID         int64         `json:"id"`
	Status     job.JobStatus `json:"status"`
	ResultText string        `json:"resultText"`
}
