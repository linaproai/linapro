// This file executes the embedded GoFrame CLI in the hidden child command.

package goframecli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/cmd/gf/v2/gfcmd"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// RunEmbedded executes a whitelisted GoFrame code generation command in the
// current process. Callers should use it only from linactl's hidden child
// command so GoFrame CLI exits cannot terminate the parent linactl process.
func RunEmbedded(ctx context.Context, configDir string, args ...string) error {
	if err := validateArgs(args); err != nil {
		return err
	}
	cleanup, err := configureGoFrameCLI(configDir, args[1] == "dao")
	if err != nil {
		return err
	}
	defer cleanup()
	command, err := gfcmd.GetCommand(ctx)
	if err != nil {
		return fmt.Errorf("initialize embedded GoFrame CLI: %w", err)
	}
	if command == nil {
		return fmt.Errorf("initialize embedded GoFrame CLI: command is nil")
	}
	_, err = command.RunWithSpecificArgs(ctx, append([]string{"gf"}, args...))
	if err != nil {
		return fmt.Errorf("run embedded GoFrame %s %s: %w", args[0], args[1], err)
	}
	return nil
}

// configureGoFrameCLI sets the gf CLI config directory without calling
// gfcmd.Command.Run, whose default lookup path can override the explicit
// plugin-root config directory and whose error path can exit the process.
func configureGoFrameCLI(configDir string, requireConfig bool) (func(), error) {
	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		return func() {}, nil
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve embedded GoFrame working directory: %w", err)
	}
	if err := ValidateProjectDir(workingDir); err != nil {
		return nil, err
	}
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return nil, fmt.Errorf("embedded GoFrame config directory is empty")
	}
	if requireConfig {
		if err := ValidateConfigDir(configDir); err != nil {
			return nil, err
		}
	}
	configPath := configDir
	cleanup := func() {}
	if !requireConfig {
		if _, err := os.Stat(filepath.Join(configDir, "config.yaml")); err != nil {
			if os.IsNotExist(err) {
				tempDir, mkErr := os.MkdirTemp("", "linactl-goframe-config-*")
				if mkErr != nil {
					return nil, fmt.Errorf("create temporary embedded GoFrame config directory: %w", mkErr)
				}
				configPath = tempDir
				cleanup = func() {
					if removeErr := os.RemoveAll(tempDir); removeErr != nil {
						fmt.Fprintf(os.Stderr, "warning: remove temporary embedded GoFrame config directory %s: %v\n", tempDir, removeErr)
					}
				}
			} else {
				return nil, fmt.Errorf("check embedded GoFrame CLI config: %w", err)
			}
		}
	}
	if err := adapter.SetPath(configPath); err != nil {
		cleanup()
		return nil, fmt.Errorf("configure embedded GoFrame CLI: %w", err)
	}
	return cleanup, nil
}
