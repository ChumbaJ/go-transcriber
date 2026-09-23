// Package postgres
package postgres

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TranscriptionsRepository struct {
	db *pgxpool.Pool
}

func NewTranscripRepo(db *pgxpool.Pool) *TranscriptionsRepository {
	return &TranscriptionsRepository{
		db: db,
	}
}

func (r *TranscriptionsRepository) Create(ctx context.Context, t job.Transcribtion) error {
	_, err := r.db.Exec(ctx, `INSERT INTO transcriptions (job_id, chunk_order, text) VALUES ($1, $2, $3)`, t.JobID, t.ChunkOrder, t.Text)
	if err != nil {
		return fmt.Errorf("insert into transcriptoins: %w", err)
	}

	return nil
}

func (r *TranscriptionsRepository) CountByJobID(ctx context.Context, jobID int64) (int, error) {
	var count int

	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transcriptions WHERE job_id = $1`, jobID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count transcriptions: %w", err)
	}
	return count, nil
}

func (r *TranscriptionsRepository) ListByJobID(ctx context.Context, jobID int64) ([]job.Transcribtion, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM transcriptions WHERE job_id = $1 ORDER BY chunk_order`, jobID)
	if err != nil {
		return nil, fmt.Errorf("select transcriptions: %w", err)
	}
	defer rows.Close()

	var result []job.Transcribtion

	for rows.Next() {
		var t job.Transcribtion

		if err := rows.Scan(
			&t.JobID,
			&t.ChunkOrder,
			&t.Text,
		); err != nil {
			return nil, fmt.Errorf("scan transcription: %w", err)
		}

		result = append(result, t)
	}

	return result, nil
}
