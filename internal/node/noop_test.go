package node

import (
	"context"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

func TestNoopExecutor_Execute(t *testing.T) {
	executor := &NoopExecutor{}

	node := &dsl.NoopNode{
		BaseNode: dsl.BaseNode{
			ID:   "test",
			Type: dsl.NodeNoop,
		},
	}

	mem := memory.NewGlobal(nil)
	eventChan := make(chan engine.Event, 10)
	ctx := context.Background()

	nextID, err := executor.Execute(ctx, node, mem, eventChan)
	close(eventChan)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if nextID != "" {
		t.Errorf("Expected empty nextID, got %s", nextID)
	}

	// Should produce no events
	eventCount := 0
	for range eventChan {
		eventCount++
	}

	if eventCount != 0 {
		t.Errorf("Expected 0 events, got %d", eventCount)
	}

	// Memory should be unchanged
	if mem.Len() != 0 {
		t.Errorf("Expected memory to be empty, got %d messages", mem.Len())
	}
}
