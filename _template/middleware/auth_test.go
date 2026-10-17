package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionCookieSecureFlag(t *testing.T) {
	for _, secure := range []bool{true, false} {
		w := httptest.NewRecorder()
		SetSessionCookie(w, "tok", secure)
		ClearSessionCookie(httptest.NewRecorder(), secure)

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(cookies))
		}
		c := cookies[0]
		if c.Secure != secure {
			t.Errorf("Secure = %v, want %v", c.Secure, secure)
		}
		if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
			t.Errorf("cookie must be HttpOnly and SameSite=Lax: %+v", c)
		}
	}
}

func TestClearSessionCookieSecureFlag(t *testing.T) {
	w := httptest.NewRecorder()
	ClearSessionCookie(w, true)
	c := w.Result().Cookies()[0]
	if !c.Secure || c.MaxAge >= 0 {
		t.Errorf("expected expired Secure cookie, got %+v", c)
	}
}

func TestIsHTTPS(t *testing.T) {
	cases := map[string]bool{
		"https://example.com":   true,
		"HTTPS://example.com":   true,
		"http://localhost:8080": false,
		"":                      false,
	}
	for in, want := range cases {
		if got := IsHTTPS(in); got != want {
			t.Errorf("IsHTTPS(%q) = %v, want %v", in, got, want)
		}
	}
}
