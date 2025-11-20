# Implementation Plan

## Phase 1: Scaffolding & DSL Loader (Completed)
- [x] Define project structure.
- [x] Implement DSL types (`dsl/types.go`).
- [x] Implement YAML loader and validation (`dsl/loader.go`).
- [x] Create basic tests.

## Phase 2: Basic Linear Runtime & Memory (Completed)
- [x] Implement `Engine` struct.
- [x] Implement `NodeExecutor` interface.
- [x] Implement `GenerateExecutor` (Basic LLM call).
- [x] Implement `GlobalMemory` (Thread-safe storage).
- [x] Implement linear workflow execution (A -> B -> C).

## Phase 3: LLM Integration & Tools (Completed)
- [x] Refine `LLM Provider` interface (Streaming).
- [x] Implement `OpenAIProvider` (with `openai-go` v3).
- [x] Implement `ToolNode` and `ToolExecutor`.
- [x] Implement `ToolRegistry`.
- [x] Support user input in CLI.

## Phase 4: Parallelism, Scopes & Expressions (CEL)
**Goal**: Enable parallel execution, variable scopes, and advanced expression evaluation using CEL.

- [ ] **CEL Integration**
    - [ ] Add `github.com/google/cel-go` dependency.
    - [ ] Implement `ExpressionEvaluator` struct.
        - [ ] Define CEL Environment with `inputs`, `nodes`, `vars` scopes.
        - [ ] Implement `EvaluateBool(expr, context)` for conditions.
        - [ ] Implement `EvaluateTemplate(template, context)` for string interpolation (`${ expr }`).
    - [ ] Integrate `ExpressionEvaluator` into `Engine`.
- [ ] **Parallel Node**
    - [ ] Define `ParallelNode` in DSL.
    - [ ] Implement `ParallelExecutor` using goroutines.
    - [ ] Handle synchronization (WaitAll).
- [ ] **Variable Scoping**
    - [ ] Refactor `Memory` to support hierarchical scopes (Global -> Parallel Branch).
    - [ ] Ensure thread-safe variable access.
- [ ] **Update Existing Nodes**
    - [ ] Update `GenerateNode` to support CEL templates in prompts.
    - [ ] Update `ToolNode` to support CEL in parameters.

## Phase 5: Branching & Control
**Goal**: Implement complex control flow.

- [ ] **Selector Node (Branching)**
    - [ ] Define `SelectorNode` in DSL.
    - [ ] Implement `SelectorExecutor` using CEL for conditions.
- [ ] **Iteration (Loop)** (Optional/Future)
    - [ ] Define `LoopNode`.

## Phase 6: Observability & Polish
- [ ] Structured Logging / Tracing.
- [ ] Metrics.
- [ ] CLI improvements.
