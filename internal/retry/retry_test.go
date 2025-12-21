package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriableError_NilError(t *testing.T) {
	if IsRetriableError(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsRetriableError_PostgreSQLConnectionException(t *testing.T) {
	tests := []struct {
		name     string
		pgError  *pgconn.PgError
		expected bool
	}{
		{
			name:     "connection exception",
			pgError:  &pgconn.PgError{Code: pgerrcode.ConnectionException},
			expected: true,
		},
		{
			name:     "connection does not exist",
			pgError:  &pgconn.PgError{Code: pgerrcode.ConnectionDoesNotExist},
			expected: true,
		},
		{
			name:     "connection failure",
			pgError:  &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			expected: true,
		},
		{
			name:     "unique violation - not retriable",
			pgError:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			expected: false,
		},
		{
			name:     "syntax error - not retriable",
			pgError:  &pgconn.PgError{Code: pgerrcode.SyntaxError},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetriableError(tt.pgError)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for error code %s", tt.expected, result, tt.pgError.Code)
			}
		})
	}
}

func TestIsRetriableError_NonPostgreSQLError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "generic error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "wrapped generic error",
			err:  errors.New("wrapped: some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetriableError(tt.err); got != tt.want {
				t.Errorf("IsRetriableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithRetry_SuccessFirstAttempt(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return nil
	}

	err := WithRetry(operation)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return &pgconn.PgError{Code: pgerrcode.ConnectionException}
		}
		return nil
	}

	start := time.Now()
	err := WithRetry(operation)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("expected no error after retries, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	expectedDuration := 1 * time.Second
	if elapsed < expectedDuration || elapsed > expectedDuration+500*time.Millisecond {
		t.Logf("warning: expected duration around %v, got %v", expectedDuration, elapsed)
	}
}

func TestWithRetry_AllRetriesFailed(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return &pgconn.PgError{Code: pgerrcode.ConnectionException}
	}

	start := time.Now()
	err := WithRetry(operation)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error after all retries failed")
	}

	expectedAttempts := 4
	if attempts != expectedAttempts {
		t.Errorf("expected %d attempts, got %d", expectedAttempts, attempts)
	}

	expectedDuration := 4 * time.Second
	if elapsed < expectedDuration || elapsed > expectedDuration+500*time.Millisecond {
		t.Logf("warning: expected duration around %v, got %v", expectedDuration, elapsed)
	}
}

func TestWithRetry_NonRetriableError(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	}

	err := WithRetry(operation)

	if err == nil {
		t.Error("expected error for non-retriable error")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for non-retriable error, got %d", attempts)
	}
}

func TestWithRetry_NonRetriableErrorAfterRetriable(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		if attempts == 1 {
			return &pgconn.PgError{Code: pgerrcode.ConnectionException}
		}
		return &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	}

	err := WithRetry(operation)

	if err == nil {
		t.Error("expected error")
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts (retriable then non-retriable), got %d", attempts)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Error("expected pgconn.PgError")
	} else if pgErr.Code != pgerrcode.UniqueViolation {
		t.Errorf("expected UniqueViolation error, got %s", pgErr.Code)
	}
}

func TestWithRetry_GenericError(t *testing.T) {
	attempts := 0
	expectedErr := errors.New("generic error")
	operation := func() error {
		attempts++
		return expectedErr
	}

	err := WithRetry(operation)

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt for generic error, got %d", attempts)
	}
}

func TestWithRetry_MaxRetries(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
	}

	WithRetry(operation)

	if attempts != 4 {
		t.Errorf("expected 4 total attempts (1 initial + 3 retries), got %d", attempts)
	}
}

func TestWithRetry_RetryIntervals(t *testing.T) {
	attempts := 0
	timestamps := make([]time.Time, 0, 4)

	operation := func() error {
		attempts++
		timestamps = append(timestamps, time.Now())
		return &pgconn.PgError{Code: pgerrcode.ConnectionException}
	}

	WithRetry(operation)

	if len(timestamps) != 4 {
		t.Fatalf("expected 4 timestamps, got %d", len(timestamps))
	}

	tolerance := 200 * time.Millisecond

	interval1 := timestamps[1].Sub(timestamps[0])
	if interval1 > tolerance {
		t.Errorf("expected first interval ~0s, got %v", interval1)
	}

	interval2 := timestamps[2].Sub(timestamps[1])
	if interval2 < time.Second-tolerance || interval2 > time.Second+tolerance {
		t.Errorf("expected second interval ~1s, got %v", interval2)
	}

	interval3 := timestamps[3].Sub(timestamps[2])
	if interval3 < 3*time.Second-tolerance || interval3 > 3*time.Second+tolerance {
		t.Errorf("expected third interval ~3s, got %v", interval3)
	}
}
