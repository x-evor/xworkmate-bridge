package acp

import "testing"

func TestResolveSingleAgentForwardEndpoint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		provider syncedProvider
		want     string
	}{
		{
			name: "preserves upstream endpoint",
			provider: syncedProvider{
				ProviderID: "opencode",
				Endpoint:   "https://acp-server.svc.plus/opencode/acp/rpc",
			},
			want: "https://acp-server.svc.plus/opencode/acp/rpc",
		},
		{
			name: "does not rewrite bridge endpoint placeholder for codex",
			provider: syncedProvider{
				ProviderID: "codex",
				Endpoint:   "https://xworkmate-bridge.svc.plus",
			},
			want: "https://xworkmate-bridge.svc.plus",
		},
		{
			name: "does not rewrite bridge endpoint placeholder for gemini",
			provider: syncedProvider{
				ProviderID: "gemini",
				Endpoint:   "https://xworkmate-bridge.svc.plus",
			},
			want: "https://xworkmate-bridge.svc.plus",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveSingleAgentForwardEndpoint(tc.provider); got != tc.want {
				t.Fatalf("resolveSingleAgentForwardEndpoint() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeAuthorizationHeader(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":                 "",
		"Bearer bridge":    "Bearer bridge",
		"bridge-token":     "Bearer bridge-token",
		"   bridge-token ": "Bearer bridge-token",
	}
	for raw, want := range cases {
		raw, want := raw, want
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			if got := normalizeAuthorizationHeader(raw); got != want {
				t.Fatalf("normalizeAuthorizationHeader(%q) = %q, want %q", raw, got, want)
			}
		})
	}
}
