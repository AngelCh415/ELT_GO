package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/config"
	"github.com/AngelCh415/ELT_GO/internal/httpx"
	"github.com/AngelCh415/ELT_GO/internal/ingest"

	"github.com/AngelCh415/ELT_GO/internal/metrics"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

func main() {
	cfg := config.FromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	cl := ingest.NewHTTPClient(cfg.HTTPTimeout)
	st := store.NewMemoryStore()
	etl := ingest.NewETL(cl, st, logger, cfg)
	mSvc := metrics.NewService(st)

	r := httpx.NewRouter(logger, etl, mSvc)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Info("starting server", slog.String("port", cfg.Port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
