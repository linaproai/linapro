// Package repository validates linactl and repository tooling conventions used
// by smoke checks. It keeps governance scans separate from the test.scripts
// command entrypoint and avoids platform-specific shell dependencies.
package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateTooling checks that local tooling entrypoints stay portable.
func ValidateTooling(root string, commandNames []string) error {
	makeCmd := filepath.Join(root, "make.cmd")
	content, err := os.ReadFile(makeCmd)
	if err != nil {
		return fmt.Errorf("read make.cmd wrapper: %w", err)
	}
	text := string(content)
	if !strings.Contains(text, "go run . %*") {
		return errors.New("make.cmd must delegate to linactl through go run . %*")
	}
	if strings.Contains(text, "GOWORK=off") {
		return errors.New("make.cmd must not force GOWORK=off")
	}
	linactlDir := filepath.Join(root, "hack", "tools", "linactl")
	if info, statErr := os.Stat(linactlDir); statErr == nil && info.IsDir() {
		if err = ValidateLinactlCommandFiles(root, commandNames); err != nil {
			return err
		}
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("stat linactl tool directory: %w", statErr)
	}

	legacyDir := filepath.Join(root, "hack", "scripts")
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy hack/scripts directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("hack/scripts contains legacy script %q; move maintained tooling into hack/tools/linactl or another Go tool", entries[0].Name())
	}
	return nil
}

// ValidateLinactlCommandFiles checks that concrete command implementations use
// command-name based filenames. Commands that collide with Go toolchain suffix
// rules use explicit command-specific suffixes documented in commandFileName.
func ValidateLinactlCommandFiles(root string, commandNames []string) error {
	linactlDir := filepath.Join(root, "hack", "tools", "linactl")
	for _, name := range commandNames {
		filename := CommandFileName(name)
		path := filepath.Join(linactlDir, filename)
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("linactl command %q must be implemented in %s", name, filename)
			}
			return fmt.Errorf("check linactl command file %s: %w", filename, err)
		}
	}
	if _, err := os.Stat(filepath.Join(linactlDir, "command_ops.go")); err == nil {
		return errors.New("command_ops.go is a legacy catch-all; move concrete commands into command_<name>.go files")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check legacy command_ops.go: %w", err)
	}
	return nil
}

// CommandFileName maps a linactl command to its implementation filename.
func CommandFileName(name string) string {
	switch name {
	case "test":
		// command_test.go would be excluded from normal builds as a test file.
		return "command_testcmd.go"
	case "wasm":
		// command_wasm.go would be treated as GOARCH=wasm-only.
		return "command_wasmcmd.go"
	default:
		return "command_" + name + ".go"
	}
}
