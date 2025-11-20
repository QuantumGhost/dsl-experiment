package node

import (
	"context"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

type NoopExecutor struct{}

func (e *NoopExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	// Do nothing
	return "", nil
}
