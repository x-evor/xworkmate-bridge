package acp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleBridgeBootstrapConsumeReturnsSetupCode(t *testing.T) {
	accounts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/xworkmate/bridge/bootstrap/consume" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("X-Service-Token"); got != "internal-test-token" {
			http.Error(w, "missing service token", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(accountsBridgeBootstrapConsumeResponse{
			TicketID:      "ticket-1",
			TargetBridge:  "https://xworkmate-bridge.svc.plus",
			OpenclawURL:   "wss://openclaw.svc.plus",
			AuthMode:      "shared-token",
			ExchangeToken: "shared-token-value",
			ExpiresAt:     "2026-04-10T00:00:00Z",
			Scopes:        []string{"connect", "pairing.bootstrap"},
		})
	}))
	defer accounts.Close()

	t.Setenv("ACCOUNTS_BASE_URL", accounts.URL)
	t.Setenv("INTERNAL_SERVICE_TOKEN", "internal-test-token")
	t.Setenv("BRIDGE_PUBLIC_BASE_URL", "https://xworkmate-bridge.svc.plus")

	server := NewServer()
	body := bytes.NewBufferString(`{"ticket":"ticket-1","bridge":"https://xworkmate-bridge.svc.plus"}`)
	request := httptest.NewRequest(http.MethodPost, "/bridge/bootstrap/consume", body)
	recorder := httptest.NewRecorder()

	server.HandleBridgeBootstrapConsume(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected consume success, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var payload bridgeBootstrapResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.SetupCode == "" {
		t.Fatalf("expected non-empty setup code")
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(payload.SetupCode), &decoded); err != nil {
		t.Fatalf("decode setup code payload: %v", err)
	}
	if decoded["url"] != "wss://openclaw.svc.plus" {
		t.Fatalf("expected openclaw url in setup payload, got %#v", decoded)
	}
	if decoded["token"] != "shared-token-value" {
		t.Fatalf("expected exchange token in setup payload, got %#v", decoded)
	}
}
