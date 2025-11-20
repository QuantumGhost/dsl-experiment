# dsl-exp - Project Architecture Summary

## Overview

dsl-exp is a minimal, event-driven runtime for executing dsl-exp workflows in Go. The project implements a message-centric architecture where conversation history serves as the core state, enabling natural AI Agent workflows with support for parallel execution, streaming responses, and extensible LLM providers.

## Core Philosophy

The system is built around the principle that **"Conversation is State"**:
- Input: A list of messages
- Output: A list of messages  
- Execution: A pipeline of nodes that append messages to history

This eliminates the complexity of traditional variable-based systems while maintaining expressiveness for AI workflows.

## Architecture Components

### 1. DSL Layer (`internal/dsl/`)

**Purpose**: YAML parsing, validation, and type definitions

**Key Files**:
- `types.go`: Defines all node types and workflow structure using sealed interfaces
- `loader.go`: YAML unmarshaling with custom node type detection and validation

**Core Design**:
- **Sealed Interface Pattern**: `Node` interface with compile-time type safety using go-sumtype
- **Custom Unmarshaling**: Dynamic type instantiation based on `type` field in YAML
- **Comprehensive Validation**: ID uniqueness, next pointer validation, and type-specific rules

**Node Types**:
- `GenerateNode`: LLM interaction with model, system_prompt, temperature, context_mode
- `ToolNode`: Tool execution with tool_name and CEL-evaluated parameters
- `ParallelNode`: Concurrent execution with branches and optional reducer
- `SelectorNode`: Conditional routing with LLM-based classification
- `GroupNode`: Nested workflow containers
- `NoopNode`: No-operation placeholders

### 2. Runtime Engine (`internal/engine/`)

**Purpose**: Core execution engine with event streaming and linear execution

**Key Files**:
- `engine.go`: Main execution loop with node mapping and sequential processing
- `types.go`: Event types and Executor interface

**Execution Model**:
- **Linear Execution**: Nodes executed sequentially using `next` pointers or executor return values
- **Event Streaming**: Real-time events through Go channels
- **Node Registry**: Dynamic executor registration by node type

**Event Types**:
- `EventNodeStart`: Node begins execution
- `EventNodeEnd`: Node completes execution
- `EventTokenGenerated`: LLM streams a token
- `EventToolCall`: Tool invocation
- `EventError`: Execution error

**Execution Flow**:
1. Load and validate YAML workflow
2. Initialize global memory with input messages and CEL context
3. Create node map for ID lookup
4. Execute nodes sequentially starting from first node
5. Stream events to consumer via channels

### 3. Memory System (`internal/memory/`)

**Purpose**: Thread-safe conversation state with CEL integration

**Key Files**:
- `memory.go`: Global memory with mutex protection and CEL context builder

**Design Principles**:
- **Thread Safety**: All operations protected by `sync.RWMutex`
- **Immutable Snapshots**: Copy-on-write for parallel branches
- **Rich Context**: Support for inputs, node outputs, and global variables

**Data Structure**:
```go
type Memory struct {
    mu          sync.RWMutex
    messages    []Message              // Linear conversation timeline
    Inputs      map[string]interface{} // Workflow inputs (e.g., user query)
    NodeOutputs map[string]interface{} // Node execution results
    Variables   map[string]interface{} // Global variables
}
```

**Message Types**:
- `RoleSystem`, `RoleUser`, `RoleAssistant`, `RoleTool`
- Support for `ToolCall` and `ToolCallID` for tool interactions

### 4. Node Executors (`internal/node/`)

**Purpose**: Implementation of execution logic for each node type

**Key Files**:
- `generate.go`: LLM interaction with context mode handling and memory accumulation
- `tool.go`: Tool execution with CEL parameter evaluation and registry system
- `noop.go`: Minimal implementation

**Features**:
- **CEL Integration**: Template evaluation for dynamic parameters (`${ expression }`)
- **Context Modes**: Full (default), window, none for memory control
- **Tool Registry**: Extensible tool system with function registration
- **Memory Accumulation**: Assistant responses automatically appended to global memory

**Tool Registry Pattern**:
```go
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)
type ToolRegistry struct { tools map[string]ToolHandler }
```

### 5. LLM Abstraction (`internal/llm/`)

**Purpose**: Provider abstraction with streaming support and mock testing

**Key Files**:
- `types.go`: Provider interface, stream events, and sealed interface pattern
- `factory.go`: Provider instantiation from environment or config
- `openai.go`: OpenAI provider implementation with streaming

**Design Patterns**:
- **Sealed Interfaces**: `StreamEvent` interface with compile-time exhaustiveness
- **Streaming First**: All responses are streamed for real-time experience
- **Environment Configuration**: Easy provider switching via environment variables

**Stream Events**:
- `TextEvent`: Text chunk from LLM
- `ToolCallEvent`: Tool call request from LLM
- Future: `ToolResultEvent`, `ErrorEvent`

**Provider Types**:
- `OpenAIProvider`: Full OpenAI API integration with openai-go v3
- `MockProvider`: Character-by-character streaming for testing

### 6. Expression Evaluation (`internal/expression/`)

**Purpose**: CEL (Common Expression Language) integration for dynamic evaluation

**Key Files**:
- `evaluator.go`: CEL environment setup with inputs/nodes/vars declarations

**Capabilities**:
- **Boolean Evaluation**: `EvaluateBool()` for conditional logic
- **Template Interpolation**: `EvaluateTemplate()` with `${ expr }` syntax
- **Rich Context**: Access to inputs, node outputs, and variables

**CEL Environment**:
```go
cel.Declarations(
    decls.NewVar("inputs", decls.NewMapType(decls.String, decls.Dyn)),
    decls.NewVar("nodes", decls.NewMapType(decls.String, decls.Dyn)),
    decls.NewVar("vars", decls.NewMapType(decls.String, decls.Dyn)),
)
```

## Entry Points

### CLI Application (`cmd/dsl-run/`)

**Purpose**: Command-line interface for workflow execution

**Features**:
- YAML workflow file loading with validation
- Interactive and batch input modes
- Real-time event streaming display
- LLM provider configuration via environment variables
- Tool registry with sample weather tool
- CEL context initialization with user input

**Key Implementation**:
```go
// Initialize global memory with user input
input := []memory.Message{{Role: memory.RoleUser, Content: userInput}}

// Set up CEL context for expression evaluation
mem.SetInputs(map[string]interface{}{"query": userInput})

// Event-driven execution loop
eventChan := eng.Run(ctx, wf, input)
for event := range eventChan {
    switch event.Type {
    case engine.EventTokenGenerated:
        fmt.Print(event.Payload.(string))
    // ...
    }
}
```

### Test Application (`cmd/test-cel/`)

**Purpose**: CEL expression testing utility for development

## Configuration

### Environment Variables
- `DSL_EXP_LLM_PROVIDER`: Choose provider (`openai` or `mock`, defaults to `openai`)
- `OPENAI_API_KEY`: API key for OpenAI provider
- `DSL_EXP_MOCK_RESPONSE`: Mock response text (defaults to "Mock LLM Response")

### Provider Factory Pattern
```go
// From environment (auto-detects)
provider := llm.NewProviderFromEnv()

// Explicit configuration
config := llm.ProviderConfig{
    Type:         llm.ProviderOpenAI,
    APIKey:       "your-api-key",
    MockResponse: "test response",
}
provider := llm.NewProvider(config)

// Direct mock for testing
mockProvider := llm.NewMockProvider("Custom mock response")
```

## Event System

The runtime uses an event-driven architecture with streaming through Go channels:

### Core Event Types
- `EventNodeStart`: Node begins execution (with NodeID)
- `EventNodeEnd`: Node completes execution (with NodeID and optional Payload)
- `EventTokenGenerated`: LLM streams a token (Payload: string)
- `EventToolCall`: Tool invocation (planned)
- `EventError`: Execution error (Payload: error)

### Event Flow
1. **Event Creation**: Executors emit events during execution
2. **Channel Streaming**: Events flow through Go channels to consumer
3. **Real-time Processing**: CLI displays tokens as they arrive
4. **Error Handling**: Errors propagate through event system

### Event Payload Patterns
```go
// Token streaming
eventChan <- engine.Event{
    Type:    engine.EventTokenGenerated,
    NodeID:  node.GetID(),
    Payload: "Hello",
}

// Tool result
eventChan <- engine.Event{
    Type:    engine.EventNodeEnd,
    NodeID:  node.GetID(),
    Payload: "Tool execution result",
}
```

## Workflow Examples

### Simple Chat
```yaml
name: "Simple Chat"
nodes:
  - id: "greeting"
    type: "generate"
    model: "gpt-4o"
    system_prompt: "You are a helpful assistant."
```

### Parallel Execution
```yaml
- id: "parallel_research"
  type: "parallel"
  branches:
    - nodes: # Branch A
        - id: "search_flights"
          type: "generate"
        - id: "exec_flights" 
          type: "tool"
    - nodes: # Branch B
        - id: "search_hotels"
          type: "generate"
        - id: "exec_hotels"
          type: "tool"
```

### Tool Use with CEL
```yaml
- id: "get_weather"
  type: "tool"
  tool_name: "weather"
  parameters:
    city: "${ inputs.query }"  # CEL template evaluation
```

### Selector-Based Routing
```yaml
- id: "classify_intent"
  type: "selector"
  model: "gpt-4o"
  system_prompt: "Classify user intent as question, task, or chat."
  options:
    - case: "question"
      next: "answer_question"
    - case: "task"
      next: "handle_task"
    - case: "chat"
      next: "casual_chat"
```

## Testing Strategy

### Test Coverage: 67.8%
- **Memory System**: 100% - Complete thread-safety and CEL context coverage
- **Node Executors**: 88.9% - Core functionality tested, some edge cases pending
- **Engine**: 87.1% - Linear execution and event system covered
- **DSL Loader**: 84.3% - YAML parsing and validation thoroughly tested

### Test Categories
- **Unit Tests**: Individual component testing with mocks
- **Integration Tests**: End-to-end workflow execution
- **Mock Providers**: Isolated testing without external dependencies
- **Race Detection**: `go test -race` for concurrent safety
- **Custom Mocks**: Test-specific behavior injection

### Test Implementation Pattern
```go
// Custom mock provider for behavior injection
type customMockProvider struct {
    chatFunc func(context.Context, []memory.Message, llm.Options) (<-chan llm.StreamEvent, error)
}

// Integration test with temporary workflow files
func TestIntegration_SimpleChatWorkflow(t *testing.T) {
    tmpDir := t.TempDir()
    workflowFile := filepath.Join(tmpDir, "chat.yaml")
    // ...
}
```

## Current Implementation Status

### ✅ Completed Features
- **Linear Execution**: Sequential node execution with next pointers
- **Event Streaming**: Real-time token streaming and event channels
- **LLM Integration**: OpenAI provider with streaming and mock support
- **Tool System**: Registry-based tool execution with CEL parameter evaluation
- **Memory Management**: Thread-safe conversation memory with snapshots
- **CEL Integration**: Template evaluation and boolean expressions
- **DSL Validation**: Comprehensive YAML parsing and validation
- **Sealed Interfaces**: Type safety with compile-time checking

### 🚧 In Progress (Phase 4)
- **Parallel Execution**: Parallel node executor with goroutines
- **Selector Routing**: LLM-based conditional routing implementation
- **Advanced Memory Policies**: result_only, full, summary strategies

### 📋 Planned (Phase 5-6)
- **Loop Constructs**: Explicit iteration over data
- **Reducer Nodes**: Result synthesis for parallel branches
- **Context Window**: Sliding window memory management
- **Observability**: Structured logging and metrics

## Development Workflow

### Adding New Node Types
1. **Define Type**: Add struct in `internal/dsl/types.go`
   ```go
   type CustomNode struct {
       BaseNode `yaml:",inline"`
       CustomField string `yaml:"custom_field"`
   }
   ```

2. **Update Loader**: Add case in `UnmarshalYAML` method
3. **Implement Executor**: Create in `internal/node/custom.go`
4. **Register**: Add to CLI executor registration
5. **Add Tests**: Unit and integration tests

### Adding New LLM Providers
1. **Implement Interface**: Satisfy `llm.Provider` interface
   ```go
   type CustomProvider struct { /* fields */ }
   func (p *CustomProvider) ChatCompletion(ctx context.Context, messages []memory.Message, opts llm.Options) (<-chan llm.StreamEvent, error) { /* implementation */ }
   ```

2. **Update Factory**: Add case in `internal/llm/factory.go`
3. **Add Configuration**: Extend `ProviderConfig` and environment variables
4. **Write Tests**: Mock and integration tests

## Key Design Decisions

### 1. Message-Centric State Model
- **Rationale**: Simplified mental model for conversational AI
- **Benefits**: Natural debugging, inspection, and persistence
- **Trade-offs**: Less suitable for non-conversational workflows

### 2. Event-Driven Streaming Architecture
- **Rationale**: Real-time user experience and efficient resource usage
- **Implementation**: Go channels with context cancellation
- **Benefits**: Natural backpressure handling and cancellation

### 3. Linear Execution Model
- **Rationale**: Predictable execution and easier debugging
- **Current Status**: Sequential node execution implemented
- **Future**: Parallel execution planned with goroutine coordination

### 4. Type Safety with Sealed Interfaces
- **Rationale**: Compile-time guarantee and better developer experience
- **Implementation**: go-sumtype for exhaustive pattern matching
- **Benefits**: IDE support, refactoring safety, and runtime performance

### 5. Plugin Architecture with Factory Pattern
- **Rationale**: Easy extensibility and clear separation of concerns
- **Implementation**: Interface-based design with environment-driven instantiation
- **Benefits**: Test isolation, provider swapping, and modular design

## Future Roadmap

### Phase 4: Parallelism & Expressions (In Progress)
- [x] CEL integration for dynamic evaluation
- [ ] Parallel node execution with goroutines
- [ ] Variable scoping with hierarchical memory
- [ ] Advanced context control

### Phase 5: Advanced Control Flow
- [ ] Selector nodes for intelligent routing
- [ ] Loop constructs for iteration
- [ ] Advanced error handling

### Phase 6: Observability & Polish
- [ ] Structured logging and tracing
- [ ] Metrics and monitoring
- [ ] CLI improvements and tooling

## Technology Stack

- **Language**: Go 1.25.3
- **YAML Processing**: gopkg.in/yaml.v3 with custom unmarshaling
- **LLM Integration**: github.com/openai/openai-go/v3 with streaming support
- **Expression Language**: github.com/google/cel-go for dynamic evaluation
- **Type Safety**: github.com/ajg/form through go-sumtype for sealed interfaces
- **Testing**: Standard Go testing with custom mocks and race detection
- **Concurrency**: Go channels and goroutines for streaming and parallelism

## Project Structure

```
.
├── cmd/
│   ├── dsl-run/          # CLI entrypoint with tool registration
│   └── test-cel/          # CEL expression testing utility
├── internal/
│   ├── dsl/               # YAML parsing with custom node unmarshaling
│   ├── engine/            # Linear execution engine with event streaming
│   ├── memory/            # Thread-safe memory with CEL context
│   ├── node/              # Executor implementations (generate, tool, noop)
│   ├── llm/               # LLM provider abstraction (OpenAI, Mock)
│   ├── expression/        # CEL evaluator for templates and booleans
│   └── integration/       # End-to-end integration tests
├── examples/              # Example workflows (simple, tool, conditional)
└── docs/                  # Documentation and design specs
```

## Implementation Notes

### Current Limitations
- **Linear Execution Only**: Parallel executor not yet implemented
- **Selector Not Implemented**: Conditional routing planned but not coded
- **No Memory Policies**: Context modes defined but not fully enforced
- **Limited Tool Call Integration**: OpenAI tool call streaming not complete

### Architecture Strengths
- **Clean Separation**: Each component has clear responsibilities
- **Extensible Design**: New node types and providers easily added
- **Type Safe**: Sealed interfaces prevent runtime errors
- **Test Friendly**: Mock providers and comprehensive test coverage
- **Production Ready**: Thread-safe memory and proper error handling

This architecture provides a solid foundation for building sophisticated AI Agent workflows while maintaining simplicity and extensibility. The current implementation focuses on core functionality with linear execution, laying the groundwork for advanced features like parallel processing and intelligent routing in future phases.
