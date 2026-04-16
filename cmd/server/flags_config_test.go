package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	unsetServerEnv()
	os.Exit(m.Run())
}

func unsetServerEnv() {
	for _, key := range []string{
		"ADDRESS",
		"STORE_INTERVAL",
		"STORE_FILE",
		"FILE_STORAGE_PATH",
		"RESTORE",
		"DATABASE_DSN",
		"KEY",
		"AUDIT_FILE",
		"AUDIT_URL",
		"CRYPTO_KEY",
		"CONFIG",
	} {
		os.Unsetenv(key)
	}
}

func TestParseFlags_ConfigFile(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	unsetServerEnv()

	configPath := filepath.Join(t.TempDir(), "server.json")
	data := []byte(`{
		"address": "localhost:8282",
		"restore": false,
		"store_interval": "11s",
		"store_file": "/tmp/server.db",
		"database_dsn": "postgres://localhost/db",
		"key": "file-key",
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://localhost/audit",
		"crypto_key": "/tmp/private.pem"
	}`)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	os.Args = []string{"cmd", "-config", configPath}

	parseFlags()

	if flagRunAddr != "localhost:8282" {
		t.Fatalf("expected address from config, got %q", flagRunAddr)
	}
	if flagRestore {
		t.Fatalf("expected restore from config to be false")
	}
	if flagStoreInterval != 11 {
		t.Fatalf("expected store interval from config, got %d", flagStoreInterval)
	}
	if flagFileStoragePath != "/tmp/server.db" {
		t.Fatalf("expected store file from config, got %q", flagFileStoragePath)
	}
	if flagDatabaseDSN != "postgres://localhost/db" {
		t.Fatalf("expected database dsn from config, got %q", flagDatabaseDSN)
	}
	if flagKey != "file-key" {
		t.Fatalf("expected key from config, got %q", flagKey)
	}
	if flagAuditFile != "/tmp/audit.log" {
		t.Fatalf("expected audit file from config, got %q", flagAuditFile)
	}
	if flagAuditURL != "http://localhost/audit" {
		t.Fatalf("expected audit url from config, got %q", flagAuditURL)
	}
	if flagCryptoKey != "/tmp/private.pem" {
		t.Fatalf("expected crypto key from config, got %q", flagCryptoKey)
	}
}

func TestParseFlags_ConfigPriority(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	unsetServerEnv()

	configPath := filepath.Join(t.TempDir(), "server.json")
	data := []byte(`{
		"address": "localhost:8282",
		"restore": false,
		"store_interval": "11s",
		"store_file": "/tmp/server.db",
		"database_dsn": "postgres://localhost/db",
		"key": "file-key",
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://localhost/audit",
		"crypto_key": "/tmp/private.pem"
	}`)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	os.Setenv("RESTORE", "true")
	os.Setenv("STORE_FILE", "/tmp/from-env.db")
	defer unsetServerEnv()

	os.Args = []string{"cmd", "-c", configPath, "-a", "localhost:9393", "-i", "7", "-d", "postgres://flag/db"}

	parseFlags()

	if flagRunAddr != "localhost:9393" {
		t.Fatalf("expected address from flag, got %q", flagRunAddr)
	}
	if !flagRestore {
		t.Fatalf("expected restore from env to be true")
	}
	if flagStoreInterval != 7 {
		t.Fatalf("expected store interval from flag, got %d", flagStoreInterval)
	}
	if flagFileStoragePath != "/tmp/from-env.db" {
		t.Fatalf("expected store file from env, got %q", flagFileStoragePath)
	}
	if flagDatabaseDSN != "postgres://flag/db" {
		t.Fatalf("expected database dsn from flag, got %q", flagDatabaseDSN)
	}
	if flagKey != "file-key" {
		t.Fatalf("expected key from config, got %q", flagKey)
	}
	if flagAuditFile != "/tmp/audit.log" {
		t.Fatalf("expected audit file from config, got %q", flagAuditFile)
	}
	if flagAuditURL != "http://localhost/audit" {
		t.Fatalf("expected audit url from config, got %q", flagAuditURL)
	}
	if flagCryptoKey != "/tmp/private.pem" {
		t.Fatalf("expected crypto key from config, got %q", flagCryptoKey)
	}
}

func TestParseFlags_FileStoragePathEnvPriority(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	unsetServerEnv()

	os.Setenv("STORE_FILE", "/tmp/store-file.db")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/file-storage-path.db")
	defer unsetServerEnv()

	os.Args = []string{"cmd"}

	parseFlags()

	if flagFileStoragePath != "/tmp/file-storage-path.db" {
		t.Fatalf("expected FILE_STORAGE_PATH to have priority, got %q", flagFileStoragePath)
	}
}
