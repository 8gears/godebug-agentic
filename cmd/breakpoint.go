package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-delve/delve/service/api"
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/output"
)

var (
	breakCond string
)

var breakCmd = &cobra.Command{
	Use:   "break <location>",
	Short: "Set a breakpoint",
	Long: `Set a breakpoint at the specified location.

Location formats:
  file.go:line    - Set at file and line number
  pkg.Function    - Set at function entry

Options:
  --cond "expr"   - Only trigger when expression is true

Examples:
  godebug --addr $ADDR break main.go:42
  godebug --addr $ADDR break main.handleRequest
  godebug --addr $ADDR break main.go:42 --cond "x > 10"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("break")
		defer func() { _ = c.Close() }()

		location := args[0]
		bp := &api.Breakpoint{}

		// Parse location: file:line or function name
		if strings.Contains(location, ":") {
			parts := strings.SplitN(location, ":", 2)
			file := parts[0]
			line, err := strconv.Atoi(parts[1])
			if err != nil {
				output.ErrorMsg("break", fmt.Sprintf("invalid line number: %s", parts[1])).Print(GetOutputFormat())
				return
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
			output.Error("break", err).Print(GetOutputFormat())
			return
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

		output.Success("break", data, fmt.Sprintf("Breakpoint %d set", created.ID)).Print(GetOutputFormat())
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear <id>",
	Short: "Clear a breakpoint by ID",
	Long: `Remove a breakpoint by its ID.

Example:
  godebug --addr $ADDR clear 1`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("clear")
		defer func() { _ = c.Close() }()

		id, err := strconv.Atoi(args[0])
		if err != nil {
			output.ErrorMsg("clear", fmt.Sprintf("invalid breakpoint ID: %s", args[0])).Print(GetOutputFormat())
			return
		}

		cleared, err := c.ClearBreakpoint(id)
		if err != nil {
			output.Error("clear", err).Print(GetOutputFormat())
			return
		}

		data := map[string]any{
			"id":   cleared.ID,
			"file": cleared.File,
			"line": cleared.Line,
		}

		output.Success("clear", data, fmt.Sprintf("Breakpoint %d cleared", id)).Print(GetOutputFormat())
	},
}

var breakpointsCmd = &cobra.Command{
	Use:   "breakpoints",
	Short: "List all breakpoints",
	Long: `List all currently set breakpoints.

Example:
  godebug --addr $ADDR breakpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("breakpoints")
		defer func() { _ = c.Close() }()

		bps, err := c.ListBreakpoints()
		if err != nil {
			output.Error("breakpoints", err).Print(GetOutputFormat())
			return
		}

		breakpoints := make([]map[string]any, 0, len(bps))
		for _, bp := range bps {
			// Skip internal breakpoints (negative IDs or special names)
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

		output.Success("breakpoints", data, fmt.Sprintf("%d breakpoints", len(breakpoints))).Print(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(breakCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(breakpointsCmd)

	breakCmd.Flags().StringVar(&breakCond, "cond", "", "Conditional expression")
}
