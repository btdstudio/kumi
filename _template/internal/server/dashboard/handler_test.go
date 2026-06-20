package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"goph/middleware"
)

func TestDashboardPage_Unauthenticated(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dashboard", nil)

	h.DashboardPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDashboardPage_Authenticated(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dashboard", nil)

	claims := &middleware.Claims{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Theme:       "catppuccin-mocha",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-1",
		},
	}
	ctx := middleware.ContextWithClaims(r.Context(), claims)
	r = r.WithContext(ctx)

	h.DashboardPage(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Error("expected body to contain 'Dashboard'")
	}
	if !strings.Contains(body, "Test User") {
		t.Error("expected body to contain display name")
	}
	if !strings.Contains(body, "catppuccin-mocha") {
		t.Error("expected body to contain theme")
	}
}
