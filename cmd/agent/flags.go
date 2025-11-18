package main

import (
	"flag"

	"github.com/user/practicum-metrics/internal/config"
)

var (
	flagRunAddr        string
	flagReportInterval int
	flagPollInterval   int
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port of metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagReportInterval = config.GetEnvAsInt("REPORT_INTERVAL", flagReportInterval)
	flagPollInterval = config.GetEnvAsInt("POLL_INTERVAL", flagPollInterval)
}
