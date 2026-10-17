package middleware

import (
	"net/http"

	gocsrf "github.com/gorilla/csrf"
)

func CSRFProtect(secretKey []byte, secure bool) func(http.Handler) http.Handler {
	return gocsrf.Protect(
		secretKey,
		gocsrf.Secure(secure),
		gocsrf.Path("/"),
		gocsrf.SameSite(gocsrf.SameSiteLaxMode),
		gocsrf.FieldName("csrf_token"),
	)
}

func CSRFToken(r *http.Request) string {
	return gocsrf.Token(r)
}
