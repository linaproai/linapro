// This file configures detached development-service processes on Windows.

//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// detachedProcessCreationFlag starts a child process detached from the parent console.
const detachedProcessCreationFlag = 0x00000008

// configureDetachedProcess lets development services outlive the linactl
// invocation that launched them.
func configureDetachedProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcessCreationFlag,
		HideWindow:    true,
	}
}
