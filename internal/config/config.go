package config

import (
	"log/slog"
	"os"
	"time"
)

type Config struct {
	AdsURL      string
	CrmURL      string
	SinkURL     string
	SinkSecret  string
	Port        string
	HTTPTimeout time.Duration
	LogLevel    slog.Level
}

func FromEnv() Config {
	to := 15 * time.Second
	if v := os.Getenv("HTTP_TIMEOUT_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			to = d
		}
	}
	lvl := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		lvl = slog.LevelDebug
	}
	return Config{
		AdsURL:      os.Getenv("ADS_API_URL"),
		CrmURL:      os.Getenv("CRM_API_URL"),
		SinkURL:     os.Getenv("SINK_URL"),
		SinkSecret:  os.Getenv("SINK_SECRET"),
		Port:        envOr("PORT", "8080"),
		HTTPTimeout: to,
		LogLevel:    lvl,
	}
}

func envOr(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
