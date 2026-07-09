// Package process provides cross-platform helpers for managing development
// service child processes spawned by linactl. It exposes two capabilities:
//
//  1. ConfigureDetached attaches platform-specific syscall attributes to an
//     exec.Cmd so the spawned service can outlive the linactl invocation
//     that started it (Setsid on Unix, DETACHED_PROCESS / CREATE_NEW_PROCESS_GROUP
//     on Windows). This lets `linactl dev` start long-running backend and
//     frontend processes that survive after the CLI returns.
//  2. Alive reports whether a previously recorded PID currently belongs to
//     a live, non-zombie process. It is used by `linactl status` and the
//     readiness loop in internal/devservice to distinguish a still-running
//     service from a stale PID file or a process that has fatal-exited but
//     not yet been reaped.
//
// Platform behavior is split across process_unix.go and process_windows.go.
// Both files implement the same exported surface so callers can depend on
// process.Alive and process.ConfigureDetached without writing build tags.
package process

import (
	"fmt"
	"os"
	"strings"
)

// Info describes one visible operating-system process.
type Info struct {
	PID  int
	Args []string
	CWD  string
}

// CommandLine returns a readable command string for matching and diagnostics.
func (i Info) CommandLine() string {
	return strings.Join(i.Args, " ")
}

// Kill sends a platform default termination request to the process.
func Kill(pid int) error {
	if pid <= 1 {
		return fmt.Errorf("invalid pid %d", pid)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
