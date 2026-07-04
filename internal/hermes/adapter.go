package hermes

import "context"

// Event represents a single streaming event from Hermes.
type Event struct {
	// Type is one of: message_start, message_delta, message_end,
	// tool_call_start, tool_call_update, tool_call_end, status, error, session_update
	Type    string
	Content string
	Meta    map[string]string
}

// Message is a single chat message in a session.
type Message struct {
	ID        string
	Role      string // user, assistant
	Content   string
	Timestamp string
	ToolCalls []ToolCall
}

// ToolCall represents a tool/function call made by the assistant.
type ToolCall struct {
	ID     string
	Name   string
	Input  string
	Output string
	Status string // pending, running, done, error
}

// Skill is a Hermes skill definition.
type Skill struct {
	ID          string
	Name        string
	Description string
	Category    string
	Tags        []string
	Source      string
	Content     string
}

// ModelMonthlyUsage holds 30-day aggregate stats for a single model.
type ModelMonthlyUsage struct {
	Calls        int
	InputTokens  int
	OutputTokens int
	CostUSD      float64
}

// Model represents an LLM model available to Hermes.
type Model struct {
	ID           string
	Name         string
	Provider     string
	IsDefault    bool
	IsActive     bool
	IsLocal      bool
	Available    bool
	MonthlyUsage *ModelMonthlyUsage
}

// Session is a Hermes chat session.
type Session struct {
	ID           string
	Title        string
	LastUpdated  string
	Preview      string
	MessageCount int
	Status       string // "running", "idle"
}

// Usage holds aggregate token and activity counters for the current context.
type Usage struct {
	InputTokens    int
	OutputTokens   int
	ActiveSessions int
	TotalMessages  int
	TodayMessages  int
}

// Status is the overall system status snapshot.
type Status struct {
	HermesProcess  string
	APIStatus      string
	GatewayStatus  string
	ActiveProfile  string
	LoadedToolsets []string
	SessionCount   int
	JobCount       int
}

// Job is a background job managed by Hermes.
type Job struct {
	ID        string
	Name      string
	Status    string
	StartTime string
	EndTime   string
	Output    string
}

// Process is an OS-level process tracked by Hermes.
type Process struct {
	PID    int
	Name   string
	Status string
	CPU    string
	Memory string
	Uptime string
}

// Adapter is the interface that all Hermes backend drivers must satisfy.
// Switch between implementations by setting env vars:
//
//	ENABLE_MOCK_MODE=true      → MockAdapter  (default, no Hermes required)
//	ENABLE_DIRECT_CLI_MODE=true → CLIAdapter  (invokes hermes binary)
//	ENABLE_API_MODE=true        → APIAdapter  (calls Hermes REST API)
type Adapter interface {
	// Chat
	StreamChat(ctx context.Context, sessionID, message string) (<-chan Event, error)
	StopChat(sessionID string) error
	GetSessionMessages(sessionID string) ([]Message, error)

	// Skills
	ListSkills() ([]Skill, error)
	GetSkill(id string) (*Skill, error)

	// Models
	ListModels() ([]Model, error)
	SwitchModel(modelID string) error

	// Sessions
	ListSessions() ([]Session, error)
	GetSession(id string) (*Session, error)
	NewSession(title string) (*Session, error)
	DeleteSession(id string) error

	// Usage
	GetUsage() (*Usage, error)

	// System
	GetStatus() (*Status, error)
	ListJobs() ([]Job, error)
	ListProcesses() ([]Process, error)
	GetLogs(n int) ([]string, error)
	RunAction(action string, args map[string]string) (string, error)
}
