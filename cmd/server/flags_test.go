package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_Defaults(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")

	os.Args = []string{"cmd"}

	parseFlags()

	if flagRunAddr != "localhost:8080" {
		t.Errorf("expected default address 'localhost:8080', got '%s'", flagRunAddr)
	}

	if flagStoreInterval != 300 {
		t.Errorf("expected default store interval 300, got %d", flagStoreInterval)
	}

	if flagFileStoragePath != "/tmp/metrics-db.json" {
		t.Errorf("expected default file storage path '/tmp/metrics-db.json', got '%s'", flagFileStoragePath)
	}

	if flagRestore != true {
		t.Errorf("expected default restore true, got %v", flagRestore)
	}
}

func TestParseFlags_CommandLineFlag(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "localhost:9090"}

	parseFlags()

	if flagRunAddr != "localhost:9090" {
		t.Errorf("expected address from flag 'localhost:9090', got '%s'", flagRunAddr)
	}
}

func TestParseFlags_EnvironmentVariable(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "localhost:7070")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd"}

	parseFlags()

	if flagRunAddr != "localhost:7070" {
		t.Errorf("expected address from env 'localhost:7070', got '%s'", flagRunAddr)
	}
}

func TestParseFlags_EnvironmentOverridesFlag(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "localhost:6060")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "localhost:9090"}

	parseFlags()

	if flagRunAddr != "localhost:6060" {
		t.Errorf("expected env to override flag, got address '%s'", flagRunAddr)
	}
}

func TestParseFlags_EmptyEnvironmentVariable(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "localhost:9090"}

	parseFlags()

	if flagRunAddr != "localhost:9090" {
		t.Errorf("expected flag value when env is empty, got address '%s'", flagRunAddr)
	}
}

func TestParseFlags_PriorityOrder(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		flagValue   string
		expected    string
		description string
	}{
		{
			name:        "env_and_flag_set",
			envValue:    "localhost:3000",
			flagValue:   "localhost:4000",
			expected:    "localhost:3000",
			description: "environment variable should override flag",
		},
		{
			name:        "only_flag_set",
			envValue:    "",
			flagValue:   "localhost:4000",
			expected:    "localhost:4000",
			description: "flag should be used when env is not set",
		},
		{
			name:        "only_env_set",
			envValue:    "localhost:3000",
			flagValue:   "",
			expected:    "localhost:3000",
			description: "environment variable should be used when flag is not set",
		},
		{
			name:        "nothing_set",
			envValue:    "",
			flagValue:   "",
			expected:    "localhost:8080",
			description: "default value should be used when neither env nor flag is set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			if tt.envValue != "" {
				os.Setenv("ADDRESS", tt.envValue)
				defer os.Unsetenv("ADDRESS")
			} else {
				os.Unsetenv("ADDRESS")
			}

			if tt.flagValue != "" {
				os.Args = []string{"cmd", "-a", tt.flagValue}
			} else {
				os.Args = []string{"cmd"}
			}

			parseFlags()

			if flagRunAddr != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, flagRunAddr)
			}
		})
	}
}

func TestParseFlags_StoreInterval(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue string
		expected  int
	}{
		{
			name:      "env_overrides_flag",
			envValue:  "60",
			flagValue: "120",
			expected:  60,
		},
		{
			name:      "flag_only",
			envValue:  "",
			flagValue: "180",
			expected:  180,
		},
		{
			name:      "env_only",
			envValue:  "240",
			flagValue: "",
			expected:  240,
		},
		{
			name:      "synchronous_mode",
			envValue:  "0",
			flagValue: "",
			expected:  0,
		},
		{
			name:      "default_value",
			envValue:  "",
			flagValue: "",
			expected:  300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Unsetenv("ADDRESS")
			os.Unsetenv("FILE_STORAGE_PATH")
			os.Unsetenv("RESTORE")

			if tt.envValue != "" {
				os.Setenv("STORE_INTERVAL", tt.envValue)
				defer os.Unsetenv("STORE_INTERVAL")
			} else {
				os.Unsetenv("STORE_INTERVAL")
			}

			if tt.flagValue != "" {
				os.Args = []string{"cmd", "-i", tt.flagValue}
			} else {
				os.Args = []string{"cmd"}
			}

			parseFlags()

			if flagStoreInterval != tt.expected {
				t.Errorf("expected store interval %d, got %d", tt.expected, flagStoreInterval)
			}
		})
	}
}

func TestParseFlags_FileStoragePath(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue string
		expected  string
	}{
		{
			name:      "env_overrides_flag",
			envValue:  "/custom/env/path.json",
			flagValue: "/custom/flag/path.json",
			expected:  "/custom/env/path.json",
		},
		{
			name:      "flag_only",
			envValue:  "",
			flagValue: "/flag/path.json",
			expected:  "/flag/path.json",
		},
		{
			name:      "env_only",
			envValue:  "/env/path.json",
			flagValue: "",
			expected:  "/env/path.json",
		},
		{
			name:      "default_value",
			envValue:  "",
			flagValue: "",
			expected:  "/tmp/metrics-db.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Unsetenv("ADDRESS")
			os.Unsetenv("STORE_INTERVAL")
			os.Unsetenv("RESTORE")

			if tt.envValue != "" {
				os.Setenv("FILE_STORAGE_PATH", tt.envValue)
				defer os.Unsetenv("FILE_STORAGE_PATH")
			} else {
				os.Unsetenv("FILE_STORAGE_PATH")
			}

			if tt.flagValue != "" {
				os.Args = []string{"cmd", "-f", tt.flagValue}
			} else {
				os.Args = []string{"cmd"}
			}

			parseFlags()

			if flagFileStoragePath != tt.expected {
				t.Errorf("expected file storage path '%s', got '%s'", tt.expected, flagFileStoragePath)
			}
		})
	}
}

func TestParseFlags_Restore(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		flagValue string
		expected  bool
	}{
		{
			name:      "env_overrides_flag_true",
			envValue:  "true",
			flagValue: "false",
			expected:  true,
		},
		{
			name:      "env_overrides_flag_false",
			envValue:  "false",
			flagValue: "true",
			expected:  false,
		},
		{
			name:      "flag_only_false",
			envValue:  "",
			flagValue: "false",
			expected:  false,
		},
		{
			name:      "flag_only_true",
			envValue:  "",
			flagValue: "true",
			expected:  true,
		},
		{
			name:      "env_only_false",
			envValue:  "false",
			flagValue: "",
			expected:  false,
		},
		{
			name:      "env_only_true",
			envValue:  "true",
			flagValue: "",
			expected:  true,
		},
		{
			name:      "default_value",
			envValue:  "",
			flagValue: "",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Unsetenv("ADDRESS")
			os.Unsetenv("STORE_INTERVAL")
			os.Unsetenv("FILE_STORAGE_PATH")

			if tt.envValue != "" {
				os.Setenv("RESTORE", tt.envValue)
				defer os.Unsetenv("RESTORE")
			} else {
				os.Unsetenv("RESTORE")
			}

			if tt.flagValue != "" {
				os.Args = []string{"cmd", "-r=" + tt.flagValue}
			} else {
				os.Args = []string{"cmd"}
			}

			parseFlags()

			if flagRestore != tt.expected {
				t.Errorf("expected restore %v, got %v", tt.expected, flagRestore)
			}
		})
	}
}

func TestParseFlags_AllParameters(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "localhost:9000")
	os.Setenv("STORE_INTERVAL", "60")
	os.Setenv("FILE_STORAGE_PATH", "/custom/metrics.json")
	os.Setenv("RESTORE", "false")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
	}()

	os.Args = []string{"cmd", "-a", "localhost:8000", "-i", "120", "-f", "/other/path.json", "-r=true"}

	parseFlags()

	if flagRunAddr != "localhost:9000" {
		t.Errorf("expected address 'localhost:9000', got '%s'", flagRunAddr)
	}

	if flagStoreInterval != 60 {
		t.Errorf("expected store interval 60, got %d", flagStoreInterval)
	}

	if flagFileStoragePath != "/custom/metrics.json" {
		t.Errorf("expected file storage path '/custom/metrics.json', got '%s'", flagFileStoragePath)
	}

	if flagRestore != false {
		t.Errorf("expected restore false, got %v", flagRestore)
	}
}
