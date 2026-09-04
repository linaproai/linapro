// This file implements ConfigureDetached and Alive on Windows.
// ConfigureDetached attaches DETACHED_PROCESS plus CREATE_NEW_PROCESS_GROUP
// so spawned services outlive the linactl console, and Alive opens the
// process with the minimal query right and inspects GetExitCodeProcess.

//go:build windows

package process

import (
	"log"
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

// ConfigureDetached lets development services outlive the linactl
// invocation that launched them.
func ConfigureDetached(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcessCreationFlag,
		HideWindow:    true,
	}
}

// Alive reports whether the given PID currently belongs to a live
// process on Windows. It opens the process with limited query rights and
// inspects the exit code: STILL_ACTIVE means the process is still running.
// Windows 平台下检测进程是否存活：使用最小权限打开进程并通过退出码判断。
// 句柄清理失败单独记录诊断，不覆盖查询得到的存活结果。
func Alive(pid int) bool {
	if pid <= 1 {
		return false
	}
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	alive, closeErr := aliveAndCloseHandle(handle, syscall.GetExitCodeProcess, syscall.CloseHandle)
	if closeErr != nil {
		log.Printf("警告：关闭进程 %d 的查询句柄失败：%v", pid, closeErr)
	}
	return alive
}

// aliveAndCloseHandle 查询已打开的进程句柄并关闭一次。
// 存活值仅由查询结果决定，返回的 error 仅表示清理失败。
// 系统调用作为显式参数传入，避免测试修改全局状态。
func aliveAndCloseHandle(
	handle syscall.Handle,
	query func(syscall.Handle, *uint32) error,
	closeHandle func(syscall.Handle) error,
) (bool, error) {
	var (
		code     uint32
		queryErr = query(handle, &code)
		closeErr = closeHandle(handle)
	)
	if queryErr != nil {
		return false, closeErr
	}
	return code == stillActiveExitCode, closeErr
}
