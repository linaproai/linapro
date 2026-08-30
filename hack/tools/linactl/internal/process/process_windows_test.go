//go:build windows

// This file verifies Windows process liveness detection for live, invalid,
// reserved, and exited process identifiers.

package process

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

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
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for short-lived process: %v", err)
	}
	if Alive(pid) {
		t.Fatalf("Alive(%d) = true, want false after process exits", pid)
	}
}
