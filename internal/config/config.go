// Package config loads and exposes all application configuration from
// environment variables. It also supports a simple .env file parser so that
// the binary can be run without a shell-level env setup.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration values.
type Config struct {
	// Hermes integration
	HermesExecutablePath string
	HermesHome           string
	HermesConfigPath     string
	HermesEnvPath        string
	HermesAPIBaseURL     string
	HermesAPIToken       string

	// Server
	AppHost          string
	AppPort          string
	AppSessionSecret string
	AppDebug         bool

	// Hermes defaults
	DefaultProfile    string
	DefaultChatModel  string

	// Mode flags
	EnableDirectCLIMode bool
	EnableAPIMode       bool
	EnableMockMode      bool

	// Logging
	LogLevel string
}

// Load reads configuration from environment variables. If a .env file path is
// provided and exists, it is parsed first (env vars already set in the process
// take precedence over .env values).
func Load(dotenvPath string) (*Config, error) {
	if dotenvPath != "" {
		if err := loadDotenv(dotenvPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: loading .env: %w", err)
		}
	}

	cfg := &Config{
		HermesExecutablePath: getEnv("HERMES_EXECUTABLE_PATH", ""),
		HermesHome:           getEnv("HERMES_HOME", ""),
		HermesConfigPath:     getEnv("HERMES_CONFIG_PATH", ""),
		HermesEnvPath:        getEnv("HERMES_ENV_PATH", ""),
		HermesAPIBaseURL:     getEnv("HERMES_API_BASE_URL", ""),
		HermesAPIToken:       getEnv("HERMES_API_TOKEN", ""),

		AppHost:          getEnv("APP_HOST", "localhost"),
		AppPort:          getEnv("APP_PORT", "8080"),
		AppSessionSecret: getEnv("APP_SESSION_SECRET", "change-me-in-production"),
		AppDebug:         getBool("APP_DEBUG", false),

		DefaultProfile:   getEnv("DEFAULT_PROFILE", "default"),
		DefaultChatModel: getEnv("DEFAULT_CHAT_MODEL", "claude-3-5-sonnet-20241022"),

		EnableDirectCLIMode: getBool("ENABLE_DIRECT_CLI_MODE", false),
		EnableAPIMode:       getBool("ENABLE_API_MODE", false),
		EnableMockMode:      getBool("ENABLE_MOCK_MODE", true),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

// Addr returns the full host:port address for the HTTP server.
func (c *Config) Addr() string {
	return c.AppHost + ":" + c.AppPort
}

// ActiveMode returns a human-readable string describing the active adapter mode.
func (c *Config) ActiveMode() string {
	switch {
	case c.EnableDirectCLIMode:
		return "CLI"
	case c.EnableAPIMode:
		return "API"
	default:
		return "Mock"
	}
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getBool parses a boolean environment variable. Accepts "true", "1", "yes"
// (case-insensitive) as true; anything else is false.
func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// loadDotenv parses a .env file and sets unset environment variables from it.
// Lines starting with '#' and empty lines are ignored. Each non-comment line
// must be in KEY=VALUE format (no export prefix, no quoting handling beyond
// stripping surrounding double-quotes).
func loadDotenv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip surrounding double-quotes
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		// Only set if not already present
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
