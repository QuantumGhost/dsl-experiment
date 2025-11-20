package engine

import (
	"context"
	"testing"
	"time"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// MockExecutor for testing
type MockExecutor struct {
	ExecuteFunc func(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error)
}

func (m *MockExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, node, mem, eventChan)
	}
	return "", nil
}

func TestEngine_BasicExecution(t *testing.T) {
	eng := NewEngine()

	executed := false
	mockExec := &MockExecutor{
		ExecuteFunc: func(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error) {
			executed = true
			return "", nil
		},
	}

	eng.RegisterExecutor(dsl.NodeNoop, mockExec)

	wf := &dsl.Workflow{
		Name: "Test",
		Nodes: dsl.NodeList{
			&dsl.NoopNode{
				BaseNode: dsl.BaseNode{
					ID:   "node1",
					Type: dsl.NodeNoop,
				},
			},
		},
	}

	ctx := context.Background()
	eventChan := eng.Run(ctx, wf, nil)

	events := []Event{}
	for event := range eventChan {
		events = append(events, event)
	}

	if !executed {
		t.Error("Node was not executed")
	}

	// Should have NodeStart and NodeEnd events
	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}

	if events[0].Type != EventNodeStart {
		t.Errorf("Expected first event to be NodeStart, got %s", events[0].Type)
	}
}

func TestEngine_SequentialExecution(t *testing.T) {
	eng := NewEngine()

	executionOrder := []string{}
	mockExec := &MockExecutor{
		ExecuteFunc: func(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error) {
			executionOrder = append(executionOrder, node.GetID())
			return "", nil
		},
	}

	eng.RegisterExecutor(dsl.NodeNoop, mockExec)

	wf := &dsl.Workflow{
		Name: "Test",
		Nodes: dsl.NodeList{
			&dsl.NoopNode{
				BaseNode: dsl.BaseNode{
					ID:   "node1",
					Type: dsl.NodeNoop,
					Next: "node2",
				},
			},
			&dsl.NoopNode{
				BaseNode: dsl.BaseNode{
					ID:   "node2",
					Type: dsl.NodeNoop,
				},
			},
		},
	}

	ctx := context.Background()
	eventChan := eng.Run(ctx, wf, nil)

	for range eventChan {
		// Drain events
	}

	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executionOrder))
	}

	if executionOrder[0] != "node1" || executionOrder[1] != "node2" {
		t.Errorf("Execution order incorrect: %v", executionOrder)
	}
}

func TestEngine_ContextCancellation(t *testing.T) {
	eng := NewEngine()

	mockExec := &MockExecutor{
		ExecuteFunc: func(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error) {
			// Simulate long-running operation
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return "", nil
			}
		},
	}

	eng.RegisterExecutor(dsl.NodeNoop, mockExec)

	wf := &dsl.Workflow{
		Name: "Test",
		Nodes: dsl.NodeList{
			&dsl.NoopNode{
				BaseNode: dsl.BaseNode{
					ID:   "node1",
					Type: dsl.NodeNoop,
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	eventChan := eng.Run(ctx, wf, nil)

	// Cancel immediately
	cancel()

	hasError := false
	for event := range eventChan {
		if event.Type == EventError {
			hasError = true
		}
	}

	if !hasError {
		t.Error("Expected error event after context cancellation")
	}
}

func TestEngine_MissingExecutor(t *testing.T) {
	eng := NewEngine()

	wf := &dsl.Workflow{
		Name: "Test",
		Nodes: dsl.NodeList{
			&dsl.GenerateNode{
				BaseNode: dsl.BaseNode{
					ID:   "node1",
					Type: dsl.NodeGenerate,
				},
			},
		},
	}

	ctx := context.Background()
	eventChan := eng.Run(ctx, wf, nil)

	hasError := false
	for event := range eventChan {
		if event.Type == EventError {
			hasError = true
		}
	}

	if !hasError {
		t.Error("Expected error event for missing executor")
	}
}
