package hermes

// =============================================================================
// APIAdapter — Hermes REST API wiring stub
// =============================================================================
//
// This adapter talks to a running Hermes API server over HTTP.
// It is activated when ENABLE_API_MODE=true.
//
// WIRING GUIDE
// ============
//
// 1. Set HERMES_API_BASE_URL to the base URL of the Hermes API server.
//    e.g. http://localhost:7700  or  https://hermes.internal
//
// 2. Set HERMES_API_TOKEN to a bearer token if the API requires auth.
//
// 3. Implement each method by calling the appropriate endpoint:
//
//    StreamChat:
//      POST {baseURL}/v1/sessions/{sessionID}/chat
//      body: {"message": "..."}
//      response: text/event-stream (SSE)
//      Decode SSE events → Event structs
//
//    StopChat:
//      POST {baseURL}/v1/sessions/{sessionID}/stop
//
//    GetSessionMessages:
//      GET {baseURL}/v1/sessions/{sessionID}/messages
//      response: {"messages": [...]}
//
//    ListSkills:
//      GET {baseURL}/v1/skills
//      response: {"skills": [...]}
//
//    GetSkill:
//      GET {baseURL}/v1/skills/{id}
//
//    ListModels:
//      GET {baseURL}/v1/models
//
//    SwitchModel:
//      POST {baseURL}/v1/models/active
//      body: {"model_id": "..."}
//
//    ListSessions:
//      GET {baseURL}/v1/sessions
//
//    GetSession:
//      GET {baseURL}/v1/sessions/{id}
//
//    NewSession:
//      POST {baseURL}/v1/sessions
//      response: {"session": {...}}
//
//    GetStatus:
//      GET {baseURL}/v1/status
//
//    ListJobs:
//      GET {baseURL}/v1/jobs
//
//    ListProcesses:
//      GET {baseURL}/v1/processes
//
//    GetLogs:
//      GET {baseURL}/v1/logs?n={n}
//
//    RunAction:
//      POST {baseURL}/v1/actions/{action}
//      body: {args}
//
// 4. All HTTP requests should:
//    - Set Authorization: Bearer {HERMES_API_TOKEN} if token is set
//    - Set Accept: application/json
//    - Use a context-aware http.Client with reasonable timeouts
//
// 5. Response JSON should map 1:1 to the structs in adapter.go.
//    Add intermediate DTO types here if the API response format differs.
//
// =============================================================================

import (
	"context"
	"net/http"
)

// APIAdapter drives Hermes via its REST API.
type APIAdapter struct {
	// baseURL is the base URL of the Hermes API server.
	baseURL string
	// token is the bearer token for API authentication (may be empty).
	token string
	// client is the HTTP client used for all requests.
	client *http.Client
}

// NewAPIAdapter creates an APIAdapter.
func NewAPIAdapter(baseURL, token string) *APIAdapter {
	return &APIAdapter{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

func (a *APIAdapter) StreamChat(_ context.Context, _, _ string) (<-chan Event, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) StopChat(_ string) error {
	return ErrNotImplemented
}

func (a *APIAdapter) GetSessionMessages(_ string) ([]Message, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) ListSkills() ([]Skill, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) GetSkill(_ string) (*Skill, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) ListModels() ([]Model, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) SwitchModel(_ string) error {
	return ErrNotImplemented
}

func (a *APIAdapter) ListSessions() ([]Session, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) GetSession(_ string) (*Session, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) NewSession(_ string) (*Session, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) GetStatus() (*Status, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) ListJobs() ([]Job, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) ListProcesses() ([]Process, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) GetLogs(_ int) ([]string, error) {
	return nil, ErrNotImplemented
}

func (a *APIAdapter) RunAction(_ string, _ map[string]string) (string, error) {
	return "", ErrNotImplemented
}

func (a *APIAdapter) DeleteSession(_ string) error {
	// Wire: DELETE {baseURL}/v1/sessions/{id}
	return ErrNotImplemented
}

func (a *APIAdapter) GetUsage() (*Usage, error) {
	// Wire: GET {baseURL}/v1/usage
	return nil, ErrNotImplemented
}
