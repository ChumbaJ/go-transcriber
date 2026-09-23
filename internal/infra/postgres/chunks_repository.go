// Package postgres
package postgres

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChunksRepository struct {
	db *pgxpool.Pool
}

func NewChunksRepository(db *pgxpool.Pool) *ChunksRepository {
	return &ChunksRepository{
		db: db,
	}
}

func (cr *ChunksRepository) CreateBatch(ctx context.Context, jobID int64, chunks []*job.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	jobIDs := make([]int64, len(chunks))
	chunk_orders := make([]int, len(chunks))
	addrs := make([]string, len(chunks))

	for i, c := range chunks {
		jobIDs[i] = jobID
		chunk_orders[i] = c.Order
		addrs[i] = c.Addr
	}

	_, err := cr.db.Exec(ctx, `
		INSERT INTO CHUNKS (job_id, chunk_order, addr) 
		SELECT * FROM UNNEST($1::bigint[], $2::int[], $3::text[])
		`, jobIDs, chunk_orders, addrs)
	if err != nil {
		return fmt.Errorf("error while inserting chunk: %w", err)
	}

	return nil
}

func (cr *ChunksRepository) CountByJobID(ctx context.Context, jobID int64) (int, error) {
	var count int

	err := cr.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM chunks
			WHERE job_id = $1
		`, jobID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count chunks: ", err)
	}

	return count, nil
}
