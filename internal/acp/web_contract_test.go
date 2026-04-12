package acp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPHandlerRootAndPingExposeRuntimeVersionInfo(t *testing.T) {
	t.Setenv("IMAGE", "ghcr.io/x-evor/xworkmate-bridge:0123456789abcdef0123456789abcdef01234567")

	server := NewServer()
	handler := server.Handler()

	rootRecorder := httptest.NewRecorder()
	rootRequest := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	handler.ServeHTTP(rootRecorder, rootRequest)

	if rootRecorder.Code != http.StatusOK {
		t.Fatalf("expected root 200, got %d", rootRecorder.Code)
	}
	if !strings.Contains(rootRecorder.Body.String(), "xworkmate-bridge is running") {
		t.Fatalf("expected root body to contain service banner, got %q", rootRecorder.Body.String())
	}

	pingRecorder := httptest.NewRecorder()
	pingRequest := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/ping", nil)
	handler.ServeHTTP(pingRecorder, pingRequest)

	if pingRecorder.Code != http.StatusOK {
		t.Fatalf("expected ping 200, got %d", pingRecorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(pingRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode ping payload: %v", err)
	}

	if got := payload["status"]; got != "ok" {
		t.Fatalf("expected status ok, got %#v", got)
	}
	if got := payload["image"]; got != "ghcr.io/x-evor/xworkmate-bridge:0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("expected full image ref, got %#v", got)
	}
	if got := payload["tag"]; got != "0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("expected full image tag, got %#v", got)
	}
	if got := payload["commit"]; got != "0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("expected full image commit, got %#v", got)
	}
	if got := payload["version"]; got != "0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("expected full image version, got %#v", got)
	}
}

func TestParseImageVersionInfoHandlesTaggedImageRef(t *testing.T) {
	info := parseImageVersionInfo("ghcr.io/x-evor/xworkmate-bridge:main-2026-04-12")

	if info.ImageRef != "ghcr.io/x-evor/xworkmate-bridge:main-2026-04-12" {
		t.Fatalf("expected full image ref, got %q", info.ImageRef)
	}
	if info.Tag != "main-2026-04-12" {
		t.Fatalf("expected tag main-2026-04-12, got %q", info.Tag)
	}
	if info.Commit != "" {
		t.Fatalf("expected empty commit for non-hex tag, got %q", info.Commit)
	}
	if info.Version != "main-2026-04-12" {
		t.Fatalf("expected version main-2026-04-12, got %q", info.Version)
	}
}

func TestHandleWebSocketRejectsUnknownOrigin(t *testing.T) {
	t.Setenv("ACP_ALLOWED_ORIGINS", "https://xworkmate.svc.plus")

	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/acp", nil)
	request.Header.Set("Origin", "https://evil.example.com")

	server.HandleWebSocket(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected application/json content type, got %q", got)
	}
}

func TestHandleRPCAllowsPreflightForConfiguredOrigin(t *testing.T) {
	t.Setenv("ACP_ALLOWED_ORIGINS", "https://xworkmate.svc.plus,http://localhost:*")

	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1/acp/rpc", nil)
	request.Header.Set("Origin", "https://xworkmate.svc.plus")
	request.Header.Set("Access-Control-Request-Method", "POST")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://xworkmate.svc.plus" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
}

func TestHandleRPCRequiresBearerAuthorization(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"http://127.0.0.1/acp/rpc",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"acp.capabilities"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandleRPCRejectsUnknownOrigin(t *testing.T) {
	t.Setenv("ACP_ALLOWED_ORIGINS", "https://xworkmate.svc.plus")

	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"http://127.0.0.1/acp/rpc",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"acp.capabilities"}`),
	)
	request.Header.Set("Origin", "https://evil.example.com")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer test")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if _, ok := envelope["error"]; !ok {
		t.Fatalf("expected JSON-RPC error envelope, got %v", envelope)
	}
}

func TestHandleRPCMethodErrorUsesJSONEnvelope(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/acp/rpc", nil)
	request.Header.Set("Authorization", "Bearer test")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected application/json content type, got %q", got)
	}
}

func TestHandleRPCCapabilitiesStillReturnsJSONResult(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"http://127.0.0.1/acp/rpc",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"acp.capabilities"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer test")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected application/json content type, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), `"providerCatalog"`) {
		t.Fatalf("expected capabilities response, got %q", recorder.Body.String())
	}
}

func TestHandleWebSocketRequiresBearerAuthorization(t *testing.T) {
	t.Setenv("ACP_ALLOWED_ORIGINS", "https://xworkmate.svc.plus")

	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/acp", nil)
	request.Header.Set("Origin", "https://xworkmate.svc.plus")

	server.HandleWebSocket(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}
