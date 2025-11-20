package llm

import (
	"context"
	"os"

	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

// ProviderType defines the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderMock   ProviderType = "mock"
)

// ProviderConfig holds configuration for creating providers
type ProviderConfig struct {
	Type   ProviderType
	APIKey string

	// Mock-specific config
	MockResponse string
}

// NewProvider creates a provider based on configuration
func NewProvider(config ProviderConfig) Provider {
	switch config.Type {
	case ProviderMock:
		return NewMockProvider(config.MockResponse)
	case ProviderOpenAI:
		fallthrough
	default:
		return NewOpenAIProvider(config.APIKey)
	}
}

// NewProviderFromEnv creates a provider from environment variables
func NewProviderFromEnv() Provider {
	providerType := os.Getenv("DSL_EXP_LLM_PROVIDER")
	if providerType == "" {
		providerType = "openai"
	}

	config := ProviderConfig{
		Type:         ProviderType(providerType),
		APIKey:       os.Getenv("OPENAI_API_KEY"),
		MockResponse: os.Getenv("DSL_EXP_MOCK_RESPONSE"),
	}

	return NewProvider(config)
}

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	Response string
}

func NewMockProvider(response string) *MockProvider {
	if response == "" {
		response = "Mock LLM Response"
	}
	return &MockProvider{Response: response}
}

func (m *MockProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts Options) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		for _, char := range m.Response {
			select {
			case <-ctx.Done():
				return
			case ch <- &TextEvent{Text: string(char)}:
			}
		}
	}()
	return ch, nil
}
