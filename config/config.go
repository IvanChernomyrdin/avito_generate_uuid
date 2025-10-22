package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Addr           string
	DatabaseDSN    string
	LogDir         string
	LogFileMaxSize int
}

func NewConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "address-server", "localhost:8080", "HTTP-server address (host:port)")
	flag.StringVar(&cfg.DatabaseDSN, "database-dsn", "postgres://postgres:postgres@localhost:5432/avito?sslmode=disable", "The postgres connection string is passed")
	flag.StringVar(&cfg.LogDir, "dir-logs", "runtime/logs", "The directory of the folder for incoming logs entries is spoecified")
	flag.IntVar(&cfg.LogFileMaxSize, "log-max-size", 128, "The maximum file size in MB for log rotation")
	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DatabaseDSN = envDSN
	}
	if envLog := os.Getenv("DIRNAME_LOGS"); envLog != "" {
		cfg.LogDir = envLog
	}
	if envLogMaxSize := os.Getenv("FILELOG_MAXSIZE"); envLogMaxSize != "" {
		cfg.LogFileMaxSize = ParseInt(envLogMaxSize)
	}
	if envLogMaxSize := os.Getenv("LOG_MAXSIZE_FILE"); envLogMaxSize != "" {
		cfg.LogFileMaxSize = ParseInt(envLogMaxSize)
	}

	return cfg
}
func ParseInt(s string) int {
	integer64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(integer64)
}
