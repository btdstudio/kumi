package infrastructure

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
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

func (s PasswordStrength) Error() string {
	var missing []string
	if !s.HasMinLength {
		missing = append(missing, "at least 12 characters")
	}
	if !s.HasUppercase {
		missing = append(missing, "an uppercase letter")
	}
	if !s.HasLowercase {
		missing = append(missing, "a lowercase letter")
	}
	if !s.HasDigit {
		missing = append(missing, "a number")
	}
	if !s.HasSpecial {
		missing = append(missing, "a special character")
	}
	return "Password must include " + strings.Join(missing, ", ") + "."
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

func CreateToken(data, secret string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"data": data,
		"exp":  time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyToken(token, secret string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	data, _ := claims["data"].(string)
	if data == "" {
		return "", fmt.Errorf("invalid token")
	}
	return data, nil
}
