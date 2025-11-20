package dsl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidWorkflow(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test Workflow"
description: "A test workflow"
version: "1.0"
nodes:
  - id: "node1"
    type: "generate"
    model: "gpt-4o"
    system_prompt: "You are helpful"
    next: "node2"
  - id: "node2"
    type: "noop"
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	wf, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if wf.Name != "Test Workflow" {
		t.Errorf("Expected name 'Test Workflow', got %s", wf.Name)
	}

	if len(wf.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(wf.Nodes))
	}

	// Check first node
	node1 := wf.Nodes[0]
	if node1.GetID() != "node1" {
		t.Errorf("Expected node1 ID, got %s", node1.GetID())
	}
	if node1.GetType() != NodeGenerate {
		t.Errorf("Expected generate type, got %s", node1.GetType())
	}

	genNode, ok := node1.(*GenerateNode)
	if !ok {
		t.Fatalf("Failed to cast to GenerateNode")
	}
	if genNode.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", genNode.Model)
	}
}

func TestLoad_AllNodeTypes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "All Types"
nodes:
  - id: "gen1"
    type: "generate"
    model: "gpt-4o"
  - id: "tool1"
    type: "tool"
  - id: "noop1"
    type: "noop"
  - id: "selector1"
    type: "selector"
    model: "gpt-4o"
    options:
      - case: "yes"
        next: "noop1"
  - id: "group1"
    type: "group"
    nodes:
      - id: "inner1"
        type: "noop"
  - id: "parallel1"
    type: "parallel"
    branches:
      - nodes:
          - id: "branch1"
            type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	wf, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(wf.Nodes) != 6 {
		t.Errorf("Expected 6 nodes, got %d", len(wf.Nodes))
	}

	// Verify each type
	types := []NodeType{}
	for _, node := range wf.Nodes {
		types = append(types, node.GetType())
	}

	expectedTypes := []NodeType{NodeGenerate, NodeTool, NodeNoop, NodeSelector, NodeGroup, NodeParallel}
	for i, expected := range expectedTypes {
		if types[i] != expected {
			t.Errorf("Expected node %d to be %s, got %s", i, expected, types[i])
		}
	}
}

func TestLoad_ParallelNodeWithReducer(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Parallel with Reducer"
nodes:
  - id: "parallel1"
    type: "parallel"
    branches:
      - nodes:
          - id: "branch1"
            type: "noop"
      - nodes:
          - id: "branch2"
            type: "noop"
    reducer:
      id: "reducer1"
      type: "generate"
      model: "gpt-4o"
`

	os.WriteFile(testFile, []byte(content), 0644)

	wf, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	parallel, ok := wf.Nodes[0].(*ParallelNode)
	if !ok {
		t.Fatalf("Failed to cast to ParallelNode")
	}

	if parallel.Reducer == nil {
		t.Error("Expected reducer to be set")
	}

	if parallel.Reducer.GetID() != "reducer1" {
		t.Errorf("Expected reducer ID 'reducer1', got %s", parallel.Reducer.GetID())
	}
}

func TestLoad_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `nodes:
  - id: "node1"
    type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for missing name, got nil")
	}
}

func TestLoad_EmptyNodes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes: []
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for empty nodes, got nil")
	}
}

func TestLoad_DuplicateIDs(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "node1"
    type: "noop"
  - id: "node1"
    type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for duplicate IDs, got nil")
	}
}

func TestLoad_InvalidNextPointer(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "node1"
    type: "noop"
    next: "nonexistent"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for invalid next pointer, got nil")
	}
}

func TestLoad_MissingRequiredField_Generate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "node1"
    type: "generate"
    # Missing required 'model' field
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for missing model field, got nil")
	}
}

func TestLoad_SelectorValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "selector1"
    type: "selector"
    model: "gpt-4o"
    options:
      - case: "option1"
        next: "node1"
      - case: "option2"
        next: "node2"
  - id: "node1"
    type: "noop"
  - id: "node2"
    type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	wf, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	selector, ok := wf.Nodes[0].(*SelectorNode)
	if !ok {
		t.Fatal("Failed to cast to SelectorNode")
	}

	if len(selector.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(selector.Options))
	}
}

func TestLoad_SelectorEmptyOptions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "selector1"
    type: "selector"
    model: "gpt-4o"
    options: []
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for empty options, got nil")
	}
}

func TestLoad_SelectorInvalidNext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "selector1"
    type: "selector"
    model: "gpt-4o"
    options:
      - case: "option1"
        next: "invalid"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for invalid option next, got nil")
	}
}

func TestLoad_ParallelEmptyBranches(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "parallel1"
    type: "parallel"
    branches: []
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for empty branches, got nil")
	}
}

func TestLoad_GroupEmptyNodes(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "group1"
    type: "group"
    nodes: []
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for empty group nodes, got nil")
	}
}

func TestLoad_InvalidIDPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "123-invalid"
    type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for invalid ID pattern, got nil")
	}
}

func TestLoad_EmptyID(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: ""
    type: "noop"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
}

func TestLoad_NestedParallelValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "parallel1"
    type: "parallel"
    branches:
      - nodes:
          - id: "inner1"
            type: "noop"
            next: "invalid"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for invalid next in nested node, got nil")
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	_, err := Load("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `invalid: yaml: syntax:
  - [broken
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoad_UnknownNodeType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")

	content := `name: "Test"
nodes:
  - id: "node1"
    type: "unknown_type"
`

	os.WriteFile(testFile, []byte(content), 0644)

	_, err := Load(testFile)
	if err == nil {
		t.Error("Expected error for unknown node type, got nil")
	}
}
