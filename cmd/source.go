package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/8gears/godebug-agentic/internal/output"
)

var (
	listContext int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show source code at current location",
	Long: `Show source code around the current execution point.

Options:
  --context N   Number of lines before and after (default 5)

Example:
  godebug --addr $ADDR list
  godebug --addr $ADDR list --context 10`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("list")
		defer func() { _ = c.Close() }()

		state, err := c.GetState()
		if err != nil {
			output.Error("list", err).PrintAndExit(GetOutputFormat())
		}

		if state.SelectedGoroutine == nil {
			output.ErrorWithInfo("list", output.NotFound("goroutine", "none selected")).PrintAndExit(GetOutputFormat())
		}

		loc := state.SelectedGoroutine.CurrentLoc
		if loc.File == "" {
			output.ErrorWithInfo("list", output.NotFound("source location", "none available")).PrintAndExit(GetOutputFormat())
		}

		// Read the source file
		file, err := os.Open(loc.File)
		if err != nil {
			output.ErrorWithInfo("list", output.NotFound("source file", loc.File)).PrintAndExit(GetOutputFormat())
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
			output.Error("list", err).PrintAndExit(GetOutputFormat())
		}

		data := map[string]any{
			"file":        loc.File,
			"currentLine": loc.Line,
			"lines":       lines,
		}
		if loc.Function != nil {
			data["function"] = loc.Function.Name()
		}

		output.Success("list", data, fmt.Sprintf("%s:%d", loc.File, loc.Line)).PrintAndExit(GetOutputFormat())
	},
}

var sourcesCmd = &cobra.Command{
	Use:   "sources [filter]",
	Short: "List all source files",
	Long: `List all source files in the debugged program.

Optional filter argument matches file paths.

Example:
  godebug --addr $ADDR sources
  godebug --addr $ADDR sources main`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("sources")
		defer func() { _ = c.Close() }()

		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		sources, err := c.ListSources(filter)
		if err != nil {
			output.Error("sources", err).PrintAndExit(GetOutputFormat())
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

		output.Success("sources", data, fmt.Sprintf("%d source files", len(filtered))).PrintAndExit(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(sourcesCmd)

	listCmd.Flags().IntVar(&listContext, "context", 5, "Lines of context before and after")
}
