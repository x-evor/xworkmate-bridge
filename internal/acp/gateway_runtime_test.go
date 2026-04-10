package acp

import (
	"testing"

	"xworkmate-bridge/internal/gatewayruntime"
)

func TestResolveGatewayReportedRemoteAddressUsesSyncedOpenClawEndpoint(t *testing.T) {
	t.Parallel()

	server := NewServer()
	server.syncProviders([]syncedProvider{
		{
			ProviderID: "openclaw",
			Label:      "OpenClaw",
			Endpoint:   "wss://gateway.example.com",
			Enabled:    true,
		},
	})

	got := resolveGatewayReportedRemoteAddress(server, gatewayruntime.ConnectRequest{
		Mode: "remote",
		Endpoint: gatewayruntime.Endpoint{
			Host: "127.0.0.1",
			Port: 18789,
			TLS:  false,
		},
	})

	const want = "gateway.example.com:443"
	if got != want {
		t.Fatalf("resolveGatewayReportedRemoteAddress() = %q, want %q", got, want)
	}
}

func TestResolveGatewayReportedRemoteAddressPreservesExplicitPublicRemoteHost(t *testing.T) {
	t.Parallel()

	server := NewServer()

	got := resolveGatewayReportedRemoteAddress(server, gatewayruntime.ConnectRequest{
		Mode: "remote",
		Endpoint: gatewayruntime.Endpoint{
			Host: "openclaw.svc.plus",
			Port: 443,
			TLS:  true,
		},
	})

	if got != "" {
		t.Fatalf("expected explicit public remote host to bypass override, got %q", got)
	}
}
