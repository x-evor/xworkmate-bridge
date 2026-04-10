package acp

import (
	"strings"
)

type syncedProvider struct {
	ProviderID          string
	Label               string
	Endpoint            string
	AuthorizationHeader string
	Enabled             bool
}

func parseSyncedProviders(raw any) []syncedProvider {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	providers := make([]syncedProvider, 0, len(list))
	for _, item := range list {
		entry := asMap(item)
		providerID := strings.TrimSpace(sharedString(entry, "providerId"))
		if providerID == "" {
			continue
		}
		providers = append(providers, syncedProvider{
			ProviderID:          providerID,
			Label:               strings.TrimSpace(sharedString(entry, "label")),
			Endpoint:            strings.TrimSpace(sharedString(entry, "endpoint")),
			AuthorizationHeader: strings.TrimSpace(sharedString(entry, "authorizationHeader")),
			Enabled:             parseBool(entry["enabled"]),
		})
	}
	return providers
}

func (s *Server) syncProviders(providers []syncedProvider) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providerCatalog = make(map[string]syncedProvider, len(providers))
	s.providerOrder = make([]string, 0, len(providers))
	for _, provider := range providers {
		providerID := strings.TrimSpace(provider.ProviderID)
		if providerID == "" {
			continue
		}
		provider.ProviderID = providerID
		s.providerCatalog[providerID] = provider
		s.providerOrder = append(s.providerOrder, providerID)
	}
	return map[string]any{
		"ok":        true,
		"providers": syncedProvidersResult(providers),
	}
}

func (s *Server) syncedProviderByID(providerID string) (syncedProvider, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	provider, ok := s.providerCatalog[strings.TrimSpace(providerID)]
	if !ok || !provider.Enabled || strings.TrimSpace(provider.Endpoint) == "" {
		return syncedProvider{}, false
	}
	return provider, true
}

func (s *Server) availableProviders() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ordered := make([]string, 0, len(s.providerOrder))
	for _, providerID := range s.providerOrder {
		provider, ok := s.providerCatalog[providerID]
		if !ok {
			continue
		}
		if !provider.Enabled || strings.TrimSpace(provider.Endpoint) == "" {
			continue
		}
		ordered = append(ordered, provider.ProviderID)
	}
	return ordered
}

func (s *Server) availableProviderCatalog() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]map[string]any, 0, len(s.providerOrder))
	for _, providerID := range s.providerOrder {
		provider, ok := s.providerCatalog[providerID]
		if !ok {
			continue
		}
		if !provider.Enabled || strings.TrimSpace(provider.Endpoint) == "" {
			continue
		}
		result = append(result, map[string]any{
			"providerId": provider.ProviderID,
			"label":      providerLabel(provider),
		})
	}
	return result
}

func syncedProvidersResult(providers []syncedProvider) []map[string]any {
	result := make([]map[string]any, 0, len(providers))
	for _, provider := range providers {
		result = append(result, map[string]any{
			"providerId": provider.ProviderID,
			"label":      providerLabel(provider),
			"endpoint":   provider.Endpoint,
			"enabled":    provider.Enabled,
		})
	}
	return result
}

func providerLabel(provider syncedProvider) string {
	if label := strings.TrimSpace(provider.Label); label != "" {
		return label
	}
	return provider.ProviderID
}
