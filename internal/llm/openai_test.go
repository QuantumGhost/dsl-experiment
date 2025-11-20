package llm

import (
	"os"
	"testing"
)

func TestNewOpenAIProvider_WithAPIKey(t *testing.T) {
	provider := NewOpenAIProvider("test-api-key")

	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}

func TestNewOpenAIProvider_WithEmptyKey(t *testing.T) {
	// Ensure no env key is set
	oldKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("OPENAI_API_KEY", oldKey)
		}
	}()
	os.Unsetenv("OPENAI_API_KEY")

	provider := NewOpenAIProvider("")

	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}

func TestNewOpenAIProvider_FromEnv(t *testing.T) {
	// Set env key
	oldKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("OPENAI_API_KEY", oldKey)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	os.Setenv("OPENAI_API_KEY", "env-test-key")

	provider := NewOpenAIProvider("")

	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}
