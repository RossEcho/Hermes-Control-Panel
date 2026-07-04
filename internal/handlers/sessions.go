package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SessionsPage renders the session list page.
func (h *Handler) SessionsPage(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.adapter.ListSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderTemplate(w, "sessions.html", PageData{
		Title:  "Sessions",
		Active: "sessions",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Sessions": sessions,
		},
	})
}

// SessionGetJSON returns a single session as JSON.
func (h *Handler) SessionGetJSON(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := h.adapter.GetSession(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

// SessionMessagesJSON returns the messages for a session as JSON.
func (h *Handler) SessionMessagesJSON(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	msgs, err := h.adapter.GetSessionMessages(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": msgs})
}
