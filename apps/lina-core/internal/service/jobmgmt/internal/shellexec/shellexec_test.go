// This file verifies shell-task execution guards, cancellation, and output truncation.

package shellexec

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeShellGate returns one fixed shell enablement state for tests.
type fakeShellGate struct {
	enabled bool
	err     error
}

// IsCronShellEnabled reports the configured shell enablement flag.
func (f fakeShellGate) IsCronShellEnabled(_ context.Context) (bool, error) {
	return f.enabled, f.err
}

// TestExecuteTruncatesOutput verifies stdout capture is bounded and marked as truncated.
func TestExecuteTruncatesOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell executor truncation test targets Unix-like shells")
	}

	svc := &serviceImpl{
		configSvc:         fakeShellGate{enabled: true},
		defaultWorkDir:    t.TempDir(),
		cancelGracePeriod: 100 * time.Millisecond,
	}

	out, err := svc.Execute(context.Background(), ExecuteInput{
		ShellCmd: "head -c 70000 /dev/zero | tr '\\000' 'a'",
		Timeout:  3 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected large output shell command to succeed, got error: %v", err)
	}
	if !strings.Contains(out.Stdout, truncatedOutputMarker) {
		t.Fatalf("expected stdout to contain truncation marker, got length=%d", len(out.Stdout))
	}
	if out.ExitCode != 0 {
		t.Fatalf("expected shell command exit code 0, got %d", out.ExitCode)
	}
}

// TestExecuteTimesOut verifies timeout cancellation terminates the process group.
func TestExecuteTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell executor timeout test targets Unix-like shells")
	}

	svc := &serviceImpl{
		configSvc:         fakeShellGate{enabled: true},
		defaultWorkDir:    t.TempDir(),
		cancelGracePeriod: 100 * time.Millisecond,
	}

	out, err := svc.Execute(context.Background(), ExecuteInput{
		ShellCmd: "sleep 5",
		Timeout:  200 * time.Millisecond,
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if out == nil || !out.TimedOut {
		t.Fatalf("expected output snapshot to report timeout, got %#v", out)
	}
}

// TestExecuteCanBeCancelled verifies manual cancellation stops the shell task.
func TestExecuteCanBeCancelled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell executor cancellation test targets Unix-like shells")
	}

	svc := &serviceImpl{
		configSvc:         fakeShellGate{enabled: true},
		defaultWorkDir:    t.TempDir(),
		cancelGracePeriod: 100 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	out, err := svc.Execute(ctx, ExecuteInput{
		ShellCmd: "sleep 5",
		Timeout:  3 * time.Second,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got %v", err)
	}
	if out == nil || !out.Cancelled {
		t.Fatalf("expected output snapshot to report cancellation, got %#v", out)
	}
}
