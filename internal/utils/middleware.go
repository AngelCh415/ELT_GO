package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey string

const requestIDKey ctxKey = "rid"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := newRID()
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey, rid))
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r)
	})
}

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Info("http", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.String("rid", RID(r.Context())), slog.Duration("latency", time.Since(start)))
		})
	}
}

func RID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func newRID() string { b := make([]byte, 8); rand.Read(b); return hex.EncodeToString(b) }
