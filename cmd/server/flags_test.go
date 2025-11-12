package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_Defaults(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd"}

	parseFlags()

	if flagRunAddr != "localhost:8080" {
		t.Errorf("expected default address 'localhost:8080', got '%s'", flagRunAddr)
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
