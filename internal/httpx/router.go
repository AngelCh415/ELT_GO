package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/AngelCh415/ELT_GO/internal/ingest"
	"github.com/AngelCh415/ELT_GO/internal/metrics"
	"github.com/AngelCh415/ELT_GO/internal/utils"
)

type router struct{ mux *chi.Mux }

func NewRouter(log *slog.Logger, etl *ingest.ETL, mSvc *metrics.Service) http.Handler {
	mux := chi.NewRouter()
	mux.Use(utils.RequestID)
	mux.Use(utils.Logger(log))

	mux.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte("ok")) })
	mux.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte("ready")) })

	mux.Post("/ingest/run", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("since")
		var since *time.Time
		if q != "" {
			if t, err := time.Parse("2006-01-02", q); err == nil {
				since = &t
			}
		}
		if err := etl.Run(r.Context(), since); err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.WriteHeader(202)
		w.Write([]byte("ingest started"))
	})

	mux.Post("/export/run", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("date")
		if q == "" {
			http.Error(w, "date required (YYYY-MM-DD)", 400)
			return
		}
		t, err := time.Parse("2006-01-02", q)
		if err != nil {
			http.Error(w, "bad date", 400)
			return
		}
		n, err := etl.ExportDay(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		writeJSON(w, map[string]any{"exported": n})
	})

	mux.Get("/metrics/channel", func(w http.ResponseWriter, r *http.Request) {
		rows, err := mSvc.QueryChannel(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, rows)
	})

	mux.Get("/metrics/funnel", func(w http.ResponseWriter, r *http.Request) {
		rows, err := mSvc.QueryFunnel(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, rows)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	enc.Encode(v)
}
