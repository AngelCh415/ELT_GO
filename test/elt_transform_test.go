package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/ingest"
	"github.com/AngelCh415/ELT_GO/internal/models"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

// helper: hace la petición y devuelve código HTTP + error de red (si hubo)
func fetchURL(c ingest.HTTPClient, url string) (int, error) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	resp, err := c.Do(req)
	if err != nil {
		return 0, err // error de transporte (timeout, conexión, etc.)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func TestHTTPClientHandles500(t *testing.T) {
	// servidor fake que devuelve 500
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(2 * time.Second)
	code, err := fetchURL(client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected network error: %v", err)
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", code)
	}
}

func TestHTTPClientHandles404(t *testing.T) {
	// servidor fake que devuelve 404
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(2 * time.Second)
	code, err := fetchURL(client, srv.URL)
	if err != nil {
		t.Fatalf("unexpected network error: %v", err)
	}
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}

func TestHTTPClientHandlesTimeout(t *testing.T) {
	// servidor fake que se tarda más del timeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer srv.Close()

	client := ingest.NewHTTPClient(1 * time.Second) // timeout corto
	_, err := fetchURL(client, srv.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestDerivedMetricsSafeDiv(t *testing.T) {
	st := store.NewMemoryStore()

	// fecha base
	d, _ := time.Parse("2006-01-02", "2025-08-01")

	// agrego un Ads con clicks = 0
	st.UpsertAds(models.AdsPerformance{
		Date:        d,
		Channel:     "google_ads",
		CampaignID:  "C-1001",
		Clicks:      0,
		Impressions: 100,
		Cost:        100,
		UTMCampaign: "camp",
		UTMSource:   "src",
		UTMMedium:   "med",
	})

	// agrego un CRM con closed_won
	st.UpsertCRM(models.Opportunity{
		CreatedAt:   d,
		Stage:       "closed_won",
		Amount:      500,
		UTMCampaign: "camp",
		UTMSource:   "src",
		UTMMedium:   "med",
	})

	aggs := st.All()
	for _, agg := range aggs {
		t.Logf("Agg: %+v", agg)
	}
	if len(aggs) == 0 {
		t.Fatal("expected aggs")
	}
	if aggs[0].ClosedWon != 1 {
		t.Fatalf("expected closed_won=1, got=%d", aggs[0].ClosedWon)
	}
}
