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
	"time"

	"linactl/internal/devservice"
	"linactl/internal/fileutil"
	"linactl/internal/plugins"
	"linactl/internal/process"
	"linactl/internal/toolrun"
)

// newApp creates a command application with default process dependencies.
func newApp(stdout io.Writer, stderr io.Writer, stdin io.Reader) *app {
	a := &app{
		stdout:       stdout,
		stderr:       stderr,
		stdin:        stdin,
		env:          os.Environ(),
		execCommand:  exec.CommandContext,
		lookPath:     exec.LookPath,
		portInUse:    devservice.IsTCPListening,
		processAlive: process.Alive,
	}
	// Default waitHTTP wraps devservice.WaitHTTP so the readiness loop can
	// dispatch into the injectable processAlive on this app instance. Tests
	// override this field directly to control readiness behavior without
	// reaching into devservice internals.
	// 默认 waitHTTP 包装 devservice.WaitHTTP，并注入 app.processAlive，便于
	// 测试通过覆盖该字段直接控制就绪行为。
	a.waitHTTP = func(name string, url string, pidPath string, logPath string, timeout time.Duration) error {
		return devservice.WaitHTTP(name, url, pidPath, logPath, timeout, a.processAlive)
	}
	return a
}

// run parses the command and dispatches to the command handler.
func (a *app) run(ctx context.Context, args []string) error {
	repoRoot, err := fileutil.DiscoverRepoRoot()
	if err != nil {
		return err
	}
	a.root = repoRoot

	if len(args) == 0 {
		return a.printHelp(false)
	}

	name := normalizeCommandName(args[0])
	if name == "help" {
		input, err := parseCommandInput(args[1:])
		if err != nil {
			return err
		}
		if input.HasBool("all") {
			return a.printHelp(true)
		}
		if len(args) > 1 {
			name = normalizeCommandName(args[1])
			if spec, ok := commandRegistry()[name]; ok {
				printCommandHelp(a.stdout, spec)
				return nil
			}
			return fmt.Errorf("unknown command %q", args[1])
		}
		return a.printHelp(false)
	}
	if name == "-h" || name == "--help" {
		return a.printHelp(false)
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

type commandOptions = toolrun.Options

// runCommand executes a child command with consistent error messages.
func (a *app) runCommand(ctx context.Context, options commandOptions, name string, args ...string) error {
	if _, err := a.lookPath(name); err != nil && !filepath.IsAbs(name) {
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

// runCommandOutput executes a child command and returns stdout.
func (a *app) runCommandOutput(ctx context.Context, options commandOptions, name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	options.Quiet = false
	options.Stdout = &stdout
	options.Stderr = &stderr
	err := a.runCommand(ctx, options, name, args...)
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return "", err
	}
	return stdout.String(), nil
}

// pluginRuntime adapts the current app dependencies for plugin subcomponents.
func pluginRuntime(a *app) plugins.Runtime {
	return plugins.Runtime{
		Root:             a.root,
		Env:              a.env,
		Stdout:           a.stdout,
		Stderr:           a.stderr,
		RunCommand:       a.runCommand,
		RunCommandOutput: a.runCommandOutput,
	}
}

// prepareOfficialPluginBuildEnv resolves official plugin build mode using the
// plugins subcomponent while preserving the command-level app dependency shape.
func prepareOfficialPluginBuildEnv(ctx context.Context, a *app, input commandInput) (bool, []string, error) {
	return plugins.PrepareBuildEnv(ctx, pluginRuntime(a), input)
}
