package repository

import (
	"context"
	"database/sql"
	"time"
)

type CounterRepository interface {
	Add(ctx context.Context, name string, delta int64) error
	Set(ctx context.Context, name string, value int64) error
	Get(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]int64, error)
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
	_, err := r.db.ExecContext(ctx, query, name, delta, time.Now())
	return err
}

func (r *counterRepository) Set(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO counters (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, name, value, time.Now())
	return err
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
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return result, err
		}
		result[name] = value
	}

	return result, rows.Err()
}
