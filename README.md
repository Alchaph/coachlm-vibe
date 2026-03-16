# CoachLM

AI-powered running coach desktop app. Syncs your Strava activities, builds a persistent athlete context, and provides personalized coaching through Claude, ChatGPT, or a local Ollama model — all running locally on your machine.

## Installation

Download the latest release for your platform from the [Releases page](https://github.com/Alchaph/coachlm-vibe/releases/latest).

| Platform | File | Instructions |
|----------|------|--------------|
| **Linux** | `coachlm-linux` | `chmod +x coachlm-linux && ./coachlm-linux` |
| **macOS** | `coachlm-macos.zip` | Unzip, move `CoachLM.app` to Applications, open |
| **Windows** | `coachlm-windows.exe` | Download and run |

### Linux dependencies

CoachLM requires GTK3 and WebKit2GTK at runtime:

```bash
# Ubuntu / Debian
sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0

# Arch
sudo pacman -S webkit2gtk-4.1

# Fedora
sudo dnf install gtk3 webkit2gtk4.1
```

## Features

**Coaching Chat**
- Chat with your AI running coach — context-aware responses based on your actual training data
- Pin useful insights from conversations to build a persistent knowledge base
- Markdown rendering with code blocks, lists, and formatting

**Strava Integration**
- OAuth2 authentication flow with encrypted token storage
- Sync recent activities including distance, pace, heart rate, and cadence
- Webhook support for automatic activity ingestion
- Activity stream data (HR, pace, cadence time series)

**LLM Backends**
- **Claude** (Anthropic) — configurable model selection
- **ChatGPT** (OpenAI) — configurable model selection
- **Ollama** (local) — browse and select from installed models, no API key needed

**Context Engine**
- Assembles athlete profile, recent training load summary, and pinned insights into every prompt
- Token-budget-aware: compresses older summaries, never drops pinned insights
- Ensures the LLM always has relevant context without exceeding limits

**Additional**
- Activity dashboard with recent runs
- FIT file import for manual uploads (Garmin, Wahoo, etc.)
- First-run onboarding wizard for initial setup
- Athlete profile management (age, max HR, threshold pace, goals, injury history)
- AES-256-GCM encryption for all API keys and OAuth tokens
- SQLite-based — fully local, no cloud dependency, data stays on your machine

## Tech Stack

- **Backend**: Go 1.24, [Wails v2](https://wails.io) (desktop framework)
- **Frontend**: Svelte + TypeScript + Vite
- **Database**: SQLite (via modernc.org/sqlite, pure Go)
- **LLM**: Claude API, OpenAI API, Ollama (local)

## Building from source

### Prerequisites

- Go 1.24+
- Node.js 20+
- npm
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Linux**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (or 4.0)

### Build

```bash
git clone https://github.com/Alchaph/coachlm-vibe.git
cd coachlm-vibe
make build
```

Or without Make:

```bash
wails build                    # webkit2gtk-4.0
wails build -tags webkit2_41   # webkit2gtk-4.1
```

Output binary: `build/bin/coachlm`

### Development

```bash
make dev
```

### Run tests

```bash
go test ./...
```

## Project Structure

```
├── app.go                # Wails app bindings
├── main.go               # Entry point
├── wails.json            # Wails build configuration
├── internal/
│   ├── strava/           # Strava OAuth, webhook, activity sync
│   ├── storage/          # SQLite layer (activities, chat, settings, tokens, profiles)
│   ├── context/          # Context engine + prompt assembler
│   ├── llm/              # LLM router (Claude / OpenAI / Ollama)
│   └── fit/              # FIT file parser
├── frontend/             # Svelte frontend (Chat, Dashboard, Settings, Onboarding)
├── build/                # Build assets (icons, platform configs)
├── stories/              # Feature stories (spec + status tracking)
└── .github/workflows/    # CI + release pipelines
```

## Configuration

On first launch, the onboarding wizard guides you through initial setup. CoachLM defaults to local Ollama if no API keys are configured. LLM backends and Strava OAuth credentials can be changed at any time through the Settings tab.

## License

Unlicensed
