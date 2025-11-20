# Loop Examples

This directory contains examples demonstrating different loop patterns in Dify DSL Next.

## Overview

Loops in Dify DSL Next are implemented using **selector nodes** that conditionally jump back to previous nodes. The `selector` node evaluates CEL expressions and routes to different nodes based on conditions.

## Loop Patterns

### 1. Interactive Loop (`quiz_loop.yaml`)

**Pattern**: User-driven continuation with exit condition

**Use Cases**:
- Interactive quizzes
- Conversational games
- Step-by-step tutorials

**Key Features**:
- User can exit anytime by saying "exit", "quit", or "stop"
- LLM detects exit intent and routes to goodbye message
- Loop continues until user exits or task completes

**Loop Structure**:
```
start → check_answer → evaluate_response
           ↑                  ↓
           └──────────────────┘ (loop back)
                    or
                    ↓
              goodbye / final_score
```

**CEL Expression**:
```yaml
- if: "nodes.check_answer.output.contains('EXIT')"
  next: goodbye  # Exit loop

- if: "true"
  next: check_answer  # Continue loop
```

### 2. Iterative Processing (`iterative_processor.yaml`)

**Pattern**: Process multiple items with nested loop control

**Use Cases**:
- Task lists
- Multi-step workflows
- Data processing pipelines

**Key Features**:
- Processes items one at a time
- Allows user to request more help (inner loop)
- Checks for completion after each item
- Multiple exit conditions

**Loop Structure**:
```
get_tasks → parse_tasks → process_next_task
                              ↓
                         check_continuation
                         ↓    ↓    ↓
                         ↓    ↓    clarify_intent
                         ↓    ↓         ↓
                         ↓    └─────────┘ (inner loop)
                         ↓
                    check_completion
                         ↓
                   completion_router
                    ↓         ↓
                summary      └→ process_next_task (outer loop)
```

**CEL Expressions**:
```yaml
# Inner loop - more help needed
- if: "nodes.process_next_task.output.toLowerCase().contains('help')"
  next: process_next_task

# Outer loop - next item
- if: "!nodes.check_completion.output.contains('COMPLETE')"
  next: process_next_task

# Exit condition
- if: "nodes.check_completion.output.contains('COMPLETE')"
  next: summary
```

### 3. Retry with Counter (`retry_loop.yaml`)

**Pattern**: Bounded retry loop with maximum attempts

**Use Cases**:
- API retry logic
- Error recovery
- Rate-limited operations

**Key Features**:
- Tracks retry count using `vars.retry_count`
- Maximum 3 attempts
- Different paths for success, retry, and max retries exceeded

**Loop Structure**:
```
start → call_api → check_result
           ↑            ↓
           │      ┌─────┴──────┬──────────┬─────────┐
           │      ↓            ↓          ↓         ↓
           └── retry_handler  success  max_retry  error
                              handler   handler   handler
```

**CEL Expressions**:
```yaml
# Success - exit loop
- if: "nodes.call_api.output.contains('SUCCESS')"
  next: success_handler

# Retry - continue loop
- if: "nodes.call_api.output.contains('FAILURE') && (vars.retry_count == null || vars.retry_count < 2)"
  next: retry_handler  # Which loops back to call_api

# Max retries - exit loop
- if: "nodes.call_api.output.contains('FAILURE')"
  next: max_retry_handler
```

## Common Loop Techniques

### 1. Exit Condition in LLM Output

The LLM can signal loop exit by including a specific marker in its response:

```yaml
- id: check_status
  type: generate
  system_prompt: |
    If the task is complete, respond with exactly: DONE
    Otherwise, provide next steps.

- id: loop_controller
  type: selector
  cases:
    - if: "nodes.check_status.output.contains('DONE')"
      next: finish
    - if: "true"
      next: continue_task  # Loop back
```

### 2. Counter-Based Loops

Track iterations using variables:

```yaml
- id: increment_counter
  type: generate
  system_prompt: |
    Current count: ${ vars.counter != null ? vars.counter + 1 : 1 }

- id: check_counter
  type: selector
  cases:
    - if: "vars.counter != null && vars.counter >= 5"
      next: max_iterations_reached
    - if: "true"
      next: increment_counter  # Loop back
```

### 3. Conditional String Matching

Check user input for specific keywords:

```yaml
- id: route_decision
  type: selector
  cases:
    - if: "nodes.user_response.output.toLowerCase().contains('yes')"
      next: continue_loop
    - if: "nodes.user_response.output.toLowerCase().contains('no')"
      next: exit_loop
```

### 4. Multi-Level Loops

Nested loops for complex scenarios:

```yaml
# Outer loop: process each category
- id: category_router
  type: selector
  cases:
    - if: "vars.current_category < vars.total_categories"
      next: process_category
    - if: "true"
      next: all_done

# Inner loop: process items in category
- id: item_router
  type: selector
  cases:
    - if: "vars.current_item < vars.items_in_category"
      next: process_item
    - if: "true"
      next: category_complete  # Exits inner loop, continues outer loop
```

## Best Practices

### 1. Always Have an Exit Condition

❌ **Bad** - Infinite loop:
```yaml
- id: loop_node
  type: selector
  cases:
    - if: "true"
      next: loop_node  # No way to exit!
```

✅ **Good** - Bounded loop:
```yaml
- id: loop_node
  type: selector
  cases:
    - if: "vars.counter >= 10"
      next: exit_node  # Exit condition
    - if: "true"
      next: loop_node
```

### 2. Provide User Control

Allow users to exit interactive loops:

```yaml
- if: "nodes.user_input.output.toLowerCase().contains('exit')"
  next: graceful_exit
```

### 3. Prevent Infinite Loops

- Set maximum iteration counts
- Use timeouts (implementation-dependent)
- Provide multiple exit paths

### 4. Clear Loop State

Use descriptive variable names and clear conditions:

```yaml
# Clear intent
- if: "vars.items_processed >= vars.total_items"
  next: processing_complete

# vs unclear
- if: "vars.x >= vars.y"
  next: done
```

### 5. Handle Edge Cases

```yaml
cases:
  # Normal completion
  - if: "vars.tasks_remaining == 0"
    next: success

  # User cancellation
  - if: "nodes.last_response.output.contains('cancel')"
    next: cancelled

  # Error state
  - if: "vars.error_count > 3"
    next: error_handler

  # Max iterations
  - if: "vars.iterations >= 100"
    next: timeout

  # Continue loop (default)
  - if: "true"
    next: process_next
```

## Limitations

1. **No built-in loop constructs**: Loops are implemented via conditional jumps
2. **State management**: Variables (`vars.*`) are used to track state, but increment operations are not atomic
3. **No automatic iteration**: You must manually manage counters and conditions
4. **Stack depth**: Very deep loops may hit recursion limits (implementation-dependent)

## Future Enhancements

These patterns could be simplified with dedicated loop constructs:

```yaml
# Hypothetical future syntax
- id: process_items
  type: foreach
  items: "${ inputs.task_list }"
  iterator: task
  body:
    - id: process_task
      type: tool
      tool_name: process
      parameters:
        item: "${ task }"
```

## Testing Loop Examples

To test these examples:

```bash
# Quiz loop
go run cmd/dify-run/main.go -input "Start quiz" examples/quiz_loop.yaml

# Iterative processor
go run cmd/dify-run/main.go -input "Write email, Review doc, Call client" examples/iterative_processor.yaml

# Retry loop
go run cmd/dify-run/main.go -input "Process payment" examples/retry_loop.yaml
```

Remember to set up your OpenAI API key:
```bash
export OPENAI_API_KEY=your-key-here
```

Or use mock provider for testing:
```bash
export DIFY_LLM_PROVIDER=mock
export DIFY_MOCK_RESPONSE="Test response"
```

## References

- [CEL Reference](../docs/CEL_REFERENCE.md) - Expression syntax
- [Selector Node Specification](../docs/planning/dsl-schema.json) - Schema definition
- [Expression Language Design](../docs/EXPRESSION_LANGUAGE_DESIGN.md) - Architecture
