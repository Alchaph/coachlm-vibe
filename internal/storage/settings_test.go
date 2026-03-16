package storage

import (
	"bytes"
	"testing"
)

func TestGetSettings_EmptyDB(t *testing.T) {
	db := newTestDB(t)

	s, err := db.GetSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Fatal("expected nil settings on empty DB")
	}
}

func TestSaveAndGetSettings_RoundTrip(t *testing.T) {
	db := newTestDB(t)

	want := &Settings{
		ClaudeAPIKey:   []byte("encrypted-claude-key"),
		OpenAIAPIKey:   []byte("encrypted-openai-key"),
		ActiveLLM:      "claude",
		OllamaEndpoint: "http://localhost:11434",
	}

	if err := db.SaveSettings(want); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	got, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil settings")
	}

	if !bytes.Equal(got.ClaudeAPIKey, want.ClaudeAPIKey) {
		t.Errorf("ClaudeAPIKey = %q, want %q", got.ClaudeAPIKey, want.ClaudeAPIKey)
	}
	if !bytes.Equal(got.OpenAIAPIKey, want.OpenAIAPIKey) {
		t.Errorf("OpenAIAPIKey = %q, want %q", got.OpenAIAPIKey, want.OpenAIAPIKey)
	}
	if got.ActiveLLM != want.ActiveLLM {
		t.Errorf("ActiveLLM = %q, want %q", got.ActiveLLM, want.ActiveLLM)
	}
	if got.OllamaEndpoint != want.OllamaEndpoint {
		t.Errorf("OllamaEndpoint = %q, want %q", got.OllamaEndpoint, want.OllamaEndpoint)
	}
}

func TestSaveSettings_Upsert(t *testing.T) {
	db := newTestDB(t)

	first := &Settings{
		ClaudeAPIKey:   []byte("key-v1"),
		OpenAIAPIKey:   []byte("openai-v1"),
		ActiveLLM:      "claude",
		OllamaEndpoint: "http://localhost:11434",
	}
	if err := db.SaveSettings(first); err != nil {
		t.Fatalf("SaveSettings first: %v", err)
	}

	second := &Settings{
		ClaudeAPIKey:   []byte("key-v2"),
		OpenAIAPIKey:   []byte("openai-v2"),
		ActiveLLM:      "openai",
		OllamaEndpoint: "http://remote:11434",
	}
	if err := db.SaveSettings(second); err != nil {
		t.Fatalf("SaveSettings second: %v", err)
	}

	got, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	if !bytes.Equal(got.ClaudeAPIKey, second.ClaudeAPIKey) {
		t.Errorf("ClaudeAPIKey = %q, want %q", got.ClaudeAPIKey, second.ClaudeAPIKey)
	}
	if !bytes.Equal(got.OpenAIAPIKey, second.OpenAIAPIKey) {
		t.Errorf("OpenAIAPIKey = %q, want %q", got.OpenAIAPIKey, second.OpenAIAPIKey)
	}
	if got.ActiveLLM != second.ActiveLLM {
		t.Errorf("ActiveLLM = %q, want %q", got.ActiveLLM, second.ActiveLLM)
	}
	if got.OllamaEndpoint != second.OllamaEndpoint {
		t.Errorf("OllamaEndpoint = %q, want %q", got.OllamaEndpoint, second.OllamaEndpoint)
	}
}

func TestSaveSettings_InvalidActiveLLM(t *testing.T) {
	db := newTestDB(t)

	invalid := []string{"gpt4", "anthropic", "", "LOCAL", "Claude"}
	for _, llm := range invalid {
		s := &Settings{
			ActiveLLM:      llm,
			OllamaEndpoint: "http://localhost:11434",
		}
		if err := db.SaveSettings(s); err == nil {
			t.Errorf("expected error for ActiveLLM=%q, got nil", llm)
		}
	}
}

func TestSaveSettings_ValidActiveLLMValues(t *testing.T) {
	db := newTestDB(t)

	for _, llm := range []string{"claude", "openai", "local"} {
		s := &Settings{
			ActiveLLM:      llm,
			OllamaEndpoint: "http://localhost:11434",
		}
		if err := db.SaveSettings(s); err != nil {
			t.Errorf("SaveSettings with ActiveLLM=%q: %v", llm, err)
		}

		got, err := db.GetSettings()
		if err != nil {
			t.Fatalf("GetSettings: %v", err)
		}
		if got.ActiveLLM != llm {
			t.Errorf("ActiveLLM = %q, want %q", got.ActiveLLM, llm)
		}
	}
}

func TestSaveSettings_NilAPIKeys(t *testing.T) {
	db := newTestDB(t)

	s := &Settings{
		ClaudeAPIKey:   nil,
		OpenAIAPIKey:   nil,
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
	}
	if err := db.SaveSettings(s); err != nil {
		t.Fatalf("SaveSettings with nil keys: %v", err)
	}

	got, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got.ClaudeAPIKey != nil {
		t.Errorf("ClaudeAPIKey = %v, want nil", got.ClaudeAPIKey)
	}
	if got.OpenAIAPIKey != nil {
		t.Errorf("OpenAIAPIKey = %v, want nil", got.OpenAIAPIKey)
	}
}

func TestSaveSettings_NilSettings(t *testing.T) {
	db := newTestDB(t)

	if err := db.SaveSettings(nil); err == nil {
		t.Error("expected error for nil settings, got nil")
	}
}
