package cmd

import (
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/debugger"
	"github.com/8gears/godebug/internal/output"
)

var (
	startMode string
)

var startCmd = &cobra.Command{
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
			Timeout: GetTimeout(),
		}

		result, err := debugger.Launch(config)
		if err != nil {
			output.Error("start", err).PrintAndExit(GetOutputFormat())
		}

		data := map[string]any{
			"addr":   result.Addr,
			"pid":    result.PID,
			"target": result.Target,
			"mode":   result.Mode,
		}

		output.Success("start", data, "Debug server started").PrintAndExit(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&startMode, "mode", "debug", "Debug mode: debug, test, or exec")
}
