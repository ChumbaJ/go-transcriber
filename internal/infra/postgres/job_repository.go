// Package postgres

package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepo(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{
		db: db,
	}
}
