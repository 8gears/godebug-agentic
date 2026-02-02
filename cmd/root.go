package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-delve/delve/service/api"
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/debugger"
	"github.com/8gears/godebug/internal/output"
)

var (
	// Global flags
	addr         string
	outputFormat string
	timeout      time.Duration

	// Shared client (initialized per command if --addr is provided)
	client *debugger.Client
)

// GetOutputFormat returns the current output format
func GetOutputFormat() output.OutputFormat {
	if outputFormat == "text" {
		return output.FormatText
	}
	return output.FormatJSON
}

// GetTimeout returns the configured operation timeout
func GetTimeout() time.Duration {
	return timeout
}

// GetClient returns the debug client, connecting if necessary
func GetClient() (*debugger.Client, error) {
	if client != nil {
		return client, nil
	}
	var err error
	client, err = debugger.Connect(addr)
	return client, err
}

// MustGetClient returns the client or exits with error
func MustGetClient(cmdName string) *debugger.Client {
	if addr == "" {
		output.ErrorWithInfo(cmdName, output.InvalidArgument("--addr flag is required")).PrintAndExit(GetOutputFormat())
	}
	c, err := GetClient()
	if err != nil {
		output.Error(cmdName, err).PrintAndExit(GetOutputFormat())
	}
	return c
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "godebug",
	Short: "AI-friendly Go debugger CLI using Delve",
	Long: `godebug is a single-command CLI debugger for Go applications.
Each invocation runs one command, outputs structured JSON, and exits.
Designed for AI agent tool calling.

Start a debug session:
  godebug start ./myapp

Then use --addr with all subsequent commands:
  godebug --addr 127.0.0.1:38697 break main.go:42
  godebug --addr 127.0.0.1:38697 continue
  godebug --addr 127.0.0.1:38697 locals`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(output.ExitGenericError)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "", "Delve server address (host:port)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "Output format: json or text")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "Operation timeout (e.g., 10s, 1m, 30s)")
}

// NewRootCmd creates a fresh root command for testing.
// Each call returns an isolated command with its own flag state,
// avoiding the shared global state that would persist between tests.
func NewRootCmd() *cobra.Command {
	// Fresh state for this command instance
	var cmdAddr string
	var cmdOutputFormat string
	var cmdTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "godebug",
		Short: "AI-friendly Go debugger CLI using Delve",
		Long: `godebug is a single-command CLI debugger for Go applications.
Each invocation runs one command, outputs structured JSON, and exits.
Designed for AI agent tool calling.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&cmdAddr, "addr", "", "Delve server address (host:port)")
	cmd.PersistentFlags().StringVar(&cmdOutputFormat, "output", "json", "Output format: json or text")
	cmd.PersistentFlags().DurationVar(&cmdTimeout, "timeout", 30*time.Second, "Operation timeout (e.g., 10s, 1m, 30s)")

	// Helper functions for this command's context
	getOutputFormat := func() output.OutputFormat {
		if cmdOutputFormat == "text" {
			return output.FormatText
		}
		return output.FormatJSON
	}

	getTimeout := func() time.Duration {
		return cmdTimeout
	}

	mustGetClient := func(cmdName string) *debugger.Client {
		if cmdAddr == "" {
			output.ErrorWithInfo(cmdName, output.InvalidArgument("--addr flag is required")).PrintAndExit(getOutputFormat())
		}
		c, err := debugger.Connect(cmdAddr)
		if err != nil {
			output.Error(cmdName, err).PrintAndExit(getOutputFormat())
		}
		return c
	}

	// Add all subcommands with fresh state
	addStartCommand(cmd, getOutputFormat, getTimeout)
	addConnectCommand(cmd, getOutputFormat)
	addStatusCommand(cmd, mustGetClient, getOutputFormat)
	addExecutionCommands(cmd, mustGetClient, getOutputFormat, getTimeout)
	addBreakpointCommands(cmd, mustGetClient, getOutputFormat)
	addInspectCommands(cmd, mustGetClient, getOutputFormat)
	addNavigationCommands(cmd, mustGetClient, getOutputFormat)
	addSourceCommands(cmd, mustGetClient, getOutputFormat)
	addQuitCommand(cmd, mustGetClient, getOutputFormat)

	return cmd
}

// addStartCommand adds the start command to the root
func addStartCommand(root *cobra.Command, getOutputFormat func() output.OutputFormat, getTimeout func() time.Duration) {
	var startMode string

	startCmd := &cobra.Command{
		Use:   "start [target]",
		Short: "Start a debug session",
		Long: `Start a Delve debug server for a Go application.

Modes:
  debug (default) - Compile and debug a Go package
  test            - Compile and debug tests
  exec            - Debug a pre-compiled binary

Examples:
  godebug start ./cmd/myapp           # Debug mode (default)
  godebug start --mode test ./...     # Test mode
  godebug start --mode exec ./binary  # Exec mode
  godebug start ./cmd/myapp -- -port 8080  # With program args`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target := args[0]
			var programArgs []string
			if cmd.ArgsLenAtDash() > 0 {
				programArgs = args[cmd.ArgsLenAtDash():]
			}

			mode := debugger.ModeDebug
			switch startMode {
			case "test":
				mode = debugger.ModeTest
			case "exec":
				mode = debugger.ModeExec
			}

			config := debugger.LaunchConfig{
				Mode:    mode,
				Target:  target,
				Args:    programArgs,
				Timeout: getTimeout(),
			}

			result, err := debugger.Launch(config)
			if err != nil {
				output.Error("start", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"addr":   result.Addr,
				"pid":    result.PID,
				"target": result.Target,
				"mode":   result.Mode,
			}

			output.Success("start", data, "Debug server started").PrintAndExit(getOutputFormat())
		},
	}

	startCmd.Flags().StringVar(&startMode, "mode", "debug", "Debug mode: debug, test, or exec")
	root.AddCommand(startCmd)
}

// addConnectCommand adds the connect command to the root
func addConnectCommand(root *cobra.Command, getOutputFormat func() output.OutputFormat) {
	connectCmd := &cobra.Command{
		Use:   "connect <addr>",
		Short: "Connect to an existing Delve server",
		Long: `Connect to an existing Delve debug server.

This is useful for remote debugging or attaching to a manually started Delve server.

Example:
  dlv debug ./myapp --headless --api-version=2 --listen=:38697
  godebug connect localhost:38697`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serverAddr := args[0]

			c, err := debugger.Connect(serverAddr)
			if err != nil {
				output.Error("connect", err).PrintAndExit(getOutputFormat())
			}
			defer func() { _ = c.Close() }()

			// Verify connection by getting state
			state, err := c.GetState()
			if err != nil {
				output.Error("connect", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"addr":    serverAddr,
				"running": state.Running,
			}
			if state.SelectedGoroutine != nil {
				data["goroutineId"] = state.SelectedGoroutine.ID
			}

			output.Success("connect", data, "Connected to debug server").PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(connectCmd)
}

// addStatusCommand adds the status command to the root
func addStatusCommand(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current debug state",
		Long: `Show the current state of the debug session.

Returns whether the process is running, paused, or exited,
along with the current location if paused.

Example:
  godebug --addr 127.0.0.1:38697 status`,
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("status")
			defer func() { _ = c.Close() }()

			state, err := c.GetState()
			if err != nil {
				output.Error("status", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"running": state.Running,
				"exited":  state.Exited,
			}

			if state.Exited {
				data["exitStatus"] = state.ExitStatus
			}

			if state.SelectedGoroutine != nil {
				g := state.SelectedGoroutine
				data["goroutine"] = map[string]any{
					"id": g.ID,
				}
				if g.CurrentLoc.File != "" {
					data["location"] = map[string]any{
						"file":     g.CurrentLoc.File,
						"line":     g.CurrentLoc.Line,
						"function": g.CurrentLoc.Function.Name(),
					}
				}
			}

			var msg string
			if state.Exited {
				msg = "Process exited"
			} else if state.Running {
				msg = "Process running"
			} else {
				msg = "Process paused"
			}

			output.Success("status", data, msg).PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(statusCmd)
}

// addExecutionCommands adds execution control commands (continue, next, step, etc.)
func addExecutionCommands(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat, getTimeout func() time.Duration) {
	// continue
	continueCmd := &cobra.Command{
		Use:   "continue",
		Short: "Continue execution until breakpoint",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("continue")
			defer func() { _ = c.Close() }()
			c.SetTimeout(getTimeout())

			state, err := c.Continue()
			if err != nil {
				output.Error("continue", err).PrintAndExit(getOutputFormat())
			}

			var msg string
			if state.Exited {
				msg = "Process exited"
			} else if state.CurrentThread != nil && state.CurrentThread.Breakpoint != nil {
				msg = "Stopped at breakpoint"
			} else {
				msg = "Process stopped"
			}

			output.Success("continue", stateToData(state), msg).PrintAndExit(getOutputFormat())
		},
	}

	// next
	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "Step over to next source line",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("next")
			defer func() { _ = c.Close() }()
			c.SetTimeout(getTimeout())

			state, err := c.Next()
			if err != nil {
				output.Error("next", err).PrintAndExit(getOutputFormat())
			}

			output.Success("next", stateToData(state), "Stepped to next line").PrintAndExit(getOutputFormat())
		},
	}

	// step
	stepCmd := &cobra.Command{
		Use:   "step",
		Short: "Step into function call",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("step")
			defer func() { _ = c.Close() }()
			c.SetTimeout(getTimeout())

			state, err := c.Step()
			if err != nil {
				output.Error("step", err).PrintAndExit(getOutputFormat())
			}

			output.Success("step", stateToData(state), "Stepped into function").PrintAndExit(getOutputFormat())
		},
	}

	// stepout
	stepoutCmd := &cobra.Command{
		Use:   "stepout",
		Short: "Step out of current function",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("stepout")
			defer func() { _ = c.Close() }()
			c.SetTimeout(getTimeout())

			state, err := c.StepOut()
			if err != nil {
				output.Error("stepout", err).PrintAndExit(getOutputFormat())
			}

			output.Success("stepout", stateToData(state), "Stepped out of function").PrintAndExit(getOutputFormat())
		},
	}

	// restart
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the debugged program",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("restart")
			defer func() { _ = c.Close() }()

			state, err := c.Restart()
			if err != nil {
				output.Error("restart", err).PrintAndExit(getOutputFormat())
			}

			output.Success("restart", stateToData(state), "Program restarted").PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(continueCmd)
	root.AddCommand(nextCmd)
	root.AddCommand(stepCmd)
	root.AddCommand(stepoutCmd)
	root.AddCommand(restartCmd)
}

// addBreakpointCommands adds breakpoint management commands
func addBreakpointCommands(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	var breakCond string

	// break
	breakCmd := &cobra.Command{
		Use:   "break <location>",
		Short: "Set a breakpoint",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("break")
			defer func() { _ = c.Close() }()

			location := args[0]
			bp := &api.Breakpoint{}

			// Parse location: file:line or function name
			if strings.Contains(location, ":") {
				parts := strings.SplitN(location, ":", 2)
				file := parts[0]
				line, err := strconv.Atoi(parts[1])
				if err != nil {
					output.ErrorWithInfo("break", output.InvalidArgumentWithDetails(
						fmt.Sprintf("invalid line number: %s", parts[1]),
						map[string]any{"location": location, "line": parts[1]},
					)).PrintAndExit(getOutputFormat())
				}
				// Convert to absolute path if relative
				if !filepath.IsAbs(file) {
					absPath, err := filepath.Abs(file)
					if err == nil {
						file = absPath
					}
				}
				bp.File = file
				bp.Line = line
			} else {
				bp.FunctionName = location
			}

			// Add condition if specified
			if breakCond != "" {
				bp.Cond = breakCond
			}

			created, err := c.CreateBreakpoint(bp)
			if err != nil {
				output.Error("break", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"id":       created.ID,
				"file":     created.File,
				"line":     created.Line,
				"function": created.FunctionName,
			}
			if created.Cond != "" {
				data["condition"] = created.Cond
			}

			output.Success("break", data, fmt.Sprintf("Breakpoint %d set", created.ID)).PrintAndExit(getOutputFormat())
		},
	}
	breakCmd.Flags().StringVar(&breakCond, "cond", "", "Conditional expression")

	// clear
	clearCmd := &cobra.Command{
		Use:   "clear <id>",
		Short: "Clear a breakpoint by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("clear")
			defer func() { _ = c.Close() }()

			id, err := strconv.Atoi(args[0])
			if err != nil {
				output.ErrorWithInfo("clear", output.InvalidArgumentWithDetails(
					fmt.Sprintf("invalid breakpoint ID: %s", args[0]),
					map[string]any{"id": args[0]},
				)).PrintAndExit(getOutputFormat())
			}

			cleared, err := c.ClearBreakpoint(id)
			if err != nil {
				output.Error("clear", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"id":   cleared.ID,
				"file": cleared.File,
				"line": cleared.Line,
			}

			output.Success("clear", data, fmt.Sprintf("Breakpoint %d cleared", id)).PrintAndExit(getOutputFormat())
		},
	}

	// breakpoints
	breakpointsCmd := &cobra.Command{
		Use:   "breakpoints",
		Short: "List all breakpoints",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("breakpoints")
			defer func() { _ = c.Close() }()

			bps, err := c.ListBreakpoints()
			if err != nil {
				output.Error("breakpoints", err).PrintAndExit(getOutputFormat())
			}

			breakpoints := make([]map[string]any, 0, len(bps))
			for _, bp := range bps {
				if bp.ID < 0 {
					continue
				}

				bpData := map[string]any{
					"id":       bp.ID,
					"file":     bp.File,
					"line":     bp.Line,
					"function": bp.FunctionName,
					"enabled":  !bp.Disabled,
				}
				if bp.Cond != "" {
					bpData["condition"] = bp.Cond
				}
				if bp.TotalHitCount > 0 {
					bpData["hitCount"] = bp.TotalHitCount
				}
				breakpoints = append(breakpoints, bpData)
			}

			data := map[string]any{
				"breakpoints": breakpoints,
				"count":       len(breakpoints),
			}

			output.Success("breakpoints", data, fmt.Sprintf("%d breakpoints", len(breakpoints))).PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(breakCmd)
	root.AddCommand(clearCmd)
	root.AddCommand(breakpointsCmd)
}

// addInspectCommands adds variable inspection commands (locals, args, eval)
func addInspectCommands(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	// locals
	localsCmd := &cobra.Command{
		Use:   "locals",
		Short: "Show local variables",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("locals")
			defer func() { _ = c.Close() }()

			state, err := c.GetState()
			if err != nil {
				output.Error("locals", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("locals", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			vars, err := c.ListLocalVars(state.SelectedGoroutine.ID, 0, debugger.DefaultLoadConfig())
			if err != nil {
				output.Error("locals", err).PrintAndExit(getOutputFormat())
			}

			variables := make([]map[string]any, len(vars))
			for i, v := range vars {
				variables[i] = variableToMap(v)
			}

			data := map[string]any{
				"variables": variables,
				"count":     len(variables),
			}

			output.Success("locals", data, fmt.Sprintf("%d local variables", len(variables))).PrintAndExit(getOutputFormat())
		},
	}

	// args
	argsCmd := &cobra.Command{
		Use:   "args",
		Short: "Show function arguments",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("args")
			defer func() { _ = c.Close() }()

			state, err := c.GetState()
			if err != nil {
				output.Error("args", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("args", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			funcArgs, err := c.ListFunctionArgs(state.SelectedGoroutine.ID, 0, debugger.DefaultLoadConfig())
			if err != nil {
				output.Error("args", err).PrintAndExit(getOutputFormat())
			}

			arguments := make([]map[string]any, len(funcArgs))
			for i, v := range funcArgs {
				arguments[i] = variableToMap(v)
			}

			data := map[string]any{
				"arguments": arguments,
				"count":     len(arguments),
			}

			output.Success("args", data, fmt.Sprintf("%d arguments", len(arguments))).PrintAndExit(getOutputFormat())
		},
	}

	// eval
	evalCmd := &cobra.Command{
		Use:   "eval <expression>",
		Short: "Evaluate an expression",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("eval")
			defer func() { _ = c.Close() }()

			expr := args[0]

			state, err := c.GetState()
			if err != nil {
				output.Error("eval", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("eval", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			result, err := c.Eval(state.SelectedGoroutine.ID, 0, expr, debugger.DefaultLoadConfig())
			if err != nil {
				output.Error("eval", err).PrintAndExit(getOutputFormat())
			}

			data := variableToMap(*result)
			data["expression"] = expr

			output.Success("eval", data, "").PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(localsCmd)
	root.AddCommand(argsCmd)
	root.AddCommand(evalCmd)
}

// addNavigationCommands adds stack and goroutine navigation commands
func addNavigationCommands(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	var stackDepth int

	// stack
	stackCmd := &cobra.Command{
		Use:   "stack",
		Short: "Show stack trace",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("stack")
			defer func() { _ = c.Close() }()

			state, err := c.GetState()
			if err != nil {
				output.Error("stack", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("stack", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			cfg := debugger.DefaultLoadConfig()
			frames, err := c.Stacktrace(state.SelectedGoroutine.ID, stackDepth, &cfg)
			if err != nil {
				output.Error("stack", err).PrintAndExit(getOutputFormat())
			}

			stackFrames := make([]map[string]any, len(frames))
			for i, frame := range frames {
				frameData := map[string]any{
					"index": i,
					"file":  frame.File,
					"line":  frame.Line,
				}
				if frame.Function != nil {
					frameData["function"] = frame.Function.Name()
				}
				stackFrames[i] = frameData
			}

			data := map[string]any{
				"frames":      stackFrames,
				"count":       len(stackFrames),
				"goroutineId": state.SelectedGoroutine.ID,
			}

			output.Success("stack", data, fmt.Sprintf("%d frames", len(stackFrames))).PrintAndExit(getOutputFormat())
		},
	}
	stackCmd.Flags().IntVar(&stackDepth, "depth", 50, "Maximum stack depth")

	// frame
	frameCmd := &cobra.Command{
		Use:   "frame <index>",
		Short: "Switch to a stack frame",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("frame")
			defer func() { _ = c.Close() }()

			frameIdx, err := strconv.Atoi(args[0])
			if err != nil {
				output.ErrorWithInfo("frame", output.InvalidArgumentWithDetails(
					fmt.Sprintf("invalid frame index: %s", args[0]),
					map[string]any{"index": args[0]},
				)).PrintAndExit(getOutputFormat())
			}

			state, err := c.GetState()
			if err != nil {
				output.Error("frame", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("frame", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			cfg := debugger.DefaultLoadConfig()
			frames, err := c.Stacktrace(state.SelectedGoroutine.ID, frameIdx+1, &cfg)
			if err != nil {
				output.Error("frame", err).PrintAndExit(getOutputFormat())
			}

			if frameIdx >= len(frames) {
				output.ErrorWithInfo("frame", output.NotFound("frame", fmt.Sprintf("%d (stack has %d frames)", frameIdx, len(frames)))).PrintAndExit(getOutputFormat())
			}

			frame := frames[frameIdx]
			data := map[string]any{
				"index": frameIdx,
				"file":  frame.File,
				"line":  frame.Line,
			}
			if frame.Function != nil {
				data["function"] = frame.Function.Name()
			}

			output.Success("frame", data, fmt.Sprintf("Switched to frame %d", frameIdx)).PrintAndExit(getOutputFormat())
		},
	}

	// goroutines
	goroutinesCmd := &cobra.Command{
		Use:   "goroutines",
		Short: "List all goroutines",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("goroutines")
			defer func() { _ = c.Close() }()

			goroutines, _, err := c.ListGoroutines(0, 0)
			if err != nil {
				output.Error("goroutines", err).PrintAndExit(getOutputFormat())
			}

			state, _ := c.GetState()
			var selectedID int64
			if state != nil && state.SelectedGoroutine != nil {
				selectedID = state.SelectedGoroutine.ID
			}

			gs := make([]map[string]any, len(goroutines))
			for i, g := range goroutines {
				gData := map[string]any{
					"id":       g.ID,
					"selected": g.ID == selectedID,
				}
				if g.CurrentLoc.File != "" {
					gData["location"] = map[string]any{
						"file":     g.CurrentLoc.File,
						"line":     g.CurrentLoc.Line,
						"function": g.CurrentLoc.Function.Name(),
					}
				}
				if g.UserCurrentLoc.File != "" && g.UserCurrentLoc.File != g.CurrentLoc.File {
					gData["userLocation"] = map[string]any{
						"file":     g.UserCurrentLoc.File,
						"line":     g.UserCurrentLoc.Line,
						"function": g.UserCurrentLoc.Function.Name(),
					}
				}
				gs[i] = gData
			}

			data := map[string]any{
				"goroutines": gs,
				"count":      len(gs),
			}
			if selectedID > 0 {
				data["selectedId"] = selectedID
			}

			output.Success("goroutines", data, fmt.Sprintf("%d goroutines", len(gs))).PrintAndExit(getOutputFormat())
		},
	}

	// goroutine
	goroutineCmd := &cobra.Command{
		Use:   "goroutine <id>",
		Short: "Switch to a goroutine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("goroutine")
			defer func() { _ = c.Close() }()

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				output.ErrorWithInfo("goroutine", output.InvalidArgumentWithDetails(
					fmt.Sprintf("invalid goroutine ID: %s", args[0]),
					map[string]any{"id": args[0]},
				)).PrintAndExit(getOutputFormat())
			}

			state, err := c.SwitchGoroutine(id)
			if err != nil {
				output.Error("goroutine", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"id": id,
			}

			if state.SelectedGoroutine != nil {
				g := state.SelectedGoroutine
				if g.CurrentLoc.File != "" {
					data["location"] = map[string]any{
						"file":     g.CurrentLoc.File,
						"line":     g.CurrentLoc.Line,
						"function": g.CurrentLoc.Function.Name(),
					}
				}
			}

			output.Success("goroutine", data, fmt.Sprintf("Switched to goroutine %d", id)).PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(stackCmd)
	root.AddCommand(frameCmd)
	root.AddCommand(goroutinesCmd)
	root.AddCommand(goroutineCmd)
}

// addSourceCommands adds source viewing commands (list, sources)
func addSourceCommands(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	var listContext int

	// list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Show source code at current location",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("list")
			defer func() { _ = c.Close() }()

			state, err := c.GetState()
			if err != nil {
				output.Error("list", err).PrintAndExit(getOutputFormat())
			}

			if state.SelectedGoroutine == nil {
				output.ErrorWithInfo("list", output.NotFound("goroutine", "none selected")).PrintAndExit(getOutputFormat())
			}

			loc := state.SelectedGoroutine.CurrentLoc
			if loc.File == "" {
				output.ErrorWithInfo("list", output.NotFound("source location", "none available")).PrintAndExit(getOutputFormat())
			}

			// Read the source file
			file, err := os.Open(loc.File)
			if err != nil {
				output.ErrorWithInfo("list", output.NotFound("source file", loc.File)).PrintAndExit(getOutputFormat())
			}
			defer func() { _ = file.Close() }()

			startLine := loc.Line - listContext
			if startLine < 1 {
				startLine = 1
			}
			endLine := loc.Line + listContext

			// Read lines
			scanner := bufio.NewScanner(file)
			lineNum := 0
			var lines []map[string]any

			for scanner.Scan() {
				lineNum++
				if lineNum < startLine {
					continue
				}
				if lineNum > endLine {
					break
				}

				lineData := map[string]any{
					"lineNumber": lineNum,
					"content":    scanner.Text(),
					"current":    lineNum == loc.Line,
				}
				lines = append(lines, lineData)
			}

			if err := scanner.Err(); err != nil {
				output.Error("list", err).PrintAndExit(getOutputFormat())
			}

			data := map[string]any{
				"file":        loc.File,
				"currentLine": loc.Line,
				"lines":       lines,
			}
			if loc.Function != nil {
				data["function"] = loc.Function.Name()
			}

			output.Success("list", data, fmt.Sprintf("%s:%d", loc.File, loc.Line)).PrintAndExit(getOutputFormat())
		},
	}
	listCmd.Flags().IntVar(&listContext, "context", 5, "Lines of context before and after")

	// sources
	sourcesCmd := &cobra.Command{
		Use:   "sources [filter]",
		Short: "List all source files",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("sources")
			defer func() { _ = c.Close() }()

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}

			sources, err := c.ListSources(filter)
			if err != nil {
				output.Error("sources", err).PrintAndExit(getOutputFormat())
			}

			// Filter out runtime/internal sources for cleaner output
			var filtered []string
			for _, src := range sources {
				// Skip standard library and internal paths
				if strings.Contains(src, "/go/src/") ||
					strings.Contains(src, "/runtime/") ||
					strings.HasPrefix(src, "<") {
					continue
				}
				filtered = append(filtered, src)
			}

			data := map[string]any{
				"sources": filtered,
				"count":   len(filtered),
				"total":   len(sources),
			}

			output.Success("sources", data, fmt.Sprintf("%d source files", len(filtered))).PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(listCmd)
	root.AddCommand(sourcesCmd)
}

// addQuitCommand adds the quit command
func addQuitCommand(root *cobra.Command, mustGetClient func(string) *debugger.Client, getOutputFormat func() output.OutputFormat) {
	quitCmd := &cobra.Command{
		Use:   "quit",
		Short: "Stop debugging and terminate the debug server",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustGetClient("quit")
			// Note: don't defer close, we're detaching

			err := c.Detach(true)
			if err != nil {
				output.Error("quit", err).PrintAndExit(getOutputFormat())
			}

			output.Success("quit", nil, "Debug session terminated").PrintAndExit(getOutputFormat())
		},
	}

	root.AddCommand(quitCmd)
}
