package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagRunAddr         string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval in seconds (0 for synchronous)")
	flag.StringVar(&flagFileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore previously saved values on startup")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		flagRunAddr = envAddr
	}

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		if interval, err := strconv.Atoi(envInterval); err == nil {
			flagStoreInterval = interval
		}
	}

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		flagFileStoragePath = envPath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restore, err := strconv.ParseBool(envRestore); err == nil {
			flagRestore = restore
		}
	}
}
