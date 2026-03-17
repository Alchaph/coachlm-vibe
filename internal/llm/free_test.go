package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFree_Name(t *testing.T) {
	client, err := NewFree(FreeConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}
	if client.Name() != "free" {
		t.Errorf("Name() = %q, want %q", client.Name(), "free")
	}
}

func TestFree_EmptyMessages(t *testing.T) {
	client, err := NewFree(FreeConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{})
	if err == nil {
		t.Error("expected error for empty messages")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestFree_NoSystemMessage(t *testing.T) {
	client, err := NewFree(FreeConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are a coach"},
	})
	if err == nil {
		t.Error("expected error for only system message")
	}
	if !strings.Contains(err.Error(), "at least one non-system message") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestFree_APIKeyRequired(t *testing.T) {
	_, err := NewFree(FreeConfig{})
	if err == nil {
		t.Error("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key is required") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestFree_SystemPromptSent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("x-goog-api-key")
		if auth != "test-key" {
			t.Errorf("missing API key header")
		}
		if auth != "test-key" {
			t.Errorf("wrong API key: got %q", auth)
		}

		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		config, ok := reqBody["generationConfig"].(map[string]interface{})
		if !ok {
			t.Error("missing generationConfig in request")
			return
		}

		sysInstr, ok := config["systemInstruction"].(string)
		if !ok {
			t.Error("missing systemInstruction")
			return
		}

		if sysInstr != "You are a helpful coach" {
			t.Errorf("wrong system instruction: got %q", sysInstr)
		}

		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Response text"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewFree(FreeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	result, err := client.Chat(context.Background(), []Message{
		{Role: RoleSystem, Content: "You are a helpful coach"},
		{Role: RoleUser, Content: "Hello"},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result != "Response text" {
		t.Errorf("response = %q, want %q", result, "Response text")
	}
}

func TestFree_RateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Rate limit exceeded",
				"status":  429,
			},
		})
	}))
	defer server.Close()

	client, err := NewFree(FreeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	})
	if err == nil {
		t.Error("expected rate limit error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("wrong error message: %v", err)
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestFree_UnreachableError(t *testing.T) {
	client, err := NewFree(FreeConfig{
		APIKey:  "test-key",
		BaseURL: "http://invalid-nonexistent-host-12345:9999",
	})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "Hello"},
	})
	if err == nil {
		t.Error("expected unreachable error")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestFree_DefaultModel(t *testing.T) {
	client, err := NewFree(FreeConfig{
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	if client.config.Model != defaultFreeModel {
		t.Errorf("Model = %q, want %q", client.config.Model, defaultFreeModel)
	}
}

func TestFree_MultiTurnConversation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request: %v", err)
			return
		}

		contents, ok := reqBody["contents"].([]interface{})
		if !ok || len(contents) != 3 {
			t.Errorf("wrong number of contents: got %d, want 3", len(contents))
			return
		}

		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Final response"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewFree(FreeConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewFree: %v", err)
	}

	result, err := client.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "First message"},
		{Role: RoleAssistant, Content: "First response"},
		{Role: RoleUser, Content: "Second message"},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result != "Final response" {
		t.Errorf("response = %q, want %q", result, "Final response")
	}
}
