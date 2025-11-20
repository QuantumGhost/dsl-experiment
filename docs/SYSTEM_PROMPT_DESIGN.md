# System Prompt vs User Prompt Design

## Architecture Decision

### Separation of Concerns

The runtime maintains a clear separation between:

1. **System Prompts** - Instructions for the LLM (ephemeral, per-node)
2. **Conversation History** - User/Assistant/Tool messages (persistent, in GlobalMemory)

## Design Principles

### 1. System Prompts Are NOT Stored in Memory

**Rationale**:
- System prompts are configuration, not conversation
- Each `generate` node may have a different system prompt
- Storing them would pollute the conversation history
- They should be ephemeral and node-specific

**Implementation**:
```go
// System prompt comes from node configuration
if genNode.SystemPrompt != "" {
    messages = append(messages, memory.Message{
        Role:    memory.RoleSystem,
        Content: genNode.SystemPrompt,
    })
}

// Then add conversation history (does NOT contain system messages)
messages = append(messages, mem.Snapshot()...)
```

### 2. GlobalMemory Only Contains Conversation

GlobalMemory stores the actual conversation that happened:
- ✅ User messages
- ✅ Assistant responses
- ✅ Tool calls/results
- ❌ NOT system prompts

**Example Flow**:

```yaml
nodes:
  - id: "greeting"
    type: "generate"
    system_prompt: "You are friendly"  # Not stored in memory
    
  - id: "analysis"  
    type: "generate"
    system_prompt: "You are analytical"  # Different prompt, not stored
```

**Memory Timeline**:
```
User: "Hello"
Assistant: "Hi! How can I help?"  # From greeting node
Assistant: "Let me analyze..."    # From analysis node, different behavior
```

The memory only has the conversation, not the instructions.

### 3. Message Structure for LLM Calls

When calling the LLM, messages are structured as:

```
[System Message]      ← From node.system_prompt
[User Message]        ← From GlobalMemory
[Assistant Message]   ← From GlobalMemory
[User Message]        ← From GlobalMemory
...
```

The system message is **prepended** at call time, not stored.

## Context Modes

The `context_mode` controls what history is sent:

### Full (default)
```go
case dsl.ContextFull:
    messages = append(messages, history...)
```
Sends: System + All conversation history

### None
```go
case dsl.ContextNone:
    // Don't add history
```
Sends: System only (no conversation context)

### Window
```go
case dsl.ContextWindow:
    // TODO: Implement window logic
    messages = append(messages, history[len(history)-10:]...)
```
Sends: System + Last N messages

## Testing

Tests verify this separation:

### Test: System Prompts Not in Memory
```go
func TestGenerateExecutor_WithSystemPrompt(t *testing.T) {
    // ... execute node with system prompt
    
    snapshot := mem.Snapshot()
    for _, msg := range snapshot {
        if msg.Role == memory.RoleSystem {
            t.Error("System message should not be in memory")
        }
    }
}
```

### Test: Context Modes
```go
func TestGenerateExecutor_ContextModes(t *testing.T) {
    // Verify ContextNone sends no history
    // Verify ContextFull sends all history
}
```

## Benefits

1. **Clarity**: Clear separation between instructions and conversation
2. **Flexibility**: Each node can have different system prompts
3. **Cleaner Memory**: GlobalMemory only has actual conversation
4. **Testability**: Easy to verify what gets sent to LLM vs stored
5. **Performance**: Don't store redundant system messages

## Example: Multi-Agent Workflow

```yaml
nodes:
  - id: "researcher"
    type: "generate"
    system_prompt: "You are a research assistant. Focus on facts."
    next: "writer"
    
  - id: "writer"
    type: "generate"  
    system_prompt: "You are a creative writer. Make it engaging."
```

**Conversation in Memory**:
```
User: "Tell me about AI"
Assistant: "AI is a field of computer science..." [from researcher]
Assistant: "Let me tell you a story about AI..." [from writer]
```

**What LLM Sees**:

*Researcher node*:
```
System: You are a research assistant. Focus on facts.
User: Tell me about AI
```

*Writer node*:
```
System: You are a creative writer. Make it engaging.
User: Tell me about AI
Assistant: AI is a field of computer science...
```

Note how the system prompts are different but not in memory!

## Migration Guide

If you have workflows that expect system messages in memory:

### Before (Incorrect)
```yaml
# DON'T: Expecting system message in history
- id: "node1"
  system_prompt: "Be helpful"
- id: "node2"  
  system_prompt: "{{memory[-1]}}"  # Trying to reference previous system
```

### After (Correct)
```yaml
# DO: Each node has its own system prompt
- id: "node1"
  system_prompt: "Be helpful"
- id: "node2"
  system_prompt: "Continue being helpful"  # Explicit instruction
```

Or use a shared system prompt if needed:
```yaml
system_prompt: &shared_prompt "Be helpful and concise"

nodes:
  - id: "node1"
    system_prompt: *shared_prompt
  - id: "node2"
    system_prompt: *shared_prompt
```

## Implementation Details

See: `internal/node/generate.go` for the complete implementation.

Key code sections:
- Lines 24-38: System prompt addition
- Lines 40-61: Context mode handling
- Lines 79-84: Memory append (no system messages)
