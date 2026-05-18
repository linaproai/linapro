// Package shellexec implements the guarded shell-task executor used by
// scheduled-job management.
package shellexec

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// maxCapturedOutputBytes bounds each output stream persisted into sys_job_log.
const maxCapturedOutputBytes = 64 * 1024

// truncatedOutputMarker is appended once captured output exceeds the configured limit.
const truncatedOutputMarker = "...[truncated]"

// Executor defines the shell-task execution contract.
type Executor interface {
	// Execute runs one shell command with timeout, environment merging, and output capture.
	Execute(ctx context.Context, in ExecuteInput) (*ExecuteOutput, error)
}

// ExecuteInput stores one shell-task execution request.
type ExecuteInput struct {
	ShellCmd string            // ShellCmd is executed through /bin/sh -c.
	WorkDir  string            // WorkDir overrides the host working directory when non-empty.
	Env      map[string]string // Env appends and overrides process environment variables.
	Timeout  time.Duration     // Timeout bounds the total execution time.
}

// ExecuteOutput stores one completed shell-task execution snapshot.
type ExecuteOutput struct {
	Stdout    string // Stdout stores the captured standard output.
	Stderr    string // Stderr stores the captured standard error.
	ExitCode  int    // ExitCode stores the process exit code.
	Cancelled bool   // Cancelled reports whether cancellation interrupted the process.
	TimedOut  bool   // TimedOut reports whether timeout interrupted the process.
}

// shellGate exposes the runtime shell switch needed by the executor.
type shellGate interface {
	// IsCronShellEnabled reports whether shell tasks are currently allowed.
	IsCronShellEnabled(ctx context.Context) (bool, error)
}

// serviceImpl implements Executor.
type serviceImpl struct {
	configSvc         shellGate     // configSvc exposes runtime shell switches.
	goos              string        // goos stores the current runtime platform.
	defaultWorkDir    string        // defaultWorkDir stores the host process working directory.
	cancelGracePeriod time.Duration // cancelGracePeriod bounds SIGTERM-to-SIGKILL escalation.
}

// Ensure serviceImpl implements Executor.
var _ Executor = (*serviceImpl)(nil)

// New creates and returns one guarded shell executor.
func New(configSvc shellGate) (Executor, error) {
	if configSvc == nil {
		return nil, gerror.New("shell executor requires a non-nil config service")
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	return &serviceImpl{
		configSvc:         configSvc,
		goos:              strings.TrimSpace(os.Getenv("GOOS")),
		defaultWorkDir:    workDir,
		cancelGracePeriod: 5 * time.Second,
	}, nil
}
