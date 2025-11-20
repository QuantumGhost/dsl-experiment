# dsl-exp - Design Proposal (Message-Centric)

## 1. Core Philosophy: "Conversation is State"

To simplify the system while maintaining expressiveness, we remove the concept of "Variables". The entire state of the application is the **Conversation History (Memory)**.

- **Input**: A list of messages.
- **Output**: A list of messages.
- **Execution**: A pipeline of nodes that append messages to the history.


## 2. Memory Model

### 2.1 Global Memory (The Timeline)
- A linear sequence of `[System, User, Assistant, Tool]` messages.
- All nodes read from this timeline.
- Nodes append their results (Assistant replies, Tool outputs) to this timeline.

### 2.2 Parallelism & Shared Memory
- **Requirement**: "Parallel nodes share memory".
- **Implementation**:
    - Parallel nodes start with the **same snapshot** of the Global Memory.
    - They execute independently (Context Memory).
    - **Join Strategy (The Merge Problem)**:
        - *Challenge*: When Branch A and Branch B both produce messages, simply interleaving them might confuse the LLM (e.g., two different tool calls appearing out of context).
        - *Solution 1: Sequential Append (Default)*: Append Branch A's messages, then Branch B's messages. Simple but might create long contexts.
        - *Solution 2: Block/Group*: Wrap parallel outputs in a `<parallel_results>` block to help the LLM understand they happened simultaneously.
        - *Solution 3: Reduce/Synthesize*: Require a `reducer` step after the parallel block to summarize the branches into a single message before appending to the Global Memory.
        - *Decision*: For this minimal runtime, we will use **Sequential Append** (in definition order) for simplicity, but allow the `parallel` node to specify a `reducer` (an LLM node) if needed.

### 2.4 Scopes & Memory Policy (Context Management)
To solve the "Limited Context vs Infinite Information" problem, we introduce **Scopes**.

- **Concept**:
    - Every `parallel` block or `group` node creates a **Local Scope**.
    - Steps inside the scope read from (Global + Local) memory, but write to **Local Memory**.

- **Memory Policy (Export Strategy)**:
    - When a Scope ends, how do we merge back to Global?
    - **`result_only` (Default)**: Only the *last* message (the result) is appended to Global Memory. The intermediate steps (CoT, Tool Calls) are discarded (forgotten).
    - **`full`**: All messages happened in the scope are appended to Global Memory.
    - **`summary`**: (Advanced) Trigger an automatic summarization to compress the scope into one message.

- **UI Visibility**:
    - Distinct from Memory. A step can be `visible: true` (stream to user) but `memory: false` (ephemeral).
    - By default, `result_only` policy implies that intermediate steps might be visible to user (streaming) but forgotten by the Agent.

### 2.5 Context Control (Memory Input)
Sometimes we don't want to pass the entire Global Memory to an LLM (e.g., to save tokens or avoid distraction).
- **`context_mode`**:
    - `full` (Default): Pass the entire visible history.
    - `none`: Pass ONLY the System Prompt and the current User Prompt (if any). Useful for summarization or isolated tasks.
    - `window`: Pass the last N messages.

### 2.6 Streaming & Events
The Runtime is designed to be **Event-Driven** to support real-time streaming.
- **Architecture**:
    - `Executor` returns a `Stream` (Go Channel) instead of a static result.
    - Events: `NodeStart`, `NodeEnd`, `TextChunk`, `ToolCall`, `Error`.
- **Visibility Integration**:
    - If `visibility: public`, `TextChunk` events are pushed to the client.
    - If `visibility: silent`, `TextChunk` events are consumed internally but not pushed.
    - If `visibility: ephemeral`, `TextChunk` events are pushed, but the final message is NOT added to Memory.

## 3. Node Types

### 3.1 `generate` (LLM)
- **Role**: The brain.
- **Action**: Sends current History to LLM.
- **Output**: Appends `AssistantMessage` (Text or Tool Call).
- **Configuration**: `model`, `system_prompt` (optional override).

### 3.2 `tool` (Executor)
- **Role**: The hands.
- **Action**: If the last message is a `ToolCall`, execute it.
- **Output**: Appends `ToolMessage` (Result).
- **Auto-Loop**: Often combined with `generate` to form a loop (Think -> Call -> Result -> Think).

### 3.3 `parallel` (Flow)
- **Role**: Multi-tasking.
- **Action**: Executes a list of branches concurrently.
- **Memory**: Each branch forks the history.
- **Merge**: Collects all new messages from branches and appends them to the main history.

### 3.4 `selector` (Branching)
- **Role**: Decision making.
- **Action**: Uses LLM (or rule) to classify the current history.
- **Output**: Selects the *ID* of the next node to execute.
- **Note**: This replaces "If/Else" with a natural "Router".

## 4. DSL Syntax Example (YAML)

```yaml
name: "Travel Agent"

# The workflow is a list of nodes. 
# Default flow is sequential.
nodes:
  - id: "ask_preference"
    type: "generate"
    model: "gpt-4"
    system_prompt: "You are a travel assistant. Ask the user where they want to go."

  - id: "parallel_research"
    type: "parallel"
    branches:
      - nodes: # Branch A: Search Flights
          - id: "search_flights"
            type: "generate"
            system_prompt: "Based on the user's input, generate a tool call to search flights."
          - id: "exec_flights"
            type: "tool" 
      
      - nodes: # Branch B: Search Hotels
          - id: "search_hotels"
            type: "generate"
            system_prompt: "Based on the user's input, generate a tool call to search hotels."
          - id: "exec_hotels"
            type: "tool"

  - id: "synthesize"
    type: "generate"
    system_prompt: "Review the flight and hotel options above and give a recommendation."
```

## 5. Critical Analysis (Self-Correction)

### How to handle Loops?
- **Implicit**: The `tool` node can automatically trigger a re-run of the previous `generate` node if the LLM wants to continue (ReAct loop).
- **Explicit**: A `loop` node or `goto` mechanism.
- **Proposal**: Use a `router` or `selector` for explicit loops.
    - Example:
      ```yaml
      - id: "check_finished"
        type: "selector"
        options:
          - case: "User is satisfied"
            next: "end"
          - case: "Need more info"
            next: "ask_preference" # GOTO loop
      ```

### How to handle "Map" (Loop over data) without variables?
- If we strictly have no variables, we can't iterate over a JSON array easily.
- **Compromise**: The LLM is the iterator.
    - We ask the LLM: "Generate 3 parallel tasks".
    - The LLM outputs 3 tool calls.
    - The Runtime sees 3 tool calls and runs them (possibly in parallel).
    - This fits the "Agentic" philosophy better than a code-like `foreach`.

## 6. Runtime Architecture
- **State**: `[]Message`
- **Runner**:
    - `pointer`: Current Node ID.
    - `stack`: For nested blocks (Parallel/Sequence).
- **Loop**:
    1. Get Current Node.
    2. Execute(Node, History) -> (NewMessages, NextNodeID).
    3. Append NewMessages to History.
    4. Move to NextNodeID.
