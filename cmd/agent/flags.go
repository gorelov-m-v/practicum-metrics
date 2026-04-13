package main

import (
	"flag"

	"github.com/user/practicum-metrics/internal/config"
)

var (
	flagRunAddr        string
	flagReportInterval int
	flagPollInterval   int
	flagKey            string
	flagRateLimit      int
	flagCryptoKey      string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port of metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&flagKey, "k", "", "secret key for signing requests")
	flag.IntVar(&flagRateLimit, "l", 3, "maximum number of concurrent outgoing requests")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "path to public key for request encryption")
	flag.Parse()

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagReportInterval = config.GetEnvAsInt("REPORT_INTERVAL", flagReportInterval)
	flagPollInterval = config.GetEnvAsInt("POLL_INTERVAL", flagPollInterval)
	flagKey = config.GetEnvAsString("KEY", flagKey)
	flagRateLimit = config.GetEnvAsInt("RATE_LIMIT", flagRateLimit)
	flagCryptoKey = config.GetEnvAsString("CRYPTO_KEY", flagCryptoKey)
}
