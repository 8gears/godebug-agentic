# godebug

A stateless CLI debugger for Go applications, designed for AI agent tool calling. A simpler alternative to MCP-based debugging. AI agents use CLI tools equally good as MCPs, no protocol overhead required.

## Overview

`godebug` wraps the [Delve](https://github.com/go-delve/delve) debugger with a single-command interface that outputs structured JSON. Each invocation runs one command, outputs the result, and exits—perfect for AI agents that work with stateless command → response patterns.

```
┌─────────────────────────────────────────────────────────────┐
│                     Delve Headless Server                   │
│                   (persistent, maintains state)             │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │ JSON-RPC
                              │
┌─────────────────────────────────────────────────────────────┐
│  godebug start ./app    │  godebug break main.go:42        │
│  godebug continue       │  godebug locals                  │
│  (each command connects, executes, outputs JSON, exits)    │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.21+
- [Delve](https://github.com/go-delve/delve) debugger (`go install github.com/go-delve/delve/cmd/dlv@latest`)

### From Source

```bash
git clone https://github.com/8gears/godebug.git
cd godebug
go install .
```

Or using Task:

```bash
task install
```

## Quick Start

```bash
# 1. Start a debug session
godebug start ./cmd/myapp
# Returns: {"success":true,"data":{"addr":"127.0.0.1:38697",...}}

# 2. Use the returned address for all subsequent commands
ADDR="127.0.0.1:38697"

# 3. Set a breakpoint
godebug --addr $ADDR break main.go:42

# 4. Continue execution
godebug --addr $ADDR continue

# 5. Inspect variables
godebug --addr $ADDR locals

# 6. End the session
godebug --addr $ADDR quit
```

## Commands

### Session Management

| Command | Description |
|---------|-------------|
| `start <target>` | Start debugging a Go package/binary |
| `connect <addr>` | Connect to an existing Delve server |
| `status` | Show current debug state (running/paused/exited) |
| `restart` | Restart the debugged program |
| `quit` | Stop debugging and terminate the server |

### Breakpoints

| Command | Description |
|---------|-------------|
| `break <location>` | Set a breakpoint (file:line or function) |
| `break <loc> --cond "expr"` | Set a conditional breakpoint |
| `clear <id>` | Remove a breakpoint by ID |
| `breakpoints` | List all breakpoints |

### Execution Control

| Command | Description |
|---------|-------------|
| `continue` | Continue until next breakpoint |
| `next` | Step over to next line |
| `step` | Step into function call |
| `stepout` | Step out of current function |

### Inspection

| Command | Description |
|---------|-------------|
| `locals` | Show local variables |
| `args` | Show function arguments |
| `eval <expr>` | Evaluate a Go expression |

### Navigation

| Command | Description |
|---------|-------------|
| `stack` | Show stack trace |
| `stack --depth N` | Show stack trace with depth limit |
| `frame <index>` | Switch to stack frame |
| `goroutines` | List all goroutines |
| `goroutine <id>` | Switch to goroutine |

### Source

| Command | Description |
|---------|-------------|
| `list` | Show source at current location |
| `list --context N` | Show source with N lines of context |
| `sources [filter]` | List all source files |

## Global Flags

| Flag | Description |
|------|-------------|
| `--addr <host:port>` | Delve server address (required for all commands except `start`/`connect`) |
| `--output json` | JSON output (default) |
| `--output text` | Human-readable output |

## Output Format

All commands return a consistent JSON envelope:

```json
{
  "success": true,
  "command": "continue",
  "data": {
    "location": {
      "file": "/path/to/main.go",
      "line": 42,
      "function": "main.handleRequest"
    },
    "goroutine": {"id": 1},
    "breakpoint": {"id": 1, "file": "main.go", "line": 42}
  },
  "message": "Stopped at breakpoint"
}
```

Error responses:

```json
{
  "success": false,
  "command": "break",
  "error": "could not find file /path/to/nonexistent.go"
}
```

## AI Agent Example

```bash
# AI extracts 'addr' from response and tracks it
RESPONSE=$(godebug start ./cmd/server)
ADDR=$(echo $RESPONSE | jq -r '.data.addr')

# AI uses --addr for all commands
godebug --addr $ADDR break handlers/user.go:42
godebug --addr $ADDR continue
godebug --addr $ADDR locals
godebug --addr $ADDR eval "user.Name"
godebug --addr $ADDR stack
godebug --addr $ADDR quit
```

## Start Modes

```bash
# Debug mode (default) - compile and debug
godebug start ./cmd/myapp

# Test mode - debug tests
godebug start --mode test ./...

# Exec mode - debug pre-compiled binary
godebug start --mode exec ./myapp

# With program arguments
godebug start ./cmd/myapp -- -port 8080 -config prod.yaml
```

## Development

### Prerequisites

- [Task](https://taskfile.dev) (optional, for task runner)
- [golangci-lint](https://golangci-lint.run) (for linting)

### Available Tasks

```bash
task              # Build the binary
task build        # Build optimized binary to ./bin/godebug
task build:dev    # Build with debug symbols
task test         # Run tests with race detection
task test:cover   # Run tests with coverage
task lint         # Run golangci-lint
task lint:fix     # Run linter with auto-fix
task fmt          # Format code
task verify       # Run tidy + lint + test + build
task clean        # Remove build artifacts
```

### Project Structure

```
godebug/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Cobra root, --addr/--output flags
│   ├── start.go                # Start debug session
│   ├── connect.go              # Connect to existing server
│   ├── status.go               # Show state
│   ├── quit.go                 # Stop debugging
│   ├── breakpoint.go           # break, clear, breakpoints
│   ├── execution.go            # continue, step, next, stepout, restart
│   ├── inspect.go              # locals, args, eval
│   ├── navigation.go           # stack, frame, goroutines, goroutine
│   └── source.go               # list, sources
├── internal/
│   ├── debugger/
│   │   ├── client.go           # Delve RPC2 client wrapper
│   │   └── launcher.go         # Spawns dlv headless
│   └── output/
│       └── response.go         # JSON response envelope
└── testdata/
    └── debugme/                # Test application
```

## Why godebug?

### Static vs Dynamic Analysis

Most AI coding tools rely on **static analysis**—reading the code and "hallucinating" the execution flow based on patterns.

| Approach | What the AI does |
|----------|------------------|
| **Without Debugger** | Guesses what *should* happen |
| **With Debugger** | Sees what *is* happening |

In Go, where concurrency is a first-class citizen, reading the code is often insufficient:

- **Race conditions** depend on timing that can't be determined statically
- **Deadlocks** emerge from lock ordering that spans multiple goroutines
- **Channel behavior** depends on buffer sizes and goroutine scheduling
- **Context cancellation** propagates in ways that require runtime observation

Static analysis produces high false positives/negatives for Go concurrency. A debugger lets the AI observe actual goroutine states, lock contention, and channel operations as they happen.

```bash
# AI debugging a suspected race condition
godebug --addr $ADDR goroutines                    # List all goroutines
godebug --addr $ADDR goroutine 7                   # Switch to worker goroutine
godebug --addr $ADDR locals                        # See its local state
godebug --addr $ADDR eval "len(ch)"                # Check channel buffer
godebug --addr $ADDR break sync.(*Mutex).Lock      # Break on lock acquisition
godebug --addr $ADDR continue                      # Run until lock contention
godebug --addr $ADDR stack                         # See who's holding the lock
```

### Surgical Data Extraction

With `eval`, an AI can extract exactly the data it needs without log spam:

```bash
# Instead of adding print statements and re-running...
godebug --addr $ADDR eval "myStruct.InnerField.Map[\"key\"]"
godebug --addr $ADDR eval "len(users)"
godebug --addr $ADDR eval "err.Error()"
godebug --addr $ADDR eval "ctx.Err()"
```

**Benefits:**
- **No code modification**: Inspect any expression without adding print statements
- **Token efficient**: Fetch only the relevant data, not entire log dumps
- **Iterative exploration**: Drill down into nested structures on demand

Traditional debugging workflow with logs:
```
1. AI guesses what to log
2. Adds print statements
3. User rebuilds and runs
4. AI reads massive log output (thousands of tokens)
5. Repeat until found
```

With `godebug`:
```
1. AI sets breakpoint at crash site
2. Evaluates specific expressions (tens of tokens each)
3. Finds root cause in one session
```

### Context Window Efficiency

LLMs have limited context windows. Dumping entire log files or massive codebases into the prompt:
- Fills up the context quickly
- Buries relevant information in noise
- Degrades reasoning quality

`godebug` returns **focused, structured data**:

```json
{
  "success": true,
  "command": "eval",
  "data": {
    "name": "user.Profile.Settings[\"theme\"]",
    "type": "string",
    "value": "\"dark\""
  }
}
```

Compare context usage:

| Approach | Tokens | Signal-to-noise |
|----------|--------|-----------------|
| Dump 1000-line log | ~4000 | Low (mostly irrelevant) |
| Dump entire file | ~2000 | Medium (need to find the line) |
| `godebug eval` | ~50 | High (exactly what was asked) |
| `godebug locals` | ~200 | High (current scope only) |

The AI stays focused on the problem instead of parsing through noise.

### Design Principles

- **Stateless**: No session files or hidden state—just pass `--addr`
- **JSON output**: Structured responses for programmatic consumption
- **Single command**: Each invocation is independent, perfect for AI agents
- **Full Delve power**: All debugging capabilities through a clean CLI

## License

MIT

## Related Projects

- [Delve](https://github.com/go-delve/delve) - The underlying Go debugger
- [dlv-mcp](https://github.com/orisano/dlv-mcp) - MCP server for Delve
