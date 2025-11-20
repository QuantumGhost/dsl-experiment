package node

import (
	"context"
	"fmt"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/expression"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// SelectorExecutor executes a selector node by evaluating CEL conditions
type SelectorExecutor struct{}

// Execute evaluates each case condition and routes to the first matching next node
func (e *SelectorExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	selectorNode, ok := node.(*dsl.SelectorNode)
	if !ok {
		return "", fmt.Errorf("invalid node type for SelectorExecutor")
	}

	if len(selectorNode.Cases) == 0 {
		return "", fmt.Errorf("selector node has no cases")
	}

	// Create evaluator
	eval, err := expression.NewEvaluator()
	if err != nil {
		return "", fmt.Errorf("failed to create expression evaluator: %w", err)
	}

	// Get CEL context
	celContext := mem.GetCELContext()

	// Evaluate each case in order
	for i, c := range selectorNode.Cases {
		// Evaluate the condition
		result, err := eval.EvaluateBool(c.If, celContext)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate case[%d] condition '%s': %w", i, c.If, err)
		}

		// If condition is true, route to this case's next node
		if result {
			return c.Next, nil
		}
	}

	// No case matched (should not happen if there's a default case with "true")
	return "", fmt.Errorf("no selector case matched")
}
