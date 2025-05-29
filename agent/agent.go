package agent

import (
	"testgo/models"
)

// Agent orchestrates ad placement logic using a HeadaiClient.
type Agent struct {
	headaiClient HeadaiAPIClient // Use the interface type
}

// NewAgent creates a new Agent with the given HeadaiClient.
func NewAgent(client HeadaiAPIClient) *Agent { // Use the interface type
	// It's good practice to check for nil interface
	if client == nil {
		// Depending on how critical this is, you might panic,
		// return an error, or return an agent that behaves in a specific "null" way.
		// For now, let's assume a valid client is required.
		panic("HeadaiAPIClient cannot be nil in NewAgent")
	}
	return &Agent{
		headaiClient: client,
	}
}

// GetAdPlacement processes an ad placement request.
// It currently calls the HeadaiClient's GetAdRecommendations method directly.
func (a *Agent) GetAdPlacement(req models.AdPlacementRequest) (*models.AdPlacementResponse, error) {
	// Call the GetAdRecommendations method of the headaiClient
	resp, err := a.headaiClient.GetAdRecommendations(req)
	if err != nil {
		// Handle errors returned by the client
		return nil, err
	}

	// For now, directly return the response from the headaiClient.
	// In a real scenario, more complex business logic, filtering,
	// or result modification might happen here.
	return resp, nil
}
