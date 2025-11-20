# CEL Expression Reference

## Overview

Dify DSL Next uses **CEL (Common Expression Language)** for dynamic content and conditional logic. CEL expressions can be embedded in specific fields using the `${ expression }` syntax.

## Where Can You Use CEL?

### 1. Generate Node - `system_prompt`

The system prompt supports CEL template interpolation for dynamic prompts.

**Example:**

```yaml
- id: personalized_greeting
  type: generate
  model: gpt-4
  system_prompt: "You are a helpful assistant for ${ inputs.user_name }. The user is from ${ inputs.city } and prefers ${ inputs.language } language."
```

**Available Variables:**
- `inputs.*` - Workflow inputs (e.g., `inputs.query`, `inputs.user_name`)
- `nodes.<node_id>.output` - Text output from previous nodes
- `nodes.<node_id>.data.*` - Structured data from tool nodes
- `vars.*` - Global variables

### 2. Tool Node - `parameters`

Tool parameter values support CEL expressions for dynamic values.

**Example:**

```yaml
- id: get_weather
  type: tool
  tool_name: weather
  parameters:
    city: "${ inputs.city }"
    units: "${ vars.preferred_units }"
    language: "${ inputs.language }"
```

**Complex Example with Expressions:**

```yaml
- id: send_email
  type: tool
  tool_name: email
  parameters:
    to: "${ inputs.recipient }"
    subject: "${ 'Report for ' + inputs.date }"
    body: "${ 'Weather: ' + nodes.weather.data.condition + ', Temp: ' + string(nodes.weather.data.temp) + '°C' }"
```

### 3. Selector Node - `cases[].if`

Conditional expressions use pure CEL (not template interpolation).

**Example:**

```yaml
- id: score_classifier
  type: selector
  cases:
    - if: "inputs.score >= 90"
      next: excellent
    - if: "inputs.score >= 60 && inputs.score < 90"
      next: pass  
    - if: "inputs.score < 60"
      next: fail
    - if: "true"  # Default case
      next: unknown
```

**Advanced Example:**

```yaml
- id: smart_router
  type: selector
  cases:
    - if: "inputs.urgent == true && inputs.user_level > 5"
      next: priority_handler
    - if: "nodes.classifier.output == 'technical' && inputs.tags.exists(t, t.startsWith('bug-'))"
      next: tech_support
    - if: "nodes.sentiment.data.score < 0.3"
      next: escalation
    - if: "true"
      next: general_queue
```

## CEL Context

### `inputs`

The workflow's initial inputs. Set by the engine from user messages.

**Structure:**
```javascript
{
  "query": "string - First user message content"
  // Custom fields can be added by the runtime
}
```

### `nodes`

Outputs from executed nodes, keyed by node ID.

**Structure:**
```javascript
{
  "<node_id>": {
    "output": "string - Primary text output (for LLM nodes)",
    "data": {
      // Structured data (for Tool nodes)
      // Example from weather tool:
      "temperature": 25,
      "condition": "Sunny",
      "humidity": 60
    }
  }
}
```

### `vars`

Global variables that persist across the workflow.

**Structure:**
```javascript
{
  "counter": 5,
  "user_preferences": {...},
  // Any variables set during execution
}
```

## CEL Language Features

### Supported Operations

| Category | Examples |
|----------|----------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `%` |
| **Comparison** | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| **Logical** | `&&`, `\|\|`, `!` |
| **String** | `+` (concatenation), `.startsWith()`, `.endsWith()`, `.contains()` |
| **Ternary** | `condition ? value_if_true : value_if_false` |
| **List/Map** | `in`, `.exists()`, `.filter()`, `.map()` |

### Type Conversion

```cel
string(42)           // "42"
int("42")            // 42
double("3.14")       // 3.14
```

### String Operations

```cel
"hello " + "world"                        // "hello world"
"Hello".startsWith("He")                  // true
"test@example.com".contains("@")          // true
"UPPERCASE".toLowerCase()                 // "uppercase"
```

### Ternary Operator

```cel
inputs.score >= 60 ? 'Pass' : 'Fail'
```

### List Operations

```cel
inputs.tags.exists(t, t == 'urgent')      // Check if 'urgent' exists in tags
inputs.scores.filter(s, s > 80)           // Filter scores > 80
inputs.values.map(v, v * 2)               // Double all values
```

## Common Patterns

### 1. Default Values

```cel
inputs.city != null ? inputs.city : 'New York'
```

### 2. Complex Conditions

```cel
(inputs.score >= 60 && inputs.attendance > 0.8) || inputs.extra_credit == true
```

### 3. String Formatting

```yaml
system_prompt: "User ${ inputs.name } (Level ${ inputs.level }) asks: ${ inputs.query }"
```

### 4. Nested Property Access

```cel
nodes.api_call.data.response.users[0].email
```

### 5. Safe Navigation (with defaults)

```cel
nodes.weather.data.temperature != null ? nodes.weather.data.temperature : 20
```

## Debugging Tips

1. **Use simple expressions first**: Start with `${ inputs.query }` and gradually add complexity
2. **Check types**: Remember that CEL is typed - use `string()`, `int()` etc. for conversions
3. **Test conditions**: Use true/false literals to test selector branching
4. **Inspect context**: Add a debug node that outputs `${ inputs }` or `${ nodes }`

## Limitations

1. **No side effects**: CEL cannot modify state, only read and compute
2. **No loops**: Use `.map()`, `.filter()` instead of for-loops
3. **Function scope**: Only built-in CEL functions are available (no custom UDFs yet)
4. **Evaluation timeout**: Complex expressions may timeout (implementation-dependent)

## Examples

### Complete Workflow with CEL

```yaml
name: Smart Customer Support
version: "1.0"
nodes:
  - id: classify_intent
    type: generate
    model: gpt-3.5-turbo
    system_prompt: "Classify this query: ${ inputs.query }. Respond with: technical, billing, or general"
  
  - id: route_request
    type: selector  
    cases:
      - if: "nodes.classify_intent.output.contains('technical')"
        next: tech_handler
      - if: "nodes.classify_intent.output.contains('billing')"
        next: billing_handler
      - if: "true"
        next: general_handler
  
  - id: tech_handler
    type: generate
    model: gpt-4
    system_prompt: "You are a technical support expert. User query: ${ inputs.query }"
  
  - id: get_account_info
    type: tool
    tool_name: database_query
    parameters:
      query: "SELECT * FROM accounts WHERE email = '${ inputs.email }'"
  
  - id: billing_handler
    type: generate
    model: gpt-4
    system_prompt: "You are a billing specialist. User: ${ inputs.email }, Balance: $${ nodes.get_account_info.data.balance }"
  
  - id: general_handler
    type: generate
    model: gpt-3.5-turbo
    system_prompt: "You are a helpful assistant. Query: ${ inputs.query }"
```

## References

- [CEL Specification](https://github.com/google/cel-spec)
- [CEL Go Implementation](https://github.com/google/cel-go)
- [Expression Language Design](./EXPRESSION_LANGUAGE_DESIGN.md)
