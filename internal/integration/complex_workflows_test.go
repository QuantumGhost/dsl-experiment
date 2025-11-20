package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/llm"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
	"github.com/QuantumGhost/dsl-exp/internal/node"
)

// MockLLMProvider for complex workflow tests
type ComplexMockProvider struct {
	ResponseMap map[string]string // nodeID -> response
	CallOrder   []string          // Track execution order
}

func (m *ComplexMockProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	go func() {
		defer close(ch)

		// Determine which node by checking system prompt or use default
		response := "Default response"
		for nodeID, resp := range m.ResponseMap {
			if strings.Contains(opts.Model, nodeID) {
				response = resp
				break
			}
		}

		// Track call order based on system prompt
		if len(messages) > 0 && messages[0].Role == memory.RoleSystem {
			m.CallOrder = append(m.CallOrder, messages[0].Content)
		}

		for _, char := range response {
			ch <- &llm.TextEvent{Text: string(char)}
		}
	}()
	return ch, nil
}

func TestComplexWorkflow_MultiStep(t *testing.T) {
	wf, err := dsl.Load("../../examples/multi_step_conversation.yaml")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	mockProvider := &ComplexMockProvider{
		ResponseMap: make(map[string]string),
		CallOrder:   []string{},
	}

	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: mockProvider})

	ctx := context.Background()
	input := []memory.Message{
		{Role: memory.RoleUser, Content: "Hello, I need help with my project"},
	}

	eventChan := eng.Run(ctx, wf, input)

	nodeExecutions := []string{}
	for event := range eventChan {
		if event.Type == engine.EventNodeStart {
			nodeExecutions = append(nodeExecutions, event.NodeID)
		}
		if event.Type == engine.EventError {
			t.Fatalf("Unexpected error: %v", event.Payload)
		}
	}

	// Verify correct execution order
	expectedOrder := []string{"greeting", "understand_intent", "provide_response", "followup"}
	if len(nodeExecutions) != len(expectedOrder) {
		t.Errorf("Expected %d nodes, got %d", len(expectedOrder), len(nodeExecutions))
	}

	for i, expected := range expectedOrder {
		if i >= len(nodeExecutions) || nodeExecutions[i] != expected {
			t.Errorf("Expected node %d to be %s, got %s", i, expected, nodeExecutions[i])
		}
	}
}

func TestComplexWorkflow_ConditionalRouting(t *testing.T) {
	wf, err := dsl.Load("../../examples/conditional_routing.yaml")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	// Test each routing path
	testCases := []struct {
		name           string
		selectorChoice string
		expectedNode   string
	}{
		{"Question route", "question", "answer_question"},
		{"Task route", "task", "handle_task"},
		{"Chat route", "chat", "casual_chat"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock that returns specific selector choice
			mockProvider := &SelectorMockProvider{
				SelectorResponse: tc.selectorChoice,
				GenerateResponse: "Test response",
			}

			eng := engine.NewEngine()
			eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: mockProvider})
			eng.RegisterExecutor(dsl.NodeSelector, &MockSelectorExecutor{
				Choice: tc.selectorChoice,
			})

			ctx := context.Background()
			input := []memory.Message{
				{Role: memory.RoleUser, Content: "Test input"},
			}

			eventChan := eng.Run(ctx, wf, input)

			nodeExecutions := []string{}
			for event := range eventChan {
				if event.Type == engine.EventNodeStart {
					nodeExecutions = append(nodeExecutions, event.NodeID)
				}
			}

			// Should execute: classify_intent -> expected_node -> closing
			found := false
			for _, nodeID := range nodeExecutions {
				if nodeID == tc.expectedNode {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to execute %s, but it was not found in %v", tc.expectedNode, nodeExecutions)
			}
		})
	}
}

func TestComplexWorkflow_ParallelExecution(t *testing.T) {
	wf, err := dsl.Load("../../examples/parallel_research.yaml")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	mockProvider := &ComplexMockProvider{
		ResponseMap: make(map[string]string),
		CallOrder:   []string{},
	}

	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: mockProvider})
	eng.RegisterExecutor(dsl.NodeParallel, &MockParallelExecutor{})

	ctx := context.Background()
	input := []memory.Message{
		{Role: memory.RoleUser, Content: "Research AI technology"},
	}

	eventChan := eng.Run(ctx, wf, input)

	parallelNodesSeen := make(map[string]bool)
	for event := range eventChan {
		if event.Type == engine.EventNodeStart {
			if event.NodeID == "research_technical" ||
				event.NodeID == "research_business" ||
				event.NodeID == "research_social" {
				parallelNodesSeen[event.NodeID] = true
			}
		}
	}

	// All three parallel branches should execute
	expectedParallel := []string{"research_technical", "research_business", "research_social"}
	for _, nodeID := range expectedParallel {
		if !parallelNodesSeen[nodeID] {
			t.Errorf("Expected parallel node %s to execute", nodeID)
		}
	}
}

func TestComplexWorkflow_NestedStructure(t *testing.T) {
	wf, err := dsl.Load("../../examples/nested_workflow.yaml")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	// Verify structure
	if len(wf.Nodes) != 3 {
		t.Errorf("Expected 3 top-level nodes, got %d", len(wf.Nodes))
	}

	// Find group node
	var groupNode *dsl.GroupNode
	for _, node := range wf.Nodes {
		if g, ok := node.(*dsl.GroupNode); ok {
			groupNode = g
			break
		}
	}

	if groupNode == nil {
		t.Fatal("Expected to find a GroupNode")
	}

	// Verify group has 2 nodes
	if len(groupNode.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in group, got %d", len(groupNode.Nodes))
	}

	// Find parallel node within group
	var parallelNode *dsl.ParallelNode
	for _, node := range groupNode.Nodes {
		if p, ok := node.(*dsl.ParallelNode); ok {
			parallelNode = p
			break
		}
	}

	if parallelNode == nil {
		t.Fatal("Expected to find a ParallelNode in group")
	}

	// Verify parallel has 2 branches
	if len(parallelNode.Branches) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(parallelNode.Branches))
	}

	// Verify reducer exists
	if parallelNode.Reducer == nil {
		t.Error("Expected reducer to be set")
	}
}

func TestComplexWorkflow_LoadAllExamples(t *testing.T) {
	examples := []string{
		"../../examples/simple_chat.yaml",
		"../../examples/multi_step_conversation.yaml",
		"../../examples/conditional_routing.yaml",
		"../../examples/parallel_research.yaml",
		"../../examples/nested_workflow.yaml",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			wf, err := dsl.Load(example)
			if err != nil {
				t.Fatalf("Failed to load %s: %v", example, err)
			}

			if wf.Name == "" {
				t.Errorf("Workflow name is empty for %s", example)
			}

			if len(wf.Nodes) == 0 {
				t.Errorf("No nodes found in %s", example)
			}
		})
	}
}

// Helper mock providers and executors

type SelectorMockProvider struct {
	SelectorResponse string
	GenerateResponse string
}

func (m *SelectorMockProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	go func() {
		defer close(ch)
		response := m.GenerateResponse
		if strings.Contains(strings.ToLower(opts.Model), "classificat") {
			response = m.SelectorResponse
		}
		for _, char := range response {
			ch <- &llm.TextEvent{Text: string(char)}
		}
	}()
	return ch, nil
}

type MockSelectorExecutor struct {
	Choice string
}

func (m *MockSelectorExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	selectorNode, ok := node.(*dsl.SelectorNode)
	if !ok {
		return "", nil
	}

	// For mock, just return first case's next (in real implementation, would evaluate CEL)
	if len(selectorNode.Cases) > 0 {
		return selectorNode.Cases[0].Next, nil
	}

	return "", nil
}

type MockParallelExecutor struct{}

func (m *MockParallelExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	parallelNode, ok := node.(*dsl.ParallelNode)
	if !ok {
		return "", nil
	}

	// Simply emit start events for each branch node for testing
	for _, branch := range parallelNode.Branches {
		for _, branchNode := range branch.Nodes {
			eventChan <- engine.Event{
				Type:   engine.EventNodeStart,
				NodeID: branchNode.GetID(),
			}
			eventChan <- engine.Event{
				Type:   engine.EventNodeEnd,
				NodeID: branchNode.GetID(),
			}
		}
	}

	return "", nil
}
