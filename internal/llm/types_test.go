package llm

import (
	"testing"
)

func TestOptions_Defaults(t *testing.T) {
	opts := Options{}

	if opts.Model != "" {
		t.Errorf("Expected empty model, got %s", opts.Model)
	}

	if opts.Temperature != 0 {
		t.Errorf("Expected 0 temperature, got %f", opts.Temperature)
	}
}

func TestOptions_WithValues(t *testing.T) {
	opts := Options{
		Model:       "gpt-4o",
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	if opts.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", opts.Model)
	}

	if opts.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", opts.Temperature)
	}

	if opts.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", opts.MaxTokens)
	}
}
