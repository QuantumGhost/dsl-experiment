package dsl

import (
	"testing"
)

func TestNodeTypes_Constants(t *testing.T) {
	tests := []struct {
		name     string
		nodeType NodeType
		expected string
	}{
		{"Generate", NodeGenerate, "generate"},
		{"Tool", NodeTool, "tool"},
		{"Parallel", NodeParallel, "parallel"},
		{"Selector", NodeSelector, "selector"},
		{"Group", NodeGroup, "group"},
		{"Noop", NodeNoop, "noop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nodeType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.nodeType)
			}
		})
	}
}

func TestVisibility_Constants(t *testing.T) {
	tests := []struct {
		name       string
		visibility Visibility
		expected   string
	}{
		{"Public", VisibilityPublic, "public"},
		{"Silent", VisibilitySilent, "silent"},
		{"Ephemeral", VisibilityEphemeral, "ephemeral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.visibility) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.visibility)
			}
		})
	}
}

func TestMemoryPolicy_Constants(t *testing.T) {
	tests := []struct {
		name     string
		policy   MemoryPolicy
		expected string
	}{
		{"ResultOnly", MemoryResultOnly, "result_only"},
		{"Full", MemoryFull, "full"},
		{"Summary", MemorySummary, "summary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.policy) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.policy)
			}
		})
	}
}

func TestContextMode_Constants(t *testing.T) {
	tests := []struct {
		name     string
		mode     ContextMode
		expected string
	}{
		{"Full", ContextFull, "full"},
		{"None", ContextNone, "none"},
		{"Window", ContextWindow, "window"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.mode)
			}
		})
	}
}

func TestBaseNode_Methods(t *testing.T) {
	node := BaseNode{
		ID:   "test-id",
		Type: NodeGenerate,
		Next: "next-id",
	}

	if node.GetID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", node.GetID())
	}

	if node.GetType() != NodeGenerate {
		t.Errorf("Expected type NodeGenerate, got %s", node.GetType())
	}

	if node.GetNext() != "next-id" {
		t.Errorf("Expected next 'next-id', got %s", node.GetNext())
	}
}

func TestGenerateNode_Implementation(t *testing.T) {
	node := &GenerateNode{
		BaseNode: BaseNode{
			ID:   "gen1",
			Type: NodeGenerate,
		},
		Model:        "gpt-4o",
		SystemPrompt: "Test prompt",
		Temperature:  0.7,
		ContextMode:  ContextFull,
	}

	// Test interface implementation
	var _ Node = node

	if node.GetID() != "gen1" {
		t.Errorf("Expected ID 'gen1', got %s", node.GetID())
	}

	if node.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got %s", node.Model)
	}
}

func TestToolNode_Implementation(t *testing.T) {
	node := &ToolNode{
		BaseNode: BaseNode{
			ID:   "tool1",
			Type: NodeTool,
		},
	}

	var _ Node = node

	if node.GetID() != "tool1" {
		t.Errorf("Expected ID 'tool1', got %s", node.GetID())
	}
}

func TestParallelNode_Implementation(t *testing.T) {
	node := &ParallelNode{
		BaseNode: BaseNode{
			ID:   "parallel1",
			Type: NodeParallel,
		},
		Branches: []Branch{
			{Nodes: []Node{}},
		},
	}

	var _ Node = node

	if len(node.Branches) != 1 {
		t.Errorf("Expected 1 branch, got %d", len(node.Branches))
	}
}

func TestSelectorNode_Implementation(t *testing.T) {
	node := &SelectorNode{
		BaseNode: BaseNode{
			ID:   "selector1",
			Type: NodeSelector,
		},
		Model: "gpt-4o",
		Options: []Option{
			{Case: "yes", Next: "path1"},
			{Case: "no", Next: "path2"},
		},
	}

	var _ Node = node

	if len(node.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(node.Options))
	}
}

func TestGroupNode_Implementation(t *testing.T) {
	node := &GroupNode{
		BaseNode: BaseNode{
			ID:           "group1",
			Type:         NodeGroup,
			MemoryPolicy: MemoryResultOnly,
		},
		Nodes: []Node{},
	}

	var _ Node = node

	if node.MemoryPolicy != MemoryResultOnly {
		t.Errorf("Expected memory policy result_only, got %s", node.MemoryPolicy)
	}
}

func TestNoopNode_Implementation(t *testing.T) {
	node := &NoopNode{
		BaseNode: BaseNode{
			ID:   "noop1",
			Type: NodeNoop,
		},
	}

	var _ Node = node

	if node.GetType() != NodeNoop {
		t.Errorf("Expected type NodeNoop, got %s", node.GetType())
	}
}
