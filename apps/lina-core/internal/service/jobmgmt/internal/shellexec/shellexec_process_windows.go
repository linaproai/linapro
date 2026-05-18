//go:build windows

// This file keeps the shell executor buildable on Windows where process-group
// signalling is not used by the current iteration.

package shellexec

import (
	"errors"
	"os"
	"os/exec"
)

// configureCommandProcess is a no-op on Windows.
func configureCommandProcess(cmd *exec.Cmd) {}

// terminateProcessGroupGracefully falls back to regular process kill on Windows.
func terminateProcessGroupGracefully(process *os.Process) error {
	if process == nil {
		return nil
	}
	return process.Kill()
}

// forceKillProcessGroup falls back to regular process kill on Windows.
func forceKillProcessGroup(process *os.Process) error {
	if process == nil {
		return nil
	}
	return process.Kill()
}

// errorsAs delegates to the standard library errors.As on Windows platforms.
func errorsAs(err error, target any) bool {
	return errors.As(err, target)
}
