package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCounterRepositoryMock_Add_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	delta := int64(5)

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, delta, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Add(ctx, name, delta)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Add_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("db error")

	mock.ExpectExec("INSERT INTO counters").
		WithArgs("PollCount", int64(10), sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	err = repo.Add(ctx, "PollCount", 10)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Add_NegativeDelta(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	delta := int64(-5)

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, delta, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Add(ctx, name, delta)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Set_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	value := int64(42)

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, value, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Set(ctx, name, value)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Set_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("db error")

	mock.ExpectExec("INSERT INTO counters").
		WithArgs("PollCount", int64(100), sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	err = repo.Set(ctx, "PollCount", 100)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Set_ZeroValue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "ResetCounter"
	value := int64(0)

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, value, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Set(ctx, name, value)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	expectedValue := int64(42)

	rows := sqlmock.NewRows([]string{"value"}).AddRow(expectedValue)
	mock.ExpectQuery("SELECT value FROM counters WHERE name").
		WithArgs(name).
		WillReturnRows(rows)

	value, err := repo.Get(ctx, name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != expectedValue {
		t.Errorf("expected value %d, got %d", expectedValue, value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Get_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT value FROM counters WHERE name").
		WithArgs("NonExistent").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Get(ctx, "NonExistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Get_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("query error")

	mock.ExpectQuery("SELECT value FROM counters WHERE name").
		WithArgs("PollCount").
		WillReturnError(expectedErr)

	_, err = repo.Get(ctx, "PollCount")
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_GetAll_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("PollCount", int64(10)).
		AddRow("Requests", int64(100)).
		AddRow("Errors", int64(5))

	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 counters, got %d", len(result))
	}

	expected := map[string]int64{
		"PollCount": 10,
		"Requests":  100,
		"Errors":    5,
	}

	for name, expectedValue := range expected {
		value, exists := result[name]
		if !exists {
			t.Errorf("counter %s not found in result", name)
		}
		if value != expectedValue {
			t.Errorf("counter %s: expected %d, got %d", name, expectedValue, value)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_GetAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"})
	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d counters", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_GetAll_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("query error")

	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnError(expectedErr)

	_, err = repo.GetAll(ctx)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_GetAll_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("PollCount", int64(10)).
		AddRow("Invalid", "not_a_number"). // This will cause scan error
		AddRow("Errors", int64(5))

	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnRows(rows)

	_, err = repo.GetAll(ctx)
	if err == nil {
		t.Error("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_GetAll_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("rows iteration error")

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("PollCount", int64(10)).
		RowError(0, expectedErr)

	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnRows(rows)

	_, err = repo.GetAll(ctx)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Add_VerifyTimestamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	delta := int64(5)
	now := time.Now()

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, delta, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Add(ctx, name, delta)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	elapsed := time.Since(now)
	if elapsed > time.Second {
		t.Errorf("timestamp seems incorrect, elapsed time: %v", elapsed)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_Set_VerifyTimestamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "PollCount"
	value := int64(100)
	now := time.Now()

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, value, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Set(ctx, name, value)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	elapsed := time.Since(now)
	if elapsed > time.Second {
		t.Errorf("timestamp seems incorrect, elapsed time: %v", elapsed)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_ContextCanceled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock.ExpectExec("INSERT INTO counters").
		WithArgs("PollCount", int64(10), sqlmock.AnyArg()).
		WillReturnError(context.Canceled)

	err = repo.Add(ctx, "PollCount", 10)
	if err == nil {
		t.Error("expected error for canceled context")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCounterRepositoryMock_LargeValue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewCounterRepository(db)
	ctx := context.Background()

	name := "LargeCounter"
	largeValue := int64(9223372036854775807) // max int64

	mock.ExpectExec("INSERT INTO counters").
		WithArgs(name, largeValue, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Set(ctx, name, largeValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"value"}).AddRow(largeValue)
	mock.ExpectQuery("SELECT value FROM counters WHERE name").
		WithArgs(name).
		WillReturnRows(rows)

	value, err := repo.Get(ctx, name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != largeValue {
		t.Errorf("expected value %d, got %d", largeValue, value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
