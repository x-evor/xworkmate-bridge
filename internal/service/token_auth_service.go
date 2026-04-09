package service

import "strings"

type StaticTokenAuthService struct {
	expectedToken string
}

func NewStaticTokenAuthService(expectedToken string) *StaticTokenAuthService {
	return &StaticTokenAuthService{
		expectedToken: strings.TrimSpace(expectedToken),
	}
}

func (s *StaticTokenAuthService) ValidateToken(token string) bool {
	token = strings.TrimSpace(token)
	return token != "" && token == s.expectedToken
}
