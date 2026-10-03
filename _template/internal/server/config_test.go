package server

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateSecretKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"too short", "short", true},
		{"31 chars", strings.Repeat("a", 31), true},
		{"placeholder", "change-this-to-a-random-secret-key-please", true},
		{"32 chars", strings.Repeat("a", 32), false},
		{"hex 32 bytes", strings.Repeat("ab", 32), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSecretKey(tt.key); (err != nil) != tt.wantErr {
				t.Errorf("validateSecretKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://example.com", "https://example.com", false},
		{"https://example.com/", "https://example.com", false},
		{"http://localhost:8080", "http://localhost:8080", false},
		{"example.com", "", true},
		{"ftp://example.com", "", true},
		{"https://", "", true},
		{"https://example.com/app", "", true},
		{"https://example.com?x=1", "", true},
		{"https://user:pw@example.com", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := validateBaseURL(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateBaseURL(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("validateBaseURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDeriveKey(t *testing.T) {
	secret := strings.Repeat("a", 32)
	a := deriveKey(secret, "csrf")
	if !bytes.Equal(a, deriveKey(secret, "csrf")) {
		t.Error("deriveKey is not deterministic")
	}
	if bytes.Equal(a, deriveKey(secret, "other")) {
		t.Error("different purposes must yield different keys")
	}
	if bytes.Equal(a, []byte(secret)) {
		t.Error("derived key must differ from the secret")
	}
}
