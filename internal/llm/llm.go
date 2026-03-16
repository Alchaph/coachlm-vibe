// Package llm defines the LLM router interface and message types.
// All LLM backends (Claude, OpenAI, local) implement the LLM interface.
package llm

import "context"

// Role constants for Message.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message represents a single chat message with a role and content.
type Message struct {
	Role    string
	Content string
}

// LLM is the core interface that all LLM backends must implement.
// Defined in AGENTS.md lines 110-113. Do not change this signature.
type LLM interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	Name() string
}
