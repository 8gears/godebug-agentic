package cmd

import (
	"fmt"

	"github.com/go-delve/delve/service/api"
	"github.com/spf13/cobra"

	"github.com/8gears/godebug/internal/debugger"
	"github.com/8gears/godebug/internal/output"
)

// variableToMap converts a Variable to a map for JSON output
func variableToMap(v api.Variable) map[string]any {
	m := map[string]any{
		"name":  v.Name,
		"type":  v.Type,
		"value": v.Value,
	}

	// Include children for complex types
	if len(v.Children) > 0 {
		children := make([]map[string]any, len(v.Children))
		for i, child := range v.Children {
			children[i] = variableToMap(child)
		}
		m["children"] = children
	}

	return m
}

var localsCmd = &cobra.Command{
	Use:   "locals",
	Short: "Show local variables",
	Long: `List all local variables in the current scope.

Example:
  godebug --addr $ADDR locals`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("locals")
		defer func() { _ = c.Close() }()

		state, err := c.GetState()
		if err != nil {
			output.Error("locals", err).Print(GetOutputFormat())
			return
		}

		if state.SelectedGoroutine == nil {
			output.ErrorMsg("locals", "no goroutine selected").Print(GetOutputFormat())
			return
		}

		vars, err := c.ListLocalVars(state.SelectedGoroutine.ID, 0, debugger.DefaultLoadConfig())
		if err != nil {
			output.Error("locals", err).Print(GetOutputFormat())
			return
		}

		variables := make([]map[string]any, len(vars))
		for i, v := range vars {
			variables[i] = variableToMap(v)
		}

		data := map[string]any{
			"variables": variables,
			"count":     len(variables),
		}

		output.Success("locals", data, fmt.Sprintf("%d local variables", len(variables))).Print(GetOutputFormat())
	},
}

var argsCmd = &cobra.Command{
	Use:   "args",
	Short: "Show function arguments",
	Long: `List all arguments to the current function.

Example:
  godebug --addr $ADDR args`,
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("args")
		defer func() { _ = c.Close() }()

		state, err := c.GetState()
		if err != nil {
			output.Error("args", err).Print(GetOutputFormat())
			return
		}

		if state.SelectedGoroutine == nil {
			output.ErrorMsg("args", "no goroutine selected").Print(GetOutputFormat())
			return
		}

		funcArgs, err := c.ListFunctionArgs(state.SelectedGoroutine.ID, 0, debugger.DefaultLoadConfig())
		if err != nil {
			output.Error("args", err).Print(GetOutputFormat())
			return
		}

		arguments := make([]map[string]any, len(funcArgs))
		for i, v := range funcArgs {
			arguments[i] = variableToMap(v)
		}

		data := map[string]any{
			"arguments": arguments,
			"count":     len(arguments),
		}

		output.Success("args", data, fmt.Sprintf("%d arguments", len(arguments))).Print(GetOutputFormat())
	},
}

var evalCmd = &cobra.Command{
	Use:   "eval <expression>",
	Short: "Evaluate an expression",
	Long: `Evaluate a Go expression in the current context.

Examples:
  godebug --addr $ADDR eval "x"
  godebug --addr $ADDR eval "user.Name"
  godebug --addr $ADDR eval "len(items)"
  godebug --addr $ADDR eval "x > 10"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := MustGetClient("eval")
		defer func() { _ = c.Close() }()

		expr := args[0]

		state, err := c.GetState()
		if err != nil {
			output.Error("eval", err).Print(GetOutputFormat())
			return
		}

		if state.SelectedGoroutine == nil {
			output.ErrorMsg("eval", "no goroutine selected").Print(GetOutputFormat())
			return
		}

		result, err := c.Eval(state.SelectedGoroutine.ID, 0, expr, debugger.DefaultLoadConfig())
		if err != nil {
			output.Error("eval", err).Print(GetOutputFormat())
			return
		}

		data := variableToMap(*result)
		data["expression"] = expr

		output.Success("eval", data, "").Print(GetOutputFormat())
	},
}

func init() {
	rootCmd.AddCommand(localsCmd)
	rootCmd.AddCommand(argsCmd)
	rootCmd.AddCommand(evalCmd)
}
