package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServerFileConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.json")
	data := []byte(`{
		"address": "localhost:8080",
		"restore": false,
		"store_interval": "15s",
		"store_file": "/tmp/metrics.db",
		"database_dsn": "postgres://localhost/db",
		"key": "secret",
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://localhost/audit",
		"crypto_key": "/tmp/private.pem",
		"trusted_subnet": "192.168.1.0/24",
		"grpc_address": "localhost:3200"
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadServerFileConfig(path)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.Address == nil || *cfg.Address != "localhost:8080" {
		t.Fatalf("unexpected address: %#v", cfg.Address)
	}

	if cfg.Restore == nil || *cfg.Restore {
		t.Fatalf("unexpected restore: %#v", cfg.Restore)
	}

	if cfg.StoreInterval == nil || *cfg.StoreInterval != 15 {
		t.Fatalf("unexpected store interval: %#v", cfg.StoreInterval)
	}

	if cfg.StoreFile == nil || *cfg.StoreFile != "/tmp/metrics.db" {
		t.Fatalf("unexpected store file: %#v", cfg.StoreFile)
	}

	if cfg.DatabaseDSN == nil || *cfg.DatabaseDSN != "postgres://localhost/db" {
		t.Fatalf("unexpected database dsn: %#v", cfg.DatabaseDSN)
	}

	if cfg.Key == nil || *cfg.Key != "secret" {
		t.Fatalf("unexpected key: %#v", cfg.Key)
	}

	if cfg.AuditFile == nil || *cfg.AuditFile != "/tmp/audit.log" {
		t.Fatalf("unexpected audit file: %#v", cfg.AuditFile)
	}

	if cfg.AuditURL == nil || *cfg.AuditURL != "http://localhost/audit" {
		t.Fatalf("unexpected audit url: %#v", cfg.AuditURL)
	}

	if cfg.CryptoKey == nil || *cfg.CryptoKey != "/tmp/private.pem" {
		t.Fatalf("unexpected crypto key: %#v", cfg.CryptoKey)
	}

	if cfg.TrustedSubnet == nil || *cfg.TrustedSubnet != "192.168.1.0/24" {
		t.Fatalf("unexpected trusted subnet: %#v", cfg.TrustedSubnet)
	}

	if cfg.GRPCAddress == nil || *cfg.GRPCAddress != "localhost:3200" {
		t.Fatalf("unexpected gRPC address: %#v", cfg.GRPCAddress)
	}
}

func TestLoadAgentFileConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	data := []byte(`{
		"address": "localhost:8080",
		"report_interval": "12s",
		"poll_interval": "3s",
		"key": "secret",
		"rate_limit": 7,
		"crypto_key": "/tmp/public.pem",
		"grpc_address": "localhost:3200"
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadAgentFileConfig(path)
	if err != nil {
		t.Fatalf("load agent config: %v", err)
	}

	if cfg.Address == nil || *cfg.Address != "localhost:8080" {
		t.Fatalf("unexpected address: %#v", cfg.Address)
	}

	if cfg.ReportInterval == nil || *cfg.ReportInterval != 12 {
		t.Fatalf("unexpected report interval: %#v", cfg.ReportInterval)
	}

	if cfg.PollInterval == nil || *cfg.PollInterval != 3 {
		t.Fatalf("unexpected poll interval: %#v", cfg.PollInterval)
	}

	if cfg.Key == nil || *cfg.Key != "secret" {
		t.Fatalf("unexpected key: %#v", cfg.Key)
	}

	if cfg.RateLimit == nil || *cfg.RateLimit != 7 {
		t.Fatalf("unexpected rate limit: %#v", cfg.RateLimit)
	}

	if cfg.CryptoKey == nil || *cfg.CryptoKey != "/tmp/public.pem" {
		t.Fatalf("unexpected crypto key: %#v", cfg.CryptoKey)
	}

	if cfg.GRPCAddress == nil || *cfg.GRPCAddress != "localhost:3200" {
		t.Fatalf("unexpected gRPC address: %#v", cfg.GRPCAddress)
	}
}

func TestLoadFileConfig_IgnoresUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.json")
	data := []byte(`{
		"address": "localhost:8080",
		"store_interval": "15s",
		"unknown_field": "kept for forward compatibility"
	}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadServerFileConfig(path)
	if err != nil {
		t.Fatalf("load server config: %v", err)
	}

	if cfg.Address == nil || *cfg.Address != "localhost:8080" {
		t.Fatalf("unexpected address: %#v", cfg.Address)
	}
	if cfg.StoreInterval == nil || *cfg.StoreInterval != 15 {
		t.Fatalf("unexpected store interval: %#v", cfg.StoreInterval)
	}
}

func TestParseIntervalSeconds(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "duration", value: "5s", want: 5},
		{name: "seconds", value: "10", want: 10},
		{name: "zero", value: "0s", want: 0},
		{name: "fractional", value: "1500ms", wantErr: true},
		{name: "invalid", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntervalSeconds(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
