package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testgo/conf"
	"testgo/models"
)

func TestNewHeadaiClient(t *testing.T) {
	client := NewHeadaiClient()
	if client == nil {
		t.Error("NewHeadaiClient() returned nil")
	}
	if client.HTTPClient == nil {
		t.Error("NewHeadaiClient().HTTPClient is nil")
	}
}

func TestGetAdRecommendations(t *testing.T) {
	// Backup original config and restore after test
	originalExternalConf := conf.Cfg.External
	defer func() {
		conf.Cfg.External = originalExternalConf
	}()

	// Mock server handler
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check method and content type
		if r.Method != http.MethodPost {
			http.Error(w, "Expected POST", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Expected Content-Type: application/json", http.StatusUnsupportedMediaType)
			return
		}

		var req models.AdPlacementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return
		}

		// Simulate different scenarios based on request content for more granular tests
		// For this example, we'll use a simple "happy path" and a "server error" path.

		if req.UserID == "user_causes_error" {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if req.UserID == "user_bad_json_response" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"request_id": "123", "ad_creative_id": "abc", "bid_amount": 1.5, "targeting_parameters": {"key": "value"}, "estimated_performance": {"ctr": 0.05, "conversion_rate": 0.02}}garbage`)
			return
		}
		
		// Happy path response
		resp := models.AdPlacementResponse{
			RequestID:    "test-req-id",
			AdCreativeID: "creative-123",
			BidAmount:    0.05,
			TargetingParameters: map[string]string{"geo": "US"},
			EstimatedPerformance: models.EstimatedPerformance{
				CTR:            0.1,
				ConversionRate: 0.02,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("Mock server failed to encode response: %v", err)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer server.Close()

	// Set the config to use the mock server's URL
	if conf.Cfg.External == nil {
		conf.Cfg.External = make(map[string]string)
	}
	conf.Cfg.External["HeadaiApiEndpoint"] = server.URL

	client := NewHeadaiClient()

	tests := []struct {
		name          string
		request       models.AdPlacementRequest
		expectError   bool
		expectedErrorContains string
		checkResponse func(t *testing.T, resp *models.AdPlacementResponse)
	}{
		{
			name: "Successful recommendation",
			request: models.AdPlacementRequest{
				UserID:     "test-user",
				Context:    "homepage",
				CampaignID: "campaign-xyz",
			},
			expectError: false,
			checkResponse: func(t *testing.T, resp *models.AdPlacementResponse) {
				if resp == nil {
					t.Fatal("Expected non-nil response")
				}
				if resp.RequestID != "test-req-id" {
					t.Errorf("Expected RequestID 'test-req-id', got '%s'", resp.RequestID)
				}
				if resp.AdCreativeID != "creative-123" {
					t.Errorf("Expected AdCreativeID 'creative-123', got '%s'", resp.AdCreativeID)
				}
			},
		},
		{
			name: "Server error",
			request: models.AdPlacementRequest{
				UserID: "user_causes_error", // This will trigger server error in mock
			},
			expectError:   true,
			expectedErrorContains: "received non-OK HTTP status: 500 Internal Server Error",
		},
		{
			name: "Bad JSON response from server",
			request: models.AdPlacementRequest{
				UserID: "user_bad_json_response", // This will trigger bad JSON in mock
			},
			expectError: true,
			expectedErrorContains: "failed to unmarshal response",
		},
		{
			name: "Missing endpoint in config",
			request: models.AdPlacementRequest{
				UserID: "test-user-no-endpoint",
			},
			expectError: true,
			expectedErrorContains: "HeadaiApiEndpoint not found or empty",
			// This test case needs to run with the endpoint removed from config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Missing endpoint in config" {
				// Temporarily remove the endpoint for this specific test case
				originalEndpoint := conf.Cfg.External["HeadaiApiEndpoint"]
				delete(conf.Cfg.External, "HeadaiApiEndpoint")
				defer func() { conf.Cfg.External["HeadaiApiEndpoint"] = originalEndpoint }()
			}


			resp, err := client.GetAdRecommendations(tt.request)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected an error, but got nil")
				}
				if tt.expectedErrorContains != "" && !strings.Contains(err.Error(), tt.expectedErrorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErrorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if tt.checkResponse != nil {
					tt.checkResponse(t, resp)
				}
			}
		})
	}
}
