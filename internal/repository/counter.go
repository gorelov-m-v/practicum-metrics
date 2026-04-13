// Package repository implements data access layer for PostgreSQL storage of metrics.
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/user/practicum-metrics/internal/retry"
)

type CounterRepository interface {
	Add(ctx context.Context, name string, delta int64) error
	Set(ctx context.Context, name string, value int64) error
	Get(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]int64, error)
	AddBatch(ctx context.Context, tx *sql.Tx, counters map[string]int64) error
}

type counterRepository struct {
	db *sql.DB
}

func NewCounterRepository(db *sql.DB) CounterRepository {
	return &counterRepository{db: db}
}

func (r *counterRepository) Add(ctx context.Context, name string, delta int64) error {
	query := `
		INSERT INTO counters (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = counters.value + EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`
	return retry.WithRetry(func() error {
		_, err := r.db.ExecContext(ctx, query, name, delta, time.Now())
		return err
	})
}

func (r *counterRepository) Set(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO counters (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`
	return retry.WithRetry(func() error {
		_, err := r.db.ExecContext(ctx, query, name, value, time.Now())
		return err
	})
}

func (r *counterRepository) Get(ctx context.Context, name string) (int64, error) {
	var value int64
	query := `SELECT value FROM counters WHERE name = $1`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&value)
	return value, err
}

func (r *counterRepository) GetAll(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)
	query := `SELECT name, value FROM counters`

	err := retry.WithRetry(func() error {
		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		result = make(map[string]int64)
		for rows.Next() {
			var name string
			var value int64
			if err := rows.Scan(&name, &value); err != nil {
				return err
			}
			result[name] = value
		}

		return rows.Err()
	})

	return result, err
}

func (r *counterRepository) AddBatch(ctx context.Context, tx *sql.Tx, counters map[string]int64) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO counters (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = counters.value + EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for name, delta := range counters {
		if _, err := stmt.ExecContext(ctx, name, delta, now); err != nil {
			return err
		}
	}
	return nil
}
