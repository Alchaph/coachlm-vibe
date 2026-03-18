package storage

import (
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
		ActiveLLM:      "gemini",
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
		ActiveLLM:      "gemini",
		OllamaEndpoint: "http://localhost:11434",
	}
	if err := db.SaveSettings(first); err != nil {
		t.Fatalf("SaveSettings first: %v", err)
	}

	second := &Settings{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://remote:11434",
	}
	if err := db.SaveSettings(second); err != nil {
		t.Fatalf("SaveSettings second: %v", err)
	}

	got, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
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

	invalid := []string{"claude", "openai", "free", "gpt4", "anthropic", "", "LOCAL", "Gemini"}
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

	for _, llm := range []string{"gemini", "local"} {
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

func TestSaveSettings_NilSettings(t *testing.T) {
	db := newTestDB(t)

	if err := db.SaveSettings(nil); err == nil {
		t.Error("expected error for nil settings, got nil")
	}
}

func TestSaveAndGetSettings_OllamaModel(t *testing.T) {
	db := newTestDB(t)

	want := &Settings{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
		OllamaModel:    "llama3.1",
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

	if got.OllamaModel != want.OllamaModel {
		t.Errorf("OllamaModel = %q, want %q", got.OllamaModel, want.OllamaModel)
	}
}

func TestSaveAndGetSettings_EmptyModelFieldsDefaultToEmpty(t *testing.T) {
	db := newTestDB(t)

	want := &Settings{
		ActiveLLM:      "local",
		OllamaEndpoint: "http://localhost:11434",
	}

	if err := db.SaveSettings(want); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	got, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}

	if got.OllamaModel != "" {
		t.Errorf("OllamaModel = %q, want empty", got.OllamaModel)
	}
}
