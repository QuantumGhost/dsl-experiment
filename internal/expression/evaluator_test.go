package expression

import (
	"testing"
)

func TestEvaluator_SimpleExpression(t *testing.T) {
	eval, err := NewEvaluator()
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	context := map[string]interface{}{
		"inputs": map[string]interface{}{
			"name": "Alice",
			"age":  30,
		},
		"nodes": map[string]interface{}{},
		"vars":  map[string]interface{}{},
	}

	// Test simple property access
	result, err := eval.Evaluate("inputs.name", context)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if result != "Alice" {
		t.Errorf("Expected 'Alice', got %v", result)
	}
}

func TestEvaluator_BooleanExpression(t *testing.T) {
	eval, err := NewEvaluator()
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	context := map[string]interface{}{
		"inputs": map[string]interface{}{
			"score": 85,
		},
		"nodes": map[string]interface{}{},
		"vars":  map[string]interface{}{},
	}

	// Test boolean condition
	result, err := eval.EvaluateBool("inputs.score >= 60", context)
	if err != nil {
		t.Fatalf("Evaluation failed: %v", err)
	}

	if !result {
		t.Error("Expected true")
	}
}

func TestEvaluator_Template(t *testing.T) {
	eval, err := NewEvaluator()
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	context := map[string]interface{}{
		"inputs": map[string]interface{}{
			"city": "San Francisco",
		},
		"nodes": map[string]interface{}{
			"weather": map[string]interface{}{
				"data": map[string]interface{}{
					"temp": 25,
				},
			},
		},
		"vars": map[string]interface{}{},
	}

	template := "The weather in ${ inputs.city } is ${ nodes.weather.data.temp }°C"
	result, err := eval.EvaluateTemplate(template, context)
	if err != nil {
		t.Fatalf("Template evaluation failed: %v", err)
	}

	expected := "The weather in San Francisco is 25°C"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
