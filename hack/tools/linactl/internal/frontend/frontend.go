// Package frontend provides dependency checks for linactl commands that run the
// Vite frontend. It owns the node_modules sentinel logic so command files only
// decide when frontend dependencies are required.
package frontend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"linactl/internal/toolutil"
)

// CommandRunner runs one child command with caller-provided working directory.
type CommandRunner func(context.Context, string, string, ...string) error

// EnsureDeps checks that the frontend node_modules are installed.
// If the vite binary is missing, it runs pnpm install automatically.
func EnsureDeps(ctx context.Context, root string, out io.Writer, run CommandRunner) error {
	vite := toolutil.ViteCommand(root)
	if _, err := os.Stat(vite); err == nil {
		return nil
	}
	fmt.Fprintln(out, "Frontend dependencies not installed; running pnpm install...")
	return run(ctx, filepath.Join(root, "apps", "lina-vben"), "pnpm", "install")
}
