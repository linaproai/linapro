//go:build !windows

// This file contains Unix-specific process-group management for shell tasks.

package shellexec

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// configureCommandProcess places the shell command in its own process group.
func configureCommandProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProcessGroupGracefully sends SIGTERM to the process group.
func terminateProcessGroupGracefully(process *os.Process) error {
	return signalProcessGroup(process, syscall.SIGTERM)
}

// forceKillProcessGroup sends SIGKILL to the process group.
func forceKillProcessGroup(process *os.Process) error {
	return signalProcessGroup(process, syscall.SIGKILL)
}

// errorsAs delegates to the standard library errors.As on Unix platforms.
func errorsAs(err error, target any) bool {
	return errors.As(err, target)
}

// signalProcessGroup sends one signal to the target process group.
func signalProcessGroup(process *os.Process, signal syscall.Signal) error {
	if process == nil {
		return nil
	}
	if err := syscall.Kill(-process.Pid, signal); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}
