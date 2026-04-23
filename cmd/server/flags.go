package main

import (
	"flag"
	"log"

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
	flagConfigPath      string
	flagTrustedSubnet   string
	flagGRPCAddr        string
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
	flag.StringVar(&flagTrustedSubnet, "t", "", "trusted subnet in CIDR notation")
	flag.StringVar(&flagGRPCAddr, "g", "", "address and port to run gRPC server")
	flag.StringVar(&flagGRPCAddr, "grpc-address", "", "address and port to run gRPC server")
	flag.StringVar(&flagConfigPath, "c", "", "path to JSON config file")
	flag.StringVar(&flagConfigPath, "config", "", "path to JSON config file")
	flag.Parse()

	visitedFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		visitedFlags[f.Name] = true
	})

	configPath := config.GetEnvAsString("CONFIG", flagConfigPath)
	if configPath != "" {
		fileConfig, err := config.LoadServerFileConfig(configPath)
		if err != nil {
			log.Fatalf("failed to load config file: %v", err)
		}

		if !visitedFlags["a"] && fileConfig.Address != nil {
			flagRunAddr = *fileConfig.Address
		}
		if !visitedFlags["i"] && fileConfig.StoreInterval != nil {
			flagStoreInterval = *fileConfig.StoreInterval
		}
		if !visitedFlags["f"] && fileConfig.StoreFile != nil {
			flagFileStoragePath = *fileConfig.StoreFile
		}
		if !visitedFlags["r"] && fileConfig.Restore != nil {
			flagRestore = *fileConfig.Restore
		}
		if !visitedFlags["d"] && fileConfig.DatabaseDSN != nil {
			flagDatabaseDSN = *fileConfig.DatabaseDSN
		}
		if !visitedFlags["k"] && fileConfig.Key != nil {
			flagKey = *fileConfig.Key
		}
		if !visitedFlags["audit-file"] && fileConfig.AuditFile != nil {
			flagAuditFile = *fileConfig.AuditFile
		}
		if !visitedFlags["audit-url"] && fileConfig.AuditURL != nil {
			flagAuditURL = *fileConfig.AuditURL
		}
		if !visitedFlags["crypto-key"] && fileConfig.CryptoKey != nil {
			flagCryptoKey = *fileConfig.CryptoKey
		}
		if !visitedFlags["t"] && fileConfig.TrustedSubnet != nil {
			flagTrustedSubnet = *fileConfig.TrustedSubnet
		}
		if !visitedFlags["g"] && !visitedFlags["grpc-address"] && fileConfig.GRPCAddress != nil {
			flagGRPCAddr = *fileConfig.GRPCAddress
		}
	}

	flagRunAddr = config.GetEnvAsString("ADDRESS", flagRunAddr)
	flagStoreInterval = config.GetEnvAsInt("STORE_INTERVAL", flagStoreInterval)
	flagFileStoragePath = config.GetEnvAsString("STORE_FILE", flagFileStoragePath)
	flagFileStoragePath = config.GetEnvAsString("FILE_STORAGE_PATH", flagFileStoragePath)
	flagRestore = config.GetEnvAsBool("RESTORE", flagRestore)
	flagDatabaseDSN = config.GetEnvAsString("DATABASE_DSN", flagDatabaseDSN)
	flagKey = config.GetEnvAsString("KEY", flagKey)
	flagAuditFile = config.GetEnvAsString("AUDIT_FILE", flagAuditFile)
	flagAuditURL = config.GetEnvAsString("AUDIT_URL", flagAuditURL)
	flagCryptoKey = config.GetEnvAsString("CRYPTO_KEY", flagCryptoKey)
	flagTrustedSubnet = config.GetEnvAsString("TRUSTED_SUBNET", flagTrustedSubnet)
	flagGRPCAddr = config.GetEnvAsString("GRPC_ADDRESS", flagGRPCAddr)
}
