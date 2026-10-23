package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/uptrace/bun"

	"kumi/internal/models"
	"kumi/internal/token"
)

type ContextKey string

const UserKey ContextKey = "user"
const UserProfileKey ContextKey = "user_profile"
const SessionCookieName = "session"

func UserFromContext(ctx context.Context) *token.SessionClaims {
	c, _ := ctx.Value(UserKey).(*token.SessionClaims)
	return c
}

func ContextWithClaims(ctx context.Context, claims *token.SessionClaims) context.Context {
	return context.WithValue(ctx, UserKey, claims)
}

func ContextWithUserProfile(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserProfileKey, user)
}

func UserProfileFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(UserProfileKey).(*models.User)
	return u
}

func LoadUserProfile(db *bun.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := UserFromContext(r.Context())
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			user := new(models.User)
			if err := db.NewSelect().Model(user).Where("id = ?", claims.Subject).Scan(r.Context()); err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := ContextWithUserProfile(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWTAuth(tokens *token.Tokens, secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := tokens.VerifySession(token.SessionToken(cookie.Value))
			if err != nil {
				ClearSessionCookie(w, secureCookies)
				next.ServeHTTP(w, r)
				return
			}

			ctx := ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Redirect(w, r, "/auth/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   14 * 24 * 3600,
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// IsHTTPS reports whether baseURL uses the https scheme.
func IsHTTPS(baseURL string) bool {
	return strings.HasPrefix(strings.ToLower(baseURL), "https://")
}
