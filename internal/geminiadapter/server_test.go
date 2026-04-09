package geminiadapter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"

	"xworkmate-bridge/internal/shared"
)

type stubClient struct {
	initResult initializeResult
	initErr    error
	callResult map[string]any
	callErr    error
	lastMethod string
	lastParams map[string]any
}

func (s *stubClient) Initialize() (initializeResult, error) {
	return s.initResult, s.initErr
}

func (s *stubClient) Call(method string, params map[string]any) (map[string]any, error) {
	s.lastMethod = method
	s.lastParams = params
	return s.callResult, s.callErr
}

func (s *stubClient) Close() error {
	return nil
}

func TestHandleCapabilitiesSynthesizesProviderResponse(t *testing.T) {
	server := NewServer(&stubClient{
		initResult: initializeResult{
			ProtocolVersion: 1,
			AuthMethods: []map[string]any{
				{"id": "gemini-api-key"},
			},
			AgentCapabilities: map[string]any{
				"mcpCapabilities": map[string]any{"http": true},
			},
		},
	})

	result := server.handleRequest(shared.RPCRequest{
		Method: "acp.capabilities",
		Params: map[string]any{},
	})
	if got := result["singleAgent"]; got != true {
		t.Fatalf("expected singleAgent true, got %#v", result)
	}
	providers, _ := result["providers"].([]string)
	if len(providers) != 1 || providers[0] != "gemini" {
		t.Fatalf("expected gemini provider, got %#v", result)
	}
}

func TestHandleRPCSessionStartReturnsUpstreamResult(t *testing.T) {
	stub := &stubClient{
		initResult: initializeResult{ProtocolVersion: 1},
		callResult: map[string]any{
			"result": map[string]any{
				"success": true,
				"output":  "hello",
			},
		},
	}
	server := NewServer(stub)

	body, _ := json.Marshal(shared.RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "session.start",
		Params: map[string]any{
			"taskPrompt": "hello",
		},
	})
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/acp/rpc", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.HandleRPC(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var envelope map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result := envelope["result"].(map[string]any)
	if got := result["output"]; got != "hello" {
		t.Fatalf("expected output hello, got %#v", result)
	}
	if stub.lastMethod != "session.start" {
		t.Fatalf("expected upstream method session.start, got %q", stub.lastMethod)
	}
}

func TestHandleWebSocketCapabilities(t *testing.T) {
	server := NewServer(&stubClient{
		initResult: initializeResult{ProtocolVersion: 1},
	})
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(w, r)
	}))
	defer httpServer.Close()

	wsURL := "ws" + httpServer.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(shared.RPCRequest{
		JSONRPC: "2.0",
		ID:      "cap-1",
		Method:  "acp.capabilities",
		Params:  map[string]any{},
	}); err != nil {
		t.Fatalf("write json: %v", err)
	}
	var envelope map[string]any
	if err := conn.ReadJSON(&envelope); err != nil {
		t.Fatalf("read json: %v", err)
	}
	result := envelope["result"].(map[string]any)
	providers := result["providers"].([]any)
	if len(providers) != 1 || providers[0] != "gemini" {
		t.Fatalf("expected gemini provider over websocket, got %#v", result)
	}
}
