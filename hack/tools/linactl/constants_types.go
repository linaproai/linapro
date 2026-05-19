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
	defaultBackendPort = 8080
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
}

// targetPlatform stores one normalized Go target platform.
type targetPlatform struct {
	OS   string
	Arch string
}
