package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags_Defaults(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")

	os.Args = []string{"cmd"}

	parseFlags()

	if flagRunAddr != "localhost:8080" {
		t.Errorf("expected default address 'localhost:8080', got '%s'", flagRunAddr)
	}

	if flagReportInterval != 10 {
		t.Errorf("expected default report interval 10, got %d", flagReportInterval)
	}

	if flagPollInterval != 2 {
		t.Errorf("expected default poll interval 2, got %d", flagPollInterval)
	}
}

func TestParseFlags_CommandLineFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")

	os.Args = []string{"cmd", "-a", "localhost:9090", "-r", "5", "-p", "1"}

	parseFlags()

	if flagRunAddr != "localhost:9090" {
		t.Errorf("expected address 'localhost:9090', got '%s'", flagRunAddr)
	}

	if flagReportInterval != 5 {
		t.Errorf("expected report interval 5, got %d", flagReportInterval)
	}

	if flagPollInterval != 1 {
		t.Errorf("expected poll interval 1, got %d", flagPollInterval)
	}
}

func TestParseFlags_EnvironmentVariables(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "localhost:7070")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("POLL_INTERVAL", "3")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	os.Args = []string{"cmd"}

	parseFlags()

	if flagRunAddr != "localhost:7070" {
		t.Errorf("expected address from env 'localhost:7070', got '%s'", flagRunAddr)
	}

	if flagReportInterval != 15 {
		t.Errorf("expected report interval from env 15, got %d", flagReportInterval)
	}

	if flagPollInterval != 3 {
		t.Errorf("expected poll interval from env 3, got %d", flagPollInterval)
	}
}

func TestParseFlags_EnvironmentOverridesFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("ADDRESS", "localhost:6060")
	os.Setenv("REPORT_INTERVAL", "20")
	os.Setenv("POLL_INTERVAL", "4")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	os.Args = []string{"cmd", "-a", "localhost:9090", "-r", "5", "-p", "1"}

	parseFlags()

	if flagRunAddr != "localhost:6060" {
		t.Errorf("expected env to override flag, got address '%s'", flagRunAddr)
	}

	if flagReportInterval != 20 {
		t.Errorf("expected env to override flag, got report interval %d", flagReportInterval)
	}

	if flagPollInterval != 4 {
		t.Errorf("expected env to override flag, got poll interval %d", flagPollInterval)
	}
}

func TestParseFlags_InvalidEnvironmentVariables(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Setenv("REPORT_INTERVAL", "invalid")
	os.Setenv("POLL_INTERVAL", "not_a_number")
	defer func() {
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
	}()

	os.Args = []string{"cmd", "-r", "5", "-p", "1"}

	parseFlags()

	if flagReportInterval != 5 {
		t.Errorf("expected flag value when env is invalid, got report interval %d", flagReportInterval)
	}

	if flagPollInterval != 1 {
		t.Errorf("expected flag value when env is invalid, got poll interval %d", flagPollInterval)
	}
}

func TestParseFlags_PartialEnvironmentVariables(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")

	os.Setenv("ADDRESS", "localhost:5050")
	defer os.Unsetenv("ADDRESS")

	os.Args = []string{"cmd", "-a", "localhost:9090", "-r", "5", "-p", "1"}

	parseFlags()

	if flagRunAddr != "localhost:5050" {
		t.Errorf("expected address from env 'localhost:5050', got '%s'", flagRunAddr)
	}

	if flagReportInterval != 5 {
		t.Errorf("expected report interval from flag 5, got %d", flagReportInterval)
	}

	if flagPollInterval != 1 {
		t.Errorf("expected poll interval from flag 1, got %d", flagPollInterval)
	}
}
