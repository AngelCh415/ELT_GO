package test

import (
	"testing"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/models"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

func TestDerivedMetricsSafeDiv(t *testing.T) {
	st := store.NewMemoryStore()

	// 2025-08-01
	d, _ := time.Parse("2006-01-02", "2025-08-01")

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
	st.UpsertCRM(models.Opportunity{
		CreatedAt:   d,
		Stage:       "closed_won",
		Amount:      500,
		UTMCampaign: "camp",
		UTMSource:   "src",
		UTMMedium:   "med",
	})

	aggs := st.All()
	if len(aggs) == 0 {
		t.Fatal("expected aggs")
	}
	if aggs[0].Leads != 1 {
		t.Fatalf("leads=1 got=%d", aggs[0].Leads)
	}
	if aggs[0].ClosedWon != 1 {
		t.Fatalf("won=1 got=%d", aggs[0].ClosedWon)
	}
}
