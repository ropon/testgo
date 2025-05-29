package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testgo/agent"
	"testgo/conf"
	"testgo/models"
)

// MockAgent is a mock implementation of agent.AgentInterface
type MockAgent struct {
	GetAdPlacementFunc func(req models.AdPlacementRequest) (*models.AdPlacementResponse, error)
}

func (m *MockAgent) GetAdPlacement(req models.AdPlacementRequest) (*models.AdPlacementResponse, error) {
	if m.GetAdPlacementFunc != nil {
		return m.GetAdPlacementFunc(req)
	}
	return nil, errors.New("MockAgent.GetAdPlacementFunc is not set")
}

func TestAdPlacementHandler(t *testing.T) {
	// Backup original createAgentFunc and restore after test
	originalCreateAgentFunc := createAgentFunc
	defer func() {
		createAgentFunc = originalCreateAgentFunc
	}()

	// Minimal config setup if needed by any underlying component accessed via default createAgentFunc
	// For most tests here, createAgentFunc will be overridden.
	if conf.Cfg.External == nil {
		conf.Cfg.External = make(map[string]string)
	}
	conf.Cfg.External["HeadaiApiEndpoint"] = "http://dummy.endpoint.for.real.client/if/not/mocked"

	mockSuccessResponse := &models.AdPlacementResponse{
		RequestID:    "mock-req-id",
		AdCreativeID: "mock-creative-id",
		BidAmount:    0.75,
		TargetingParameters: map[string]string{"test": "value"},
		EstimatedPerformance: models.EstimatedPerformance{CTR: 0.1, ConversionRate: 0.02},
	}
	mockAgentError := errors.New("mock agent error")

	tests := []struct {
		name               string
		method             string
		path               string
		body               interface{}
		setupMockAgent     func() agent.AgentInterface
		expectedStatusCode int
		expectedBody       string // Substring to check in body, or full JSON for success
		expectedContentType string
	}{
		{
			name:   "Valid Request (POST)",
			method: http.MethodPost,
			path:   "/api/v1/adplacement",
			body: models.AdPlacementRequest{
				UserID: "test-user", Context: "homepage", CampaignID: "camp1",
			},
			setupMockAgent: func() agent.AgentInterface {
				return &MockAgent{
					GetAdPlacementFunc: func(req models.AdPlacementRequest) (*models.AdPlacementResponse, error) {
						return mockSuccessResponse, nil
					},
				}
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"request_id":"mock-req-id","ad_creative_id":"mock-creative-id","bid_amount":0.75,"targeting_parameters":{"test":"value"},"estimated_performance":{"ctr":0.1,"conversion_rate":0.02}}`,
			expectedContentType: "application/json",
		},
		{
			name:   "Invalid HTTP Method (GET)",
			method: http.MethodGet,
			path:   "/api/v1/adplacement",
			body:   nil,
			setupMockAgent: func() agent.AgentInterface { // Agent won't be called
				return &MockAgent{}
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedBody:       "Only POST method is allowed",
		},
		{
			name:   "Invalid JSON Payload",
			method: http.MethodPost,
			path:   "/api/v1/adplacement",
			body:   "this is not json", // Invalid JSON string
			setupMockAgent: func() agent.AgentInterface { // Agent won't be called
				return &MockAgent{}
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Error decoding request",
		},
		{
			name:   "Agent Error",
			method: http.MethodPost,
			path:   "/api/v1/adplacement",
			body: models.AdPlacementRequest{
				UserID: "error-user", Context: "productpage", CampaignID: "camp2",
			},
			setupMockAgent: func() agent.AgentInterface {
				return &MockAgent{
					GetAdPlacementFunc: func(req models.AdPlacementRequest) (*models.AdPlacementResponse, error) {
						return nil, mockAgentError
					},
				}
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Error processing request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the createAgentFunc to return our mock for this test case
			createAgentFunc = tt.setupMockAgent

			var reqBodyBytes []byte
			var err error
			if tt.body != nil {
				if strBody, ok := tt.body.(string); ok {
					reqBodyBytes = []byte(strBody)
				} else {
					reqBodyBytes, err = json.Marshal(tt.body)
					if err != nil {
						t.Fatalf("Failed to marshal request body for test %s: %v", tt.name, err)
					}
				}
			}

			req, err := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(reqBodyBytes))
			if err != nil {
				t.Fatalf("Failed to create request for test %s: %v", tt.name, err)
			}
			if tt.method == http.MethodPost && tt.body != nil && !strings.Contains(tt.body.(string), "this is not json") { // quick check to avoid setting content type for bad json string test
                 req.Header.Set("Content-Type", "application/json")
            }


			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(adPlacementHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Response body: %s", tt.expectedStatusCode, rr.Code, rr.Body.String())
			}

			if tt.expectedContentType != "" {
				contentType := rr.Header().Get("Content-Type")
				if contentType != tt.expectedContentType {
					t.Errorf("Expected Content-Type '%s', got '%s'", tt.expectedContentType, contentType)
				}
			}
			
			// For exact JSON match, unmarshal and compare, or use a JSON comparison library.
			// For substring, strings.Contains is fine.
			if tt.expectedStatusCode == http.StatusOK && tt.expectedContentType == "application/json" {
				// Try to unmarshal both expected and actual to compare content meaningfully
				var expectedResp models.AdPlacementResponse
				if err := json.Unmarshal([]byte(tt.expectedBody), &expectedResp); err != nil {
					t.Fatalf("Failed to unmarshal expectedBody for test %s: %v", tt.name, err)
				}
				var actualResp models.AdPlacementResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &actualResp); err != nil {
					t.Errorf("Failed to unmarshal actual response body for test %s: %v. Body: %s", tt.name, err, rr.Body.String())
				} else {
					// Basic comparison, for more complex structs, use reflect.DeepEqual or a dedicated library
					if expectedResp.RequestID != actualResp.RequestID || expectedResp.AdCreativeID != actualResp.AdCreativeID {
						t.Errorf("Response body mismatch for test %s.\nExpected: %+v\nActual:   %+v", tt.name, expectedResp, actualResp)
					}
				}
			} else if tt.expectedBody != "" && !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
