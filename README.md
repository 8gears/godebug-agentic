# godebug agentic

A stateless CLI debugger for Go applications, designed for AI agent tool calling. Enables runtime verification of application behavior beyond what logs and tests can provide. See [Beyond Tests: Runtime Verification for AI Agents](docs/ai-debugger-verification.md) for the paradigm shift from stochastic generation to deterministic engineering. You can also use godebug for traditional debugging with AI agents. 

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
│  godebug start ./app    │  godebug break main.go:42         │
│  godebug continue       │  godebug locals                   │
│  (each command connects, executes, outputs JSON, exits)     │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.24+
- [Delve](https://github.com/go-delve/delve) debugger (`go install github.com/go-delve/delve/cmd/dlv@latest`)

### Install

```bash
go install github.com/8gears/godebug-agentic@latest
```

## Quick Start

```bash
# Start debug session (returns JSON with server address)
godebug start ./cmd/myapp
# {"data":{"addr":"127.0.0.1:38697",...}}

# Use --addr for all subsequent commands
godebug --addr 127.0.0.1:38697 break main.go:42
godebug --addr 127.0.0.1:38697 continue
godebug --addr 127.0.0.1:38697 locals
godebug --addr 127.0.0.1:38697 quit
```

## Documentation

**For complete command reference, examples, and workflows, see the Claude Code skill:**

📖 [**.claude/skills/godebug/SKILL.md**](.claude/skills/godebug/SKILL.md)

The skill documents all 21 commands with verified JSON output examples:
- Session management: `start`, `connect`, `quit`, `status`, `restart`
- Breakpoints: `break` (with `--cond` for conditionals), `clear`, `breakpoints`
- Execution: `continue`, `next`, `step`, `stepout`
- Inspection: `locals`, `args`, `eval`
- Navigation: `stack`, `frame`, `goroutines`, `goroutine`
- Source: `list`, `sources`

## Why Debug?

### Traditional Debugging

AI agents can use `godebug` for classic debugging workflows—finding and fixing bugs:

- **Set breakpoints** at suspicious locations
- **Step through** code line by line
- **Inspect variables** to understand state
- **Examine call stacks** to trace execution flow
- **Debug goroutines** to diagnose concurrency issues

In Go, where concurrency is a first-class citizen, this is especially valuable:

- **Race conditions** depend on timing that can't be determined statically
- **Deadlocks** emerge from lock ordering that spans multiple goroutines
- **Channel behavior** depends on buffer sizes and goroutine scheduling
- **Context cancellation** propagates in ways that require runtime observation

### Runtime Verification (Beyond Tests)

Beyond bug-finding, debuggers enable a new paradigm: **verifying application behavior** that logs and tests cannot capture.

| Verification Method | What It Provides | What It Misses |
|---------------------|------------------|----------------|
| **Logs** | What code chose to report | Everything between log statements |
| **Unit Tests** | Pass/fail for predicted scenarios | Intermediate states, timing, actual paths |
| **Debugger** | Actual runtime state | Nothing—direct observation |

The core insight: **tests verify outcomes, debuggers verify behavior**.

An AI agent that can inspect runtime state doesn't guess—it knows. This transforms code generation from stochastic output to verified engineering.

For the full paradigm shift, see [Beyond Tests: Runtime Verification for AI Agents](docs/ai-debugger-verification.md).

### Surgical Data Extraction

With `eval`, an AI can extract exactly the data it needs without log spam:

```bash
# Instead of adding print statements and re-running...
godebug --addr $ADDR eval "myStruct.InnerField.Map[\"key\"]"
godebug --addr $ADDR eval "len(users)"
godebug --addr $ADDR eval "err.Error()"
```

**Benefits:**
- **No code modification**: Inspect any expression without adding print statements
- **Token efficient**: Fetch only the relevant data, not entire log dumps
- **Iterative exploration**: Drill down into nested structures on demand

### Context Window Efficiency

LLMs have limited context windows. `godebug` returns **focused, structured data**:

| Approach | Tokens | Signal-to-noise |
|----------|--------|-----------------|
| Dump 1000-line log | ~4000 | Low (mostly irrelevant) |
| Dump entire file | ~2000 | Medium (need to find the line) |
| `godebug eval` | ~50 | High (exactly what was asked) |
| `godebug locals` | ~200 | High (current scope only) |

### Design Principles

- **Stateless**: No session files or hidden state—just pass `--addr`
- **JSON output**: Structured responses for programmatic consumption
- **Single command**: Each invocation is independent, perfect for AI agents
- **Full Delve power**: All debugging capabilities through a clean CLI

## Example: Debugging a Concurrency Bug

This example shows how to debug a WaitGroup race condition in `testdata/concurrency_bugs/waitgroup_race`.

### The Bug

```go
// BUGGY: Add() races with Wait()
for i := 0; i < 10; i++ {
    go func(id int) {
        wg.Add(1)  // May execute AFTER Wait() returns
        defer wg.Done()
    }(i)
}
wg.Wait()  // Returns immediately if no Add() called yet
```

### Debug Session

```bash
# Start debug session and capture address
ADDR=$(godebug start ./testdata/concurrency_bugs/waitgroup_race | jq -r '.data.addr')

# Set breakpoints
godebug --addr $ADDR break main.go:18  # wg.Add(1)
godebug --addr $ADDR break main.go:27  # wg.Wait()

# Continue - hits Wait() FIRST (proving the race)
godebug --addr $ADDR continue
# {"data":{"location":{"line":27}}}  <- Wait() reached before Add()!

# Check WaitGroup state
godebug --addr $ADDR locals
# wg.state.v = 0  <- No Add() called yet!

# Check goroutines
godebug --addr $ADDR goroutines
# Main at Wait() (line 27), workers still at lines 17-18

godebug --addr $ADDR quit
```

**Finding:** Main goroutine reaches `Wait()` before any worker calls `Add()`, so `Wait()` returns immediately.

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
task verify       # Run tidy + lint + test + build
task clean        # Remove build artifacts
```

### Project Structure

```
godebug-agentic/
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Cobra root, --addr/--output flags
│   ├── start.go                # Start debug session
│   ├── connect.go              # Connect to existing server
│   ├── quit.go                 # Quit debug session
│   ├── status.go               # Check server status
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
│       ├── response.go         # JSON response envelope
│       ├── errors.go           # Error types and handling
│       └── exitcodes.go        # CLI exit codes
├── testdata/
│   ├── debugme/                # Basic test application
│   └── concurrency_bugs/       # Concurrency bug examples
└── .claude/
    └── skills/godebug/         # Claude Code skill documentation
```

### Test Data: Concurrency Bug Examples

The `testdata/concurrency_bugs/` directory contains intentionally buggy Go programs for practicing debugging:

| Example | Bug Type |
|---------|----------|
| `waitgroup_race/` | WaitGroup Add/Wait race condition |
| `race_counter/` | Unsynchronized counter increment |
| `deadlock_circular/` | Circular lock ordering deadlock |
| `closure_loop/` | Loop variable capture in closures |
| `mutex_copy/` | Mutex copied via value receiver |
| `channel_nil/` | Operations on nil channels |
| `leak_forgotten_sender/` | Goroutine leak from abandoned channel send |
| `select_timeout_leak/` | Timer leak from `time.After` in loops |

```bash
# Build all examples
task build:examples

# Debug an example
godebug start ./testdata/concurrency_bugs/waitgroup_race
```

## License

MIT

## Related Projects

- [Delve](https://github.com/go-delve/delve) - The underlying Go debugger
- [dlv-mcp](https://github.com/orisano/dlv-mcp) - MCP server for Delve
