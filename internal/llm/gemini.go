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
	defaultGeminiModel   = "gemini-2.0-flash"
	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"
	// Built-in Gemini API key - injected at build time with -ldflags
	// Usage: -ldflags "-X coachlm/internal/llm.builtinGeminiApiKey=YOUR_KEY"
	builtinGeminiApiKey = ""
)

type GeminiConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Gemini struct {
	config GeminiConfig
	client *http.Client
}

func NewGemini(config GeminiConfig) (*Gemini, error) {
	if config.Model == "" {
		config.Model = defaultGeminiModel
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultGeminiBaseURL
	}
	if config.APIKey == "" {
		config.APIKey = os.Getenv("GEMINI_API_KEY")
	}
	// Fall back to built-in key if available
	if config.APIKey == "" && builtinGeminiApiKey != "" {
		config.APIKey = builtinGeminiApiKey
	}
	if config.APIKey == "" {
		return nil, errors.New("gemini: API key is required. Please set GEMINI_API_KEY environment variable or use a local Ollama backend")
	}
	return &Gemini{
		config: config,
		client: &http.Client{},
	}, nil
}

func (g *Gemini) Name() string {
	return "gemini"
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig *geminiConfig   `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiConfig struct {
	SystemInstruction string `json:"systemInstruction,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content *geminiContent `json:"content"`
}

type geminiErrorResponse struct {
	Error *geminiError `json:"error"`
}

type geminiError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (g *Gemini) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("gemini: messages must not be empty")
	}

	var systemInstruction string
	var apiContents []geminiContent

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemInstruction = m.Content
			continue
		}
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		apiContents = append(apiContents, geminiContent{
			Parts: []geminiPart{{Text: m.Content}},
			Role:  role,
		})
	}

	if len(apiContents) == 0 {
		return "", errors.New("gemini: at least one non-system message is required")
	}

	reqBody := geminiRequest{
		Contents: apiContents,
	}
	if systemInstruction != "" {
		reqBody.GenerationConfig = &geminiConfig{
			SystemInstruction: systemInstruction,
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", g.config.BaseURL, g.config.Model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", g.handleErrorResponse(resp.StatusCode, respBody)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("gemini: failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return "", errors.New("gemini: response contained no candidates")
	}

	if geminiResp.Candidates[0].Content == nil {
		return "", errors.New("gemini: candidate contained no content")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: content contained no parts")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (g *Gemini) handleErrorResponse(statusCode int, body []byte) error {
	var errResp geminiErrorResponse
	_ = json.Unmarshal(body, &errResp)

	var detail string
	if errResp.Error != nil {
		detail = errResp.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("gemini: authentication failed (401): %s", detail)
	case http.StatusTooManyRequests:
		return fmt.Errorf("gemini: rate limited (429): %s. Try again later", detail)
	case http.StatusBadRequest:
		return fmt.Errorf("gemini: bad request (400): %s", detail)
	default:
		if statusCode >= 500 {
			return fmt.Errorf("gemini: server error (%d): %s", statusCode, detail)
		}
		return fmt.Errorf("gemini: unexpected status %d: %s", statusCode, detail)
	}
}
