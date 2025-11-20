package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/llm"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

type GenerateExecutor struct {
	Provider llm.Provider
}

func (e *GenerateExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	genNode, ok := node.(*dsl.GenerateNode)
	if !ok {
		return "", fmt.Errorf("invalid node type for GenerateExecutor")
	}

	// Prepare Messages for LLM
	// The message list should be structured as:
	// 1. System message (if provided in node config)
	// 2. Conversation history (User/Assistant/Tool messages from GlobalMemory)

	messages := []memory.Message{}

	// 1. Add System Prompt from node configuration
	// This is the instruction for how the LLM should behave in this specific node
	if genNode.SystemPrompt != "" {
		messages = append(messages, memory.Message{
			Role:    memory.RoleSystem,
			Content: genNode.SystemPrompt,
		})
	}

	// 2. Add Conversation History from GlobalMemory
	// GlobalMemory should only contain User/Assistant/Tool messages, NOT system messages
	// System messages are ephemeral and specific to each generate node
	history := mem.Snapshot()

	// TODO: Handle Context Mode (Window, None, Full)
	// - Full: Include all history (default)
	// - Window: Include only last N messages
	// - None: No history, only system prompt
	switch genNode.ContextMode {
	case dsl.ContextNone:
		// Don't add any history, only system prompt
	case dsl.ContextWindow:
		// TODO: Implement window logic (e.g., last 10 messages)
		// For now, treat as Full
		messages = append(messages, history...)
	case dsl.ContextFull:
		fallthrough
	default:
		// Include all conversation history
		messages = append(messages, history...)
	}

	// Call LLM
	opts := llm.Options{
		Model:       genNode.Model,
		Temperature: genNode.Temperature,
	}

	stream, err := e.Provider.ChatCompletion(ctx, messages, opts)
	if err != nil {
		return "", err
	}

	// Stream Response
	var fullResponse strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *llm.TextEvent:
			eventChan <- engine.Event{
				Type:    engine.EventTokenGenerated,
				NodeID:  node.GetID(),
				Payload: e.Text,
			}
			fullResponse.WriteString(e.Text)
		case *llm.ToolCallEvent:
			// TODO: Handle Tool Calls
		}
	}

	// Append Assistant's response to GlobalMemory
	// Only the actual conversation (User/Assistant/Tool) goes into memory
	// System prompts are NOT stored in memory
	mem.Append(memory.Message{
		Role:    memory.RoleAssistant,
		Content: fullResponse.String(),
	})

	return "", nil
}
