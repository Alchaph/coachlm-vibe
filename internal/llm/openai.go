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
	defaultOpenAIModel   = "gpt-4o"
	defaultOpenAIBaseURL = "https://api.openai.com"
)

type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

type OpenAI struct {
	config OpenAIConfig
	client *http.Client
}

func NewOpenAI(config OpenAIConfig) (*OpenAI, error) {
	if config.APIKey == "" {
		return nil, errors.New("openai: API key is required")
	}
	if config.Model == "" {
		config.Model = defaultOpenAIModel
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultOpenAIBaseURL
	}
	return &OpenAI{
		config: config,
		client: &http.Client{},
	}, nil
}

func (o *OpenAI) Name() string {
	return "openai"
}

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Error   *openaiError   `json:"error,omitempty"`
}

type openaiChoice struct {
	Message openaiMessage `json:"message"`
}

type openaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (o *OpenAI) Chat(ctx context.Context, messages []Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("openai: messages must not be empty")
	}

	var apiMessages []openaiMessage
	for _, m := range messages {
		apiMessages = append(apiMessages, openaiMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	reqBody := openaiRequest{
		Model:    o.config.Model,
		Messages: apiMessages,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openai: failed to marshal request: %w", err)
	}

	url := o.config.BaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("openai: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", o.handleErrorResponse(resp.StatusCode, respBody)
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return "", fmt.Errorf("openai: failed to parse response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", errors.New("openai: response contained no choices")
	}

	return openaiResp.Choices[0].Message.Content, nil
}

func (o *OpenAI) handleErrorResponse(statusCode int, body []byte) error {
	var errResp openaiResponse
	_ = json.Unmarshal(body, &errResp)

	var detail string
	if errResp.Error != nil {
		detail = errResp.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("openai: authentication failed (401): %s", detail)
	case http.StatusTooManyRequests:
		return fmt.Errorf("openai: rate limited (429): %s", detail)
	case http.StatusBadRequest:
		return fmt.Errorf("openai: bad request (400): %s", detail)
	default:
		if statusCode >= 500 {
			return fmt.Errorf("openai: server error (%d): %s", statusCode, detail)
		}
		return fmt.Errorf("openai: unexpected status %d: %s", statusCode, detail)
	}
}
