package store

import (
	"sync"
	"time"

	"github.com/AngelCh415/ELT_GO/internal/models"
)

type MemoryStore struct {
	mu   sync.RWMutex
	agg  map[models.DailyAggKey]*models.DailyAgg
	seen map[string]struct{} // idempotencia por-record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		agg:  make(map[models.DailyAggKey]*models.DailyAgg),
		seen: make(map[string]struct{}),
	}
}
func (s *MemoryStore) MarkSeen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return false
	}
	s.seen[key] = struct{}{}
	return true
}
func (s *MemoryStore) UpsertAds(a models.AdsPerformance) {
	k := models.DailyAggKey{Date: day(a.Date), Channel: a.Channel, CampaignID: a.CampaignID, UTMCampaign: a.UTMCampaign, UTMSource: a.UTMSource, UTMMedium: a.UTMMedium}
	s.mu.Lock()
	defer s.mu.Unlock()
	agg, ok := s.agg[k]
	if !ok {
		agg = &models.DailyAgg{Key: k}
		s.agg[k] = agg
	}
	agg.Clicks += max0(a.Clicks)
	agg.Impressions += max0(a.Impressions)
	agg.Cost += maxf(a.Cost)
}

func (s *MemoryStore) UpsertCRM(o models.Opportunity) {
	k := models.DailyAggKey{Date: day(o.CreatedAt), Channel: "", CampaignID: "", UTMCampaign: o.UTMCampaign, UTMSource: o.UTMSource, UTMMedium: o.UTMMedium}
	s.mu.Lock()
	defer s.mu.Unlock()
	agg, ok := s.agg[k]
	if !ok {
		agg = &models.DailyAgg{Key: k}
		s.agg[k] = agg
	}
	agg.Leads += 1
	switch o.Stage {
	case "opportunity":
		agg.Opportunities += 1
	case "closed_won":
		agg.Opportunities += 1
		agg.ClosedWon += 1
		agg.Revenue += o.Amount
	}
}

func (s *MemoryStore) All() []models.DailyAgg {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.DailyAgg, 0, len(s.agg))
	for _, v := range s.agg {
		out = append(out, *v)
	}
	return out
}

func (s *MemoryStore) Query(from, to time.Time, f func(models.DailyAgg) bool) []models.DailyAgg {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []models.DailyAgg
	for _, v := range s.agg {
		if !v.Key.Date.Before(from) && !v.Key.Date.After(to) {
			if f == nil || f(*v) {
				out = append(out, *v)
			}
		}
	}
	return out
}

func day(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
func max0(i int) int {
	if i < 0 {
		return 0
	}
	return i
}
func maxf(f float64) float64 {
	if f < 0 {
		return 0
	}
	return f
}
