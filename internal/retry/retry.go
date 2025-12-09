package retry

import (
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	maxRetries = 3
)

var retryIntervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}

	return false
}

func WithRetry(operation func() error) error {
	var lastErr error

	lastErr = operation()
	if lastErr == nil {
		return nil
	}

	if !IsRetriableError(lastErr) {
		return lastErr
	}

	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryIntervals[i])

		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if !IsRetriableError(lastErr) {
			return lastErr
		}
	}

	return lastErr
}
