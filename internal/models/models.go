package models

import "time"

type AdsPerformance struct {
	Date        time.Time
	CampaignID  string
	Channel     string
	Clicks      int
	Impressions int
	Cost        float64
	UTMCampaign string
	UTMSource   string
	UTMMedium   string
}

type Opportunity struct {
	OpportunityID string
	ContactEmail  string
	Stage         string // e.g., lead, opportunity, closed_won, closed_lost
	Amount        float64
	CreatedAt     time.Time
	UTMCampaign   string
	UTMSource     string
	UTMMedium     string
}
type DailyAggKey struct {
	Date        time.Time
	Channel     string
	CampaignID  string
	UTMCampaign string
	UTMSource   string
	UTMMedium   string
}
type DailyAgg struct {
	Key           DailyAggKey
	Clicks        int
	Impressions   int
	Cost          float64
	Leads         int
	Opportunities int
	ClosedWon     int
	Revenue       float64
}
type Metrics struct {
	Date          string  `json:"date"`
	Channel       string  `json:"channel"`
	CampaignID    string  `json:"campaign_id"`
	UTMCampaign   string  `json:"utm_campaign"`
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
	Clicks        int     `json:"clicks"`
	Impressions   int     `json:"impressions"`
	Cost          float64 `json:"cost"`
	Leads         int     `json:"leads"`
	Opportunities int     `json:"opportunities"`
	ClosedWon     int     `json:"closed_won"`
	Revenue       float64 `json:"revenue"`
	CPC           float64 `json:"cpc"`
	CPA           float64 `json:"cpa"`
	CVRLeadToOpp  float64 `json:"cvr_lead_to_opp"`
	CVROppToWon   float64 `json:"cvr_opp_to_won"`
	ROAS          float64 `json:"roas"`
}
