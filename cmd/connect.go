package cmd

import (
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/debugger"
	"github.com/8gears/godebug/internal/output"
)

var connectCmd = &cobra.Command{
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
			output.Error("connect", err).Print(GetOutputFormat())
			return
		}
		defer func() { _ = c.Close() }()

		// Verify connection by getting state
		state, err := c.GetState()
		if err != nil {
			output.Error("connect", err).Print(GetOutputFormat())
			return
		}

		data := map[string]any{
			"addr":    serverAddr,
			"running": state.Running,
		}
		if state.SelectedGoroutine != nil {
			data["goroutineId"] = state.SelectedGoroutine.ID
		}

		output.Success("connect", data, "Connected to debug server").Print(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
