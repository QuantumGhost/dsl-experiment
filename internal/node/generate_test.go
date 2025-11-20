package node

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/llm"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// MockLLMProvider for testing
type MockLLMProvider struct {
	Response string
}

func (m *MockLLMProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	go func() {
		defer close(ch)
		// Stream the response character by character
		for _, char := range m.Response {
			ch <- &llm.TextEvent{Text: string(char)}
		}
	}()
	return ch, nil
}

func TestGenerateExecutor_BasicExecution(t *testing.T) {
	mockProvider := &MockLLMProvider{Response: "Hello, world!"}
	executor := &GenerateExecutor{Provider: mockProvider}

	node := &dsl.GenerateNode{
		BaseNode: dsl.BaseNode{
			ID:   "test",
			Type: dsl.NodeGenerate,
		},
		Model:        "gpt-4o",
		SystemPrompt: "You are helpful",
	}

	mem := memory.NewGlobal([]memory.Message{
		{Role: memory.RoleUser, Content: "Hi"},
	})

	eventChan := make(chan engine.Event, 100)
	ctx := context.Background()

	nextID, err := executor.Execute(ctx, node, mem, eventChan)
	close(eventChan)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if nextID != "" {
		t.Errorf("Expected empty nextID, got %s", nextID)
	}

	// Collect token events
	var tokens []string
	for event := range eventChan {
		if event.Type == engine.EventTokenGenerated {
			tokens = append(tokens, event.Payload.(string))
		}
	}

	fullResponse := strings.Join(tokens, "")
	if fullResponse != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!', got %s", fullResponse)
	}

	// Check memory was updated
	snapshot := mem.Snapshot()
	// Should have: Original user message + assistant response + system message = 3 messages
	if len(snapshot) != 2 {
		t.Errorf("Expected 2 messages in memory, got %d", len(snapshot))
	}

	// First message should be user message
	if snapshot[0].Role != memory.RoleUser {
		t.Errorf("Expected first message to be user, got %s", snapshot[0].Role)
	}

	// Last message should be assistant
	lastMsg := snapshot[len(snapshot)-1]
	if lastMsg.Role != memory.RoleAssistant {
		t.Errorf("Expected last message to be assistant, got %s", lastMsg.Role)
	}
	if lastMsg.Content != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!' in memory, got %s", lastMsg.Content)
	}
}

func TestGenerateExecutor_WithSystemPrompt(t *testing.T) {
	var receivedMessages []memory.Message

	mockProvider := &customMockProvider{
		chatFunc: func(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
			receivedMessages = messages
			ch := make(chan llm.StreamEvent)
			go func() {
				defer close(ch)
				ch <- &llm.TextEvent{Text: "OK"}
			}()
			return ch, nil
		},
	}

	executor := &GenerateExecutor{Provider: mockProvider}

	node := &dsl.GenerateNode{
		BaseNode: dsl.BaseNode{
			ID:   "test",
			Type: dsl.NodeGenerate,
		},
		Model:        "gpt-4o",
		SystemPrompt: "You are a helpful assistant",
	}

	mem := memory.NewGlobal([]memory.Message{
		{Role: memory.RoleUser, Content: "Hello"},
	})

	eventChan := make(chan engine.Event, 100)
	ctx := context.Background()

	_, err := executor.Execute(ctx, node, mem, eventChan)
	close(eventChan)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Messages sent to LLM should be:
	// 1. System message (from node config)
	// 2. User message (from history)
	if len(receivedMessages) != 2 {
		t.Fatalf("Expected 2 messages sent to LLM, got %d", len(receivedMessages))
	}

	// First should be system
	if receivedMessages[0].Role != memory.RoleSystem {
		t.Errorf("Expected first message to be system, got %s", receivedMessages[0].Role)
	}

	if receivedMessages[0].Content != "You are a helpful assistant" {
		t.Errorf("System prompt mismatch: %s", receivedMessages[0].Content)
	}

	// Second should be user (from history)
	if receivedMessages[1].Role != memory.RoleUser {
		t.Errorf("Expected second message to be user, got %s", receivedMessages[1].Role)
	}

	if receivedMessages[1].Content != "Hello" {
		t.Errorf("User message mismatch: %s", receivedMessages[1].Content)
	}

	// Memory should NOT contain system message, only user + assistant
	snapshot := mem.Snapshot()
	if len(snapshot) != 2 {
		t.Errorf("Expected 2 messages in memory (user + assistant), got %d", len(snapshot))
	}

	// Verify no system messages in memory
	for i, msg := range snapshot {
		if msg.Role == memory.RoleSystem {
			t.Errorf("System message should not be in memory, found at position %d", i)
		}
	}
}

func TestGenerateExecutor_ContextModes(t *testing.T) {
	tests := []struct {
		name          string
		contextMode   dsl.ContextMode
		expectHistory bool
	}{
		{"Full context", dsl.ContextFull, true},
		{"No context", dsl.ContextNone, false},
		{"Window context", dsl.ContextWindow, true}, // Currently treated as Full
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedMessages []memory.Message

			mockProvider := &customMockProvider{
				chatFunc: func(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
					receivedMessages = messages
					ch := make(chan llm.StreamEvent)
					go func() {
						defer close(ch)
						ch <- &llm.TextEvent{Text: "Response"}
					}()
					return ch, nil
				},
			}

			executor := &GenerateExecutor{Provider: mockProvider}

			node := &dsl.GenerateNode{
				BaseNode: dsl.BaseNode{
					ID:   "test",
					Type: dsl.NodeGenerate,
				},
				Model:        "gpt-4o",
				SystemPrompt: "System",
				ContextMode:  tt.contextMode,
			}

			mem := memory.NewGlobal([]memory.Message{
				{Role: memory.RoleUser, Content: "Message 1"},
				{Role: memory.RoleAssistant, Content: "Response 1"},
			})

			eventChan := make(chan engine.Event, 100)
			ctx := context.Background()

			_, err := executor.Execute(ctx, node, mem, eventChan)
			close(eventChan)

			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// First message should always be system
			if receivedMessages[0].Role != memory.RoleSystem {
				t.Errorf("Expected first message to be system")
			}

			if tt.expectHistory {
				// Should have system + history
				if len(receivedMessages) != 3 { // System + 2 history messages
					t.Errorf("Expected 3 messages (system + history), got %d", len(receivedMessages))
				}
			} else {
				// Should only have system, no history
				if len(receivedMessages) != 1 {
					t.Errorf("Expected only 1 message (system), got %d", len(receivedMessages))
				}
			}
		})
	}
}

// Custom mock provider that allows capturing function calls
type customMockProvider struct {
	chatFunc func(context.Context, []memory.Message, llm.Options) (<-chan llm.StreamEvent, error)
}

func (c *customMockProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	return c.chatFunc(ctx, messages, opts)
}
