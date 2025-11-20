# Loop Testing Summary

## Status: ✅ Loops Working with Safety Limits

### Tests Performed

#### 1. ✅ Selector Functionality Test
**File**: `examples/selector_test.yaml`

**Test**: CEL-based conditional routing without loops

**Result**: PASS
```
✓ Selector evaluated: nodes.start.output.contains('complete')
✓ Correctly routed to 'success' node
✓ Workflow completed successfully
```

#### 2. ✅ Infinite Loop Detection
**File**: `examples/infinite_loop_test.yaml`

**Test**: Intentional infinite loop (always routes back to start)

**Result**: PASS - Safety limit triggered
```
✓ Engine executed 100 iterations
✓ Detected infinite loop
✓ Error: "maximum iteration limit (100) exceeded - possible infinite loop"
✓ Gracefully terminated
```

## Implementation Details

### SelectorExecutor
**Location**: `internal/node/selector.go`

**Functionality**:
- Evaluates CEL expressions in `cases[].if` fields sequentially
- Routes to first matching case's `next` node
- Supports complex CEL expressions for routing logic

**Example**:
```yaml
- id: router
  type: selector
  cases:
    - if: "nodes.classifier.output.contains('urgent')"
      next: priority_handler
    - if: "inputs.score >= 90"
      next: excellent
    - if: "true"  # Default case
      next: normal
```

### Node Output Storage
**Location**: `internal/node/generate.go`

**Enhancement**: Generate nodes now store outputs in memory:
```go
mem.SetNodeOutput(node.GetID(), map[string]interface{}{
    "output": responseText,
})
```

**Access Pattern**:
```cel
nodes.<node_id>.output  // Access full text output
```

### Safety: Maximum Iteration Limit
**Location**: `internal/engine/engine.go`

**Protection**: 100 iteration limit to prevent runaway loops

**Behavior**:
- Counts each node execution
- Terminates workflow if limit exceeded
- Emits descriptive error event
- Prevents infinite loops from hanging system

## Loop Examples Status

### ✅ Working Examples

1. **`selector_test.yaml`** - Non-loop conditional routing
2. **`infinite_loop_test.yaml`** - Demonstrates safety limit

### ⚠️ Complex Loop Examples (Needs Variable Support)

The following examples are **syntactically valid** but have **functional limitations**:

1. **`quiz_loop.yaml`** 
   - ✅ Loads and parses correctly
   - ⚠️ Loop continuation depends on LLM output detection
   - ⚠️ No iteration counter without `vars.*` support

2. **`iterative_processor.yaml`**
   - ✅ Loads and parses correctly  
   - ⚠️ Multiple nested loops
   - ⚠️ Requires state management (not yet implemented)

3. **`retry_loop.yaml`**
   - ✅ Loads and parses correctly
   - ❌ Requires `vars.retry_count` (not implemented)
   - ❌ Cannot increment counters in CEL (read-only)

## Identified Limitations

### 1. No Variable Mutation
**Problem**: CEL expressions are read-only
```yaml
# This DOESN'T work - CEL cannot mutate values
- if: "vars.counter != null ? vars.counter + 1 : 1"
```

**Impact**: Cannot implement counters for bounded loops

**Workaround**: Loop conditions must rely on:
- Node outputs (LLM responses)
- Input data
- Fixed thresholds

### 2. No Side Effects in CEL
**Problem**: CEL expressions cannot set variables

**What's Needed**: A dedicated node type for variable assignment:
```yaml
# Proposed future feature
- id: increment_counter
  type: "set_variable"
  variable: "counter"
  value: "${ vars.counter + 1 }"
```

### 3. LLM-Based Exit Conditions
**Current Approach**: Loops depend on LLM output content

**Example**:
```yaml
- if: "nodes.ask_user.output.contains('EXIT')"
  next: goodbye
```

**Risk**: Unreliable - LLM may not output exact keyword

**Better Approach**: Explicit user input classification (not yet implemented)

## Recommendations

### Use Cases That Work NOW

✅ **Conditional Branching** (no loops):
```yaml
- type: selector
  cases:
    - if: "inputs.priority == 'high'"
      next: fast_track
    - if: "true"
      next: normal_processing
```

✅ **Fixed Iteration with Safety Limit**:
```yaml
# Will run until safety limit (100) or exit condition
- if: "nodes.check.output.contains('done')"
  next: finish
- if: "true"
  next: process_next  # Loop back
```

### Use Cases That Need More Features

❌ **Bounded Loops with Exact Count**:
```yaml
# Needs variable mutation
for i in range(1, 10):
    process_item(i)
```

❌ **Stateful Iteration**:
```yaml
# Needs persistent variables across iterations
while items_remaining > 0:
    process_one()
    items_remaining -= 1
```

## Next Steps to Full Loop Support

1. **Implement `SetVariable` Node Type**
   - Allow explicit variable assignment
   - Enable counter increments
   - Support accumulator patterns

2. **Add `foreach` Node Type** (future)
   - Iterate over arrays
   - Built-in iteration variable
   - Automatic counter management

3. **Enhanced Safety**
   - Configurable iteration limits
   - Per-workflow limits
   - Timeout mechanisms
   - Circuit breakers

4. **Better Exit Detection**
   - Structured responses from LLM
   - Input parsing for commands
   - Explicit intent classification

## Testing Checklist

- [x] SelectorExecutor registered in engine
- [x] CEL evaluation working for conditions
- [x] Node outputs stored in memory
- [x] `nodes.*` context accessible in CEL
- [x] Infinite loop detection working
- [x] Safety limit (100 iterations) tested
- [x] Graceful error handling
- [ ] Variable mutation (not implemented)
- [ ] Counter-based loops (blocked by above)
- [ ] Nested loop stress testing (blocked by above)

## Conclusion

**Loops are WORKING with important caveats**:

✅ **What Works**:
- Conditional routing with CEL
- Loop-back navigation
- Infinite loop protection
- LLM output-based exit conditions

⚠️ **What's Limited**:
- No loop counters (requires variable mutation)
- Relies on LLM consistency for exit conditions
- Cannot implement exact iteration counts

❌ **What Doesn't Work**:
- Counter-based loops (`for i = 1 to N`)
- Stateful iteration with accumulators
- Precise iteration control

**Recommendation**: Current implementation is **production-ready for conditional workflows** but **NOT recommended for precise loop control** until variable mutation is implemented.
