package handlers

import (
	"net/http"
	"strings"
)

// maskSecret masks a secret string, showing only the first 4 characters.
func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}

// ConfigPage renders the read-only configuration display.
func (h *Handler) ConfigPage(w http.ResponseWriter, r *http.Request) {
	cfg := h.cfg
	h.renderTemplate(w, "config.html", PageData{
		Title:  "Config",
		Active: "config",
		Mode:   cfg.ActiveMode(),
		Payload: map[string]interface{}{
			"HermesExecutablePath": cfg.HermesExecutablePath,
			"HermesHome":           cfg.HermesHome,
			"HermesConfigPath":     cfg.HermesConfigPath,
			"HermesEnvPath":        cfg.HermesEnvPath,
			"HermesAPIBaseURL":     cfg.HermesAPIBaseURL,
			"HermesAPIToken":       maskSecret(cfg.HermesAPIToken),

			"AppHost":          cfg.AppHost,
			"AppPort":          cfg.AppPort,
			"AppSessionSecret": maskSecret(cfg.AppSessionSecret),
			"AppDebug":         cfg.AppDebug,

			"DefaultProfile":   cfg.DefaultProfile,
			"DefaultChatModel": cfg.DefaultChatModel,

			"EnableDirectCLIMode": cfg.EnableDirectCLIMode,
			"EnableAPIMode":       cfg.EnableAPIMode,
			"EnableMockMode":      cfg.EnableMockMode,

			"LogLevel": cfg.LogLevel,
		},
	})
}
