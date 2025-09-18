package metrics

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/models"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

type Service struct{ st *store.MemoryStore }

func NewService(st *store.MemoryStore) *Service { return &Service{st: st} }

func (s *Service) QueryChannel(v url.Values) ([]models.Metrics, error) {
	from, _ := time.Parse("2006-01-02", v.Get("from"))
	to, _ := time.Parse("2006-01-02", v.Get("to"))
	channel := strings.TrimSpace(v.Get("channel"))
	limit := atoiDef(v.Get("limit"), 100)
	offset := atoiDef(v.Get("offset"), 0)

	aggs := s.st.Query(from, to, func(a models.DailyAgg) bool {
		if channel != "" && a.Key.Channel != channel {
			return false
		}
		return true
	})
	rows := toMetricsSlice(aggs)
	return paginate(rows, limit, offset), nil
}

func (s *Service) QueryFunnel(v url.Values) ([]models.Metrics, error) {
	from, _ := time.Parse("2006-01-02", v.Get("from"))
	to, _ := time.Parse("2006-01-02", v.Get("to"))
	utmC := strings.TrimSpace(v.Get("utm_campaign"))
	utmS := strings.TrimSpace(v.Get("utm_source"))
	utmM := strings.TrimSpace(v.Get("utm_medium"))
	limit := atoiDef(v.Get("limit"), 100)
	offset := atoiDef(v.Get("offset"), 0)

	aggs := s.st.Query(from, to, func(a models.DailyAgg) bool {
		if utmC != "" && a.Key.UTMCampaign != utmC {
			return false
		}
		if utmS != "" && a.Key.UTMSource != utmS {
			return false
		}
		if utmM != "" && a.Key.UTMMedium != utmM {
			return false
		}
		return true
	})
	rows := toMetricsSlice(aggs)
	return paginate(rows, limit, offset), nil
}

func toMetricsSlice(aggs []models.DailyAgg) []models.Metrics {
	// Reusa cálculos del ETL: duplicamos lógica mínima para evitar dependencia circular
	rows := make([]models.Metrics, 0, len(aggs))
	for _, a := range aggs {
		m := models.Metrics{
			Date:          a.Key.Date.Format("2006-01-02"),
			Channel:       a.Key.Channel,
			CampaignID:    a.Key.CampaignID,
			UTMCampaign:   a.Key.UTMCampaign,
			UTMSource:     a.Key.UTMSource,
			UTMMedium:     a.Key.UTMMedium,
			Clicks:        a.Clicks,
			Impressions:   a.Impressions,
			Cost:          a.Cost,
			Leads:         a.Leads,
			Opportunities: a.Opportunities,
			ClosedWon:     a.ClosedWon,
			Revenue:       a.Revenue,
		}
		// métricas derivadas
		if a.Clicks > 0 {
			m.CPC = round3(a.Cost / float64(a.Clicks))
		}
		if a.Leads > 0 {
			m.CPA = round2(a.Cost / float64(a.Leads))
		}
		if a.Leads > 0 {
			m.CVRLeadToOpp = round3(float64(a.Opportunities) / float64(a.Leads))
		}
		if a.Opportunities > 0 {
			m.CVROppToWon = round3(float64(a.ClosedWon) / float64(a.Opportunities))
		}
		if a.Cost > 0 {
			m.ROAS = round2(a.Revenue / a.Cost)
		}
		rows = append(rows, m)
	}
	return rows
}

func paginate[T any](rows []T, limit, offset int) []T {
	if offset >= len(rows) {
		return []T{}
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end]
}

func atoiDef(s string, d int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return d
}
func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }
func round3(f float64) float64 { return float64(int64(f*1000+0.5)) / 1000 }
