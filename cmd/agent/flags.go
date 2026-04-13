package main

import (
	"flag"
	"log"

	"github.com/user/practicum-metrics/internal/config"
)

var (
	flagRunAddr        string
	flagReportInterval int
	flagPollInterval   int
	flagKey            string
	flagRateLimit      int
	flagCryptoKey      string
	flagConfigPath     string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port of metrics server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&flagKey, "k", "", "secret key for signing requests")
	flag.IntVar(&flagRateLimit, "l", 3, "maximum number of concurrent outgoing requests")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "path to public key for request encryption")
	flag.StringVar(&flagConfigPath, "c", "", "path to JSON config file")
	flag.StringVar(&flagConfigPath, "config", "", "path to JSON config file")
	flag.Parse()

	visitedFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		visitedFlags[f.Name] = true
	})

	configPath := config.GetEnvAsString("CONFIG", flagConfigPath)
	if configPath != "" {
		fileConfig, err := config.LoadAgentFileConfig(configPath)
		if err != nil {
			log.Fatalf("failed to load config file: %v", err)
		}

		if !visitedFlags["a"] && fileConfig.Address != nil {
			flagRunAddr = *fileConfig.Address
		}
		if !visitedFlags["r"] && fileConfig.ReportInterval != nil {
			flagReportInterval = *fileConfig.ReportInterval
		}
		if !visitedFlags["p"] && fileConfig.PollInterval != nil {
			flagPollInterval = *fileConfig.PollInterval
		}
		if !visitedFlags["k"] && fileConfig.Key != nil {
			flagKey = *fileConfig.Key
		}
		if !visitedFlags["l"] && fileConfig.RateLimit != nil {
			flagRateLimit = *fileConfig.RateLimit
		}
		if !visitedFlags["crypto-key"] && fileConfig.CryptoKey != nil {
			flagCryptoKey = *fileConfig.CryptoKey
		}
	}

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagReportInterval = config.GetEnvAsInt("REPORT_INTERVAL", flagReportInterval)
	flagPollInterval = config.GetEnvAsInt("POLL_INTERVAL", flagPollInterval)
	flagKey = config.GetEnvAsString("KEY", flagKey)
	flagRateLimit = config.GetEnvAsInt("RATE_LIMIT", flagRateLimit)
	flagCryptoKey = config.GetEnvAsString("CRYPTO_KEY", flagCryptoKey)
}
