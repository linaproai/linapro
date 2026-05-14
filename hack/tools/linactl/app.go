// This file wires the linactl application runtime and child command execution.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// newApp creates a command application with default process dependencies.
func newApp(stdout io.Writer, stderr io.Writer, stdin io.Reader) *app {
	return &app{
		stdout:      stdout,
		stderr:      stderr,
		stdin:       stdin,
		env:         os.Environ(),
		execCommand: exec.CommandContext,
		waitHTTP:    waitHTTP,
	}
}

// run parses the command and dispatches to the command handler.
func (a *app) run(ctx context.Context, args []string) error {
	repoRoot, err := discoverRepoRoot()
	if err != nil {
		return err
	}
	a.root = repoRoot

	if len(args) == 0 {
		return a.printHelp()
	}

	name := normalizeCommandName(args[0])
	if name == "help" {
		if len(args) > 1 {
			name = normalizeCommandName(args[1])
			if spec, ok := commandRegistry()[name]; ok {
				printCommandHelp(a.stdout, spec)
				return nil
			}
			return fmt.Errorf("unknown command %q", args[1])
		}
		return a.printHelp()
	}
	if name == "-h" || name == "--help" {
		return a.printHelp()
	}

	spec, ok := commandRegistry()[name]
	if !ok {
		return fmt.Errorf("unknown command %q; run linactl help", args[0])
	}

	input, err := parseCommandInput(args[1:])
	if err != nil {
		return err
	}
	if input.HasBool("help") || input.HasBool("h") {
		printCommandHelp(a.stdout, spec)
		return errHelpRequested
	}
	return spec.Run(ctx, a, input)
}

type commandOptions struct {
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

// runCommand executes a child command with consistent error messages.
func (a *app) runCommand(ctx context.Context, options commandOptions, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil && !filepath.IsAbs(name) {
		return fmt.Errorf("required tool %q is not available in PATH while running %s: %w", name, strings.Join(append([]string{name}, args...), " "), err)
	}

	cmd := a.execCommand(ctx, name, args...)
	if options.Dir != "" {
		cmd.Dir = options.Dir
	}
	if len(options.Env) > 0 {
		cmd.Env = options.Env
	} else {
		cmd.Env = a.env
	}
	cmd.Stdin = a.stdin

	stdout := options.Stdout
	stderr := options.Stderr
	if stdout == nil {
		stdout = a.stdout
	}
	if stderr == nil {
		stderr = a.stderr
	}
	if options.Quiet {
		var buffer bytes.Buffer
		cmd.Stdout = &buffer
		cmd.Stderr = &buffer
		err := cmd.Run()
		if err != nil {
			fmt.Fprint(stderr, buffer.String())
			return fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, args...), " "), err)
		}
		return nil
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}
	return nil
}
