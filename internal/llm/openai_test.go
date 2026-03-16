package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpenAI_EmptyAPIKey(t *testing.T) {
	_, err := NewOpenAI(OpenAIConfig{})
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestNewOpenAI_Defaults(t *testing.T) {
	o, err := NewOpenAI(OpenAIConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.config.Model != defaultOpenAIModel {
		t.Errorf("expected model %q, got %q", defaultOpenAIModel, o.config.Model)
	}
	if o.config.BaseURL != defaultOpenAIBaseURL {
		t.Errorf("expected base URL %q, got %q", defaultOpenAIBaseURL, o.config.BaseURL)
	}
}

func TestOpenAI_Name(t *testing.T) {
	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key"})
	if o.Name() != "openai" {
		t.Errorf("expected name %q, got %q", "openai", o.Name())
	}
}

func TestOpenAI_ImplementsLLMInterface(t *testing.T) {
	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key"})
	var _ LLM = o
}

func TestOpenAI_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req openaiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected 1 user message, got %+v", req.Messages)
		}

		resp := openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{Role: "assistant", Content: "Go run 5k today."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	o, _ := NewOpenAI(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	result, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "What should I do today?"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Go run 5k today." {
		t.Errorf("expected %q, got %q", "Go run 5k today.", result)
	}
}

func TestOpenAI_Chat_WithSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openaiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" || req.Messages[0].Content != "You are a coach." {
			t.Errorf("expected system message first, got %+v", req.Messages[0])
		}
		if req.Messages[1].Role != "user" {
			t.Errorf("expected user message second, got %+v", req.Messages[1])
		}

		resp := openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{Role: "assistant", Content: "Rest day."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	result, err := o.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are a coach."},
		{Role: RoleUser, Content: "What should I do today?"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Rest day." {
		t.Errorf("expected %q, got %q", "Rest day.", result)
	}
}

func TestOpenAI_Chat_EmptyMessages(t *testing.T) {
	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key"})
	_, err := o.Chat(context.Background(), []Message{})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestOpenAI_Chat_RateLimit(t *testing.T) {
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

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 429")
	}
	expected := "openai: rate limited (429): too many requests"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestOpenAI_Chat_AuthFailure(t *testing.T) {
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

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "bad-key", BaseURL: server.URL})
	_, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	expected := "openai: authentication failed (401): invalid api key"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestOpenAI_Chat_ServerError(t *testing.T) {
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

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	expected := "openai: server error (500): internal error"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestOpenAI_Chat_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

func TestOpenAI_Chat_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiResponse{Choices: []openaiChoice{}})
	}))
	defer server.Close()

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	_, err := o.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestOpenAI_Chat_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiResponse{
			Choices: []openaiChoice{
				{Message: openaiMessage{Role: "assistant", Content: "ok"}},
			},
		})
	}))
	defer server.Close()

	o, _ := NewOpenAI(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := o.Chat(ctx, []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
