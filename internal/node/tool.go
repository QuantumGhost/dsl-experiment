package node

import (
	"context"
	"fmt"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/expression"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// ToolHandler defines the function signature for a tool implementation.
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ToolRegistry holds the available tools.
type ToolRegistry struct {
	tools map[string]ToolHandler
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolHandler),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(name string, handler ToolHandler) {
	r.tools[name] = handler
}

// Get retrieves a tool handler by name.
func (r *ToolRegistry) Get(name string) (ToolHandler, bool) {
	h, ok := r.tools[name]
	return h, ok
}

// ToolExecutor executes a tool node.
type ToolExecutor struct {
	Registry *ToolRegistry
}

func (e *ToolExecutor) Execute(ctx context.Context, node dsl.Node, mem *memory.Memory, eventChan chan<- engine.Event) (string, error) {
	toolNode, ok := node.(*dsl.ToolNode)
	if !ok {
		return "", fmt.Errorf("invalid node type for ToolExecutor")
	}

	handler, ok := e.Registry.Get(toolNode.ToolName)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolNode.ToolName)
	}

	// Evaluate parameters using CEL
	evaluatedParams := make(map[string]interface{})
	celContext := mem.GetCELContext()

	// Import expression evaluator
	eval, err := createEvaluator()
	if err != nil {
		return "", fmt.Errorf("failed to create expression evaluator: %w", err)
	}

	for key, value := range toolNode.Parameters {
		// Check if value is a string that might contain CEL expressions
		if strValue, ok := value.(string); ok {
			// Evaluate template (supports ${ expr })
			evaluated, err := eval.EvaluateTemplate(strValue, celContext)
			if err != nil {
				return "", fmt.Errorf("failed to evaluate parameter '%s': %w", key, err)
			}
			evaluatedParams[key] = evaluated
		} else {
			// Non-string values pass through as-is
			evaluatedParams[key] = value
		}
	}

	// Execute tool with evaluated parameters
	result, err := handler(ctx, evaluatedParams)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// For explicit ToolNode execution, we append the result as a User message
	// to provide context to the LLM, as there is no preceding Assistant tool call.
	mem.Append(memory.Message{
		Role:    memory.RoleUser,
		Content: fmt.Sprintf("Tool '%s' output: %s", toolNode.ToolName, result),
	})

	eventChan <- engine.Event{
		Type:    engine.EventNodeEnd, // Or a specific ToolResult event
		NodeID:  node.GetID(),
		Payload: result,
	}

	return "", nil
}

// createEvaluator is a helper to create an expression evaluator
func createEvaluator() (*expression.Evaluator, error) {
	return expression.NewEvaluator()
}
