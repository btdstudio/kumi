package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	sessionTTL      = 14 * 24 * time.Hour
	resetTTL        = 1 * time.Hour
	emailConfirmTTL = 1 * time.Hour

	purposeSession      = "session"
	purposeReset        = "reset"
	purposeEmailConfirm = "email-confirm"
)

// Tokens issues and verifies purpose-typed tokens. Each purpose is signed with
// its own HMAC-SHA256 subkey derived from a single secret, so a token issued
// for one purpose can never be verified as another.
type Tokens struct {
	sessionKey []byte
	resetKey   []byte
	emailKey   []byte
}

// New derives one subkey per purpose from secret. The raw secret is never used
// to sign a token directly.
func New(secret string) *Tokens {
	return &Tokens{
		sessionKey: DeriveKey(secret, purposeSession),
		resetKey:   DeriveKey(secret, purposeReset),
		emailKey:   DeriveKey(secret, purposeEmailConfirm),
	}
}

// DeriveKey returns a purpose-specific subkey so a single secret is never used
// directly for more than one cryptographic purpose.
func DeriveKey(secret, purpose string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(purpose))
	return mac.Sum(nil)
}

// Opaque token types make cross-purpose use a compile error.
type SessionToken string
type ResetToken string
type EmailConfirmToken string

type SessionClaims struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	Theme       string `json:"theme,omitempty"`
	Purpose     string `json:"purpose"`
	jwt.RegisteredClaims
}

type EmailConfirmClaims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type resetClaims struct {
	UserID  string `json:"user_id"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func (t *Tokens) IssueSession(c SessionClaims) (SessionToken, error) {
	now := time.Now()
	c.Purpose = purposeSession
	c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(sessionTTL))
	c.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	s, err := token.SignedString(t.sessionKey)
	return SessionToken(s), err
}

func (t *Tokens) VerifySession(tok SessionToken) (*SessionClaims, error) {
	claims := &SessionClaims{}
	if err := t.parse(string(tok), claims, t.sessionKey); err != nil {
		return nil, err
	}
	if claims.Purpose != purposeSession {
		return nil, fmt.Errorf("invalid token purpose")
	}
	return claims, nil
}

func (t *Tokens) IssueReset(userID string) (ResetToken, error) {
	now := time.Now()
	claims := resetClaims{
		UserID:  userID,
		Purpose: purposeReset,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(resetTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(t.resetKey)
	return ResetToken(s), err
}

func (t *Tokens) VerifyReset(tok ResetToken) (string, error) {
	claims := &resetClaims{}
	if err := t.parse(string(tok), claims, t.resetKey); err != nil {
		return "", err
	}
	if claims.Purpose != purposeReset {
		return "", fmt.Errorf("invalid token purpose")
	}
	return claims.UserID, nil
}

func (t *Tokens) IssueEmailConfirm(userID, email string) (EmailConfirmToken, error) {
	now := time.Now()
	claims := EmailConfirmClaims{
		UserID:  userID,
		Email:   email,
		Purpose: purposeEmailConfirm,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(emailConfirmTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(t.emailKey)
	return EmailConfirmToken(s), err
}

func (t *Tokens) VerifyEmailConfirm(tok EmailConfirmToken) (*EmailConfirmClaims, error) {
	claims := &EmailConfirmClaims{}
	if err := t.parse(string(tok), claims, t.emailKey); err != nil {
		return nil, err
	}
	if claims.Purpose != purposeEmailConfirm {
		return nil, fmt.Errorf("invalid token purpose")
	}
	return claims, nil
}

func (t *Tokens) parse(tokenString string, claims jwt.Claims, key []byte) error {
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithLeeway(30*time.Second),
		jwt.WithIssuedAt(),
		jwt.WithExpirationRequired(),
	)
	return err
}
