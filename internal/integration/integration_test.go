package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/llm"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
	"github.com/QuantumGhost/dsl-exp/internal/node"
)

// MockLLMProvider for integration tests
type MockLLMProvider struct {
	Response string
}

func (m *MockLLMProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	go func() {
		defer close(ch)
		for _, char := range m.Response {
			ch <- &llm.TextEvent{Text: string(char)}
		}
	}()
	return ch, nil
}

func TestIntegration_SimpleChatWorkflow(t *testing.T) {
	// Create test workflow file
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "chat.yaml")

	workflowContent := `name: "Simple Chat"
description: "A basic chat workflow"
nodes:
  - id: "greeting"
    type: "generate"
    model: "gpt-4o"
    system_prompt: "You are a friendly assistant"
    next: "followup"
  - id: "followup"
    type: "generate"
    model: "gpt-4o"
    system_prompt: "Continue the conversation"
`

	err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create workflow file: %v", err)
	}

	// Load workflow
	wf, err := dsl.Load(workflowFile)
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	// Create engine with mock LLM
	mockProvider := &MockLLMProvider{Response: "Hello! How can I help you?"}
	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: mockProvider})
	eng.RegisterExecutor(dsl.NodeNoop, &node.NoopExecutor{})

	// Run workflow
	ctx := context.Background()
	input := []memory.Message{
		{Role: memory.RoleUser, Content: "Hi there!"},
	}

	eventChan := eng.Run(ctx, wf, input)

	// Collect events
	nodeStarts := 0
	nodeEnds := 0
	tokens := []string{}

	for event := range eventChan {
		switch event.Type {
		case engine.EventNodeStart:
			nodeStarts++
		case engine.EventNodeEnd:
			nodeEnds++
		case engine.EventTokenGenerated:
			tokens = append(tokens, event.Payload.(string))
		case engine.EventError:
			t.Fatalf("Unexpected error: %v", event.Payload)
		}
	}

	// Verify execution
	if nodeStarts != 2 {
		t.Errorf("Expected 2 node starts, got %d", nodeStarts)
	}

	if nodeEnds != 2 {
		t.Errorf("Expected 2 node ends, got %d", nodeEnds)
	}

	fullResponse := strings.Join(tokens, "")
	// Should have two responses (one from each generate node)
	expectedTokens := len("Hello! How can I help you?") * 2
	if len(tokens) != expectedTokens {
		t.Errorf("Expected %d tokens, got %d", expectedTokens, len(tokens))
	}

	if !strings.Contains(fullResponse, "Hello") {
		t.Errorf("Expected response to contain 'Hello', got: %s", fullResponse)
	}
}

func TestIntegration_LinearWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "linear.yaml")

	workflowContent := `name: "Linear Workflow"
nodes:
  - id: "step1"
    type: "noop"
    next: "step2"
  - id: "step2"
    type: "noop"
    next: "step3"
  - id: "step3"
    type: "noop"
`

	os.WriteFile(workflowFile, []byte(workflowContent), 0644)

	wf, err := dsl.Load(workflowFile)
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeNoop, &node.NoopExecutor{})

	ctx := context.Background()
	eventChan := eng.Run(ctx, wf, nil)

	nodeOrder := []string{}
	for event := range eventChan {
		if event.Type == engine.EventNodeStart {
			nodeOrder = append(nodeOrder, event.NodeID)
		}
	}

	expectedOrder := []string{"step1", "step2", "step3"}
	if len(nodeOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d nodes, got %d", len(expectedOrder), len(nodeOrder))
	}

	for i, expected := range expectedOrder {
		if nodeOrder[i] != expected {
			t.Errorf("Expected node %s at position %d, got %s", expected, i, nodeOrder[i])
		}
	}
}

func TestIntegration_MemoryAccumulation(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "memory.yaml")

	workflowContent := `name: "Memory Test"
nodes:
  - id: "gen1"
    type: "generate"
    model: "gpt-4o"
    next: "gen2"
  - id: "gen2"
    type: "generate"
    model: "gpt-4o"
`

	os.WriteFile(workflowFile, []byte(workflowContent), 0644)

	wf, err := dsl.Load(workflowFile)
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	// Mock provider that echoes the number of messages it receives
	callCount := 0
	mockProvider := &customMockProvider{
		chatFunc: func(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
			callCount++
			response := "Response"
			ch := make(chan llm.StreamEvent)
			go func() {
				defer close(ch)
				for _, char := range response {
					ch <- &llm.TextEvent{Text: string(char)}
				}
			}()
			return ch, nil
		},
	}

	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: mockProvider})

	ctx := context.Background()
	input := []memory.Message{
		{Role: memory.RoleUser, Content: "Initial message"},
	}

	eventChan := eng.Run(ctx, wf, input)

	for range eventChan {
		// Drain events
	}

	// Both generate nodes should have been called
	if callCount != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", callCount)
	}
}

// Custom mock provider for capturing function calls
type customMockProvider struct {
	chatFunc func(context.Context, []memory.Message, llm.Options) (<-chan llm.StreamEvent, error)
}

func (c *customMockProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	return c.chatFunc(ctx, messages, opts)
}
