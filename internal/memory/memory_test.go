package memory

import (
	"fmt"
	"sync"
	"testing"
)

func TestGlobalMemory_Concurrency(t *testing.T) {
	mem := NewGlobal(nil)
	wg := sync.WaitGroup{}
	numGoroutines := 100
	numMessages := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				mem.Append(Message{
					Role:    RoleUser,
					Content: fmt.Sprintf("goroutine-%d-msg-%d", id, j),
				})
			}
		}(i)
	}

	wg.Wait()

	expected := numGoroutines * numMessages
	if mem.Len() != expected {
		t.Errorf("expected %d messages, got %d", expected, mem.Len())
	}
}

func TestGlobalMemory_Snapshot(t *testing.T) {
	mem := NewGlobal(nil)
	mem.Append(Message{Role: RoleUser, Content: "Hello"})

	snapshot := mem.Snapshot()

	if len(snapshot) != 1 {
		t.Errorf("Expected snapshot length 1, got %d", len(snapshot))
	}

	// Modifying snapshot should not affect original
	snapshot = append(snapshot, Message{Role: RoleAssistant, Content: "Hi"})
	mem.Append(Message{Role: RoleSystem, Content: "Test"})

	snapshot2 := mem.Snapshot()
	if len(snapshot2) != 2 {
		t.Errorf("Expected snapshot2 length 2, got %d", len(snapshot2))
	}

	if len(snapshot) != 2 {
		t.Errorf("Snapshot should not be affected by new appends")
	}
}

func TestMemory_SetInputs(t *testing.T) {
	mem := NewMemory()

	inputs := map[string]interface{}{
		"query":   "Hello world",
		"user_id": 123,
	}

	mem.SetInputs(inputs)
	retrieved := mem.GetInputs()

	if retrieved["query"] != "Hello world" {
		t.Errorf("Expected query 'Hello world', got %v", retrieved["query"])
	}

	if retrieved["user_id"] != 123 {
		t.Errorf("Expected user_id 123, got %v", retrieved["user_id"])
	}
}

func TestMemory_NodeOutputs(t *testing.T) {
	mem := NewMemory()

	output := map[string]interface{}{
		"output": "Test output",
		"data": map[string]interface{}{
			"temperature": 25,
		},
	}

	mem.SetNodeOutput("node1", output)

	retrieved, ok := mem.GetNodeOutput("node1")
	if !ok {
		t.Fatal("Expected to find node1 output")
	}

	outputMap, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Fatal("Expected output to be a map")
	}

	if outputMap["output"] != "Test output" {
		t.Errorf("Expected 'Test output', got %v", outputMap["output"])
	}

	_, ok = mem.GetNodeOutput("nonexistent")
	if ok {
		t.Error("Should not find nonexistent node")
	}
}

func TestMemory_Variables(t *testing.T) {
	mem := NewMemory()

	mem.SetVariable("counter", 5)
	mem.SetVariable("name", "Alice")

	counter, ok := mem.GetVariable("counter")
	if !ok || counter != 5 {
		t.Errorf("Expected counter 5, got %v", counter)
	}

	name, ok := mem.GetVariable("name")
	if !ok || name != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", name)
	}

	_, ok = mem.GetVariable("nonexistent")
	if ok {
		t.Error("Should not find nonexistent variable")
	}
}

func TestMemory_GetCELContext(t *testing.T) {
	mem := NewMemory()

	// Set inputs
	mem.SetInputs(map[string]interface{}{
		"query": "test query",
	})

	// Set node output
	mem.SetNodeOutput("node1", map[string]interface{}{
		"output": "node output",
	})

	// Set variable
	mem.SetVariable("counter", 10)

	// Get CEL context
	ctx := mem.GetCELContext()

	// Verify inputs
	inputs, ok := ctx["inputs"].(map[string]interface{})
	if !ok {
		t.Fatal("inputs should be a map")
	}
	if inputs["query"] != "test query" {
		t.Errorf("Expected query 'test query', got %v", inputs["query"])
	}

	// Verify nodes
	nodes, ok := ctx["nodes"].(map[string]interface{})
	if !ok {
		t.Fatal("nodes should be a map")
	}
	node1, ok := nodes["node1"].(map[string]interface{})
	if !ok || node1["output"] != "node output" {
		t.Errorf("Expected node1 output, got %v", nodes["node1"])
	}

	// Verify vars
	vars, ok := ctx["vars"].(map[string]interface{})
	if !ok {
		t.Fatal("vars should be a map")
	}
	if vars["counter"] != 10 {
		t.Errorf("Expected counter 10, got %v", vars["counter"])
	}
}
