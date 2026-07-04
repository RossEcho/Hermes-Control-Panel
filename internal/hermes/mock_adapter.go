package hermes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MockAdapter is a fully self-contained adapter that returns realistic sample
// data. It requires no Hermes binary or API. Use it during development and
// demos by setting ENABLE_MOCK_MODE=true (the default).
type MockAdapter struct {
	mu           sync.RWMutex
	sessions     map[string]*Session
	messages     map[string][]Message
	activeModel  string
	stopChannels map[string]context.CancelFunc
}

// NewMockAdapter returns a ready-to-use MockAdapter pre-populated with sample
// sessions, skills, models, jobs, and processes.
func NewMockAdapter() *MockAdapter {
	m := &MockAdapter{
		sessions:     make(map[string]*Session),
		messages:     make(map[string][]Message),
		stopChannels: make(map[string]context.CancelFunc),
		activeModel:  "claude-3-5-sonnet-20241022",
	}
	// Pre-seed sessions
	seeds := []Session{
		{
			ID:           "sess-001",
			Title:        "Refactor auth middleware",
			LastUpdated:  "2026-07-04T09:12:00Z",
			Preview:      "Let's extract the JWT validation into its own package...",
			MessageCount: 14,
			Status:       "running",
		},
		{
			ID:           "sess-002",
			Title:        "Deploy pipeline debug",
			LastUpdated:  "2026-07-03T22:45:00Z",
			Preview:      "The GitHub Actions workflow fails on the docker push step...",
			MessageCount: 8,
			Status:       "running",
		},
		{
			ID:           "sess-003",
			Title:        "Database query optimization",
			LastUpdated:  "2026-07-03T16:00:00Z",
			Preview:      "EXPLAIN ANALYZE shows a sequential scan on users table...",
			MessageCount: 21,
			Status:       "idle",
		},
		{
			ID:           "sess-004",
			Title:        "New feature brainstorm",
			LastUpdated:  "2026-07-02T11:30:00Z",
			Preview:      "I want to add a real-time notification system using SSE...",
			MessageCount: 5,
			Status:       "idle",
		},
	}
	for i := range seeds {
		s := seeds[i]
		m.sessions[s.ID] = &s
		m.messages[s.ID] = []Message{}
	}
	return m
}

// --- Chat ---

func (m *MockAdapter) StreamChat(ctx context.Context, sessionID, message string) (<-chan Event, error) {
	ch := make(chan Event, 32)

	// Store user message
	m.mu.Lock()
	if _, ok := m.sessions[sessionID]; !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("session %q not found", sessionID)
	}
	userMsg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	m.messages[sessionID] = append(m.messages[sessionID], userMsg)

	// Create cancellable context for this stream
	streamCtx, cancel := context.WithCancel(ctx)
	m.stopChannels[sessionID] = cancel
	if sess, ok := m.sessions[sessionID]; ok {
		sess.Status = "running"
	}
	m.mu.Unlock()

	go func() {
		defer close(ch)
		defer cancel()

		send := func(e Event) bool {
			select {
			case <-streamCtx.Done():
				return false
			case ch <- e:
				return true
			}
		}

		sleep := func(d time.Duration) bool {
			select {
			case <-streamCtx.Done():
				return false
			case <-time.After(d):
				return true
			}
		}

		if !send(Event{Type: "message_start", Meta: map[string]string{"session_id": sessionID}}) {
			return
		}
		if !sleep(120 * time.Millisecond) {
			return
		}

		// Simulate a tool call
		if !send(Event{
			Type:    "tool_call_start",
			Content: "",
			Meta:    map[string]string{"tool_id": "tc-001", "tool_name": "read_file"},
		}) {
			return
		}
		if !sleep(150 * time.Millisecond) {
			return
		}
		if !send(Event{
			Type:    "tool_call_update",
			Content: `{"path": "internal/auth/jwt.go"}`,
			Meta:    map[string]string{"tool_id": "tc-001"},
		}) {
			return
		}
		if !sleep(200 * time.Millisecond) {
			return
		}
		if !send(Event{
			Type:    "tool_call_end",
			Content: `// JWT validation logic loaded (284 lines)`,
			Meta:    map[string]string{"tool_id": "tc-001", "status": "done"},
		}) {
			return
		}
		if !sleep(100 * time.Millisecond) {
			return
		}

		// Stream the assistant reply in chunks
		chunks := []string{
			"I've reviewed the file. ",
			"Here's what I recommend:\n\n",
			"1. Extract `validateToken()` into `pkg/auth/token.go`\n",
			"2. Add a `TokenClaims` struct with proper field validation\n",
			"3. Replace the raw `jwt.Parse` call with a typed wrapper that returns `(*TokenClaims, error)`\n\n",
			"This will make testing significantly easier and remove the duplication you mentioned.",
		}

		for _, chunk := range chunks {
			if !send(Event{Type: "message_delta", Content: chunk}) {
				return
			}
			if !sleep(100 * time.Millisecond) {
				return
			}
		}

		fullContent := ""
		for _, c := range chunks {
			fullContent += c
		}

		// Store assistant message
		m.mu.Lock()
		assistantMsg := Message{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "assistant",
			Content:   fullContent,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			ToolCalls: []ToolCall{
				{
					ID:     "tc-001",
					Name:   "read_file",
					Input:  `{"path": "internal/auth/jwt.go"}`,
					Output: "// JWT validation logic loaded (284 lines)",
					Status: "done",
				},
			},
		}
		m.messages[sessionID] = append(m.messages[sessionID], assistantMsg)

		// Update session preview and count
		if sess, ok := m.sessions[sessionID]; ok {
			sess.Preview = fullContent
			if len(fullContent) > 80 {
				sess.Preview = fullContent[:80] + "..."
			}
			sess.MessageCount += 2
			sess.LastUpdated = time.Now().UTC().Format(time.RFC3339)
		}
		m.mu.Unlock()

		m.mu.Lock()
		if sess, ok := m.sessions[sessionID]; ok {
			sess.Status = "idle"
		}
		m.mu.Unlock()

		send(Event{Type: "message_end", Meta: map[string]string{"session_id": sessionID}})
	}()

	return ch, nil
}

func (m *MockAdapter) StopChat(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cancel, ok := m.stopChannels[sessionID]; ok {
		cancel()
		delete(m.stopChannels, sessionID)
	}
	return nil
}

func (m *MockAdapter) GetSessionMessages(sessionID string) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	msgs, ok := m.messages[sessionID]
	if !ok {
		return nil, fmt.Errorf("session %q not found", sessionID)
	}
	out := make([]Message, len(msgs))
	copy(out, msgs)
	return out, nil
}

// --- Skills ---

func (m *MockAdapter) ListSkills() ([]Skill, error) {
	return []Skill{
		{
			ID:          "skill-git-ops",
			Name:        "Git Operations",
			Description: "Manages git repositories: clone, commit, push, branch, merge, and conflict resolution.",
			Category:    "Version Control",
			Tags:        []string{"git", "vcs", "devops"},
			Source:      "builtin",
			Content:     "# Git Operations\n\nProvides full git lifecycle management including staging, committing, branching, and remote operations.\n\n## Commands\n- `git_clone(url, dest)` — Clone a repository\n- `git_commit(message, files)` — Stage and commit files\n- `git_push(remote, branch)` — Push to remote\n- `git_branch(name)` — Create a new branch\n- `git_merge(branch)` — Merge a branch into HEAD\n",
		},
		{
			ID:          "skill-file-ops",
			Name:        "File System",
			Description: "Read, write, search, and manipulate files and directories on the local file system.",
			Category:    "System",
			Tags:        []string{"files", "io", "system"},
			Source:      "builtin",
			Content:     "# File System\n\nProvides low-level file I/O including read, write, append, delete, move, copy, and directory traversal.\n\n## Commands\n- `read_file(path)` — Read file contents\n- `write_file(path, content)` — Write or overwrite a file\n- `list_dir(path)` — List directory entries\n- `find_files(pattern)` — Glob-based file search\n",
		},
		{
			ID:          "skill-web-fetch",
			Name:        "Web Fetch",
			Description: "Fetch URLs, parse HTML/JSON, and interact with HTTP APIs.",
			Category:    "Network",
			Tags:        []string{"http", "web", "api", "scraping"},
			Source:      "builtin",
			Content:     "# Web Fetch\n\nFetches web pages and API endpoints with support for custom headers, cookies, and body payloads.\n\n## Commands\n- `fetch(url, method, headers, body)` — HTTP request\n- `parse_html(html)` — Extract text and links\n- `parse_json(text)` — Parse JSON payload\n",
		},
		{
			ID:          "skill-code-search",
			Name:        "Code Search",
			Description: "Ripgrep-backed code search across local repositories with regex support.",
			Category:    "Development",
			Tags:        []string{"search", "grep", "regex", "code"},
			Source:      "builtin",
			Content:     "# Code Search\n\nHigh-performance code search using ripgrep under the hood.\n\n## Commands\n- `search(pattern, path, flags)` — Regex search\n- `find_symbol(name, lang)` — Language-aware symbol search\n- `grep_files(glob, pattern)` — Filter files by glob then grep\n",
		},
		{
			ID:          "skill-shell-exec",
			Name:        "Shell Execution",
			Description: "Execute shell commands and scripts in a sandboxed environment with output capture.",
			Category:    "System",
			Tags:        []string{"shell", "bash", "exec", "terminal"},
			Source:      "builtin",
			Content:     "# Shell Execution\n\nRuns shell commands with timeout, working-directory control, and env-var injection.\n\n## Commands\n- `run(cmd, cwd, env, timeout)` — Execute a shell command\n- `run_script(path, args)` — Execute a script file\n- `pipe(cmds)` — Pipe multiple commands\n",
		},
		{
			ID:          "skill-db-query",
			Name:        "Database Query",
			Description: "Run SQL queries against PostgreSQL, MySQL, or SQLite databases and return structured results.",
			Category:    "Data",
			Tags:        []string{"sql", "database", "postgres", "mysql"},
			Source:      "community",
			Content:     "# Database Query\n\nConnects to relational databases and executes queries, returning results as typed JSON.\n\n## Commands\n- `query(dsn, sql, params)` — Execute a SELECT query\n- `exec(dsn, sql, params)` — Execute an INSERT/UPDATE/DELETE\n- `schema(dsn, table)` — Describe a table schema\n",
		},
	}, nil
}

func (m *MockAdapter) GetSkill(id string) (*Skill, error) {
	skills, _ := m.ListSkills()
	for i := range skills {
		if skills[i].ID == id {
			return &skills[i], nil
		}
	}
	return nil, fmt.Errorf("skill %q not found", id)
}

// --- Models ---

func (m *MockAdapter) ListModels() ([]Model, error) {
	m.mu.RLock()
	active := m.activeModel
	m.mu.RUnlock()

	models := []Model{
		{
			ID:        "claude-3-5-sonnet-20241022",
			Name:      "Claude 3.5 Sonnet",
			Provider:  "Anthropic",
			IsDefault: true,
			IsActive:  active == "claude-3-5-sonnet-20241022",
			IsLocal:   false,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        1842,
				InputTokens:  4_210_500,
				OutputTokens: 1_388_200,
				CostUSD:      31.47,
			},
		},
		{
			ID:        "claude-3-opus-20240229",
			Name:      "Claude 3 Opus",
			Provider:  "Anthropic",
			IsDefault: false,
			IsActive:  active == "claude-3-opus-20240229",
			IsLocal:   false,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        214,
				InputTokens:  892_000,
				OutputTokens: 310_400,
				CostUSD:      48.21,
			},
		},
		{
			ID:        "claude-3-haiku-20240307",
			Name:      "Claude 3 Haiku",
			Provider:  "Anthropic",
			IsDefault: false,
			IsActive:  active == "claude-3-haiku-20240307",
			IsLocal:   false,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        5_103,
				InputTokens:  9_841_000,
				OutputTokens: 2_914_000,
				CostUSD:      3.12,
			},
		},
		{
			ID:        "gpt-4o",
			Name:      "GPT-4o",
			Provider:  "OpenAI",
			IsDefault: false,
			IsActive:  active == "gpt-4o",
			IsLocal:   false,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        387,
				InputTokens:  1_204_000,
				OutputTokens: 498_700,
				CostUSD:      22.84,
			},
		},
		{
			ID:        "gpt-4o-mini",
			Name:      "GPT-4o Mini",
			Provider:  "OpenAI",
			IsDefault: false,
			IsActive:  active == "gpt-4o-mini",
			IsLocal:   false,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        2_910,
				InputTokens:  6_330_000,
				OutputTokens: 1_870_000,
				CostUSD:      1.94,
			},
		},
		{
			ID:        "ollama/llama3.1:8b",
			Name:      "Llama 3.1 8B",
			Provider:  "Ollama (local)",
			IsDefault: false,
			IsActive:  active == "ollama/llama3.1:8b",
			IsLocal:   true,
			Available: true,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        720,
				InputTokens:  1_540_000,
				OutputTokens: 620_000,
				CostUSD:      0,
			},
		},
		{
			ID:        "ollama/codestral:22b",
			Name:      "Codestral 22B",
			Provider:  "Ollama (local)",
			IsDefault: false,
			IsActive:  active == "ollama/codestral:22b",
			IsLocal:   true,
			Available: false,
			MonthlyUsage: &ModelMonthlyUsage{
				Calls:        0,
				InputTokens:  0,
				OutputTokens: 0,
				CostUSD:      0,
			},
		},
	}
	return models, nil
}

func (m *MockAdapter) SwitchModel(modelID string) error {
	models, _ := m.ListModels()
	for _, mod := range models {
		if mod.ID == modelID {
			m.mu.Lock()
			m.activeModel = modelID
			m.mu.Unlock()
			return nil
		}
	}
	return fmt.Errorf("model %q not found", modelID)
}

// --- Sessions ---

func (m *MockAdapter) ListSessions() ([]Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, *s)
	}
	return out, nil
}

func (m *MockAdapter) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %q not found", id)
	}
	cp := *s
	return &cp, nil
}

func (m *MockAdapter) NewSession(title string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	if title == "" {
		title = "New Session"
	}
	s := &Session{
		ID:           id,
		Title:        title,
		LastUpdated:  time.Now().UTC().Format(time.RFC3339),
		Preview:      "",
		MessageCount: 0,
		Status:       "idle",
	}
	m.sessions[id] = s
	m.messages[id] = []Message{}
	cp := *s
	return &cp, nil
}

func (m *MockAdapter) DeleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cancel, ok := m.stopChannels[id]; ok {
		cancel()
		delete(m.stopChannels, id)
	}
	delete(m.sessions, id)
	delete(m.messages, id)
	return nil
}

func (m *MockAdapter) GetUsage() (*Usage, error) {
	m.mu.RLock()
	sessCount := len(m.sessions)
	totalMsgs := 0
	for _, msgs := range m.messages {
		totalMsgs += len(msgs)
	}
	m.mu.RUnlock()
	return &Usage{
		InputTokens:    12438 + totalMsgs*120,
		OutputTokens:   4219 + totalMsgs*85,
		ActiveSessions: sessCount,
		TotalMessages:  48 + totalMsgs,
		TodayMessages:  23 + totalMsgs,
	}, nil
}

// --- System ---

func (m *MockAdapter) GetStatus() (*Status, error) {
	return &Status{
		HermesProcess:  "running",
		APIStatus:      "connected",
		GatewayStatus:  "ok",
		ActiveProfile:  "default",
		LoadedToolsets: []string{"builtin", "filesystem", "network", "shell"},
		SessionCount:   len(m.sessions),
		JobCount:       3,
	}, nil
}

func (m *MockAdapter) ListJobs() ([]Job, error) {
	return []Job{
		{
			ID:        "job-001",
			Name:      "Index codebase",
			Status:    "completed",
			StartTime: "2026-07-04T08:00:00Z",
			EndTime:   "2026-07-04T08:02:31Z",
			Output:    "Indexed 1,482 files across 23 packages.",
		},
		{
			ID:        "job-002",
			Name:      "Sync skills from registry",
			Status:    "running",
			StartTime: "2026-07-04T09:00:00Z",
			EndTime:   "",
			Output:    "Fetching skill manifests... (3/6 complete)",
		},
		{
			ID:        "job-003",
			Name:      "Cleanup old sessions",
			Status:    "failed",
			StartTime: "2026-07-04T07:55:00Z",
			EndTime:   "2026-07-04T07:55:04Z",
			Output:    "Error: permission denied on /var/hermes/sessions/archived",
		},
	}, nil
}

func (m *MockAdapter) ListProcesses() ([]Process, error) {
	return []Process{
		{
			PID:    12340,
			Name:   "hermes",
			Status: "running",
			CPU:    "1.2%",
			Memory: "124 MB",
			Uptime: "4h 32m",
		},
		{
			PID:    12341,
			Name:   "hermes-gateway",
			Status: "running",
			CPU:    "0.4%",
			Memory: "48 MB",
			Uptime: "4h 32m",
		},
		{
			PID:    12399,
			Name:   "ollama",
			Status: "running",
			CPU:    "0.1%",
			Memory: "892 MB",
			Uptime: "12h 10m",
		},
	}, nil
}

func (m *MockAdapter) GetLogs(n int) ([]string, error) {
	all := []string{
		"[2026-07-04T09:00:00Z] INFO  hermes: starting session manager",
		"[2026-07-04T09:00:01Z] INFO  hermes: loaded 6 skills from builtin toolset",
		"[2026-07-04T09:00:01Z] INFO  hermes: loaded 4 skills from filesystem toolset",
		"[2026-07-04T09:00:02Z] INFO  gateway: listening on :7700",
		"[2026-07-04T09:00:02Z] INFO  hermes: connected to anthropic api",
		"[2026-07-04T09:12:00Z] INFO  session sess-001: message received (user)",
		"[2026-07-04T09:12:01Z] INFO  session sess-001: invoking tool read_file",
		"[2026-07-04T09:12:01Z] INFO  tool read_file: path=internal/auth/jwt.go",
		"[2026-07-04T09:12:02Z] INFO  session sess-001: streaming response start",
		"[2026-07-04T09:12:03Z] INFO  session sess-001: streaming response end (284 tokens)",
	}
	if n > len(all) {
		n = len(all)
	}
	return all[len(all)-n:], nil
}

func (m *MockAdapter) RunAction(action string, args map[string]string) (string, error) {
	argsJSON, _ := json.Marshal(args)
	return fmt.Sprintf("[mock] action=%q args=%s — OK", action, argsJSON), nil
}
