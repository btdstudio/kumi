package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

func RateLimit(count int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitByIP(count, window)
}

var (
	LoginLimit        = RateLimit(5, 1*time.Minute)
	RegisterLimit     = RateLimit(2, 1*time.Minute)
	ForgotPasswordLimit = RateLimit(2, 1*time.Minute)
	ResetPasswordLimit  = RateLimit(5, 1*time.Minute)
	DefaultRateLimit  = RateLimit(30, 1*time.Minute)
)

func RateLimitByIP(next http.Handler) http.Handler {
	return httprate.LimitByIP(30, 1*time.Minute)(next)
}
