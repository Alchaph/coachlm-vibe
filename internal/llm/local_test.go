package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewLocal_Defaults(t *testing.T) {
	l := NewLocal(LocalConfig{})
	if l.config.Endpoint != defaultLocalEndpoint {
		t.Errorf("expected endpoint %q, got %q", defaultLocalEndpoint, l.config.Endpoint)
	}
	if l.config.Model != defaultLocalModel {
		t.Errorf("expected model %q, got %q", defaultLocalModel, l.config.Model)
	}
}

func TestNewLocal_CustomConfig(t *testing.T) {
	l := NewLocal(LocalConfig{
		Endpoint: "http://remote:11434",
		Model:    "mistral",
	})
	if l.config.Endpoint != "http://remote:11434" {
		t.Errorf("expected endpoint %q, got %q", "http://remote:11434", l.config.Endpoint)
	}
	if l.config.Model != "mistral" {
		t.Errorf("expected model %q, got %q", "mistral", l.config.Model)
	}
}

func TestLocal_Name(t *testing.T) {
	l := NewLocal(LocalConfig{})
	if l.Name() != "local" {
		t.Errorf("expected name %q, got %q", "local", l.Name())
	}
}

func TestLocal_ImplementsLLMInterface(t *testing.T) {
	var _ LLM = NewLocal(LocalConfig{})
}

func TestLocal_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("expected path /api/chat, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var req ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != "llama3" {
			t.Errorf("expected model llama3, got %q", req.Model)
		}
		if req.Stream != false {
			t.Error("expected stream to be false")
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected 1 user message, got %+v", req.Messages)
		}

		resp := ollamaChatResponse{
			Message: &ollamaMessage{Role: "assistant", Content: "Run easy 5k today."},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})

	result, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "What should I do today?"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Run easy 5k today." {
		t.Errorf("expected %q, got %q", "Run easy 5k today.", result)
	}
}

func TestLocal_Chat_WithSystemMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" || req.Messages[0].Content != "You are a running coach." {
			t.Errorf("expected system message first, got %+v", req.Messages[0])
		}
		if req.Messages[1].Role != "user" {
			t.Errorf("expected user message second, got %+v", req.Messages[1])
		}

		resp := ollamaChatResponse{
			Message: &ollamaMessage{Role: "assistant", Content: "Great workout plan!"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	result, err := l.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are a running coach."},
		{Role: RoleUser, Content: "Plan my week"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Great workout plan!" {
		t.Errorf("expected %q, got %q", "Great workout plan!", result)
	}
}

func TestLocal_Chat_EmptyMessages(t *testing.T) {
	l := NewLocal(LocalConfig{})
	_, err := l.Chat(context.Background(), []Message{})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
	if !strings.Contains(err.Error(), "messages must not be empty") {
		t.Errorf("expected 'messages must not be empty' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_ConnectionRefused(t *testing.T) {
	l := NewLocal(LocalConfig{Endpoint: "http://127.0.0.1:1"})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
	if !strings.Contains(err.Error(), "local: request failed") {
		t.Errorf("expected 'local: request failed' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected 'failed to parse response' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: &ollamaMessage{Role: "assistant", Content: "ok"},
		})
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := l.Chat(ctx, []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestLocal_Chat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
		})
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "server error (500)") {
		t.Errorf("expected 'server error (500)' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_ModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "model 'nonexistent' not found",
		})
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL, Model: "nonexistent"})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for model not found")
	}
	if !strings.Contains(err.Error(), "model not found (404)") {
		t.Errorf("expected 'model not found (404)' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_OllamaErrorInBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Error: "model requires more memory than available",
		})
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for ollama error in body")
	}
	if !strings.Contains(err.Error(), "ollama error") {
		t.Errorf("expected 'ollama error' in error, got %q", err.Error())
	}
}

func TestLocal_Chat_NoMessageInResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{})
	}))
	defer server.Close()

	l := NewLocal(LocalConfig{Endpoint: server.URL})
	_, err := l.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for no message in response")
	}
	if !strings.Contains(err.Error(), "no message") {
		t.Errorf("expected 'no message' in error, got %q", err.Error())
	}
}

func TestListOllamaModels_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected path /api/tags, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models":[{"name":"llama3:8b"},{"name":"mistral:latest"},{"name":"codellama:7b"}]}`))
	}))
	defer server.Close()

	models, err := ListOllamaModels(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}
	want := []string{"llama3:8b", "mistral:latest", "codellama:7b"}
	for i, m := range models {
		if m != want[i] {
			t.Errorf("model[%d]: expected %q, got %q", i, want[i], m)
		}
	}
}

func TestListOllamaModels_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	models, err := ListOllamaModels(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected 0 models, got %d", len(models))
	}
}

func TestListOllamaModels_DefaultEndpoint(t *testing.T) {
	_, err := ListOllamaModels(context.Background(), "")
	if err == nil {
		t.Skip("Ollama is running locally; skipping connection error test")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected 'request failed' in error, got %q", err.Error())
	}
}

func TestListOllamaModels_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	_, err := ListOllamaModels(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("expected 'unexpected status 500' in error, got %q", err.Error())
	}
}

func TestListOllamaModels_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid`))
	}))
	defer server.Close()

	_, err := ListOllamaModels(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
	if !strings.Contains(err.Error(), "parse response") {
		t.Errorf("expected 'parse response' in error, got %q", err.Error())
	}
}

func TestListOllamaModels_ConnectionRefused(t *testing.T) {
	_, err := ListOllamaModels(context.Background(), "http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("expected 'request failed' in error, got %q", err.Error())
	}
}
