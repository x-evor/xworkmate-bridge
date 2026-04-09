package handler

import (
	"encoding/json"
	"net/http"

	"xworkmate-bridge/internal/service"
)

type TokenAuthHandler struct {
	service *service.StaticTokenAuthService
}

func NewTokenAuthHandler(service *service.StaticTokenAuthService) *TokenAuthHandler {
	return &TokenAuthHandler{service: service}
}

func (h *TokenAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	token := r.Header.Get("Authorization")
	if !h.service.ValidateAuthorizationHeader(token) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
