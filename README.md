# CoachLM

AI-powered running coach desktop app — syncs your Strava data, understands your training history, and provides personalized coaching via LLM.

## Features

- Strava activity sync (OAuth2 + webhooks)
- LLM-powered coaching chat (Claude, ChatGPT, or local Ollama)
- Context engine assembles athlete profile, training load, and pinned insights into every prompt
- Activity dashboard with recent runs
- FIT file import for manual uploads
- Encrypted token and API key storage (AES-256-GCM)
- SQLite-based — fully local, no cloud required

## Prerequisites

- Go 1.24+
- Node.js 20+
- npm
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Linux**: `gtk3` and `webkit2gtk` (4.0 or 4.1)

## Quick Start

```bash
git clone https://github.com/YOUR_USERNAME/coachlm.git
cd coachlm
make dev
```

Or without Make:

```bash
wails dev                    # if you have webkit2gtk-4.0
wails dev -tags webkit2_41   # if you have webkit2gtk-4.1 (Arch, Fedora 39+)
```

## Build

```bash
make build
```

Or without Make:

```bash
wails build                    # webkit2gtk-4.0
wails build -tags webkit2_41   # webkit2gtk-4.1
```

Output binary location: `build/bin/`

## Run Tests

```bash
go test ./...
```

## Project Structure

```
├── app.go              # Wails bindings
├── main.go             # Entry point
├── internal/
│   ├── strava/         # Strava API client + OAuth + webhook
│   ├── storage/        # SQLite layer (all CRUD)
│   ├── context/        # Context engine + prompt assembler
│   ├── llm/            # LLM router (Claude / OpenAI / Ollama)
│   └── fit/            # FIT file parser
├── frontend/           # Svelte frontend
└── stories/            # Feature stories
```

## Configuration

On first launch, the app defaults to local Ollama. Configure LLM backends and Strava OAuth credentials through the app settings.

## License

Unlicensed
