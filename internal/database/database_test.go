package database

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDatabase(t *testing.T) string {
	t.Helper()
	return "postgres://postgres:postgres@localhost:5432/metrics_test?sslmode=disable"
}

func TestNewDB(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("NewDB returned nil")
	}

	if db.conn == nil {
		t.Fatal("connection is nil")
	}
}

func TestNewDB_InvalidDSN(t *testing.T) {
	_, err := NewDB("invalid_dsn")
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}

func TestNewDB_UnreachableHost(t *testing.T) {
	dsn := "postgres://postgres:postgres@nonexistent:5432/metrics?sslmode=disable"
	_, err := NewDB(dsn)
	if err == nil {
		t.Error("expected error for unreachable host")
	}
}

func TestDB_Ping(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestDB_PingContextCanceled(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = db.Ping(ctx)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestDB_Close(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = db.Ping(ctx)
	if err == nil {
		t.Error("expected error after closing connection")
	}
}

func TestDB_GetConn(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	conn := db.GetConn()
	if conn == nil {
		t.Fatal("GetConn returned nil")
	}

	err = conn.Ping()
	if err != nil {
		t.Errorf("Ping via GetConn failed: %v", err)
	}
}

func TestDB_RunMigrations(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	_, _ = db.conn.Exec("DROP TABLE IF EXISTS gauges CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS counters CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE")

	err = db.RunMigrations("migrations")
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	tables := []string{"gauges", "counters"}
	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = $1
			)
		`
		err := db.conn.QueryRow(query, table).Scan(&exists)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s was not created", table)
		}
	}

	err = db.RunMigrations("migrations")
	if err != nil {
		t.Errorf("Running migrations second time failed: %v", err)
	}
}

func TestDB_RunMigrations_InvalidPath(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	err = db.RunMigrations("nonexistent_path")
	if err == nil {
		t.Error("expected error for invalid migrations path")
	}
}

func TestDB_RunMigrations_VerifySchema(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	_, _ = db.conn.Exec("DROP TABLE IF EXISTS gauges CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS counters CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE")

	err = db.RunMigrations("migrations")
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	var dataType string
	query := `
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = 'gauges' AND column_name = 'value'
	`
	err = db.conn.QueryRow(query).Scan(&dataType)
	if err != nil {
		t.Errorf("Failed to get gauges.value column type: %v", err)
	}
	if dataType != "double precision" {
		t.Errorf("Expected gauges.value to be 'double precision', got '%s'", dataType)
	}

	query = `
		SELECT data_type
		FROM information_schema.columns
		WHERE table_name = 'counters' AND column_name = 'value'
	`
	err = db.conn.QueryRow(query).Scan(&dataType)
	if err != nil {
		t.Errorf("Failed to get counters.value column type: %v", err)
	}
	if dataType != "bigint" {
		t.Errorf("Expected counters.value to be 'bigint', got '%s'", dataType)
	}
}

func TestDB_InsertAndQuery(t *testing.T) {
	dsn := setupTestDatabase(t)

	db, err := NewDB(dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing:", err)
	}
	defer db.Close()

	_, _ = db.conn.Exec("DROP TABLE IF EXISTS gauges CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS counters CASCADE")
	_, _ = db.conn.Exec("DROP TABLE IF EXISTS schema_migrations CASCADE")

	err = db.RunMigrations("migrations")
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	_, err = db.conn.Exec(
		"INSERT INTO gauges (name, value, updated_at) VALUES ($1, $2, $3)",
		"test_gauge", 123.456, time.Now(),
	)
	if err != nil {
		t.Errorf("Failed to insert gauge: %v", err)
	}

	var value float64
	err = db.conn.QueryRow("SELECT value FROM gauges WHERE name = $1", "test_gauge").Scan(&value)
	if err != nil {
		t.Errorf("Failed to query gauge: %v", err)
	}
	if value != 123.456 {
		t.Errorf("Expected gauge value 123.456, got %f", value)
	}

	_, err = db.conn.Exec(
		"INSERT INTO counters (name, value, updated_at) VALUES ($1, $2, $3)",
		"test_counter", int64(42), time.Now(),
	)
	if err != nil {
		t.Errorf("Failed to insert counter: %v", err)
	}

	var counter int64
	err = db.conn.QueryRow("SELECT value FROM counters WHERE name = $1", "test_counter").Scan(&counter)
	if err != nil {
		t.Errorf("Failed to query counter: %v", err)
	}
	if counter != 42 {
		t.Errorf("Expected counter value 42, got %d", counter)
	}
}
