---
name: godebug
description: Stateless CLI debugger for Go applications using Delve. Use when debugging Go programs via command line, setting breakpoints, inspecting variables, stepping through code, or analyzing goroutines. Each command outputs JSON and exits - perfect for AI agents.
---

# godebug - Stateless Go Debugger CLI

A single-command CLI debugger for Go applications. Each invocation runs one command, outputs structured JSON, and exits. Designed for AI agent tool calling.

## Quick Start

```bash
# 1. Start debug session (auto-assigns port)
godebug start ./myapp
# Returns: {"data": {"addr": "127.0.0.1:58656", ...}}

# 2. Set breakpoint
godebug --addr 127.0.0.1:58656 break main.go:42

# 3. Run to breakpoint
godebug --addr 127.0.0.1:58656 continue

# 4. Inspect state
godebug --addr 127.0.0.1:58656 locals
godebug --addr 127.0.0.1:58656 stack

# 5. End session
godebug --addr 127.0.0.1:58656 quit
```

## Prerequisites

### Debug Symbols Required

Binaries **must** be compiled with debug symbols for breakpoints to work. Without debug symbols, breakpoints will not be hit and the program will run to completion.

```bash
# CORRECT - includes debug symbols, disables optimizations
go build -gcflags="all=-N -l" -o ./myapp .

# CORRECT - default go build includes symbols
go build -o ./myapp .

# WRONG - strips symbols, breakpoints won't work!
go build -ldflags "-s -w" -o ./myapp .
```

**Flags explained:**
- `-gcflags="all=-N -l"`: Disables optimizations (`-N`) and inlining (`-l`) for all packages
- `-ldflags "-s -w"`: Strips debug symbols (`-s`) and DWARF info (`-w`) - **avoid this for debugging**

If breakpoints are not being hit, rebuild the binary with debug symbols.

## Starting Delve

There are two ways to start a debug session. Choose based on whether your program produces output.

### Recommended: Manual Delve Start (Programs with Output)

**Use this approach for most Go programs.** Any program that uses `fmt.Println`, `log.Println`, or writes to stdout/stderr will cause SIGPIPE (exit code 13) with `godebug start`.

```bash
# 1. Start Delve in background (keeps stdout connected to terminal)
dlv debug ./myapp --headless --api-version=2 --listen=:4445 --accept-multiclient &
sleep 2  # Wait for server to start

# 2. Connect with godebug
godebug connect localhost:4445

# 3. Debug as normal
godebug --addr localhost:4445 break main.main
godebug --addr localhost:4445 continue
```

**For pre-compiled binaries:**
```bash
dlv exec ./myapp --headless --api-version=2 --listen=:4445 --accept-multiclient &
sleep 2
godebug connect localhost:4445
```

**For tests:**
```bash
dlv test ./pkg/... --headless --api-version=2 --listen=:4445 --accept-multiclient &
sleep 2
godebug connect localhost:4445
```

**Key flags explained:**
- `--headless`: Run without terminal UI (required for godebug)
- `--api-version=2`: Use Delve API v2 (required)
- `--listen=:4445`: Port to listen on (choose any free port)
- `--accept-multiclient`: Allow reconnection if connection drops
- `&`: Run in background so stdout stays connected to terminal

### Alternative: godebug start (Silent Programs Only)

Use `godebug start` only for programs that produce **no output**:

```bash
# Only for programs without any fmt.Println, log.Println, etc.
godebug start ./myapp
# Returns: {"data": {"addr": "127.0.0.1:58656", ...}}

godebug --addr 127.0.0.1:58656 break main.main
godebug --addr 127.0.0.1:58656 continue
```

**Why this limitation exists:** `godebug start` launches Delve as a subprocess. When the command returns, the pipe for the target program's stdout is closed. Any subsequent `fmt.Println` in the target causes SIGPIPE (exit code 13), terminating the program before breakpoints are hit.

### Quick Reference

| Program Type | Start Method |
|--------------|--------------|
| Has `fmt.Println`, `log.*`, stdout writes | `dlv debug ... &` + `godebug connect` |
| Silent (no output) | `godebug start` |
| Tests | `dlv test ... &` + `godebug connect` |
| Pre-compiled binary | `dlv exec ... &` + `godebug connect` |
| Remote debugging | `dlv debug --listen=0.0.0.0:4445` on remote |

### Cleanup

When done debugging, clean up the Delve process:

```bash
# End the debug session
godebug --addr localhost:4445 quit

# If Delve is still running in background
pkill -f "dlv debug"
# or
pkill -f "dlv exec"
```

## Detection Criteria

Use this skill when:
- Debugging Go applications from command line
- Setting breakpoints and stepping through code
- Inspecting variables, stack traces, goroutines
- Attaching to running Go processes
- Testing with conditional breakpoints
- The user mentions `godebug` or stateless debugging

## Debugging Workflow Order

**The debugger is a tool of last resort, not first resort.** Before starting a debug session, use faster tools in this order:

```
1. Run the program     →  Observe error messages, panic traces, output
2. Static analysis     →  Read the code, look for known antipatterns
3. Race detector       →  go run -race / go test -race (for concurrency bugs)
4. Targeted logging    →  Add temporary log statements at key points
5. Debugger            →  Only when above tools don't reveal the issue
```

### When Simpler Tools Are Faster

| Situation | Better Tool | Why |
|-----------|-------------|-----|
| Concurrency bug suspected | `go test -race` | Pinpoints exact lines with no stepping |
| Panic with stack trace | Read the trace | Location already provided |
| Known bug pattern | Code review | Pattern recognition beats stepping |
| Need to see one value | Temporary `log.Printf` | Faster than breakpoint setup |
| Regression after change | `git diff` + review | Compare what changed |

### When the Debugger IS the Right Tool

Use godebug when:
- **Runtime state inspection**: You need to see actual variable values at specific moments
- **Complex interactions**: Multiple goroutines, timing-dependent behavior that's hard to reason about
- **Unclear reproduction**: Bug is intermittent or behavior doesn't match code reading
- **No obvious pattern**: You've reviewed code and used race detector but can't spot the issue
- **Exploration**: Understanding unfamiliar code paths by stepping through execution

## AI Debugging Strategy

### The Scientific Protocol

Stop "flailing" (randomly changing code). Follow this cycle:

1. **Observation:** Gather raw data (logs, panic traces) without interpretation
2. **Hypothesis:** Formulate a falsifiable theory (e.g., "The goroutine hangs because the channel is unbuffered")
3. **Prediction:** Define what *must* happen if the hypothesis is true
4. **Experiment:** Change **one** variable only. Run the test
5. **Analysis:** If prediction failed, revert the change. You eliminated one possibility

### Symptom → Command Sequence

| Symptom | First Command | If Result Shows | Then |
|---------|---------------|-----------------|------|
| Wrong variable value | `locals` | unexpected value | `eval` parent struct, `stack` for call path |
| Program crashes/panics | `break` on error line | stack trace | `frame N` + `locals` for each frame |
| Program hangs | `goroutines` | blocked goroutines | `goroutine N` + `stack` to find blocker |
| Race condition suspected | `go test -race` first | race location | `break` both locations |
| Wrong control flow | `break` at branch point | wrong branch taken | `eval` condition expression |
| Test fails | `start --mode test` | test location | `break` in test, then `step` |
| Intermittent bug | `break --cond` | specific state | `locals` + `goroutines` |

### Bug Type → Command Sequence

**Data Bug (wrong value):**
1. `break <file>:<line>` at assignment
2. `continue`
3. `locals` to see current state
4. `eval "expr"` for specific expressions
5. `stack` to trace where value came from
6. `frame N` + `locals` to inspect caller's state

**Control Flow Bug (wrong branch):**
1. `break <file>:<line>` at decision point
2. `continue`
3. `eval "condition"` to see boolean result
4. `step` to confirm which branch executes

**Concurrency Bug (race/deadlock):**
1. First: `go test -race ./...` outside debugger
2. `break` on sync primitive or shared variable
3. `goroutines` to see all goroutine states
4. `goroutine N` → `stack` → `locals` for each relevant goroutine
5. Look for: missing locks, wrong order, blocked channels

**Deadlock Investigation:**
1. `goroutines` - identify which are in `runtime.gopark`
2. `goroutine N` → `stack` for each blocked goroutine
3. Draw dependency: "Goroutine A holds Lock 1, waiting for Lock 2. Goroutine B holds Lock 2, waiting for Lock 1"
4. The cycle is the deadlock

**Memory/Resource Leak:**
1. Use `pprof` first: `go tool pprof http://localhost:6060/debug/pprof/heap`
2. `break` at allocation site identified by pprof
3. `stack` to see who allocates
4. `eval "len(slice)"` or `eval "cap(slice)"` to check growth

## Debugging Do's and Don'ts

### Mindset & Strategy

| Category | DON'T | DO |
|----------|-------|-----|
| **Bias** | Confirmation bias: "I know X is fine" → ignore X's logs | Falsification: "If it's NOT X, then this must return Y. Verify." |
| **Focus** | Tunnel vision: 4 hours in one function | Divide & conquer: verify inputs/outputs at boundaries first |
| **Process** | Shotgun: change 3 things at once | Isolation: change ONE thing, revert if no fix, try next |
| **Ego** | "The library is broken" | "What assumption have I made that is incorrect?" |

### Go-Specific Tactics

| Category | DON'T | DO |
|----------|-------|-----|
| **Observability** | `fmt.Println("HERE 1")` spam | `godebug break` + `locals` - no code changes |
| **Concurrency** | Stare at code imagining races | Run `go test -race` - let runtime prove it |
| **Logging** | `log.Println("Error happened")` | Structured: `slog.Error("failed", "userID", id, "error", err)` |
| **Regressions** | Manual checkout of random commits | `git bisect run go test -run TestBroken ./...` |
| **Testing** | Happy path only | Table-driven with nil, empty, edge cases |
| **Deadlocks** | Random `time.Sleep()` calls | Stack dump → identify `runtime.gopark` → draw lock graph |

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--addr` | Delve server address (host:port) | Required for all commands except `start` |
| `--output` | Output format: `json` or `text` | `json` |
| `--timeout` | Operation timeout (e.g., `10s`, `1m`) | `30s` |

## Command Reference

### Session Management

#### `start` - Start Debug Session

Starts a Delve debug server and returns the connection address.

```bash
# Debug mode (default) - compile and debug
godebug start ./cmd/myapp

# Test mode - debug tests
godebug start --mode test ./...

# Exec mode - debug pre-compiled binary
godebug start --mode exec ./binary

# With program arguments
godebug start ./cmd/myapp -- -port 8080
```

**Flags:**
- `--mode`: Debug mode: `debug` (default), `test`, or `exec`

**Output:**
```json
{
  "success": true,
  "command": "start",
  "data": {
    "addr": "127.0.0.1:58656",
    "mode": "debug",
    "pid": 87833,
    "target": "./testdata/debugme"
  },
  "message": "Debug server started"
}
```

#### `connect` - Connect to Existing Server

Connect to a manually started Delve server.

```bash
# First, start Delve manually
dlv debug ./myapp --headless --api-version=2 --listen=:2345

# Then connect
godebug connect localhost:2345
```

**Output:**
```json
{
  "success": true,
  "command": "connect",
  "data": {
    "addr": "127.0.0.1:2345",
    "running": false
  },
  "message": "Connected to debug server"
}
```

#### `status` - Show Debug State

```bash
godebug --addr 127.0.0.1:2345 status
```

**Output:**
```json
{
  "success": true,
  "command": "status",
  "data": {
    "exited": false,
    "running": false
  },
  "message": "Process paused"
}
```

#### `restart` - Restart Program

```bash
godebug --addr 127.0.0.1:2345 restart
```

#### `quit` - End Session

```bash
godebug --addr 127.0.0.1:2345 quit
```

**Output:**
```json
{
  "success": true,
  "command": "quit",
  "message": "Debug session terminated"
}
```

### Breakpoints

#### `break` - Set Breakpoint

```bash
# Set breakpoint at file:line
godebug --addr 127.0.0.1:2345 break main.go:36

# Set breakpoint at function
godebug --addr 127.0.0.1:2345 break main.innerFunc

# Conditional breakpoint - only stops when condition is true
godebug --addr 127.0.0.1:2345 break --cond "i > 2" main.go:42
```

**Flags:**
- `--cond`: Condition expression (e.g., `"x > 10"`, `"name == \"test\""`)

**Shell Quoting for Method Names:**

Function names with special characters (parentheses, asterisks) must be quoted:

```bash
# WRONG - shell interprets ( and * as glob/subshell
godebug --addr $ADDR break sync.(*WaitGroup).Wait

# CORRECT - quote the function name
godebug --addr $ADDR break "sync.(*WaitGroup).Wait"
godebug --addr $ADDR break "sync.(*Mutex).Lock"
godebug --addr $ADDR break "bytes.(*Buffer).Write"
```

**Output (standard breakpoint):**
```json
{
  "success": true,
  "command": "break",
  "data": {
    "file": "/path/to/main.go",
    "function": "main.innerFunc",
    "id": 1,
    "line": 36
  },
  "message": "Breakpoint 1 set"
}
```

**Output (conditional breakpoint):**
```json
{
  "success": true,
  "command": "break",
  "data": {
    "condition": "i > 2",
    "file": "/path/to/main.go",
    "function": "main.processItems",
    "id": 1,
    "line": 42
  },
  "message": "Breakpoint 1 set"
}
```

#### `breakpoints` - List Breakpoints

```bash
godebug --addr 127.0.0.1:2345 breakpoints
```

**Output:**
```json
{
  "success": true,
  "command": "breakpoints",
  "data": {
    "breakpoints": [
      {
        "enabled": true,
        "file": "/path/to/main.go",
        "function": "main.innerFunc",
        "id": 1,
        "line": 36
      }
    ],
    "count": 1
  },
  "message": "1 breakpoints"
}
```

#### `clear` - Remove Breakpoint

```bash
godebug --addr 127.0.0.1:2345 clear 1
```

**Output:**
```json
{
  "success": true,
  "command": "clear",
  "data": {
    "file": "/path/to/main.go",
    "id": 1,
    "line": 36
  },
  "message": "Breakpoint 1 cleared"
}
```

### Execution Control

#### `continue` - Resume Execution

```bash
godebug --addr 127.0.0.1:2345 continue
```

**Output:**
```json
{
  "success": true,
  "command": "continue",
  "data": {
    "breakpoint": {
      "file": "/path/to/main.go",
      "id": 1,
      "line": 36
    },
    "exited": false,
    "goroutine": {
      "id": 1
    },
    "location": {
      "file": "/path/to/main.go",
      "function": "main.innerFunc",
      "line": 36
    },
    "running": false
  },
  "message": "Stopped at breakpoint"
}
```

#### `next` - Step Over

Execute next line, stepping over function calls.

```bash
godebug --addr 127.0.0.1:2345 next
```

**Output:**
```json
{
  "success": true,
  "command": "next",
  "data": {
    "exited": false,
    "goroutine": {"id": 1},
    "location": {
      "file": "/path/to/main.go",
      "function": "main.middleFunc",
      "line": 31
    },
    "running": false
  },
  "message": "Stepped to next line"
}
```

#### `step` - Step Into

Step into function calls.

```bash
godebug --addr 127.0.0.1:2345 step
```

**Output:**
```json
{
  "success": true,
  "command": "step",
  "data": {
    "exited": false,
    "goroutine": {"id": 1},
    "location": {
      "file": "/path/to/main.go",
      "function": "main.middleFunc",
      "line": 32
    },
    "running": false
  },
  "message": "Stepped into function"
}
```

#### `stepout` - Step Out

Step out of current function.

```bash
godebug --addr 127.0.0.1:2345 stepout
```

**Output:**
```json
{
  "success": true,
  "command": "stepout",
  "data": {
    "exited": false,
    "goroutine": {"id": 1},
    "location": {
      "file": "/path/to/main.go",
      "function": "main.outerFunc",
      "line": 26
    },
    "running": false
  },
  "message": "Stepped out of function"
}
```

### Variable Inspection

#### `locals` - Show Local Variables

```bash
godebug --addr 127.0.0.1:2345 locals
```

**Output:**
```json
{
  "success": true,
  "command": "locals",
  "data": {
    "count": 3,
    "variables": [
      {
        "children": [
          {"name": "", "type": "int", "value": "2"},
          {"name": "", "type": "int", "value": "4"},
          {"name": "", "type": "int", "value": "6"}
        ],
        "name": "result",
        "type": "[]int",
        "value": ""
      },
      {"name": "i", "type": "int", "value": "3"},
      {"name": "item", "type": "int", "value": "4"}
    ]
  },
  "message": "3 local variables"
}
```

#### `args` - Show Function Arguments

```bash
godebug --addr 127.0.0.1:2345 args
```

**Output:**
```json
{
  "success": true,
  "command": "args",
  "data": {
    "arguments": [
      {"name": "x", "type": "int", "value": "25"},
      {"name": "~r0", "type": "int", "value": "0"}
    ],
    "count": 2
  },
  "message": "2 arguments"
}
```

#### `eval` - Evaluate Expression

```bash
godebug --addr 127.0.0.1:2345 eval "x"
godebug --addr 127.0.0.1:2345 eval "x * 2"
godebug --addr 127.0.0.1:2345 eval "len(result)"
```

**Output:**
```json
{
  "success": true,
  "command": "eval",
  "data": {
    "expression": "x",
    "name": "x",
    "type": "int",
    "value": "25"
  }
}
```

### Stack Navigation

#### `stack` - Show Stack Trace

```bash
# Full stack trace
godebug --addr 127.0.0.1:2345 stack

# Limited depth
godebug --addr 127.0.0.1:2345 stack --depth 3
```

**Flags:**
- `--depth`: Maximum number of frames to show

**Output:**
```json
{
  "success": true,
  "command": "stack",
  "data": {
    "count": 4,
    "frames": [
      {"file": "/path/to/main.go", "function": "main.innerFunc", "index": 0, "line": 36},
      {"file": "/path/to/main.go", "function": "main.middleFunc", "index": 1, "line": 31},
      {"file": "/path/to/main.go", "function": "main.outerFunc", "index": 2, "line": 26},
      {"file": "/path/to/main.go", "function": "main.main", "index": 3, "line": 7}
    ],
    "goroutineId": 1
  },
  "message": "4 frames"
}
```

#### `frame` - Switch Stack Frame

```bash
godebug --addr 127.0.0.1:2345 frame 1
```

**Output:**
```json
{
  "success": true,
  "command": "frame",
  "data": {
    "file": "/path/to/main.go",
    "function": "main.middleFunc",
    "index": 1,
    "line": 31
  },
  "message": "Switched to frame 1"
}
```

### Goroutine Management

#### `goroutines` - List All Goroutines

```bash
godebug --addr 127.0.0.1:2345 goroutines
```

**Output:**
```json
{
  "success": true,
  "command": "goroutines",
  "data": {
    "count": 6,
    "goroutines": [
      {
        "id": 1,
        "location": {
          "file": "/path/to/main.go",
          "function": "main.innerFunc",
          "line": 36
        },
        "selected": true
      },
      {
        "id": 2,
        "location": {
          "file": "/usr/local/go/src/runtime/proc.go",
          "function": "runtime.gopark",
          "line": 461
        },
        "selected": false
      }
    ],
    "selectedId": 1
  },
  "message": "6 goroutines"
}
```

#### `goroutine` - Switch Goroutine

```bash
godebug --addr 127.0.0.1:2345 goroutine 2
```

**Output:**
```json
{
  "success": true,
  "command": "goroutine",
  "data": {
    "id": 2,
    "location": {
      "file": "/usr/local/go/src/runtime/proc.go",
      "function": "runtime.gopark",
      "line": 461
    }
  },
  "message": "Switched to goroutine 2"
}
```

### Source Code

#### `list` - Show Source Code

```bash
# Show source at current location
godebug --addr 127.0.0.1:2345 list

# Show with custom context (lines before/after)
godebug --addr 127.0.0.1:2345 list --context 3
```

**Flags:**
- `--context`: Number of lines before and after current line (default: 5)

**Output:**
```json
{
  "success": true,
  "command": "list",
  "data": {
    "currentLine": 36,
    "file": "/path/to/main.go",
    "function": "main.innerFunc",
    "lines": [
      {"content": "func innerFunc(x int) int {", "current": false, "lineNumber": 35},
      {"content": "\treturn x * x // Breakpoint here", "current": true, "lineNumber": 36},
      {"content": "}", "current": false, "lineNumber": 37}
    ]
  },
  "message": "/path/to/main.go:36"
}
```

#### `sources` - List Source Files

```bash
godebug --addr 127.0.0.1:2345 sources
```

**Output:**
```json
{
  "success": true,
  "command": "sources",
  "data": {
    "count": 310,
    "sources": [
      "/path/to/main.go",
      "/path/to/other.go",
      "/usr/local/go/src/fmt/print.go"
    ]
  },
  "message": "310 sources"
}
```

## Core Workflows

### Basic Debugging Workflow

```bash
# 1. Start session
godebug start ./cmd/myapp
# Save the addr from output: 127.0.0.1:58656

# 2. Set breakpoints
godebug --addr 127.0.0.1:58656 break main.go:42
godebug --addr 127.0.0.1:58656 break pkg/handler.go:100

# 3. Run to breakpoint
godebug --addr 127.0.0.1:58656 continue

# 4. Inspect state
godebug --addr 127.0.0.1:58656 locals
godebug --addr 127.0.0.1:58656 args
godebug --addr 127.0.0.1:58656 stack

# 5. Step through code
godebug --addr 127.0.0.1:58656 next    # Step over
godebug --addr 127.0.0.1:58656 step    # Step into
godebug --addr 127.0.0.1:58656 stepout # Step out

# 6. Continue or quit
godebug --addr 127.0.0.1:58656 continue
godebug --addr 127.0.0.1:58656 quit
```

### Conditional Breakpoint Workflow

Use conditional breakpoints to stop only when specific conditions are met:

```bash
# Start session
godebug start ./myapp
# addr: 127.0.0.1:58656

# Break only when loop variable exceeds threshold
godebug --addr 127.0.0.1:58656 break --cond "i > 100" main.go:50

# Break only for specific user
godebug --addr 127.0.0.1:58656 break --cond "user.ID == 42" handlers.go:75

# Break when slice is empty
godebug --addr 127.0.0.1:58656 break --cond "len(items) == 0" process.go:30

# Continue - will only stop when condition is true
godebug --addr 127.0.0.1:58656 continue
```

### Test Debugging Workflow

```bash
# 1. Start in test mode
godebug start --mode test ./pkg/...
# Save addr: 127.0.0.1:58656

# 2. Set breakpoint in test or code under test
godebug --addr 127.0.0.1:58656 break pkg/handler_test.go:25
godebug --addr 127.0.0.1:58656 break pkg/handler.go:50

# 3. Run tests to breakpoint
godebug --addr 127.0.0.1:58656 continue

# 4. Debug as normal
godebug --addr 127.0.0.1:58656 locals
```

### Goroutine Debugging Workflow

```bash
# 1. Start session and run to interesting point
godebug start ./concurrent-app
godebug --addr 127.0.0.1:58656 break worker.go:30
godebug --addr 127.0.0.1:58656 continue

# 2. List all goroutines
godebug --addr 127.0.0.1:58656 goroutines

# 3. Switch to specific goroutine
godebug --addr 127.0.0.1:58656 goroutine 5

# 4. Inspect that goroutine's state
godebug --addr 127.0.0.1:58656 stack
godebug --addr 127.0.0.1:58656 locals

# 5. Switch back to main goroutine
godebug --addr 127.0.0.1:58656 goroutine 1
```

### Remote Debugging Workflow

```bash
# On remote machine: start Delve server
dlv debug ./myapp --headless --api-version=2 --accept-multiclient --listen=0.0.0.0:2345

# Locally: connect to remote server
godebug connect remote-host:2345

# Debug as normal with --addr
godebug --addr remote-host:2345 break main.go:42
godebug --addr remote-host:2345 continue
```

### Race Condition Debugging Workflow

Race conditions in concurrent Go code can be challenging to debug because they may execute faster than breakpoints can catch them. Use this workflow:

#### Step 1: Confirm the Race with Go's Race Detector

Before using the debugger, confirm the race condition exists:

```bash
# Run with race detector
go run -race ./myapp

# Or test with race detector
go test -race ./...
```

The race detector will report data races with stack traces showing where they occur.

#### Step 2: Set Breakpoints on Sync Primitives

For WaitGroup, Mutex, or channel issues, set breakpoints on the sync package methods:

```bash
# Start debug session
godebug start ./myapp

# Breakpoints on WaitGroup methods (note: must quote due to special chars)
godebug --addr $ADDR break "sync.(*WaitGroup).Add"
godebug --addr $ADDR break "sync.(*WaitGroup).Done"
godebug --addr $ADDR break "sync.(*WaitGroup).Wait"

# Breakpoints on Mutex methods
godebug --addr $ADDR break "sync.(*Mutex).Lock"
godebug --addr $ADDR break "sync.(*Mutex).Unlock"

# Breakpoints on channel operations (runtime)
godebug --addr $ADDR break "runtime.chansend1"
godebug --addr $ADDR break "runtime.chanrecv1"
```

#### Step 3: Set Breakpoints BEFORE Goroutine Creation

For race conditions involving goroutine startup, set breakpoints **before** the `go` statement, not inside the goroutine:

```go
// Example buggy code:
for i := 0; i < 10; i++ {
    go func(id int) {
        wg.Add(1)  // BUG: Add() inside goroutine
        // ...
    }(i)
}
wg.Wait()  // May return immediately!
```

```bash
# Set breakpoint BEFORE the for loop, not inside the goroutine
godebug --addr $ADDR break main.go:15  # Line with 'for'

# Use step to watch goroutine creation
godebug --addr $ADDR step
```

#### Step 4: Use Conditional Breakpoints for Specific States

```bash
# Break when WaitGroup counter might be wrong
godebug --addr $ADDR break --cond "counter == 0" main.go:50

# Break on specific goroutine count
godebug --addr $ADDR break --cond "len(goroutines) > 5" worker.go:30
```

#### Step 5: Inspect All Goroutines

When stopped, examine all goroutines to understand the race:

```bash
# List all goroutines
godebug --addr $ADDR goroutines

# Switch to each goroutine and inspect
godebug --addr $ADDR goroutine 5
godebug --addr $ADDR stack
godebug --addr $ADDR locals
```

#### Common Race Patterns

| Pattern | Symptom | Debug Strategy |
|---------|---------|----------------|
| WaitGroup.Add inside goroutine | Wait() returns early | Break on `sync.(*WaitGroup).Add`, check call location |
| Missing mutex lock | Data corruption | Break on shared variable access |
| Channel send/recv mismatch | Deadlock or panic | Break on channel operations |
| Closure capturing loop var | Wrong values | Break inside goroutine, check captured values |

#### When Races Are Too Fast

If the race always wins and breakpoints never hit:

1. The bug is **deterministic** - analyze the code statically
2. Add `time.Sleep()` temporarily to slow down the race
3. Use `GOMAXPROCS=1` to serialize goroutine execution:
   ```bash
   GOMAXPROCS=1 godebug start ./myapp
   ```
4. Set breakpoint at `main.main` and use `step` instead of `continue`

## Best Practices

### 1. Track the Address

The `--addr` flag is required for all commands after `start`. Always capture and reuse the address:

```bash
# Good: Save the address
ADDR=$(godebug start ./myapp | jq -r '.data.addr')
godebug --addr $ADDR break main.go:42

# Or extract from JSON manually
# {"data": {"addr": "127.0.0.1:58656"}} -> use 127.0.0.1:58656
```

### 2. Use Conditional Breakpoints for Loops

Don't step through 1000 iterations - use conditions:

```bash
# Bad: Will stop on every iteration
godebug --addr $ADDR break main.go:42

# Good: Only stop when interesting
godebug --addr $ADDR break --cond "i == 999" main.go:42
godebug --addr $ADDR break --cond "err != nil" main.go:42
```

### 3. Check Status Before Commands

```bash
# Verify session state before continuing
godebug --addr $ADDR status
# If running: true, wait or use another command
# If exited: true, session ended
```

### 4. Use Stack Depth for Large Call Stacks

```bash
# Limit output for deep recursion
godebug --addr $ADDR stack --depth 5
```

### 5. Clean Up Sessions

Always quit when done to release resources:

```bash
godebug --addr $ADDR quit
```

## Exit Codes

There are **two types** of exit codes to understand:

1. **godebug CLI exit codes** - The exit code returned by the `godebug` command itself
2. **Target process exit status** - The exit status of the debugged program (in JSON `data.exitStatus`)

### godebug CLI Exit Codes

These are the exit codes returned by godebug commands. Use `echo $?` after running a command to check.

| Code | Constant | Meaning | JSON Error Code |
|------|----------|---------|-----------------|
| 0 | `ExitSuccess` | Command completed successfully | - |
| 1 | `ExitGenericError` | Unspecified error | `INTERNAL_ERROR`, `EVAL_FAILED` |
| 2 | `ExitUsageError` | Invalid arguments or flags | `INVALID_ARGUMENT` |
| 3 | `ExitConnectionError` | Cannot connect to Delve server | `CONNECTION_FAILED`, `CONNECTION_REFUSED` |
| 4 | `ExitNotFound` | Resource not found (breakpoint, goroutine, frame) | `NOT_FOUND` |
| 124 | `ExitTimeout` | Operation timed out (GNU timeout convention) | `TIMEOUT` |
| 125 | `ExitProcessError` | Target process error | `PROCESS_EXITED` |

**JSON Error Codes Reference:**

| Error Code | Description |
|------------|-------------|
| `CONNECTION_FAILED` | Cannot reach the Delve server |
| `CONNECTION_REFUSED` | Server actively refused the connection |
| `TIMEOUT` | Operation exceeded time limit |
| `INVALID_ARGUMENT` | Bad input from user |
| `NOT_FOUND` | Requested resource doesn't exist |
| `PROCESS_EXITED` | Target program terminated |
| `EVAL_FAILED` | Expression evaluation failed |
| `INTERNAL_ERROR` | Unexpected internal error |

**Example:**
```bash
godebug --addr 127.0.0.1:2345 break main.go:999
echo $?  # Returns 4 if line doesn't exist (NOT_FOUND)

godebug --addr 127.0.0.1:9999 status
echo $?  # Returns 3 if server not running (CONNECTION_REFUSED)
```

### Target Process Exit Status (`data.exitStatus`)

When the debugged program exits, the `exitStatus` field in the JSON response shows how it terminated. This is separate from the CLI exit code.

**Common values:**

| Status | Meaning |
|--------|---------|
| 0 | Program completed successfully |
| 1 | General error (often `os.Exit(1)` or `log.Fatal()`) |
| 2 | Panic without recovery, or deadlock detected |
| n | Value passed to `os.Exit(n)` |

**Signal-based exits** (128 + signal number):

| Status | Signal | Meaning |
|--------|--------|---------|
| 130 | SIGINT (2) | Interrupted (Ctrl+C) |
| 131 | SIGQUIT (3) | Quit with core dump |
| 134 | SIGABRT (6) | Aborted |
| 137 | SIGKILL (9) | Killed forcefully |
| 139 | SIGSEGV (11) | Segmentation fault |
| 143 | SIGTERM (15) | Terminated |

**Go-specific patterns:**

| Scenario | Exit Status | Notes |
|----------|-------------|-------|
| `panic()` without recovery | 2 | Stack trace printed to stderr |
| `os.Exit(n)` | n | Deferred functions NOT called |
| `log.Fatal()` | 1 | Calls `os.Exit(1)` after logging |
| Deadlock detected | 2 | "fatal error: all goroutines are asleep" |
| Race detector found race | 66 | When running with `-race` |

### Example JSON Response

When the target process exits:

```json
{
  "success": true,
  "command": "continue",
  "data": {
    "exitStatus": 0,
    "exited": true,
    "running": false
  },
  "message": "Process exited"
}
```

**Key fields:**
- `exited: true` - The target process has terminated
- `exitStatus` - The exit code of the target process (not godebug)
- The godebug CLI itself returns exit code 0 (success) because the command worked

### Distinguishing the Two

```bash
# Run godebug and capture both exit codes
OUTPUT=$(godebug --addr 127.0.0.1:2345 continue)
CLI_EXIT=$?

# CLI exit code (was the godebug command successful?)
echo "CLI exit: $CLI_EXIT"

# Target process exit status (how did the debugged program end?)
TARGET_EXIT=$(echo "$OUTPUT" | jq -r '.data.exitStatus // empty')
echo "Target exit: $TARGET_EXIT"
```

## Troubleshooting

### "Connection refused"

The Delve server isn't running or wrong address:
```bash
# Verify server is running
ps aux | grep dlv

# Try starting fresh
godebug start ./myapp
```

### "No such file or directory"

Build the target first or use correct path:
```bash
# Ensure target exists
go build ./cmd/myapp
godebug start --mode exec ./cmd/myapp
```

### "Could not attach to pid"

On macOS, you may need to codesign Delve:
```bash
# Check if dlv is codesigned
codesign -d -v $(which dlv)
```

### Breakpoints Never Hit / Program Exits Immediately

This is often caused by **missing debug symbols** or **race conditions**.

**Check 1: Debug symbols present?**
```bash
# Rebuild with debug symbols
go build -gcflags="all=-N -l" -o ./myapp .

# Verify symbols exist (should show DWARF info)
go tool objdump ./myapp | head -20
```

**Check 2: Is binary stripped?**
```bash
# If built with -ldflags "-s -w", symbols are stripped
# Rebuild WITHOUT those flags
go build -o ./myapp .
```

**Check 3: Race condition causing early exit?**
```bash
# Set breakpoint at main.main first
godebug --addr $ADDR break main.main
godebug --addr $ADDR continue

# Then use step instead of continue
godebug --addr $ADDR step
godebug --addr $ADDR step
```

**Check 4: Code path not executed?**
```bash
# List sources to verify file is loaded
godebug --addr $ADDR sources | grep myfile

# Check breakpoints are actually set
godebug --addr $ADDR breakpoints
```

### Breakpoint on Wrong Line

Compiler optimizations can move code. Disable optimizations:
```bash
go build -gcflags="all=-N -l" -o ./myapp .
```

### "could not find function" Error

Function name may need quoting or different format:
```bash
# For methods with pointer receivers, quote the name
godebug --addr $ADDR break "sync.(*WaitGroup).Wait"

# For generic functions, include type parameters
godebug --addr $ADDR break "slices.Sort[int]"
```

### Timeout Errors

Increase timeout for slow operations:
```bash
godebug --addr $ADDR --timeout 60s continue
```

### Program Exits with Code 13

Exit code 13 (SIGPIPE) usually means the program completed normally but output was closed. This is common when:
- The program finishes before breakpoints are hit (race condition)
- Debug symbols are missing
- The code path with breakpoints isn't executed

See "Breakpoints Never Hit" above for solutions.

## Output Interpretation

### When `continue` Returns

| JSON Field | Value | Meaning | Next Action |
|------------|-------|---------|-------------|
| `exited` | `true` | Program ended | Check `exitStatus`, may need `restart` |
| `exited` | `false` | Stopped at breakpoint | `locals`, `stack`, `eval` to inspect |
| `exitStatus` | `0` | Clean exit | Bug may be logic, not crash |
| `exitStatus` | `2` | Panic or deadlock | Check stderr, `goroutines` |
| `exitStatus` | `66` | Race detected | Run with `-race` outside debugger |
| `breakpoint.id` | present | Stopped at breakpoint | Inspect state |
| `running` | `true` | Still executing | Wait or `status` to poll |

### When `locals` Returns

| Pattern | Interpretation | Next Action |
|---------|---------------|-------------|
| `value: "<nil>"` | Nil pointer/interface | Check initialization, `stack` for caller |
| `value: ""` (string) | Empty string | May be intentional or missing assignment |
| `children: []` | Empty slice/map | Check if expected, `eval "cap(x)"` |
| Complex nested struct | Large data | `eval "x.SpecificField"` for targeted inspection |

### When `goroutines` Returns

| Pattern | Interpretation | Next Action |
|---------|---------------|-------------|
| Many at `runtime.gopark` | Blocked waiting | Check what they wait for |
| At `sync.(*Mutex).Lock` | Waiting for lock | Find who holds lock |
| At `runtime.chanrecv` | Waiting on channel | Find sender |
| Only 1 at user code | Others blocked | Likely deadlock |
| Growing count | Goroutine leak | `pprof` goroutine profile |

### Error Recovery

| `error.code` | Meaning | Recovery |
|--------------|---------|----------|
| `CONNECTION_REFUSED` | Server not running | `start` new session |
| `NOT_FOUND` | Invalid breakpoint/goroutine | `breakpoints` or `goroutines` to list valid |
| `PROCESS_EXITED` | Target ended | `restart` or `quit` + new `start` |
| `TIMEOUT` | Operation too slow | Increase `--timeout`, check for infinite loop |
| `EVAL_FAILED` | Bad expression | Check variable exists in scope |

## Multi-Tool Integration

### godebug + Race Detector

Race detector finds locations; godebug inspects state.

```bash
# 1. Find race locations
go test -race ./... 2>&1 | tee race.log
# Extract: "main.go:42" and "main.go:67"

# 2. Debug with godebug
ADDR=$(godebug start ./myapp | jq -r '.data.addr')
godebug --addr $ADDR break main.go:42
godebug --addr $ADDR break main.go:67
godebug --addr $ADDR continue

# 3. Check goroutine states
godebug --addr $ADDR goroutines

# 4. Inspect each racing goroutine
godebug --addr $ADDR goroutine N
godebug --addr $ADDR locals
godebug --addr $ADDR stack
```

### godebug + pprof (Memory Leaks)

pprof identifies allocation sites; godebug inspects context.

```bash
# 1. Capture heap profile
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top10
# Identifies: "cache.go:89 allocates 500MB"

# 2. Debug allocation
ADDR=$(godebug start ./myapp | jq -r '.data.addr')
godebug --addr $ADDR break cache.go:89
godebug --addr $ADDR continue
godebug --addr $ADDR stack    # Who called this?
godebug --addr $ADDR locals   # What's allocated?
```

### godebug + pprof (Goroutine Leaks)

```bash
# 1. Get goroutine profile
curl http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt
# Look for many goroutines with same stack

# 2. Debug the leak
ADDR=$(godebug start ./myapp | jq -r '.data.addr')
godebug --addr $ADDR break worker.go:50  # Creation site
godebug --addr $ADDR continue
godebug --addr $ADDR locals    # What state causes leak?
godebug --addr $ADDR stack     # Full context
```

### godebug + git bisect (Regressions)

```bash
# 1. Find breaking commit
git bisect start
git bisect bad HEAD
git bisect good v1.0.0
git bisect run go test -run TestBroken ./...
# Result: "abc123 is first bad commit"

# 2. Debug that commit
git checkout abc123
ADDR=$(godebug start ./myapp | jq -r '.data.addr')
godebug --addr $ADDR break <changed_file>:<line>
godebug --addr $ADDR continue
godebug --addr $ADDR locals

git bisect reset
```

## Output Format

All commands output JSON with this structure:

```json
{
  "success": true|false,
  "command": "command-name",
  "data": { ... },      // Command-specific data
  "message": "Human-readable summary",
  "error": {            // Only on failure
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

Use `--output text` for human-readable output instead of JSON.
