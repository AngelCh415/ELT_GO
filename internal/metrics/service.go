package metrics

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/models"
	"github.com/AngelCh415/ELT_GO/internal/store"
)

type Service struct{ st *store.MemoryStore }

func NewService(st *store.MemoryStore) *Service { return &Service{st: st} }
func norm(s string) string                      { return strings.ToLower(strings.TrimSpace(s)) }

func csvSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, p := range strings.Split(s, ",") {
		p = norm(p)
		if p != "" {
			out[p] = struct{}{}
		}
	}
	return out
}

func (s *Service) QueryChannel(v url.Values) ([]models.Metrics, error) {
	from, _ := time.Parse("2006-01-02", v.Get("from"))
	to, _ := time.Parse("2006-01-02", v.Get("to"))
	chSet := csvSet(v.Get("channel"))
	limit := atoiDef(v.Get("limit"), 100)
	offset := atoiDef(v.Get("offset"), 0)

	aggs := s.st.Query(from, to, func(a models.DailyAgg) bool {
		if len(chSet) > 0 {
			_, ok := chSet[norm(a.Key.Channel)]
			if !ok {
				return false
			}
		}
		return true
	})

	// orden determinista
	sort.Slice(aggs, func(i, j int) bool {
		if !aggs[i].Key.Date.Equal(aggs[j].Key.Date) {
			return aggs[i].Key.Date.Before(aggs[j].Key.Date)
		}
		if aggs[i].Key.Channel != aggs[j].Key.Channel {
			return aggs[i].Key.Channel < aggs[j].Key.Channel
		}
		return aggs[i].Key.CampaignID < aggs[j].Key.CampaignID
	})

	rows := toMetricsSlice(aggs)
	limit, offset = clampLimitOffset(limit, offset, len(rows))
	return paginate(rows, limit, offset), nil
}

func (s *Service) QueryFunnel(v url.Values) ([]models.Metrics, error) {
	from, _ := time.Parse("2006-01-02", v.Get("from"))
	to, _ := time.Parse("2006-01-02", v.Get("to"))
	utmC := norm(v.Get("utm_campaign"))
	utmS := norm(v.Get("utm_source"))
	utmM := norm(v.Get("utm_medium"))
	limit := atoiDef(v.Get("limit"), 100)
	offset := atoiDef(v.Get("offset"), 0)

	aggs := s.st.Query(from, to, func(a models.DailyAgg) bool {
		if utmC != "" && norm(a.Key.UTMCampaign) != utmC {
			return false
		}
		if utmS != "" && norm(a.Key.UTMSource) != utmS {
			return false
		}
		if utmM != "" && norm(a.Key.UTMMedium) != utmM {
			return false
		}
		return true
	})

	sort.Slice(aggs, func(i, j int) bool {
		if !aggs[i].Key.Date.Equal(aggs[j].Key.Date) {
			return aggs[i].Key.Date.Before(aggs[j].Key.Date)
		}
		if aggs[i].Key.UTMCampaign != aggs[j].Key.UTMCampaign {
			return aggs[i].Key.UTMCampaign < aggs[j].Key.UTMCampaign
		}
		if aggs[i].Key.UTMSource != aggs[j].Key.UTMSource {
			return aggs[i].Key.UTMSource < aggs[j].Key.UTMSource
		}
		return aggs[i].Key.UTMMedium < aggs[j].Key.UTMMedium
	})

	rows := toMetricsSlice(aggs)
	limit, offset = clampLimitOffset(limit, offset, len(rows))
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
	v, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return v
}
func clampLimitOffset(limit, offset, n int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = n
	}
	if limit > 1000 {
		limit = 1000
	} // tope sano
	if offset > n {
		offset = n
	}
	return limit, offset
}
func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }
func round3(f float64) float64 { return float64(int64(f*1000+0.5)) / 1000 }
