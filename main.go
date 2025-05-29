package main

import (
	"encoding/json"
	"fmt"
	"github.com/ropon/logger"
	"net/http"
	"testgo/agent"
	"testgo/conf"
	"testgo/models"
)

// createAgentFunc is a factory function for creating an agent.AgentInterface.
// This can be replaced in tests to inject a mock agent.
var createAgentFunc = func() agent.AgentInterface {
	headaiClient := agent.NewHeadaiClient() // Real HeadaiClient
	return agent.NewAgent(headaiClient)     // Real Agent
}

// @title work_api
// @version 1.0
// @description 后端快速Api脚手架

// @contact.name Ropon
// @contact.url https://www.ropon.top
// @contact.email ropon@xxx.com

// @license.name Apache 2.0
// @license.url https://www.apache.org/licenses/LICENSE-2.0.html

// @host work-api.xxx.com:2345
// @BasePath /
func main() {
	err := conf.Init()
	if err != nil {
		fmt.Printf("init failed, err: %v\n", err)
		return
	}

	logger.Info("Starting...")
	logger.Error("Error log demo...")

	http.HandleFunc("/", sayHello)
	http.HandleFunc("/api/v1/adplacement", adPlacementHandler)
	logger.Fatal("Server Start Fatal, Error: %s", http.ListenAndServe(conf.Cfg.Listen, nil))
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello CodoonOps！")
}

func adPlacementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AdPlacementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Use the factory to create an agent instance
	appAgent := createAgentFunc()

	resp, err := appAgent.GetAdPlacement(req)
	if err != nil {
		logger.Error("Error getting ad placement: %v", err)
		http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Error encoding response: %v", err)
		// If headers are already sent, this might not be effective, but good practice
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
