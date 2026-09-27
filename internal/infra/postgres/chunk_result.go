package postgres

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
)

func (p *Postgres) SaveChunkResult(ctx context.Context, t job.Transcribtion) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// TODO:
	// прочитать как работает тут Rollback
	// доделать транзакцию

	_, err = tx.Exec(ctx, `
		INSERT INTO transcriptions (job_id, chunk_order, text)
		VALUES ($1, $2, $3)
		ON CONFLICT (job_id, chunk_order) DO NOTHING
	`, t.JobID, t.ChunkOrder, t.Text)
	if err != nil {
		return fmt.Errorf("insert transcription: %w", err)
	}

	return nil

}
