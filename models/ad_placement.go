package models

// Demographics represents user demographic information.
type Demographics struct {
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Location string `json:"location"`
}

// AdPlacementRequest represents a request to place an ad.
type AdPlacementRequest struct {
	UserID       string       `json:"user_id"`
	Context      string       `json:"context"`
	CampaignID   string       `json:"campaign_id"`
	Demographics Demographics `json:"demographics"`
	Interests    []string     `json:"interests"`
}

// EstimatedPerformance represents the estimated performance of an ad.
type EstimatedPerformance struct {
	CTR            float64 `json:"ctr"`
	ConversionRate float64 `json:"conversion_rate"`
}

// AdPlacementResponse represents the response for an ad placement request.
type AdPlacementResponse struct {
	RequestID           string            `json:"request_id"`
	AdCreativeID        string            `json:"ad_creative_id"`
	BidAmount           float64           `json:"bid_amount"`
	TargetingParameters map[string]string `json:"targeting_parameters"`
	EstimatedPerformance EstimatedPerformance `json:"estimated_performance"`
}
