package output

import "fmt"

// Error codes for machine-readable error classification
const (
	// ErrCodeConnectionFailed indicates the client cannot reach the Delve server
	ErrCodeConnectionFailed = "CONNECTION_FAILED"

	// ErrCodeConnectionRefused indicates the server actively refused the connection
	ErrCodeConnectionRefused = "CONNECTION_REFUSED"

	// ErrCodeTimeout indicates the operation exceeded its time limit
	ErrCodeTimeout = "TIMEOUT"

	// ErrCodeInvalidArgument indicates bad input from the user
	ErrCodeInvalidArgument = "INVALID_ARGUMENT"

	// ErrCodeNotFound indicates a requested resource doesn't exist (breakpoint, goroutine, frame)
	ErrCodeNotFound = "NOT_FOUND"

	// ErrCodeProcessExited indicates the target program terminated
	ErrCodeProcessExited = "PROCESS_EXITED"

	// ErrCodeEvalFailed indicates expression evaluation failed
	ErrCodeEvalFailed = "EVAL_FAILED"

	// ErrCodeInternalError indicates an unexpected internal error
	ErrCodeInternalError = "INTERNAL_ERROR"
)

// ErrorInfo provides structured error information for AI consumption
type ErrorInfo struct {
	Code    string `json:"code"`              // Machine-readable error code
	Message string `json:"message"`           // Human-readable description
	Details any    `json:"details,omitempty"` // Additional context
}

// Error implements the error interface
func (e *ErrorInfo) Error() string {
	return e.Message
}

// NewErrorInfo creates a new ErrorInfo with the given code and message
func NewErrorInfo(code, message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    code,
		Message: message,
	}
}

// WithDetails returns a copy of the ErrorInfo with additional details
func (e *ErrorInfo) WithDetails(details any) *ErrorInfo {
	return &ErrorInfo{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// ConnectionFailed creates an error for connection failures
func ConnectionFailed(addr string, err error) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeConnectionFailed,
		Message: fmt.Sprintf("cannot connect to Delve server at %s: %v", addr, err),
		Details: map[string]any{"addr": addr},
	}
}

// ConnectionRefused creates an error for connection refused
func ConnectionRefused(addr string) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeConnectionRefused,
		Message: fmt.Sprintf("connection refused by Delve server at %s", addr),
		Details: map[string]any{"addr": addr},
	}
}

// Timeout creates an error for operation timeouts
func Timeout(operation string, duration float64) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeTimeout,
		Message: fmt.Sprintf("operation timed out after %.1fs", duration),
		Details: map[string]any{
			"operation":       operation,
			"timeout_seconds": duration,
		},
	}
}

// InvalidArgument creates an error for invalid user input
func InvalidArgument(message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeInvalidArgument,
		Message: message,
	}
}

// InvalidArgumentWithDetails creates an error for invalid user input with details
func InvalidArgumentWithDetails(message string, details any) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeInvalidArgument,
		Message: message,
		Details: details,
	}
}

// NotFound creates an error for missing resources
func NotFound(resourceType, identifier string) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found: %s", resourceType, identifier),
		Details: map[string]any{
			"resource_type": resourceType,
			"identifier":    identifier,
		},
	}
}

// ProcessExited creates an error for when the target process has terminated
func ProcessExited(exitStatus int) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeProcessExited,
		Message: fmt.Sprintf("target process exited with status %d", exitStatus),
		Details: map[string]any{"exit_status": exitStatus},
	}
}

// EvalFailed creates an error for expression evaluation failures
func EvalFailed(expr string, err error) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeEvalFailed,
		Message: fmt.Sprintf("failed to evaluate expression '%s': %v", expr, err),
		Details: map[string]any{"expression": expr},
	}
}

// InternalError creates an error for unexpected internal errors
func InternalError(message string) *ErrorInfo {
	return &ErrorInfo{
		Code:    ErrCodeInternalError,
		Message: message,
	}
}

// FromError creates an ErrorInfo from a standard error
// It attempts to classify the error based on its message
func FromError(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	// Check if it's already an ErrorInfo
	if ei, ok := err.(*ErrorInfo); ok {
		return ei
	}

	msg := err.Error()

	// Try to classify common error patterns
	switch {
	case contains(msg, "connection refused"):
		return NewErrorInfo(ErrCodeConnectionRefused, msg)
	case contains(msg, "timeout") || contains(msg, "timed out"):
		return NewErrorInfo(ErrCodeTimeout, msg)
	case contains(msg, "not found") || contains(msg, "does not exist"):
		return NewErrorInfo(ErrCodeNotFound, msg)
	case contains(msg, "failed to connect") || contains(msg, "connection failed"):
		return NewErrorInfo(ErrCodeConnectionFailed, msg)
	case contains(msg, "process exited") || contains(msg, "has exited"):
		return NewErrorInfo(ErrCodeProcessExited, msg)
	default:
		return NewErrorInfo(ErrCodeInternalError, msg)
	}
}

// contains performs case-insensitive substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(toLower(s), toLower(substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
