package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_CryptoKey(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("KEY")
	os.Unsetenv("AUDIT_FILE")
	os.Unsetenv("AUDIT_URL")
	os.Unsetenv("CRYPTO_KEY")

	os.Args = []string{"cmd", "-crypto-key", "/tmp/private.pem"}

	parseFlags()

	if flagCryptoKey != "/tmp/private.pem" {
		t.Fatalf("expected crypto key path from flag, got %q", flagCryptoKey)
	}
}

func TestParseFlags_CryptoKeyFromEnv(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("KEY")
	os.Unsetenv("AUDIT_FILE")
	os.Unsetenv("AUDIT_URL")
	os.Setenv("CRYPTO_KEY", "/tmp/env-private.pem")
	defer os.Unsetenv("CRYPTO_KEY")

	os.Args = []string{"cmd", "-crypto-key", "/tmp/flag-private.pem"}

	parseFlags()

	if flagCryptoKey != "/tmp/env-private.pem" {
		t.Fatalf("expected crypto key path from env, got %q", flagCryptoKey)
	}
}
