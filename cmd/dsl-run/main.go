package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/QuantumGhost/dsl-exp/internal/dsl"
	"github.com/QuantumGhost/dsl-exp/internal/engine"
	"github.com/QuantumGhost/dsl-exp/internal/llm"
	"github.com/QuantumGhost/dsl-exp/internal/memory"
	"github.com/QuantumGhost/dsl-exp/internal/node"
)

func main() {
	inputPtr := flag.String("input", "", "User input message")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: dsl-run <workflow.yaml> [-input 'message']")
		os.Exit(1)
	}

	path := args[0]
	wf, err := dsl.Load(path)
	if err != nil {
		log.Fatalf("Failed to load workflow: %v", err)
	}

	fmt.Printf("Loaded workflow: %s\n", wf.Name)

	// Initialize LLM Provider
	llmProvider := llm.NewProviderFromEnv()

	// Initialize Tool Registry
	toolRegistry := node.NewToolRegistry()
	toolRegistry.Register("weather", func(ctx context.Context, args map[string]interface{}) (string, error) {
		city, _ := args["city"].(string)
		if city == "" {
			city = "Unknown"
		}
		return fmt.Sprintf("The weather in %s is Sunny, 25°C", city), nil
	})

	// Initialize Engine
	eng := engine.NewEngine()
	eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{Provider: llmProvider})
	eng.RegisterExecutor(dsl.NodeTool, &node.ToolExecutor{Registry: toolRegistry})
	eng.RegisterExecutor(dsl.NodeSelector, &node.SelectorExecutor{})
	eng.RegisterExecutor(dsl.NodeNoop, &node.NoopExecutor{})

	// Prepare Input
	userInput := *inputPtr
	if userInput == "" {
		// Interactive mode if no input flag
		fmt.Print("User Input: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			userInput = scanner.Text()
		}
	}

	ctx := context.Background()
	input := []memory.Message{{Role: memory.RoleUser, Content: userInput}}

	fmt.Println("--- Start Execution ---")
	eventChan := eng.Run(ctx, wf, input)

	for event := range eventChan {
		switch event.Type {
		case engine.EventTokenGenerated:
			fmt.Print(event.Payload.(string))
		case engine.EventNodeStart:
			fmt.Printf("\n[Node Start: %s]\n", event.NodeID)
		case engine.EventNodeEnd:
			fmt.Printf("\n[Node End: %s]\n", event.NodeID)
			if str, ok := event.Payload.(string); ok && str != "" {
				// Print tool output or other results
				fmt.Printf("Result: %s\n", str)
			}
		case engine.EventError:
			fmt.Printf("\n[Error] %v\n", event.Payload)
		}
	}
	fmt.Println("\n--- End Execution ---")
}
