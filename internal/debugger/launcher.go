package debugger

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// LaunchMode specifies how to start the debugger
type LaunchMode string

const (
	ModeDebug LaunchMode = "debug" // dlv debug - compile and debug
	ModeTest  LaunchMode = "test"  // dlv test - debug tests
	ModeExec  LaunchMode = "exec"  // dlv exec - debug pre-compiled binary
)

// LaunchConfig holds configuration for launching Delve
type LaunchConfig struct {
	Mode       LaunchMode
	Target     string   // Path to package/binary
	Args       []string // Arguments to pass to the program
	BuildFlags string   // Additional build flags
}

// LaunchResult contains the result of launching Delve
type LaunchResult struct {
	Addr    string `json:"addr"`
	PID     int    `json:"pid"`
	Target  string `json:"target"`
	Mode    string `json:"mode"`
	process *os.Process
}

// Launch starts a Delve headless server
func Launch(config LaunchConfig) (*LaunchResult, error) {
	// Find dlv binary
	dlvPath, err := exec.LookPath("dlv")
	if err != nil {
		return nil, fmt.Errorf("dlv not found in PATH: %w", err)
	}

	// Build command arguments
	args := []string{string(config.Mode)}

	// Add target
	if config.Target != "" {
		args = append(args, config.Target)
	}

	// Add headless mode options
	args = append(args,
		"--headless",
		"--api-version=2",
		"--accept-multiclient",
		"--listen=127.0.0.1:0", // Let OS pick a port
	)

	// Note: Delve already uses -gcflags="all=-N -l" by default when compiling
	// so we don't need to specify build flags explicitly

	// Add program arguments after --
	if len(config.Args) > 0 {
		args = append(args, "--")
		args = append(args, config.Args...)
	}

	cmd := exec.Command(dlvPath, args...) //nolint:gosec // dlvPath is from exec.LookPath, args are controlled
	cmd.Dir = "."                         // Use current directory

	// Capture both stdout and stderr - dlv outputs to both
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dlv: %w", err)
	}

	// Parse the address from stdout/stderr
	addrChan := make(chan string, 1)
	errChan := make(chan error, 1)

	addrRegex := regexp.MustCompile(`API server listening at: (.+)`)

	// Scanner function for both pipes
	scanPipe := func(scanner *bufio.Scanner) {
		for scanner.Scan() {
			line := scanner.Text()
			if matches := addrRegex.FindStringSubmatch(line); len(matches) > 1 {
				select {
				case addrChan <- matches[1]:
				default:
				}
				return
			}
			// Check for errors
			if strings.Contains(line, "error") || strings.Contains(line, "Error") {
				select {
				case errChan <- fmt.Errorf("dlv error: %s", line):
				default:
				}
				return
			}
		}
	}

	go scanPipe(bufio.NewScanner(stdout))
	go scanPipe(bufio.NewScanner(stderr))

	// Wait for address or timeout
	select {
	case addr := <-addrChan:
		return &LaunchResult{
			Addr:    addr,
			PID:     cmd.Process.Pid,
			Target:  config.Target,
			Mode:    string(config.Mode),
			process: cmd.Process,
		}, nil
	case err := <-errChan:
		_ = cmd.Process.Kill()
		return nil, err
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("timeout waiting for dlv to start")
	}
}

// Kill terminates the Delve process
func (r *LaunchResult) Kill() error {
	if r.process != nil {
		return r.process.Kill()
	}
	return nil
}
