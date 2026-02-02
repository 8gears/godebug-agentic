package output

// Exit codes for AI-friendly CLI consumption
// These semantic codes allow agents to branch logic without parsing JSON
const (
	// ExitSuccess indicates the command completed successfully
	ExitSuccess = 0

	// ExitGenericError indicates an unspecified error occurred
	ExitGenericError = 1

	// ExitUsageError indicates invalid arguments or flags
	ExitUsageError = 2

	// ExitConnectionError indicates failure to connect to the Delve server
	ExitConnectionError = 3

	// ExitNotFound indicates a requested resource was not found (breakpoint, goroutine, etc.)
	ExitNotFound = 4

	// ExitTimeout indicates the operation timed out (matches GNU timeout convention)
	ExitTimeout = 124

	// ExitProcessError indicates a target process error
	ExitProcessError = 125
)
