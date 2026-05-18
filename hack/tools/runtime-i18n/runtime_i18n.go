// This file implements the command dispatcher shared by runtime i18n scan and
// message-coverage checks.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// commandScan runs the runtime-visible hard-coded copy scanner.
	commandScan = "scan"
	// commandMessages runs runtime i18n message key coverage validation.
	commandMessages = "messages"
)

// run parses the top-level command and executes the selected verification flow.
func run(args []string, out io.Writer) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("missing command, expected %s or %s", commandScan, commandMessages)
	}

	repoRoot, err := resolveRepoRoot()
	if err != nil {
		return 1, err
	}

	switch strings.TrimSpace(args[0]) {
	case commandScan:
		return runScanCommand(repoRoot, args[1:], out)
	case commandMessages:
		return runMessagesCommand(repoRoot, args[1:], out)
	default:
		return 1, fmt.Errorf("unknown command %q, expected %s or %s", args[0], commandScan, commandMessages)
	}
}

// runScanCommand parses scanner flags and emits text or JSON findings.
func runScanCommand(repoRoot string, args []string, out io.Writer) (int, error) {
	options := scanOptions{
		allowlistPath: filepath.Join(repoRoot, "hack", "tools", "runtime-i18n", "allowlist.json"),
		format:        "text",
	}

	flagSet := flag.NewFlagSet(commandScan, flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.StringVar(&options.allowlistPath, "allowlist", options.allowlistPath, "JSON allowlist file path")
	flagSet.StringVar(&options.format, "format", options.format, "output format: text or json")
	if err := flagSet.Parse(args); err != nil {
		return 1, fmt.Errorf("parse scan flags: %w", err)
	}
	if options.format != "text" && options.format != "json" {
		return 1, fmt.Errorf("unsupported scan format %q", options.format)
	}

	report, err := scanRuntimeI18NReport(repoRoot, options)
	if err != nil {
		return 1, err
	}
	if err = emitScanReport(out, report, options.format); err != nil {
		return 1, err
	}
	if len(report.Findings) > 0 {
		return 1, nil
	}
	return 0, nil
}

// runMessagesCommand validates runtime i18n message key coverage.
func runMessagesCommand(repoRoot string, args []string, out io.Writer) (int, error) {
	flagSet := flag.NewFlagSet(commandMessages, flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	if err := flagSet.Parse(args); err != nil {
		return 1, fmt.Errorf("parse messages flags: %w", err)
	}

	errors, err := validateRuntimeI18NMessages(repoRoot)
	if err != nil {
		return 1, err
	}
	if err = emitMessageCoverage(out, errors); err != nil {
		return 1, err
	}
	if len(errors) > 0 {
		return 1, nil
	}
	return 0, nil
}

// resolveRepoRoot finds the repository root by walking upward from the current directory.
func resolveRepoRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	current, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute working directory: %w", err)
	}
	for {
		if isRepoRoot(current) {
			return filepath.Clean(current), nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("repository root not found from %s", workingDir)
}

// isRepoRoot reports whether dir has the expected LinaPro repository markers.
func isRepoRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "apps", "lina-core", "go.mod")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "apps", "lina-vben", "package.json")); err != nil {
		return false
	}
	return true
}
