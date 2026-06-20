package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	"goph/internal/models"
)

type ContextKey string

const UserKey ContextKey = "user"
const UserProfileKey ContextKey = "user_profile"
const SessionCookieName = "session"

type Claims struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	Theme       string `json:"theme,omitempty"`
	jwt.RegisteredClaims
}

func UserFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(UserKey).(*Claims)
	return c
}

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
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

func JWTAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			}, jwt.WithLeeway(30*time.Second))

			if err != nil || !token.Valid {
				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookieName,
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
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

func RequireAuthAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CreateSessionToken(secretKey string, claims *Claims) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.Subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(14 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   14 * 24 * 3600,
	})
}
