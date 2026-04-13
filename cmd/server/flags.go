package main

import (
	"flag"

	"github.com/user/practicum-metrics/internal/config"
)

var (
	flagRunAddr         string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagDatabaseDSN     string
	flagKey             string
	flagAuditFile       string
	flagAuditURL        string
	flagCryptoKey       string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval in seconds (0 for synchronous)")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore previously saved values on startup")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&flagKey, "k", "", "secret key for signing requests")
	flag.StringVar(&flagAuditFile, "audit-file", "", "path to audit log file")
	flag.StringVar(&flagAuditURL, "audit-url", "", "URL to send audit events to")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "path to private key for request decryption")
	flag.Parse()

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagStoreInterval = config.GetEnvAsInt("STORE_INTERVAL", flagStoreInterval)
	flagFileStoragePath = config.GetEnvAsString("FILE_STORAGE_PATH", flagFileStoragePath)
	flagRestore = config.GetEnvAsBool("RESTORE", flagRestore)
	flagDatabaseDSN = config.GetEnvAsString("DATABASE_DSN", flagDatabaseDSN)
	flagKey = config.GetEnvAsString("KEY", flagKey)
	flagAuditFile = config.GetEnvAsString("AUDIT_FILE", flagAuditFile)
	flagAuditURL = config.GetEnvAsString("AUDIT_URL", flagAuditURL)
	flagCryptoKey = config.GetEnvAsString("CRYPTO_KEY", flagCryptoKey)
}
