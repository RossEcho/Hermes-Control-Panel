package handlers

import (
	"net/http"
	"strconv"
)

// JobsPage renders the jobs, processes, and logs page.
func (h *Handler) JobsPage(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.adapter.ListJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	processes, err := h.adapter.ListProcesses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logs, err := h.adapter.GetLogs(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderTemplate(w, "jobs.html", PageData{
		Title:  "Jobs",
		Active: "jobs",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Jobs":      jobs,
			"Processes": processes,
			"Logs":      logs,
		},
	})
}

// JobsJSON returns jobs as JSON.
func (h *Handler) JobsJSON(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.adapter.ListJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"jobs": jobs})
}

// ProcessesJSON returns processes as JSON.
func (h *Handler) ProcessesJSON(w http.ResponseWriter, r *http.Request) {
	procs, err := h.adapter.ListProcesses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"processes": procs})
}

// LogsJSON returns log lines as JSON.
func (h *Handler) LogsJSON(w http.ResponseWriter, r *http.Request) {
	n := 50
	if nStr := r.URL.Query().Get("n"); nStr != "" {
		if parsed, err := strconv.Atoi(nStr); err == nil && parsed > 0 {
			n = parsed
		}
	}
	logs, err := h.adapter.GetLogs(n)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
}
