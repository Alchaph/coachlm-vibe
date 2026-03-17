package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	defaultFreeModel   = "gemini-2.0-flash"
	defaultFreeBaseURL = "https://generativelanguage.googleapis.com"
	// Built-in free tier API key - should be injected at build time with -ldflags
	// This is a placeholder; in production builds, use: -ldflags "-X coachlm/internal/llm.builtinFreeApiKey=YOUR_KEY"
	builtinFreeApiKey = ""
)

type FreeConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Free struct {
	config FreeConfig
	client *http.Client
}

func NewFree(config FreeConfig) (*Free, error) {
	if config.Model == "" {
		config.Model = defaultFreeModel
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultFreeBaseURL
	}
	if config.APIKey == "" {
		config.APIKey = os.Getenv("GEMINI_API_KEY")
	}
	// Fall back to built-in key if available
	if config.APIKey == "" && builtinFreeApiKey != "" {
		config.APIKey = builtinFreeApiKey
	}
	if config.APIKey == "" {
		return nil, errors.New("free: API key is required for Gemini backend. Please set GEMINI_API_KEY environment variable, or switch to Claude, OpenAI, or Ollama backend")
	}
	return &Free{
		config: config,
		client: &http.Client{},
	}, nil
}

func (f *Free) Name() string {
	return "free"
}

type freeRequest struct {
	Contents         []freeContent `json:"contents"`
	GenerationConfig *freeConfig   `json:"generationConfig,omitempty"`
}

type freeContent struct {
	Parts []freePart `json:"parts"`
	Role  string     `json:"role,omitempty"`
}

type freePart struct {
	Text string `json:"text"`
}

type freeConfig struct {
	SystemInstruction string `json:"systemInstruction,omitempty"`
}

type freeResponse struct {
	Candidates []freeCandidate `json:"candidates"`
}

type freeCandidate struct {
	Content *freeContent `json:"content"`
}

type freeErrorResponse struct {
	Error *freeError `json:"error"`
}

type freeError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (f *Free) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("free: messages must not be empty")
	}

	var systemInstruction string
	var apiContents []freeContent

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemInstruction = m.Content
			continue
		}
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		apiContents = append(apiContents, freeContent{
			Parts: []freePart{{Text: m.Content}},
			Role:  role,
		})
	}

	if len(apiContents) == 0 {
		return "", errors.New("free: at least one non-system message is required")
	}

	reqBody := freeRequest{
		Contents: apiContents,
	}
	if systemInstruction != "" {
		reqBody.GenerationConfig = &freeConfig{
			SystemInstruction: systemInstruction,
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("free: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", f.config.BaseURL, f.config.Model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("free: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", f.config.APIKey)

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("free: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("free: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", f.handleErrorResponse(resp.StatusCode, respBody)
	}

	var freeResp freeResponse
	if err := json.Unmarshal(respBody, &freeResp); err != nil {
		return "", fmt.Errorf("free: failed to parse response: %w", err)
	}

	if len(freeResp.Candidates) == 0 {
		return "", errors.New("free: response contained no candidates")
	}

	if freeResp.Candidates[0].Content == nil {
		return "", errors.New("free: candidate contained no content")
	}

	if len(freeResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("free: content contained no parts")
	}

	return freeResp.Candidates[0].Content.Parts[0].Text, nil
}

func (f *Free) handleErrorResponse(statusCode int, body []byte) error {
	var errResp freeErrorResponse
	_ = json.Unmarshal(body, &errResp)

	var detail string
	if errResp.Error != nil {
		detail = errResp.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("free: authentication failed (401): %s. Try switching to Claude, OpenAI, or Ollama backend", detail)
	case http.StatusTooManyRequests:
		return fmt.Errorf("free: rate limited (429): %s. Try again later or switch to another backend", detail)
	case http.StatusBadRequest:
		return fmt.Errorf("free: bad request (400): %s", detail)
	default:
		if statusCode >= 500 {
			return fmt.Errorf("free: server error (%d): %s. Try another backend", statusCode, detail)
		}
		return fmt.Errorf("free: unexpected status %d: %s", statusCode, detail)
	}
}
