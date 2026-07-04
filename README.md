# Bot Factory — Hermes Control Panel

A local web control panel for the Hermes AI agent CLI. Built in Go with no frontend build step — just run and open a browser.

---

## Overview

Bot Factory is a dark-themed operator dashboard that gives you a real-time view of every AI session running under Hermes. You can spin up new bots, watch them stream responses, browse the skill library, switch models, and monitor token spend — all without leaving the browser.

The app runs fully in **mock mode** by default (no Hermes binary required), making it usable as a standalone demo or UI prototype. Swapping in the real CLI or REST API adapter requires only an environment variable change.

---

## Features

### Bots (Main View)
- **Active Bots grid** — session blobs showing each bot's name, status (running / idle), and message count. Click any blob to open a floating chat window.
- **New Bot button** — opens the Bot Factory modal to configure and launch a new session.
- **System health cards** — live indicators for Hermes process, API connection, gateway, active profile, session count, and job count.

### Bot Factory Modal
Launched via the **New Bot** button. Fields:
- **Session Name** — human-readable label for the bot.
- **Short Context** — initial instructions or framing passed to the session.
- **Model** — choose from cloud models (Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku, GPT-4o, GPT-4o Mini) or local Ollama models.
- **Harness** — execution profile (Default, Code, Research, Ops, Custom).
- **Workspace Path** — local directory the bot operates in.

On submit the session is created and immediately opens as a floating chat window.

### Floating Chat Windows
Windows Messenger-style chat panels that stack along the bottom of the screen. Each window has:
- **Live status dot** — pulses green while the bot is streaming a response.
- **Minimize / Close** controls.
- **Enlarge (⛶)** — expands the window into a large centered modal with a full composer.
- **Real-time streaming** — responses arrive token-by-token via Server-Sent Events (SSE).
- **Stop button** — cancels an in-progress stream.

Multiple windows can be open simultaneously; they reposition automatically when one is closed.

### Skills Browser
Browse and search the full library of Hermes skills (tools). Each skill shows its name, category, tags, source, and full markdown documentation. Filter by name or category using the search bar.

### Models
- **Active model card** — highlights the currently selected LLM.
- **Cloud and local model grids** — availability status and one-click switching.
- **Monthly Usage table** — last-30-day breakdown per model: call count, tokens in, tokens out, and estimated cost. Local models show as free.

### Sidebar
- **SPA navigation** — clicking Bots, Skills, Models, or Config swaps only the main content area; the sidebar never reloads.
- **Running Agents panel** — lists all sessions with their live status dot and a remove (×) button to delete a session.
- **Usage strip** — auto-refreshes every 30 seconds showing aggregate token counts (input, output, active sessions, messages today).

### Config
Displays the active runtime configuration: mode, server address, adapter settings, and masked secrets.

---

## Architecture

| Layer | Technology |
|---|---|
| Server | Go 1.21, `net/http`, `github.com/go-chi/chi/v5` |
| Templates | `html/template` — one isolated set per page to prevent content-block collisions |
| Streaming | Server-Sent Events (SSE) — no WebSocket dependency |
| Progressive enhancement | HTMX (usage fragment polling) |
| SPA nav | Vanilla JS fetch + `DOMParser` + `history.pushState` |
| Floating UI | Pure CSS + vanilla JS — no React, no npm |

### Adapter pattern

All Hermes interaction is abstracted behind a single `Adapter` interface. Switch modes via environment variables:

| Env var | Adapter | Description |
|---|---|---|
| `ENABLE_MOCK_MODE=true` *(default)* | `MockAdapter` | Fully self-contained; no Hermes required |
| `ENABLE_DIRECT_CLI_MODE=true` | `CLIAdapter` | Invokes the Hermes binary via `os/exec` |
| `ENABLE_API_MODE=true` | `APIAdapter` | Calls a running Hermes REST API server |

---

## Running

```bash
# Default — mock mode on port 8081
APP_PORT=8081 go run ./cmd/server
```

Open **http://localhost:8081** in your browser.

### Connect to a real Hermes CLI

```bash
ENABLE_DIRECT_CLI_MODE=true \
ENABLE_MOCK_MODE=false \
HERMES_EXECUTABLE_PATH=/usr/local/bin/hermes \
go run ./cmd/server
```

### Connect to a Hermes REST API

```bash
ENABLE_API_MODE=true \
ENABLE_MOCK_MODE=false \
HERMES_API_BASE_URL=http://localhost:7700 \
go run ./cmd/server
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | Server bind port |
| `APP_HOST` | `localhost` | Server bind host |
| `ENABLE_MOCK_MODE` | `true` | Use MockAdapter (no Hermes needed) |
| `ENABLE_DIRECT_CLI_MODE` | `false` | Use CLIAdapter |
| `ENABLE_API_MODE` | `false` | Use APIAdapter |
| `HERMES_EXECUTABLE_PATH` | | Path to Hermes binary (CLI mode) |
| `HERMES_API_BASE_URL` | | Hermes API base URL (API mode) |
| `HERMES_API_TOKEN` | | Hermes API bearer token |
| `DEFAULT_CHAT_MODEL` | `claude-3-5-sonnet-20241022` | Default model |

Copy `.env.example` to `.env` and edit as needed.

---

## Project Structure

```
cmd/server/          — entry point
internal/
  config/            — env-based configuration
  handlers/          — HTTP handlers and router
  hermes/
    adapter.go       — shared types + Adapter interface
    mock_adapter.go  — mock implementation (default)
    cli_adapter.go   — CLI wiring stub
    api_adapter.go   — REST API wiring stub
web/
  static/
    style.css        — dark operator theme (CSS custom properties)
    app.js           — chat SSE, floating windows, modal, SPA nav
  templates/
    layout.html      — sidebar shell, usage strip, nav
    status.html      — Bots / Active Bots view
    chat.html        — full chat page (session sidebar + message pane)
    skills.html      — skill browser
    models.html      — model switcher + monthly usage table
    config.html      — configuration viewer
```

---

## Known Limitations

- Session state is in-memory only — restarting the server clears sessions created at runtime (pre-seeded mock sessions are always regenerated).
- No authentication or access control. Intended for local/trusted use only.
- Templates must be present relative to the working directory (`web/templates/`). Run from the project root.
- The CLI and API adapters are stubs that return `ErrNotImplemented` until wired up.
- No mobile layout; designed for desktop operator use.
