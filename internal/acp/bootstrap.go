package acp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"xworkmate-bridge/internal/shared"
)

type bridgeBootstrapConsumeRequest struct {
	Ticket string `json:"ticket"`
	Bridge string `json:"bridge"`
}

type accountsBridgeBootstrapConsumeResponse struct {
	TicketID        string   `json:"ticketId"`
	TargetBridge    string   `json:"targetBridge"`
	BridgeServerURL string   `json:"BRIDGE_SERVER_URL"`
	AuthMode        string   `json:"authMode"`
	BridgeAuthToken string   `json:"BRIDGE_AUTH_TOKEN"`
	ExpiresAt       string   `json:"expiresAt"`
	Scopes          []string `json:"scopes"`
}

type bridgeBootstrapResponse struct {
	SetupCode    string   `json:"setupCode"`
	BridgeOrigin string   `json:"bridgeOrigin"`
	AuthMode     string   `json:"authMode"`
	ExpiresAt    string   `json:"expiresAt"`
	IssuedBy     string   `json:"issuedBy"`
	Scopes       []string `json:"scopes"`
}

func (s *Server) HandleBridgeBootstrapHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":           true,
		"bridgeOrigin": bridgePublicBaseURL(),
		"issuedBy":     "xworkmate-bridge",
	})
}

func (s *Server) HandleBridgeBootstrapConsume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req bridgeBootstrapConsumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Ticket = strings.TrimSpace(req.Ticket)
	req.Bridge = strings.TrimSpace(req.Bridge)
	if req.Ticket == "" {
		http.Error(w, "ticket is required", http.StatusBadRequest)
		return
	}
	if req.Bridge == "" {
		req.Bridge = bridgePublicBaseURL()
	}

	payload, status, err := consumeBootstrapFromAccounts(req)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	setupCodePayload := map[string]any{
		"url":               payload.BridgeServerURL,
		"token":             payload.BridgeAuthToken,
		"BRIDGE_SERVER_URL": payload.BridgeServerURL,
		"BRIDGE_AUTH_TOKEN": payload.BridgeAuthToken,
		"authMode":          payload.AuthMode,
		"expiresAt":         payload.ExpiresAt,
		"bridgeOrigin":      req.Bridge,
		"issuedBy":          "xworkmate-bridge",
	}
	setupCodeBytes, err := json.Marshal(setupCodePayload)
	if err != nil {
		http.Error(w, "failed to encode setup payload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bridgeBootstrapResponse{
		SetupCode:    string(setupCodeBytes),
		BridgeOrigin: req.Bridge,
		AuthMode:     payload.AuthMode,
		ExpiresAt:    payload.ExpiresAt,
		IssuedBy:     "xworkmate-bridge",
		Scopes:       append([]string(nil), payload.Scopes...),
	})
}

func consumeBootstrapFromAccounts(req bridgeBootstrapConsumeRequest) (accountsBridgeBootstrapConsumeResponse, int, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(shared.EnvOrDefault("ACCOUNTS_BASE_URL", "https://accounts.svc.plus")), "/")
	if baseURL == "" {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusInternalServerError, fmt.Errorf("accounts base url is not configured")
	}
	serviceToken := strings.TrimSpace(shared.EnvOrDefault("INTERNAL_SERVICE_TOKEN", ""))
	if serviceToken == "" {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusInternalServerError, fmt.Errorf("internal service token is not configured")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to encode consume request")
	}
	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/api/internal/xworkmate/bridge/bootstrap/consume", bytes.NewReader(body))
	if err != nil {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusInternalServerError, fmt.Errorf("failed to create consume request")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Service-Token", serviceToken)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusBadGateway, fmt.Errorf("failed to contact accounts service")
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return accountsBridgeBootstrapConsumeResponse{}, resp.StatusCode, fmt.Errorf("accounts bootstrap consume failed")
	}
	var payload accountsBridgeBootstrapConsumeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return accountsBridgeBootstrapConsumeResponse{}, http.StatusBadGateway, fmt.Errorf("failed to decode accounts bootstrap response")
	}
	return payload, http.StatusOK, nil
}

func bridgePublicBaseURL() string {
	value := strings.TrimSpace(shared.EnvOrDefault("BRIDGE_PUBLIC_BASE_URL", "https://xworkmate-bridge.svc.plus"))
	if value == "" {
		return "https://xworkmate-bridge.svc.plus"
	}
	return strings.TrimRight(value, "/")
}
