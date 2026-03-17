# CoachLM

Your personal AI running coach that lives on your computer.

Connect your Strava account, chat with an AI coach that knows your training history, and get personalized insights about your running — all without your data ever leaving your machine.

---

## What can it do?

**💬 Chat with a smart running coach**
- Ask about training plans, recovery, race strategies, or anything running-related
- The AI knows your actual running history — past workouts, pace trends, heart rate patterns
- Save the best advice so the AI remembers it for future conversations

**🏃 See your running history**
- Auto-syncs all your Strava activities
- View recent runs with distance, pace, heart rate, and cadence
- Import FIT files directly from Garmin, Wahoo, and other devices

**🤖 Choose your AI**
- Use Claude, ChatGPT, or run a local model with Ollama (no internet needed after setup)
- Your API keys are stored securely — never sent anywhere

**🔒 Your data stays yours**
- Everything runs locally on your computer
- No cloud accounts, no subscription, no data mining
- SQLite database means you own your data — export it anytime

---

## Get started

### Download

Head to the [Releases page](https://github.com/Alchaph/coachlm-vibe/releases/latest) and grab the version for your computer:

| Your computer | Download this |
|---------------|---------------|
| Windows | `coachlm-windows.exe` |
| macOS | `coachlm-macos.zip` |
| Linux | `coachlm-linux` |

### Run it

**Windows**: Just double-click the `.exe` file

**macOS**: Unzip, then drag `CoachLM.app` to your Applications folder and open it

**Linux**: 
```bash
chmod +x coachlm-linux
./coachlm-linux
```

*Linux users may need to install GTK3 and WebKit2GTK first — see below*

### First time setup

When you launch CoachLM for the first time, a setup wizard will help you:
1. Connect your Strava account (optional)
2. Choose an AI model (Claude, ChatGPT, or local Ollama)
3. Set up your athlete profile (age, max heart rate, goals)

That's it — you're ready to chat with your coach!

---

## Linux users only

If you're on Linux, you'll need these libraries:

```bash
# Ubuntu / Debian
sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0

# Arch
sudo pacman -S webkit2gtk-4.1

# Fedora
sudo dnf install gtk3 webkit2gtk4.1
```

---

## Tech details (for the curious)

CoachLM is built with:
- **Backend**: Go with Wails v2
- **Frontend**: Svelte + TypeScript
- **Database**: SQLite (pure Go, no external dependencies)
- **LLM**: Supports Claude, OpenAI, and Ollama APIs

All your data is encrypted with AES-256-GCM — your API keys and Strava tokens are safe.

---

## Building from source

If you're a developer and want to build it yourself:

```bash
git clone https://github.com/Alchaph/coachlm-vibe.git
cd coachlm-vibe
make build
```

You'll need:
- Go 1.24+
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

---

## License

Unlicensed
