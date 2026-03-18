# CoachLM

Your personal AI running coach

Connect your Strava account, chat with an AI coach that knows your training history, and get personalized insights about your running

---

### Download

Head to the [Releases page](https://github.com/Alchaph/coachlm-vibe/releases/latest) and grab the version for your computer:

| Your computer | Download this |
|---------------|---------------|
| Windows | `coachlm-windows.exe` |
| macOS | `coachlm-macos.zip` |
| Linux | `coachlm-linux` |

### Run it

**Windows**: Just double-click the `.exe` file

**macOS**: Unzip, then drag `coachlm.app` to your Applications folder and open it

```bash
xattr -d com.apple.quarantine /Applications/coachlm.app
```

**Linux**: 
```bash
chmod +x coachlm-linux
./coachlm-linux
```

*Linux users may need to install GTK3 and WebKit2GTK first — see below*

### First time setup

When you launch CoachLM for the first time, a setup wizard will help you:
1. Connect your Strava account (optional)
2. Configure your Ollama endpoint and model
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
