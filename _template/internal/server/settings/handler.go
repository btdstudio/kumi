package settings

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	"goph/internal/infrastructure"
	"goph/internal/models"
	views "goph/internal/views/settings"
	"goph/middleware"
)

var themeOptions = []views.ThemeOption{
	{"catppuccin-frappe", "Catppuccin Frappé"},
	{"catppuccin-latte", "Catppuccin Latte"},
	{"catppuccin-macchiato", "Catppuccin Macchiato"},
	{"catppuccin-mocha", "Catppuccin Mocha"},
	{"dracula", "Dracula"},
	{"solarized-dark", "Solarized Dark"},
	{"solarized-light", "Solarized Light"},
}

var themes = func() map[string]string {
	m := make(map[string]string, len(themeOptions))
	for _, opt := range themeOptions {
		m[opt.Key] = opt.Label
	}
	return m
}()

type Handler struct {
	DB              *bun.DB
	SQLDB           *sql.DB
	SecretKey       string
	EmailService    infrastructure.EmailService
}

func NewHandler(db *bun.DB, sqldb *sql.DB, secretKey string, emailSvc infrastructure.EmailService) *Handler {
	return &Handler{
		DB:           db,
		SQLDB:        sqldb,
		SecretKey:    secretKey,
		EmailService: emailSvc,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.IndexPage)
	r.Post("/profile", h.ProfilePost)
	r.Post("/email", h.EmailPost)
	r.Get("/confirm-email/{token}", h.ConfirmEmail)
	r.Post("/password", h.PasswordPost)
	r.Post("/theme", h.ThemePost)
	r.Post("/delete", h.DeletePost)

	return r
}

func (h *Handler) IndexPage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	user := middleware.UserProfileFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	flash := r.URL.Query().Get("flash")
	msg := r.URL.Query().Get("msg")

	views.IndexPage(
		user.DisplayName,
		user.Email,
		user.PendingEmail,
		user.AvatarURL,
		user.Theme,
		themeOptions,
		middleware.CSRFToken(r),
		flash, msg,
	).Render(r.Context(), w)
}

func (h *Handler) ProfilePost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	displayName := strings.TrimSpace(r.FormValue("display_name"))

	r.ParseMultipartForm(5 << 20)
	file, header, err := r.FormFile("avatar")

	user := middleware.UserProfileFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		views.Error("User not found.").Render(r.Context(), w)
		return
	}

	if displayName != "" {
		user.DisplayName = displayName
	}

	if err == nil && header != nil && header.Filename != "" {
		defer file.Close()

		var buf bytes.Buffer
		size, _ := io.CopyN(&buf, file, 5<<20)
		if size > 2<<20 {
			w.WriteHeader(http.StatusBadRequest)
			views.Error("Avatar must be under 2 MB.").Render(r.Context(), w)
			return
		}

		mime := header.Header.Get("Content-Type")
		if mime == "" {
			mime = "image/png"
		}
		b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
		user.AvatarURL = "data:" + mime + ";base64," + b64
	}

	if _, err := h.DB.NewUpdate().Model(user).Column("display_name", "avatar_url").WherePK().Exec(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to update profile.").Render(r.Context(), w)
		return
	}

	newToken, err := middleware.CreateSessionToken(h.SecretKey, &middleware.Claims{
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Theme:       user.Theme,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
		},
	})
	if err != nil {
		log.Printf("CreateSessionToken failed: %v (user.ID=%q user.Email=%q)", err, user.ID, user.Email)
	} else {
		middleware.SetSessionCookie(w, newToken)
	}

	w.Header().Set("HX-Redirect", "/settings/?flash=success&msg=Profile+updated.")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) EmailPost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Email and password are required.").Render(r.Context(), w)
		return
	}

	if !infrastructure.IsValidEmail(email) {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Please enter a valid email address.").Render(r.Context(), w)
		return
	}

	user := middleware.UserProfileFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		views.Error("User not found.").Render(r.Context(), w)
		return
	}

	if !infrastructure.CheckPassword(password, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		views.Error("Password is incorrect.").Render(r.Context(), w)
		return
	}

	if email == user.Email {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("That is already your current email.").Render(r.Context(), w)
		return
	}

	existing := new(models.User)
	if err := h.DB.NewSelect().Model(existing).Where("email = ? AND id != ?", email, user.ID).Scan(r.Context()); err == nil {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("That email is already in use.").Render(r.Context(), w)
		return
	}

	user.PendingEmail = email
	if _, err := h.DB.NewUpdate().Model(user).Column("pending_email").WherePK().Exec(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to update email.").Render(r.Context(), w)
		return
	}

	token, err := CreateEmailConfirmToken(h.SecretKey, user.ID, email)
	if err == nil {
		confirmURL := "http://" + r.Host + "/settings/confirm-email/" + token
		SendConfirmationEmail(h.EmailService, email, confirmURL)
	}

	w.Header().Set("HX-Redirect", "/settings/?flash=info&msg=Confirmation+link+sent+—+check+your+new+email+inbox.")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	data, err := VerifyEmailConfirmToken(h.SecretKey, token)
	if err != nil || data == nil {
		views.ConfirmEmail(false, "This confirmation link is invalid or has expired. Please request a new one.").Render(r.Context(), w)
		return
	}

	user := new(models.User)
	if err := h.DB.NewSelect().Model(user).Where("id = ?", data.UserID).Scan(r.Context()); err != nil {
		views.ConfirmEmail(false, "User not found.").Render(r.Context(), w)
		return
	}

	if user.PendingEmail != data.Email {
		views.ConfirmEmail(false, "Confirmation link is no longer valid.").Render(r.Context(), w)
		return
	}

	user.Email = user.PendingEmail
	user.PendingEmail = ""
	if _, err := h.DB.NewUpdate().Model(user).Column("email", "pending_email").WherePK().Exec(r.Context()); err != nil {
		views.ConfirmEmail(false, "Failed to confirm email.").Render(r.Context(), w)
		return
	}

	views.ConfirmEmail(true, "Your email address has been confirmed.").Render(r.Context(), w)
}

func (h *Handler) PasswordPost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	currentPw := r.FormValue("current_password")
	newPw := r.FormValue("new_password")
	confirmPw := r.FormValue("confirm_password")

	if currentPw == "" || newPw == "" || confirmPw == "" {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("All password fields are required.").Render(r.Context(), w)
		return
	}

	user := middleware.UserProfileFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		views.Error("User not found.").Render(r.Context(), w)
		return
	}

	if !infrastructure.CheckPassword(currentPw, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		views.Error("Current password is incorrect.").Render(r.Context(), w)
		return
	}

	strength := infrastructure.CheckPasswordStrength(newPw)
	if !strength.IsValid {
		var missing []string
		if !strength.HasMinLength {
			missing = append(missing, "at least 12 characters")
		}
		if !strength.HasUppercase {
			missing = append(missing, "an uppercase letter")
		}
		if !strength.HasLowercase {
			missing = append(missing, "a lowercase letter")
		}
		if !strength.HasDigit {
			missing = append(missing, "a number")
		}
		if !strength.HasSpecial {
			missing = append(missing, "a special character")
		}
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Password must include " + strings.Join(missing, ", ") + ".").Render(r.Context(), w)
		return
	}

	if newPw != confirmPw {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("New passwords do not match.").Render(r.Context(), w)
		return
	}

	hash, err := infrastructure.HashPassword(newPw)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to update password.").Render(r.Context(), w)
		return
	}

	user.PasswordHash = hash
	if _, err := h.DB.NewUpdate().Model(user).Column("password_hash").WherePK().Exec(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to update password.").Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", "/settings/?flash=success&msg=Password+updated.")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ThemePost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	theme := strings.TrimSpace(r.FormValue("theme"))
	if _, ok := themes[theme]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Invalid theme.").Render(r.Context(), w)
		return
	}

	if _, err := h.DB.NewUpdate().Model(&models.User{}).Set("theme = ?", theme).Where("id = ?", claims.Subject).Exec(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to update theme.").Render(r.Context(), w)
		return
	}

	claims.Theme = theme
	newToken, err := middleware.CreateSessionToken(h.SecretKey, claims)
	if err == nil {
		middleware.SetSessionCookie(w, newToken)
	}

	w.Header().Set("HX-Trigger", "theme-updated")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserFromContext(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	confirm := strings.TrimSpace(r.FormValue("confirm"))
	password := r.FormValue("password")

	if password == "" {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Password is required.").Render(r.Context(), w)
		return
	}

	if confirm != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		views.Error("Type DELETE to confirm.").Render(r.Context(), w)
		return
	}

	user := middleware.UserProfileFromContext(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		views.Error("User not found.").Render(r.Context(), w)
		return
	}

	if !infrastructure.CheckPassword(password, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		views.Error("Password is incorrect.").Render(r.Context(), w)
		return
	}

	if _, err := h.DB.NewDelete().Model(user).WherePK().Exec(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		views.Error("Failed to delete account.").Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", "/auth/login")
	w.WriteHeader(http.StatusNoContent)
}
