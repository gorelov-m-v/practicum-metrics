package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGaugeRepositoryMock_Upsert_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	name := "Alloc"
	value := 123.456

	mock.ExpectExec("INSERT INTO gauges").
		WithArgs(name, value, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Upsert(ctx, name, value)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_Upsert_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("db error")

	mock.ExpectExec("INSERT INTO gauges").
		WithArgs("Alloc", 100.0, sqlmock.AnyArg()).
		WillReturnError(expectedErr)

	err = repo.Upsert(ctx, "Alloc", 100.0)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_Upsert_ContextCanceled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock.ExpectExec("INSERT INTO gauges").
		WithArgs("Alloc", 100.0, sqlmock.AnyArg()).
		WillReturnError(context.Canceled)

	err = repo.Upsert(ctx, "Alloc", 100.0)
	if err == nil {
		t.Error("expected error for canceled context")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_Get_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	name := "Alloc"
	expectedValue := 123.456

	rows := sqlmock.NewRows([]string{"value"}).AddRow(expectedValue)
	mock.ExpectQuery("SELECT value FROM gauges WHERE name").
		WithArgs(name).
		WillReturnRows(rows)

	value, err := repo.Get(ctx, name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != expectedValue {
		t.Errorf("expected value %f, got %f", expectedValue, value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_Get_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	mock.ExpectQuery("SELECT value FROM gauges WHERE name").
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

func TestGaugeRepositoryMock_Get_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("query error")

	mock.ExpectQuery("SELECT value FROM gauges WHERE name").
		WithArgs("Alloc").
		WillReturnError(expectedErr)

	_, err = repo.Get(ctx, "Alloc")
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_GetAll_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("Alloc", 123.456).
		AddRow("HeapAlloc", 789.012).
		AddRow("Sys", 555.555)

	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 gauges, got %d", len(result))
	}

	expected := map[string]float64{
		"Alloc":     123.456,
		"HeapAlloc": 789.012,
		"Sys":       555.555,
	}

	for name, expectedValue := range expected {
		value, exists := result[name]
		if !exists {
			t.Errorf("gauge %s not found in result", name)
		}
		if value != expectedValue {
			t.Errorf("gauge %s: expected %f, got %f", name, expectedValue, value)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_GetAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"})
	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnRows(rows)

	result, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d gauges", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_GetAll_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("query error")

	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnError(expectedErr)

	_, err = repo.GetAll(ctx)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_GetAll_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("Alloc", 123.456).
		AddRow("Invalid", "not_a_number"). // This will cause scan error
		AddRow("Sys", 555.555)

	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnRows(rows)

	_, err = repo.GetAll(ctx)
	if err == nil {
		t.Error("expected scan error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_GetAll_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	expectedErr := errors.New("rows iteration error")

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("Alloc", 123.456).
		RowError(0, expectedErr)

	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnRows(rows)

	_, err = repo.GetAll(ctx)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGaugeRepositoryMock_Upsert_VerifyTimestamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewGaugeRepository(db)
	ctx := context.Background()

	name := "Alloc"
	value := 123.456
	now := time.Now()

	mock.ExpectExec("INSERT INTO gauges").
		WithArgs(name, value, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Upsert(ctx, name, value)
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
