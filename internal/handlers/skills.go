package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SkillsPage renders the skills browser page.
func (h *Handler) SkillsPage(w http.ResponseWriter, r *http.Request) {
	skills, err := h.adapter.ListSkills()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.renderTemplate(w, "skills.html", PageData{
		Title:  "Skills",
		Active: "skills",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Skills": skills,
		},
	})
}

// SkillsList returns the JSON list of all skills.
func (h *Handler) SkillsList(w http.ResponseWriter, r *http.Request) {
	skills, err := h.adapter.ListSkills()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"skills": skills})
}

// SkillGet returns a single skill by ID as JSON.
func (h *Handler) SkillGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	skill, err := h.adapter.GetSkill(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, skill)
}
