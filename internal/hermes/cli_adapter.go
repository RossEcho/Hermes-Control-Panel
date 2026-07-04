package hermes

// =============================================================================
// CLIAdapter — Hermes CLI wiring stub
// =============================================================================
//
// This adapter invokes the Hermes binary directly via os/exec.
// It is activated when ENABLE_DIRECT_CLI_MODE=true.
//
// WIRING GUIDE
// ============
//
// 1. Set HERMES_EXECUTABLE_PATH to the path of the Hermes binary.
//    e.g. /usr/local/bin/hermes   or   C:\Program Files\Hermes\hermes.exe
//
// 2. Set HERMES_HOME to the Hermes data directory.
//    e.g. ~/.hermes
//
// 3. Set HERMES_CONFIG_PATH to the Hermes config file.
//    e.g. ~/.hermes/config.yaml
//
// 4. Implement each method by constructing an exec.Cmd:
//
//    StreamChat:
//      cmd := exec.CommandContext(ctx, cfg.HermesExecutablePath,
//          "chat", "--session", sessionID, "--message", message,
//          "--stream", "--output", "json")
//      // pipe cmd.Stdout and decode NDJSON events into Event structs
//
//    ListSkills:
//      cmd := exec.Command(cfg.HermesExecutablePath, "skills", "list", "--output", "json")
//      // decode stdout JSON array into []Skill
//
//    ListModels:
//      cmd := exec.Command(cfg.HermesExecutablePath, "models", "list", "--output", "json")
//
//    SwitchModel:
//      cmd := exec.Command(cfg.HermesExecutablePath, "models", "use", modelID)
//
//    ListSessions:
//      cmd := exec.Command(cfg.HermesExecutablePath, "sessions", "list", "--output", "json")
//
//    NewSession:
//      cmd := exec.Command(cfg.HermesExecutablePath, "sessions", "new")
//      // parse session ID from stdout
//
//    GetStatus:
//      cmd := exec.Command(cfg.HermesExecutablePath, "status", "--output", "json")
//
//    ListJobs:
//      cmd := exec.Command(cfg.HermesExecutablePath, "jobs", "list", "--output", "json")
//
//    GetLogs:
//      cmd := exec.Command(cfg.HermesExecutablePath, "logs", "--tail", strconv.Itoa(n))
//
//    RunAction:
//      cmd := exec.Command(cfg.HermesExecutablePath, "run", action, ...)
//
// 5. All JSON models returned by the Hermes CLI should map 1:1 to the structs
//    defined in adapter.go. Add intermediate struct types here if the CLI
//    output format differs.
//
// =============================================================================

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by all stub methods.
var ErrNotImplemented = errors.New("not implemented: real Hermes wiring required")

// CLIAdapter drives Hermes via its command-line interface.
type CLIAdapter struct {
	// executablePath is the absolute path to the Hermes binary.
	executablePath string
	// hermesHome is the Hermes data directory.
	hermesHome string
	// configPath is the path to the Hermes config file.
	configPath string
}

// NewCLIAdapter creates a CLIAdapter.
func NewCLIAdapter(executablePath, hermesHome, configPath string) *CLIAdapter {
	return &CLIAdapter{
		executablePath: executablePath,
		hermesHome:     hermesHome,
		configPath:     configPath,
	}
}

func (a *CLIAdapter) StreamChat(_ context.Context, _, _ string) (<-chan Event, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) StopChat(_ string) error {
	return ErrNotImplemented
}

func (a *CLIAdapter) GetSessionMessages(_ string) ([]Message, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) ListSkills() ([]Skill, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) GetSkill(_ string) (*Skill, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) ListModels() ([]Model, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) SwitchModel(_ string) error {
	return ErrNotImplemented
}

func (a *CLIAdapter) ListSessions() ([]Session, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) GetSession(_ string) (*Session, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) NewSession(_ string) (*Session, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) GetStatus() (*Status, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) ListJobs() ([]Job, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) ListProcesses() ([]Process, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) GetLogs(_ int) ([]string, error) {
	return nil, ErrNotImplemented
}

func (a *CLIAdapter) RunAction(_ string, _ map[string]string) (string, error) {
	return "", ErrNotImplemented
}

func (a *CLIAdapter) DeleteSession(_ string) error {
	// Wire: exec.Command(executablePath, "sessions", "delete", id)
	return ErrNotImplemented
}

func (a *CLIAdapter) GetUsage() (*Usage, error) {
	// Wire: exec.Command(executablePath, "usage", "--output", "json")
	return nil, ErrNotImplemented
}
