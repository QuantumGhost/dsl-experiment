# Expression Language Design: CEL Integration

## Overview

dsl-exp adopts **CEL (Common Expression Language)** as its standard expression language. CEL is a non-Turing complete language designed for simplicity, speed, safety, and portability. It is widely used in cloud-native ecosystems (Kubernetes, Envoy) and by OpenAI's Agent Builder.

## Why CEL?

1.  **Safety**: CEL is designed to terminate in bounded time (no infinite loops) and has no side effects. This is critical for executing user-defined logic in a secure SaaS environment.
2.  **Type Safety**: CEL supports compile-time type checking, allowing us to validate workflow logic before execution.
3.  **Performance**: CEL evaluates efficiently (often in microseconds) and has a high-performance Go implementation (`github.com/google/cel-go`).
4.  **Expressiveness**: It supports logical operations, math, string manipulation, list/map operations, and macros, covering all needs for variable substitution, conditional branching, and data transformation.

## Scope & Context

All CEL expressions in dsl-exp are evaluated within a specific context. The top-level variables available in the global scope are:

| Variable | Type | Description |
| :--- | :--- | :--- |
| `inputs` | `Map<String, Any>` | The initial inputs provided to the workflow execution. |
| `nodes` | `Map<String, NodeOutput>` | The outputs of all executed nodes, keyed by Node ID. |
| `vars` | `Map<String, Any>` | Global variables that can be read/written during execution (e.g., conversation context). |
| `env` | `Map<String, String>` | (Optional) Safe environment variables or secrets allowed for the workflow. |

### Node Output Structure (`NodeOutput`)

Each node's output is structured. For example:

```go
type NodeOutput struct {
    Output string                 // The primary text output (for LLM nodes)
    Data   map[string]interface{} // Structured data (for Tool/JSON nodes)
    // Metadata, Status, etc.
}
```

Accessing a tool's result: `nodes.weather_tool.data.temperature`
Accessing an LLM's text: `nodes.llm_step.output`

## Usage Scenarios

### 1. String Interpolation (Template Substitution)

In text fields (like Prompts, User Inputs), CEL expressions can be embedded using the syntax `${ expression }`.

**Example:**
```yaml
system_prompt: "Hello ${ inputs.user_name }, the weather in ${ inputs.city } is ${ nodes.weather.data.condition }."
```

**Implementation:**
The runtime will parse strings, identify `${...}` blocks, compile and evaluate the CEL expression, and replace the block with the string representation of the result.

### 2. Conditional Logic (Selector Node)

In `SelectorNode` (Branching), CEL expressions are used directly as boolean predicates.

**Example:**
```yaml
type: selector
cases:
  - if: "inputs.score >= 60 && inputs.score < 90"
    next: pass
  - if: "inputs.score >= 90"
    next: distinction
  - if: "true" # Default case
    next: fail
```

### 3. Parameter Mapping (Tool Node)

When passing parameters to tools, CEL can be used to construct complex values.

**Example:**
```yaml
type: tool
tool_name: http_request
parameters:
  url: "https://api.example.com/data"
  method: "POST"
  # Constructing a JSON object dynamically
  body:
    user_id: "${ inputs.user_id }"
    timestamp: "${ nodes.timer.data.now }"
    tags: "${ inputs.tags.filter(t, t.startsWith('admin_')) }" # Using CEL macros
```

## Implementation Strategy

1.  **Dependency**: Import `github.com/google/cel-go`.
2.  **Environment Setup**: Create a shared CEL `env` that defines the types for `inputs`, `nodes`, etc.
3.  **Evaluation**:
    *   For **Conditions**: Compile as `bool`.
    *   For **Interpolation**: Parse string, extract `${ expr }`, compile `expr` (expecting `string` or auto-converting), evaluate, and join.
    *   For **Values**: Compile and evaluate to `any`.

## Migration from Old Syntax

*   Old: `{{#node_id.output#}}`
*   New: `${ nodes.node_id.output }`

The new syntax is more standard, supports complex logic within the brackets, and is clearly distinguishable.
