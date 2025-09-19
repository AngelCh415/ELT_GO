package ingest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/config"
	"github.com/AngelCh415/ELT_GO/internal/models"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

type ETL struct {
	c   HTTPClient
	st  *store.MemoryStore
	log *slog.Logger
	cfg config.Config
}

func NewETL(c HTTPClient, st *store.MemoryStore, log *slog.Logger, cfg config.Config) *ETL {
	return &ETL{c: c, st: st, log: log, cfg: cfg}
}

type adsResp []struct {
	Date        string  `json:"date"`
	CampaignID  string  `json:"campaign_id"`
	Channel     string  `json:"channel"`
	Clicks      int     `json:"clicks"`
	Impressions int     `json:"impressions"`
	Cost        float64 `json:"cost"`
	UTMCampaign string  `json:"utm_campaign"`
	UTMSource   string  `json:"utm_source"`
	UTMMedium   string  `json:"utm_medium"`
}

type crmResp []struct {
	OpportunityID string  `json:"opportunity_id"`
	ContactEmail  string  `json:"contact_email"`
	Stage         string  `json:"stage"`
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"created_at"`
	UTMCampaign   string  `json:"utm_campaign"`
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
}

func (e *ETL) Run(ctx context.Context, since *time.Time) error {
	// ADS
	var aResp adsResp
	if err := GetJSONWithRetry(e.c, e.cfg.AdsURL, &aResp); err != nil {
		return err
	}
	// CRM
	var cResp crmResp
	if err := GetJSONWithRetry(e.c, e.cfg.CrmURL, &cResp); err != nil {
		return err
	}

	// Normalizar + filtrar fechas
	for _, r := range aResp {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(r.Date))
		if err != nil {
			continue
		}
		if since != nil && dayUTC(d).Before(dayUTC(*since)) {
			continue
		}
		key := "ads|" + r.Date + "|" + r.CampaignID + "|" + r.Channel
		if !e.st.MarkSeen(key) {
			continue
		} // idempotencia
		e.st.UpsertAds(models.AdsPerformance{
			Date:        d,
			CampaignID:  strings.TrimSpace(r.CampaignID),
			Channel:     strings.TrimSpace(r.Channel),
			Clicks:      max0(r.Clicks),
			Impressions: max0(r.Impressions),
			Cost:        maxf(r.Cost),
			UTMCampaign: coalesce(r.UTMCampaign, "unknown"),
			UTMSource:   coalesce(r.UTMSource, "unknown"),
			UTMMedium:   coalesce(r.UTMMedium, "unknown"),
		})
	}

	for _, r := range cResp {
		if r.CreatedAt == "" {
			continue
		}
		d, err := time.Parse(time.RFC3339, r.CreatedAt)
		if err != nil {
			continue
		}
		if since != nil && dayUTC(d).Before(dayUTC(*since)) {
			continue
		}
		key := "crm|" + r.OpportunityID
		if r.OpportunityID == "" {
			key = "crm|" + d.Format(time.RFC3339) + "|" + r.ContactEmail
		}
		if !e.st.MarkSeen(key) {
			continue
		}
		e.st.UpsertCRM(models.Opportunity{
			OpportunityID: r.OpportunityID,
			ContactEmail:  strings.ToLower(strings.TrimSpace(r.ContactEmail)),
			Stage:         strings.ToLower(strings.TrimSpace(r.Stage)),
			Amount:        maxf(r.Amount),
			CreatedAt:     d,
			UTMCampaign:   coalesce(r.UTMCampaign, "unknown"),
			UTMSource:     coalesce(r.UTMSource, "unknown"),
			UTMMedium:     coalesce(r.UTMMedium, "unknown"),
		})
	}

	e.log.Info("ingest complete", slog.Int("agg_count", len(e.st.All())))
	return nil

}
func (e *ETL) ExportDay(ctx context.Context, date time.Time) (int, error) {
	if e.cfg.SinkURL == "" || e.cfg.SinkSecret == "" {
		return 0, errors.New("sink not configured")
	}
	to := dayUTC(date)
	from := to
	rows := e.toMetrics(e.st.Query(from, to, nil))
	if len(rows) == 0 {
		return 0, nil
	}
	payload := rows // exportamos arreglo
	b, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(e.cfg.SinkSecret))
	mac.Write(b)
	sig := hex.EncodeToString(mac.Sum(nil))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.SinkURL, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sig)
	resp, err := e.c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, errors.New("export sink non-2xx")
	}
	return len(rows), nil
}

// Helpers de métricas calculadas
func (e *ETL) toMetrics(aggs []models.DailyAgg) []models.Metrics {
	sort.Slice(aggs, func(i, j int) bool { return aggs[i].Key.Date.Before(aggs[j].Key.Date) })
	out := make([]models.Metrics, 0, len(aggs))
	for _, a := range aggs {
		cpc := safeDivF(a.Cost, float64(max1(a.Clicks)))
		cpa := safeDivF(a.Cost, float64(max1(a.Leads)))
		cvr1 := safeDivF(float64(a.Opportunities), float64(max1(a.Leads)))
		cvr2 := safeDivF(float64(a.ClosedWon), float64(max1(a.Opportunities)))
		roas := safeDivF(a.Revenue, maxf(a.Cost))
		out = append(out, models.Metrics{
			Date:          a.Key.Date.Format("2006-01-02"),
			Channel:       a.Key.Channel,
			CampaignID:    a.Key.CampaignID,
			UTMCampaign:   a.Key.UTMCampaign,
			UTMSource:     a.Key.UTMSource,
			UTMMedium:     a.Key.UTMMedium,
			Clicks:        a.Clicks,
			Impressions:   a.Impressions,
			Cost:          round2(a.Cost),
			Leads:         a.Leads,
			Opportunities: a.Opportunities,
			ClosedWon:     a.ClosedWon,
			Revenue:       round2(a.Revenue),
			CPC:           round3(cpc),
			CPA:           round2(cpa),
			CVRLeadToOpp:  round3(cvr1),
			CVROppToWon:   round3(cvr2),
			ROAS:          round2(roas),
		})
	}
	return out
}

func coalesce(s, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}
func dayUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
func safeDivF(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
func max0(i int) int {
	if i < 0 {
		return 0
	}
	return i
}
func max1(i int) int {
	if i <= 0 {
		return 1
	}
	return i
}
func maxf(f float64) float64 {
	if f < 0 {
		return 0
	}
	return f
}
func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }
func round3(f float64) float64 { return float64(int64(f*1000+0.5)) / 1000 }
