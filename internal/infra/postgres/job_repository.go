// Package postgres
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/jackc/pgx/v5"
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

func (r *JobRepository) Create(ctx context.Context, startedAt time.Time) (*job.Job, error) {
	j := &job.Job{}

	err := r.db.QueryRow(ctx, `
		INSERT INTO jobs (status, created_at)
		VALUES ($1, $2)
		RETURNING id, status, result_text, created_at
	`, job.JobStatusPending, startedAt).Scan(&j.ID, &j.Status, &j.ResultText, &j.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting job into database: %w", err)
	}

	return j, nil
}

func (r *JobRepository) Complete(ctx context.Context, jobID int64, result string) (time.Duration, bool, error) {
	var createdAt, completedAt time.Time
	err := r.db.QueryRow(ctx, `
		UPDATE jobs
		SET result_text = $1, status = $2, completed_at = clock_timestamp()
		WHERE id = $3 AND status = $4
		RETURNING created_at, completed_at
	`, result, job.JobStatusCompleted, jobID, job.JobStatusPending).Scan(&createdAt, &completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("complete job: %w", err)
	}
	return completedAt.Sub(createdAt), true, nil
}

func (r *JobRepository) Get(ctx context.Context, jobID int64) (*job.Job, error) {
	var j job.Job
	err := r.db.QueryRow(ctx, `
		SELECT id, status, result_text, created_at, completed_at
		FROM jobs WHERE id = $1
	`, jobID).Scan(&j.ID, &j.Status, &j.ResultText, &j.CreatedAt, &j.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, job.ErrNotFound
		}
		return nil, fmt.Errorf("error querying job by id: %w", err)
	}
	return &j, nil
}

func (r *JobRepository) MarkFailed(ctx context.Context, jobID int64) error {
	_, err := r.db.Exec(ctx,
		`
		UPDATE jobs
		SET status = $2
		WHERE id = $1
		`, jobID, job.JobStatusFailed)
	if err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}

	return nil
}
