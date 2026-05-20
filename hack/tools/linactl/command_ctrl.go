// This file implements the ctrl command for GoFrame controller generation.

package main

import (
	"context"

	"linactl/internal/goframecli"
)

// runCtrl runs gf gen ctrl in the core application directory.
func runCtrl(ctx context.Context, a *app, input commandInput) error {
	return goframecli.Run(ctx, a.root, a.runCommand, func(installCtx context.Context) error {
		return runCLIInstallIfMissing(installCtx, a, input)
	}, "gen", "ctrl")
}
