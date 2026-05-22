// This file implements ConfigureDetached and Alive for Unix-like platforms.
// ConfigureDetached enables Setsid so spawned services outlive the linactl
// invocation, and Alive combines kill(0) with /proc/<pid>/status (Linux) to
// detect liveness while explicitly rejecting zombie processes.

//go:build !windows

package process

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// ConfigureDetached lets development services outlive the linactl
// invocation that launched them by attaching Setsid to the child process.
func ConfigureDetached(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// Alive reports whether the given PID currently belongs to a live process.
//
// We combine two checks because each one has a known blind spot:
//  1. syscall.Kill(pid, 0) reports success on zombie processes too, so a
//     fatal-exited-but-not-yet-reaped child would be treated as alive. We
//     mitigate this by reading /proc/<pid>/status when available and
//     rejecting the "Z" (zombie) state. This is Linux-specific; on non-Linux
//     Unix systems /proc may be absent and the kill check stands alone.
//  2. os.FindProcess + Process.Signal can be polluted by Go's pidfd
//     caching when the same Go process previously started this PID via
//     exec.Cmd, so we deliberately call syscall.Kill directly instead of
//     going through os.Process.
//
// 双重校验：先用 kill(0) 排除 ESRCH，再读 /proc/<pid>/status 排除 zombie 状态；
// 直接调用 syscall.Kill 绕开 os.Process pidfd 缓存导致的误报。
func Alive(pid int) bool {
	if pid <= 1 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err != nil {
		// EPERM means the process exists but the caller cannot signal it.
		// EPERM 表示进程存在但当前用户无权限，仍视为存活。
		if errors.Is(err, syscall.EPERM) {
			return true
		}
		// ESRCH and other errors (including the kernel-level "no such process")
		// indicate the PID does not refer to a running process.
		// ESRCH 等错误意味着 PID 已不对应任何运行中的进程。
		return false
	}
	// kill(0) succeeded; reject zombies via /proc when available.
	// 通过 /proc/<pid>/status 排除 zombie，避免对未被 reap 的已退出进程误报存活。
	if isZombieStatus(pid) {
		return false
	}
	return true
}

// isZombieStatus reports whether /proc/<pid>/status (Linux) lists the process
// in the "Z" state. Returns false when /proc is unavailable or the file
// cannot be read so non-Linux Unix platforms fall back to the kill(0) result.
// 读取 /proc/<pid>/status 判断是否为 zombie；/proc 不可用时返回 false 不阻塞判断。
func isZombieStatus(pid int) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "State:") {
			continue
		}
		// Example: "State:\tZ (zombie)"
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return fields[1] == "Z"
		}
	}
	return false
}
