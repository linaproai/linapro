// Package shellexec implements the guarded shell-task executor used by
// scheduled-job management.
package shellexec

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
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
func New(configSvc shellGate) Executor {
	if configSvc == nil {
		panic("shell executor requires a non-nil config service")
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
	}
}

// Execute runs one shell command with timeout, environment merging, and output capture.
func (s *serviceImpl) Execute(ctx context.Context, in ExecuteInput) (*ExecuteOutput, error) {
	if s == nil {
		return nil, bizerr.NewCode(jobmeta.CodeJobShellExecutorUninitialized)
	}
	shellEnabled, err := s.configSvc.IsCronShellEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !shellEnabled {
		return nil, bizerr.NewCode(jobmeta.CodeJobShellDisabled)
	}
	commandText := strings.TrimSpace(in.ShellCmd)
	if commandText == "" {
		return nil, bizerr.NewCode(jobmeta.CodeJobShellCommandRequired)
	}
	if in.Timeout <= 0 {
		return nil, bizerr.NewCode(jobmeta.CodeJobShellTimeoutInvalid)
	}

	workDir, err := s.resolveWorkDir(in.WorkDir)
	if err != nil {
		return nil, err
	}

	execCtx, cancel := context.WithTimeout(ctx, in.Timeout)
	defer cancel()

	cmd := exec.Command("/bin/sh", "-c", commandText)
	cmd.Dir = workDir
	cmd.Env = mergeEnv(os.Environ(), in.Env)
	configureCommandProcess(cmd)

	var (
		stdoutBuffer = newLimitedBuffer(maxCapturedOutputBytes)
		stderrBuffer = newLimitedBuffer(maxCapturedOutputBytes)
	)
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer

	if err = cmd.Start(); err != nil {
		return nil, bizerr.WrapCode(err, jobmeta.CodeJobShellStartFailed)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	select {
	case waitErr := <-waitCh:
		return buildExecuteOutput(stdoutBuffer, stderrBuffer, waitErr), wrapCommandWaitError(waitErr)

	case <-execCtx.Done():
		out := s.cancelCommand(ctx, cmd, waitCh, execCtx.Err(), stdoutBuffer, stderrBuffer)
		return out, execCtx.Err()
	}
}

// resolveWorkDir validates and normalizes the requested working directory.
func (s *serviceImpl) resolveWorkDir(workDir string) (string, error) {
	trimmed := strings.TrimSpace(workDir)
	if trimmed == "" {
		return s.defaultWorkDir, nil
	}

	cleaned := filepath.Clean(trimmed)
	if cleaned == string(filepath.Separator) {
		return "", bizerr.NewCode(jobmeta.CodeJobShellWorkdirRootDenied)
	}
	info, err := os.Stat(cleaned)
	if err != nil {
		return "", bizerr.WrapCode(err, jobmeta.CodeJobShellWorkdirValidateFailed)
	}
	if !info.IsDir() {
		return "", bizerr.NewCode(jobmeta.CodeJobShellWorkdirNotDirectory)
	}
	return cleaned, nil
}

// cancelCommand terminates the running process group and waits for final exit.
func (s *serviceImpl) cancelCommand(
	ctx context.Context,
	cmd *exec.Cmd,
	waitCh <-chan error,
	cancelErr error,
	stdoutBuffer *limitedBuffer,
	stderrBuffer *limitedBuffer,
) *ExecuteOutput {
	if err := terminateProcessGroupGracefully(cmd.Process); err != nil {
		logger.Warningf(ctx, "terminate shell process group failed err=%v", err)
	}

	select {
	case waitErr := <-waitCh:
		return buildCancelledOutput(stdoutBuffer, stderrBuffer, waitErr, cancelErr)

	case <-time.After(s.cancelGracePeriod):
		if err := forceKillProcessGroup(cmd.Process); err != nil {
			logger.Warningf(ctx, "force kill shell process group failed err=%v", err)
		}
		waitErr := <-waitCh
		return buildCancelledOutput(stdoutBuffer, stderrBuffer, waitErr, cancelErr)
	}
}

// wrapCommandWaitError converts one process wait error into a user-facing error.
func wrapCommandWaitError(waitErr error) error {
	if waitErr == nil {
		return nil
	}
	return bizerr.WrapCode(waitErr, jobmeta.CodeJobShellExecutionFailed)
}

// buildExecuteOutput builds one completed shell result snapshot.
func buildExecuteOutput(
	stdoutBuffer *limitedBuffer,
	stderrBuffer *limitedBuffer,
	waitErr error,
) *ExecuteOutput {
	return &ExecuteOutput{
		Stdout:   stdoutBuffer.String(),
		Stderr:   stderrBuffer.String(),
		ExitCode: resolveExitCode(waitErr),
	}
}

// buildCancelledOutput builds one cancelled or timed-out shell result snapshot.
func buildCancelledOutput(
	stdoutBuffer *limitedBuffer,
	stderrBuffer *limitedBuffer,
	waitErr error,
	cancelErr error,
) *ExecuteOutput {
	out := buildExecuteOutput(stdoutBuffer, stderrBuffer, waitErr)
	out.Cancelled = cancelErr == context.Canceled
	out.TimedOut = cancelErr == context.DeadlineExceeded
	return out
}

// resolveExitCode extracts the process exit code when available.
func resolveExitCode(waitErr error) int {
	if waitErr == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errorsAs(waitErr, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

// mergeEnv overlays one environment map onto the base process environment.
func mergeEnv(base []string, overrides map[string]string) []string {
	envMap := make(map[string]string, len(base)+len(overrides))
	for _, item := range base {
		key, value, found := strings.Cut(item, "=")
		if !found {
			continue
		}
		envMap[key] = value
	}
	for key, value := range overrides {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			continue
		}
		envMap[trimmedKey] = value
	}

	keys := make([]string, 0, len(envMap))
	for key := range envMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+envMap[key])
	}
	return result
}

// limitedBuffer captures at most a fixed number of bytes and appends one
// truncation marker once the limit is exceeded.
type limitedBuffer struct {
	limit     int          // limit bounds the captured payload size.
	buffer    bytes.Buffer // buffer stores the captured output.
	truncated bool         // truncated reports whether the marker was already appended.
}

// newLimitedBuffer creates one bounded output buffer.
func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

// Write captures as much of the input as still fits in the buffer limit.
func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	if b == nil {
		return len(p), nil
	}
	if b.buffer.Len() >= b.limit {
		if !b.truncated {
			if _, err := b.buffer.WriteString(truncatedOutputMarker); err != nil {
				return 0, err
			}
			b.truncated = true
		}
		return len(p), nil
	}

	remaining := b.limit - b.buffer.Len()
	if len(p) <= remaining {
		if _, err := b.buffer.Write(p); err != nil {
			return 0, err
		}
		return len(p), nil
	}

	if _, err := b.buffer.Write(p[:remaining]); err != nil {
		return 0, err
	}
	if !b.truncated {
		if _, err := b.buffer.WriteString(truncatedOutputMarker); err != nil {
			return 0, err
		}
		b.truncated = true
	}
	return len(p), nil
}

// String returns the captured output.
func (b *limitedBuffer) String() string {
	if b == nil {
		return ""
	}
	return b.buffer.String()
}
