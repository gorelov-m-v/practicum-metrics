package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerFileConfig struct {
	Address       *string
	Restore       *bool
	StoreInterval *int
	StoreFile     *string
	DatabaseDSN   *string
	Key           *string
	AuditFile     *string
	AuditURL      *string
	CryptoKey     *string
	TrustedSubnet *string
}

type AgentFileConfig struct {
	Address        *string
	ReportInterval *int
	PollInterval   *int
	Key            *string
	RateLimit      *int
	CryptoKey      *string
}

type serverFileConfigRaw struct {
	Address       *string `json:"address"`
	Restore       *bool   `json:"restore"`
	StoreInterval *string `json:"store_interval"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	Key           *string `json:"key"`
	AuditFile     *string `json:"audit_file"`
	AuditURL      *string `json:"audit_url"`
	CryptoKey     *string `json:"crypto_key"`
	TrustedSubnet *string `json:"trusted_subnet"`
}

type agentFileConfigRaw struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	Key            *string `json:"key"`
	RateLimit      *int    `json:"rate_limit"`
	CryptoKey      *string `json:"crypto_key"`
}

func LoadServerFileConfig(path string) (ServerFileConfig, error) {
	var raw serverFileConfigRaw
	if err := loadJSONFile(path, &raw); err != nil {
		return ServerFileConfig{}, err
	}

	cfg := ServerFileConfig{
		Address:       raw.Address,
		Restore:       raw.Restore,
		StoreFile:     raw.StoreFile,
		DatabaseDSN:   raw.DatabaseDSN,
		Key:           raw.Key,
		AuditFile:     raw.AuditFile,
		AuditURL:      raw.AuditURL,
		CryptoKey:     raw.CryptoKey,
		TrustedSubnet: raw.TrustedSubnet,
	}

	if raw.StoreInterval != nil {
		value, err := parseIntervalSeconds(*raw.StoreInterval)
		if err != nil {
			return ServerFileConfig{}, fmt.Errorf("parse store_interval: %w", err)
		}
		cfg.StoreInterval = &value
	}

	return cfg, nil
}

func LoadAgentFileConfig(path string) (AgentFileConfig, error) {
	var raw agentFileConfigRaw
	if err := loadJSONFile(path, &raw); err != nil {
		return AgentFileConfig{}, err
	}

	cfg := AgentFileConfig{
		Address:   raw.Address,
		Key:       raw.Key,
		RateLimit: raw.RateLimit,
		CryptoKey: raw.CryptoKey,
	}

	if raw.ReportInterval != nil {
		value, err := parseIntervalSeconds(*raw.ReportInterval)
		if err != nil {
			return AgentFileConfig{}, fmt.Errorf("parse report_interval: %w", err)
		}
		cfg.ReportInterval = &value
	}

	if raw.PollInterval != nil {
		value, err := parseIntervalSeconds(*raw.PollInterval)
		if err != nil {
			return AgentFileConfig{}, fmt.Errorf("parse poll_interval: %w", err)
		}
		cfg.PollInterval = &value
	}

	return cfg, nil
}

func loadJSONFile(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("decode config file: %w", err)
	}

	return nil
}

func parseIntervalSeconds(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, fmt.Errorf("negative value")
		}
		return seconds, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}

	if duration < 0 {
		return 0, fmt.Errorf("negative value")
	}

	if duration%time.Second != 0 {
		return 0, fmt.Errorf("duration must be a whole number of seconds")
	}

	return int(duration / time.Second), nil
}
