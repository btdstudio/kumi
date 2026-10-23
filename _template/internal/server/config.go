package server

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"kumi/middleware"
)

const minSecretKeyLength = 32

type Config struct {
	DatabaseURL  string
	SecretKey    string
	BaseURL      string
	TrustProxy   bool
	ResendAPIKey string
	MailFromAddr string
}

// SecureCookies reports whether cookies should carry the Secure attribute,
// which is the case whenever the app is served over https.
func (c Config) SecureCookies() bool {
	return middleware.IsHTTPS(c.BaseURL)
}

func loadConfig() (Config, error) {
	dbURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return Config{}, err
	}
	secretKey, err := requireEnv("SECRET_KEY")
	if err != nil {
		return Config{}, err
	}
	if err := validateSecretKey(secretKey); err != nil {
		return Config{}, err
	}
	rawBaseURL, err := requireEnv("BASE_URL")
	if err != nil {
		return Config{}, err
	}
	baseURL, err := validateBaseURL(rawBaseURL)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL:  dbURL,
		SecretKey:    secretKey,
		BaseURL:      baseURL,
		TrustProxy:   os.Getenv("TRUST_PROXY") == "true",
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		MailFromAddr: os.Getenv("MAIL_FROM_ADDRESS"),
	}, nil
}

func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return val, nil
}

func validateSecretKey(key string) error {
	if len(key) < minSecretKeyLength {
		return fmt.Errorf("SECRET_KEY must be at least %d characters (generate one with: openssl rand -hex 32)", minSecretKeyLength)
	}
	if strings.Contains(strings.ToLower(key), "change-this") {
		return fmt.Errorf("SECRET_KEY is still set to a placeholder value (generate one with: openssl rand -hex 32)")
	}
	return nil
}

// validateBaseURL checks that raw is an absolute http(s) URL with no path,
// query or fragment, and returns it without a trailing slash.
func validateBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("BASE_URL must be an absolute http(s) URL such as https://example.com, got %q", raw)
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("BASE_URL must contain only scheme and host, got %q", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}
