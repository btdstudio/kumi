package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/uptrace/bun"

	"goph/db"
	"goph/internal/infrastructure"
	"goph/internal/server/auth"
	"goph/internal/server/dashboard"
	"goph/internal/server/settings"
	"goph/middleware"
)

type Config struct {
	DatabaseURL  string
	SecretKey    string
	ResendAPIKey string
	MailFromAddr string
}

func loadConfig() Config {
	return Config{
		DatabaseURL:  requireEnv("DATABASE_URL"),
		SecretKey:    requireEnv("SECRET_KEY"),
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		MailFromAddr: os.Getenv("MAIL_FROM_ADDRESS"),
	}
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return val
}

type App struct {
	Router       chi.Router
	Config       Config
	DB           *bun.DB
	SQLDB        *sql.DB
	EmailService infrastructure.EmailService
}

func NewApp() (*App, error) {
	cfg := loadConfig()

	bdb, sqldb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(sqldb, "migrations"); err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(securityHeaders)
	r.Use(middleware.CSRFProtect([]byte(cfg.SecretKey), false))
	r.Use(middleware.JWTAuth(cfg.SecretKey))

	var emailSvc infrastructure.EmailService
	if cfg.ResendAPIKey != "" {
		emailSvc = infrastructure.NewResendEmailService(cfg.ResendAPIKey, cfg.MailFromAddr)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := sqldb.Ping(); err != nil {
			http.Error(w, "DB unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	app := &App{
		Router:       r,
		Config:       cfg,
		DB:           bdb,
		SQLDB:        sqldb,
		EmailService: emailSvc,
	}

	app.mountRoutes()

	return app, nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func (app *App) mountRoutes() {
	r := app.Router

	authHandler := auth.NewHandler(app.DB, app.SQLDB, app.Config.SecretKey, app.EmailService)
	dashHandler := dashboard.NewHandler(app.DB, app.SQLDB)
	settingsHandler := settings.NewHandler(app.DB, app.SQLDB, app.Config.SecretKey, app.EmailService)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.LoginPage)
		r.With(httprate.LimitByIP(5, 1*time.Minute)).Post("/login", authHandler.LoginPost)

		r.Get("/register", authHandler.RegisterPage)
		r.With(httprate.LimitByIP(2, 1*time.Minute)).Post("/register", authHandler.RegisterPost)

		r.Post("/logout", authHandler.LogoutPost)

		r.Get("/forgot-password", authHandler.ForgotPasswordPage)
		r.With(httprate.LimitByIP(2, 1*time.Minute)).Post("/forgot-password", authHandler.ForgotPasswordPost)

		r.Get("/reset-password", authHandler.ResetPasswordPage)
		r.With(httprate.LimitByIP(5, 1*time.Minute)).Post("/reset-password", authHandler.ResetPasswordPost)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Use(middleware.LoadUserProfile(app.DB))

		r.Mount("/settings", settingsHandler.Routes())

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		})

		r.Get("/dashboard", dashHandler.DashboardPage)
	})
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}


