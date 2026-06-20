package infrastructure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 12

var passwordPatterns = map[string]*regexp.Regexp{
	"has_uppercase": regexp.MustCompile(`[A-Z]`),
	"has_lowercase": regexp.MustCompile(`[a-z]`),
	"has_digit":     regexp.MustCompile(`\d`),
	"has_special":   regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;':",./<>?~]`),
}

type PasswordStrength struct {
	IsValid       bool
	HasMinLength  bool
	HasUppercase  bool
	HasLowercase  bool
	HasDigit      bool
	HasSpecial    bool
}

func CheckPasswordStrength(password string) PasswordStrength {
	var s PasswordStrength
	s.HasMinLength = len(password) >= minPasswordLength
	for _, r := range password {
		if unicode.IsUpper(r) {
			s.HasUppercase = true
		}
		if unicode.IsLower(r) {
			s.HasLowercase = true
		}
		if unicode.IsDigit(r) {
			s.HasDigit = true
		}
	}
	s.HasSpecial = passwordPatterns["has_special"].MatchString(password)
	s.IsValid = s.HasMinLength && s.HasUppercase && s.HasLowercase && s.HasDigit && s.HasSpecial
	return s
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(b), nil
}

func CheckPassword(plain, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

type timedToken struct {
	Data string `json:"data"`
	Exp  int64  `json:"exp"`
}

func CreateToken(data, secret string, ttl time.Duration) (string, error) {
	t := timedToken{
		Data: data,
		Exp:  time.Now().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("marshal token: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := mac.Sum(nil)
	combined := append(payload, sig...)
	return base64.URLEncoding.EncodeToString(combined), nil
}

func VerifyToken(token, secret string) (string, error) {
	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	if len(raw) < sha256.Size {
		return "", fmt.Errorf("token too short")
	}
	payload := raw[:len(raw)-sha256.Size]
	sig := raw[len(raw)-sha256.Size:]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return "", fmt.Errorf("invalid signature")
	}

	var t timedToken
	if err := json.Unmarshal(payload, &t); err != nil {
		return "", fmt.Errorf("unmarshal token: %w", err)
	}
	if time.Now().Unix() > t.Exp {
		return "", fmt.Errorf("token expired")
	}
	return t.Data, nil
}
