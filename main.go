package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"coachlm/internal/llm"
	"coachlm/internal/storage"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("get home directory: %v", err)
	}

	dataDir := filepath.Join(homeDir, ".coachlm")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatalf("create data directory: %v", err)
	}

	db, err := storage.New(filepath.Join(dataDir, "coachlm.db"))
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	llmClient, err := createLLMClient(db)
	if err != nil {
		log.Fatalf("create LLM client: %v", err)
	}

	app := NewApp(db, llmClient)

	if err := wails.Run(&options.App{
		Title:  "CoachLM",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}); err != nil {
		fmt.Println("Error:", err.Error())
	}
}

func createLLMClient(db *storage.DB) (llm.LLM, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	if settings == nil {
		return llm.NewLocal(llm.LocalConfig{}), nil
	}

	switch settings.ActiveLLM {
	case "free":
		client, err := llm.NewFree(llm.FreeConfig{})
		if err != nil {
			return nil, fmt.Errorf("create free LLM: %w", err)
		}
		return client, nil
	case "claude":
		return llm.NewClaude(llm.ClaudeConfig{APIKey: string(settings.ClaudeAPIKey), Model: settings.ClaudeModel})
	case "openai":
		return llm.NewOpenAI(llm.OpenAIConfig{APIKey: string(settings.OpenAIAPIKey), Model: settings.OpenAIModel})
	default:
		return llm.NewLocal(llm.LocalConfig{Endpoint: settings.OllamaEndpoint, Model: settings.OllamaModel}), nil
	}
}
