package auth

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	"kumi/internal/infrastructure"
	"kumi/internal/models"
	"kumi/internal/token"
	"kumi/internal/views/auth"
	"kumi/middleware"
)

type Handler struct {
	DB           *bun.DB
	SQLDB        *sql.DB
	Tokens       *token.Tokens
	BaseURL      string
	EmailService infrastructure.EmailService
	Log          *log.Logger
}

func NewHandler(db *bun.DB, sqldb *sql.DB, tokens *token.Tokens, baseURL string, emailSvc infrastructure.EmailService) *Handler {
	return &Handler{
		DB:           db,
		SQLDB:        sqldb,
		Tokens:       tokens,
		BaseURL:      baseURL,
		EmailService: emailSvc,
		Log:          log.Default(),
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", h.LoginPage)
	r.Post("/login", h.LoginPost)

	r.Get("/register", h.RegisterPage)
	r.Post("/register", h.RegisterPost)

	r.Post("/logout", h.LogoutPost)

	r.Get("/forgot-password", h.ForgotPasswordPage)
	r.Post("/forgot-password", h.ForgotPasswordPost)

	r.Get("/reset-password", h.ResetPasswordPage)
	r.Post("/reset-password", h.ResetPasswordPost)

	return r
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	reset := r.URL.Query().Get("reset")
	auth.LoginPage(reset == "1", middleware.CSRFToken(r)).Render(r.Context(), w)
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Email and password are required.").Render(r.Context(), w)
		return
	}

	if !infrastructure.IsValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Please enter a valid email address.").Render(r.Context(), w)
		return
	}

	user := new(models.User)
	err := h.DB.NewSelect().Model(user).Where("email = ?", email).Scan(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		auth.Error("Invalid email or password.").Render(r.Context(), w)
		return
	}

	if !infrastructure.CheckPassword(password, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		auth.Error("Invalid email or password.").Render(r.Context(), w)
		return
	}

	token, err := h.createSessionToken(user)
	if err != nil {
		h.Log.Printf("failed to create session token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		auth.Error("An error occurred. Please try again.").Render(r.Context(), w)
		return
	}

	middleware.SetSessionCookie(w, token, middleware.IsHTTPS(h.BaseURL))
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	auth.RegisterPage(middleware.CSRFToken(r)).Render(r.Context(), w)
}

func (h *Handler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Email and password are required.").Render(r.Context(), w)
		return
	}

	if !infrastructure.IsValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Please enter a valid email address.").Render(r.Context(), w)
		return
	}

	strength := infrastructure.CheckPasswordStrength(password)
	if !strength.IsValid {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error(strength.Error()).Render(r.Context(), w)
		return
	}

	hash, err := infrastructure.HashPassword(password)
	if err != nil {
		h.Log.Printf("failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		auth.Error("An error occurred. Please try again.").Render(r.Context(), w)
		return
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  strings.Split(email, "@")[0],
	}

	if _, err := h.DB.NewInsert().Model(user).Returning("*").Exec(r.Context()); err != nil {
		auth.Error("An account with that email already exists.").Render(r.Context(), w)
		return
	}

	token, err := h.createSessionToken(user)
	if err != nil {
		h.Log.Printf("failed to create session token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		auth.Error("An error occurred. Please try again.").Render(r.Context(), w)
		return
	}

	middleware.SetSessionCookie(w, token, middleware.IsHTTPS(h.BaseURL))
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LogoutPost(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w, middleware.IsHTTPS(h.BaseURL))
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

func (h *Handler) ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	if middleware.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	auth.ForgotPasswordPage(middleware.CSRFToken(r)).Render(r.Context(), w)
}

func (h *Handler) ForgotPasswordPost(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))

	if email == "" || !infrastructure.IsValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Please enter a valid email address.").Render(r.Context(), w)
		return
	}

	user := new(models.User)
	err := h.DB.NewSelect().Model(user).Where("email = ?", email).Scan(r.Context())
	if err == nil {
		resetTok, tokErr := h.Tokens.IssueReset(user.ID)
		if tokErr != nil {
			h.Log.Printf("failed to create reset token: %v", tokErr)
		} else if h.EmailService != nil {
			resetURL := h.BaseURL + "/auth/reset-password?token=" + string(resetTok)
			if sendErr := h.EmailService.Send(user.Email, "Reset your kumi password",
				"<p>Click the link below to reset your password. This link expires in 1 hour.</p>"+
					"<p><a href=\""+resetURL+"\">"+resetURL+"</a></p>"+
					"<p>If you didn't request this, you can safely ignore this email.</p>",
			); sendErr != nil {
				h.Log.Printf("failed to send reset email: %v", sendErr)
			}
		} else {
			h.Log.Printf("email service not configured; skipping password reset email to %s", user.Email)
		}
	}

	auth.Error("If that email is registered you'll receive a reset link shortly.").Render(r.Context(), w)
}

func (h *Handler) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")

	userID, err := h.Tokens.VerifyReset(token.ResetToken(tokenStr))
	if err != nil || userID == "" {
		auth.ResetPasswordPage(false, "", "This reset link is invalid or has expired. Please request a new one.", "").Render(r.Context(), w)
		return
	}

	auth.ResetPasswordPage(true, tokenStr, "", middleware.CSRFToken(r)).Render(r.Context(), w)
}

func (h *Handler) ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.FormValue("token")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	strength := infrastructure.CheckPasswordStrength(password)
	if !strength.IsValid {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error(strength.Error()).Render(r.Context(), w)
		return
	}

	if password != confirm {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Passwords do not match.").Render(r.Context(), w)
		return
	}

	userID, err := h.Tokens.VerifyReset(token.ResetToken(tokenStr))
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		auth.Error("Reset link is invalid or expired. Please request a new one.").Render(r.Context(), w)
		return
	}

	user := new(models.User)
	if err := h.DB.NewSelect().Model(user).Where("id = ?", userID).Scan(r.Context()); err != nil {
		w.WriteHeader(http.StatusNotFound)
		auth.Error("User not found.").Render(r.Context(), w)
		return
	}

	hash, hashErr := infrastructure.HashPassword(password)
	if hashErr != nil {
		h.Log.Printf("failed to hash password: %v", hashErr)
		w.WriteHeader(http.StatusInternalServerError)
		auth.Error("An error occurred. Please try again.").Render(r.Context(), w)
		return
	}

	user.PasswordHash = hash
	if _, err := h.DB.NewUpdate().Model(user).Column("password_hash").WherePK().Exec(r.Context()); err != nil {
		h.Log.Printf("failed to update password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		auth.Error("An error occurred. Please try again.").Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", "/auth/login?reset=1")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createSessionToken(user *models.User) (string, error) {
	tok, err := h.Tokens.IssueSession(token.SessionClaims{
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Theme:       user.Theme,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
		},
	})
	return string(tok), err
}
