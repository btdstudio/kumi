package token

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueVerifySession(t *testing.T) {
	tokens := New("test-secret")
	tok, err := tokens.IssueSession(SessionClaims{
		Email:       "user@example.com",
		DisplayName: "User",
		Theme:       "catppuccin-mocha",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-1",
		},
	})
	if err != nil {
		t.Fatalf("IssueSession failed: %v", err)
	}
	claims, err := tokens.VerifySession(tok)
	if err != nil {
		t.Fatalf("VerifySession failed: %v", err)
	}
	if claims.Email != "user@example.com" || claims.DisplayName != "User" || claims.Theme != "catppuccin-mocha" {
		t.Errorf("unexpected claims: %+v", claims)
	}
	if claims.Subject != "user-1" {
		t.Errorf("expected Subject user-1, got %q", claims.Subject)
	}
}

func TestIssueVerifyReset(t *testing.T) {
	tokens := New("test-secret")
	tok, err := tokens.IssueReset("user-1")
	if err != nil {
		t.Fatalf("IssueReset failed: %v", err)
	}
	userID, err := tokens.VerifyReset(tok)
	if err != nil {
		t.Fatalf("VerifyReset failed: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("expected user-1, got %q", userID)
	}
}

func TestIssueVerifyEmailConfirm(t *testing.T) {
	tokens := New("test-secret")
	tok, err := tokens.IssueEmailConfirm("user-1", "new@example.com")
	if err != nil {
		t.Fatalf("IssueEmailConfirm failed: %v", err)
	}
	claims, err := tokens.VerifyEmailConfirm(tok)
	if err != nil {
		t.Fatalf("VerifyEmailConfirm failed: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "new@example.com" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestCrossPurposeRejection(t *testing.T) {
	tokens := New("test-secret")
	resetTok, err := tokens.IssueReset("user-1")
	if err != nil {
		t.Fatalf("IssueReset failed: %v", err)
	}
	if _, err := tokens.VerifySession(SessionToken(resetTok)); err == nil {
		t.Error("expected reset token to fail session verification")
	}
	if _, err := tokens.VerifyEmailConfirm(EmailConfirmToken(resetTok)); err == nil {
		t.Error("expected reset token to fail email-confirm verification")
	}
}

func TestVerifySession_WrongSecret(t *testing.T) {
	tokens := New("secret-a")
	tok, err := tokens.IssueSession(SessionClaims{Email: "a@b.c"})
	if err != nil {
		t.Fatalf("IssueSession failed: %v", err)
	}
	other := New("secret-b")
	if _, err := other.VerifySession(tok); err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestVerifySession_Expired(t *testing.T) {
	tokens := New("test-secret")
	claims := SessionClaims{
		Email:   "a@b.c",
		Purpose: purposeSession,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(DeriveKey("test-secret", purposeSession))
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if _, err := tokens.VerifySession(SessionToken(s)); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifySession_Tampered(t *testing.T) {
	tokens := New("test-secret")
	tok, err := tokens.IssueSession(SessionClaims{Email: "a@b.c"})
	if err != nil {
		t.Fatalf("IssueSession failed: %v", err)
	}
	s := string(tok)
	last := s[len(s)-1]
	replacement := byte('A')
	if last == 'A' {
		replacement = 'B'
	}
	tampered := s[:len(s)-1] + string(replacement)
	if tampered == s {
		t.Fatal("tampered token is identical to the original")
	}
	if _, err := tokens.VerifySession(SessionToken(tampered)); err == nil {
		t.Error("expected error for tampered token")
	}
}

func TestVerifySession_AlgNoneRejected(t *testing.T) {
	tokens := New("test-secret")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"email":"a@b.c","purpose":"session","exp":9999999999}`))
	noneToken := header + "." + claims + "."
	if _, err := tokens.VerifySession(SessionToken(noneToken)); err == nil {
		t.Error("expected alg=none to be rejected")
	}
}

func TestVerifySession_AlgRS256Rejected(t *testing.T) {
	tokens := New("test-secret")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"email":"a@b.c","purpose":"session","exp":9999999999}`))
	rsToken := header + "." + claims + ".sig"
	if _, err := tokens.VerifySession(SessionToken(rsToken)); err == nil {
		t.Error("expected alg=RS256 to be rejected")
	}
}

func TestDeriveKey(t *testing.T) {
	secret := strings.Repeat("a", 32)
	a := DeriveKey(secret, "csrf")
	if !equalBytes(a, DeriveKey(secret, "csrf")) {
		t.Error("DeriveKey is not deterministic")
	}
	if equalBytes(a, DeriveKey(secret, "other")) {
		t.Error("different purposes must yield different keys")
	}
	if equalBytes(a, []byte(secret)) {
		t.Error("derived key must differ from the secret")
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
