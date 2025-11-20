package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/QuantumGhost/dsl-exp/internal/memory"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIProvider struct {
	client openai.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIProvider{client: client}
}

// ChatCompletion streams the response from OpenAI.
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts Options) (<-chan StreamEvent, error) {
	chatMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case memory.RoleSystem:
			chatMessages[i] = openai.SystemMessage(msg.Content)
		case memory.RoleUser:
			chatMessages[i] = openai.UserMessage(msg.Content)
		case memory.RoleAssistant:
			chatMessages[i] = openai.AssistantMessage(msg.Content)
		case memory.RoleTool:
			// Handle tool outputs
			chatMessages[i] = openai.ToolMessage(msg.ToolCallID, msg.Content)
		default:
			chatMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Messages: chatMessages,
		Model:    openai.ChatModel(opts.Model),
	}

	if opts.Temperature != 0 {
		params.Temperature = openai.Float(opts.Temperature)
	}

	// TODO: Add Tools to params if present in opts

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)
	ch := make(chan StreamEvent)

	go func() {
		defer close(ch)
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					select {
					case <-ctx.Done():
						return
					case ch <- &TextEvent{Text: delta.Content}:
					}
				}
				// TODO: Handle Tool Calls in Delta
			}
		}
		if err := stream.Err(); err != nil {
			// Log error or handle it
			fmt.Printf("Stream error: %v\n", err)
		}
	}()

	return ch, nil
}
