// Package runtimei18n implements repository-level runtime i18n governance
// checks for linactl. It owns hard-coded runtime copy scanning, host/plugin
// locale key parity, bizerr messageKey coverage (host + i18n-enabled plugins),
// plugin management display metadata keys (plugin.<id>.name/description),
// config-management display metadata keys (config.<sys_config.key>.name/remark),
// frontend static $t key coverage, and the default allowlist path used by the
// consolidated i18n.check command.
package runtimei18n

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	// commandScan runs the runtime-visible hard-coded copy scanner.
	commandScan = "scan"
	// commandMessages runs runtime i18n message key coverage validation.
	commandMessages = "messages"
	// commandFrontendKeys validates frontend static $t key references.
	commandFrontendKeys = "frontend-keys"
)

// RunCheck runs the scanner and message coverage checks, preserving the
// previous behavior of attempting message validation even when scanning fails.
func RunCheck(repoRoot string, out io.Writer) error {
	var (
		scanErr        = runAsError(repoRoot, []string{commandScan}, out)
		messageErr     = runAsError(repoRoot, []string{commandMessages}, out)
		frontendKeyErr = runAsError(repoRoot, []string{commandFrontendKeys}, out)
	)
	return errors.Join(scanErr, messageErr, frontendKeyErr)
}

// Run parses a runtime i18n subcommand and executes the selected verification
// flow against the provided repository root.
func Run(repoRoot string, args []string, out io.Writer) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("missing command, expected %s, %s, or %s", commandScan, commandMessages, commandFrontendKeys)
	}
	root := filepath.Clean(strings.TrimSpace(repoRoot))
	if root == "" || root == "." {
		return 1, fmt.Errorf("repository root is required")
	}

	switch strings.TrimSpace(args[0]) {
	case commandScan:
		return runScanCommand(root, args[1:], out)
	case commandMessages:
		return runMessagesCommand(root, args[1:], out)
	case commandFrontendKeys:
		return runFrontendKeysCommand(root, args[1:], out)
	default:
		return 1, fmt.Errorf("unknown command %q, expected %s, %s, or %s", args[0], commandScan, commandMessages, commandFrontendKeys)
	}
}

// runAsError converts command-style exit codes into errors for linactl command
// composition.
func runAsError(repoRoot string, args []string, out io.Writer) error {
	exitCode, err := Run(repoRoot, args, out)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("runtime i18n %s failed with exit code %d", args[0], exitCode)
	}
	return nil
}

// runScanCommand parses scanner flags and emits text or JSON findings.
func runScanCommand(repoRoot string, args []string, out io.Writer) (int, error) {
	options := scanOptions{
		allowlistPath: filepath.Join(repoRoot, "hack", "tools", "linactl", "internal", "runtimei18n", "allowlist.json"),
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

// runFrontendKeysCommand validates frontend static $t key references.
func runFrontendKeysCommand(repoRoot string, args []string, out io.Writer) (int, error) {
	flagSet := flag.NewFlagSet(commandFrontendKeys, flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	if err := flagSet.Parse(args); err != nil {
		return 1, fmt.Errorf("parse frontend key flags: %w", err)
	}

	errors, err := validateFrontendI18NKeyReferences(repoRoot)
	if err != nil {
		return 1, err
	}
	if err = emitFrontendKeyCoverage(out, errors); err != nil {
		return 1, err
	}
	if len(errors) > 0 {
		return 1, nil
	}
	return 0, nil
}
