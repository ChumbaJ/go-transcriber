package postgres

import (
	"context"
	"fmt"
	"time"

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

	pgCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := pool.Ping(pgCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
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
