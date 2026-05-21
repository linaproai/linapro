// This file defines shared linactl constants and data types.

package main

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"time"
)

const (
	// defaultBackendPort is the standard backend development port.
	defaultBackendPort = 9120
	// defaultFrontendPort is the standard frontend development port.
	defaultFrontendPort = 5666
	// defaultWaitTimeout bounds development service readiness checks.
	defaultWaitTimeout = 60 * time.Second
)

// errHelpRequested marks help output as a successful early return.
var errHelpRequested = errors.New("help requested")

// commandSpec describes one supported linactl command.
type commandSpec struct {
	Name        string
	Description string
	Usage       string
	Internal    bool
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
}

// targetPlatform stores one normalized Go target platform.
type targetPlatform struct {
	OS   string
	Arch string
}
