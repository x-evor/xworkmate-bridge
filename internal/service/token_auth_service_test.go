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

func TestStaticTokenAuthServiceValidateAuthorizationHeaderAsBearer(t *testing.T) {
	svc := NewStaticTokenAuthService("")
	if !svc.ValidateAuthorizationHeader("Bearer test-token") {
		t.Fatal("expected bearer header to be accepted")
	}
	if svc.ValidateAuthorizationHeader("Basic abc") {
		t.Fatal("expected non-bearer header to be rejected")
	}
	if svc.ValidateAuthorizationHeader("Bearer   ") {
		t.Fatal("expected empty bearer token to be rejected")
	}
}
