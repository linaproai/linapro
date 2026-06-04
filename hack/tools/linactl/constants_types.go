// This file defines shared linactl constants and data types.

package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"time"

	"linactl/internal/process"
)

const (
	// defaultBackendPort is the standard backend development port.
	defaultBackendPort = 9120
	// defaultFrontendPort is the standard frontend development port.
	defaultFrontendPort = 5666
	// defaultWaitTimeout bounds development service readiness checks.
	defaultWaitTimeout = 60 * time.Second
	// defaultPortReleaseTimeout bounds how long dev waits for managed old
	// services to release their TCP ports after StopService sends a kill.
	defaultPortReleaseTimeout = 5 * time.Second
)

// errHelpRequested marks help output as a successful early return.
var errHelpRequested = errors.New("help requested")

// commandSpec describes one supported linactl command.
type commandSpec struct {
	Name        string
	Description string
	Usage       string
	Internal    bool
	Hidden      bool
	Run         func(context.Context, *app, commandInput) error
}

// commandInput stores parsed command arguments.
type commandInput struct {
	Args   []string
	Params map[string]string
}

// app stores one linactl invocation's process dependencies and repository paths.
type app struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader

	root string
	env  []string

	execCommand func(context.Context, string, ...string) *exec.Cmd
	executable  func() (string, error)
	lookPath    func(string) (string, error)
	waitHTTP    func(string, string, string, string, time.Duration) error
	// portInUse reports whether the given TCP port on localhost is currently
	// bound. It is exposed as an injectable field so unit tests can simulate
	// arbitrary port-availability scenarios without binding real sockets.
	// 端口占用检测函数，通过依赖注入便于单元测试覆盖端口可用与不可用两种场景。
	portInUse func(int) bool
	// processAlive reports whether the given PID currently belongs to a live
	// process. It is injectable for the same reason as portInUse: tests can
	// simulate "process exited" without spawning real subprocesses.
	// 进程存活检测函数，通过依赖注入便于单元测试覆盖。
	processAlive func(int) bool
	// processList returns visible operating-system processes. It lets dev
	// clean up current-project service processes that lost their PID file,
	// while tests can inject synthetic process tables.
	// 进程列表函数，用于识别丢失 PID 文件的当前项目服务进程，并便于单测注入。
	processList func() ([]process.Info, error)
	// processKill stops a process by PID. Tests inject this to avoid sending
	// signals to real user processes.
	// 进程停止函数，便于单测避免触碰真实系统进程。
	processKill func(int) error
}

// targetPlatform stores one normalized Go target platform.
type targetPlatform struct {
	OS   string
	Arch string
}
