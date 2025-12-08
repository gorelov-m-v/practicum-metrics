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
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval in seconds (0 for synchronous)")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore previously saved values on startup")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database connection string")
	flag.Parse()

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagStoreInterval = config.GetEnvAsInt("STORE_INTERVAL", flagStoreInterval)
	flagFileStoragePath = config.GetEnvAsString("FILE_STORAGE_PATH", flagFileStoragePath)
	flagRestore = config.GetEnvAsBool("RESTORE", flagRestore)
	flagDatabaseDSN = config.GetEnvAsString("DATABASE_DSN", flagDatabaseDSN)
}
