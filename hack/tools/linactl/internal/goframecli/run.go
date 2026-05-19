// Package goframecli wraps GoFrame CLI installation and generation execution
// for linactl. It keeps controller/DAO command files focused on the specific
// generation command while centralizing the shared "install if missing" flow.

package goframecli

import (
	"context"
	"path/filepath"

	"linactl/internal/toolrun"
)

// Installer installs GoFrame CLI when it is missing.
type Installer func(context.Context) error

// Run wraps a GoFrame CLI command inside the core application directory.
func Run(ctx context.Context, root string, run toolrun.Runner, install Installer, args ...string) error {
	if err := install(ctx); err != nil {
		return err
	}
	return run(ctx, toolrun.Options{Dir: filepath.Join(root, "apps", "lina-core")}, "gf", args...)
}
