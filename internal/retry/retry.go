package retry

import (
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	MaxRetries = 3
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
	err := operation()
	if err == nil || !IsRetriableError(err) {
		return err
	}

	return retry.Do(
		operation,
		retry.Attempts(uint(MaxRetries)),
		retry.DelayType(delayFunc),
		retry.RetryIf(func(err error) bool {
			return IsRetriableError(err)
		}),
		retry.LastErrorOnly(true),
	)
}

func Do(operation func() error) error {
	return retry.Do(
		operation,
		retry.Attempts(uint(MaxRetries)),
		retry.DelayType(delayFunc),
		retry.LastErrorOnly(true),
	)
}

func delayFunc(n uint, err error, config *retry.Config) time.Duration {
	if n > 0 && int(n-1) < len(retryIntervals) {
		return retryIntervals[n-1]
	}
	return 1 * time.Second
}
