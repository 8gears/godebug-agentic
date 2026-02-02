---
name: godebug-verify
description: Proactive runtime verification for AI-generated Go code. Use when you need to CONFIRM code correctness before marking tasks done - closing the loop from generation to runtime proof. Complements godebug (reactive debugging) with verification mindset.
---

# godebug-verify - Runtime Verification for AI Agents

Proactive verification skill that teaches AI agents to **confirm their own code works** at runtime. This is "closing the loop" - never mark code done until you've verified it executes correctly.

> "How to be effective with coding agent is always like you have to close the loop. It needs to be able to debug and test itself."

## 1. Verification vs Debugging

### When to Use Which Skill

| Situation | Use | Skill |
|-----------|-----|-------|
| Code you just wrote | Verify it works | **godebug-verify** |
| Existing bug report | Find the cause | godebug |
| "Is this correct?" | Confirm behavior | **godebug-verify** |
| "What's broken?" | Investigate failure | godebug |
| Before marking done | Runtime proof | **godebug-verify** |
| After test failure | Debug the failure | godebug |
| Concurrency concerns | Verify synchronization | **godebug-verify** |
| Crash or panic | Find root cause | godebug |

### The Mindset Difference

**Debugging (godebug):**
- Reactive: Something broke, find why
- Question: "What went wrong?"
- Goal: Fix the bug

**Verification (godebug-verify):**
- Proactive: Code looks right, prove it runs right
- Question: "Does this actually work?"
- Goal: Confirm before shipping

### The "Closing the Loop" Principle

```
[Write Code] → [Looks Correct] → [VERIFY AT RUNTIME] → [Mark Done]
                      ↓                   ↓
              Static analysis      Actually executes?
              Code review          Values are correct?
              "Seems right"        Timing is right?
```

**Never mark code as complete based only on:**
- "The logic looks correct"
- "Tests pass" (tests might not cover this path)
- "It compiles"

**Always verify when:**
- You wrote concurrent code (goroutines, channels, sync primitives)
- You changed control flow logic
- You modified data structures
- The fix was non-obvious

## 2. The Verification Protocol

### HAEC: Hypothesis → Assertion → Execution → Conclusion

Every verification follows this four-step protocol:

```
HYPOTHESIS  → State what MUST be true if code is correct
ASSERTION   → Define observable evidence (breakpoints, values)
EXECUTION   → Run to verification points
CONCLUSION  → CONFIRMED or REFUTED based on evidence
```

### Protocol in Practice

**Note on file paths:** Breakpoint file paths can be specified as:
- **Short filename** (e.g., `main.go:42`) - works if the file is unique in the debug session
- **Relative path** (e.g., `pkg/myapp/main.go:42`) - use when multiple files share the same name
- **Absolute path** - always unambiguous

If a breakpoint fails with "could not find file", run `godebug sources` to see how files are listed, then use that path.

```bash
# HYPOTHESIS: wg.Add() is called BEFORE goroutine starts

# ASSERTION: Set breakpoints to observe call order
ADDR=localhost:4445
godebug --addr $ADDR break pkg/myapp/main.go:42    # wg.Add(1) line
godebug --addr $ADDR break pkg/myapp/main.go:44    # go func() line

# EXECUTION: Run and observe which hits first
godebug --addr $ADDR continue
# Check: which breakpoint ID appears in response?

# CONCLUSION:
# - Breakpoint at :42 first → CONFIRMED (Add before goroutine)
# - Breakpoint at :44 first → REFUTED (race condition!)
```

### Conclusion States

| State | Meaning | Action |
|-------|---------|--------|
| **CONFIRMED** | Evidence matches hypothesis | Proceed with confidence |
| **REFUTED** | Evidence contradicts hypothesis | Diagnose → Fix → Re-verify |
| **INCONCLUSIVE** | Can't determine from evidence | Adjust breakpoints, re-run |

## 3. Verification Scenarios

### 3.1 Function Return Values

**Hypothesis:** Function returns expected value for given input.

```bash
# Verify: calculateTotal() returns correct sum

# ASSERTION: Break at return, check value
godebug --addr $ADDR break "main.calculateTotal"
godebug --addr $ADDR continue

# Step to return statement
godebug --addr $ADDR next
godebug --addr $ADDR next

# EXECUTION: Check return value
godebug --addr $ADDR eval "total"
# Expected: {"value": "150"}

# CONCLUSION: value matches expected → CONFIRMED
```

### 3.2 WaitGroup Synchronization

**Hypothesis:** All wg.Add() calls complete before wg.Wait() returns.

```bash
# Verify: WaitGroup pattern in testdata/concurrency_bugs/waitgroup_race

# ASSERTION: Track Add/Done/Wait sequence
godebug --addr $ADDR break "sync.(*WaitGroup).Add"
godebug --addr $ADDR break "sync.(*WaitGroup).Wait"

# EXECUTION: Run and observe call order
godebug --addr $ADDR continue
# Response shows which breakpoint hit

# If Add() breakpoint hit:
godebug --addr $ADDR stack  # Where was Add() called from?
godebug --addr $ADDR continue  # Next call

# CONCLUSION:
# Pattern: Add, Add, Add..., Wait → CONFIRMED
# Pattern: Wait hits with Add still pending → REFUTED
```

**Real Example (buggy code):**
```go
// testdata/concurrency_bugs/waitgroup_race/main.go:16-27
for i := 0; i < 10; i++ {
    go func(id int) {
        wg.Add(1)  // BUG: Add() inside goroutine
        defer wg.Done()
        // work...
    }(i)
}
wg.Wait()  // May return before all Add() calls!
```

```bash
# Verification reveals the race:
godebug --addr $ADDR break main.go:18    # wg.Add(1)
godebug --addr $ADDR break main.go:27    # wg.Wait()
godebug --addr $ADDR continue

# If Wait breakpoint hits first → REFUTED
# The race condition is proven
```

### 3.3 Mutex Protection

**Hypothesis:** Shared variable is always accessed while holding mutex.

```bash
# Verify: counter is protected by mutex

# ASSERTION: Break on Lock and variable access
godebug --addr $ADDR break "sync.(*Mutex).Lock"
godebug --addr $ADDR break main.go:25  # counter++ line

# EXECUTION: Check if Lock always precedes access
godebug --addr $ADDR continue

# Check goroutine state when at counter++
godebug --addr $ADDR stack
# Look for Mutex.Lock in the stack

# CONCLUSION:
# Lock in stack → CONFIRMED (protected)
# No Lock in stack → REFUTED (unprotected access!)
```

**Real Example (buggy code):**
```go
// testdata/concurrency_bugs/race_counter/main.go:14-17
for i := 0; i < 1000; i++ {
    go func() {
        counter++  // Unprotected!
    }()
}
```

### 3.4 Channel Communication

**Hypothesis:** Channel receive gets the expected value from sender.

```bash
# Verify: value sent on channel matches value received

# ASSERTION: Break on send and receive
godebug --addr $ADDR break "runtime.chansend1"
godebug --addr $ADDR break "runtime.chanrecv1"

# Or break on your specific lines:
godebug --addr $ADDR break main.go:30  # ch <- value
godebug --addr $ADDR break main.go:45  # result := <-ch

# EXECUTION: Check values at each point
godebug --addr $ADDR continue
godebug --addr $ADDR eval "value"  # At send
godebug --addr $ADDR continue
godebug --addr $ADDR eval "result"  # At receive

# CONCLUSION: values match → CONFIRMED
```

### 3.5 Deadlock Prevention

**Hypothesis:** Lock acquisition order is consistent (no circular wait).

```bash
# Verify: deadlock_circular pattern

# ASSERTION: Track lock acquisition order
godebug --addr $ADDR break "sync.(*Mutex).Lock"
godebug --addr $ADDR continue

# EXECUTION: Record which mutex and from which goroutine
godebug --addr $ADDR stack
godebug --addr $ADDR goroutines

# Build lock order graph:
# Goroutine 1: resourceA → resourceB
# Goroutine 2: resourceB → resourceA  ← CYCLE!

# CONCLUSION:
# Same order in all goroutines → CONFIRMED
# Different orders → REFUTED (deadlock risk)
```

**Real Example:**
```go
// testdata/concurrency_bugs/deadlock_circular/main.go
func transferAtoB() {
    resourceA.Lock()  // Acquires A first
    resourceB.Lock()  // Then B
}
func transferBtoA() {
    resourceB.Lock()  // Acquires B first
    resourceA.Lock()  // Then A - OPPOSITE ORDER!
}
```

### 3.6 Control Flow Verification

**Hypothesis:** Code takes the expected branch.

```bash
# Verify: error handling branch executes

# ASSERTION: Break at the branch point
godebug --addr $ADDR break main.go:50  # if err != nil

# EXECUTION: Check condition and which path
godebug --addr $ADDR continue
godebug --addr $ADDR eval "err"
godebug --addr $ADDR next  # Step to see which branch

# Check current location
godebug --addr $ADDR list

# CONCLUSION:
# At error handling code → CONFIRMED (error path taken)
# At normal path → REFUTED (error not handled)
```

## 4. The Self-Correction Loop

When verification is REFUTED, don't just fix blindly. Follow this loop:

```
REFUTED
   ↓
DIAGNOSE: Why did reality differ from expectation?
   ↓
FIX: Make minimal change to correct the issue
   ↓
RE-VERIFY: Run the SAME verification again
   ↓
CONFIRMED? → Done
   ↓ No
Back to DIAGNOSE
```

### Self-Correction Example

```bash
# Initial verification: WaitGroup synchronization
# RESULT: REFUTED - Wait() returned before all workers

# DIAGNOSE
godebug --addr $ADDR stack
# Shows: wg.Add(1) called from goroutine, not main

# FIX (in code)
# Move wg.Add(1) BEFORE the go statement:
#   wg.Add(1)
#   go func() { defer wg.Done(); ... }()

# RE-VERIFY: Same test
godebug --addr $ADDR restart

godebug --addr $ADDR break "sync.(*WaitGroup).Add"
godebug --addr $ADDR break "sync.(*WaitGroup).Wait"
godebug --addr $ADDR continue

# Check: Add() now hits from main goroutine
godebug --addr $ADDR stack
# Shows: main.main calls Add()

# CONCLUSION: CONFIRMED - Fix verified
```

### Never Mark Done Until CONFIRMED

```
[Code Written] → [REFUTED] → [Fix] → [Still REFUTED] → [Fix] → [CONFIRMED] → [Mark Done]
                    ↑                       ↑
                    └───────────────────────┘
                    Keep iterating until CONFIRMED
```

## 5. Token-Efficient Verification Patterns

Verification should be surgical. Don't dump all state; query exactly what you need.

### Efficient Queries

| Need | Inefficient | Efficient |
|------|-------------|-----------|
| One field | `locals` (~200 tokens) | `eval "user.ID"` (~50 tokens) |
| Slice length | `locals` | `eval "len(items)"` |
| Map key | `locals` | `eval "cache[\"key\"]"` |
| Condition | `locals` + mental eval | `eval "count > threshold"` |
| Pointer nil check | `locals` | `eval "ptr == nil"` |

### Targeted Verification Commands

```bash
# Instead of: dump everything
godebug --addr $ADDR locals
godebug --addr $ADDR args
godebug --addr $ADDR stack

# Do: surgical queries for your hypothesis
godebug --addr $ADDR eval "wg"           # Just the WaitGroup
godebug --addr $ADDR eval "len(results)" # Just the count
godebug --addr $ADDR eval "err != nil"   # Just the condition
```

### Minimal Breakpoint Strategy

```bash
# Instead of: breakpoints everywhere
godebug --addr $ADDR break main.go:10
godebug --addr $ADDR break main.go:15
godebug --addr $ADDR break main.go:20
godebug --addr $ADDR break main.go:25

# Do: breakpoint only at verification point
godebug --addr $ADDR break main.go:25  # The critical assertion point
```

## 6. Workflow Examples

### Example 1: Verify Concurrent Counter

You wrote code to increment a counter from multiple goroutines.

```bash
# Setup
dlv debug ./testdata/concurrency_bugs/race_counter --headless \
    --api-version=2 --listen=:4445 --accept-multiclient &
sleep 2
ADDR=localhost:4445

# HYPOTHESIS: counter equals 1000 after all goroutines complete

# ASSERTION: Check final counter value
godebug --addr $ADDR break main.go:22  # After time.Sleep

# EXECUTION
godebug --addr $ADDR continue
godebug --addr $ADDR eval "counter"

# CONCLUSION:
# counter == 1000 → CONFIRMED
# counter < 1000 → REFUTED (race condition, needs sync)
```

### Example 2: Verify Error Propagation

You wrote error handling code.

```bash
# HYPOTHESIS: Error from inner function propagates to caller

# ASSERTION: Track error through call chain
godebug --addr $ADDR break inner.go:15  # return fmt.Errorf(...)
godebug --addr $ADDR break caller.go:30  # if err != nil

# EXECUTION
godebug --addr $ADDR continue
godebug --addr $ADDR eval "err"  # At inner - should be non-nil
godebug --addr $ADDR continue
godebug --addr $ADDR eval "err"  # At caller - same error?

# CONCLUSION:
# Same error at both points → CONFIRMED
# nil at caller → REFUTED (error swallowed!)
```

### Example 3: Verify Mutex-Protected Critical Section

```bash
# HYPOTHESIS: Only one goroutine in critical section at a time

# Setup with concurrent workers
dlv debug ./concurrent-app --headless \
    --api-version=2 --listen=:4445 --accept-multiclient &
sleep 2
ADDR=localhost:4445

# ASSERTION: When in critical section, no other goroutine is there
godebug --addr $ADDR break main.go:50  # Inside critical section

# EXECUTION
godebug --addr $ADDR continue

# Check all goroutines
godebug --addr $ADDR goroutines

# For each goroutine, check location:
godebug --addr $ADDR goroutine 1
godebug --addr $ADDR list
godebug --addr $ADDR goroutine 2
godebug --addr $ADDR list

# CONCLUSION:
# Only one at main.go:50 → CONFIRMED
# Multiple at main.go:50 → REFUTED (mutex not working)
```

### Example 4: Verify Channel Doesn't Block Forever

```bash
# HYPOTHESIS: Channel receive completes (doesn't hang)

# ASSERTION: Break after receive
godebug --addr $ADDR break main.go:60  # result := <-ch
godebug --addr $ADDR break main.go:61  # Line after receive

# EXECUTION
godebug --addr $ADDR continue
# If we reach line 61, receive completed

# CONCLUSION:
# Hit breakpoint at :61 → CONFIRMED (didn't block)
# Program hangs at :60 → REFUTED (channel deadlock)
```

## 7. Integration with godebug Skill

This skill focuses on the **verification mindset**. For command syntax and detailed options, see the `godebug` skill.

### Command Reference (in godebug skill)

| Command | Use in Verification |
|---------|---------------------|
| `break` | Set assertion points |
| `continue` | Run to next assertion |
| `eval` | Check specific values |
| `goroutines` | Verify concurrent state |
| `stack` | Verify call path |
| `list` | Confirm current location |

### When to Hand Off to godebug

Verification discovers a bug → switch to debugging:

| Verification Result | Next Step |
|---------------------|-----------|
| CONFIRMED | Done - proceed with confidence |
| REFUTED (simple) | Fix inline, re-verify |
| REFUTED (complex) | Switch to godebug for deep investigation |
| INCONCLUSIVE | Adjust assertions, re-verify |

### Handoff Trigger

```
[Verification REFUTED]
        ↓
  Is root cause obvious?
        ↓
   YES → Fix → Re-verify
   NO  → Use godebug skill for deep debugging
```

## 8. Result Interpretation

### CONFIRMED - Proceed with Confidence

```json
// Expected value matches
{"value": "100"}  // Hypothesis: total == 100 ✓

// Expected breakpoint hit
{"breakpoint": {"id": 1, "line": 42}}  // Hypothesis: reaches line 42 ✓

// Expected call order
// First hit: Add(), Second hit: Wait() ✓
```

**Action:** Code verified. Mark task complete.

### REFUTED - Capture State, Fix, Re-verify

```json
// Value mismatch
{"value": "97"}  // Hypothesis: counter == 100, but got 97 ✗

// Unexpected breakpoint order
// First hit: Wait(), expected Add() first ✗

// Concurrent access without lock
// Two goroutines at critical section simultaneously ✗
```

**Action:**
1. Capture the actual state (save the JSON output)
2. Identify the discrepancy
3. Fix the code
4. Re-run exact same verification
5. Repeat until CONFIRMED

### INCONCLUSIVE - Adjust and Retry

```
// Breakpoint never hit - wrong location?
// Program exited before reaching assertion point
// Value is correct but for wrong reason
```

**Action:**
1. Check breakpoint locations
2. Verify code path actually executes
3. Add more assertion points
4. Re-run with adjusted strategy

## 9. Detection Criteria

Use this skill when:

- You just wrote Go code and want to verify it works
- You're working with goroutines, channels, or sync primitives
- You need to prove concurrent code is correctly synchronized
- Before marking a coding task as complete
- User asks to "verify", "confirm", or "prove" code behavior
- You want to "close the loop" on generated code

Do NOT use this skill when:

- You're investigating an existing bug (use godebug)
- You need to find why something crashed (use godebug)
- You just need to run tests (use `go test`)
- The code is trivial and obviously correct

## 10. Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│                   VERIFICATION PROTOCOL                      │
├─────────────────────────────────────────────────────────────┤
│  HYPOTHESIS: State what MUST be true                        │
│  ASSERTION:  Set breakpoints at verification points         │
│  EXECUTION:  godebug continue → eval specific values        │
│  CONCLUSION: CONFIRMED → done | REFUTED → fix → re-verify   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   COMMON ASSERTIONS                          │
├─────────────────────────────────────────────────────────────┤
│  WaitGroup:    break "sync.(*WaitGroup).Add"                │
│                break "sync.(*WaitGroup).Wait"               │
│                                                             │
│  Mutex:        break "sync.(*Mutex).Lock"                   │
│                                                             │
│  Value check:  eval "variable == expected"                  │
│                                                             │
│  Control flow: break at branch, then list after next        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   TOKEN-EFFICIENT                            │
├─────────────────────────────────────────────────────────────┤
│  DO:   eval "x.Field"         (targeted)                    │
│  NOT:  locals                 (dumps everything)            │
│                                                             │
│  DO:   Single assertion point                               │
│  NOT:  Breakpoints everywhere                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   SELF-CORRECTION                            │
├─────────────────────────────────────────────────────────────┤
│  REFUTED → Diagnose → Fix → Re-verify → CONFIRMED?          │
│                ↑                              │ No           │
│                └──────────────────────────────┘              │
│                                                             │
│  Never mark done until CONFIRMED                            │
└─────────────────────────────────────────────────────────────┘
```
