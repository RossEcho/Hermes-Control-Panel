package handlers

import (
	"net/http"
)

// StatusPage renders the system status dashboard.
func (h *Handler) StatusPage(w http.ResponseWriter, r *http.Request) {
	status, err := h.adapter.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, _ := h.adapter.ListSessions()
	h.renderTemplate(w, "status.html", PageData{
		Title:  "Status",
		Active: "status",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Status":   status,
			"Sessions": sessions,
		},
	})
}

// StatusJSON returns the system status as JSON.
func (h *Handler) StatusJSON(w http.ResponseWriter, r *http.Request) {
	status, err := h.adapter.GetStatus()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}
