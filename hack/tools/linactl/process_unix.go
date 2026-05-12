// This file configures detached development-service processes on Unix systems.

//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// configureDetachedProcess lets development services outlive the linactl
// invocation that launched them.
func configureDetachedProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
