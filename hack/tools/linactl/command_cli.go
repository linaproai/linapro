// This file implements the cli command for GoFrame CLI installation.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"linactl/internal/toolutil"
)

// runCLIInstall downloads and installs the GoFrame CLI for this platform.
func runCLIInstall(ctx context.Context, a *app, _ commandInput) error {
	tmpDir, err := os.MkdirTemp("", "linapro-gf-*")
	if err != nil {
		return err
	}
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			fmt.Fprintf(a.stderr, "warning: remove temporary gf directory: %v\n", removeErr)
		}
	}()

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	binary := toolutil.ExecutableName("gf")
	url := fmt.Sprintf("https://github.com/gogf/gf/releases/latest/download/gf_%s_%s", goos, goarch)
	archive := filepath.Join(tmpDir, binary)

	fmt.Fprintf(a.stdout, "Downloading GoFrame CLI: %s\n", url)
	if err = toolutil.DownloadFile(ctx, url, archive); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err = os.Chmod(archive, 0o755); err != nil {
			return fmt.Errorf("chmod gf binary: %w", err)
		}
	}
	return a.runCommand(ctx, commandOptions{Dir: tmpDir}, archive, "install", "-y")
}
