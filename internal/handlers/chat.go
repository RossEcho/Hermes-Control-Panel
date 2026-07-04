package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
)

// pendingStreams maps sessionID → message, held until the SSE endpoint is opened.
var (
	pendingMu      sync.Mutex
	pendingStreams  = map[string]string{}
)

// ChatPage renders the main chat UI.
func (h *Handler) ChatPage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")

	sessions, err := h.adapter.ListSessions()
	if err != nil {
		log.Printf("ChatPage ListSessions: %v", err)
	}

	// Pick a default session if none specified
	if sessionID == "" && len(sessions) > 0 {
		// Sort-stable: just use the first available session (mock returns unsorted).
		sessionID = sessions[0].ID
	}

	var messages interface{}
	if sessionID != "" {
		msgs, err := h.adapter.GetSessionMessages(sessionID)
		if err == nil {
			messages = msgs
		}
	}

	models, _ := h.adapter.ListModels()
	activeModel := "—"
	for _, m := range models {
		if m.IsActive {
			activeModel = m.Name
			break
		}
	}

	h.renderTemplate(w, "chat.html", PageData{
		Title:  "Chat",
		Active: "chat",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Sessions":    sessions,
			"SessionID":   sessionID,
			"Messages":    messages,
			"ActiveModel": activeModel,
		},
	})
}

// ChatSend accepts a new chat message and registers it for the SSE stream.
// The client should open /api/chat/stream/{sessionID} immediately after.
func (h *Handler) ChatSend(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "bad form")
		return
	}
	sessionID := r.FormValue("session_id")
	message := r.FormValue("message")
	if sessionID == "" || message == "" {
		writeError(w, http.StatusBadRequest, "session_id and message are required")
		return
	}

	pendingMu.Lock()
	pendingStreams[sessionID] = message
	pendingMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "ok",
		"session_id": sessionID,
		"stream_url": fmt.Sprintf("/api/chat/stream/%s", sessionID),
	})
}

// ChatStream opens an SSE stream for the given session, consuming the pending
// message registered by ChatSend.
func (h *Handler) ChatStream(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")

	pendingMu.Lock()
	message, ok := pendingStreams[sessionID]
	if ok {
		delete(pendingStreams, sessionID)
	}
	pendingMu.Unlock()

	if !ok {
		writeError(w, http.StatusBadRequest, "no pending message for this session")
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	eventCh, err := h.adapter.StreamChat(ctx, sessionID, message)
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return
	}

	for event := range eventCh {
		payload := map[string]interface{}{
			"type":    event.Type,
			"content": event.Content,
		}
		if len(event.Meta) > 0 {
			payload["meta"] = event.Meta
		}
		sendSSE(w, flusher, event.Type, payload)
	}
}

// ChatStop cancels the streaming for a session.
func (h *Handler) ChatStop(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "bad form")
		return
	}
	sessionID := r.FormValue("session_id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session_id required")
		return
	}
	if err := h.adapter.StopChat(sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// SessionNew creates a new session. Accepts optional form fields:
// name, context, model, harness, workspace.
// Returns JSON so it can be called via fetch from JS.
func (h *Handler) SessionNew(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "bad form")
		return
	}
	name := r.FormValue("name")
	model := r.FormValue("model")

	if model != "" {
		_ = h.adapter.SwitchModel(model) // best-effort
	}

	sess, err := h.adapter.NewSession(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"session_id": sess.ID,
		"title":      sess.Title,
	})
}

// SessionsList returns the JSON list of sessions.
func (h *Handler) SessionsList(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.adapter.ListSessions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": sessions})
}

// SessionDelete removes a session via DELETE /api/sessions/{id}.
// On success it returns an empty 200, which HTMX uses to remove the element.
func (h *Handler) SessionDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return
	}
	if err := h.adapter.DeleteSession(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UsageJSON returns usage stats as JSON.
func (h *Handler) UsageJSON(w http.ResponseWriter, r *http.Request) {
	usage, err := h.adapter.GetUsage()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, usage)
}

// UsageFragment returns a small HTML snippet for the sidebar usage section.
func (h *Handler) UsageFragment(w http.ResponseWriter, r *http.Request) {
	usage, err := h.adapter.GetUsage()
	if err != nil {
		http.Error(w, "usage error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="usage-stats">
  <div class="usage-row"><span class="usage-icon">↑</span><span class="usage-label">in</span><span class="usage-val">%s</span></div>
  <div class="usage-row"><span class="usage-icon">↓</span><span class="usage-label">out</span><span class="usage-val">%s</span></div>
  <div class="usage-row"><span class="usage-icon">◉</span><span class="usage-label">sessions</span><span class="usage-val">%s</span></div>
  <div class="usage-row"><span class="usage-icon">✉</span><span class="usage-label">today</span><span class="usage-val">%s</span></div>
</div>`,
		fmtTokens(usage.InputTokens),
		fmtTokens(usage.OutputTokens),
		strconv.Itoa(usage.ActiveSessions),
		strconv.Itoa(usage.TodayMessages),
	)
}

func fmtTokens(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return strconv.Itoa(n)
}

// sendSSE writes a single SSE event to the response.
func sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("sendSSE marshal: %v", err)
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, b)
	flusher.Flush()
}
