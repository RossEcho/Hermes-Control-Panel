package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RossEcho/hermes-control-panel/internal/config"
	"github.com/RossEcho/hermes-control-panel/internal/handlers"
	"github.com/RossEcho/hermes-control-panel/internal/hermes"
)

func main() {
	// Load configuration from .env and environment variables.
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Select the adapter based on config flags.
	var adapter hermes.Adapter
	switch {
	case cfg.EnableDirectCLIMode:
		log.Println("mode: CLI (stub — ErrNotImplemented will be returned)")
		adapter = hermes.NewCLIAdapter(
			cfg.HermesExecutablePath,
			cfg.HermesHome,
			cfg.HermesConfigPath,
		)
	case cfg.EnableAPIMode:
		log.Println("mode: API (stub — ErrNotImplemented will be returned)")
		adapter = hermes.NewAPIAdapter(cfg.HermesAPIBaseURL, cfg.HermesAPIToken)
	default:
		log.Println("mode: Mock (all data is simulated)")
		adapter = hermes.NewMockAdapter()
	}

	// Build HTTP handlers.
	h, err := handlers.New(adapter, cfg)
	if err != nil {
		log.Fatalf("handlers: %v", err)
	}

	addr := cfg.Addr()
	fmt.Printf("\n  Hermes Control Panel\n")
	fmt.Printf("  Mode   : %s\n", cfg.ActiveMode())
	fmt.Printf("  Listen : http://%s\n\n", addr)

	if err := http.ListenAndServe(addr, h.Router()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
