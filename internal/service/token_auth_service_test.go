package service

import "testing"

func TestStaticTokenAuthServiceValidateToken(t *testing.T) {
	svc := NewStaticTokenAuthService("secret")
	if !svc.ValidateToken("secret") {
		t.Fatal("expected valid token")
	}
	if svc.ValidateToken("wrong") {
		t.Fatal("expected invalid token")
	}
}
