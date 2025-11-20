# Dify DSL Next - Go Runtime

A production-ready, event-driven runtime for executing Dify DSL workflows in Go with CEL expression support.

## Features

- ✅ **CEL Expressions**: Dynamic content with Common Expression Language (`${ }` syntax)
- ✅ **Message-Centric Design**: Conversation history is the core state
- ✅ **Event-Driven**: Streaming execution with real-time events
- ✅ **Tool Integration**: Extensible tool registry with dynamic parameters
- ✅ **Type-Safe**: Strong typing with Go structs and sealed interfaces
- ✅ **LLM Provider Abstraction**: Support for OpenAI (with streaming) and mock providers
- ✅ **Parallel Execution**: Run multiple branches concurrently (planned)
- ✅ **High Test Coverage**: >80% average, 100% for core memory module
- ✅ **Testable**: Comprehensive mocking for unit and integration tests

## Quick Start

### Installation

```bash
go build -o dsl-run cmd/dsl-run/main.go
```

### Running a Workflow

```bash
## Running Examples

### Simple Chat
```bash
go run cmd/dsl-run/main.go examples/simple_chat.yaml
```

### Tool Use Example
To run the tool use example with a real LLM (OpenAI), you need to set your API key:

```bash
export OPENAI_API_KEY=your-key
go run cmd/dsl-run/main.go -input "San Francisco" examples/tool_use.yaml
```

Or using Mock provider:
```bash
DSL_EXP_LLM_PROVIDER=mock DSL_EXP_MOCK_RESPONSE="Mock response" go run cmd/dsl-run/main.go -input "San Francisco" examples/tool_use.yaml
```


## Documentation

- **[CEL Expression Reference](docs/CEL_REFERENCE.md)** - Complete guide to using CEL expressions
- **[Expression Language Design](docs/EXPRESSION_LANGUAGE_DESIGN.md)** - Architecture and design decisions
- **[JSONSchema](docs/planning/dsl-schema.json)** - Formal schema definition for workflows
- **[System Prompt Design](docs/SYSTEM_PROMPT_DESIGN.md)** - Memory and prompt engineering
- **[Sealed Interfaces](docs/SEALED_INTERFACES.md)** - Tagged union pattern in Go

## CEL Expressions

Dynamic expressions can be used in:

### 1. Tool Parameters
```yaml
- id: get_weather
  type: tool
  tool_name: weather
  parameters:
    city: "${ inputs.query }"  # CEL expression
```

### 2. System Prompts
```yaml
- id: personalized_greeting
  type: generate
  model: gpt-4
  system_prompt: "Hello ${ inputs.user_name }, the weather in ${ inputs.city } is ${ nodes.weather.data.condition }"
```

### 3. Conditional Branching (Planned)
```yaml
- id: score_router
  type: selector
  cases:
    - if: "inputs.score >= 90"
      next: excellent
    - if: "inputs.score >= 60"
      next: pass
```

See [CEL_REFERENCE.md](docs/CEL_REFERENCE.md) for full syntax and examples.

## Project Structure

```
.
├── cmd/dsl-run/           # CLI entrypoint
├── internal/
│   ├── dsl/                # YAML parsing & validation
│   ├── engine/             # Runtime execution engine
│   ├── memory/             # Thread-safe conversation memory
│   ├── node/               # Node executors
│   ├── llm/                # LLM provider abstraction
│   └── integration/        # Integration tests
├── examples/               # Example workflows
└── docs/                   # Documentation
```

## Example Workflows

### 1. Simple Chat
```yaml
name: "Simple Chat"
nodes:
  - id: "greeting"
    type: "generate"
    model: "gpt-4o"
    system_prompt: "You are a helpful assistant."
```

### 2. Multi-Step Conversation
See `examples/multi_step_conversation.yaml` for a complex workflow with multiple generate nodes.

### 3. Conditional Routing
See `examples/conditional_routing.yaml` for intent-based routing using selectors.

### 4. Parallel Execution
See `examples/parallel_research.yaml` for concurrent execution with result synthesis.

### 5. Nested Workflows
See `examples/nested_workflow.yaml` for groups and nested parallel blocks.

## Testing

### Run all tests
```bash
go test ./...
```

### With coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### With race detection
```bash
go test -race ./...
```

### Integration tests
```bash
go test ./internal/integration/... -v
```

## LLM Provider Configuration

The runtime supports multiple LLM providers through a plugin architecture:

### Environment Variables

- `DSL_EXP_LLM_PROVIDER`: Choose provider (`openai` or `mock`)
- `OPENAI_API_KEY`: API key for OpenAI (when using openai provider)
- `DSL_EXP_MOCK_RESPONSE`: Response text for mock provider (testing only)

### Using Mock Provider for Testing

```bash
export DSL_EXP_LLM_PROVIDER=mock
export DSL_EXP_MOCK_RESPONSE="Hello from mock LLM!"
./dsl-run examples/simple_chat.yaml
```

### Programmatic Usage

```go
import "github.com/QuantumGhost/dsl-exp/internal/llm"

// Option 1: From environment
provider := llm.NewProviderFromEnv()

// Option 2: Explicit configuration
config := llm.ProviderConfig{
    Type:   llm.ProviderOpenAI,
    APIKey: "your-api-key",
}
provider := llm.NewProvider(config)

// Option 3: Mock for testing
mockProvider := llm.NewMockProvider("Test response")
```

## Development

### Adding a New Node Type

1. Define the node struct in `internal/dsl/types.go`
2. Add unmarshaling logic in `internal/dsl/loader.go`
3. Implement executor in `internal/node/`
4. Register executor in `cmd/dsl-run/main.go`

### Adding a New LLM Provider

1. Implement the `llm.Provider` interface
2. Add factory case in `internal/llm/factory.go`
3. Add tests in `internal/llm/`

## Architecture

### Memory Model
- **Global Memory**: Linear conversation timeline
- **Thread-Safe**: Uses `sync.RWMutex` for concurrent access
- **Snapshot**: Copy-on-write for parallel branches

### Event System
- `NodeStart`: Node begins execution
- `NodeEnd`: Node completes execution  
- `TokenGenerated`: LLM streams a token
- `ToolCall`: Tool invocation
- `Error`: Execution error

### Execution Flow
1. Load and validate YAML workflow
2. Initialize global memory with input messages
3. Create engine and register executors
4. Execute nodes sequentially/parallel
5. Stream events to consumer

## Test Coverage

Current coverage: **67.8%**

- Memory System: 100%
- Node Executors: 88.9%
- Engine: 87.1%
- DSL Loader: 84.3%

See `TEST_COVERAGE.md` for detailed analysis.

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch
3. Write tests for your changes
4. Ensure all tests pass with `go test -race ./...`
5. Submit a pull request
