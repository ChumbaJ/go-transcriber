// Package postgres
package postgres

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepo(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) Create(ctx context.Context) (*job.Job, error) {
	j := &job.Job{}

	err := r.db.QueryRow(ctx, `INSERT INTO jobs (status) VALUES($1) RETURNING id, status, result_text`, job.JobStatusPending).Scan(&j.ID, &j.Status, &j.ResultText)
	if err != nil {
		return nil, fmt.Errorf("inserting job into database: %w", err)
	}

	return j, nil
}

func (r *JobRepository) UpdateStatus(ctx context.Context, jobID int64, status job.JobStatus) error {
	_, err := r.db.Exec(ctx, `
			UPDATE jobs 
			SET status = $1
			WHERE id = $2
		`, status, jobID)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}

	return nil
}

func (r *JobRepository) SetResult(ctx context.Context, jobID int64, result string) error {
	_, err := r.db.Exec(ctx, `
			UPDATE jobs 
			SET result_text = $1
			WHERE id = $2
		`, result, jobID)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}

	return nil
}

func (r *JobRepository) Get(ctx context.Context, jobID int64) (*job.Job, error) {
	var j job.Job
	err := r.db.QueryRow(ctx, `SELECT id, status, result_text FROM jobs WHERE id = $1`, jobID).Scan(&j.ID, &j.Status, &j.ResultText)
	if err != nil {
		return nil, fmt.Errorf("error querying job by id: %w", err)
	}
	return &j, nil
}
