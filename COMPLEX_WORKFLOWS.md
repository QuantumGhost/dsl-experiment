# Complex Workflow Testing Guide

## Overview

This document describes the complex test workflows and how to run them.

## Test Workflows

### 1. Multi-Step Conversation (`multi_step_conversation.yaml`)

**Purpose**: Tests sequential execution with multiple generate nodes

**Nodes**: 4 sequential generate nodes
- `greeting` → `understand_intent` → `provide_response` → `followup`

**Test**:
```bash
DSL_EXP_LLM_PROVIDER=mock DSL_EXP_MOCK_RESPONSE="Test" ./dsl-run examples/multi_step_conversation.yaml
```

**Expected Behavior**:
- All 4 nodes execute in order
- Each node streams mock response
- Memory accumulates with each step

### 2. Conditional Routing (`conditional_routing.yaml`)

**Purpose**: Tests selector node and conditional branching

**Nodes**:
- 1 selector node (`classify_intent`)
- 3 conditional branches (`answer_question`, `handle_task`, `casual_chat`)
- 1 final node (`closing`)

**Structure**:
```
classify_intent
├─→ answer_question → closing
├─→ handle_task → closing
└─→ casual_chat → closing
```

**Test**:
```bash
go test ./internal/integration/... -run TestComplexWorkflow_ConditionalRouting -v
```

**Expected Behavior**:
- Selector chooses one of three paths
- Only selected path executes
- Final `closing` node always executes

### 3. Parallel Research (`parallel_research.yaml`)

**Purpose**: Tests parallel execution with reducer

**Nodes**:
- Initial node (`initial_query`)
- Parallel block with 3 branches
  - `research_technical`
  - `research_business`  
  - `research_social`
- Reducer node (`synthesize`)
- Final node (`final_summary`)

**Test**:
```bash
go test ./internal/integration/... -run TestComplexWorkflow_ParallelExecution -v
```

**Expected Behavior**:
- All 3 branches execute concurrently
- Reducer synthesizes results
- Memory contains all branch outputs

### 4. Nested Workflow (`nested_workflow.yaml`)

**Purpose**: Tests nested structures (group + parallel)

**Structure**:
```
start
└─→ analysis_group (group)
    ├─→ prepare_analysis
    └─→ parallel_analysis (parallel)
        ├─→ analyze_pros
        └─→ analyze_cons
        └─→ compare (reducer)
└─→ conclusion
```

**Test**:
```bash
go test ./internal/integration/... -run TestComplexWorkflow_NestedStructure -v
```

**Expected Behavior**:
- Group creates local scope
- Nested parallel executes within group
- Memory policy applies correctly

## Running Tests

### All Complex Workflow Tests

```bash
go test ./internal/integration/... -run TestComplexWorkflow -v
```

### Specific Test

```bash
go test ./internal/integration/... -run TestComplexWorkflow_MultiStep -v
```

### With Coverage

```bash
go test ./internal/integration/... -coverprofile=integration_coverage.out
go tool cover -html=integration_coverage.out
```

## Mock Provider Usage

### Environment Variable Method

```bash
export DSL_EXP_LLM_PROVIDER=mock
export DSL_EXP_MOCK_RESPONSE="Your custom response"
./dsl-run examples/multi_step_conversation.yaml
```

### Programmatic Method

```go
provider := llm.NewMockProvider("Custom response")
eng := engine.NewEngine()
eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{
    Provider: provider,
})
```

## Test Coverage

Complex workflow tests verify:

✅ **Sequential Execution**
- Node ordering
- Memory accumulation
- Event ordering

✅ **Conditional Branching**
- Selector logic
- Path selection
- Conditional execution

✅ **Parallel Execution**
- Concurrent branch execution
- Result merging
- Reducer functionality

✅ **Nested Structures**
- Group scoping
- Nested parallel blocks
- Memory policy application

✅ **Workflow Validation**
- YAML parsing
- Type checking
- ID resolution
- Next pointer validation

## Integration Test Structure

```go
func TestComplexWorkflow_X(t *testing.T) {
    // 1. Load workflow from YAML
    wf, err := dsl.Load("path/to/workflow.yaml")
    
    // 2. Create mock provider
    mockProvider := &MockProvider{...}
    
    // 3. Setup engine
    eng := engine.NewEngine()
    eng.RegisterExecutor(dsl.NodeGenerate, &node.GenerateExecutor{
        Provider: mockProvider,
    })
    
    // 4. Run workflow
    eventChan := eng.Run(ctx, wf, input)
    
    // 5. Verify behavior
    for event := range eventChan {
        // Assert expected events
    }
}
```

## Debugging

### Enable Verbose Output

```bash
go test ./internal/integration/... -v
```

### Check Node Execution Order

Events are emitted in this pattern:
```
NodeStart (node1) → TokenGenerated → NodeEnd (node1) → NodeStart (node2) → ...
```

### Inspect Memory

Add logging in tests to see memory state:
```go
snapshot := mem.Snapshot()
for i, msg := range snapshot {
    t.Logf("Message %d: Role=%s, Content=%s", i, msg.Role, msg.Content)
}
```

## Common Issues

### Issue: Nodes execute out of order
**Cause**: Missing or incorrect `next` pointers
**Fix**: Verify YAML workflow structure and next pointers

### Issue: Parallel branches don't execute
**Cause**: Missing executor registration for NodeParallel
**Fix**: Register all node executors before running

### Issue: Mock responses not appearing
**Cause**: Provider not properly injected
**Fix**: Ensure mock provider is registered with GenerateExecutor

## Performance

Complex workflows have been tested with:
- ✅ Up to 10+ sequential nodes
- ✅ Up to 5 parallel branches
- ✅ 3 levels of nesting
- ✅ 100+ goroutines (concurrency test)

All tests pass with race detection enabled:
```bash
go test -race ./internal/integration/...
```
