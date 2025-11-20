package llm

import (
	"context"

	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// ToolDefinition defines a tool that can be called by the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// ToolCall represents a request from the LLM to execute a tool.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type Options struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Tools       []ToolDefinition
}

// StreamEvent represents a streaming response event (sealed interface)
//
// go-sumtype:decl StreamEvent
type StreamEvent interface {
	isStreamEvent()
}

// TextEvent represents a text chunk in the stream
type TextEvent struct {
	Text string
}

func (*TextEvent) isStreamEvent() {}

// ToolCallEvent represents a tool call in the stream
type ToolCallEvent struct {
	ToolCall ToolCall
}

func (*ToolCallEvent) isStreamEvent() {}

// Provider defines the interface for LLM providers
type Provider interface {
	// ChatCompletion performs a chat completion request with streaming
	ChatCompletion(ctx context.Context, messages []memory.Message, opts Options) (<-chan StreamEvent, error)
}
