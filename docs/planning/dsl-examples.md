# dsl-exp - Comprehensive Examples

This document provides detailed examples of the dsl-exp syntax (YAML). It covers all proposed node types and common patterns.

## 1. Basic Linear Chat
A simple conversation flow where the system prompts the user, and the LLM responds.

```yaml
name: "Simple Chat"
description: "A basic linear conversation."

nodes:
  - id: "system_intro"
    type: "generate"
    model: "gpt-4"
    system_prompt: "You are a helpful assistant. Say hello to the user."
    # Implicitly waits for user input if it's the first node or after a stop? 
    # For this DSL, we assume the 'input' messages are already in the history when execution starts.
```

## 2. Tool Use (ReAct Pattern)
Demonstrates how an LLM can generate a tool call, which is then executed by a `tool` node.

```yaml
name: "Weather Bot"
description: "Checks weather using a tool."

nodes:
  - id: "think_and_call"
    type: "generate"
    model: "gpt-4"
    system_prompt: |
      You are a weather assistant. 
      If the user asks for weather, call the `get_weather` tool.
    # LLM Output: ToolCall(name="get_weather", args={"city": "Beijing"})

  - id: "execute_tool"
    type: "tool"
    # This node automatically looks at the last message. 
    # If it's a ToolCall, it executes it and appends the ToolResult.
    # If not, it does nothing (no-op).

  - id: "final_response"
    type: "generate"
    model: "gpt-4"
    system_prompt: "Summarize the tool result for the user."
```

## 3. Parallel Execution (Shared Memory)
Demonstrates running two independent tasks concurrently. The results are merged sequentially into the history.

```yaml
name: "Parallel Research"
description: "Research two topics at once."

nodes:
  - id: "start"
    type: "generate"
    model: "gpt-3.5-turbo"
    system_prompt: "Plan the research tasks."

  - id: "do_research"
    type: "parallel"
    branches:
      - nodes: # Branch 1: History
        - id: "research_history"
          type: "generate"
          model: "gpt-3.5-turbo"
          system_prompt: "Write a short paragraph about the history of AI."
      
      - nodes: # Branch 2: Future
        - id: "research_future"
          type: "generate"
          model: "gpt-3.5-turbo"
          system_prompt: "Write a short paragraph about the future of AI."

  # After 'do_research', both paragraphs are appended to the history.
  
  - id: "synthesize"
    type: "generate"
    model: "gpt-4"
    system_prompt: "Combine the history and future paragraphs into a coherent article."
```

## 4. Branching (Selector / Router)
Demonstrates conditional logic using an LLM to decide the path.

```yaml
name: "Customer Support Router"
description: "Route user to technical support or sales."

nodes:
  - id: "classify_intent"
    type: "selector" # Special node that returns a Node ID
    model: "gpt-3.5-turbo"
    system_prompt: |
      Classify the user's intent.
      - If they have a bug, choose 'tech_support'.
      - If they want to buy, choose 'sales'.
      - Otherwise, choose 'general_chat'.
    options:
      - case: "tech_support"
        next: "tech_support_agent"
      - case: "sales"
        next: "sales_agent"
      - case: "general_chat"
        next: "default_reply"

  - id: "tech_support_agent"
    type: "generate"
    model: "gpt-4"
    system_prompt: "You are Tech Support. Ask for error logs."
    next: "end" # Explicit end

  - id: "sales_agent"
    type: "generate"
    model: "gpt-4"
    system_prompt: "You are Sales. Ask for their budget."
    next: "end"

  - id: "default_reply"
    type: "generate"
    model: "gpt-3.5-turbo"
    system_prompt: "I can only help with Tech Support or Sales."
    next: "end"
    
  - id: "end"
    type: "noop" # Marker for end of flow
```

## 5. Complex Agent (Loop + Tools)
A simplified "AutoGPT" style loop.

```yaml
name: "Autonomous Agent"

nodes:
  - id: "entry"
    type: "noop"

  - id: "think"
    type: "generate"
    model: "gpt-4"
    system_prompt: |
      Goal: Solve the user's request.
      1. Decide if you need to use a tool.
      2. If yes, output a tool call.
      3. If you are done, output "DONE".

  - id: "check_done"
    type: "selector"
    model: "gpt-3.5-turbo" # A cheaper model or rule-based check
    system_prompt: "Did the assistant say DONE?"
    options:
      - case: "yes"
        next: "finish"
      - case: "no"
        next: "act"

  - id: "act"
    type: "tool"
    # Executes the tool call from 'think'
    next: "think" # Loop back to think

  - id: "finish"
    type: "generate"
    system_prompt: "Give a final answer."
```
