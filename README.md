# Hermes Control Panel

A Go-based local web control panel for the **Hermes** AI assistant system.
It provides a dark-themed operator UI with chat, skill browsing, model switching,
session management, system status, job monitoring, and configuration viewing.

The panel is built as a **shell/control-layer** around Hermes. All Hermes
interaction is funneled through a single `Adapter` interface with three
implementations:

| Mode | Adapter | Status |
|------|---------|--------|
| **Mock** (default) | `MockAdapter` | Fully implemented — no Hermes required |
| **CLI** | `CLIAdapter` | Stub — see wiring guide below |
| **API** | `APIAdapter` | Stub — see wiring guide below |

---

## What is implemented

- Full server-rendered UI (dark operator theme, fixed sidebar, 7 pages)
- Chat page with session sidebar, streaming message pane via SSE, tool-call
  collapsible blocks, and Stop button
- Skills browser with live search filter and detail panel
- Models page with provider grouping and switch-model action
- Sessions page with table view and Open button
- Status dashboard with colour-coded status cards
- Jobs page with jobs table, processes table, and log tail
- Config page (read-only) with masked secrets and wiring hints
- Complete `MockAdapter` with realistic sample data (6 models, 6 skills,
  4 sessions, 3 jobs, 3 processes, 10 log lines, streaming chat simulation
  with a tool_call event)
- Simple `.env` file parser (no external dependency)
- `CLIAdapter` stub with full wiring documentation
- `APIAdapter` stub with full wiring documentation
- Config test suite (env vars, defaults, .env loading)
- Mock adapter test suite (all interface methods)
- Handler test suite (health endpoint, redirects, JSON APIs)

## What is mocked

Everything in `MockAdapter` is simulated in-process with no real Hermes
process required:

- Chat streaming: ~5–8 SSE events with 100ms delays, including one tool call
- Skills: 6 built-in/community skills with full content
- Models: 5 cloud models (Anthropic, OpenAI) + 2 local Ollama models
- Sessions: 4 pre-seeded sessions; in-memory accumulation for new messages
- Status: all green/running
- Jobs: completed, running, and failed samples
- Processes: hermes, hermes-gateway, ollama
- Logs: 10 realistic log lines

---

## Where real Hermes wiring connects

### CLI mode (`internal/hermes/cli_adapter.go`)

Implement each method by constructing an `exec.Cmd` with the Hermes binary.
The file contains a detailed wiring guide as a comment block:

- `HERMES_EXECUTABLE_PATH` → path to the `hermes` binary
- `HERMES_HOME` → Hermes data directory
- `HERMES_CONFIG_PATH` → Hermes config file
- Expected CLI subcommands: `chat --stream`, `skills list`, `models list/use`,
  `sessions list/new`, `status`, `jobs list`, `logs --tail N`

### API mode (`internal/hermes/api_adapter.go`)

Implement each method with `http.Client` calls against the Hermes REST API.
The file contains a detailed wiring guide:

- `HERMES_API_BASE_URL` → e.g. `http://localhost:7700`
- `HERMES_API_TOKEN` → bearer token (optional)
- Expected endpoints: `POST /v1/sessions/{id}/chat` (SSE),
  `GET /v1/skills`, `GET /v1/models`, `POST /v1/models/active`, etc.

---

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HERMES_EXECUTABLE_PATH` | | Path to Hermes binary (CLI mode) |
| `HERMES_HOME` | | Hermes data directory |
| `HERMES_CONFIG_PATH` | | Hermes config file path |
| `HERMES_ENV_PATH` | | Hermes .env file path |
| `HERMES_API_BASE_URL` | | Hermes API base URL (API mode) |
| `HERMES_API_TOKEN` | | Hermes API bearer token |
| `APP_HOST` | `localhost` | Server bind host |
| `APP_PORT` | `8080` | Server bind port |
| `APP_SESSION_SECRET` | `change-me-in-production` | Session secret |
| `APP_DEBUG` | `false` | Enable debug output |
| `DEFAULT_PROFILE` | `default` | Default Hermes profile |
| `DEFAULT_CHAT_MODEL` | `claude-3-5-sonnet-20241022` | Default chat model |
| `ENABLE_MOCK_MODE` | `true` | Use MockAdapter |
| `ENABLE_DIRECT_CLI_MODE` | `false` | Use CLIAdapter |
| `ENABLE_API_MODE` | `false` | Use APIAdapter |
| `LOG_LEVEL` | `info` | Log level |

Copy `.env.example` to `.env` and edit as needed. The `.env` file is optional —
all variables can also be set in the process environment.

---

## Running in mock mode (default)

```bash
# From the project root
go run ./cmd/server
```

Open http://localhost:8080 in your browser.

## Building

```bash
go build -o hermes-ui ./cmd/server
./hermes-ui
```

## Running with the real Hermes CLI

1. Fill in `HERMES_EXECUTABLE_PATH` in `.env`
2. Set `ENABLE_DIRECT_CLI_MODE=true` and `ENABLE_MOCK_MODE=false`
3. Implement `CLIAdapter` methods in `internal/hermes/cli_adapter.go`
4. Run as normal

```bash
ENABLE_DIRECT_CLI_MODE=true ENABLE_MOCK_MODE=false go run ./cmd/server
```

## Running with the Hermes REST API

1. Fill in `HERMES_API_BASE_URL` (and optionally `HERMES_API_TOKEN`) in `.env`
2. Set `ENABLE_API_MODE=true` and `ENABLE_MOCK_MODE=false`
3. Implement `APIAdapter` methods in `internal/hermes/api_adapter.go`
4. Run as normal

```bash
ENABLE_API_MODE=true ENABLE_MOCK_MODE=false go run ./cmd/server
```

## Running tests

```bash
# Run all tests from project root (templates are accessible)
go test ./...

# Run specific packages
go test ./internal/config/...
go test ./internal/hermes/...
go test ./internal/handlers/...
```

---

## Known limitations

- Session state is stored in-memory only — restarting the server loses all
  sessions created during the run (pre-seeded mock sessions are always
  regenerated).
- No authentication or access control. Intended for local/trusted use.
- Template files must be present relative to the working directory where the
  binary is invoked (`web/templates/*.html`). Run the binary from the project
  root.
- No mobile responsiveness; designed for desktop operator use.
- The CLI and API adapters are stubs that return `ErrNotImplemented` until
  wired up.
- SSE streaming does not use WebSockets; browser support for EventSource is
  required (all modern browsers).
