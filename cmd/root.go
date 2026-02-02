package cmd

import (
	"os"
	"time"

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
