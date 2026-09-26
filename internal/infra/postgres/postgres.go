package postgres

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool

	Jobs        *JobRepository
	Chunks      *ChunksRepository
	Transcripts *TranscriptionsRepository
}

func New(ctx context.Context, cfg *config.Config) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return &Postgres{
		pool:        pool,
		Jobs:        NewJobRepo(pool),
		Chunks:      NewChunksRepository(pool),
		Transcripts: NewTranscripRepo(pool),
	}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}
