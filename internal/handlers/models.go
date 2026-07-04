package handlers

import (
	"net/http"
)

// ModelsPage renders the model browser and switcher.
func (h *Handler) ModelsPage(w http.ResponseWriter, r *http.Request) {
	models, err := h.adapter.ListModels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group by provider
	type ProviderGroup struct {
		Provider string
		Models   interface{}
	}

	h.renderTemplate(w, "models.html", PageData{
		Title:  "Models",
		Active: "models",
		Mode:   h.cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"Models": models,
		},
	})
}

// ModelSelect switches the active model.
func (h *Handler) ModelSelect(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "bad form")
		return
	}
	modelID := r.FormValue("model_id")
	if modelID == "" {
		writeError(w, http.StatusBadRequest, "model_id required")
		return
	}
	if err := h.adapter.SwitchModel(modelID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Redirect back to models page
	http.Redirect(w, r, "/models", http.StatusSeeOther)
}
