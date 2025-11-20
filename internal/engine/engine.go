package engine

import (
	"context"
	"fmt"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

type Engine struct {
	executors map[dsl.NodeType]Executor
}

func NewEngine() *Engine {
	return &Engine{
		executors: make(map[dsl.NodeType]Executor),
	}
}

func (e *Engine) RegisterExecutor(t dsl.NodeType, exec Executor) {
	e.executors[t] = exec
}

func (e *Engine) Run(ctx context.Context, wf *dsl.Workflow, input []memory.Message) <-chan Event {
	ch := make(chan Event)
	go func() {
		defer close(ch)

		// Initialize Global Memory with Input
		mem := memory.NewGlobal(input)

		// Prepare inputs context for CEL
		// Extract user input from messages (typically the first user message)
		inputs := make(map[string]interface{})
		for _, msg := range input {
			if msg.Role == memory.RoleUser {
				// Store the first user message content as "query"
				if _, exists := inputs["query"]; !exists {
					inputs["query"] = msg.Content
				}
			}
		}
		// Store inputs in Memory for access by executors
		mem.SetInputs(inputs)

		if len(wf.Nodes) == 0 {
			return
		}

		// Start from first node
		// In a real graph, we might need a lookup map. For now, assuming linear or ID lookup.
		// We need a helper to find node by ID.
		nodeMap := make(map[string]dsl.Node)
		for _, n := range wf.Nodes {
			nodeMap[n.GetID()] = n
		}

		currentNodeID := wf.Nodes[0].GetID()

		// Safety: prevent infinite loops
		const maxIterations = 100
		iterationCount := 0

		for currentNodeID != "" {
			iterationCount++
			if iterationCount > maxIterations {
				ch <- Event{
					Type:    EventError,
					Payload: fmt.Errorf("maximum iteration limit (%d) exceeded - possible infinite loop", maxIterations),
				}
				return
			}

			node, exists := nodeMap[currentNodeID]
			if !exists {
				ch <- Event{Type: EventError, Payload: fmt.Errorf("node not found: %s", currentNodeID)}
				return
			}

			// Emit NodeStart
			ch <- Event{Type: EventNodeStart, NodeID: node.GetID()}

			// Get executor for this node type
			executor, exists := e.executors[node.GetType()]
			if !exists {
				ch <- Event{Type: EventError, Payload: fmt.Errorf("no executor for node type: %s", node.GetType())}
				return
			}

			// Execute node
			nextID, err := executor.Execute(ctx, node, mem, ch)
			if err != nil {
				ch <- Event{Type: EventError, Payload: err}
				return
			}

			// Emit NodeEnd
			ch <- Event{Type: EventNodeEnd, NodeID: node.GetID(), Payload: nextID}

			// Determine next node
			if nextID != "" {
				currentNodeID = nextID
			} else {
				currentNodeID = node.GetNext()
			}
		}
	}()
	return ch
}
