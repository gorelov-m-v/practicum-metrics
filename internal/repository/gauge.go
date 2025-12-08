package repository

import (
	"context"
	"database/sql"
	"time"
)

type GaugeRepository interface {
	Upsert(ctx context.Context, name string, value float64) error
	Get(ctx context.Context, name string) (float64, error)
	GetAll(ctx context.Context) (map[string]float64, error)
	UpsertBatch(ctx context.Context, tx *sql.Tx, gauges map[string]float64) error
}

type gaugeRepository struct {
	db *sql.DB
}

func NewGaugeRepository(db *sql.DB) GaugeRepository {
	return &gaugeRepository{db: db}
}

func (r *gaugeRepository) Upsert(ctx context.Context, name string, value float64) error {
	query := `
		INSERT INTO gauges (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, name, value, time.Now())
	return err
}

func (r *gaugeRepository) Get(ctx context.Context, name string) (float64, error) {
	var value float64
	query := `SELECT value FROM gauges WHERE name = $1`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&value)
	return value, err
}

func (r *gaugeRepository) GetAll(ctx context.Context) (map[string]float64, error) {
	result := make(map[string]float64)
	query := `SELECT name, value FROM gauges`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return result, err
		}
		result[name] = value
	}

	return result, rows.Err()
}

func (r *gaugeRepository) UpsertBatch(ctx context.Context, tx *sql.Tx, gauges map[string]float64) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO gauges (name, value, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for name, value := range gauges {
		if _, err := stmt.ExecContext(ctx, name, value, now); err != nil {
			return err
		}
	}
	return nil
}
