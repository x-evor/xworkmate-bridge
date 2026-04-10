package acp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xworkmate-bridge/internal/shared"
)

func TestCapabilitiesIgnoreLocalProviderAutodetectUntilSync(t *testing.T) {
	fakeProvider := t.TempDir() + "/fake-claude"
	if err := os.WriteFile(fakeProvider, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake provider: %v", err)
	}
	t.Setenv("ACP_CLAUDE_BIN", fakeProvider)
	t.Setenv("ACP_CODEX_BIN", "")
	t.Setenv("ACP_GEMINI_BIN", "")
	t.Setenv("ACP_OPENCODE_BIN", "")

	server := NewServer()
	result, rpcErr := server.handleRequest(shared.RPCRequest{
		Method: "acp.capabilities",
		Params: map[string]any{},
	}, func(map[string]any) {})
	if rpcErr != nil {
		t.Fatalf("expected capabilities success, got %v", rpcErr)
	}

	providers, _ := result["providers"].([]string)
	found := false
	for _, provider := range providers {
		if provider == "claude" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected autodetected local provider before sync, got %#v", providers)
	}
}

func TestProvidersSyncUpdatesCapabilities(t *testing.T) {
	server := NewServer()

	_, rpcErr := server.handleRequest(shared.RPCRequest{
		Method: "xworkmate.providers.sync",
		Params: map[string]any{
			"providers": []any{
				map[string]any{
					"providerId":          "claude",
					"label":               "Claude",
					"endpoint":            "http://127.0.0.1:9999",
					"authorizationHeader": "Bearer test",
					"enabled":             true,
				},
			},
		},
	}, func(map[string]any) {})
	if rpcErr != nil {
		t.Fatalf("expected sync success, got %v", rpcErr)
	}

	result, rpcErr := server.handleRequest(shared.RPCRequest{
		Method: "acp.capabilities",
		Params: map[string]any{},
	}, func(map[string]any) {})
	if rpcErr != nil {
		t.Fatalf("expected capabilities success, got %v", rpcErr)
	}
	providers, _ := result["providers"].([]string)
	if len(providers) == 0 {
		t.Fatalf("expected synced provider in capabilities, got %#v", result)
	}
	found := false
	for _, provider := range providers {
		if provider == "claude" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected claude provider after sync, got %#v", providers)
	}
}

func TestExecuteSessionTaskUsesSyncedExternalProvider(t *testing.T) {
	var lastForwardedParams map[string]any
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acp/rpc" {
			http.NotFound(w, r)
			return
		}
		defer r.Body.Close()
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		lastForwardedParams = asMap(request["params"])
		method, _ := request["method"].(string)
		switch method {
		case "session.start":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result": map[string]any{
					"success":  true,
					"output":   "external-provider-ok",
					"turnId":   "turn-external",
					"provider": "claude",
					"mode":     "single-agent",
				},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result":  map[string]any{"ok": true},
			})
		}
	}))
	defer externalServer.Close()

	server := NewServer()
	server.syncProviders([]syncedProvider{
		{
			ProviderID:          "claude",
			Label:               "Claude",
			Endpoint:            externalServer.URL,
			AuthorizationHeader: "Bearer test",
			Enabled:             true,
		},
	})

	response, rpcErr := server.executeSessionTask(task{
		req: shared.RPCRequest{
			Method: "session.start",
			Params: map[string]any{
				"sessionId":        "session-external",
				"threadId":         "thread-external",
				"taskPrompt":       "hello from external provider",
				"workingDirectory": t.TempDir(),
				"routing": map[string]any{
					"routingMode":             "explicit",
					"explicitExecutionTarget": "singleAgent",
					"explicitProviderId":      "claude",
				},
			},
		},
	})
	if rpcErr != nil {
		t.Fatalf("expected success, got rpc error: %v", rpcErr)
	}
	if got := response["output"]; got != "external-provider-ok" {
		t.Fatalf("expected external provider output, got %#v", response)
	}
	if got := response["resolvedProviderId"]; got != "claude" {
		t.Fatalf("expected resolved provider claude, got %#v", response)
	}
	if _, exists := lastForwardedParams["metadata"]; exists {
		t.Fatalf("expected metadata to be stripped for external provider request, got %#v", lastForwardedParams)
	}
	if _, exists := lastForwardedParams[externalProviderEndpointKey]; exists {
		t.Fatalf("expected internal endpoint key to be stripped, got %#v", lastForwardedParams)
	}
}

func TestExecuteSessionTaskEnrichesExternalProviderResultWithArtifactsAndRemoteMetadata(t *testing.T) {
	workingDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workingDir, "outputs"), 0o755); err != nil {
		t.Fatalf("mkdir outputs: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(workingDir, "outputs", "report.txt"),
		[]byte("artifact-body"),
		0o644,
	); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acp/rpc" {
			http.NotFound(w, r)
			return
		}
		defer r.Body.Close()
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"result": map[string]any{
				"success":                true,
				"output":                 "external-provider-ok",
				"turnId":                 "turn-external-artifacts",
				"provider":               "claude",
				"mode":                   "single-agent",
				"resolvedWorkingDirectory": "/remote/threads/task-42",
				"resolvedWorkspaceRefKind": "remotePath",
			},
		})
	}))
	defer externalServer.Close()

	server := NewServer()
	server.syncProviders([]syncedProvider{
		{
			ProviderID:          "claude",
			Label:               "Claude",
			Endpoint:            externalServer.URL,
			AuthorizationHeader: "Bearer test",
			Enabled:             true,
		},
	})

	response, rpcErr := server.executeSessionTask(task{
		req: shared.RPCRequest{
			Method: "session.start",
			Params: map[string]any{
				"sessionId":        "session-external-artifacts",
				"threadId":         "thread-external-artifacts",
				"taskPrompt":       "hello from external provider",
				"workingDirectory": workingDir,
				"routing": map[string]any{
					"routingMode":             "explicit",
					"explicitExecutionTarget": "singleAgent",
					"explicitProviderId":      "claude",
				},
			},
		},
	})
	if rpcErr != nil {
		t.Fatalf("expected success, got rpc error: %v", rpcErr)
	}
	if got := response["remoteWorkingDirectory"]; got != "/remote/threads/task-42" {
		t.Fatalf("expected remoteWorkingDirectory to be preserved, got %#v", got)
	}
	if got := response["remoteWorkspaceRefKind"]; got != "remotePath" {
		t.Fatalf("expected remoteWorkspaceRefKind remotePath, got %#v", got)
	}
	artifacts, ok := response["artifacts"].([]map[string]any)
	if !ok || len(artifacts) == 0 {
		t.Fatalf("expected enriched artifacts, got %#v", response["artifacts"])
	}
	artifact := artifacts[0]
	if got := artifact["relativePath"]; got != "outputs/report.txt" {
		t.Fatalf("expected relativePath outputs/report.txt, got %#v", got)
	}
	if got := artifact["content"]; got != "artifact-body" {
		t.Fatalf("expected inline artifact content, got %#v", got)
	}
	if got := artifact["encoding"]; got != "utf8" {
		t.Fatalf("expected utf8 artifact encoding, got %#v", got)
	}
	remoteExecution, ok := response["remoteExecution"].(map[string]any)
	if !ok {
		t.Fatalf("expected remoteExecution metadata, got %#v", response["remoteExecution"])
	}
	if got := remoteExecution["remoteWorkingDirectory"]; got != "/remote/threads/task-42" {
		t.Fatalf("expected remoteExecution remoteWorkingDirectory, got %#v", got)
	}
}

func TestRunSingleAgentUsesFrozenExternalProviderParams(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acp/rpc" {
			http.NotFound(w, r)
			return
		}
		defer r.Body.Close()
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      request["id"],
			"result": map[string]any{
				"success":  true,
				"output":   "frozen-provider-ok",
				"turnId":   "turn-frozen",
				"provider": "custom-agent-1",
				"mode":     "single-agent",
			},
		})
	}))
	defer externalServer.Close()

	server := NewServer()
	session := server.getOrCreateSession("session-frozen", "thread-frozen")
	result := server.runSingleAgent(
		context.Background(),
		"session.start",
		session,
		map[string]any{
			"provider":                             "custom-agent-1",
			"taskPrompt":                           "hello",
			"workingDirectory":                     t.TempDir(),
			externalProviderEndpointKey:            externalServer.URL,
			externalProviderAuthorizationHeaderKey: "Bearer test",
			externalProviderLabelKey:               "Codex",
		},
		"turn-frozen",
		func(map[string]any) {},
	)
	if result.err != nil {
		t.Fatalf("expected success, got rpc error: %v", result.err)
	}
	if got := result.response["output"]; got != "frozen-provider-ok" {
		t.Fatalf("expected frozen provider output, got %#v", result.response)
	}
}

func TestRunSingleAgentFallsBackWorkingDirectoryToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	fakeOpencode := filepath.Join(t.TempDir(), "opencode")
	if err := os.WriteFile(fakeOpencode, []byte("#!/bin/sh\necho local-ok\n"), 0o755); err != nil {
		t.Fatalf("write fake opencode: %v", err)
	}
	t.Setenv("ACP_OPENCODE_BIN", fakeOpencode)

	server := NewServer()
	session := server.getOrCreateSession("session-local", "thread-local")
	result := server.runSingleAgent(
		context.Background(),
		"session.start",
		session,
		map[string]any{
			"provider":         "opencode",
			"taskPrompt":       "hello",
			"workingDirectory": filepath.Join(t.TempDir(), "missing"),
		},
		"turn-local",
		func(map[string]any) {},
	)
	if result.err != nil {
		t.Fatalf("expected success, got rpc error: %v", result.err)
	}
	if got := result.response["output"]; got != "local-ok" {
		t.Fatalf("expected local provider output, got %#v", result.response)
	}
	if got := result.response["effectiveWorkingDirectory"]; got != home {
		t.Fatalf("expected effectiveWorkingDirectory %q, got %#v", home, got)
	}
}

func TestHandleRPCForwardsInboundBearerToExternalProvider(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer bridge-token" {
			t.Fatalf("expected forwarded bearer header, got %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      "run-auth",
			"result": map[string]any{
				"success": true,
				"output":  "forwarded-auth-ok",
			},
		})
	}))
	defer externalServer.Close()

	server := NewServer()
	server.syncProviders([]syncedProvider{{
		ProviderID: "codex",
		Label:      "Codex",
		Endpoint:   externalServer.URL,
		Enabled:    true,
	}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"http://127.0.0.1/acp/rpc",
		strings.NewReader(`{"jsonrpc":"2.0","id":"run-auth","method":"session.start","params":{"sessionId":"s1","threadId":"t1","taskPrompt":"hello","workingDirectory":"`+t.TempDir()+`","routing":{"routingMode":"explicit","explicitExecutionTarget":"singleAgent","explicitProviderId":"codex"}}}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer bridge-token")

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "forwarded-auth-ok") {
		t.Fatalf("expected forwarded provider response, got %q", recorder.Body.String())
	}
}

func TestExternalACPNotificationCollectorSynthesizesOutputAndWorkspace(t *testing.T) {
	collector := &externalACPNotificationCollector{}
	collector.observe(map[string]any{
		"jsonrpc": "2.0",
		"method":  "session.update",
		"params": map[string]any{
			"sessionId":                "session-streamed",
			"threadId":                 "thread-streamed",
			"turnId":                   "turn-streamed",
			"type":                     "delta",
			"delta":                    "streamed external output",
			"resolvedWorkingDirectory": "/tmp/thread-streamed",
			"pending":                  false,
			"error":                    false,
		},
	})

	result := collector.apply(map[string]any{
		"success": true,
	})

	if got := result["output"]; got != "streamed external output" {
		t.Fatalf("expected synthesized output from notifications, got %#v", result)
	}
	if got := result["summary"]; got != "streamed external output" {
		t.Fatalf("expected synthesized summary from notifications, got %#v", result)
	}
	if got := result["turnId"]; got != "turn-streamed" {
		t.Fatalf("expected synthesized turnId, got %#v", result)
	}
	if got := result["resolvedWorkingDirectory"]; got != "/tmp/thread-streamed" {
		t.Fatalf("expected synthesized working directory, got %#v", result)
	}
}
