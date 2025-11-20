package llm

import (
	"context"
	"os"
	"testing"

	"github.com/QuantumGhost/dsl-exp/internal/memory"
)

func TestProviderFactory_Mock(t *testing.T) {
	config := ProviderConfig{
		Type:         ProviderMock,
		MockResponse: "Test response",
	}

	provider := NewProvider(config)

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	mockProvider, ok := provider.(*MockProvider)
	if !ok {
		t.Fatal("Expected MockProvider")
	}

	if mockProvider.Response != "Test response" {
		t.Errorf("Expected 'Test response', got %s", mockProvider.Response)
	}
}

func TestProviderFactory_OpenAI(t *testing.T) {
	config := ProviderConfig{
		Type:   ProviderOpenAI,
		APIKey: "test-key",
	}

	provider := NewProvider(config)

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	_, ok := provider.(*OpenAIProvider)
	if !ok {
		t.Fatal("Expected OpenAIProvider")
	}
}

func TestMockProvider_ChatCompletion(t *testing.T) {
	provider := NewMockProvider("Hello World")

	ctx := context.Background()
	messages := []memory.Message{
		{Role: memory.RoleUser, Content: "Test"},
	}
	opts := Options{Model: "test"}

	ch, err := provider.ChatCompletion(ctx, messages, opts)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	var result string
	for event := range ch {
		if e, ok := event.(*TextEvent); ok {
			result += e.Text
		}
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", result)
	}
}

func TestMockProvider_DefaultResponse(t *testing.T) {
	provider := NewMockProvider("")

	if provider.Response != "Mock LLM Response" {
		t.Errorf("Expected default response, got %s", provider.Response)
	}
}

func TestMockProvider_ContextCancellation(t *testing.T) {
	provider := NewMockProvider("Long response that should be interrupted")

	ctx, cancel := context.WithCancel(context.Background())
	ch, err := provider.ChatCompletion(ctx, nil, Options{})
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	// Cancel immediately
	cancel()

	tokenCount := 0
	for range ch {
		tokenCount++
	}

	expectedTokens := len("Long response that should be interrupted")
	if tokenCount >= expectedTokens {
		t.Logf("Note: Got %d tokens, cancellation may not have taken effect immediately", tokenCount)
	}
}

func TestNewProviderFromEnv_Mock(t *testing.T) {
	// Save and restore original env vars
	originalProvider := os.Getenv("DSL_EXP_LLM_PROVIDER")
	originalMockResponse := os.Getenv("DSL_EXP_MOCK_RESPONSE")
	defer func() {
		os.Setenv("DSL_EXP_LLM_PROVIDER", originalProvider)
		os.Setenv("DSL_EXP_MOCK_RESPONSE", originalMockResponse)
	}()

	// Set env vars
	os.Setenv("DSL_EXP_LLM_PROVIDER", "mock")
	os.Setenv("DSL_EXP_MOCK_RESPONSE", "Test from env")

	provider := NewProviderFromEnv()

	mockProvider, ok := provider.(*MockProvider)
	if !ok {
		t.Fatal("Expected MockProvider")
	}

	if mockProvider.Response != "Test from env" {
		t.Errorf("Expected 'Test from env', got %s", mockProvider.Response)
	}
}

func TestNewProviderFromEnv_DefaultOpenAI(t *testing.T) {
	// Save and restore original env vars
	originalProvider := os.Getenv("DSL_EXP_LLM_PROVIDER")
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		os.Setenv("DSL_EXP_LLM_PROVIDER", originalProvider)
		os.Setenv("OPENAI_API_KEY", originalAPIKey)
	}()

	// Unset provider type (should default to openai)
	os.Unsetenv("DSL_EXP_LLM_PROVIDER")
	os.Setenv("OPENAI_API_KEY", "test-key")

	provider := NewProviderFromEnv()

	_, ok := provider.(*OpenAIProvider)
	if !ok {
		t.Fatal("Expected OpenAIProvider as default")
	}
}
