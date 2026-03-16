package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultLocalEndpoint = "http://localhost:11434"
	defaultLocalModel    = "llama3"
)

// LocalConfig holds configuration for the local Ollama LLM backend.
type LocalConfig struct {
	Endpoint string
	Model    string
}

// Local is an LLM backend that connects to a local Ollama instance.
type Local struct {
	config LocalConfig
	client *http.Client
}

// NewLocal creates a new Local LLM backend with the given config.
// No API key is required for local operation. Defaults are applied for
// empty Endpoint and Model fields.
func NewLocal(config LocalConfig) *Local {
	if config.Endpoint == "" {
		config.Endpoint = defaultLocalEndpoint
	}
	if config.Model == "" {
		config.Model = defaultLocalModel
	}
	return &Local{
		config: config,
		client: &http.Client{},
	}
}

// Name returns "local".
func (l *Local) Name() string {
	return "local"
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message *ollamaMessage `json:"message,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// Chat sends messages to the local Ollama instance and returns the response.
func (l *Local) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("local: messages must not be empty")
	}

	var apiMessages []ollamaMessage
	for _, m := range messages {
		apiMessages = append(apiMessages, ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	reqBody := ollamaChatRequest{
		Model:    l.config.Model,
		Messages: apiMessages,
		Stream:   false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("local: failed to marshal request: %w", err)
	}

	url := l.config.Endpoint + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("local: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("local: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("local: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", l.handleErrorResponse(resp.StatusCode, respBody)
	}

	var ollamaResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", fmt.Errorf("local: failed to parse response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("local: ollama error: %s", ollamaResp.Error)
	}

	if ollamaResp.Message == nil {
		return "", errors.New("local: response contained no message")
	}

	return ollamaResp.Message.Content, nil
}

func (l *Local) handleErrorResponse(statusCode int, body []byte) error {
	var errResp ollamaChatResponse
	_ = json.Unmarshal(body, &errResp)

	detail := errResp.Error
	if detail == "" {
		detail = string(body)
	}

	switch {
	case statusCode == http.StatusNotFound:
		return fmt.Errorf("local: model not found (404): %s", detail)
	case statusCode >= 500:
		return fmt.Errorf("local: server error (%d): %s", statusCode, detail)
	default:
		return fmt.Errorf("local: unexpected status %d: %s", statusCode, detail)
	}
}

// ListOllamaModels queries the Ollama API at the given endpoint for installed models.
// Returns a list of model name strings (e.g. "llama3:8b", "mistral:latest").
func ListOllamaModels(ctx context.Context, endpoint string) ([]string, error) {
	if endpoint == "" {
		endpoint = defaultLocalEndpoint
	}

	url := endpoint + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("list ollama models: create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list ollama models: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list ollama models: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("list ollama models: parse response: %w", err)
	}

	names := make([]string, 0, len(tagsResp.Models))
	for _, m := range tagsResp.Models {
		names = append(names, m.Name)
	}

	return names, nil
}
