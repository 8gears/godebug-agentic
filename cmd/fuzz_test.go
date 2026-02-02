package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/8gears/godebug/internal/output"
	"pgregory.net/rapid"
)

// setupFuzzTest prevents os.Exit from killing tests by replacing ExitFunc
// with a function that panics instead. The panic is caught by runCLI.
func setupFuzzTest(t testing.TB) {
	t.Helper()
	originalExit := output.ExitFunc
	output.ExitFunc = func(code int) {
		// Don't exit - panic with a recognizable format that runCLI can catch
		panic(fmt.Sprintf("exit:%d", code))
	}
	t.Cleanup(func() {
		output.ExitFunc = originalExit
	})
}

// runCLI executes the CLI with the given args and recovers from exit panics.
// Returns the exit code (if exit was called) and whether a real panic occurred.
func runCLI(args []string) (exitCode int, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok && len(s) > 5 && s[:5] == "exit:" {
				fmt.Sscanf(s, "exit:%d", &exitCode)
				return
			}
			panicked = true // Real panic, not our exit replacement
		}
	}()

	cmd := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		return 1, false
	}
	return 0, false
}

// TestFuzzRandomArgs tests with completely unbiased random arguments.
// This covers cases where the CLI receives arbitrary string input.
func TestFuzzRandomArgs(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Random number of args (0-15)
		numArgs := rapid.IntRange(0, 15).Draw(t, "numArgs")
		args := make([]string, numArgs)
		for i := range args {
			args[i] = rapid.String().Draw(t, fmt.Sprintf("arg%d", i))
		}

		_, panicked := runCLI(args)
		if panicked {
			t.Fatalf("CLI panicked with args: %v", args)
		}
	})
}

// TestFuzzWithBrokenJSON tests with random strings that look like broken JSON.
// This helps find edge cases in JSON parsing or handling.
func TestFuzzWithBrokenJSON(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		brokenJSON := rapid.OneOf(
			rapid.Just("{"),
			rapid.Just(`{"key":`),
			rapid.Just(`{"key": "value"`),
			rapid.Just(`[1, 2, 3`),
			rapid.Just(`{"nested": {"broken"`),
			rapid.StringMatching(`\{[^}]*`),  // Opening brace, random, no close
			rapid.StringMatching(`"[^"]*`),   // Unclosed string
			rapid.Map(rapid.String(), func(s string) string {
				return "{" + s // Prefix with brace
			}),
		).Draw(t, "brokenJSON")

		// Test as argument to various commands
		testCases := [][]string{
			{"eval", brokenJSON},
			{"break", brokenJSON},
			{"--addr", brokenJSON, "status"},
		}

		for _, args := range testCases {
			_, panicked := runCLI(args)
			if panicked {
				t.Fatalf("CLI panicked with broken JSON args: %v", args)
			}
		}
	})
}

// TestFuzzWithValidJSON tests with valid JSON as arguments.
// This helps find issues when valid JSON is passed as CLI arguments.
func TestFuzzWithValidJSON(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Generate random valid JSON
		validJSON := generateRandomJSON(t)

		testCases := [][]string{
			{"eval", validJSON},
			{"break", validJSON},
			{"--addr", validJSON, "status"},
		}

		for _, args := range testCases {
			_, panicked := runCLI(args)
			if panicked {
				t.Fatalf("CLI panicked with valid JSON args: %v", args)
			}
		}
	})
}

// generateFuzzTimeout generates a timeout value for fuzz testing.
// Valid timeouts are capped at 100ms to keep tests fast.
// Also includes invalid values to test error handling.
func generateFuzzTimeout(t *rapid.T) string {
	return rapid.OneOf(
		// Valid short durations for fast tests (1-100ms)
		rapid.SampledFrom([]string{"1ms", "5ms", "10ms", "50ms", "100ms"}),
		// Valid durations with different units (capped at 100ms)
		rapid.Map(rapid.IntRange(1, 100), func(ms int) string {
			return fmt.Sprintf("%dms", ms)
		}),
		// Invalid durations (for error handling testing)
		rapid.SampledFrom([]string{"", "invalid", "-1s", "abc", "0", "3.5.2s"}),
	).Draw(t, "timeout")
}

// generateRandomJSON creates a random valid JSON value.
func generateRandomJSON(t *rapid.T) string {
	jsonType := rapid.IntRange(0, 4).Draw(t, "jsonType")
	switch jsonType {
	case 0: // null
		return "null"
	case 1: // bool
		if rapid.Bool().Draw(t, "bool") {
			return "true"
		}
		return "false"
	case 2: // number
		return fmt.Sprintf("%d", rapid.Int().Draw(t, "num"))
	case 3: // string
		s := rapid.String().Draw(t, "str")
		b, _ := json.Marshal(s)
		return string(b)
	default: // object
		key := rapid.StringMatching(`[a-z]{1,10}`).Draw(t, "key")
		val := rapid.String().Draw(t, "val")
		b, _ := json.Marshal(map[string]string{key: val})
		return string(b)
	}
}

// TestFuzzAddressVariations tests with various --addr flag formats.
// This covers missing addr, valid formats, invalid formats, empty values, etc.
func TestFuzzAddressVariations(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Commands that typically need --addr
		command := rapid.SampledFrom([]string{
			"status", "continue", "next", "step", "stepout",
			"locals", "args", "stack", "goroutines", "list",
			"breakpoints", "restart", "quit",
		}).Draw(t, "command")

		addrMode := rapid.IntRange(0, 5).Draw(t, "addrMode")

		var args []string
		switch addrMode {
		case 0: // No addr at all
			args = []string{command}
		case 1: // Valid-looking addr
			host := rapid.SampledFrom([]string{
				"localhost", "127.0.0.1", "0.0.0.0", "::1",
			}).Draw(t, "host")
			port := rapid.IntRange(1, 65535).Draw(t, "port")
			args = []string{"--addr", fmt.Sprintf("%s:%d", host, port), command}
		case 2: // Invalid addr format
			garbage := rapid.String().Draw(t, "garbage")
			args = []string{"--addr", garbage, command}
		case 3: // Empty addr
			args = []string{"--addr", "", command}
		case 4: // Missing port
			args = []string{"--addr", "localhost", command}
		case 5: // Only port
			port := rapid.IntRange(1, 65535).Draw(t, "port")
			args = []string{"--addr", fmt.Sprintf(":%d", port), command}
		}

		_, panicked := runCLI(args)
		if panicked {
			t.Fatalf("CLI panicked with addr variation: %v", args)
		}
	})
}

// TestFuzzBreakpointLocations tests with random breakpoint location formats.
// This covers file:line, package.Function, and various invalid formats.
func TestFuzzBreakpointLocations(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		location := rapid.OneOf(
			// Completely random
			rapid.String(),
			// File:line-ish
			rapid.StringMatching(`[a-z./]+:[0-9]+`),
			// Package.Func-ish
			rapid.StringMatching(`[a-z]+\.[A-Z][a-z]+`),
			// Broken patterns
			rapid.Just(":"),
			rapid.Just("::"),
			rapid.Just("."),
			rapid.Just(".."),
			rapid.StringMatching(`[^a-zA-Z0-9]+`), // Only special chars
		).Draw(t, "location")

		_, panicked := runCLI([]string{"--addr", "localhost:12345", "break", location})
		if panicked {
			t.Fatalf("CLI panicked with breakpoint location: %s", location)
		}
	})
}

// TestFuzzEvalExpressions tests with random eval expressions.
// This covers field access, index access, arithmetic, and random strings.
func TestFuzzEvalExpressions(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		expr := rapid.OneOf(
			rapid.String(),                                // Totally random
			rapid.StringMatching(`[a-z]+\.[A-Z][a-z]+`),   // Field access
			rapid.StringMatching(`[a-z]+\[[0-9]+\]`),      // Index access
			rapid.StringMatching(`[a-z]+ [+\-*/] [0-9]+`), // Arithmetic
		).Draw(t, "expr")

		_, panicked := runCLI([]string{"--addr", "localhost:12345", "eval", expr})
		if panicked {
			t.Fatalf("CLI panicked with eval expression: %s", expr)
		}
	})
}

// TestFuzzAllCommands tests every command with random args and flags.
// This is the most comprehensive test, exercising all commands with various inputs.
func TestFuzzAllCommands(t *testing.T) {
	setupFuzzTest(t)

	commands := []string{
		"start", "connect", "status", "restart", "quit",
		"break", "clear", "breakpoints",
		"continue", "next", "step", "stepout",
		"locals", "args", "eval",
		"stack", "frame", "goroutines", "goroutine",
		"list", "sources",
	}

	rapid.Check(t, func(t *rapid.T) {
		command := rapid.SampledFrom(commands).Draw(t, "command")

		// Random extra args (0-5)
		numExtra := rapid.IntRange(0, 5).Draw(t, "numExtra")
		args := []string{command}
		for i := 0; i < numExtra; i++ {
			args = append(args, rapid.String().Draw(t, fmt.Sprintf("extra%d", i)))
		}

		// Randomly add --addr or not
		if rapid.Bool().Draw(t, "hasAddr") {
			addr := rapid.String().Draw(t, "addr")
			args = append([]string{"--addr", addr}, args...)
		}

		// Randomly add --output or not
		if rapid.Bool().Draw(t, "hasOutput") {
			outputVal := rapid.String().Draw(t, "output")
			args = append([]string{"--output", outputVal}, args...)
		}

		// Randomly add --timeout or not (capped at 3s to prevent slow tests)
		if rapid.Bool().Draw(t, "hasTimeout") {
			timeout := generateFuzzTimeout(t)
			args = append([]string{"--timeout", timeout}, args...)
		}

		_, panicked := runCLI(args)
		if panicked {
			t.Fatalf("CLI panicked with args: %v", args)
		}
	})
}

// TestFuzzUnicodeAndSpecialChars tests with Unicode and special character inputs.
// This helps find encoding issues or character handling problems.
func TestFuzzUnicodeAndSpecialChars(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Generate strings with specific character classes
		specialInput := rapid.OneOf(
			// Unicode strings (using string matching for unicode ranges)
			rapid.StringMatching(`[\x00-\x7F]+`),
			// Strings with control characters
			rapid.StringMatching(`[\x00-\x1F\x7F]+`),
			// Strings with quotes and escapes
			rapid.SampledFrom([]string{`"`, `'`, `\`, "`", `$`, `"test"`, `'test'`, `\n`, `$HOME`}),
			// Strings with shell metacharacters
			rapid.SampledFrom([]string{"|", "&", ";", "<", ">", "()", "{}", "[]", "*", "?", "!", "| cat", "&& ls", "; rm"}),
			// Empty and whitespace
			rapid.Just(""),
			rapid.SampledFrom([]string{" ", "\t", "\n", "\r", "  ", "\t\t", " \n "}),
		).Draw(t, "specialInput")

		testCases := [][]string{
			{"eval", specialInput},
			{"break", specialInput},
			{"--addr", specialInput, "status"},
			{"sources", specialInput},
			{"frame", specialInput},
			{"goroutine", specialInput},
			{"clear", specialInput},
		}

		for _, args := range testCases {
			_, panicked := runCLI(args)
			if panicked {
				t.Fatalf("CLI panicked with special chars in args: %v", args)
			}
		}
	})
}

// TestFuzzLongStrings tests with very long string inputs.
// This helps find buffer overflow or memory issues.
func TestFuzzLongStrings(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Generate strings of various lengths
		length := rapid.OneOf(
			rapid.Just(0),
			rapid.Just(1),
			rapid.Just(100),
			rapid.Just(1000),
			rapid.Just(10000),
			rapid.IntRange(0, 100000),
		).Draw(t, "length")

		longString := rapid.StringOfN(rapid.Rune(), length, length, length).Draw(t, "longString")

		testCases := [][]string{
			{"eval", longString},
			{"break", longString},
			{"--addr", longString, "status"},
		}

		for _, args := range testCases {
			_, panicked := runCLI(args)
			if panicked {
				t.Fatalf("CLI panicked with long string (len=%d) in args: %v", length, args)
			}
		}
	})
}

// TestFuzzNumericInputs tests commands that expect numeric inputs.
// This covers frame, goroutine, clear, and other commands with numeric args.
func TestFuzzNumericInputs(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		// Various numeric representations
		numericInput := rapid.OneOf(
			// Valid integers
			rapid.Map(rapid.Int(), func(i int) string { return fmt.Sprintf("%d", i) }),
			// Float-like
			rapid.Map(rapid.Float64(), func(f float64) string { return fmt.Sprintf("%f", f) }),
			// Hex
			rapid.Map(rapid.Int(), func(i int) string { return fmt.Sprintf("0x%x", i) }),
			// Octal
			rapid.Map(rapid.IntRange(0, 1000), func(i int) string { return fmt.Sprintf("0%o", i) }),
			// Scientific notation
			rapid.Map(rapid.Float64(), func(f float64) string { return fmt.Sprintf("%e", f) }),
			// Negative
			rapid.Map(rapid.Int(), func(i int) string { return fmt.Sprintf("-%d", i) }),
			// Invalid
			rapid.Just("NaN"),
			rapid.Just("Inf"),
			rapid.Just("-Inf"),
			rapid.Just(""),
			rapid.String(),
		).Draw(t, "numericInput")

		testCases := [][]string{
			{"--addr", "localhost:12345", "frame", numericInput},
			{"--addr", "localhost:12345", "goroutine", numericInput},
			{"--addr", "localhost:12345", "clear", numericInput},
		}

		for _, args := range testCases {
			_, panicked := runCLI(args)
			if panicked {
				t.Fatalf("CLI panicked with numeric input: %v", args)
			}
		}
	})
}

// TestFuzzFlagCombinations tests various combinations of global flags.
// This helps find issues with flag parsing and validation.
func TestFuzzFlagCombinations(t *testing.T) {
	setupFuzzTest(t)

	rapid.Check(t, func(t *rapid.T) {
		var args []string

		// Randomly include --addr
		if rapid.Bool().Draw(t, "include_addr") {
			args = append(args, "--addr")
			if rapid.Bool().Draw(t, "addr_hasValue") {
				args = append(args, rapid.String().Draw(t, "addr_value"))
			}
		}

		// Randomly include --output
		if rapid.Bool().Draw(t, "include_output") {
			args = append(args, "--output")
			if rapid.Bool().Draw(t, "output_hasValue") {
				args = append(args, rapid.String().Draw(t, "output_value"))
			}
		}

		// Randomly include --timeout (capped at 3s to prevent slow tests)
		if rapid.Bool().Draw(t, "include_timeout") {
			args = append(args, "--timeout")
			if rapid.Bool().Draw(t, "timeout_hasValue") {
				args = append(args, generateFuzzTimeout(t))
			}
		}

		// Add a command
		commands := []string{"status", "continue", "locals", "stack", "breakpoints"}
		args = append(args, rapid.SampledFrom(commands).Draw(t, "command"))

		_, panicked := runCLI(args)
		if panicked {
			t.Fatalf("CLI panicked with flag combinations: %v", args)
		}
	})
}
