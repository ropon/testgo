package agent

import "testgo/models"

// HeadaiAPIClient defines the interface for clients interacting with the Headai API.
// This allows for mocking in tests.
type HeadaiAPIClient interface {
	GetAdRecommendations(req models.AdPlacementRequest) (*models.AdPlacementResponse, error)
}

// AgentInterface defines the interface for the ad placement agent.
// This allows for mocking in tests.
type AgentInterface interface {
	GetAdPlacement(req models.AdPlacementRequest) (*models.AdPlacementResponse, error)
}
