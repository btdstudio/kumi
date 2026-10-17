package infrastructure

import (
	"testing"
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
