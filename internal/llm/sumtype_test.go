package llm

// TestIncompleteSwitch intentionally has an incomplete type switch to test go-sumtype
func checkIncompleteSwitch() {
	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		ch <- &TextEvent{Text: "test"}
	}()

	for event := range ch {
		switch e := event.(type) {
		case *TextEvent:
			_ = e.Text
			// Intentionally missing case ToolCallEvent to test go-sumtype detection
		}
	}
}

// TestCompleteSwitch has a complete type switch
func checkCompleteSwitch() {
	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		ch <- &TextEvent{Text: "test"}
	}()

	for event := range ch {
		switch e := event.(type) {
		case *TextEvent:
			_ = e.Text
		case *ToolCallEvent:
			_ = e.ToolCall
		}
	}
}

// TestWithDefault uses a default clause (should pass)
func checkWithDefault() {
	ch := make(chan StreamEvent)
	go func() {
		defer close(ch)
		ch <- &TextEvent{Text: "test"}
	}()

	for event := range ch {
		switch event.(type) {
		case *TextEvent:
			// handle text
		default:
			// handle others
		}
	}
}
