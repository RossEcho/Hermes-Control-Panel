// Package handlers wires HTTP routes and shared handler infrastructure.
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/RossEcho/hermes-control-panel/internal/config"
	"github.com/RossEcho/hermes-control-panel/internal/hermes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	adapter   hermes.Adapter
	cfg       *config.Config
	templates *template.Template
}

// New creates a Handler and parses all HTML templates.
func New(adapter hermes.Adapter, cfg *config.Config) (*Handler, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"join": func(sep string, items []string) string {
			result := ""
			for i, s := range items {
				if i > 0 {
					result += sep
				}
				result += s
			}
			return result
		},
		"maskSecret": func(s string) string {
			if s == "" {
				return "(not set)"
			}
			if len(s) <= 4 {
				return "****"
			}
			return s[:4] + strings.Repeat("*", len(s)-4)
		},
	}).ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{adapter: adapter, cfg: cfg, templates: tmpl}, nil
}

// Router builds and returns the complete chi router.
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Pages
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/chat", http.StatusFound)
	})
	r.Get("/chat", h.ChatPage)
	r.Get("/chat/{sessionID}", h.ChatPage)
	r.Get("/skills", h.SkillsPage)
	r.Get("/models", h.ModelsPage)
	r.Get("/sessions", h.SessionsPage)
	r.Get("/status", h.StatusPage)
	r.Get("/jobs", h.JobsPage)
	r.Get("/config", h.ConfigPage)

	// Chat API
	r.Post("/api/chat/send", h.ChatSend)
	r.Get("/api/chat/stream/{sessionID}", h.ChatStream)
	r.Post("/api/chat/stop", h.ChatStop)

	// Sessions API
	r.Post("/api/sessions/new", h.SessionNew)
	r.Get("/api/sessions", h.SessionsList)
	r.Get("/api/sessions/{id}", h.SessionGetJSON)
	r.Get("/api/sessions/{id}/messages", h.SessionMessagesJSON)
	r.Delete("/api/sessions/{id}", h.SessionDelete)

	// Usage API
	r.Get("/api/usage", h.UsageJSON)
	r.Get("/api/usage/fragment", h.UsageFragment)

	// Skills API
	r.Get("/api/skills", h.SkillsList)
	r.Get("/api/skills/{id}", h.SkillGet)

	// Models API
	r.Post("/api/models/select", h.ModelSelect)

	// System API
	r.Get("/api/status", h.StatusJSON)
	r.Get("/api/jobs", h.JobsJSON)
	r.Get("/api/processes", h.ProcessesJSON)
	r.Get("/api/logs", h.LogsJSON)

	// Health
	r.Get("/health", h.Health)

	// Static assets
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return r
}

// Health returns a simple JSON health response.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "mode": h.cfg.ActiveMode()})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %q: %v", name, err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// PageData is passed to every page template.
type PageData struct {
	Title   string
	Active  string // nav item: chat, skills, models, sessions, status, jobs, config
	Mode    string
	Payload interface{}
}
