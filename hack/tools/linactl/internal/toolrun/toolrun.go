// Package toolrun defines process execution contracts shared between linactl's
// root command package and internal components. It avoids coupling reusable
// components to the main package while keeping command execution centralized.
package toolrun

import (
	"context"
	"io"
)

// Options configures one child process invocation.
type Options struct {
	// Dir sets the child process working directory.
	Dir string
	// Env overrides the child process environment.
	Env []string
	// Quiet buffers child output unless the command fails.
	Quiet bool
	// Stdout overrides stdout forwarding.
	Stdout io.Writer
	// Stderr overrides stderr forwarding.
	Stderr io.Writer
}

// Runner executes a child process.
type Runner func(context.Context, Options, string, ...string) error

// OutputRunner executes a child process and returns stdout.
type OutputRunner func(context.Context, Options, string, ...string) (string, error)
