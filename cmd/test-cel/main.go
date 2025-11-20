package main

import (
	"fmt"
	"log"

	"github.com/QuantumGhost/dsl-exp/internal/expression"
)

func main() {
	eval, err := expression.NewEvaluator()
	if err != nil {
		log.Fatalf("Failed to create evaluator: %v", err)
	}

	// Test context
	context := map[string]interface{}{
		"inputs": map[string]interface{}{
			"user_name": "Alice",
			"city":      "San Francisco",
			"score":     85,
		},
		"nodes": map[string]interface{}{
			"weather": map[string]interface{}{
				"output": "Sunny",
				"data": map[string]interface{}{
					"temperature": 25,
					"humidity":    60,
				},
			},
			"classifier": map[string]interface{}{
				"output": "urgent",
			},
		},
		"vars": map[string]interface{}{
			"counter": 5,
		},
	}

	fmt.Println("=== CEL Expression Evaluator Tests ===")

	// Test 1: Simple property access
	fmt.Println("Test 1: Simple property access")
	result, err := eval.Evaluate("inputs.user_name", context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  inputs.user_name = %v\n", result)
	}

	// Test 2: Nested property access
	fmt.Println("\nTest 2: Nested property access")
	result, err = eval.Evaluate("nodes.weather.data.temperature", context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  nodes.weather.data.temperature = %v\n", result)
	}

	// Test 3: Boolean condition
	fmt.Println("\nTest 3: Boolean condition")
	boolResult, err := eval.EvaluateBool("inputs.score >= 60 && inputs.score < 90", context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  inputs.score >= 60 && inputs.score < 90 = %v\n", boolResult)
	}

	// Test 4: String comparison
	fmt.Println("\nTest 4: String comparison")
	boolResult, err = eval.EvaluateBool("nodes.classifier.output == 'urgent'", context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  nodes.classifier.output == 'urgent' = %v\n", boolResult)
	}

	// Test 5: Arithmetic
	fmt.Println("\nTest 5: Arithmetic")
	result, err = eval.Evaluate("inputs.score + 15", context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  inputs.score + 15 = %v\n", result)
	}

	// Test 6: Template interpolation
	fmt.Println("\nTest 6: Template interpolation")
	template := "Hello ${ inputs.user_name }! The weather in ${ inputs.city } is ${ nodes.weather.output } with temperature ${ nodes.weather.data.temperature }°C."
	templateResult, err := eval.EvaluateTemplate(template, context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Template: %s\n", template)
		fmt.Printf("  Result: %s\n", templateResult)
	}

	// Test 7: Complex template with expressions
	fmt.Println("\nTest 7: Complex template with expressions")
	template2 := "Score: ${ inputs.score }, Grade: ${ inputs.score >= 90 ? 'A' : (inputs.score >= 60 ? 'B' : 'C') }"
	templateResult2, err := eval.EvaluateTemplate(template2, context)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Template: %s\n", template2)
		fmt.Printf("  Result: %s\n", templateResult2)
	}

	fmt.Println("\n=== All tests completed ===")
}
