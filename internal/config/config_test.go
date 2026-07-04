package config_test

import (
	"os"
	"testing"

	"github.com/RossEcho/hermes-control-panel/internal/config"
)

func clearEnv() {
	vars := []string{
		"HERMES_EXECUTABLE_PATH", "HERMES_HOME", "HERMES_CONFIG_PATH",
		"HERMES_ENV_PATH", "HERMES_API_BASE_URL", "HERMES_API_TOKEN",
		"APP_HOST", "APP_PORT", "APP_SESSION_SECRET", "APP_DEBUG",
		"DEFAULT_PROFILE", "DEFAULT_CHAT_MODEL",
		"ENABLE_DIRECT_CLI_MODE", "ENABLE_API_MODE", "ENABLE_MOCK_MODE",
		"LOG_LEVEL",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

func TestDefaults(t *testing.T) {
	clearEnv()

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppHost != "localhost" {
		t.Errorf("expected AppHost=localhost, got %q", cfg.AppHost)
	}
	if cfg.AppPort != "8080" {
		t.Errorf("expected AppPort=8080, got %q", cfg.AppPort)
	}
	if !cfg.EnableMockMode {
		t.Error("expected EnableMockMode=true by default")
	}
	if cfg.EnableDirectCLIMode {
		t.Error("expected EnableDirectCLIMode=false by default")
	}
	if cfg.EnableAPIMode {
		t.Error("expected EnableAPIMode=false by default")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %q", cfg.LogLevel)
	}
}

func TestEnvOverride(t *testing.T) {
	clearEnv()
	os.Setenv("APP_HOST", "0.0.0.0")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("ENABLE_MOCK_MODE", "false")
	os.Setenv("ENABLE_DIRECT_CLI_MODE", "true")
	defer clearEnv()

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.AppHost != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %q", cfg.AppHost)
	}
	if cfg.AppPort != "9090" {
		t.Errorf("expected 9090, got %q", cfg.AppPort)
	}
	if !cfg.AppDebug {
		t.Error("expected AppDebug=true")
	}
	if cfg.EnableMockMode {
		t.Error("expected EnableMockMode=false")
	}
	if !cfg.EnableDirectCLIMode {
		t.Error("expected EnableDirectCLIMode=true")
	}
}

func TestAddr(t *testing.T) {
	clearEnv()
	os.Setenv("APP_HOST", "127.0.0.1")
	os.Setenv("APP_PORT", "4000")
	defer clearEnv()

	cfg, _ := config.Load("")
	if cfg.Addr() != "127.0.0.1:4000" {
		t.Errorf("expected 127.0.0.1:4000, got %q", cfg.Addr())
	}
}

func TestActiveMode(t *testing.T) {
	clearEnv()

	// Default = Mock
	cfg, _ := config.Load("")
	if cfg.ActiveMode() != "Mock" {
		t.Errorf("expected Mock, got %q", cfg.ActiveMode())
	}

	// CLI mode
	os.Setenv("ENABLE_DIRECT_CLI_MODE", "true")
	os.Setenv("ENABLE_MOCK_MODE", "false")
	cfg, _ = config.Load("")
	if cfg.ActiveMode() != "CLI" {
		t.Errorf("expected CLI, got %q", cfg.ActiveMode())
	}
	clearEnv()

	// API mode
	os.Setenv("ENABLE_API_MODE", "true")
	os.Setenv("ENABLE_MOCK_MODE", "false")
	cfg, _ = config.Load("")
	if cfg.ActiveMode() != "API" {
		t.Errorf("expected API, got %q", cfg.ActiveMode())
	}
}

func TestDotenvLoad(t *testing.T) {
	clearEnv()

	// Write a temp .env file
	f, err := os.CreateTemp("", "hermes-test-*.env")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString("APP_HOST=envhost\n")
	f.WriteString("APP_PORT=7777\n")
	f.WriteString("# comment line\n")
	f.WriteString("\n")
	f.WriteString("LOG_LEVEL=debug\n")
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("Load with dotenv: %v", err)
	}
	if cfg.AppHost != "envhost" {
		t.Errorf("expected envhost, got %q", cfg.AppHost)
	}
	if cfg.AppPort != "7777" {
		t.Errorf("expected 7777, got %q", cfg.AppPort)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected debug, got %q", cfg.LogLevel)
	}
}
