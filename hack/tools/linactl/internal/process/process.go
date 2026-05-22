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
