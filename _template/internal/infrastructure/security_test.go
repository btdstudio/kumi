package infrastructure

import (
	"testing"
	"time"
)

func TestCheckPasswordStrength_Valid(t *testing.T) {
	s := CheckPasswordStrength("HelloWorld123!")
	if !s.IsValid {
		t.Error("expected valid password")
	}
	if !s.HasMinLength {
		t.Error("expected HasMinLength")
	}
	if !s.HasUppercase {
		t.Error("expected HasUppercase")
	}
	if !s.HasLowercase {
		t.Error("expected HasLowercase")
	}
	if !s.HasDigit {
		t.Error("expected HasDigit")
	}
	if !s.HasSpecial {
		t.Error("expected HasSpecial")
	}
}

func TestCheckPasswordStrength_TooShort(t *testing.T) {
	s := CheckPasswordStrength("Ab1!")
	if s.IsValid {
		t.Error("expected invalid for short password")
	}
	if s.HasMinLength {
		t.Error("expected no HasMinLength for short password")
	}
}

func TestCheckPasswordStrength_MissingUppercase(t *testing.T) {
	s := CheckPasswordStrength("helloworld123!")
	if s.IsValid {
		t.Error("expected invalid without uppercase")
	}
	if s.HasUppercase {
		t.Error("expected no HasUppercase")
	}
}

func TestCheckPasswordStrength_MissingLowercase(t *testing.T) {
	s := CheckPasswordStrength("HELLOWORLD123!")
	if s.IsValid {
		t.Error("expected invalid without lowercase")
	}
	if s.HasLowercase {
		t.Error("expected no HasLowercase")
	}
}

func TestCheckPasswordStrength_MissingDigit(t *testing.T) {
	s := CheckPasswordStrength("HelloWorld!")
	if s.IsValid {
		t.Error("expected invalid without digit")
	}
	if s.HasDigit {
		t.Error("expected no HasDigit")
	}
}

func TestCheckPasswordStrength_MissingSpecial(t *testing.T) {
	s := CheckPasswordStrength("HelloWorld123")
	if s.IsValid {
		t.Error("expected invalid without special char")
	}
	if s.HasSpecial {
		t.Error("expected no HasSpecial")
	}
}

func TestHashPassword(t *testing.T) {
	plain := "mySecureP@ss1"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !CheckPassword(plain, hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword("wrongPassword1!", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestCreateAndVerifyToken(t *testing.T) {
	secret := "test-secret-key"
	data := "user@example.com"
	ttl := 1 * time.Hour

	token, err := CreateToken(data, secret, ttl)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}

	extracted, err := VerifyToken(token, secret)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if extracted != data {
		t.Errorf("expected %q, got %q", data, extracted)
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	token, err := CreateToken("test@example.com", "secret1", 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}
	_, err = VerifyToken(token, "secret2")
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	token, err := CreateToken("test@example.com", "secret", -1*time.Second)
	if err != nil {
		t.Fatalf("CreateToken failed: %v", err)
	}
	_, err = VerifyToken(token, "secret")
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestVerifyToken_InvalidBase64(t *testing.T) {
	_, err := VerifyToken("!!!invalid!!!", "secret")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestVerifyToken_TooShort(t *testing.T) {
	_, err := VerifyToken("aGVsbG8=", "secret")
	if err == nil {
		t.Error("expected error for too-short token")
	}
}
