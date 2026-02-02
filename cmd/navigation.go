package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/debugger"
	"github.com/8gears/godebug/internal/output"
)

var (
	stackDepth int
)

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Show stack trace",
	Long: `Show the current stack trace.

Options:
  --depth N   Maximum stack depth (default 50)

Example:
  godebug --addr $ADDR stack
  godebug --addr $ADDR stack --depth 20`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("stack")
		defer func() { _ = c.Close() }()

		state, err := c.GetState()
		if err != nil {
			output.Error("stack", err).Print(GetOutputFormat())
			return
		}

		if state.SelectedGoroutine == nil {
			output.ErrorMsg("stack", "no goroutine selected").Print(GetOutputFormat())
			return
		}

		cfg := debugger.DefaultLoadConfig()
		frames, err := c.Stacktrace(state.SelectedGoroutine.ID, stackDepth, &cfg)
		if err != nil {
			output.Error("stack", err).Print(GetOutputFormat())
			return
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

		output.Success("stack", data, fmt.Sprintf("%d frames", len(stackFrames))).Print(GetOutputFormat())
	},
}

var frameCmd = &cobra.Command{
	Use:   "frame <index>",
	Short: "Switch to a stack frame",
	Long: `Switch to a specific stack frame by index.

Frame 0 is the current (innermost) frame.

Example:
  godebug --addr $ADDR frame 2`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("frame")
		defer func() { _ = c.Close() }()

		frameIdx, err := strconv.Atoi(args[0])
		if err != nil {
			output.ErrorMsg("frame", fmt.Sprintf("invalid frame index: %s", args[0])).Print(GetOutputFormat())
			return
		}

		state, err := c.GetState()
		if err != nil {
			output.Error("frame", err).Print(GetOutputFormat())
			return
		}

		if state.SelectedGoroutine == nil {
			output.ErrorMsg("frame", "no goroutine selected").Print(GetOutputFormat())
			return
		}

		cfg := debugger.DefaultLoadConfig()
		frames, err := c.Stacktrace(state.SelectedGoroutine.ID, frameIdx+1, &cfg)
		if err != nil {
			output.Error("frame", err).Print(GetOutputFormat())
			return
		}

		if frameIdx >= len(frames) {
			output.ErrorMsg("frame", fmt.Sprintf("frame %d does not exist (stack has %d frames)", frameIdx, len(frames))).Print(GetOutputFormat())
			return
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

		output.Success("frame", data, fmt.Sprintf("Switched to frame %d", frameIdx)).Print(GetOutputFormat())
	},
}

var goroutinesCmd = &cobra.Command{
	Use:   "goroutines",
	Short: "List all goroutines",
	Long: `List all goroutines in the debugged process.

Example:
  godebug --addr $ADDR goroutines`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("goroutines")
		defer func() { _ = c.Close() }()

		goroutines, _, err := c.ListGoroutines(0, 0)
		if err != nil {
			output.Error("goroutines", err).Print(GetOutputFormat())
			return
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

		output.Success("goroutines", data, fmt.Sprintf("%d goroutines", len(gs))).Print(GetOutputFormat())
	},
}

var goroutineCmd = &cobra.Command{
	Use:   "goroutine <id>",
	Short: "Switch to a goroutine",
	Long: `Switch to a specific goroutine by ID.

Example:
  godebug --addr $ADDR goroutine 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("goroutine")
		defer func() { _ = c.Close() }()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			output.ErrorMsg("goroutine", fmt.Sprintf("invalid goroutine ID: %s", args[0])).Print(GetOutputFormat())
			return
		}

		state, err := c.SwitchGoroutine(id)
		if err != nil {
			output.Error("goroutine", err).Print(GetOutputFormat())
			return
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

		output.Success("goroutine", data, fmt.Sprintf("Switched to goroutine %d", id)).Print(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(stackCmd)
	rootCmd.AddCommand(frameCmd)
	rootCmd.AddCommand(goroutinesCmd)
	rootCmd.AddCommand(goroutineCmd)

	stackCmd.Flags().IntVar(&stackDepth, "depth", 50, "Maximum stack depth")
}
