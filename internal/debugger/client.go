package debugger

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/rpc2"
)

// Client wraps the Delve RPC2 client
type Client struct {
	addr   string
	client *rpc.Client
}

// Connect creates a new client connected to the Delve server
func Connect(addr string) (*Client, error) {
	client, err := jsonrpc.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	return &Client{addr: addr, client: client}, nil
}

// Close closes the connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Addr returns the server address
func (c *Client) Addr() string {
	return c.addr
}

// call is a helper for RPC calls
func (c *Client) call(method string, args any, reply any) error {
	return c.client.Call("RPCServer."+method, args, reply)
}

// GetState returns the current debugger state
func (c *Client) GetState() (*api.DebuggerState, error) {
	var state rpc2.StateOut
	err := c.call("State", rpc2.StateIn{NonBlocking: true}, &state)
	if err != nil {
		return nil, err
	}
	return state.State, nil
}

// Continue resumes execution until a breakpoint is hit
func (c *Client) Continue() (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{Name: api.Continue}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// Next steps over to the next source line
func (c *Client) Next() (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{Name: api.Next}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// Step steps into a function call
func (c *Client) Step() (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{Name: api.Step}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// StepOut steps out of the current function
func (c *Client) StepOut() (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{Name: api.StepOut}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// Halt stops execution
func (c *Client) Halt() (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{Name: api.Halt}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// Restart restarts the debugged process
func (c *Client) Restart() (*api.DebuggerState, error) {
	var out rpc2.RestartOut
	err := c.call("Restart", rpc2.RestartIn{}, &out)
	if err != nil {
		return nil, err
	}
	// Get fresh state after restart
	return c.GetState()
}

// CreateBreakpoint creates a new breakpoint
func (c *Client) CreateBreakpoint(bp *api.Breakpoint) (*api.Breakpoint, error) {
	var out rpc2.CreateBreakpointOut
	err := c.call("CreateBreakpoint", rpc2.CreateBreakpointIn{Breakpoint: *bp}, &out)
	if err != nil {
		return nil, err
	}
	return &out.Breakpoint, nil
}

// ClearBreakpoint removes a breakpoint by ID
func (c *Client) ClearBreakpoint(id int) (*api.Breakpoint, error) {
	var out rpc2.ClearBreakpointOut
	err := c.call("ClearBreakpoint", rpc2.ClearBreakpointIn{Id: id}, &out)
	if err != nil {
		return nil, err
	}
	return out.Breakpoint, nil
}

// ListBreakpoints returns all breakpoints
func (c *Client) ListBreakpoints() ([]*api.Breakpoint, error) {
	var out rpc2.ListBreakpointsOut
	err := c.call("ListBreakpoints", rpc2.ListBreakpointsIn{}, &out)
	if err != nil {
		return nil, err
	}
	return out.Breakpoints, nil
}

// ListLocalVars returns local variables in the current scope
func (c *Client) ListLocalVars(goroutineID int64, frame int, cfg api.LoadConfig) ([]api.Variable, error) {
	var out rpc2.ListLocalVarsOut
	err := c.call("ListLocalVars", rpc2.ListLocalVarsIn{
		Scope: api.EvalScope{
			GoroutineID:  goroutineID,
			Frame:        frame,
			DeferredCall: 0,
		},
		Cfg: cfg,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out.Variables, nil
}

// ListFunctionArgs returns function arguments
func (c *Client) ListFunctionArgs(goroutineID int64, frame int, cfg api.LoadConfig) ([]api.Variable, error) {
	var out rpc2.ListFunctionArgsOut
	err := c.call("ListFunctionArgs", rpc2.ListFunctionArgsIn{
		Scope: api.EvalScope{
			GoroutineID:  goroutineID,
			Frame:        frame,
			DeferredCall: 0,
		},
		Cfg: cfg,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out.Args, nil
}

// Eval evaluates an expression
func (c *Client) Eval(goroutineID int64, frame int, expr string, cfg api.LoadConfig) (*api.Variable, error) {
	var out rpc2.EvalOut
	err := c.call("Eval", rpc2.EvalIn{
		Scope: api.EvalScope{
			GoroutineID:  goroutineID,
			Frame:        frame,
			DeferredCall: 0,
		},
		Expr: expr,
		Cfg:  &cfg,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out.Variable, nil
}

// Stacktrace returns the stack trace
func (c *Client) Stacktrace(goroutineID int64, depth int, cfg *api.LoadConfig) ([]api.Stackframe, error) {
	var out rpc2.StacktraceOut
	err := c.call("Stacktrace", rpc2.StacktraceIn{
		Id:    goroutineID,
		Depth: depth,
		Cfg:   cfg,
	}, &out)
	if err != nil {
		return nil, err
	}
	return out.Locations, nil
}

// ListGoroutines returns all goroutines
func (c *Client) ListGoroutines(start, count int) ([]*api.Goroutine, int, error) {
	var out rpc2.ListGoroutinesOut
	err := c.call("ListGoroutines", rpc2.ListGoroutinesIn{
		Start: start,
		Count: count,
	}, &out)
	if err != nil {
		return nil, 0, err
	}
	return out.Goroutines, out.Nextg, nil
}

// SwitchGoroutine switches to a different goroutine
func (c *Client) SwitchGoroutine(goroutineID int64) (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{
		Name:        api.SwitchGoroutine,
		GoroutineID: goroutineID,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// SwitchThread switches to a different thread
func (c *Client) SwitchThread(threadID int) (*api.DebuggerState, error) {
	var out rpc2.CommandOut
	err := c.call("Command", &api.DebuggerCommand{
		Name:     api.SwitchThread,
		ThreadID: threadID,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out.State, nil
}

// ListSources returns all source files
func (c *Client) ListSources(filter string) ([]string, error) {
	var out rpc2.ListSourcesOut
	err := c.call("ListSources", rpc2.ListSourcesIn{Filter: filter}, &out)
	if err != nil {
		return nil, err
	}
	return out.Sources, nil
}

// Detach detaches from the debugged process
func (c *Client) Detach(kill bool) error {
	var out rpc2.DetachOut
	return c.call("Detach", rpc2.DetachIn{Kill: kill}, &out)
}

// CreateCheckpoint creates an execution checkpoint
func (c *Client) CreateCheckpoint(where string) (*api.Checkpoint, error) {
	var out rpc2.CheckpointOut
	err := c.call("Checkpoint", rpc2.CheckpointIn{Where: where}, &out)
	if err != nil {
		return nil, err
	}
	return &api.Checkpoint{ID: out.ID, Where: where}, nil
}

// ListCheckpoints returns all checkpoints
func (c *Client) ListCheckpoints() ([]api.Checkpoint, error) {
	var out rpc2.ListCheckpointsOut
	err := c.call("ListCheckpoints", rpc2.ListCheckpointsIn{}, &out)
	if err != nil {
		return nil, err
	}
	return out.Checkpoints, nil
}

// ClearCheckpoint removes a checkpoint
func (c *Client) ClearCheckpoint(id int) error {
	var out rpc2.ClearCheckpointOut
	return c.call("ClearCheckpoint", rpc2.ClearCheckpointIn{ID: id}, &out)
}

// ExamineMemory reads memory at the given address
func (c *Client) ExamineMemory(address uint64, length int) ([]byte, bool, error) {
	var out rpc2.ExaminedMemoryOut
	err := c.call("ExamineMemory", rpc2.ExamineMemoryIn{
		Address: address,
		Length:  length,
	}, &out)
	if err != nil {
		return nil, false, err
	}
	return out.Mem, out.IsLittleEndian, nil
}

// DefaultLoadConfig returns a sensible default config for loading variables
func DefaultLoadConfig() api.LoadConfig {
	return api.LoadConfig{
		FollowPointers:     true,
		MaxVariableRecurse: 3,
		MaxStringLen:       512,
		MaxArrayValues:     64,
		MaxStructFields:    -1,
	}
}
