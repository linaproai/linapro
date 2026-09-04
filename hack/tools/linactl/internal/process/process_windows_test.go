//go:build windows

// 本文件验证 Windows 进程存活判断、系统调用失败时的句柄清理及测试子进程回收。

package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// TestAliveAndCloseHandle 验证清理错误不会覆盖存活结果，且查询失败后仍清理一次。
// 回归来源：https://github.com/linaproai/linapro/pull/103#issuecomment-5537085367
func TestAliveAndCloseHandle(t *testing.T) {
	closeFailure := errors.New("模拟关闭句柄失败")
	for _, tt := range []struct {
		name     string
		code     uint32
		queryErr error
		closeErr error
		want     bool
	}{
		{name: "running", code: 259, want: true},
		{name: "running_close_failed", code: 259, closeErr: closeFailure, want: true},
		{name: "exited", code: 0, want: false},
		{name: "exited_close_failed", code: 0, closeErr: closeFailure, want: false},
		{name: "query_failed", code: 259, queryErr: syscall.ERROR_ACCESS_DENIED, want: false},
		{name: "query_and_close_failed", code: 259, queryErr: syscall.ERROR_ACCESS_DENIED, closeErr: closeFailure, want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				handle     = syscall.Handle(123)
				queryCalls int
				closeCalls int
			)
			alive, closeErr := aliveAndCloseHandle(handle, func(got syscall.Handle, code *uint32) error {
				queryCalls++
				if got != handle || closeCalls != 0 {
					t.Errorf("查询句柄或调用顺序错误：handle=%d, closeCalls=%d", got, closeCalls)
				}
				*code = tt.code
				return tt.queryErr
			}, func(got syscall.Handle) error {
				closeCalls++
				if got != handle || queryCalls != 1 {
					t.Errorf("关闭句柄或调用顺序错误：handle=%d, queryCalls=%d", got, queryCalls)
				}
				return tt.closeErr
			})
			if alive != tt.want {
				t.Errorf("存活结果=%t，期望=%t", alive, tt.want)
			}
			if !errors.Is(closeErr, tt.closeErr) {
				t.Errorf("关闭错误=%v，期望=%v", closeErr, tt.closeErr)
			}
			if queryCalls != 1 || closeCalls != 1 {
				t.Errorf("查询和关闭必须各执行一次：queryCalls=%d, closeCalls=%d", queryCalls, closeCalls)
			}
		})
	}
}

func TestAliveReturnsTrueForCurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !Alive(pid) {
		t.Fatalf("Alive(%d) = false, want true for current process", pid)
	}
}

func TestAliveReturnsFalseForInvalidPIDs(t *testing.T) {
	for _, pid := range []int{-1, 0, 1} {
		t.Run(fmt.Sprintf("PID_%d", pid), func(t *testing.T) {
			if Alive(pid) {
				t.Fatalf("Alive(%d) = true, want false for invalid or reserved PID", pid)
			}
		})
	}
}

func TestAliveReturnsFalseAfterProcessExits(t *testing.T) {
	cmd := exec.Command("cmd.exe", "/c", "exit", "0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start short-lived process: %v", err)
	}
	t.Cleanup(func() {
		// Wait 已取得退出状态时，进程已经回收；避免再次 Kill 或 Wait。
		if cmd.ProcessState != nil {
			return
		}
		if killErr := cmd.Process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			t.Logf("终止测试子进程 %d：%v", cmd.Process.Pid, killErr)
		}
		if waitErr := cmd.Wait(); waitErr != nil {
			var exitErr *exec.ExitError
			// 清理主动终止进程会产生非零退出码，其他等待错误必须暴露。
			if !errors.As(waitErr, &exitErr) {
				t.Errorf("回收测试子进程 %d：%v", cmd.Process.Pid, waitErr)
			}
		}
	})
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for short-lived process: %v", err)
	}
	if Alive(pid) {
		t.Fatalf("Alive(%d) = true, want false after process exits", pid)
	}
}
