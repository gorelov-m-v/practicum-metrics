package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_CryptoKey(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("KEY")
	os.Unsetenv("RATE_LIMIT")
	os.Unsetenv("CRYPTO_KEY")

	os.Args = []string{"cmd", "-crypto-key", "/tmp/public.pem"}

	parseFlags()

	if flagCryptoKey != "/tmp/public.pem" {
		t.Fatalf("expected crypto key path from flag, got %q", flagCryptoKey)
	}
}

func TestParseFlags_CryptoKeyFromEnv(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("KEY")
	os.Unsetenv("RATE_LIMIT")
	os.Setenv("CRYPTO_KEY", "/tmp/env-public.pem")
	defer os.Unsetenv("CRYPTO_KEY")

	os.Args = []string{"cmd", "-crypto-key", "/tmp/flag-public.pem"}

	parseFlags()

	if flagCryptoKey != "/tmp/env-public.pem" {
		t.Fatalf("expected crypto key path from env, got %q", flagCryptoKey)
	}
}
