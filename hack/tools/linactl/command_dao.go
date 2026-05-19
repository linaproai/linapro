// This file implements the dao command for GoFrame DAO generation.

package main

import (
	"context"

	"linactl/internal/goframecli"
)

// runDao runs gf gen dao in the core application directory.
func runDao(ctx context.Context, a *app, input commandInput) error {
	return goframecli.Run(ctx, a.root, a.runCommand, func(installCtx context.Context) error {
		return runCLIInstallIfMissing(installCtx, a, input)
	}, "gen", "dao")
}
