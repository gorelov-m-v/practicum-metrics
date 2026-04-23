package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	unsetAgentEnv()
	os.Exit(m.Run())
}

func unsetAgentEnv() {
	for _, key := range []string{
		"ADDRESS",
		"REPORT_INTERVAL",
		"POLL_INTERVAL",
		"KEY",
		"RATE_LIMIT",
		"CRYPTO_KEY",
		"GRPC_ADDRESS",
		"CONFIG",
	} {
		os.Unsetenv(key)
	}
}

func TestParseFlags_ConfigFile(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	unsetAgentEnv()

	configPath := filepath.Join(t.TempDir(), "agent.json")
	data := []byte(`{
		"address": "localhost:8181",
		"report_interval": "9s",
		"poll_interval": "4s",
		"key": "file-key",
		"rate_limit": 6,
		"crypto_key": "/tmp/public.pem",
		"grpc_address": "localhost:3200"
	}`)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	os.Args = []string{"cmd", "-config", configPath}

	parseFlags()

	if flagRunAddr != "localhost:8181" {
		t.Fatalf("expected address from config, got %q", flagRunAddr)
	}
	if flagReportInterval != 9 {
		t.Fatalf("expected report interval from config, got %d", flagReportInterval)
	}
	if flagPollInterval != 4 {
		t.Fatalf("expected poll interval from config, got %d", flagPollInterval)
	}
	if flagKey != "file-key" {
		t.Fatalf("expected key from config, got %q", flagKey)
	}
	if flagRateLimit != 6 {
		t.Fatalf("expected rate limit from config, got %d", flagRateLimit)
	}
	if flagCryptoKey != "/tmp/public.pem" {
		t.Fatalf("expected crypto key from config, got %q", flagCryptoKey)
	}
	if flagGRPCAddr != "localhost:3200" {
		t.Fatalf("expected gRPC address from config, got %q", flagGRPCAddr)
	}
}

func TestParseFlags_ConfigPriority(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	unsetAgentEnv()

	configPath := filepath.Join(t.TempDir(), "agent.json")
	data := []byte(`{
		"address": "localhost:8181",
		"report_interval": "9s",
		"poll_interval": "4s",
		"key": "file-key",
		"rate_limit": 6,
		"crypto_key": "/tmp/public.pem",
		"grpc_address": "localhost:3200"
	}`)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("KEY", "env-key")
	os.Setenv("GRPC_ADDRESS", "localhost:4200")
	defer unsetAgentEnv()

	os.Args = []string{"cmd", "-c", configPath, "-a", "localhost:9191", "-l", "8", "-g", "localhost:5200"}

	parseFlags()

	if flagRunAddr != "localhost:9191" {
		t.Fatalf("expected address from flag, got %q", flagRunAddr)
	}
	if flagReportInterval != 15 {
		t.Fatalf("expected report interval from env, got %d", flagReportInterval)
	}
	if flagPollInterval != 4 {
		t.Fatalf("expected poll interval from config, got %d", flagPollInterval)
	}
	if flagKey != "env-key" {
		t.Fatalf("expected key from env, got %q", flagKey)
	}
	if flagRateLimit != 8 {
		t.Fatalf("expected rate limit from flag, got %d", flagRateLimit)
	}
	if flagCryptoKey != "/tmp/public.pem" {
		t.Fatalf("expected crypto key from config, got %q", flagCryptoKey)
	}
	if flagGRPCAddr != "localhost:4200" {
		t.Fatalf("expected gRPC address from env, got %q", flagGRPCAddr)
	}
}
