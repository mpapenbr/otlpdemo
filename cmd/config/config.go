package config

import (
	"os"

	"github.com/mpapenbr/otlpdemo/log"
)

type Config struct {
	EnableTelemetry   bool
	TelemetryEndpoint string
	Worker            int
	LogFormat         string
	LogLevel          string
}

var (
	config *Config     = nil
	logger *log.Logger = nil
)

func DefaultConfig() *Config {
	if config == nil {
		config = NewConfig()
	}
	return config
}

func NewConfig() *Config {
	return &Config{}
}

func parseLogLevel(l string, defaultVal log.Level) log.Level {
	level, err := log.ParseLevel(l)
	if err != nil {
		return defaultVal
	}
	return level
}

func InitLogger(cfg *Config) {
	switch cfg.LogFormat {
	case "json":
		logger = log.New(
			os.Stderr,
			parseLogLevel(cfg.LogLevel, log.InfoLevel),
			log.WithCaller(true),
			log.AddCallerSkip(1))
	default:
		logger = log.DevLogger(
			os.Stderr,
			parseLogLevel(cfg.LogLevel, log.DebugLevel),
			log.WithCaller(true),
			log.AddCallerSkip(1))
	}

	log.ResetDefault(logger)
}
