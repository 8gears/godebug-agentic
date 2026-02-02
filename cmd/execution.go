package cmd

import (
	"github.com/go-delve/delve/service/api"
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/output"
)

// stateToData converts a DebuggerState to a response data map
func stateToData(state *api.DebuggerState) map[string]any {
	data := map[string]any{
		"running": state.Running,
		"exited":  state.Exited,
	}

	if state.Exited {
		data["exitStatus"] = state.ExitStatus
		return data
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

	// Include breakpoint info if we stopped at one
	if state.CurrentThread != nil && state.CurrentThread.Breakpoint != nil {
		bp := state.CurrentThread.Breakpoint
		data["breakpoint"] = map[string]any{
			"id":   bp.ID,
			"file": bp.File,
			"line": bp.Line,
		}
	}

	return data
}

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue execution until breakpoint",
	Long: `Continue execution until the next breakpoint is hit or the program exits.

Example:
  godebug --addr $ADDR continue`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("continue")
		defer func() { _ = c.Close() }()

		// Set the timeout from global flag
		c.SetTimeout(GetTimeout())

		state, err := c.Continue()
		if err != nil {
			output.Error("continue", err).PrintAndExit(GetOutputFormat())
		}

		var msg string
		if state.Exited {
			msg = "Process exited"
		} else if state.CurrentThread != nil && state.CurrentThread.Breakpoint != nil {
			msg = "Stopped at breakpoint"
		} else {
			msg = "Process stopped"
		}

		output.Success("continue", stateToData(state), msg).PrintAndExit(GetOutputFormat())
	},
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Step over to next source line",
	Long: `Step to the next source line, stepping over function calls.

Example:
  godebug --addr $ADDR next`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("next")
		defer func() { _ = c.Close() }()

		c.SetTimeout(GetTimeout())

		state, err := c.Next()
		if err != nil {
			output.Error("next", err).PrintAndExit(GetOutputFormat())
		}

		output.Success("next", stateToData(state), "Stepped to next line").PrintAndExit(GetOutputFormat())
	},
}

var stepCmd = &cobra.Command{
	Use:   "step",
	Short: "Step into function call",
	Long: `Step into the next function call.

Example:
  godebug --addr $ADDR step`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("step")
		defer func() { _ = c.Close() }()

		c.SetTimeout(GetTimeout())

		state, err := c.Step()
		if err != nil {
			output.Error("step", err).PrintAndExit(GetOutputFormat())
		}

		output.Success("step", stateToData(state), "Stepped into function").PrintAndExit(GetOutputFormat())
	},
}

var stepoutCmd = &cobra.Command{
	Use:   "stepout",
	Short: "Step out of current function",
	Long: `Step out of the current function to the caller.

Example:
  godebug --addr $ADDR stepout`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("stepout")
		defer func() { _ = c.Close() }()

		c.SetTimeout(GetTimeout())

		state, err := c.StepOut()
		if err != nil {
			output.Error("stepout", err).PrintAndExit(GetOutputFormat())
		}

		output.Success("stepout", stateToData(state), "Stepped out of function").PrintAndExit(GetOutputFormat())
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the debugged program",
	Long: `Restart the program from the beginning.

All breakpoints are preserved.

Example:
  godebug --addr $ADDR restart`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("restart")
		defer func() { _ = c.Close() }()

		state, err := c.Restart()
		if err != nil {
			output.Error("restart", err).PrintAndExit(GetOutputFormat())
		}

		output.Success("restart", stateToData(state), "Program restarted").PrintAndExit(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(continueCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(stepCmd)
	rootCmd.AddCommand(stepoutCmd)
	rootCmd.AddCommand(restartCmd)
}
