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

func (r *JobRepository) Get(ctx context.Context, jobID string) *job.Job {
	return nil
}
