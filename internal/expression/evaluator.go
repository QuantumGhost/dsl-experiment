package expression

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

// Evaluator handles CEL expression evaluation
type Evaluator struct {
	env *cel.Env
}

// NewEvaluator creates a new CEL evaluator with standard environment
func NewEvaluator() (*Evaluator, error) {
	// Define the environment with standard variables
	// inputs: map[string]any
	// nodes: map[string]any
	// vars: map[string]any
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("inputs", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("nodes", decls.NewMapType(decls.String, decls.Dyn)),
			decls.NewVar("vars", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL env: %w", err)
	}
	return &Evaluator{env: env}, nil
}

// Evaluate compiles and evaluates a CEL expression
func (e *Evaluator) Evaluate(expr string, context map[string]interface{}) (interface{}, error) {
	ast, issues := e.env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile error: %w", issues.Err())
	}

	prg, err := e.env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program creation error: %w", err)
	}

	out, _, err := prg.Eval(context)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %w", err)
	}

	return out.Value(), nil
}

// EvaluateBool evaluates an expression expecting a boolean result
func (e *Evaluator) EvaluateBool(expr string, context map[string]interface{}) (bool, error) {
	val, err := e.Evaluate(expr, context)
	if err != nil {
		return false, err
	}
	if b, ok := val.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("expression result is not a boolean: %v", val)
}

var templateRegex = regexp.MustCompile(`\$\{\s*(.*?)\s*\}`)

// EvaluateTemplate replaces ${ expr } in a string with evaluated results
func (e *Evaluator) EvaluateTemplate(template string, context map[string]interface{}) (string, error) {
	var err error
	result := templateRegex.ReplaceAllStringFunc(template, func(match string) string {
		if err != nil {
			return match
		}
		// Extract expression inside ${ }
		// match is like "${ expr }"
		// We need to be careful with parsing.
		// Simple regex might fail on nested braces, but CEL doesn't use {} for blocks usually.
		// Let's assume simple cases for now.

		expr := match[2 : len(match)-1]
		expr = strings.TrimSpace(expr)

		val, evalErr := e.Evaluate(expr, context)
		if evalErr != nil {
			err = evalErr // Capture error
			return match  // Return original on error
		}
		return fmt.Sprintf("%v", val)
	})

	if err != nil {
		return "", err
	}
	return result, nil
}
