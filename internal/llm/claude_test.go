package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClaude_EmptyAPIKey(t *testing.T) {
	_, err := NewClaude(ClaudeConfig{})
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestNewClaude_Defaults(t *testing.T) {
	c, err := NewClaude(ClaudeConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.config.Model != defaultClaudeModel {
		t.Errorf("expected model %q, got %q", defaultClaudeModel, c.config.Model)
	}
	if c.config.BaseURL != defaultClaudeBaseURL {
		t.Errorf("expected base URL %q, got %q", defaultClaudeBaseURL, c.config.BaseURL)
	}
}

func TestClaude_Name(t *testing.T) {
	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key"})
	if c.Name() != "claude" {
		t.Errorf("expected name %q, got %q", "claude", c.Name())
	}
}

func TestClaude_ImplementsLLMInterface(t *testing.T) {
	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key"})
	var _ LLM = c
}

func TestClaude_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("expected path /v1/messages, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header test-key, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req claudeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.System != "You are a coach." {
			t.Errorf("expected system prompt, got %q", req.System)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected 1 user message, got %+v", req.Messages)
		}

		resp := claudeResponse{
			Content: []claudeContentBlock{
				{Type: "text", Text: "Go run 5k today."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	result, err := c.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are a coach."},
		{Role: RoleUser, Content: "What should I do today?"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Go run 5k today." {
		t.Errorf("expected %q, got %q", "Go run 5k today.", result)
	}
}

func TestClaude_Chat_EmptyMessages(t *testing.T) {
	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key"})
	_, err := c.Chat(context.Background(), []Message{})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestClaude_Chat_OnlySystemMessage(t *testing.T) {
	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key"})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "System only"},
	})
	if err == nil {
		t.Fatal("expected error when only system message provided")
	}
}

func TestClaude_Chat_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"type":    "rate_limit_error",
				"message": "too many requests",
			},
		})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 429")
	}
	expected := "claude: rate limited (429): too many requests"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClaude_Chat_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"type":    "authentication_error",
				"message": "invalid api key",
			},
		})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "bad-key", BaseURL: server.URL})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	expected := "claude: authentication failed (401): invalid api key"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClaude_Chat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"type":    "server_error",
				"message": "internal error",
			},
		})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	expected := "claude: server error (500): internal error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClaude_Chat_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

func TestClaude_Chat_EmptyContentBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse{Content: []claudeContentBlock{}})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for empty content blocks")
	}
}

func TestClaude_Chat_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse{
			Content: []claudeContentBlock{{Type: "text", Text: "ok"}},
		})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Chat(ctx, []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestClaude_Chat_NoSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req claudeRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.System != "" {
			t.Errorf("expected empty system, got %q", req.System)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claudeResponse{
			Content: []claudeContentBlock{{Type: "text", Text: "response"}},
		})
	}))
	defer server.Close()

	c, _ := NewClaude(ClaudeConfig{APIKey: "test-key", BaseURL: server.URL})
	result, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "response" {
		t.Errorf("expected %q, got %q", "response", result)
	}
}
