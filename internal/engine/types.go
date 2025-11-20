package engine

import (
	"context"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

type EventType string

const (
	EventNodeStart      EventType = "node_start"
	EventNodeEnd        EventType = "node_end"
	EventTokenGenerated EventType = "token_generated"
	EventToolCall       EventType = "tool_call"
	EventError          EventType = "error"
)

type Event struct {
	Type    EventType
	NodeID  string
	Payload interface{} // Can be string (token), error, or struct
}

// Executor is the interface for executing a node.
type Executor interface {
	// Execute runs the node logic.
	// It streams events to eventChan.
	// Returns the ID of the next node to execute, or empty string for default next.
	Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- Event) (string, error)
}
