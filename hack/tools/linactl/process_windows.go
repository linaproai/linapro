// This file configures detached development-service processes on Windows.

//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// detachedProcessCreationFlag starts a child process detached from the parent console.
const detachedProcessCreationFlag = 0x00000008

// processQueryLimitedInformation is the minimal access right needed to query
// process exit status without elevated privileges.
// 仅用于查询进程退出码所需的最小访问权限。
const processQueryLimitedInformation = 0x1000

// stillActiveExitCode is the value Windows returns from GetExitCodeProcess
// while the process is still running.
// Windows 进程仍在运行时 GetExitCodeProcess 返回的固定值。
const stillActiveExitCode uint32 = 259

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

// processAlive reports whether the given PID currently belongs to a live
// process on Windows. It opens the process with limited query rights and
// inspects the exit code: STILL_ACTIVE means the process is still running.
// Windows 平台下检测进程是否存活：使用最小权限打开进程并通过退出码判断。
func processAlive(pid int) bool {
	if pid <= 1 {
		return false
	}
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	var code uint32
	if err = syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}
	return code == stillActiveExitCode
}
