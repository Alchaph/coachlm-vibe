package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultClaudeModel     = "claude-sonnet-4-20250514"
	defaultClaudeBaseURL   = "https://api.anthropic.com"
	claudeAPIVersion       = "2023-06-01"
	claudeDefaultMaxTokens = 4096
)

type ClaudeConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Claude struct {
	config ClaudeConfig
	client *http.Client
}

func NewClaude(config ClaudeConfig) (*Claude, error) {
	if config.APIKey == "" {
		return nil, errors.New("claude: API key is required")
	}
	if config.Model == "" {
		config.Model = defaultClaudeModel
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultClaudeBaseURL
	}
	return &Claude{
		config: config,
		client: &http.Client{},
	}, nil
}

func (c *Claude) Name() string {
	return "claude"
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []claudeContentBlock `json:"content"`
	Error   *claudeError         `json:"error,omitempty"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *Claude) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("claude: messages must not be empty")
	}

	var systemPrompt string
	var apiMessages []claudeMessage

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemPrompt = m.Content
			continue
		}
		apiMessages = append(apiMessages, claudeMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	if len(apiMessages) == 0 {
		return "", errors.New("claude: at least one non-system message is required")
	}

	reqBody := claudeRequest{
		Model:     c.config.Model,
		MaxTokens: claudeDefaultMaxTokens,
		System:    systemPrompt,
		Messages:  apiMessages,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("claude: failed to marshal request: %w", err)
	}

	url := c.config.BaseURL + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("claude: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("claude: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.handleErrorResponse(resp.StatusCode, respBody)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("claude: failed to parse response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", errors.New("claude: response contained no content blocks")
	}

	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", errors.New("claude: response contained no text content")
}

func (c *Claude) handleErrorResponse(statusCode int, body []byte) error {
	var errResp claudeResponse
	_ = json.Unmarshal(body, &errResp)

	var detail string
	if errResp.Error != nil {
		detail = errResp.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("claude: authentication failed (401): %s", detail)
	case http.StatusTooManyRequests:
		return fmt.Errorf("claude: rate limited (429): %s", detail)
	case http.StatusBadRequest:
		return fmt.Errorf("claude: bad request (400): %s", detail)
	default:
		if statusCode >= 500 {
			return fmt.Errorf("claude: server error (%d): %s", statusCode, detail)
		}
		return fmt.Errorf("claude: unexpected status %d: %s", statusCode, detail)
	}
}
