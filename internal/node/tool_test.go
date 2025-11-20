package node

import (
	"context"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

func TestToolExecutor_Execute(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register("weather", func(ctx context.Context, args map[string]interface{}) (string, error) {
		city, ok := args["city"].(string)
		if !ok {
			return "", nil
		}
		return "Sunny in " + city, nil
	})

	executor := &ToolExecutor{Registry: registry}

	node := &dsl.ToolNode{
		BaseNode: dsl.BaseNode{
			ID:   "get_weather",
			Type: dsl.NodeTool,
		},
		ToolName: "weather",
		Parameters: map[string]interface{}{
			"city": "San Francisco",
		},
	}

	mem := memory.NewGlobal(nil)
	eventChan := make(chan engine.Event, 10)
	ctx := context.Background()

	_, err := executor.Execute(ctx, node, mem, eventChan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	close(eventChan)

	// Verify memory
	snapshot := mem.Snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("Expected 1 message in memory, got %d", len(snapshot))
	}
	if snapshot[0].Role != memory.RoleUser {
		t.Errorf("Expected user role, got %s", snapshot[0].Role)
	}
	expectedContent := "Tool 'weather' output: Sunny in San Francisco"
	if snapshot[0].Content != expectedContent {
		t.Errorf("Expected '%s', got '%s'", expectedContent, snapshot[0].Content)
	}
}
