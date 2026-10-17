package dashboard

import (
	"database/sql"
	"net/http"

	"github.com/uptrace/bun"

	"goph/internal/views"
	"goph/middleware"
)

type Handler struct {
	DB    *bun.DB
	SQLDB *sql.DB
}

func NewHandler(db *bun.DB, sqldb *sql.DB) *Handler {
	return &Handler{DB: db, SQLDB: sqldb}
}

type pageData struct {
	Title           string
	PageTitle       string
	BsTheme         string
	Theme           string
	DisplayName     string
	Email           string
	AvatarURL       string
	CSRFToken       string
	NavExtraClass   string
	IsAuthenticated bool
}

func renderPage(w http.ResponseWriter, r *http.Request, data pageData) {
	views.Base(
		data.Title,
		data.PageTitle,
		data.BsTheme,
		data.Theme,
		data.DisplayName,
		data.Email,
		data.AvatarURL,
		data.CSRFToken,
		data.NavExtraClass,
		data.IsAuthenticated,
	).Render(r.Context(), w)
}

func (h *Handler) userPageData(r *http.Request, title, pageTitle, navExtraClass string) pageData {
	data := pageData{
		Title:         title,
		PageTitle:     pageTitle,
		NavExtraClass: navExtraClass,
	}

	if claims := middleware.UserFromContext(r.Context()); claims != nil {
		data.IsAuthenticated = true
		data.Email = claims.Email
		data.DisplayName = claims.DisplayName
		data.Theme = claims.Theme
		data.CSRFToken = middleware.CSRFToken(r)

		if profile := middleware.UserProfileFromContext(r.Context()); profile != nil {
			data.AvatarURL = profile.AvatarURL
		}

		switch claims.Theme {
		case "catppuccin-latte", "solarized-light":
			data.BsTheme = "light"
		default:
			data.BsTheme = "dark"
		}
	}

	return data
}

func (h *Handler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	data := h.userPageData(r, "Dashboard — goph", "Dashboard", " active")
	renderPage(w, r, data)
}
