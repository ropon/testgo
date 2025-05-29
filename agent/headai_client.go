package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testgo/conf"
	"testgo/models"
	"time"
)

// HeadaiClient is a client for interacting with the Headai API.
type HeadaiClient struct {
	HTTPClient *http.Client
}

// NewHeadaiClient creates a new HeadaiClient.
func NewHeadaiClient() *HeadaiClient {
	return &HeadaiClient{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Example timeout
		},
	}
}

// GetAdRecommendations fetches ad recommendations from the Headai API.
func (c *HeadaiClient) GetAdRecommendations(req models.AdPlacementRequest) (*models.AdPlacementResponse, error) {
	// Marshal the AdPlacementRequest into JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Get the endpoint from configuration
	// As per subtask: conf.Cfg.External["HeadaiApiEndpoint"]
	endpoint, ok := conf.Cfg.External["HeadaiApiEndpoint"]
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("HeadaiApiEndpoint not found or empty in configuration External map")
	}

	// Create a new HTTP POST request
	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	// Set the Content-Type header
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Unmarshal the HTTP response body into an AdPlacementResponse struct
	var adResponse models.AdPlacementResponse
	if err := json.NewDecoder(resp.Body).Decode(&adResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &adResponse, nil
}
