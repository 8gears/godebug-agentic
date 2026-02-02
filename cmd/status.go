package cmd

import (
	"github.com/spf13/cobra"

	"github.com/8gears/godebug-agentic/internal/output"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current debug state",
	Long: `Show the current state of the debug session.

Returns whether the process is running, paused, or exited,
along with the current location if paused.

Example:
  godebug --addr 127.0.0.1:38697 status`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("status")
		defer func() { _ = c.Close() }()

		state, err := c.GetState()
		if err != nil {
			output.Error("status", err).PrintAndExit(GetOutputFormat())
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

		output.Success("status", data, msg).PrintAndExit(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
