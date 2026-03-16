package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// App struct holds the application state and dependencies.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the Wails runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SendMessage sends a user message to the LLM backend and returns the response.
// This is a stub that echoes the message until the LLM router (S08/S09) is wired up.
func (a *App) SendMessage(message string) (string, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "", errors.New("message cannot be empty")
	}
	return fmt.Sprintf("Echo: %s", trimmed), nil
}
