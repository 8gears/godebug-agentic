package cmd

import (
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/output"
)

var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Stop debugging and terminate the debug server",
	Long: `Stop the debug session and terminate the debugged process.

This cleanly detaches from the process and shuts down the Delve server.

Example:
  godebug --addr 127.0.0.1:38697 quit`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("quit")
		// Note: don't defer close, we're detaching

		err := c.Detach(true)
		if err != nil {
			output.Error("quit", err).Print(GetOutputFormat())
			return
		}

		output.Success("quit", nil, "Debug session terminated").Print(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(quitCmd)
}
