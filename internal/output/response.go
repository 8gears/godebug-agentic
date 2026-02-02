package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// ExitFunc can be replaced in tests to prevent os.Exit from killing the test process
var ExitFunc = os.Exit

// Response is the standard JSON response envelope for all commands
type Response struct {
	Success bool        `json:"success"`
	Command string      `json:"command"`
	Data    any         `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// OutputFormat specifies the output format
type OutputFormat string

const (
	FormatJSON OutputFormat = "json"
	FormatText OutputFormat = "text"
)

// Print outputs the response in the specified format
func (r *Response) Print(format OutputFormat) {
	switch format {
	case FormatText:
		r.printText()
	default:
		r.printJSON()
	}
}

// PrintAndExit outputs the response and exits with the appropriate code
func (r *Response) PrintAndExit(format OutputFormat) {
	r.Print(format)
	ExitFunc(r.ExitCode())
}

// ExitCode returns the appropriate exit code based on the response status and error code
func (r *Response) ExitCode() int {
	if r.Success {
		return ExitSuccess
	}

	if r.Error == nil {
		return ExitGenericError
	}

	// Map error codes to exit codes
	switch r.Error.Code {
	case ErrCodeTimeout:
		return ExitTimeout
	case ErrCodeConnectionFailed, ErrCodeConnectionRefused:
		return ExitConnectionError
	case ErrCodeNotFound:
		return ExitNotFound
	case ErrCodeInvalidArgument:
		return ExitUsageError
	case ErrCodeProcessExited:
		return ExitProcessError
	default:
		return ExitGenericError
	}
}

func (r *Response) printJSON() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(r)
}

func (r *Response) printText() {
	if !r.Success {
		if r.Error != nil {
			fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", r.Error.Code, r.Error.Message)
		}
		return
	}
	if r.Message != "" {
		fmt.Println(r.Message)
	}
	if r.Data != nil {
		// Pretty print data for text mode
		data, _ := json.MarshalIndent(r.Data, "", "  ")
		fmt.Println(string(data))
	}
}

// Success creates a successful response
func Success(command string, data any, message string) *Response {
	return &Response{
		Success: true,
		Command: command,
		Data:    data,
		Message: message,
	}
}

// Error creates an error response from a standard error
// It classifies the error and creates appropriate ErrorInfo
func Error(command string, err error) *Response {
	return &Response{
		Success: false,
		Command: command,
		Error:   FromError(err),
	}
}

// ErrorWithInfo creates an error response with a specific ErrorInfo
func ErrorWithInfo(command string, errInfo *ErrorInfo) *Response {
	return &Response{
		Success: false,
		Command: command,
		Error:   errInfo,
	}
}

// ErrorMsg creates an error response with a message string
// Deprecated: Use ErrorWithInfo with proper error codes instead
func ErrorMsg(command string, msg string) *Response {
	return &Response{
		Success: false,
		Command: command,
		Error:   InternalError(msg),
	}
}
