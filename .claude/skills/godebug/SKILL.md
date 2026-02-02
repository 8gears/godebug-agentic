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

## Detection Criteria

Use this skill when:
- Debugging Go applications from command line
- Setting breakpoints and stepping through code
- Inspecting variables, stack traces, goroutines
- Attaching to running Go processes
- Testing with conditional breakpoints
- The user mentions `godebug` or stateless debugging

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

### Breakpoint Not Hitting

- Verify line number has executable code (not comment/blank)
- Check file path matches exactly
- Ensure code path is actually executed

```bash
# List sources to verify file path
godebug --addr $ADDR sources | grep myfile

# Check breakpoints are set
godebug --addr $ADDR breakpoints
```

### Timeout Errors

Increase timeout for slow operations:
```bash
godebug --addr $ADDR --timeout 60s continue
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
