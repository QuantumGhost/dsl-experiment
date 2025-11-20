# Sealed Interfaces in dsl-exp

## Overview

This project uses **sealed interfaces** (also known as tagged unions or sum types) to model discriminated unions in Go. This pattern provides better type safety and eliminates redundant fields in structs.

## What is a Sealed Interface?

A sealed interface is an interface with a limited set of implementations, enforced by an unexported marker method. This allows us to use type switches to handle all possible cases exhaustively.

## Implementation Pattern

```go
// 1. Define the sealed interface with an unexported marker method
//sumtype:decl
type StreamEvent interface {
    isStreamEvent() // Unexported marker method
}

// 2. Define concrete implementations
type TextEvent struct {
    Text string
}

func (TextEvent) isStreamEvent() {} // Implements the marker

type ToolCallEvent struct {
    ToolCall ToolCall
}

func (ToolCallEvent) isStreamEvent() {}

// 3. Use type switches to handle all cases
func handleEvent(event StreamEvent) {
    switch e := event.(type) {
    case TextEvent:
        fmt.Println("Text:", e.Text)
    case ToolCallEvent:
        fmt.Println("Tool call:", e.ToolCall.Name)
    }
}
```

## Benefits

1. **Type Safety**: The compiler ensures all fields are properly typed
2. **No Redundant Fields**: Each variant has only the fields it needs
3. **Exhaustiveness Checking**: `go-sumtype` verifies all cases are handled
4. **Clear Intent**: Code explicitly shows which variant is being used

## Current Usage

### `llm.StreamEvent`

Located in `internal/llm/types.go`:

```go
//sumtype:decl
type StreamEvent interface {
    isStreamEvent()
}

type TextEvent struct {
    Text string
}

type ToolCallEvent struct {
    ToolCall ToolCall
}
```

**Usage example** (from `internal/node/generate.go`):

```go
for event := range stream {
    switch e := event.(type) {
    case llm.TextEvent:
        // Handle text chunk
        fullResponse.WriteString(e.Text)
    case llm.ToolCallEvent:
        // Handle tool call
        handleToolCall(e.ToolCall)
    }
}
```

### `memory.Message` (Planned)

Currently `memory.Message` is still a struct with optional fields marked with `omitempty`. This will be refactored to a sealed interface in a future change.

Planned design:

```go
//sumtype:decl
type Message interface {
    isMessage()
    GetContent() string
}

type SystemMessage struct {
    Content string
}

type UserMessage struct {
    Content string
}

type AssistantMessage struct {
    Content   string
    ToolCalls []ToolCall
}

type ToolMessage struct {
    Content    string
    ToolCallID string
}
```

## Verification with `go-sumtype`

We use [`go-sumtype`](https://github.com/BurntSushi/go-sumtype) to verify that all type switches on sealed interfaces are exhaustive.

### Running the Check

```bash
# Standalone
go-sumtype ./...

# Via Makefile
make lint
```

### Adding a New Sealed Interface

1. Define the interface with the `//sumtype:decl` comment
2. Implement all variants with the marker method
3. Run `go-sumtype ./...` to verify exhaustiveness
4. Document it in this file

## Anti-Patterns to Avoid

❌ **Don't use type assertions without checking all cases**:

```go
// Bad: Missing ToolCallEvent case
if e, ok := event.(llm.TextEvent); ok {
    handleText(e.Text)
}
// What if it's a ToolCallEvent?
```

✅ **Use type switches and handle all cases**:

```go
// Good: Exhaustive
switch e := event.(type) {
case llm.TextEvent:
    handleText(e.Text)
case llm.ToolCallEvent:
    handleToolCall(e.ToolCall)
}
```

❌ **Don't add fields to the interface for specific variants**:

```go
// Bad: Not all events have ToolCall
type StreamEvent interface {
    isStreamEvent()
    GetToolCall() *ToolCall // Wrong!
}
```

✅ **Fields belong to concrete types**:

```go
// Good: Only ToolCallEvent has ToolCall
type ToolCallEvent struct {
    ToolCall ToolCall
}
```

## References

- [Go Proverbs: "Accept interfaces, return structs"](https://go-proverbs.github.io/)
- [go-sumtype](https://github.com/BurntSushi/go-sumtype)
- [Tagged Unions in Go](https://www.jerf.org/iri/post/2017/go_interfaces_for_sum_types/)
